package middleware

import (
	"net"
	"strings"

	"github.com/gin-gonic/gin"
)

var realIPHeaders []string

func SetRealIPHeaders(headers []string) {
	realIPHeaders = headers
}

func GetRealIP(c *gin.Context) string {
	for _, header := range realIPHeaders {
		val := c.GetHeader(header)
		if val == "" {
			continue
		}
		ip := strings.TrimSpace(strings.Split(val, ",")[0])
		if ip != "" && net.ParseIP(ip) != nil {
			return ip
		}
	}
	return c.ClientIP()
}
