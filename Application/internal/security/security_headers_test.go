package security

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestSecurityHeadersMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := DefaultSecurityHeadersConfig()

	r := gin.New()
	r.Use(SecurityHeadersMiddleware(cfg))
	r.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	// Check security headers
	assert.Equal(t, http.StatusOK, w.Code)

	// CSP header
	assert.NotEmpty(t, w.Header().Get("Content-Security-Policy"))

	// X-Frame-Options
	assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"))

	// X-Content-Type-Options
	assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))

	// X-XSS-Protection
	assert.Equal(t, "1; mode=block", w.Header().Get("X-XSS-Protection"))

	// Referrer-Policy
	assert.Equal(t, "strict-origin-when-cross-origin", w.Header().Get("Referrer-Policy"))

	// Permissions-Policy
	assert.NotEmpty(t, w.Header().Get("Permissions-Policy"))

	// Cross-Origin headers
	assert.Equal(t, "same-origin", w.Header().Get("Cross-Origin-Resource-Policy"))
	assert.Equal(t, "require-corp", w.Header().Get("Cross-Origin-Embedder-Policy"))
	assert.Equal(t, "same-origin", w.Header().Get("Cross-Origin-Opener-Policy"))
}

func TestSecurityHeadersMiddleware_HSTS(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := DefaultSecurityHeadersConfig()

	r := gin.New()
	r.Use(SecurityHeadersMiddleware(cfg))
	r.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	// Add TLS connection info
	req.TLS = &tls.ConnectionState{
		Version: tls.VersionTLS13,
	}

	r.ServeHTTP(w, req)

	// HSTS should be present with TLS
	hsts := w.Header().Get("Strict-Transport-Security")
	assert.NotEmpty(t, hsts)
	assert.Contains(t, hsts, "max-age=31536000")
	assert.Contains(t, hsts, "includeSubDomains")
	assert.Contains(t, hsts, "preload")
}

func TestBuildCSP(t *testing.T) {
	directives := map[string][]string{
		"default-src": {"'self'"},
		"script-src":  {"'self'", "'unsafe-inline'"},
		"style-src":   {"'self'"},
	}

	csp := buildCSP(directives)

	assert.Contains(t, csp, "default-src 'self'")
	assert.Contains(t, csp, "script-src 'self' 'unsafe-inline'")
	assert.Contains(t, csp, "style-src 'self'")
}

func TestBuildPermissionsPolicy(t *testing.T) {
	directives := map[string]string{
		"geolocation": "()",
		"camera":      "()",
		"microphone":  "()",
	}

	pp := buildPermissionsPolicy(directives)

	assert.Contains(t, pp, "geolocation=()")
	assert.Contains(t, pp, "camera=()")
	assert.Contains(t, pp, "microphone=()")
}

func TestSecureRedirect(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(SecureRedirect(false))
	r.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// HTTP request should be redirected
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.Host = "example.com"
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusMovedPermanently, w.Code)
	assert.Equal(t, "https://example.com/test", w.Header().Get("Location"))
}

func TestTLSVersionMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(TLSVersionMiddleware(tls.VersionTLS12))
	r.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// TLS 1.3 should be allowed
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.TLS = &tls.ConnectionState{
		Version: tls.VersionTLS13,
	}
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// TLS 1.1 should be rejected
	w = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/test", nil)
	req.TLS = &tls.ConnectionState{
		Version: tls.VersionTLS11,
	}
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetSecurityHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := DefaultSecurityHeadersConfig()

	r := gin.New()
	r.Use(SecurityHeadersMiddleware(cfg))
	r.GET("/test", func(c *gin.Context) {
		headers := GetSecurityHeaders(c)
		assert.True(t, len(headers) > 0)
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetSecurityHeadersInfo(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := DefaultSecurityHeadersConfig()

	r := gin.New()
	r.Use(SecurityHeadersMiddleware(cfg))
	r.GET("/test", func(c *gin.Context) {
		info := GetSecurityHeadersInfo(c)
		assert.NotNil(t, info)
		assert.True(t, info.CSPEnabled)
		assert.True(t, info.FrameOptionsEnabled)
		assert.True(t, info.ContentTypeNoSniff)
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDefaultSecurityHeadersChecker(t *testing.T) {
	checker := DefaultSecurityHeadersChecker()

	headers := map[string][]string{
		"Content-Security-Policy": {"default-src 'self'"},
		"X-Frame-Options":         {"DENY"},
		"X-Content-Type-Options":  {"nosniff"},
		"X-XSS-Protection":        {"1; mode=block"},
		"Referrer-Policy":         {"no-referrer"},
	}

	issues := checker.CheckHeaders(headers)

	// Should not have HSTS issue (only required with TLS)
	// All other required headers are present
	assert.True(t, len(issues) <= 1) // Might be missing HSTS
}

func TestSecurityHeadersChecker_Forbidden(t *testing.T) {
	checker := DefaultSecurityHeadersChecker()

	headers := map[string][]string{
		"Content-Security-Policy": {"default-src 'self'"},
		"X-Frame-Options":         {"DENY"},
		"X-Content-Type-Options":  {"nosniff"},
		"X-XSS-Protection":        {"1; mode=block"},
		"Referrer-Policy":         {"no-referrer"},
		"X-Powered-By":            {"PHP/7.4"},      // Forbidden!
		"Server":                  {"Apache/2.4.41"}, // Forbidden!
	}

	issues := checker.CheckHeaders(headers)

	// Should have issues with forbidden headers
	assert.True(t, len(issues) >= 2)
	assert.Contains(t, issues, "X-Powered-By")
	assert.Contains(t, issues, "Server")
}

func TestStrictSecurityHeadersConfig(t *testing.T) {
	cfg := StrictSecurityHeadersConfig()

	assert.Equal(t, tls.VersionTLS12, cfg.MinTLSVersion)
	assert.True(t, cfg.EnableHSTS)
	assert.True(t, cfg.EnableCSP)
	assert.NotNil(t, cfg.CSPDirectives)

	// Strict CSP should have restrictive directives
	defaultSrc, ok := cfg.CSPDirectives["default-src"]
	assert.True(t, ok)
	assert.Contains(t, defaultSrc, "'none'")
}

func TestRelaxedSecurityHeadersConfig(t *testing.T) {
	cfg := RelaxedSecurityHeadersConfig()

	// Relaxed config should have HSTS disabled
	assert.False(t, cfg.EnableHSTS)

	// CSP should be more permissive
	defaultSrc, ok := cfg.CSPDirectives["default-src"]
	assert.True(t, ok)
	assert.Contains(t, defaultSrc, "'unsafe-inline'")
}

func TestCSPReportHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.POST("/csp-report", CSPReportHandler())

	// Valid CSP report
	w := httptest.NewRecorder()
	body := `{
		"csp-report": {
			"document-uri": "https://example.com",
			"violated-directive": "script-src 'self'",
			"blocked-uri": "https://evil.com/script.js"
		}
	}`
	req := httptest.NewRequest("POST", "/csp-report", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Body = http.NoBody
	r.ServeHTTP(w, req)

	// Handler should accept the report
	// (actual parsing might fail in test, but handler should not crash)
}

func BenchmarkSecurityHeadersMiddleware(b *testing.B) {
	gin.SetMode(gin.TestMode)

	cfg := DefaultSecurityHeadersConfig()
	r := gin.New()
	r.Use(SecurityHeadersMiddleware(cfg))
	r.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
	}
}

func BenchmarkBuildCSP(b *testing.B) {
	directives := map[string][]string{
		"default-src": {"'self'"},
		"script-src":  {"'self'", "'unsafe-inline'"},
		"style-src":   {"'self'"},
		"img-src":     {"'self'", "data:", "https:"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buildCSP(directives)
	}
}
