package http

import (
	"strings"

	"github.com/gin-gonic/gin"
)

func GetUserIPAddress(c *gin.Context) string {
	if ip := c.Request.Header.Get("CF-Connecting-IP"); ip != "" {
		return ip
	}

	if ip := c.Request.Header.Get("X-Forwarded-For"); ip != "" {
		return strings.Split(ip, ",")[0]
	}

	return c.ClientIP()
}
