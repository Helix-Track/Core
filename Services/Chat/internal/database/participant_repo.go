package database

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"helixtrack.ru/chat/internal/models"
)

// ParticipantAdd adds a user to a chat room
func (p *PostgresDB) ParticipantAdd(ctx context.Context, participant *models.ChatParticipant) error {
	participant.ID = uuid.New()
	participant.CreatedAt = getNow()
	participant.UpdatedAt = participant.CreatedAt
	participant.JoinedAt = participant.CreatedAt

	query := `
		INSERT INTO chat_participant (
			id, chat_room_id, user_id, role, is_muted, joined_at, created_at, updated_at, deleted
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (chat_room_id, user_id) DO UPDATE
		SET deleted = false, left_at = NULL, updated_at = EXCLUDED.updated_at
	`

	_, err := p.db.ExecContext(ctx, query,
		participant.ID, participant.ChatRoomID, participant.UserID, participant.Role,
		participant.IsMuted, participant.JoinedAt, participant.CreatedAt, participant.UpdatedAt,
		participant.Deleted,
	)

	return err
}

// ParticipantRemove removes a user from a chat room
func (p *PostgresDB) ParticipantRemove(ctx context.Context, chatRoomID, userID string) error {
	now := getNow()

	query := `
		UPDATE chat_participant
		SET deleted = true, deleted_at = $3, left_at = $3, updated_at = $3
		WHERE chat_room_id = $1 AND user_id = $2 AND deleted = false
	`

	result, err := p.db.ExecContext(ctx, query, chatRoomID, userID, now)
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

// ParticipantList retrieves all participants in a chat room
func (p *PostgresDB) ParticipantList(ctx context.Context, chatRoomID string) ([]*models.ChatParticipant, error) {
	query := `
		SELECT id, chat_room_id, user_id, role, is_muted, joined_at, left_at,
		       created_at, updated_at, deleted, deleted_at
		FROM chat_participant
		WHERE chat_room_id = $1 AND deleted = false
		ORDER BY joined_at ASC
	`

	rows, err := p.db.QueryContext(ctx, query, chatRoomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var participants []*models.ChatParticipant
	for rows.Next() {
		participant := &models.ChatParticipant{}
		err := rows.Scan(
			&participant.ID, &participant.ChatRoomID, &participant.UserID, &participant.Role,
			&participant.IsMuted, &participant.JoinedAt, &participant.LeftAt, &participant.CreatedAt,
			&participant.UpdatedAt, &participant.Deleted, &participant.DeletedAt,
		)
		if err != nil {
			return nil, err
		}
		participants = append(participants, participant)
	}

	return participants, rows.Err()
}

// ParticipantUpdate updates a participant's role or mute status
func (p *PostgresDB) ParticipantUpdate(ctx context.Context, participant *models.ChatParticipant) error {
	participant.UpdatedAt = getNow()

	query := `
		UPDATE chat_participant
		SET role = $3, is_muted = $4, updated_at = $5
		WHERE chat_room_id = $1 AND user_id = $2 AND deleted = false
	`

	result, err := p.db.ExecContext(ctx, query,
		participant.ChatRoomID, participant.UserID, participant.Role,
		participant.IsMuted, participant.UpdatedAt,
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

// ParticipantGet retrieves a specific participant
func (p *PostgresDB) ParticipantGet(ctx context.Context, chatRoomID, userID string) (*models.ChatParticipant, error) {
	participant := &models.ChatParticipant{}

	query := `
		SELECT id, chat_room_id, user_id, role, is_muted, joined_at, left_at,
		       created_at, updated_at, deleted, deleted_at
		FROM chat_participant
		WHERE chat_room_id = $1 AND user_id = $2 AND deleted = false
	`

	err := p.db.QueryRowContext(ctx, query, chatRoomID, userID).Scan(
		&participant.ID, &participant.ChatRoomID, &participant.UserID, &participant.Role,
		&participant.IsMuted, &participant.JoinedAt, &participant.LeftAt, &participant.CreatedAt,
		&participant.UpdatedAt, &participant.Deleted, &participant.DeletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, models.ErrNotFound
	}

	return participant, err
}

// ParticipantMute mutes a participant in a chat room
func (p *PostgresDB) ParticipantMute(ctx context.Context, chatRoomID, userID string) error {
	query := `
		UPDATE chat_participant
		SET is_muted = true, updated_at = NOW()
		WHERE chat_room_id = $1 AND user_id = $2 AND deleted = false
	`

	result, err := p.db.ExecContext(ctx, query, chatRoomID, userID)
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

// ParticipantUnmute unmutes a participant in a chat room
func (p *PostgresDB) ParticipantUnmute(ctx context.Context, chatRoomID, userID string) error {
	query := `
		UPDATE chat_participant
		SET is_muted = false, updated_at = NOW()
		WHERE chat_room_id = $1 AND user_id = $2 AND deleted = false
	`

	result, err := p.db.ExecContext(ctx, query, chatRoomID, userID)
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
