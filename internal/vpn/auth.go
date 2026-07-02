package vpn

import (
	"encoding/json"

	"lmvpn/internal/model"

	"github.com/gorilla/websocket"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type authMessage struct {
	Type     string `json:"type"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type authResponse struct {
	Type    string `json:"type"`
	Message string `json:"message,omitempty"`
}

func authenticate(conn *websocket.Conn, db *gorm.DB) (*model.User, error) {
	_, msgBytes, err := conn.ReadMessage()
	if err != nil {
		return nil, err
	}

	var msg authMessage
	if err := json.Unmarshal(msgBytes, &msg); err != nil || msg.Type != "auth" {
		resp := authResponse{Type: "auth_err", Message: "消息格式错误"}
		sendJSON(conn, resp)
		conn.Close()
		return nil, nil
	}

	var user model.User
	if err := db.Where("username = ? AND status = 1", msg.Username).First(&user).Error; err != nil {
		resp := authResponse{Type: "auth_err", Message: "用户名或密码错误"}
		sendJSON(conn, resp)
		conn.Close()
		return nil, nil
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(msg.Password)); err != nil {
		resp := authResponse{Type: "auth_err", Message: "用户名或密码错误"}
		sendJSON(conn, resp)
		conn.Close()
		return nil, nil
	}

	resp := authResponse{Type: "auth_ok"}
	if err := sendJSON(conn, resp); err != nil {
		conn.Close()
		return nil, nil
	}

	return &user, nil
}

func sendJSON(conn *websocket.Conn, v interface{}) error {
	data, _ := json.Marshal(v)
	return conn.WriteMessage(websocket.TextMessage, data)
}
