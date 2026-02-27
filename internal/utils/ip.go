// internal/utils/ip.go
package utils

import (
	"net"
	"strings"

	"github.com/gin-gonic/gin"
)

func GetClientIP(c *gin.Context) string {
	// Check X-Forwarded-For header
	forwarded := c.GetHeader("X-Forwarded-For")
	if forwarded != "" {
		ips := strings.Split(forwarded, ",")
		if len(ips) > 0 {
			ip := strings.TrimSpace(ips[0])
			if net.ParseIP(ip) != nil {
				return ip
			}
		}
	}

	// Check X-Real-IP header
	realIP := c.GetHeader("X-Real-IP")
	if realIP != "" {
		if net.ParseIP(realIP) != nil {
			return realIP
		}
	}

	// Fall back to remote address
	ip, _, err := net.SplitHostPort(c.Request.RemoteAddr)
	if err != nil {
		return c.Request.RemoteAddr
	}
	return ip
}
