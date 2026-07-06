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

	m, note := checkMasquerade()
	r.Masquerade = m
	r.MasqueradeNote = note
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

// findExecutable 在常见路径中查找可执行文件，弥补 systemd 服务 PATH 不含 /usr/sbin、/sbin 的问题
func findExecutable(names ...string) string {
	for _, name := range names {
		if p, err := exec.LookPath(name); err == nil {
			return p
		}
		for _, dir := range []string{"/usr/sbin", "/sbin", "/usr/bin", "/bin"} {
			full := dir + "/" + name
			if fi, err := os.Stat(full); err == nil && !fi.IsDir() {
				return full
			}
		}
	}
	return ""
}

// checkMasquerade 检测 NAT masquerade 规则，优先 iptables，回退 nft
// 返回 (结果, 说明)；结果为 nil 表示无法判定
func checkMasquerade() (*bool, string) {
	// 优先 iptables
	iptPath := findExecutable("iptables")
	if iptPath != "" {
		out, err := exec.Command(iptPath, "-t", "nat", "-L", "POSTROUTING", "-n").Output()
		if err != nil {
			return nil, "iptables 不可执行（权限不足？），无法检测 MASQUERADE"
		}
		has := strings.Contains(string(out), "MASQUERADE")
		if has {
			return &has, ""
		}
		return &has, "未检测到 MASQUERADE 规则，客户端无法出网"
	}

	// 回退 nft
	nftPath := findExecutable("nft")
	if nftPath != "" {
		out, err := exec.Command(nftPath, "list", "ruleset").Output()
		if err != nil {
			return nil, "nft 不可执行（权限不足？），无法检测 MASQUERADE"
		}
		has := strings.Contains(string(out), "masquerade")
		if has {
			return &has, ""
		}
		return &has, "未检测到 masquerade 规则，客户端无法出网"
	}

	return nil, "iptables 与 nft 均未安装，无法检测 NAT 规则。Debian/Ubuntu 安装: apt install iptables iptables-persistent"
}
