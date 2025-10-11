package integration

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"helixtrack.ru/core/internal/cache"
	"helixtrack.ru/core/internal/config"
	"helixtrack.ru/core/internal/database"
)

// TestDatabaseCache_Integration tests database with cache layer
func TestDatabaseCache_Integration(t *testing.T) {
	// Setup database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "cache_test.db")

	dbCfg := config.DatabaseConfig{
		Type:       "sqlite",
		SQLitePath: dbPath,
	}

	db, err := database.NewDatabase(dbCfg)
	require.NoError(t, err)
	defer db.Close()

	// Setup cache
	cacheCfg := cache.DefaultCacheConfig()
	cacheCfg.MaxSize = 100
	c := cache.NewInMemoryCache(cacheCfg)

	ctx := context.Background()

	// Create test table
	_, err = db.Exec(ctx, `
		CREATE TABLE users (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			email TEXT NOT NULL,
			created INTEGER NOT NULL
		)
	`)
	require.NoError(t, err)

	// Insert data
	_, err = db.Exec(ctx,
		"INSERT INTO users (id, name, email, created) VALUES (?, ?, ?, ?)",
		"user-1", "John Doe", "john@example.com", time.Now().Unix(),
	)
	require.NoError(t, err)

	// First read - from database (cache miss)
	cacheKey := "user:user-1"
	cachedData, found := c.Get(ctx, cacheKey)
	assert.False(t, found)
	assert.Nil(t, cachedData)

	// Read from database
	row := db.QueryRow(ctx, "SELECT name, email FROM users WHERE id = ?", "user-1")
	var name, email string
	err = row.Scan(&name, &email)
	require.NoError(t, err)
	assert.Equal(t, "John Doe", name)

	// Cache the result
	userData := map[string]string{
		"name":  name,
		"email": email,
	}
	c.Set(ctx, cacheKey, userData, 5*time.Minute)

	// Second read - from cache (cache hit)
	cachedData, found = c.Get(ctx, cacheKey)
	assert.True(t, found)
	assert.NotNil(t, cachedData)

	cachedUser := cachedData.(map[string]string)
	assert.Equal(t, "John Doe", cachedUser["name"])
	assert.Equal(t, "john@example.com", cachedUser["email"])
}

// TestDatabaseCache_WriteThrough tests write-through cache pattern
func TestDatabaseCache_WriteThrough(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "write_through_test.db")

	dbCfg := config.DatabaseConfig{
		Type:       "sqlite",
		SQLitePath: dbPath,
	}

	db, err := database.NewDatabase(dbCfg)
	require.NoError(t, err)
	defer db.Close()

	cacheCfg := cache.DefaultCacheConfig()
	c := cache.NewInMemoryCache(cacheCfg)

	ctx := context.Background()

	// Create table
	_, err = db.Exec(ctx, "CREATE TABLE items (id TEXT PRIMARY KEY, value TEXT)")
	require.NoError(t, err)

	// Write-through: Write to DB and cache simultaneously
	itemID := "item-1"
	itemValue := "test value"

	// Write to database
	_, err = db.Exec(ctx, "INSERT INTO items (id, value) VALUES (?, ?)", itemID, itemValue)
	require.NoError(t, err)

	// Write to cache
	c.Set(ctx, "item:"+itemID, itemValue, 5*time.Minute)

	// Read from cache (should be immediate)
	cached, found := c.Get(ctx, "item:"+itemID)
	assert.True(t, found)
	assert.Equal(t, itemValue, cached)

	// Verify database also has the data
	row := db.QueryRow(ctx, "SELECT value FROM items WHERE id = ?", itemID)
	var dbValue string
	err = row.Scan(&dbValue)
	require.NoError(t, err)
	assert.Equal(t, itemValue, dbValue)
}

// TestDatabaseCache_CacheInvalidation tests cache invalidation on updates
func TestDatabaseCache_CacheInvalidation(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "invalidation_test.db")

	dbCfg := config.DatabaseConfig{
		Type:       "sqlite",
		SQLitePath: dbPath,
	}

	db, err := database.NewDatabase(dbCfg)
	require.NoError(t, err)
	defer db.Close()

	cacheCfg := cache.DefaultCacheConfig()
	c := cache.NewInMemoryCache(cacheCfg)

	ctx := context.Background()

	// Create table
	_, err = db.Exec(ctx, "CREATE TABLE products (id TEXT PRIMARY KEY, name TEXT, price INTEGER)")
	require.NoError(t, err)

	// Insert initial data
	productID := "prod-1"
	_, err = db.Exec(ctx, "INSERT INTO products (id, name, price) VALUES (?, ?, ?)",
		productID, "Widget", 100)
	require.NoError(t, err)

	// Cache the product
	cacheKey := "product:" + productID
	c.Set(ctx, cacheKey, map[string]interface{}{"name": "Widget", "price": 100}, 5*time.Minute)

	// Verify cache has old data
	cached, found := c.Get(ctx, cacheKey)
	assert.True(t, found)
	oldData := cached.(map[string]interface{})
	assert.Equal(t, 100, oldData["price"])

	// Update database
	_, err = db.Exec(ctx, "UPDATE products SET price = ? WHERE id = ?", 150, productID)
	require.NoError(t, err)

	// Invalidate cache
	c.Delete(ctx, cacheKey)

	// Verify cache is empty
	_, found = c.Get(ctx, cacheKey)
	assert.False(t, found)

	// Read from database and refresh cache
	row := db.QueryRow(ctx, "SELECT name, price FROM products WHERE id = ?", productID)
	var name string
	var price int
	err = row.Scan(&name, &price)
	require.NoError(t, err)
	assert.Equal(t, 150, price)

	// Update cache with new data
	c.Set(ctx, cacheKey, map[string]interface{}{"name": name, "price": price}, 5*time.Minute)

	// Verify cache has new data
	cached, found = c.Get(ctx, cacheKey)
	assert.True(t, found)
	newData := cached.(map[string]interface{})
	assert.Equal(t, 150, newData["price"])
}

// TestOptimizedDatabase_WithCache tests optimized database with cache
func TestOptimizedDatabase_WithCache(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "optimized_cache_test.db")

	dbCfg := config.DatabaseConfig{
		Type:       "sqlite",
		SQLitePath: dbPath,
	}

	optCfg := database.DefaultOptimizationConfig()
	db, err := database.NewOptimizedDatabase(dbCfg, optCfg)
	require.NoError(t, err)
	defer db.Close()

	cacheCfg := cache.DefaultCacheConfig()
	c := cache.NewInMemoryCache(cacheCfg)

	ctx := context.Background()

	// Create table
	_, err = db.Exec(ctx, `
		CREATE TABLE queries (
			id TEXT PRIMARY KEY,
			query TEXT NOT NULL,
			result TEXT,
			created INTEGER NOT NULL
		)
	`)
	require.NoError(t, err)

	// Insert test data
	for i := 1; i <= 100; i++ {
		queryID := "query-" + string(rune(i))
		_, err = db.Exec(ctx,
			"INSERT INTO queries (id, query, result, created) VALUES (?, ?, ?, ?)",
			queryID, "SELECT * FROM test", "result", time.Now().Unix(),
		)
		require.NoError(t, err)
	}

	// Test prepared query with cache
	query := "SELECT query, result FROM queries WHERE id = ?"

	// First execution - cache miss
	startTime := time.Now()
	cacheKey := "query:query-1"
	_, found := c.Get(ctx, cacheKey)
	assert.False(t, found)

	// Execute prepared query
	row := db.PreparedQueryRow(ctx, query, "query-1")
	var queryText, result string
	err = row.Scan(&queryText, &result)
	require.NoError(t, err)
	firstDuration := time.Since(startTime)

	// Cache the result
	c.Set(ctx, cacheKey, map[string]string{"query": queryText, "result": result}, 5*time.Minute)

	// Second execution - cache hit
	startTime = time.Now()
	cached, found := c.Get(ctx, cacheKey)
	assert.True(t, found)
	secondDuration := time.Since(startTime)

	// Cache access should be faster
	assert.Less(t, secondDuration, firstDuration)

	// Verify cached data
	cachedData := cached.(map[string]string)
	assert.Equal(t, queryText, cachedData["query"])
}

// TestDatabaseCache_ConcurrentAccess tests concurrent database and cache access
func TestDatabaseCache_ConcurrentAccess(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "concurrent_test.db")

	dbCfg := config.DatabaseConfig{
		Type:       "sqlite",
		SQLitePath: dbPath,
	}

	db, err := database.NewDatabase(dbCfg)
	require.NoError(t, err)
	defer db.Close()

	cacheCfg := cache.DefaultCacheConfig()
	c := cache.NewInMemoryCache(cacheCfg)

	ctx := context.Background()

	// Create table
	_, err = db.Exec(ctx, "CREATE TABLE counters (id TEXT PRIMARY KEY, count INTEGER)")
	require.NoError(t, err)

	// Insert initial data
	_, err = db.Exec(ctx, "INSERT INTO counters (id, count) VALUES (?, ?)", "counter-1", 0)
	require.NoError(t, err)

	// Concurrent reads and writes
	done := make(chan bool)
	numGoroutines := 20

	for i := 0; i < numGoroutines; i++ {
		go func(index int) {
			defer func() { done <- true }()

			// Try to get from cache first
			cacheKey := "counter:counter-1"
			_, found := c.Get(ctx, cacheKey)

			if !found {
				// Read from database
				row := db.QueryRow(ctx, "SELECT count FROM counters WHERE id = ?", "counter-1")
				var count int
				if err := row.Scan(&count); err == nil {
					// Cache the result
					c.Set(ctx, cacheKey, count, 1*time.Second)
				}
			}

			// Simulate some work
			time.Sleep(10 * time.Millisecond)
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Verify cache has the data
	_, found := c.Get(ctx, "counter:counter-1")
	assert.True(t, found)
}

// TestDatabaseCache_ExpirationSync tests cache expiration synchronized with database
func TestDatabaseCache_ExpirationSync(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "expiration_test.db")

	dbCfg := config.DatabaseConfig{
		Type:       "sqlite",
		SQLitePath: dbPath,
	}

	db, err := database.NewDatabase(dbCfg)
	require.NoError(t, err)
	defer db.Close()

	cacheCfg := cache.DefaultCacheConfig()
	cacheCfg.CleanupInterval = 100 * time.Millisecond
	c := cache.NewInMemoryCache(cacheCfg)
	defer c.Close()

	ctx := context.Background()

	// Create table with TTL
	_, err = db.Exec(ctx, `
		CREATE TABLE sessions (
			id TEXT PRIMARY KEY,
			data TEXT,
			expires INTEGER NOT NULL
		)
	`)
	require.NoError(t, err)

	// Insert session with expiration
	sessionID := "session-1"
	expiresAt := time.Now().Add(1 * time.Second).Unix()
	_, err = db.Exec(ctx,
		"INSERT INTO sessions (id, data, expires) VALUES (?, ?, ?)",
		sessionID, "session data", expiresAt,
	)
	require.NoError(t, err)

	// Cache the session with same TTL
	cacheKey := "session:" + sessionID
	c.Set(ctx, cacheKey, "session data", 1*time.Second)

	// Verify both have the data
	_, found := c.Get(ctx, cacheKey)
	assert.True(t, found)

	row := db.QueryRow(ctx, "SELECT data FROM sessions WHERE id = ? AND expires > ?",
		sessionID, time.Now().Unix())
	var data string
	err = row.Scan(&data)
	assert.NoError(t, err)

	// Wait for expiration
	time.Sleep(2 * time.Second)

	// Verify cache has expired
	_, found = c.Get(ctx, cacheKey)
	assert.False(t, found)

	// Verify database session is also expired
	row = db.QueryRow(ctx, "SELECT data FROM sessions WHERE id = ? AND expires > ?",
		sessionID, time.Now().Unix())
	err = row.Scan(&data)
	assert.Error(t, err) // Should be no rows
}

// TestDatabaseCache_Statistics tests database stats with cache metrics
func TestDatabaseCache_Statistics(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "stats_test.db")

	dbCfg := config.DatabaseConfig{
		Type:       "sqlite",
		SQLitePath: dbPath,
	}

	optCfg := database.DefaultOptimizationConfig()
	db, err := database.NewOptimizedDatabase(dbCfg, optCfg)
	require.NoError(t, err)
	defer db.Close()

	cacheCfg := cache.DefaultCacheConfig()
	c := cache.NewInMemoryCache(cacheCfg)

	ctx := context.Background()

	// Create table
	_, err = db.Exec(ctx, "CREATE TABLE metrics (id TEXT PRIMARY KEY, value INTEGER)")
	require.NoError(t, err)

	// Perform operations and track stats
	for i := 1; i <= 50; i++ {
		// Check cache first
		cacheKey := "metric:metric-" + string(rune(i))
		_, found := c.Get(ctx, cacheKey)

		if !found {
			// Database query (cache miss)
			row := db.PreparedQueryRow(ctx, "SELECT value FROM metrics WHERE id = ?", "metric-"+string(rune(i)))
			var value int
			if err := row.Scan(&value); err == nil {
				c.Set(ctx, cacheKey, value, 5*time.Minute)
			}
		}
	}

	// Get database statistics
	dbStats := db.GetStats()
	assert.Greater(t, dbStats.QueryCount, int64(0))
	assert.GreaterOrEqual(t, dbStats.PreparedQueryCount, int64(0))

	// Get cache statistics
	cacheStats := c.GetStats()
	totalAccesses := cacheStats.Hits + cacheStats.Misses
	assert.GreaterOrEqual(t, totalAccesses, int64(50))
	assert.Greater(t, cacheStats.Misses, int64(0))
}
