package cache

import (
	"context"
	"time"
)

// Cache interface defines caching operations
type Cache interface {
	// Get retrieves a value from cache
	Get(ctx context.Context, key string) (string, error)

	// Set stores a value in cache with TTL
	Set(ctx context.Context, key string, value string, ttl time.Duration) error

	// Delete removes a value from cache
	Delete(ctx context.Context, key string) error

	// DeletePattern removes all keys matching a pattern
	DeletePattern(ctx context.Context, pattern string) error

	// Exists checks if a key exists in cache
	Exists(ctx context.Context, key string) (bool, error)

	// Close closes the cache connection
	Close() error
}

// CacheKey generates a cache key
func CacheKey(parts ...string) string {
	result := "l10n"
	for _, part := range parts {
		if part != "" {
			result += ":" + part
		}
	}
	return result
}
