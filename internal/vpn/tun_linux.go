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
	args := []string{"addr", "add", "dev", t.Name(), localCidr, "peer", peerIP.String()}
	if err := execCmd("ip", args...); err != nil {
		if err2 := execCmd("ip", "addr", "add", "dev", t.Name(), localCidr); err2 != nil {
			return err
		}
	}
	return nil
}

func (t *TUNInterface) AddSubnetRoute(subnet *net.IPNet) error {
	return execCmd("ip", "route", "add", subnet.String(), "dev", t.Name())
}

func (t *TUNInterface) SetMTU(mtu int) error {
	return execCmd("ip", "link", "set", "dev", t.Name(), "mtu", fmt.Sprintf("%d", mtu))
}
