package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// ================================================================
// Document Tests
// ================================================================

func TestDocument_Validate(t *testing.T) {
	validDocument := &Document{
		ID:          "doc-123",
		Title:       "Test Document",
		SpaceID:     "space-123",
		TypeID:      "type-page",
		CreatorID:   "user-123",
		Version:     1,
		Position:    0,
		IsPublished: false,
		IsArchived:  false,
		Created:     time.Now().Unix(),
		Modified:    time.Now().Unix(),
		Deleted:     false,
	}

	tests := []struct {
		name      string
		document  *Document
		wantError bool
		errorMsg  string
	}{
		{
			name:      "Valid document",
			document:  validDocument,
			wantError: false,
		},
		{
			name: "Empty ID",
			document: &Document{
				ID:        "",
				Title:     "Test",
				SpaceID:   "space-123",
				TypeID:    "type-page",
				CreatorID: "user-123",
				Version:   1,
				Created:   time.Now().Unix(),
				Modified:  time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "document ID cannot be empty",
		},
		{
			name: "Empty Title",
			document: &Document{
				ID:        "doc-123",
				Title:     "",
				SpaceID:   "space-123",
				TypeID:    "type-page",
				CreatorID: "user-123",
				Version:   1,
				Created:   time.Now().Unix(),
				Modified:  time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "document title cannot be empty",
		},
		{
			name: "Empty SpaceID",
			document: &Document{
				ID:        "doc-123",
				Title:     "Test",
				SpaceID:   "",
				TypeID:    "type-page",
				CreatorID: "user-123",
				Version:   1,
				Created:   time.Now().Unix(),
				Modified:  time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "document space ID cannot be empty",
		},
		{
			name: "Empty TypeID",
			document: &Document{
				ID:        "doc-123",
				Title:     "Test",
				SpaceID:   "space-123",
				TypeID:    "",
				CreatorID: "user-123",
				Version:   1,
				Created:   time.Now().Unix(),
				Modified:  time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "document type ID cannot be empty",
		},
		{
			name: "Empty CreatorID",
			document: &Document{
				ID:        "doc-123",
				Title:     "Test",
				SpaceID:   "space-123",
				TypeID:    "type-page",
				CreatorID: "",
				Version:   1,
				Created:   time.Now().Unix(),
				Modified:  time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "document creator ID cannot be empty",
		},
		{
			name: "Version less than 1",
			document: &Document{
				ID:        "doc-123",
				Title:     "Test",
				SpaceID:   "space-123",
				TypeID:    "type-page",
				CreatorID: "user-123",
				Version:   0,
				Created:   time.Now().Unix(),
				Modified:  time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "document version must be at least 1",
		},
		{
			name: "Zero Created timestamp",
			document: &Document{
				ID:        "doc-123",
				Title:     "Test",
				SpaceID:   "space-123",
				TypeID:    "type-page",
				CreatorID: "user-123",
				Version:   1,
				Created:   0,
				Modified:  time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "document created timestamp cannot be zero",
		},
		{
			name: "Zero Modified timestamp",
			document: &Document{
				ID:        "doc-123",
				Title:     "Test",
				SpaceID:   "space-123",
				TypeID:    "type-page",
				CreatorID: "user-123",
				Version:   1,
				Created:   time.Now().Unix(),
				Modified:  0,
			},
			wantError: true,
			errorMsg:  "document modified timestamp cannot be zero",
		},
		{
			name: "Valid with optional fields",
			document: &Document{
				ID:          "doc-123",
				Title:       "Test",
				SpaceID:     "space-123",
				ParentID:    stringPtr("parent-123"),
				TypeID:      "type-page",
				ProjectID:   stringPtr("project-123"),
				CreatorID:   "user-123",
				Version:     1,
				Position:    5,
				IsPublished: true,
				IsArchived:  false,
				PublishDate: int64Ptr(time.Now().Unix()),
				Created:     time.Now().Unix(),
				Modified:    time.Now().Unix(),
				Deleted:     false,
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.document.Validate()
			if tt.wantError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDocument_SetTimestamps(t *testing.T) {
	tests := []struct {
		name     string
		document *Document
		checkFn  func(*testing.T, *Document, int64)
	}{
		{
			name: "Set both timestamps when zero",
			document: &Document{
				Created:  0,
				Modified: 0,
			},
			checkFn: func(t *testing.T, d *Document, before int64) {
				assert.GreaterOrEqual(t, d.Created, before)
				assert.GreaterOrEqual(t, d.Modified, before)
				assert.Equal(t, d.Created, d.Modified)
			},
		},
		{
			name: "Only update modified when created exists",
			document: &Document{
				Created:  1234567890,
				Modified: 0,
			},
			checkFn: func(t *testing.T, d *Document, before int64) {
				assert.Equal(t, int64(1234567890), d.Created)
				assert.GreaterOrEqual(t, d.Modified, before)
			},
		},
		{
			name: "Update modified when both exist",
			document: &Document{
				Created:  1234567890,
				Modified: 1234567890,
			},
			checkFn: func(t *testing.T, d *Document, before int64) {
				assert.Equal(t, int64(1234567890), d.Created)
				assert.Greater(t, d.Modified, int64(1234567890))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			before := time.Now().Unix()
			tt.document.SetTimestamps()
			tt.checkFn(t, tt.document, before)
		})
	}
}

func TestDocument_IncrementVersion(t *testing.T) {
	tests := []struct {
		name            string
		initialVersion  int
		expectedVersion int
	}{
		{"From version 1", 1, 2},
		{"From version 5", 5, 6},
		{"From version 100", 100, 101},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := &Document{
				Version:  tt.initialVersion,
				Modified: 1234567890,
			}
			before := time.Now().Unix()

			doc.IncrementVersion()

			assert.Equal(t, tt.expectedVersion, doc.Version)
			assert.GreaterOrEqual(t, doc.Modified, before)
		})
	}
}

func TestDocument_Structure(t *testing.T) {
	parentID := "parent-123"
	projectID := "project-123"
	publishDate := time.Now().Unix()

	doc := Document{
		ID:          "doc-123",
		Title:       "Test Document",
		SpaceID:     "space-123",
		ParentID:    &parentID,
		TypeID:      "type-page",
		ProjectID:   &projectID,
		CreatorID:   "user-123",
		Version:     1,
		Position:    0,
		IsPublished: true,
		IsArchived:  false,
		PublishDate: &publishDate,
		Created:     time.Now().Unix(),
		Modified:    time.Now().Unix(),
		Deleted:     false,
	}

	// Test all fields are accessible
	assert.Equal(t, "doc-123", doc.ID)
	assert.Equal(t, "Test Document", doc.Title)
	assert.Equal(t, "space-123", doc.SpaceID)
	assert.NotNil(t, doc.ParentID)
	assert.Equal(t, "parent-123", *doc.ParentID)
	assert.Equal(t, "type-page", doc.TypeID)
	assert.NotNil(t, doc.ProjectID)
	assert.Equal(t, "project-123", *doc.ProjectID)
	assert.Equal(t, "user-123", doc.CreatorID)
	assert.Equal(t, 1, doc.Version)
	assert.Equal(t, 0, doc.Position)
	assert.True(t, doc.IsPublished)
	assert.False(t, doc.IsArchived)
	assert.NotNil(t, doc.PublishDate)
	assert.Greater(t, doc.Created, int64(0))
	assert.Greater(t, doc.Modified, int64(0))
	assert.False(t, doc.Deleted)
}

func TestDocument_OptimisticLocking(t *testing.T) {
	doc := &Document{
		ID:        "doc-123",
		Title:     "Test",
		SpaceID:   "space-123",
		TypeID:    "type-page",
		CreatorID: "user-123",
		Version:   1,
		Created:   time.Now().Unix(),
		Modified:  time.Now().Unix(),
	}

	// Simulate multiple version increments
	for i := 1; i <= 5; i++ {
		assert.Equal(t, i, doc.Version)
		doc.IncrementVersion()
		assert.Equal(t, i+1, doc.Version)
	}
}

// ================================================================
// DocumentContent Tests
// ================================================================

func TestDocumentContent_Validate(t *testing.T) {
	validContent := &DocumentContent{
		ID:          "content-123",
		DocumentID:  "doc-123",
		Version:     1,
		ContentType: "html",
		SizeBytes:   1024,
		Created:     time.Now().Unix(),
		Modified:    time.Now().Unix(),
		Deleted:     false,
	}

	tests := []struct {
		name      string
		content   *DocumentContent
		wantError bool
		errorMsg  string
	}{
		{
			name:      "Valid content",
			content:   validContent,
			wantError: false,
		},
		{
			name: "Empty ID",
			content: &DocumentContent{
				ID:          "",
				DocumentID:  "doc-123",
				Version:     1,
				ContentType: "html",
				Created:     time.Now().Unix(),
				Modified:    time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "document content ID cannot be empty",
		},
		{
			name: "Empty DocumentID",
			content: &DocumentContent{
				ID:          "content-123",
				DocumentID:  "",
				Version:     1,
				ContentType: "html",
				Created:     time.Now().Unix(),
				Modified:    time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "document content document ID cannot be empty",
		},
		{
			name: "Version less than 1",
			content: &DocumentContent{
				ID:          "content-123",
				DocumentID:  "doc-123",
				Version:     0,
				ContentType: "html",
				Created:     time.Now().Unix(),
				Modified:    time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "document content version must be at least 1",
		},
		{
			name: "Empty ContentType",
			content: &DocumentContent{
				ID:          "content-123",
				DocumentID:  "doc-123",
				Version:     1,
				ContentType: "",
				Created:     time.Now().Unix(),
				Modified:    time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "document content type cannot be empty",
		},
		{
			name: "Invalid ContentType",
			content: &DocumentContent{
				ID:          "content-123",
				DocumentID:  "doc-123",
				Version:     1,
				ContentType: "xml",
				Created:     time.Now().Unix(),
				Modified:    time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "invalid content type",
		},
		{
			name: "Zero Created timestamp",
			content: &DocumentContent{
				ID:          "content-123",
				DocumentID:  "doc-123",
				Version:     1,
				ContentType: "html",
				Created:     0,
				Modified:    time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "document content created timestamp cannot be zero",
		},
		{
			name: "Zero Modified timestamp",
			content: &DocumentContent{
				ID:          "content-123",
				DocumentID:  "doc-123",
				Version:     1,
				ContentType: "html",
				Created:     time.Now().Unix(),
				Modified:    0,
			},
			wantError: true,
			errorMsg:  "document content modified timestamp cannot be zero",
		},
		{
			name: "Valid HTML content",
			content: &DocumentContent{
				ID:          "content-123",
				DocumentID:  "doc-123",
				Version:     1,
				ContentType: "html",
				Created:     time.Now().Unix(),
				Modified:    time.Now().Unix(),
			},
			wantError: false,
		},
		{
			name: "Valid Markdown content",
			content: &DocumentContent{
				ID:          "content-123",
				DocumentID:  "doc-123",
				Version:     1,
				ContentType: "markdown",
				Created:     time.Now().Unix(),
				Modified:    time.Now().Unix(),
			},
			wantError: false,
		},
		{
			name: "Valid Plain content",
			content: &DocumentContent{
				ID:          "content-123",
				DocumentID:  "doc-123",
				Version:     1,
				ContentType: "plain",
				Created:     time.Now().Unix(),
				Modified:    time.Now().Unix(),
			},
			wantError: false,
		},
		{
			name: "Valid Storage format",
			content: &DocumentContent{
				ID:          "content-123",
				DocumentID:  "doc-123",
				Version:     1,
				ContentType: "storage",
				Created:     time.Now().Unix(),
				Modified:    time.Now().Unix(),
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.content.Validate()
			if tt.wantError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDocumentContent_SetTimestamps(t *testing.T) {
	tests := []struct {
		name    string
		content *DocumentContent
		checkFn func(*testing.T, *DocumentContent, int64)
	}{
		{
			name: "Set both timestamps when zero",
			content: &DocumentContent{
				Created:  0,
				Modified: 0,
			},
			checkFn: func(t *testing.T, dc *DocumentContent, before int64) {
				assert.GreaterOrEqual(t, dc.Created, before)
				assert.GreaterOrEqual(t, dc.Modified, before)
				assert.Equal(t, dc.Created, dc.Modified)
			},
		},
		{
			name: "Only update modified when created exists",
			content: &DocumentContent{
				Created:  1234567890,
				Modified: 0,
			},
			checkFn: func(t *testing.T, dc *DocumentContent, before int64) {
				assert.Equal(t, int64(1234567890), dc.Created)
				assert.GreaterOrEqual(t, dc.Modified, before)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			before := time.Now().Unix()
			tt.content.SetTimestamps()
			tt.checkFn(t, tt.content, before)
		})
	}
}

func TestDocumentContent_Structure(t *testing.T) {
	content := "# Test Document\n\nThis is a test."
	hash := "sha256hash123"

	dc := DocumentContent{
		ID:          "content-123",
		DocumentID:  "doc-123",
		Version:     1,
		ContentType: "markdown",
		Content:     &content,
		ContentHash: &hash,
		SizeBytes:   1024,
		Created:     time.Now().Unix(),
		Modified:    time.Now().Unix(),
		Deleted:     false,
	}

	assert.Equal(t, "content-123", dc.ID)
	assert.Equal(t, "doc-123", dc.DocumentID)
	assert.Equal(t, 1, dc.Version)
	assert.Equal(t, "markdown", dc.ContentType)
	assert.NotNil(t, dc.Content)
	assert.Equal(t, "# Test Document\n\nThis is a test.", *dc.Content)
	assert.NotNil(t, dc.ContentHash)
	assert.Equal(t, "sha256hash123", *dc.ContentHash)
	assert.Equal(t, 1024, dc.SizeBytes)
	assert.Greater(t, dc.Created, int64(0))
	assert.Greater(t, dc.Modified, int64(0))
	assert.False(t, dc.Deleted)
}

func TestDocumentContent_AllContentTypes(t *testing.T) {
	contentTypes := []string{"html", "markdown", "plain", "storage"}

	for _, ct := range contentTypes {
		t.Run("ContentType: "+ct, func(t *testing.T) {
			dc := &DocumentContent{
				ID:          "content-123",
				DocumentID:  "doc-123",
				Version:     1,
				ContentType: ct,
				Created:     time.Now().Unix(),
				Modified:    time.Now().Unix(),
			}

			err := dc.Validate()
			assert.NoError(t, err)
		})
	}
}

// ================================================================
// Benchmark Tests
// ================================================================

func BenchmarkDocument_Validate(b *testing.B) {
	doc := &Document{
		ID:        "doc-123",
		Title:     "Test Document",
		SpaceID:   "space-123",
		TypeID:    "type-page",
		CreatorID: "user-123",
		Version:   1,
		Created:   time.Now().Unix(),
		Modified:  time.Now().Unix(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = doc.Validate()
	}
}

func BenchmarkDocument_SetTimestamps(b *testing.B) {
	doc := &Document{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		doc.SetTimestamps()
	}
}

func BenchmarkDocument_IncrementVersion(b *testing.B) {
	doc := &Document{Version: 1}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		doc.IncrementVersion()
	}
}

func BenchmarkDocumentContent_Validate(b *testing.B) {
	dc := &DocumentContent{
		ID:          "content-123",
		DocumentID:  "doc-123",
		Version:     1,
		ContentType: "html",
		Created:     time.Now().Unix(),
		Modified:    time.Now().Unix(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = dc.Validate()
	}
}

// ================================================================
// Helper Functions
// ================================================================

func stringPtr(s string) *string {
	return &s
}

func int64Ptr(i int64) *int64 {
	return &i
}
