package middleware

import (
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"golang.org/x/time/rate"

	"helixtrack.ru/chat/internal/logger"
	"helixtrack.ru/chat/internal/models"
)

// RateLimiter manages rate limiting per IP
type RateLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
	rate     int    // requests per second
	burst    int    // burst size
	cleanup  *time.Ticker
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(ratePerSecond, burst int) *RateLimiter {
	rl := &RateLimiter{
		limiters: make(map[string]*rate.Limiter),
		rate:     ratePerSecond,
		burst:    burst,
		cleanup:  time.NewTicker(5 * time.Minute),
	}

	// Start cleanup goroutine
	go rl.cleanupRoutine()

	return rl
}

// getLimiter returns the rate limiter for an IP
func (rl *RateLimiter) getLimiter(ip string) *rate.Limiter {
	rl.mu.RLock()
	limiter, exists := rl.limiters[ip]
	rl.mu.RUnlock()

	if exists {
		return limiter
	}

	// Create new limiter
	limiter = rate.NewLimiter(rate.Limit(rl.rate), rl.burst)

	rl.mu.Lock()
	rl.limiters[ip] = limiter
	rl.mu.Unlock()

	return limiter
}

// Allow checks if a request from an IP is allowed
func (rl *RateLimiter) Allow(ip string) bool {
	limiter := rl.getLimiter(ip)
	return limiter.Allow()
}

// cleanupRoutine periodically removes old limiters
func (rl *RateLimiter) cleanupRoutine() {
	for range rl.cleanup.C {
		rl.mu.Lock()
		// Clear all limiters (simple approach)
		// In production, you might want to track last access time
		if len(rl.limiters) > 1000 {
			rl.limiters = make(map[string]*rate.Limiter)
			logger.Debug("Rate limiter cleanup executed", zap.Int("cleared", len(rl.limiters)))
		}
		rl.mu.Unlock()
	}
}

// Stop stops the cleanup routine
func (rl *RateLimiter) Stop() {
	rl.cleanup.Stop()
}

// RateLimitMiddleware creates a rate limiting middleware
func RateLimitMiddleware(rl *RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()

		if !rl.Allow(ip) {
			logger.Warn("Rate limit exceeded",
				zap.String("ip", ip),
				zap.String("path", c.Request.URL.Path),
			)
			c.JSON(429, models.ErrorResponse(
				models.ErrorCodeRateLimitExceeded,
				"Rate limit exceeded. Please try again later.",
			))
			c.Abort()
			return
		}

		c.Next()
	}
}

// DDOSProtectionMiddleware provides DDOS protection
func DDOSProtectionMiddleware(config *models.SecurityConfig) gin.HandlerFunc {
	rl := NewRateLimiter(config.RateLimitPerSecond, config.RateLimitBurst)

	return func(c *gin.Context) {
		if !config.EnableDDOSProtection {
			c.Next()
			return
		}

		ip := c.ClientIP()

		if !rl.Allow(ip) {
			logger.Warn("DDOS protection triggered",
				zap.String("ip", ip),
				zap.String("path", c.Request.URL.Path),
				zap.String("method", c.Request.Method),
			)
			c.JSON(429, models.ErrorResponse(
				models.ErrorCodeRateLimitExceeded,
				"Too many requests. Please slow down.",
			))
			c.Abort()
			return
		}

		c.Next()
	}
}

// MessageSizeMiddleware limits request body size
func MessageSizeMiddleware(maxSize int) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check content length
		if c.Request.ContentLength > int64(maxSize) {
			logger.Warn("Message size exceeded",
				zap.String("ip", c.ClientIP()),
				zap.Int64("size", c.Request.ContentLength),
				zap.Int("max_size", maxSize),
			)
			c.JSON(413, models.ErrorResponse(
				models.ErrorCodeMessageTooLarge,
				"Request body too large",
			))
			c.Abort()
			return
		}

		// Limit request body reader
		c.Request.Body = &limitedReader{
			reader:   c.Request.Body,
			maxBytes: int64(maxSize),
		}

		c.Next()
	}
}

// limitedReader wraps io.ReadCloser with size limit
type limitedReader struct {
	reader   interface {
		Read([]byte) (int, error)
		Close() error
	}
	maxBytes int64
	read     int64
}

func (lr *limitedReader) Read(p []byte) (int, error) {
	if lr.read >= lr.maxBytes {
		return 0, &messageTooLargeError{}
	}

	n, err := lr.reader.Read(p)
	lr.read += int64(n)

	if lr.read > lr.maxBytes {
		return 0, &messageTooLargeError{}
	}

	return n, err
}

func (lr *limitedReader) Close() error {
	return lr.reader.Close()
}

type messageTooLargeError struct{}

func (e *messageTooLargeError) Error() string {
	return "message too large"
}
