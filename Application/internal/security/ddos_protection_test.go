package security

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestDDoSProtectionMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := DefaultDDoSProtectionConfig()
	cfg.MaxRequestsPerSecond = 5
	cfg.MaxRequestSize = 1024 // 1KB
	cfg.MaxURILength = 100

	tests := []struct {
		name           string
		requests       int
		expectedStatus int
		sleepBetween   time.Duration
	}{
		{
			name:           "Under rate limit",
			requests:       3,
			expectedStatus: http.StatusOK,
			sleepBetween:   0,
		},
		{
			name:           "Exceed rate limit",
			requests:       10,
			expectedStatus: http.StatusTooManyRequests,
			sleepBetween:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			r.Use(DDoSProtectionMiddleware(cfg))
			r.GET("/test", func(c *gin.Context) {
				c.Status(http.StatusOK)
			})

			lastStatus := 0
			for i := 0; i < tt.requests; i++ {
				w := httptest.NewRecorder()
				req := httptest.NewRequest("GET", "/test", nil)
				req.RemoteAddr = "192.168.1.1:12345"
				r.ServeHTTP(w, req)
				lastStatus = w.Code

				if tt.sleepBetween > 0 {
					time.Sleep(tt.sleepBetween)
				}
			}

			assert.Equal(t, tt.expectedStatus, lastStatus)
		})
	}
}

func TestDDoSRequestSizeLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := DefaultDDoSProtectionConfig()
	cfg.MaxRequestSize = 1024 // 1KB

	r := gin.New()
	r.Use(DDoSProtectionMiddleware(cfg))
	r.POST("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// Test request exceeding size limit
	w := httptest.NewRecorder()
	largeBody := strings.NewReader(string(make([]byte, 2048))) // 2KB
	req := httptest.NewRequest("POST", "/test", largeBody)
	req.ContentLength = 2048
	req.RemoteAddr = "192.168.1.1:12345"
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusRequestEntityTooLarge, w.Code)
}

func TestDDoSURILengthLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := DefaultDDoSProtectionConfig()
	cfg.MaxURILength = 100

	r := gin.New()
	r.Use(DDoSProtectionMiddleware(cfg))
	r.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// Test URI exceeding length limit
	w := httptest.NewRecorder()
	longURI := "/test?" + strings.Repeat("a", 200)
	req := httptest.NewRequest("GET", longURI, nil)
	req.RemoteAddr = "192.168.1.1:12345"
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusRequestURITooLong, w.Code)
}

func TestDDoSProtectorCheckRequest(t *testing.T) {
	cfg := DefaultDDoSProtectionConfig()
	cfg.MaxRequestsPerSecond = 5
	protector := newDDoSProtector(cfg)
	defer protector.close()

	ip := "192.168.1.1"

	// First 5 requests should be allowed
	for i := 0; i < 5; i++ {
		allowed, _ := protector.checkRequest(ip)
		assert.True(t, allowed, "Request %d should be allowed", i+1)
	}

	// 6th request should be blocked (rate limit exceeded)
	allowed, reason := protector.checkRequest(ip)
	assert.False(t, allowed)
	assert.Contains(t, reason, "Rate limit exceeded")
}

func TestDDoSProtectorWhitelist(t *testing.T) {
	cfg := DefaultDDoSProtectionConfig()
	cfg.MaxRequestsPerSecond = 5
	protector := newDDoSProtector(cfg)
	defer protector.close()

	ip := "192.168.1.100"

	// Whitelist the IP
	protector.whitelistIP(ip)

	// Even after many requests, whitelisted IP should be allowed
	for i := 0; i < 20; i++ {
		allowed, _ := protector.checkRequest(ip)
		assert.True(t, allowed, "Whitelisted IP should always be allowed")
	}
}

func TestDDoSProtectorBlockIP(t *testing.T) {
	cfg := DefaultDDoSProtectionConfig()
	cfg.EnableIPBlocking = true
	cfg.BlockDuration = 100 * time.Millisecond
	protector := newDDoSProtector(cfg)
	defer protector.close()

	ip := "192.168.1.2"

	// Block the IP
	protector.blockIP(ip, "Test block")

	// Request should be blocked
	allowed, reason := protector.checkRequest(ip)
	assert.False(t, allowed)
	assert.Contains(t, reason, "blocked")

	// Wait for block to expire
	time.Sleep(150 * time.Millisecond)

	// Request should be allowed again
	allowed, _ = protector.checkRequest(ip)
	assert.True(t, allowed)
}

func TestDDoSProtectorConcurrentLimit(t *testing.T) {
	cfg := DefaultDDoSProtectionConfig()
	cfg.MaxConcurrentPerIP = 3
	protector := newDDoSProtector(cfg)
	defer protector.close()

	ip := "192.168.1.3"

	// Open 3 concurrent connections
	for i := 0; i < 3; i++ {
		allowed, _ := protector.checkRequest(ip)
		assert.True(t, allowed)
	}

	// 4th concurrent connection should be blocked
	allowed, reason := protector.checkRequest(ip)
	assert.False(t, allowed)
	assert.Contains(t, reason, "concurrent")

	// Release one connection
	protector.releaseRequest(ip)

	// Now should be allowed again
	allowed, _ = protector.checkRequest(ip)
	assert.True(t, allowed)
}

func TestDDoSProtectorCleanup(t *testing.T) {
	cfg := DefaultDDoSProtectionConfig()
	cfg.CleanupInterval = 50 * time.Millisecond
	protector := newDDoSProtector(cfg)
	defer protector.close()

	// Create some entries
	for i := 0; i < 10; i++ {
		ip := "192.168.1." + string(rune(100+i))
		protector.checkRequest(ip)
	}

	// Wait for cleanup
	time.Sleep(100 * time.Millisecond)

	// Cleanup should have run
	// (This is a basic test, more sophisticated testing would check actual cleanup)
}

func TestDDoSProtectorStatistics(t *testing.T) {
	cfg := DefaultDDoSProtectionConfig()
	protector := newDDoSProtector(cfg)
	defer protector.close()

	// Generate some traffic
	for i := 0; i < 5; i++ {
		ip := "192.168.1." + string(rune(100+i))
		protector.checkRequest(ip)
	}

	// Block one IP
	protector.blockIP("192.168.1.100", "Test")

	// Whitelist one IP
	protector.whitelistIP("192.168.1.200")

	stats := protector.GetStatistics()

	assert.True(t, stats.TrackedIPs > 0)
	assert.True(t, stats.BlockedIPs >= 1)
	assert.True(t, stats.WhitelistedIPs >= 1)
}

func TestExtractIPFromContext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		remoteAddr string
		headers    map[string]string
		expectedIP string
	}{
		{
			name:       "Direct IP",
			remoteAddr: "192.168.1.1:12345",
			headers:    map[string]string{},
			expectedIP: "192.168.1.1",
		},
		{
			name:       "X-Forwarded-For",
			remoteAddr: "10.0.0.1:12345",
			headers: map[string]string{
				"X-Forwarded-For": "203.0.113.1",
			},
			expectedIP: "203.0.113.1",
		},
		{
			name:       "X-Real-IP",
			remoteAddr: "10.0.0.1:12345",
			headers: map[string]string{
				"X-Real-IP": "203.0.113.2",
			},
			expectedIP: "203.0.113.2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			req := httptest.NewRequest("GET", "/test", nil)
			req.RemoteAddr = tt.remoteAddr
			for key, value := range tt.headers {
				req.Header.Set(key, value)
			}
			c.Request = req

			ip := extractIPFromContext(c)
			assert.Equal(t, tt.expectedIP, ip)
		})
	}
}

func BenchmarkDDoSCheckRequest(b *testing.B) {
	cfg := DefaultDDoSProtectionConfig()
	protector := newDDoSProtector(cfg)
	defer protector.close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		protector.checkRequest("192.168.1.1")
	}
}

func BenchmarkDDoSReleaseRequest(b *testing.B) {
	cfg := DefaultDDoSProtectionConfig()
	protector := newDDoSProtector(cfg)
	defer protector.close()

	// Setup
	protector.checkRequest("192.168.1.1")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		protector.releaseRequest("192.168.1.1")
	}
}
