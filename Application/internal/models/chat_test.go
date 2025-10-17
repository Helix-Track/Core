package models

import (
	"testing"
)

// TestUserPresence tests
func TestUserPresence_IsValidStatus(t *testing.T) {
	tests := []struct {
		name     string
		status   string
		expected bool
	}{
		{"Online status", PresenceStatusOnline, true},
		{"Offline status", PresenceStatusOffline, true},
		{"Away status", PresenceStatusAway, true},
		{"Busy status", PresenceStatusBusy, true},
		{"DND status", PresenceStatusDND, true},
		{"Invalid status", "invisible", false},
		{"Empty status", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			presence := &UserPresence{Status: tt.status}
			result := presence.IsValidStatus()
			if result != tt.expected {
				t.Errorf("IsValidStatus() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestChatRoom tests
func TestChatRoom_IsValidType(t *testing.T) {
	tests := []struct {
		name     string
		roomType string
		expected bool
	}{
		{"Direct room", ChatRoomTypeDirect, true},
		{"Group room", ChatRoomTypeGroup, true},
		{"Team room", ChatRoomTypeTeam, true},
		{"Project room", ChatRoomTypeProject, true},
		{"Ticket room", ChatRoomTypeTicket, true},
		{"Account room", ChatRoomTypeAccount, true},
		{"Organization room", ChatRoomTypeOrganization, true},
		{"Attachment room", ChatRoomTypeAttachment, true},
		{"Custom room", ChatRoomTypeCustom, true},
		{"Invalid type", "invalid", false},
		{"Empty type", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			room := &ChatRoom{Type: tt.roomType}
			result := room.IsValidType()
			if result != tt.expected {
				t.Errorf("IsValidType() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestChatRoom_PrivacyAndArchive(t *testing.T) {
	room := &ChatRoom{
		ID:         "room-1",
		Name:       "Private Room",
		Type:       ChatRoomTypeDirect,
		CreatedBy:  "user-1",
		IsPrivate:  true,
		IsArchived: false,
	}

	if !room.IsPrivate {
		t.Error("Expected room to be private")
	}

	if room.IsArchived {
		t.Error("Expected room not to be archived")
	}

	room.IsArchived = true
	if !room.IsArchived {
		t.Error("Expected room to be archived after update")
	}
}

// TestChatParticipant tests
func TestChatParticipant_IsValidRole(t *testing.T) {
	tests := []struct {
		name     string
		role     string
		expected bool
	}{
		{"Owner role", ChatParticipantRoleOwner, true},
		{"Admin role", ChatParticipantRoleAdmin, true},
		{"Moderator role", ChatParticipantRoleModerator, true},
		{"Member role", ChatParticipantRoleMember, true},
		{"Guest role", ChatParticipantRoleGuest, true},
		{"Invalid role", "superuser", false},
		{"Empty role", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			participant := &ChatParticipant{Role: tt.role}
			result := participant.IsValidRole()
			if result != tt.expected {
				t.Errorf("IsValidRole() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestChatParticipant_MuteStatus(t *testing.T) {
	participant := &ChatParticipant{
		ID:         "participant-1",
		ChatRoomID: "room-1",
		UserID:     "user-1",
		Role:       ChatParticipantRoleMember,
		IsMuted:    false,
	}

	if participant.IsMuted {
		t.Error("Expected participant not to be muted")
	}

	participant.IsMuted = true
	if !participant.IsMuted {
		t.Error("Expected participant to be muted after update")
	}
}

// TestMessage tests
func TestMessage_IsValidType(t *testing.T) {
	tests := []struct {
		name        string
		messageType string
		expected    bool
	}{
		{"Text message", MessageTypeText, true},
		{"Reply message", MessageTypeReply, true},
		{"Quote message", MessageTypeQuote, true},
		{"System message", MessageTypeSystem, true},
		{"File message", MessageTypeFile, true},
		{"Code message", MessageTypeCode, true},
		{"Poll message", MessageTypePoll, true},
		{"Invalid type", "audio", false},
		{"Empty type", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			message := &Message{Type: tt.messageType}
			result := message.IsValidType()
			if result != tt.expected {
				t.Errorf("IsValidType() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestMessage_IsValidContentFormat(t *testing.T) {
	tests := []struct {
		name     string
		format   string
		expected bool
	}{
		{"Plain format", ContentFormatPlain, true},
		{"Markdown format", ContentFormatMarkdown, true},
		{"HTML format", ContentFormatHTML, true},
		{"Invalid format", "rtf", false},
		{"Empty format", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			message := &Message{ContentFormat: tt.format}
			result := message.IsValidContentFormat()
			if result != tt.expected {
				t.Errorf("IsValidContentFormat() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestMessage_EditAndPinStatus(t *testing.T) {
	message := &Message{
		ID:         "msg-1",
		ChatRoomID: "room-1",
		SenderID:   "user-1",
		Type:       MessageTypeText,
		Content:    "Hello",
		IsEdited:   false,
		IsPinned:   false,
	}

	if message.IsEdited {
		t.Error("Expected message not to be edited")
	}

	if message.IsPinned {
		t.Error("Expected message not to be pinned")
	}

	message.IsEdited = true
	message.IsPinned = true

	if !message.IsEdited {
		t.Error("Expected message to be edited after update")
	}

	if !message.IsPinned {
		t.Error("Expected message to be pinned after update")
	}
}

func TestMessage_Threading(t *testing.T) {
	parentID := "parent-msg-1"
	quotedID := "quoted-msg-1"

	message := &Message{
		ID:              "msg-1",
		ChatRoomID:      "room-1",
		SenderID:        "user-1",
		ParentID:        &parentID,
		QuotedMessageID: &quotedID,
		Type:            MessageTypeReply,
		Content:         "This is a reply",
	}

	if message.ParentID == nil {
		t.Error("Expected message to have parent ID")
	}

	if *message.ParentID != parentID {
		t.Errorf("Expected parent ID %s, got %s", parentID, *message.ParentID)
	}

	if message.QuotedMessageID == nil {
		t.Error("Expected message to have quoted message ID")
	}

	if *message.QuotedMessageID != quotedID {
		t.Errorf("Expected quoted message ID %s, got %s", quotedID, *message.QuotedMessageID)
	}
}

func TestMessage_Metadata(t *testing.T) {
	message := &Message{
		ID:         "msg-1",
		ChatRoomID: "room-1",
		SenderID:   "user-1",
		Type:       MessageTypeText,
		Content:    "Test message with metadata",
		Metadata: map[string]interface{}{
			"mentions": []string{"user-2", "user-3"},
			"links":    []string{"https://example.com"},
		},
	}

	if message.Metadata == nil {
		t.Fatal("Expected message to have metadata")
	}

	mentions, ok := message.Metadata["mentions"]
	if !ok {
		t.Error("Expected metadata to have mentions field")
	}

	mentionsList, ok := mentions.([]string)
	if !ok {
		t.Error("Expected mentions to be a string slice")
	}

	if len(mentionsList) != 2 {
		t.Errorf("Expected 2 mentions, got %d", len(mentionsList))
	}
}

// TestTypingIndicator tests
func TestTypingIndicator_ExpirationLogic(t *testing.T) {
	now := int64(1000)
	expiresAt := now + 5 // 5 seconds later

	indicator := &TypingIndicator{
		ID:         "typing-1",
		ChatRoomID: "room-1",
		UserID:     "user-1",
		IsTyping:   true,
		StartedAt:  now,
		ExpiresAt:  expiresAt,
	}

	if !indicator.IsTyping {
		t.Error("Expected user to be typing")
	}

	if indicator.ExpiresAt != expiresAt {
		t.Errorf("Expected expiration at %d, got %d", expiresAt, indicator.ExpiresAt)
	}

	// Simulate expiration
	currentTime := int64(1006) // After expiration
	if currentTime > indicator.ExpiresAt {
		indicator.IsTyping = false
	}

	if indicator.IsTyping {
		t.Error("Expected typing indicator to be expired")
	}
}

// TestMessageReadReceipt tests
func TestMessageReadReceipt_CreationAndRead(t *testing.T) {
	receipt := &MessageReadReceipt{
		ID:        "receipt-1",
		MessageID: "msg-1",
		UserID:    "user-1",
		ReadAt:    1234567890,
		CreatedAt: 1234567890,
	}

	if receipt.MessageID == "" {
		t.Error("Expected receipt to have message ID")
	}

	if receipt.UserID == "" {
		t.Error("Expected receipt to have user ID")
	}

	if receipt.ReadAt == 0 {
		t.Error("Expected receipt to have read timestamp")
	}
}

// TestMessageAttachment tests
func TestMessageAttachment_FileInfo(t *testing.T) {
	attachment := &MessageAttachment{
		ID:           "attachment-1",
		MessageID:    "msg-1",
		FileName:     "document.pdf",
		FileType:     "application/pdf",
		FileSize:     1024000, // 1 MB
		FileURL:      "https://cdn.example.com/files/document.pdf",
		ThumbnailURL: "https://cdn.example.com/thumbnails/document.png",
		UploadedBy:   "user-1",
		Metadata: map[string]interface{}{
			"pages": 10,
		},
	}

	if attachment.FileName != "document.pdf" {
		t.Errorf("Expected filename document.pdf, got %s", attachment.FileName)
	}

	if attachment.FileSize != 1024000 {
		t.Errorf("Expected file size 1024000, got %d", attachment.FileSize)
	}

	if attachment.FileType != "application/pdf" {
		t.Errorf("Expected file type application/pdf, got %s", attachment.FileType)
	}

	pages, ok := attachment.Metadata["pages"]
	if !ok {
		t.Error("Expected metadata to have pages field")
	}

	if pages != 10 {
		t.Errorf("Expected 10 pages, got %v", pages)
	}
}

// TestMessageReaction tests
func TestMessageReaction_EmojiHandling(t *testing.T) {
	tests := []struct {
		name  string
		emoji string
	}{
		{"Thumbs up unicode", "\U0001F44D"},
		{"Heart emoji", "❤️"},
		{"Named emoji", ":heart:"},
		{"Multiple emojis", "\U0001F44D\U0001F389"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reaction := &MessageReaction{
				ID:        "reaction-1",
				MessageID: "msg-1",
				UserID:    "user-1",
				Emoji:     tt.emoji,
				CreatedAt: 1234567890,
			}

			if reaction.Emoji != tt.emoji {
				t.Errorf("Expected emoji %s, got %s", tt.emoji, reaction.Emoji)
			}
		})
	}
}

// TestChatExternalIntegration tests
func TestChatExternalIntegration_IsValidProvider(t *testing.T) {
	tests := []struct {
		name     string
		provider string
		expected bool
	}{
		{"Slack provider", ChatProviderSlack, true},
		{"Telegram provider", ChatProviderTelegram, true},
		{"Yandex provider", ChatProviderYandex, true},
		{"Google provider", ChatProviderGoogle, true},
		{"WhatsApp provider", ChatProviderWhatsApp, true},
		{"Custom provider", ChatProviderCustom, true},
		{"Invalid provider", "discord", false},
		{"Empty provider", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			integration := &ChatExternalIntegration{Provider: tt.provider}
			result := integration.IsValidProvider()
			if result != tt.expected {
				t.Errorf("IsValidProvider() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestChatExternalIntegration_Config(t *testing.T) {
	integration := &ChatExternalIntegration{
		ID:         "integration-1",
		ChatRoomID: "room-1",
		Provider:   ChatProviderSlack,
		ExternalID: "slack-channel-123",
		Config: map[string]interface{}{
			"webhook_url": "https://hooks.slack.com/services/XXX/YYY/ZZZ",
			"channel":     "#general",
			"bot_token":   "xoxb-token",
		},
		IsActive: true,
	}

	if !integration.IsActive {
		t.Error("Expected integration to be active")
	}

	webhookURL, ok := integration.Config["webhook_url"]
	if !ok {
		t.Error("Expected config to have webhook_url")
	}

	if webhookURL != "https://hooks.slack.com/services/XXX/YYY/ZZZ" {
		t.Errorf("Unexpected webhook URL: %v", webhookURL)
	}

	channel, ok := integration.Config["channel"]
	if !ok {
		t.Error("Expected config to have channel")
	}

	if channel != "#general" {
		t.Errorf("Expected channel #general, got %v", channel)
	}
}

// TestChatConstants tests
func TestPresenceStatusConstants(t *testing.T) {
	if PresenceStatusOnline != "online" {
		t.Errorf("Expected PresenceStatusOnline to be 'online', got %s", PresenceStatusOnline)
	}
	if PresenceStatusOffline != "offline" {
		t.Errorf("Expected PresenceStatusOffline to be 'offline', got %s", PresenceStatusOffline)
	}
	if PresenceStatusAway != "away" {
		t.Errorf("Expected PresenceStatusAway to be 'away', got %s", PresenceStatusAway)
	}
	if PresenceStatusBusy != "busy" {
		t.Errorf("Expected PresenceStatusBusy to be 'busy', got %s", PresenceStatusBusy)
	}
	if PresenceStatusDND != "dnd" {
		t.Errorf("Expected PresenceStatusDND to be 'dnd', got %s", PresenceStatusDND)
	}
}

func TestChatRoomTypeConstants(t *testing.T) {
	types := []string{
		ChatRoomTypeDirect,
		ChatRoomTypeGroup,
		ChatRoomTypeTeam,
		ChatRoomTypeProject,
		ChatRoomTypeTicket,
		ChatRoomTypeAccount,
		ChatRoomTypeOrganization,
		ChatRoomTypeAttachment,
		ChatRoomTypeCustom,
	}

	if len(types) != 9 {
		t.Errorf("Expected 9 chat room types, got %d", len(types))
	}
}

func TestMessageTypeConstants(t *testing.T) {
	types := []string{
		MessageTypeText,
		MessageTypeReply,
		MessageTypeQuote,
		MessageTypeSystem,
		MessageTypeFile,
		MessageTypeCode,
		MessageTypePoll,
	}

	if len(types) != 7 {
		t.Errorf("Expected 7 message types, got %d", len(types))
	}
}

func TestChatProviderConstants(t *testing.T) {
	providers := []string{
		ChatProviderSlack,
		ChatProviderTelegram,
		ChatProviderYandex,
		ChatProviderGoogle,
		ChatProviderWhatsApp,
		ChatProviderCustom,
	}

	if len(providers) != 6 {
		t.Errorf("Expected 6 chat providers, got %d", len(providers))
	}
}
