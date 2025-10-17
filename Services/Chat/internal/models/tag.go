package models

import (
	"time"

	"github.com/google/uuid"
)

// MessageTag represents a tag associated with a message
type MessageTag struct {
	ID        uuid.UUID `json:"id" db:"id"`
	MessageID uuid.UUID `json:"message_id" db:"message_id"`
	Tag       string    `json:"tag" db:"tag"`
	Color     string    `json:"color" db:"color"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	CreatedBy uuid.UUID `json:"created_by" db:"created_by"`
}

// MessageTagRequest represents a request to add a tag
type MessageTagRequest struct {
	MessageID uuid.UUID `json:"message_id" binding:"required"`
	Tag       string    `json:"tag" binding:"required"`
	Color     string    `json:"color"`
}

// MessageTagResponse represents the response for tag operations
type MessageTagResponse struct {
	Tag *MessageTag `json:"tag"`
}

// ChatRoomTag represents a tag for a chat room
type ChatRoomTag struct {
	ID         uuid.UUID `json:"id" db:"id"`
	ChatRoomID uuid.UUID `json:"chat_room_id" db:"chat_room_id"`
	Tag        string    `json:"tag" db:"tag"`
	Color      string    `json:"color" db:"color"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	CreatedBy  uuid.UUID `json:"created_by" db:"created_by"`
}

// ChatRoomTagRequest represents a request to add a room tag
type ChatRoomTagRequest struct {
	ChatRoomID uuid.UUID `json:"chat_room_id" binding:"required"`
	Tag        string    `json:"tag" binding:"required"`
	Color      string    `json:"color"`
}

// ChatRoomTagResponse represents the response for room tag operations
type ChatRoomTagResponse struct {
	Tag *ChatRoomTag `json:"tag"`
}
