package vpn

import (
	"encoding/json"
	"log"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"lmvpn/internal/model"

	"github.com/gorilla/websocket"
)

const (
	readTimeout     = 60 * time.Second
	writeTimeout    = 10 * time.Second
	readyTimeout    = 30 * time.Second
	pingPeriod      = 30 * time.Second
	maxMessageSize  = 1 << 20
	maxConnsPerUser = 3
)

var (
	activeConns   = make(map[uint]int)
	activeConnsMu sync.Mutex
)

type tunnelConn struct {
	conn        *websocket.Conn
	user        *model.User
	svc         *VpnService
	assignedIP  net.IP
	connectedAt time.Time
	writeMu     sync.Mutex
	ready       atomic.Bool
	rxBytes     atomic.Int64
	txBytes     atomic.Int64
}

func (c *tunnelConn) AssignedIP() net.IP { return c.assignedIP }

func (c *tunnelConn) WritePacket(data []byte) error {
	if !c.ready.Load() || len(data) == 0 {
		return nil
	}
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	c.conn.SetWriteDeadline(time.Now().Add(writeTimeout))
	if err := c.conn.WriteMessage(websocket.BinaryMessage, data); err != nil {
		return err
	}
	c.txBytes.Add(int64(len(data)))
	return nil
}

func (c *tunnelConn) writeControl(v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	c.conn.SetWriteDeadline(time.Now().Add(writeTimeout))
	return c.conn.WriteMessage(websocket.TextMessage, data)
}

func (c *tunnelConn) close() {
	c.writeMu.Lock()
	_ = c.conn.Close()
	c.writeMu.Unlock()
}

func (c *tunnelConn) info() ClientInfo {
	return ClientInfo{
		Username:    c.user.Username,
		IP:          c.assignedIP.String(),
		ConnectedAt: c.connectedAt.Format("2006-01-02 15:04:05"),
	}
}

func runTunnel(conn *websocket.Conn, user *model.User) {
	defer conn.Close()

	if VPN == nil || !VPN.Running() {
		_ = sendJSON(conn, controlMessage{Type: "error", Message: "VPN 服务未启用"})
		return
	}

	activeConnsMu.Lock()
	if activeConns[user.ID] >= maxConnsPerUser {
		activeConnsMu.Unlock()
		_ = sendJSON(conn, controlMessage{Type: "error", Message: "连接数已达上限"})
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

	ip, err := VPN.Allocate(user)
	if err != nil {
		_ = sendJSON(conn, controlMessage{Type: "error", Message: "分配 IP 失败: " + err.Error()})
		return
	}

	tc := &tunnelConn{
		conn:        conn,
		user:        user,
		svc:         VPN,
		assignedIP:  ip,
		connectedAt: time.Now(),
	}

	VPN.registerClient(tc)
	defer VPN.unregisterClient(tc)

	settings := VPN.Settings()
	initMsg := initMessage{
		Type:     "init",
		IP:       ip.String(),
		Prefix:   VPN.Prefix(),
		MTU:      settings.MTU,
		ServerIP: VPN.ServerIP().String(),
	}
	if err := tc.writeControl(initMsg); err != nil {
		log.Printf("用户 %s 发送 init 失败: %v", user.Username, err)
		return
	}

	log.Printf("用户 %s 已连接，分配 IP %s", user.Username, ip.String())

	conn.SetReadLimit(maxMessageSize)
	conn.SetReadDeadline(time.Now().Add(readyTimeout))
	readyDeadline := time.Now().Add(readyTimeout)

	go func() {
		ticker := time.NewTicker(pingPeriod)
		defer ticker.Stop()
		for range ticker.C {
			tc.writeMu.Lock()
			if err := conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(writeTimeout)); err != nil {
				tc.writeMu.Unlock()
				return
			}
			tc.writeMu.Unlock()
		}
	}()

	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(readTimeout))
		return nil
	})

	for {
		messageType, data, err := conn.ReadMessage()
		if err != nil {
			log.Printf("用户 %s 断开连接: %v", user.Username, err)
			return
		}

		if messageType == websocket.TextMessage {
			var msg controlMessage
			if err := json.Unmarshal(data, &msg); err != nil {
				continue
			}
			if msg.Type == "ready" && !tc.ready.Load() {
				tc.ready.Store(true)
				conn.SetReadDeadline(time.Now().Add(readTimeout))
				log.Printf("用户 %s 就绪 (IP %s)", user.Username, ip.String())
			}
			continue
		}

		if messageType != websocket.BinaryMessage {
			continue
		}

		if !tc.ready.Load() {
			if time.Now().After(readyDeadline) {
				log.Printf("用户 %s 等待 ready 超时", user.Username)
				return
			}
			conn.SetReadDeadline(readyDeadline)
			continue
		}

		tc.rxBytes.Add(int64(len(data)))

		targets := VPN.RouteFromClient(tc, data)
		if len(targets) == 0 {
			if err := VPN.WriteToTUN(data); err != nil {
				log.Printf("用户 %s 写入 TUN 失败: %v", user.Username, err)
			}
			continue
		}
		for _, t := range targets {
			_ = t.WritePacket(data)
		}
	}
}
