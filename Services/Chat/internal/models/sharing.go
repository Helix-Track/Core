package models

import (
	"time"

	"github.com/google/uuid"
)

// ChatShare represents a shared chat link
type ChatShare struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	ChatRoomID  uuid.UUID  `json:"chat_room_id" db:"chat_room_id"`
	SharedBy    uuid.UUID  `json:"shared_by" db:"shared_by"`
	ShareToken  string     `json:"share_token" db:"share_token"`
	ExpiresAt   *time.Time `json:"expires_at" db:"expires_at"`
	AccessLevel string     `json:"access_level" db:"access_level"` // "read", "write"
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	IsActive    bool       `json:"is_active" db:"is_active"`
}

// ChatShareRequest represents a request to create a share link
type ChatShareRequest struct {
	ChatRoomID  uuid.UUID  `json:"chat_room_id" binding:"required"`
	ExpiresAt   *time.Time `json:"expires_at"`
	AccessLevel string     `json:"access_level" binding:"required"` // "read" or "write"
}

// ChatShareResponse represents the response for share operations
type ChatShareResponse struct {
	Share *ChatShare `json:"share"`
}

// ChatShareAccess represents access via a share link
type ChatShareAccess struct {
	ID         uuid.UUID `json:"id" db:"id"`
	ShareID    uuid.UUID `json:"share_id" db:"share_id"`
	UserID     uuid.UUID `json:"user_id" db:"user_id"`
	AccessedAt time.Time `json:"accessed_at" db:"accessed_at"`
	IPAddress  string    `json:"ip_address" db:"ip_address"`
	UserAgent  string    `json:"user_agent" db:"user_agent"`
}
