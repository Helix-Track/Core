package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/helixtrack/attachments-service/internal/database"
	"github.com/helixtrack/attachments-service/internal/models"
	"github.com/helixtrack/attachments-service/internal/storage/deduplication"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// Note: MockDatabase is defined in admin_test.go and shared across all handler tests
// Note: MockDeduplicationEngine methods are defined in upload_test.go

// Helper function to create metadata handler
func createMetadataHandler() (*MetadataHandler, *MockDatabase, *MockDeduplicationEngine) {
	mockDB := &MockDatabase{}
	mockEngine := &MockDeduplicationEngine{}
	logger := zap.NewNop()

	// Pass mockEngine so tests can set expectations on it
	handler := NewMetadataHandler(mockDB, mockEngine, nil, logger)

	return handler, mockDB, mockEngine
}

// Tests

func TestMetadataHandler_ListByEntity_Success(t *testing.T) {
	handler, mockDB, _ := createMetadataHandler()

	references := []*models.AttachmentReference{
		{
			ID:         "ref-1",
			FileHash:   "hash1",
			Filename:   "file1.pdf",
			EntityType: "ticket",
			EntityID:   "TICKET-123",
			UploaderID: "user1",
			Tags:       []string{"important"},
			Created:    time.Now().Unix(),
		},
		{
			ID:         "ref-2",
			FileHash:   "hash2",
			Filename:   "file2.pdf",
			EntityType: "ticket",
			EntityID:   "TICKET-123",
			UploaderID: "user2",
			Created:    time.Now().Unix(),
		},
	}

	file1 := &models.AttachmentFile{
		Hash:      "hash1",
		SizeBytes: 1024,
		MimeType:  "application/pdf",
	}

	file2 := &models.AttachmentFile{
		Hash:      "hash2",
		SizeBytes: 2048,
		MimeType:  "application/pdf",
	}

	mockDB.On("ListReferencesByEntity", mock.Anything, "ticket", "TICKET-123").Return(references, nil)
	mockDB.On("GetFile", mock.Anything, "hash1").Return(file1, nil)
	mockDB.On("GetFile", mock.Anything, "hash2").Return(file2, nil)

	// Create request
	req, _ := http.NewRequest("GET", "/attachments/ticket/TICKET-123", nil)
	w := httptest.NewRecorder()

	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Params = gin.Params{
		{Key: "entity_type", Value: "ticket"},
		{Key: "entity_id", Value: "TICKET-123"},
	}

	// Execute
	handler.ListByEntity(ctx)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "ref-1")
	assert.Contains(t, w.Body.String(), "ref-2")
	assert.Contains(t, w.Body.String(), "file1.pdf")
	assert.Contains(t, w.Body.String(), "file2.pdf")
	assert.Contains(t, w.Body.String(), `"total_count":2`)
	mockDB.AssertExpectations(t)
}

func TestMetadataHandler_ListByEntity_MissingParams(t *testing.T) {
	handler, _, _ := createMetadataHandler()

	// Create request without entity_id
	req, _ := http.NewRequest("GET", "/attachments/ticket/", nil)
	w := httptest.NewRecorder()

	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Params = gin.Params{
		{Key: "entity_type", Value: "ticket"},
	}

	// Execute
	handler.ListByEntity(ctx)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "entity_type and entity_id are required")
}

func TestMetadataHandler_ListByEntity_DatabaseError(t *testing.T) {
	handler, mockDB, _ := createMetadataHandler()

	mockDB.On("ListReferencesByEntity", mock.Anything, "ticket", "TICKET-123").Return(
		nil,
		errors.New("database error"),
	)

	// Create request
	req, _ := http.NewRequest("GET", "/attachments/ticket/TICKET-123", nil)
	w := httptest.NewRecorder()

	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Params = gin.Params{
		{Key: "entity_type", Value: "ticket"},
		{Key: "entity_id", Value: "TICKET-123"},
	}

	// Execute
	handler.ListByEntity(ctx)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "failed to list attachments")
	mockDB.AssertExpectations(t)
}

func TestMetadataHandler_Delete_Success(t *testing.T) {
	handler, _, mockEngine := createMetadataHandler()

	mockEngine.On("DeleteReference", mock.Anything, "ref-123").Return(nil)

	// Create request
	req, _ := http.NewRequest("DELETE", "/attachments/ref-123", nil)
	w := httptest.NewRecorder()

	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Params = gin.Params{
		{Key: "reference_id", Value: "ref-123"},
	}
	ctx.Set("user_id", "user123")

	// Execute
	handler.Delete(ctx)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "attachment deleted successfully")
	assert.Contains(t, w.Body.String(), "ref-123")
	mockEngine.AssertExpectations(t)
}

func TestMetadataHandler_Delete_MissingAuth(t *testing.T) {
	handler, _, _ := createMetadataHandler()

	// Create request without user_id
	req, _ := http.NewRequest("DELETE", "/attachments/ref-123", nil)
	w := httptest.NewRecorder()

	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Params = gin.Params{
		{Key: "reference_id", Value: "ref-123"},
	}

	// Execute
	handler.Delete(ctx)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "authentication required")
}

func TestMetadataHandler_Delete_MissingReferenceID(t *testing.T) {
	handler, _, _ := createMetadataHandler()

	// Create request without reference_id
	req, _ := http.NewRequest("DELETE", "/attachments/", nil)
	w := httptest.NewRecorder()

	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Set("user_id", "user123")

	// Execute
	handler.Delete(ctx)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "reference_id is required")
}

func TestMetadataHandler_Delete_Failed(t *testing.T) {
	handler, _, mockEngine := createMetadataHandler()

	mockEngine.On("DeleteReference", mock.Anything, "ref-123").Return(errors.New("deletion failed"))

	// Create request
	req, _ := http.NewRequest("DELETE", "/attachments/ref-123", nil)
	w := httptest.NewRecorder()

	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Params = gin.Params{
		{Key: "reference_id", Value: "ref-123"},
	}
	ctx.Set("user_id", "user123")

	// Execute
	handler.Delete(ctx)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "failed to delete attachment")
	mockEngine.AssertExpectations(t)
}

func TestMetadataHandler_Update_Success(t *testing.T) {
	handler, mockDB, _ := createMetadataHandler()

	description := "Updated description"
	reference := &models.AttachmentReference{
		ID:          "ref-123",
		FileHash:    "hash123",
		Filename:    "test.pdf",
		Description: &description,
		Tags:        []string{"tag1", "tag2"},
	}

	mockDB.On("GetReference", mock.Anything, "ref-123").Return(reference, nil)
	mockDB.On("UpdateReference", mock.Anything, mock.Anything).Return(nil)

	// Create request body
	reqBody := map[string]interface{}{
		"description": "Updated description",
		"tags":        []string{"tag1", "tag2"},
	}
	body, _ := json.Marshal(reqBody)

	// Create request
	req, _ := http.NewRequest("PATCH", "/attachments/ref-123", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Params = gin.Params{
		{Key: "reference_id", Value: "ref-123"},
	}

	// Execute
	handler.Update(ctx)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "attachment updated successfully")
	mockDB.AssertExpectations(t)
}

func TestMetadataHandler_Update_InvalidJSON(t *testing.T) {
	handler, _, _ := createMetadataHandler()

	// Create request with invalid JSON
	req, _ := http.NewRequest("PATCH", "/attachments/ref-123", bytes.NewReader([]byte("{invalid")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Params = gin.Params{
		{Key: "reference_id", Value: "ref-123"},
	}

	// Execute
	handler.Update(ctx)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid request body")
}

func TestMetadataHandler_Update_ReferenceNotFound(t *testing.T) {
	handler, mockDB, _ := createMetadataHandler()

	mockDB.On("GetReference", mock.Anything, "nonexistent").Return(nil, errors.New("not found"))

	// Create request body
	reqBody := map[string]interface{}{
		"description": "Updated",
	}
	body, _ := json.Marshal(reqBody)

	// Create request
	req, _ := http.NewRequest("PATCH", "/attachments/nonexistent", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Params = gin.Params{
		{Key: "reference_id", Value: "nonexistent"},
	}

	// Execute
	handler.Update(ctx)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "reference not found")
	mockDB.AssertExpectations(t)
}

func TestMetadataHandler_GetStats_Success(t *testing.T) {
	handler, _, mockEngine := createMetadataHandler()

	stats := &deduplication.DeduplicationStats{
		TotalFiles:        100,
		TotalReferences:   250,
		UniqueFiles:       100,
		SharedFiles:       30,
		DeduplicationRate: 0.60,
		SavedFiles:        150,
	}

	mockEngine.On("GetDeduplicationStats", mock.Anything).Return(stats, nil)

	// Create request
	req, _ := http.NewRequest("GET", "/stats", nil)
	w := httptest.NewRecorder()

	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req

	// Execute
	handler.GetStats(ctx)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"total_files":100`)
	assert.Contains(t, w.Body.String(), `"total_references":250`)
	assert.Contains(t, w.Body.String(), `"deduplication_rate":0.6`)
	mockEngine.AssertExpectations(t)
}

func TestMetadataHandler_GetStats_Failed(t *testing.T) {
	handler, _, mockEngine := createMetadataHandler()

	mockEngine.On("GetDeduplicationStats", mock.Anything).Return(nil, errors.New("stats error"))

	// Create request
	req, _ := http.NewRequest("GET", "/stats", nil)
	w := httptest.NewRecorder()

	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req

	// Execute
	handler.GetStats(ctx)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "failed to get statistics")
	mockEngine.AssertExpectations(t)
}

func TestMetadataHandler_Search_Success(t *testing.T) {
	handler, mockDB, _ := createMetadataHandler()

	references := []*models.AttachmentReference{
		{
			ID:         "ref-1",
			FileHash:   "hash1",
			Filename:   "test.pdf",
			EntityType: "ticket",
			EntityID:   "TICKET-123",
			UploaderID: "user1",
			Tags:       []string{"important"},
			Created:    time.Now().Unix(),
		},
	}

	file := &models.AttachmentFile{
		Hash:      "hash1",
		SizeBytes: 1024,
		MimeType:  "application/pdf",
	}

	mockDB.On("ListReferences", mock.Anything, mock.MatchedBy(func(f *database.ReferenceFilter) bool {
		return f.Limit == 50 && f.Offset == 0
	})).Return(references, int64(1), nil)

	mockDB.On("GetFile", mock.Anything, "hash1").Return(file, nil)

	// Create request
	req, _ := http.NewRequest("GET", "/search?filename=test.pdf&limit=50&offset=0", nil)
	w := httptest.NewRecorder()

	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req

	// Execute
	handler.Search(ctx)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "ref-1")
	assert.Contains(t, w.Body.String(), "test.pdf")
	assert.Contains(t, w.Body.String(), `"total_count":1`)
	mockDB.AssertExpectations(t)
}

func TestMetadataHandler_Search_WithMimeFilter(t *testing.T) {
	handler, mockDB, _ := createMetadataHandler()

	references := []*models.AttachmentReference{
		{
			ID:       "ref-1",
			FileHash: "hash1",
			Filename: "test.pdf",
		},
		{
			ID:       "ref-2",
			FileHash: "hash2",
			Filename: "test.png",
		},
	}

	pdfFile := &models.AttachmentFile{
		Hash:      "hash1",
		SizeBytes: 1024,
		MimeType:  "application/pdf",
	}

	pngFile := &models.AttachmentFile{
		Hash:      "hash2",
		SizeBytes: 2048,
		MimeType:  "image/png",
	}

	mockDB.On("ListReferences", mock.Anything, mock.Anything).Return(references, int64(2), nil)
	mockDB.On("GetFile", mock.Anything, "hash1").Return(pdfFile, nil)
	mockDB.On("GetFile", mock.Anything, "hash2").Return(pngFile, nil)

	// Create request with MIME type filter
	req, _ := http.NewRequest("GET", "/search?mime_type=application/pdf", nil)
	w := httptest.NewRecorder()

	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req

	// Execute
	handler.Search(ctx)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	// Should only contain PDF, not PNG
	assert.Contains(t, w.Body.String(), "ref-1")
	assert.NotContains(t, w.Body.String(), "ref-2")
	mockDB.AssertExpectations(t)
}

func TestMetadataHandler_Search_Failed(t *testing.T) {
	handler, mockDB, _ := createMetadataHandler()

	mockDB.On("ListReferences", mock.Anything, mock.Anything).Return(nil, int64(0), errors.New("search error"))

	// Create request
	req, _ := http.NewRequest("GET", "/search", nil)
	w := httptest.NewRecorder()

	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req

	// Execute
	handler.Search(ctx)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "search failed")
	mockDB.AssertExpectations(t)
}

func TestMetadataHandler_GetByHash_Success(t *testing.T) {
	handler, mockDB, _ := createMetadataHandler()

	file := &models.AttachmentFile{
		Hash:      "abc123",
		SizeBytes: 1024,
		MimeType:  "application/pdf",
		RefCount:  3,
	}

	references := []*models.AttachmentReference{
		{
			ID:         "ref-1",
			Filename:   "doc1.pdf",
			EntityType: "ticket",
			EntityID:   "TICKET-1",
			UploaderID: "user1",
			Created:    time.Now().Unix(),
		},
		{
			ID:         "ref-2",
			Filename:   "doc2.pdf",
			EntityType: "project",
			EntityID:   "PROJECT-1",
			UploaderID: "user2",
			Created:    time.Now().Unix(),
		},
	}

	mockDB.On("GetFile", mock.Anything, "abc123").Return(file, nil)
	mockDB.On("ListReferencesByHash", mock.Anything, "abc123").Return(references, nil)

	// Create request
	req, _ := http.NewRequest("GET", "/files/abc123", nil)
	w := httptest.NewRecorder()

	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Params = gin.Params{
		{Key: "file_hash", Value: "abc123"},
	}

	// Execute
	handler.GetByHash(ctx)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "abc123")
	assert.Contains(t, w.Body.String(), `"ref_count":3`)
	assert.Contains(t, w.Body.String(), "ref-1")
	assert.Contains(t, w.Body.String(), "ref-2")
	mockDB.AssertExpectations(t)
}

func TestMetadataHandler_GetByHash_FileNotFound(t *testing.T) {
	handler, mockDB, _ := createMetadataHandler()

	mockDB.On("GetFile", mock.Anything, "nonexistent").Return(nil, errors.New("not found"))

	// Create request
	req, _ := http.NewRequest("GET", "/files/nonexistent", nil)
	w := httptest.NewRecorder()

	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Params = gin.Params{
		{Key: "file_hash", Value: "nonexistent"},
	}

	// Execute
	handler.GetByHash(ctx)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "file not found")
	mockDB.AssertExpectations(t)
}

func TestMetadataHandler_GetByHash_MissingHash(t *testing.T) {
	handler, _, _ := createMetadataHandler()

	// Create request without file_hash
	req, _ := http.NewRequest("GET", "/files/", nil)
	w := httptest.NewRecorder()

	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req

	// Execute
	handler.GetByHash(ctx)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "file_hash is required")
}

func TestNewMetadataHandler(t *testing.T) {
	mockDB := &MockDatabase{}
	mockEngine := &MockDeduplicationEngine{}
	mockMetrics := &MockPrometheusMetrics{}
	logger := zap.NewNop()

	handler := NewMetadataHandler(mockDB, mockEngine, mockMetrics, logger)

	assert.NotNil(t, handler)
	assert.Equal(t, mockDB, handler.db)
	assert.Equal(t, mockEngine, handler.deduplicationEngine)
	assert.Equal(t, mockMetrics, handler.metrics)
	assert.Equal(t, logger, handler.logger)
}
