package models

import "github.com/google/uuid"

// ParticipantRole represents the role of a participant in a chat room
type ParticipantRole string

const (
	ParticipantRoleOwner     ParticipantRole = "owner"
	ParticipantRoleAdmin     ParticipantRole = "admin"
	ParticipantRoleModerator ParticipantRole = "moderator"
	ParticipantRoleMember    ParticipantRole = "member"
	ParticipantRoleGuest     ParticipantRole = "guest"
)

// ChatParticipant represents a user participating in a chat room
type ChatParticipant struct {
	ID         uuid.UUID       `json:"id" db:"id"`
	ChatRoomID uuid.UUID       `json:"chat_room_id" db:"chat_room_id"`
	UserID     uuid.UUID       `json:"user_id" db:"user_id"`
	Role       ParticipantRole `json:"role" db:"role"`
	IsMuted    bool            `json:"is_muted" db:"is_muted"`
	JoinedAt   int64           `json:"joined_at" db:"joined_at"`
	LeftAt     *int64          `json:"left_at,omitempty" db:"left_at"`
	CreatedAt  int64           `json:"created_at" db:"created_at"`
	UpdatedAt  int64           `json:"updated_at" db:"updated_at"`
	Deleted    bool            `json:"deleted" db:"deleted"`
	DeletedAt  *int64          `json:"deleted_at,omitempty" db:"deleted_at"`
}

// ParticipantRequest represents the request to add/update a participant
type ParticipantRequest struct {
	ChatRoomID uuid.UUID       `json:"chat_room_id" binding:"required"`
	UserID     uuid.UUID       `json:"user_id" binding:"required"`
	Role       ParticipantRole `json:"role"`
}

// ParticipantResponse represents the response with participant details
type ParticipantResponse struct {
	Participant *ChatParticipant `json:"participant"`
	UserInfo    *UserInfo        `json:"user_info,omitempty"`
	Presence    *UserPresence    `json:"presence,omitempty"`
}

// Validate checks if the participant request is valid
func (r *ParticipantRequest) Validate() error {
	validRoles := map[ParticipantRole]bool{
		ParticipantRoleOwner:     true,
		ParticipantRoleAdmin:     true,
		ParticipantRoleModerator: true,
		ParticipantRoleMember:    true,
		ParticipantRoleGuest:     true,
	}

	if r.Role != "" && !validRoles[r.Role] {
		return ErrInvalidParticipantRole
	}

	// Set default role
	if r.Role == "" {
		r.Role = ParticipantRoleMember
	}

	return nil
}
