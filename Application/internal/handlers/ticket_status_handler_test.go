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

// TestTicketStatusHandler_Create_Success tests successful ticket status creation
func TestTicketStatusHandler_Create_Success(t *testing.T) {
	handler := setupTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionTicketStatusCreate,
		Data: map[string]interface{}{
			"title":       "In Progress",
			"description": "Work is currently in progress",
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

	ticketStatus, ok := resp.Data["ticketStatus"].(map[string]interface{})
	require.True(t, ok)
	assert.NotEmpty(t, ticketStatus["id"])
	assert.Equal(t, "In Progress", ticketStatus["title"])
	assert.Equal(t, "Work is currently in progress", ticketStatus["description"])
}

// TestTicketStatusHandler_Create_MinimalFields tests ticket status creation with only required fields
func TestTicketStatusHandler_Create_MinimalFields(t *testing.T) {
	handler := setupTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionTicketStatusCreate,
		Data: map[string]interface{}{
			"title": "Open",
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

	ticketStatus, ok := resp.Data["ticketStatus"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "Open", ticketStatus["title"])
}

// TestTicketStatusHandler_Create_MissingTitle tests ticket status creation with missing title
func TestTicketStatusHandler_Create_MissingTitle(t *testing.T) {
	handler := setupTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionTicketStatusCreate,
		Data: map[string]interface{}{
			"description": "Some description",
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

// TestTicketStatusHandler_Create_MultipleStatuses tests creating multiple common ticket statuses
func TestTicketStatusHandler_Create_MultipleStatuses(t *testing.T) {
	handler := setupTestHandler(t)

	statuses := []struct {
		title       string
		description string
	}{
		{"Open", "Ticket is open and awaiting assignment"},
		{"In Progress", "Work is currently in progress"},
		{"Code Review", "Code is under review"},
		{"Testing", "Ticket is being tested"},
		{"Done", "Work is completed"},
		{"Closed", "Ticket is closed"},
	}

	for _, status := range statuses {
		reqBody := models.Request{
			Action: models.ActionTicketStatusCreate,
			Data: map[string]interface{}{
				"title":       status.title,
				"description": status.description,
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
	}
}

// TestTicketStatusHandler_Read_Success tests successful ticket status read
func TestTicketStatusHandler_Read_Success(t *testing.T) {
	handler := setupTestHandler(t)

	// Insert test ticket status
	_, err := handler.db.Exec(context.Background(),
		"INSERT INTO ticket_status (id, title, description, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)",
		"test-status-id", "In Progress", "Work in progress", 1000, 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionTicketStatusRead,
		Data: map[string]interface{}{
			"id": "test-status-id",
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

	ticketStatus, ok := resp.Data["ticketStatus"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "test-status-id", ticketStatus["id"])
	assert.Equal(t, "In Progress", ticketStatus["title"])
	assert.Equal(t, "Work in progress", ticketStatus["description"])
}

// TestTicketStatusHandler_Read_NotFound tests reading non-existent ticket status
func TestTicketStatusHandler_Read_NotFound(t *testing.T) {
	handler := setupTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionTicketStatusRead,
		Data: map[string]interface{}{
			"id": "non-existent-status",
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

// TestTicketStatusHandler_Read_MissingId tests reading without providing ID
func TestTicketStatusHandler_Read_MissingId(t *testing.T) {
	handler := setupTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionTicketStatusRead,
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

// TestTicketStatusHandler_List_Empty tests listing ticket statuses when none exist
func TestTicketStatusHandler_List_Empty(t *testing.T) {
	handler := setupTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionTicketStatusList,
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

	ticketStatuses, ok := resp.Data["ticketStatuses"].([]interface{})
	require.True(t, ok)
	assert.Empty(t, ticketStatuses)
}

// TestTicketStatusHandler_List_Multiple tests listing multiple ticket statuses
func TestTicketStatusHandler_List_Multiple(t *testing.T) {
	handler := setupTestHandler(t)

	// Insert multiple ticket statuses
	statuses := []struct {
		id    string
		title string
	}{
		{"status-1", "Open"},
		{"status-2", "In Progress"},
		{"status-3", "Done"},
	}

	for _, status := range statuses {
		_, err := handler.db.Exec(context.Background(),
			"INSERT INTO ticket_status (id, title, description, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)",
			status.id, status.title, "Description for "+status.title, 1000, 1000, 0)
		require.NoError(t, err)
	}

	reqBody := models.Request{
		Action: models.ActionTicketStatusList,
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

	ticketStatuses, ok := resp.Data["ticketStatuses"].([]interface{})
	require.True(t, ok)
	assert.Len(t, ticketStatuses, 3)
}

// TestTicketStatusHandler_List_ExcludesDeleted tests that list excludes soft-deleted statuses
func TestTicketStatusHandler_List_ExcludesDeleted(t *testing.T) {
	handler := setupTestHandler(t)

	// Insert active and deleted ticket statuses
	_, err := handler.db.Exec(context.Background(),
		"INSERT INTO ticket_status (id, title, description, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)",
		"active-status", "Open", "Active status", 1000, 1000, 0)
	require.NoError(t, err)

	_, err = handler.db.Exec(context.Background(),
		"INSERT INTO ticket_status (id, title, description, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)",
		"deleted-status", "Archived", "Deleted status", 1000, 1000, 1)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionTicketStatusList,
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
	err = json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)

	ticketStatuses, ok := resp.Data["ticketStatuses"].([]interface{})
	require.True(t, ok)
	assert.Len(t, ticketStatuses, 1)

	status := ticketStatuses[0].(map[string]interface{})
	assert.Equal(t, "active-status", status["id"])
	assert.Equal(t, "Open", status["title"])
}

// TestTicketStatusHandler_List_OrderedByTitle tests that statuses are ordered by title
func TestTicketStatusHandler_List_OrderedByTitle(t *testing.T) {
	handler := setupTestHandler(t)

	// Insert statuses in non-alphabetical order
	statuses := []struct {
		id    string
		title string
	}{
		{"status-1", "Zebra"},
		{"status-2", "Apple"},
		{"status-3", "Mango"},
	}

	for _, status := range statuses {
		_, err := handler.db.Exec(context.Background(),
			"INSERT INTO ticket_status (id, title, description, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)",
			status.id, status.title, "Description", 1000, 1000, 0)
		require.NoError(t, err)
	}

	reqBody := models.Request{
		Action: models.ActionTicketStatusList,
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

	ticketStatuses, ok := resp.Data["ticketStatuses"].([]interface{})
	require.True(t, ok)
	assert.Len(t, ticketStatuses, 3)

	// Verify alphabetical ordering
	assert.Equal(t, "Apple", ticketStatuses[0].(map[string]interface{})["title"])
	assert.Equal(t, "Mango", ticketStatuses[1].(map[string]interface{})["title"])
	assert.Equal(t, "Zebra", ticketStatuses[2].(map[string]interface{})["title"])
}

// TestTicketStatusHandler_Modify_Success tests successful ticket status modification
func TestTicketStatusHandler_Modify_Success(t *testing.T) {
	handler := setupTestHandler(t)

	// Insert test ticket status
	_, err := handler.db.Exec(context.Background(),
		"INSERT INTO ticket_status (id, title, description, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)",
		"test-status-id", "In Progress", "Old description", 1000, 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionTicketStatusModify,
		Data: map[string]interface{}{
			"id":          "test-status-id",
			"title":       "In Development",
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

	var resp models.Response
	err = json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)

	// Verify update in database
	var title, description string
	err = handler.db.QueryRow(context.Background(),
		"SELECT title, description FROM ticket_status WHERE id = ?",
		"test-status-id").Scan(&title, &description)
	require.NoError(t, err)
	assert.Equal(t, "In Development", title)
	assert.Equal(t, "Updated description", description)
}

// TestTicketStatusHandler_Modify_TitleOnly tests modifying only the title
func TestTicketStatusHandler_Modify_TitleOnly(t *testing.T) {
	handler := setupTestHandler(t)

	// Insert test ticket status
	_, err := handler.db.Exec(context.Background(),
		"INSERT INTO ticket_status (id, title, description, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)",
		"test-status-id", "In Progress", "Original description", 1000, 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionTicketStatusModify,
		Data: map[string]interface{}{
			"id":    "test-status-id",
			"title": "Under Development",
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

	// Verify title update and description unchanged
	var title, description string
	err = handler.db.QueryRow(context.Background(),
		"SELECT title, description FROM ticket_status WHERE id = ?",
		"test-status-id").Scan(&title, &description)
	require.NoError(t, err)
	assert.Equal(t, "Under Development", title)
	assert.Equal(t, "Original description", description)
}

// TestTicketStatusHandler_Modify_NotFound tests modifying non-existent ticket status
func TestTicketStatusHandler_Modify_NotFound(t *testing.T) {
	handler := setupTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionTicketStatusModify,
		Data: map[string]interface{}{
			"id":    "non-existent-status",
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

// TestTicketStatusHandler_Modify_NoFieldsToUpdate tests modifying without providing any fields
func TestTicketStatusHandler_Modify_NoFieldsToUpdate(t *testing.T) {
	handler := setupTestHandler(t)

	// Insert test ticket status
	_, err := handler.db.Exec(context.Background(),
		"INSERT INTO ticket_status (id, title, description, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)",
		"test-status-id", "In Progress", "Description", 1000, 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionTicketStatusModify,
		Data: map[string]interface{}{
			"id": "test-status-id",
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

	handler.DoAction(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp models.Response
	err = json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeMissingData, resp.ErrorCode)
}

// TestTicketStatusHandler_Remove_Success tests successful ticket status deletion
func TestTicketStatusHandler_Remove_Success(t *testing.T) {
	handler := setupTestHandler(t)

	// Insert test ticket status
	_, err := handler.db.Exec(context.Background(),
		"INSERT INTO ticket_status (id, title, description, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)",
		"test-status-id", "Deprecated", "Old status", 1000, 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionTicketStatusRemove,
		Data: map[string]interface{}{
			"id": "test-status-id",
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
		"SELECT deleted FROM ticket_status WHERE id = ?",
		"test-status-id").Scan(&deleted)
	require.NoError(t, err)
	assert.True(t, deleted)
}

// TestTicketStatusHandler_Remove_NotFound tests deleting non-existent ticket status
func TestTicketStatusHandler_Remove_NotFound(t *testing.T) {
	handler := setupTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionTicketStatusRemove,
		Data: map[string]interface{}{
			"id": "non-existent-status",
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

// TestTicketStatusHandler_CRUD_FullCycle tests complete ticket status lifecycle
func TestTicketStatusHandler_CRUD_FullCycle(t *testing.T) {
	handler := setupTestHandler(t)

	// 1. Create ticket status
	createReq := models.Request{
		Action: models.ActionTicketStatusCreate,
		Data: map[string]interface{}{
			"title":       "In Review",
			"description": "Under review",
		},
	}
	body, _ := json.Marshal(createReq)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")
	handler.DoAction(c)

	var createResp models.Response
	json.NewDecoder(w.Body).Decode(&createResp)
	statusData := createResp.Data["ticketStatus"].(map[string]interface{})
	statusID := statusData["id"].(string)

	// 2. Read ticket status
	readReq := models.Request{
		Action: models.ActionTicketStatusRead,
		Data:   map[string]interface{}{"id": statusID},
	}
	body, _ = json.Marshal(readReq)
	req = httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")
	handler.DoAction(c)
	assert.Equal(t, http.StatusOK, w.Code)

	// 3. Modify ticket status
	modifyReq := models.Request{
		Action: models.ActionTicketStatusModify,
		Data: map[string]interface{}{
			"id":    statusID,
			"title": "Under Review",
		},
	}
	body, _ = json.Marshal(modifyReq)
	req = httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")
	handler.DoAction(c)
	assert.Equal(t, http.StatusOK, w.Code)

	// 4. Delete ticket status
	deleteReq := models.Request{
		Action: models.ActionTicketStatusRemove,
		Data:   map[string]interface{}{"id": statusID},
	}
	body, _ = json.Marshal(deleteReq)
	req = httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")
	handler.DoAction(c)
	assert.Equal(t, http.StatusOK, w.Code)

	// 5. Verify deletion - status should not be found
	readReq = models.Request{
		Action: models.ActionTicketStatusRead,
		Data:   map[string]interface{}{"id": statusID},
	}
	body, _ = json.Marshal(readReq)
	req = httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")
	handler.DoAction(c)
	assert.Equal(t, http.StatusNotFound, w.Code)
}
