package middleware

import (
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"helixtrack.ru/chat/internal/models"
)

func TestJWTMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := "test-secret"

	// Create valid token
	validToken := createTestToken(t, secret, map[string]interface{}{
		"sub":           "test",
		"username":      "testuser",
		"user_id":       uuid.New().String(),
		"role":          "admin",
		"permissions":   "READ|CREATE|UPDATE|DELETE",
		"htCoreAddress": "http://localhost:8080",
	})

	tests := []struct {
		name         string
		token        string
		expectStatus int
		expectAbort  bool
	}{
		{
			name:         "valid token in header",
			token:        "Bearer " + validToken,
			expectStatus: 200,
			expectAbort:  false,
		},
		{
			name:         "valid token in query",
			token:        "?token=" + validToken,
			expectStatus: 200,
			expectAbort:  false,
		},
		{
			name:         "missing token",
			token:        "",
			expectStatus: 401,
			expectAbort:  true,
		},
		{
			name:         "invalid token",
			token:        "Bearer invalid",
			expectStatus: 401,
			expectAbort:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			if strings.HasPrefix(tt.token, "Bearer") {
				c.Request = httptest.NewRequest("GET", "/test", nil)
				c.Request.Header.Set("Authorization", tt.token)
			} else if strings.HasPrefix(tt.token, "?token=") {
				c.Request = httptest.NewRequest("GET", "/test"+tt.token, nil)
			} else {
				c.Request = httptest.NewRequest("GET", "/test", nil)
			}

			middleware := JWTMiddleware(secret)
			middleware(c)

			if tt.expectAbort {
				assert.True(t, c.IsAborted())
			} else {
				assert.False(t, c.IsAborted())
				_, exists := c.Get("claims")
				assert.True(t, exists)
			}
		})
	}
}

func TestRateLimiter(t *testing.T) {
	rl := NewRateLimiter(2, 5) // 2 req/sec, burst of 5
	defer rl.Stop()

	ip := "192.168.1.1"

	// First 5 requests should succeed (burst)
	for i := 0; i < 5; i++ {
		assert.True(t, rl.Allow(ip), "Request %d should be allowed", i+1)
	}

	// Next request should be blocked
	assert.False(t, rl.Allow(ip), "Request 6 should be blocked")

	// Wait for rate limit to refill
	time.Sleep(600 * time.Millisecond) // Allow ~1 request

	// Should allow one more
	assert.True(t, rl.Allow(ip), "Request after wait should be allowed")
}

func TestRateLimitMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rl := NewRateLimiter(1, 2)
	defer rl.Stop()

	router := gin.New()
	router.Use(RateLimitMiddleware(rl))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	// First 2 requests should succeed
	for i := 0; i < 2; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)
		assert.Equal(t, 200, w.Code, "Request %d should succeed", i+1)
	}

	// Third request should be rate limited
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, 429, w.Code, "Request 3 should be rate limited")
}

func TestCORSMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		allowedOrigins []string
		requestOrigin  string
		expectCORS     bool
	}{
		{
			name:           "wildcard allows all",
			allowedOrigins: []string{"*"},
			requestOrigin:  "http://example.com",
			expectCORS:     true,
		},
		{
			name:           "exact match",
			allowedOrigins: []string{"http://example.com"},
			requestOrigin:  "http://example.com",
			expectCORS:     true,
		},
		{
			name:           "no match",
			allowedOrigins: []string{"http://allowed.com"},
			requestOrigin:  "http://notallowed.com",
			expectCORS:     false,
		},
		{
			name:           "pattern match",
			allowedOrigins: []string{"*.example.com"},
			requestOrigin:  "http://app.example.com",
			expectCORS:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(CORSMiddleware(tt.allowedOrigins))
			router.GET("/test", func(c *gin.Context) {
				c.JSON(200, gin.H{"ok": true})
			})

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/test", nil)
			req.Header.Set("Origin", tt.requestOrigin)
			router.ServeHTTP(w, req)

			if tt.expectCORS {
				assert.Equal(t, tt.requestOrigin, w.Header().Get("Access-Control-Allow-Origin"))
			} else {
				assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
			}
		})
	}
}

func TestCORSPreflight(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(CORSMiddleware([]string{"*"}))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("OPTIONS", "/test", nil)
	req.Header.Set("Origin", "http://example.com")
	router.ServeHTTP(w, req)

	assert.Equal(t, 204, w.Code)
	assert.NotEmpty(t, w.Header().Get("Access-Control-Allow-Methods"))
}

func TestRequestLoggerMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(RequestLoggerMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func TestRecoveryMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(RecoveryMiddleware())
	router.GET("/panic", func(c *gin.Context) {
		panic("test panic")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/panic", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 500, w.Code)
}

func TestMessageSizeMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(MessageSizeMiddleware(100)) // 100 bytes max
	router.POST("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	t.Run("within limit", func(t *testing.T) {
		w := httptest.NewRecorder()
		body := strings.NewReader("small body")
		req := httptest.NewRequest("POST", "/test", body)
		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
	})

	t.Run("exceeds limit", func(t *testing.T) {
		w := httptest.NewRecorder()
		body := strings.NewReader(strings.Repeat("x", 200))
		req := httptest.NewRequest("POST", "/test", body)
		req.ContentLength = 200
		router.ServeHTTP(w, req)

		assert.Equal(t, 413, w.Code)
	})
}

// Helper function to create test JWT token
func createTestToken(t *testing.T, secret string, claims map[string]interface{}) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims(claims))
	tokenString, err := token.SignedString([]byte(secret))
	require.NoError(t, err)
	return tokenString
}

func TestGetClaims(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Test missing claims
	_, err := GetClaims(c)
	assert.Error(t, err)

	// Test valid claims
	claims := &models.JWTClaims{
		Username: "test",
		UserID:   uuid.New(),
	}
	c.Set("claims", claims)

	retrievedClaims, err := GetClaims(c)
	assert.NoError(t, err)
	assert.Equal(t, claims.Username, retrievedClaims.Username)
}

func TestGetUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Test missing user_id
	_, err := GetUserID(c)
	assert.Error(t, err)

	// Test valid user_id
	userID := uuid.New().String()
	c.Set("user_id", userID)

	retrievedID, err := GetUserID(c)
	assert.NoError(t, err)
	assert.Equal(t, userID, retrievedID)
}
