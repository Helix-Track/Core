package cache

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNewMemoryCache(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	mc := NewMemoryCache(100, 1*time.Hour, 5*time.Minute, logger)

	assert.NotNil(t, mc)
	assert.Equal(t, int64(100*1024*1024), mc.maxSizeBytes)
	assert.Equal(t, 1*time.Hour, mc.defaultTTL)
}

func TestMemoryCache_SetAndGet(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	mc := NewMemoryCache(100, 1*time.Hour, 5*time.Minute, logger)
	ctx := context.Background()

	// Set a value
	err := mc.Set(ctx, "test_key", "test_value", 0)
	assert.NoError(t, err)

	// Get the value
	value, err := mc.Get(ctx, "test_key")
	assert.NoError(t, err)
	assert.Equal(t, "test_value", value)
}

func TestMemoryCache_GetNonExistent(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	mc := NewMemoryCache(100, 1*time.Hour, 5*time.Minute, logger)
	ctx := context.Background()

	value, err := mc.Get(ctx, "non_existent")
	assert.Error(t, err)
	assert.Equal(t, ErrCacheMiss, err)
	assert.Empty(t, value)
}

func TestMemoryCache_SetWithTTL(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	mc := NewMemoryCache(100, 1*time.Hour, 5*time.Minute, logger)
	ctx := context.Background()

	// Set with short TTL
	err := mc.Set(ctx, "ttl_key", "ttl_value", 100*time.Millisecond)
	assert.NoError(t, err)

	// Get immediately - should work
	value, err := mc.Get(ctx, "ttl_key")
	assert.NoError(t, err)
	assert.Equal(t, "ttl_value", value)

	// Wait for expiration
	time.Sleep(200 * time.Millisecond)

	// Get after expiration - should fail
	value, err = mc.Get(ctx, "ttl_key")
	assert.Error(t, err)
	assert.Equal(t, ErrCacheMiss, err)
	assert.Empty(t, value)
}

func TestMemoryCache_Delete(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	mc := NewMemoryCache(100, 1*time.Hour, 5*time.Minute, logger)
	ctx := context.Background()

	// Set a value
	mc.Set(ctx, "delete_key", "delete_value", 0)

	// Verify it exists
	value, err := mc.Get(ctx, "delete_key")
	assert.NoError(t, err)
	assert.Equal(t, "delete_value", value)

	// Delete it
	err = mc.Delete(ctx, "delete_key")
	assert.NoError(t, err)

	// Verify it's gone
	value, err = mc.Get(ctx, "delete_key")
	assert.Error(t, err)
	assert.Equal(t, ErrCacheMiss, err)
}

func TestMemoryCache_DeletePattern(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	mc := NewMemoryCache(100, 1*time.Hour, 5*time.Minute, logger)
	ctx := context.Background()

	// Set multiple values
	mc.Set(ctx, "user:1:profile", "profile1", 0)
	mc.Set(ctx, "user:2:profile", "profile2", 0)
	mc.Set(ctx, "user:1:settings", "settings1", 0)
	mc.Set(ctx, "post:1:content", "content1", 0)

	// Delete user:* pattern
	err := mc.DeletePattern(ctx, "user:*")
	assert.NoError(t, err)

	// Verify user keys are gone
	_, err = mc.Get(ctx, "user:1:profile")
	assert.Error(t, err)
	_, err = mc.Get(ctx, "user:2:profile")
	assert.Error(t, err)
	_, err = mc.Get(ctx, "user:1:settings")
	assert.Error(t, err)

	// Verify post key still exists
	value, err := mc.Get(ctx, "post:1:content")
	assert.NoError(t, err)
	assert.Equal(t, "content1", value)
}

func TestMemoryCache_Exists(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	mc := NewMemoryCache(100, 1*time.Hour, 5*time.Minute, logger)
	ctx := context.Background()

	// Check non-existent key
	exists, err := mc.Exists(ctx, "non_existent")
	assert.NoError(t, err)
	assert.False(t, exists)

	// Set a key
	mc.Set(ctx, "exists_key", "exists_value", 0)

	// Check existing key
	exists, err = mc.Exists(ctx, "exists_key")
	assert.NoError(t, err)
	assert.True(t, exists)

	// Set with short TTL and check after expiration
	mc.Set(ctx, "ttl_exists", "value", 100*time.Millisecond)
	time.Sleep(200 * time.Millisecond)
	exists, err = mc.Exists(ctx, "ttl_exists")
	assert.NoError(t, err)
	assert.False(t, exists)
}

func TestMemoryCache_Update(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	mc := NewMemoryCache(100, 1*time.Hour, 5*time.Minute, logger)
	ctx := context.Background()

	// Set initial value
	mc.Set(ctx, "update_key", "initial_value", 0)

	// Update value
	mc.Set(ctx, "update_key", "updated_value", 0)

	// Verify updated value
	value, err := mc.Get(ctx, "update_key")
	assert.NoError(t, err)
	assert.Equal(t, "updated_value", value)
}

func TestMemoryCache_Stats(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	mc := NewMemoryCache(100, 1*time.Hour, 5*time.Minute, logger)
	ctx := context.Background()

	// Set some values
	mc.Set(ctx, "key1", "value1", 0)
	mc.Set(ctx, "key2", "value2value2", 0) // Longer value

	stats := mc.Stats()

	assert.Equal(t, 2, stats["entries"])
	assert.Greater(t, stats["size_bytes"].(int64), int64(0))
	assert.Less(t, stats["utilization_%"].(float64), 100.0)
}

func TestMemoryCache_Close(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	mc := NewMemoryCache(100, 1*time.Hour, 5*time.Minute, logger)
	ctx := context.Background()

	// Set some values
	mc.Set(ctx, "key1", "value1", 0)

	// Close cache
	err := mc.Close()
	assert.NoError(t, err)

	// After close, cache should be empty
	stats := mc.Stats()
	assert.Equal(t, 0, stats["entries"])
	assert.Equal(t, int64(0), stats["size_bytes"])
}

func TestMatchPattern(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		str      string
		expected bool
	}{
		{
			name:     "exact match",
			pattern:  "test",
			str:      "test",
			expected: true,
		},
		{
			name:     "no match",
			pattern:  "test",
			str:      "other",
			expected: false,
		},
		{
			name:     "wildcard all",
			pattern:  "*",
			str:      "anything",
			expected: true,
		},
		{
			name:     "prefix wildcard",
			pattern:  "user:*",
			str:      "user:123",
			expected: true,
		},
		{
			name:     "suffix wildcard",
			pattern:  "*:profile",
			str:      "user:profile",
			expected: true,
		},
		{
			name:     "middle wildcard",
			pattern:  "user:*:profile",
			str:      "user:123:profile",
			expected: true,
		},
		{
			name:     "prefix mismatch",
			pattern:  "user:*",
			str:      "post:123",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchPattern(tt.pattern, tt.str)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMemoryCache_Cleanup(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	mc := NewMemoryCache(100, 1*time.Hour, 100*time.Millisecond, logger) // Short cleanup interval
	ctx := context.Background()

	// Set value with short TTL
	mc.Set(ctx, "cleanup_key", "cleanup_value", 50*time.Millisecond)

	// Verify it exists
	value, err := mc.Get(ctx, "cleanup_key")
	assert.NoError(t, err)
	assert.Equal(t, "cleanup_value", value)

	// Wait for cleanup to run
	time.Sleep(200 * time.Millisecond)

	// Verify entry was cleaned up
	value, err = mc.Get(ctx, "cleanup_key")
	assert.Error(t, err)
	assert.Equal(t, ErrCacheMiss, err)

	// Close to stop cleanup goroutine
	mc.Close()
}

func TestCacheKey(t *testing.T) {
	tests := []struct {
		name     string
		parts    []string
		expected string
	}{
		{
			name:     "single part",
			parts:    []string{"catalog"},
			expected: "l10n:catalog",
		},
		{
			name:     "multiple parts",
			parts:    []string{"catalog", "en", "error"},
			expected: "l10n:catalog:en:error",
		},
		{
			name:     "with empty parts",
			parts:    []string{"catalog", "", "en"},
			expected: "l10n:catalog:en",
		},
		{
			name:     "no parts",
			parts:    []string{},
			expected: "l10n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CacheKey(tt.parts...)
			assert.Equal(t, tt.expected, result)
		})
	}
}
