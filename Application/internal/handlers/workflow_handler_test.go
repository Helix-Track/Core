package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"helixtrack.ru/core/internal/models"
)

// setupWorkflowTestHandler creates a test handler for workflow tests
func setupWorkflowTestHandler(t *testing.T) *Handler {
	return setupTestHandler(t)
}

// =============================================================================
// handleWorkflowCreate Tests
// =============================================================================

func TestWorkflowHandler_Create_Success(t *testing.T) {
	handler := setupWorkflowTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionWorkflowCreate,
		Data: map[string]interface{}{
			"title":       "Test Workflow",
			"description": "Test workflow description",
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

	workflow, ok := resp.Data["workflow"].(map[string]interface{})
	require.True(t, ok)
	assert.NotEmpty(t, workflow["id"])
	assert.Equal(t, "Test Workflow", workflow["title"])
	assert.Equal(t, "Test workflow description", workflow["description"])
	assert.NotEmpty(t, workflow["created"])
	assert.NotEmpty(t, workflow["modified"])
}

func TestWorkflowHandler_Create_MinimalFields(t *testing.T) {
	handler := setupWorkflowTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionWorkflowCreate,
		Data: map[string]interface{}{
			"title": "Minimal Workflow",
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

	workflow, ok := resp.Data["workflow"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "Minimal Workflow", workflow["title"])
}

func TestWorkflowHandler_Create_MissingTitle(t *testing.T) {
	handler := setupWorkflowTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionWorkflowCreate,
		Data: map[string]interface{}{
			"description": "Description without title",
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
	assert.Contains(t, resp.ErrorMessage, "title")
}

func TestWorkflowHandler_Create_Unauthorized(t *testing.T) {
	handler := setupWorkflowTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionWorkflowCreate,
		Data: map[string]interface{}{
			"title": "Test Workflow",
		},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	// No username set - should be unauthorized

	handler.DoAction(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeUnauthorized, resp.ErrorCode)
}

func TestWorkflowHandler_Create_EmptyTitle(t *testing.T) {
	handler := setupWorkflowTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionWorkflowCreate,
		Data: map[string]interface{}{
			"title": "",
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
}

// =============================================================================
// handleWorkflowRead Tests
// =============================================================================

func TestWorkflowHandler_Read_Success(t *testing.T) {
	handler := setupWorkflowTestHandler(t)

	// Create workflow first
	createReq := models.Request{
		Action: models.ActionWorkflowCreate,
		Data: map[string]interface{}{
			"title":       "Read Test Workflow",
			"description": "Description for read test",
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
	workflowData := createResp.Data["workflow"].(map[string]interface{})
	workflowID := workflowData["id"].(string)

	// Read the workflow
	readReq := models.Request{
		Action: models.ActionWorkflowRead,
		Data: map[string]interface{}{
			"id": workflowID,
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

	workflow := readResp.Data["workflow"].(map[string]interface{})
	assert.Equal(t, workflowID, workflow["id"])
	assert.Equal(t, "Read Test Workflow", workflow["title"])
	assert.Equal(t, "Description for read test", workflow["description"])
}

func TestWorkflowHandler_Read_MissingID(t *testing.T) {
	handler := setupWorkflowTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionWorkflowRead,
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
	assert.Contains(t, resp.ErrorMessage, "workflow ID")
}

func TestWorkflowHandler_Read_NotFound(t *testing.T) {
	handler := setupWorkflowTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionWorkflowRead,
		Data: map[string]interface{}{
			"id": "non-existent-workflow-id",
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
}

func TestWorkflowHandler_Read_Unauthorized(t *testing.T) {
	handler := setupWorkflowTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionWorkflowRead,
		Data: map[string]interface{}{
			"id": "some-id",
		},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	// No username set

	handler.DoAction(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// =============================================================================
// handleWorkflowList Tests
// =============================================================================

func TestWorkflowHandler_List_Empty(t *testing.T) {
	handler := setupWorkflowTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionWorkflowList,
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

	workflows, ok := resp.Data["workflows"].([]interface{})
	require.True(t, ok)
	assert.Equal(t, 0, len(workflows))
	assert.Equal(t, float64(0), resp.Data["count"])
}

func TestWorkflowHandler_List_Multiple(t *testing.T) {
	handler := setupWorkflowTestHandler(t)

	// Create 3 workflows
	workflowNames := []string{"Workflow A", "Workflow B", "Workflow C"}
	for _, name := range workflowNames {
		createReq := models.Request{
			Action: models.ActionWorkflowCreate,
			Data: map[string]interface{}{
				"title": name,
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

	// List workflows
	reqBody := models.Request{
		Action: models.ActionWorkflowList,
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

	workflows, ok := resp.Data["workflows"].([]interface{})
	require.True(t, ok)
	assert.Equal(t, 3, len(workflows))
	assert.Equal(t, float64(3), resp.Data["count"])

	// Verify ordering by title (ASC)
	firstWorkflow := workflows[0].(map[string]interface{})
	assert.Equal(t, "Workflow A", firstWorkflow["title"])
}

func TestWorkflowHandler_List_Unauthorized(t *testing.T) {
	handler := setupWorkflowTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionWorkflowList,
		Data:   map[string]interface{}{},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	// No username set

	handler.DoAction(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestWorkflowHandler_List_ExcludesDeleted(t *testing.T) {
	handler := setupWorkflowTestHandler(t)

	// Create 2 workflows
	var workflowID string
	for i := 1; i <= 2; i++ {
		createReq := models.Request{
			Action: models.ActionWorkflowCreate,
			Data: map[string]interface{}{
				"title": fmt.Sprintf("Workflow %d", i),
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

		if i == 1 {
			var createResp models.Response
			json.NewDecoder(wCreate.Body).Decode(&createResp)
			workflowData := createResp.Data["workflow"].(map[string]interface{})
			workflowID = workflowData["id"].(string)
		}
	}

	// Delete first workflow
	_, err := handler.db.Exec(context.Background(),
		"UPDATE workflow SET deleted = 1 WHERE id = ?", workflowID)
	require.NoError(t, err)

	// List workflows
	reqBody := models.Request{
		Action: models.ActionWorkflowList,
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

	var resp models.Response
	json.NewDecoder(w.Body).Decode(&resp)
	workflows := resp.Data["workflows"].([]interface{})

	// Should have only 1 workflow
	assert.Equal(t, 1, len(workflows))
}

// =============================================================================
// handleWorkflowModify Tests
// =============================================================================

func TestWorkflowHandler_Modify_Success(t *testing.T) {
	handler := setupWorkflowTestHandler(t)

	// Create workflow
	createReq := models.Request{
		Action: models.ActionWorkflowCreate,
		Data: map[string]interface{}{
			"title":       "Original Title",
			"description": "Original Description",
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
	workflowData := createResp.Data["workflow"].(map[string]interface{})
	workflowID := workflowData["id"].(string)

	// Modify workflow
	modifyReq := models.Request{
		Action: models.ActionWorkflowModify,
		Data: map[string]interface{}{
			"id":          workflowID,
			"title":       "Updated Title",
			"description": "Updated Description",
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
	assert.True(t, modifyResp.Data["updated"].(bool))
	assert.Equal(t, workflowID, modifyResp.Data["id"])
}

func TestWorkflowHandler_Modify_OnlyTitle(t *testing.T) {
	handler := setupWorkflowTestHandler(t)

	// Create workflow
	createReq := models.Request{
		Action: models.ActionWorkflowCreate,
		Data: map[string]interface{}{
			"title": "Original",
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
	workflowData := createResp.Data["workflow"].(map[string]interface{})
	workflowID := workflowData["id"].(string)

	// Modify only title
	modifyReq := models.Request{
		Action: models.ActionWorkflowModify,
		Data: map[string]interface{}{
			"id":    workflowID,
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

func TestWorkflowHandler_Modify_MissingID(t *testing.T) {
	handler := setupWorkflowTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionWorkflowModify,
		Data: map[string]interface{}{
			"title": "Updated Title",
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

func TestWorkflowHandler_Modify_NotFound(t *testing.T) {
	handler := setupWorkflowTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionWorkflowModify,
		Data: map[string]interface{}{
			"id":    "non-existent-id",
			"title": "Updated Title",
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
}

func TestWorkflowHandler_Modify_NoFieldsToUpdate(t *testing.T) {
	handler := setupWorkflowTestHandler(t)

	// Create workflow
	createReq := models.Request{
		Action: models.ActionWorkflowCreate,
		Data: map[string]interface{}{
			"title": "Test",
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
	workflowData := createResp.Data["workflow"].(map[string]interface{})
	workflowID := workflowData["id"].(string)

	// Try to modify with no fields
	modifyReq := models.Request{
		Action: models.ActionWorkflowModify,
		Data: map[string]interface{}{
			"id": workflowID,
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

	assert.Equal(t, http.StatusBadRequest, wModify.Code)
}

// =============================================================================
// handleWorkflowRemove Tests
// =============================================================================

func TestWorkflowHandler_Remove_Success(t *testing.T) {
	handler := setupWorkflowTestHandler(t)

	// Create workflow
	createReq := models.Request{
		Action: models.ActionWorkflowCreate,
		Data: map[string]interface{}{
			"title": "To Delete",
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
	workflowData := createResp.Data["workflow"].(map[string]interface{})
	workflowID := workflowData["id"].(string)

	// Remove workflow
	removeReq := models.Request{
		Action: models.ActionWorkflowRemove,
		Data: map[string]interface{}{
			"id": workflowID,
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
	assert.True(t, removeResp.Data["deleted"].(bool))
	assert.Equal(t, workflowID, removeResp.Data["id"])
}

func TestWorkflowHandler_Remove_MissingID(t *testing.T) {
	handler := setupWorkflowTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionWorkflowRemove,
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
}

func TestWorkflowHandler_Remove_NotFound(t *testing.T) {
	handler := setupWorkflowTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionWorkflowRemove,
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
}

func TestWorkflowHandler_Remove_Unauthorized(t *testing.T) {
	handler := setupWorkflowTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionWorkflowRemove,
		Data: map[string]interface{}{
			"id": "some-id",
		},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	// No username set

	handler.DoAction(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
