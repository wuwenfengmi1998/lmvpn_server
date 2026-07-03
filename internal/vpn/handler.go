package vpn

import (
	"log"
	"net/http"
	"net/url"

	"lmvpn/internal/db"
	"lmvpn/internal/middleware"
	"lmvpn/internal/model"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		if origin == "" {
			return true
		}
		u, err := url.Parse(origin)
		if err != nil {
			return false
		}
		return u.Host == r.Host
	},
}

func HandleWS(c *gin.Context) {
	tokenStr := c.Query("token")

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket 升级失败: %v", err)
		return
	}

	if tokenStr != "" {
		claims, err := middleware.ParseToken(tokenStr)
		if err != nil {
			sendJSON(conn, authResponse{Type: "auth_err", Message: "令牌无效或已过期"})
			conn.Close()
			return
		}
		var u model.User
		if err := db.DB.First(&u, claims.UserID).Error; err != nil || u.Status != 1 {
			sendJSON(conn, authResponse{Type: "auth_err", Message: "用户不存在或已禁用"})
			conn.Close()
			return
		}
		runTunnel(conn, &u)
		return
	}

	user, err := authenticate(conn, db.DB, c.ClientIP())
	if err != nil {
		log.Printf("认证读取失败: %v", err)
		conn.Close()
		return
	}
	if user == nil {
		return
	}

	runTunnel(conn, user)
}
