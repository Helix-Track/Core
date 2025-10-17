package models

import "github.com/google/uuid"

// UserInfo represents basic user information (from Core service)
type UserInfo struct {
	ID       uuid.UUID `json:"id"`
	Username string    `json:"username"`
	FullName string    `json:"full_name"`
	Email    string    `json:"email,omitempty"`
	Avatar   string    `json:"avatar,omitempty"`
}

// APIResponse represents the standard API response structure
type APIResponse struct {
	ErrorCode              int         `json:"error_code"`
	ErrorMessage           string      `json:"error_message,omitempty"`
	ErrorMessageLocalised  string      `json:"error_message_localised,omitempty"`
	Data                   interface{} `json:"data,omitempty"`
}

// PaginationMeta represents pagination metadata
type PaginationMeta struct {
	Total      int  `json:"total"`
	Limit      int  `json:"limit"`
	Offset     int  `json:"offset"`
	HasMore    bool `json:"has_more"`
	NextOffset int  `json:"next_offset,omitempty"`
}

// ListResponse represents a paginated list response
type ListResponse struct {
	Items      interface{}     `json:"items"`
	Pagination *PaginationMeta `json:"pagination,omitempty"`
}

// WSEvent represents a WebSocket event
type WSEvent struct {
	Type      string      `json:"type"`
	ChatRoomID uuid.UUID   `json:"chat_room_id,omitempty"`
	Data      interface{} `json:"data"`
	Timestamp int64       `json:"timestamp"`
}

// WebSocket event types
const (
	WSEventTypeMessageNew        = "message.new"
	WSEventTypeMessageUpdated    = "message.updated"
	WSEventTypeMessageDeleted    = "message.deleted"
	WSEventTypeTypingStarted     = "typing.started"
	WSEventTypeTypingStopped     = "typing.stopped"
	WSEventTypeReadReceipt       = "read.receipt"
	WSEventTypeReactionAdded     = "reaction.added"
	WSEventTypeReactionRemoved   = "reaction.removed"
	WSEventTypeParticipantJoined = "participant.joined"
	WSEventTypeParticipantLeft   = "participant.left"
	WSEventTypeParticipantUpdated = "participant.updated"
	WSEventTypePresenceChanged   = "presence.changed"
	WSEventTypeChatRoomCreated   = "chatroom.created"
	WSEventTypeChatRoomUpdated   = "chatroom.updated"
	WSEventTypeChatRoomDeleted   = "chatroom.deleted"
	WSEventTypeChatRoomArchived  = "chatroom.archived"
)

// JWTClaims represents the JWT token claims
type JWTClaims struct {
	Sub         string    `json:"sub"`
	Name        string    `json:"name"`
	Username    string    `json:"username"`
	UserID      uuid.UUID `json:"user_id,omitempty"`
	Role        string    `json:"role"`
	Permissions string    `json:"permissions"`
	CoreAddress string    `json:"htCoreAddress"`
}

// Config represents the chat service configuration
type Config struct {
	Server   ServerConfig   `json:"server"`
	Database DatabaseConfig `json:"database"`
	JWT      JWTConfig      `json:"jwt"`
	Logger   LoggerConfig   `json:"logger"`
	Security SecurityConfig `json:"security"`
}

// ServerConfig represents server configuration
type ServerConfig struct {
	Address         string `json:"address"`
	Port            int    `json:"port"`
	HTTPS           bool   `json:"https"`
	CertFile        string `json:"cert_file"`
	KeyFile         string `json:"key_file"`
	EnableHTTP3     bool   `json:"enable_http3"`
	ReadTimeout     int    `json:"read_timeout"`
	WriteTimeout    int    `json:"write_timeout"`
	MaxHeaderBytes  int    `json:"max_header_bytes"`
}

// DatabaseConfig represents database configuration
type DatabaseConfig struct {
	Type              string `json:"type"`
	Host              string `json:"host"`
	Port              int    `json:"port"`
	Database          string `json:"database"`
	User              string `json:"user"`
	Password          string `json:"password"`
	SSLMode           string `json:"ssl_mode"`
	MaxConnections    int    `json:"max_connections"`
	ConnectionTimeout int    `json:"connection_timeout"`
}

// JWTConfig represents JWT configuration
type JWTConfig struct {
	Secret         string `json:"secret"`
	Issuer         string `json:"issuer"`
	Audience       string `json:"audience"`
	ExpiryHours    int    `json:"expiry_hours"`
}

// LoggerConfig represents logger configuration
type LoggerConfig struct {
	LogPath         string `json:"log_path"`
	LogfileBaseName string `json:"logfile_base_name"`
	LogSizeLimit    int    `json:"log_size_limit"`
	Level           string `json:"level"`
}

// SecurityConfig represents security configuration
type SecurityConfig struct {
	EnableDDOSProtection bool  `json:"enable_ddos_protection"`
	RateLimitPerSecond   int   `json:"rate_limit_per_second"`
	RateLimitBurst       int   `json:"rate_limit_burst"`
	MaxMessageSize       int   `json:"max_message_size"`
	MaxAttachmentSize    int64 `json:"max_attachment_size"`
	AllowedOrigins       []string `json:"allowed_origins"`
}
