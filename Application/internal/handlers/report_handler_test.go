package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"helixtrack.ru/core/internal/models"
)

// setupReportTestHandler creates test handler with report tables
func setupReportTestHandler(t *testing.T) *Handler {
	handler := setupTestHandler(t)

	// Create report table
	_, err := handler.db.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS report (
			id TEXT PRIMARY KEY,
			title TEXT NOT NULL,
			description TEXT,
			query TEXT,
			created INTEGER NOT NULL,
			modified INTEGER NOT NULL,
			deleted INTEGER DEFAULT 0
		)
	`)
	require.NoError(t, err)

	// Create report_metadata table
	_, err = handler.db.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS report_metadata (
			id TEXT PRIMARY KEY,
			report_id TEXT NOT NULL,
			property TEXT NOT NULL,
			value TEXT,
			created INTEGER NOT NULL,
			modified INTEGER NOT NULL,
			deleted INTEGER DEFAULT 0
		)
	`)
	require.NoError(t, err)

	return handler
}

// TestReportHandler_Create_Success tests successful report creation
func TestReportHandler_Create_Success(t *testing.T) {
	handler := setupReportTestHandler(t)
	gin.SetMode(gin.TestMode)

	reqBody := models.Request{
		Action: models.ActionReportCreate,
		Data: map[string]interface{}{
			"title":       "Test Report",
			"description": "Test report description",
			"query":       map[string]interface{}{"status": "open"},
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")

	c.Set("request", &reqBody)
	handler.handleReportCreate(c, &reqBody)

	assert.Equal(t, http.StatusCreated, w.Code)

	// Verify database insertion
	var count int
	err := handler.db.QueryRow(context.Background(), "SELECT COUNT(*) FROM report WHERE title = ?", "Test Report").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

// TestReportHandler_Create_MissingTitle tests creation with missing title
func TestReportHandler_Create_MissingTitle(t *testing.T) {
	handler := setupReportTestHandler(t)
	gin.SetMode(gin.TestMode)

	reqBody := models.Request{
		Action: models.ActionReportCreate,
		Data: map[string]interface{}{
			"description": "No title",
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")

	c.Set("request", &reqBody)
	handler.handleReportCreate(c, &reqBody)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestReportHandler_Read_Success tests reading a report
func TestReportHandler_Read_Success(t *testing.T) {
	handler := setupReportTestHandler(t)
	gin.SetMode(gin.TestMode)

	// Insert test report
	reportID := "test-report-id"
	now := time.Now().Unix()
	_, err := handler.db.Exec(context.Background(), `
		INSERT INTO report (id, title, description, query, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, reportID, "Test Report", "Description", `{"status":"open"}`, now, now, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionReportRead,
		Data: map[string]interface{}{
			"id": reportID,
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")

	c.Set("request", &reqBody)
	handler.handleReportRead(c, &reqBody)

	assert.Equal(t, http.StatusOK, w.Code)
}

// TestReportHandler_Read_NotFound tests reading non-existent report
func TestReportHandler_Read_NotFound(t *testing.T) {
	handler := setupReportTestHandler(t)
	gin.SetMode(gin.TestMode)

	reqBody := models.Request{
		Action: models.ActionReportRead,
		Data: map[string]interface{}{
			"id": "non-existent-id",
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")

	c.Set("request", &reqBody)
	handler.handleReportRead(c, &reqBody)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// TestReportHandler_List_EmptyList tests listing when no reports exist
func TestReportHandler_List_EmptyList(t *testing.T) {
	handler := setupReportTestHandler(t)
	gin.SetMode(gin.TestMode)

	reqBody := models.Request{
		Action: models.ActionReportList,
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")

	c.Set("request", &reqBody)
	handler.handleReportList(c, &reqBody)

	assert.Equal(t, http.StatusOK, w.Code)
}

// TestReportHandler_List_MultipleReports tests listing multiple reports
func TestReportHandler_List_MultipleReports(t *testing.T) {
	handler := setupReportTestHandler(t)
	gin.SetMode(gin.TestMode)

	// Insert test reports
	now := time.Now().Unix()
	reports := []string{"Report A", "Report B", "Report C"}
	for _, title := range reports {
		_, err := handler.db.Exec(context.Background(), `
			INSERT INTO report (id, title, query, created, modified, deleted)
			VALUES (?, ?, ?, ?, ?, ?)
		`, generateTestID(), title, `{}`, now, now, 0)
		require.NoError(t, err)
	}

	reqBody := models.Request{
		Action: models.ActionReportList,
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")

	c.Set("request", &reqBody)
	handler.handleReportList(c, &reqBody)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	count := int(response.Data["count"].(float64))
	assert.Equal(t, 3, count)
}

// TestReportHandler_Modify_Success tests modifying a report
func TestReportHandler_Modify_Success(t *testing.T) {
	handler := setupReportTestHandler(t)
	gin.SetMode(gin.TestMode)

	// Insert test report
	reportID := "test-report-id"
	now := time.Now().Unix()
	_, err := handler.db.Exec(context.Background(), `
		INSERT INTO report (id, title, query, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?)
	`, reportID, "Original Title", `{}`, now, now, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionReportModify,
		Data: map[string]interface{}{
			"id":          reportID,
			"title":       "Updated Title",
			"description": "New description",
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")

	c.Set("request", &reqBody)
	handler.handleReportModify(c, &reqBody)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify update
	var title string
	err = handler.db.QueryRow(context.Background(), "SELECT title FROM report WHERE id = ?", reportID).Scan(&title)
	require.NoError(t, err)
	assert.Equal(t, "Updated Title", title)
}

// TestReportHandler_Modify_NotFound tests modifying non-existent report
func TestReportHandler_Modify_NotFound(t *testing.T) {
	handler := setupReportTestHandler(t)
	gin.SetMode(gin.TestMode)

	reqBody := models.Request{
		Action: models.ActionReportModify,
		Data: map[string]interface{}{
			"id":    "non-existent-id",
			"title": "New Title",
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")

	c.Set("request", &reqBody)
	handler.handleReportModify(c, &reqBody)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// TestReportHandler_Remove_Success tests removing a report
func TestReportHandler_Remove_Success(t *testing.T) {
	handler := setupReportTestHandler(t)
	gin.SetMode(gin.TestMode)

	// Insert test report
	reportID := "test-report-id"
	now := time.Now().Unix()
	_, err := handler.db.Exec(context.Background(), `
		INSERT INTO report (id, title, query, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?)
	`, reportID, "Test Report", `{}`, now, now, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionReportRemove,
		Data: map[string]interface{}{
			"id": reportID,
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")

	c.Set("request", &reqBody)
	handler.handleReportRemove(c, &reqBody)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify soft delete
	var deleted int
	err = handler.db.QueryRow(context.Background(), "SELECT deleted FROM report WHERE id = ?", reportID).Scan(&deleted)
	require.NoError(t, err)
	assert.Equal(t, 1, deleted)
}

// TestReportHandler_Execute_Success tests executing a report
func TestReportHandler_Execute_Success(t *testing.T) {
	handler := setupReportTestHandler(t)
	gin.SetMode(gin.TestMode)

	// Insert test report
	reportID := "test-report-id"
	now := time.Now().Unix()
	_, err := handler.db.Exec(context.Background(), `
		INSERT INTO report (id, title, query, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?)
	`, reportID, "Test Report", `{"status":"open"}`, now, now, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionReportExecute,
		Data: map[string]interface{}{
			"id": reportID,
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")

	c.Set("request", &reqBody)
	handler.handleReportExecute(c, &reqBody)

	assert.Equal(t, http.StatusOK, w.Code)
}

// TestReportHandler_Execute_NotFound tests executing non-existent report
func TestReportHandler_Execute_NotFound(t *testing.T) {
	handler := setupReportTestHandler(t)
	gin.SetMode(gin.TestMode)

	reqBody := models.Request{
		Action: models.ActionReportExecute,
		Data: map[string]interface{}{
			"id": "non-existent-id",
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")

	c.Set("request", &reqBody)
	handler.handleReportExecute(c, &reqBody)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// TestReportHandler_SetMetadata_Success tests setting report metadata
func TestReportHandler_SetMetadata_Success(t *testing.T) {
	handler := setupReportTestHandler(t)
	gin.SetMode(gin.TestMode)

	// Insert test report
	reportID := "test-report-id"
	now := time.Now().Unix()
	_, err := handler.db.Exec(context.Background(), `
		INSERT INTO report (id, title, query, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?)
	`, reportID, "Test Report", `{}`, now, now, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionReportSetMetadata,
		Data: map[string]interface{}{
			"reportId": reportID,
			"property": "author",
			"value":    "Test User",
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")

	c.Set("request", &reqBody)
	handler.handleReportSetMetadata(c, &reqBody)

	assert.Equal(t, http.StatusCreated, w.Code)

	// Verify metadata insertion
	var count int
	err = handler.db.QueryRow(context.Background(), "SELECT COUNT(*) FROM report_metadata WHERE report_id = ? AND property = ?", reportID, "author").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

// TestReportHandler_GetMetadata_Success tests getting report metadata
func TestReportHandler_GetMetadata_Success(t *testing.T) {
	handler := setupReportTestHandler(t)
	gin.SetMode(gin.TestMode)

	// Insert test metadata
	reportID := "test-report-id"
	now := time.Now().Unix()
	_, err := handler.db.Exec(context.Background(), `
		INSERT INTO report_metadata (id, report_id, property, value, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, generateTestID(), reportID, "author", `"Test User"`, now, now, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionReportGetMetadata,
		Data: map[string]interface{}{
			"reportId": reportID,
			"property": "author",
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")

	c.Set("request", &reqBody)
	handler.handleReportGetMetadata(c, &reqBody)

	assert.Equal(t, http.StatusOK, w.Code)
}

// TestReportHandler_GetMetadata_NotFound tests getting non-existent metadata
func TestReportHandler_GetMetadata_NotFound(t *testing.T) {
	handler := setupReportTestHandler(t)
	gin.SetMode(gin.TestMode)

	reqBody := models.Request{
		Action: models.ActionReportGetMetadata,
		Data: map[string]interface{}{
			"reportId": "test-report-id",
			"property": "non-existent",
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")

	c.Set("request", &reqBody)
	handler.handleReportGetMetadata(c, &reqBody)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// TestReportHandler_RemoveMetadata_Success tests removing report metadata
func TestReportHandler_RemoveMetadata_Success(t *testing.T) {
	handler := setupReportTestHandler(t)
	gin.SetMode(gin.TestMode)

	// Insert test metadata
	metadataID := "test-metadata-id"
	now := time.Now().Unix()
	_, err := handler.db.Exec(context.Background(), `
		INSERT INTO report_metadata (id, report_id, property, value, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, metadataID, "test-report-id", "author", `"Test User"`, now, now, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionReportRemoveMetadata,
		Data: map[string]interface{}{
			"id": metadataID,
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")

	c.Set("request", &reqBody)
	handler.handleReportRemoveMetadata(c, &reqBody)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify soft delete
	var deleted int
	err = handler.db.QueryRow(context.Background(), "SELECT deleted FROM report_metadata WHERE id = ?", metadataID).Scan(&deleted)
	require.NoError(t, err)
	assert.Equal(t, 1, deleted)
}

// TestReportHandler_RemoveMetadata_NotFound tests removing non-existent metadata
func TestReportHandler_RemoveMetadata_NotFound(t *testing.T) {
	handler := setupReportTestHandler(t)
	gin.SetMode(gin.TestMode)

	reqBody := models.Request{
		Action: models.ActionReportRemoveMetadata,
		Data: map[string]interface{}{
			"id": "non-existent-id",
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")

	c.Set("request", &reqBody)
	handler.handleReportRemoveMetadata(c, &reqBody)

	assert.Equal(t, http.StatusNotFound, w.Code)
}
