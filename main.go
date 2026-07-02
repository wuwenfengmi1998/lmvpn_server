package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"lmvpn/internal/config"
	"lmvpn/internal/db"
	"lmvpn/internal/handler"
	"lmvpn/internal/middleware"
	"lmvpn/internal/vpn"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := config.Load("data/config.yml")
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	if err := db.Init(&cfg.Database); err != nil {
		log.Fatalf("数据库初始化失败: %v", err)
	}

	r := gin.Default()

	r.GET("/ws", vpn.HandleWS)

	r.POST("/api/login", handler.Login)

	auth := r.Group("/api")
	auth.Use(middleware.AuthMiddleware())
	{
		auth.GET("/me", handler.Me)
	}

	fs := http.FileServer(http.Dir("./dist"))
	r.Use(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/ws") {
			return
		}
		if strings.HasPrefix(c.Request.URL.Path, "/api") {
			c.Next()
			return
		}

		fs.ServeHTTP(c.Writer, c.Request)
		c.Abort()
	})

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
		if err := os.MkdirAll(filepath.Dir(cfg.Web.Sock), 0755); err != nil {
			log.Fatalf("创建 sock 目录失败: %v", err)
		}
		listener, err := net.Listen("unix", cfg.Web.Sock)
		if err != nil {
			log.Fatalf("Unix socket 监听失败: %v", err)
		}
		if err := os.Chmod(cfg.Web.Sock, 0666); err != nil {
			log.Fatalf("设置 sock 权限失败: %v", err)
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
