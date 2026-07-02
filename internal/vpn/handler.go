package vpn

import (
	"log"
	"net/http"

	"lmvpn/internal/db"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func HandleWS(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket 升级失败: %v", err)
		return
	}

	user, err := authenticate(conn, db.DB)
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
