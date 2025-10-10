package security

import (
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// TLSConfig contains TLS/SSL enforcement configuration
type TLSConfig struct {
	// TLS version enforcement
	MinTLSVersion uint16 // Minimum TLS version (default: tls.VersionTLS12)
	MaxTLSVersion uint16 // Maximum TLS version (default: tls.VersionTLS13)

	// Cipher suites
	CipherSuites  []uint16 // Allowed cipher suites
	PreferServerCipherSuites bool // Prefer server cipher suites

	// Certificate configuration
	CertFile      string // Path to certificate file
	KeyFile       string // Path to private key file
	ClientCAFile  string // Path to client CA file (for mutual TLS)

	// Client authentication
	ClientAuth    tls.ClientAuthType // Client authentication mode

	// HTTPS enforcement
	EnforceHTTPS  bool // Redirect HTTP to HTTPS
	HTTPSPort     int  // HTTPS port (default: 443)

	// HSTS
	EnableHSTS    bool // Enable HTTP Strict Transport Security
	HSTSMaxAge    int  // HSTS max age in seconds

	// Certificate verification
	InsecureSkipVerify bool // Skip certificate verification (NOT RECOMMENDED)

	// Session tickets
	SessionTicketsDisabled bool // Disable session tickets

	// Renegotiation
	Renegotiation tls.RenegotiationSupport // Renegotiation support
}

// DefaultTLSConfig returns secure default TLS settings
func DefaultTLSConfig() TLSConfig {
	return TLSConfig{
		MinTLSVersion:            tls.VersionTLS12,
		MaxTLSVersion:            tls.VersionTLS13,
		CipherSuites:             getSecureCipherSuites(),
		PreferServerCipherSuites: true,
		ClientAuth:               tls.NoClientCert,
		EnforceHTTPS:             true,
		HTTPSPort:                443,
		EnableHSTS:               true,
		HSTSMaxAge:               31536000, // 1 year
		InsecureSkipVerify:       false,
		SessionTicketsDisabled:   false,
		Renegotiation:            tls.RenegotiateNever,
	}
}

// StrictTLSConfig returns very strict TLS settings
func StrictTLSConfig() TLSConfig {
	cfg := DefaultTLSConfig()
	cfg.MinTLSVersion = tls.VersionTLS13 // TLS 1.3 only
	cfg.CipherSuites = getTLS13CipherSuites()
	cfg.SessionTicketsDisabled = true
	cfg.ClientAuth = tls.RequireAndVerifyClientCert // Mutual TLS
	return cfg
}

// getSecureCipherSuites returns recommended cipher suites for TLS 1.2
func getSecureCipherSuites() []uint16 {
	return []uint16{
		// TLS 1.3 cipher suites (always enabled in TLS 1.3)
		tls.TLS_AES_128_GCM_SHA256,
		tls.TLS_AES_256_GCM_SHA384,
		tls.TLS_CHACHA20_POLY1305_SHA256,

		// TLS 1.2 cipher suites (recommended)
		tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
		tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
		tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
		tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256,
		tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256,
	}
}

// getTLS13CipherSuites returns TLS 1.3 cipher suites
func getTLS13CipherSuites() []uint16 {
	return []uint16{
		tls.TLS_AES_128_GCM_SHA256,
		tls.TLS_AES_256_GCM_SHA384,
		tls.TLS_CHACHA20_POLY1305_SHA256,
	}
}

// GetWeakCipherSuites returns a list of weak cipher suites to avoid
func GetWeakCipherSuites() []uint16 {
	return []uint16{
		tls.TLS_RSA_WITH_RC4_128_SHA,
		tls.TLS_RSA_WITH_3DES_EDE_CBC_SHA,
		tls.TLS_RSA_WITH_AES_128_CBC_SHA,
		tls.TLS_RSA_WITH_AES_256_CBC_SHA,
		tls.TLS_ECDHE_ECDSA_WITH_RC4_128_SHA,
		tls.TLS_ECDHE_RSA_WITH_RC4_128_SHA,
		tls.TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA,
		tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,
		tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
		tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
		tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
	}
}

// CreateTLSConfig creates a *tls.Config from TLSConfig
func CreateTLSConfig(cfg TLSConfig) *tls.Config {
	return &tls.Config{
		MinVersion:               cfg.MinTLSVersion,
		MaxVersion:               cfg.MaxTLSVersion,
		CipherSuites:             cfg.CipherSuites,
		PreferServerCipherSuites: cfg.PreferServerCipherSuites,
		ClientAuth:               cfg.ClientAuth,
		InsecureSkipVerify:       cfg.InsecureSkipVerify,
		SessionTicketsDisabled:   cfg.SessionTicketsDisabled,
		Renegotiation:            cfg.Renegotiation,
	}
}

// TLSEnforcementMiddleware enforces TLS/SSL requirements
func TLSEnforcementMiddleware(cfg TLSConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if connection is TLS
		if c.Request.TLS == nil {
			// Not using TLS
			if cfg.EnforceHTTPS {
				// Redirect to HTTPS
				host := c.Request.Host
				if cfg.HTTPSPort != 443 {
					host = fmt.Sprintf("%s:%d", c.Request.Host, cfg.HTTPSPort)
				}

				target := fmt.Sprintf("https://%s%s", host, c.Request.RequestURI)
				LogSecurityEvent("HTTP_TO_HTTPS_REDIRECT", c.ClientIP(),
					fmt.Sprintf("Redirecting to %s", target))

				c.Redirect(http.StatusMovedPermanently, target)
				c.Abort()
				return
			} else {
				// Just log warning
				LogSecurityEvent("INSECURE_CONNECTION", c.ClientIP(),
					"Request received over HTTP instead of HTTPS")
			}
		} else {
			// Using TLS, verify version and cipher suite
			if c.Request.TLS.Version < cfg.MinTLSVersion {
				LogSecurityEvent("TLS_VERSION_TOO_LOW", c.ClientIP(),
					fmt.Sprintf("TLS version %d is below minimum %d",
						c.Request.TLS.Version, cfg.MinTLSVersion))

				c.JSON(http.StatusBadRequest, gin.H{
					"error": "TLS version too low",
				})
				c.Abort()
				return
			}

			if cfg.MaxTLSVersion > 0 && c.Request.TLS.Version > cfg.MaxTLSVersion {
				LogSecurityEvent("TLS_VERSION_TOO_HIGH", c.ClientIP(),
					fmt.Sprintf("TLS version %d is above maximum %d",
						c.Request.TLS.Version, cfg.MaxTLSVersion))

				c.JSON(http.StatusBadRequest, gin.H{
					"error": "TLS version not supported",
				})
				c.Abort()
				return
			}

			// Check cipher suite
			if len(cfg.CipherSuites) > 0 {
				cipherAllowed := false
				for _, allowed := range cfg.CipherSuites {
					if c.Request.TLS.CipherSuite == allowed {
						cipherAllowed = true
						break
					}
				}

				if !cipherAllowed {
					LogSecurityEvent("WEAK_CIPHER_SUITE", c.ClientIP(),
						fmt.Sprintf("Cipher suite 0x%04X not allowed",
							c.Request.TLS.CipherSuite))

					c.JSON(http.StatusBadRequest, gin.H{
						"error": "Cipher suite not allowed",
					})
					c.Abort()
					return
				}
			}

			// Add HSTS header if enabled
			if cfg.EnableHSTS {
				hsts := fmt.Sprintf("max-age=%d; includeSubDomains; preload", cfg.HSTSMaxAge)
				c.Header("Strict-Transport-Security", hsts)
			}
		}

		c.Next()
	}
}

// MutualTLSMiddleware enforces mutual TLS (client certificate authentication)
func MutualTLSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.TLS == nil {
			LogSecurityEvent("MTLS_NO_TLS", c.ClientIP(), "Mutual TLS required but no TLS connection")
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "TLS required",
			})
			c.Abort()
			return
		}

		if len(c.Request.TLS.PeerCertificates) == 0 {
			LogSecurityEvent("MTLS_NO_CLIENT_CERT", c.ClientIP(),
				"Mutual TLS required but no client certificate provided")

			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Client certificate required",
			})
			c.Abort()
			return
		}

		// Get client certificate
		clientCert := c.Request.TLS.PeerCertificates[0]

		// Verify certificate is not expired
		if clientCert.NotAfter.Before(c.Request.TLS.PeerCertificates[0].NotBefore) {
			LogSecurityEvent("MTLS_CERT_EXPIRED", c.ClientIP(),
				fmt.Sprintf("Client certificate expired: %s", clientCert.Subject.CommonName))

			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Client certificate expired",
			})
			c.Abort()
			return
		}

		// Store client certificate info in context
		c.Set("client_cert_subject", clientCert.Subject.CommonName)
		c.Set("client_cert_issuer", clientCert.Issuer.CommonName)
		c.Set("client_cert", clientCert)

		LogSecurityEvent("MTLS_SUCCESS", c.ClientIP(),
			fmt.Sprintf("Mutual TLS authentication successful for %s", clientCert.Subject.CommonName))

		c.Next()
	}
}

// TLSVersionInfo contains information about TLS connection
type TLSVersionInfo struct {
	Version              string   `json:"version"`
	VersionNumber        uint16   `json:"version_number"`
	CipherSuite          string   `json:"cipher_suite"`
	CipherSuiteNumber    uint16   `json:"cipher_suite_number"`
	ServerName           string   `json:"server_name"`
	NegotiatedProtocol   string   `json:"negotiated_protocol"`
	HandshakeComplete    bool     `json:"handshake_complete"`
	ClientCertPresent    bool     `json:"client_cert_present"`
	ClientCertSubject    string   `json:"client_cert_subject,omitempty"`
}

// GetTLSVersionInfo extracts TLS connection information
func GetTLSVersionInfo(c *gin.Context) *TLSVersionInfo {
	if c.Request.TLS == nil {
		return nil
	}

	info := &TLSVersionInfo{
		Version:            getTLSVersionName(c.Request.TLS.Version),
		VersionNumber:      c.Request.TLS.Version,
		CipherSuite:        getCipherSuiteName(c.Request.TLS.CipherSuite),
		CipherSuiteNumber:  c.Request.TLS.CipherSuite,
		ServerName:         c.Request.TLS.ServerName,
		NegotiatedProtocol: c.Request.TLS.NegotiatedProtocol,
		HandshakeComplete:  c.Request.TLS.HandshakeComplete,
		ClientCertPresent:  len(c.Request.TLS.PeerCertificates) > 0,
	}

	if info.ClientCertPresent {
		info.ClientCertSubject = c.Request.TLS.PeerCertificates[0].Subject.CommonName
	}

	return info
}

// getTLSVersionName returns the name of a TLS version
func getTLSVersionName(version uint16) string {
	switch version {
	case tls.VersionTLS10:
		return "TLS 1.0"
	case tls.VersionTLS11:
		return "TLS 1.1"
	case tls.VersionTLS12:
		return "TLS 1.2"
	case tls.VersionTLS13:
		return "TLS 1.3"
	default:
		return fmt.Sprintf("Unknown (0x%04X)", version)
	}
}

// getCipherSuiteName returns the name of a cipher suite
func getCipherSuiteName(suite uint16) string {
	// TLS 1.3 cipher suites
	switch suite {
	case tls.TLS_AES_128_GCM_SHA256:
		return "TLS_AES_128_GCM_SHA256"
	case tls.TLS_AES_256_GCM_SHA384:
		return "TLS_AES_256_GCM_SHA384"
	case tls.TLS_CHACHA20_POLY1305_SHA256:
		return "TLS_CHACHA20_POLY1305_SHA256"

	// TLS 1.2 cipher suites
	case tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256:
		return "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256"
	case tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384:
		return "TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384"
	case tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256:
		return "TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256"
	case tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384:
		return "TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384"
	case tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256:
		return "TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256"
	case tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256:
		return "TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256"

	// Weak cipher suites
	case tls.TLS_RSA_WITH_RC4_128_SHA:
		return "TLS_RSA_WITH_RC4_128_SHA (WEAK)"
	case tls.TLS_RSA_WITH_3DES_EDE_CBC_SHA:
		return "TLS_RSA_WITH_3DES_EDE_CBC_SHA (WEAK)"
	case tls.TLS_RSA_WITH_AES_128_CBC_SHA:
		return "TLS_RSA_WITH_AES_128_CBC_SHA (WEAK)"
	case tls.TLS_RSA_WITH_AES_256_CBC_SHA:
		return "TLS_RSA_WITH_AES_256_CBC_SHA (WEAK)"

	default:
		return fmt.Sprintf("Unknown (0x%04X)", suite)
	}
}

// ValidateTLSConfig validates a TLS configuration
func ValidateTLSConfig(cfg TLSConfig) []string {
	var issues []string

	// Check TLS version
	if cfg.MinTLSVersion < tls.VersionTLS12 {
		issues = append(issues, "Minimum TLS version is below TLS 1.2 (insecure)")
	}

	// Check if HTTPS is enforced
	if !cfg.EnforceHTTPS {
		issues = append(issues, "HTTPS enforcement is disabled")
	}

	// Check HSTS
	if !cfg.EnableHSTS {
		issues = append(issues, "HSTS is disabled")
	} else if cfg.HSTSMaxAge < 31536000 {
		issues = append(issues, "HSTS max-age is less than 1 year")
	}

	// Check certificate verification
	if cfg.InsecureSkipVerify {
		issues = append(issues, "Certificate verification is disabled (CRITICAL SECURITY ISSUE)")
	}

	// Check for weak cipher suites
	weakCiphers := GetWeakCipherSuites()
	for _, configured := range cfg.CipherSuites {
		for _, weak := range weakCiphers {
			if configured == weak {
				issues = append(issues, fmt.Sprintf("Weak cipher suite enabled: %s",
					getCipherSuiteName(configured)))
			}
		}
	}

	return issues
}

// TLSAuditMiddleware logs TLS connection details for audit
func TLSAuditMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.TLS != nil {
			info := GetTLSVersionInfo(c)
			LogSecurityEvent("TLS_CONNECTION", c.ClientIP(),
				fmt.Sprintf("TLS %s with %s", info.Version, info.CipherSuite))
		} else {
			LogSecurityEvent("INSECURE_HTTP_CONNECTION", c.ClientIP(),
				"Request received over insecure HTTP")
		}

		c.Next()
	}
}

// RequireHTTPSMiddleware strictly requires HTTPS (aborts non-HTTPS requests)
func RequireHTTPSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.TLS == nil {
			LogSecurityEvent("HTTPS_REQUIRED", c.ClientIP(),
				"HTTPS required but request was HTTP")

			c.JSON(http.StatusBadRequest, gin.H{
				"error": "HTTPS required",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// TLSStatistics contains TLS usage statistics
type TLSStatistics struct {
	TotalTLSConnections    int            `json:"total_tls_connections"`
	TotalHTTPConnections   int            `json:"total_http_connections"`
	TLSVersions            map[string]int `json:"tls_versions"`
	CipherSuites           map[string]int `json:"cipher_suites"`
	ClientCertConnections  int            `json:"client_cert_connections"`
	WeakCipherDetections   int            `json:"weak_cipher_detections"`
}

// Global TLS statistics (would need to be implemented with proper tracking)
var tlsStats = &TLSStatistics{
	TLSVersions:  make(map[string]int),
	CipherSuites: make(map[string]int),
}

// GetTLSStatistics returns TLS statistics
func GetTLSStatistics() *TLSStatistics {
	return tlsStats
}
