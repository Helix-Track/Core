package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// ================================================================
// DocumentSpace Tests
// ================================================================

func TestDocumentSpace_Validate(t *testing.T) {
	validSpace := &DocumentSpace{
		ID:          "space-123",
		Key:         "DOCS",
		Name:        "Documentation",
		Description: "Technical documentation space",
		OwnerID:     "user-123",
		IsPublic:    true,
		Created:     time.Now().Unix(),
		Modified:    time.Now().Unix(),
		Deleted:     false,
	}

	tests := []struct {
		name      string
		space     *DocumentSpace
		wantError bool
		errorMsg  string
	}{
		{
			name:      "Valid space",
			space:     validSpace,
			wantError: false,
		},
		{
			name: "Empty ID",
			space: &DocumentSpace{
				ID:       "",
				Key:      "DOCS",
				Name:     "Documentation",
				OwnerID:  "user-123",
				Created:  time.Now().Unix(),
				Modified: time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "document space ID cannot be empty",
		},
		{
			name: "Empty Key",
			space: &DocumentSpace{
				ID:       "space-123",
				Key:      "",
				Name:     "Documentation",
				OwnerID:  "user-123",
				Created:  time.Now().Unix(),
				Modified: time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "document space key cannot be empty",
		},
		{
			name: "Empty Name",
			space: &DocumentSpace{
				ID:       "space-123",
				Key:      "DOCS",
				Name:     "",
				OwnerID:  "user-123",
				Created:  time.Now().Unix(),
				Modified: time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "document space name cannot be empty",
		},
		{
			name: "Empty OwnerID",
			space: &DocumentSpace{
				ID:       "space-123",
				Key:      "DOCS",
				Name:     "Documentation",
				OwnerID:  "",
				Created:  time.Now().Unix(),
				Modified: time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "document space owner ID cannot be empty",
		},
		{
			name: "Zero Created timestamp",
			space: &DocumentSpace{
				ID:       "space-123",
				Key:      "DOCS",
				Name:     "Documentation",
				OwnerID:  "user-123",
				Created:  0,
				Modified: time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "document space created timestamp cannot be zero",
		},
		{
			name: "Zero Modified timestamp",
			space: &DocumentSpace{
				ID:       "space-123",
				Key:      "DOCS",
				Name:     "Documentation",
				OwnerID:  "user-123",
				Created:  time.Now().Unix(),
				Modified: 0,
			},
			wantError: true,
			errorMsg:  "document space modified timestamp cannot be zero",
		},
		{
			name: "Public space",
			space: &DocumentSpace{
				ID:       "space-123",
				Key:      "PUBLIC",
				Name:     "Public Space",
				OwnerID:  "user-123",
				IsPublic: true,
				Created:  time.Now().Unix(),
				Modified: time.Now().Unix(),
			},
			wantError: false,
		},
		{
			name: "Private space",
			space: &DocumentSpace{
				ID:       "space-123",
				Key:      "PRIVATE",
				Name:     "Private Space",
				OwnerID:  "user-123",
				IsPublic: false,
				Created:  time.Now().Unix(),
				Modified: time.Now().Unix(),
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.space.Validate()
			if tt.wantError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDocumentSpace_SetTimestamps(t *testing.T) {
	tests := []struct {
		name    string
		space   *DocumentSpace
		checkFn func(*testing.T, *DocumentSpace, int64)
	}{
		{
			name: "Set both timestamps when zero",
			space: &DocumentSpace{
				Created:  0,
				Modified: 0,
			},
			checkFn: func(t *testing.T, ds *DocumentSpace, before int64) {
				assert.GreaterOrEqual(t, ds.Created, before)
				assert.GreaterOrEqual(t, ds.Modified, before)
				assert.Equal(t, ds.Created, ds.Modified)
			},
		},
		{
			name: "Only update modified when created exists",
			space: &DocumentSpace{
				Created:  1234567890,
				Modified: 0,
			},
			checkFn: func(t *testing.T, ds *DocumentSpace, before int64) {
				assert.Equal(t, int64(1234567890), ds.Created)
				assert.GreaterOrEqual(t, ds.Modified, before)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			before := time.Now().Unix()
			tt.space.SetTimestamps()
			tt.checkFn(t, tt.space, before)
		})
	}
}

func TestDocumentSpace_Structure(t *testing.T) {
	space := DocumentSpace{
		ID:          "space-123",
		Key:         "TECH",
		Name:        "Technical Documentation",
		Description: "All technical documents",
		OwnerID:     "user-123",
		IsPublic:    true,
		Created:     time.Now().Unix(),
		Modified:    time.Now().Unix(),
		Deleted:     false,
	}

	assert.Equal(t, "space-123", space.ID)
	assert.Equal(t, "TECH", space.Key)
	assert.Equal(t, "Technical Documentation", space.Name)
	assert.Equal(t, "All technical documents", space.Description)
	assert.Equal(t, "user-123", space.OwnerID)
	assert.True(t, space.IsPublic)
	assert.Greater(t, space.Created, int64(0))
	assert.Greater(t, space.Modified, int64(0))
	assert.False(t, space.Deleted)
}

func TestDocumentSpace_Keys(t *testing.T) {
	keys := []string{"DOCS", "TECH", "HR", "SALES", "ENG"}

	for _, key := range keys {
		t.Run("Key: "+key, func(t *testing.T) {
			space := &DocumentSpace{
				ID:       "space-123",
				Key:      key,
				Name:     "Test Space",
				OwnerID:  "user-123",
				Created:  time.Now().Unix(),
				Modified: time.Now().Unix(),
			}

			err := space.Validate()
			assert.NoError(t, err)
		})
	}
}

// ================================================================
// DocumentType Tests
// ================================================================

func TestDocumentType_Validate(t *testing.T) {
	validType := &DocumentType{
		ID:          "type-page",
		Key:         "page",
		Name:        "Page",
		Description: "Standard document page",
		Icon:        "📄",
		SchemaJSON:  `{"type": "object"}`,
		Created:     time.Now().Unix(),
		Modified:    time.Now().Unix(),
		Deleted:     false,
	}

	tests := []struct {
		name      string
		docType   *DocumentType
		wantError bool
		errorMsg  string
	}{
		{
			name:      "Valid document type",
			docType:   validType,
			wantError: false,
		},
		{
			name: "Empty ID",
			docType: &DocumentType{
				ID:       "",
				Key:      "page",
				Name:     "Page",
				Created:  time.Now().Unix(),
				Modified: time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "document type ID cannot be empty",
		},
		{
			name: "Empty Key",
			docType: &DocumentType{
				ID:       "type-page",
				Key:      "",
				Name:     "Page",
				Created:  time.Now().Unix(),
				Modified: time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "document type key cannot be empty",
		},
		{
			name: "Empty Name",
			docType: &DocumentType{
				ID:       "type-page",
				Key:      "page",
				Name:     "",
				Created:  time.Now().Unix(),
				Modified: time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "document type name cannot be empty",
		},
		{
			name: "Zero Created timestamp",
			docType: &DocumentType{
				ID:       "type-page",
				Key:      "page",
				Name:     "Page",
				Created:  0,
				Modified: time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "document type created timestamp cannot be zero",
		},
		{
			name: "Zero Modified timestamp",
			docType: &DocumentType{
				ID:       "type-page",
				Key:      "page",
				Name:     "Page",
				Created:  time.Now().Unix(),
				Modified: 0,
			},
			wantError: true,
			errorMsg:  "document type modified timestamp cannot be zero",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.docType.Validate()
			if tt.wantError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDocumentType_SetTimestamps(t *testing.T) {
	tests := []struct {
		name    string
		docType *DocumentType
		checkFn func(*testing.T, *DocumentType, int64)
	}{
		{
			name: "Set both timestamps when zero",
			docType: &DocumentType{
				Created:  0,
				Modified: 0,
			},
			checkFn: func(t *testing.T, dt *DocumentType, before int64) {
				assert.GreaterOrEqual(t, dt.Created, before)
				assert.GreaterOrEqual(t, dt.Modified, before)
				assert.Equal(t, dt.Created, dt.Modified)
			},
		},
		{
			name: "Only update modified when created exists",
			docType: &DocumentType{
				Created:  1234567890,
				Modified: 0,
			},
			checkFn: func(t *testing.T, dt *DocumentType, before int64) {
				assert.Equal(t, int64(1234567890), dt.Created)
				assert.GreaterOrEqual(t, dt.Modified, before)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			before := time.Now().Unix()
			tt.docType.SetTimestamps()
			tt.checkFn(t, tt.docType, before)
		})
	}
}

func TestDocumentType_Structure(t *testing.T) {
	docType := DocumentType{
		ID:          "type-blog",
		Key:         "blog",
		Name:        "Blog Post",
		Description: "Blog post document",
		Icon:        "📝",
		SchemaJSON:  `{"type": "object", "properties": {"title": {"type": "string"}}}`,
		Created:     time.Now().Unix(),
		Modified:    time.Now().Unix(),
		Deleted:     false,
	}

	assert.Equal(t, "type-blog", docType.ID)
	assert.Equal(t, "blog", docType.Key)
	assert.Equal(t, "Blog Post", docType.Name)
	assert.Equal(t, "Blog post document", docType.Description)
	assert.Equal(t, "📝", docType.Icon)
	assert.Contains(t, docType.SchemaJSON, "title")
	assert.Greater(t, docType.Created, int64(0))
	assert.Greater(t, docType.Modified, int64(0))
	assert.False(t, docType.Deleted)
}

func TestDocumentType_StandardTypes(t *testing.T) {
	types := []struct {
		key  string
		name string
		icon string
	}{
		{"page", "Page", "📄"},
		{"blog", "Blog Post", "📝"},
		{"template", "Template", "📋"},
		{"whiteboard", "Whiteboard", "🎨"},
	}

	for _, typ := range types {
		t.Run("Type: "+typ.key, func(t *testing.T) {
			docType := &DocumentType{
				ID:       "type-" + typ.key,
				Key:      typ.key,
				Name:     typ.name,
				Icon:     typ.icon,
				Created:  time.Now().Unix(),
				Modified: time.Now().Unix(),
			}

			err := docType.Validate()
			assert.NoError(t, err)
			assert.Equal(t, typ.key, docType.Key)
			assert.Equal(t, typ.name, docType.Name)
			assert.Equal(t, typ.icon, docType.Icon)
		})
	}
}

// ================================================================
// Benchmark Tests
// ================================================================

func BenchmarkDocumentSpace_Validate(b *testing.B) {
	space := &DocumentSpace{
		ID:       "space-123",
		Key:      "DOCS",
		Name:     "Documentation",
		OwnerID:  "user-123",
		Created:  time.Now().Unix(),
		Modified: time.Now().Unix(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = space.Validate()
	}
}

func BenchmarkDocumentSpace_SetTimestamps(b *testing.B) {
	space := &DocumentSpace{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		space.SetTimestamps()
	}
}

func BenchmarkDocumentType_Validate(b *testing.B) {
	docType := &DocumentType{
		ID:       "type-page",
		Key:      "page",
		Name:     "Page",
		Created:  time.Now().Unix(),
		Modified: time.Now().Unix(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = docType.Validate()
	}
}

func BenchmarkDocumentType_SetTimestamps(b *testing.B) {
	docType := &DocumentType{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		docType.SetTimestamps()
	}
}
