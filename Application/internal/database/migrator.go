package database

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"helixtrack.ru/core/internal/config"
)

// Migrator handles DDL file-based migrations for first-start database initialization.
type Migrator struct {
	ddlDir string
}

// NewMigrator creates a new Migrator reading .sql files from ddlDir.
func NewMigrator(ddlDir string) *Migrator {
	return &Migrator{ddlDir: ddlDir}
}

// RunMigrations reads all .sql files from the DDL directory (excluding
// test-data files) and executes them in sorted order against a fresh database.
func (m *Migrator) RunMigrations(dbCfg config.DatabaseConfig) error {
	entries, err := os.ReadDir(m.ddlDir)
	if err != nil {
		return fmt.Errorf("failed to read DDL directory %s: %w", m.ddlDir, err)
	}

	// Collect .sql files, skip test data files
	var sqlFiles []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(strings.ToLower(name), ".sql") {
			continue
		}
		if strings.Contains(strings.ToLower(name), "test") {
			continue
		}
		sqlFiles = append(sqlFiles, name)
	}

	// Sort for deterministic execution order
	sort.Strings(sqlFiles)

	// Open a temporary connection to run migrations
	db, err := NewDatabase(dbCfg)
	if err != nil {
		return fmt.Errorf("failed to open database for migration: %w", err)
	}
	defer db.Close()

	ctx := context.Background()
	for _, file := range sqlFiles {
		path := filepath.Join(m.ddlDir, file)
		content, readErr := os.ReadFile(path)
		if readErr != nil {
			return fmt.Errorf("failed to read migration file %s: %w", file, readErr)
		}

		// Split on semicolons to execute statements individually
		statements := splitSQL(string(content))
		for _, stmt := range statements {
			trimmed := strings.TrimSpace(stmt)
			if trimmed == "" {
				continue
			}
			if _, execErr := db.Exec(ctx, trimmed); execErr != nil {
				return fmt.Errorf("migration %s failed on statement: %w\nStatement: %s",
					file, execErr, trimmed[:min(len(trimmed), 200)])
			}
		}
	}

	return nil
}

// splitSQL splits a SQL script into individual statements on semicolons.
func splitSQL(script string) []string {
	parts := strings.Split(script, ";")
	var result []string
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			result = append(result, trimmed+";")
		}
	}
	return result
}
