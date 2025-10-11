package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"helixtrack.ru/core/internal/models"
)

// setupVersionTable creates the version table for testing
func setupVersionTable(t *testing.T, handler *Handler) {
	_, err := handler.db.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS version (
			id TEXT PRIMARY KEY,
			title TEXT NOT NULL,
			description TEXT,
			project_id TEXT NOT NULL,
			start_date INTEGER,
			release_date INTEGER,
			released INTEGER NOT NULL DEFAULT 0,
			archived INTEGER NOT NULL DEFAULT 0,
			created INTEGER NOT NULL,
			modified INTEGER NOT NULL,
			deleted INTEGER NOT NULL DEFAULT 0
		)
	`)
	require.NoError(t, err)
}

func setupVersionProjectTable(t *testing.T, handler *Handler) {
	_, err := handler.db.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS project (
			id TEXT PRIMARY KEY,
			title TEXT NOT NULL,
			description TEXT,
			created INTEGER NOT NULL,
			modified INTEGER NOT NULL,
			deleted INTEGER NOT NULL DEFAULT 0
		)
	`)
	require.NoError(t, err)
}

func setupVersionTicketTable(t *testing.T, handler *Handler) {
	_, err := handler.db.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS ticket (
			id TEXT PRIMARY KEY,
			title TEXT NOT NULL,
			description TEXT,
			status TEXT,
			created INTEGER NOT NULL,
			modified INTEGER NOT NULL,
			deleted INTEGER NOT NULL DEFAULT 0
		)
	`)
	require.NoError(t, err)
}

func setupTicketAffectedVersionMappingTable(t *testing.T, handler *Handler) {
	_, err := handler.db.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS ticket_affected_version_mapping (
			id TEXT PRIMARY KEY,
			ticket_id TEXT NOT NULL,
			version_id TEXT NOT NULL,
			created INTEGER NOT NULL,
			deleted INTEGER NOT NULL DEFAULT 0
		)
	`)
	require.NoError(t, err)
}

func setupTicketFixVersionMappingTable(t *testing.T, handler *Handler) {
	_, err := handler.db.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS ticket_fix_version_mapping (
			id TEXT PRIMARY KEY,
			ticket_id TEXT NOT NULL,
			version_id TEXT NOT NULL,
			created INTEGER NOT NULL,
			deleted INTEGER NOT NULL DEFAULT 0
		)
	`)
	require.NoError(t, err)
}

// setupVersionTestHandler creates a test handler with version test data
func setupVersionTestHandler(t *testing.T) *Handler {
	handler := setupTestHandler(t)

	// Setup tables
	setupVersionProjectTable(t, handler)
	setupVersionTicketTable(t, handler)
	setupVersionTable(t, handler)
	setupTicketAffectedVersionMappingTable(t, handler)
	setupTicketFixVersionMappingTable(t, handler)

	// Insert test project
	_, err := handler.db.Exec(context.Background(),
		"INSERT INTO project (id, title, description, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)",
		"test-project-id", "Test Project", "Test project description", 1000, 1000, 0)
	require.NoError(t, err)

	// Insert test ticket for version mappings
	_, err = handler.db.Exec(context.Background(),
		"INSERT INTO ticket (id, title, description, status, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?)",
		"test-ticket-id", "Test Ticket", "Test ticket description", "open", 1000, 1000, 0)
	require.NoError(t, err)

	return handler
}

// setupVersionTestHandlerWithPublisher creates a test handler with version test data and mock publisher
func setupVersionTestHandlerWithPublisher(t *testing.T) (*Handler, *MockEventPublisher) {
	handler, mockPublisher := setupTestHandlerWithPublisher(t)

	// Setup tables
	setupVersionProjectTable(t, handler)
	setupVersionTicketTable(t, handler)
	setupVersionTable(t, handler)
	setupTicketAffectedVersionMappingTable(t, handler)
	setupTicketFixVersionMappingTable(t, handler)

	return handler, mockPublisher
}

// TestVersionHandler_Create_Success tests successful version creation with all fields
func TestVersionHandler_Create_Success(t *testing.T) {
	handler := setupVersionTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionVersionCreate,
		Data: map[string]interface{}{
			"title":       "v1.0.0",
			"description": "First major release",
			"projectId":   "test-project-id",
			"startDate":   float64(1000000),
			"releaseDate": float64(2000000),
		},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")
	c.Set("request", &reqBody)

	handler.DoAction(c)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)

	version, ok := resp.Data["version"].(map[string]interface{})
	require.True(t, ok)
	assert.NotEmpty(t, version["id"])
	assert.Equal(t, "v1.0.0", version["title"])
	assert.Equal(t, "First major release", version["description"])
	assert.Equal(t, "test-project-id", version["projectId"])
}

// TestVersionHandler_Create_MinimalFields tests version creation with only required fields
func TestVersionHandler_Create_MinimalFields(t *testing.T) {
	handler := setupVersionTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionVersionCreate,
		Data: map[string]interface{}{
			"title":     "v2.0.0",
			"projectId": "test-project-id",
		},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")
	c.Set("request", &reqBody)

	handler.DoAction(c)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)

	version, ok := resp.Data["version"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "v2.0.0", version["title"])
	assert.False(t, version["released"].(bool))
	assert.False(t, version["archived"].(bool))
}

// TestVersionHandler_Create_MissingTitle tests version creation without title
func TestVersionHandler_Create_MissingTitle(t *testing.T) {
	handler := setupVersionTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionVersionCreate,
		Data: map[string]interface{}{
			"projectId": "test-project-id",
		},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")
	c.Set("request", &reqBody)

	handler.DoAction(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeMissingData, resp.ErrorCode)
}

// TestVersionHandler_Create_MissingProjectId tests version creation without projectId
func TestVersionHandler_Create_MissingProjectId(t *testing.T) {
	handler := setupVersionTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionVersionCreate,
		Data: map[string]interface{}{
			"title": "v1.0.0",
		},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")
	c.Set("request", &reqBody)

	handler.DoAction(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeMissingData, resp.ErrorCode)
}

// TestVersionHandler_Read_Success tests successful version read
func TestVersionHandler_Read_Success(t *testing.T) {
	handler := setupVersionTestHandler(t)

	// Insert test version
	_, err := handler.db.Exec(context.Background(),
		"INSERT INTO version (id, title, description, project_id, start_date, release_date, released, archived, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		"test-version-id", "v1.0.0", "First release", "test-project-id", nil, nil, 0, 0, 1000, 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionVersionRead,
		Data: map[string]interface{}{
			"id": "test-version-id",
		},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")
	c.Set("request", &reqBody)

	handler.DoAction(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err = json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)

	version, ok := resp.Data["version"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "test-version-id", version["id"])
	assert.Equal(t, "v1.0.0", version["title"])
}

// TestVersionHandler_Read_NotFound tests reading non-existent version
func TestVersionHandler_Read_NotFound(t *testing.T) {
	handler := setupVersionTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionVersionRead,
		Data: map[string]interface{}{
			"id": "non-existent-version",
		},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")
	c.Set("request", &reqBody)

	handler.DoAction(c)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeEntityNotFound, resp.ErrorCode)
}

// TestVersionHandler_List_Empty tests listing versions when none exist
func TestVersionHandler_List_Empty(t *testing.T) {
	handler := setupVersionTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionVersionList,
		Data:   map[string]interface{}{},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")
	c.Set("request", &reqBody)

	handler.DoAction(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)

	versions, ok := resp.Data["versions"].([]interface{})
	require.True(t, ok)
	assert.Empty(t, versions)
}

// TestVersionHandler_List_Multiple tests listing multiple versions
func TestVersionHandler_List_Multiple(t *testing.T) {
	handler := setupVersionTestHandler(t)

	// Insert multiple versions
	versions := []string{"v1.0.0", "v1.1.0", "v2.0.0"}
	for i, title := range versions {
		versionID := "version-" + string(rune('1'+i))
		_, err := handler.db.Exec(context.Background(),
			"INSERT INTO version (id, title, description, project_id, start_date, release_date, released, archived, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
			versionID, title, "Description", "test-project-id", nil, nil, 0, 0, 1000, 1000, 0)
		require.NoError(t, err)
	}

	reqBody := models.Request{
		Action: models.ActionVersionList,
		Data:   map[string]interface{}{},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")
	c.Set("request", &reqBody)

	handler.DoAction(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)

	versionList, ok := resp.Data["versions"].([]interface{})
	require.True(t, ok)
	assert.Len(t, versionList, 3)
}

// TestVersionHandler_List_FilterByProject tests listing versions filtered by project
func TestVersionHandler_List_FilterByProject(t *testing.T) {
	handler := setupVersionTestHandler(t)

	// Insert another project
	_, err := handler.db.Exec(context.Background(),
		"INSERT INTO project (id, title, description, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)",
		"project-2-id", "Project 2", "Second project", 1000, 1000, 0)
	require.NoError(t, err)

	// Insert versions for different projects
	_, err = handler.db.Exec(context.Background(),
		"INSERT INTO version (id, title, description, project_id, start_date, release_date, released, archived, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		"version-1", "v1.0.0", "Version 1", "test-project-id", nil, nil, 0, 0, 1000, 1000, 0)
	require.NoError(t, err)

	_, err = handler.db.Exec(context.Background(),
		"INSERT INTO version (id, title, description, project_id, start_date, release_date, released, archived, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		"version-2", "v2.0.0", "Version 2", "project-2-id", nil, nil, 0, 0, 1000, 1000, 0)
	require.NoError(t, err)

	// Filter by test-project-id
	reqBody := models.Request{
		Action: models.ActionVersionList,
		Data: map[string]interface{}{
			"projectId": "test-project-id",
		},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")
	c.Set("request", &reqBody)

	handler.DoAction(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err = json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)

	versions, ok := resp.Data["versions"].([]interface{})
	require.True(t, ok)
	assert.Len(t, versions, 1)

	version := versions[0].(map[string]interface{})
	assert.Equal(t, "v1.0.0", version["title"])
}

// TestVersionHandler_Modify_Success tests successful version modification
func TestVersionHandler_Modify_Success(t *testing.T) {
	handler := setupVersionTestHandler(t)

	// Insert test version
	_, err := handler.db.Exec(context.Background(),
		"INSERT INTO version (id, title, description, project_id, start_date, release_date, released, archived, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		"test-version-id", "v1.0.0", "Old description", "test-project-id", nil, nil, 0, 0, 1000, 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionVersionModify,
		Data: map[string]interface{}{
			"id":          "test-version-id",
			"title":       "v1.0.1",
			"description": "Updated description",
			"releaseDate": float64(2000000),
		},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")
	c.Set("request", &reqBody)

	handler.DoAction(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err = json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)

	// Verify update in database
	var title, description string
	err = handler.db.QueryRow(context.Background(),
		"SELECT title, description FROM version WHERE id = ?",
		"test-version-id").Scan(&title, &description)
	require.NoError(t, err)
	assert.Equal(t, "v1.0.1", title)
	assert.Equal(t, "Updated description", description)
}

// TestVersionHandler_Modify_NotFound tests modifying non-existent version
func TestVersionHandler_Modify_NotFound(t *testing.T) {
	handler := setupVersionTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionVersionModify,
		Data: map[string]interface{}{
			"id":    "non-existent-version",
			"title": "Updated",
		},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")
	c.Set("request", &reqBody)

	handler.DoAction(c)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeEntityNotFound, resp.ErrorCode)
}

// TestVersionHandler_Remove_Success tests successful version deletion
func TestVersionHandler_Remove_Success(t *testing.T) {
	handler := setupVersionTestHandler(t)

	// Insert test version
	_, err := handler.db.Exec(context.Background(),
		"INSERT INTO version (id, title, description, project_id, start_date, release_date, released, archived, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		"test-version-id", "v0.9.0", "Deprecated version", "test-project-id", nil, nil, 0, 0, 1000, 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionVersionRemove,
		Data: map[string]interface{}{
			"id": "test-version-id",
		},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")
	c.Set("request", &reqBody)

	handler.DoAction(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err = json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)

	// Verify soft delete in database
	var deleted bool
	err = handler.db.QueryRow(context.Background(),
		"SELECT deleted FROM version WHERE id = ?",
		"test-version-id").Scan(&deleted)
	require.NoError(t, err)
	assert.True(t, deleted)
}

// TestVersionHandler_Remove_NotFound tests deleting non-existent version
func TestVersionHandler_Remove_NotFound(t *testing.T) {
	handler := setupVersionTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionVersionRemove,
		Data: map[string]interface{}{
			"id": "non-existent-version",
		},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")
	c.Set("request", &reqBody)

	handler.DoAction(c)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeEntityNotFound, resp.ErrorCode)
}

// TestVersionHandler_Release_Success tests successful version release
func TestVersionHandler_Release_Success(t *testing.T) {
	handler := setupVersionTestHandler(t)

	// Insert test version
	_, err := handler.db.Exec(context.Background(),
		"INSERT INTO version (id, title, description, project_id, start_date, release_date, released, archived, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		"test-version-id", "v1.0.0", "Ready to release", "test-project-id", nil, nil, 0, 0, 1000, 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionVersionRelease,
		Data: map[string]interface{}{
			"id": "test-version-id",
		},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")
	c.Set("request", &reqBody)

	handler.DoAction(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err = json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)

	// Verify released flag in database
	var released bool
	err = handler.db.QueryRow(context.Background(),
		"SELECT released FROM version WHERE id = ?",
		"test-version-id").Scan(&released)
	require.NoError(t, err)
	assert.True(t, released)
}

// TestVersionHandler_Release_NotFound tests releasing non-existent version
func TestVersionHandler_Release_NotFound(t *testing.T) {
	handler := setupVersionTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionVersionRelease,
		Data: map[string]interface{}{
			"id": "non-existent-version",
		},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")
	c.Set("request", &reqBody)

	handler.DoAction(c)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeEntityNotFound, resp.ErrorCode)
}

// TestVersionHandler_Archive_Success tests successful version archiving
func TestVersionHandler_Archive_Success(t *testing.T) {
	handler := setupVersionTestHandler(t)

	// Insert test version
	_, err := handler.db.Exec(context.Background(),
		"INSERT INTO version (id, title, description, project_id, start_date, release_date, released, archived, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		"test-version-id", "v0.1.0", "Old version", "test-project-id", nil, nil, 1, 0, 1000, 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionVersionArchive,
		Data: map[string]interface{}{
			"id": "test-version-id",
		},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")
	c.Set("request", &reqBody)

	handler.DoAction(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err = json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)

	// Verify archived flag in database
	var archived bool
	err = handler.db.QueryRow(context.Background(),
		"SELECT archived FROM version WHERE id = ?",
		"test-version-id").Scan(&archived)
	require.NoError(t, err)
	assert.True(t, archived)
}

// TestVersionHandler_Archive_NotFound tests archiving non-existent version
func TestVersionHandler_Archive_NotFound(t *testing.T) {
	handler := setupVersionTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionVersionArchive,
		Data: map[string]interface{}{
			"id": "non-existent-version",
		},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")
	c.Set("request", &reqBody)

	handler.DoAction(c)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeEntityNotFound, resp.ErrorCode)
}

// TestVersionHandler_AddAffected_Success tests adding affected version to ticket
func TestVersionHandler_AddAffected_Success(t *testing.T) {
	handler := setupVersionTestHandler(t)

	// Insert test version
	_, err := handler.db.Exec(context.Background(),
		"INSERT INTO version (id, title, description, project_id, start_date, release_date, released, archived, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		"test-version-id", "v1.0.0", "Version 1", "test-project-id", nil, nil, 0, 0, 1000, 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionVersionAddAffected,
		Data: map[string]interface{}{
			"ticketId":  "test-ticket-id",
			"versionId": "test-version-id",
		},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")
	c.Set("request", &reqBody)

	handler.DoAction(c)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp models.Response
	err = json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)

	mapping, ok := resp.Data["mapping"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "test-ticket-id", mapping["ticketId"])
	assert.Equal(t, "test-version-id", mapping["versionId"])
}

// TestVersionHandler_RemoveAffected_Success tests removing affected version from ticket
func TestVersionHandler_RemoveAffected_Success(t *testing.T) {
	handler := setupVersionTestHandler(t)

	// Insert test version
	_, err := handler.db.Exec(context.Background(),
		"INSERT INTO version (id, title, description, project_id, start_date, release_date, released, archived, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		"test-version-id", "v1.0.0", "Version 1", "test-project-id", nil, nil, 0, 0, 1000, 1000, 0)
	require.NoError(t, err)

	// Insert mapping
	_, err = handler.db.Exec(context.Background(),
		"INSERT INTO ticket_affected_version_mapping (id, ticket_id, version_id, created, deleted) VALUES (?, ?, ?, ?, ?)",
		"mapping-id", "test-ticket-id", "test-version-id", 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionVersionRemoveAffected,
		Data: map[string]interface{}{
			"ticketId":  "test-ticket-id",
			"versionId": "test-version-id",
		},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")
	c.Set("request", &reqBody)

	handler.DoAction(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err = json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)

	// Verify soft delete in database
	var deleted bool
	err = handler.db.QueryRow(context.Background(),
		"SELECT deleted FROM ticket_affected_version_mapping WHERE ticket_id = ? AND version_id = ?",
		"test-ticket-id", "test-version-id").Scan(&deleted)
	require.NoError(t, err)
	assert.True(t, deleted)
}

// TestVersionHandler_ListAffected_Success tests listing affected versions for ticket
func TestVersionHandler_ListAffected_Success(t *testing.T) {
	handler := setupVersionTestHandler(t)

	// Insert test version
	_, err := handler.db.Exec(context.Background(),
		"INSERT INTO version (id, title, description, project_id, start_date, release_date, released, archived, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		"test-version-id", "v1.0.0", "Version 1", "test-project-id", nil, nil, 0, 0, 1000, 1000, 0)
	require.NoError(t, err)

	// Insert mapping
	_, err = handler.db.Exec(context.Background(),
		"INSERT INTO ticket_affected_version_mapping (id, ticket_id, version_id, created, deleted) VALUES (?, ?, ?, ?, ?)",
		"mapping-id", "test-ticket-id", "test-version-id", 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionVersionListAffected,
		Data: map[string]interface{}{
			"ticketId": "test-ticket-id",
		},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")
	c.Set("request", &reqBody)

	handler.DoAction(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err = json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)

	versions, ok := resp.Data["versions"].([]interface{})
	require.True(t, ok)
	assert.Len(t, versions, 1)

	version := versions[0].(map[string]interface{})
	assert.Equal(t, "v1.0.0", version["title"])
}

// TestVersionHandler_AddFix_Success tests adding fix version to ticket
func TestVersionHandler_AddFix_Success(t *testing.T) {
	handler := setupVersionTestHandler(t)

	// Insert test version
	_, err := handler.db.Exec(context.Background(),
		"INSERT INTO version (id, title, description, project_id, start_date, release_date, released, archived, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		"test-version-id", "v1.0.1", "Fix version", "test-project-id", nil, nil, 0, 0, 1000, 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionVersionAddFix,
		Data: map[string]interface{}{
			"ticketId":  "test-ticket-id",
			"versionId": "test-version-id",
		},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")
	c.Set("request", &reqBody)

	handler.DoAction(c)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp models.Response
	err = json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)

	mapping, ok := resp.Data["mapping"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "test-ticket-id", mapping["ticketId"])
	assert.Equal(t, "test-version-id", mapping["versionId"])
}

// TestVersionHandler_RemoveFix_Success tests removing fix version from ticket
func TestVersionHandler_RemoveFix_Success(t *testing.T) {
	handler := setupVersionTestHandler(t)

	// Insert test version
	_, err := handler.db.Exec(context.Background(),
		"INSERT INTO version (id, title, description, project_id, start_date, release_date, released, archived, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		"test-version-id", "v1.0.1", "Fix version", "test-project-id", nil, nil, 0, 0, 1000, 1000, 0)
	require.NoError(t, err)

	// Insert mapping
	_, err = handler.db.Exec(context.Background(),
		"INSERT INTO ticket_fix_version_mapping (id, ticket_id, version_id, created, deleted) VALUES (?, ?, ?, ?, ?)",
		"mapping-id", "test-ticket-id", "test-version-id", 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionVersionRemoveFix,
		Data: map[string]interface{}{
			"ticketId":  "test-ticket-id",
			"versionId": "test-version-id",
		},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")
	c.Set("request", &reqBody)

	handler.DoAction(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err = json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)

	// Verify soft delete in database
	var deleted bool
	err = handler.db.QueryRow(context.Background(),
		"SELECT deleted FROM ticket_fix_version_mapping WHERE ticket_id = ? AND version_id = ?",
		"test-ticket-id", "test-version-id").Scan(&deleted)
	require.NoError(t, err)
	assert.True(t, deleted)
}

// TestVersionHandler_ListFix_Success tests listing fix versions for ticket
func TestVersionHandler_ListFix_Success(t *testing.T) {
	handler := setupVersionTestHandler(t)

	// Insert test version
	_, err := handler.db.Exec(context.Background(),
		"INSERT INTO version (id, title, description, project_id, start_date, release_date, released, archived, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		"test-version-id", "v1.0.1", "Fix version", "test-project-id", nil, nil, 0, 0, 1000, 1000, 0)
	require.NoError(t, err)

	// Insert mapping
	_, err = handler.db.Exec(context.Background(),
		"INSERT INTO ticket_fix_version_mapping (id, ticket_id, version_id, created, deleted) VALUES (?, ?, ?, ?, ?)",
		"mapping-id", "test-ticket-id", "test-version-id", 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionVersionListFix,
		Data: map[string]interface{}{
			"ticketId": "test-ticket-id",
		},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")
	c.Set("request", &reqBody)

	handler.DoAction(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err = json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)

	versions, ok := resp.Data["versions"].([]interface{})
	require.True(t, ok)
	assert.Len(t, versions, 1)

	version := versions[0].(map[string]interface{})
	assert.Equal(t, "v1.0.1", version["title"])
}

// Event Publishing Tests

// TestVersionHandler_Create_PublishesEvent tests that version creation publishes an event
func TestVersionHandler_Create_PublishesEvent(t *testing.T) {
	handler, mockPublisher := setupVersionTestHandlerWithPublisher(t)

	// Insert test project for project-based context
	_, err := handler.db.Exec(context.Background(),
		"INSERT INTO project (id, title, description, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)",
		"test-project-id", "Test Project", "Test project description", 1000, 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionVersionCreate,
		Data: map[string]interface{}{
			"title":       "v1.0.0",
			"description": "First major release",
			"projectId":   "test-project-id",
			"startDate":   float64(1000000),
			"releaseDate": float64(2000000),
		},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")
	c.Set("request", &reqBody)

	handler.DoAction(c)

	assert.Equal(t, http.StatusCreated, w.Code)

	// Verify event was published
	assert.Equal(t, 1, mockPublisher.GetEventCount())
	lastCall := mockPublisher.GetLastEntityCall()
	require.NotNil(t, lastCall)

	// Verify event details
	assert.Equal(t, models.ActionCreate, lastCall.Action)
	assert.Equal(t, "version", lastCall.Object)
	assert.Equal(t, "testuser", lastCall.Username)
	assert.NotEmpty(t, lastCall.EntityID)

	// Verify event data
	assert.Equal(t, "v1.0.0", lastCall.Data["title"])
	assert.Equal(t, "First major release", lastCall.Data["description"])
	assert.Equal(t, "test-project-id", lastCall.Data["project_id"])
	assert.Equal(t, float64(1000000), lastCall.Data["start_date"])
	assert.Equal(t, float64(2000000), lastCall.Data["release_date"])
	assert.Equal(t, false, lastCall.Data["released"])
	assert.Equal(t, false, lastCall.Data["archived"])

	// Verify project-based context
	assert.Equal(t, "test-project-id", lastCall.Context.ProjectID)
	assert.Contains(t, lastCall.Context.Permissions, "READ")
}

// TestVersionHandler_Modify_PublishesEvent tests that version modification publishes an event
func TestVersionHandler_Modify_PublishesEvent(t *testing.T) {
	handler, mockPublisher := setupVersionTestHandlerWithPublisher(t)

	// Insert test project
	_, err := handler.db.Exec(context.Background(),
		"INSERT INTO project (id, title, description, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)",
		"test-project-id", "Test Project", "Test project description", 1000, 1000, 0)
	require.NoError(t, err)

	// Insert test version
	_, err = handler.db.Exec(context.Background(),
		"INSERT INTO version (id, title, description, project_id, start_date, release_date, released, archived, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		"test-version-id", "v1.0.0", "Old description", "test-project-id", nil, nil, 0, 0, 1000, 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionVersionModify,
		Data: map[string]interface{}{
			"id":          "test-version-id",
			"title":       "v1.0.1",
			"description": "Updated description",
			"releaseDate": float64(2000000),
		},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")
	c.Set("request", &reqBody)

	handler.DoAction(c)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify event was published
	assert.Equal(t, 1, mockPublisher.GetEventCount())
	lastCall := mockPublisher.GetLastEntityCall()
	require.NotNil(t, lastCall)

	// Verify event details
	assert.Equal(t, models.ActionModify, lastCall.Action)
	assert.Equal(t, "version", lastCall.Object)
	assert.Equal(t, "test-version-id", lastCall.EntityID)
	assert.Equal(t, "testuser", lastCall.Username)

	// Verify event data
	assert.Equal(t, "v1.0.1", lastCall.Data["title"])
	assert.Equal(t, "Updated description", lastCall.Data["description"])
	assert.Equal(t, "test-project-id", lastCall.Data["project_id"])

	// Verify project-based context
	assert.Equal(t, "test-project-id", lastCall.Context.ProjectID)
	assert.Contains(t, lastCall.Context.Permissions, "READ")
}

// TestVersionHandler_Remove_PublishesEvent tests that version deletion publishes an event
func TestVersionHandler_Remove_PublishesEvent(t *testing.T) {
	handler, mockPublisher := setupVersionTestHandlerWithPublisher(t)

	// Insert test project
	_, err := handler.db.Exec(context.Background(),
		"INSERT INTO project (id, title, description, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)",
		"test-project-id", "Test Project", "Test project description", 1000, 1000, 0)
	require.NoError(t, err)

	// Insert test version
	_, err = handler.db.Exec(context.Background(),
		"INSERT INTO version (id, title, description, project_id, start_date, release_date, released, archived, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		"test-version-id", "v0.9.0", "Deprecated version", "test-project-id", nil, nil, 0, 0, 1000, 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionVersionRemove,
		Data: map[string]interface{}{
			"id": "test-version-id",
		},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")
	c.Set("request", &reqBody)

	handler.DoAction(c)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify event was published
	assert.Equal(t, 1, mockPublisher.GetEventCount())
	lastCall := mockPublisher.GetLastEntityCall()
	require.NotNil(t, lastCall)

	// Verify event details
	assert.Equal(t, models.ActionRemove, lastCall.Action)
	assert.Equal(t, "version", lastCall.Object)
	assert.Equal(t, "test-version-id", lastCall.EntityID)
	assert.Equal(t, "testuser", lastCall.Username)

	// Verify event data
	assert.Equal(t, "test-version-id", lastCall.Data["id"])
	assert.Equal(t, "v0.9.0", lastCall.Data["title"])
	assert.Equal(t, "test-project-id", lastCall.Data["project_id"])

	// Verify project-based context
	assert.Equal(t, "test-project-id", lastCall.Context.ProjectID)
	assert.Contains(t, lastCall.Context.Permissions, "READ")
}

// TestVersionHandler_Release_PublishesEvent tests that version release publishes an event
func TestVersionHandler_Release_PublishesEvent(t *testing.T) {
	handler, mockPublisher := setupVersionTestHandlerWithPublisher(t)

	// Insert test project
	_, err := handler.db.Exec(context.Background(),
		"INSERT INTO project (id, title, description, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)",
		"test-project-id", "Test Project", "Test project description", 1000, 1000, 0)
	require.NoError(t, err)

	// Insert test version
	_, err = handler.db.Exec(context.Background(),
		"INSERT INTO version (id, title, description, project_id, start_date, release_date, released, archived, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		"test-version-id", "v1.0.0", "Ready to release", "test-project-id", nil, nil, 0, 0, 1000, 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionVersionRelease,
		Data: map[string]interface{}{
			"id": "test-version-id",
		},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")
	c.Set("request", &reqBody)

	handler.DoAction(c)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify event was published
	assert.Equal(t, 1, mockPublisher.GetEventCount())
	lastCall := mockPublisher.GetLastEntityCall()
	require.NotNil(t, lastCall)

	// Verify event details (special release operation)
	assert.Equal(t, "versionRelease", lastCall.Action)
	assert.Equal(t, "version", lastCall.Object)
	assert.Equal(t, "test-version-id", lastCall.EntityID)
	assert.Equal(t, "testuser", lastCall.Username)

	// Verify event data
	assert.Equal(t, "test-version-id", lastCall.Data["id"])
	assert.Equal(t, "v1.0.0", lastCall.Data["title"])
	assert.Equal(t, "test-project-id", lastCall.Data["project_id"])
	assert.Equal(t, true, lastCall.Data["released"])

	// Verify project-based context
	assert.Equal(t, "test-project-id", lastCall.Context.ProjectID)
	assert.Contains(t, lastCall.Context.Permissions, "READ")
}

// TestVersionHandler_Archive_PublishesEvent tests that version archive publishes an event
func TestVersionHandler_Archive_PublishesEvent(t *testing.T) {
	handler, mockPublisher := setupVersionTestHandlerWithPublisher(t)

	// Insert test project
	_, err := handler.db.Exec(context.Background(),
		"INSERT INTO project (id, title, description, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)",
		"test-project-id", "Test Project", "Test project description", 1000, 1000, 0)
	require.NoError(t, err)

	// Insert test version
	_, err = handler.db.Exec(context.Background(),
		"INSERT INTO version (id, title, description, project_id, start_date, release_date, released, archived, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		"test-version-id", "v0.1.0", "Old version", "test-project-id", nil, nil, 1, 0, 1000, 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionVersionArchive,
		Data: map[string]interface{}{
			"id": "test-version-id",
		},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")
	c.Set("request", &reqBody)

	handler.DoAction(c)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify event was published
	assert.Equal(t, 1, mockPublisher.GetEventCount())
	lastCall := mockPublisher.GetLastEntityCall()
	require.NotNil(t, lastCall)

	// Verify event details (special archive operation)
	assert.Equal(t, "versionArchive", lastCall.Action)
	assert.Equal(t, "version", lastCall.Object)
	assert.Equal(t, "test-version-id", lastCall.EntityID)
	assert.Equal(t, "testuser", lastCall.Username)

	// Verify event data
	assert.Equal(t, "test-version-id", lastCall.Data["id"])
	assert.Equal(t, "v0.1.0", lastCall.Data["title"])
	assert.Equal(t, "test-project-id", lastCall.Data["project_id"])
	assert.Equal(t, true, lastCall.Data["archived"])

	// Verify project-based context
	assert.Equal(t, "test-project-id", lastCall.Context.ProjectID)
	assert.Contains(t, lastCall.Context.Permissions, "READ")
}

// TestVersionHandler_Create_NoEventOnFailure tests that no event is published on create failure
func TestVersionHandler_Create_NoEventOnFailure(t *testing.T) {
	handler, mockPublisher := setupVersionTestHandlerWithPublisher(t)

	// Insert test project
	_, err := handler.db.Exec(context.Background(),
		"INSERT INTO project (id, title, description, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)",
		"test-project-id", "Test Project", "Test project description", 1000, 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionVersionCreate,
		Data: map[string]interface{}{
			// Missing required field 'title'
			"projectId": "test-project-id",
		},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")
	c.Set("request", &reqBody)

	handler.DoAction(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Verify no event was published
	assert.Equal(t, 0, mockPublisher.GetEventCount())
}

// TestVersionHandler_Modify_NoEventOnFailure tests that no event is published on modify failure
func TestVersionHandler_Modify_NoEventOnFailure(t *testing.T) {
	handler, mockPublisher := setupVersionTestHandlerWithPublisher(t)

	// Insert test project
	_, err := handler.db.Exec(context.Background(),
		"INSERT INTO project (id, title, description, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)",
		"test-project-id", "Test Project", "Test project description", 1000, 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionVersionModify,
		Data: map[string]interface{}{
			"id":    "non-existent-version",
			"title": "Updated",
		},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")
	c.Set("request", &reqBody)

	handler.DoAction(c)

	assert.Equal(t, http.StatusNotFound, w.Code)

	// Verify no event was published
	assert.Equal(t, 0, mockPublisher.GetEventCount())
}

// TestVersionHandler_Remove_NoEventOnFailure tests that no event is published on remove failure
func TestVersionHandler_Remove_NoEventOnFailure(t *testing.T) {
	handler, mockPublisher := setupVersionTestHandlerWithPublisher(t)

	// Insert test project
	_, err := handler.db.Exec(context.Background(),
		"INSERT INTO project (id, title, description, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)",
		"test-project-id", "Test Project", "Test project description", 1000, 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionVersionRemove,
		Data: map[string]interface{}{
			"id": "non-existent-version",
		},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")
	c.Set("request", &reqBody)

	handler.DoAction(c)

	assert.Equal(t, http.StatusNotFound, w.Code)

	// Verify no event was published
	assert.Equal(t, 0, mockPublisher.GetEventCount())
}

// TestVersionHandler_Release_NoEventOnFailure tests that no event is published on release failure
func TestVersionHandler_Release_NoEventOnFailure(t *testing.T) {
	handler, mockPublisher := setupVersionTestHandlerWithPublisher(t)

	// Insert test project
	_, err := handler.db.Exec(context.Background(),
		"INSERT INTO project (id, title, description, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)",
		"test-project-id", "Test Project", "Test project description", 1000, 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionVersionRelease,
		Data: map[string]interface{}{
			"id": "non-existent-version",
		},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")
	c.Set("request", &reqBody)

	handler.DoAction(c)

	assert.Equal(t, http.StatusNotFound, w.Code)

	// Verify no event was published
	assert.Equal(t, 0, mockPublisher.GetEventCount())
}

// TestVersionHandler_Archive_NoEventOnFailure tests that no event is published on archive failure
func TestVersionHandler_Archive_NoEventOnFailure(t *testing.T) {
	handler, mockPublisher := setupVersionTestHandlerWithPublisher(t)

	// Insert test project
	_, err := handler.db.Exec(context.Background(),
		"INSERT INTO project (id, title, description, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)",
		"test-project-id", "Test Project", "Test project description", 1000, 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionVersionArchive,
		Data: map[string]interface{}{
			"id": "non-existent-version",
		},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")
	c.Set("request", &reqBody)

	handler.DoAction(c)

	assert.Equal(t, http.StatusNotFound, w.Code)

	// Verify no event was published
	assert.Equal(t, 0, mockPublisher.GetEventCount())
}
