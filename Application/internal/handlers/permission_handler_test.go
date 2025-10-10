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

// setupPermissionTestHandler creates test handler with permission tables
func setupPermissionTestHandler(t *testing.T) *Handler {
	handler := setupTestHandler(t)

	// Create permission table
	_, err := handler.db.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS permission (
			id TEXT PRIMARY KEY,
			title TEXT NOT NULL,
			description TEXT,
			value INTEGER NOT NULL,
			created INTEGER NOT NULL,
			modified INTEGER NOT NULL,
			deleted INTEGER DEFAULT 0
		)
	`)
	require.NoError(t, err)

	// Create permission_context table
	_, err = handler.db.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS permission_context (
			id TEXT PRIMARY KEY,
			context TEXT NOT NULL,
			created INTEGER NOT NULL,
			modified INTEGER NOT NULL,
			deleted INTEGER DEFAULT 0
		)
	`)
	require.NoError(t, err)

	// Create permission_user_mapping table
	_, err = handler.db.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS permission_user_mapping (
			id TEXT PRIMARY KEY,
			permission_id TEXT NOT NULL,
			user_id TEXT NOT NULL,
			permission_context_id TEXT NOT NULL,
			created INTEGER NOT NULL,
			deleted INTEGER DEFAULT 0
		)
	`)
	require.NoError(t, err)

	// Create permission_team_mapping table
	_, err = handler.db.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS permission_team_mapping (
			id TEXT PRIMARY KEY,
			permission_id TEXT NOT NULL,
			team_id TEXT NOT NULL,
			permission_context_id TEXT NOT NULL,
			created INTEGER NOT NULL,
			deleted INTEGER DEFAULT 0
		)
	`)
	require.NoError(t, err)

	return handler
}

// TestPermissionHandler_Create_Success tests creating a permission
func TestPermissionHandler_Create_Success(t *testing.T) {
	handler := setupPermissionTestHandler(t)
	gin.SetMode(gin.TestMode)

	reqBody := models.Request{
		Action: models.ActionPermissionCreate,
		Data: map[string]interface{}{
			"title":       "Read Permission",
			"description": "Can read entities",
			"value":       float64(models.PermissionRead),
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")

	handler.handlePermissionCreate(c, &reqBody)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response models.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeSuccess, response.ErrorCode)
}

// TestPermissionHandler_Create_AllPermissionValues tests creating permissions with all valid values
func TestPermissionHandler_Create_AllPermissionValues(t *testing.T) {
	handler := setupPermissionTestHandler(t)
	gin.SetMode(gin.TestMode)

	testCases := []struct {
		name  string
		title string
		value int
	}{
		{"READ", "Read Permission", models.PermissionRead},
		{"CREATE", "Create Permission", models.PermissionCreate},
		{"UPDATE", "Update Permission", models.PermissionUpdate},
		{"DELETE", "Delete Permission", models.PermissionDelete},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			reqBody := models.Request{
				Action: models.ActionPermissionCreate,
				Data: map[string]interface{}{
					"title": tc.title,
					"value": float64(tc.value),
				},
			}

			body, _ := json.Marshal(reqBody)
			req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
			w := httptest.NewRecorder()

			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.Set("username", "testuser")

			handler.handlePermissionCreate(c, &reqBody)

			assert.Equal(t, http.StatusCreated, w.Code)
		})
	}
}

// TestPermissionHandler_Create_InvalidValue tests creating with invalid permission value
func TestPermissionHandler_Create_InvalidValue(t *testing.T) {
	handler := setupPermissionTestHandler(t)
	gin.SetMode(gin.TestMode)

	reqBody := models.Request{
		Action: models.ActionPermissionCreate,
		Data: map[string]interface{}{
			"title": "Invalid Permission",
			"value": float64(99), // Invalid value
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")

	handler.handlePermissionCreate(c, &reqBody)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestPermissionHandler_Create_MissingTitle tests creating without required title
func TestPermissionHandler_Create_MissingTitle(t *testing.T) {
	handler := setupPermissionTestHandler(t)
	gin.SetMode(gin.TestMode)

	reqBody := models.Request{
		Action: models.ActionPermissionCreate,
		Data: map[string]interface{}{
			"value": float64(models.PermissionRead),
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")

	handler.handlePermissionCreate(c, &reqBody)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestPermissionHandler_Read_Success tests reading a permission
func TestPermissionHandler_Read_Success(t *testing.T) {
	handler := setupPermissionTestHandler(t)
	gin.SetMode(gin.TestMode)

	// Insert test permission
	permID := generateTestID()
	now := time.Now().Unix()
	_, err := handler.db.Exec(context.Background(),
		`INSERT INTO permission (id, title, description, value, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		permID, "Test Permission", "Description", models.PermissionRead, now, now, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionPermissionRead,
		Data: map[string]interface{}{
			"id": permID,
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")

	handler.handlePermissionRead(c, &reqBody)

	assert.Equal(t, http.StatusOK, w.Code)
}

// TestPermissionHandler_Read_NotFound tests reading non-existent permission
func TestPermissionHandler_Read_NotFound(t *testing.T) {
	handler := setupPermissionTestHandler(t)
	gin.SetMode(gin.TestMode)

	reqBody := models.Request{
		Action: models.ActionPermissionRead,
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

	handler.handlePermissionRead(c, &reqBody)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// TestPermissionHandler_List_Success tests listing permissions
func TestPermissionHandler_List_Success(t *testing.T) {
	handler := setupPermissionTestHandler(t)
	gin.SetMode(gin.TestMode)

	// Insert test permissions
	now := time.Now().Unix()
	_, err := handler.db.Exec(context.Background(),
		`INSERT INTO permission (id, title, description, value, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		generateTestID(), "Read", "Read perm", models.PermissionRead, now, now, 0)
	require.NoError(t, err)

	_, err = handler.db.Exec(context.Background(),
		`INSERT INTO permission (id, title, description, value, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		generateTestID(), "Create", "Create perm", models.PermissionCreate, now, now, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionPermissionList,
		Data:   map[string]interface{}{},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")

	handler.handlePermissionList(c, &reqBody)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Response
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	dataMap := response.Data.(map[string]interface{})
	count := int(dataMap["count"].(float64))
	assert.Equal(t, 2, count)
}

// TestPermissionHandler_Modify_Success tests modifying a permission
func TestPermissionHandler_Modify_Success(t *testing.T) {
	handler := setupPermissionTestHandler(t)
	gin.SetMode(gin.TestMode)

	// Insert test permission
	permID := generateTestID()
	now := time.Now().Unix()
	_, err := handler.db.Exec(context.Background(),
		`INSERT INTO permission (id, title, description, value, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		permID, "Old Title", "Old Desc", models.PermissionRead, now, now, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionPermissionModify,
		Data: map[string]interface{}{
			"id":          permID,
			"title":       "New Title",
			"description": "New Description",
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")

	handler.handlePermissionModify(c, &reqBody)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify changes
	var title, description string
	err = handler.db.QueryRow(context.Background(), "SELECT title, description FROM permission WHERE id = ?", permID).Scan(&title, &description)
	require.NoError(t, err)
	assert.Equal(t, "New Title", title)
	assert.Equal(t, "New Description", description)
}

// TestPermissionHandler_Modify_NotFound tests modifying non-existent permission
func TestPermissionHandler_Modify_NotFound(t *testing.T) {
	handler := setupPermissionTestHandler(t)
	gin.SetMode(gin.TestMode)

	reqBody := models.Request{
		Action: models.ActionPermissionModify,
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

	handler.handlePermissionModify(c, &reqBody)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// TestPermissionHandler_Remove_Success tests soft-deleting a permission
func TestPermissionHandler_Remove_Success(t *testing.T) {
	handler := setupPermissionTestHandler(t)
	gin.SetMode(gin.TestMode)

	// Insert test permission
	permID := generateTestID()
	now := time.Now().Unix()
	_, err := handler.db.Exec(context.Background(),
		`INSERT INTO permission (id, title, description, value, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		permID, "Test", "Desc", models.PermissionRead, now, now, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionPermissionRemove,
		Data: map[string]interface{}{
			"id": permID,
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")

	handler.handlePermissionRemove(c, &reqBody)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify soft delete
	var deleted int
	err = handler.db.QueryRow(context.Background(), "SELECT deleted FROM permission WHERE id = ?", permID).Scan(&deleted)
	require.NoError(t, err)
	assert.Equal(t, 1, deleted)
}

// TestPermissionContextHandler_Create_Success tests creating a permission context
func TestPermissionContextHandler_Create_Success(t *testing.T) {
	handler := setupPermissionTestHandler(t)
	gin.SetMode(gin.TestMode)

	reqBody := models.Request{
		Action: models.ActionPermissionContextCreate,
		Data: map[string]interface{}{
			"context": "project",
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")

	handler.handlePermissionContextCreate(c, &reqBody)

	assert.Equal(t, http.StatusCreated, w.Code)
}

// TestPermissionContextHandler_Create_AllContexts tests creating all valid contexts
func TestPermissionContextHandler_Create_AllContexts(t *testing.T) {
	handler := setupPermissionTestHandler(t)
	gin.SetMode(gin.TestMode)

	contexts := []string{"node", "account", "organization", "team", "project"}

	for _, ctx := range contexts {
		t.Run(ctx, func(t *testing.T) {
			reqBody := models.Request{
				Action: models.ActionPermissionContextCreate,
				Data: map[string]interface{}{
					"context": ctx,
				},
			}

			body, _ := json.Marshal(reqBody)
			req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
			w := httptest.NewRecorder()

			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.Set("username", "testuser")

			handler.handlePermissionContextCreate(c, &reqBody)

			assert.Equal(t, http.StatusCreated, w.Code)
		})
	}
}

// TestPermissionContextHandler_Create_InvalidContext tests creating with invalid context
func TestPermissionContextHandler_Create_InvalidContext(t *testing.T) {
	handler := setupPermissionTestHandler(t)
	gin.SetMode(gin.TestMode)

	reqBody := models.Request{
		Action: models.ActionPermissionContextCreate,
		Data: map[string]interface{}{
			"context": "invalid-context",
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")

	handler.handlePermissionContextCreate(c, &reqBody)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestPermissionContextHandler_Read_Success tests reading a permission context
func TestPermissionContextHandler_Read_Success(t *testing.T) {
	handler := setupPermissionTestHandler(t)
	gin.SetMode(gin.TestMode)

	// Insert test context
	contextID := generateTestID()
	now := time.Now().Unix()
	_, err := handler.db.Exec(context.Background(),
		`INSERT INTO permission_context (id, context, created, modified, deleted) VALUES (?, ?, ?, ?, ?)`,
		contextID, "project", now, now, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionPermissionContextRead,
		Data: map[string]interface{}{
			"id": contextID,
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")

	handler.handlePermissionContextRead(c, &reqBody)

	assert.Equal(t, http.StatusOK, w.Code)
}

// TestPermissionContextHandler_List_Success tests listing permission contexts
func TestPermissionContextHandler_List_Success(t *testing.T) {
	handler := setupPermissionTestHandler(t)
	gin.SetMode(gin.TestMode)

	// Insert test contexts
	now := time.Now().Unix()
	_, err := handler.db.Exec(context.Background(),
		`INSERT INTO permission_context (id, context, created, modified, deleted) VALUES (?, ?, ?, ?, ?)`,
		generateTestID(), "project", now, now, 0)
	require.NoError(t, err)

	_, err = handler.db.Exec(context.Background(),
		`INSERT INTO permission_context (id, context, created, modified, deleted) VALUES (?, ?, ?, ?, ?)`,
		generateTestID(), "team", now, now, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionPermissionContextList,
		Data:   map[string]interface{}{},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")

	handler.handlePermissionContextList(c, &reqBody)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Response
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	dataMap := response.Data.(map[string]interface{})
	count := int(dataMap["count"].(float64))
	assert.Equal(t, 2, count)
}

// TestPermissionContextHandler_Modify_Success tests modifying a permission context
func TestPermissionContextHandler_Modify_Success(t *testing.T) {
	handler := setupPermissionTestHandler(t)
	gin.SetMode(gin.TestMode)

	// Insert test context
	contextID := generateTestID()
	now := time.Now().Unix()
	_, err := handler.db.Exec(context.Background(),
		`INSERT INTO permission_context (id, context, created, modified, deleted) VALUES (?, ?, ?, ?, ?)`,
		contextID, "project", now, now, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionPermissionContextModify,
		Data: map[string]interface{}{
			"id":      contextID,
			"context": "team",
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")

	handler.handlePermissionContextModify(c, &reqBody)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify change
	var context string
	err = handler.db.QueryRow(context.Background(), "SELECT context FROM permission_context WHERE id = ?", contextID).Scan(&context)
	require.NoError(t, err)
	assert.Equal(t, "team", context)
}

// TestPermissionContextHandler_Remove_Success tests soft-deleting a permission context
func TestPermissionContextHandler_Remove_Success(t *testing.T) {
	handler := setupPermissionTestHandler(t)
	gin.SetMode(gin.TestMode)

	// Insert test context
	contextID := generateTestID()
	now := time.Now().Unix()
	_, err := handler.db.Exec(context.Background(),
		`INSERT INTO permission_context (id, context, created, modified, deleted) VALUES (?, ?, ?, ?, ?)`,
		contextID, "project", now, now, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionPermissionContextRemove,
		Data: map[string]interface{}{
			"id": contextID,
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")

	handler.handlePermissionContextRemove(c, &reqBody)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify soft delete
	var deleted int
	err = handler.db.QueryRow(context.Background(), "SELECT deleted FROM permission_context WHERE id = ?", contextID).Scan(&deleted)
	require.NoError(t, err)
	assert.Equal(t, 1, deleted)
}

// TestPermissionHandler_AssignUser_Success tests assigning permission to user
func TestPermissionHandler_AssignUser_Success(t *testing.T) {
	handler := setupPermissionTestHandler(t)
	gin.SetMode(gin.TestMode)

	permissionID := generateTestID()
	userID := generateTestID()
	contextID := generateTestID()

	reqBody := models.Request{
		Action: models.ActionPermissionAssignUser,
		Data: map[string]interface{}{
			"permissionId":        permissionID,
			"userId":              userID,
			"permissionContextId": contextID,
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")

	handler.handlePermissionAssignUser(c, &reqBody)

	assert.Equal(t, http.StatusCreated, w.Code)

	// Verify mapping created
	var count int
	err := handler.db.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM permission_user_mapping WHERE permission_id = ? AND user_id = ? AND permission_context_id = ? AND deleted = 0",
		permissionID, userID, contextID).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

// TestPermissionHandler_UnassignUser_Success tests unassigning permission from user
func TestPermissionHandler_UnassignUser_Success(t *testing.T) {
	handler := setupPermissionTestHandler(t)
	gin.SetMode(gin.TestMode)

	mappingID := generateTestID()
	now := time.Now().Unix()

	// Create mapping
	_, err := handler.db.Exec(context.Background(),
		`INSERT INTO permission_user_mapping (id, permission_id, user_id, permission_context_id, created, deleted) VALUES (?, ?, ?, ?, ?, ?)`,
		mappingID, generateTestID(), generateTestID(), generateTestID(), now, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionPermissionUnassignUser,
		Data: map[string]interface{}{
			"id": mappingID,
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")

	handler.handlePermissionUnassignUser(c, &reqBody)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify soft delete
	var deleted int
	err = handler.db.QueryRow(context.Background(), "SELECT deleted FROM permission_user_mapping WHERE id = ?", mappingID).Scan(&deleted)
	require.NoError(t, err)
	assert.Equal(t, 1, deleted)
}

// TestPermissionHandler_AssignTeam_Success tests assigning permission to team
func TestPermissionHandler_AssignTeam_Success(t *testing.T) {
	handler := setupPermissionTestHandler(t)
	gin.SetMode(gin.TestMode)

	permissionID := generateTestID()
	teamID := generateTestID()
	contextID := generateTestID()

	reqBody := models.Request{
		Action: models.ActionPermissionAssignTeam,
		Data: map[string]interface{}{
			"permissionId":        permissionID,
			"teamId":              teamID,
			"permissionContextId": contextID,
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")

	handler.handlePermissionAssignTeam(c, &reqBody)

	assert.Equal(t, http.StatusCreated, w.Code)

	// Verify mapping created
	var count int
	err := handler.db.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM permission_team_mapping WHERE permission_id = ? AND team_id = ? AND permission_context_id = ? AND deleted = 0",
		permissionID, teamID, contextID).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

// TestPermissionHandler_UnassignTeam_Success tests unassigning permission from team
func TestPermissionHandler_UnassignTeam_Success(t *testing.T) {
	handler := setupPermissionTestHandler(t)
	gin.SetMode(gin.TestMode)

	mappingID := generateTestID()
	now := time.Now().Unix()

	// Create mapping
	_, err := handler.db.Exec(context.Background(),
		`INSERT INTO permission_team_mapping (id, permission_id, team_id, permission_context_id, created, deleted) VALUES (?, ?, ?, ?, ?, ?)`,
		mappingID, generateTestID(), generateTestID(), generateTestID(), now, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionPermissionUnassignTeam,
		Data: map[string]interface{}{
			"id": mappingID,
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")

	handler.handlePermissionUnassignTeam(c, &reqBody)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify soft delete
	var deleted int
	err = handler.db.QueryRow(context.Background(), "SELECT deleted FROM permission_team_mapping WHERE id = ?", mappingID).Scan(&deleted)
	require.NoError(t, err)
	assert.Equal(t, 1, deleted)
}

// TestPermissionHandler_Check_Allowed tests permission check returning allowed
func TestPermissionHandler_Check_Allowed(t *testing.T) {
	handler := setupPermissionTestHandler(t)
	gin.SetMode(gin.TestMode)

	// Mock permission service will return true for this check
	reqBody := models.Request{
		Action: models.ActionPermissionCheck,
		Data: map[string]interface{}{
			"userId":     "testuser",
			"resource":   "test-resource",
			"permission": float64(models.PermissionRead),
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")

	handler.handlePermissionCheck(c, &reqBody)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	dataMap := response.Data.(map[string]interface{})
	allowed := dataMap["allowed"].(bool)
	assert.True(t, allowed)
}

// TestPermissionHandler_Check_MissingParameters tests permission check with missing parameters
func TestPermissionHandler_Check_MissingParameters(t *testing.T) {
	handler := setupPermissionTestHandler(t)
	gin.SetMode(gin.TestMode)

	testCases := []struct {
		name string
		data map[string]interface{}
	}{
		{
			"MissingUserId",
			map[string]interface{}{
				"resource":   "test",
				"permission": float64(1),
			},
		},
		{
			"MissingResource",
			map[string]interface{}{
				"userId":     "user1",
				"permission": float64(1),
			},
		},
		{
			"MissingPermission",
			map[string]interface{}{
				"userId":   "user1",
				"resource": "test",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			reqBody := models.Request{
				Action: models.ActionPermissionCheck,
				Data:   tc.data,
			}

			body, _ := json.Marshal(reqBody)
			req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
			w := httptest.NewRecorder()

			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.Set("username", "testuser")

			handler.handlePermissionCheck(c, &reqBody)

			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	}
}

// TestPermissionHandler_FullCRUDCycle tests complete permission lifecycle
func TestPermissionHandler_FullCRUDCycle(t *testing.T) {
	handler := setupPermissionTestHandler(t)
	gin.SetMode(gin.TestMode)

	// 1. Create Permission
	reqBody := models.Request{
		Action: models.ActionPermissionCreate,
		Data: map[string]interface{}{
			"title": "Lifecycle Permission",
			"value": float64(models.PermissionRead),
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")

	handler.handlePermissionCreate(c, &reqBody)
	assert.Equal(t, http.StatusCreated, w.Code)

	var createResponse models.Response
	json.Unmarshal(w.Body.Bytes(), &createResponse)
	dataMap := createResponse.Data.(map[string]interface{})
	permission := dataMap["permission"].(map[string]interface{})
	permID := permission["id"].(string)

	// 2. Read Permission
	reqBody.Action = models.ActionPermissionRead
	reqBody.Data = map[string]interface{}{"id": permID}
	body, _ = json.Marshal(reqBody)
	req = httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")

	handler.handlePermissionRead(c, &reqBody)
	assert.Equal(t, http.StatusOK, w.Code)

	// 3. Modify Permission
	reqBody.Action = models.ActionPermissionModify
	reqBody.Data = map[string]interface{}{
		"id":    permID,
		"title": "Modified Permission",
	}
	body, _ = json.Marshal(reqBody)
	req = httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")

	handler.handlePermissionModify(c, &reqBody)
	assert.Equal(t, http.StatusOK, w.Code)

	// 4. Remove Permission
	reqBody.Action = models.ActionPermissionRemove
	reqBody.Data = map[string]interface{}{"id": permID}
	body, _ = json.Marshal(reqBody)
	req = httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")

	handler.handlePermissionRemove(c, &reqBody)
	assert.Equal(t, http.StatusOK, w.Code)

	// 5. Verify deleted
	var deleted int
	err := handler.db.QueryRow(context.Background(), "SELECT deleted FROM permission WHERE id = ?", permID).Scan(&deleted)
	require.NoError(t, err)
	assert.Equal(t, 1, deleted)
}
