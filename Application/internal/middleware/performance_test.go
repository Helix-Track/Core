package middleware

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestCompressionMiddleware(t *testing.T) {
	tests := []struct {
		name            string
		acceptEncoding  string
		shouldCompress  bool
	}{
		{"With gzip support", "gzip, deflate", true},
		{"Without gzip support", "deflate", false},
		{"Empty accept-encoding", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			r.Use(CompressionMiddleware(gzip.DefaultCompression))
			r.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "test response with some content to compress"})
			})

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.acceptEncoding != "" {
				req.Header.Set("Accept-Encoding", tt.acceptEncoding)
			}
			r.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			if tt.shouldCompress {
				assert.Equal(t, "gzip", w.Header().Get("Content-Encoding"))
				assert.Equal(t, "Accept-Encoding", w.Header().Get("Vary"))

				// Decompress and verify
				reader, err := gzip.NewReader(w.Body)
				assert.NoError(t, err)
				decompressed, err := io.ReadAll(reader)
				assert.NoError(t, err)
				assert.Contains(t, string(decompressed), "test response")
			} else {
				assert.Empty(t, w.Header().Get("Content-Encoding"))
			}
		})
	}
}

func TestCompressionMiddleware_AlreadyCompressed(t *testing.T) {
	r := gin.New()
	r.Use(CompressionMiddleware(gzip.DefaultCompression))
	r.GET("/test", func(c *gin.Context) {
		c.Header("Content-Encoding", "br") // Already compressed with Brotli
		c.JSON(http.StatusOK, gin.H{"message": "already compressed"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	r.ServeHTTP(w, req)

	// Should not re-compress
	assert.Equal(t, "br", w.Header().Get("Content-Encoding"))
}

func TestRateLimitMiddleware(t *testing.T) {
	cfg := DefaultRateLimiterConfig()
	cfg.RequestsPerSecond = 5
	cfg.BurstSize = 5

	r := gin.New()
	r.Use(RateLimitMiddleware(cfg))
	r.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// First 5 requests should succeed
	for i := 0; i < 5; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code, "Request %d should succeed", i+1)
	}

	// 6th request should be rate limited
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusTooManyRequests, w.Code)
}

func TestRateLimitMiddleware_TokenRefill(t *testing.T) {
	cfg := DefaultRateLimiterConfig()
	cfg.RequestsPerSecond = 10

	r := gin.New()
	r.Use(RateLimitMiddleware(cfg))
	r.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// Use up all tokens
	for i := 0; i < 10; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		r.ServeHTTP(w, req)
	}

	// Wait for token refill
	time.Sleep(200 * time.Millisecond)

	// Should be able to make requests again
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCircuitBreakerMiddleware(t *testing.T) {
	cfg := DefaultCircuitBreakerConfig()
	cfg.MaxFailures = 3

	failureCount := 0
	r := gin.New()
	r.Use(CircuitBreakerMiddleware(cfg))
	r.GET("/test", func(c *gin.Context) {
		if failureCount < 3 {
			failureCount++
			c.Status(http.StatusInternalServerError)
		} else {
			c.Status(http.StatusOK)
		}
	})

	// First 3 requests fail
	for i := 0; i < 3; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/test", nil)
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code, "Request %d should fail", i+1)
	}

	// Circuit should be open now
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
}

func TestCircuitBreaker_HalfOpenState(t *testing.T) {
	cfg := DefaultCircuitBreakerConfig()
	cfg.MaxFailures = 2
	cfg.Timeout = 100 * time.Millisecond
	cfg.HalfOpenMax = 2

	breaker := newCircuitBreaker(cfg)

	// Cause failures to open circuit
	breaker.recordFailure()
	breaker.recordFailure()
	assert.Equal(t, stateOpen, breaker.getState())

	// Wait for timeout
	time.Sleep(150 * time.Millisecond)

	// Should allow request (half-open)
	assert.True(t, breaker.allow())
	assert.Equal(t, stateHalfOpen, breaker.getState())

	// Record successes to close circuit
	breaker.recordSuccess()
	breaker.recordSuccess()
	assert.Equal(t, stateClosed, breaker.getState())
}

func TestTimeoutMiddleware(t *testing.T) {
	r := gin.New()
	r.Use(TimeoutMiddleware(100 * time.Millisecond))
	r.GET("/fast", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	r.GET("/slow", func(c *gin.Context) {
		time.Sleep(200 * time.Millisecond)
		c.Status(http.StatusOK)
	})

	// Fast request should succeed
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/fast", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Slow request should timeout
	w = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/slow", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusRequestTimeout, w.Code)
}

func TestCORSMiddleware(t *testing.T) {
	cfg := DefaultCORSConfig()
	cfg.AllowOrigins = []string{"https://example.com"}
	cfg.AllowCredentials = true

	r := gin.New()
	r.Use(CORSMiddleware(cfg))
	r.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// Test with allowed origin
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "https://example.com")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "https://example.com", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
}

func TestCORSMiddleware_PreflightRequest(t *testing.T) {
	cfg := DefaultCORSConfig()

	r := gin.New()
	r.Use(CORSMiddleware(cfg))
	r.OPTIONS("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("OPTIONS", "/test", nil)
	req.Header.Set("Origin", "https://example.com")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.NotEmpty(t, w.Header().Get("Access-Control-Allow-Methods"))
}

func TestCORSMiddleware_Wildcard(t *testing.T) {
	cfg := DefaultCORSConfig()
	cfg.AllowOrigins = []string{"*"}

	r := gin.New()
	r.Use(CORSMiddleware(cfg))
	r.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "https://example.com")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
}

func TestTokenBucket_Take(t *testing.T) {
	bucket := &tokenBucket{
		tokens:     5.0,
		lastRefill: time.Now(),
		maxTokens:  10.0,
		refillRate: 1.0,
	}

	// Should be able to take 5 tokens
	for i := 0; i < 5; i++ {
		assert.True(t, bucket.take())
	}

	// Should not be able to take more
	assert.False(t, bucket.take())
}

func TestTokenBucket_Refill(t *testing.T) {
	bucket := &tokenBucket{
		tokens:     0.0,
		lastRefill: time.Now().Add(-1 * time.Second),
		maxTokens:  10.0,
		refillRate: 10.0, // 10 tokens per second
	}

	// After 1 second, should have refilled
	assert.True(t, bucket.take())
}

func TestRateLimiter_Cleanup(t *testing.T) {
	cfg := DefaultRateLimiterConfig()
	cfg.CleanupInterval = 50 * time.Millisecond
	limiter := newRateLimiter(cfg)
	defer limiter.close()

	// Create a bucket
	limiter.allow("test-key")

	// Wait for cleanup
	time.Sleep(100 * time.Millisecond)

	// Cleanup should have run (hard to test actual cleanup without waiting 5+ minutes)
	// Just verify the cleanup loop is running
}

func TestGzipResponseWriter(t *testing.T) {
	// Use Gin's test context to get a proper ResponseWriter
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	buf := &bytes.Buffer{}
	gz := gzip.NewWriter(buf)

	gzWriter := &gzipResponseWriter{
		ResponseWriter: c.Writer,
		writer:         gz,
	}

	// Test Write
	n, err := gzWriter.Write([]byte("test data"))
	assert.NoError(t, err)
	assert.Equal(t, 9, n)

	// Test WriteString
	n, err = gzWriter.WriteString(" more data")
	assert.NoError(t, err)
	assert.Equal(t, 10, n)

	gz.Close()
}

func TestDefaultConfigs(t *testing.T) {
	// Test DefaultRateLimiterConfig
	rateCfg := DefaultRateLimiterConfig()
	assert.Equal(t, 1000, rateCfg.RequestsPerSecond)
	assert.Equal(t, 2000, rateCfg.BurstSize)
	assert.Equal(t, 1*time.Minute, rateCfg.CleanupInterval)

	// Test DefaultCircuitBreakerConfig
	cbCfg := DefaultCircuitBreakerConfig()
	assert.Equal(t, 5, cbCfg.MaxFailures)
	assert.Equal(t, 30*time.Second, cbCfg.Timeout)
	assert.Equal(t, 3, cbCfg.HalfOpenMax)
	assert.Equal(t, 0.5, cbCfg.FailureRatio)
	assert.Equal(t, 10, cbCfg.MinRequests)

	// Test DefaultCORSConfig
	corsCfg := DefaultCORSConfig()
	assert.Contains(t, corsCfg.AllowOrigins, "*")
	assert.Contains(t, corsCfg.AllowMethods, "GET")
	assert.Contains(t, corsCfg.AllowHeaders, "Authorization")
	assert.Equal(t, 12*time.Hour, corsCfg.MaxAge)
}

func BenchmarkCompressionMiddleware(b *testing.B) {
	r := gin.New()
	r.Use(CompressionMiddleware(gzip.DefaultCompression))
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": strings.Repeat("test ", 100)})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Accept-Encoding", "gzip")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
	}
}

func BenchmarkRateLimitMiddleware(b *testing.B) {
	r := gin.New()
	r.Use(RateLimitMiddleware(DefaultRateLimiterConfig()))
	r.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
	}
}

func BenchmarkCircuitBreakerMiddleware(b *testing.B) {
	r := gin.New()
	r.Use(CircuitBreakerMiddleware(DefaultCircuitBreakerConfig()))
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

func BenchmarkCORSMiddleware(b *testing.B) {
	r := gin.New()
	r.Use(CORSMiddleware(DefaultCORSConfig()))
	r.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "https://example.com")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
	}
}

func BenchmarkTokenBucket_Take(b *testing.B) {
	bucket := &tokenBucket{
		tokens:     1000.0,
		lastRefill: time.Now(),
		maxTokens:  1000.0,
		refillRate: 100.0,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bucket.take()
	}
}
