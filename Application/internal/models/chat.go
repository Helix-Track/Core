package models

// UserPresence represents online/offline status for a user
type UserPresence struct {
	ID            string `json:"id" db:"id"`
	UserID        string `json:"user_id" db:"user_id" binding:"required"`
	Status        string `json:"status" db:"status" binding:"required"` // online, offline, away, busy, dnd
	StatusMessage string `json:"status_message,omitempty" db:"status_message"`
	LastSeen      int64  `json:"last_seen" db:"last_seen"`
	CreatedAt     int64  `json:"created_at" db:"created_at"`
	UpdatedAt     int64  `json:"updated_at" db:"updated_at"`
}

// Presence status constants
const (
	PresenceStatusOnline  = "online"
	PresenceStatusOffline = "offline"
	PresenceStatusAway    = "away"
	PresenceStatusBusy    = "busy"
	PresenceStatusDND     = "dnd"
)

// IsValidStatus checks if the presence status is valid
func (up *UserPresence) IsValidStatus() bool {
	return up.Status == PresenceStatusOnline ||
		up.Status == PresenceStatusOffline ||
		up.Status == PresenceStatusAway ||
		up.Status == PresenceStatusBusy ||
		up.Status == PresenceStatusDND
}

// ChatRoom represents a chat room that can be associated with any entity
type ChatRoom struct {
	ID          string `json:"id" db:"id"`
	Name        string `json:"name,omitempty" db:"name"`
	Description string `json:"description,omitempty" db:"description"`
	Type        string `json:"type" db:"type" binding:"required"` // direct, group, team, project, ticket, account, organization, attachment, custom
	EntityType  string `json:"entity_type,omitempty" db:"entity_type"`
	EntityID    string `json:"entity_id,omitempty" db:"entity_id"`
	CreatedBy   string `json:"created_by" db:"created_by" binding:"required"`
	IsPrivate   bool   `json:"is_private" db:"is_private"`
	IsArchived  bool   `json:"is_archived" db:"is_archived"`
	CreatedAt   int64  `json:"created_at" db:"created_at"`
	UpdatedAt   int64  `json:"updated_at" db:"updated_at"`
	Deleted     bool   `json:"deleted" db:"deleted"`
	DeletedAt   *int64 `json:"deleted_at,omitempty" db:"deleted_at"`
}

// Chat room type constants
const (
	ChatRoomTypeDirect       = "direct"
	ChatRoomTypeGroup        = "group"
	ChatRoomTypeTeam         = "team"
	ChatRoomTypeProject      = "project"
	ChatRoomTypeTicket       = "ticket"
	ChatRoomTypeAccount      = "account"
	ChatRoomTypeOrganization = "organization"
	ChatRoomTypeAttachment   = "attachment"
	ChatRoomTypeCustom       = "custom"
)

// IsValidType checks if the chat room type is valid
func (cr *ChatRoom) IsValidType() bool {
	return cr.Type == ChatRoomTypeDirect ||
		cr.Type == ChatRoomTypeGroup ||
		cr.Type == ChatRoomTypeTeam ||
		cr.Type == ChatRoomTypeProject ||
		cr.Type == ChatRoomTypeTicket ||
		cr.Type == ChatRoomTypeAccount ||
		cr.Type == ChatRoomTypeOrganization ||
		cr.Type == ChatRoomTypeAttachment ||
		cr.Type == ChatRoomTypeCustom
}

// ChatParticipant represents a user in a chat room
type ChatParticipant struct {
	ID         string `json:"id" db:"id"`
	ChatRoomID string `json:"chat_room_id" db:"chat_room_id" binding:"required"`
	UserID     string `json:"user_id" db:"user_id" binding:"required"`
	Role       string `json:"role" db:"role" binding:"required"` // owner, admin, moderator, member, guest
	IsMuted    bool   `json:"is_muted" db:"is_muted"`
	JoinedAt   int64  `json:"joined_at" db:"joined_at"`
	LeftAt     *int64 `json:"left_at,omitempty" db:"left_at"`
	CreatedAt  int64  `json:"created_at" db:"created_at"`
	UpdatedAt  int64  `json:"updated_at" db:"updated_at"`
	Deleted    bool   `json:"deleted" db:"deleted"`
	DeletedAt  *int64 `json:"deleted_at,omitempty" db:"deleted_at"`
}

// Participant role constants
const (
	ChatParticipantRoleOwner     = "owner"
	ChatParticipantRoleAdmin     = "admin"
	ChatParticipantRoleModerator = "moderator"
	ChatParticipantRoleMember    = "member"
	ChatParticipantRoleGuest     = "guest"
)

// IsValidRole checks if the participant role is valid
func (cp *ChatParticipant) IsValidRole() bool {
	return cp.Role == ChatParticipantRoleOwner ||
		cp.Role == ChatParticipantRoleAdmin ||
		cp.Role == ChatParticipantRoleModerator ||
		cp.Role == ChatParticipantRoleMember ||
		cp.Role == ChatParticipantRoleGuest
}

// Message represents a chat message
type Message struct {
	ID              string                 `json:"id" db:"id"`
	ChatRoomID      string                 `json:"chat_room_id" db:"chat_room_id" binding:"required"`
	SenderID        string                 `json:"sender_id" db:"sender_id" binding:"required"`
	ParentID        *string                `json:"parent_id,omitempty" db:"parent_id"`         // for threads/replies
	QuotedMessageID *string                `json:"quoted_message_id,omitempty" db:"quoted_message_id"` // for quotes
	Type            string                 `json:"type" db:"type" binding:"required"`          // text, reply, quote, system, file, code, poll
	Content         string                 `json:"content" db:"content" binding:"required"`
	ContentFormat   string                 `json:"content_format" db:"content_format"` // plain, markdown, html
	Metadata        map[string]interface{} `json:"metadata,omitempty" db:"metadata"`   // JSONB in database
	IsEdited        bool                   `json:"is_edited" db:"is_edited"`
	EditedAt        *int64                 `json:"edited_at,omitempty" db:"edited_at"`
	IsPinned        bool                   `json:"is_pinned" db:"is_pinned"`
	PinnedAt        *int64                 `json:"pinned_at,omitempty" db:"pinned_at"`
	PinnedBy        *string                `json:"pinned_by,omitempty" db:"pinned_by"`
	CreatedAt       int64                  `json:"created_at" db:"created_at"`
	UpdatedAt       int64                  `json:"updated_at" db:"updated_at"`
	Deleted         bool                   `json:"deleted" db:"deleted"`
	DeletedAt       *int64                 `json:"deleted_at,omitempty" db:"deleted_at"`
}

// Message type constants
const (
	MessageTypeText   = "text"
	MessageTypeReply  = "reply"
	MessageTypeQuote  = "quote"
	MessageTypeSystem = "system"
	MessageTypeFile   = "file"
	MessageTypeCode   = "code"
	MessageTypePoll   = "poll"
)

// Content format constants
const (
	ContentFormatPlain    = "plain"
	ContentFormatMarkdown = "markdown"
	ContentFormatHTML     = "html"
)

// IsValidType checks if the message type is valid
func (m *Message) IsValidType() bool {
	return m.Type == MessageTypeText ||
		m.Type == MessageTypeReply ||
		m.Type == MessageTypeQuote ||
		m.Type == MessageTypeSystem ||
		m.Type == MessageTypeFile ||
		m.Type == MessageTypeCode ||
		m.Type == MessageTypePoll
}

// IsValidContentFormat checks if the content format is valid
func (m *Message) IsValidContentFormat() bool {
	return m.ContentFormat == ContentFormatPlain ||
		m.ContentFormat == ContentFormatMarkdown ||
		m.ContentFormat == ContentFormatHTML
}

// TypingIndicator represents real-time typing status
type TypingIndicator struct {
	ID         string `json:"id" db:"id"`
	ChatRoomID string `json:"chat_room_id" db:"chat_room_id" binding:"required"`
	UserID     string `json:"user_id" db:"user_id" binding:"required"`
	IsTyping   bool   `json:"is_typing" db:"is_typing"`
	StartedAt  int64  `json:"started_at" db:"started_at"`
	ExpiresAt  int64  `json:"expires_at" db:"expires_at"` // auto-expire after 5 seconds
}

// MessageReadReceipt represents message read status
type MessageReadReceipt struct {
	ID        string `json:"id" db:"id"`
	MessageID string `json:"message_id" db:"message_id" binding:"required"`
	UserID    string `json:"user_id" db:"user_id" binding:"required"`
	ReadAt    int64  `json:"read_at" db:"read_at"`
	CreatedAt int64  `json:"created_at" db:"created_at"`
}

// MessageAttachment represents file attachments for messages
type MessageAttachment struct {
	ID           string                 `json:"id" db:"id"`
	MessageID    string                 `json:"message_id" db:"message_id" binding:"required"`
	FileName     string                 `json:"file_name" db:"file_name" binding:"required"`
	FileType     string                 `json:"file_type,omitempty" db:"file_type"`
	FileSize     int64                  `json:"file_size" db:"file_size" binding:"required"`
	FileURL      string                 `json:"file_url" db:"file_url" binding:"required"`
	ThumbnailURL string                 `json:"thumbnail_url,omitempty" db:"thumbnail_url"`
	Metadata     map[string]interface{} `json:"metadata,omitempty" db:"metadata"` // dimensions, duration, etc.
	UploadedBy   string                 `json:"uploaded_by" db:"uploaded_by" binding:"required"`
	CreatedAt    int64                  `json:"created_at" db:"created_at"`
	Deleted      bool                   `json:"deleted" db:"deleted"`
	DeletedAt    *int64                 `json:"deleted_at,omitempty" db:"deleted_at"`
}

// MessageReaction represents emoji reactions to messages
type MessageReaction struct {
	ID        string `json:"id" db:"id"`
	MessageID string `json:"message_id" db:"message_id" binding:"required"`
	UserID    string `json:"user_id" db:"user_id" binding:"required"`
	Emoji     string `json:"emoji" db:"emoji" binding:"required"` // emoji unicode or :emoji_name:
	CreatedAt int64  `json:"created_at" db:"created_at"`
}

// ChatExternalIntegration represents integration with external chat providers
type ChatExternalIntegration struct {
	ID         string                 `json:"id" db:"id"`
	ChatRoomID string                 `json:"chat_room_id" db:"chat_room_id" binding:"required"`
	Provider   string                 `json:"provider" db:"provider" binding:"required"` // slack, telegram, yandex, google, whatsapp, custom
	ExternalID string                 `json:"external_id" db:"external_id" binding:"required"`
	Config     map[string]interface{} `json:"config,omitempty" db:"config"` // provider-specific configuration
	IsActive   bool                   `json:"is_active" db:"is_active"`
	CreatedAt  int64                  `json:"created_at" db:"created_at"`
	UpdatedAt  int64                  `json:"updated_at" db:"updated_at"`
	Deleted    bool                   `json:"deleted" db:"deleted"`
	DeletedAt  *int64                 `json:"deleted_at,omitempty" db:"deleted_at"`
}

// External integration provider constants
const (
	ChatProviderSlack    = "slack"
	ChatProviderTelegram = "telegram"
	ChatProviderYandex   = "yandex"
	ChatProviderGoogle   = "google"
	ChatProviderWhatsApp = "whatsapp"
	ChatProviderCustom   = "custom"
)

// IsValidProvider checks if the provider is valid
func (cei *ChatExternalIntegration) IsValidProvider() bool {
	return cei.Provider == ChatProviderSlack ||
		cei.Provider == ChatProviderTelegram ||
		cei.Provider == ChatProviderYandex ||
		cei.Provider == ChatProviderGoogle ||
		cei.Provider == ChatProviderWhatsApp ||
		cei.Provider == ChatProviderCustom
}
