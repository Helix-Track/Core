package security

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// CSRFProtectionConfig contains CSRF protection configuration
type CSRFProtectionConfig struct {
	// Token configuration
	TokenLength       int           // Length of CSRF token in bytes (default: 32)
	TokenLifetime     time.Duration // Lifetime of CSRF token (default: 1 hour)
	CookieName        string        // Name of CSRF cookie (default: "csrf_token")
	HeaderName        string        // Name of CSRF header (default: "X-CSRF-Token")
	FormFieldName     string        // Name of form field (default: "csrf_token")

	// Cookie configuration
	CookiePath        string        // Cookie path (default: "/")
	CookieDomain      string        // Cookie domain (default: "")
	CookieSecure      bool          // Secure cookie (HTTPS only, default: true)
	CookieHTTPOnly    bool          // HTTPOnly cookie (default: true)
	CookieSameSite    http.SameSite // SameSite cookie attribute (default: Strict)

	// Protection options
	RequireTokenRefresh bool          // Require token refresh on each request
	EnableDoubleSubmit  bool          // Enable double-submit cookie pattern
	TrustedOrigins      []string      // List of trusted origins
	ExcludePaths        []string      // Paths to exclude from CSRF protection
	ExcludeMethods      []string      // HTTP methods to exclude (default: GET, HEAD, OPTIONS)

	// Error handling
	ErrorHandler      func(*gin.Context) // Custom error handler
	RegenerateOnError bool               // Regenerate token on validation error
}

// DefaultCSRFProtectionConfig returns secure default settings
func DefaultCSRFProtectionConfig() CSRFProtectionConfig {
	return CSRFProtectionConfig{
		TokenLength:         32,
		TokenLifetime:       1 * time.Hour,
		CookieName:          "csrf_token",
		HeaderName:          "X-CSRF-Token",
		FormFieldName:       "csrf_token",
		CookiePath:          "/",
		CookieSecure:        true,
		CookieHTTPOnly:      true,
		CookieSameSite:      http.SameSiteLaxMode,
		RequireTokenRefresh: false,
		EnableDoubleSubmit:  true,
		TrustedOrigins:      []string{},
		ExcludePaths:        []string{},
		ExcludeMethods:      []string{"GET", "HEAD", "OPTIONS"},
		RegenerateOnError:   true,
	}
}

// StrictCSRFProtectionConfig returns very strict CSRF protection settings
func StrictCSRFProtectionConfig() CSRFProtectionConfig {
	cfg := DefaultCSRFProtectionConfig()
	cfg.TokenLifetime = 15 * time.Minute
	cfg.CookieSameSite = http.SameSiteStrictMode
	cfg.RequireTokenRefresh = true
	cfg.RegenerateOnError = true
	return cfg
}

// csrfToken represents a CSRF token with metadata
type csrfToken struct {
	Value      string
	CreatedAt  time.Time
	ExpiresAt  time.Time
	Used       bool
	IPAddress  string
	UserAgent  string
}

// csrfTokenStore manages CSRF tokens
type csrfTokenStore struct {
	tokens    map[string]*csrfToken
	mu        sync.RWMutex
	maxTokens int
}

// newCSRFTokenStore creates a new token store
func newCSRFTokenStore(maxTokens int) *csrfTokenStore {
	store := &csrfTokenStore{
		tokens:    make(map[string]*csrfToken),
		maxTokens: maxTokens,
	}

	// Start cleanup goroutine
	go store.cleanupExpired()

	return store
}

// generateToken generates a new CSRF token
func (store *csrfTokenStore) generateToken(length int, lifetime time.Duration, ip, userAgent string) (*csrfToken, error) {
	// Generate random bytes
	tokenBytes := make([]byte, length)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// Encode to base64
	tokenValue := base64.URLEncoding.EncodeToString(tokenBytes)

	// Create token
	now := time.Now()
	token := &csrfToken{
		Value:     tokenValue,
		CreatedAt: now,
		ExpiresAt: now.Add(lifetime),
		Used:      false,
		IPAddress: ip,
		UserAgent: userAgent,
	}

	// Store token
	store.mu.Lock()
	defer store.mu.Unlock()

	// Check if we need to cleanup
	if len(store.tokens) >= store.maxTokens {
		store.removeOldest()
	}

	store.tokens[tokenValue] = token

	return token, nil
}

// validateToken validates a CSRF token
func (store *csrfTokenStore) validateToken(tokenValue, ip, userAgent string, requireMatch bool) bool {
	store.mu.RLock()
	token, exists := store.tokens[tokenValue]
	store.mu.RUnlock()

	if !exists {
		return false
	}

	// Check if token is expired
	if time.Now().After(token.ExpiresAt) {
		store.removeToken(tokenValue)
		return false
	}

	// Check if token was already used (one-time use)
	if token.Used {
		return false
	}

	// Check IP address match if required
	if requireMatch && token.IPAddress != ip {
		return false
	}

	// Check user agent match if required
	if requireMatch && token.UserAgent != userAgent {
		return false
	}

	return true
}

// markTokenUsed marks a token as used
func (store *csrfTokenStore) markTokenUsed(tokenValue string) {
	store.mu.Lock()
	defer store.mu.Unlock()

	if token, exists := store.tokens[tokenValue]; exists {
		token.Used = true
	}
}

// removeToken removes a token from the store
func (store *csrfTokenStore) removeToken(tokenValue string) {
	store.mu.Lock()
	defer store.mu.Unlock()
	delete(store.tokens, tokenValue)
}

// removeOldest removes the oldest token
func (store *csrfTokenStore) removeOldest() {
	var oldestKey string
	var oldestTime time.Time

	for key, token := range store.tokens {
		if oldestKey == "" || token.CreatedAt.Before(oldestTime) {
			oldestKey = key
			oldestTime = token.CreatedAt
		}
	}

	if oldestKey != "" {
		delete(store.tokens, oldestKey)
	}
}

// cleanupExpired removes expired tokens
func (store *csrfTokenStore) cleanupExpired() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		store.mu.Lock()
		now := time.Now()
		for key, token := range store.tokens {
			if now.After(token.ExpiresAt) {
				delete(store.tokens, key)
			}
		}
		store.mu.Unlock()
	}
}

// getStats returns token store statistics
func (store *csrfTokenStore) getStats() map[string]interface{} {
	store.mu.RLock()
	defer store.mu.RUnlock()

	totalTokens := len(store.tokens)
	usedTokens := 0
	expiredTokens := 0
	now := time.Now()

	for _, token := range store.tokens {
		if token.Used {
			usedTokens++
		}
		if now.After(token.ExpiresAt) {
			expiredTokens++
		}
	}

	return map[string]interface{}{
		"total_tokens":   totalTokens,
		"used_tokens":    usedTokens,
		"expired_tokens": expiredTokens,
		"active_tokens":  totalTokens - usedTokens - expiredTokens,
	}
}

// Global token store
var globalCSRFStore = newCSRFTokenStore(10000)

// CSRFProtectionMiddleware creates CSRF protection middleware
func CSRFProtectionMiddleware(cfg CSRFProtectionConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if path is excluded
		for _, path := range cfg.ExcludePaths {
			if c.Request.URL.Path == path {
				c.Next()
				return
			}
		}

		// Check if method is excluded (GET, HEAD, OPTIONS are safe)
		for _, method := range cfg.ExcludeMethods {
			if c.Request.Method == method {
				// Generate and set token for safe methods
				token, err := generateAndSetCSRFToken(c, cfg)
				if err != nil {
					LogSecurityEvent("CSRF_TOKEN_GENERATION_FAILED", c.ClientIP(),
						fmt.Sprintf("Failed to generate CSRF token: %v", err))
				} else {
					// Make token available to response
					c.Set("csrf_token", token)
				}
				c.Next()
				return
			}
		}

		// For state-changing methods (POST, PUT, DELETE, PATCH), validate token
		valid := validateCSRFToken(c, cfg)

		if !valid {
			LogSecurityEvent("CSRF_DETECTED", c.ClientIP(),
				fmt.Sprintf("CSRF token validation failed for %s %s",
					c.Request.Method, c.Request.URL.Path))

			// Call custom error handler if provided
			if cfg.ErrorHandler != nil {
				cfg.ErrorHandler(c)
			} else {
				c.JSON(http.StatusForbidden, gin.H{
					"error": "CSRF token validation failed",
				})
			}

			c.Abort()
			return
		}

		// Token is valid, mark as used if one-time use is enabled
		if cfg.RequireTokenRefresh {
			tokenValue := getCSRFTokenFromRequest(c, cfg)
			globalCSRFStore.markTokenUsed(tokenValue)
		}

		// Generate new token for response
		newToken, err := generateAndSetCSRFToken(c, cfg)
		if err != nil {
			LogSecurityEvent("CSRF_TOKEN_GENERATION_FAILED", c.ClientIP(),
				fmt.Sprintf("Failed to generate new CSRF token: %v", err))
		} else {
			c.Set("csrf_token", newToken)
		}

		c.Next()
	}
}

// generateAndSetCSRFToken generates a new CSRF token and sets it in cookie
func generateAndSetCSRFToken(c *gin.Context, cfg CSRFProtectionConfig) (string, error) {
	ip := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	token, err := globalCSRFStore.generateToken(cfg.TokenLength, cfg.TokenLifetime, ip, userAgent)
	if err != nil {
		return "", err
	}

	// Set cookie
	c.SetCookie(
		cfg.CookieName,
		token.Value,
		int(cfg.TokenLifetime.Seconds()),
		cfg.CookiePath,
		cfg.CookieDomain,
		cfg.CookieSecure,
		cfg.CookieHTTPOnly,
	)

	// Also set SameSite attribute
	// Note: Gin's SetCookie doesn't support SameSite directly,
	// so we need to set it manually
	if cfg.CookieSameSite != http.SameSiteDefaultMode {
		sameSite := "Lax"
		if cfg.CookieSameSite == http.SameSiteStrictMode {
			sameSite = "Strict"
		} else if cfg.CookieSameSite == http.SameSiteNoneMode {
			sameSite = "None"
		}

		cookie := fmt.Sprintf("%s=%s; Path=%s; Max-Age=%d; SameSite=%s",
			cfg.CookieName, token.Value, cfg.CookiePath,
			int(cfg.TokenLifetime.Seconds()), sameSite)

		if cfg.CookieSecure {
			cookie += "; Secure"
		}
		if cfg.CookieHTTPOnly {
			cookie += "; HttpOnly"
		}
		if cfg.CookieDomain != "" {
			cookie += fmt.Sprintf("; Domain=%s", cfg.CookieDomain)
		}

		c.Header("Set-Cookie", cookie)
	}

	return token.Value, nil
}

// validateCSRFToken validates the CSRF token from request
func validateCSRFToken(c *gin.Context, cfg CSRFProtectionConfig) bool {
	// Get token from request
	requestToken := getCSRFTokenFromRequest(c, cfg)
	if requestToken == "" {
		return false
	}

	// Get token from cookie
	cookieToken, err := c.Cookie(cfg.CookieName)
	if err != nil {
		return false
	}

	// Double-submit cookie pattern: compare request token with cookie token
	if cfg.EnableDoubleSubmit {
		// Use constant-time comparison to prevent timing attacks
		if subtle.ConstantTimeCompare([]byte(requestToken), []byte(cookieToken)) != 1 {
			return false
		}
	}

	// Validate token in store
	ip := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	return globalCSRFStore.validateToken(requestToken, ip, userAgent, true)
}

// getCSRFTokenFromRequest extracts CSRF token from request
func getCSRFTokenFromRequest(c *gin.Context, cfg CSRFProtectionConfig) string {
	// Try header first
	token := c.GetHeader(cfg.HeaderName)
	if token != "" {
		return token
	}

	// Try form field
	token = c.PostForm(cfg.FormFieldName)
	if token != "" {
		return token
	}

	// Try JSON body
	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err == nil {
		if tokenValue, ok := body[cfg.FormFieldName].(string); ok {
			return tokenValue
		}
	}

	return ""
}

// GetCSRFToken returns the current CSRF token for the request
func GetCSRFToken(c *gin.Context) string {
	if token, exists := c.Get("csrf_token"); exists {
		if tokenStr, ok := token.(string); ok {
			return tokenStr
		}
	}
	return ""
}

// CSRFTokenResponse adds CSRF token to response
func CSRFTokenResponse() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Add CSRF token to response headers
		if token := GetCSRFToken(c); token != "" {
			c.Header("X-CSRF-Token", token)
		}
	}
}

// OriginValidationMiddleware validates the Origin header
func OriginValidationMiddleware(trustedOrigins []string) gin.HandlerFunc {
	trustedMap := make(map[string]bool)
	for _, origin := range trustedOrigins {
		trustedMap[origin] = true
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin == "" {
			// No Origin header, allow (might be same-origin request)
			c.Next()
			return
		}

		// Check if origin is trusted
		if !trustedMap[origin] {
			LogSecurityEvent("UNTRUSTED_ORIGIN", c.ClientIP(),
				fmt.Sprintf("Request from untrusted origin: %s", origin))

			c.JSON(http.StatusForbidden, gin.H{
				"error": "Untrusted origin",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RefererValidationMiddleware validates the Referer header
func RefererValidationMiddleware(trustedDomains []string) gin.HandlerFunc {
	trustedMap := make(map[string]bool)
	for _, domain := range trustedDomains {
		trustedMap[domain] = true
	}

	return func(c *gin.Context) {
		referer := c.GetHeader("Referer")
		if referer == "" {
			// No Referer header for state-changing methods
			if c.Request.Method != "GET" && c.Request.Method != "HEAD" && c.Request.Method != "OPTIONS" {
				LogSecurityEvent("MISSING_REFERER", c.ClientIP(),
					fmt.Sprintf("State-changing request without Referer: %s %s",
						c.Request.Method, c.Request.URL.Path))

				c.JSON(http.StatusForbidden, gin.H{
					"error": "Missing Referer header",
				})
				c.Abort()
				return
			}
			c.Next()
			return
		}

		// Extract domain from referer
		// Simple validation - can be enhanced
		trusted := false
		for domain := range trustedMap {
			if len(referer) >= len(domain) && referer[:len(domain)] == domain {
				trusted = true
				break
			}
		}

		if !trusted {
			LogSecurityEvent("UNTRUSTED_REFERER", c.ClientIP(),
				fmt.Sprintf("Request from untrusted referer: %s", referer))

			c.JSON(http.StatusForbidden, gin.H{
				"error": "Untrusted referer",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// CSRFStatistics contains CSRF protection statistics
type CSRFStatistics struct {
	TotalTokens       int `json:"total_tokens"`
	ActiveTokens      int `json:"active_tokens"`
	UsedTokens        int `json:"used_tokens"`
	ExpiredTokens     int `json:"expired_tokens"`
	ValidationsFailed int `json:"validations_failed"`
	ValidationsSucceeded int `json:"validations_succeeded"`
}

// GetCSRFStatistics returns CSRF protection statistics
func GetCSRFStatistics() *CSRFStatistics {
	stats := globalCSRFStore.getStats()

	return &CSRFStatistics{
		TotalTokens:   stats["total_tokens"].(int),
		ActiveTokens:  stats["active_tokens"].(int),
		UsedTokens:    stats["used_tokens"].(int),
		ExpiredTokens: stats["expired_tokens"].(int),
	}
}

// ClearExpiredCSRFTokens manually clears expired tokens
func ClearExpiredCSRFTokens() int {
	globalCSRFStore.mu.Lock()
	defer globalCSRFStore.mu.Unlock()

	count := 0
	now := time.Now()

	for key, token := range globalCSRFStore.tokens {
		if now.After(token.ExpiresAt) {
			delete(globalCSRFStore.tokens, key)
			count++
		}
	}

	return count
}
