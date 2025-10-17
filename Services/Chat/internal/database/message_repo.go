package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"helixtrack.ru/chat/internal/models"
)

// MessageCreate creates a new message
func (p *PostgresDB) MessageCreate(ctx context.Context, message *models.Message) error {
	message.ID = uuid.New()
	message.CreatedAt = getNow()
	message.UpdatedAt = message.CreatedAt

	query := `
		INSERT INTO message (
			id, chat_room_id, sender_id, parent_id, quoted_message_id,
			type, content, content_format, metadata, is_edited, is_pinned,
			created_at, updated_at, deleted
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`

	_, err := p.db.ExecContext(ctx, query,
		message.ID, message.ChatRoomID, message.SenderID, message.ParentID, message.QuotedMessageID,
		message.Type, message.Content, message.ContentFormat, message.Metadata, message.IsEdited,
		message.IsPinned, message.CreatedAt, message.UpdatedAt, message.Deleted,
	)

	return err
}

// MessageRead retrieves a message by ID
func (p *PostgresDB) MessageRead(ctx context.Context, id string) (*models.Message, error) {
	message := &models.Message{}

	query := `
		SELECT id, chat_room_id, sender_id, parent_id, quoted_message_id,
		       type, content, content_format, metadata, is_edited, edited_at,
		       is_pinned, pinned_at, pinned_by, created_at, updated_at, deleted, deleted_at
		FROM message
		WHERE id = $1 AND deleted = false
	`

	err := p.db.QueryRowContext(ctx, query, id).Scan(
		&message.ID, &message.ChatRoomID, &message.SenderID, &message.ParentID, &message.QuotedMessageID,
		&message.Type, &message.Content, &message.ContentFormat, &message.Metadata, &message.IsEdited,
		&message.EditedAt, &message.IsPinned, &message.PinnedAt, &message.PinnedBy, &message.CreatedAt,
		&message.UpdatedAt, &message.Deleted, &message.DeletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, models.ErrNotFound
	}

	return message, err
}

// MessageUpdate updates a message
func (p *PostgresDB) MessageUpdate(ctx context.Context, message *models.Message) error {
	now := getNow()
	message.UpdatedAt = now
	message.IsEdited = true
	message.EditedAt = &now

	query := `
		UPDATE message
		SET content = $2, content_format = $3, metadata = $4, is_edited = $5,
		    edited_at = $6, updated_at = $7
		WHERE id = $1 AND deleted = false
	`

	result, err := p.db.ExecContext(ctx, query,
		message.ID, message.Content, message.ContentFormat, message.Metadata,
		message.IsEdited, message.EditedAt, message.UpdatedAt,
	)

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

// MessageDelete soft deletes a message
func (p *PostgresDB) MessageDelete(ctx context.Context, id string) error {
	now := getNow()

	query := `
		UPDATE message
		SET deleted = true, deleted_at = $2, updated_at = $2
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

// MessageList retrieves paginated messages for a chat room
func (p *PostgresDB) MessageList(ctx context.Context, req *models.MessageListRequest) ([]*models.Message, int, error) {
	// Build query conditions
	conditions := "chat_room_id = $1 AND deleted = false"
	args := []interface{}{req.ChatRoomID}
	argPos := 2

	if req.ParentID != nil {
		conditions += fmt.Sprintf(" AND parent_id = $%d", argPos)
		args = append(args, *req.ParentID)
		argPos++
	}

	if req.Before != nil {
		conditions += fmt.Sprintf(" AND created_at < $%d", argPos)
		args = append(args, *req.Before)
		argPos++
	}

	if req.After != nil {
		conditions += fmt.Sprintf(" AND created_at > $%d", argPos)
		args = append(args, *req.After)
		argPos++
	}

	// Get total count
	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM message WHERE %s", conditions)
	if err := p.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Get paginated results
	limit := req.Limit
	if limit <= 0 || limit > 100 {
		limit = 50 // Default
	}

	offset := req.Offset
	if offset < 0 {
		offset = 0
	}

	query := fmt.Sprintf(`
		SELECT id, chat_room_id, sender_id, parent_id, quoted_message_id,
		       type, content, content_format, metadata, is_edited, edited_at,
		       is_pinned, pinned_at, pinned_by, created_at, updated_at, deleted, deleted_at
		FROM message
		WHERE %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, conditions, argPos, argPos+1)

	args = append(args, limit, offset)

	rows, err := p.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var messages []*models.Message
	for rows.Next() {
		message := &models.Message{}
		err := rows.Scan(
			&message.ID, &message.ChatRoomID, &message.SenderID, &message.ParentID, &message.QuotedMessageID,
			&message.Type, &message.Content, &message.ContentFormat, &message.Metadata, &message.IsEdited,
			&message.EditedAt, &message.IsPinned, &message.PinnedAt, &message.PinnedBy, &message.CreatedAt,
			&message.UpdatedAt, &message.Deleted, &message.DeletedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		messages = append(messages, message)
	}

	return messages, total, rows.Err()
}

// MessageSearch performs full-text search on messages
func (p *PostgresDB) MessageSearch(ctx context.Context, chatRoomID, query string, limit, offset int) ([]*models.Message, int, error) {
	// Get total count
	var total int
	countQuery := `
		SELECT COUNT(*)
		FROM message
		WHERE chat_room_id = $1 AND deleted = false
		  AND to_tsvector('english', content) @@ plainto_tsquery('english', $2)
	`
	if err := p.db.QueryRowContext(ctx, countQuery, chatRoomID, query).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Get paginated results
	searchQuery := `
		SELECT id, chat_room_id, sender_id, parent_id, quoted_message_id,
		       type, content, content_format, metadata, is_edited, edited_at,
		       is_pinned, pinned_at, pinned_by, created_at, updated_at, deleted, deleted_at
		FROM message
		WHERE chat_room_id = $1 AND deleted = false
		  AND to_tsvector('english', content) @@ plainto_tsquery('english', $2)
		ORDER BY ts_rank(to_tsvector('english', content), plainto_tsquery('english', $2)) DESC,
		         created_at DESC
		LIMIT $3 OFFSET $4
	`

	rows, err := p.db.QueryContext(ctx, searchQuery, chatRoomID, query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var messages []*models.Message
	for rows.Next() {
		message := &models.Message{}
		err := rows.Scan(
			&message.ID, &message.ChatRoomID, &message.SenderID, &message.ParentID, &message.QuotedMessageID,
			&message.Type, &message.Content, &message.ContentFormat, &message.Metadata, &message.IsEdited,
			&message.EditedAt, &message.IsPinned, &message.PinnedAt, &message.PinnedBy, &message.CreatedAt,
			&message.UpdatedAt, &message.Deleted, &message.DeletedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		messages = append(messages, message)
	}

	return messages, total, rows.Err()
}

// MessageEditHistoryCreate creates a new edit history entry
func (p *PostgresDB) MessageEditHistoryCreate(ctx context.Context, history *models.MessageEditHistory) error {
	history.ID = uuid.New()
	history.CreatedAt = getNow()

	query := `
		INSERT INTO message_edit_history (
			id, message_id, previous_content, previous_content_format,
			previous_metadata, editor_id, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := p.db.ExecContext(ctx, query,
		history.ID, history.MessageID, history.PreviousContent,
		history.PreviousContentFormat, history.PreviousMetadata,
		history.EditorID, history.CreatedAt,
	)

	return err
}

// MessageEditHistoryList retrieves edit history for a message
func (p *PostgresDB) MessageEditHistoryList(ctx context.Context, messageID string) ([]*models.MessageEditHistory, error) {
	query := `
		SELECT id, message_id, previous_content, previous_content_format,
		       previous_metadata, editor_id, created_at
		FROM message_edit_history
		WHERE message_id = $1
		ORDER BY created_at DESC
	`

	rows, err := p.db.QueryContext(ctx, query, messageID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var histories []*models.MessageEditHistory
	for rows.Next() {
		history := &models.MessageEditHistory{}
		err := rows.Scan(
			&history.ID, &history.MessageID, &history.PreviousContent,
			&history.PreviousContentFormat, &history.PreviousMetadata,
			&history.EditorID, &history.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		histories = append(histories, history)
	}

	return histories, rows.Err()
}

// MessageEditHistoryGet retrieves a specific edit history entry
func (p *PostgresDB) MessageEditHistoryGet(ctx context.Context, id string) (*models.MessageEditHistory, error) {
	history := &models.MessageEditHistory{}

	query := `
		SELECT id, message_id, previous_content, previous_content_format,
		       previous_metadata, editor_id, created_at
		FROM message_edit_history
		WHERE id = $1
	`

	err := p.db.QueryRowContext(ctx, query, id).Scan(
		&history.ID, &history.MessageID, &history.PreviousContent,
		&history.PreviousContentFormat, &history.PreviousMetadata,
		&history.EditorID, &history.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, models.ErrNotFound
	}

	return history, err
}

// MessageEditHistoryCount returns the count of edit history entries for a message
func (p *PostgresDB) MessageEditHistoryCount(ctx context.Context, messageID string) (int, error) {
	var count int
	query := `
		SELECT COUNT(*)
		FROM message_edit_history
		WHERE message_id = $1
	`
	err := p.db.QueryRowContext(ctx, query, messageID).Scan(&count)
	return count, err
}
