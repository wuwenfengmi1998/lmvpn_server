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

	r.POST("/api/login", middleware.LoginRateLimit(), handler.Login)

	auth := r.Group("/api")
	auth.Use(middleware.AuthMiddleware())
	{
		auth.GET("/me", handler.Me)
		auth.PUT("/me/password", handler.ChangePassword)
		auth.GET("/me/sessions", handler.ListMySessions)
		auth.DELETE("/me/sessions/:sessionId", handler.RevokeMySession)
	}

	admin := r.Group("/api/admin")
	admin.Use(middleware.AuthMiddleware(), middleware.AdminMiddleware())
	{
		admin.GET("/stats", handler.GetAdminStats)
		admin.GET("/users/count", handler.GetUserCount)
		admin.GET("/users", handler.ListUsers)
		admin.POST("/users", handler.CreateUser)
		admin.PUT("/users/:id", handler.UpdateUser)
		admin.DELETE("/users/:id", handler.DeleteUser)
		admin.DELETE("/users/:id/sessions", handler.AdminRevokeUserSessions)

		admin.GET("/vpn/settings", handler.GetVpnSettings)
		admin.PUT("/vpn/settings", handler.UpdateVpnSettings)
		admin.GET("/vpn/status", handler.GetVpnStatus)
		admin.GET("/vpn/diag", handler.GetVpnDiag)
		admin.GET("/vpn/reservations", handler.ListVpnReservations)
		admin.POST("/vpn/reservations", handler.CreateVpnReservation)
		admin.DELETE("/vpn/reservations/:id", handler.DeleteVpnReservation)
	}

	distDir := http.Dir("./dist")
	fs := http.FileServer(distDir)
	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		if strings.HasPrefix(path, "/api") || strings.HasPrefix(path, "/ws") {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		f, err := distDir.Open(path)
		if err != nil {
			c.Header("Content-Type", "text/html")
			c.File("./dist/index.html")
			return
		}
		f.Close()
		fs.ServeHTTP(c.Writer, c.Request)
	})
}
