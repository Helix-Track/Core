package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"helixtrack.ru/chat/internal/logger"
)

// RequestLoggerMiddleware logs HTTP requests
func RequestLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)

		// Log request
		fields := []zap.Field{
			zap.Int("status", c.Writer.Status()),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.String("ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
			zap.Duration("latency", latency),
			zap.Int("body_size", c.Writer.Size()),
		}

		// Add user info if available
		if username, exists := c.Get("username"); exists {
			fields = append(fields, zap.String("username", username.(string)))
		}

		// Add error if any
		if len(c.Errors) > 0 {
			fields = append(fields, zap.String("errors", c.Errors.String()))
		}

		// Log based on status code
		status := c.Writer.Status()
		switch {
		case status >= 500:
			logger.Error("Server error", fields...)
		case status >= 400:
			logger.Warn("Client error", fields...)
		case status >= 300:
			logger.Info("Redirect", fields...)
		default:
			logger.Debug("Request completed", fields...)
		}
	}
}

// RecoveryMiddleware recovers from panics
func RecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error("Panic recovered",
					zap.Any("error", err),
					zap.String("path", c.Request.URL.Path),
					zap.String("method", c.Request.Method),
					zap.String("ip", c.ClientIP()),
				)

				c.JSON(500, gin.H{
					"error_code":    2002,
					"error_message": "Internal server error",
				})
				c.Abort()
			}
		}()

		c.Next()
	}
}
