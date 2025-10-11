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

// setupProjectTestHandler creates a test handler with a default workflow
func setupProjectTestHandler(t *testing.T) *Handler {
	handler := setupTestHandler(t)

	// Insert default workflow for project creation
	_, err := handler.db.Exec(context.Background(),
		"INSERT INTO workflow (id, title, description, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)",
		"default-workflow-id", "Default Workflow", "Default workflow for testing", 1000, 1000, 0)
	require.NoError(t, err)

	return handler
}

// =============================================================================
// handleCreateProject Tests
// =============================================================================

func TestProjectHandler_Create_Success(t *testing.T) {
	handler := setupProjectTestHandler(t)
	router := gin.New()
	router.POST("/do", handler.DoAction)

	reqBody := models.Request{
		Action: models.ActionCreate,
		Object: "project",
		Data: map[string]interface{}{
			"name":        "Test Project",
			"key":         "TEST",
			"description": "A test project",
			"type":        "software",
		},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Add username to context to pass authorization
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")

	handler.DoAction(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)

	project, ok := resp.Data["project"].(map[string]interface{})
	require.True(t, ok)
	assert.NotEmpty(t, project["id"])
	assert.Equal(t, "TEST", project["identifier"])
	assert.Equal(t, "Test Project", project["title"])
	assert.Equal(t, "A test project", project["description"])
}

func TestProjectHandler_Create_MinimalFields(t *testing.T) {
	handler := setupProjectTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionCreate,
		Object: "project",
		Data: map[string]interface{}{
			"name": "Minimal Project",
			"key":  "MIN",
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
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)

	project, ok := resp.Data["project"].(map[string]interface{})
	require.True(t, ok)
	assert.NotEmpty(t, project["id"])
	assert.Equal(t, "MIN", project["identifier"])
	assert.Equal(t, "Minimal Project", project["title"])
}

func TestProjectHandler_Create_MissingName(t *testing.T) {
	handler := setupProjectTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionCreate,
		Object: "project",
		Data: map[string]interface{}{
			"key": "TEST",
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
	assert.Contains(t, resp.ErrorMessage, "project name")
}

func TestProjectHandler_Create_MissingKey(t *testing.T) {
	handler := setupProjectTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionCreate,
		Object: "project",
		Data: map[string]interface{}{
			"name": "Test Project",
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
	assert.Contains(t, resp.ErrorMessage, "project key")
}

func TestProjectHandler_Create_DuplicateKey(t *testing.T) {
	handler := setupProjectTestHandler(t)

	// Create first project
	reqBody := models.Request{
		Action: models.ActionCreate,
		Object: "project",
		Data: map[string]interface{}{
			"name": "First Project",
			"key":  "DUP",
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

	// Try to create second project with same key
	reqBody2 := models.Request{
		Action: models.ActionCreate,
		Object: "project",
		Data: map[string]interface{}{
			"name": "Second Project",
			"key":  "DUP",
		},
	}
	body2, _ := json.Marshal(reqBody2)

	req2 := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(body2))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()

	c2, _ := gin.CreateTestContext(w2)
	c2.Request = req2
	c2.Set("username", "testuser")

	handler.DoAction(c2)

	assert.Equal(t, http.StatusConflict, w2.Code)

	var resp models.Response
	err := json.NewDecoder(w2.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeEntityAlreadyExists, resp.ErrorCode)
	assert.Contains(t, resp.ErrorMessage, "already exists")
}

func TestProjectHandler_Create_DefaultType(t *testing.T) {
	handler := setupProjectTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionCreate,
		Object: "project",
		Data: map[string]interface{}{
			"name": "Type Test",
			"key":  "TYPE",
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
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)

	project, ok := resp.Data["project"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "software", project["type"])
}

// =============================================================================
// handleModifyProject Tests
// =============================================================================

func TestProjectHandler_Modify_Success(t *testing.T) {
	handler := setupProjectTestHandler(t)

	// Create a project first
	createReq := models.Request{
		Action: models.ActionCreate,
		Object: "project",
		Data: map[string]interface{}{
			"name": "Original Project",
			"key":  "ORIG",
		},
	}
	createBody, _ := json.Marshal(createReq)
	createHttpReq := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(createBody))
	createHttpReq.Header.Set("Content-Type", "application/json")
	wCreate := httptest.NewRecorder()
	cCreate, _ := gin.CreateTestContext(wCreate)
	cCreate.Request = createHttpReq
	cCreate.Set("username", "testuser")
	handler.DoAction(cCreate)

	var createResp models.Response
	json.NewDecoder(wCreate.Body).Decode(&createResp)
	projectData := createResp.Data["project"].(map[string]interface{})
	projectID := projectData["id"].(string)

	// Now modify the project
	modifyReq := models.Request{
		Action: models.ActionModify,
		Object: "project",
		Data: map[string]interface{}{
			"id":          projectID,
			"title":       "Updated Project",
			"description": "Updated description",
		},
	}
	modifyBody, _ := json.Marshal(modifyReq)
	modifyHttpReq := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(modifyBody))
	modifyHttpReq.Header.Set("Content-Type", "application/json")
	wModify := httptest.NewRecorder()
	cModify, _ := gin.CreateTestContext(wModify)
	cModify.Request = modifyHttpReq
	cModify.Set("username", "testuser")
	handler.DoAction(cModify)

	assert.Equal(t, http.StatusOK, wModify.Code)

	var modifyResp models.Response
	err := json.NewDecoder(wModify.Body).Decode(&modifyResp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, modifyResp.ErrorCode)

	modifiedProject := modifyResp.Data["project"].(map[string]interface{})
	assert.Equal(t, projectID, modifiedProject["id"])
	assert.True(t, modifiedProject["updated"].(bool))
}

func TestProjectHandler_Modify_MissingID(t *testing.T) {
	handler := setupProjectTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionModify,
		Object: "project",
		Data: map[string]interface{}{
			"title": "Updated Project",
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
	assert.Contains(t, resp.ErrorMessage, "project ID")
}

func TestProjectHandler_Modify_NotFound(t *testing.T) {
	handler := setupProjectTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionModify,
		Object: "project",
		Data: map[string]interface{}{
			"id":    "non-existent-id",
			"title": "Updated Project",
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
	assert.Equal(t, models.ErrorCodeEntityNotFound, resp.ErrorCode)
	assert.Contains(t, resp.ErrorMessage, "not found")
}

func TestProjectHandler_Modify_OnlyTitle(t *testing.T) {
	handler := setupProjectTestHandler(t)

	// Create project
	createReq := models.Request{
		Action: models.ActionCreate,
		Object: "project",
		Data: map[string]interface{}{
			"name": "Original",
			"key":  "MOD1",
		},
	}
	createBody, _ := json.Marshal(createReq)
	createHttpReq := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(createBody))
	createHttpReq.Header.Set("Content-Type", "application/json")
	wCreate := httptest.NewRecorder()
	cCreate, _ := gin.CreateTestContext(wCreate)
	cCreate.Request = createHttpReq
	cCreate.Set("username", "testuser")
	handler.DoAction(cCreate)

	var createResp models.Response
	json.NewDecoder(wCreate.Body).Decode(&createResp)
	projectID := createResp.Data["project"].(map[string]interface{})["id"].(string)

	// Modify only title
	modifyReq := models.Request{
		Action: models.ActionModify,
		Object: "project",
		Data: map[string]interface{}{
			"id":    projectID,
			"title": "New Title Only",
		},
	}
	modifyBody, _ := json.Marshal(modifyReq)
	modifyHttpReq := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(modifyBody))
	modifyHttpReq.Header.Set("Content-Type", "application/json")
	wModify := httptest.NewRecorder()
	cModify, _ := gin.CreateTestContext(wModify)
	cModify.Request = modifyHttpReq
	cModify.Set("username", "testuser")
	handler.DoAction(cModify)

	assert.Equal(t, http.StatusOK, wModify.Code)
}

// =============================================================================
// handleRemoveProject Tests
// =============================================================================

func TestProjectHandler_Remove_Success(t *testing.T) {
	handler := setupProjectTestHandler(t)

	// Create project
	createReq := models.Request{
		Action: models.ActionCreate,
		Object: "project",
		Data: map[string]interface{}{
			"name": "To Delete",
			"key":  "DEL",
		},
	}
	createBody, _ := json.Marshal(createReq)
	createHttpReq := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(createBody))
	createHttpReq.Header.Set("Content-Type", "application/json")
	wCreate := httptest.NewRecorder()
	cCreate, _ := gin.CreateTestContext(wCreate)
	cCreate.Request = createHttpReq
	cCreate.Set("username", "testuser")
	handler.DoAction(cCreate)

	var createResp models.Response
	json.NewDecoder(wCreate.Body).Decode(&createResp)
	projectID := createResp.Data["project"].(map[string]interface{})["id"].(string)

	// Remove project
	removeReq := models.Request{
		Action: models.ActionRemove,
		Object: "project",
		Data: map[string]interface{}{
			"id": projectID,
		},
	}
	removeBody, _ := json.Marshal(removeReq)
	removeHttpReq := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(removeBody))
	removeHttpReq.Header.Set("Content-Type", "application/json")
	wRemove := httptest.NewRecorder()
	cRemove, _ := gin.CreateTestContext(wRemove)
	cRemove.Request = removeHttpReq
	cRemove.Set("username", "testuser")
	handler.DoAction(cRemove)

	assert.Equal(t, http.StatusOK, wRemove.Code)

	var removeResp models.Response
	err := json.NewDecoder(wRemove.Body).Decode(&removeResp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, removeResp.ErrorCode)

	removedProject := removeResp.Data["project"].(map[string]interface{})
	assert.Equal(t, projectID, removedProject["id"])
	assert.True(t, removedProject["deleted"].(bool))
}

func TestProjectHandler_Remove_MissingID(t *testing.T) {
	handler := setupProjectTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionRemove,
		Object: "project",
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

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeMissingData, resp.ErrorCode)
	assert.Contains(t, resp.ErrorMessage, "project ID")
}

// =============================================================================
// handleReadProject Tests
// =============================================================================

func TestProjectHandler_Read_Success(t *testing.T) {
	handler := setupProjectTestHandler(t)

	// Create project
	createReq := models.Request{
		Action: models.ActionCreate,
		Object: "project",
		Data: map[string]interface{}{
			"name":        "Read Test",
			"key":         "READ",
			"description": "Test description",
		},
	}
	createBody, _ := json.Marshal(createReq)
	createHttpReq := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(createBody))
	createHttpReq.Header.Set("Content-Type", "application/json")
	wCreate := httptest.NewRecorder()
	cCreate, _ := gin.CreateTestContext(wCreate)
	cCreate.Request = createHttpReq
	cCreate.Set("username", "testuser")
	handler.DoAction(cCreate)

	var createResp models.Response
	json.NewDecoder(wCreate.Body).Decode(&createResp)
	projectID := createResp.Data["project"].(map[string]interface{})["id"].(string)

	// Read project
	readReq := models.Request{
		Action: models.ActionRead,
		Object: "project",
		Data: map[string]interface{}{
			"id": projectID,
		},
	}
	readBody, _ := json.Marshal(readReq)
	readHttpReq := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(readBody))
	readHttpReq.Header.Set("Content-Type", "application/json")
	wRead := httptest.NewRecorder()
	cRead, _ := gin.CreateTestContext(wRead)
	cRead.Request = readHttpReq
	cRead.Set("username", "testuser")
	handler.DoAction(cRead)

	assert.Equal(t, http.StatusOK, wRead.Code)

	var readResp models.Response
	err := json.NewDecoder(wRead.Body).Decode(&readResp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, readResp.ErrorCode)

	project := readResp.Data["project"].(map[string]interface{})
	assert.Equal(t, projectID, project["id"])
	assert.Equal(t, "READ", project["identifier"])
	assert.Equal(t, "Read Test", project["title"])
	assert.Equal(t, "Test description", project["description"])
	assert.NotEmpty(t, project["workflowId"])
}

func TestProjectHandler_Read_MissingID(t *testing.T) {
	handler := setupProjectTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionRead,
		Object: "project",
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

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeMissingData, resp.ErrorCode)
	assert.Contains(t, resp.ErrorMessage, "project ID")
}

func TestProjectHandler_Read_NotFound(t *testing.T) {
	handler := setupProjectTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionRead,
		Object: "project",
		Data: map[string]interface{}{
			"id": "non-existent-id",
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
	assert.Equal(t, models.ErrorCodeEntityNotFound, resp.ErrorCode)
	assert.Contains(t, resp.ErrorMessage, "not found")
}

func TestProjectHandler_Read_DeletedProject(t *testing.T) {
	handler := setupProjectTestHandler(t)

	// Create and then delete project
	createReq := models.Request{
		Action: models.ActionCreate,
		Object: "project",
		Data: map[string]interface{}{
			"name": "To Delete",
			"key":  "DELD",
		},
	}
	createBody, _ := json.Marshal(createReq)
	createHttpReq := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(createBody))
	createHttpReq.Header.Set("Content-Type", "application/json")
	wCreate := httptest.NewRecorder()
	cCreate, _ := gin.CreateTestContext(wCreate)
	cCreate.Request = createHttpReq
	cCreate.Set("username", "testuser")
	handler.DoAction(cCreate)

	var createResp models.Response
	json.NewDecoder(wCreate.Body).Decode(&createResp)
	projectID := createResp.Data["project"].(map[string]interface{})["id"].(string)

	// Delete project
	removeReq := models.Request{
		Action: models.ActionRemove,
		Object: "project",
		Data: map[string]interface{}{
			"id": projectID,
		},
	}
	removeBody, _ := json.Marshal(removeReq)
	removeHttpReq := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(removeBody))
	removeHttpReq.Header.Set("Content-Type", "application/json")
	wRemove := httptest.NewRecorder()
	cRemove, _ := gin.CreateTestContext(wRemove)
	cRemove.Request = removeHttpReq
	cRemove.Set("username", "testuser")
	handler.DoAction(cRemove)

	// Try to read deleted project
	readReq := models.Request{
		Action: models.ActionRead,
		Object: "project",
		Data: map[string]interface{}{
			"id": projectID,
		},
	}
	readBody, _ := json.Marshal(readReq)
	readHttpReq := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(readBody))
	readHttpReq.Header.Set("Content-Type", "application/json")
	wRead := httptest.NewRecorder()
	cRead, _ := gin.CreateTestContext(wRead)
	cRead.Request = readHttpReq
	cRead.Set("username", "testuser")
	handler.DoAction(cRead)

	assert.Equal(t, http.StatusNotFound, wRead.Code)
}

// =============================================================================
// handleListProjects Tests
// =============================================================================

func TestProjectHandler_List_Empty(t *testing.T) {
	handler := setupProjectTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionList,
		Object: "project",
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

	items, ok := resp.Data["items"].([]interface{})
	require.True(t, ok)
	assert.Equal(t, 0, len(items))
	assert.Equal(t, float64(0), resp.Data["total"])
}

func TestProjectHandler_List_Multiple(t *testing.T) {
	handler := setupProjectTestHandler(t)

	// Create multiple projects
	projectNames := []string{"Project 1", "Project 2", "Project 3"}
	for i, name := range projectNames {
		createReq := models.Request{
			Action: models.ActionCreate,
			Object: "project",
			Data: map[string]interface{}{
				"name": name,
				"key":  fmt.Sprintf("P%d", i+1),
			},
		}
		createBody, _ := json.Marshal(createReq)
		createHttpReq := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(createBody))
		createHttpReq.Header.Set("Content-Type", "application/json")
		wCreate := httptest.NewRecorder()
		cCreate, _ := gin.CreateTestContext(wCreate)
		cCreate.Request = createHttpReq
		cCreate.Set("username", "testuser")
		handler.DoAction(cCreate)
	}

	// List all projects
	reqBody := models.Request{
		Action: models.ActionList,
		Object: "project",
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

	items, ok := resp.Data["items"].([]interface{})
	require.True(t, ok)
	assert.Equal(t, 3, len(items))
	assert.Equal(t, float64(3), resp.Data["total"])

	// Verify projects are ordered by created DESC
	firstProject := items[0].(map[string]interface{})
	assert.NotEmpty(t, firstProject["id"])
	assert.NotEmpty(t, firstProject["identifier"])
	assert.NotEmpty(t, firstProject["title"])
}

func TestProjectHandler_List_ExcludesDeleted(t *testing.T) {
	handler := setupProjectTestHandler(t)

	// Create 2 projects
	for i := 1; i <= 2; i++ {
		createReq := models.Request{
			Action: models.ActionCreate,
			Object: "project",
			Data: map[string]interface{}{
				"name": fmt.Sprintf("Project %d", i),
				"key":  fmt.Sprintf("DEL%d", i),
			},
		}
		createBody, _ := json.Marshal(createReq)
		createHttpReq := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(createBody))
		createHttpReq.Header.Set("Content-Type", "application/json")
		wCreate := httptest.NewRecorder()
		cCreate, _ := gin.CreateTestContext(wCreate)
		cCreate.Request = createHttpReq
		cCreate.Set("username", "testuser")
		handler.DoAction(cCreate)
	}

	// List before deletion
	listReq := models.Request{
		Action: models.ActionList,
		Object: "project",
		Data:   map[string]interface{}{},
	}
	listBody, _ := json.Marshal(listReq)
	listHttpReq := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(listBody))
	listHttpReq.Header.Set("Content-Type", "application/json")
	wList := httptest.NewRecorder()
	cList, _ := gin.CreateTestContext(wList)
	cList.Request = listHttpReq
	cList.Set("username", "testuser")
	handler.DoAction(cList)

	var listResp models.Response
	json.NewDecoder(wList.Body).Decode(&listResp)
	itemsBefore := listResp.Data["items"].([]interface{})
	projectID := itemsBefore[0].(map[string]interface{})["id"].(string)

	// Delete one project
	removeReq := models.Request{
		Action: models.ActionRemove,
		Object: "project",
		Data: map[string]interface{}{
			"id": projectID,
		},
	}
	removeBody, _ := json.Marshal(removeReq)
	removeHttpReq := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(removeBody))
	removeHttpReq.Header.Set("Content-Type", "application/json")
	wRemove := httptest.NewRecorder()
	cRemove, _ := gin.CreateTestContext(wRemove)
	cRemove.Request = removeHttpReq
	cRemove.Set("username", "testuser")
	handler.DoAction(cRemove)

	// List after deletion
	listReq2 := models.Request{
		Action: models.ActionList,
		Object: "project",
		Data:   map[string]interface{}{},
	}
	listBody2, _ := json.Marshal(listReq2)
	listHttpReq2 := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(listBody2))
	listHttpReq2.Header.Set("Content-Type", "application/json")
	wList2 := httptest.NewRecorder()
	cList2, _ := gin.CreateTestContext(wList2)
	cList2.Request = listHttpReq2
	cList2.Set("username", "testuser")
	handler.DoAction(cList2)

	var listResp2 models.Response
	json.NewDecoder(wList2.Body).Decode(&listResp2)
	itemsAfter := listResp2.Data["items"].([]interface{})

	// Should have 1 less project
	assert.Equal(t, 1, len(itemsAfter))
}

// =============================================================================
// Helper Function Tests
// =============================================================================

func TestJoinWithComma(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected string
	}{
		{
			name:     "Empty slice",
			input:    []string{},
			expected: "",
		},
		{
			name:     "Single element",
			input:    []string{"field1 = ?"},
			expected: "field1 = ?",
		},
		{
			name:     "Multiple elements",
			input:    []string{"field1 = ?", "field2 = ?", "field3 = ?"},
			expected: "field1 = ?, field2 = ?, field3 = ?",
		},
		{
			name:     "Two elements",
			input:    []string{"title = ?", "description = ?"},
			expected: "title = ?, description = ?",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := joinWithComma(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// =============================================================================
// Event Publishing Tests
// =============================================================================

// TestProjectHandler_Create_PublishesEvent tests that project creation publishes an event
func TestProjectHandler_Create_PublishesEvent(t *testing.T) {
	handler, mockPublisher := setupTestHandlerWithPublisher(t)

	// Insert default workflow for project creation
	_, err := handler.db.Exec(context.Background(),
		"INSERT INTO workflow (id, title, description, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)",
		"default-workflow-id", "Default Workflow", "Default workflow for testing", 1000, 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionCreate,
		Object: "project",
		Data: map[string]interface{}{
			"name":        "Test Project",
			"key":         "TEST",
			"description": "A test project",
			"type":        "software",
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

	// Verify event was published
	assert.Equal(t, 1, mockPublisher.GetEventCount())
	lastCall := mockPublisher.GetLastEntityCall()
	require.NotNil(t, lastCall)

	// Verify event details
	assert.Equal(t, models.ActionCreate, lastCall.Action)
	assert.Equal(t, "project", lastCall.Object)
	assert.Equal(t, "testuser", lastCall.Username)
	assert.NotEmpty(t, lastCall.EntityID)

	// Verify event data
	assert.Equal(t, lastCall.EntityID, lastCall.Data["id"])
	assert.Equal(t, "TEST", lastCall.Data["identifier"])
	assert.Equal(t, "Test Project", lastCall.Data["title"])
	assert.Equal(t, "A test project", lastCall.Data["description"])
	assert.Equal(t, "software", lastCall.Data["type"])

	// Verify project-based context (self-referential)
	assert.Equal(t, lastCall.EntityID, lastCall.Context.ProjectID)
	assert.Contains(t, lastCall.Context.Permissions, "READ")
}

// TestProjectHandler_Modify_PublishesEvent tests that project modification publishes an event
func TestProjectHandler_Modify_PublishesEvent(t *testing.T) {
	handler, mockPublisher := setupTestHandlerWithPublisher(t)

	// Insert default workflow
	_, err := handler.db.Exec(context.Background(),
		"INSERT INTO workflow (id, title, description, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)",
		"default-workflow-id", "Default Workflow", "Default workflow for testing", 1000, 1000, 0)
	require.NoError(t, err)

	// Insert test project
	_, err = handler.db.Exec(context.Background(),
		"INSERT INTO project (id, identifier, title, description, workflow_id, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		"test-project-id", "TEST", "Original Project", "Original description", "default-workflow-id", 1000, 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionModify,
		Object: "project",
		Data: map[string]interface{}{
			"id":          "test-project-id",
			"title":       "Updated Project",
			"description": "Updated description",
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

	// Verify event was published
	assert.Equal(t, 1, mockPublisher.GetEventCount())
	lastCall := mockPublisher.GetLastEntityCall()
	require.NotNil(t, lastCall)

	// Verify event details
	assert.Equal(t, models.ActionModify, lastCall.Action)
	assert.Equal(t, "project", lastCall.Object)
	assert.Equal(t, "test-project-id", lastCall.EntityID)
	assert.Equal(t, "testuser", lastCall.Username)

	// Verify event data
	assert.Equal(t, "test-project-id", lastCall.Data["id"])
	assert.Equal(t, "Updated Project", lastCall.Data["title"])
	assert.Equal(t, "Updated description", lastCall.Data["description"])

	// Verify project-based context (self-referential)
	assert.Equal(t, "test-project-id", lastCall.Context.ProjectID)
	assert.Contains(t, lastCall.Context.Permissions, "READ")
}

// TestProjectHandler_Remove_PublishesEvent tests that project deletion publishes an event
func TestProjectHandler_Remove_PublishesEvent(t *testing.T) {
	handler, mockPublisher := setupTestHandlerWithPublisher(t)

	// Insert default workflow
	_, err := handler.db.Exec(context.Background(),
		"INSERT INTO workflow (id, title, description, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)",
		"default-workflow-id", "Default Workflow", "Default workflow for testing", 1000, 1000, 0)
	require.NoError(t, err)

	// Insert test project
	_, err = handler.db.Exec(context.Background(),
		"INSERT INTO project (id, identifier, title, description, workflow_id, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		"test-project-id", "TEST", "To Delete", "Project to be deleted", "default-workflow-id", 1000, 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionRemove,
		Object: "project",
		Data: map[string]interface{}{
			"id": "test-project-id",
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

	// Verify event was published
	assert.Equal(t, 1, mockPublisher.GetEventCount())
	lastCall := mockPublisher.GetLastEntityCall()
	require.NotNil(t, lastCall)

	// Verify event details
	assert.Equal(t, models.ActionRemove, lastCall.Action)
	assert.Equal(t, "project", lastCall.Object)
	assert.Equal(t, "test-project-id", lastCall.EntityID)
	assert.Equal(t, "testuser", lastCall.Username)

	// Verify event data
	assert.Equal(t, "test-project-id", lastCall.Data["id"])

	// Verify project-based context (self-referential)
	assert.Equal(t, "test-project-id", lastCall.Context.ProjectID)
	assert.Contains(t, lastCall.Context.Permissions, "READ")
}

// TestProjectHandler_Create_NoEventOnFailure tests that no event is published on create failure
func TestProjectHandler_Create_NoEventOnFailure(t *testing.T) {
	handler, mockPublisher := setupTestHandlerWithPublisher(t)

	// Insert default workflow
	_, err := handler.db.Exec(context.Background(),
		"INSERT INTO workflow (id, title, description, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)",
		"default-workflow-id", "Default Workflow", "Default workflow for testing", 1000, 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionCreate,
		Object: "project",
		Data: map[string]interface{}{
			// Missing required field 'name'
			"key": "TEST",
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

	// Verify no event was published
	assert.Equal(t, 0, mockPublisher.GetEventCount())
}

// TestProjectHandler_Modify_NoEventOnFailure tests that no event is published on modify failure
func TestProjectHandler_Modify_NoEventOnFailure(t *testing.T) {
	handler, mockPublisher := setupTestHandlerWithPublisher(t)

	reqBody := models.Request{
		Action: models.ActionModify,
		Object: "project",
		Data: map[string]interface{}{
			"id":    "non-existent-project",
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

	// Verify no event was published
	assert.Equal(t, 0, mockPublisher.GetEventCount())
}

// TestProjectHandler_Remove_NoEventOnFailure tests that no event is published on remove failure
func TestProjectHandler_Remove_NoEventOnFailure(t *testing.T) {
	handler, mockPublisher := setupTestHandlerWithPublisher(t)

	reqBody := models.Request{
		Action: models.ActionRemove,
		Object: "project",
		Data: map[string]interface{}{
			"id": "non-existent-project",
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

	// Verify no event was published
	assert.Equal(t, 0, mockPublisher.GetEventCount())
}
