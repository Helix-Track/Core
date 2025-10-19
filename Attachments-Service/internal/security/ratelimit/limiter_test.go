package ratelimit

import (
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestNewLimiter(t *testing.T) {
	logger := zap.NewNop()

	t.Run("with nil config uses defaults", func(t *testing.T) {
		limiter := NewLimiter(nil, logger)
		if limiter == nil {
			t.Fatal("expected limiter, got nil")
		}
		if limiter.config == nil {
			t.Fatal("expected default config")
		}
	})

	t.Run("with custom config", func(t *testing.T) {
		config := &LimiterConfig{
			IPRequestsPerSecond: 5,
			IPBurstSize:         10,
			CleanupInterval:     5 * time.Minute,
		}
		limiter := NewLimiter(config, logger)
		if limiter.config.IPRequestsPerSecond != 5 {
			t.Errorf("expected IPRequestsPerSecond 5, got %d", limiter.config.IPRequestsPerSecond)
		}
	})
}

func TestTokenBucket_Allow(t *testing.T) {
	t.Run("allows requests within rate", func(t *testing.T) {
		bucket := NewTokenBucket(10, 10) // 10 req/sec, burst 10

		// First 10 requests should succeed (burst)
		for i := 0; i < 10; i++ {
			if !bucket.Allow() {
				t.Errorf("request %d should be allowed", i)
			}
		}

		// 11th request should fail (no tokens left)
		if bucket.Allow() {
			t.Error("request 11 should be denied")
		}
	})

	t.Run("refills tokens over time", func(t *testing.T) {
		bucket := NewTokenBucket(10, 10) // 10 req/sec, burst 10

		// Consume all tokens
		for i := 0; i < 10; i++ {
			bucket.Allow()
		}

		// Wait for refill (100ms = 1 token at 10/sec)
		time.Sleep(150 * time.Millisecond)

		// Should have at least 1 token now
		if !bucket.Allow() {
			t.Error("expected at least 1 token after refill")
		}
	})

	t.Run("caps tokens at burst size", func(t *testing.T) {
		bucket := NewTokenBucket(10, 5) // 10 req/sec, burst 5

		// Wait to accumulate tokens
		time.Sleep(1 * time.Second)

		// Should only allow burst size (5), not more
		count := 0
		for i := 0; i < 10; i++ {
			if bucket.Allow() {
				count++
			}
		}

		if count > 5 {
			t.Errorf("expected max 5 requests, got %d", count)
		}
	})
}

func TestTokenBucket_Available(t *testing.T) {
	bucket := NewTokenBucket(10, 10)

	available := bucket.Available()
	if available != 10 {
		t.Errorf("expected 10 available tokens, got %d", available)
	}

	// Consume 5 tokens
	for i := 0; i < 5; i++ {
		bucket.Allow()
	}

	available = bucket.Available()
	if available != 5 {
		t.Errorf("expected 5 available tokens, got %d", available)
	}
}

func TestTokenBucket_Reset(t *testing.T) {
	bucket := NewTokenBucket(10, 10)

	// Consume all tokens
	for i := 0; i < 10; i++ {
		bucket.Allow()
	}

	if bucket.Available() != 0 {
		t.Error("expected 0 available tokens")
	}

	bucket.Reset()

	if bucket.Available() != 10 {
		t.Error("expected 10 tokens after reset")
	}
}

func TestLimiter_Allow(t *testing.T) {
	logger := zap.NewNop()
	config := &LimiterConfig{
		EnableIPRateLimit: true,
		IPRequestsPerSecond: 5,
		IPBurstSize: 5,
		EnableUserRateLimit: true,
		UserRequestsPerSecond: 10,
		UserBurstSize: 10,
		EnableGlobalRateLimit: true,
		GlobalRequestsPerSecond: 100,
		GlobalBurstSize: 100,
	}
	limiter := NewLimiter(config, logger)

	t.Run("allows requests within limits", func(t *testing.T) {
		allowed, err := limiter.Allow("192.168.1.1", "user1")
		if !allowed || err != nil {
			t.Errorf("expected request to be allowed, got allowed=%v, err=%v", allowed, err)
		}
	})

	t.Run("blocks requests exceeding IP limit", func(t *testing.T) {
		ip := "192.168.1.2"

		// Make 5 requests (burst size)
		for i := 0; i < 5; i++ {
			allowed, _ := limiter.Allow(ip, "user1")
			if !allowed {
				t.Errorf("request %d should be allowed", i)
			}
		}

		// 6th request should be blocked
		allowed, err := limiter.Allow(ip, "user1")
		if allowed {
			t.Error("expected request to be blocked")
		}
		if err == nil {
			t.Error("expected error for rate limit exceeded")
		}
	})

	t.Run("blocks requests exceeding user limit", func(t *testing.T) {
		user := "user2"

		// Make 10 requests (burst size)
		for i := 0; i < 10; i++ {
			limiter.Allow("192.168.1."+string(rune(i+3)), user)
		}

		// Next request should be blocked
		allowed, err := limiter.Allow("192.168.1.99", user)
		if allowed {
			t.Error("expected request to be blocked")
		}
		if err == nil {
			t.Error("expected error for rate limit exceeded")
		}
	})
}

func TestLimiter_Whitelist(t *testing.T) {
	logger := zap.NewNop()
	config := &LimiterConfig{
		EnableIPRateLimit: true,
		IPRequestsPerSecond: 1,
		IPBurstSize: 1,
		WhitelistedIPs: []string{"127.0.0.1"},
	}
	limiter := NewLimiter(config, logger)

	// Whitelisted IP should always be allowed
	for i := 0; i < 100; i++ {
		allowed, err := limiter.Allow("127.0.0.1", "")
		if !allowed || err != nil {
			t.Errorf("whitelisted IP should always be allowed (request %d)", i)
		}
	}
}

func TestLimiter_Blacklist(t *testing.T) {
	logger := zap.NewNop()
	config := &LimiterConfig{
		BlacklistedIPs: []string{"10.0.0.1"},
	}
	limiter := NewLimiter(config, logger)

	allowed, err := limiter.Allow("10.0.0.1", "")
	if allowed {
		t.Error("blacklisted IP should be blocked")
	}
	if err == nil {
		t.Error("expected error for blacklisted IP")
	}
}

func TestLimiter_AddRemoveBlacklist(t *testing.T) {
	logger := zap.NewNop()
	limiter := NewLimiter(nil, logger)

	ip := "192.168.1.100"

	// Initially not blacklisted
	allowed, _ := limiter.Allow(ip, "")
	if !allowed {
		t.Error("IP should be allowed initially")
	}

	// Add to blacklist
	limiter.AddToBlacklist(ip)

	// Should be blocked now
	allowed, err := limiter.Allow(ip, "")
	if allowed {
		t.Error("IP should be blocked after blacklisting")
	}
	if err == nil {
		t.Error("expected error for blacklisted IP")
	}

	// Remove from blacklist
	limiter.RemoveFromBlacklist(ip)

	// Should be allowed again
	allowed, _ = limiter.Allow(ip, "")
	if !allowed {
		t.Error("IP should be allowed after removing from blacklist")
	}
}

func TestLimiter_AllowUpload(t *testing.T) {
	logger := zap.NewNop()
	config := &LimiterConfig{
		UploadRequestsPerMinute: 6, // 6 per minute = 0.1 per second
		UploadBurstSize: 2,
	}
	limiter := NewLimiter(config, logger)

	ip := "192.168.1.50"

	// First 2 uploads should succeed (burst)
	for i := 0; i < 2; i++ {
		allowed, err := limiter.AllowUpload(ip, "")
		if !allowed || err != nil {
			t.Errorf("upload %d should be allowed", i)
		}
	}

	// 3rd upload should fail
	allowed, err := limiter.AllowUpload(ip, "")
	if allowed {
		t.Error("upload should be rate limited")
	}
	if err == nil {
		t.Error("expected error for upload rate limit")
	}
}

func TestLimiter_AllowDownload(t *testing.T) {
	logger := zap.NewNop()
	config := &LimiterConfig{
		DownloadRequestsPerMinute: 60, // 60 per minute = 1 per second
		DownloadBurstSize: 10,
	}
	limiter := NewLimiter(config, logger)

	ip := "192.168.1.60"

	// First 10 downloads should succeed (burst)
	for i := 0; i < 10; i++ {
		allowed, err := limiter.AllowDownload(ip, "")
		if !allowed || err != nil {
			t.Errorf("download %d should be allowed", i)
		}
	}

	// 11th download should fail
	allowed, err := limiter.AllowDownload(ip, "")
	if allowed {
		t.Error("download should be rate limited")
	}
	if err == nil {
		t.Error("expected error for download rate limit")
	}
}

func TestLimiter_GetStats(t *testing.T) {
	logger := zap.NewNop()
	config := &LimiterConfig{
		WhitelistedIPs: []string{"127.0.0.1"},
		BlacklistedIPs: []string{"10.0.0.1"},
	}
	limiter := NewLimiter(config, logger)

	// Make some requests to create buckets
	limiter.Allow("192.168.1.1", "user1")
	limiter.Allow("192.168.1.2", "user2")

	stats := limiter.GetStats()

	if stats.IPBuckets < 2 {
		t.Errorf("expected at least 2 IP buckets, got %d", stats.IPBuckets)
	}

	if stats.UserBuckets < 2 {
		t.Errorf("expected at least 2 user buckets, got %d", stats.UserBuckets)
	}

	if stats.WhitelistedIPs != 1 {
		t.Errorf("expected 1 whitelisted IP, got %d", stats.WhitelistedIPs)
	}

	if stats.BlacklistedIPs != 1 {
		t.Errorf("expected 1 blacklisted IP, got %d", stats.BlacklistedIPs)
	}
}

func BenchmarkLimiter_Allow(b *testing.B) {
	logger := zap.NewNop()
	limiter := NewLimiter(nil, logger)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		limiter.Allow("192.168.1.1", "user1")
	}
}

func BenchmarkTokenBucket_Allow(b *testing.B) {
	bucket := NewTokenBucket(1000, 1000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bucket.Allow()
	}
}
