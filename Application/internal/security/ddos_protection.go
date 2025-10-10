package security

import (
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// DDoSProtectionConfig contains DDoS protection configuration
type DDoSProtectionConfig struct {
	// Rate limiting (per IP)
	MaxRequestsPerSecond  int           // Maximum requests per second per IP
	MaxRequestsPerMinute  int           // Maximum requests per minute per IP
	MaxRequestsPerHour    int           // Maximum requests per hour per IP
	BurstSize             int           // Maximum burst size

	// Connection limits
	MaxConcurrentPerIP    int           // Maximum concurrent connections per IP
	MaxTotalConcurrent    int           // Maximum total concurrent connections

	// Request size limits
	MaxRequestSize        int64         // Maximum request size in bytes
	MaxHeaderSize         int           // Maximum header size in bytes
	MaxURILength          int           // Maximum URI length

	// Timeouts
	RequestTimeout        time.Duration // Maximum request processing time
	SlowlorisTimeout      time.Duration // Timeout for slow requests (Slowloris attack)
	ReadTimeout           time.Duration // Read timeout
	WriteTimeout          time.Duration // Write timeout

	// Protection features
	EnableIPBlocking      bool          // Enable automatic IP blocking
	BlockDuration         time.Duration // Duration to block IP
	SuspiciousThreshold   int           // Requests before marking as suspicious
	BanThreshold          int           // Failed attempts before banning

	// Cleanup
	CleanupInterval       time.Duration // Cleanup interval for old entries
}

// DefaultDDoSProtectionConfig returns secure default settings
func DefaultDDoSProtectionConfig() DDoSProtectionConfig {
	return DDoSProtectionConfig{
		// Rate limiting
		MaxRequestsPerSecond: 100,              // 100 req/sec per IP
		MaxRequestsPerMinute: 3000,             // 3000 req/min per IP
		MaxRequestsPerHour:   50000,            // 50k req/hour per IP
		BurstSize:            200,              // Allow short bursts

		// Connection limits
		MaxConcurrentPerIP:   50,               // 50 concurrent per IP
		MaxTotalConcurrent:   10000,            // 10k total concurrent

		// Request size limits
		MaxRequestSize:       10 * 1024 * 1024, // 10MB max request
		MaxHeaderSize:        8192,             // 8KB max headers
		MaxURILength:         4096,             // 4KB max URI

		// Timeouts
		RequestTimeout:       30 * time.Second,
		SlowlorisTimeout:     10 * time.Second,
		ReadTimeout:          10 * time.Second,
		WriteTimeout:         10 * time.Second,

		// Protection
		EnableIPBlocking:     true,
		BlockDuration:        15 * time.Minute, // Block for 15 minutes
		SuspiciousThreshold:  500,              // 500 req/sec is suspicious
		BanThreshold:         10,               // 10 failed attempts = ban

		// Cleanup
		CleanupInterval:      1 * time.Minute,
	}
}

// ipStats tracks statistics for an IP address
type ipStats struct {
	// Request counts
	requestsLastSecond   int
	requestsLastMinute   int
	requestsLastHour     int
	lastSecond           time.Time
	lastMinute           time.Time
	lastHour             time.Time

	// Connection tracking
	concurrentRequests   int

	// Blocking
	blocked              bool
	blockExpiry          time.Time
	failedAttempts       int
	suspiciousActivity   bool

	// Timing
	lastRequest          time.Time
	firstRequest         time.Time

	mu                   sync.RWMutex
}

// ddosProtector implements DDoS protection
type ddosProtector struct {
	config              DDoSProtectionConfig
	ipStats             map[string]*ipStats
	ipStatsMu           sync.RWMutex

	// Global statistics
	totalConcurrent     int
	totalConcurrentMu   sync.RWMutex

	// Blocked IPs
	blockedIPs          map[string]time.Time
	blockedIPsMu        sync.RWMutex

	// Whitelisted IPs
	whitelistedIPs      map[string]bool
	whitelistedIPsMu    sync.RWMutex

	// Cleanup
	stopCleanup         chan struct{}
	cleanupDone         sync.WaitGroup
}

// newDDoSProtector creates a new DDoS protector
func newDDoSProtector(cfg DDoSProtectionConfig) *ddosProtector {
	dp := &ddosProtector{
		config:         cfg,
		ipStats:        make(map[string]*ipStats),
		blockedIPs:     make(map[string]time.Time),
		whitelistedIPs: make(map[string]bool),
		stopCleanup:    make(chan struct{}),
	}

	// Start background cleanup
	dp.cleanupDone.Add(1)
	go dp.cleanupLoop()

	return dp
}

// checkRequest checks if a request should be allowed
func (dp *ddosProtector) checkRequest(ip string) (allowed bool, reason string) {
	// Check if IP is whitelisted
	dp.whitelistedIPsMu.RLock()
	if dp.whitelistedIPs[ip] {
		dp.whitelistedIPsMu.RUnlock()
		return true, ""
	}
	dp.whitelistedIPsMu.RUnlock()

	// Check if IP is blocked
	dp.blockedIPsMu.RLock()
	if blockExpiry, blocked := dp.blockedIPs[ip]; blocked {
		if time.Now().Before(blockExpiry) {
			dp.blockedIPsMu.RUnlock()
			return false, "IP address is blocked"
		}
		// Block expired, remove it
		dp.blockedIPsMu.RUnlock()
		dp.blockedIPsMu.Lock()
		delete(dp.blockedIPs, ip)
		dp.blockedIPsMu.Unlock()
	} else {
		dp.blockedIPsMu.RUnlock()
	}

	// Check global concurrent limit
	dp.totalConcurrentMu.RLock()
	if dp.totalConcurrent >= dp.config.MaxTotalConcurrent {
		dp.totalConcurrentMu.RUnlock()
		return false, "Server at maximum capacity"
	}
	dp.totalConcurrentMu.RUnlock()

	// Get or create IP stats
	stats := dp.getOrCreateIPStats(ip)
	stats.mu.Lock()
	defer stats.mu.Unlock()

	now := time.Now()

	// Check if IP is blocked
	if stats.blocked && now.Before(stats.blockExpiry) {
		return false, "IP address is temporarily blocked"
	}
	if stats.blocked && now.After(stats.blockExpiry) {
		stats.blocked = false
		stats.failedAttempts = 0
	}

	// Update rate counters
	if now.Sub(stats.lastSecond) >= time.Second {
		stats.requestsLastSecond = 0
		stats.lastSecond = now
	}
	if now.Sub(stats.lastMinute) >= time.Minute {
		stats.requestsLastMinute = 0
		stats.lastMinute = now
	}
	if now.Sub(stats.lastHour) >= time.Hour {
		stats.requestsLastHour = 0
		stats.lastHour = now
	}

	// Check rate limits
	if stats.requestsLastSecond >= dp.config.MaxRequestsPerSecond {
		stats.suspiciousActivity = true
		if stats.requestsLastSecond >= dp.config.SuspiciousThreshold {
			dp.blockIP(ip, "Excessive requests per second")
			return false, "Rate limit exceeded - IP blocked"
		}
		return false, "Rate limit exceeded (per second)"
	}

	if stats.requestsLastMinute >= dp.config.MaxRequestsPerMinute {
		return false, "Rate limit exceeded (per minute)"
	}

	if stats.requestsLastHour >= dp.config.MaxRequestsPerHour {
		return false, "Rate limit exceeded (per hour)"
	}

	// Check concurrent connections
	if stats.concurrentRequests >= dp.config.MaxConcurrentPerIP {
		return false, "Too many concurrent connections"
	}

	// Increment counters
	stats.requestsLastSecond++
	stats.requestsLastMinute++
	stats.requestsLastHour++
	stats.concurrentRequests++
	stats.lastRequest = now

	if stats.firstRequest.IsZero() {
		stats.firstRequest = now
	}

	// Increment global counter
	dp.totalConcurrentMu.Lock()
	dp.totalConcurrent++
	dp.totalConcurrentMu.Unlock()

	return true, ""
}

// releaseRequest releases a concurrent request slot
func (dp *ddosProtector) releaseRequest(ip string) {
	stats := dp.getIPStats(ip)
	if stats != nil {
		stats.mu.Lock()
		if stats.concurrentRequests > 0 {
			stats.concurrentRequests--
		}
		stats.mu.Unlock()
	}

	dp.totalConcurrentMu.Lock()
	if dp.totalConcurrent > 0 {
		dp.totalConcurrent--
	}
	dp.totalConcurrentMu.Unlock()
}

// blockIP blocks an IP address
func (dp *ddosProtector) blockIP(ip, reason string) {
	if !dp.config.EnableIPBlocking {
		return
	}

	stats := dp.getIPStats(ip)
	if stats != nil {
		stats.mu.Lock()
		stats.blocked = true
		stats.blockExpiry = time.Now().Add(dp.config.BlockDuration)
		stats.failedAttempts++
		stats.mu.Unlock()
	}

	dp.blockedIPsMu.Lock()
	dp.blockedIPs[ip] = time.Now().Add(dp.config.BlockDuration)
	dp.blockedIPsMu.Unlock()

	LogSecurityEvent("IP_BLOCKED", ip, reason)
}

// whitelistIP adds an IP to whitelist
func (dp *ddosProtector) whitelistIP(ip string) {
	dp.whitelistedIPsMu.Lock()
	dp.whitelistedIPs[ip] = true
	dp.whitelistedIPsMu.Unlock()
}

// removeWhitelistIP removes an IP from whitelist
func (dp *ddosProtector) removeWhitelistIP(ip string) {
	dp.whitelistedIPsMu.Lock()
	delete(dp.whitelistedIPs, ip)
	dp.whitelistedIPsMu.Unlock()
}

// unblockIP unblocks an IP address
func (dp *ddosProtector) unblockIP(ip string) {
	stats := dp.getIPStats(ip)
	if stats != nil {
		stats.mu.Lock()
		stats.blocked = false
		stats.failedAttempts = 0
		stats.mu.Unlock()
	}

	dp.blockedIPsMu.Lock()
	delete(dp.blockedIPs, ip)
	dp.blockedIPsMu.Unlock()
}

// getOrCreateIPStats gets or creates IP statistics
func (dp *ddosProtector) getOrCreateIPStats(ip string) *ipStats {
	dp.ipStatsMu.RLock()
	stats, exists := dp.ipStats[ip]
	dp.ipStatsMu.RUnlock()

	if exists {
		return stats
	}

	dp.ipStatsMu.Lock()
	// Double-check after acquiring write lock
	stats, exists = dp.ipStats[ip]
	if !exists {
		stats = &ipStats{
			lastSecond: time.Now(),
			lastMinute: time.Now(),
			lastHour:   time.Now(),
		}
		dp.ipStats[ip] = stats
	}
	dp.ipStatsMu.Unlock()

	return stats
}

// getIPStats gets IP statistics if they exist
func (dp *ddosProtector) getIPStats(ip string) *ipStats {
	dp.ipStatsMu.RLock()
	defer dp.ipStatsMu.RUnlock()
	return dp.ipStats[ip]
}

// cleanupLoop runs background cleanup
func (dp *ddosProtector) cleanupLoop() {
	defer dp.cleanupDone.Done()

	ticker := time.NewTicker(dp.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			dp.cleanup()
		case <-dp.stopCleanup:
			return
		}
	}
}

// cleanup removes old entries
func (dp *ddosProtector) cleanup() {
	now := time.Now()

	// Cleanup IP stats
	dp.ipStatsMu.Lock()
	for ip, stats := range dp.ipStats {
		stats.mu.RLock()
		inactive := now.Sub(stats.lastRequest) > 5*time.Minute
		stats.mu.RUnlock()

		if inactive {
			delete(dp.ipStats, ip)
		}
	}
	dp.ipStatsMu.Unlock()

	// Cleanup expired blocks
	dp.blockedIPsMu.Lock()
	for ip, expiry := range dp.blockedIPs {
		if now.After(expiry) {
			delete(dp.blockedIPs, ip)
		}
	}
	dp.blockedIPsMu.Unlock()
}

// close stops the DDoS protector
func (dp *ddosProtector) close() {
	close(dp.stopCleanup)
	dp.cleanupDone.Wait()
}

// DDoSProtectionMiddleware creates DDoS protection middleware
func DDoSProtectionMiddleware(cfg DDoSProtectionConfig) gin.HandlerFunc {
	protector := newDDoSProtector(cfg)

	return func(c *gin.Context) {
		// Get client IP
		ip := c.ClientIP()

		// Check request size
		if c.Request.ContentLength > cfg.MaxRequestSize {
			LogSecurityEvent("REQUEST_TOO_LARGE", ip, "Request size exceeded limit")
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{
				"error": "Request entity too large",
			})
			c.Abort()
			return
		}

		// Check URI length
		if len(c.Request.RequestURI) > cfg.MaxURILength {
			LogSecurityEvent("URI_TOO_LONG", ip, "URI length exceeded limit")
			c.JSON(http.StatusRequestURITooLong, gin.H{
				"error": "Request URI too long",
			})
			c.Abort()
			return
		}

		// Check if request should be allowed
		allowed, reason := protector.checkRequest(ip)
		if !allowed {
			LogSecurityEvent("REQUEST_BLOCKED", ip, reason)
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": reason,
				"retry_after": cfg.BlockDuration.String(),
			})
			c.Abort()
			return
		}

		// Process request
		c.Next()

		// Release concurrent request slot
		protector.releaseRequest(ip)

		// Check for failed authentication (401, 403)
		if c.Writer.Status() == http.StatusUnauthorized || c.Writer.Status() == http.StatusForbidden {
			stats := protector.getIPStats(ip)
			if stats != nil {
				stats.mu.Lock()
				stats.failedAttempts++
				if stats.failedAttempts >= cfg.BanThreshold {
					protector.blockIP(ip, "Too many failed authentication attempts")
				}
				stats.mu.Unlock()
			}
		}
	}
}

// extractIPFromContext extracts IP address from request
func extractIPFromContext(c *gin.Context) string {
	// Try X-Forwarded-For header first
	forwarded := c.GetHeader("X-Forwarded-For")
	if forwarded != "" {
		// Take first IP in the list
		if ip := net.ParseIP(forwarded); ip != nil {
			return ip.String()
		}
	}

	// Try X-Real-IP header
	realIP := c.GetHeader("X-Real-IP")
	if realIP != "" {
		if ip := net.ParseIP(realIP); ip != nil {
			return ip.String()
		}
	}

	// Fall back to RemoteAddr
	ip, _, _ := net.SplitHostPort(c.Request.RemoteAddr)
	return ip
}

// DDoSStatistics contains DDoS protection statistics
type DDoSStatistics struct {
	TotalRequests     int64
	BlockedRequests   int64
	BlockedIPs        int
	WhitelistedIPs    int
	ActiveConnections int
	TrackedIPs        int
}

// GetStatistics returns DDoS protection statistics
func (dp *ddosProtector) GetStatistics() *DDoSStatistics {
	dp.ipStatsMu.RLock()
	trackedIPs := len(dp.ipStats)
	dp.ipStatsMu.RUnlock()

	dp.blockedIPsMu.RLock()
	blockedIPs := len(dp.blockedIPs)
	dp.blockedIPsMu.RUnlock()

	dp.whitelistedIPsMu.RLock()
	whitelistedIPs := len(dp.whitelistedIPs)
	dp.whitelistedIPsMu.RUnlock()

	dp.totalConcurrentMu.RLock()
	activeConnections := dp.totalConcurrent
	dp.totalConcurrentMu.RUnlock()

	return &DDoSStatistics{
		TrackedIPs:        trackedIPs,
		BlockedIPs:        blockedIPs,
		WhitelistedIPs:    whitelistedIPs,
		ActiveConnections: activeConnections,
	}
}
