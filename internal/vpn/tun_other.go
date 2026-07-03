//go:build !linux && !darwin

package vpn

import (
	"errors"
	"net"
)

func (t *TUNInterface) Configure(localIP net.IP, prefix int, peerIP net.IP) error {
	return errors.New("TUN 配置当前平台不支持")
}

func (t *TUNInterface) AddSubnetRoute(subnet *net.IPNet) error {
	return errors.New("TUN 路由当前平台不支持")
}

func (t *TUNInterface) SetMTU(mtu int) error {
	return errors.New("TUN MTU 当前平台不支持")
}
