package cache

import (
	"container/list"
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

var (
	ErrCacheMiss = errors.New("cache miss")
	ErrCacheFull = errors.New("cache full")
)

// cacheEntry represents a single cache entry
type cacheEntry struct {
	value      string
	expiration int64
	size       int
	element    *list.Element // For LRU tracking
}

// lruItem represents an item in the LRU list
type lruItem struct {
	key   string
	entry *cacheEntry
}

// MemoryCache implements in-memory LRU cache
type MemoryCache struct {
	mu         sync.RWMutex
	entries    map[string]*cacheEntry
	lru        *list.List // Most recently used at front
	maxSizeBytes int64
	currentSize  int64
	defaultTTL   time.Duration
	logger       *zap.Logger
	cleanupTicker *time.Ticker
	done         chan struct{}
}

// NewMemoryCache creates a new in-memory cache with LRU eviction
func NewMemoryCache(maxSizeMB int, defaultTTL time.Duration, cleanupInterval time.Duration, logger *zap.Logger) *MemoryCache {
	mc := &MemoryCache{
		entries:      make(map[string]*cacheEntry),
		lru:           list.New(),
		maxSizeBytes: int64(maxSizeMB * 1024 * 1024),
		currentSize:   0,
		defaultTTL:    defaultTTL,
		logger:        logger,
		done:          make(chan struct{}),
	}

	// Start cleanup goroutine
	mc.cleanupTicker = time.NewTicker(cleanupInterval)
	go mc.cleanup()

	logger.Info("memory cache initialized",
		zap.Int("max_size_mb", maxSizeMB),
		zap.Duration("default_ttl", defaultTTL),
		zap.Duration("cleanup_interval", cleanupInterval),
	)

	return mc
}

// Get retrieves a value from cache and moves it to the front of LRU
func (mc *MemoryCache) Get(ctx context.Context, key string) (string, error) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	entry, exists := mc.entries[key]
	if !exists {
		return "", ErrCacheMiss
	}

	// Check expiration (using milliseconds for precision)
	if entry.expiration > 0 && time.Now().UnixMilli() > entry.expiration {
		// Entry expired, remove it
		mc.removeEntry(key, entry)
		return "", ErrCacheMiss
	}

	// Move to front of LRU (most recently used)
	if entry.element != nil {
		mc.lru.MoveToFront(entry.element)
	}

	return entry.value, nil
}

// Set stores a value in cache with TTL, evicting LRU entries if needed
func (mc *MemoryCache) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	valueSize := len(value)
	
	// Check if entry already exists
	if existing, exists := mc.entries[key]; exists {
		// Update existing entry
		mc.currentSize -= int64(existing.size)
		mc.lru.Remove(existing.element)
		delete(mc.entries, key)
	}

	// Make space if needed by evicting LRU entries
	for mc.currentSize+int64(valueSize) > mc.maxSizeBytes && mc.lru.Len() > 0 {
		mc.evictLRU()
	}

	// If still not enough space, reject the entry
	if mc.currentSize+int64(valueSize) > mc.maxSizeBytes {
		mc.logger.Warn("cache entry rejected - too large",
			zap.String("key", key),
			zap.Int("size", valueSize),
			zap.Int64("available", mc.maxSizeBytes-mc.currentSize),
		)
		return ErrCacheFull
	}

	expiration := int64(0)
	if ttl > 0 {
		expiration = time.Now().Add(ttl).UnixMilli()
	} else if mc.defaultTTL > 0 {
		expiration = time.Now().Add(mc.defaultTTL).UnixMilli()
	}

	// Create new entry
	entry := &cacheEntry{
		value:      value,
		expiration: expiration,
		size:       valueSize,
	}

	// Add to front of LRU
	item := &lruItem{
		key:   key,
		entry: entry,
	}
	entry.element = mc.lru.PushFront(item)
	mc.entries[key] = entry

	mc.currentSize += int64(valueSize)

	return nil
}

// Delete removes a value from cache
func (mc *MemoryCache) Delete(ctx context.Context, key string) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	if entry, exists := mc.entries[key]; exists {
		mc.removeEntry(key, entry)
	}

	return nil
}

// DeletePattern removes all keys matching a pattern
func (mc *MemoryCache) DeletePattern(ctx context.Context, pattern string) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	// Collect keys to remove (avoid modifying map while iterating)
	var keysToRemove []string
	for key := range mc.entries {
		if matchPattern(pattern, key) {
			keysToRemove = append(keysToRemove, key)
		}
	}

	// Remove collected keys
	for _, key := range keysToRemove {
		if entry, exists := mc.entries[key]; exists {
			mc.removeEntry(key, entry)
		}
	}

	return nil
}

// Exists checks if a key exists in cache and moves it to front of LRU
func (mc *MemoryCache) Exists(ctx context.Context, key string) (bool, error) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	entry, exists := mc.entries[key]
	if !exists {
		return false, nil
	}

	// Check expiration (using milliseconds for precision)
	if entry.expiration > 0 && time.Now().UnixMilli() > entry.expiration {
		// Entry expired, remove it
		mc.removeEntry(key, entry)
		return false, nil
	}

	// Move to front of LRU (most recently used)
	if entry.element != nil {
		mc.lru.MoveToFront(entry.element)
	}

	return true, nil
}

// Close closes the cache and stops cleanup goroutine
func (mc *MemoryCache) Close() error {
	close(mc.done)
	mc.cleanupTicker.Stop()

	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.entries = make(map[string]*cacheEntry)
	mc.lru = list.New()
	mc.currentSize = 0

	mc.logger.Info("memory cache closed")
	return nil
}

// cleanup removes expired entries periodically
func (mc *MemoryCache) cleanup() {
	for {
		select {
		case <-mc.cleanupTicker.C:
			mc.removeExpired()
		case <-mc.done:
			return
		}
	}
}

// removeExpired removes all expired entries
func (mc *MemoryCache) removeExpired() {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	now := time.Now().UnixMilli()
	removed := 0

	// Iterate through LRU list for expired entries
	for elem := mc.lru.Front(); elem != nil; {
		item := elem.Value.(*lruItem)
		next := elem.Next() // Get next before potentially removing current

		if item.entry.expiration > 0 && now > item.entry.expiration {
			mc.removeEntry(item.key, item.entry)
			removed++
		}

		elem = next
	}

	if removed > 0 {
		mc.logger.Debug("expired cache entries removed", zap.Int("count", removed))
	}
}

// evictLRU removes the least recently used entry
func (mc *MemoryCache) evictLRU() {
	// Get the last element (least recently used)
	elem := mc.lru.Back()
	if elem == nil {
		return
	}

	item := elem.Value.(*lruItem)
	mc.removeEntry(item.key, item.entry)
}

// removeEntry removes an entry from both map and LRU list
func (mc *MemoryCache) removeEntry(key string, entry *cacheEntry) {
	mc.currentSize -= int64(entry.size)
	delete(mc.entries, key)
	
	if entry.element != nil {
		mc.lru.Remove(entry.element)
	}
}

// Stats returns cache statistics
func (mc *MemoryCache) Stats() map[string]interface{} {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	return map[string]interface{}{
		"entries":       len(mc.entries),
		"size_bytes":    mc.currentSize,
		"size_mb":       float64(mc.currentSize) / (1024 * 1024),
		"max_size_mb":   float64(mc.maxSizeBytes) / (1024 * 1024),
		"utilization_%": float64(mc.currentSize) / float64(mc.maxSizeBytes) * 100,
	}
}

// matchPattern performs simple pattern matching with * wildcard
func matchPattern(pattern, str string) bool {
	if pattern == "*" {
		return true
	}

	if !strings.Contains(pattern, "*") {
		return pattern == str
	}

	// Split by * and check each part
	parts := strings.Split(pattern, "*")
	if len(parts) == 0 {
		return false
	}

	// Check prefix
	if parts[0] != "" && !strings.HasPrefix(str, parts[0]) {
		return false
	}

	// Check suffix
	if parts[len(parts)-1] != "" && !strings.HasSuffix(str, parts[len(parts)-1]) {
		return false
	}

	// Check middle parts (simplified)
	currentPos := 0
	for i, part := range parts {
		if i == 0 || i == len(parts)-1 || part == "" {
			continue
		}

		index := strings.Index(str[currentPos:], part)
		if index == -1 {
			return false
		}
		currentPos += index + len(part)
	}

	return true
}