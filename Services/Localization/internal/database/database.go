package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/helixtrack/localization-service/internal/config"
	"github.com/helixtrack/localization-service/internal/models"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

// Database interface defines all database operations
type Database interface {
	// Connection management
	Ping() error
	Close() error

	// Language operations
	CreateLanguage(ctx context.Context, lang *models.Language) error
	GetLanguageByID(ctx context.Context, id string) (*models.Language, error)
	GetLanguageByCode(ctx context.Context, code string) (*models.Language, error)
	GetLanguages(ctx context.Context, activeOnly bool) ([]*models.Language, error)
	UpdateLanguage(ctx context.Context, lang *models.Language) error
	DeleteLanguage(ctx context.Context, id string) error
	GetDefaultLanguage(ctx context.Context) (*models.Language, error)

	// Localization Key operations
	CreateLocalizationKey(ctx context.Context, key *models.LocalizationKey) error
	GetLocalizationKeyByID(ctx context.Context, id string) (*models.LocalizationKey, error)
	GetLocalizationKeyByKey(ctx context.Context, key string) (*models.LocalizationKey, error)
	GetLocalizationKeysByCategory(ctx context.Context, category string) ([]*models.LocalizationKey, error)
	UpdateLocalizationKey(ctx context.Context, key *models.LocalizationKey) error
	DeleteLocalizationKey(ctx context.Context, id string) error

	// Localization operations
	CreateLocalization(ctx context.Context, loc *models.Localization) error
	GetLocalizationByID(ctx context.Context, id string) (*models.Localization, error)
	GetLocalizationByKeyAndLanguage(ctx context.Context, keyID, languageID string) (*models.Localization, error)
	GetLocalizationsByLanguage(ctx context.Context, languageID string) ([]*models.Localization, error)
	GetLocalizationsByKeyID(ctx context.Context, keyID string) ([]*models.Localization, error)
	UpdateLocalization(ctx context.Context, loc *models.Localization) error
	DeleteLocalization(ctx context.Context, id string) error
	ApproveLocalization(ctx context.Context, id, username string) error

	// Catalog operations
	CreateCatalog(ctx context.Context, catalog *models.LocalizationCatalog) error
	GetCatalogByLanguage(ctx context.Context, languageID string, category string) (*models.LocalizationCatalog, error)
	GetLatestCatalog(ctx context.Context, languageID string, category string) (*models.LocalizationCatalog, error)
	UpdateCatalog(ctx context.Context, catalog *models.LocalizationCatalog) error
	DeleteCatalog(ctx context.Context, id string) error
	BuildCatalog(ctx context.Context, languageID string, category string) (*models.LocalizationCatalog, error)

	// Audit operations
	CreateAuditLog(ctx context.Context, action, entityType, entityID, username string, changes interface{}, ipAddress, userAgent string) error

	// Utility operations
	GetStats(ctx context.Context) (map[string]interface{}, error)
}

// PostgresDatabase implements the Database interface for PostgreSQL
type PostgresDatabase struct {
	db     *sql.DB
	config *config.DatabaseConfig
	logger *zap.Logger
}

// New creates a new database connection
func New(cfg *config.DatabaseConfig, logger *zap.Logger) (Database, error) {
	dsn := cfg.GetDSN()

	db, err := sql.Open(cfg.Driver, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(cfg.MaxConnections)
	db.SetMaxIdleConns(cfg.IdleConnections)
	db.SetConnMaxLifetime(time.Duration(cfg.ConnectionLifetime) * time.Second)
	db.SetConnMaxIdleTime(time.Duration(cfg.ConnectionTimeout) * time.Second)

	// Ping to verify connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("database connection established",
		zap.String("driver", cfg.Driver),
		zap.String("host", cfg.Host),
		zap.Int("port", cfg.Port),
		zap.String("database", cfg.Database),
	)

	return &PostgresDatabase{
		db:     db,
		config: cfg,
		logger: logger,
	}, nil
}

// Ping checks database connectivity
func (d *PostgresDatabase) Ping() error {
	return d.db.Ping()
}

// Close closes the database connection
func (d *PostgresDatabase) Close() error {
	return d.db.Close()
}

// execContext executes a query with context
func (d *PostgresDatabase) execContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	result, err := d.db.ExecContext(ctx, query, args...)
	if err != nil {
		d.logger.Error("database exec error",
			zap.Error(err),
			zap.String("query", query),
		)
		return nil, models.ErrDatabase(err)
	}
	return result, nil
}

// queryRowContext executes a query that returns a single row
func (d *PostgresDatabase) queryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return d.db.QueryRowContext(ctx, query, args...)
}

// queryContext executes a query that returns multiple rows
func (d *PostgresDatabase) queryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	rows, err := d.db.QueryContext(ctx, query, args...)
	if err != nil {
		d.logger.Error("database query error",
			zap.Error(err),
			zap.String("query", query),
		)
		return nil, models.ErrDatabase(err)
	}
	return rows, nil
}

// encrypt encrypts a value using the configured encryption key
func (d *PostgresDatabase) encrypt(value string) string {
	// PostgreSQL pgcrypto encryption
	// This is a simplified version - actual implementation should use pgcrypto functions
	return value // Placeholder - will be encrypted by PostgreSQL
}

// decrypt decrypts a value using the configured encryption key
func (d *PostgresDatabase) decrypt(value string) string {
	// PostgreSQL pgcrypto decryption
	// This is a simplified version - actual implementation should use pgcrypto functions
	return value // Placeholder - will be decrypted by PostgreSQL
}
