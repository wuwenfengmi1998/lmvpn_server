package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/user"
	"path/filepath"
	"strconv"

	"lmvpn/internal/config"
	"lmvpn/internal/db"
	"lmvpn/internal/handler"
	"lmvpn/internal/middleware"
	"lmvpn/internal/router"
	"lmvpn/internal/vpn"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := config.Load("data/config.yml")
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	middleware.SetJWTSecret(cfg.Web.JWTSecret)
	middleware.SetRealIPHeaders(cfg.Web.RealIPHeaders)

	if err := db.Init(&cfg.Database); err != nil {
		log.Fatalf("数据库初始化失败: %v", err)
	}

	vpn.VPN = vpn.NewVpnService()
	if err := handler.ApplyVpnFromDB(vpn.VPN); err != nil {
		log.Printf("警告: 应用 VPN 设置失败: %v", err)
	}

	r := gin.Default()

	if len(cfg.Web.TrustedProxies) > 0 {
		_ = r.SetTrustedProxies(cfg.Web.TrustedProxies)
	} else {
		_ = r.SetTrustedProxies(nil)
	}

	router.Setup(r)

	if cfg.Web.Port == 0 && cfg.Web.Sock == "" {
		log.Fatal("配置错误: port 和 sock 不能同时为空")
	}

	if cfg.Web.Port > 0 {
		go func() {
			log.Printf("TCP 监听 :%d", cfg.Web.Port)
			if err := r.Run(fmt.Sprintf(":%d", cfg.Web.Port)); err != nil {
				log.Fatalf("TCP 启动失败: %v", err)
			}
		}()
	}

	if cfg.Web.Sock != "" {
		if err := os.Remove(cfg.Web.Sock); err != nil && !os.IsNotExist(err) {
			log.Fatalf("删除残留 sock 文件失败: %v", err)
		}
		dirMode := parseFileMode(cfg.Web.SockDirMode, 0755)
		if err := os.MkdirAll(filepath.Dir(cfg.Web.Sock), dirMode); err != nil {
			log.Fatalf("创建 sock 目录失败: %v", err)
		}
		listener, err := net.Listen("unix", cfg.Web.Sock)
		if err != nil {
			log.Fatalf("Unix socket 监听失败: %v", err)
		}
		sockMode := parseFileMode(cfg.Web.SockMode, 0666)
		if err := os.Chmod(cfg.Web.Sock, sockMode); err != nil {
			log.Printf("警告: 设置 sock 权限失败: %v", err)
		}
		if cfg.Web.SockGroup != "" {
			if err := chownGroup(cfg.Web.Sock, cfg.Web.SockGroup); err != nil {
				log.Printf("警告: 设置 sock group 失败: %v", err)
			}
		}
		go func() {
			log.Printf("Unix socket 监听 %s", cfg.Web.Sock)
			if err := r.RunListener(listener); err != nil {
				log.Fatalf("Unix socket 启动失败: %v", err)
			}
		}()
	}

	select {}
}

func parseFileMode(s string, defaultMode os.FileMode) os.FileMode {
	if s == "" {
		return defaultMode
	}
	m, err := strconv.ParseUint(s, 8, 32)
	if err != nil {
		log.Printf("警告: 解析文件权限 %q 失败，使用默认值 %o: %v", s, defaultMode, err)
		return defaultMode
	}
	return os.FileMode(m)
}

func chownGroup(path, group string) error {
	g, err := user.LookupGroup(group)
	if err != nil {
		return err
	}
	gid, err := strconv.Atoi(g.Gid)
	if err != nil {
		return err
	}
	return os.Chown(path, -1, gid)
}
