package vpn

import (
	"errors"
	"fmt"
	"log"
	"net"
	"sync"

	"lmvpn/internal/model"

	"github.com/apparentlymart/go-cidr/cidr"
)

type VpnService struct {
	mu       sync.RWMutex
	settings model.VpnSetting
	net      *net.IPNet
	serverIP net.IP
	prefix   int
	alloc    *AllocationManager
	switchx  *PacketSwitch
	tun      *TUNInterface
	tunDone  chan struct{}
	running  bool
	clients  map[*tunnelConn]struct{}
}

func NewVpnService() *VpnService {
	return &VpnService{
		clients: make(map[*tunnelConn]struct{}),
	}
}

func (s *VpnService) Running() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

func (s *VpnService) Settings() model.VpnSetting {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.settings
}

func (s *VpnService) parseNet(subnet string) (*net.IPNet, net.IP, int, error) {
	_, ipNet, err := net.ParseCIDR(subnet)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("子网格式错误: %w", err)
	}
	ones, _ := ipNet.Mask.Size()
	serverIP, err := cidr.Host(ipNet, 1)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("计算服务器 IP 失败: %w", err)
	}
	return ipNet, serverIP, ones, nil
}

func (s *VpnService) ApplySettings(settings model.VpnSetting, reservations map[uint]string) error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		_ = s.Stop()
		s.mu.Lock()
	}
	s.settings = settings
	s.mu.Unlock()

	if !settings.Enabled {
		return nil
	}

	ipNet, serverIP, prefix, err := s.parseNet(settings.Subnet)
	if err != nil {
		return err
	}

	tun, err := CreateTUN(settings.InterfaceName)
	if err != nil {
		return err
	}

	var peerIP net.IP = serverIP
	if settings.DoLocalIPConfig {
		if err := tun.Configure(serverIP, prefix, peerIP); err != nil {
			_ = tun.Close()
			return fmt.Errorf("配置 TUN 失败: %w", err)
		}
	}
	if err := tun.SetMTU(settings.MTU); err != nil {
		log.Printf("警告: 设置 MTU 失败: %v", err)
	}

	s.mu.Lock()
	s.net = ipNet
	s.serverIP = serverIP
	s.prefix = prefix
	s.alloc = NewAllocationManager(ipNet, serverIP, reservations)
	s.switchx = NewPacketSwitch(settings.AllowClientToClient)
	s.tun = tun
	s.tunDone = make(chan struct{})
	s.running = true
	s.mu.Unlock()

	go s.serveTUN()
	log.Printf("VPN 服务已启动: tun=%s subnet=%s server=%s mtu=%d", tun.Name(), ipNet.String(), serverIP.String(), settings.MTU)
	return nil
}

func (s *VpnService) serveTUN() {
	s.mu.RLock()
	tun := s.tun
	switchx := s.switchx
	done := s.tunDone
	bufSize := s.settings.MTU + 64
	s.mu.RUnlock()

	packet := make([]byte, bufSize)
	for {
		n, err := tun.Iface.Read(packet)
		if err != nil {
			log.Printf("TUN 读取结束: %v", err)
			close(done)
			return
		}
		if n < 1 {
			continue
		}
		targets := switchx.RouteFromTUN(packet[:n])
		for _, t := range targets {
			_ = t.WritePacket(packet[:n])
		}
	}
}

func (s *VpnService) Stop() error {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return nil
	}
	s.running = false
	tun := s.tun
	done := s.tunDone
	clients := s.clients
	s.clients = make(map[*tunnelConn]struct{})
	s.mu.Unlock()

	for c := range clients {
		c.close()
	}
	if tun != nil {
		_ = tun.Close()
		if done != nil {
			<-done
		}
	}
	log.Printf("VPN 服务已停止")
	return nil
}

func (s *VpnService) Allocate(user *model.User) (net.IP, error) {
	s.mu.RLock()
	alloc := s.alloc
	s.mu.RUnlock()
	if alloc == nil {
		return nil, errors.New("VPN 服务未运行")
	}
	return alloc.Allocate(user.ID)
}

func (s *VpnService) WriteToTUN(packet []byte) error {
	s.mu.RLock()
	tun := s.tun
	s.mu.RUnlock()
	if tun == nil {
		return errors.New("TUN 未就绪")
	}
	_, err := tun.Iface.Write(packet)
	return err
}

func (s *VpnService) RouteFromClient(src SwitchConn, packet []byte) []SwitchConn {
	s.mu.RLock()
	switchx := s.switchx
	s.mu.RUnlock()
	if switchx == nil {
		return nil
	}
	return switchx.RouteFromClient(src, packet)
}

func (s *VpnService) registerClient(c *tunnelConn) {
	s.mu.Lock()
	s.switchx.Register(c)
	s.clients[c] = struct{}{}
	s.mu.Unlock()
}

func (s *VpnService) unregisterClient(c *tunnelConn) {
	s.mu.Lock()
	if s.switchx != nil {
		s.switchx.Unregister(c)
	}
	delete(s.clients, c)
	if s.alloc != nil {
		s.alloc.Release(c.assignedIP)
	}
	s.mu.Unlock()
}

func (s *VpnService) ServerIP() net.IP {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.serverIP
}

func (s *VpnService) Prefix() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.prefix
}

func (s *VpnService) AllocStats() (used int, capacity uint64) {
	s.mu.RLock()
	alloc := s.alloc
	s.mu.RUnlock()
	if alloc == nil {
		return 0, 0
	}
	return alloc.UsedCount(), alloc.Capacity()
}

func (s *VpnService) AddReservation(userID uint, ipStr string) {
	s.mu.RLock()
	alloc := s.alloc
	s.mu.RUnlock()
	if alloc != nil {
		alloc.AddReservation(userID, ipStr)
	}
}

func (s *VpnService) RemoveReservation(userID uint) {
	s.mu.RLock()
	alloc := s.alloc
	s.mu.RUnlock()
	if alloc != nil {
		alloc.RemoveReservation(userID)
	}
}

func (s *VpnService) ClientList() []ClientInfo {
	s.mu.RLock()
	out := make([]ClientInfo, 0, len(s.clients))
	for c := range s.clients {
		out = append(out, c.info())
	}
	s.mu.RUnlock()
	return out
}

type ClientInfo struct {
	Username    string `json:"username"`
	IP          string `json:"ip"`
	ConnectedAt string `json:"connected_at"`
}

var VPN *VpnService
