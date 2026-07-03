package handler

import (
	"net/http"
	"strconv"
	"time"

	"lmvpn/internal/db"
	"lmvpn/internal/model"

	"github.com/gin-gonic/gin"
)

type sessionResponse struct {
	SessionID string `json:"session_id"`
	IP        string `json:"ip"`
	UserAgent string `json:"user_agent"`
	Current   bool   `json:"current"`
	CreatedAt string `json:"created_at"`
	ExpiresAt string `json:"expires_at"`
}

func ListMySessions(c *gin.Context) {
	userID, _ := c.Get("user_id")
	currentSessionID, _ := c.Get("session_id")

	var sessions []model.Session
	db.DB.Where("user_id = ? AND invalid = ? AND expires_at > ?", userID, false, time.Now()).
		Order("created_at desc").
		Find(&sessions)

	result := make([]sessionResponse, len(sessions))
	for i, s := range sessions {
		result[i] = sessionResponse{
			SessionID: s.SessionID,
			IP:        s.IP,
			UserAgent: s.UserAgent,
			Current:   s.SessionID == currentSessionID,
			CreatedAt: s.CreatedAt.Format("2006-01-02 15:04:05"),
			ExpiresAt: s.ExpiresAt.Format("2006-01-02 15:04:05"),
		}
	}

	c.JSON(http.StatusOK, gin.H{"sessions": result})
}

func RevokeMySession(c *gin.Context) {
	sessionID := c.Param("sessionId")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	userID, _ := c.Get("user_id")
	currentSessionID, _ := c.Get("session_id")

	if sessionID == currentSessionID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "不能撤销当前会话"})
		return
	}

	var session model.Session
	if err := db.DB.Where("session_id = ? AND user_id = ?", sessionID, userID).First(&session).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "会话不存在"})
		return
	}

	db.DB.Model(&session).Update("invalid", true)
	c.JSON(http.StatusOK, gin.H{"message": "已撤销会话"})
}

func AdminRevokeUserSessions(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	var user model.User
	if err := db.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	db.DB.Model(&model.Session{}).Where("user_id = ?", id).Update("invalid", true)
	c.JSON(http.StatusOK, gin.H{"message": "已强制下线该用户"})
}
