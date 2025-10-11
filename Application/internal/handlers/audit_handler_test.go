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

// setupAuditTestHandler creates a test handler with audit tables
func setupAuditTestHandler(t *testing.T) *Handler {
	handler := setupTestHandler(t)

	// Create audit table
	_, err := handler.db.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS audit (
			id TEXT PRIMARY KEY,
			action TEXT NOT NULL,
			user_id TEXT,
			entity_id TEXT,
			entity_type TEXT,
			details TEXT,
			created INTEGER NOT NULL,
			modified INTEGER NOT NULL,
			deleted INTEGER DEFAULT 0
		)
	`)
	require.NoError(t, err)

	// Create audit_metadata table
	_, err = handler.db.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS audit_metadata (
			id TEXT PRIMARY KEY,
			audit_id TEXT NOT NULL,
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

// TestAuditHandler_Create_Success tests successful audit entry creation
func TestAuditHandler_Create_Success(t *testing.T) {
	handler := setupAuditTestHandler(t)
	gin.SetMode(gin.TestMode)

	reqBody := models.Request{
		Action: models.ActionAuditCreate,
		Data: map[string]interface{}{
			"action":     "create",
			"userId":     "user123",
			"entityId":   "ticket123",
			"entityType": "ticket",
			"details":    `{"field": "status", "old": "open", "new": "closed"}`,
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")

	handler.handleAuditCreate(c, &reqBody)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response models.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, models.ErrorCodeNoError, response.ErrorCode)

	// Verify audit entry was inserted into database
	var count int
	err = handler.db.QueryRow(context.Background(), "SELECT COUNT(*) FROM audit WHERE action = ? AND deleted = 0", "create").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

// TestAuditHandler_Create_MissingAction tests creation with missing action
func TestAuditHandler_Create_MissingAction(t *testing.T) {
	handler := setupAuditTestHandler(t)
	gin.SetMode(gin.TestMode)

	reqBody := models.Request{
		Action: models.ActionAuditCreate,
		Data: map[string]interface{}{
			"userId":     "user123",
			"entityType": "ticket",
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")

	handler.handleAuditCreate(c, &reqBody)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response models.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, models.ErrorCodeMissingData, response.ErrorCode)
}

// TestAuditHandler_Create_InvalidAction tests creation with invalid action
func TestAuditHandler_Create_InvalidAction(t *testing.T) {
	handler := setupAuditTestHandler(t)
	gin.SetMode(gin.TestMode)

	reqBody := models.Request{
		Action: models.ActionAuditCreate,
		Data: map[string]interface{}{
			"action":     "invalid_action_name",
			"entityType": "ticket",
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")

	handler.handleAuditCreate(c, &reqBody)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response models.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, models.ErrorCodeInvalidData, response.ErrorCode)
}

// TestAuditHandler_Create_Unauthorized tests creation without authentication
func TestAuditHandler_Create_Unauthorized(t *testing.T) {
	handler := setupAuditTestHandler(t)
	gin.SetMode(gin.TestMode)

	reqBody := models.Request{
		Action: models.ActionAuditCreate,
		Data: map[string]interface{}{
			"action": "create",
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	// No username set - unauthorized

	handler.handleAuditCreate(c, &reqBody)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestAuditHandler_Read_Success tests reading an audit entry
func TestAuditHandler_Read_Success(t *testing.T) {
	handler := setupAuditTestHandler(t)
	gin.SetMode(gin.TestMode)

	// Insert test audit entry
	auditID := "test-audit-id"
	now := time.Now().Unix()
	_, err := handler.db.Exec(context.Background(), `
		INSERT INTO audit (id, action, user_id, entity_id, entity_type, details, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, auditID, "create", "user123", "ticket123", "ticket", `{"test": "data"}`, now, now, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionAuditRead,
		Data: map[string]interface{}{
			"id": auditID,
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")

	handler.handleAuditRead(c, &reqBody)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Response
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, models.ErrorCodeNoError, response.ErrorCode)
}

// TestAuditHandler_Read_NotFound tests reading non-existent audit entry
func TestAuditHandler_Read_NotFound(t *testing.T) {
	handler := setupAuditTestHandler(t)
	gin.SetMode(gin.TestMode)

	reqBody := models.Request{
		Action: models.ActionAuditRead,
		Data: map[string]interface{}{
			"id": "non-existent-audit-id",
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")

	handler.handleAuditRead(c, &reqBody)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response models.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, models.ErrorCodeEntityNotFound, response.ErrorCode)
}

// TestAuditHandler_Read_MissingID tests reading without providing ID
func TestAuditHandler_Read_MissingID(t *testing.T) {
	handler := setupAuditTestHandler(t)
	gin.SetMode(gin.TestMode)

	reqBody := models.Request{
		Action: models.ActionAuditRead,
		Data:   map[string]interface{}{},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")

	handler.handleAuditRead(c, &reqBody)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestAuditHandler_List_EmptyList tests listing when no audit entries exist
func TestAuditHandler_List_EmptyList(t *testing.T) {
	handler := setupAuditTestHandler(t)
	gin.SetMode(gin.TestMode)

	reqBody := models.Request{
		Action: models.ActionAuditList,
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")

	handler.handleAuditList(c, &reqBody)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, models.ErrorCodeNoError, response.ErrorCode)
}

// TestAuditHandler_List_MultipleEntries tests listing multiple audit entries
func TestAuditHandler_List_MultipleEntries(t *testing.T) {
	handler := setupAuditTestHandler(t)
	gin.SetMode(gin.TestMode)

	// Insert test audit entries
	now := time.Now().Unix()
	for i := 0; i < 5; i++ {
		_, err := handler.db.Exec(context.Background(), `
			INSERT INTO audit (id, action, user_id, entity_type, created, modified, deleted)
			VALUES (?, ?, ?, ?, ?, ?, ?)
		`, generateTestID(), "create", "user123", "ticket", now-int64(i), now-int64(i), 0)
		require.NoError(t, err)
	}

	reqBody := models.Request{
		Action: models.ActionAuditList,
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")

	handler.handleAuditList(c, &reqBody)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, models.ErrorCodeNoError, response.ErrorCode)

	count := int(response.Data["count"].(float64))
	assert.Equal(t, 5, count)
}

// TestAuditHandler_List_ExcludesDeleted tests that deleted entries are not listed
func TestAuditHandler_List_ExcludesDeleted(t *testing.T) {
	handler := setupAuditTestHandler(t)
	gin.SetMode(gin.TestMode)

	// Insert active and deleted audit entries
	now := time.Now().Unix()
	_, err := handler.db.Exec(context.Background(), `
		INSERT INTO audit (id, action, user_id, entity_type, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, generateTestID(), "create", "user123", "ticket", now, now, 0)
	require.NoError(t, err)

	_, err = handler.db.Exec(context.Background(), `
		INSERT INTO audit (id, action, user_id, entity_type, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, generateTestID(), "delete", "user123", "ticket", now, now, 1) // deleted=1
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionAuditList,
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")

	handler.handleAuditList(c, &reqBody)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Response
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	count := int(response.Data["count"].(float64))
	assert.Equal(t, 1, count) // Only non-deleted entry
}

// TestAuditHandler_Query_ByUserID tests querying audit entries by user ID
func TestAuditHandler_Query_ByUserID(t *testing.T) {
	handler := setupAuditTestHandler(t)
	gin.SetMode(gin.TestMode)

	// Insert audit entries for different users
	now := time.Now().Unix()
	_, err := handler.db.Exec(context.Background(), `
		INSERT INTO audit (id, action, user_id, entity_type, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, generateTestID(), "create", "user123", "ticket", now, now, 0)
	require.NoError(t, err)

	_, err = handler.db.Exec(context.Background(), `
		INSERT INTO audit (id, action, user_id, entity_type, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, generateTestID(), "create", "user456", "ticket", now, now, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionAuditQuery,
		Data: map[string]interface{}{
			"userId": "user123",
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")

	handler.handleAuditQuery(c, &reqBody)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Response
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	count := int(response.Data["count"].(float64))
	assert.Equal(t, 1, count) // Only user123's entry
}

// TestAuditHandler_Query_ByAction tests querying audit entries by action
func TestAuditHandler_Query_ByAction(t *testing.T) {
	handler := setupAuditTestHandler(t)
	gin.SetMode(gin.TestMode)

	// Insert audit entries with different actions
	now := time.Now().Unix()
	_, err := handler.db.Exec(context.Background(), `
		INSERT INTO audit (id, action, user_id, entity_type, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, generateTestID(), "create", "user123", "ticket", now, now, 0)
	require.NoError(t, err)

	_, err = handler.db.Exec(context.Background(), `
		INSERT INTO audit (id, action, user_id, entity_type, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, generateTestID(), "modify", "user123", "ticket", now, now, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionAuditQuery,
		Data: map[string]interface{}{
			"action": "create",
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")

	handler.handleAuditQuery(c, &reqBody)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Response
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	count := int(response.Data["count"].(float64))
	assert.Equal(t, 1, count) // Only "create" action
}

// TestAuditHandler_Query_ByTimeRange tests querying audit entries by time range
func TestAuditHandler_Query_ByTimeRange(t *testing.T) {
	handler := setupAuditTestHandler(t)
	gin.SetMode(gin.TestMode)

	// Insert audit entries at different times
	now := time.Now().Unix()
	_, err := handler.db.Exec(context.Background(), `
		INSERT INTO audit (id, action, user_id, entity_type, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, generateTestID(), "create", "user123", "ticket", now-3600, now-3600, 0) // 1 hour ago
	require.NoError(t, err)

	_, err = handler.db.Exec(context.Background(), `
		INSERT INTO audit (id, action, user_id, entity_type, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, generateTestID(), "create", "user123", "ticket", now, now, 0) // now
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionAuditQuery,
		Data: map[string]interface{}{
			"startTime": float64(now - 1800), // 30 minutes ago
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")

	handler.handleAuditQuery(c, &reqBody)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Response
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	count := int(response.Data["count"].(float64))
	assert.Equal(t, 1, count) // Only recent entry
}

// TestAuditHandler_Query_MultipleFilters tests querying with multiple filters
func TestAuditHandler_Query_MultipleFilters(t *testing.T) {
	handler := setupAuditTestHandler(t)
	gin.SetMode(gin.TestMode)

	// Insert various audit entries
	now := time.Now().Unix()
	_, err := handler.db.Exec(context.Background(), `
		INSERT INTO audit (id, action, user_id, entity_id, entity_type, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, generateTestID(), "create", "user123", "ticket123", "ticket", now, now, 0)
	require.NoError(t, err)

	_, err = handler.db.Exec(context.Background(), `
		INSERT INTO audit (id, action, user_id, entity_id, entity_type, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, generateTestID(), "modify", "user123", "ticket456", "ticket", now, now, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionAuditQuery,
		Data: map[string]interface{}{
			"userId":     "user123",
			"action":     "create",
			"entityType": "ticket",
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")

	handler.handleAuditQuery(c, &reqBody)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Response
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	count := int(response.Data["count"].(float64))
	assert.Equal(t, 1, count) // Only entry matching all filters
}

// TestAuditHandler_AddMeta_Success tests adding metadata to audit entry
func TestAuditHandler_AddMeta_Success(t *testing.T) {
	handler := setupAuditTestHandler(t)
	gin.SetMode(gin.TestMode)

	// Insert test audit entry
	auditID := "test-audit-id"
	now := time.Now().Unix()
	_, err := handler.db.Exec(context.Background(), `
		INSERT INTO audit (id, action, user_id, entity_type, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, auditID, "create", "user123", "ticket", now, now, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionAuditAddMeta,
		Data: map[string]interface{}{
			"auditId":  auditID,
			"property": "ipAddress",
			"value":    "192.168.1.100",
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")

	handler.handleAuditAddMeta(c, &reqBody)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response models.Response
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, models.ErrorCodeNoError, response.ErrorCode)

	// Verify metadata was inserted
	var count int
	err = handler.db.QueryRow(context.Background(), "SELECT COUNT(*) FROM audit_metadata WHERE audit_id = ? AND property = ?", auditID, "ipAddress").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

// TestAuditHandler_AddMeta_MissingAuditID tests adding metadata without audit ID
func TestAuditHandler_AddMeta_MissingAuditID(t *testing.T) {
	handler := setupAuditTestHandler(t)
	gin.SetMode(gin.TestMode)

	reqBody := models.Request{
		Action: models.ActionAuditAddMeta,
		Data: map[string]interface{}{
			"property": "ipAddress",
			"value":    "192.168.1.100",
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")

	handler.handleAuditAddMeta(c, &reqBody)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestAuditHandler_AddMeta_MissingProperty tests adding metadata without property
func TestAuditHandler_AddMeta_MissingProperty(t *testing.T) {
	handler := setupAuditTestHandler(t)
	gin.SetMode(gin.TestMode)

	reqBody := models.Request{
		Action: models.ActionAuditAddMeta,
		Data: map[string]interface{}{
			"auditId": "test-audit-id",
			"value":   "192.168.1.100",
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")

	handler.handleAuditAddMeta(c, &reqBody)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
