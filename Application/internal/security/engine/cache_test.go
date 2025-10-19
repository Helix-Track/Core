package engine

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestNewPermissionCache tests cache creation
func TestNewPermissionCache(t *testing.T) {
	cache := NewPermissionCache(1000, 5*time.Minute)

	assert.NotNil(t, cache)
	assert.Equal(t, 1000, cache.maxSize)
	assert.Equal(t, 5*time.Minute, cache.defaultTTL)
	assert.NotNil(t, cache.entries)
	assert.NotNil(t, cache.contexts)
}

// TestPermissionCache_SetAndGet tests basic cache operations
func TestPermissionCache_SetAndGet(t *testing.T) {
	cache := NewPermissionCache(100, 5*time.Minute)

	req := AccessRequest{
		Username: "testuser",
		Resource: "ticket",
		Action:   ActionRead,
	}

	response := AccessResponse{
		Allowed: true,
		Reason:  "Access granted",
	}

	// Set cache entry
	cache.Set(req, response)

	// Get cache entry
	cached, found := cache.Get(req)

	assert.True(t, found)
	assert.Equal(t, response.Allowed, cached.Allowed)
	assert.Equal(t, response.Reason, cached.Reason)
}

// TestPermissionCache_CacheMiss tests cache miss
func TestPermissionCache_CacheMiss(t *testing.T) {
	cache := NewPermissionCache(100, 5*time.Minute)

	req := AccessRequest{
		Username: "testuser",
		Resource: "ticket",
		Action:   ActionRead,
	}

	// Try to get non-existent entry
	_, found := cache.Get(req)

	assert.False(t, found)
}

// TestPermissionCache_Expiration tests cache expiration
func TestPermissionCache_Expiration(t *testing.T) {
	cache := NewPermissionCache(100, 1*time.Millisecond)

	req := AccessRequest{
		Username: "testuser",
		Resource: "ticket",
		Action:   ActionRead,
	}

	response := AccessResponse{
		Allowed: true,
	}

	// Set cache entry with very short TTL
	cache.Set(req, response)

	// Verify it exists immediately
	_, found := cache.Get(req)
	assert.True(t, found)

	// Wait for expiration
	time.Sleep(10 * time.Millisecond)

	// Verify it's expired
	_, found = cache.Get(req)
	assert.False(t, found)
}

// TestPermissionCache_SetWithTTL tests custom TTL
func TestPermissionCache_SetWithTTL(t *testing.T) {
	cache := NewPermissionCache(100, 5*time.Minute)

	req := AccessRequest{
		Username: "testuser",
		Resource: "ticket",
		Action:   ActionRead,
	}

	response := AccessResponse{
		Allowed: true,
	}

	// Set with custom short TTL
	cache.SetWithTTL(req, response, 1*time.Millisecond)

	// Verify it exists
	_, found := cache.Get(req)
	assert.True(t, found)

	// Wait for expiration
	time.Sleep(10 * time.Millisecond)

	// Verify it's expired
	_, found = cache.Get(req)
	assert.False(t, found)
}

// TestPermissionCache_MultipleEntries tests multiple cache entries
func TestPermissionCache_MultipleEntries(t *testing.T) {
	cache := NewPermissionCache(100, 5*time.Minute)

	for i := 0; i < 10; i++ {
		req := AccessRequest{
			Username:   "testuser",
			Resource:   "ticket",
			ResourceID: string(rune('A' + i)),
			Action:     ActionRead,
		}

		response := AccessResponse{
			Allowed: i%2 == 0, // Alternate allowed/denied
		}

		cache.Set(req, response)
	}

	stats := cache.GetStats()
	assert.Equal(t, 10, stats.EntryCount)
}

// TestPermissionCache_Eviction tests cache eviction when full
func TestPermissionCache_Eviction(t *testing.T) {
	cache := NewPermissionCache(5, 5*time.Minute) // Small cache

	// Fill cache beyond capacity
	for i := 0; i < 10; i++ {
		req := AccessRequest{
			Username:   "testuser",
			Resource:   "ticket",
			ResourceID: string(rune('A' + i)),
			Action:     ActionRead,
		}

		response := AccessResponse{
			Allowed: true,
		}

		cache.Set(req, response)
	}

	stats := cache.GetStats()
	assert.LessOrEqual(t, stats.EntryCount, 5)
	assert.Greater(t, stats.EvictCount, uint64(0))
}

// TestPermissionCache_InvalidateUser tests user-specific invalidation
func TestPermissionCache_InvalidateUser(t *testing.T) {
	cache := NewPermissionCache(100, 5*time.Minute)

	// Add entries for multiple users
	users := []string{"user1", "user2", "user3"}
	for _, username := range users {
		req := AccessRequest{
			Username: username,
			Resource: "ticket",
			Action:   ActionRead,
		}

		response := AccessResponse{
			Allowed: true,
		}

		cache.Set(req, response)
	}

	// Invalidate user1
	cache.InvalidateUser("user1")

	// Verify user1's entry is gone
	req1 := AccessRequest{
		Username: "user1",
		Resource: "ticket",
		Action:   ActionRead,
	}
	_, found := cache.Get(req1)
	assert.False(t, found)

	// Verify other users' entries still exist
	req2 := AccessRequest{
		Username: "user2",
		Resource: "ticket",
		Action:   ActionRead,
	}
	_, found = cache.Get(req2)
	assert.True(t, found)
}

// TestPermissionCache_Clear tests clearing entire cache
func TestPermissionCache_Clear(t *testing.T) {
	cache := NewPermissionCache(100, 5*time.Minute)

	// Add multiple entries
	for i := 0; i < 10; i++ {
		req := AccessRequest{
			Username:   "testuser",
			Resource:   "ticket",
			ResourceID: string(rune('A' + i)),
			Action:     ActionRead,
		}

		response := AccessResponse{
			Allowed: true,
		}

		cache.Set(req, response)
	}

	// Verify cache has entries
	stats := cache.GetStats()
	assert.Equal(t, 10, stats.EntryCount)

	// Clear cache
	cache.Clear()

	// Verify cache is empty
	stats = cache.GetStats()
	assert.Equal(t, 0, stats.EntryCount)
}

// TestPermissionCache_SecurityContext tests security context caching
func TestPermissionCache_SecurityContext(t *testing.T) {
	cache := NewPermissionCache(100, 5*time.Minute)

	ctx := &SecurityContext{
		Username:  "testuser",
		Roles:     []Role{{ID: "role1", Title: "Developer"}},
		Teams:     []string{"team1"},
		CachedAt:  time.Now(),
		ExpiresAt: time.Now().Add(5 * time.Minute),
	}

	// Set context
	cache.SetContext("testuser", ctx)

	// Get context
	retrieved, found := cache.GetContext("testuser")

	assert.True(t, found)
	assert.Equal(t, ctx.Username, retrieved.Username)
	assert.Len(t, retrieved.Roles, 1)
	assert.Len(t, retrieved.Teams, 1)
}

// TestPermissionCache_HitRate tests hit rate calculation
func TestPermissionCache_HitRate(t *testing.T) {
	cache := NewPermissionCache(100, 5*time.Minute)

	req := AccessRequest{
		Username: "testuser",
		Resource: "ticket",
		Action:   ActionRead,
	}

	response := AccessResponse{
		Allowed: true,
	}

	// Initial hit rate should be 0
	assert.Equal(t, 0.0, cache.GetHitRate())

	// Add entry
	cache.Set(req, response)

	// First get - cache hit
	_, found := cache.Get(req)
	assert.True(t, found)

	// Hit rate should be 100%
	assert.Equal(t, 1.0, cache.GetHitRate())

	// Try to get non-existent entry - cache miss
	req2 := AccessRequest{
		Username: "otheruser",
		Resource: "ticket",
		Action:   ActionRead,
	}
	_, found = cache.Get(req2)
	assert.False(t, found)

	// Hit rate should be 50% (1 hit, 1 miss)
	assert.Equal(t, 0.5, cache.GetHitRate())
}

// TestPermissionCache_Stats tests statistics retrieval
func TestPermissionCache_Stats(t *testing.T) {
	cache := NewPermissionCache(100, 5*time.Minute)

	// Add some entries
	for i := 0; i < 5; i++ {
		req := AccessRequest{
			Username:   "testuser",
			Resource:   "ticket",
			ResourceID: string(rune('A' + i)),
			Action:     ActionRead,
		}

		response := AccessResponse{
			Allowed: true,
		}

		cache.Set(req, response)
	}

	// Add a context
	ctx := &SecurityContext{
		Username: "testuser",
	}
	cache.SetContext("testuser", ctx)

	stats := cache.GetStats()

	assert.Equal(t, 5, stats.EntryCount)
	assert.Equal(t, 1, stats.ContextCount)
	assert.Equal(t, 100, stats.MaxSize)
}

// TestPermissionCache_ConcurrentAccess tests thread safety
func TestPermissionCache_ConcurrentAccess(t *testing.T) {
	cache := NewPermissionCache(1000, 5*time.Minute)

	// Run concurrent operations
	done := make(chan bool)

	// Writer goroutines
	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 100; j++ {
				req := AccessRequest{
					Username:   "testuser",
					Resource:   "ticket",
					ResourceID: string(rune(id*100 + j)),
					Action:     ActionRead,
				}

				response := AccessResponse{
					Allowed: true,
				}

				cache.Set(req, response)
			}
			done <- true
		}(i)
	}

	// Reader goroutines
	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 100; j++ {
				req := AccessRequest{
					Username:   "testuser",
					Resource:   "ticket",
					ResourceID: string(rune(id*100 + j)),
					Action:     ActionRead,
				}

				cache.Get(req)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 20; i++ {
		<-done
	}

	// No assertion - just verifying no race conditions
	assert.NotNil(t, cache)
}

// Benchmark tests
func BenchmarkCacheSet(b *testing.B) {
	cache := NewPermissionCache(10000, 5*time.Minute)

	req := AccessRequest{
		Username: "testuser",
		Resource: "ticket",
		Action:   ActionRead,
	}

	response := AccessResponse{
		Allowed: true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Set(req, response)
	}
}

func BenchmarkCacheGet_Hit(b *testing.B) {
	cache := NewPermissionCache(10000, 5*time.Minute)

	req := AccessRequest{
		Username: "testuser",
		Resource: "ticket",
		Action:   ActionRead,
	}

	response := AccessResponse{
		Allowed: true,
	}

	cache.Set(req, response)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Get(req)
	}
}

func BenchmarkCacheGet_Miss(b *testing.B) {
	cache := NewPermissionCache(10000, 5*time.Minute)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := AccessRequest{
			Username:   "testuser",
			Resource:   "ticket",
			ResourceID: string(rune(i)),
			Action:     ActionRead,
		}
		cache.Get(req)
	}
}
