package security

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestCSRFTokenGeneration(t *testing.T) {
	store := newCSRFTokenStore(1000)

	token, err := store.generateToken(32, 1*time.Hour, "192.168.1.1", "TestAgent")

	assert.NoError(t, err)
	assert.NotEmpty(t, token.Value)
	assert.Equal(t, "192.168.1.1", token.IPAddress)
	assert.Equal(t, "TestAgent", token.UserAgent)
	assert.False(t, token.Used)
}

func TestCSRFTokenValidation(t *testing.T) {
	store := newCSRFTokenStore(1000)

	token, _ := store.generateToken(32, 1*time.Hour, "192.168.1.1", "TestAgent")

	// Valid token
	valid := store.validateToken(token.Value, "192.168.1.1", "TestAgent", true)
	assert.True(t, valid)

	// Invalid IP
	valid = store.validateToken(token.Value, "192.168.1.2", "TestAgent", true)
	assert.False(t, valid)

	// Invalid User-Agent
	valid = store.validateToken(token.Value, "192.168.1.1", "DifferentAgent", true)
	assert.False(t, valid)

	// Non-existent token
	valid = store.validateToken("invalid-token", "192.168.1.1", "TestAgent", true)
	assert.False(t, valid)
}

func TestCSRFTokenExpiration(t *testing.T) {
	store := newCSRFTokenStore(1000)

	// Create token with very short lifetime
	token, _ := store.generateToken(32, 100*time.Millisecond, "192.168.1.1", "TestAgent")

	// Token should be valid initially
	valid := store.validateToken(token.Value, "192.168.1.1", "TestAgent", true)
	assert.True(t, valid)

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Token should be invalid now
	valid = store.validateToken(token.Value, "192.168.1.1", "TestAgent", true)
	assert.False(t, valid)
}

func TestCSRFTokenOneTimeUse(t *testing.T) {
	store := newCSRFTokenStore(1000)

	token, _ := store.generateToken(32, 1*time.Hour, "192.168.1.1", "TestAgent")

	// Mark as used
	store.markTokenUsed(token.Value)

	// Token should be invalid after being used
	valid := store.validateToken(token.Value, "192.168.1.1", "TestAgent", true)
	assert.False(t, valid)
}

func TestCSRFProtectionMiddleware_GET(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := DefaultCSRFProtectionConfig()

	r := gin.New()
	r.Use(CSRFProtectionMiddleware(cfg))
	r.GET("/test", func(c *gin.Context) {
		token := GetCSRFToken(c)
		c.JSON(http.StatusOK, gin.H{"token": token})
	})

	// GET request should succeed and receive a token
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Check that a cookie was set
	cookies := w.Result().Cookies()
	assert.True(t, len(cookies) > 0)

	var csrfCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == cfg.CookieName {
			csrfCookie = cookie
			break
		}
	}

	assert.NotNil(t, csrfCookie)
	assert.NotEmpty(t, csrfCookie.Value)
}

func TestCSRFProtectionMiddleware_POST_Valid(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := DefaultCSRFProtectionConfig()

	// First, get a CSRF token
	r := gin.New()
	r.Use(CSRFProtectionMiddleware(cfg))
	r.GET("/token", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	r.POST("/submit", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// Get token
	w1 := httptest.NewRecorder()
	req1 := httptest.NewRequest("GET", "/token", nil)
	req1.RemoteAddr = "192.168.1.1:12345"
	r.ServeHTTP(w1, req1)

	// Extract token from cookie
	var tokenValue string
	for _, cookie := range w1.Result().Cookies() {
		if cookie.Name == cfg.CookieName {
			tokenValue = cookie.Value
			break
		}
	}

	assert.NotEmpty(t, tokenValue)

	// Now submit with token
	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest("POST", "/submit", nil)
	req2.RemoteAddr = "192.168.1.1:12345"
	req2.Header.Set(cfg.HeaderName, tokenValue)
	req2.AddCookie(&http.Cookie{
		Name:  cfg.CookieName,
		Value: tokenValue,
	})
	r.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusOK, w2.Code)
}

func TestCSRFProtectionMiddleware_POST_Invalid(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := DefaultCSRFProtectionConfig()

	r := gin.New()
	r.Use(CSRFProtectionMiddleware(cfg))
	r.POST("/submit", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// POST without token should fail
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/submit", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestGetCSRFToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// No token set
	token := GetCSRFToken(c)
	assert.Empty(t, token)

	// Set token
	c.Set("csrf_token", "test-token-value")
	token = GetCSRFToken(c)
	assert.Equal(t, "test-token-value", token)
}

func TestCSRFStatistics(t *testing.T) {
	// Clear global store
	globalCSRFStore = newCSRFTokenStore(10000)

	// Generate some tokens
	for i := 0; i < 10; i++ {
		globalCSRFStore.generateToken(32, 1*time.Hour, "192.168.1.1", "TestAgent")
	}

	stats := GetCSRFStatistics()

	assert.Equal(t, 10, stats.TotalTokens)
	assert.Equal(t, 10, stats.ActiveTokens)
	assert.Equal(t, 0, stats.UsedTokens)
}

func TestClearExpiredCSRFTokens(t *testing.T) {
	// Clear global store
	globalCSRFStore = newCSRFTokenStore(10000)

	// Generate expired tokens
	for i := 0; i < 5; i++ {
		globalCSRFStore.generateToken(32, 1*time.Millisecond, "192.168.1.1", "TestAgent")
	}

	// Wait for expiration
	time.Sleep(10 * time.Millisecond)

	// Clear expired tokens
	count := ClearExpiredCSRFTokens()

	assert.Equal(t, 5, count)
}

func BenchmarkCSRFTokenGeneration(b *testing.B) {
	store := newCSRFTokenStore(10000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		store.generateToken(32, 1*time.Hour, "192.168.1.1", "TestAgent")
	}
}

func BenchmarkCSRFTokenValidation(b *testing.B) {
	store := newCSRFTokenStore(10000)
	token, _ := store.generateToken(32, 1*time.Hour, "192.168.1.1", "TestAgent")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		store.validateToken(token.Value, "192.168.1.1", "TestAgent", true)
	}
}
