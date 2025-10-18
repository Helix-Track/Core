package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// ================================================================
// DocumentTagMapping Tests
// ================================================================

func TestDocumentTagMapping_Validate(t *testing.T) {
	tests := []struct {
		name      string
		mapping   *DocumentTagMapping
		wantError bool
		errorMsg  string
	}{
		{
			name: "Valid mapping",
			mapping: &DocumentTagMapping{
				ID:         "mapping-123",
				DocumentID: "doc-123",
				TagID:      "tag-456",
				UserID:     "user-123",
				Created:    time.Now().Unix(),
			},
			wantError: false,
		},
		{
			name: "Empty ID",
			mapping: &DocumentTagMapping{
				ID:         "",
				DocumentID: "doc-123",
				TagID:      "tag-456",
				UserID:     "user-123",
				Created:    time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "tag mapping ID cannot be empty",
		},
		{
			name: "Empty DocumentID",
			mapping: &DocumentTagMapping{
				ID:         "mapping-123",
				DocumentID: "",
				TagID:      "tag-456",
				UserID:     "user-123",
				Created:    time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "tag mapping document ID cannot be empty",
		},
		{
			name: "Empty TagID",
			mapping: &DocumentTagMapping{
				ID:         "mapping-123",
				DocumentID: "doc-123",
				TagID:      "",
				UserID:     "user-123",
				Created:    time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "tag mapping tag ID cannot be empty",
		},
		{
			name: "Empty UserID",
			mapping: &DocumentTagMapping{
				ID:         "mapping-123",
				DocumentID: "doc-123",
				TagID:      "tag-456",
				UserID:     "",
				Created:    time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "tag mapping user ID cannot be empty",
		},
		{
			name: "Zero Created timestamp",
			mapping: &DocumentTagMapping{
				ID:         "mapping-123",
				DocumentID: "doc-123",
				TagID:      "tag-456",
				UserID:     "user-123",
				Created:    0,
			},
			wantError: true,
			errorMsg:  "tag mapping created timestamp cannot be zero",
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
// DocumentEntityLink Tests
// ================================================================

func TestDocumentEntityLink_Validate(t *testing.T) {
	tests := []struct {
		name      string
		link      *DocumentEntityLink
		wantError bool
		errorMsg  string
	}{
		{
			name: "Valid entity link",
			link: &DocumentEntityLink{
				ID:         "link-123",
				DocumentID: "doc-123",
				EntityType: "ticket",
				EntityID:   "ticket-456",
				LinkType:   "relates-to",
				UserID:     "user-123",
				Created:    time.Now().Unix(),
			},
			wantError: false,
		},
		{
			name: "Empty ID",
			link: &DocumentEntityLink{
				ID:         "",
				DocumentID: "doc-123",
				EntityType: "ticket",
				EntityID:   "ticket-456",
				LinkType:   "relates-to",
				UserID:     "user-123",
				Created:    time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "entity link ID cannot be empty",
		},
		{
			name: "Empty DocumentID",
			link: &DocumentEntityLink{
				ID:         "link-123",
				DocumentID: "",
				EntityType: "ticket",
				EntityID:   "ticket-456",
				LinkType:   "relates-to",
				UserID:     "user-123",
				Created:    time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "entity link document ID cannot be empty",
		},
		{
			name: "Empty EntityType",
			link: &DocumentEntityLink{
				ID:         "link-123",
				DocumentID: "doc-123",
				EntityType: "",
				EntityID:   "ticket-456",
				LinkType:   "relates-to",
				UserID:     "user-123",
				Created:    time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "entity link entity type cannot be empty",
		},
		{
			name: "Empty EntityID",
			link: &DocumentEntityLink{
				ID:         "link-123",
				DocumentID: "doc-123",
				EntityType: "ticket",
				EntityID:   "",
				LinkType:   "relates-to",
				UserID:     "user-123",
				Created:    time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "entity link entity ID cannot be empty",
		},
		{
			name: "Empty LinkType",
			link: &DocumentEntityLink{
				ID:         "link-123",
				DocumentID: "doc-123",
				EntityType: "ticket",
				EntityID:   "ticket-456",
				LinkType:   "",
				UserID:     "user-123",
				Created:    time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "entity link link type cannot be empty",
		},
		{
			name: "Empty UserID",
			link: &DocumentEntityLink{
				ID:         "link-123",
				DocumentID: "doc-123",
				EntityType: "ticket",
				EntityID:   "ticket-456",
				LinkType:   "relates-to",
				UserID:     "",
				Created:    time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "entity link user ID cannot be empty",
		},
		{
			name: "Zero Created timestamp",
			link: &DocumentEntityLink{
				ID:         "link-123",
				DocumentID: "doc-123",
				EntityType: "ticket",
				EntityID:   "ticket-456",
				LinkType:   "relates-to",
				UserID:     "user-123",
				Created:    0,
			},
			wantError: true,
			errorMsg:  "entity link created timestamp cannot be zero",
		},
		{
			name: "Link to project",
			link: &DocumentEntityLink{
				ID:          "link-123",
				DocumentID:  "doc-123",
				EntityType:  "project",
				EntityID:    "project-789",
				LinkType:    "documents",
				Description: stringPtr("Project documentation"),
				UserID:      "user-123",
				Created:     time.Now().Unix(),
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.link.Validate()
			if tt.wantError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDocumentEntityLink_AllEntityTypes(t *testing.T) {
	entityTypes := []struct {
		entityType string
		entityID   string
	}{
		{"ticket", "ticket-123"},
		{"project", "project-456"},
		{"user", "user-789"},
		{"epic", "epic-012"},
		{"sprint", "sprint-345"},
	}

	for _, et := range entityTypes {
		t.Run("EntityType: "+et.entityType, func(t *testing.T) {
			link := &DocumentEntityLink{
				ID:         "link-123",
				DocumentID: "doc-123",
				EntityType: et.entityType,
				EntityID:   et.entityID,
				LinkType:   "relates-to",
				UserID:     "user-123",
				Created:    time.Now().Unix(),
			}

			err := link.Validate()
			assert.NoError(t, err)
		})
	}
}

// ================================================================
// DocumentRelationship Tests
// ================================================================

func TestDocumentRelationship_Validate(t *testing.T) {
	tests := []struct {
		name      string
		rel       *DocumentRelationship
		wantError bool
		errorMsg  string
	}{
		{
			name: "Valid relationship",
			rel: &DocumentRelationship{
				ID:               "rel-123",
				SourceDocumentID: "doc-123",
				TargetDocumentID: "doc-456",
				RelationshipType: "references",
				UserID:           "user-123",
				Created:          time.Now().Unix(),
			},
			wantError: false,
		},
		{
			name: "Empty ID",
			rel: &DocumentRelationship{
				ID:               "",
				SourceDocumentID: "doc-123",
				TargetDocumentID: "doc-456",
				RelationshipType: "references",
				UserID:           "user-123",
				Created:          time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "relationship ID cannot be empty",
		},
		{
			name: "Empty SourceDocumentID",
			rel: &DocumentRelationship{
				ID:               "rel-123",
				SourceDocumentID: "",
				TargetDocumentID: "doc-456",
				RelationshipType: "references",
				UserID:           "user-123",
				Created:          time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "relationship source document ID cannot be empty",
		},
		{
			name: "Empty TargetDocumentID",
			rel: &DocumentRelationship{
				ID:               "rel-123",
				SourceDocumentID: "doc-123",
				TargetDocumentID: "",
				RelationshipType: "references",
				UserID:           "user-123",
				Created:          time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "relationship target document ID cannot be empty",
		},
		{
			name: "Same source and target",
			rel: &DocumentRelationship{
				ID:               "rel-123",
				SourceDocumentID: "doc-123",
				TargetDocumentID: "doc-123",
				RelationshipType: "references",
				UserID:           "user-123",
				Created:          time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "source and target document cannot be the same",
		},
		{
			name: "Empty RelationshipType",
			rel: &DocumentRelationship{
				ID:               "rel-123",
				SourceDocumentID: "doc-123",
				TargetDocumentID: "doc-456",
				RelationshipType: "",
				UserID:           "user-123",
				Created:          time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "relationship type cannot be empty",
		},
		{
			name: "Empty UserID",
			rel: &DocumentRelationship{
				ID:               "rel-123",
				SourceDocumentID: "doc-123",
				TargetDocumentID: "doc-456",
				RelationshipType: "references",
				UserID:           "",
				Created:          time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "relationship user ID cannot be empty",
		},
		{
			name: "Zero Created timestamp",
			rel: &DocumentRelationship{
				ID:               "rel-123",
				SourceDocumentID: "doc-123",
				TargetDocumentID: "doc-456",
				RelationshipType: "references",
				UserID:           "user-123",
				Created:          0,
			},
			wantError: true,
			errorMsg:  "relationship created timestamp cannot be zero",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.rel.Validate()
			if tt.wantError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDocumentRelationship_RelationshipTypes(t *testing.T) {
	relationshipTypes := []string{
		"references",
		"related-to",
		"blocks",
		"blocked-by",
		"duplicates",
		"supersedes",
		"child-of",
		"parent-of",
	}

	for _, relType := range relationshipTypes {
		t.Run("Type: "+relType, func(t *testing.T) {
			rel := &DocumentRelationship{
				ID:               "rel-123",
				SourceDocumentID: "doc-123",
				TargetDocumentID: "doc-456",
				RelationshipType: relType,
				UserID:           "user-123",
				Created:          time.Now().Unix(),
			}

			err := rel.Validate()
			assert.NoError(t, err)
		})
	}
}

// ================================================================
// Benchmark Tests
// ================================================================

func BenchmarkDocumentTagMapping_Validate(b *testing.B) {
	mapping := &DocumentTagMapping{
		ID:         "mapping-123",
		DocumentID: "doc-123",
		TagID:      "tag-456",
		UserID:     "user-123",
		Created:    time.Now().Unix(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = mapping.Validate()
	}
}

func BenchmarkDocumentEntityLink_Validate(b *testing.B) {
	link := &DocumentEntityLink{
		ID:         "link-123",
		DocumentID: "doc-123",
		EntityType: "ticket",
		EntityID:   "ticket-456",
		LinkType:   "relates-to",
		UserID:     "user-123",
		Created:    time.Now().Unix(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = link.Validate()
	}
}

func BenchmarkDocumentRelationship_Validate(b *testing.B) {
	rel := &DocumentRelationship{
		ID:               "rel-123",
		SourceDocumentID: "doc-123",
		TargetDocumentID: "doc-456",
		RelationshipType: "references",
		UserID:           "user-123",
		Created:          time.Now().Unix(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = rel.Validate()
	}
}
