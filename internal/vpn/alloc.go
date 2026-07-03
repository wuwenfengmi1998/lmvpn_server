package vpn

import (
	"errors"
	"fmt"
	"net"
	"sync"

	"github.com/apparentlymart/go-cidr/cidr"
)

type AllocationManager struct {
	mu             sync.Mutex
	net            *net.IPNet
	serverIP       net.IP
	used           map[string]bool
	reservedByUser map[uint]string
	reservedSet    map[string]bool
}

func NewAllocationManager(ipNet *net.IPNet, serverIP net.IP, reservations map[uint]string) *AllocationManager {
	m := &AllocationManager{
		net:            ipNet,
		serverIP:       serverIP,
		used:           make(map[string]bool),
		reservedByUser: make(map[uint]string),
		reservedSet:    make(map[string]bool),
	}
	for uid, ip := range reservations {
		m.reservedByUser[uid] = ip
		m.reservedSet[ip] = true
	}
	return m
}

func (m *AllocationManager) ServerIP() net.IP { return m.serverIP }
func (m *AllocationManager) Subnet() *net.IPNet { return m.net }

func (m *AllocationManager) Allocate(userID uint) (net.IP, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if ipStr, ok := m.reservedByUser[userID]; ok {
		if m.used[ipStr] {
			return nil, fmt.Errorf("用户预留 IP %s 已被占用", ipStr)
		}
		m.used[ipStr] = true
		return net.ParseIP(ipStr), nil
	}

	count := cidr.AddressCount(m.net)
	maxIndex := int(count - 1)
	for i := 2; i < maxIndex; i++ {
		ip, err := cidr.Host(m.net, i)
		if err != nil {
			continue
		}
		ipStr := ip.String()
		if m.used[ipStr] || m.reservedSet[ipStr] {
			continue
		}
		m.used[ipStr] = true
		return ip, nil
	}
	return nil, errors.New("可用 IP 地址已耗尽")
}

func (m *AllocationManager) Release(ip net.IP) {
	if ip == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.used, ip.String())
}

func (m *AllocationManager) IsReserved(ipStr string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.reservedSet[ipStr]
}

func (m *AllocationManager) ReservedByUser(userID uint) (string, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	ip, ok := m.reservedByUser[userID]
	return ip, ok
}

func (m *AllocationManager) ReservedList() map[uint]string {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make(map[uint]string, len(m.reservedByUser))
	for k, v := range m.reservedByUser {
		out[k] = v
	}
	return out
}

func (m *AllocationManager) UsedCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.used)
}

func (m *AllocationManager) Capacity() uint64 {
	count := cidr.AddressCount(m.net)
	if count < 3 {
		return 0
	}
	return count - 3
}

func (m *AllocationManager) AddReservation(userID uint, ipStr string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if old, ok := m.reservedByUser[userID]; ok {
		delete(m.reservedSet, old)
	}
	m.reservedByUser[userID] = ipStr
	m.reservedSet[ipStr] = true
}

func (m *AllocationManager) RemoveReservation(userID uint) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if old, ok := m.reservedByUser[userID]; ok {
		delete(m.reservedByUser, userID)
		delete(m.reservedSet, old)
	}
}
