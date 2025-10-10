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
	"helixtrack.ru/core/internal/middleware"
	"helixtrack.ru/core/internal/models"
)

// setupExtensionTestHandler creates test handler with extension tables
func setupExtensionTestHandler(t *testing.T) *Handler {
	handler := setupTestHandler(t)

	// Create extension table
	_, err := handler.db.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS extension (
			id TEXT PRIMARY KEY,
			title TEXT NOT NULL,
			description TEXT,
			version TEXT,
			enabled INTEGER DEFAULT 0,
			created INTEGER NOT NULL,
			modified INTEGER NOT NULL,
			deleted INTEGER DEFAULT 0
		)
	`)
	require.NoError(t, err)

	// Create extension_metadata table
	_, err = handler.db.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS extension_metadata (
			id TEXT PRIMARY KEY,
			extension_id TEXT NOT NULL,
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

// TestExtensionHandler_Create_Success tests successful extension creation
func TestExtensionHandler_Create_Success(t *testing.T) {
	handler := setupExtensionTestHandler(t)
	gin.SetMode(gin.TestMode)

	reqBody := models.Request{
		Action: models.ActionExtensionCreate,
		Data: map[string]interface{}{
			"title":       "Test Extension",
			"description": "Test extension description",
			"version":     "1.0.0",
			"enabled":     true,
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set(middleware.UsernameKey, "testuser")

	handler.handleExtensionCreate(c, &reqBody)

	assert.Equal(t, http.StatusCreated, w.Code)

	// Verify database insertion
	var count int
	err := handler.db.QueryRow(context.Background(), "SELECT COUNT(*) FROM extension WHERE title = ?", "Test Extension").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

// TestExtensionHandler_Create_MissingTitle tests creation with missing title
func TestExtensionHandler_Create_MissingTitle(t *testing.T) {
	handler := setupExtensionTestHandler(t)
	gin.SetMode(gin.TestMode)

	reqBody := models.Request{
		Action: models.ActionExtensionCreate,
		Data: map[string]interface{}{
			"description": "No title",
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set(middleware.UsernameKey, "testuser")

	handler.handleExtensionCreate(c, &reqBody)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestExtensionHandler_Read_Success tests reading an extension
func TestExtensionHandler_Read_Success(t *testing.T) {
	handler := setupExtensionTestHandler(t)
	gin.SetMode(gin.TestMode)

	// Insert test extension
	extID := "test-ext-id"
	now := time.Now().Unix()
	_, err := handler.db.Exec(context.Background(), `
		INSERT INTO extension (id, title, description, version, enabled, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, extID, "Test Extension", "Description", "1.0.0", 1, now, now, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionExtensionRead,
		Data: map[string]interface{}{
			"id": extID,
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set(middleware.UsernameKey, "testuser")

	handler.handleExtensionRead(c, &reqBody)

	assert.Equal(t, http.StatusOK, w.Code)
}

// TestExtensionHandler_Read_NotFound tests reading non-existent extension
func TestExtensionHandler_Read_NotFound(t *testing.T) {
	handler := setupExtensionTestHandler(t)
	gin.SetMode(gin.TestMode)

	reqBody := models.Request{
		Action: models.ActionExtensionRead,
		Data: map[string]interface{}{
			"id": "non-existent-id",
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set(middleware.UsernameKey, "testuser")

	handler.handleExtensionRead(c, &reqBody)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// TestExtensionHandler_List_EmptyList tests listing when no extensions exist
func TestExtensionHandler_List_EmptyList(t *testing.T) {
	handler := setupExtensionTestHandler(t)
	gin.SetMode(gin.TestMode)

	reqBody := models.Request{
		Action: models.ActionExtensionList,
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set(middleware.UsernameKey, "testuser")

	handler.handleExtensionList(c, &reqBody)

	assert.Equal(t, http.StatusOK, w.Code)
}

// TestExtensionHandler_List_MultipleExtensions tests listing multiple extensions
func TestExtensionHandler_List_MultipleExtensions(t *testing.T) {
	handler := setupExtensionTestHandler(t)
	gin.SetMode(gin.TestMode)

	// Insert test extensions
	now := time.Now().Unix()
	extensions := []string{"Extension A", "Extension B", "Extension C"}
	for _, title := range extensions {
		_, err := handler.db.Exec(context.Background(), `
			INSERT INTO extension (id, title, version, enabled, created, modified, deleted)
			VALUES (?, ?, ?, ?, ?, ?, ?)
		`, generateTestID(), title, "1.0.0", 1, now, now, 0)
		require.NoError(t, err)
	}

	reqBody := models.Request{
		Action: models.ActionExtensionList,
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set(middleware.UsernameKey, "testuser")

	handler.handleExtensionList(c, &reqBody)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	responseData := response.Data.(map[string]interface{})
	count := int(responseData["count"].(float64))
	assert.Equal(t, 3, count)
}

// TestExtensionHandler_Modify_Success tests modifying an extension
func TestExtensionHandler_Modify_Success(t *testing.T) {
	handler := setupExtensionTestHandler(t)
	gin.SetMode(gin.TestMode)

	// Insert test extension
	extID := "test-ext-id"
	now := time.Now().Unix()
	_, err := handler.db.Exec(context.Background(), `
		INSERT INTO extension (id, title, version, enabled, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, extID, "Original Title", "1.0.0", 1, now, now, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionExtensionModify,
		Data: map[string]interface{}{
			"id":          extID,
			"title":       "Updated Title",
			"description": "New description",
			"version":     "2.0.0",
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set(middleware.UsernameKey, "testuser")

	handler.handleExtensionModify(c, &reqBody)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify update
	var title, version string
	err = handler.db.QueryRow(context.Background(), "SELECT title, version FROM extension WHERE id = ?", extID).Scan(&title, &version)
	require.NoError(t, err)
	assert.Equal(t, "Updated Title", title)
	assert.Equal(t, "2.0.0", version)
}

// TestExtensionHandler_Modify_NotFound tests modifying non-existent extension
func TestExtensionHandler_Modify_NotFound(t *testing.T) {
	handler := setupExtensionTestHandler(t)
	gin.SetMode(gin.TestMode)

	reqBody := models.Request{
		Action: models.ActionExtensionModify,
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
	c.Set(middleware.UsernameKey, "testuser")

	handler.handleExtensionModify(c, &reqBody)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// TestExtensionHandler_Remove_Success tests removing an extension
func TestExtensionHandler_Remove_Success(t *testing.T) {
	handler := setupExtensionTestHandler(t)
	gin.SetMode(gin.TestMode)

	// Insert test extension
	extID := "test-ext-id"
	now := time.Now().Unix()
	_, err := handler.db.Exec(context.Background(), `
		INSERT INTO extension (id, title, version, enabled, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, extID, "Test Extension", "1.0.0", 1, now, now, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionExtensionRemove,
		Data: map[string]interface{}{
			"id": extID,
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set(middleware.UsernameKey, "testuser")

	handler.handleExtensionRemove(c, &reqBody)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify soft delete
	var deleted int
	err = handler.db.QueryRow(context.Background(), "SELECT deleted FROM extension WHERE id = ?", extID).Scan(&deleted)
	require.NoError(t, err)
	assert.Equal(t, 1, deleted)
}

// TestExtensionHandler_Enable_Success tests enabling an extension
func TestExtensionHandler_Enable_Success(t *testing.T) {
	handler := setupExtensionTestHandler(t)
	gin.SetMode(gin.TestMode)

	// Insert test extension (disabled)
	extID := "test-ext-id"
	now := time.Now().Unix()
	_, err := handler.db.Exec(context.Background(), `
		INSERT INTO extension (id, title, version, enabled, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, extID, "Test Extension", "1.0.0", 0, now, now, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionExtensionEnable,
		Data: map[string]interface{}{
			"id": extID,
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set(middleware.UsernameKey, "testuser")

	handler.handleExtensionEnable(c, &reqBody)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify enabled
	var enabled int
	err = handler.db.QueryRow(context.Background(), "SELECT enabled FROM extension WHERE id = ?", extID).Scan(&enabled)
	require.NoError(t, err)
	assert.Equal(t, 1, enabled)
}

// TestExtensionHandler_Enable_NotFound tests enabling non-existent extension
func TestExtensionHandler_Enable_NotFound(t *testing.T) {
	handler := setupExtensionTestHandler(t)
	gin.SetMode(gin.TestMode)

	reqBody := models.Request{
		Action: models.ActionExtensionEnable,
		Data: map[string]interface{}{
			"id": "non-existent-id",
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set(middleware.UsernameKey, "testuser")

	handler.handleExtensionEnable(c, &reqBody)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// TestExtensionHandler_Disable_Success tests disabling an extension
func TestExtensionHandler_Disable_Success(t *testing.T) {
	handler := setupExtensionTestHandler(t)
	gin.SetMode(gin.TestMode)

	// Insert test extension (enabled)
	extID := "test-ext-id"
	now := time.Now().Unix()
	_, err := handler.db.Exec(context.Background(), `
		INSERT INTO extension (id, title, version, enabled, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, extID, "Test Extension", "1.0.0", 1, now, now, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionExtensionDisable,
		Data: map[string]interface{}{
			"id": extID,
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set(middleware.UsernameKey, "testuser")

	handler.handleExtensionDisable(c, &reqBody)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify disabled
	var enabled int
	err = handler.db.QueryRow(context.Background(), "SELECT enabled FROM extension WHERE id = ?", extID).Scan(&enabled)
	require.NoError(t, err)
	assert.Equal(t, 0, enabled)
}

// TestExtensionHandler_SetMetadata_Success tests setting extension metadata
func TestExtensionHandler_SetMetadata_Success(t *testing.T) {
	handler := setupExtensionTestHandler(t)
	gin.SetMode(gin.TestMode)

	// Insert test extension
	extID := "test-ext-id"
	now := time.Now().Unix()
	_, err := handler.db.Exec(context.Background(), `
		INSERT INTO extension (id, title, version, enabled, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, extID, "Test Extension", "1.0.0", 1, now, now, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionExtensionSetMetadata,
		Data: map[string]interface{}{
			"extensionId": extID,
			"property":    "config_url",
			"value":       "https://example.com/config",
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set(middleware.UsernameKey, "testuser")

	handler.handleExtensionSetMetadata(c, &reqBody)

	assert.Equal(t, http.StatusCreated, w.Code)

	// Verify metadata insertion
	var count int
	err = handler.db.QueryRow(context.Background(), "SELECT COUNT(*) FROM extension_metadata WHERE extension_id = ? AND property = ?", extID, "config_url").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

// TestExtensionHandler_SetMetadata_MissingExtensionID tests setting metadata with missing extension ID
func TestExtensionHandler_SetMetadata_MissingExtensionID(t *testing.T) {
	handler := setupExtensionTestHandler(t)
	gin.SetMode(gin.TestMode)

	reqBody := models.Request{
		Action: models.ActionExtensionSetMetadata,
		Data: map[string]interface{}{
			"property": "config_url",
			"value":    "https://example.com/config",
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set(middleware.UsernameKey, "testuser")

	handler.handleExtensionSetMetadata(c, &reqBody)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestExtensionHandler_SetMetadata_MissingProperty tests setting metadata with missing property
func TestExtensionHandler_SetMetadata_MissingProperty(t *testing.T) {
	handler := setupExtensionTestHandler(t)
	gin.SetMode(gin.TestMode)

	reqBody := models.Request{
		Action: models.ActionExtensionSetMetadata,
		Data: map[string]interface{}{
			"extensionId": "test-ext-id",
			"value":       "https://example.com/config",
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set(middleware.UsernameKey, "testuser")

	handler.handleExtensionSetMetadata(c, &reqBody)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
