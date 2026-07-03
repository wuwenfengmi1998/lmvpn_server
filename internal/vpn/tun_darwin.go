//go:build darwin

package vpn

import (
	"fmt"
	"net"
)

func inetFamily(ip net.IP) string {
	if ip.To4() == nil {
		return "inet6"
	}
	return "inet"
}

func (t *TUNInterface) Configure(localIP net.IP, prefix int, peerIP net.IP) error {
	if localIP == nil {
		return execCmd("ifconfig", t.Name(), "up")
	}
	localCidr := fmt.Sprintf("%s/%d", localIP.String(), prefix)
	inetType := inetFamily(localIP)
	var err error
	if t.Iface.IsTUN() && inetType == "inet" {
		err = execCmd("ifconfig", t.Name(), inetType, localCidr, peerIP.String(), "up")
	} else {
		err = execCmd("ifconfig", t.Name(), inetType, localCidr, "up")
	}
	return err
}

func (t *TUNInterface) AddSubnetRoute(subnet *net.IPNet) error {
	inetType := inetFamily(subnet.IP)
	return execCmd("route", "add", fmt.Sprintf("-%s", inetType), "-net", subnet.String(), "-interface", t.Name())
}

func (t *TUNInterface) SetMTU(mtu int) error {
	return execCmd("ifconfig", t.Name(), "mtu", fmt.Sprintf("%d", mtu))
}
