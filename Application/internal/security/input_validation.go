package security

import (
	"fmt"
	"html"
	"net/http"
	"regexp"
	"strings"
	"unicode"

	"github.com/gin-gonic/gin"
)

// InputValidationConfig contains input validation configuration
type InputValidationConfig struct {
	// SQL Injection Protection
	EnableSQLInjectionCheck   bool
	BlockSQLKeywords          bool

	// XSS Protection
	EnableXSSCheck            bool
	SanitizeHTML              bool
	AllowedHTMLTags           []string

	// Path Traversal Protection
	EnablePathTraversalCheck  bool

	// Command Injection Protection
	EnableCommandInjectionCheck bool

	// LDAP Injection Protection
	EnableLDAPInjectionCheck  bool

	// Max lengths
	MaxStringLength           int
	MaxArrayLength            int
	MaxJSONDepth              int

	// Character restrictions
	AllowUnicode              bool
	AllowSpecialChars         bool
	AllowedSpecialChars       string
}

// DefaultInputValidationConfig returns secure default settings
func DefaultInputValidationConfig() InputValidationConfig {
	return InputValidationConfig{
		EnableSQLInjectionCheck:     true,
		BlockSQLKeywords:            true,
		EnableXSSCheck:              true,
		SanitizeHTML:                true,
		AllowedHTMLTags:             []string{}, // No HTML allowed by default
		EnablePathTraversalCheck:    true,
		EnableCommandInjectionCheck: true,
		EnableLDAPInjectionCheck:    true,
		MaxStringLength:             10000,
		MaxArrayLength:              1000,
		MaxJSONDepth:                10,
		AllowUnicode:                true,
		AllowSpecialChars:           true,
		AllowedSpecialChars:         "!@#$%^&*()_+-=[]{}|;:',.<>?/~` ",
	}
}

// SQL injection patterns
var sqlInjectionPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)(union.*select)`),
	regexp.MustCompile(`(?i)(select.*from)`),
	regexp.MustCompile(`(?i)(insert.*into)`),
	regexp.MustCompile(`(?i)(delete.*from)`),
	regexp.MustCompile(`(?i)(drop.*table)`),
	regexp.MustCompile(`(?i)(update.*set)`),
	regexp.MustCompile(`(?i)(exec(ute)?)`),
	regexp.MustCompile(`(?i)(--)`),
	regexp.MustCompile(`(?i)(;.*--)`),
	regexp.MustCompile(`(?i)('.*or.*'.*=.*')`),
	regexp.MustCompile(`(?i)(or.*1.*=.*1)`),
	regexp.MustCompile(`(?i)(and.*1.*=.*1)`),
	regexp.MustCompile(`(?i)(having)`),
	regexp.MustCompile(`(?i)(group.*by)`),
	regexp.MustCompile(`(?i)(order.*by)`),
	regexp.MustCompile(`(?i)(waitfor.*delay)`),
	regexp.MustCompile(`(?i)(benchmark)`),
	regexp.MustCompile(`(?i)(sleep\()`),
}

// XSS patterns
var xssPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)<script`),
	regexp.MustCompile(`(?i)</script>`),
	regexp.MustCompile(`(?i)javascript:`),
	regexp.MustCompile(`(?i)onerror\s*=`),
	regexp.MustCompile(`(?i)onload\s*=`),
	regexp.MustCompile(`(?i)onclick\s*=`),
	regexp.MustCompile(`(?i)onmouseover\s*=`),
	regexp.MustCompile(`(?i)<iframe`),
	regexp.MustCompile(`(?i)<object`),
	regexp.MustCompile(`(?i)<embed`),
	regexp.MustCompile(`(?i)<img.*src`),
	regexp.MustCompile(`(?i)eval\(`),
	regexp.MustCompile(`(?i)expression\(`),
	regexp.MustCompile(`(?i)vbscript:`),
	regexp.MustCompile(`(?i)data:text/html`),
}

// Path traversal patterns
var pathTraversalPatterns = []*regexp.Regexp{
	regexp.MustCompile(`\.\.\/`),
	regexp.MustCompile(`\.\.\\`),
	regexp.MustCompile(`%2e%2e%2f`),
	regexp.MustCompile(`%2e%2e\\`),
	regexp.MustCompile(`\.\.%2f`),
	regexp.MustCompile(`\.\.%5c`),
}

// Command injection patterns
var commandInjectionPatterns = []*regexp.Regexp{
	regexp.MustCompile(`[;&|]`),
	regexp.MustCompile(`\$\(`),
	regexp.MustCompile("`.*`"),
	regexp.MustCompile(`>\s*/dev/null`),
	regexp.MustCompile(`&&`),
	regexp.MustCompile(`\|\|`),
}

// LDAP injection patterns
var ldapInjectionPatterns = []*regexp.Regexp{
	regexp.MustCompile(`\*`),
	regexp.MustCompile(`\(\)`),
	regexp.MustCompile(`\|\|`),
	regexp.MustCompile(`&&`),
}

// ValidateString validates a string input
func ValidateString(input string, cfg InputValidationConfig) (valid bool, sanitized string, reason string) {
	// Check length
	if len(input) > cfg.MaxStringLength {
		return false, "", fmt.Sprintf("String exceeds maximum length of %d", cfg.MaxStringLength)
	}

	// Check for SQL injection
	if cfg.EnableSQLInjectionCheck {
		for _, pattern := range sqlInjectionPatterns {
			if pattern.MatchString(input) {
				return false, "", "Potential SQL injection detected"
			}
		}
	}

	// Check for XSS
	if cfg.EnableXSSCheck {
		for _, pattern := range xssPatterns {
			if pattern.MatchString(input) {
				return false, "", "Potential XSS attack detected"
			}
		}
	}

	// Check for path traversal
	if cfg.EnablePathTraversalCheck {
		for _, pattern := range pathTraversalPatterns {
			if pattern.MatchString(input) {
				return false, "", "Potential path traversal detected"
			}
		}
	}

	// Check for command injection
	if cfg.EnableCommandInjectionCheck {
		for _, pattern := range commandInjectionPatterns {
			if pattern.MatchString(input) {
				return false, "", "Potential command injection detected"
			}
		}
	}

	// Check for LDAP injection
	if cfg.EnableLDAPInjectionCheck {
		for _, pattern := range ldapInjectionPatterns {
			if pattern.MatchString(input) {
				return false, "", "Potential LDAP injection detected"
			}
		}
	}

	// Sanitize HTML if enabled
	sanitized = input
	if cfg.SanitizeHTML {
		sanitized = html.EscapeString(input)
	}

	// Check character restrictions
	if !cfg.AllowUnicode || !cfg.AllowSpecialChars {
		for _, r := range input {
			if !cfg.AllowUnicode && r > unicode.MaxASCII {
				return false, "", "Unicode characters not allowed"
			}

			if !cfg.AllowSpecialChars {
				if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != ' ' {
					if !strings.ContainsRune(cfg.AllowedSpecialChars, r) {
						return false, "", fmt.Sprintf("Special character not allowed: %c", r)
					}
				}
			}
		}
	}

	return true, sanitized, ""
}

// SanitizeInput sanitizes input by removing dangerous characters
func SanitizeInput(input string) string {
	// Remove null bytes
	input = strings.ReplaceAll(input, "\x00", "")

	// Trim whitespace
	input = strings.TrimSpace(input)

	// Escape HTML
	input = html.EscapeString(input)

	return input
}

// SanitizeFilename sanitizes a filename
func SanitizeFilename(filename string) string {
	// Remove path separators
	filename = strings.ReplaceAll(filename, "/", "")
	filename = strings.ReplaceAll(filename, "\\", "")

	// Remove null bytes
	filename = strings.ReplaceAll(filename, "\x00", "")

	// Remove dangerous characters
	dangerousChars := []string{"<", ">", ":", "\"", "|", "?", "*"}
	for _, char := range dangerousChars {
		filename = strings.ReplaceAll(filename, char, "")
	}

	return filename
}

// SanitizeURL sanitizes a URL
func SanitizeURL(url string) (string, error) {
	// Check for javascript: protocol
	if strings.HasPrefix(strings.ToLower(url), "javascript:") {
		return "", fmt.Errorf("javascript: protocol not allowed")
	}

	// Check for data: protocol
	if strings.HasPrefix(strings.ToLower(url), "data:") {
		return "", fmt.Errorf("data: protocol not allowed")
	}

	// Only allow http: and https:
	if !strings.HasPrefix(strings.ToLower(url), "http://") &&
		!strings.HasPrefix(strings.ToLower(url), "https://") {
		return "", fmt.Errorf("only http: and https: protocols allowed")
	}

	return url, nil
}

// ValidateEmail validates an email address
func ValidateEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

// ValidateUsername validates a username
func ValidateUsername(username string) (bool, string) {
	if len(username) < 3 {
		return false, "Username must be at least 3 characters"
	}
	if len(username) > 50 {
		return false, "Username must be at most 50 characters"
	}

	// Allow only alphanumeric and underscore
	usernameRegex := regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
	if !usernameRegex.MatchString(username) {
		return false, "Username can only contain letters, numbers, and underscores"
	}

	return true, ""
}

// ValidatePassword validates a password
func ValidatePassword(password string) (bool, string) {
	if len(password) < 8 {
		return false, "Password must be at least 8 characters"
	}
	if len(password) > 128 {
		return false, "Password must be at most 128 characters"
	}

	// Check for at least one uppercase, one lowercase, one digit
	hasUpper := false
	hasLower := false
	hasDigit := false
	hasSpecial := false

	for _, r := range password {
		if unicode.IsUpper(r) {
			hasUpper = true
		}
		if unicode.IsLower(r) {
			hasLower = true
		}
		if unicode.IsDigit(r) {
			hasDigit = true
		}
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			hasSpecial = true
		}
	}

	if !hasUpper {
		return false, "Password must contain at least one uppercase letter"
	}
	if !hasLower {
		return false, "Password must contain at least one lowercase letter"
	}
	if !hasDigit {
		return false, "Password must contain at least one digit"
	}
	if !hasSpecial {
		return false, "Password must contain at least one special character"
	}

	return true, ""
}

// InputValidationMiddleware creates input validation middleware
func InputValidationMiddleware(cfg InputValidationConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Validate query parameters
		for key, values := range c.Request.URL.Query() {
			for _, value := range values {
				valid, _, reason := ValidateString(value, cfg)
				if !valid {
					LogSecurityEvent("INVALID_INPUT", c.ClientIP(),
						fmt.Sprintf("Invalid query parameter %s: %s", key, reason))
					c.JSON(http.StatusBadRequest, gin.H{
						"error": "Invalid input detected",
					})
					c.Abort()
					return
				}
			}
		}

		// Validate headers
		suspiciousHeaders := []string{
			"X-Forwarded-Host",
			"X-Original-URL",
			"X-Rewrite-URL",
		}

		for _, header := range suspiciousHeaders {
			if value := c.GetHeader(header); value != "" {
				valid, _, reason := ValidateString(value, cfg)
				if !valid {
					LogSecurityEvent("INVALID_INPUT", c.ClientIP(),
						fmt.Sprintf("Invalid header %s: %s", header, reason))
					c.JSON(http.StatusBadRequest, gin.H{
						"error": "Invalid input detected",
					})
					c.Abort()
					return
				}
			}
		}

		c.Next()
	}
}

// SQLInjectionPattern checks for SQL injection patterns
func SQLInjectionPattern(input string) bool {
	for _, pattern := range sqlInjectionPatterns {
		if pattern.MatchString(input) {
			return true
		}
	}
	return false
}

// XSSPattern checks for XSS patterns
func XSSPattern(input string) bool {
	for _, pattern := range xssPatterns {
		if pattern.MatchString(input) {
			return true
		}
	}
	return false
}

// PathTraversalPattern checks for path traversal patterns
func PathTraversalPattern(input string) bool {
	for _, pattern := range pathTraversalPatterns {
		if pattern.MatchString(input) {
			return true
		}
	}
	return false
}
