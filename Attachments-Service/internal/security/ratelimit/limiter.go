package ratelimit

import (
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// Limiter implements token bucket rate limiting with multiple strategies
type Limiter struct {
	config      *LimiterConfig
	logger      *zap.Logger
	ipBuckets   map[string]*TokenBucket
	userBuckets map[string]*TokenBucket
	globalBucket *TokenBucket
	mu          sync.RWMutex
	cleanup     *time.Ticker
	stopCleanup chan struct{}
}

// LimiterConfig contains rate limiter configuration
type LimiterConfig struct {
	// Per-IP rate limiting
	EnableIPRateLimit bool
	IPRequestsPerSecond int
	IPBurstSize int

	// Per-user rate limiting
	EnableUserRateLimit bool
	UserRequestsPerSecond int
	UserBurstSize int

	// Global rate limiting
	EnableGlobalRateLimit bool
	GlobalRequestsPerSecond int
	GlobalBurstSize int

	// Upload-specific limits
	UploadRequestsPerMinute int
	UploadBurstSize int

	// Download-specific limits
	DownloadRequestsPerMinute int
	DownloadBurstSize int

	// Whitelist/Blacklist
	WhitelistedIPs []string
	BlacklistedIPs []string

	// Cleanup settings
	CleanupInterval time.Duration
	BucketExpiry time.Duration
}

// DefaultLimiterConfig returns default rate limiter configuration
func DefaultLimiterConfig() *LimiterConfig {
	return &LimiterConfig{
		EnableIPRateLimit: true,
		IPRequestsPerSecond: 10,
		IPBurstSize: 20,

		EnableUserRateLimit: true,
		UserRequestsPerSecond: 20,
		UserBurstSize: 40,

		EnableGlobalRateLimit: true,
		GlobalRequestsPerSecond: 1000,
		GlobalBurstSize: 2000,

		UploadRequestsPerMinute: 100,
		UploadBurstSize: 20,

		DownloadRequestsPerMinute: 500,
		DownloadBurstSize: 100,

		WhitelistedIPs: []string{},
		BlacklistedIPs: []string{},

		CleanupInterval: 5 * time.Minute,
		BucketExpiry: 15 * time.Minute,
	}
}

// NewLimiter creates a new rate limiter
func NewLimiter(config *LimiterConfig, logger *zap.Logger) *Limiter {
	if config == nil {
		config = DefaultLimiterConfig()
	}

	limiter := &Limiter{
		config: config,
		logger: logger,
		ipBuckets: make(map[string]*TokenBucket),
		userBuckets: make(map[string]*TokenBucket),
		stopCleanup: make(chan struct{}),
	}

	// Create global bucket
	if config.EnableGlobalRateLimit {
		limiter.globalBucket = NewTokenBucket(
			config.GlobalRequestsPerSecond,
			config.GlobalBurstSize,
		)
	}

	// Start cleanup goroutine
	limiter.cleanup = time.NewTicker(config.CleanupInterval)
	go limiter.cleanupExpiredBuckets()

	return limiter
}

// Allow checks if a request should be allowed
func (l *Limiter) Allow(ip, userID string) (bool, error) {
	// Check blacklist
	if l.isBlacklisted(ip) {
		l.logger.Warn("request from blacklisted IP",
			zap.String("ip", ip),
		)
		return false, fmt.Errorf("IP address is blacklisted")
	}

	// Check whitelist (skip rate limiting)
	if l.isWhitelisted(ip) {
		return true, nil
	}

	// Check global rate limit
	if l.config.EnableGlobalRateLimit {
		if !l.globalBucket.Allow() {
			l.logger.Warn("global rate limit exceeded",
				zap.String("ip", ip),
				zap.String("user_id", userID),
			)
			return false, fmt.Errorf("global rate limit exceeded")
		}
	}

	// Check IP rate limit
	if l.config.EnableIPRateLimit && ip != "" {
		bucket := l.getIPBucket(ip)
		if !bucket.Allow() {
			l.logger.Warn("IP rate limit exceeded",
				zap.String("ip", ip),
			)
			return false, fmt.Errorf("rate limit exceeded for IP: %s", ip)
		}
	}

	// Check user rate limit
	if l.config.EnableUserRateLimit && userID != "" {
		bucket := l.getUserBucket(userID)
		if !bucket.Allow() {
			l.logger.Warn("user rate limit exceeded",
				zap.String("user_id", userID),
			)
			return false, fmt.Errorf("rate limit exceeded for user: %s", userID)
		}
	}

	return true, nil
}

// AllowUpload checks if an upload request should be allowed
func (l *Limiter) AllowUpload(ip, userID string) (bool, error) {
	// First check general rate limits
	allowed, err := l.Allow(ip, userID)
	if !allowed {
		return false, err
	}

	// Then check upload-specific limit
	key := fmt.Sprintf("upload:%s", ip)
	if userID != "" {
		key = fmt.Sprintf("upload:%s", userID)
	}

	bucket := l.getCustomBucket(key, l.config.UploadRequestsPerMinute/60, l.config.UploadBurstSize)
	if !bucket.Allow() {
		l.logger.Warn("upload rate limit exceeded",
			zap.String("ip", ip),
			zap.String("user_id", userID),
		)
		return false, fmt.Errorf("upload rate limit exceeded")
	}

	return true, nil
}

// AllowDownload checks if a download request should be allowed
func (l *Limiter) AllowDownload(ip, userID string) (bool, error) {
	// First check general rate limits
	allowed, err := l.Allow(ip, userID)
	if !allowed {
		return false, err
	}

	// Then check download-specific limit
	key := fmt.Sprintf("download:%s", ip)
	if userID != "" {
		key = fmt.Sprintf("download:%s", userID)
	}

	bucket := l.getCustomBucket(key, l.config.DownloadRequestsPerMinute/60, l.config.DownloadBurstSize)
	if !bucket.Allow() {
		l.logger.Warn("download rate limit exceeded",
			zap.String("ip", ip),
			zap.String("user_id", userID),
		)
		return false, fmt.Errorf("download rate limit exceeded")
	}

	return true, nil
}

// getIPBucket returns token bucket for an IP address
func (l *Limiter) getIPBucket(ip string) *TokenBucket {
	l.mu.RLock()
	bucket, exists := l.ipBuckets[ip]
	l.mu.RUnlock()

	if exists {
		return bucket
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	// Double-check after acquiring write lock
	bucket, exists = l.ipBuckets[ip]
	if exists {
		return bucket
	}

	// Create new bucket
	bucket = NewTokenBucket(l.config.IPRequestsPerSecond, l.config.IPBurstSize)
	l.ipBuckets[ip] = bucket

	return bucket
}

// getUserBucket returns token bucket for a user
func (l *Limiter) getUserBucket(userID string) *TokenBucket {
	l.mu.RLock()
	bucket, exists := l.userBuckets[userID]
	l.mu.RUnlock()

	if exists {
		return bucket
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	// Double-check after acquiring write lock
	bucket, exists = l.userBuckets[userID]
	if exists {
		return bucket
	}

	// Create new bucket
	bucket = NewTokenBucket(l.config.UserRequestsPerSecond, l.config.UserBurstSize)
	l.userBuckets[userID] = bucket

	return bucket
}

// getCustomBucket returns or creates a custom token bucket
func (l *Limiter) getCustomBucket(key string, rate, burst int) *TokenBucket {
	l.mu.RLock()
	bucket, exists := l.ipBuckets[key] // Reuse ipBuckets map for custom buckets
	l.mu.RUnlock()

	if exists {
		return bucket
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	// Double-check
	bucket, exists = l.ipBuckets[key]
	if exists {
		return bucket
	}

	bucket = NewTokenBucket(rate, burst)
	l.ipBuckets[key] = bucket

	return bucket
}

// isWhitelisted checks if an IP is whitelisted
func (l *Limiter) isWhitelisted(ip string) bool {
	for _, whitelisted := range l.config.WhitelistedIPs {
		if ip == whitelisted {
			return true
		}
	}
	return false
}

// isBlacklisted checks if an IP is blacklisted
func (l *Limiter) isBlacklisted(ip string) bool {
	for _, blacklisted := range l.config.BlacklistedIPs {
		if ip == blacklisted {
			return true
		}
	}
	return false
}

// AddToBlacklist adds an IP to the blacklist
func (l *Limiter) AddToBlacklist(ip string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	for _, existing := range l.config.BlacklistedIPs {
		if existing == ip {
			return // Already blacklisted
		}
	}

	l.config.BlacklistedIPs = append(l.config.BlacklistedIPs, ip)

	l.logger.Info("IP added to blacklist",
		zap.String("ip", ip),
	)
}

// RemoveFromBlacklist removes an IP from the blacklist
func (l *Limiter) RemoveFromBlacklist(ip string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	for i, existing := range l.config.BlacklistedIPs {
		if existing == ip {
			l.config.BlacklistedIPs = append(l.config.BlacklistedIPs[:i], l.config.BlacklistedIPs[i+1:]...)
			l.logger.Info("IP removed from blacklist",
				zap.String("ip", ip),
			)
			return
		}
	}
}

// cleanupExpiredBuckets removes inactive token buckets
func (l *Limiter) cleanupExpiredBuckets() {
	for {
		select {
		case <-l.cleanup.C:
			l.mu.Lock()

			// Clean IP buckets
			for ip, bucket := range l.ipBuckets {
				if time.Since(bucket.lastAccess) > l.config.BucketExpiry {
					delete(l.ipBuckets, ip)
				}
			}

			// Clean user buckets
			for userID, bucket := range l.userBuckets {
				if time.Since(bucket.lastAccess) > l.config.BucketExpiry {
					delete(l.userBuckets, userID)
				}
			}

			l.mu.Unlock()

			l.logger.Debug("rate limiter cleanup complete",
				zap.Int("ip_buckets", len(l.ipBuckets)),
				zap.Int("user_buckets", len(l.userBuckets)),
			)

		case <-l.stopCleanup:
			l.cleanup.Stop()
			return
		}
	}
}

// Close stops the rate limiter
func (l *Limiter) Close() {
	close(l.stopCleanup)
}

// GetStats returns rate limiter statistics
func (l *Limiter) GetStats() *LimiterStats {
	l.mu.RLock()
	defer l.mu.RUnlock()

	stats := &LimiterStats{
		IPBuckets: len(l.ipBuckets),
		UserBuckets: len(l.userBuckets),
		BlacklistedIPs: len(l.config.BlacklistedIPs),
		WhitelistedIPs: len(l.config.WhitelistedIPs),
	}

	if l.globalBucket != nil {
		stats.GlobalTokens = l.globalBucket.Available()
	}

	return stats
}

// LimiterStats contains rate limiter statistics
type LimiterStats struct {
	IPBuckets int
	UserBuckets int
	BlacklistedIPs int
	WhitelistedIPs int
	GlobalTokens int
}

// TokenBucket implements the token bucket algorithm
type TokenBucket struct {
	rate int // Tokens per second
	burst int // Maximum bucket size
	tokens float64
	lastRefill time.Time
	lastAccess time.Time
	mu sync.Mutex
}

// NewTokenBucket creates a new token bucket
func NewTokenBucket(rate, burst int) *TokenBucket {
	now := time.Now()
	return &TokenBucket{
		rate: rate,
		burst: burst,
		tokens: float64(burst),
		lastRefill: now,
		lastAccess: now,
		mu: sync.Mutex{},
	}
}

// Allow checks if a request is allowed and consumes a token
func (tb *TokenBucket) Allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	tb.lastAccess = now

	// Refill tokens based on time elapsed
	elapsed := now.Sub(tb.lastRefill).Seconds()
	tb.tokens += elapsed * float64(tb.rate)

	// Cap at burst size
	if tb.tokens > float64(tb.burst) {
		tb.tokens = float64(tb.burst)
	}

	tb.lastRefill = now

	// Check if we have tokens available
	if tb.tokens >= 1.0 {
		tb.tokens -= 1.0
		return true
	}

	return false
}

// Available returns the number of available tokens
func (tb *TokenBucket) Available() int {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	// Refill first
	now := time.Now()
	elapsed := now.Sub(tb.lastRefill).Seconds()
	tokens := tb.tokens + elapsed*float64(tb.rate)

	if tokens > float64(tb.burst) {
		tokens = float64(tb.burst)
	}

	return int(tokens)
}

// Reset resets the token bucket to full
func (tb *TokenBucket) Reset() {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	tb.tokens = float64(tb.burst)
	tb.lastRefill = time.Now()
}
