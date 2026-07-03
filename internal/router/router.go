package router

import (
	"net/http"
	"strings"

	"lmvpn/internal/handler"
	"lmvpn/internal/middleware"
	"lmvpn/internal/vpn"

	"github.com/gin-gonic/gin"
)

func Setup(r *gin.Engine) {
	r.GET("/ws", vpn.HandleWS)

	r.POST("/api/login", handler.Login)

	auth := r.Group("/api")
	auth.Use(middleware.AuthMiddleware())
	{
		auth.GET("/me", handler.Me)
		auth.PUT("/me/password", handler.ChangePassword)
	}

	admin := r.Group("/api/admin")
	admin.Use(middleware.AuthMiddleware(), middleware.AdminMiddleware())
	{
		admin.GET("/users/count", handler.GetUserCount)
		admin.GET("/users", handler.ListUsers)
		admin.POST("/users", handler.CreateUser)
		admin.PUT("/users/:id", handler.UpdateUser)
		admin.DELETE("/users/:id", handler.DeleteUser)
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
}
