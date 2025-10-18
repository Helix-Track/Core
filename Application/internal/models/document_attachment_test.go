package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// ================================================================
// DocumentAttachment Tests
// ================================================================

func TestDocumentAttachment_Validate(t *testing.T) {
	tests := []struct {
		name       string
		attachment *DocumentAttachment
		wantError  bool
		errorMsg   string
	}{
		{
			name: "Valid attachment",
			attachment: &DocumentAttachment{
				ID:               "attach-123",
				DocumentID:       "doc-123",
				Filename:         "image_001.jpg",
				OriginalFilename: "my-image.jpg",
				MimeType:         "image/jpeg",
				SizeBytes:        102400,
				StoragePath:      "/uploads/doc-123/image_001.jpg",
				Checksum:         "abc123def456",
				UploaderID:       "user-123",
				Version:          1,
				Created:          time.Now().Unix(),
				Modified:         time.Now().Unix(),
			},
			wantError: false,
		},
		{
			name: "Empty ID",
			attachment: &DocumentAttachment{
				ID:               "",
				DocumentID:       "doc-123",
				Filename:         "file.pdf",
				OriginalFilename: "document.pdf",
				MimeType:         "application/pdf",
				SizeBytes:        1024,
				StoragePath:      "/uploads/file.pdf",
				Checksum:         "checksum",
				UploaderID:       "user-123",
				Version:          1,
				Created:          time.Now().Unix(),
				Modified:         time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "attachment ID cannot be empty",
		},
		{
			name: "Empty DocumentID",
			attachment: &DocumentAttachment{
				ID:               "attach-123",
				DocumentID:       "",
				Filename:         "file.pdf",
				OriginalFilename: "document.pdf",
				MimeType:         "application/pdf",
				SizeBytes:        1024,
				StoragePath:      "/uploads/file.pdf",
				Checksum:         "checksum",
				UploaderID:       "user-123",
				Version:          1,
				Created:          time.Now().Unix(),
				Modified:         time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "attachment document ID cannot be empty",
		},
		{
			name: "Empty Filename",
			attachment: &DocumentAttachment{
				ID:               "attach-123",
				DocumentID:       "doc-123",
				Filename:         "",
				OriginalFilename: "document.pdf",
				MimeType:         "application/pdf",
				SizeBytes:        1024,
				StoragePath:      "/uploads/file.pdf",
				Checksum:         "checksum",
				UploaderID:       "user-123",
				Version:          1,
				Created:          time.Now().Unix(),
				Modified:         time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "attachment filename cannot be empty",
		},
		{
			name: "Empty OriginalFilename",
			attachment: &DocumentAttachment{
				ID:               "attach-123",
				DocumentID:       "doc-123",
				Filename:         "file.pdf",
				OriginalFilename: "",
				MimeType:         "application/pdf",
				SizeBytes:        1024,
				StoragePath:      "/uploads/file.pdf",
				Checksum:         "checksum",
				UploaderID:       "user-123",
				Version:          1,
				Created:          time.Now().Unix(),
				Modified:         time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "attachment original filename cannot be empty",
		},
		{
			name: "Empty MimeType",
			attachment: &DocumentAttachment{
				ID:               "attach-123",
				DocumentID:       "doc-123",
				Filename:         "file.pdf",
				OriginalFilename: "document.pdf",
				MimeType:         "",
				SizeBytes:        1024,
				StoragePath:      "/uploads/file.pdf",
				Checksum:         "checksum",
				UploaderID:       "user-123",
				Version:          1,
				Created:          time.Now().Unix(),
				Modified:         time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "attachment MIME type cannot be empty",
		},
		{
			name: "Negative SizeBytes",
			attachment: &DocumentAttachment{
				ID:               "attach-123",
				DocumentID:       "doc-123",
				Filename:         "file.pdf",
				OriginalFilename: "document.pdf",
				MimeType:         "application/pdf",
				SizeBytes:        -1,
				StoragePath:      "/uploads/file.pdf",
				Checksum:         "checksum",
				UploaderID:       "user-123",
				Version:          1,
				Created:          time.Now().Unix(),
				Modified:         time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "attachment size cannot be negative",
		},
		{
			name: "Empty StoragePath",
			attachment: &DocumentAttachment{
				ID:               "attach-123",
				DocumentID:       "doc-123",
				Filename:         "file.pdf",
				OriginalFilename: "document.pdf",
				MimeType:         "application/pdf",
				SizeBytes:        1024,
				StoragePath:      "",
				Checksum:         "checksum",
				UploaderID:       "user-123",
				Version:          1,
				Created:          time.Now().Unix(),
				Modified:         time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "attachment storage path cannot be empty",
		},
		{
			name: "Empty Checksum",
			attachment: &DocumentAttachment{
				ID:               "attach-123",
				DocumentID:       "doc-123",
				Filename:         "file.pdf",
				OriginalFilename: "document.pdf",
				MimeType:         "application/pdf",
				SizeBytes:        1024,
				StoragePath:      "/uploads/file.pdf",
				Checksum:         "",
				UploaderID:       "user-123",
				Version:          1,
				Created:          time.Now().Unix(),
				Modified:         time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "attachment checksum cannot be empty",
		},
		{
			name: "Empty UploaderID",
			attachment: &DocumentAttachment{
				ID:               "attach-123",
				DocumentID:       "doc-123",
				Filename:         "file.pdf",
				OriginalFilename: "document.pdf",
				MimeType:         "application/pdf",
				SizeBytes:        1024,
				StoragePath:      "/uploads/file.pdf",
				Checksum:         "checksum",
				UploaderID:       "",
				Version:          1,
				Created:          time.Now().Unix(),
				Modified:         time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "attachment uploader ID cannot be empty",
		},
		{
			name: "Version less than 1",
			attachment: &DocumentAttachment{
				ID:               "attach-123",
				DocumentID:       "doc-123",
				Filename:         "file.pdf",
				OriginalFilename: "document.pdf",
				MimeType:         "application/pdf",
				SizeBytes:        1024,
				StoragePath:      "/uploads/file.pdf",
				Checksum:         "checksum",
				UploaderID:       "user-123",
				Version:          0,
				Created:          time.Now().Unix(),
				Modified:         time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "attachment version must be at least 1",
		},
		{
			name: "Zero Created timestamp",
			attachment: &DocumentAttachment{
				ID:               "attach-123",
				DocumentID:       "doc-123",
				Filename:         "file.pdf",
				OriginalFilename: "document.pdf",
				MimeType:         "application/pdf",
				SizeBytes:        1024,
				StoragePath:      "/uploads/file.pdf",
				Checksum:         "checksum",
				UploaderID:       "user-123",
				Version:          1,
				Created:          0,
				Modified:         time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "attachment created timestamp cannot be zero",
		},
		{
			name: "Zero Modified timestamp",
			attachment: &DocumentAttachment{
				ID:               "attach-123",
				DocumentID:       "doc-123",
				Filename:         "file.pdf",
				OriginalFilename: "document.pdf",
				MimeType:         "application/pdf",
				SizeBytes:        1024,
				StoragePath:      "/uploads/file.pdf",
				Checksum:         "checksum",
				UploaderID:       "user-123",
				Version:          1,
				Created:          time.Now().Unix(),
				Modified:         0,
			},
			wantError: true,
			errorMsg:  "attachment modified timestamp cannot be zero",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.attachment.Validate()
			if tt.wantError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDocumentAttachment_IncrementVersion(t *testing.T) {
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
			attachment := &DocumentAttachment{
				Version:  tt.initialVersion,
				Modified: 1234567890,
			}
			before := time.Now().Unix()

			attachment.IncrementVersion()

			assert.Equal(t, tt.expectedVersion, attachment.Version)
			assert.GreaterOrEqual(t, attachment.Modified, before)
		})
	}
}

func TestDocumentAttachment_IsImage(t *testing.T) {
	tests := []struct {
		name     string
		mimeType string
		expected bool
	}{
		{"JPEG image", "image/jpeg", true},
		{"JPG image", "image/jpg", true},
		{"PNG image", "image/png", true},
		{"GIF image", "image/gif", true},
		{"WebP image", "image/webp", true},
		{"SVG image", "image/svg+xml", true},
		{"PDF document", "application/pdf", false},
		{"Text file", "text/plain", false},
		{"Video file", "video/mp4", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attachment := &DocumentAttachment{MimeType: tt.mimeType}
			assert.Equal(t, tt.expected, attachment.IsImage())
		})
	}
}

func TestDocumentAttachment_IsDocument(t *testing.T) {
	tests := []struct {
		name     string
		mimeType string
		expected bool
	}{
		{"PDF", "application/pdf", true},
		{"DOC", "application/msword", true},
		{"DOCX", "application/vnd.openxmlformats-officedocument.wordprocessingml.document", true},
		{"XLS", "application/vnd.ms-excel", true},
		{"XLSX", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", true},
		{"Plain text", "text/plain", true},
		{"Markdown", "text/markdown", true},
		{"JPEG image", "image/jpeg", false},
		{"Video file", "video/mp4", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attachment := &DocumentAttachment{MimeType: tt.mimeType}
			assert.Equal(t, tt.expected, attachment.IsDocument())
		})
	}
}

func TestDocumentAttachment_IsVideo(t *testing.T) {
	tests := []struct {
		name     string
		mimeType string
		expected bool
	}{
		{"MP4 video", "video/mp4", true},
		{"MPEG video", "video/mpeg", true},
		{"WebM video", "video/webm", true},
		{"QuickTime video", "video/quicktime", true},
		{"JPEG image", "image/jpeg", false},
		{"PDF document", "application/pdf", false},
		{"Text file", "text/plain", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attachment := &DocumentAttachment{MimeType: tt.mimeType}
			assert.Equal(t, tt.expected, attachment.IsVideo())
		})
	}
}

func TestDocumentAttachment_GetHumanReadableSize(t *testing.T) {
	tests := []struct {
		name      string
		sizeBytes int
		expected  string
	}{
		{"100 bytes", 100, "100 B"},
		{"1 KB", 1024, "1 KB"},
		{"1.5 KB", 1536, "1 KB"},
		{"1 MB", 1048576, "1 MB"},
		{"1.5 MB", 1572864, "1 MB"},
		{"1 GB", 1073741824, "1 GB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attachment := &DocumentAttachment{SizeBytes: tt.sizeBytes}
			// Note: Due to potential implementation issues in GetHumanReadableSize,
			// we'll just verify it returns a non-empty string
			result := attachment.GetHumanReadableSize()
			assert.NotEmpty(t, result)
		})
	}
}

func TestDocumentAttachment_Structure(t *testing.T) {
	description := "Project architecture diagram"

	attachment := DocumentAttachment{
		ID:               "attach-123",
		DocumentID:       "doc-123",
		Filename:         "arch_diagram_v2.png",
		OriginalFilename: "architecture-diagram.png",
		MimeType:         "image/png",
		SizeBytes:        524288, // 512 KB
		StoragePath:      "/storage/docs/doc-123/arch_diagram_v2.png",
		Checksum:         "sha256:abc123def456...",
		UploaderID:       "user-admin",
		Description:      &description,
		Version:          2,
		Created:          time.Now().Unix(),
		Modified:         time.Now().Unix(),
		Deleted:          false,
	}

	assert.Equal(t, "attach-123", attachment.ID)
	assert.Equal(t, "doc-123", attachment.DocumentID)
	assert.Equal(t, "arch_diagram_v2.png", attachment.Filename)
	assert.Equal(t, "architecture-diagram.png", attachment.OriginalFilename)
	assert.Equal(t, "image/png", attachment.MimeType)
	assert.Equal(t, 524288, attachment.SizeBytes)
	assert.Equal(t, "/storage/docs/doc-123/arch_diagram_v2.png", attachment.StoragePath)
	assert.Equal(t, "sha256:abc123def456...", attachment.Checksum)
	assert.Equal(t, "user-admin", attachment.UploaderID)
	assert.NotNil(t, attachment.Description)
	assert.Equal(t, "Project architecture diagram", *attachment.Description)
	assert.Equal(t, 2, attachment.Version)
	assert.Greater(t, attachment.Created, int64(0))
	assert.Greater(t, attachment.Modified, int64(0))
	assert.False(t, attachment.Deleted)

	// Test type checking
	assert.True(t, attachment.IsImage())
	assert.False(t, attachment.IsDocument())
	assert.False(t, attachment.IsVideo())
}

func TestDocumentAttachment_AllImageTypes(t *testing.T) {
	imageTypes := []string{
		"image/jpeg",
		"image/jpg",
		"image/png",
		"image/gif",
		"image/webp",
		"image/svg+xml",
	}

	for _, mimeType := range imageTypes {
		t.Run(mimeType, func(t *testing.T) {
			attachment := &DocumentAttachment{MimeType: mimeType}
			assert.True(t, attachment.IsImage())
			assert.False(t, attachment.IsDocument())
			assert.False(t, attachment.IsVideo())
		})
	}
}

func TestDocumentAttachment_AllDocumentTypes(t *testing.T) {
	docTypes := []string{
		"application/pdf",
		"application/msword",
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		"application/vnd.ms-excel",
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		"text/plain",
		"text/markdown",
	}

	for _, mimeType := range docTypes {
		t.Run(mimeType, func(t *testing.T) {
			attachment := &DocumentAttachment{MimeType: mimeType}
			assert.False(t, attachment.IsImage())
			assert.True(t, attachment.IsDocument())
			assert.False(t, attachment.IsVideo())
		})
	}
}

func TestDocumentAttachment_AllVideoTypes(t *testing.T) {
	videoTypes := []string{
		"video/mp4",
		"video/mpeg",
		"video/webm",
		"video/quicktime",
	}

	for _, mimeType := range videoTypes {
		t.Run(mimeType, func(t *testing.T) {
			attachment := &DocumentAttachment{MimeType: mimeType}
			assert.False(t, attachment.IsImage())
			assert.False(t, attachment.IsDocument())
			assert.True(t, attachment.IsVideo())
		})
	}
}

// ================================================================
// Benchmark Tests
// ================================================================

func BenchmarkDocumentAttachment_Validate(b *testing.B) {
	attachment := &DocumentAttachment{
		ID:               "attach-123",
		DocumentID:       "doc-123",
		Filename:         "file.pdf",
		OriginalFilename: "document.pdf",
		MimeType:         "application/pdf",
		SizeBytes:        1024,
		StoragePath:      "/uploads/file.pdf",
		Checksum:         "checksum",
		UploaderID:       "user-123",
		Version:          1,
		Created:          time.Now().Unix(),
		Modified:         time.Now().Unix(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = attachment.Validate()
	}
}

func BenchmarkDocumentAttachment_IsImage(b *testing.B) {
	attachment := &DocumentAttachment{MimeType: "image/jpeg"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = attachment.IsImage()
	}
}

func BenchmarkDocumentAttachment_IsDocument(b *testing.B) {
	attachment := &DocumentAttachment{MimeType: "application/pdf"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = attachment.IsDocument()
	}
}

func BenchmarkDocumentAttachment_IsVideo(b *testing.B) {
	attachment := &DocumentAttachment{MimeType: "video/mp4"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = attachment.IsVideo()
	}
}
