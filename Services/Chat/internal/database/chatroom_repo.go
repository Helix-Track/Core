package database

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"helixtrack.ru/chat/internal/models"
)

// ChatRoomCreate creates a new chat room
func (p *PostgresDB) ChatRoomCreate(ctx context.Context, room *models.ChatRoom) error {
	room.ID = uuid.New()
	room.CreatedAt = getNow()
	room.UpdatedAt = room.CreatedAt

	query := `
		INSERT INTO chat_room (
			id, name, description, type, entity_type, entity_id,
			created_by, is_private, is_archived, created_at, updated_at, deleted
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	_, err := p.db.ExecContext(ctx, query,
		room.ID, room.Name, room.Description, room.Type, room.EntityType, room.EntityID,
		room.CreatedBy, room.IsPrivate, room.IsArchived, room.CreatedAt, room.UpdatedAt, room.Deleted,
	)

	return err
}

// ChatRoomRead retrieves a chat room by ID
func (p *PostgresDB) ChatRoomRead(ctx context.Context, id string) (*models.ChatRoom, error) {
	room := &models.ChatRoom{}

	query := `
		SELECT id, name, description, type, entity_type, entity_id,
		       created_by, is_private, is_archived, created_at, updated_at, deleted, deleted_at
		FROM chat_room
		WHERE id = $1 AND deleted = false
	`

	err := p.db.QueryRowContext(ctx, query, id).Scan(
		&room.ID, &room.Name, &room.Description, &room.Type, &room.EntityType, &room.EntityID,
		&room.CreatedBy, &room.IsPrivate, &room.IsArchived, &room.CreatedAt, &room.UpdatedAt,
		&room.Deleted, &room.DeletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, models.ErrNotFound
	}

	return room, err
}

// ChatRoomUpdate updates a chat room
func (p *PostgresDB) ChatRoomUpdate(ctx context.Context, room *models.ChatRoom) error {
	room.UpdatedAt = getNow()

	query := `
		UPDATE chat_room
		SET name = $2, description = $3, is_private = $4, is_archived = $5, updated_at = $6
		WHERE id = $1 AND deleted = false
	`

	result, err := p.db.ExecContext(ctx, query,
		room.ID, room.Name, room.Description, room.IsPrivate, room.IsArchived, room.UpdatedAt,
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

// ChatRoomDelete soft deletes a chat room
func (p *PostgresDB) ChatRoomDelete(ctx context.Context, id string) error {
	now := getNow()

	query := `
		UPDATE chat_room
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

// ChatRoomList retrieves a paginated list of chat rooms
func (p *PostgresDB) ChatRoomList(ctx context.Context, limit, offset int) ([]*models.ChatRoom, int, error) {
	// Get total count
	var total int
	countQuery := `SELECT COUNT(*) FROM chat_room WHERE deleted = false`
	if err := p.db.QueryRowContext(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Get paginated results
	query := `
		SELECT id, name, description, type, entity_type, entity_id,
		       created_by, is_private, is_archived, created_at, updated_at, deleted, deleted_at
		FROM chat_room
		WHERE deleted = false
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := p.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var rooms []*models.ChatRoom
	for rows.Next() {
		room := &models.ChatRoom{}
		err := rows.Scan(
			&room.ID, &room.Name, &room.Description, &room.Type, &room.EntityType, &room.EntityID,
			&room.CreatedBy, &room.IsPrivate, &room.IsArchived, &room.CreatedAt, &room.UpdatedAt,
			&room.Deleted, &room.DeletedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		rooms = append(rooms, room)
	}

	return rooms, total, rows.Err()
}

// ChatRoomGetByEntity retrieves a chat room by entity type and ID
func (p *PostgresDB) ChatRoomGetByEntity(ctx context.Context, entityType string, entityID string) (*models.ChatRoom, error) {
	room := &models.ChatRoom{}

	query := `
		SELECT id, name, description, type, entity_type, entity_id,
		       created_by, is_private, is_archived, created_at, updated_at, deleted, deleted_at
		FROM chat_room
		WHERE entity_type = $1 AND entity_id = $2 AND deleted = false
		LIMIT 1
	`

	err := p.db.QueryRowContext(ctx, query, entityType, entityID).Scan(
		&room.ID, &room.Name, &room.Description, &room.Type, &room.EntityType, &room.EntityID,
		&room.CreatedBy, &room.IsPrivate, &room.IsArchived, &room.CreatedAt, &room.UpdatedAt,
		&room.Deleted, &room.DeletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, models.ErrNotFound
	}

	return room, err
}
