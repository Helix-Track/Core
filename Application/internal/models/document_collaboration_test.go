package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// ================================================================
// DocumentComment Tests
// ================================================================

func TestDocumentComment_Validate(t *testing.T) {
	tests := []struct {
		name      string
		comment   *DocumentComment
		wantError bool
		errorMsg  string
	}{
		{
			name: "Valid comment",
			comment: &DocumentComment{
				ID:         "comment-123",
				DocumentID: "doc-123",
				UserID:     "user-123",
				Content:    "This is a comment",
				Version:    1,
				IsResolved: false,
				Created:    time.Now().Unix(),
				Modified:   time.Now().Unix(),
			},
			wantError: false,
		},
		{
			name: "Empty ID",
			comment: &DocumentComment{
				ID:         "",
				DocumentID: "doc-123",
				UserID:     "user-123",
				Content:    "This is a comment",
				Version:    1,
				Created:    time.Now().Unix(),
				Modified:   time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "comment ID cannot be empty",
		},
		{
			name: "Empty DocumentID",
			comment: &DocumentComment{
				ID:         "comment-123",
				DocumentID: "",
				UserID:     "user-123",
				Content:    "This is a comment",
				Version:    1,
				Created:    time.Now().Unix(),
				Modified:   time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "comment document ID cannot be empty",
		},
		{
			name: "Empty UserID",
			comment: &DocumentComment{
				ID:         "comment-123",
				DocumentID: "doc-123",
				UserID:     "",
				Content:    "This is a comment",
				Version:    1,
				Created:    time.Now().Unix(),
				Modified:   time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "comment user ID cannot be empty",
		},
		{
			name: "Empty Content",
			comment: &DocumentComment{
				ID:         "comment-123",
				DocumentID: "doc-123",
				UserID:     "user-123",
				Content:    "",
				Version:    1,
				Created:    time.Now().Unix(),
				Modified:   time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "comment content cannot be empty",
		},
		{
			name: "Version less than 1",
			comment: &DocumentComment{
				ID:         "comment-123",
				DocumentID: "doc-123",
				UserID:     "user-123",
				Content:    "This is a comment",
				Version:    0,
				Created:    time.Now().Unix(),
				Modified:   time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "comment version must be at least 1",
		},
		{
			name: "Zero Created timestamp",
			comment: &DocumentComment{
				ID:         "comment-123",
				DocumentID: "doc-123",
				UserID:     "user-123",
				Content:    "This is a comment",
				Version:    1,
				Created:    0,
				Modified:   time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "comment created timestamp cannot be zero",
		},
		{
			name: "Zero Modified timestamp",
			comment: &DocumentComment{
				ID:         "comment-123",
				DocumentID: "doc-123",
				UserID:     "user-123",
				Content:    "This is a comment",
				Version:    1,
				Created:    time.Now().Unix(),
				Modified:   0,
			},
			wantError: true,
			errorMsg:  "comment modified timestamp cannot be zero",
		},
		{
			name: "Threaded comment with parent",
			comment: &DocumentComment{
				ID:         "comment-456",
				DocumentID: "doc-123",
				UserID:     "user-123",
				Content:    "Reply comment",
				ParentID:   stringPtr("comment-123"),
				Version:    1,
				Created:    time.Now().Unix(),
				Modified:   time.Now().Unix(),
			},
			wantError: false,
		},
		{
			name: "Resolved comment",
			comment: &DocumentComment{
				ID:         "comment-789",
				DocumentID: "doc-123",
				UserID:     "user-123",
				Content:    "Fixed issue",
				Version:    1,
				IsResolved: true,
				Created:    time.Now().Unix(),
				Modified:   time.Now().Unix(),
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.comment.Validate()
			if tt.wantError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ================================================================
// DocumentInlineComment Tests
// ================================================================

func TestDocumentInlineComment_Validate(t *testing.T) {
	tests := []struct {
		name      string
		inline    *DocumentInlineComment
		wantError bool
		errorMsg  string
	}{
		{
			name: "Valid inline comment",
			inline: &DocumentInlineComment{
				ID:            "inline-123",
				DocumentID:    "doc-123",
				CommentID:     "comment-123",
				PositionStart: 10,
				PositionEnd:   20,
				SelectedText:  stringPtr("selected text"),
				IsResolved:    false,
				Created:       time.Now().Unix(),
			},
			wantError: false,
		},
		{
			name: "Empty ID",
			inline: &DocumentInlineComment{
				ID:            "",
				DocumentID:    "doc-123",
				CommentID:     "comment-123",
				PositionStart: 10,
				PositionEnd:   20,
				Created:       time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "inline comment ID cannot be empty",
		},
		{
			name: "Empty DocumentID",
			inline: &DocumentInlineComment{
				ID:            "inline-123",
				DocumentID:    "",
				CommentID:     "comment-123",
				PositionStart: 10,
				PositionEnd:   20,
				Created:       time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "inline comment document ID cannot be empty",
		},
		{
			name: "Empty CommentID",
			inline: &DocumentInlineComment{
				ID:            "inline-123",
				DocumentID:    "doc-123",
				CommentID:     "",
				PositionStart: 10,
				PositionEnd:   20,
				Created:       time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "inline comment comment ID cannot be empty",
		},
		{
			name: "Negative PositionStart",
			inline: &DocumentInlineComment{
				ID:            "inline-123",
				DocumentID:    "doc-123",
				CommentID:     "comment-123",
				PositionStart: -1,
				PositionEnd:   20,
				Created:       time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "position start cannot be negative",
		},
		{
			name: "PositionEnd < PositionStart",
			inline: &DocumentInlineComment{
				ID:            "inline-123",
				DocumentID:    "doc-123",
				CommentID:     "comment-123",
				PositionStart: 20,
				PositionEnd:   10,
				Created:       time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "position end must be >= position start",
		},
		{
			name: "Zero Created timestamp",
			inline: &DocumentInlineComment{
				ID:            "inline-123",
				DocumentID:    "doc-123",
				CommentID:     "comment-123",
				PositionStart: 10,
				PositionEnd:   20,
				Created:       0,
			},
			wantError: true,
			errorMsg:  "inline comment created timestamp cannot be zero",
		},
		{
			name: "Same start and end position",
			inline: &DocumentInlineComment{
				ID:            "inline-123",
				DocumentID:    "doc-123",
				CommentID:     "comment-123",
				PositionStart: 10,
				PositionEnd:   10,
				Created:       time.Now().Unix(),
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.inline.Validate()
			if tt.wantError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ================================================================
// DocumentMention Tests
// ================================================================

func TestDocumentMention_Validate(t *testing.T) {
	tests := []struct {
		name      string
		mention   *DocumentMention
		wantError bool
		errorMsg  string
	}{
		{
			name: "Valid mention",
			mention: &DocumentMention{
				ID:               "mention-123",
				DocumentID:       "doc-123",
				MentionedUserID:  "user-456",
				MentioningUserID: "user-123",
				Created:          time.Now().Unix(),
			},
			wantError: false,
		},
		{
			name: "Empty ID",
			mention: &DocumentMention{
				ID:               "",
				DocumentID:       "doc-123",
				MentionedUserID:  "user-456",
				MentioningUserID: "user-123",
				Created:          time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "mention ID cannot be empty",
		},
		{
			name: "Empty DocumentID",
			mention: &DocumentMention{
				ID:               "mention-123",
				DocumentID:       "",
				MentionedUserID:  "user-456",
				MentioningUserID: "user-123",
				Created:          time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "mention document ID cannot be empty",
		},
		{
			name: "Empty MentionedUserID",
			mention: &DocumentMention{
				ID:               "mention-123",
				DocumentID:       "doc-123",
				MentionedUserID:  "",
				MentioningUserID: "user-123",
				Created:          time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "mentioned user ID cannot be empty",
		},
		{
			name: "Empty MentioningUserID",
			mention: &DocumentMention{
				ID:               "mention-123",
				DocumentID:       "doc-123",
				MentionedUserID:  "user-456",
				MentioningUserID: "",
				Created:          time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "mentioning user ID cannot be empty",
		},
		{
			name: "Zero Created timestamp",
			mention: &DocumentMention{
				ID:               "mention-123",
				DocumentID:       "doc-123",
				MentionedUserID:  "user-456",
				MentioningUserID: "user-123",
				Created:          0,
			},
			wantError: true,
			errorMsg:  "mention created timestamp cannot be zero",
		},
		{
			name: "With context and position",
			mention: &DocumentMention{
				ID:               "mention-123",
				DocumentID:       "doc-123",
				MentionedUserID:  "user-456",
				MentioningUserID: "user-123",
				MentionContext:   stringPtr("Check this section"),
				Position:         intPtr(150),
				IsAcknowledged:   false,
				Created:          time.Now().Unix(),
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.mention.Validate()
			if tt.wantError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ================================================================
// DocumentReaction Tests
// ================================================================

func TestDocumentReaction_Validate(t *testing.T) {
	tests := []struct {
		name      string
		reaction  *DocumentReaction
		wantError bool
		errorMsg  string
	}{
		{
			name: "Valid reaction",
			reaction: &DocumentReaction{
				ID:           "reaction-123",
				DocumentID:   "doc-123",
				UserID:       "user-123",
				ReactionType: "like",
				Created:      time.Now().Unix(),
			},
			wantError: false,
		},
		{
			name: "Empty ID",
			reaction: &DocumentReaction{
				ID:           "",
				DocumentID:   "doc-123",
				UserID:       "user-123",
				ReactionType: "like",
				Created:      time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "reaction ID cannot be empty",
		},
		{
			name: "Empty DocumentID",
			reaction: &DocumentReaction{
				ID:           "reaction-123",
				DocumentID:   "",
				UserID:       "user-123",
				ReactionType: "like",
				Created:      time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "reaction document ID cannot be empty",
		},
		{
			name: "Empty UserID",
			reaction: &DocumentReaction{
				ID:           "reaction-123",
				DocumentID:   "doc-123",
				UserID:       "",
				ReactionType: "like",
				Created:      time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "reaction user ID cannot be empty",
		},
		{
			name: "Empty ReactionType",
			reaction: &DocumentReaction{
				ID:           "reaction-123",
				DocumentID:   "doc-123",
				UserID:       "user-123",
				ReactionType: "",
				Created:      time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "reaction type cannot be empty",
		},
		{
			name: "Zero Created timestamp",
			reaction: &DocumentReaction{
				ID:           "reaction-123",
				DocumentID:   "doc-123",
				UserID:       "user-123",
				ReactionType: "like",
				Created:      0,
			},
			wantError: true,
			errorMsg:  "reaction created timestamp cannot be zero",
		},
		{
			name: "With emoji",
			reaction: &DocumentReaction{
				ID:           "reaction-123",
				DocumentID:   "doc-123",
				UserID:       "user-123",
				ReactionType: "emoji",
				Emoji:        stringPtr("👍"),
				Created:      time.Now().Unix(),
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.reaction.Validate()
			if tt.wantError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ================================================================
// DocumentWatcher Tests
// ================================================================

func TestDocumentWatcher_Validate(t *testing.T) {
	tests := []struct {
		name      string
		watcher   *DocumentWatcher
		wantError bool
		errorMsg  string
	}{
		{
			name: "Valid watcher with 'all' notification",
			watcher: &DocumentWatcher{
				ID:                "watcher-123",
				DocumentID:        "doc-123",
				UserID:            "user-123",
				NotificationLevel: "all",
				Created:           time.Now().Unix(),
			},
			wantError: false,
		},
		{
			name: "Empty ID",
			watcher: &DocumentWatcher{
				ID:                "",
				DocumentID:        "doc-123",
				UserID:            "user-123",
				NotificationLevel: "all",
				Created:           time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "watcher ID cannot be empty",
		},
		{
			name: "Empty DocumentID",
			watcher: &DocumentWatcher{
				ID:                "watcher-123",
				DocumentID:        "",
				UserID:            "user-123",
				NotificationLevel: "all",
				Created:           time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "watcher document ID cannot be empty",
		},
		{
			name: "Empty UserID",
			watcher: &DocumentWatcher{
				ID:                "watcher-123",
				DocumentID:        "doc-123",
				UserID:            "",
				NotificationLevel: "all",
				Created:           time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "watcher user ID cannot be empty",
		},
		{
			name: "Empty NotificationLevel",
			watcher: &DocumentWatcher{
				ID:                "watcher-123",
				DocumentID:        "doc-123",
				UserID:            "user-123",
				NotificationLevel: "",
				Created:           time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "notification level cannot be empty",
		},
		{
			name: "Invalid NotificationLevel",
			watcher: &DocumentWatcher{
				ID:                "watcher-123",
				DocumentID:        "doc-123",
				UserID:            "user-123",
				NotificationLevel: "invalid",
				Created:           time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "invalid notification level",
		},
		{
			name: "Zero Created timestamp",
			watcher: &DocumentWatcher{
				ID:                "watcher-123",
				DocumentID:        "doc-123",
				UserID:            "user-123",
				NotificationLevel: "all",
				Created:           0,
			},
			wantError: true,
			errorMsg:  "watcher created timestamp cannot be zero",
		},
		{
			name: "Valid 'mentions' notification",
			watcher: &DocumentWatcher{
				ID:                "watcher-123",
				DocumentID:        "doc-123",
				UserID:            "user-123",
				NotificationLevel: "mentions",
				Created:           time.Now().Unix(),
			},
			wantError: false,
		},
		{
			name: "Valid 'none' notification",
			watcher: &DocumentWatcher{
				ID:                "watcher-123",
				DocumentID:        "doc-123",
				UserID:            "user-123",
				NotificationLevel: "none",
				Created:           time.Now().Unix(),
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.watcher.Validate()
			if tt.wantError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDocumentWatcher_AllNotificationLevels(t *testing.T) {
	levels := []string{"all", "mentions", "none"}

	for _, level := range levels {
		t.Run("Level: "+level, func(t *testing.T) {
			watcher := &DocumentWatcher{
				ID:                "watcher-123",
				DocumentID:        "doc-123",
				UserID:            "user-123",
				NotificationLevel: level,
				Created:           time.Now().Unix(),
			}

			err := watcher.Validate()
			assert.NoError(t, err)
		})
	}
}

// ================================================================
// DocumentLabel Tests
// ================================================================

func TestDocumentLabel_Validate(t *testing.T) {
	tests := []struct {
		name      string
		label     *DocumentLabel
		wantError bool
		errorMsg  string
	}{
		{
			name: "Valid label",
			label: &DocumentLabel{
				ID:      "label-123",
				Name:    "Important",
				Created: time.Now().Unix(),
			},
			wantError: false,
		},
		{
			name: "Empty ID",
			label: &DocumentLabel{
				ID:      "",
				Name:    "Important",
				Created: time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "label ID cannot be empty",
		},
		{
			name: "Empty Name",
			label: &DocumentLabel{
				ID:      "label-123",
				Name:    "",
				Created: time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "label name cannot be empty",
		},
		{
			name: "Zero Created timestamp",
			label: &DocumentLabel{
				ID:      "label-123",
				Name:    "Important",
				Created: 0,
			},
			wantError: true,
			errorMsg:  "label created timestamp cannot be zero",
		},
		{
			name: "With description and color",
			label: &DocumentLabel{
				ID:          "label-123",
				Name:        "Important",
				Description: stringPtr("High priority items"),
				Color:       stringPtr("#FF0000"),
				Created:     time.Now().Unix(),
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.label.Validate()
			if tt.wantError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ================================================================
// DocumentTag Tests
// ================================================================

func TestDocumentTag_Validate(t *testing.T) {
	tests := []struct {
		name      string
		tag       *DocumentTag
		wantError bool
		errorMsg  string
	}{
		{
			name: "Valid tag",
			tag: &DocumentTag{
				ID:      "tag-123",
				Name:    "api",
				Created: time.Now().Unix(),
			},
			wantError: false,
		},
		{
			name: "Empty ID",
			tag: &DocumentTag{
				ID:      "",
				Name:    "api",
				Created: time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "tag ID cannot be empty",
		},
		{
			name: "Empty Name",
			tag: &DocumentTag{
				ID:      "tag-123",
				Name:    "",
				Created: time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "tag name cannot be empty",
		},
		{
			name: "Zero Created timestamp",
			tag: &DocumentTag{
				ID:      "tag-123",
				Name:    "api",
				Created: 0,
			},
			wantError: true,
			errorMsg:  "tag created timestamp cannot be zero",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.tag.Validate()
			if tt.wantError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ================================================================
// Benchmark Tests
// ================================================================

func BenchmarkDocumentComment_Validate(b *testing.B) {
	comment := &DocumentComment{
		ID:         "comment-123",
		DocumentID: "doc-123",
		UserID:     "user-123",
		Content:    "Test comment",
		Version:    1,
		Created:    time.Now().Unix(),
		Modified:   time.Now().Unix(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = comment.Validate()
	}
}

func BenchmarkDocumentWatcher_Validate(b *testing.B) {
	watcher := &DocumentWatcher{
		ID:                "watcher-123",
		DocumentID:        "doc-123",
		UserID:            "user-123",
		NotificationLevel: "all",
		Created:           time.Now().Unix(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = watcher.Validate()
	}
}

// ================================================================
// Helper Functions
// ================================================================

func intPtr(i int) *int {
	return &i
}
