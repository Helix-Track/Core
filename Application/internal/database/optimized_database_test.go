package database

import (
	"context"
	"fmt"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"helixtrack.ru/core/internal/config"
)

var dbCounter atomic.Uint64

func TestDefaultOptimizationConfig(t *testing.T) {
	cfg := DefaultOptimizationConfig()

	// Connection Pool Settings
	assert.Equal(t, 100, cfg.MaxOpenConns)
	assert.Equal(t, 25, cfg.MaxIdleConns)
	assert.Equal(t, time.Hour, cfg.ConnMaxLifetime)
	assert.Equal(t, 15*time.Minute, cfg.ConnMaxIdleTime)

	// SQLCipher Settings
	assert.Equal(t, 256000, cfg.KDFIterations)
	assert.Equal(t, 4096, cfg.CipherPageSize)
	assert.True(t, cfg.CipherUseHMAC)

	// SQLite Performance Settings
	assert.True(t, cfg.EnableWAL)
	assert.Equal(t, -64000, cfg.CacheSize)
	assert.Equal(t, 5000, cfg.BusyTimeout)
	assert.Equal(t, "WAL", cfg.JournalMode)
	assert.Equal(t, "NORMAL", cfg.Synchronous)
	assert.Equal(t, "MEMORY", cfg.TempStore)
	assert.Equal(t, int64(268435456), cfg.MMAPSize)

	// PostgreSQL Settings
	assert.Equal(t, 30000, cfg.StatementTimeout)
	assert.Equal(t, 60000, cfg.IdleInTxTimeout)
	assert.True(t, cfg.EnableJIT)
}

func TestNewOptimizedDatabase_SQLite(t *testing.T) {
	dbID := dbCounter.Add(1)
	dbCfg := config.DatabaseConfig{
		Type:       "sqlite",
		SQLitePath: fmt.Sprintf("file:testdb_%d?mode=memory&cache=shared", dbID),
	}

	optCfg := DefaultOptimizationConfig()

	db, err := NewOptimizedDatabase(dbCfg, optCfg)
	require.NoError(t, err)
	require.NotNil(t, db)
	defer db.Close()

	assert.Equal(t, "sqlite", db.GetType())

	// Verify connection works
	ctx := context.Background()
	err = db.Ping(ctx)
	assert.NoError(t, err)
}

func TestNewOptimizedDatabase_SQLite_WithEncryption(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "encrypted_test.db")

	dbCfg := config.DatabaseConfig{
		Type:       "sqlite",
		SQLitePath: dbPath,
	}

	optCfg := DefaultOptimizationConfig()
	optCfg.EncryptionKey = "test-encryption-key-12345"

	db, err := NewOptimizedDatabase(dbCfg, optCfg)
	require.NoError(t, err)
	require.NotNil(t, db)
	defer db.Close()

	// Verify connection works
	ctx := context.Background()
	err = db.Ping(ctx)
	assert.NoError(t, err)

	// Verify we can create tables and query
	_, err = db.Exec(ctx, "CREATE TABLE test (id INTEGER PRIMARY KEY, data TEXT)")
	assert.NoError(t, err)

	_, err = db.Exec(ctx, "INSERT INTO test (id, data) VALUES (?, ?)", 1, "encrypted data")
	assert.NoError(t, err)

	row := db.QueryRow(ctx, "SELECT data FROM test WHERE id = ?", 1)
	var data string
	err = row.Scan(&data)
	assert.NoError(t, err)
	assert.Equal(t, "encrypted data", data)
}

func TestNewOptimizedDatabase_SQLite_CustomOptimizations(t *testing.T) {
	dbID := dbCounter.Add(1)
	dbCfg := config.DatabaseConfig{
		Type:       "sqlite",
		SQLitePath: fmt.Sprintf("file:testdb_%d?mode=memory&cache=shared", dbID),
	}

	optCfg := OptimizationConfig{
		MaxOpenConns:    50,
		MaxIdleConns:    10,
		ConnMaxLifetime: 30 * time.Minute,
		ConnMaxIdleTime: 5 * time.Minute,
		EnableWAL:       true,
		CacheSize:       -32000, // 32MB cache
		BusyTimeout:     3000,
		JournalMode:     "WAL",
		Synchronous:     "NORMAL",
		TempStore:       "MEMORY",
		MMAPSize:        134217728, // 128MB
	}

	db, err := NewOptimizedDatabase(dbCfg, optCfg)
	require.NoError(t, err)
	require.NotNil(t, db)
	defer db.Close()

	ctx := context.Background()
	err = db.Ping(ctx)
	assert.NoError(t, err)
}

func TestNewOptimizedDatabase_InvalidType(t *testing.T) {
	dbCfg := config.DatabaseConfig{
		Type: "invalid-db-type",
	}

	optCfg := DefaultOptimizationConfig()

	db, err := NewOptimizedDatabase(dbCfg, optCfg)
	assert.Error(t, err)
	assert.Nil(t, db)
	assert.Contains(t, err.Error(), "unsupported database type")
}

func TestOptimizedDatabase_PreparedQuery(t *testing.T) {
	db := setupOptimizedTestDB(t)
	defer db.Close()

	ctx := context.Background()

	// Create test table
	_, err := db.Exec(ctx, "CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT, value INTEGER)")
	require.NoError(t, err)

	// Insert test data
	for i := 1; i <= 10; i++ {
		_, err = db.Exec(ctx, "INSERT INTO test (id, name, value) VALUES (?, ?, ?)", i, "test", i*10)
		require.NoError(t, err)
	}

	// Use prepared query (should cache the statement)
	query := "SELECT id, name, value FROM test WHERE value > ? ORDER BY id"

	rows, err := db.PreparedQuery(ctx, query, 50)
	require.NoError(t, err)
	defer rows.Close()

	var results []int
	for rows.Next() {
		var id, value int
		var name string
		err := rows.Scan(&id, &name, &value)
		require.NoError(t, err)
		results = append(results, id)
	}

	assert.Len(t, results, 5) // IDs 6, 7, 8, 9, 10
	assert.Equal(t, []int{6, 7, 8, 9, 10}, results)

	// Execute same query again (should use cached statement)
	rows2, err := db.PreparedQuery(ctx, query, 30)
	require.NoError(t, err)
	defer rows2.Close()

	results2 := []int{}
	for rows2.Next() {
		var id, value int
		var name string
		err := rows2.Scan(&id, &name, &value)
		require.NoError(t, err)
		results2 = append(results2, id)
	}

	assert.Len(t, results2, 7) // IDs 4, 5, 6, 7, 8, 9, 10
}

func TestOptimizedDatabase_PreparedQueryRow(t *testing.T) {
	db := setupOptimizedTestDB(t)
	defer db.Close()

	ctx := context.Background()

	// Create test table
	_, err := db.Exec(ctx, "CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT)")
	require.NoError(t, err)

	// Insert test data
	_, err = db.Exec(ctx, "INSERT INTO test (id, name) VALUES (?, ?)", 1, "test1")
	require.NoError(t, err)

	// Use prepared query row (should cache the statement)
	query := "SELECT id, name FROM test WHERE id = ?"
	row := db.PreparedQueryRow(ctx, query, 1)

	var id int
	var name string
	err = row.Scan(&id, &name)
	require.NoError(t, err)

	assert.Equal(t, 1, id)
	assert.Equal(t, "test1", name)

	// Execute same query again (should use cached statement)
	row2 := db.PreparedQueryRow(ctx, query, 1)
	err = row2.Scan(&id, &name)
	require.NoError(t, err)
	assert.Equal(t, 1, id)
}

func TestOptimizedDatabase_PreparedExec(t *testing.T) {
	db := setupOptimizedTestDB(t)
	defer db.Close()

	ctx := context.Background()

	// Create test table
	_, err := db.Exec(ctx, "CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT, count INTEGER)")
	require.NoError(t, err)

	// Insert initial data
	_, err = db.Exec(ctx, "INSERT INTO test (id, name, count) VALUES (?, ?, ?)", 1, "test", 0)
	require.NoError(t, err)

	// Use prepared exec to update (should cache the statement)
	updateQuery := "UPDATE test SET count = count + ? WHERE id = ?"

	result, err := db.PreparedExec(ctx, updateQuery, 1, 1)
	require.NoError(t, err)

	rowsAffected, err := result.RowsAffected()
	require.NoError(t, err)
	assert.Equal(t, int64(1), rowsAffected)

	// Execute same query again (should use cached statement)
	result2, err := db.PreparedExec(ctx, updateQuery, 5, 1)
	require.NoError(t, err)

	rowsAffected2, err := result2.RowsAffected()
	require.NoError(t, err)
	assert.Equal(t, int64(1), rowsAffected2)

	// Verify final count
	row := db.QueryRow(ctx, "SELECT count FROM test WHERE id = ?", 1)
	var count int
	err = row.Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 6, count) // 0 + 1 + 5 = 6
}

func TestOptimizedDatabase_GetStats(t *testing.T) {
	db := setupOptimizedTestDB(t)
	defer db.Close()

	ctx := context.Background()

	// Create test table
	_, err := db.Exec(ctx, "CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT)")
	require.NoError(t, err)

	// Get initial stats
	stats := db.GetStats()
	assert.NotNil(t, stats)
	assert.Equal(t, 0, stats.PreparedStmtCount)
	initialQueryCount := stats.QueryCount

	// Execute some queries
	_, err = db.Exec(ctx, "INSERT INTO test (id, name) VALUES (?, ?)", 1, "test1")
	require.NoError(t, err)

	_, err = db.Exec(ctx, "INSERT INTO test (id, name) VALUES (?, ?)", 2, "test2")
	require.NoError(t, err)

	// Get updated stats
	stats = db.GetStats()
	assert.Greater(t, stats.QueryCount, initialQueryCount)

	// Execute prepared queries
	_, err = db.PreparedQuery(ctx, "SELECT * FROM test WHERE id = ?", 1)
	require.NoError(t, err)

	_, err = db.PreparedQuery(ctx, "SELECT * FROM test WHERE name = ?", "test1")
	require.NoError(t, err)

	// Get final stats
	stats = db.GetStats()
	assert.Equal(t, 2, stats.PreparedStmtCount) // 2 different prepared statements
	assert.Greater(t, stats.PreparedQueryCount, int64(0))
	assert.GreaterOrEqual(t, stats.OpenConnections, 0)
}

func TestOptimizedDatabase_ClearPreparedStatements(t *testing.T) {
	db := setupOptimizedTestDB(t)
	defer db.Close()

	ctx := context.Background()

	// Create test table
	_, err := db.Exec(ctx, "CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT)")
	require.NoError(t, err)

	// Insert test data
	_, err = db.Exec(ctx, "INSERT INTO test (id, name) VALUES (?, ?)", 1, "test1")
	require.NoError(t, err)

	// Create multiple prepared statements
	_, err = db.PreparedQuery(ctx, "SELECT * FROM test WHERE id = ?", 1)
	require.NoError(t, err)

	_, err = db.PreparedQuery(ctx, "SELECT * FROM test WHERE name = ?", "test1")
	require.NoError(t, err)

	db.PreparedQueryRow(ctx, "SELECT COUNT(*) FROM test")

	// Verify statements are cached
	stats := db.GetStats()
	assert.Greater(t, stats.PreparedStmtCount, 0)
	initialCount := stats.PreparedStmtCount

	// Clear prepared statements
	err = db.ClearPreparedStatements()
	assert.NoError(t, err)

	// Verify cache is cleared
	stats = db.GetStats()
	assert.Equal(t, 0, stats.PreparedStmtCount)
	assert.Less(t, stats.PreparedStmtCount, initialCount)
}

func TestOptimizedDatabase_StatementCaching(t *testing.T) {
	db := setupOptimizedTestDB(t)
	defer db.Close()

	ctx := context.Background()

	// Create test table
	_, err := db.Exec(ctx, "CREATE TABLE test (id INTEGER PRIMARY KEY, value INTEGER)")
	require.NoError(t, err)

	// Insert data
	for i := 1; i <= 5; i++ {
		_, err = db.Exec(ctx, "INSERT INTO test (id, value) VALUES (?, ?)", i, i*10)
		require.NoError(t, err)
	}

	query := "SELECT value FROM test WHERE id = ?"

	// First execution - should create and cache statement
	row1 := db.PreparedQueryRow(ctx, query, 1)
	var value1 int
	err = row1.Scan(&value1)
	require.NoError(t, err)
	assert.Equal(t, 10, value1)

	stats1 := db.GetStats()
	assert.Equal(t, 1, stats1.PreparedStmtCount)

	// Second execution - should use cached statement
	row2 := db.PreparedQueryRow(ctx, query, 2)
	var value2 int
	err = row2.Scan(&value2)
	require.NoError(t, err)
	assert.Equal(t, 20, value2)

	stats2 := db.GetStats()
	assert.Equal(t, 1, stats2.PreparedStmtCount) // Same count, used cache
	assert.Greater(t, stats2.PreparedQueryCount, stats1.PreparedQueryCount)
}

func TestOptimizedDatabase_QueryMetrics(t *testing.T) {
	db := setupOptimizedTestDB(t)
	defer db.Close()

	ctx := context.Background()

	// Create test table
	_, err := db.Exec(ctx, "CREATE TABLE test (id INTEGER PRIMARY KEY)")
	require.NoError(t, err)

	// Get initial metrics
	initialStats := db.GetStats()
	initialQueryCount := initialStats.QueryCount
	initialPreparedCount := initialStats.PreparedQueryCount

	// Execute regular queries
	_, err = db.Query(ctx, "SELECT * FROM test")
	require.NoError(t, err)

	_, err = db.Exec(ctx, "INSERT INTO test (id) VALUES (?)", 1)
	require.NoError(t, err)

	// Execute prepared queries
	_, err = db.PreparedQuery(ctx, "SELECT * FROM test WHERE id = ?", 1)
	require.NoError(t, err)

	_, err = db.PreparedExec(ctx, "DELETE FROM test WHERE id = ?", 1)
	require.NoError(t, err)

	// Check metrics
	finalStats := db.GetStats()
	assert.Greater(t, finalStats.QueryCount, initialQueryCount)
	assert.Greater(t, finalStats.PreparedQueryCount, initialPreparedCount)
	assert.GreaterOrEqual(t, finalStats.AvgQueryDuration, time.Duration(0))
}

func TestOptimizedDatabase_Close(t *testing.T) {
	db := setupOptimizedTestDB(t)

	ctx := context.Background()

	// Create prepared statements
	_, err := db.Exec(ctx, "CREATE TABLE test (id INTEGER PRIMARY KEY)")
	require.NoError(t, err)

	_, err = db.PreparedQuery(ctx, "SELECT * FROM test")
	require.NoError(t, err)

	stats := db.GetStats()
	assert.Greater(t, stats.PreparedStmtCount, 0)

	// Close database
	err = db.Close()
	assert.NoError(t, err)

	// Verify connection is closed
	err = db.Ping(ctx)
	assert.Error(t, err)
}

func TestOptimizedDatabase_ConcurrentPreparedQueries(t *testing.T) {
	db := setupOptimizedTestDB(t)
	defer db.Close()

	ctx := context.Background()

	// Create test table
	_, err := db.Exec(ctx, "CREATE TABLE test (id INTEGER PRIMARY KEY, value INTEGER)")
	require.NoError(t, err)

	// Insert test data
	for i := 1; i <= 100; i++ {
		_, err = db.Exec(ctx, "INSERT INTO test (id, value) VALUES (?, ?)", i, i)
		require.NoError(t, err)
	}

	// Execute concurrent prepared queries
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			defer func() { done <- true }()

			for j := 0; j < 10; j++ {
				row := db.PreparedQueryRow(ctx, "SELECT value FROM test WHERE id = ?", id*10+j+1)
				var value int
				_ = row.Scan(&value)
			}
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify statement cache (should only have 1 unique query)
	stats := db.GetStats()
	assert.Equal(t, 1, stats.PreparedStmtCount)
	assert.GreaterOrEqual(t, stats.PreparedQueryCount, int64(100))
}

func TestOptimizedDatabase_MultipleStatements(t *testing.T) {
	db := setupOptimizedTestDB(t)
	defer db.Close()

	ctx := context.Background()

	// Create test table
	_, err := db.Exec(ctx, "CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT, value INTEGER)")
	require.NoError(t, err)

	// Create multiple different prepared statements
	queries := []string{
		"SELECT * FROM test WHERE id = ?",
		"SELECT * FROM test WHERE name = ?",
		"SELECT * FROM test WHERE value > ?",
		"SELECT COUNT(*) FROM test",
		"SELECT name FROM test WHERE id = ?",
	}

	for _, query := range queries {
		if query == "SELECT COUNT(*) FROM test" {
			db.PreparedQueryRow(ctx, query)
		} else {
			_, _ = db.PreparedQuery(ctx, query, 1)
		}
	}

	// Verify all statements are cached
	stats := db.GetStats()
	assert.Equal(t, len(queries), stats.PreparedStmtCount)
}

// Helper function to set up an optimized test database
func setupOptimizedTestDB(t *testing.T) OptimizedDatabase {
	// Generate unique database name for each test to avoid table conflicts
	dbID := dbCounter.Add(1)
	dbCfg := config.DatabaseConfig{
		Type:       "sqlite",
		SQLitePath: fmt.Sprintf("file:testdb_%d?mode=memory&cache=shared", dbID),
	}

	optCfg := DefaultOptimizationConfig()

	db, err := NewOptimizedDatabase(dbCfg, optCfg)
	require.NoError(t, err)

	return db
}

// Benchmarks

func BenchmarkOptimizedDatabase_RegularQuery(b *testing.B) {
	tmpDir := b.TempDir()
	dbPath := filepath.Join(tmpDir, "bench_test.db")

	dbCfg := config.DatabaseConfig{
		Type:       "sqlite",
		SQLitePath: dbPath,
	}

	optCfg := DefaultOptimizationConfig()
	db, err := NewOptimizedDatabase(dbCfg, optCfg)
	require.NoError(b, err)
	defer db.Close()

	ctx := context.Background()
	_, err = db.Exec(ctx, "CREATE TABLE test (id INTEGER PRIMARY KEY, value INTEGER)")
	require.NoError(b, err)

	for i := 1; i <= 1000; i++ {
		_, err = db.Exec(ctx, "INSERT INTO test (id, value) VALUES (?, ?)", i, i*10)
		require.NoError(b, err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		row := db.QueryRow(ctx, "SELECT value FROM test WHERE id = ?", i%1000+1)
		var value int
		_ = row.Scan(&value)
	}
}

func BenchmarkOptimizedDatabase_PreparedQuery(b *testing.B) {
	tmpDir := b.TempDir()
	dbPath := filepath.Join(tmpDir, "bench_test.db")

	dbCfg := config.DatabaseConfig{
		Type:       "sqlite",
		SQLitePath: dbPath,
	}

	optCfg := DefaultOptimizationConfig()
	db, err := NewOptimizedDatabase(dbCfg, optCfg)
	require.NoError(b, err)
	defer db.Close()

	ctx := context.Background()
	_, err = db.Exec(ctx, "CREATE TABLE test (id INTEGER PRIMARY KEY, value INTEGER)")
	require.NoError(b, err)

	for i := 1; i <= 1000; i++ {
		_, err = db.Exec(ctx, "INSERT INTO test (id, value) VALUES (?, ?)", i, i*10)
		require.NoError(b, err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		row := db.PreparedQueryRow(ctx, "SELECT value FROM test WHERE id = ?", i%1000+1)
		var value int
		_ = row.Scan(&value)
	}
}

func BenchmarkOptimizedDatabase_StatementCacheHit(b *testing.B) {
	tmpDir := b.TempDir()
	dbPath := filepath.Join(tmpDir, "bench_test.db")

	dbCfg := config.DatabaseConfig{
		Type:       "sqlite",
		SQLitePath: dbPath,
	}

	optCfg := DefaultOptimizationConfig()
	db, err := NewOptimizedDatabase(dbCfg, optCfg)
	require.NoError(b, err)
	defer db.Close()

	ctx := context.Background()
	_, err = db.Exec(ctx, "CREATE TABLE test (id INTEGER PRIMARY KEY)")
	require.NoError(b, err)

	// Pre-warm the cache
	db.PreparedQueryRow(ctx, "SELECT * FROM test WHERE id = ?", 1)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		db.PreparedQueryRow(ctx, "SELECT * FROM test WHERE id = ?", i)
	}
}

func BenchmarkOptimizedDatabase_GetStats(b *testing.B) {
	tmpDir := b.TempDir()
	dbPath := filepath.Join(tmpDir, "bench_test.db")

	dbCfg := config.DatabaseConfig{
		Type:       "sqlite",
		SQLitePath: dbPath,
	}

	optCfg := DefaultOptimizationConfig()
	db, err := NewOptimizedDatabase(dbCfg, optCfg)
	require.NoError(b, err)
	defer db.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = db.GetStats()
	}
}
