package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/helixtrack/localization-service/internal/models"
	"golang.org/x/time/rate"
)

// RateLimiter implements token bucket rate limiting
type RateLimiter struct {
	mu           sync.RWMutex
	ipLimiters   map[string]*rate.Limiter
	userLimiters map[string]*rate.Limiter
	globalLimiter *rate.Limiter
	ipRate       float64
	ipBurst      int
	userRate     float64
	userBurst    int
	cleanup      *time.Ticker
	done         chan struct{}
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(ipRequestsPerMin, userRequestsPerMin, globalRequestsPerMin int) *RateLimiter {
	rl := &RateLimiter{
		ipLimiters:   make(map[string]*rate.Limiter),
		userLimiters: make(map[string]*rate.Limiter),
		ipRate:       float64(ipRequestsPerMin) / 60.0,
		ipBurst:      ipRequestsPerMin / 10,
		userRate:     float64(userRequestsPerMin) / 60.0,
		userBurst:    userRequestsPerMin / 10,
		globalLimiter: rate.NewLimiter(rate.Limit(float64(globalRequestsPerMin)/60.0), globalRequestsPerMin/10),
		cleanup:      time.NewTicker(5 * time.Minute),
		done:         make(chan struct{}),
	}

	// Start cleanup goroutine
	go rl.cleanupOldLimiters()

	return rl
}

// RateLimit middleware
func (rl *RateLimiter) RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check global limit
		if !rl.globalLimiter.Allow() {
			c.JSON(http.StatusTooManyRequests, models.ErrorResponse(
				429,
				"global rate limit exceeded",
			))
			c.Abort()
			return
		}

		// Check IP limit
		ip := c.ClientIP()
		ipLimiter := rl.getIPLimiter(ip)
		if !ipLimiter.Allow() {
			c.JSON(http.StatusTooManyRequests, models.ErrorResponse(
				429,
				"IP rate limit exceeded",
			))
			c.Abort()
			return
		}

		// Check user limit (if authenticated)
		claims := GetClaims(c)
		if claims != nil && claims.Username != "" {
			userLimiter := rl.getUserLimiter(claims.Username)
			if !userLimiter.Allow() {
				c.JSON(http.StatusTooManyRequests, models.ErrorResponse(
					429,
					"user rate limit exceeded",
				))
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// getIPLimiter gets or creates a rate limiter for an IP
func (rl *RateLimiter) getIPLimiter(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limiter, exists := rl.ipLimiters[ip]
	if !exists {
		limiter = rate.NewLimiter(rate.Limit(rl.ipRate), rl.ipBurst)
		rl.ipLimiters[ip] = limiter
	}

	return limiter
}

// getUserLimiter gets or creates a rate limiter for a user
func (rl *RateLimiter) getUserLimiter(username string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limiter, exists := rl.userLimiters[username]
	if !exists {
		limiter = rate.NewLimiter(rate.Limit(rl.userRate), rl.userBurst)
		rl.userLimiters[username] = limiter
	}

	return limiter
}

// cleanupOldLimiters removes inactive rate limiters
func (rl *RateLimiter) cleanupOldLimiters() {
	for {
		select {
		case <-rl.cleanup.C:
			rl.mu.Lock()
			// In production, track last access time and remove inactive limiters
			// For simplicity, we'll keep all limiters for now
			rl.mu.Unlock()
		case <-rl.done:
			return
		}
	}
}

// Close stops the cleanup goroutine
func (rl *RateLimiter) Close() {
	close(rl.done)
	rl.cleanup.Stop()
}
