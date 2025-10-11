package security

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// BruteForceProtectionConfig contains brute force protection configuration
type BruteForceProtectionConfig struct {
	// Failure tracking
	MaxFailedAttempts    int           // Maximum failed attempts before blocking
	FailureWindow        time.Duration // Time window for counting failures
	BlockDuration        time.Duration // Duration to block after max failures
	PermanentBlockThreshold int        // Failures before permanent block

	// Progressive delays
	EnableProgressiveDelay bool          // Enable progressive delays
	BaseDelay              time.Duration // Base delay (increases exponentially)
	MaxDelay               time.Duration // Maximum delay

	// Account lockout
	EnableAccountLockout   bool          // Enable account lockout
	LockoutDuration        time.Duration // Duration to lock account
	NotifyOnLockout        bool          // Notify on account lockout

	// IP tracking
	TrackByIP              bool          // Track failures by IP
	TrackByUsername        bool          // Track failures by username
	TrackByIPAndUsername   bool          // Track failures by IP+username combination

	// Whitelist
	WhitelistedIPs         []string      // IPs exempt from protection
	WhitelistedUsernames   []string      // Usernames exempt from protection

	// CAPTCHA integration
	EnableCAPTCHA          bool          // Enable CAPTCHA after failures
	CAPTCHAThreshold       int           // Failures before requiring CAPTCHA

	// Cleanup
	CleanupInterval        time.Duration // Cleanup interval for old entries
}

// DefaultBruteForceProtectionConfig returns secure default settings
func DefaultBruteForceProtectionConfig() BruteForceProtectionConfig {
	return BruteForceProtectionConfig{
		MaxFailedAttempts:       5,
		FailureWindow:           15 * time.Minute,
		BlockDuration:           30 * time.Minute,
		PermanentBlockThreshold: 20,
		EnableProgressiveDelay:  true,
		BaseDelay:               1 * time.Second,
		MaxDelay:                30 * time.Second,
		EnableAccountLockout:    true,
		LockoutDuration:         1 * time.Hour,
		NotifyOnLockout:         true,
		TrackByIP:               true,
		TrackByUsername:         true,
		TrackByIPAndUsername:    true,
		WhitelistedIPs:          []string{},
		WhitelistedUsernames:    []string{},
		EnableCAPTCHA:           false,
		CAPTCHAThreshold:        3,
		CleanupInterval:         5 * time.Minute,
	}
}

// StrictBruteForceProtectionConfig returns very strict settings
func StrictBruteForceProtectionConfig() BruteForceProtectionConfig {
	cfg := DefaultBruteForceProtectionConfig()
	cfg.MaxFailedAttempts = 3
	cfg.BlockDuration = 1 * time.Hour
	cfg.PermanentBlockThreshold = 10
	cfg.LockoutDuration = 24 * time.Hour
	cfg.EnableCAPTCHA = true
	cfg.CAPTCHAThreshold = 2
	return cfg
}

// failureRecord tracks failed login attempts
type failureRecord struct {
	Attempts      int
	FirstAttempt  time.Time
	LastAttempt   time.Time
	BlockedUntil  time.Time
	PermanentBlock bool
	TotalFailures int
}

// bruteForceProtector implements brute force protection
type bruteForceProtector struct {
	config           BruteForceProtectionConfig
	ipFailures       map[string]*failureRecord
	usernameFailures map[string]*failureRecord
	combinedFailures map[string]*failureRecord // IP+Username
	mu               sync.RWMutex
	stopCleanup      chan struct{}
	cleanupDone      sync.WaitGroup
}

// newBruteForceProtector creates a new brute force protector
func newBruteForceProtector(cfg BruteForceProtectionConfig) *bruteForceProtector {
	bp := &bruteForceProtector{
		config:           cfg,
		ipFailures:       make(map[string]*failureRecord),
		usernameFailures: make(map[string]*failureRecord),
		combinedFailures: make(map[string]*failureRecord),
		stopCleanup:      make(chan struct{}),
	}

	// Start background cleanup
	bp.cleanupDone.Add(1)
	go bp.cleanupLoop()

	return bp
}

// checkAttempt checks if an attempt should be allowed
func (bp *bruteForceProtector) checkAttempt(ip, username string) (allowed bool, reason string, delay time.Duration) {
	// Check whitelists
	if bp.isWhitelisted(ip, username) {
		return true, "", 0
	}

	bp.mu.Lock()
	defer bp.mu.Unlock()

	now := time.Now()
	blocked := false
	var blockReason string
	maxDelay := time.Duration(0)

	// Check IP-based failures
	if bp.config.TrackByIP {
		if record, exists := bp.ipFailures[ip]; exists {
			isBlocked, reason, delay := bp.checkRecord(record, now, "IP")
			if isBlocked {
				blocked = true
				blockReason = reason
			}
			if delay > maxDelay {
				maxDelay = delay
			}
		}
	}

	// Check username-based failures
	if bp.config.TrackByUsername && username != "" {
		if record, exists := bp.usernameFailures[username]; exists {
			isBlocked, reason, delay := bp.checkRecord(record, now, "username")
			if isBlocked {
				blocked = true
				blockReason = reason
			}
			if delay > maxDelay {
				maxDelay = delay
			}
		}
	}

	// Check combined IP+username failures
	if bp.config.TrackByIPAndUsername && username != "" {
		key := fmt.Sprintf("%s:%s", ip, username)
		if record, exists := bp.combinedFailures[key]; exists {
			isBlocked, reason, delay := bp.checkRecord(record, now, "IP+username")
			if isBlocked {
				blocked = true
				blockReason = reason
			}
			if delay > maxDelay {
				maxDelay = delay
			}
		}
	}

	if blocked {
		return false, blockReason, maxDelay
	}

	return true, "", maxDelay
}

// checkRecord checks if a failure record indicates blocking
func (bp *bruteForceProtector) checkRecord(record *failureRecord, now time.Time, recordType string) (blocked bool, reason string, delay time.Duration) {
	// Check permanent block
	if record.PermanentBlock {
		return true, fmt.Sprintf("Permanently blocked (%s)", recordType), 0
	}

	// Check temporary block
	if now.Before(record.BlockedUntil) {
		remaining := record.BlockedUntil.Sub(now)
		return true, fmt.Sprintf("Temporarily blocked (%s) - %v remaining", recordType, remaining.Round(time.Second)), 0
	}

	// If block has expired, reset attempts
	if !record.BlockedUntil.IsZero() && now.After(record.BlockedUntil) {
		record.Attempts = 0
		record.BlockedUntil = time.Time{} // Clear block time
		return false, "", 0
	}

	// Check if we're in the failure window
	if now.Sub(record.FirstAttempt) > bp.config.FailureWindow {
		// Outside window, reset
		record.Attempts = 0
		record.FirstAttempt = now
		return false, "", 0
	}

	// Check if we've exceeded max attempts
	if record.Attempts >= bp.config.MaxFailedAttempts {
		return true, fmt.Sprintf("Too many failed attempts (%s)", recordType), 0
	}

	// Calculate progressive delay if enabled
	if bp.config.EnableProgressiveDelay && record.Attempts > 0 {
		delay = bp.calculateDelay(record.Attempts)
		return false, "", delay
	}

	return false, "", 0
}

// recordFailure records a failed attempt
func (bp *bruteForceProtector) recordFailure(ip, username string) {
	// Check whitelists
	if bp.isWhitelisted(ip, username) {
		return
	}

	bp.mu.Lock()
	defer bp.mu.Unlock()

	now := time.Now()

	// Record IP failure
	if bp.config.TrackByIP {
		bp.updateFailureRecord(bp.ipFailures, ip, now)
	}

	// Record username failure
	if bp.config.TrackByUsername && username != "" {
		bp.updateFailureRecord(bp.usernameFailures, username, now)
	}

	// Record combined failure
	if bp.config.TrackByIPAndUsername && username != "" {
		key := fmt.Sprintf("%s:%s", ip, username)
		bp.updateFailureRecord(bp.combinedFailures, key, now)
	}
}

// updateFailureRecord updates a failure record
func (bp *bruteForceProtector) updateFailureRecord(records map[string]*failureRecord, key string, now time.Time) {
	record, exists := records[key]
	if !exists {
		record = &failureRecord{
			FirstAttempt: now,
		}
		records[key] = record
	}

	// Reset if outside failure window
	if now.Sub(record.FirstAttempt) > bp.config.FailureWindow {
		record.Attempts = 0
		record.FirstAttempt = now
	}

	record.Attempts++
	record.LastAttempt = now
	record.TotalFailures++

	// Check if we should block
	if record.Attempts >= bp.config.MaxFailedAttempts {
		record.BlockedUntil = now.Add(bp.config.BlockDuration)

		// Check for permanent block
		if record.TotalFailures >= bp.config.PermanentBlockThreshold {
			record.PermanentBlock = true
		}
	}
}

// recordSuccess records a successful attempt (resets counters)
func (bp *bruteForceProtector) recordSuccess(ip, username string) {
	bp.mu.Lock()
	defer bp.mu.Unlock()

	// Reset IP failures
	if bp.config.TrackByIP {
		delete(bp.ipFailures, ip)
	}

	// Reset username failures
	if bp.config.TrackByUsername && username != "" {
		delete(bp.usernameFailures, username)
	}

	// Reset combined failures
	if bp.config.TrackByIPAndUsername && username != "" {
		key := fmt.Sprintf("%s:%s", ip, username)
		delete(bp.combinedFailures, key)
	}
}

// isWhitelisted checks if IP or username is whitelisted
func (bp *bruteForceProtector) isWhitelisted(ip, username string) bool {
	// Check IP whitelist
	for _, whitelistedIP := range bp.config.WhitelistedIPs {
		if ip == whitelistedIP {
			return true
		}
	}

	// Check username whitelist
	for _, whitelistedUsername := range bp.config.WhitelistedUsernames {
		if username == whitelistedUsername {
			return true
		}
	}

	return false
}

// calculateDelay calculates progressive delay based on attempt count
func (bp *bruteForceProtector) calculateDelay(attempts int) time.Duration {
	// Exponential backoff: baseDelay * 2^(attempts-1)
	delay := bp.config.BaseDelay * time.Duration(1<<uint(attempts-1))

	if delay > bp.config.MaxDelay {
		delay = bp.config.MaxDelay
	}

	return delay
}

// cleanupLoop runs background cleanup
func (bp *bruteForceProtector) cleanupLoop() {
	defer bp.cleanupDone.Done()

	ticker := time.NewTicker(bp.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			bp.cleanup()
		case <-bp.stopCleanup:
			return
		}
	}
}

// cleanup removes old entries
func (bp *bruteForceProtector) cleanup() {
	bp.mu.Lock()
	defer bp.mu.Unlock()

	now := time.Now()

	// Cleanup IP failures
	for ip, record := range bp.ipFailures {
		if !record.PermanentBlock && now.After(record.BlockedUntil) &&
			now.Sub(record.LastAttempt) > bp.config.FailureWindow {
			delete(bp.ipFailures, ip)
		}
	}

	// Cleanup username failures
	for username, record := range bp.usernameFailures {
		if !record.PermanentBlock && now.After(record.BlockedUntil) &&
			now.Sub(record.LastAttempt) > bp.config.FailureWindow {
			delete(bp.usernameFailures, username)
		}
	}

	// Cleanup combined failures
	for key, record := range bp.combinedFailures {
		if !record.PermanentBlock && now.After(record.BlockedUntil) &&
			now.Sub(record.LastAttempt) > bp.config.FailureWindow {
			delete(bp.combinedFailures, key)
		}
	}
}

// close stops the brute force protector
func (bp *bruteForceProtector) close() {
	close(bp.stopCleanup)
	bp.cleanupDone.Wait()
}

// getStatistics returns statistics
func (bp *bruteForceProtector) getStatistics() *BruteForceStatistics {
	bp.mu.RLock()
	defer bp.mu.RUnlock()

	stats := &BruteForceStatistics{
		TrackedIPs:       len(bp.ipFailures),
		TrackedUsernames: len(bp.usernameFailures),
		TrackedCombined:  len(bp.combinedFailures),
	}

	now := time.Now()

	// Count blocked entries
	for _, record := range bp.ipFailures {
		if record.PermanentBlock {
			stats.PermanentlyBlockedIPs++
		} else if now.Before(record.BlockedUntil) {
			stats.BlockedIPs++
		}
	}

	for _, record := range bp.usernameFailures {
		if record.PermanentBlock {
			stats.PermanentlyBlockedUsernames++
		} else if now.Before(record.BlockedUntil) {
			stats.BlockedUsernames++
		}
	}

	return stats
}

// Global brute force protector
var globalBruteForceProtector *bruteForceProtector

// InitBruteForceProtection initializes brute force protection
func InitBruteForceProtection(cfg BruteForceProtectionConfig) {
	if globalBruteForceProtector != nil {
		globalBruteForceProtector.close()
	}
	globalBruteForceProtector = newBruteForceProtector(cfg)
}

// BruteForceProtectionMiddleware creates brute force protection middleware
func BruteForceProtectionMiddleware(cfg BruteForceProtectionConfig) gin.HandlerFunc {
	protector := newBruteForceProtector(cfg)

	return func(c *gin.Context) {
		// Only protect authentication endpoints
		// This should be applied to login/authentication routes only

		ip := c.ClientIP()

		// Try to get username from request
		username := ""
		if user, exists := c.Get("username"); exists {
			username = user.(string)
		} else {
			// Try to extract from request body
			var body map[string]interface{}
			if err := c.ShouldBindJSON(&body); err == nil {
				if u, ok := body["username"].(string); ok {
					username = u
				} else if u, ok := body["email"].(string); ok {
					username = u
				}
			}
		}

		// Check if attempt should be allowed
		allowed, reason, delay := protector.checkAttempt(ip, username)

		if !allowed {
			LogSecurityEvent("BRUTE_FORCE_DETECTED", ip,
				fmt.Sprintf("Brute force attempt blocked for username '%s': %s",
					username, reason))

			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Too many failed attempts",
				"reason":      reason,
				"retry_after": cfg.BlockDuration.String(),
			})
			c.Abort()
			return
		}

		// Apply progressive delay if needed
		if delay > 0 {
			time.Sleep(delay)
		}

		// Process request
		c.Next()

		// Check if authentication was successful
		statusCode := c.Writer.Status()

		if statusCode == http.StatusOK || statusCode == http.StatusCreated {
			// Successful authentication
			protector.recordSuccess(ip, username)
		} else if statusCode == http.StatusUnauthorized || statusCode == http.StatusForbidden {
			// Failed authentication
			protector.recordFailure(ip, username)

			LogSecurityEvent("FAILED_LOGIN_ATTEMPT", ip,
				fmt.Sprintf("Failed login attempt for username '%s'", username))
		}
	}
}

// RecordLoginFailure manually records a login failure
func RecordLoginFailure(ip, username string) {
	if globalBruteForceProtector != nil {
		globalBruteForceProtector.recordFailure(ip, username)
	}
}

// RecordLoginSuccess manually records a login success
func RecordLoginSuccess(ip, username string) {
	if globalBruteForceProtector != nil {
		globalBruteForceProtector.recordSuccess(ip, username)
	}
}

// IsBlocked checks if IP or username is currently blocked
func IsBlocked(ip, username string) (bool, string) {
	if globalBruteForceProtector != nil {
		allowed, reason, _ := globalBruteForceProtector.checkAttempt(ip, username)
		return !allowed, reason
	}
	return false, ""
}

// UnblockIP unblocks an IP address
func UnblockIP(ip string) {
	if globalBruteForceProtector != nil {
		globalBruteForceProtector.mu.Lock()
		defer globalBruteForceProtector.mu.Unlock()

		// Delete IP-specific record
		delete(globalBruteForceProtector.ipFailures, ip)

		// Delete all combined IP+username records for this IP
		for key := range globalBruteForceProtector.combinedFailures {
			// Combined key format is "ip:username"
			if len(key) > len(ip) && key[:len(ip)] == ip && key[len(ip)] == ':' {
				delete(globalBruteForceProtector.combinedFailures, key)
			}
		}
	}
}

// UnblockUsername unblocks a username
func UnblockUsername(username string) {
	if globalBruteForceProtector != nil {
		globalBruteForceProtector.mu.Lock()
		defer globalBruteForceProtector.mu.Unlock()

		// Delete username-specific record
		delete(globalBruteForceProtector.usernameFailures, username)

		// Delete all combined IP+username records for this username
		for key := range globalBruteForceProtector.combinedFailures {
			// Combined key format is "ip:username"
			// Find the colon and check if username matches
			colonIdx := -1
			for i := range key {
				if key[i] == ':' {
					colonIdx = i
					break
				}
			}
			if colonIdx >= 0 && colonIdx+1 < len(key) && key[colonIdx+1:] == username {
				delete(globalBruteForceProtector.combinedFailures, key)
			}
		}
	}
}

// BruteForceStatistics contains brute force protection statistics
type BruteForceStatistics struct {
	TrackedIPs                  int `json:"tracked_ips"`
	TrackedUsernames            int `json:"tracked_usernames"`
	TrackedCombined             int `json:"tracked_combined"`
	BlockedIPs                  int `json:"blocked_ips"`
	BlockedUsernames            int `json:"blocked_usernames"`
	PermanentlyBlockedIPs       int `json:"permanently_blocked_ips"`
	PermanentlyBlockedUsernames int `json:"permanently_blocked_usernames"`
}

// GetBruteForceStatistics returns brute force protection statistics
func GetBruteForceStatistics() *BruteForceStatistics {
	if globalBruteForceProtector != nil {
		return globalBruteForceProtector.getStatistics()
	}
	return &BruteForceStatistics{}
}
