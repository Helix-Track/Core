package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"helixtrack.ru/core/internal/config"
	"helixtrack.ru/core/internal/database"
	"helixtrack.ru/core/internal/models"
	"helixtrack.ru/core/internal/services"
)

func setupSubtaskTestHandler(t *testing.T) (*Handler, database.Database) {
	db, err := database.NewDatabase(config.DatabaseConfig{
		Type:       "sqlite",
		SQLitePath: ":memory:",
	})
	require.NoError(t, err)

	// Initialize schema
	err = InitializeProjectTables(db)
	require.NoError(t, err)

	mockAuth := &services.MockAuthService{
		IsEnabledFunc: func() bool { return true },
	}

	mockPerm := &services.MockPermissionService{
		IsEnabledFunc: func() bool { return true },
		CheckPermissionFunc: func(ctx context.Context, username, permissionContext string, requiredLevel models.PermissionLevel) (bool, error) {
			return username == "testuser", nil
		},
	}

	handler := NewHandler(db, mockAuth, mockPerm, "1.0.0-test")
	mockPublisher := NewMockEventPublisher(true)
	handler.SetEventPublisher(mockPublisher)

	return handler, db
}

// Helper to create a test ticket with required fields for subtask tests
func createTestTicketForSubtask(t *testing.T, db database.Database, ticketID string, ticketNumber int, title, key string, isSubtask bool, parentTicketID string) {
	var projectID, typeID, statusID string
	db.QueryRow(context.Background(), "SELECT id FROM workflow LIMIT 1").Scan(&projectID)
	db.QueryRow(context.Background(), "SELECT id FROM ticket_type WHERE title = 'task' LIMIT 1").Scan(&typeID)
	db.QueryRow(context.Background(), "SELECT id FROM ticket_status WHERE title = 'open' LIMIT 1").Scan(&statusID)

	query := `INSERT INTO ticket (id, ticket_number, title, ticket_key, ticket_type_id, ticket_status_id, project_id, creator, is_subtask, parent_ticket_id, created, modified, deleted)
	          VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := db.Exec(context.Background(),
		query,
		ticketID, ticketNumber, title, key, typeID, statusID, projectID, "testuser", isSubtask,
		nullStringOrValue(parentTicketID), time.Now().Unix(), time.Now().Unix(), false)
	require.NoError(t, err)
}

// Helper to convert string to NULL or value
func nullStringOrValue(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

func TestSubtaskHandler_Create_Success(t *testing.T) {
	handler, db := setupSubtaskTestHandler(t)
	defer db.Close()

	// Create a parent ticket first
	parentID := uuid.New().String()
	createTestTicketForSubtask(t, db, parentID, 1, "Parent Ticket", "PARENT-1", false, "")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionSubtaskCreate,
		Data: map[string]interface{}{
			"parentTicketId": parentID,
			"title":          "My Subtask",
			"description":    "Subtask description",
		},
	}

	handler.handleSubtaskCreate(c, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.Equal(t, "My Subtask", resp.Data["title"])
	assert.Equal(t, parentID, resp.Data["parentTicketId"])
	assert.True(t, resp.Data["isSubtask"].(bool))
}

func TestSubtaskHandler_Create_MissingParentTicketID(t *testing.T) {
	handler, db := setupSubtaskTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionSubtaskCreate,
		Data: map[string]interface{}{
			"title": "My Subtask",
		},
	}

	handler.handleSubtaskCreate(c, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeMissingData, resp.ErrorCode)
}

func TestSubtaskHandler_Create_ParentNotFound(t *testing.T) {
	handler, db := setupSubtaskTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionSubtaskCreate,
		Data: map[string]interface{}{
			"parentTicketId": "nonexistent-id",
			"title":          "My Subtask",
		},
	}

	handler.handleSubtaskCreate(c, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeEntityNotFound, resp.ErrorCode)
}

func TestSubtaskHandler_Create_ParentIsSubtask(t *testing.T) {
	handler, db := setupSubtaskTestHandler(t)
	defer db.Close()

	// Create a subtask (cannot be a parent)
	parentID := uuid.New().String()
	createTestTicketForSubtask(t, db, parentID, 1, "Subtask Parent", "SUB-1", true, "some-parent-id")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionSubtaskCreate,
		Data: map[string]interface{}{
			"parentTicketId": parentID,
			"title":          "My Subtask",
		},
	}

	handler.handleSubtaskCreate(c, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeEntityNotFound, resp.ErrorCode)
}

func TestSubtaskHandler_List_Success(t *testing.T) {
	handler, db := setupSubtaskTestHandler(t)
	defer db.Close()

	// Create parent ticket
	parentID := uuid.New().String()
	createTestTicketForSubtask(t, db, parentID, 1, "Parent", "PARENT-1", false, "")

	// Create multiple subtasks
	for i := 0; i < 3; i++ {
		subtaskID := uuid.New().String()
		createTestTicketForSubtask(t, db, subtaskID, i+2, "Subtask "+string(rune(i+1)), "SUB-"+string(rune(i+1)), true, parentID)
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionSubtaskList,
		Data:   map[string]interface{}{},
	}

	handler.handleSubtaskList(c, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.Equal(t, float64(3), resp.Data["count"])
}

func TestSubtaskHandler_MoveToParent_Success(t *testing.T) {
	handler, db := setupSubtaskTestHandler(t)
	defer db.Close()

	// Create two parent tickets
	parent1ID := uuid.New().String()
	createTestTicketForSubtask(t, db, parent1ID, 1, "Parent 1", "PARENT-1", false, "")

	parent2ID := uuid.New().String()
	createTestTicketForSubtask(t, db, parent2ID, 2, "Parent 2", "PARENT-2", false, "")

	// Create a subtask under parent1
	subtaskID := uuid.New().String()
	createTestTicketForSubtask(t, db, subtaskID, 3, "Subtask", "SUB-1", true, parent1ID)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionSubtaskMoveToParent,
		Data: map[string]interface{}{
			"subtaskId":        subtaskID,
			"newParentTicketId": parent2ID,
		},
	}

	handler.handleSubtaskMoveToParent(c, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.True(t, resp.Data["moved"].(bool))
}

func TestSubtaskHandler_MoveToParent_SubtaskNotFound(t *testing.T) {
	handler, db := setupSubtaskTestHandler(t)
	defer db.Close()

	// Create a parent ticket
	parentID := uuid.New().String()
	createTestTicketForSubtask(t, db, parentID, 1, "Parent", "PARENT-1", false, "")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionSubtaskMoveToParent,
		Data: map[string]interface{}{
			"subtaskId":        "nonexistent-id",
			"newParentTicketId": parentID,
		},
	}

	handler.handleSubtaskMoveToParent(c, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeEntityNotFound, resp.ErrorCode)
}

func TestSubtaskHandler_MoveToParent_NewParentIsSubtask(t *testing.T) {
	handler, db := setupSubtaskTestHandler(t)
	defer db.Close()

	// Create a parent ticket
	parent1ID := uuid.New().String()
	createTestTicketForSubtask(t, db, parent1ID, 1, "Parent 1", "PARENT-1", false, "")

	// Create another subtask (cannot be new parent)
	parent2ID := uuid.New().String()
	createTestTicketForSubtask(t, db, parent2ID, 2, "Subtask Parent", "SUB-PARENT", true, parent1ID)

	// Create the subtask to move
	subtaskID := uuid.New().String()
	createTestTicketForSubtask(t, db, subtaskID, 3, "Subtask", "SUB-1", true, parent1ID)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionSubtaskMoveToParent,
		Data: map[string]interface{}{
			"subtaskId":        subtaskID,
			"newParentTicketId": parent2ID,
		},
	}

	handler.handleSubtaskMoveToParent(c, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeEntityNotFound, resp.ErrorCode)
}

func TestSubtaskHandler_ConvertToIssue_Success(t *testing.T) {
	handler, db := setupSubtaskTestHandler(t)
	defer db.Close()

	// Create a parent ticket
	parentID := uuid.New().String()
	createTestTicketForSubtask(t, db, parentID, 1, "Parent", "PARENT-1", false, "")

	// Create a subtask
	subtaskID := uuid.New().String()
	createTestTicketForSubtask(t, db, subtaskID, 2, "Subtask", "SUB-1", true, parentID)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionSubtaskConvertToIssue,
		Data: map[string]interface{}{
			"subtaskId": subtaskID,
		},
	}

	handler.handleSubtaskConvertToIssue(c, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.True(t, resp.Data["converted"].(bool))
	assert.False(t, resp.Data["isSubtask"].(bool))
}

func TestSubtaskHandler_ConvertToIssue_NotFound(t *testing.T) {
	handler, db := setupSubtaskTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionSubtaskConvertToIssue,
		Data: map[string]interface{}{
			"subtaskId": "nonexistent-id",
		},
	}

	handler.handleSubtaskConvertToIssue(c, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeEntityNotFound, resp.ErrorCode)
}

func TestSubtaskHandler_ListByParent_Success(t *testing.T) {
	handler, db := setupSubtaskTestHandler(t)
	defer db.Close()

	// Create a parent ticket
	parentID := uuid.New().String()
	createTestTicketForSubtask(t, db, parentID, 1, "Parent", "PARENT-1", false, "")

	// Create multiple subtasks under this parent
	for i := 0; i < 3; i++ {
		subtaskID := uuid.New().String()
		createTestTicketForSubtask(t, db, subtaskID, i+2, "Subtask "+string(rune(i+1)), "SUB-"+string(rune(i+1)), true, parentID)
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionSubtaskListByParent,
		Data: map[string]interface{}{
			"parentTicketId": parentID,
		},
	}

	handler.handleSubtaskListByParent(c, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.Equal(t, float64(3), resp.Data["totalCount"])
	assert.Equal(t, parentID, resp.Data["parentTicketId"])
}

func TestSubtaskHandler_ListByParent_MissingParentID(t *testing.T) {
	handler, db := setupSubtaskTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionSubtaskListByParent,
		Data:   map[string]interface{}{},
	}

	handler.handleSubtaskListByParent(c, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeMissingData, resp.ErrorCode)
}

func TestSubtaskHandler_Unauthorized(t *testing.T) {
	handler, db := setupSubtaskTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	// No username set

	req := &models.Request{
		Action: models.ActionSubtaskCreate,
		Data:   map[string]interface{}{},
	}

	handler.handleSubtaskCreate(c, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
