package database

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"helixtrack.ru/chat/internal/models"
)

// PresenceUpsert creates or updates user presence
func (p *PostgresDB) PresenceUpsert(ctx context.Context, presence *models.UserPresence) error {
	if presence.ID == uuid.Nil {
		presence.ID = uuid.New()
	}

	now := getNow()
	presence.LastSeen = now
	presence.UpdatedAt = now

	if presence.CreatedAt == 0 {
		presence.CreatedAt = now
	}

	query := `
		INSERT INTO user_presence (id, user_id, status, status_message, last_seen, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (user_id) DO UPDATE
		SET status = EXCLUDED.status,
		    status_message = EXCLUDED.status_message,
		    last_seen = EXCLUDED.last_seen,
		    updated_at = EXCLUDED.updated_at
	`

	_, err := p.db.ExecContext(ctx, query,
		presence.ID, presence.UserID, presence.Status, presence.StatusMessage,
		presence.LastSeen, presence.CreatedAt, presence.UpdatedAt,
	)

	return err
}

// PresenceGet retrieves user presence
func (p *PostgresDB) PresenceGet(ctx context.Context, userID string) (*models.UserPresence, error) {
	presence := &models.UserPresence{}

	query := `
		SELECT id, user_id, status, status_message, last_seen, created_at, updated_at
		FROM user_presence
		WHERE user_id = $1
	`

	err := p.db.QueryRowContext(ctx, query, userID).Scan(
		&presence.ID, &presence.UserID, &presence.Status, &presence.StatusMessage,
		&presence.LastSeen, &presence.CreatedAt, &presence.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, models.ErrNotFound
	}

	return presence, err
}

// PresenceGetMultiple retrieves presence for multiple users
func (p *PostgresDB) PresenceGetMultiple(ctx context.Context, userIDs []string) ([]*models.UserPresence, error) {
	if len(userIDs) == 0 {
		return []*models.UserPresence{}, nil
	}

	query := `
		SELECT id, user_id, status, status_message, last_seen, created_at, updated_at
		FROM user_presence
		WHERE user_id = ANY($1)
	`

	rows, err := p.db.QueryContext(ctx, query, userIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var presences []*models.UserPresence
	for rows.Next() {
		presence := &models.UserPresence{}
		err := rows.Scan(
			&presence.ID, &presence.UserID, &presence.Status, &presence.StatusMessage,
			&presence.LastSeen, &presence.CreatedAt, &presence.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		presences = append(presences, presence)
	}

	return presences, rows.Err()
}

// TypingUpsert creates or updates typing indicator
func (p *PostgresDB) TypingUpsert(ctx context.Context, indicator *models.TypingIndicator) error {
	if indicator.ID == uuid.Nil {
		indicator.ID = uuid.New()
	}

	now := getNow()
	indicator.StartedAt = now
	indicator.ExpiresAt = now + 5 // 5 seconds expiry

	query := `
		INSERT INTO typing_indicator (id, chat_room_id, user_id, is_typing, started_at, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (chat_room_id, user_id) DO UPDATE
		SET is_typing = EXCLUDED.is_typing,
		    started_at = EXCLUDED.started_at,
		    expires_at = EXCLUDED.expires_at
	`

	_, err := p.db.ExecContext(ctx, query,
		indicator.ID, indicator.ChatRoomID, indicator.UserID, indicator.IsTyping,
		indicator.StartedAt, indicator.ExpiresAt,
	)

	return err
}

// TypingDelete removes typing indicator
func (p *PostgresDB) TypingDelete(ctx context.Context, chatRoomID, userID string) error {
	query := `DELETE FROM typing_indicator WHERE chat_room_id = $1 AND user_id = $2`
	_, err := p.db.ExecContext(ctx, query, chatRoomID, userID)
	return err
}

// TypingGetActive retrieves active typing indicators for a chat room
func (p *PostgresDB) TypingGetActive(ctx context.Context, chatRoomID string) ([]*models.TypingIndicator, error) {
	now := getNow()

	query := `
		SELECT id, chat_room_id, user_id, is_typing, started_at, expires_at
		FROM typing_indicator
		WHERE chat_room_id = $1 AND is_typing = true AND expires_at > $2
		ORDER BY started_at DESC
	`

	rows, err := p.db.QueryContext(ctx, query, chatRoomID, now)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var indicators []*models.TypingIndicator
	for rows.Next() {
		indicator := &models.TypingIndicator{}
		err := rows.Scan(
			&indicator.ID, &indicator.ChatRoomID, &indicator.UserID, &indicator.IsTyping,
			&indicator.StartedAt, &indicator.ExpiresAt,
		)
		if err != nil {
			return nil, err
		}
		indicators = append(indicators, indicator)
	}

	return indicators, rows.Err()
}

// ReadReceiptCreate creates a read receipt
func (p *PostgresDB) ReadReceiptCreate(ctx context.Context, receipt *models.MessageReadReceipt) error {
	if receipt.ID == uuid.Nil {
		receipt.ID = uuid.New()
	}

	now := getNow()
	receipt.ReadAt = now
	receipt.CreatedAt = now

	query := `
		INSERT INTO message_read_receipt (id, message_id, user_id, read_at, created_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (message_id, user_id) DO UPDATE
		SET read_at = EXCLUDED.read_at
	`

	_, err := p.db.ExecContext(ctx, query,
		receipt.ID, receipt.MessageID, receipt.UserID, receipt.ReadAt, receipt.CreatedAt,
	)

	return err
}

// ReadReceiptGet retrieves all read receipts for a message
func (p *PostgresDB) ReadReceiptGet(ctx context.Context, messageID string) ([]*models.MessageReadReceipt, error) {
	query := `
		SELECT id, message_id, user_id, read_at, created_at
		FROM message_read_receipt
		WHERE message_id = $1
		ORDER BY read_at ASC
	`

	rows, err := p.db.QueryContext(ctx, query, messageID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var receipts []*models.MessageReadReceipt
	for rows.Next() {
		receipt := &models.MessageReadReceipt{}
		err := rows.Scan(
			&receipt.ID, &receipt.MessageID, &receipt.UserID, &receipt.ReadAt, &receipt.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		receipts = append(receipts, receipt)
	}

	return receipts, rows.Err()
}

// ReadReceiptGetByUser retrieves a specific read receipt
func (p *PostgresDB) ReadReceiptGetByUser(ctx context.Context, messageID, userID string) (*models.MessageReadReceipt, error) {
	receipt := &models.MessageReadReceipt{}

	query := `
		SELECT id, message_id, user_id, read_at, created_at
		FROM message_read_receipt
		WHERE message_id = $1 AND user_id = $2
	`

	err := p.db.QueryRowContext(ctx, query, messageID, userID).Scan(
		&receipt.ID, &receipt.MessageID, &receipt.UserID, &receipt.ReadAt, &receipt.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, models.ErrNotFound
	}

	return receipt, err
}

// ReactionCreate creates a message reaction
func (p *PostgresDB) ReactionCreate(ctx context.Context, reaction *models.MessageReaction) error {
	if reaction.ID == uuid.Nil {
		reaction.ID = uuid.New()
	}

	reaction.CreatedAt = getNow()

	query := `
		INSERT INTO message_reaction (id, message_id, user_id, emoji, created_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (message_id, user_id, emoji) DO NOTHING
	`

	_, err := p.db.ExecContext(ctx, query,
		reaction.ID, reaction.MessageID, reaction.UserID, reaction.Emoji, reaction.CreatedAt,
	)

	return err
}

// ReactionDelete removes a message reaction
func (p *PostgresDB) ReactionDelete(ctx context.Context, messageID, userID, emoji string) error {
	query := `DELETE FROM message_reaction WHERE message_id = $1 AND user_id = $2 AND emoji = $3`
	_, err := p.db.ExecContext(ctx, query, messageID, userID, emoji)
	return err
}

// ReactionList retrieves all reactions for a message
func (p *PostgresDB) ReactionList(ctx context.Context, messageID string) ([]*models.MessageReaction, error) {
	query := `
		SELECT id, message_id, user_id, emoji, created_at
		FROM message_reaction
		WHERE message_id = $1
		ORDER BY created_at ASC
	`

	rows, err := p.db.QueryContext(ctx, query, messageID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reactions []*models.MessageReaction
	for rows.Next() {
		reaction := &models.MessageReaction{}
		err := rows.Scan(
			&reaction.ID, &reaction.MessageID, &reaction.UserID, &reaction.Emoji, &reaction.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		reactions = append(reactions, reaction)
	}

	return reactions, rows.Err()
}

// AttachmentCreate creates a message attachment
func (p *PostgresDB) AttachmentCreate(ctx context.Context, attachment *models.MessageAttachment) error {
	if attachment.ID == uuid.Nil {
		attachment.ID = uuid.New()
	}

	attachment.CreatedAt = getNow()

	query := `
		INSERT INTO message_attachment (
			id, message_id, file_name, file_type, file_size, file_url,
			thumbnail_url, metadata, uploaded_by, created_at, deleted
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err := p.db.ExecContext(ctx, query,
		attachment.ID, attachment.MessageID, attachment.FileName, attachment.FileType,
		attachment.FileSize, attachment.FileURL, attachment.ThumbnailURL, attachment.Metadata,
		attachment.UploadedBy, attachment.CreatedAt, attachment.Deleted,
	)

	return err
}

// AttachmentDelete soft deletes an attachment
func (p *PostgresDB) AttachmentDelete(ctx context.Context, id string) error {
	now := getNow()

	query := `
		UPDATE message_attachment
		SET deleted = true, deleted_at = $2
		WHERE id = $1 AND deleted = false
	`

	result, err := p.db.ExecContext(ctx, query, id, now)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return models.ErrNotFound
	}

	return nil
}

// AttachmentList retrieves all attachments for a message
func (p *PostgresDB) AttachmentList(ctx context.Context, messageID string) ([]*models.MessageAttachment, error) {
	query := `
		SELECT id, message_id, file_name, file_type, file_size, file_url,
		       thumbnail_url, metadata, uploaded_by, created_at, deleted, deleted_at
		FROM message_attachment
		WHERE message_id = $1 AND deleted = false
		ORDER BY created_at ASC
	`

	rows, err := p.db.QueryContext(ctx, query, messageID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var attachments []*models.MessageAttachment
	for rows.Next() {
		attachment := &models.MessageAttachment{}
		err := rows.Scan(
			&attachment.ID, &attachment.MessageID, &attachment.FileName, &attachment.FileType,
			&attachment.FileSize, &attachment.FileURL, &attachment.ThumbnailURL, &attachment.Metadata,
			&attachment.UploadedBy, &attachment.CreatedAt, &attachment.Deleted, &attachment.DeletedAt,
		)
		if err != nil {
			return nil, err
		}
		attachments = append(attachments, attachment)
	}

	return attachments, rows.Err()
}

// CleanupExpiredTypingIndicators removes expired typing indicators (run periodically)
func (p *PostgresDB) CleanupExpiredTypingIndicators(ctx context.Context) error {
	now := time.Now().Unix()
	query := `DELETE FROM typing_indicator WHERE expires_at <= $1`
	_, err := p.db.ExecContext(ctx, query, now)
	return err
}
