package models

import "github.com/google/uuid"

// ChatRoomType represents the type of chat room
type ChatRoomType string

const (
	ChatRoomTypeDirect       ChatRoomType = "direct"
	ChatRoomTypeGroup        ChatRoomType = "group"
	ChatRoomTypeTeam         ChatRoomType = "team"
	ChatRoomTypeProject      ChatRoomType = "project"
	ChatRoomTypeTicket       ChatRoomType = "ticket"
	ChatRoomTypeAccount      ChatRoomType = "account"
	ChatRoomTypeOrganization ChatRoomType = "organization"
	ChatRoomTypeAttachment   ChatRoomType = "attachment"
	ChatRoomTypeCustom       ChatRoomType = "custom"
)

// ChatRoom represents a chat room/channel
type ChatRoom struct {
	ID          uuid.UUID    `json:"id" db:"id"`
	Name        string       `json:"name" db:"name"`
	Description string       `json:"description" db:"description"`
	Type        ChatRoomType `json:"type" db:"type"`
	EntityType  string       `json:"entity_type,omitempty" db:"entity_type"`
	EntityID    *uuid.UUID   `json:"entity_id,omitempty" db:"entity_id"`
	CreatedBy   uuid.UUID    `json:"created_by" db:"created_by"`
	IsPrivate   bool         `json:"is_private" db:"is_private"`
	IsArchived  bool         `json:"is_archived" db:"is_archived"`
	CreatedAt   int64        `json:"created_at" db:"created_at"`
	UpdatedAt   int64        `json:"updated_at" db:"updated_at"`
	Deleted     bool         `json:"deleted" db:"deleted"`
	DeletedAt   *int64       `json:"deleted_at,omitempty" db:"deleted_at"`
}

// ChatRoomRequest represents the request to create/update a chat room
type ChatRoomRequest struct {
	Name        string       `json:"name" binding:"required,min=1,max=255"`
	Description string       `json:"description"`
	Type        ChatRoomType `json:"type" binding:"required"`
	EntityType  string       `json:"entity_type"`
	EntityID    *uuid.UUID   `json:"entity_id"`
	IsPrivate   bool         `json:"is_private"`
}

// ChatRoomResponse represents the response with chat room details
type ChatRoomResponse struct {
	ChatRoom         *ChatRoom         `json:"chat_room"`
	Participants     []*ChatParticipant `json:"participants,omitempty"`
	UnreadCount      int               `json:"unread_count,omitempty"`
	LastMessage      *Message          `json:"last_message,omitempty"`
	ParticipantCount int               `json:"participant_count,omitempty"`
}

// Validate checks if the chat room type is valid
func (r *ChatRoomRequest) Validate() error {
	validTypes := map[ChatRoomType]bool{
		ChatRoomTypeDirect:       true,
		ChatRoomTypeGroup:        true,
		ChatRoomTypeTeam:         true,
		ChatRoomTypeProject:      true,
		ChatRoomTypeTicket:       true,
		ChatRoomTypeAccount:      true,
		ChatRoomTypeOrganization: true,
		ChatRoomTypeAttachment:   true,
		ChatRoomTypeCustom:       true,
	}

	if !validTypes[r.Type] {
		return ErrInvalidChatRoomType
	}

	return nil
}
