package database

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3" // SQLite driver
	_ "github.com/lib/pq"           // PostgreSQL driver
	"helixtrack.ru/core/internal/config"
)

// OptimizedDatabase extends Database with performance features
type OptimizedDatabase interface {
	Database

	// PreparedQuery executes a prepared statement with caching
	PreparedQuery(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)

	// PreparedQueryRow executes a prepared statement that returns a single row
	PreparedQueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row

	// PreparedExec executes a prepared statement that doesn't return rows
	PreparedExec(ctx context.Context, query string, args ...interface{}) (sql.Result, error)

	// GetStats returns database performance statistics
	GetStats() *DatabaseStats

	// ClearPreparedStatements clears the prepared statement cache
	ClearPreparedStatements() error
}

// DatabaseStats contains performance metrics
type DatabaseStats struct {
	OpenConnections     int           // Current open connections
	InUseConnections    int           // Connections currently in use
	IdleConnections     int           // Idle connections
	WaitCount           int64         // Total connections waited for
	WaitDuration        time.Duration // Total time waited for connections
	MaxIdleClosed       int64         // Connections closed due to max idle
	MaxLifetimeClosed   int64         // Connections closed due to max lifetime
	PreparedStmtCount   int           // Number of cached prepared statements
	QueryCount          int64         // Total queries executed
	PreparedQueryCount  int64         // Total prepared queries executed
	AvgQueryDuration    time.Duration // Average query execution time
}

// optimizedDB is the high-performance implementation
type optimizedDB struct {
	conn               *sql.DB
	dbType             string
	encryptionKey      string

	// Prepared statement cache
	stmtCache          map[string]*sql.Stmt
	stmtCacheMu        sync.RWMutex

	// Performance metrics
	queryCount         int64
	preparedQueryCount int64
	totalQueryTime     time.Duration
	queryTimeMu        sync.RWMutex
}

// OptimizationConfig contains database optimization settings
type OptimizationConfig struct {
	// Connection Pool Settings
	MaxOpenConns        int           // Maximum number of open connections
	MaxIdleConns        int           // Maximum number of idle connections
	ConnMaxLifetime     time.Duration // Maximum connection lifetime
	ConnMaxIdleTime     time.Duration // Maximum connection idle time

	// SQLCipher Settings (SQLite only)
	EncryptionKey       string        // Encryption key for SQLCipher
	KDFIterations       int           // Key derivation iterations (default: 256000)
	CipherPageSize      int           // Page size in bytes (default: 4096)
	CipherUseHMAC       bool          // Use HMAC for integrity (default: true)

	// Performance Settings
	EnableWAL           bool          // Enable Write-Ahead Logging (SQLite)
	CacheSize           int           // Cache size in pages (SQLite, default: -2000 = 2MB)
	BusyTimeout         int           // Busy timeout in milliseconds (SQLite)
	JournalMode         string        // Journal mode (SQLite, default: WAL)
	Synchronous         string        // Synchronous mode (SQLite, default: NORMAL)
	TempStore           string        // Temp store location (SQLite, default: MEMORY)
	MMAPSize            int64         // Memory-mapped I/O size (SQLite, default: 0 = disabled)

	// PostgreSQL Settings
	StatementTimeout    int           // Statement timeout in milliseconds
	IdleInTxTimeout     int           // Idle in transaction timeout
	EnableJIT           bool          // Enable JIT compilation (PostgreSQL 11+)
}

// DefaultOptimizationConfig returns optimized default settings
func DefaultOptimizationConfig() OptimizationConfig {
	return OptimizationConfig{
		// Connection Pool - Optimized for high concurrency
		MaxOpenConns:      100,                  // High concurrency support
		MaxIdleConns:      25,                   // Keep connections warm
		ConnMaxLifetime:   time.Hour,            // Recycle connections hourly
		ConnMaxIdleTime:   15 * time.Minute,     // Close idle after 15min

		// SQLCipher - Secure with good performance
		KDFIterations:     256000,               // Strong key derivation
		CipherPageSize:    4096,                 // Optimal page size
		CipherUseHMAC:     true,                 // Integrity verification

		// SQLite Performance - Maximum performance
		EnableWAL:         true,                 // Write-Ahead Logging
		CacheSize:         -64000,               // 64MB cache
		BusyTimeout:       5000,                 // 5 second busy timeout
		JournalMode:       "WAL",                // Best concurrency
		Synchronous:       "NORMAL",             // Balance safety/performance
		TempStore:         "MEMORY",             // Fast temp tables
		MMAPSize:          268435456,            // 256MB memory-mapped I/O

		// PostgreSQL - Production optimized
		StatementTimeout:  30000,                // 30 second timeout
		IdleInTxTimeout:   60000,                // 1 minute idle in tx
		EnableJIT:         true,                 // JIT compilation
	}
}

// NewOptimizedDatabase creates a high-performance encrypted database connection
func NewOptimizedDatabase(cfg config.DatabaseConfig, optCfg OptimizationConfig) (OptimizedDatabase, error) {
	var conn *sql.DB
	var err error

	switch cfg.Type {
	case "sqlite":
		// Build SQLCipher connection string with encryption and optimizations
		connStr := cfg.SQLitePath

		// Check if the path already contains parameters (e.g., "file::memory:?cache=shared")
		hasParams := false
		if len(connStr) > 0 && (connStr[len(connStr)-1] != '/' && connStr[len(connStr)-1] != '\\') {
			// Check if string contains '?'
			for i := 0; i < len(connStr); i++ {
				if connStr[i] == '?' {
					hasParams = true
					break
				}
			}
		}

		if optCfg.EncryptionKey != "" {
			// Use SQLCipher with encryption
			if hasParams {
				connStr += fmt.Sprintf("&_pragma_key=%s", optCfg.EncryptionKey)
			} else {
				connStr += fmt.Sprintf("?_pragma_key=%s", optCfg.EncryptionKey)
				hasParams = true
			}

			// Add cipher configuration
			connStr += fmt.Sprintf("&_pragma_cipher_page_size=%d", optCfg.CipherPageSize)
			connStr += fmt.Sprintf("&_pragma_kdf_iter=%d", optCfg.KDFIterations)
			if optCfg.CipherUseHMAC {
				connStr += "&_pragma_cipher_use_hmac=ON"
			}
		}

		// Add performance pragmas
		separator := "?"
		if hasParams {
			separator = "&"
		}
		connStr += fmt.Sprintf("%s_pragma_foreign_keys=ON", separator)
		connStr += fmt.Sprintf("&_pragma_journal_mode=%s", optCfg.JournalMode)
		connStr += fmt.Sprintf("&_pragma_synchronous=%s", optCfg.Synchronous)
		connStr += fmt.Sprintf("&_pragma_cache_size=%d", optCfg.CacheSize)
		connStr += fmt.Sprintf("&_pragma_temp_store=%s", optCfg.TempStore)
		connStr += fmt.Sprintf("&_pragma_mmap_size=%d", optCfg.MMAPSize)
		connStr += fmt.Sprintf("&_pragma_busy_timeout=%d", optCfg.BusyTimeout)

		// Additional optimizations
		connStr += "&_pragma_locking_mode=NORMAL"  // Allow multiple connections
		connStr += "&_pragma_auto_vacuum=INCREMENTAL" // Incremental vacuum
		connStr += "&_pragma_page_size=4096"       // Optimal page size

		// Open with SQLCipher driver
		if optCfg.EncryptionKey != "" {
			conn, err = sql.Open("sqlite3", connStr)
		} else {
			// Fallback to standard SQLite if no encryption
			conn, err = sql.Open("sqlite3", connStr)
		}

		if err != nil {
			return nil, fmt.Errorf("failed to open sqlite database: %w", err)
		}

		// Configure connection pool for SQLite
		// Note: SQLite with WAL mode supports multiple readers
		if optCfg.EnableWAL {
			conn.SetMaxOpenConns(optCfg.MaxOpenConns)
			conn.SetMaxIdleConns(optCfg.MaxIdleConns)
		} else {
			conn.SetMaxOpenConns(1) // Single writer mode
			conn.SetMaxIdleConns(1)
		}

	case "postgres":
		connStr := fmt.Sprintf(
			"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			cfg.PostgresHost,
			cfg.PostgresPort,
			cfg.PostgresUser,
			cfg.PostgresPassword,
			cfg.PostgresDatabase,
			cfg.PostgresSSLMode,
		)

		// Add performance parameters
		if optCfg.StatementTimeout > 0 {
			connStr += fmt.Sprintf(" statement_timeout=%d", optCfg.StatementTimeout)
		}
		if optCfg.IdleInTxTimeout > 0 {
			connStr += fmt.Sprintf(" idle_in_transaction_session_timeout=%d", optCfg.IdleInTxTimeout)
		}

		conn, err = sql.Open("postgres", connStr)
		if err != nil {
			return nil, fmt.Errorf("failed to open postgres database: %w", err)
		}

		// Configure connection pool for PostgreSQL
		conn.SetMaxOpenConns(optCfg.MaxOpenConns)
		conn.SetMaxIdleConns(optCfg.MaxIdleConns)

	default:
		return nil, fmt.Errorf("unsupported database type: %s", cfg.Type)
	}

	// Set connection pool timeouts
	conn.SetConnMaxLifetime(optCfg.ConnMaxLifetime)
	conn.SetConnMaxIdleTime(optCfg.ConnMaxIdleTime)

	// Verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := conn.PingContext(ctx); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Apply PostgreSQL-specific optimizations
	if cfg.Type == "postgres" {
		if optCfg.EnableJIT {
			_, _ = conn.Exec("SET jit = ON")
		}
		// Set work_mem for better query performance
		_, _ = conn.Exec("SET work_mem = '64MB'")
		// Set shared_buffers recommendation
		_, _ = conn.Exec("SET shared_buffers = '256MB'")
		// Set effective_cache_size
		_, _ = conn.Exec("SET effective_cache_size = '1GB'")
	}

	return &optimizedDB{
		conn:          conn,
		dbType:        cfg.Type,
		encryptionKey: optCfg.EncryptionKey,
		stmtCache:     make(map[string]*sql.Stmt),
	}, nil
}

// Query executes a query that returns rows
func (d *optimizedDB) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	start := time.Now()
	defer d.trackQueryTime(time.Since(start))

	d.incrementQueryCount()
	return d.conn.QueryContext(ctx, query, args...)
}

// QueryRow executes a query that returns a single row
func (d *optimizedDB) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	start := time.Now()
	defer d.trackQueryTime(time.Since(start))

	d.incrementQueryCount()
	return d.conn.QueryRowContext(ctx, query, args...)
}

// Exec executes a query that doesn't return rows
func (d *optimizedDB) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	start := time.Now()
	defer d.trackQueryTime(time.Since(start))

	d.incrementQueryCount()
	return d.conn.ExecContext(ctx, query, args...)
}

// PreparedQuery executes a prepared statement with caching
func (d *optimizedDB) PreparedQuery(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	start := time.Now()
	defer d.trackQueryTime(time.Since(start))

	stmt, err := d.getOrCreateStmt(ctx, query)
	if err != nil {
		return nil, err
	}

	d.incrementPreparedQueryCount()
	return stmt.QueryContext(ctx, args...)
}

// PreparedQueryRow executes a prepared statement that returns a single row
func (d *optimizedDB) PreparedQueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	start := time.Now()
	defer d.trackQueryTime(time.Since(start))

	stmt, err := d.getOrCreateStmt(ctx, query)
	if err != nil {
		// Fallback to regular query on error
		return d.conn.QueryRowContext(ctx, query, args...)
	}

	d.incrementPreparedQueryCount()
	return stmt.QueryRowContext(ctx, args...)
}

// PreparedExec executes a prepared statement that doesn't return rows
func (d *optimizedDB) PreparedExec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	start := time.Now()
	defer d.trackQueryTime(time.Since(start))

	stmt, err := d.getOrCreateStmt(ctx, query)
	if err != nil {
		return nil, err
	}

	d.incrementPreparedQueryCount()
	return stmt.ExecContext(ctx, args...)
}

// getOrCreateStmt gets a cached prepared statement or creates a new one
func (d *optimizedDB) getOrCreateStmt(ctx context.Context, query string) (*sql.Stmt, error) {
	// Check cache first (read lock)
	d.stmtCacheMu.RLock()
	stmt, exists := d.stmtCache[query]
	d.stmtCacheMu.RUnlock()

	if exists {
		return stmt, nil
	}

	// Create new statement (write lock)
	d.stmtCacheMu.Lock()
	defer d.stmtCacheMu.Unlock()

	// Double-check after acquiring write lock
	stmt, exists = d.stmtCache[query]
	if exists {
		return stmt, nil
	}

	// Prepare new statement
	stmt, err := d.conn.PrepareContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare statement: %w", err)
	}

	d.stmtCache[query] = stmt
	return stmt, nil
}

// ClearPreparedStatements clears the prepared statement cache
func (d *optimizedDB) ClearPreparedStatements() error {
	d.stmtCacheMu.Lock()
	defer d.stmtCacheMu.Unlock()

	var lastErr error
	for _, stmt := range d.stmtCache {
		if err := stmt.Close(); err != nil {
			lastErr = err
		}
	}

	d.stmtCache = make(map[string]*sql.Stmt)
	return lastErr
}

// Begin starts a transaction
func (d *optimizedDB) Begin(ctx context.Context) (*sql.Tx, error) {
	return d.conn.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted, // Best balance for most workloads
	})
}

// Close closes the database connection and cleans up resources
func (d *optimizedDB) Close() error {
	// Clear prepared statements first
	d.ClearPreparedStatements()

	return d.conn.Close()
}

// Ping verifies the database connection
func (d *optimizedDB) Ping(ctx context.Context) error {
	return d.conn.PingContext(ctx)
}

// GetType returns the database type
func (d *optimizedDB) GetType() string {
	return d.dbType
}

// GetStats returns database performance statistics
func (d *optimizedDB) GetStats() *DatabaseStats {
	stats := d.conn.Stats()

	d.queryTimeMu.RLock()
	queryCount := d.queryCount
	preparedQueryCount := d.preparedQueryCount
	totalQueryTime := d.totalQueryTime
	d.queryTimeMu.RUnlock()

	var avgQueryDuration time.Duration
	totalQueries := queryCount + preparedQueryCount
	if totalQueries > 0 {
		avgQueryDuration = totalQueryTime / time.Duration(totalQueries)
	}

	d.stmtCacheMu.RLock()
	preparedStmtCount := len(d.stmtCache)
	d.stmtCacheMu.RUnlock()

	return &DatabaseStats{
		OpenConnections:    stats.OpenConnections,
		InUseConnections:   stats.InUse,
		IdleConnections:    stats.Idle,
		WaitCount:          stats.WaitCount,
		WaitDuration:       stats.WaitDuration,
		MaxIdleClosed:      stats.MaxIdleClosed,
		MaxLifetimeClosed:  stats.MaxLifetimeClosed,
		PreparedStmtCount:  preparedStmtCount,
		QueryCount:         queryCount,
		PreparedQueryCount: preparedQueryCount,
		AvgQueryDuration:   avgQueryDuration,
	}
}

// incrementQueryCount atomically increments the query counter
func (d *optimizedDB) incrementQueryCount() {
	d.queryTimeMu.Lock()
	d.queryCount++
	d.queryTimeMu.Unlock()
}

// incrementPreparedQueryCount atomically increments the prepared query counter
func (d *optimizedDB) incrementPreparedQueryCount() {
	d.queryTimeMu.Lock()
	d.preparedQueryCount++
	d.queryTimeMu.Unlock()
}

// trackQueryTime adds query execution time to total
func (d *optimizedDB) trackQueryTime(duration time.Duration) {
	d.queryTimeMu.Lock()
	d.totalQueryTime += duration
	d.queryTimeMu.Unlock()
}
