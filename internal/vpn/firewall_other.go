//go:build !linux && !darwin

package vpn

import "log"

func configureFirewall(ipNet, ipNet6 string, tunName string) {
	log.Printf("防火墙配置在当前平台不支持，请手动配置 NAT 和转发规则")
}

func checkUFWActive() bool {
	return false
}
