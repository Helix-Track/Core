package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	"go.uber.org/zap"

	"helixtrack.ru/chat/internal/logger"
	"helixtrack.ru/chat/internal/models"
)

// Database interface defines all database operations
type Database interface {
	// Connection management
	Close() error
	Ping() error
	BeginTx(ctx context.Context) (*sql.Tx, error)

	// Chat Room operations
	ChatRoomCreate(ctx context.Context, room *models.ChatRoom) error
	ChatRoomRead(ctx context.Context, id string) (*models.ChatRoom, error)
	ChatRoomUpdate(ctx context.Context, room *models.ChatRoom) error
	ChatRoomDelete(ctx context.Context, id string) error
	ChatRoomList(ctx context.Context, limit, offset int) ([]*models.ChatRoom, int, error)
	ChatRoomGetByEntity(ctx context.Context, entityType string, entityID string) (*models.ChatRoom, error)

	// Message operations
	MessageCreate(ctx context.Context, message *models.Message) error
	MessageRead(ctx context.Context, id string) (*models.Message, error)
	MessageUpdate(ctx context.Context, message *models.Message) error
	MessageDelete(ctx context.Context, id string) error
	MessageList(ctx context.Context, req *models.MessageListRequest) ([]*models.Message, int, error)
	MessageSearch(ctx context.Context, chatRoomID, query string, limit, offset int) ([]*models.Message, int, error)

	// Participant operations
	ParticipantAdd(ctx context.Context, participant *models.ChatParticipant) error
	ParticipantRemove(ctx context.Context, chatRoomID, userID string) error
	ParticipantList(ctx context.Context, chatRoomID string) ([]*models.ChatParticipant, error)
	ParticipantUpdate(ctx context.Context, participant *models.ChatParticipant) error
	ParticipantGet(ctx context.Context, chatRoomID, userID string) (*models.ChatParticipant, error)
	ParticipantMute(ctx context.Context, chatRoomID, userID string) error
	ParticipantUnmute(ctx context.Context, chatRoomID, userID string) error

	// Presence operations
	PresenceUpsert(ctx context.Context, presence *models.UserPresence) error
	PresenceGet(ctx context.Context, userID string) (*models.UserPresence, error)
	PresenceGetMultiple(ctx context.Context, userIDs []string) ([]*models.UserPresence, error)

	// Typing indicator operations
	TypingUpsert(ctx context.Context, indicator *models.TypingIndicator) error
	TypingDelete(ctx context.Context, chatRoomID, userID string) error
	TypingGetActive(ctx context.Context, chatRoomID string) ([]*models.TypingIndicator, error)

	// Read receipt operations
	ReadReceiptCreate(ctx context.Context, receipt *models.MessageReadReceipt) error
	ReadReceiptGet(ctx context.Context, messageID string) ([]*models.MessageReadReceipt, error)
	ReadReceiptGetByUser(ctx context.Context, messageID, userID string) (*models.MessageReadReceipt, error)

	// Reaction operations
	ReactionCreate(ctx context.Context, reaction *models.MessageReaction) error
	ReactionDelete(ctx context.Context, messageID, userID, emoji string) error
	ReactionList(ctx context.Context, messageID string) ([]*models.MessageReaction, error)

	// Attachment operations
	AttachmentCreate(ctx context.Context, attachment *models.MessageAttachment) error
	AttachmentDelete(ctx context.Context, id string) error
	AttachmentList(ctx context.Context, messageID string) ([]*models.MessageAttachment, error)

	// Message edit history operations
	MessageEditHistoryCreate(ctx context.Context, history *models.MessageEditHistory) error
	MessageEditHistoryList(ctx context.Context, messageID string) ([]*models.MessageEditHistory, error)
	MessageEditHistoryGet(ctx context.Context, id string) (*models.MessageEditHistory, error)
	MessageEditHistoryCount(ctx context.Context, messageID string) (int, error)
}

// PostgresDB implements the Database interface for PostgreSQL
type PostgresDB struct {
	db *sql.DB
}

// NewPostgresDB creates a new PostgreSQL database connection
func NewPostgresDB(config *models.DatabaseConfig) (*PostgresDB, error) {
	// Build connection string
	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s connect_timeout=%d",
		config.Host,
		config.Port,
		config.User,
		config.Password,
		config.Database,
		config.SSLMode,
		config.ConnectionTimeout,
	)

	// Open database connection
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(config.MaxConnections)
	db.SetMaxIdleConns(config.MaxConnections / 2)
	db.SetConnMaxLifetime(time.Hour)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("Database connected",
		zap.String("host", config.Host),
		zap.Int("port", config.Port),
		zap.String("database", config.Database),
	)

	return &PostgresDB{db: db}, nil
}

// Close closes the database connection
func (p *PostgresDB) Close() error {
	if p.db != nil {
		return p.db.Close()
	}
	return nil
}

// Ping checks the database connection
func (p *PostgresDB) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return p.db.PingContext(ctx)
}

// BeginTx starts a new transaction
func (p *PostgresDB) BeginTx(ctx context.Context) (*sql.Tx, error) {
	return p.db.BeginTx(ctx, nil)
}

// Helper function to get current Unix timestamp
func getNow() int64 {
	return time.Now().Unix()
}
