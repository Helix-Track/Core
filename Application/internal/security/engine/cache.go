package engine

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
	"helixtrack.ru/core/internal/logger"
)

// PermissionCache caches permission check results for performance
type PermissionCache struct {
	entries     map[string]*CacheEntry
	contexts    map[string]*SecurityContext
	mu          sync.RWMutex
	maxSize     int
	defaultTTL  time.Duration
	hitCount    uint64
	missCount   uint64
	evictCount  uint64
	cleanupStop chan bool
}

// NewPermissionCache creates a new permission cache
func NewPermissionCache(maxSize int, defaultTTL time.Duration) *PermissionCache {
	cache := &PermissionCache{
		entries:     make(map[string]*CacheEntry, maxSize),
		contexts:    make(map[string]*SecurityContext, maxSize/10),
		maxSize:     maxSize,
		defaultTTL:  defaultTTL,
		cleanupStop: make(chan bool),
	}

	// Start background cleanup goroutine
	go cache.cleanupExpired()

	return cache
}

// Get retrieves a cached access response
func (pc *PermissionCache) Get(req AccessRequest) (AccessResponse, bool) {
	pc.mu.RLock()
	defer pc.mu.RUnlock()

	key := pc.generateKey(req)
	entry, exists := pc.entries[key]

	if !exists {
		pc.missCount++
		return AccessResponse{}, false
	}

	// Check if expired
	if time.Now().After(entry.ExpiresAt) {
		pc.missCount++
		return AccessResponse{}, false
	}

	pc.hitCount++
	logger.Debug("Permission cache hit",
		zap.String("username", req.Username),
		zap.String("resource", req.Resource),
		zap.Float64("hit_rate", pc.GetHitRate()),
	)

	return entry.Response, true
}

// Set stores an access response in the cache with default TTL
func (pc *PermissionCache) Set(req AccessRequest, response AccessResponse) {
	pc.SetWithTTL(req, response, pc.defaultTTL)
}

// SetWithTTL stores an access response with a specific TTL
func (pc *PermissionCache) SetWithTTL(req AccessRequest, response AccessResponse, ttl time.Duration) {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	// Check size limit and evict if necessary
	if len(pc.entries) >= pc.maxSize {
		pc.evictOldest()
	}

	key := pc.generateKey(req)
	entry := &CacheEntry{
		Request:   req,
		Response:  response,
		CachedAt:  time.Now(),
		ExpiresAt: time.Now().Add(ttl),
	}

	pc.entries[key] = entry

	logger.Debug("Permission cache set",
		zap.String("username", req.Username),
		zap.String("resource", req.Resource),
		zap.Duration("ttl", ttl),
	)
}

// GetContext retrieves a cached security context
func (pc *PermissionCache) GetContext(username string) (*SecurityContext, bool) {
	pc.mu.RLock()
	defer pc.mu.RUnlock()

	ctx, exists := pc.contexts[username]
	if !exists {
		return nil, false
	}

	// Check if expired
	if time.Now().After(ctx.ExpiresAt) {
		return nil, false
	}

	return ctx, true
}

// SetContext stores a security context in the cache
func (pc *PermissionCache) SetContext(username string, ctx *SecurityContext) {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	pc.contexts[username] = ctx

	logger.Debug("Security context cached",
		zap.String("username", username),
		zap.Int("role_count", len(ctx.Roles)),
		zap.Int("team_count", len(ctx.Teams)),
	)
}

// InvalidateUser removes all cache entries for a specific user
func (pc *PermissionCache) InvalidateUser(username string) {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	// Remove security context
	delete(pc.contexts, username)

	// Remove all permission entries for this user
	for key, entry := range pc.entries {
		if entry.Request.Username == username {
			delete(pc.entries, key)
		}
	}

	logger.Info("Cache invalidated for user",
		zap.String("username", username),
	)
}

// Clear removes all entries from the cache
func (pc *PermissionCache) Clear() {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	pc.entries = make(map[string]*CacheEntry, pc.maxSize)
	pc.contexts = make(map[string]*SecurityContext, pc.maxSize/10)

	logger.Info("Permission cache cleared")
}

// GetStats returns cache statistics
func (pc *PermissionCache) GetStats() CacheStats {
	pc.mu.RLock()
	defer pc.mu.RUnlock()

	return CacheStats{
		EntryCount:  len(pc.entries),
		ContextCount: len(pc.contexts),
		HitCount:    pc.hitCount,
		MissCount:   pc.missCount,
		EvictCount:  pc.evictCount,
		HitRate:     pc.GetHitRate(),
		MaxSize:     pc.maxSize,
	}
}

// GetHitRate returns the cache hit rate
func (pc *PermissionCache) GetHitRate() float64 {
	total := pc.hitCount + pc.missCount
	if total == 0 {
		return 0.0
	}
	return float64(pc.hitCount) / float64(total)
}

// Stop stops the background cleanup goroutine
func (pc *PermissionCache) Stop() {
	close(pc.cleanupStop)
}

// generateKey generates a cache key from an access request
func (pc *PermissionCache) generateKey(req AccessRequest) string {
	// Create a deterministic key from request fields
	data := struct {
		Username   string
		Resource   string
		ResourceID string
		Action     Action
		Context    map[string]string
	}{
		Username:   req.Username,
		Resource:   req.Resource,
		ResourceID: req.ResourceID,
		Action:     req.Action,
		Context:    req.Context,
	}

	jsonBytes, _ := json.Marshal(data)
	hash := sha256.Sum256(jsonBytes)
	return fmt.Sprintf("%x", hash)
}

// evictOldest removes the oldest cache entry
func (pc *PermissionCache) evictOldest() {
	var oldestKey string
	var oldestTime time.Time = time.Now()

	for key, entry := range pc.entries {
		if entry.CachedAt.Before(oldestTime) {
			oldestTime = entry.CachedAt
			oldestKey = key
		}
	}

	if oldestKey != "" {
		delete(pc.entries, oldestKey)
		pc.evictCount++
	}
}

// cleanupExpired periodically removes expired entries
func (pc *PermissionCache) cleanupExpired() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			pc.removeExpired()
		case <-pc.cleanupStop:
			return
		}
	}
}

// removeExpired removes all expired entries
func (pc *PermissionCache) removeExpired() {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	now := time.Now()
	removedCount := 0

	// Remove expired permission entries
	for key, entry := range pc.entries {
		if now.After(entry.ExpiresAt) {
			delete(pc.entries, key)
			removedCount++
		}
	}

	// Remove expired security contexts
	for username, ctx := range pc.contexts {
		if now.After(ctx.ExpiresAt) {
			delete(pc.contexts, username)
			removedCount++
		}
	}

	if removedCount > 0 {
		logger.Debug("Removed expired cache entries",
			zap.Int("count", removedCount),
			zap.Int("remaining", len(pc.entries)+len(pc.contexts)),
		)
	}
}

// CacheStats represents cache statistics
type CacheStats struct {
	EntryCount   int
	ContextCount int
	HitCount     uint64
	MissCount    uint64
	EvictCount   uint64
	HitRate      float64
	MaxSize      int
}
