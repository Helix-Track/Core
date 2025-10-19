package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/helixtrack/attachments-service/internal/security/ratelimit"
	"github.com/helixtrack/attachments-service/internal/utils"
	"go.uber.org/zap"
)

// JWTClaims represents the JWT claims structure
type JWTClaims struct {
	Sub         string   `json:"sub"`
	Name        string   `json:"name"`
	Username    string   `json:"username"`
	Role        string   `json:"role"`
	Permissions string   `json:"permissions"`
	HtCoreAddr  string   `json:"htCoreAddress"`
	jwt.RegisteredClaims
}

// JWTMiddleware validates JWT tokens
func JWTMiddleware(jwtSecret string, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			logger.Warn("missing authorization header",
				zap.String("path", c.Request.URL.Path),
				zap.String("ip", c.ClientIP()),
			)
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "authorization header required",
			})
			c.Abort()
			return
		}

		// Check for "Bearer " prefix
		tokenString := authHeader
		if strings.HasPrefix(authHeader, "Bearer ") {
			tokenString = strings.TrimPrefix(authHeader, "Bearer ")
		}

		// Parse and validate token
		token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			// Validate signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(jwtSecret), nil
		})

		if err != nil {
			logger.Warn("invalid JWT token",
				zap.Error(err),
				zap.String("ip", c.ClientIP()),
			)
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "invalid token",
			})
			c.Abort()
			return
		}

		// Extract claims
		claims, ok := token.Claims.(*JWTClaims)
		if !ok || !token.Valid {
			logger.Warn("invalid JWT claims",
				zap.String("ip", c.ClientIP()),
			)
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "invalid token claims",
			})
			c.Abort()
			return
		}

		// Check token expiration
		if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now()) {
			logger.Warn("expired JWT token",
				zap.String("username", claims.Username),
				zap.Time("expired_at", claims.ExpiresAt.Time),
			)
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "token expired",
			})
			c.Abort()
			return
		}

		// Store claims in context
		c.Set("jwt_claims", claims)
		c.Set("username", claims.Username)
		c.Set("user_id", claims.Username) // Use username as user ID
		c.Set("role", claims.Role)
		c.Set("permissions", claims.Permissions)

		logger.Debug("JWT validated",
			zap.String("username", claims.Username),
			zap.String("role", claims.Role),
		)

		c.Next()
	}
}

// OptionalJWTMiddleware validates JWT if present, but doesn't require it
func OptionalJWTMiddleware(jwtSecret string, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// No token, continue without auth
			c.Next()
			return
		}

		// Token present, validate it
		tokenString := authHeader
		if strings.HasPrefix(authHeader, "Bearer ") {
			tokenString = strings.TrimPrefix(authHeader, "Bearer ")
		}

		token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(jwtSecret), nil
		})

		if err == nil && token.Valid {
			claims, ok := token.Claims.(*JWTClaims)
			if ok {
				c.Set("jwt_claims", claims)
				c.Set("username", claims.Username)
				c.Set("user_id", claims.Username)
				c.Set("role", claims.Role)
				c.Set("permissions", claims.Permissions)
			}
		}

		c.Next()
	}
}

// CORSMiddleware handles Cross-Origin Resource Sharing
func CORSMiddleware(allowedOrigins []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Check if origin is allowed
		allowed := false
		if len(allowedOrigins) == 0 || contains(allowedOrigins, "*") {
			allowed = true
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		} else {
			for _, allowedOrigin := range allowedOrigins {
				if origin == allowedOrigin {
					allowed = true
					c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
					break
				}
			}
		}

		if !allowed && origin != "" {
			c.Writer.Header().Set("Access-Control-Allow-Origin", "null")
		}

		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")
		c.Writer.Header().Set("Access-Control-Max-Age", "86400") // 24 hours

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// RequestLoggerMiddleware logs all requests
func RequestLoggerMiddleware(logger *zap.Logger, metrics *utils.PrometheusMetrics) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Process request
		c.Next()

		// Log after request completion
		duration := time.Since(start)

		logger.Info("request completed",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("query", c.Request.URL.RawQuery),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("duration", duration),
			zap.String("ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
			zap.Int("response_size", c.Writer.Size()),
		)

		// Record metrics
		if metrics != nil {
			status := fmt.Sprintf("%d", c.Writer.Status())
			metrics.RecordRequest(c.Request.Method, c.Request.URL.Path, status, duration)
		}
	}
}

// ErrorHandlerMiddleware handles errors and panics
func ErrorHandlerMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error("panic recovered",
					zap.Any("error", err),
					zap.String("path", c.Request.URL.Path),
					zap.String("ip", c.ClientIP()),
				)

				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "internal server error",
				})
			}
		}()

		c.Next()

		// Log errors from handlers
		if len(c.Errors) > 0 {
			for _, e := range c.Errors {
				logger.Error("handler error",
					zap.Error(e.Err),
					zap.Uint("type", uint(e.Type)),
					zap.String("path", c.Request.URL.Path),
				)
			}
		}
	}
}

// RateLimitMiddleware applies rate limiting
func RateLimitMiddleware(limiter *ratelimit.Limiter, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		userID := ""

		// Try to get user ID from context (if JWT validated)
		if val, exists := c.Get("user_id"); exists {
			if uid, ok := val.(string); ok {
				userID = uid
			}
		}

		// Check rate limit
		allowed, err := limiter.Allow(ip, userID)
		if !allowed {
			logger.Warn("rate limit exceeded",
				zap.String("ip", ip),
				zap.String("user_id", userID),
				zap.Error(err),
			)

			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "rate limit exceeded",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// SecurityHeadersMiddleware adds security headers
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Prevent clickjacking
		c.Writer.Header().Set("X-Frame-Options", "DENY")

		// Prevent MIME sniffing
		c.Writer.Header().Set("X-Content-Type-Options", "nosniff")

		// XSS protection
		c.Writer.Header().Set("X-XSS-Protection", "1; mode=block")

		// Referrer policy
		c.Writer.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// Content Security Policy
		c.Writer.Header().Set("Content-Security-Policy", "default-src 'self'")

		// HSTS (if HTTPS)
		if c.Request.TLS != nil {
			c.Writer.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		c.Next()
	}
}

// RequestIDMiddleware adds a unique request ID
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}

		c.Set("request_id", requestID)
		c.Writer.Header().Set("X-Request-ID", requestID)

		c.Next()
	}
}

// TimeoutMiddleware adds request timeout
func TimeoutMiddleware(timeout time.Duration, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		c.Request = c.Request.WithContext(ctx)

		finished := make(chan struct{})
		go func() {
			c.Next()
			finished <- struct{}{}
		}()

		select {
		case <-finished:
			// Request completed normally
			return

		case <-ctx.Done():
			// Request timed out
			logger.Warn("request timeout",
				zap.String("path", c.Request.URL.Path),
				zap.Duration("timeout", timeout),
			)

			c.JSON(http.StatusRequestTimeout, gin.H{
				"error": "request timeout",
			})
			c.Abort()
		}
	}
}

// PermissionMiddleware checks if user has required permission
func PermissionMiddleware(requiredPermission string, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		permissions, exists := c.Get("permissions")
		if !exists {
			logger.Warn("no permissions in context",
				zap.String("path", c.Request.URL.Path),
			)
			c.JSON(http.StatusForbidden, gin.H{
				"error": "insufficient permissions",
			})
			c.Abort()
			return
		}

		permStr, ok := permissions.(string)
		if !ok {
			logger.Warn("invalid permissions format",
				zap.String("path", c.Request.URL.Path),
			)
			c.JSON(http.StatusForbidden, gin.H{
				"error": "insufficient permissions",
			})
			c.Abort()
			return
		}

		// Check if user has required permission
		if !hasPermission(permStr, requiredPermission) {
			username, _ := c.Get("username")
			logger.Warn("insufficient permissions",
				zap.String("username", username.(string)),
				zap.String("required", requiredPermission),
				zap.String("actual", permStr),
			)
			c.JSON(http.StatusForbidden, gin.H{
				"error": "insufficient permissions",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// hasPermission checks if user has the required permission
func hasPermission(userPermissions, required string) bool {
	// Parse permission levels (READ=1, CREATE=2, UPDATE=3, DELETE=5, ALL=5)
	permissionLevels := map[string]int{
		"READ":   1,
		"CREATE": 2,
		"UPDATE": 3,
		"DELETE": 5,
		"ALL":    5,
	}

	userLevel := permissionLevels[userPermissions]
	requiredLevel := permissionLevels[required]

	return userLevel >= requiredLevel
}

// generateRequestID generates a unique request ID
func generateRequestID() string {
	return fmt.Sprintf("%d-%s", time.Now().UnixNano(), randomString(8))
}

// randomString generates a random string of given length
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}

// contains checks if a slice contains a string
func contains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

// RequestLogger creates a request logging middleware
func RequestLogger(logger *zap.Logger) gin.HandlerFunc {
	return RequestLoggerMiddleware(logger, utils.NewPrometheusMetrics())
}

// CORS creates a CORS middleware with default allowed origins
func CORS() gin.HandlerFunc {
	return CORSMiddleware([]string{"*"})
}

// RequestSize creates a middleware that limits request body size
func RequestSize(maxSize int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxSize)
		c.Next()
	}
}

// RateLimiter creates a rate limiting middleware
func RateLimiter(limiter *ratelimit.Limiter) gin.HandlerFunc {
	// Use a no-op logger since we already have request logging
	logger, _ := zap.NewProduction()
	return RateLimitMiddleware(limiter, logger)
}

// JWTAuth creates a JWT authentication middleware
func JWTAuth(secret string, logger *zap.Logger) gin.HandlerFunc {
	return JWTMiddleware(secret, logger)
}

// AdminOnly creates a middleware that only allows admin users
func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "authentication required",
			})
			c.Abort()
			return
		}

		roleStr, ok := role.(string)
		if !ok || (roleStr != "admin" && roleStr != "administrator") {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "admin access required",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
