package cache

import (
	"context"
	"encoding/json"
	"sync"
	"time"
)

// Cache defines the caching interface
type Cache interface {
	// Get retrieves a value from cache
	Get(ctx context.Context, key string) (interface{}, bool)

	// Set stores a value in cache with expiration
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error

	// Delete removes a value from cache
	Delete(ctx context.Context, key string) error

	// Clear removes all values from cache
	Clear(ctx context.Context) error

	// GetStats returns cache statistics
	GetStats() *CacheStats

	// Close closes the cache connection
	Close() error
}

// CacheStats contains cache performance metrics
type CacheStats struct {
	Hits           int64         // Number of cache hits
	Misses         int64         // Number of cache misses
	Sets           int64         // Number of set operations
	Deletes        int64         // Number of delete operations
	Evictions      int64         // Number of evicted entries
	Size           int           // Current number of entries
	AvgGetDuration time.Duration // Average get operation duration
	AvgSetDuration time.Duration // Average set operation duration
	HitRate        float64       // Cache hit rate (0.0 - 1.0)
}

// cacheEntry represents a single cache entry
type cacheEntry struct {
	value      interface{}
	expiration time.Time
	size       int // Approximate size in bytes
}

// isExpired checks if entry has expired
func (e *cacheEntry) isExpired() bool {
	return time.Now().After(e.expiration)
}

// inMemoryCache is a high-performance in-memory cache
type inMemoryCache struct {
	entries       map[string]*cacheEntry
	mu            sync.RWMutex

	// Statistics
	hits          int64
	misses        int64
	sets          int64
	deletes       int64
	evictions     int64
	totalGetTime  time.Duration
	totalSetTime  time.Duration
	statsMu       sync.RWMutex

	// Configuration
	maxSize       int           // Maximum number of entries
	maxMemory     int64         // Maximum memory in bytes
	currentMemory int64         // Current memory usage
	defaultTTL    time.Duration // Default time-to-live

	// Cleanup
	cleanupInterval time.Duration
	stopCleanup     chan struct{}
	cleanupDone     sync.WaitGroup
}

// CacheConfig contains cache configuration
type CacheConfig struct {
	MaxSize         int           // Maximum entries (0 = unlimited)
	MaxMemory       int64         // Maximum memory in bytes (0 = unlimited)
	DefaultTTL      time.Duration // Default expiration time
	CleanupInterval time.Duration // Cleanup interval
}

// DefaultCacheConfig returns optimized default settings
func DefaultCacheConfig() CacheConfig {
	return CacheConfig{
		MaxSize:         10000,              // 10k entries
		MaxMemory:       256 * 1024 * 1024,  // 256MB
		DefaultTTL:      5 * time.Minute,    // 5 minute default
		CleanupInterval: 1 * time.Minute,    // Cleanup every minute
	}
}

// NewInMemoryCache creates a new high-performance in-memory cache
func NewInMemoryCache(cfg CacheConfig) Cache {
	c := &inMemoryCache{
		entries:         make(map[string]*cacheEntry),
		maxSize:         cfg.MaxSize,
		maxMemory:       cfg.MaxMemory,
		defaultTTL:      cfg.DefaultTTL,
		cleanupInterval: cfg.CleanupInterval,
		stopCleanup:     make(chan struct{}),
	}

	// Start background cleanup
	c.cleanupDone.Add(1)
	go c.cleanupLoop()

	return c
}

// Get retrieves a value from cache
func (c *inMemoryCache) Get(ctx context.Context, key string) (interface{}, bool) {
	start := time.Now()
	defer c.trackGetTime(time.Since(start))

	c.mu.RLock()
	entry, exists := c.entries[key]
	c.mu.RUnlock()

	if !exists {
		c.incrementMisses()
		return nil, false
	}

	if entry.isExpired() {
		// Remove expired entry
		c.Delete(ctx, key)
		c.incrementMisses()
		return nil, false
	}

	c.incrementHits()
	return entry.value, true
}

// Set stores a value in cache with expiration
func (c *inMemoryCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	start := time.Now()
	defer c.trackSetTime(time.Since(start))

	if expiration == 0 {
		expiration = c.defaultTTL
	}

	// Estimate entry size
	size := c.estimateSize(value)

	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if we need to evict
	if c.maxSize > 0 && len(c.entries) >= c.maxSize {
		c.evictOldest()
	}

	// Check memory limit
	if c.maxMemory > 0 {
		for c.currentMemory+int64(size) > c.maxMemory && len(c.entries) > 0 {
			c.evictOldest()
		}
	}

	// Remove old entry size if exists
	if oldEntry, exists := c.entries[key]; exists {
		c.currentMemory -= int64(oldEntry.size)
	}

	// Add new entry
	c.entries[key] = &cacheEntry{
		value:      value,
		expiration: time.Now().Add(expiration),
		size:       size,
	}
	c.currentMemory += int64(size)

	c.incrementSets()
	return nil
}

// Delete removes a value from cache
func (c *inMemoryCache) Delete(ctx context.Context, key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if entry, exists := c.entries[key]; exists {
		c.currentMemory -= int64(entry.size)
		delete(c.entries, key)
		c.incrementDeletes()
	}

	return nil
}

// Clear removes all values from cache
func (c *inMemoryCache) Clear(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries = make(map[string]*cacheEntry)
	c.currentMemory = 0

	return nil
}

// GetStats returns cache statistics
func (c *inMemoryCache) GetStats() *CacheStats {
	c.statsMu.RLock()
	hits := c.hits
	misses := c.misses
	sets := c.sets
	deletes := c.deletes
	evictions := c.evictions
	totalGetTime := c.totalGetTime
	totalSetTime := c.totalSetTime
	c.statsMu.RUnlock()

	c.mu.RLock()
	size := len(c.entries)
	c.mu.RUnlock()

	var avgGetDuration, avgSetDuration time.Duration
	var hitRate float64

	totalRequests := hits + misses
	if totalRequests > 0 {
		hitRate = float64(hits) / float64(totalRequests)
	}

	if hits > 0 {
		avgGetDuration = totalGetTime / time.Duration(hits)
	}

	if sets > 0 {
		avgSetDuration = totalSetTime / time.Duration(sets)
	}

	return &CacheStats{
		Hits:           hits,
		Misses:         misses,
		Sets:           sets,
		Deletes:        deletes,
		Evictions:      evictions,
		Size:           size,
		AvgGetDuration: avgGetDuration,
		AvgSetDuration: avgSetDuration,
		HitRate:        hitRate,
	}
}

// Close closes the cache and stops background cleanup
func (c *inMemoryCache) Close() error {
	close(c.stopCleanup)
	c.cleanupDone.Wait()
	return nil
}

// cleanupLoop runs background cleanup of expired entries
func (c *inMemoryCache) cleanupLoop() {
	defer c.cleanupDone.Done()

	ticker := time.NewTicker(c.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.cleanup()
		case <-c.stopCleanup:
			return
		}
	}
}

// cleanup removes expired entries
func (c *inMemoryCache) cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, entry := range c.entries {
		if now.After(entry.expiration) {
			c.currentMemory -= int64(entry.size)
			delete(c.entries, key)
			c.incrementEvictions()
		}
	}
}

// evictOldest removes the oldest entry (called with lock held)
func (c *inMemoryCache) evictOldest() {
	var oldestKey string
	var oldestTime time.Time

	for key, entry := range c.entries {
		if oldestTime.IsZero() || entry.expiration.Before(oldestTime) {
			oldestKey = key
			oldestTime = entry.expiration
		}
	}

	if oldestKey != "" {
		c.currentMemory -= int64(c.entries[oldestKey].size)
		delete(c.entries, oldestKey)
		c.incrementEvictions()
	}
}

// estimateSize estimates the size of a value in bytes
func (c *inMemoryCache) estimateSize(value interface{}) int {
	// Try JSON serialization for size estimation
	if data, err := json.Marshal(value); err == nil {
		return len(data)
	}

	// Fallback to rough estimate
	return 100 // Default size estimate
}

// Statistics tracking methods
func (c *inMemoryCache) incrementHits() {
	c.statsMu.Lock()
	c.hits++
	c.statsMu.Unlock()
}

func (c *inMemoryCache) incrementMisses() {
	c.statsMu.Lock()
	c.misses++
	c.statsMu.Unlock()
}

func (c *inMemoryCache) incrementSets() {
	c.statsMu.Lock()
	c.sets++
	c.statsMu.Unlock()
}

func (c *inMemoryCache) incrementDeletes() {
	c.statsMu.Lock()
	c.deletes++
	c.statsMu.Unlock()
}

func (c *inMemoryCache) incrementEvictions() {
	c.statsMu.Lock()
	c.evictions++
	c.statsMu.Unlock()
}

func (c *inMemoryCache) trackGetTime(duration time.Duration) {
	c.statsMu.Lock()
	c.totalGetTime += duration
	c.statsMu.Unlock()
}

func (c *inMemoryCache) trackSetTime(duration time.Duration) {
	c.statsMu.Lock()
	c.totalSetTime += duration
	c.statsMu.Unlock()
}

// BuildCacheKey builds a cache key from components
func BuildCacheKey(components ...string) string {
	key := ""
	for i, component := range components {
		if i > 0 {
			key += ":"
		}
		key += component
	}
	return key
}

// CachedQuery executes a query with caching
func CachedQuery[T any](
	ctx context.Context,
	cache Cache,
	key string,
	ttl time.Duration,
	queryFunc func(ctx context.Context) (T, error),
) (T, error) {
	var zero T

	// Try cache first
	if cached, found := cache.Get(ctx, key); found {
		if result, ok := cached.(T); ok {
			return result, nil
		}
	}

	// Execute query
	result, err := queryFunc(ctx)
	if err != nil {
		return zero, err
	}

	// Store in cache
	_ = cache.Set(ctx, key, result, ttl)

	return result, nil
}

// InvalidatePattern invalidates all cache entries matching a pattern
func InvalidatePattern(ctx context.Context, cache Cache, pattern string) error {
	// Note: This is a simple implementation
	// For production, consider using a cache that supports pattern matching
	return cache.Clear(ctx)
}
