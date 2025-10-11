package security

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestDefaultTLSConfig(t *testing.T) {
	cfg := DefaultTLSConfig()

	assert.Equal(t, uint16(tls.VersionTLS12), cfg.MinTLSVersion)
	assert.Equal(t, uint16(tls.VersionTLS13), cfg.MaxTLSVersion)
	assert.True(t, cfg.EnforceHTTPS)
	assert.True(t, cfg.EnableHSTS)
	assert.Equal(t, 31536000, cfg.HSTSMaxAge)
	assert.False(t, cfg.InsecureSkipVerify)
}

func TestStrictTLSConfig(t *testing.T) {
	cfg := StrictTLSConfig()

	assert.Equal(t, uint16(tls.VersionTLS13), cfg.MinTLSVersion)
	assert.True(t, cfg.SessionTicketsDisabled)
	assert.Equal(t, tls.RequireAndVerifyClientCert, cfg.ClientAuth)
}

func TestGetSecureCipherSuites(t *testing.T) {
	suites := getSecureCipherSuites()

	assert.True(t, len(suites) > 0)

	// Should include TLS 1.3 suites
	assert.Contains(t, suites, tls.TLS_AES_128_GCM_SHA256)
	assert.Contains(t, suites, tls.TLS_AES_256_GCM_SHA384)
	assert.Contains(t, suites, tls.TLS_CHACHA20_POLY1305_SHA256)

	// Should include secure TLS 1.2 suites
	assert.Contains(t, suites, tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256)
	assert.Contains(t, suites, tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384)
}

func TestGetWeakCipherSuites(t *testing.T) {
	weakSuites := GetWeakCipherSuites()

	assert.True(t, len(weakSuites) > 0)

	// Should include known weak suites
	assert.Contains(t, weakSuites, tls.TLS_RSA_WITH_RC4_128_SHA)
	assert.Contains(t, weakSuites, tls.TLS_RSA_WITH_3DES_EDE_CBC_SHA)
}

func TestCreateTLSConfig(t *testing.T) {
	cfg := DefaultTLSConfig()
	tlsConfig := CreateTLSConfig(cfg)

	assert.NotNil(t, tlsConfig)
	assert.Equal(t, cfg.MinTLSVersion, tlsConfig.MinVersion)
	assert.Equal(t, cfg.MaxTLSVersion, tlsConfig.MaxVersion)
	assert.Equal(t, cfg.PreferServerCipherSuites, tlsConfig.PreferServerCipherSuites)
}

func TestTLSEnforcementMiddleware_HTTPS(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := DefaultTLSConfig()

	r := gin.New()
	r.Use(TLSEnforcementMiddleware(cfg))
	r.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// HTTPS request should succeed
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.TLS = &tls.ConnectionState{
		Version:     tls.VersionTLS13,
		CipherSuite: tls.TLS_AES_128_GCM_SHA256,
	}
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Should have HSTS header
	hsts := w.Header().Get("Strict-Transport-Security")
	assert.NotEmpty(t, hsts)
	assert.Contains(t, hsts, "max-age=31536000")
}

func TestTLSEnforcementMiddleware_HTTP_Redirect(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := DefaultTLSConfig()
	cfg.EnforceHTTPS = true

	r := gin.New()
	r.Use(TLSEnforcementMiddleware(cfg))
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

func TestTLSEnforcementMiddleware_LowVersion(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := DefaultTLSConfig()
	cfg.MinTLSVersion = tls.VersionTLS12

	r := gin.New()
	r.Use(TLSEnforcementMiddleware(cfg))
	r.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// TLS 1.1 should be rejected
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.TLS = &tls.ConnectionState{
		Version:     tls.VersionTLS11,
		CipherSuite: tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
	}
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTLSEnforcementMiddleware_WeakCipher(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := DefaultTLSConfig()
	cfg.CipherSuites = getSecureCipherSuites()

	r := gin.New()
	r.Use(TLSEnforcementMiddleware(cfg))
	r.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// Weak cipher should be rejected
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.TLS = &tls.ConnectionState{
		Version:     tls.VersionTLS12,
		CipherSuite: tls.TLS_RSA_WITH_RC4_128_SHA, // Weak cipher
	}
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestMutualTLSMiddleware_NoCert(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(MutualTLSMiddleware())
	r.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// No client certificate should be rejected
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.TLS = &tls.ConnectionState{
		Version:          tls.VersionTLS13,
		PeerCertificates: nil, // No client cert
	}
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetTLSVersionInfo(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest("GET", "/test", nil)
	req.TLS = &tls.ConnectionState{
		Version:            tls.VersionTLS13,
		CipherSuite:        tls.TLS_AES_128_GCM_SHA256,
		ServerName:         "example.com",
		HandshakeComplete:  true,
		NegotiatedProtocol: "h2",
	}
	c.Request = req

	info := GetTLSVersionInfo(c)

	assert.NotNil(t, info)
	assert.Equal(t, "TLS 1.3", info.Version)
	assert.Equal(t, uint16(tls.VersionTLS13), info.VersionNumber)
	assert.Equal(t, "TLS_AES_128_GCM_SHA256", info.CipherSuite)
	assert.Equal(t, "example.com", info.ServerName)
	assert.Equal(t, "h2", info.NegotiatedProtocol)
	assert.True(t, info.HandshakeComplete)
}

func TestGetTLSVersionInfo_NoTLS(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest("GET", "/test", nil)
	c.Request = req

	info := GetTLSVersionInfo(c)

	assert.Nil(t, info)
}

func TestGetTLSVersionName(t *testing.T) {
	tests := []struct {
		version  uint16
		expected string
	}{
		{tls.VersionTLS10, "TLS 1.0"},
		{tls.VersionTLS11, "TLS 1.1"},
		{tls.VersionTLS12, "TLS 1.2"},
		{tls.VersionTLS13, "TLS 1.3"},
		{0x9999, "Unknown (0x9999)"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			name := getTLSVersionName(tt.version)
			assert.Equal(t, tt.expected, name)
		})
	}
}

func TestGetCipherSuiteName(t *testing.T) {
	tests := []struct {
		suite    uint16
		expected string
	}{
		{tls.TLS_AES_128_GCM_SHA256, "TLS_AES_128_GCM_SHA256"},
		{tls.TLS_AES_256_GCM_SHA384, "TLS_AES_256_GCM_SHA384"},
		{tls.TLS_CHACHA20_POLY1305_SHA256, "TLS_CHACHA20_POLY1305_SHA256"},
		{tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256, "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256"},
		{tls.TLS_RSA_WITH_RC4_128_SHA, "TLS_RSA_WITH_RC4_128_SHA (WEAK)"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			name := getCipherSuiteName(tt.suite)
			assert.Equal(t, tt.expected, name)
		})
	}
}

func TestValidateTLSConfig(t *testing.T) {
	// Valid config
	cfg := DefaultTLSConfig()
	issues := ValidateTLSConfig(cfg)
	assert.Equal(t, 0, len(issues))

	// Insecure config
	insecureCfg := DefaultTLSConfig()
	insecureCfg.MinTLSVersion = tls.VersionTLS10
	insecureCfg.InsecureSkipVerify = true
	insecureCfg.EnableHSTS = false

	issues = ValidateTLSConfig(insecureCfg)
	assert.True(t, len(issues) > 0)
}

func TestRequireHTTPSMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(RequireHTTPSMiddleware())
	r.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// HTTPS should succeed
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.TLS = &tls.ConnectionState{
		Version: tls.VersionTLS13,
	}
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// HTTP should be rejected
	w = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTLSAuditMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Clear audit log
	ClearAuditLog()

	r := gin.New()
	r.Use(TLSAuditMiddleware())
	r.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// HTTPS request
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.TLS = &tls.ConnectionState{
		Version:     tls.VersionTLS13,
		CipherSuite: tls.TLS_AES_128_GCM_SHA256,
	}
	req.RemoteAddr = "192.168.1.1:12345"
	r.ServeHTTP(w, req)

	// Check audit log
	events := GetRecentEvents(10)
	assert.True(t, len(events) > 0)
	assert.Equal(t, "TLS_CONNECTION", events[0].EventType)
}

func BenchmarkTLSEnforcementMiddleware(b *testing.B) {
	gin.SetMode(gin.TestMode)

	cfg := DefaultTLSConfig()
	r := gin.New()
	r.Use(TLSEnforcementMiddleware(cfg))
	r.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.TLS = &tls.ConnectionState{
		Version:     tls.VersionTLS13,
		CipherSuite: tls.TLS_AES_128_GCM_SHA256,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
	}
}

func BenchmarkGetTLSVersionInfo(b *testing.B) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest("GET", "/test", nil)
	req.TLS = &tls.ConnectionState{
		Version:     tls.VersionTLS13,
		CipherSuite: tls.TLS_AES_128_GCM_SHA256,
	}
	c.Request = req

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetTLSVersionInfo(c)
	}
}
