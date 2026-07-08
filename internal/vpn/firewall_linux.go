//go:build linux

package vpn

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
)

// configureFirewall dynamically configures NAT masquerade, forward accept rules,
// and UFW forward rules based on the current VPN subnets.
// This is called from ApplySettings() so subnet changes in the backend
// are automatically reflected in firewall rules.
func configureFirewall(ipNet, ipNet6 string, tunName string) {
	wanIface := detectWANInterface()
	if wanIface == "" {
		log.Printf("警告: 未能检测出口网卡，跳过防火墙配置")
		return
	}
	log.Printf("出口网卡: %s", wanIface)

	configureNAT(wanIface, ipNet, ipNet6)
	configureForward(ipNet, ipNet6)
	configureUFWForward(ipNet, ipNet6)
}

func detectWANInterface() string {
	out, err := exec.Command("ip", "route", "show", "default").Output()
	if err != nil {
		return ""
	}
	fields := strings.Fields(string(out))
	for i, f := range fields {
		if f == "dev" && i+1 < len(fields) {
			return fields[i+1]
		}
	}
	return ""
}

func nftExists(args ...string) bool {
	return exec.Command("nft", args...).Run() == nil
}

func nftExec(args ...string) {
	cmd := exec.Command("nft", args...)
	cmd.Stderr = nil
	_ = cmd.Run()
}

func configureNAT(wanIface, ipNet, ipNet6 string) {
	if !nftExists("list", "table", "inet", "lmvpn_nat") {
		nftExec("add", "table", "inet", "lmvpn_nat")
	}

	// postrouting chain
	if !nftExists("list", "chain", "inet", "lmvpn_nat", "postrouting") {
		nftExec("add", "chain", "inet", "lmvpn_nat", "postrouting",
			"{ type nat hook postrouting priority 100 ; }")
	}
	nftExec("flush", "chain", "inet", "lmvpn_nat", "postrouting")
	if ipNet != "" {
		nftExec("add", "rule", "inet", "lmvpn_nat", "postrouting",
			"oifname", wanIface, "ip", "saddr", ipNet, "masquerade")
	}
	if ipNet6 != "" {
		nftExec("add", "rule", "inet", "lmvpn_nat", "postrouting",
			"oifname", wanIface, "ip6", "saddr", ipNet6, "masquerade")
	}
	log.Printf("NAT masquerade 已配置: wan=%s v4=%s", wanIface, ipNet)
	if ipNet6 != "" {
		log.Printf("NAT masquerade IPv6: %s", ipNet6)
	}
}

func configureForward(ipNet, ipNet6 string) {
	if !nftExists("list", "chain", "inet", "lmvpn_nat", "forward") {
		nftExec("add", "chain", "inet", "lmvpn_nat", "forward",
			"{ type filter hook forward priority 0 ; policy accept ; }")
	}
	nftExec("flush", "chain", "inet", "lmvpn_nat", "forward")
	if ipNet != "" {
		nftExec("add", "rule", "inet", "lmvpn_nat", "forward",
			"ip", "saddr", ipNet, "accept")
		nftExec("add", "rule", "inet", "lmvpn_nat", "forward",
			"ip", "daddr", ipNet, "accept")
	}
	if ipNet6 != "" {
		nftExec("add", "rule", "inet", "lmvpn_nat", "forward",
			"ip6", "saddr", ipNet6, "accept")
		nftExec("add", "rule", "inet", "lmvpn_nat", "forward",
			"ip6", "daddr", ipNet6, "accept")
	}
}

// configureUFWForward adds VPN subnet accept rules to UFW's user-forward chains.
// UFW's FORWARD chain has policy DROP by default, which overrides the lmvpn_nat
// forward chain's accept rules. We use dedicated chains (lmvpn-fwd/lmvpn6-fwd)
// that are jumped to from ufw-user-forward/ufw6-user-forward.
func configureUFWForward(ipNet, ipNet6 string) {
	// IPv4
	if nftExists("list", "chain", "ip", "filter", "ufw-user-forward") {
		if !nftExists("list", "chain", "ip", "filter", "lmvpn-fwd") {
			nftExec("add", "chain", "ip", "filter", "lmvpn-fwd")
		}
		nftExec("flush", "chain", "ip", "filter", "lmvpn-fwd")
		if ipNet != "" {
			nftExec("add", "rule", "ip", "filter", "lmvpn-fwd",
				"ip", "saddr", ipNet, "accept")
			nftExec("add", "rule", "ip", "filter", "lmvpn-fwd",
				"ip", "daddr", ipNet, "accept")
		}
		// Add jump if not already present
		if !nftJumpExists("ip", "filter", "ufw-user-forward", "lmvpn-fwd") {
			nftExec("add", "rule", "ip", "filter", "ufw-user-forward", "jump", "lmvpn-fwd")
		}
		log.Printf("UFW IPv4 转发规则已配置")
	}

	// IPv6
	if ipNet6 != "" && nftExists("list", "chain", "ip6", "filter", "ufw6-user-forward") {
		if !nftExists("list", "chain", "ip6", "filter", "lmvpn6-fwd") {
			nftExec("add", "chain", "ip6", "filter", "lmvpn6-fwd")
		}
		nftExec("flush", "chain", "ip6", "filter", "lmvpn6-fwd")
		nftExec("add", "rule", "ip6", "filter", "lmvpn6-fwd",
			"ip6", "saddr", ipNet6, "accept")
		nftExec("add", "rule", "ip6", "filter", "lmvpn6-fwd",
			"ip6", "daddr", ipNet6, "accept")
		if !nftJumpExists("ip6", "filter", "ufw6-user-forward", "lmvpn6-fwd") {
			nftExec("add", "rule", "ip6", "filter", "ufw6-user-forward", "jump", "lmvpn6-fwd")
		}
		log.Printf("UFW IPv6 转发规则已配置")
	}
}

// nftJumpExists checks if a jump rule to the target chain already exists
// in the source chain.
func nftJumpExists(family, table, chain, target string) bool {
	out, err := exec.Command("nft", "list", "chain", family, table, chain).Output()
	if err != nil {
		return false
	}
	return strings.Contains(string(out), fmt.Sprintf("jump %s", target))
}

// checkUFWActive checks if UFW is active by looking for its forward chain.
func checkUFWActive() bool {
	return nftExists("list", "chain", "ip", "filter", "ufw-user-forward")
}
