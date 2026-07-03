package vpn

import (
	"log"
	"sync"
	"time"

	"lmvpn/internal/model"

	"github.com/gorilla/websocket"
)

const (
	readTimeout     = 60 * time.Second
	writeTimeout    = 10 * time.Second
	pingPeriod      = 30 * time.Second
	maxMessageSize  = 1 << 20
	maxConnsPerUser = 3
)

var (
	activeConns   = make(map[uint]int)
	activeConnsMu sync.Mutex
)

func runTunnel(conn *websocket.Conn, user *model.User) {
	defer conn.Close()

	activeConnsMu.Lock()
	if activeConns[user.ID] >= maxConnsPerUser {
		activeConnsMu.Unlock()
		sendJSON(conn, authResponse{Type: "auth_err", Message: "连接数已达上限"})
		return
	}
	activeConns[user.ID]++
	activeConnsMu.Unlock()

	defer func() {
		activeConnsMu.Lock()
		activeConns[user.ID]--
		if activeConns[user.ID] <= 0 {
			delete(activeConns, user.ID)
		}
		activeConnsMu.Unlock()
	}()

	log.Printf("用户 %s 已连接", user.Username)

	conn.SetReadLimit(maxMessageSize)

	go func() {
		ticker := time.NewTicker(pingPeriod)
		defer ticker.Stop()
		for range ticker.C {
			conn.SetWriteDeadline(time.Now().Add(writeTimeout))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}()

	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(readTimeout))
		return nil
	})

	for {
		conn.SetReadDeadline(time.Now().Add(readTimeout))
		messageType, data, err := conn.ReadMessage()
		if err != nil {
			log.Printf("用户 %s 断开连接: %v", user.Username, err)
			return
		}

		conn.SetWriteDeadline(time.Now().Add(writeTimeout))
		if err := conn.WriteMessage(messageType, data); err != nil {
			log.Printf("用户 %s 发送失败: %v", user.Username, err)
			return
		}
	}
}
