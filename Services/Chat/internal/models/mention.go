package models

import (
	"time"

	"github.com/google/uuid"
)

// MessageMention represents a user mention in a message
type MessageMention struct {
	ID        uuid.UUID `json:"id" db:"id"`
	MessageID uuid.UUID `json:"message_id" db:"message_id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	Username  string    `json:"username" db:"username"`
	StartPos  int       `json:"start_pos" db:"start_pos"`
	EndPos    int       `json:"end_pos" db:"end_pos"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// MessageMentionRequest represents a request to add a mention
type MessageMentionRequest struct {
	MessageID uuid.UUID `json:"message_id" binding:"required"`
	UserID    uuid.UUID `json:"user_id" binding:"required"`
	StartPos  int       `json:"start_pos" binding:"required"`
	EndPos    int       `json:"end_pos" binding:"required"`
}

// MessageMentionResponse represents the response for mention operations
type MessageMentionResponse struct {
	Mention *MessageMention `json:"mention"`
}
