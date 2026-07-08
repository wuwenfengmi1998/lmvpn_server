package vpn

import (
	"errors"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"lmvpn/internal/model"

	"github.com/apparentlymart/go-cidr/cidr"
)

type VpnService struct {
	mu        sync.RWMutex
	settings  model.VpnSetting
	net       *net.IPNet
	serverIP  net.IP
	prefix    int
	alloc     *AllocationManager
	net6      *net.IPNet
	serverIP6 net.IP
	prefix6   int
	alloc6    *AllocationManager
	switchx   *PacketSwitch
	tun       *TUNInterface
	tunDone   chan struct{}
	running   bool
	startedAt time.Time
	clients   map[*tunnelConn]struct{}
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

func (s *VpnService) StartedAt() time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.startedAt
}

func (s *VpnService) TotalLiveTraffic() (rx, tx int64) {
	s.mu.RLock()
	for c := range s.clients {
		rx += c.rxBytes.Load()
		tx += c.txBytes.Load()
	}
	s.mu.RUnlock()
	return
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

func (s *VpnService) ApplySettings(settings model.VpnSetting, reservations4, reservations6 map[uint]string) error {
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

	var ipNet6 *net.IPNet
	var serverIP6 net.IP
	var prefix6 int
	var alloc6 *AllocationManager
	if settings.Subnet6 != "" {
		ipNet6, serverIP6, prefix6, err = s.parseNet(settings.Subnet6)
		if err != nil {
			return fmt.Errorf("IPv6 子网错误: %w", err)
		}
	}

	tun, err := CreateTUN(settings.InterfaceName)
	if err != nil {
		return err
	}

	if settings.DoLocalIPConfig {
		if err := tun.Configure(serverIP, prefix, nil); err != nil {
			_ = tun.Close()
			return fmt.Errorf("配置 TUN IPv4 失败: %w", err)
		}
	}
	if err := tun.AddSubnetRoute(ipNet); err != nil {
		log.Printf("警告: 添加 IPv4 子网路由失败: %v", err)
	}
	if ipNet6 != nil {
		if settings.DoLocalIPConfig {
			if err := tun.Configure(serverIP6, prefix6, nil); err != nil {
				log.Printf("警告: 配置 TUN IPv6 失败: %v", err)
			}
		}
		if err := tun.AddSubnetRoute(ipNet6); err != nil {
			log.Printf("警告: 添加 IPv6 子网路由失败: %v", err)
		}
		alloc6 = NewAllocationManager(ipNet6, serverIP6, reservations6)
	}
	if err := tun.SetMTU(settings.MTU); err != nil {
		log.Printf("警告: 设置 MTU 失败: %v", err)
	}

	s.mu.Lock()
	s.net = ipNet
	s.serverIP = serverIP
	s.prefix = prefix
	s.alloc = NewAllocationManager(ipNet, serverIP, reservations4)
	s.net6 = ipNet6
	s.serverIP6 = serverIP6
	s.prefix6 = prefix6
	s.alloc6 = alloc6
	s.switchx = NewPacketSwitch(settings.AllowClientToClient)
	s.tun = tun
	s.tunDone = make(chan struct{})
	s.running = true
	s.startedAt = time.Now()
	s.mu.Unlock()

	go s.serveTUN()

	subnet4 := ipNet.String()
	var subnet6Str string
	if ipNet6 != nil {
		subnet6Str = ipNet6.String()
	}
	configureFirewall(subnet4, subnet6Str, tun.Name())

	log.Printf("VPN 服务已启动: tun=%s subnet=%s server=%s", tun.Name(), ipNet.String(), serverIP.String())
	if ipNet6 != nil {
		log.Printf("  IPv6: subnet=%s server=%s", ipNet6.String(), serverIP6.String())
	}
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

func (s *VpnService) Allocate(user *model.User) (net.IP, net.IP, error) {
	s.mu.RLock()
	alloc := s.alloc
	alloc6 := s.alloc6
	s.mu.RUnlock()
	if alloc == nil {
		return nil, nil, errors.New("VPN 服务未运行")
	}
	ip4, err := alloc.Allocate(user.ID)
	if err != nil {
		return nil, nil, err
	}
	var ip6 net.IP
	if alloc6 != nil {
		ip6, err = alloc6.Allocate(user.ID)
		if err != nil {
			alloc.Release(ip4)
			return ip4, nil, fmt.Errorf("IPv6 分配失败: %w", err)
		}
	}
	return ip4, ip6, nil
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
	if s.alloc6 != nil && c.assignedIP6 != nil {
		s.alloc6.Release(c.assignedIP6)
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

func (s *VpnService) ServerIP6() net.IP {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.serverIP6
}

func (s *VpnService) Prefix6() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.prefix6
}

func (s *VpnService) HasIPv6() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.alloc6 != nil
}

func (s *VpnService) AllocStats() (used4 int, cap4 uint64, used6 int, cap6 uint64) {
	s.mu.RLock()
	alloc := s.alloc
	alloc6 := s.alloc6
	s.mu.RUnlock()
	if alloc != nil {
		used4, cap4 = alloc.UsedCount(), alloc.Capacity()
	}
	if alloc6 != nil {
		used6, cap6 = alloc6.UsedCount(), alloc6.Capacity()
	}
	return
}

func (s *VpnService) AddReservation(userID uint, ipStr string) {
	s.mu.RLock()
	alloc := s.alloc
	s.mu.RUnlock()
	if alloc != nil {
		alloc.AddReservation(userID, ipStr)
	}
}

func (s *VpnService) AddReservation6(userID uint, ipStr string) {
	s.mu.RLock()
	alloc6 := s.alloc6
	s.mu.RUnlock()
	if alloc6 != nil {
		alloc6.AddReservation(userID, ipStr)
	}
}

func (s *VpnService) RemoveReservation(userID uint) {
	s.mu.RLock()
	alloc := s.alloc
	alloc6 := s.alloc6
	s.mu.RUnlock()
	if alloc != nil {
		alloc.RemoveReservation(userID)
	}
	if alloc6 != nil {
		alloc6.RemoveReservation(userID)
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
	IP6         string `json:"ip6,omitempty"`
	ConnectedAt string `json:"connected_at"`
}

var VPN *VpnService
