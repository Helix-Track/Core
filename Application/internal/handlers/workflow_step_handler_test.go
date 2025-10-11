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

// setupWorkflowStepTable creates the workflow_step table for testing
func setupWorkflowStepTable(t *testing.T, handler *Handler) {
	_, err := handler.db.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS workflow_step (
			id TEXT PRIMARY KEY,
			workflow_id TEXT NOT NULL,
			status_id TEXT NOT NULL,
			position INTEGER NOT NULL,
			created INTEGER NOT NULL,
			modified INTEGER NOT NULL,
			deleted INTEGER NOT NULL DEFAULT 0
		)
	`)
	require.NoError(t, err)
}

// setupWorkflowStepTestHandler creates a test handler with workflow step test data
func setupWorkflowStepTestHandler(t *testing.T) *Handler {
	handler := setupTestHandler(t)

	// Create all required tables
	setupWorkflowTable(t, handler)
	setupTicketStatusTable(t, handler)
	setupWorkflowStepTable(t, handler)

	// Insert test workflow
	_, err := handler.db.Exec(context.Background(),
		"INSERT INTO workflow (id, title, description, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)",
		"test-workflow-id", "Test Workflow", "Test workflow description", 1000, 1000, 0)
	require.NoError(t, err)

	// Insert test status
	_, err = handler.db.Exec(context.Background(),
		"INSERT INTO ticket_status (id, title, description, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)",
		"test-status-id", "In Progress", "Status description", 1000, 1000, 0)
	require.NoError(t, err)

	// Insert another test status for multiple step tests
	_, err = handler.db.Exec(context.Background(),
		"INSERT INTO ticket_status (id, title, description, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)",
		"test-status-2-id", "Done", "Completed status", 1000, 1000, 0)
	require.NoError(t, err)

	return handler
}

// TestWorkflowStepHandler_Create_Success tests successful workflow step creation
func TestWorkflowStepHandler_Create_Success(t *testing.T) {
	handler := setupWorkflowStepTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionWorkflowStepCreate,
		Data: map[string]interface{}{
			"workflowId": "test-workflow-id",
			"statusId":   "test-status-id",
			"position":   float64(1),
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

	workflowStep, ok := resp.Data["workflowStep"].(map[string]interface{})
	require.True(t, ok)
	assert.NotEmpty(t, workflowStep["id"])
	assert.Equal(t, "test-workflow-id", workflowStep["workflowId"])
	assert.Equal(t, "test-status-id", workflowStep["statusId"])
	assert.Equal(t, float64(1), workflowStep["position"])
}

// TestWorkflowStepHandler_Create_MultiplePositions tests creating multiple workflow steps with different positions
func TestWorkflowStepHandler_Create_MultiplePositions(t *testing.T) {
	handler := setupWorkflowStepTestHandler(t)

	// Insert additional statuses
	_, err := handler.db.Exec(context.Background(),
		"INSERT INTO ticket_status (id, title, description, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)",
		"open-status-id", "Open", "Open status", 1000, 1000, 0)
	require.NoError(t, err)

	_, err = handler.db.Exec(context.Background(),
		"INSERT INTO ticket_status (id, title, description, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)",
		"closed-status-id", "Closed", "Closed status", 1000, 1000, 0)
	require.NoError(t, err)

	// Create workflow steps at different positions
	steps := []struct {
		statusID string
		position int
	}{
		{"open-status-id", 1},
		{"test-status-id", 2},
		{"test-status-2-id", 3},
		{"closed-status-id", 4},
	}

	for _, step := range steps {
		reqBody := models.Request{
			Action: models.ActionWorkflowStepCreate,
			Data: map[string]interface{}{
				"workflowId": "test-workflow-id",
				"statusId":   step.statusID,
				"position":   float64(step.position),
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
	}
}

// TestWorkflowStepHandler_Create_MissingWorkflowId tests workflow step creation with missing workflow ID
func TestWorkflowStepHandler_Create_MissingWorkflowId(t *testing.T) {
	handler := setupWorkflowStepTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionWorkflowStepCreate,
		Data: map[string]interface{}{
			"statusId": "test-status-id",
			"position": float64(1),
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

// TestWorkflowStepHandler_Create_MissingStatusId tests workflow step creation with missing status ID
func TestWorkflowStepHandler_Create_MissingStatusId(t *testing.T) {
	handler := setupWorkflowStepTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionWorkflowStepCreate,
		Data: map[string]interface{}{
			"workflowId": "test-workflow-id",
			"position":   float64(1),
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

// TestWorkflowStepHandler_Create_MissingPosition tests workflow step creation with missing position
func TestWorkflowStepHandler_Create_MissingPosition(t *testing.T) {
	handler := setupWorkflowStepTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionWorkflowStepCreate,
		Data: map[string]interface{}{
			"workflowId": "test-workflow-id",
			"statusId":   "test-status-id",
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

// TestWorkflowStepHandler_Read_Success tests successful workflow step read
func TestWorkflowStepHandler_Read_Success(t *testing.T) {
	handler := setupWorkflowStepTestHandler(t)

	// Insert test workflow step
	_, err := handler.db.Exec(context.Background(),
		"INSERT INTO workflow_step (id, workflow_id, status_id, position, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?)",
		"test-step-id", "test-workflow-id", "test-status-id", 1, 1000, 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionWorkflowStepRead,
		Data: map[string]interface{}{
			"id": "test-step-id",
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

	workflowStep, ok := resp.Data["workflowStep"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "test-step-id", workflowStep["id"])
	assert.Equal(t, "test-workflow-id", workflowStep["workflowId"])
	assert.Equal(t, "test-status-id", workflowStep["statusId"])
	assert.Equal(t, float64(1), workflowStep["position"])
}

// TestWorkflowStepHandler_Read_NotFound tests reading non-existent workflow step
func TestWorkflowStepHandler_Read_NotFound(t *testing.T) {
	handler := setupWorkflowStepTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionWorkflowStepRead,
		Data: map[string]interface{}{
			"id": "non-existent-step",
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

// TestWorkflowStepHandler_List_Empty tests listing workflow steps when none exist
func TestWorkflowStepHandler_List_Empty(t *testing.T) {
	handler := setupWorkflowStepTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionWorkflowStepList,
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

	workflowSteps, ok := resp.Data["workflowSteps"].([]interface{})
	require.True(t, ok)
	assert.Empty(t, workflowSteps)
}

// TestWorkflowStepHandler_List_Multiple tests listing multiple workflow steps
func TestWorkflowStepHandler_List_Multiple(t *testing.T) {
	handler := setupWorkflowStepTestHandler(t)

	// Insert multiple workflow steps
	steps := []struct {
		id       string
		position int
	}{
		{"step-1", 1},
		{"step-2", 2},
		{"step-3", 3},
	}

	for _, step := range steps {
		_, err := handler.db.Exec(context.Background(),
			"INSERT INTO workflow_step (id, workflow_id, status_id, position, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?)",
			step.id, "test-workflow-id", "test-status-id", step.position, 1000, 1000, 0)
		require.NoError(t, err)
	}

	reqBody := models.Request{
		Action: models.ActionWorkflowStepList,
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

	workflowSteps, ok := resp.Data["workflowSteps"].([]interface{})
	require.True(t, ok)
	assert.Len(t, workflowSteps, 3)
}

// TestWorkflowStepHandler_List_FilterByWorkflow tests listing workflow steps filtered by workflow ID
func TestWorkflowStepHandler_List_FilterByWorkflow(t *testing.T) {
	handler := setupWorkflowStepTestHandler(t)

	// Insert another workflow
	_, err := handler.db.Exec(context.Background(),
		"INSERT INTO workflow (id, title, description, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)",
		"workflow-2-id", "Workflow 2", "Second workflow", 1000, 1000, 0)
	require.NoError(t, err)

	// Insert workflow steps for different workflows
	_, err = handler.db.Exec(context.Background(),
		"INSERT INTO workflow_step (id, workflow_id, status_id, position, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?)",
		"step-1", "test-workflow-id", "test-status-id", 1, 1000, 1000, 0)
	require.NoError(t, err)

	_, err = handler.db.Exec(context.Background(),
		"INSERT INTO workflow_step (id, workflow_id, status_id, position, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?)",
		"step-2", "workflow-2-id", "test-status-id", 1, 1000, 1000, 0)
	require.NoError(t, err)

	// Filter by test-workflow-id
	reqBody := models.Request{
		Action: models.ActionWorkflowStepList,
		Data: map[string]interface{}{
			"workflowId": "test-workflow-id",
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

	workflowSteps, ok := resp.Data["workflowSteps"].([]interface{})
	require.True(t, ok)
	assert.Len(t, workflowSteps, 1)

	step := workflowSteps[0].(map[string]interface{})
	assert.Equal(t, "step-1", step["id"])
	assert.Equal(t, "test-workflow-id", step["workflowId"])
}

// TestWorkflowStepHandler_List_OrderedByPosition tests that steps are ordered by position
func TestWorkflowStepHandler_List_OrderedByPosition(t *testing.T) {
	handler := setupWorkflowStepTestHandler(t)

	// Insert workflow steps in random order
	steps := []struct {
		id       string
		position int
	}{
		{"step-3", 3},
		{"step-1", 1},
		{"step-2", 2},
	}

	for _, step := range steps {
		_, err := handler.db.Exec(context.Background(),
			"INSERT INTO workflow_step (id, workflow_id, status_id, position, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?)",
			step.id, "test-workflow-id", "test-status-id", step.position, 1000, 1000, 0)
		require.NoError(t, err)
	}

	reqBody := models.Request{
		Action: models.ActionWorkflowStepList,
		Data: map[string]interface{}{
			"workflowId": "test-workflow-id",
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
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)

	workflowSteps, ok := resp.Data["workflowSteps"].([]interface{})
	require.True(t, ok)
	assert.Len(t, workflowSteps, 3)

	// Verify ordering by position
	assert.Equal(t, "step-1", workflowSteps[0].(map[string]interface{})["id"])
	assert.Equal(t, "step-2", workflowSteps[1].(map[string]interface{})["id"])
	assert.Equal(t, "step-3", workflowSteps[2].(map[string]interface{})["id"])
}

// TestWorkflowStepHandler_Modify_Success tests successful workflow step modification
func TestWorkflowStepHandler_Modify_Success(t *testing.T) {
	handler := setupWorkflowStepTestHandler(t)

	// Insert test workflow step
	_, err := handler.db.Exec(context.Background(),
		"INSERT INTO workflow_step (id, workflow_id, status_id, position, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?)",
		"test-step-id", "test-workflow-id", "test-status-id", 1, 1000, 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionWorkflowStepModify,
		Data: map[string]interface{}{
			"id":       "test-step-id",
			"statusId": "test-status-2-id",
			"position": float64(2),
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
	var statusID string
	var position int
	err = handler.db.QueryRow(context.Background(),
		"SELECT status_id, position FROM workflow_step WHERE id = ?",
		"test-step-id").Scan(&statusID, &position)
	require.NoError(t, err)
	assert.Equal(t, "test-status-2-id", statusID)
	assert.Equal(t, 2, position)
}

// TestWorkflowStepHandler_Modify_PositionOnly tests modifying only the position
func TestWorkflowStepHandler_Modify_PositionOnly(t *testing.T) {
	handler := setupWorkflowStepTestHandler(t)

	// Insert test workflow step
	_, err := handler.db.Exec(context.Background(),
		"INSERT INTO workflow_step (id, workflow_id, status_id, position, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?)",
		"test-step-id", "test-workflow-id", "test-status-id", 1, 1000, 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionWorkflowStepModify,
		Data: map[string]interface{}{
			"id":       "test-step-id",
			"position": float64(5),
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
	var position int
	err = handler.db.QueryRow(context.Background(),
		"SELECT position FROM workflow_step WHERE id = ?",
		"test-step-id").Scan(&position)
	require.NoError(t, err)
	assert.Equal(t, 5, position)
}

// TestWorkflowStepHandler_Modify_NotFound tests modifying non-existent workflow step
func TestWorkflowStepHandler_Modify_NotFound(t *testing.T) {
	handler := setupWorkflowStepTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionWorkflowStepModify,
		Data: map[string]interface{}{
			"id":       "non-existent-step",
			"position": float64(2),
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

// TestWorkflowStepHandler_Modify_NoFieldsToUpdate tests modifying without providing any fields
func TestWorkflowStepHandler_Modify_NoFieldsToUpdate(t *testing.T) {
	handler := setupWorkflowStepTestHandler(t)

	// Insert test workflow step
	_, err := handler.db.Exec(context.Background(),
		"INSERT INTO workflow_step (id, workflow_id, status_id, position, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?)",
		"test-step-id", "test-workflow-id", "test-status-id", 1, 1000, 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionWorkflowStepModify,
		Data: map[string]interface{}{
			"id": "test-step-id",
			// No fields to update
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
	err = json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeMissingData, resp.ErrorCode)
}

// TestWorkflowStepHandler_Remove_Success tests successful workflow step deletion
func TestWorkflowStepHandler_Remove_Success(t *testing.T) {
	handler := setupWorkflowStepTestHandler(t)

	// Insert test workflow step
	_, err := handler.db.Exec(context.Background(),
		"INSERT INTO workflow_step (id, workflow_id, status_id, position, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?)",
		"test-step-id", "test-workflow-id", "test-status-id", 1, 1000, 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionWorkflowStepRemove,
		Data: map[string]interface{}{
			"id": "test-step-id",
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
		"SELECT deleted FROM workflow_step WHERE id = ?",
		"test-step-id").Scan(&deleted)
	require.NoError(t, err)
	assert.True(t, deleted)
}

// TestWorkflowStepHandler_Remove_NotFound tests deleting non-existent workflow step
func TestWorkflowStepHandler_Remove_NotFound(t *testing.T) {
	handler := setupWorkflowStepTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionWorkflowStepRemove,
		Data: map[string]interface{}{
			"id": "non-existent-step",
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

// TestWorkflowStepHandler_CRUD_FullCycle tests complete workflow step lifecycle
func TestWorkflowStepHandler_CRUD_FullCycle(t *testing.T) {
	handler := setupWorkflowStepTestHandler(t)

	// 1. Create workflow step
	createReq := models.Request{
		Action: models.ActionWorkflowStepCreate,
		Data: map[string]interface{}{
			"workflowId": "test-workflow-id",
			"statusId":   "test-status-id",
			"position":   float64(1),
		},
	}
	body, _ := json.Marshal(createReq)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")
	c.Set("request", &createReq)
	handler.DoAction(c)

	var createResp models.Response
	json.NewDecoder(w.Body).Decode(&createResp)
	stepData := createResp.Data["workflowStep"].(map[string]interface{})
	stepID := stepData["id"].(string)

	// 2. Read workflow step
	readReq := models.Request{
		Action: models.ActionWorkflowStepRead,
		Data:   map[string]interface{}{"id": stepID},
	}
	body, _ = json.Marshal(readReq)
	req = httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")
	c.Set("request", &readReq)
	handler.DoAction(c)
	assert.Equal(t, http.StatusOK, w.Code)

	// 3. Modify workflow step
	modifyReq := models.Request{
		Action: models.ActionWorkflowStepModify,
		Data: map[string]interface{}{
			"id":       stepID,
			"position": float64(2),
		},
	}
	body, _ = json.Marshal(modifyReq)
	req = httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")
	c.Set("request", &modifyReq)
	handler.DoAction(c)
	assert.Equal(t, http.StatusOK, w.Code)

	// 4. Delete workflow step
	deleteReq := models.Request{
		Action: models.ActionWorkflowStepRemove,
		Data:   map[string]interface{}{"id": stepID},
	}
	body, _ = json.Marshal(deleteReq)
	req = httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")
	c.Set("request", &deleteReq)
	handler.DoAction(c)
	assert.Equal(t, http.StatusOK, w.Code)

	// 5. Verify deletion - step should not be found
	readReq = models.Request{
		Action: models.ActionWorkflowStepRead,
		Data:   map[string]interface{}{"id": stepID},
	}
	body, _ = json.Marshal(readReq)
	req = httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")
	c.Set("request", &readReq)
	handler.DoAction(c)
	assert.Equal(t, http.StatusNotFound, w.Code)
}
