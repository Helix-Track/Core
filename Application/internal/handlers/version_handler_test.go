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

// setupVersionTestHandler creates a test handler with version test data
func setupVersionTestHandler(t *testing.T) *Handler {
	handler := setupTestHandler(t)

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

	handler.DoAction(c)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNotFound, resp.ErrorCode)
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

	handler.DoAction(c)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNotFound, resp.ErrorCode)
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

	handler.DoAction(c)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNotFound, resp.ErrorCode)
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

	handler.DoAction(c)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNotFound, resp.ErrorCode)
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

	handler.DoAction(c)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNotFound, resp.ErrorCode)
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
