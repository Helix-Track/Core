package database

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"helixtrack.ru/core/internal/config"
)

func TestNewDatabase_SQLite(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	cfg := config.DatabaseConfig{
		Type:       "sqlite",
		SQLitePath: dbPath,
	}

	db, err := NewDatabase(cfg)
	require.NoError(t, err)
	require.NotNil(t, db)

	defer db.Close()

	assert.Equal(t, "sqlite", db.GetType())

	// Verify connection works
	ctx := context.Background()
	err = db.Ping(ctx)
	assert.NoError(t, err)
}

func TestNewDatabase_SQLite_ForeignKeys(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	cfg := config.DatabaseConfig{
		Type:       "sqlite",
		SQLitePath: dbPath,
	}

	db, err := NewDatabase(cfg)
	require.NoError(t, err)
	defer db.Close()

	// Verify foreign keys are enabled
	ctx := context.Background()
	row := db.QueryRow(ctx, "PRAGMA foreign_keys")
	var fkEnabled int
	err = row.Scan(&fkEnabled)
	require.NoError(t, err)
	assert.Equal(t, 1, fkEnabled, "Foreign keys should be enabled")
}

func TestNewDatabase_InvalidType(t *testing.T) {
	cfg := config.DatabaseConfig{
		Type: "invalid",
	}

	db, err := NewDatabase(cfg)
	assert.Error(t, err)
	assert.Nil(t, db)
	assert.Contains(t, err.Error(), "unsupported database type")
}

func TestNewDatabase_InvalidPath(t *testing.T) {
	cfg := config.DatabaseConfig{
		Type:       "sqlite",
		SQLitePath: "/invalid/path/that/does/not/exist/db.sqlite",
	}

	db, err := NewDatabase(cfg)
	assert.Error(t, err)
	assert.Nil(t, db)
}

func TestDatabase_Query(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()

	// Create test table
	_, err := db.Exec(ctx, "CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT)")
	require.NoError(t, err)

	// Insert test data
	_, err = db.Exec(ctx, "INSERT INTO test (id, name) VALUES (?, ?)", 1, "test1")
	require.NoError(t, err)
	_, err = db.Exec(ctx, "INSERT INTO test (id, name) VALUES (?, ?)", 2, "test2")
	require.NoError(t, err)

	// Query data
	rows, err := db.Query(ctx, "SELECT id, name FROM test ORDER BY id")
	require.NoError(t, err)
	defer rows.Close()

	var results []struct {
		ID   int
		Name string
	}

	for rows.Next() {
		var id int
		var name string
		err := rows.Scan(&id, &name)
		require.NoError(t, err)
		results = append(results, struct {
			ID   int
			Name string
		}{id, name})
	}

	assert.Len(t, results, 2)
	assert.Equal(t, 1, results[0].ID)
	assert.Equal(t, "test1", results[0].Name)
	assert.Equal(t, 2, results[1].ID)
	assert.Equal(t, "test2", results[1].Name)
}

func TestDatabase_QueryRow(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()

	// Create test table
	_, err := db.Exec(ctx, "CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT)")
	require.NoError(t, err)

	// Insert test data
	_, err = db.Exec(ctx, "INSERT INTO test (id, name) VALUES (?, ?)", 1, "test1")
	require.NoError(t, err)

	// Query single row
	row := db.QueryRow(ctx, "SELECT id, name FROM test WHERE id = ?", 1)

	var id int
	var name string
	err = row.Scan(&id, &name)
	require.NoError(t, err)

	assert.Equal(t, 1, id)
	assert.Equal(t, "test1", name)
}

func TestDatabase_QueryRow_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()

	// Create test table
	_, err := db.Exec(ctx, "CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT)")
	require.NoError(t, err)

	// Query non-existent row
	row := db.QueryRow(ctx, "SELECT id, name FROM test WHERE id = ?", 999)

	var id int
	var name string
	err = row.Scan(&id, &name)
	assert.Error(t, err)
	assert.Equal(t, sql.ErrNoRows, err)
}

func TestDatabase_Exec(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()

	// Create table
	result, err := db.Exec(ctx, "CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT)")
	require.NoError(t, err)
	assert.NotNil(t, result)

	// Insert data
	result, err = db.Exec(ctx, "INSERT INTO test (id, name) VALUES (?, ?)", 1, "test")
	require.NoError(t, err)

	rowsAffected, err := result.RowsAffected()
	require.NoError(t, err)
	assert.Equal(t, int64(1), rowsAffected)

	// Update data
	result, err = db.Exec(ctx, "UPDATE test SET name = ? WHERE id = ?", "updated", 1)
	require.NoError(t, err)

	rowsAffected, err = result.RowsAffected()
	require.NoError(t, err)
	assert.Equal(t, int64(1), rowsAffected)

	// Delete data
	result, err = db.Exec(ctx, "DELETE FROM test WHERE id = ?", 1)
	require.NoError(t, err)

	rowsAffected, err = result.RowsAffected()
	require.NoError(t, err)
	assert.Equal(t, int64(1), rowsAffected)
}

func TestDatabase_Begin(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()

	// Create test table
	_, err := db.Exec(ctx, "CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT)")
	require.NoError(t, err)

	// Begin transaction
	tx, err := db.Begin(ctx)
	require.NoError(t, err)

	// Insert data in transaction
	_, err = tx.ExecContext(ctx, "INSERT INTO test (id, name) VALUES (?, ?)", 1, "test")
	require.NoError(t, err)

	// Rollback transaction
	err = tx.Rollback()
	require.NoError(t, err)

	// Verify data was not inserted
	row := db.QueryRow(ctx, "SELECT COUNT(*) FROM test")
	var count int
	err = row.Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 0, count)

	// Begin new transaction
	tx, err = db.Begin(ctx)
	require.NoError(t, err)

	// Insert data in transaction
	_, err = tx.ExecContext(ctx, "INSERT INTO test (id, name) VALUES (?, ?)", 1, "test")
	require.NoError(t, err)

	// Commit transaction
	err = tx.Commit()
	require.NoError(t, err)

	// Verify data was inserted
	row = db.QueryRow(ctx, "SELECT COUNT(*) FROM test")
	err = row.Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestDatabase_Close(t *testing.T) {
	db := setupTestDB(t)

	err := db.Close()
	assert.NoError(t, err)

	// Verify connection is closed
	ctx := context.Background()
	err = db.Ping(ctx)
	assert.Error(t, err)
}

func TestDatabase_Ping(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	err := db.Ping(ctx)
	assert.NoError(t, err)
}

func TestDatabase_GetType(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	dbType := db.GetType()
	assert.Equal(t, "sqlite", dbType)
}

// Helper function to set up a test database
func setupTestDB(t *testing.T) Database {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	cfg := config.DatabaseConfig{
		Type:       "sqlite",
		SQLitePath: dbPath,
	}

	db, err := NewDatabase(cfg)
	require.NoError(t, err)

	return db
}

func TestDatabase_ContextCancellation(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create test table
	ctx := context.Background()
	_, err := db.Exec(ctx, "CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT)")
	require.NoError(t, err)

	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Attempt to query with cancelled context
	_, err = db.Query(ctx, "SELECT * FROM test")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context canceled")
}

func TestNewDatabase_SQLite_FileCreation(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "new_database.db")

	// Verify file doesn't exist yet
	_, err := os.Stat(dbPath)
	assert.True(t, os.IsNotExist(err))

	cfg := config.DatabaseConfig{
		Type:       "sqlite",
		SQLitePath: dbPath,
	}

	db, err := NewDatabase(cfg)
	require.NoError(t, err)
	defer db.Close()

	// Verify file was created
	_, err = os.Stat(dbPath)
	assert.NoError(t, err)
}
