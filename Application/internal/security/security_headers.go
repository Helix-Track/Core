package security

import (
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// SecurityHeadersConfig contains security headers configuration
type SecurityHeadersConfig struct {
	// HSTS (HTTP Strict Transport Security)
	EnableHSTS            bool
	HSTSMaxAge            int    // Seconds (default: 31536000 = 1 year)
	HSTSIncludeSubdomains bool
	HSTSPreload           bool

	// CSP (Content Security Policy)
	EnableCSP             bool
	CSPDirectives         map[string][]string
	CSPReportOnly         bool
	CSPReportURI          string

	// X-Frame-Options (Clickjacking protection)
	EnableFrameOptions    bool
	FrameOption           string // DENY, SAMEORIGIN, or ALLOW-FROM uri

	// X-Content-Type-Options (MIME sniffing protection)
	EnableContentTypeOptions bool

	// X-XSS-Protection
	EnableXSSProtection   bool
	XSSProtectionMode     string // "0", "1", "1; mode=block"

	// Referrer-Policy
	EnableReferrerPolicy  bool
	ReferrerPolicy        string // no-referrer, strict-origin-when-cross-origin, etc.

	// Permissions-Policy (formerly Feature-Policy)
	EnablePermissionsPolicy bool
	PermissionsDirectives   map[string]string

	// Additional security headers
	EnableExpectCT        bool
	ExpectCTMaxAge        int
	ExpectCTEnforce       bool
	ExpectCTReportURI     string

	// Cross-Origin headers
	EnableCORP            bool // Cross-Origin-Resource-Policy
	CORPPolicy            string // same-site, same-origin, cross-origin

	EnableCOEP            bool // Cross-Origin-Embedder-Policy
	COEPPolicy            string // require-corp, credentialless

	EnableCOOP            bool // Cross-Origin-Opener-Policy
	COOPPolicy            string // same-origin, same-origin-allow-popups, unsafe-none

	// Server header
	RemoveServerHeader    bool
	CustomServerHeader    string

	// X-Powered-By header
	RemovePoweredByHeader bool
}

// DefaultSecurityHeadersConfig returns secure default settings
func DefaultSecurityHeadersConfig() SecurityHeadersConfig {
	return SecurityHeadersConfig{
		// HSTS - Force HTTPS for 1 year
		EnableHSTS:            true,
		HSTSMaxAge:            31536000, // 1 year
		HSTSIncludeSubdomains: true,
		HSTSPreload:           true,

		// CSP - Strict content security policy
		EnableCSP:      true,
		CSPDirectives: map[string][]string{
			"default-src": {"'self'"},
			"script-src":  {"'self'", "'unsafe-inline'"}, // Allow inline scripts for now
			"style-src":   {"'self'", "'unsafe-inline'"}, // Allow inline styles for now
			"img-src":     {"'self'", "data:", "https:"},
			"font-src":    {"'self'", "data:"},
			"connect-src": {"'self'"},
			"media-src":   {"'self'"},
			"object-src":  {"'none'"},
			"frame-src":   {"'none'"},
			"base-uri":    {"'self'"},
			"form-action": {"'self'"},
			"frame-ancestors": {"'none'"}, // Prevent clickjacking
			"upgrade-insecure-requests": {}, // Upgrade HTTP to HTTPS
		},
		CSPReportOnly: false,

		// X-Frame-Options - Prevent clickjacking
		EnableFrameOptions: true,
		FrameOption:        "DENY",

		// X-Content-Type-Options - Prevent MIME sniffing
		EnableContentTypeOptions: true,

		// X-XSS-Protection - Enable XSS filter
		EnableXSSProtection: true,
		XSSProtectionMode:   "1; mode=block",

		// Referrer-Policy - Strict referrer policy
		EnableReferrerPolicy: true,
		ReferrerPolicy:       "strict-origin-when-cross-origin",

		// Permissions-Policy - Disable dangerous features
		EnablePermissionsPolicy: true,
		PermissionsDirectives: map[string]string{
			"geolocation":           "()",
			"microphone":            "()",
			"camera":                "()",
			"payment":               "()",
			"usb":                   "()",
			"magnetometer":          "()",
			"gyroscope":             "()",
			"accelerometer":         "()",
			"ambient-light-sensor":  "()",
			"autoplay":              "()",
			"encrypted-media":       "()",
			"fullscreen":            "(self)",
			"picture-in-picture":    "()",
		},

		// Expect-CT - Require Certificate Transparency
		EnableExpectCT:  true,
		ExpectCTMaxAge:  86400, // 24 hours
		ExpectCTEnforce: true,

		// Cross-Origin policies
		EnableCORP:   true,
		CORPPolicy:   "same-origin",
		EnableCOEP:   true,
		COEPPolicy:   "require-corp",
		EnableCOOP:   true,
		COOPPolicy:   "same-origin",

		// Remove identifying headers
		RemoveServerHeader:    true,
		RemovePoweredByHeader: true,
	}
}

// StrictSecurityHeadersConfig returns very strict security settings
func StrictSecurityHeadersConfig() SecurityHeadersConfig {
	cfg := DefaultSecurityHeadersConfig()

	// Stricter CSP
	cfg.CSPDirectives = map[string][]string{
		"default-src":         {"'none'"},
		"script-src":          {"'self'"},
		"style-src":           {"'self'"},
		"img-src":             {"'self'"},
		"font-src":            {"'self'"},
		"connect-src":         {"'self'"},
		"media-src":           {"'none'"},
		"object-src":          {"'none'"},
		"frame-src":           {"'none'"},
		"base-uri":            {"'self'"},
		"form-action":         {"'self'"},
		"frame-ancestors":     {"'none'"},
		"upgrade-insecure-requests": {},
	}

	// Stricter referrer policy
	cfg.ReferrerPolicy = "no-referrer"

	return cfg
}

// RelaxedSecurityHeadersConfig returns more permissive settings for development
func RelaxedSecurityHeadersConfig() SecurityHeadersConfig {
	cfg := DefaultSecurityHeadersConfig()

	// More permissive CSP for development
	cfg.CSPDirectives = map[string][]string{
		"default-src": {"'self'", "'unsafe-inline'", "'unsafe-eval'"},
		"img-src":     {"'self'", "data:", "https:", "http:"},
		"connect-src": {"'self'", "ws:", "wss:"},
	}

	// Disable some headers for development
	cfg.EnableHSTS = false
	cfg.EnableExpectCT = false

	return cfg
}

// SecurityHeadersMiddleware creates security headers middleware
func SecurityHeadersMiddleware(cfg SecurityHeadersConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Remove identifying headers
		if cfg.RemoveServerHeader {
			c.Header("Server", cfg.CustomServerHeader)
		}
		if cfg.RemovePoweredByHeader {
			c.Header("X-Powered-By", "")
		}

		// HSTS - HTTP Strict Transport Security
		if cfg.EnableHSTS && c.Request.TLS != nil {
			hsts := fmt.Sprintf("max-age=%d", cfg.HSTSMaxAge)
			if cfg.HSTSIncludeSubdomains {
				hsts += "; includeSubDomains"
			}
			if cfg.HSTSPreload {
				hsts += "; preload"
			}
			c.Header("Strict-Transport-Security", hsts)
		}

		// CSP - Content Security Policy
		if cfg.EnableCSP {
			csp := buildCSP(cfg.CSPDirectives)
			if cfg.CSPReportURI != "" {
				csp += fmt.Sprintf("; report-uri %s", cfg.CSPReportURI)
			}
			headerName := "Content-Security-Policy"
			if cfg.CSPReportOnly {
				headerName = "Content-Security-Policy-Report-Only"
			}
			c.Header(headerName, csp)
		}

		// X-Frame-Options - Clickjacking protection
		if cfg.EnableFrameOptions {
			c.Header("X-Frame-Options", cfg.FrameOption)
		}

		// X-Content-Type-Options - MIME sniffing protection
		if cfg.EnableContentTypeOptions {
			c.Header("X-Content-Type-Options", "nosniff")
		}

		// X-XSS-Protection
		if cfg.EnableXSSProtection {
			c.Header("X-XSS-Protection", cfg.XSSProtectionMode)
		}

		// Referrer-Policy
		if cfg.EnableReferrerPolicy {
			c.Header("Referrer-Policy", cfg.ReferrerPolicy)
		}

		// Permissions-Policy (formerly Feature-Policy)
		if cfg.EnablePermissionsPolicy {
			pp := buildPermissionsPolicy(cfg.PermissionsDirectives)
			c.Header("Permissions-Policy", pp)
		}

		// Expect-CT - Certificate Transparency
		if cfg.EnableExpectCT && c.Request.TLS != nil {
			expectCT := fmt.Sprintf("max-age=%d", cfg.ExpectCTMaxAge)
			if cfg.ExpectCTEnforce {
				expectCT += ", enforce"
			}
			if cfg.ExpectCTReportURI != "" {
				expectCT += fmt.Sprintf(", report-uri=\"%s\"", cfg.ExpectCTReportURI)
			}
			c.Header("Expect-CT", expectCT)
		}

		// Cross-Origin-Resource-Policy
		if cfg.EnableCORP {
			c.Header("Cross-Origin-Resource-Policy", cfg.CORPPolicy)
		}

		// Cross-Origin-Embedder-Policy
		if cfg.EnableCOEP {
			c.Header("Cross-Origin-Embedder-Policy", cfg.COEPPolicy)
		}

		// Cross-Origin-Opener-Policy
		if cfg.EnableCOOP {
			c.Header("Cross-Origin-Opener-Policy", cfg.COOPPolicy)
		}

		c.Next()
	}
}

// buildCSP builds a Content-Security-Policy header value
func buildCSP(directives map[string][]string) string {
	var parts []string
	for directive, values := range directives {
		if len(values) == 0 {
			// Directive with no value (like upgrade-insecure-requests)
			parts = append(parts, directive)
		} else {
			// Directive with values
			parts = append(parts, fmt.Sprintf("%s %s", directive, strings.Join(values, " ")))
		}
	}
	return strings.Join(parts, "; ")
}

// buildPermissionsPolicy builds a Permissions-Policy header value
func buildPermissionsPolicy(directives map[string]string) string {
	var parts []string
	for feature, allowlist := range directives {
		parts = append(parts, fmt.Sprintf("%s=%s", feature, allowlist))
	}
	return strings.Join(parts, ", ")
}

// CSPViolationReport represents a CSP violation report
type CSPViolationReport struct {
	DocumentURI        string `json:"document-uri"`
	Referrer           string `json:"referrer"`
	BlockedURI         string `json:"blocked-uri"`
	ViolatedDirective  string `json:"violated-directive"`
	EffectiveDirective string `json:"effective-directive"`
	OriginalPolicy     string `json:"original-policy"`
	Disposition        string `json:"disposition"`
	StatusCode         int    `json:"status-code"`
}

// CSPReportHandler creates a handler for CSP violation reports
func CSPReportHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var report struct {
			CSPReport CSPViolationReport `json:"csp-report"`
		}

		if err := c.ShouldBindJSON(&report); err != nil {
			c.JSON(400, gin.H{"error": "Invalid CSP report"})
			return
		}

		// Log the CSP violation
		LogSecurityEvent("CSP_VIOLATION", c.ClientIP(),
			fmt.Sprintf("Violated directive: %s, Blocked URI: %s",
				report.CSPReport.ViolatedDirective,
				report.CSPReport.BlockedURI))

		c.Status(204) // No Content
	}
}

// SecureRedirect redirects HTTP to HTTPS
func SecureRedirect(trustProxy bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if request is HTTPS
		isHTTPS := c.Request.TLS != nil

		// Check X-Forwarded-Proto header if behind proxy
		if trustProxy {
			proto := c.GetHeader("X-Forwarded-Proto")
			if proto == "https" {
				isHTTPS = true
			}
		}

		// Redirect to HTTPS if not already
		if !isHTTPS {
			host := c.Request.Host
			if host == "" {
				host = "localhost"
			}

			target := fmt.Sprintf("https://%s%s", host, c.Request.RequestURI)
			LogSecurityEvent("HTTP_TO_HTTPS_REDIRECT", c.ClientIP(),
				fmt.Sprintf("Redirecting to %s", target))

			c.Redirect(301, target) // Permanent redirect
			c.Abort()
			return
		}

		c.Next()
	}
}

// TLSVersionMiddleware enforces minimum TLS version
func TLSVersionMiddleware(minVersion uint16) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.TLS != nil {
			if c.Request.TLS.Version < minVersion {
				LogSecurityEvent("TLS_VERSION_TOO_LOW", c.ClientIP(),
					fmt.Sprintf("TLS version %d is too low (min: %d)",
						c.Request.TLS.Version, minVersion))

				c.JSON(400, gin.H{
					"error": "TLS version too low",
				})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// SecurityHeadersChecker checks if security headers are properly set
type SecurityHeadersChecker struct {
	RequiredHeaders map[string]bool
	ForbiddenHeaders []string
}

// DefaultSecurityHeadersChecker returns a checker with recommended headers
func DefaultSecurityHeadersChecker() *SecurityHeadersChecker {
	return &SecurityHeadersChecker{
		RequiredHeaders: map[string]bool{
			"Strict-Transport-Security":    false,
			"Content-Security-Policy":      false,
			"X-Frame-Options":              false,
			"X-Content-Type-Options":       false,
			"X-XSS-Protection":             false,
			"Referrer-Policy":              false,
		},
		ForbiddenHeaders: []string{
			"X-Powered-By",
			"Server",
		},
	}
}

// CheckHeaders checks if security headers are properly set
func (checker *SecurityHeadersChecker) CheckHeaders(headers map[string][]string) map[string]string {
	issues := make(map[string]string)

	// Check required headers
	for header := range checker.RequiredHeaders {
		if _, exists := headers[header]; !exists {
			issues[header] = "Missing required security header"
		}
	}

	// Check forbidden headers
	for _, header := range checker.ForbiddenHeaders {
		if value, exists := headers[header]; exists && len(value) > 0 && value[0] != "" {
			issues[header] = "Forbidden header present (information disclosure)"
		}
	}

	return issues
}

// SecurityHeadersAuditMiddleware logs security headers for audit
func SecurityHeadersAuditMiddleware() gin.HandlerFunc {
	checker := DefaultSecurityHeadersChecker()

	return func(c *gin.Context) {
		c.Next()

		// Check response headers
		issues := checker.CheckHeaders(c.Writer.Header())

		if len(issues) > 0 {
			LogSecurityEvent("SECURITY_HEADERS_ISSUE", c.ClientIP(),
				fmt.Sprintf("Issues found: %v", issues))
		}
	}
}

// GetSecurityHeaders returns all security-related headers
func GetSecurityHeaders(c *gin.Context) map[string]string {
	headers := make(map[string]string)

	securityHeaderNames := []string{
		"Strict-Transport-Security",
		"Content-Security-Policy",
		"Content-Security-Policy-Report-Only",
		"X-Frame-Options",
		"X-Content-Type-Options",
		"X-XSS-Protection",
		"Referrer-Policy",
		"Permissions-Policy",
		"Feature-Policy",
		"Expect-CT",
		"Cross-Origin-Resource-Policy",
		"Cross-Origin-Embedder-Policy",
		"Cross-Origin-Opener-Policy",
	}

	for _, name := range securityHeaderNames {
		if value := c.Writer.Header().Get(name); value != "" {
			headers[name] = value
		}
	}

	return headers
}

// SecurityHeadersInfo provides information about current security headers
type SecurityHeadersInfo struct {
	HSTSEnabled             bool          `json:"hsts_enabled"`
	HSTSMaxAge              time.Duration `json:"hsts_max_age,omitempty"`
	CSPEnabled              bool          `json:"csp_enabled"`
	CSPPolicy               string        `json:"csp_policy,omitempty"`
	FrameOptionsEnabled     bool          `json:"frame_options_enabled"`
	FrameOption             string        `json:"frame_option,omitempty"`
	ContentTypeNoSniff      bool          `json:"content_type_no_sniff"`
	XSSProtectionEnabled    bool          `json:"xss_protection_enabled"`
	ReferrerPolicyEnabled   bool          `json:"referrer_policy_enabled"`
	ReferrerPolicy          string        `json:"referrer_policy,omitempty"`
	PermissionsPolicyEnabled bool         `json:"permissions_policy_enabled"`
	ExpectCTEnabled         bool          `json:"expect_ct_enabled"`
	CORPEnabled             bool          `json:"corp_enabled"`
	COEPEnabled             bool          `json:"coep_enabled"`
	COOPEnabled             bool          `json:"coop_enabled"`
	ServerHeaderRemoved     bool          `json:"server_header_removed"`
	PoweredByHeaderRemoved  bool          `json:"powered_by_header_removed"`
}

// GetSecurityHeadersInfo extracts security headers information
func GetSecurityHeadersInfo(c *gin.Context) *SecurityHeadersInfo {
	info := &SecurityHeadersInfo{}

	// Check HSTS
	if hsts := c.Writer.Header().Get("Strict-Transport-Security"); hsts != "" {
		info.HSTSEnabled = true
		// Parse max-age if needed
	}

	// Check CSP
	if csp := c.Writer.Header().Get("Content-Security-Policy"); csp != "" {
		info.CSPEnabled = true
		info.CSPPolicy = csp
	} else if csp := c.Writer.Header().Get("Content-Security-Policy-Report-Only"); csp != "" {
		info.CSPEnabled = true
		info.CSPPolicy = csp + " (report-only)"
	}

	// Check X-Frame-Options
	if fo := c.Writer.Header().Get("X-Frame-Options"); fo != "" {
		info.FrameOptionsEnabled = true
		info.FrameOption = fo
	}

	// Check X-Content-Type-Options
	if ctno := c.Writer.Header().Get("X-Content-Type-Options"); ctno == "nosniff" {
		info.ContentTypeNoSniff = true
	}

	// Check X-XSS-Protection
	if xss := c.Writer.Header().Get("X-XSS-Protection"); xss != "" {
		info.XSSProtectionEnabled = true
	}

	// Check Referrer-Policy
	if rp := c.Writer.Header().Get("Referrer-Policy"); rp != "" {
		info.ReferrerPolicyEnabled = true
		info.ReferrerPolicy = rp
	}

	// Check Permissions-Policy
	if pp := c.Writer.Header().Get("Permissions-Policy"); pp != "" {
		info.PermissionsPolicyEnabled = true
	}

	// Check Expect-CT
	if ect := c.Writer.Header().Get("Expect-CT"); ect != "" {
		info.ExpectCTEnabled = true
	}

	// Check Cross-Origin policies
	if corp := c.Writer.Header().Get("Cross-Origin-Resource-Policy"); corp != "" {
		info.CORPEnabled = true
	}
	if coep := c.Writer.Header().Get("Cross-Origin-Embedder-Policy"); coep != "" {
		info.COEPEnabled = true
	}
	if coop := c.Writer.Header().Get("Cross-Origin-Opener-Policy"); coop != "" {
		info.COOPEnabled = true
	}

	// Check if identifying headers are removed
	if server := c.Writer.Header().Get("Server"); server == "" {
		info.ServerHeaderRemoved = true
	}
	if poweredBy := c.Writer.Header().Get("X-Powered-By"); poweredBy == "" {
		info.PoweredByHeaderRemoved = true
	}

	return info
}
