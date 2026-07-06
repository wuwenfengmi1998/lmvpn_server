//go:build linux

package vpn

import (
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func fillPlatformDiag(r *DiagResult) {
	r.HasCapNetAdmin = ptrBool(checkCapNetAdmin())
	if r.HasCapNetAdmin == nil || !*r.HasCapNetAdmin {
		r.CapNetAdminNote = "CAP_NET_ADMIN 未授权，TUN 操作需 root 或显式授予该能力"
	}

	v := readIPForward()
	r.IPForward = v
	if v == nil || !*v {
		r.IPForwardNote = "未开启，执行: sysctl -w net.ipv4.ip_forward=1"
	}

	m := checkMasquerade()
	r.Masquerade = m
	if m == nil {
		r.MasqueradeNote = "iptables 未安装或不可执行"
	} else if !*m {
		r.MasqueradeNote = "未检测到 MASQUERADE 规则，客户端无法出网"
	}
}

func ptrBool(b bool) *bool { return &b }

func checkCapNetAdmin() bool {
	data, err := os.ReadFile("/proc/self/status")
	if err != nil {
		return false
	}
	for _, line := range strings.Split(string(data), "\n") {
		if !strings.HasPrefix(line, "CapEff:") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			return false
		}
		val, err := strconv.ParseUint(fields[1], 16, 64)
		if err != nil {
			return false
		}
		return val&(1<<12) != 0
	}
	return false
}

func readIPForward() *bool {
	data, err := os.ReadFile("/proc/sys/net/ipv4/ip_forward")
	if err != nil {
		return nil
	}
	v := strings.TrimSpace(string(data)) == "1"
	return &v
}

func checkMasquerade() *bool {
	cmd := exec.Command("iptables", "-t", "nat", "-L", "POSTROUTING", "-n")
	out, err := cmd.Output()
	if err != nil {
		return nil
	}
	has := strings.Contains(string(out), "MASQUERADE")
	return &has
}
