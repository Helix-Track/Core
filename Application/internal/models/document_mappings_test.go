package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// ================================================================
// CommentDocumentMapping Tests
// ================================================================

func TestCommentDocumentMapping_Validate(t *testing.T) {
	tests := []struct {
		name      string
		mapping   *CommentDocumentMapping
		wantError bool
		errorMsg  string
	}{
		{
			name: "Valid mapping",
			mapping: &CommentDocumentMapping{
				ID:         "mapping-123",
				CommentID:  "comment-456",
				DocumentID: "doc-789",
				UserID:     "user-123",
				IsResolved: false,
				Created:    time.Now().Unix(),
			},
			wantError: false,
		},
		{
			name: "Empty ID",
			mapping: &CommentDocumentMapping{
				ID:         "",
				CommentID:  "comment-456",
				DocumentID: "doc-789",
				UserID:     "user-123",
				Created:    time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "comment-document mapping ID cannot be empty",
		},
		{
			name: "Empty CommentID",
			mapping: &CommentDocumentMapping{
				ID:         "mapping-123",
				CommentID:  "",
				DocumentID: "doc-789",
				UserID:     "user-123",
				Created:    time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "comment-document mapping comment ID cannot be empty",
		},
		{
			name: "Empty DocumentID",
			mapping: &CommentDocumentMapping{
				ID:         "mapping-123",
				CommentID:  "comment-456",
				DocumentID: "",
				UserID:     "user-123",
				Created:    time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "comment-document mapping document ID cannot be empty",
		},
		{
			name: "Empty UserID",
			mapping: &CommentDocumentMapping{
				ID:         "mapping-123",
				CommentID:  "comment-456",
				DocumentID: "doc-789",
				UserID:     "",
				Created:    time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "comment-document mapping user ID cannot be empty",
		},
		{
			name: "Zero Created timestamp",
			mapping: &CommentDocumentMapping{
				ID:         "mapping-123",
				CommentID:  "comment-456",
				DocumentID: "doc-789",
				UserID:     "user-123",
				Created:    0,
			},
			wantError: true,
			errorMsg:  "comment-document mapping created timestamp cannot be zero",
		},
		{
			name: "Resolved comment mapping",
			mapping: &CommentDocumentMapping{
				ID:         "mapping-123",
				CommentID:  "comment-456",
				DocumentID: "doc-789",
				UserID:     "user-123",
				IsResolved: true,
				Created:    time.Now().Unix(),
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.mapping.Validate()
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
// LabelDocumentMapping Tests
// ================================================================

func TestLabelDocumentMapping_Validate(t *testing.T) {
	tests := []struct {
		name      string
		mapping   *LabelDocumentMapping
		wantError bool
		errorMsg  string
	}{
		{
			name: "Valid mapping",
			mapping: &LabelDocumentMapping{
				ID:         "mapping-123",
				LabelID:    "label-456",
				DocumentID: "doc-789",
				UserID:     "user-123",
				Created:    time.Now().Unix(),
			},
			wantError: false,
		},
		{
			name: "Empty ID",
			mapping: &LabelDocumentMapping{
				ID:         "",
				LabelID:    "label-456",
				DocumentID: "doc-789",
				UserID:     "user-123",
				Created:    time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "label-document mapping ID cannot be empty",
		},
		{
			name: "Empty LabelID",
			mapping: &LabelDocumentMapping{
				ID:         "mapping-123",
				LabelID:    "",
				DocumentID: "doc-789",
				UserID:     "user-123",
				Created:    time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "label-document mapping label ID cannot be empty",
		},
		{
			name: "Empty DocumentID",
			mapping: &LabelDocumentMapping{
				ID:         "mapping-123",
				LabelID:    "label-456",
				DocumentID: "",
				UserID:     "user-123",
				Created:    time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "label-document mapping document ID cannot be empty",
		},
		{
			name: "Empty UserID",
			mapping: &LabelDocumentMapping{
				ID:         "mapping-123",
				LabelID:    "label-456",
				DocumentID: "doc-789",
				UserID:     "",
				Created:    time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "label-document mapping user ID cannot be empty",
		},
		{
			name: "Zero Created timestamp",
			mapping: &LabelDocumentMapping{
				ID:         "mapping-123",
				LabelID:    "label-456",
				DocumentID: "doc-789",
				UserID:     "user-123",
				Created:    0,
			},
			wantError: true,
			errorMsg:  "label-document mapping created timestamp cannot be zero",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.mapping.Validate()
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
// VoteMapping Tests
// ================================================================

func TestVoteMapping_Validate(t *testing.T) {
	tests := []struct {
		name      string
		vote      *VoteMapping
		wantError bool
		errorMsg  string
	}{
		{
			name: "Valid upvote",
			vote: &VoteMapping{
				ID:         "vote-123",
				EntityType: "document",
				EntityID:   "doc-456",
				UserID:     "user-123",
				VoteType:   "upvote",
				Created:    time.Now().Unix(),
			},
			wantError: false,
		},
		{
			name: "Empty ID",
			vote: &VoteMapping{
				ID:         "",
				EntityType: "document",
				EntityID:   "doc-456",
				UserID:     "user-123",
				VoteType:   "upvote",
				Created:    time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "vote mapping ID cannot be empty",
		},
		{
			name: "Empty EntityType",
			vote: &VoteMapping{
				ID:         "vote-123",
				EntityType: "",
				EntityID:   "doc-456",
				UserID:     "user-123",
				VoteType:   "upvote",
				Created:    time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "vote mapping entity type cannot be empty",
		},
		{
			name: "Empty EntityID",
			vote: &VoteMapping{
				ID:         "vote-123",
				EntityType: "document",
				EntityID:   "",
				UserID:     "user-123",
				VoteType:   "upvote",
				Created:    time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "vote mapping entity ID cannot be empty",
		},
		{
			name: "Empty UserID",
			vote: &VoteMapping{
				ID:         "vote-123",
				EntityType: "document",
				EntityID:   "doc-456",
				UserID:     "",
				VoteType:   "upvote",
				Created:    time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "vote mapping user ID cannot be empty",
		},
		{
			name: "Empty VoteType",
			vote: &VoteMapping{
				ID:         "vote-123",
				EntityType: "document",
				EntityID:   "doc-456",
				UserID:     "user-123",
				VoteType:   "",
				Created:    time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "vote mapping vote type cannot be empty",
		},
		{
			name: "Invalid VoteType",
			vote: &VoteMapping{
				ID:         "vote-123",
				EntityType: "document",
				EntityID:   "doc-456",
				UserID:     "user-123",
				VoteType:   "invalid",
				Created:    time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "invalid vote type",
		},
		{
			name: "Zero Created timestamp",
			vote: &VoteMapping{
				ID:         "vote-123",
				EntityType: "document",
				EntityID:   "doc-456",
				UserID:     "user-123",
				VoteType:   "upvote",
				Created:    0,
			},
			wantError: true,
			errorMsg:  "vote mapping created timestamp cannot be zero",
		},
		{
			name: "Vote with emoji",
			vote: &VoteMapping{
				ID:         "vote-123",
				EntityType: "document",
				EntityID:   "doc-456",
				UserID:     "user-123",
				VoteType:   "like",
				Emoji:      stringPtr("👍"),
				Created:    time.Now().Unix(),
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.vote.Validate()
			if tt.wantError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestVoteMapping_AllVoteTypes(t *testing.T) {
	voteTypes := []string{
		"upvote", "downvote", "like", "love",
		"celebrate", "support", "insightful",
	}

	for _, voteType := range voteTypes {
		t.Run("VoteType: "+voteType, func(t *testing.T) {
			vote := &VoteMapping{
				ID:         "vote-123",
				EntityType: "document",
				EntityID:   "doc-456",
				UserID:     "user-123",
				VoteType:   voteType,
				Created:    time.Now().Unix(),
			}

			err := vote.Validate()
			assert.NoError(t, err)
		})
	}
}

func TestVoteMapping_IsPositive(t *testing.T) {
	tests := []struct {
		name     string
		voteType string
		expected bool
	}{
		{"Upvote", "upvote", true},
		{"Like", "like", true},
		{"Love", "love", true},
		{"Celebrate", "celebrate", true},
		{"Support", "support", true},
		{"Insightful", "insightful", true},
		{"Downvote", "downvote", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vote := &VoteMapping{VoteType: tt.voteType}
			assert.Equal(t, tt.expected, vote.IsPositive())
		})
	}
}

func TestVoteMapping_IsNegative(t *testing.T) {
	tests := []struct {
		name     string
		voteType string
		expected bool
	}{
		{"Downvote", "downvote", true},
		{"Upvote", "upvote", false},
		{"Like", "like", false},
		{"Love", "love", false},
		{"Celebrate", "celebrate", false},
		{"Support", "support", false},
		{"Insightful", "insightful", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vote := &VoteMapping{VoteType: tt.voteType}
			assert.Equal(t, tt.expected, vote.IsNegative())
		})
	}
}

func TestVoteMapping_AllEntityTypes(t *testing.T) {
	entityTypes := []string{"document", "ticket", "comment", "epic", "sprint"}

	for _, entityType := range entityTypes {
		t.Run("EntityType: "+entityType, func(t *testing.T) {
			vote := &VoteMapping{
				ID:         "vote-123",
				EntityType: entityType,
				EntityID:   entityType + "-456",
				UserID:     "user-123",
				VoteType:   "upvote",
				Created:    time.Now().Unix(),
			}

			err := vote.Validate()
			assert.NoError(t, err)
		})
	}
}

func TestVoteMapping_Structure(t *testing.T) {
	emoji := "❤️"

	vote := VoteMapping{
		ID:         "vote-123",
		EntityType: "document",
		EntityID:   "doc-456",
		UserID:     "user-123",
		VoteType:   "love",
		Emoji:      &emoji,
		Created:    time.Now().Unix(),
		Deleted:    false,
	}

	assert.Equal(t, "vote-123", vote.ID)
	assert.Equal(t, "document", vote.EntityType)
	assert.Equal(t, "doc-456", vote.EntityID)
	assert.Equal(t, "user-123", vote.UserID)
	assert.Equal(t, "love", vote.VoteType)
	assert.NotNil(t, vote.Emoji)
	assert.Equal(t, "❤️", *vote.Emoji)
	assert.Greater(t, vote.Created, int64(0))
	assert.False(t, vote.Deleted)

	assert.True(t, vote.IsPositive())
	assert.False(t, vote.IsNegative())
}

// ================================================================
// Benchmark Tests
// ================================================================

func BenchmarkCommentDocumentMapping_Validate(b *testing.B) {
	mapping := &CommentDocumentMapping{
		ID:         "mapping-123",
		CommentID:  "comment-456",
		DocumentID: "doc-789",
		UserID:     "user-123",
		Created:    time.Now().Unix(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = mapping.Validate()
	}
}

func BenchmarkLabelDocumentMapping_Validate(b *testing.B) {
	mapping := &LabelDocumentMapping{
		ID:         "mapping-123",
		LabelID:    "label-456",
		DocumentID: "doc-789",
		UserID:     "user-123",
		Created:    time.Now().Unix(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = mapping.Validate()
	}
}

func BenchmarkVoteMapping_Validate(b *testing.B) {
	vote := &VoteMapping{
		ID:         "vote-123",
		EntityType: "document",
		EntityID:   "doc-456",
		UserID:     "user-123",
		VoteType:   "upvote",
		Created:    time.Now().Unix(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = vote.Validate()
	}
}

func BenchmarkVoteMapping_IsPositive(b *testing.B) {
	vote := &VoteMapping{VoteType: "upvote"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = vote.IsPositive()
	}
}

func BenchmarkVoteMapping_IsNegative(b *testing.B) {
	vote := &VoteMapping{VoteType: "downvote"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = vote.IsNegative()
	}
}
