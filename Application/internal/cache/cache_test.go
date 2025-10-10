package cache

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewInMemoryCache(t *testing.T) {
	cfg := DefaultCacheConfig()
	cache := NewInMemoryCache(cfg)
	assert.NotNil(t, cache)

	// Cleanup
	err := cache.Close()
	assert.NoError(t, err)
}

func TestCache_SetAndGet(t *testing.T) {
	cfg := DefaultCacheConfig()
	cache := NewInMemoryCache(cfg)
	defer cache.Close()

	ctx := context.Background()

	// Set a value
	err := cache.Set(ctx, "key1", "value1", 1*time.Minute)
	assert.NoError(t, err)

	// Get the value
	value, found := cache.Get(ctx, "key1")
	assert.True(t, found)
	assert.Equal(t, "value1", value)
}

func TestCache_GetNonExistent(t *testing.T) {
	cfg := DefaultCacheConfig()
	cache := NewInMemoryCache(cfg)
	defer cache.Close()

	ctx := context.Background()

	// Get non-existent key
	value, found := cache.Get(ctx, "nonexistent")
	assert.False(t, found)
	assert.Nil(t, value)
}

func TestCache_Expiration(t *testing.T) {
	cfg := DefaultCacheConfig()
	cfg.DefaultTTL = 100 * time.Millisecond
	cache := NewInMemoryCache(cfg)
	defer cache.Close()

	ctx := context.Background()

	// Set a value with short expiration
	err := cache.Set(ctx, "key1", "value1", 100*time.Millisecond)
	assert.NoError(t, err)

	// Value should exist immediately
	value, found := cache.Get(ctx, "key1")
	assert.True(t, found)
	assert.Equal(t, "value1", value)

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Value should be expired
	value, found = cache.Get(ctx, "key1")
	assert.False(t, found)
	assert.Nil(t, value)
}

func TestCache_Delete(t *testing.T) {
	cfg := DefaultCacheConfig()
	cache := NewInMemoryCache(cfg)
	defer cache.Close()

	ctx := context.Background()

	// Set and delete
	err := cache.Set(ctx, "key1", "value1", 1*time.Minute)
	assert.NoError(t, err)

	err = cache.Delete(ctx, "key1")
	assert.NoError(t, err)

	// Value should not exist
	value, found := cache.Get(ctx, "key1")
	assert.False(t, found)
	assert.Nil(t, value)
}

func TestCache_Clear(t *testing.T) {
	cfg := DefaultCacheConfig()
	cache := NewInMemoryCache(cfg)
	defer cache.Close()

	ctx := context.Background()

	// Set multiple values
	cache.Set(ctx, "key1", "value1", 1*time.Minute)
	cache.Set(ctx, "key2", "value2", 1*time.Minute)
	cache.Set(ctx, "key3", "value3", 1*time.Minute)

	// Clear cache
	err := cache.Clear(ctx)
	assert.NoError(t, err)

	// All values should be gone
	_, found := cache.Get(ctx, "key1")
	assert.False(t, found)
	_, found = cache.Get(ctx, "key2")
	assert.False(t, found)
	_, found = cache.Get(ctx, "key3")
	assert.False(t, found)
}

func TestCache_MaxSize(t *testing.T) {
	cfg := DefaultCacheConfig()
	cfg.MaxSize = 3
	cache := NewInMemoryCache(cfg)
	defer cache.Close()

	ctx := context.Background()

	// Add more items than max size
	cache.Set(ctx, "key1", "value1", 1*time.Minute)
	cache.Set(ctx, "key2", "value2", 1*time.Minute)
	cache.Set(ctx, "key3", "value3", 1*time.Minute)
	cache.Set(ctx, "key4", "value4", 1*time.Minute)

	// Check stats
	stats := cache.GetStats()
	assert.LessOrEqual(t, stats.Size, 3)
	assert.Greater(t, stats.Evictions, int64(0))
}

func TestCache_Statistics(t *testing.T) {
	cfg := DefaultCacheConfig()
	cache := NewInMemoryCache(cfg)
	defer cache.Close()

	ctx := context.Background()

	// Set some values
	cache.Set(ctx, "key1", "value1", 1*time.Minute)
	cache.Set(ctx, "key2", "value2", 1*time.Minute)

	// Create hits
	cache.Get(ctx, "key1")
	cache.Get(ctx, "key1")

	// Create misses
	cache.Get(ctx, "nonexistent1")
	cache.Get(ctx, "nonexistent2")

	// Check stats
	stats := cache.GetStats()
	assert.Equal(t, int64(2), stats.Hits)
	assert.Equal(t, int64(2), stats.Misses)
	assert.Equal(t, int64(2), stats.Sets)
	assert.Equal(t, 0.5, stats.HitRate)
	assert.Greater(t, stats.AvgGetDuration, time.Duration(0))
	assert.Greater(t, stats.AvgSetDuration, time.Duration(0))
}

func TestCache_ComplexTypes(t *testing.T) {
	cfg := DefaultCacheConfig()
	cache := NewInMemoryCache(cfg)
	defer cache.Close()

	ctx := context.Background()

	// Test with struct
	type TestStruct struct {
		Name  string
		Value int
	}

	testData := TestStruct{
		Name:  "test",
		Value: 42,
	}

	err := cache.Set(ctx, "struct", testData, 1*time.Minute)
	assert.NoError(t, err)

	value, found := cache.Get(ctx, "struct")
	assert.True(t, found)
	assert.Equal(t, testData, value)

	// Test with slice
	testSlice := []string{"a", "b", "c"}
	err = cache.Set(ctx, "slice", testSlice, 1*time.Minute)
	assert.NoError(t, err)

	value, found = cache.Get(ctx, "slice")
	assert.True(t, found)
	assert.Equal(t, testSlice, value)

	// Test with map
	testMap := map[string]int{"one": 1, "two": 2}
	err = cache.Set(ctx, "map", testMap, 1*time.Minute)
	assert.NoError(t, err)

	value, found = cache.Get(ctx, "map")
	assert.True(t, found)
	assert.Equal(t, testMap, value)
}

func TestCache_Cleanup(t *testing.T) {
	cfg := DefaultCacheConfig()
	cfg.CleanupInterval = 100 * time.Millisecond
	cache := NewInMemoryCache(cfg)
	defer cache.Close()

	ctx := context.Background()

	// Set values with short expiration
	cache.Set(ctx, "key1", "value1", 50*time.Millisecond)
	cache.Set(ctx, "key2", "value2", 50*time.Millisecond)

	// Wait for cleanup to run
	time.Sleep(200 * time.Millisecond)

	// Values should be cleaned up
	stats := cache.GetStats()
	assert.Greater(t, stats.Evictions, int64(0))
}

func TestBuildCacheKey(t *testing.T) {
	tests := []struct {
		name       string
		components []string
		expected   string
	}{
		{
			name:       "Single component",
			components: []string{"key1"},
			expected:   "key1",
		},
		{
			name:       "Multiple components",
			components: []string{"user", "123", "tickets"},
			expected:   "user:123:tickets",
		},
		{
			name:       "Empty components",
			components: []string{},
			expected:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildCacheKey(tt.components...)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCachedQuery(t *testing.T) {
	cfg := DefaultCacheConfig()
	cache := NewInMemoryCache(cfg)
	defer cache.Close()

	ctx := context.Background()

	callCount := 0
	queryFunc := func(ctx context.Context) (string, error) {
		callCount++
		return "result", nil
	}

	// First call - should execute query
	result, err := CachedQuery(ctx, cache, "test-key", 1*time.Minute, queryFunc)
	assert.NoError(t, err)
	assert.Equal(t, "result", result)
	assert.Equal(t, 1, callCount)

	// Second call - should use cache
	result, err = CachedQuery(ctx, cache, "test-key", 1*time.Minute, queryFunc)
	assert.NoError(t, err)
	assert.Equal(t, "result", result)
	assert.Equal(t, 1, callCount) // Call count should not increase
}

func TestCachedQuery_Error(t *testing.T) {
	cfg := DefaultCacheConfig()
	cache := NewInMemoryCache(cfg)
	defer cache.Close()

	ctx := context.Background()

	queryFunc := func(ctx context.Context) (string, error) {
		return "", assert.AnError
	}

	// Query with error should not cache
	result, err := CachedQuery(ctx, cache, "test-key", 1*time.Minute, queryFunc)
	assert.Error(t, err)
	assert.Equal(t, "", result)

	// Value should not be in cache
	_, found := cache.Get(ctx, "test-key")
	assert.False(t, found)
}

func TestCache_ConcurrentAccess(t *testing.T) {
	cfg := DefaultCacheConfig()
	cache := NewInMemoryCache(cfg)
	defer cache.Close()

	ctx := context.Background()

	// Test concurrent writes
	done := make(chan bool)
	for i := 0; i < 100; i++ {
		go func(n int) {
			cache.Set(ctx, fmt.Sprintf("key%d", n), n, 1*time.Minute)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 100; i++ {
		<-done
	}

	// Test concurrent reads
	for i := 0; i < 100; i++ {
		go func(n int) {
			cache.Get(ctx, fmt.Sprintf("key%d", n))
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 100; i++ {
		<-done
	}

	// Verify no panics occurred
	stats := cache.GetStats()
	assert.Greater(t, stats.Sets, int64(0))
}

// Benchmark tests
func BenchmarkCache_Set(b *testing.B) {
	cfg := DefaultCacheConfig()
	cache := NewInMemoryCache(cfg)
	defer cache.Close()

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Set(ctx, fmt.Sprintf("key%d", i), i, 1*time.Minute)
	}
}

func BenchmarkCache_Get(b *testing.B) {
	cfg := DefaultCacheConfig()
	cache := NewInMemoryCache(cfg)
	defer cache.Close()

	ctx := context.Background()

	// Pre-populate cache
	for i := 0; i < 1000; i++ {
		cache.Set(ctx, fmt.Sprintf("key%d", i), i, 1*time.Minute)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Get(ctx, fmt.Sprintf("key%d", i%1000))
	}
}

func BenchmarkCache_SetGet(b *testing.B) {
	cfg := DefaultCacheConfig()
	cache := NewInMemoryCache(cfg)
	defer cache.Close()

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key%d", i%100)
		cache.Set(ctx, key, i, 1*time.Minute)
		cache.Get(ctx, key)
	}
}
