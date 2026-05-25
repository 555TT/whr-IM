package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
)

func isAllowedOrigin(origin string) bool {
	if origin == "" {
		return false
	}
	//
	//parsed, err := url.Parse(origin)
	//if err != nil {
	//	return false
	//}
	//
	//if parsed.Scheme != "http" {
	//	return false
	//}
	//
	//hostname := parsed.Hostname()
	//return hostname == "127.0.0.1" || hostname == "localhost"
	return true
}

func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := strings.TrimSpace(c.GetHeader("Origin"))
		if isAllowedOrigin(origin) {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, OPTIONS")
		}

		if c.Request.Method == "OPTIONS" {
			c.Status(204)
			c.Abort()
			return
		}

		c.Next()
	}
}
