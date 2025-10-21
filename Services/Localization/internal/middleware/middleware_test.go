package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/helixtrack/localization-service/internal/models"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func init() {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)
}

// Helper function to create a valid JWT token
func createTestToken(username, role, permissions string, secret string) string {
	claims := &models.JWTClaims{
		Username:    username,
		Role:        role,
		Permissions: permissions,
	}
	claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(1 * time.Hour))
	claims.IssuedAt = jwt.NewNumericDate(time.Now())

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(secret))
	return tokenString
}

// Helper function to create an expired JWT token
func createExpiredToken(username string, secret string) string {
	claims := &models.JWTClaims{
		Username: username,
		Role:     "user",
	}
	claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(-1 * time.Hour))
	claims.IssuedAt = jwt.NewNumericDate(time.Now().Add(-2 * time.Hour))

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(secret))
	return tokenString
}

// TestJWTAuth_Success tests successful JWT authentication
func TestJWTAuth_Success(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	secret := "test-secret"

	// Create test server
	router := gin.New()
	router.Use(JWTAuth(secret, logger))
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Create request with valid token
	token := createTestToken("testuser", "user", "READ", secret)
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	// Record response
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "success")
}

// TestJWTAuth_MissingHeader tests missing authorization header
func TestJWTAuth_MissingHeader(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	secret := "test-secret"

	router := gin.New()
	router.Use(JWTAuth(secret, logger))
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req, _ := http.NewRequest("GET", "/protected", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "missing authorization header")
}

// TestJWTAuth_InvalidFormat tests invalid header format
func TestJWTAuth_InvalidFormat(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	secret := "test-secret"

	router := gin.New()
	router.Use(JWTAuth(secret, logger))
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	tests := []struct {
		name   string
		header string
	}{
		{"no bearer prefix", "sometoken"},
		{"wrong prefix", "Basic sometoken"},
		{"only bearer", "Bearer"},
		{"too many parts", "Bearer token extra"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/protected", nil)
			req.Header.Set("Authorization", tt.header)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusUnauthorized, w.Code)
			assert.Contains(t, w.Body.String(), "invalid authorization header format")
		})
	}
}

// TestJWTAuth_InvalidToken tests invalid JWT token
func TestJWTAuth_InvalidToken(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	secret := "test-secret"

	router := gin.New()
	router.Use(JWTAuth(secret, logger))
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.here")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "invalid or expired token")
}

// TestJWTAuth_ExpiredToken tests expired JWT token
func TestJWTAuth_ExpiredToken(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	secret := "test-secret"

	router := gin.New()
	router.Use(JWTAuth(secret, logger))
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	token := createExpiredToken("testuser", secret)
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "invalid or expired token")
}

// TestJWTAuth_WrongSecret tests token signed with wrong secret
func TestJWTAuth_WrongSecret(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	secret := "test-secret"

	router := gin.New()
	router.Use(JWTAuth(secret, logger))
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Create token with different secret
	token := createTestToken("testuser", "user", "READ", "wrong-secret")
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestAdminOnly_Success tests successful admin authorization
func TestAdminOnly_Success(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	secret := "test-secret"
	adminRoles := []string{"admin", "superadmin"}

	router := gin.New()
	router.Use(JWTAuth(secret, logger))
	router.Use(AdminOnly(adminRoles))
	router.GET("/admin", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "admin access granted"})
	})

	token := createTestToken("adminuser", "admin", "ALL", secret)
	req, _ := http.NewRequest("GET", "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "admin access granted")
}

// TestAdminOnly_Forbidden tests non-admin access
func TestAdminOnly_Forbidden(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	secret := "test-secret"
	adminRoles := []string{"admin", "superadmin"}

	router := gin.New()
	router.Use(JWTAuth(secret, logger))
	router.Use(AdminOnly(adminRoles))
	router.GET("/admin", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "admin access granted"})
	})

	// Create token for regular user
	token := createTestToken("regularuser", "user", "READ", secret)
	req, _ := http.NewRequest("GET", "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "admin access required")
}

// TestAdminOnly_MissingClaims tests admin check without authentication
func TestAdminOnly_MissingClaims(t *testing.T) {
	adminRoles := []string{"admin", "superadmin"}

	router := gin.New()
	router.Use(AdminOnly(adminRoles)) // No JWTAuth middleware before
	router.GET("/admin", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "admin access granted"})
	})

	req, _ := http.NewRequest("GET", "/admin", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "authentication required")
}

// TestGetClaims tests retrieving claims from context
func TestGetClaims(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	secret := "test-secret"

	router := gin.New()
	router.Use(JWTAuth(secret, logger))
	router.GET("/test", func(c *gin.Context) {
		claims := GetClaims(c)
		if claims != nil {
			c.JSON(http.StatusOK, gin.H{"username": claims.Username})
		} else {
			c.JSON(http.StatusOK, gin.H{"username": "none"})
		}
	})

	token := createTestToken("testuser", "user", "READ", secret)
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "testuser")
}

// TestGetClaims_NoClaims tests GetClaims when no claims in context
func TestGetClaims_NoClaims(t *testing.T) {
	router := gin.New()
	router.GET("/test", func(c *gin.Context) {
		claims := GetClaims(c)
		if claims == nil {
			c.JSON(http.StatusOK, gin.H{"message": "no claims"})
		}
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "no claims")
}

// TestCORS tests CORS middleware
func TestCORS(t *testing.T) {
	router := gin.New()
	router.Use(CORS())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Headers"), "Authorization")
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "GET")
}

// TestCORS_OptionsRequest tests CORS preflight request
func TestCORS_OptionsRequest(t *testing.T) {
	router := gin.New()
	router.Use(CORS())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req, _ := http.NewRequest("OPTIONS", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
}

// TestRequestLogger tests request logging middleware
func TestRequestLogger(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	router := gin.New()
	router.Use(RequestLogger(logger))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req, _ := http.NewRequest("GET", "/test?param=value", nil)
	req.Header.Set("User-Agent", "test-agent")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	// Logger should not affect response
	assert.Contains(t, w.Body.String(), "success")
}

// TestRateLimiter_GlobalLimit tests global rate limiting
func TestRateLimiter_GlobalLimit(t *testing.T) {
	// Create rate limiter with very low global limit
	rl := NewRateLimiter(100, 100, 20) // 20 requests per minute globally (burst = 2)
	defer rl.Close()

	router := gin.New()
	router.Use(rl.RateLimit())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// First request should succeed
	req1, _ := http.NewRequest("GET", "/test", nil)
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)

	// Second request should succeed (burst allows)
	req2, _ := http.NewRequest("GET", "/test", nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)

	// Third request should be rate limited
	req3, _ := http.NewRequest("GET", "/test", nil)
	w3 := httptest.NewRecorder()
	router.ServeHTTP(w3, req3)
	assert.Equal(t, http.StatusTooManyRequests, w3.Code)
	assert.Contains(t, w3.Body.String(), "global rate limit exceeded")
}

// TestRateLimiter_IPLimit tests IP-based rate limiting
func TestRateLimiter_IPLimit(t *testing.T) {
	// Create rate limiter with low IP limit
	rl := NewRateLimiter(20, 100, 1000) // 20 requests per minute per IP (burst = 2)
	defer rl.Close()

	router := gin.New()
	router.Use(rl.RateLimit())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// First request from same IP should succeed
	req1, _ := http.NewRequest("GET", "/test", nil)
	req1.RemoteAddr = "192.168.1.1:12345"
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)

	// Second request from same IP should succeed
	req2, _ := http.NewRequest("GET", "/test", nil)
	req2.RemoteAddr = "192.168.1.1:12346"
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)

	// Third request from same IP should be rate limited
	req3, _ := http.NewRequest("GET", "/test", nil)
	req3.RemoteAddr = "192.168.1.1:12347"
	w3 := httptest.NewRecorder()
	router.ServeHTTP(w3, req3)
	assert.Equal(t, http.StatusTooManyRequests, w3.Code)
	assert.Contains(t, w3.Body.String(), "IP rate limit exceeded")
}

// TestRateLimiter_UserLimit tests user-based rate limiting
func TestRateLimiter_UserLimit(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	secret := "test-secret"

	// Create rate limiter with low user limit
	rl := NewRateLimiter(1000, 20, 1000) // 20 requests per minute per user (burst = 2)
	defer rl.Close()

	router := gin.New()
	router.Use(JWTAuth(secret, logger))
	router.Use(rl.RateLimit())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	token := createTestToken("testuser", "user", "READ", secret)

	// First request should succeed
	req1, _ := http.NewRequest("GET", "/test", nil)
	req1.Header.Set("Authorization", "Bearer "+token)
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)

	// Second request should succeed
	req2, _ := http.NewRequest("GET", "/test", nil)
	req2.Header.Set("Authorization", "Bearer "+token)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)

	// Third request should be rate limited
	req3, _ := http.NewRequest("GET", "/test", nil)
	req3.Header.Set("Authorization", "Bearer "+token)
	w3 := httptest.NewRecorder()
	router.ServeHTTP(w3, req3)
	assert.Equal(t, http.StatusTooManyRequests, w3.Code)
	assert.Contains(t, w3.Body.String(), "user rate limit exceeded")
}

// TestRateLimiter_DifferentIPs tests that different IPs have separate limits
func TestRateLimiter_DifferentIPs(t *testing.T) {
	rl := NewRateLimiter(20, 100, 1000) // 20 requests per minute per IP (burst = 2)
	defer rl.Close()

	router := gin.New()
	router.Use(rl.RateLimit())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Exhaust limit for first IP
	req1, _ := http.NewRequest("GET", "/test", nil)
	req1.RemoteAddr = "192.168.1.1:12345"
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)

	req2, _ := http.NewRequest("GET", "/test", nil)
	req2.RemoteAddr = "192.168.1.1:12346"
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)

	// Request from different IP should succeed
	req3, _ := http.NewRequest("GET", "/test", nil)
	req3.RemoteAddr = "192.168.1.2:12345"
	w3 := httptest.NewRecorder()
	router.ServeHTTP(w3, req3)
	assert.Equal(t, http.StatusOK, w3.Code)
}

// TestRateLimiter_Close tests closing the rate limiter
func TestRateLimiter_Close(t *testing.T) {
	rl := NewRateLimiter(100, 100, 1000)

	// Should not panic
	assert.NotPanics(t, func() {
		rl.Close()
	})
}
