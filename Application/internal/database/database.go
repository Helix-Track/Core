package database

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"           // PostgreSQL driver
	_ "github.com/mattn/go-sqlite3" // SQLite driver
	"helixtrack.ru/core/internal/config"
)

// Database represents a database connection interface
type Database interface {
	// Query executes a query that returns rows
	Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)

	// QueryRow executes a query that returns a single row
	QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row

	// Exec executes a query that doesn't return rows
	Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error)

	// Begin starts a transaction
	Begin(ctx context.Context) (*sql.Tx, error)

	// Close closes the database connection
	Close() error

	// Ping verifies the database connection
	Ping(ctx context.Context) error

	// GetType returns the database type (sqlite or postgres)
	GetType() string
}

// db is the concrete implementation of Database interface
type db struct {
	conn   *sql.DB
	dbType string
}

// NewDatabase creates a new database connection based on configuration
func NewDatabase(cfg config.DatabaseConfig) (Database, error) {
	var conn *sql.DB
	var err error

	switch cfg.Type {
	case "sqlite":
		conn, err = sql.Open("sqlite3", cfg.SQLitePath)
		if err != nil {
			return nil, fmt.Errorf("failed to open sqlite database: %w", err)
		}

		// Enable foreign keys for SQLite
		_, err = conn.Exec("PRAGMA foreign_keys = ON")
		if err != nil {
			conn.Close()
			return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
		}

		// Configure connection pool for SQLite
		conn.SetMaxOpenConns(1) // SQLite only supports one write connection

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

		conn, err = sql.Open("postgres", connStr)
		if err != nil {
			return nil, fmt.Errorf("failed to open postgres database: %w", err)
		}

		// Configure connection pool for PostgreSQL
		conn.SetMaxOpenConns(25)
		conn.SetMaxIdleConns(5)

	default:
		return nil, fmt.Errorf("unsupported database type: %s", cfg.Type)
	}

	// Verify connection
	if err := conn.Ping(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &db{
		conn:   conn,
		dbType: cfg.Type,
	}, nil
}

// Query executes a query that returns rows
func (d *db) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return d.conn.QueryContext(ctx, query, args...)
}

// QueryRow executes a query that returns a single row
func (d *db) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return d.conn.QueryRowContext(ctx, query, args...)
}

// Exec executes a query that doesn't return rows
func (d *db) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return d.conn.ExecContext(ctx, query, args...)
}

// Begin starts a transaction
func (d *db) Begin(ctx context.Context) (*sql.Tx, error) {
	return d.conn.BeginTx(ctx, nil)
}

// Close closes the database connection
func (d *db) Close() error {
	return d.conn.Close()
}

// Ping verifies the database connection
func (d *db) Ping(ctx context.Context) error {
	return d.conn.PingContext(ctx)
}

// GetType returns the database type
func (d *db) GetType() string {
	return d.dbType
}
