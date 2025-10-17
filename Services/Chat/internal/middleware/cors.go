package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"helixtrack.ru/chat/internal/logger"
)

// CORSMiddleware handles CORS
func CORSMiddleware(allowedOrigins []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")

		// Check if origin is allowed
		if isAllowedOrigin(origin, allowedOrigins) {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Request-ID")
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Max-Age", "86400") // 24 hours
		} else if origin != "" {
			logger.Warn("CORS: Origin not allowed",
				zap.String("origin", origin),
				zap.String("ip", c.ClientIP()),
			)
		}

		// Handle preflight requests
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// isAllowedOrigin checks if an origin is in the allowed list
func isAllowedOrigin(origin string, allowedOrigins []string) bool {
	if origin == "" {
		return false
	}

	for _, allowed := range allowedOrigins {
		// Wildcard match
		if allowed == "*" {
			return true
		}

		// Exact match
		if allowed == origin {
			return true
		}

		// Pattern match (e.g., *.example.com)
		if strings.HasPrefix(allowed, "*.") {
			domain := strings.TrimPrefix(allowed, "*.")
			if strings.HasSuffix(origin, domain) {
				return true
			}
		}
	}

	return false
}
