package middleware

import (
	"net/http"
	"strings"
	"time"

	"lmvpn/internal/db"
	"lmvpn/internal/model"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const tokenExpire = 24 * time.Hour

var jwtSecret []byte

func SetJWTSecret(secret string) {
	jwtSecret = []byte(secret)
}

type Claims struct {
	SessionID string `json:"session_id,omitempty"`
	UserID    uint   `json:"user_id"`
	Username  string `json:"username"`
	Role      string `json:"role"`
	jwt.RegisteredClaims
}

func GenerateToken(sessionID string, userID uint, username, role string) (string, error) {
	claims := Claims{
		SessionID: sessionID,
		UserID:    userID,
		Username:  username,
		Role:      role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenExpire)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func ParseToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrSignatureInvalid
}

func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, _ := c.Get("role")
		if role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "无权限访问"})
			c.Abort()
			return
		}
		c.Next()
	}
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未提供认证令牌"})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "认证格式错误"})
			c.Abort()
			return
		}

		claims, err := ParseToken(parts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "令牌无效或已过期"})
			c.Abort()
			return
		}

		var user model.User
		if err := db.DB.First(&user, claims.UserID).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "用户不存在"})
			c.Abort()
			return
		}

		if user.Status != 1 {
			c.JSON(http.StatusForbidden, gin.H{"error": "账号已被禁用"})
			c.Abort()
			return
		}

		if claims.SessionID != "" {
			var session model.Session
			if err := db.DB.Where("session_id = ?", claims.SessionID).First(&session).Error; err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "会话不存在"})
				c.Abort()
				return
			}
			if session.Invalid {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "会话已失效，请重新登录"})
				c.Abort()
				return
			}
			if time.Now().After(session.ExpiresAt) {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "会话已过期，请重新登录"})
				c.Abort()
				return
			}
		} else if user.TokenInvalidBefore != nil && claims.IssuedAt != nil {
			if claims.IssuedAt.Time.Before(*user.TokenInvalidBefore) {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "令牌已失效，请重新登录"})
				c.Abort()
				return
			}
		}

		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", user.Role)
		c.Set("session_id", claims.SessionID)
		c.Next()
	}
}
