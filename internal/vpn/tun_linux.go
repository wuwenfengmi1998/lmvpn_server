//go:build linux

package vpn

import (
	"fmt"
	"net"
)

func (t *TUNInterface) Configure(localIP net.IP, prefix int, peerIP net.IP) error {
	if err := execCmd("ip", "link", "set", "dev", t.Name(), "up"); err != nil {
		return err
	}
	if localIP == nil {
		return nil
	}
	localCidr := fmt.Sprintf("%s/%d", localIP.String(), prefix)
	if peerIP != nil {
		if err := execCmd("ip", "addr", "add", "dev", t.Name(), localCidr, "peer", peerIP.String()); err == nil {
			return nil
		}
	}
	return execCmd("ip", "addr", "add", "dev", t.Name(), localCidr)
}

func (t *TUNInterface) AddSubnetRoute(subnet *net.IPNet) error {
	return execCmd("ip", "route", "add", subnet.String(), "dev", t.Name())
}

func (t *TUNInterface) SetMTU(mtu int) error {
	return execCmd("ip", "link", "set", "dev", t.Name(), "mtu", fmt.Sprintf("%d", mtu))
}
