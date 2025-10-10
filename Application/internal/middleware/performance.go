package middleware

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// CompressionMiddleware provides response compression
func CompressionMiddleware(level int) gin.HandlerFunc {
	if level < gzip.DefaultCompression || level > gzip.BestCompression {
		level = gzip.DefaultCompression
	}

	// Pool of gzip writers for reuse
	gzipPool := sync.Pool{
		New: func() interface{} {
			gz, _ := gzip.NewWriterLevel(io.Discard, level)
			return gz
		},
	}

	return func(c *gin.Context) {
		// Check if client supports gzip
		if !strings.Contains(c.GetHeader("Accept-Encoding"), "gzip") {
			c.Next()
			return
		}

		// Don't compress if content is already compressed
		if c.GetHeader("Content-Encoding") != "" {
			c.Next()
			return
		}

		// Get gzip writer from pool
		gz := gzipPool.Get().(*gzip.Writer)
		defer gzipPool.Put(gz)

		gz.Reset(c.Writer)
		defer gz.Close()

		// Wrap response writer
		c.Writer = &gzipResponseWriter{
			ResponseWriter: c.Writer,
			writer:         gz,
		}

		c.Header("Content-Encoding", "gzip")
		c.Header("Vary", "Accept-Encoding")

		c.Next()
	}
}

// gzipResponseWriter wraps gin.ResponseWriter with gzip compression
type gzipResponseWriter struct {
	gin.ResponseWriter
	writer *gzip.Writer
}

func (g *gzipResponseWriter) Write(data []byte) (int, error) {
	return g.writer.Write(data)
}

func (g *gzipResponseWriter) WriteString(s string) (int, error) {
	return g.writer.Write([]byte(s))
}

// RateLimiterConfig contains rate limiting configuration
type RateLimiterConfig struct {
	RequestsPerSecond int           // Maximum requests per second
	BurstSize         int           // Maximum burst size
	CleanupInterval   time.Duration // Cleanup interval for old entries
}

// DefaultRateLimiterConfig returns optimized default settings
func DefaultRateLimiterConfig() RateLimiterConfig {
	return RateLimiterConfig{
		RequestsPerSecond: 1000,              // 1000 req/sec
		BurstSize:         2000,              // Allow bursts up to 2000
		CleanupInterval:   1 * time.Minute,   // Cleanup every minute
	}
}

// tokenBucket represents a token bucket for rate limiting
type tokenBucket struct {
	tokens         float64
	lastRefill     time.Time
	maxTokens      float64
	refillRate     float64 // tokens per second
	mu             sync.Mutex
}

// rateLimiter implements rate limiting
type rateLimiter struct {
	buckets         map[string]*tokenBucket
	mu              sync.RWMutex
	config          RateLimiterConfig
	stopCleanup     chan struct{}
	cleanupDone     sync.WaitGroup
}

// newRateLimiter creates a new rate limiter
func newRateLimiter(cfg RateLimiterConfig) *rateLimiter {
	rl := &rateLimiter{
		buckets:     make(map[string]*tokenBucket),
		config:      cfg,
		stopCleanup: make(chan struct{}),
	}

	// Start background cleanup
	rl.cleanupDone.Add(1)
	go rl.cleanupLoop()

	return rl
}

// allow checks if a request is allowed
func (rl *rateLimiter) allow(key string) bool {
	rl.mu.RLock()
	bucket, exists := rl.buckets[key]
	rl.mu.RUnlock()

	if !exists {
		rl.mu.Lock()
		// Double-check after acquiring write lock
		bucket, exists = rl.buckets[key]
		if !exists {
			bucket = &tokenBucket{
				tokens:     float64(rl.config.BurstSize),
				lastRefill: time.Now(),
				maxTokens:  float64(rl.config.BurstSize),
				refillRate: float64(rl.config.RequestsPerSecond),
			}
			rl.buckets[key] = bucket
		}
		rl.mu.Unlock()
	}

	return bucket.take()
}

// take attempts to take a token from the bucket
func (tb *tokenBucket) take() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(tb.lastRefill).Seconds()

	// Refill tokens based on elapsed time
	tb.tokens += elapsed * tb.refillRate
	if tb.tokens > tb.maxTokens {
		tb.tokens = tb.maxTokens
	}
	tb.lastRefill = now

	// Check if we have tokens available
	if tb.tokens >= 1.0 {
		tb.tokens -= 1.0
		return true
	}

	return false
}

// cleanupLoop removes old buckets
func (rl *rateLimiter) cleanupLoop() {
	defer rl.cleanupDone.Done()

	ticker := time.NewTicker(rl.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rl.cleanup()
		case <-rl.stopCleanup:
			return
		}
	}
}

// cleanup removes inactive buckets
func (rl *rateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	for key, bucket := range rl.buckets {
		bucket.mu.Lock()
		inactive := now.Sub(bucket.lastRefill) > 5*time.Minute
		bucket.mu.Unlock()

		if inactive {
			delete(rl.buckets, key)
		}
	}
}

// close stops the rate limiter
func (rl *rateLimiter) close() {
	close(rl.stopCleanup)
	rl.cleanupDone.Wait()
}

// RateLimitMiddleware creates rate limiting middleware
func RateLimitMiddleware(cfg RateLimiterConfig) gin.HandlerFunc {
	limiter := newRateLimiter(cfg)

	return func(c *gin.Context) {
		// Use client IP as key
		key := c.ClientIP()

		if !limiter.allow(key) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded",
				"retry_after": "1s",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// CircuitBreakerConfig contains circuit breaker configuration
type CircuitBreakerConfig struct {
	MaxFailures     int           // Maximum failures before opening
	Timeout         time.Duration // Timeout before attempting to close
	HalfOpenMax     int           // Maximum requests in half-open state
	FailureRatio    float64       // Failure ratio threshold (0.0 - 1.0)
	MinRequests     int           // Minimum requests before evaluating ratio
}

// DefaultCircuitBreakerConfig returns optimized default settings
func DefaultCircuitBreakerConfig() CircuitBreakerConfig {
	return CircuitBreakerConfig{
		MaxFailures:  5,
		Timeout:      30 * time.Second,
		HalfOpenMax:  3,
		FailureRatio: 0.5,
		MinRequests:  10,
	}
}

// circuitState represents circuit breaker state
type circuitState int

const (
	stateClosed circuitState = iota
	stateOpen
	stateHalfOpen
)

// circuitBreaker implements circuit breaker pattern
type circuitBreaker struct {
	config          CircuitBreakerConfig
	state           circuitState
	failures        int
	successes       int
	totalRequests   int
	halfOpenCount   int
	lastFailureTime time.Time
	mu              sync.RWMutex
}

// newCircuitBreaker creates a new circuit breaker
func newCircuitBreaker(cfg CircuitBreakerConfig) *circuitBreaker {
	return &circuitBreaker{
		config: cfg,
		state:  stateClosed,
	}
}

// allow checks if request is allowed through circuit breaker
func (cb *circuitBreaker) allow() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case stateClosed:
		return true

	case stateOpen:
		// Check if we should transition to half-open
		if time.Since(cb.lastFailureTime) > cb.config.Timeout {
			cb.state = stateHalfOpen
			cb.halfOpenCount = 0
			return true
		}
		return false

	case stateHalfOpen:
		// Allow limited requests in half-open state
		if cb.halfOpenCount < cb.config.HalfOpenMax {
			cb.halfOpenCount++
			return true
		}
		return false

	default:
		return false
	}
}

// recordSuccess records a successful request
func (cb *circuitBreaker) recordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.successes++
	cb.totalRequests++

	if cb.state == stateHalfOpen {
		// Transition to closed if enough successes
		if cb.successes >= cb.config.HalfOpenMax {
			cb.state = stateClosed
			cb.failures = 0
			cb.successes = 0
			cb.totalRequests = 0
		}
	}
}

// recordFailure records a failed request
func (cb *circuitBreaker) recordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failures++
	cb.totalRequests++
	cb.lastFailureTime = time.Now()

	// Check if we should open the circuit
	if cb.state == stateClosed {
		// Check max failures
		if cb.failures >= cb.config.MaxFailures {
			cb.state = stateOpen
			return
		}

		// Check failure ratio
		if cb.totalRequests >= cb.config.MinRequests {
			ratio := float64(cb.failures) / float64(cb.totalRequests)
			if ratio >= cb.config.FailureRatio {
				cb.state = stateOpen
				return
			}
		}
	} else if cb.state == stateHalfOpen {
		// Any failure in half-open state opens the circuit
		cb.state = stateOpen
	}
}

// getState returns current circuit state
func (cb *circuitBreaker) getState() circuitState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// CircuitBreakerMiddleware creates circuit breaker middleware
func CircuitBreakerMiddleware(cfg CircuitBreakerConfig) gin.HandlerFunc {
	breaker := newCircuitBreaker(cfg)

	return func(c *gin.Context) {
		if !breaker.allow() {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error": "Service temporarily unavailable",
				"retry_after": cfg.Timeout.String(),
			})
			c.Abort()
			return
		}

		c.Next()

		// Record result based on status code
		if c.Writer.Status() >= 500 {
			breaker.recordFailure()
		} else {
			breaker.recordSuccess()
		}
	}
}

// TimeoutMiddleware adds request timeout
func TimeoutMiddleware(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Create context with timeout
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		// Replace request context
		c.Request = c.Request.WithContext(ctx)

		// Channel to signal when request is done
		finished := make(chan struct{})

		go func() {
			c.Next()
			close(finished)
		}()

		select {
		case <-finished:
			// Request completed successfully
			return
		case <-ctx.Done():
			// Timeout occurred
			c.JSON(http.StatusRequestTimeout, gin.H{
				"error": "Request timeout",
			})
			c.Abort()
		}
	}
}

// CORSConfig contains CORS configuration
type CORSConfig struct {
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	ExposeHeaders    []string
	AllowCredentials bool
	MaxAge           time.Duration
}

// DefaultCORSConfig returns default CORS configuration
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}
}

// CORSMiddleware creates CORS middleware
func CORSMiddleware(cfg CORSConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Set CORS headers
		if len(cfg.AllowOrigins) > 0 {
			if cfg.AllowOrigins[0] == "*" {
				c.Header("Access-Control-Allow-Origin", "*")
			} else {
				for _, allowedOrigin := range cfg.AllowOrigins {
					if origin == allowedOrigin {
						c.Header("Access-Control-Allow-Origin", origin)
						break
					}
				}
			}
		}

		if len(cfg.AllowMethods) > 0 {
			c.Header("Access-Control-Allow-Methods", strings.Join(cfg.AllowMethods, ", "))
		}

		if len(cfg.AllowHeaders) > 0 {
			c.Header("Access-Control-Allow-Headers", strings.Join(cfg.AllowHeaders, ", "))
		}

		if len(cfg.ExposeHeaders) > 0 {
			c.Header("Access-Control-Expose-Headers", strings.Join(cfg.ExposeHeaders, ", "))
		}

		if cfg.AllowCredentials {
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		if cfg.MaxAge > 0 {
			c.Header("Access-Control-Max-Age", fmt.Sprintf("%.0f", cfg.MaxAge.Seconds()))
		}

		// Handle preflight requests
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
