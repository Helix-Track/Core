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

// setupTicketTestHandler creates a test handler with required test data
func setupTicketTestHandler(t *testing.T) (*Handler, string) {
	handler := setupTestHandler(t)

	// Insert default workflow
	_, err := handler.db.Exec(context.Background(),
		"INSERT INTO workflow (id, title, description, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)",
		"test-workflow-id", "Test Workflow", "Test workflow", 1000, 1000, 0)
	require.NoError(t, err)

	// Insert ticket types
	ticketTypes := []string{"task", "bug", "story", "epic"}
	for _, tt := range ticketTypes {
		_, err := handler.db.Exec(context.Background(),
			"INSERT INTO ticket_type (id, title, description, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)",
			"type-"+tt, tt, "Type: "+tt, 1000, 1000, 0)
		require.NoError(t, err)
	}

	// Insert ticket statuses
	statuses := []string{"open", "in_progress", "done", "closed"}
	for _, status := range statuses {
		_, err := handler.db.Exec(context.Background(),
			"INSERT INTO ticket_status (id, title, description, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)",
			"status-"+status, status, "Status: "+status, 1000, 1000, 0)
		require.NoError(t, err)
	}

	// Create a test project
	projectID := "test-project-id"
	_, err = handler.db.Exec(context.Background(),
		"INSERT INTO project (id, identifier, title, description, workflow_id, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		projectID, "TEST", "Test Project", "A test project", "test-workflow-id", 1000, 1000, 0)
	require.NoError(t, err)

	return handler, projectID
}

// =============================================================================
// handleCreateTicket Tests
// =============================================================================

func TestTicketHandler_Create_Success(t *testing.T) {
	handler, projectID := setupTicketTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionCreate,
		Object: "ticket",
		Data: map[string]interface{}{
			"project_id":  projectID,
			"title":       "Test Ticket",
			"description": "Test ticket description",
			"type":        "task",
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

	ticket, ok := resp.Data["ticket"].(map[string]interface{})
	require.True(t, ok)
	assert.NotEmpty(t, ticket["id"])
	assert.Equal(t, float64(1), ticket["ticket_number"]) // First ticket
	assert.Equal(t, "Test Ticket", ticket["title"])
	assert.Equal(t, "Test ticket description", ticket["description"])
	assert.Equal(t, "task", ticket["type"])
	assert.Equal(t, "open", ticket["status"])
	assert.Equal(t, projectID, ticket["project_id"])
}

func TestTicketHandler_Create_MinimalFields(t *testing.T) {
	handler, projectID := setupTicketTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionCreate,
		Object: "ticket",
		Data: map[string]interface{}{
			"project_id": projectID,
			"title":      "Minimal Ticket",
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

	ticket := resp.Data["ticket"].(map[string]interface{})
	assert.Equal(t, "task", ticket["type"]) // Default type
}

func TestTicketHandler_Create_MissingProjectID(t *testing.T) {
	handler, _ := setupTicketTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionCreate,
		Object: "ticket",
		Data: map[string]interface{}{
			"title": "Test Ticket",
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
	assert.Contains(t, resp.ErrorMessage, "project_id")
}

func TestTicketHandler_Create_MissingTitle(t *testing.T) {
	handler, projectID := setupTicketTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionCreate,
		Object: "ticket",
		Data: map[string]interface{}{
			"project_id": projectID,
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

func TestTicketHandler_Create_InvalidTicketType(t *testing.T) {
	handler, projectID := setupTicketTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionCreate,
		Object: "ticket",
		Data: map[string]interface{}{
			"project_id": projectID,
			"title":      "Test Ticket",
			"type":       "invalid_type",
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
	assert.Equal(t, models.ErrorCodeInvalidData, resp.ErrorCode)
	assert.Contains(t, resp.ErrorMessage, "ticket type")
}

func TestTicketHandler_Create_TicketNumberIncrement(t *testing.T) {
	handler, projectID := setupTicketTestHandler(t)

	// Create 3 tickets and verify numbers increment
	for i := 1; i <= 3; i++ {
		reqBody := models.Request{
			Action: models.ActionCreate,
			Object: "ticket",
			Data: map[string]interface{}{
				"project_id": projectID,
				"title":      fmt.Sprintf("Ticket %d", i),
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

		var resp models.Response
		json.NewDecoder(w.Body).Decode(&resp)
		ticket := resp.Data["ticket"].(map[string]interface{})
		assert.Equal(t, float64(i), ticket["ticket_number"])
	}
}

func TestTicketHandler_Create_DifferentTypes(t *testing.T) {
	handler, projectID := setupTicketTestHandler(t)

	ticketTypes := []string{"task", "bug", "story", "epic"}

	for _, ticketType := range ticketTypes {
		reqBody := models.Request{
			Action: models.ActionCreate,
			Object: "ticket",
			Data: map[string]interface{}{
				"project_id": projectID,
				"title":      fmt.Sprintf("Test %s", ticketType),
				"type":       ticketType,
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
		json.NewDecoder(w.Body).Decode(&resp)
		ticket := resp.Data["ticket"].(map[string]interface{})
		assert.Equal(t, ticketType, ticket["type"])
	}
}

// =============================================================================
// handleModifyTicket Tests
// =============================================================================

func TestTicketHandler_Modify_Success(t *testing.T) {
	handler, projectID := setupTicketTestHandler(t)

	// Create ticket first
	createReq := models.Request{
		Action: models.ActionCreate,
		Object: "ticket",
		Data: map[string]interface{}{
			"project_id": projectID,
			"title":      "Original Title",
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
	ticketID := createResp.Data["ticket"].(map[string]interface{})["id"].(string)

	// Modify ticket
	modifyReq := models.Request{
		Action: models.ActionModify,
		Object: "ticket",
		Data: map[string]interface{}{
			"id":          ticketID,
			"title":       "Updated Title",
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

	modifiedTicket := modifyResp.Data["ticket"].(map[string]interface{})
	assert.Equal(t, ticketID, modifiedTicket["id"])
	assert.True(t, modifiedTicket["updated"].(bool))
}

func TestTicketHandler_Modify_StatusChange(t *testing.T) {
	handler, projectID := setupTicketTestHandler(t)

	// Create ticket
	createReq := models.Request{
		Action: models.ActionCreate,
		Object: "ticket",
		Data: map[string]interface{}{
			"project_id": projectID,
			"title":      "Status Test",
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
	ticketID := createResp.Data["ticket"].(map[string]interface{})["id"].(string)

	// Change status to in_progress
	modifyReq := models.Request{
		Action: models.ActionModify,
		Object: "ticket",
		Data: map[string]interface{}{
			"id":     ticketID,
			"status": "in_progress",
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

func TestTicketHandler_Modify_MissingID(t *testing.T) {
	handler, _ := setupTicketTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionModify,
		Object: "ticket",
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
	assert.Contains(t, resp.ErrorMessage, "ticket ID")
}

func TestTicketHandler_Modify_OnlyTitle(t *testing.T) {
	handler, projectID := setupTicketTestHandler(t)

	// Create ticket
	createReq := models.Request{
		Action: models.ActionCreate,
		Object: "ticket",
		Data: map[string]interface{}{
			"project_id":  projectID,
			"title":       "Original",
			"description": "Keep this",
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
	ticketID := createResp.Data["ticket"].(map[string]interface{})["id"].(string)

	// Update only title
	modifyReq := models.Request{
		Action: models.ActionModify,
		Object: "ticket",
		Data: map[string]interface{}{
			"id":    ticketID,
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
// handleRemoveTicket Tests
// =============================================================================

func TestTicketHandler_Remove_Success(t *testing.T) {
	handler, projectID := setupTicketTestHandler(t)

	// Create ticket
	createReq := models.Request{
		Action: models.ActionCreate,
		Object: "ticket",
		Data: map[string]interface{}{
			"project_id": projectID,
			"title":      "To Delete",
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
	ticketID := createResp.Data["ticket"].(map[string]interface{})["id"].(string)

	// Remove ticket
	removeReq := models.Request{
		Action: models.ActionRemove,
		Object: "ticket",
		Data: map[string]interface{}{
			"id": ticketID,
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

	removedTicket := removeResp.Data["ticket"].(map[string]interface{})
	assert.Equal(t, ticketID, removedTicket["id"])
	assert.True(t, removedTicket["deleted"].(bool))
}

func TestTicketHandler_Remove_MissingID(t *testing.T) {
	handler, _ := setupTicketTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionRemove,
		Object: "ticket",
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
	assert.Contains(t, resp.ErrorMessage, "ticket ID")
}

// =============================================================================
// handleReadTicket Tests
// =============================================================================

func TestTicketHandler_Read_Success(t *testing.T) {
	handler, projectID := setupTicketTestHandler(t)

	// Create ticket
	createReq := models.Request{
		Action: models.ActionCreate,
		Object: "ticket",
		Data: map[string]interface{}{
			"project_id":  projectID,
			"title":       "Read Test",
			"description": "Test description",
			"type":        "bug",
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
	ticketID := createResp.Data["ticket"].(map[string]interface{})["id"].(string)

	// Read ticket
	readReq := models.Request{
		Action: models.ActionRead,
		Object: "ticket",
		Data: map[string]interface{}{
			"id": ticketID,
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

	ticket := readResp.Data["ticket"].(map[string]interface{})
	assert.Equal(t, ticketID, ticket["id"])
	assert.Equal(t, float64(1), ticket["ticket_number"])
	assert.Equal(t, "Read Test", ticket["title"])
	assert.Equal(t, "Test description", ticket["description"])
	assert.Equal(t, "bug", ticket["type"])
	assert.Equal(t, "open", ticket["status"])
	assert.Equal(t, projectID, ticket["project_id"])
}

func TestTicketHandler_Read_MissingID(t *testing.T) {
	handler, _ := setupTicketTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionRead,
		Object: "ticket",
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
	assert.Contains(t, resp.ErrorMessage, "ticket ID")
}

func TestTicketHandler_Read_NotFound(t *testing.T) {
	handler, _ := setupTicketTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionRead,
		Object: "ticket",
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

func TestTicketHandler_Read_DeletedTicket(t *testing.T) {
	handler, projectID := setupTicketTestHandler(t)

	// Create and delete ticket
	createReq := models.Request{
		Action: models.ActionCreate,
		Object: "ticket",
		Data: map[string]interface{}{
			"project_id": projectID,
			"title":      "To Delete",
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
	ticketID := createResp.Data["ticket"].(map[string]interface{})["id"].(string)

	// Delete ticket
	removeReq := models.Request{
		Action: models.ActionRemove,
		Object: "ticket",
		Data: map[string]interface{}{
			"id": ticketID,
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

	// Try to read deleted ticket
	readReq := models.Request{
		Action: models.ActionRead,
		Object: "ticket",
		Data: map[string]interface{}{
			"id": ticketID,
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
// handleListTickets Tests
// =============================================================================

func TestTicketHandler_List_Empty(t *testing.T) {
	handler, _ := setupTicketTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionList,
		Object: "ticket",
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

func TestTicketHandler_List_Multiple(t *testing.T) {
	handler, projectID := setupTicketTestHandler(t)

	// Create 3 tickets
	for i := 1; i <= 3; i++ {
		createReq := models.Request{
			Action: models.ActionCreate,
			Object: "ticket",
			Data: map[string]interface{}{
				"project_id": projectID,
				"title":      fmt.Sprintf("Ticket %d", i),
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

	// List tickets
	reqBody := models.Request{
		Action: models.ActionList,
		Object: "ticket",
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
}

func TestTicketHandler_List_FilterByProject(t *testing.T) {
	handler, projectID := setupTicketTestHandler(t)

	// Create another project
	projectID2 := "test-project-id-2"
	_, err := handler.db.Exec(context.Background(),
		"INSERT INTO project (id, identifier, title, description, workflow_id, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		projectID2, "TEST2", "Test Project 2", "Second project", "test-workflow-id", 1000, 1000, 0)
	require.NoError(t, err)

	// Create tickets in both projects
	for i := 1; i <= 2; i++ {
		createReq := models.Request{
			Action: models.ActionCreate,
			Object: "ticket",
			Data: map[string]interface{}{
				"project_id": projectID,
				"title":      fmt.Sprintf("Project1 Ticket %d", i),
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

	createReq := models.Request{
		Action: models.ActionCreate,
		Object: "ticket",
		Data: map[string]interface{}{
			"project_id": projectID2,
			"title":      "Project2 Ticket",
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

	// List tickets filtered by first project
	reqBody := models.Request{
		Action: models.ActionList,
		Object: "ticket",
		Data: map[string]interface{}{
			"project_id": projectID,
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

	var resp models.Response
	err = json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)

	items := resp.Data["items"].([]interface{})
	assert.Equal(t, 2, len(items)) // Only 2 tickets from project 1

	// Verify all tickets belong to project 1
	for _, item := range items {
		ticket := item.(map[string]interface{})
		assert.Equal(t, projectID, ticket["project_id"])
	}
}

func TestTicketHandler_List_ExcludesDeleted(t *testing.T) {
	handler, projectID := setupTicketTestHandler(t)

	// Create 2 tickets
	var ticketID string
	for i := 1; i <= 2; i++ {
		createReq := models.Request{
			Action: models.ActionCreate,
			Object: "ticket",
			Data: map[string]interface{}{
				"project_id": projectID,
				"title":      fmt.Sprintf("Ticket %d", i),
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
			ticketID = createResp.Data["ticket"].(map[string]interface{})["id"].(string)
		}
	}

	// Delete first ticket
	removeReq := models.Request{
		Action: models.ActionRemove,
		Object: "ticket",
		Data: map[string]interface{}{
			"id": ticketID,
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

	// List tickets
	listReq := models.Request{
		Action: models.ActionList,
		Object: "ticket",
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
	items := listResp.Data["items"].([]interface{})

	// Should have only 1 ticket (deleted one excluded)
	assert.Equal(t, 1, len(items))
}

// =============================================================================
// Event Publishing Tests
// =============================================================================

// setupTicketTestHandlerWithPublisher creates a test handler with mock event publisher and test data
func setupTicketTestHandlerWithPublisher(t *testing.T) (*Handler, *MockEventPublisher, string) {
	handler, mockPublisher := setupTestHandlerWithPublisher(t)

	// Insert default workflow
	_, err := handler.db.Exec(context.Background(),
		"INSERT INTO workflow (id, title, description, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)",
		"test-workflow-id", "Test Workflow", "Test workflow", 1000, 1000, 0)
	require.NoError(t, err)

	// Insert ticket types
	ticketTypes := []string{"task", "bug", "story", "epic"}
	for _, tt := range ticketTypes {
		_, err := handler.db.Exec(context.Background(),
			"INSERT INTO ticket_type (id, title, description, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)",
			"type-"+tt, tt, "Type: "+tt, 1000, 1000, 0)
		require.NoError(t, err)
	}

	// Insert ticket statuses
	statuses := []string{"open", "in_progress", "done", "closed"}
	for _, status := range statuses {
		_, err := handler.db.Exec(context.Background(),
			"INSERT INTO ticket_status (id, title, description, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)",
			"status-"+status, status, "Status: "+status, 1000, 1000, 0)
		require.NoError(t, err)
	}

	// Create a test project
	projectID := "test-project-id"
	_, err = handler.db.Exec(context.Background(),
		"INSERT INTO project (id, identifier, title, description, workflow_id, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		projectID, "TEST", "Test Project", "A test project", "test-workflow-id", 1000, 1000, 0)
	require.NoError(t, err)

	return handler, mockPublisher, projectID
}

// TestTicketHandler_Create_PublishesEvent tests that ticket creation publishes an event
func TestTicketHandler_Create_PublishesEvent(t *testing.T) {
	handler, mockPublisher, projectID := setupTicketTestHandlerWithPublisher(t)

	reqBody := models.Request{
		Action: models.ActionCreate,
		Object: "ticket",
		Data: map[string]interface{}{
			"project_id":  projectID,
			"title":       "Event Test Ticket",
			"description": "Testing event publishing",
			"type":        "bug",
			"priority":    "high",
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
	assert.Equal(t, "ticket", lastCall.Object)
	assert.Equal(t, "testuser", lastCall.Username)
	assert.NotEmpty(t, lastCall.EntityID)

	// Verify event data
	assert.Equal(t, "Event Test Ticket", lastCall.Data["title"])
	assert.Equal(t, "Testing event publishing", lastCall.Data["description"])
	assert.Equal(t, "bug", lastCall.Data["type"])
	assert.Equal(t, "high", lastCall.Data["priority"])
	assert.Equal(t, "open", lastCall.Data["status"])
	assert.Equal(t, projectID, lastCall.Data["project_id"])

	// Verify project-based context
	assert.Equal(t, projectID, lastCall.Context.ProjectID)
	assert.Contains(t, lastCall.Context.Permissions, "READ")
}

// TestTicketHandler_Modify_PublishesEvent tests that ticket modification publishes an event
func TestTicketHandler_Modify_PublishesEvent(t *testing.T) {
	handler, mockPublisher, projectID := setupTicketTestHandlerWithPublisher(t)

	// Create ticket first
	createReq := models.Request{
		Action: models.ActionCreate,
		Object: "ticket",
		Data: map[string]interface{}{
			"project_id":  projectID,
			"title":       "Original Title",
			"description": "Original description",
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
	ticketID := createResp.Data["ticket"].(map[string]interface{})["id"].(string)

	// Reset mock publisher to clear create event
	mockPublisher.Reset()

	// Modify ticket
	modifyReq := models.Request{
		Action: models.ActionModify,
		Object: "ticket",
		Data: map[string]interface{}{
			"id":          ticketID,
			"title":       "Updated Title",
			"description": "Updated description",
			"status":      "in_progress",
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

	// Verify event was published
	assert.Equal(t, 1, mockPublisher.GetEventCount())
	lastCall := mockPublisher.GetLastEntityCall()
	require.NotNil(t, lastCall)

	// Verify event details
	assert.Equal(t, models.ActionModify, lastCall.Action)
	assert.Equal(t, "ticket", lastCall.Object)
	assert.Equal(t, ticketID, lastCall.EntityID)
	assert.Equal(t, "testuser", lastCall.Username)

	// Verify event data
	assert.Equal(t, ticketID, lastCall.Data["id"])
	assert.Equal(t, "Updated Title", lastCall.Data["title"])
	assert.Equal(t, "Updated description", lastCall.Data["description"])
	assert.Equal(t, "in_progress", lastCall.Data["status"])

	// Verify project-based context
	assert.Equal(t, projectID, lastCall.Context.ProjectID)
	assert.Contains(t, lastCall.Context.Permissions, "READ")
}

// TestTicketHandler_Remove_PublishesEvent tests that ticket deletion publishes an event
func TestTicketHandler_Remove_PublishesEvent(t *testing.T) {
	handler, mockPublisher, projectID := setupTicketTestHandlerWithPublisher(t)

	// Create ticket first
	createReq := models.Request{
		Action: models.ActionCreate,
		Object: "ticket",
		Data: map[string]interface{}{
			"project_id": projectID,
			"title":      "Ticket to Delete",
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
	ticketID := createResp.Data["ticket"].(map[string]interface{})["id"].(string)

	// Reset mock publisher to clear create event
	mockPublisher.Reset()

	// Remove ticket
	removeReq := models.Request{
		Action: models.ActionRemove,
		Object: "ticket",
		Data: map[string]interface{}{
			"id": ticketID,
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

	// Verify event was published
	assert.Equal(t, 1, mockPublisher.GetEventCount())
	lastCall := mockPublisher.GetLastEntityCall()
	require.NotNil(t, lastCall)

	// Verify event details
	assert.Equal(t, models.ActionRemove, lastCall.Action)
	assert.Equal(t, "ticket", lastCall.Object)
	assert.Equal(t, ticketID, lastCall.EntityID)
	assert.Equal(t, "testuser", lastCall.Username)

	// Verify event data
	assert.Equal(t, ticketID, lastCall.Data["id"])
	assert.Equal(t, projectID, lastCall.Data["project_id"])

	// Verify project-based context
	assert.Equal(t, projectID, lastCall.Context.ProjectID)
	assert.Contains(t, lastCall.Context.Permissions, "READ")
}

// TestTicketHandler_Create_NoEventOnFailure tests that no event is published on create failure
func TestTicketHandler_Create_NoEventOnFailure(t *testing.T) {
	handler, mockPublisher, _ := setupTicketTestHandlerWithPublisher(t)

	reqBody := models.Request{
		Action: models.ActionCreate,
		Object: "ticket",
		Data: map[string]interface{}{
			// Missing required field 'project_id'
			"title": "Test Ticket",
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

// TestTicketHandler_Modify_NoEventOnFailure tests that no event is published on modify failure
func TestTicketHandler_Modify_NoEventOnFailure(t *testing.T) {
	handler, mockPublisher, _ := setupTicketTestHandlerWithPublisher(t)

	reqBody := models.Request{
		Action: models.ActionModify,
		Object: "ticket",
		Data: map[string]interface{}{
			"id":    "non-existent-ticket-id",
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

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	// Verify no event was published
	assert.Equal(t, 0, mockPublisher.GetEventCount())
}

// TestTicketHandler_Remove_NoEventOnFailure tests that no event is published on remove failure
func TestTicketHandler_Remove_NoEventOnFailure(t *testing.T) {
	handler, mockPublisher, _ := setupTicketTestHandlerWithPublisher(t)

	reqBody := models.Request{
		Action: models.ActionRemove,
		Object: "ticket",
		Data: map[string]interface{}{
			"id": "non-existent-ticket-id",
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
