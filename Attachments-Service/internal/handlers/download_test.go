package handlers

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/helixtrack/attachments-service/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// Mock reader for testing
type mockReadCloser struct {
	*bytes.Reader
}

func (m *mockReadCloser) Close() error {
	return nil
}

func newMockReadCloser(data string) *mockReadCloser {
	return &mockReadCloser{
		Reader: bytes.NewReader([]byte(data)),
	}
}

// Helper function to create download handler
func createDownloadHandler() (*DownloadHandler, *MockDeduplicationEngine, *MockPrometheusMetrics) {
	mockEngine := &MockDeduplicationEngine{}
	mockMetrics := &MockPrometheusMetrics{}
	logger := zap.NewNop()

	config := &DownloadConfig{
		EnableRangeRequests: true,
		EnableCaching:       true,
		CacheMaxAge:         3600,
		BufferSize:          32 * 1024,
	}

	handler := NewDownloadHandler(mockEngine, mockMetrics, logger, config)

	return handler, mockEngine, mockMetrics
}

// Tests

func TestNewDownloadHandler(t *testing.T) {
	t.Run("with nil config uses defaults", func(t *testing.T) {
		_ = &MockDeduplicationEngine{}
		_ = &MockPrometheusMetrics{}
		logger := zap.NewNop()

		handler := NewDownloadHandler(nil, nil, logger, nil)

		assert.NotNil(t, handler)
		assert.NotNil(t, handler.config)
		assert.True(t, handler.config.EnableRangeRequests)
		assert.True(t, handler.config.EnableCaching)
		assert.Equal(t, 3600, handler.config.CacheMaxAge)
	})

	t.Run("with custom config", func(t *testing.T) {
		_ = &MockDeduplicationEngine{}
		_ = &MockPrometheusMetrics{}
		logger := zap.NewNop()

		config := &DownloadConfig{
			EnableRangeRequests: false,
			EnableCaching:       false,
			CacheMaxAge:         7200,
			BufferSize:          64 * 1024,
		}

		handler := NewDownloadHandler(nil, nil, logger, config)

		assert.False(t, handler.config.EnableRangeRequests)
		assert.False(t, handler.config.EnableCaching)
		assert.Equal(t, 7200, handler.config.CacheMaxAge)
	})
}

func TestDownloadHandler_Handle_Success(t *testing.T) {
	handler, mockEngine, mockMetrics := createDownloadHandler()

	fileContent := "This is test file content"
	reference := &models.AttachmentReference{
		ID:         "ref-123",
		Filename:   "test.pdf",
		EntityType: "ticket",
		EntityID:   "TICKET-123",
	}
	file := &models.AttachmentFile{
		Hash:      "abc123",
		SizeBytes: int64(len(fileContent)),
		MimeType:  "application/pdf",
		Extension: ".pdf",
	}

	mockEngine.On("DownloadFile", mock.Anything, "ref-123").Return(
		newMockReadCloser(fileContent),
		reference,
		file,
		nil,
	)

	mockMetrics.On("RecordDownload", "success", int64(len(fileContent)), mock.Anything, false).Return()

	// Create request
	req, _ := http.NewRequest("GET", "/download/ref-123", nil)
	w := httptest.NewRecorder()

	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Params = gin.Params{
		{Key: "reference_id", Value: "ref-123"},
	}

	// Execute
	handler.Handle(ctx)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, fileContent, w.Body.String())
	assert.Equal(t, "application/pdf", w.Header().Get("Content-Type"))
	assert.Equal(t, "attachment; filename=\"test.pdf\"", w.Header().Get("Content-Disposition"))
	assert.Equal(t, "bytes", w.Header().Get("Accept-Ranges"))
	mockEngine.AssertExpectations(t)
	mockMetrics.AssertExpectations(t)
}

func TestDownloadHandler_Handle_MissingReferenceID(t *testing.T) {
	handler, _, _ := createDownloadHandler()

	// Create request without reference_id
	req, _ := http.NewRequest("GET", "/download/", nil)
	w := httptest.NewRecorder()

	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req

	// Execute
	handler.Handle(ctx)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "reference_id is required")
}

func TestDownloadHandler_Handle_FileNotFound(t *testing.T) {
	handler, mockEngine, _ := createDownloadHandler()
	handler.metrics = nil // Disable metrics for this test

	mockEngine.On("DownloadFile", mock.Anything, "nonexistent").Return(
		nil,
		nil,
		nil,
		errors.New("reference not found"),
	)

	// Create request
	req, _ := http.NewRequest("GET", "/download/nonexistent", nil)
	w := httptest.NewRecorder()

	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Params = gin.Params{
		{Key: "reference_id", Value: "nonexistent"},
	}

	// Execute
	handler.Handle(ctx)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "file not found")
	mockEngine.AssertExpectations(t)
}

func TestDownloadHandler_HandleInline_Success(t *testing.T) {
	handler, mockEngine, mockMetrics := createDownloadHandler()

	fileContent := "PDF content for inline viewing"
	reference := &models.AttachmentReference{
		ID:       "ref-456",
		Filename: "document.pdf",
	}
	file := &models.AttachmentFile{
		Hash:      "def456",
		SizeBytes: int64(len(fileContent)),
		MimeType:  "application/pdf",
		Extension: ".pdf",
	}

	mockEngine.On("DownloadFile", mock.Anything, "ref-456").Return(
		newMockReadCloser(fileContent),
		reference,
		file,
		nil,
	)

	mockMetrics.On("RecordDownload", "success", int64(len(fileContent)), mock.Anything, false).Return()

	// Create request
	req, _ := http.NewRequest("GET", "/download/ref-456/inline", nil)
	w := httptest.NewRecorder()

	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Params = gin.Params{
		{Key: "reference_id", Value: "ref-456"},
	}

	// Execute
	handler.HandleInline(ctx)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, fileContent, w.Body.String())
	assert.Equal(t, "application/pdf", w.Header().Get("Content-Type"))
	assert.Equal(t, "inline; filename=\"document.pdf\"", w.Header().Get("Content-Disposition"))
	assert.Equal(t, "public, max-age=3600", w.Header().Get("Cache-Control"))
	assert.Equal(t, "def456", w.Header().Get("ETag"))
	mockEngine.AssertExpectations(t)
	mockMetrics.AssertExpectations(t)
}

func TestDownloadHandler_HandleInline_NoCaching(t *testing.T) {
	handler, mockEngine, mockMetrics := createDownloadHandler()
	handler.config.EnableCaching = false

	fileContent := "Content"
	reference := &models.AttachmentReference{
		ID:       "ref-789",
		Filename: "test.txt",
	}
	file := &models.AttachmentFile{
		Hash:      "ghi789",
		SizeBytes: int64(len(fileContent)),
		MimeType:  "text/plain",
	}

	mockEngine.On("DownloadFile", mock.Anything, "ref-789").Return(
		newMockReadCloser(fileContent),
		reference,
		file,
		nil,
	)

	mockMetrics.On("RecordDownload", "success", int64(len(fileContent)), mock.Anything, false).Return()

	// Create request
	req, _ := http.NewRequest("GET", "/download/ref-789/inline", nil)
	w := httptest.NewRecorder()

	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Params = gin.Params{
		{Key: "reference_id", Value: "ref-789"},
	}

	// Execute
	handler.HandleInline(ctx)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	// Should not have caching headers
	assert.Empty(t, w.Header().Get("ETag"))
	mockEngine.AssertExpectations(t)
	mockMetrics.AssertExpectations(t)
}

func TestDownloadHandler_HandleMetadata_Success(t *testing.T) {
	handler, mockEngine, _ := createDownloadHandler()
	handler.metrics = nil // Disable metrics for this test

	testDesc := "Test file"
	reference := &models.AttachmentReference{
		ID:          "ref-123",
		Filename:    "test.pdf",
		EntityType:  "ticket",
		EntityID:    "TICKET-123",
		UploaderID:  "user123",
		Description: &testDesc,
		Tags:        []string{"important"},
		Created:     time.Now().Unix(),
	}
	file := &models.AttachmentFile{
		Hash:      "abc123",
		SizeBytes: 1024,
		MimeType:  "application/pdf",
		Extension: ".pdf",
	}

	mockEngine.On("DownloadFile", mock.Anything, "ref-123").Return(
		newMockReadCloser(""),
		reference,
		file,
		nil,
	)

	// Create request
	req, _ := http.NewRequest("GET", "/download/ref-123/metadata", nil)
	w := httptest.NewRecorder()

	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Params = gin.Params{
		{Key: "reference_id", Value: "ref-123"},
	}

	// Execute
	handler.HandleMetadata(ctx)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "ref-123")
	assert.Contains(t, w.Body.String(), "test.pdf")
	assert.Contains(t, w.Body.String(), "abc123")
	assert.Contains(t, w.Body.String(), "application/pdf")
	assert.Contains(t, w.Body.String(), "user123")
	mockEngine.AssertExpectations(t)
}

func TestDownloadHandler_HandleMetadata_NotFound(t *testing.T) {
	handler, mockEngine, _ := createDownloadHandler()
	handler.metrics = nil // Disable metrics for this test

	mockEngine.On("DownloadFile", mock.Anything, "nonexistent").Return(
		nil,
		nil,
		nil,
		errors.New("not found"),
	)

	// Create request
	req, _ := http.NewRequest("GET", "/download/nonexistent/metadata", nil)
	w := httptest.NewRecorder()

	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Params = gin.Params{
		{Key: "reference_id", Value: "nonexistent"},
	}

	// Execute
	handler.HandleMetadata(ctx)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "file not found")
	mockEngine.AssertExpectations(t)
}

func TestDownloadHandler_RangeRequest_StartEnd(t *testing.T) {
	handler, mockEngine, _ := createDownloadHandler()
	handler.metrics = nil // Disable metrics for this test

	fileContent := "0123456789ABCDEFGHIJ" // 20 bytes
	reference := &models.AttachmentReference{
		ID:       "ref-range",
		Filename: "test.txt",
	}
	file := &models.AttachmentFile{
		Hash:      "hash-range",
		SizeBytes: int64(len(fileContent)),
		MimeType:  "text/plain",
	}

	mockEngine.On("DownloadFile", mock.Anything, "ref-range").Return(
		newMockReadCloser(fileContent),
		reference,
		file,
		nil,
	)

	// Create request with range header
	req, _ := http.NewRequest("GET", "/download/ref-range", nil)
	req.Header.Set("Range", "bytes=0-9")
	w := httptest.NewRecorder()

	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Params = gin.Params{
		{Key: "reference_id", Value: "ref-range"},
	}

	// Execute
	handler.Handle(ctx)

	// Assert
	assert.Equal(t, http.StatusPartialContent, w.Code)
	assert.Equal(t, "0123456789", w.Body.String())
	assert.Equal(t, "bytes 0-9/20", w.Header().Get("Content-Range"))
	assert.Equal(t, "10", w.Header().Get("Content-Length"))
	mockEngine.AssertExpectations(t)
}

func TestDownloadHandler_RangeRequest_FromStart(t *testing.T) {
	handler, mockEngine, _ := createDownloadHandler()
	handler.metrics = nil // Disable metrics for this test

	fileContent := "0123456789ABCDEFGHIJ" // 20 bytes
	reference := &models.AttachmentReference{
		ID:       "ref-range2",
		Filename: "test.txt",
	}
	file := &models.AttachmentFile{
		Hash:      "hash-range2",
		SizeBytes: int64(len(fileContent)),
		MimeType:  "text/plain",
	}

	mockEngine.On("DownloadFile", mock.Anything, "ref-range2").Return(
		newMockReadCloser(fileContent),
		reference,
		file,
		nil,
	)

	// Create request with range header (from position to end)
	req, _ := http.NewRequest("GET", "/download/ref-range2", nil)
	req.Header.Set("Range", "bytes=10-")
	w := httptest.NewRecorder()

	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Params = gin.Params{
		{Key: "reference_id", Value: "ref-range2"},
	}

	// Execute
	handler.Handle(ctx)

	// Assert
	assert.Equal(t, http.StatusPartialContent, w.Code)
	assert.Equal(t, "ABCDEFGHIJ", w.Body.String())
	assert.Equal(t, "bytes 10-19/20", w.Header().Get("Content-Range"))
	mockEngine.AssertExpectations(t)
}

func TestDownloadHandler_RangeRequest_Suffix(t *testing.T) {
	handler, mockEngine, _ := createDownloadHandler()
	handler.metrics = nil // Disable metrics for this test

	fileContent := "0123456789ABCDEFGHIJ" // 20 bytes
	reference := &models.AttachmentReference{
		ID:       "ref-range3",
		Filename: "test.txt",
	}
	file := &models.AttachmentFile{
		Hash:      "hash-range3",
		SizeBytes: int64(len(fileContent)),
		MimeType:  "text/plain",
	}

	mockEngine.On("DownloadFile", mock.Anything, "ref-range3").Return(
		newMockReadCloser(fileContent),
		reference,
		file,
		nil,
	)

	// Create request with suffix range (last N bytes)
	req, _ := http.NewRequest("GET", "/download/ref-range3", nil)
	req.Header.Set("Range", "bytes=-5")
	w := httptest.NewRecorder()

	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Params = gin.Params{
		{Key: "reference_id", Value: "ref-range3"},
	}

	// Execute
	handler.Handle(ctx)

	// Assert
	assert.Equal(t, http.StatusPartialContent, w.Code)
	assert.Equal(t, "FGHIJ", w.Body.String())
	assert.Equal(t, "bytes 15-19/20", w.Header().Get("Content-Range"))
	mockEngine.AssertExpectations(t)
}

func TestParseRangeHeader(t *testing.T) {
	tests := []struct {
		name       string
		header     string
		fileSize   int64
		wantStart  int64
		wantEnd    int64
		wantErr    bool
	}{
		{
			name:      "start and end",
			header:    "bytes=0-999",
			fileSize:  1000,
			wantStart: 0,
			wantEnd:   999,
			wantErr:   false,
		},
		{
			name:      "from start to end",
			header:    "bytes=500-",
			fileSize:  1000,
			wantStart: 500,
			wantEnd:   999,
			wantErr:   false,
		},
		{
			name:      "suffix range",
			header:    "bytes=-200",
			fileSize:  1000,
			wantStart: 800,
			wantEnd:   999,
			wantErr:   false,
		},
		{
			name:     "invalid format",
			header:   "bytes=abc",
			fileSize: 1000,
			wantErr:  true,
		},
		{
			name:     "no bytes prefix",
			header:   "0-999",
			fileSize: 1000,
			wantErr:  true,
		},
		{
			name:     "out of bounds",
			header:   "bytes=0-1500",
			fileSize: 1000,
			wantErr:  true,
		},
		{
			name:     "start > end",
			header:   "bytes=900-100",
			fileSize: 1000,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, end, err := parseRangeHeader(tt.header, tt.fileSize)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantStart, start)
				assert.Equal(t, tt.wantEnd, end)
			}
		})
	}
}

func TestDownloadHandler_RangeDisabled(t *testing.T) {
	handler, mockEngine, _ := createDownloadHandler()
	handler.metrics = nil // Disable metrics for this test
	handler.config.EnableRangeRequests = false
	handler.metrics = nil // Disable metrics for this test

	fileContent := "Full content without range support"
	reference := &models.AttachmentReference{
		ID:       "ref-no-range",
		Filename: "test.txt",
	}
	file := &models.AttachmentFile{
		Hash:      "hash-no-range",
		SizeBytes: int64(len(fileContent)),
		MimeType:  "text/plain",
	}

	mockEngine.On("DownloadFile", mock.Anything, "ref-no-range").Return(
		newMockReadCloser(fileContent),
		reference,
		file,
		nil,
	)

	// Create request with range header (should be ignored)
	req, _ := http.NewRequest("GET", "/download/ref-no-range", nil)
	req.Header.Set("Range", "bytes=0-10")
	w := httptest.NewRecorder()

	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Params = gin.Params{
		{Key: "reference_id", Value: "ref-no-range"},
	}

	// Execute
	handler.Handle(ctx)

	// Assert - full content returned, not partial
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, fileContent, w.Body.String())
	assert.Empty(t, w.Header().Get("Accept-Ranges"))
	mockEngine.AssertExpectations(t)
}

func TestDownloadHandler_CacheHeaders(t *testing.T) {
	t.Run("caching enabled", func(t *testing.T) {
		handler, mockEngine, mockMetrics := createDownloadHandler()
		handler.config.EnableCaching = true

		fileContent := "Cacheable content"
		reference := &models.AttachmentReference{
			ID:       "ref-cache",
			Filename: "test.txt",
		}
		file := &models.AttachmentFile{
			Hash:      "hash-cache",
			SizeBytes: int64(len(fileContent)),
			MimeType:  "text/plain",
		}

		mockEngine.On("DownloadFile", mock.Anything, "ref-cache").Return(
			newMockReadCloser(fileContent),
			reference,
			file,
			nil,
		)

		mockMetrics.On("RecordDownload", "success", int64(len(fileContent)), mock.Anything, false).Return()

		req, _ := http.NewRequest("GET", "/download/ref-cache", nil)
		w := httptest.NewRecorder()

		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = req
		ctx.Params = gin.Params{
			{Key: "reference_id", Value: "ref-cache"},
		}

		handler.Handle(ctx)

		assert.Equal(t, "public, max-age=3600", w.Header().Get("Cache-Control"))
		mockEngine.AssertExpectations(t)
		mockMetrics.AssertExpectations(t)
	})

	t.Run("caching disabled", func(t *testing.T) {
		handler, mockEngine, mockMetrics := createDownloadHandler()
		handler.config.EnableCaching = false

		fileContent := "Non-cacheable content"
		reference := &models.AttachmentReference{
			ID:       "ref-no-cache",
			Filename: "test.txt",
		}
		file := &models.AttachmentFile{
			Hash:      "hash-no-cache",
			SizeBytes: int64(len(fileContent)),
			MimeType:  "text/plain",
		}

		mockEngine.On("DownloadFile", mock.Anything, "ref-no-cache").Return(
			newMockReadCloser(fileContent),
			reference,
			file,
			nil,
		)

		mockMetrics.On("RecordDownload", "success", int64(len(fileContent)), mock.Anything, false).Return()

		req, _ := http.NewRequest("GET", "/download/ref-no-cache", nil)
		w := httptest.NewRecorder()

		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = req
		ctx.Params = gin.Params{
			{Key: "reference_id", Value: "ref-no-cache"},
		}

		handler.Handle(ctx)

		assert.Equal(t, "no-cache, no-store, must-revalidate", w.Header().Get("Cache-Control"))
		mockEngine.AssertExpectations(t)
		mockMetrics.AssertExpectations(t)
	})
}

func TestDownloadHandler_SecurityHeaders(t *testing.T) {
	handler, mockEngine, _ := createDownloadHandler()
	handler.metrics = nil // Disable metrics for this test

	fileContent := "Content"
	reference := &models.AttachmentReference{
		ID:       "ref-security",
		Filename: "test.txt",
	}
	file := &models.AttachmentFile{
		Hash:      "hash-security",
		SizeBytes: int64(len(fileContent)),
		MimeType:  "text/plain",
	}

	mockEngine.On("DownloadFile", mock.Anything, "ref-security").Return(
		newMockReadCloser(fileContent),
		reference,
		file,
		nil,
	)

	req, _ := http.NewRequest("GET", "/download/ref-security", nil)
	w := httptest.NewRecorder()

	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Params = gin.Params{
		{Key: "reference_id", Value: "ref-security"},
	}

	handler.Handle(ctx)

	// Assert security headers are set
	assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
	mockEngine.AssertExpectations(t)
}

func TestDefaultDownloadConfig(t *testing.T) {
	config := DefaultDownloadConfig()

	assert.NotNil(t, config)
	assert.True(t, config.EnableRangeRequests)
	assert.True(t, config.EnableCaching)
	assert.Equal(t, 3600, config.CacheMaxAge)
	assert.Equal(t, 32*1024, config.BufferSize)
}

func BenchmarkDownloadHandler_Handle(b *testing.B) {
	handler, mockEngine, _ := createDownloadHandler()
	handler.metrics = nil // Disable metrics for this test

	fileContent := strings.Repeat("A", 1024*1024) // 1 MB
	reference := &models.AttachmentReference{
		ID:       "ref-bench",
		Filename: "test.dat",
	}
	file := &models.AttachmentFile{
		Hash:      "hash-bench",
		SizeBytes: int64(len(fileContent)),
		MimeType:  "application/octet-stream",
	}

	mockEngine.On("DownloadFile", mock.Anything, mock.Anything).Return(
		newMockReadCloser(fileContent),
		reference,
		file,
		nil,
	)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("GET", "/download/ref-bench", nil)
		w := httptest.NewRecorder()

		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = req
		ctx.Params = gin.Params{
			{Key: "reference_id", Value: "ref-bench"},
		}

		handler.Handle(ctx)
	}
}
