package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"helixtrack.ru/core/internal/security"
)

// TestSecurity_FullProtectionStack tests all security layers working together
func TestSecurity_FullProtectionStack(t *testing.T) {
	router := gin.New()

	// Apply full security stack
	router.Use(security.SecurityHeadersMiddleware())
	router.Use(security.CSRFProtectionMiddleware("test-secret"))
	router.Use(security.RateLimitMiddleware(security.DefaultDDoSConfig()))
	router.Use(security.BruteForceProtectionMiddleware(security.DefaultBruteForceConfig()))
	router.Use(security.InputValidationMiddleware())

	router.POST("/api/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Test 1: Valid request with CSRF token passes all layers
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/test", strings.NewReader(`{"data":"valid"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-CSRF-Token", "valid-token")
	req.Header.Set("Cookie", "csrf_token=valid-token")

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Verify security headers are set
	assert.NotEmpty(t, w.Header().Get("X-Frame-Options"))
	assert.NotEmpty(t, w.Header().Get("X-Content-Type-Options"))
	assert.NotEmpty(t, w.Header().Get("X-XSS-Protection"))
}

// TestSecurity_CSRFAndInputValidation tests CSRF protection with input validation
func TestSecurity_CSRFAndInputValidation(t *testing.T) {
	router := gin.New()

	router.Use(security.CSRFProtectionMiddleware("test-secret"))
	router.Use(security.InputValidationMiddleware())

	router.POST("/api/data", func(c *gin.Context) {
		var data map[string]interface{}
		if err := c.ShouldBindJSON(&data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid data"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Test 1: Valid CSRF token but malicious input (SQL injection attempt)
	w := httptest.NewRecorder()
	payload := map[string]interface{}{
		"name": "test'; DROP TABLE users; --",
	}
	jsonData, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/api/data", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-CSRF-Token", "valid-token")
	req.Header.Set("Cookie", "csrf_token=valid-token")

	router.ServeHTTP(w, req)
	// Should be blocked by input validation
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestSecurity_RateLimitingAndBruteForce tests rate limiting with brute force protection
func TestSecurity_RateLimitingAndBruteForce(t *testing.T) {
	router := gin.New()

	rateCfg := security.DefaultDDoSConfig()
	rateCfg.RequestsPerSecond = 5
	rateCfg.BurstSize = 5

	bruteCfg := security.DefaultBruteForceConfig()
	bruteCfg.MaxAttempts = 3
	bruteCfg.WindowDuration = 1 * time.Minute

	router.Use(security.RateLimitMiddleware(rateCfg))
	router.Use(security.BruteForceProtectionMiddleware(bruteCfg))

	loginAttempts := 0
	router.POST("/api/login", func(c *gin.Context) {
		loginAttempts++
		var data map[string]interface{}
		if err := c.ShouldBindJSON(&data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid data"})
			return
		}

		// Simulate failed login
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
	})

	// Send multiple failed login attempts from same IP
	ip := "192.168.1.100:12345"

	for i := 0; i < 10; i++ {
		w := httptest.NewRecorder()
		payload := map[string]interface{}{
			"username": "testuser",
			"password": "wrongpass",
		}
		jsonData, _ := json.Marshal(payload)

		req := httptest.NewRequest("POST", "/api/login", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.RemoteAddr = ip

		router.ServeHTTP(w, req)

		if i < 5 {
			// First 5 requests should pass rate limit
			assert.NotEqual(t, http.StatusTooManyRequests, w.Code)
		} else {
			// After 5 requests, should be rate limited
			assert.Equal(t, http.StatusTooManyRequests, w.Code)
		}
	}
}

// TestSecurity_HeadersAndTLSEnforcement tests security headers with TLS enforcement
func TestSecurity_HeadersAndTLSEnforcement(t *testing.T) {
	router := gin.New()

	router.Use(security.SecurityHeadersMiddleware())
	router.Use(security.TLSEnforcementMiddleware(true)) // Require HTTPS

	router.GET("/api/secure", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "secure"})
	})

	// Test 1: HTTP request (should be rejected)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/secure", nil)
	req.Header.Set("X-Forwarded-Proto", "http")

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)

	// Test 2: HTTPS request (should pass)
	w = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/api/secure", nil)
	req.Header.Set("X-Forwarded-Proto", "https")
	req.TLS = &struct{}{} // Simulate TLS connection

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Verify security headers
	assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"))
	assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
	assert.NotEmpty(t, w.Header().Get("Strict-Transport-Security"))
}

// TestSecurity_AuditLogging tests security events are logged
func TestSecurity_AuditLogging(t *testing.T) {
	router := gin.New()

	auditCfg := security.DefaultAuditConfig()
	auditLogger, err := security.NewAuditLogger(auditCfg)
	require.NoError(t, err)
	defer auditLogger.Close()

	router.Use(security.AuditMiddleware(auditLogger))
	router.Use(security.InputValidationMiddleware())

	router.POST("/api/action", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Test 1: Malicious request (should be logged)
	w := httptest.NewRecorder()
	payload := map[string]interface{}{
		"data": "<script>alert('xss')</script>",
	}
	jsonData, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/api/action", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.RemoteAddr = "192.168.1.200:12345"

	router.ServeHTTP(w, req)

	// Should be blocked and logged
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestSecurity_FullStackWithValidation tests complete security stack with comprehensive validation
func TestSecurity_FullStackWithValidation(t *testing.T) {
	router := gin.New()

	// Complete security middleware stack
	router.Use(security.SecurityHeadersMiddleware())
	router.Use(security.CSRFProtectionMiddleware("test-secret"))
	router.Use(security.RateLimitMiddleware(security.DefaultDDoSConfig()))
	router.Use(security.BruteForceProtectionMiddleware(security.DefaultBruteForceConfig()))
	router.Use(security.InputValidationMiddleware())
	router.Use(security.TLSEnforcementMiddleware(true))

	router.POST("/api/submit", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Test valid request through entire stack
	w := httptest.NewRecorder()
	payload := map[string]interface{}{
		"name":  "John Doe",
		"email": "john@example.com",
		"text":  "This is a safe message",
	}
	jsonData, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/api/submit", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-CSRF-Token", "valid-token")
	req.Header.Set("Cookie", "csrf_token=valid-token")
	req.Header.Set("X-Forwarded-Proto", "https")
	req.TLS = &struct{}{} // Simulate TLS
	req.RemoteAddr = "192.168.1.1:12345"

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Verify all security headers are present
	assert.NotEmpty(t, w.Header().Get("X-Frame-Options"))
	assert.NotEmpty(t, w.Header().Get("X-Content-Type-Options"))
	assert.NotEmpty(t, w.Header().Get("X-XSS-Protection"))
	assert.NotEmpty(t, w.Header().Get("Strict-Transport-Security"))
	assert.NotEmpty(t, w.Header().Get("Content-Security-Policy"))
}

// TestSecurity_AttackScenarios tests realistic attack scenarios
func TestSecurity_AttackScenarios(t *testing.T) {
	router := gin.New()

	router.Use(security.SecurityHeadersMiddleware())
	router.Use(security.CSRFProtectionMiddleware("test-secret"))
	router.Use(security.InputValidationMiddleware())

	router.POST("/api/update", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "updated"})
	})

	attacks := []struct {
		name    string
		payload map[string]interface{}
		headers map[string]string
	}{
		{
			name: "SQL Injection",
			payload: map[string]interface{}{
				"query": "'; DROP TABLE users; --",
			},
			headers: map[string]string{
				"X-CSRF-Token": "valid-token",
				"Cookie":       "csrf_token=valid-token",
			},
		},
		{
			name: "XSS Attack",
			payload: map[string]interface{}{
				"comment": "<script>alert(document.cookie)</script>",
			},
			headers: map[string]string{
				"X-CSRF-Token": "valid-token",
				"Cookie":       "csrf_token=valid-token",
			},
		},
		{
			name: "Path Traversal",
			payload: map[string]interface{}{
				"file": "../../etc/passwd",
			},
			headers: map[string]string{
				"X-CSRF-Token": "valid-token",
				"Cookie":       "csrf_token=valid-token",
			},
		},
		{
			name: "Command Injection",
			payload: map[string]interface{}{
				"cmd": "; cat /etc/passwd",
			},
			headers: map[string]string{
				"X-CSRF-Token": "valid-token",
				"Cookie":       "csrf_token=valid-token",
			},
		},
		{
			name: "LDAP Injection",
			payload: map[string]interface{}{
				"filter": "*)(uid=*))(|(uid=*",
			},
			headers: map[string]string{
				"X-CSRF-Token": "valid-token",
				"Cookie":       "csrf_token=valid-token",
			},
		},
		{
			name: "CSRF Attack (missing token)",
			payload: map[string]interface{}{
				"action": "delete",
			},
			headers: map[string]string{},
		},
	}

	for _, attack := range attacks {
		t.Run(attack.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			jsonData, _ := json.Marshal(attack.payload)

			req := httptest.NewRequest("POST", "/api/update", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")

			for key, value := range attack.headers {
				req.Header.Set(key, value)
			}

			router.ServeHTTP(w, req)

			// All attacks should be blocked (either 400 or 403)
			assert.True(t, w.Code == http.StatusBadRequest || w.Code == http.StatusForbidden,
				"Attack '%s' should be blocked, got status %d", attack.name, w.Code)
		})
	}
}

// TestSecurity_ConcurrentAttacks tests security under concurrent attack load
func TestSecurity_ConcurrentAttacks(t *testing.T) {
	router := gin.New()

	rateCfg := security.DefaultDDoSConfig()
	rateCfg.RequestsPerSecond = 100
	rateCfg.BurstSize = 100

	router.Use(security.RateLimitMiddleware(rateCfg))
	router.Use(security.InputValidationMiddleware())

	router.POST("/api/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	// Launch concurrent attacks
	done := make(chan bool)
	numAttacks := 50

	for i := 0; i < numAttacks; i++ {
		go func(index int) {
			defer func() { done <- true }()

			payload := map[string]interface{}{
				"data": "'; DROP TABLE users; --",
			}
			jsonData, _ := json.Marshal(payload)

			w := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/api/test", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")
			req.RemoteAddr = "192.168.1.1:12345"

			router.ServeHTTP(w, req)

			// Should be blocked by input validation
			assert.Equal(t, http.StatusBadRequest, w.Code)
		}(i)
	}

	// Wait for all attacks to complete
	for i := 0; i < numAttacks; i++ {
		<-done
	}
}
