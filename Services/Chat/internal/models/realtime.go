package models

import "github.com/google/uuid"

// PresenceStatus represents the user's online status
type PresenceStatus string

const (
	PresenceStatusOnline  PresenceStatus = "online"
	PresenceStatusOffline PresenceStatus = "offline"
	PresenceStatusAway    PresenceStatus = "away"
	PresenceStatusBusy    PresenceStatus = "busy"
	PresenceStatusDND     PresenceStatus = "dnd" // Do Not Disturb
)

// UserPresence represents user's online/offline status
type UserPresence struct {
	ID            uuid.UUID      `json:"id" db:"id"`
	UserID        uuid.UUID      `json:"user_id" db:"user_id"`
	Status        PresenceStatus `json:"status" db:"status"`
	StatusMessage string         `json:"status_message,omitempty" db:"status_message"`
	LastSeen      int64          `json:"last_seen" db:"last_seen"`
	CreatedAt     int64          `json:"created_at" db:"created_at"`
	UpdatedAt     int64          `json:"updated_at" db:"updated_at"`
}

// TypingIndicator represents a user typing in a chat room
type TypingIndicator struct {
	ID         uuid.UUID `json:"id" db:"id"`
	ChatRoomID uuid.UUID `json:"chat_room_id" db:"chat_room_id"`
	UserID     uuid.UUID `json:"user_id" db:"user_id"`
	IsTyping   bool      `json:"is_typing" db:"is_typing"`
	StartedAt  int64     `json:"started_at" db:"started_at"`
	ExpiresAt  int64     `json:"expires_at" db:"expires_at"`
}

// MessageReadReceipt represents a message read receipt
type MessageReadReceipt struct {
	ID        uuid.UUID `json:"id" db:"id"`
	MessageID uuid.UUID `json:"message_id" db:"message_id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	ReadAt    int64     `json:"read_at" db:"read_at"`
	CreatedAt int64     `json:"created_at" db:"created_at"`
}

// MessageReaction represents an emoji reaction to a message
type MessageReaction struct {
	ID        uuid.UUID `json:"id" db:"id"`
	MessageID uuid.UUID `json:"message_id" db:"message_id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	Emoji     string    `json:"emoji" db:"emoji"`
	CreatedAt int64     `json:"created_at" db:"created_at"`
}

// MessageAttachment represents a file attached to a message
type MessageAttachment struct {
	ID           uuid.UUID `json:"id" db:"id"`
	MessageID    uuid.UUID `json:"message_id" db:"message_id"`
	FileName     string    `json:"file_name" db:"file_name"`
	FileType     string    `json:"file_type" db:"file_type"`
	FileSize     int64     `json:"file_size" db:"file_size"`
	FileURL      string    `json:"file_url" db:"file_url"`
	ThumbnailURL string    `json:"thumbnail_url,omitempty" db:"thumbnail_url"`
	Metadata     []byte    `json:"metadata,omitempty" db:"metadata"`
	UploadedBy   uuid.UUID `json:"uploaded_by" db:"uploaded_by"`
	CreatedAt    int64     `json:"created_at" db:"created_at"`
	Deleted      bool      `json:"deleted" db:"deleted"`
	DeletedAt    *int64    `json:"deleted_at,omitempty" db:"deleted_at"`
}

// TypingRequest represents a typing indicator request
type TypingRequest struct {
	ChatRoomID uuid.UUID `json:"chat_room_id" binding:"required"`
	IsTyping   bool      `json:"is_typing"`
}

// ReadReceiptRequest represents a read receipt request
type ReadReceiptRequest struct {
	MessageID uuid.UUID `json:"message_id" binding:"required"`
}

// ReactionRequest represents a reaction request
type ReactionRequest struct {
	MessageID uuid.UUID `json:"message_id" binding:"required"`
	Emoji     string    `json:"emoji" binding:"required,min=1,max=50"`
}

// PresenceRequest represents a presence update request
type PresenceRequest struct {
	Status        PresenceStatus `json:"status" binding:"required"`
	StatusMessage string         `json:"status_message"`
}

// Validate checks if the presence request is valid
func (r *PresenceRequest) Validate() error {
	validStatuses := map[PresenceStatus]bool{
		PresenceStatusOnline:  true,
		PresenceStatusOffline: true,
		PresenceStatusAway:    true,
		PresenceStatusBusy:    true,
		PresenceStatusDND:     true,
	}

	if !validStatuses[r.Status] {
		return ErrInvalidPresenceStatus
	}

	return nil
}
