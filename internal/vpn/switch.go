package vpn

import (
	"net"
	"sync"

	"github.com/songgao/water/waterutil"
)

type SwitchConn interface {
	WritePacket(data []byte) error
	AssignedIP() net.IP
	AssignedIP6() net.IP
}

type ipKey [16]byte

func ipToKey(ip net.IP) ipKey {
	var k ipKey
	copy(k[:], ip.To16())
	return k
}

type PacketSwitch struct {
	allowClientToClient bool
	mu    sync.RWMutex
	table map[ipKey]SwitchConn
}

func NewPacketSwitch(allowClientToClient bool) *PacketSwitch {
	return &PacketSwitch{
		allowClientToClient: allowClientToClient,
		table:               make(map[ipKey]SwitchConn),
	}
}

func (s *PacketSwitch) SetAllowClientToClient(v bool) {
	s.mu.Lock()
	s.allowClientToClient = v
	s.mu.Unlock()
}

func (s *PacketSwitch) Register(c SwitchConn) {
	s.mu.Lock()
	s.table[ipToKey(c.AssignedIP())] = c
	if ip6 := c.AssignedIP6(); ip6 != nil {
		s.table[ipToKey(ip6)] = c
	}
	s.mu.Unlock()
}

func (s *PacketSwitch) Unregister(c SwitchConn) {
	s.mu.Lock()
	if cur, ok := s.table[ipToKey(c.AssignedIP())]; ok && cur == c {
		delete(s.table, ipToKey(c.AssignedIP()))
	}
	if ip6 := c.AssignedIP6(); ip6 != nil {
		k := ipToKey(ip6)
		if cur, ok := s.table[k]; ok && cur == c {
			delete(s.table, k)
		}
	}
	s.mu.Unlock()
}

func (s *PacketSwitch) findByIP(ip net.IP) SwitchConn {
	s.mu.RLock()
	c := s.table[ipToKey(ip)]
	s.mu.RUnlock()
	return c
}

func (s *PacketSwitch) allExcept(skip SwitchConn) []SwitchConn {
	s.mu.RLock()
	out := make([]SwitchConn, 0, len(s.table))
	for _, c := range s.table {
		if c == skip {
			continue
		}
		out = append(out, c)
	}
	s.mu.RUnlock()
	return out
}

func parseIPAddrs(packet []byte) (src, dest net.IP, ok bool) {
	if len(packet) < 1 {
		return nil, nil, false
	}
	switch {
	case waterutil.IsIPv4(packet):
		if len(packet) < 20 {
			return nil, nil, false
		}
		return waterutil.IPv4Source(packet), waterutil.IPv4Destination(packet), true
	case waterutil.IsIPv6(packet):
		if len(packet) < 40 {
			return nil, nil, false
		}
		src = make(net.IP, 16)
		copy(src, packet[8:24])
		dest = make(net.IP, 16)
		copy(dest, packet[24:40])
		return src, dest, true
	}
	return nil, nil, false
}

func (s *PacketSwitch) allowC2C() bool {
	s.mu.RLock()
	v := s.allowClientToClient
	s.mu.RUnlock()
	return v
}

func (s *PacketSwitch) RouteFromClient(src SwitchConn, packet []byte) []SwitchConn {
	srcIP, dest, ok := parseIPAddrs(packet)
	if !ok {
		return nil
	}
	// anti-spoof: enforce assigned source IP by version
	if srcIP != nil {
		if srcIP.To4() != nil {
			if !srcIP.Equal(src.AssignedIP()) {
				return nil
			}
		} else {
			assigned6 := src.AssignedIP6()
			if assigned6 == nil || !srcIP.Equal(assigned6) {
				return nil
			}
		}
	}
	if dest.IsGlobalUnicast() {
		if c := s.findByIP(dest); c != nil && s.allowC2C() {
			return []SwitchConn{c}
		}
		return nil
	}
	if s.allowC2C() {
		return s.allExcept(src)
	}
	return nil
}

func (s *PacketSwitch) RouteFromTUN(packet []byte) []SwitchConn {
	_, dest, ok := parseIPAddrs(packet)
	if !ok {
		return nil
	}
	if dest.IsGlobalUnicast() {
		if c := s.findByIP(dest); c != nil {
			return []SwitchConn{c}
		}
		return nil
	}
	return s.allExcept(nil)
}
