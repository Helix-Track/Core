package database

import (
	"context"
	"fmt"
	"time"

	"github.com/helixtrack/attachments-service/internal/config"
	"github.com/helixtrack/attachments-service/internal/models"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"           // PostgreSQL driver
	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

// Database defines the interface for database operations
type Database interface {
	// Connection management
	Ping() error
	Close() error
	Migrate() error

	// Attachment File operations
	CreateFile(ctx context.Context, file *models.AttachmentFile) error
	GetFile(ctx context.Context, hash string) (*models.AttachmentFile, error)
	UpdateFile(ctx context.Context, file *models.AttachmentFile) error
	DeleteFile(ctx context.Context, hash string) error
	ListFiles(ctx context.Context, filter *FileFilter) ([]*models.AttachmentFile, int64, error)
	IncrementRefCount(ctx context.Context, hash string) error
	DecrementRefCount(ctx context.Context, hash string) error

	// Attachment Reference operations
	CreateReference(ctx context.Context, ref *models.AttachmentReference) error
	GetReference(ctx context.Context, id string) (*models.AttachmentReference, error)
	UpdateReference(ctx context.Context, ref *models.AttachmentReference) error
	DeleteReference(ctx context.Context, id string) error
	SoftDeleteReference(ctx context.Context, id string) error
	ListReferences(ctx context.Context, filter *ReferenceFilter) ([]*models.AttachmentReference, int64, error)
	ListReferencesByEntity(ctx context.Context, entityType, entityID string) ([]*models.AttachmentReference, error)
	ListReferencesByHash(ctx context.Context, hash string) ([]*models.AttachmentReference, error)

	// Storage Endpoint operations
	CreateEndpoint(ctx context.Context, endpoint *models.StorageEndpoint) error
	GetEndpoint(ctx context.Context, id string) (*models.StorageEndpoint, error)
	UpdateEndpoint(ctx context.Context, endpoint *models.StorageEndpoint) error
	DeleteEndpoint(ctx context.Context, id string) error
	ListEndpoints(ctx context.Context, role string) ([]*models.StorageEndpoint, error)
	GetPrimaryEndpoint(ctx context.Context) (*models.StorageEndpoint, error)

	// Storage Health operations
	RecordHealth(ctx context.Context, health *models.StorageHealth) error
	GetLatestHealth(ctx context.Context, endpointID string) (*models.StorageHealth, error)
	GetHealthHistory(ctx context.Context, endpointID string, since time.Time) ([]*models.StorageHealth, error)

	// Upload Quota operations
	GetQuota(ctx context.Context, userID string) (*models.UploadQuota, error)
	CreateQuota(ctx context.Context, quota *models.UploadQuota) error
	UpdateQuota(ctx context.Context, quota *models.UploadQuota) error
	IncrementQuotaUsage(ctx context.Context, userID string, bytes int64, files int) error
	DecrementQuotaUsage(ctx context.Context, userID string, bytes int64, files int) error
	CheckQuotaAvailable(ctx context.Context, userID string, bytes int64) (bool, error)

	// Access Log operations
	LogAccess(ctx context.Context, log *models.AccessLog) error
	GetAccessLogs(ctx context.Context, filter *AccessLogFilter) ([]*models.AccessLog, int64, error)

	// Presigned URL operations
	CreatePresignedURL(ctx context.Context, url *models.PresignedURL) error
	GetPresignedURL(ctx context.Context, token string) (*models.PresignedURL, error)
	IncrementDownloadCount(ctx context.Context, token string) error
	DeleteExpiredPresignedURLs(ctx context.Context) (int64, error)

	// Cleanup Job operations
	CreateCleanupJob(ctx context.Context, job *models.CleanupJob) error
	UpdateCleanupJob(ctx context.Context, job *models.CleanupJob) error
	GetOrphanedFiles(ctx context.Context, retentionDays int) ([]*models.AttachmentFile, error)
	DeleteOrphanedFiles(ctx context.Context, hashes []string) (int64, error)

	// Statistics operations
	GetTotalStorageUsage(ctx context.Context) (int64, error)
	GetUserStorageUsage(ctx context.Context, userID string) (*models.UserStorageUsage, error)
	GetStorageStats(ctx context.Context) (*models.StorageStats, error)
}

// FileFilter defines filters for listing files
type FileFilter struct {
	MimeType  string
	MinSize   int64
	MaxSize   int64
	Deleted   *bool
	Limit     int
	Offset    int
	SortBy    string // "created", "size", "mime_type"
	SortOrder string // "asc", "desc"
}

// ReferenceFilter defines filters for listing references
type ReferenceFilter struct {
	EntityType string
	EntityID   string
	UploaderID string
	Tags       []string
	Deleted    *bool
	Limit      int
	Offset     int
	SortBy     string // "created", "modified", "filename"
	SortOrder  string // "asc", "desc"
}

// AccessLogFilter defines filters for access logs
type AccessLogFilter struct {
	UserID    string
	Action    string
	StartTime int64
	EndTime   int64
	Limit     int
	Offset    int
}

// DB is the concrete implementation of Database
type DB struct {
	conn   *sqlx.DB
	driver string
}

// New creates a new database connection
func New(cfg config.DatabaseConfig) (Database, error) {
	dsn := cfg.GetDSN()
	conn, err := sqlx.Connect(cfg.Driver, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Set connection pool parameters
	conn.SetMaxOpenConns(cfg.MaxConnections)
	conn.SetMaxIdleConns(cfg.IdleConnections)
	conn.SetConnMaxLifetime(time.Duration(cfg.ConnectionLifetime) * time.Second)
	conn.SetConnMaxIdleTime(300 * time.Second)

	// Verify connection
	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DB{
		conn:   conn,
		driver: cfg.Driver,
	}, nil
}

// Ping checks database connectivity
func (db *DB) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return db.conn.PingContext(ctx)
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.conn.Close()
}

// Migrate runs database migrations
func (db *DB) Migrate() error {
	// Check if schema_version table exists
	var exists bool
	query := "SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'schema_version')"
	if db.driver == "sqlite3" {
		query = "SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='schema_version'"
	}

	if err := db.conn.Get(&exists, query); err != nil {
		// Table doesn't exist, run initial migration
		return db.runInitialMigration()
	}

	// Get current schema version
	var currentVersion int
	if err := db.conn.Get(&currentVersion, "SELECT MAX(version) FROM schema_version"); err != nil {
		return db.runInitialMigration()
	}

	// Run pending migrations
	return db.runMigrations(currentVersion)
}

// runInitialMigration runs the initial schema creation
func (db *DB) runInitialMigration() error {
	// In production, we would embed SQL files from:
	// - Database/DDL/001_initial_schema.sql (postgres)
	// - Database/DDL/001_initial_schema_sqlite.sql (sqlite)
	// For now, we execute the schema directly
	return db.executeInitialSchema()
}

// runMigrations runs pending migrations
func (db *DB) runMigrations(currentVersion int) error {
	// Future migrations would be added here
	// For now, we only have version 1
	return nil
}

// executeInitialSchema executes the initial schema
func (db *DB) executeInitialSchema() error {
	// This would normally read from embedded SQL files
	// For now, return nil as migrations are handled externally
	return nil
}

// withContext creates a context with default timeout
func withContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 30*time.Second)
}

// Helper function to handle boolean for SQLite (INTEGER 0/1)
func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func intToBool(i int) bool {
	return i != 0
}
