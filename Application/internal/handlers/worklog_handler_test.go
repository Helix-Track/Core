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

func setupWorkLogTestHandler(t *testing.T) (*Handler, database.Database) {
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

// Helper to create a test ticket for work log tests
func createTestTicketForWorkLog(t *testing.T, db database.Database, ticketID string, ticketNumber int, title, key string) {
	var projectID, typeID, statusID string
	db.QueryRow(context.Background(), "SELECT id FROM workflow LIMIT 1").Scan(&projectID)
	db.QueryRow(context.Background(), "SELECT id FROM ticket_type WHERE title = 'task' LIMIT 1").Scan(&typeID)
	db.QueryRow(context.Background(), "SELECT id FROM ticket_status WHERE title = 'open' LIMIT 1").Scan(&statusID)

	_, err := db.Exec(context.Background(),
		"INSERT INTO ticket (id, ticket_number, title, ticket_key, ticket_type_id, ticket_status_id, project_id, creator, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		ticketID, ticketNumber, title, key, typeID, statusID, projectID, "testuser", time.Now().Unix(), time.Now().Unix(), false)
	require.NoError(t, err)
}

// Helper to create a test work log entry
func createTestWorkLog(t *testing.T, db database.Database, workLogID, ticketID, userID string, timeSpent int) {
	_, err := db.Exec(context.Background(),
		"INSERT INTO work_log (id, ticket_id, user_id, time_spent, work_date, description, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		workLogID, ticketID, userID, timeSpent, time.Now().Unix(), "Test work log", time.Now().Unix(), time.Now().Unix(), false)
	require.NoError(t, err)
}

// ============================================================================
// ActionWorkLogAdd Tests
// ============================================================================

func TestWorkLogHandler_Add_Success(t *testing.T) {
	handler, db := setupWorkLogTestHandler(t)
	defer db.Close()

	// Create a test ticket first
	ticketID := uuid.New().String()
	createTestTicketForWorkLog(t, db, ticketID, 1, "Test Ticket", "TEST-1")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionWorkLogAdd,
		Data: map[string]interface{}{
			"ticketId":    ticketID,
			"timeSpent":   120.0, // 2 hours in minutes
			"workDate":    float64(time.Now().Unix()),
			"description": "Fixed critical bug",
		},
	}

	handler.handleWorkLogAdd(c, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)

	workLogData := resp.Data["workLog"].(map[string]interface{})
	assert.Equal(t, ticketID, workLogData["ticketId"])
	assert.Equal(t, float64(120), workLogData["timeSpent"])
	assert.Equal(t, "testuser", workLogData["userId"])
	assert.Equal(t, "Fixed critical bug", workLogData["description"])
}

func TestWorkLogHandler_Add_Success_DefaultWorkDate(t *testing.T) {
	handler, db := setupWorkLogTestHandler(t)
	defer db.Close()

	// Create a test ticket
	ticketID := uuid.New().String()
	createTestTicketForWorkLog(t, db, ticketID, 1, "Test Ticket", "TEST-1")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionWorkLogAdd,
		Data: map[string]interface{}{
			"ticketId":  ticketID,
			"timeSpent": 60.0, // 1 hour
			// workDate not specified - should default to now
		},
	}

	handler.handleWorkLogAdd(c, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)

	workLogData := resp.Data["workLog"].(map[string]interface{})
	assert.NotZero(t, workLogData["workDate"])
}

func TestWorkLogHandler_Add_MissingTicketID(t *testing.T) {
	handler, db := setupWorkLogTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionWorkLogAdd,
		Data: map[string]interface{}{
			"timeSpent": 120.0,
		},
	}

	handler.handleWorkLogAdd(c, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeMissingData, resp.ErrorCode)
	assert.Contains(t, resp.ErrorMessage, "ticketId")
}

func TestWorkLogHandler_Add_MissingTimeSpent(t *testing.T) {
	handler, db := setupWorkLogTestHandler(t)
	defer db.Close()

	ticketID := uuid.New().String()
	createTestTicketForWorkLog(t, db, ticketID, 1, "Test Ticket", "TEST-1")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionWorkLogAdd,
		Data: map[string]interface{}{
			"ticketId": ticketID,
			// timeSpent missing
		},
	}

	handler.handleWorkLogAdd(c, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeMissingData, resp.ErrorCode)
	assert.Contains(t, resp.ErrorMessage, "timeSpent")
}

func TestWorkLogHandler_Add_InvalidTimeSpent_Zero(t *testing.T) {
	handler, db := setupWorkLogTestHandler(t)
	defer db.Close()

	ticketID := uuid.New().String()
	createTestTicketForWorkLog(t, db, ticketID, 1, "Test Ticket", "TEST-1")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionWorkLogAdd,
		Data: map[string]interface{}{
			"ticketId":  ticketID,
			"timeSpent": 0.0,
		},
	}

	handler.handleWorkLogAdd(c, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeMissingData, resp.ErrorCode)
}

func TestWorkLogHandler_Add_InvalidTimeSpent_Negative(t *testing.T) {
	handler, db := setupWorkLogTestHandler(t)
	defer db.Close()

	ticketID := uuid.New().String()
	createTestTicketForWorkLog(t, db, ticketID, 1, "Test Ticket", "TEST-1")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionWorkLogAdd,
		Data: map[string]interface{}{
			"ticketId":  ticketID,
			"timeSpent": -30.0,
		},
	}

	handler.handleWorkLogAdd(c, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeMissingData, resp.ErrorCode)
}

func TestWorkLogHandler_Add_TicketNotFound(t *testing.T) {
	handler, db := setupWorkLogTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionWorkLogAdd,
		Data: map[string]interface{}{
			"ticketId":  "nonexistent-id",
			"timeSpent": 60.0,
		},
	}

	handler.handleWorkLogAdd(c, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeEntityNotFound, resp.ErrorCode)
}

func TestWorkLogHandler_Add_Unauthorized(t *testing.T) {
	handler, db := setupWorkLogTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	// No username set

	req := &models.Request{
		Action: models.ActionWorkLogAdd,
		Data:   map[string]interface{}{},
	}

	handler.handleWorkLogAdd(c, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// ============================================================================
// ActionWorkLogModify Tests
// ============================================================================

func TestWorkLogHandler_Modify_Success(t *testing.T) {
	handler, db := setupWorkLogTestHandler(t)
	defer db.Close()

	// Create a ticket and work log
	ticketID := uuid.New().String()
	createTestTicketForWorkLog(t, db, ticketID, 1, "Test Ticket", "TEST-1")

	workLogID := uuid.New().String()
	createTestWorkLog(t, db, workLogID, ticketID, "testuser", 60)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionWorkLogModify,
		Data: map[string]interface{}{
			"id":          workLogID,
			"timeSpent":   180.0, // Update to 3 hours
			"description": "Updated description",
		},
	}

	handler.handleWorkLogModify(c, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.True(t, resp.Data["updated"].(bool))
	assert.Equal(t, workLogID, resp.Data["id"])
}

func TestWorkLogHandler_Modify_OnlyTimeSpent(t *testing.T) {
	handler, db := setupWorkLogTestHandler(t)
	defer db.Close()

	ticketID := uuid.New().String()
	createTestTicketForWorkLog(t, db, ticketID, 1, "Test Ticket", "TEST-1")

	workLogID := uuid.New().String()
	createTestWorkLog(t, db, workLogID, ticketID, "testuser", 60)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionWorkLogModify,
		Data: map[string]interface{}{
			"id":        workLogID,
			"timeSpent": 120.0,
		},
	}

	handler.handleWorkLogModify(c, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.True(t, resp.Data["updated"].(bool))
}

func TestWorkLogHandler_Modify_OnlyDescription(t *testing.T) {
	handler, db := setupWorkLogTestHandler(t)
	defer db.Close()

	ticketID := uuid.New().String()
	createTestTicketForWorkLog(t, db, ticketID, 1, "Test Ticket", "TEST-1")

	workLogID := uuid.New().String()
	createTestWorkLog(t, db, workLogID, ticketID, "testuser", 60)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionWorkLogModify,
		Data: map[string]interface{}{
			"id":          workLogID,
			"description": "New description",
		},
	}

	handler.handleWorkLogModify(c, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
}

func TestWorkLogHandler_Modify_OnlyWorkDate(t *testing.T) {
	handler, db := setupWorkLogTestHandler(t)
	defer db.Close()

	ticketID := uuid.New().String()
	createTestTicketForWorkLog(t, db, ticketID, 1, "Test Ticket", "TEST-1")

	workLogID := uuid.New().String()
	createTestWorkLog(t, db, workLogID, ticketID, "testuser", 60)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	newWorkDate := float64(time.Now().AddDate(0, 0, -1).Unix())
	req := &models.Request{
		Action: models.ActionWorkLogModify,
		Data: map[string]interface{}{
			"id":       workLogID,
			"workDate": newWorkDate,
		},
	}

	handler.handleWorkLogModify(c, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
}

func TestWorkLogHandler_Modify_MissingID(t *testing.T) {
	handler, db := setupWorkLogTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionWorkLogModify,
		Data: map[string]interface{}{
			"timeSpent": 120.0,
		},
	}

	handler.handleWorkLogModify(c, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeMissingData, resp.ErrorCode)
}

func TestWorkLogHandler_Modify_WorkLogNotFound(t *testing.T) {
	handler, db := setupWorkLogTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionWorkLogModify,
		Data: map[string]interface{}{
			"id":        "nonexistent-id",
			"timeSpent": 120.0,
		},
	}

	handler.handleWorkLogModify(c, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeEntityNotFound, resp.ErrorCode)
}

func TestWorkLogHandler_Modify_NoFieldsToUpdate(t *testing.T) {
	handler, db := setupWorkLogTestHandler(t)
	defer db.Close()

	ticketID := uuid.New().String()
	createTestTicketForWorkLog(t, db, ticketID, 1, "Test Ticket", "TEST-1")

	workLogID := uuid.New().String()
	createTestWorkLog(t, db, workLogID, ticketID, "testuser", 60)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionWorkLogModify,
		Data: map[string]interface{}{
			"id": workLogID,
			// No fields to update
		},
	}

	handler.handleWorkLogModify(c, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeMissingData, resp.ErrorCode)
}

func TestWorkLogHandler_Modify_Unauthorized(t *testing.T) {
	handler, db := setupWorkLogTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	// No username set

	req := &models.Request{
		Action: models.ActionWorkLogModify,
		Data:   map[string]interface{}{},
	}

	handler.handleWorkLogModify(c, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// ============================================================================
// ActionWorkLogRemove Tests
// ============================================================================

func TestWorkLogHandler_Remove_Success(t *testing.T) {
	handler, db := setupWorkLogTestHandler(t)
	defer db.Close()

	// Create a ticket and work log
	ticketID := uuid.New().String()
	createTestTicketForWorkLog(t, db, ticketID, 1, "Test Ticket", "TEST-1")

	workLogID := uuid.New().String()
	createTestWorkLog(t, db, workLogID, ticketID, "testuser", 60)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionWorkLogRemove,
		Data: map[string]interface{}{
			"id": workLogID,
		},
	}

	handler.handleWorkLogRemove(c, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.True(t, resp.Data["deleted"].(bool))
	assert.Equal(t, workLogID, resp.Data["id"])
}

func TestWorkLogHandler_Remove_MissingID(t *testing.T) {
	handler, db := setupWorkLogTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionWorkLogRemove,
		Data:   map[string]interface{}{},
	}

	handler.handleWorkLogRemove(c, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeMissingData, resp.ErrorCode)
}

func TestWorkLogHandler_Remove_NotFound(t *testing.T) {
	handler, db := setupWorkLogTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionWorkLogRemove,
		Data: map[string]interface{}{
			"id": "nonexistent-id",
		},
	}

	handler.handleWorkLogRemove(c, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeEntityNotFound, resp.ErrorCode)
}

func TestWorkLogHandler_Remove_AlreadyDeleted(t *testing.T) {
	handler, db := setupWorkLogTestHandler(t)
	defer db.Close()

	ticketID := uuid.New().String()
	createTestTicketForWorkLog(t, db, ticketID, 1, "Test Ticket", "TEST-1")

	workLogID := uuid.New().String()
	// Create already deleted work log
	_, err := db.Exec(context.Background(),
		"INSERT INTO work_log (id, ticket_id, user_id, time_spent, work_date, description, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		workLogID, ticketID, "testuser", 60, time.Now().Unix(), "Test", time.Now().Unix(), time.Now().Unix(), true)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionWorkLogRemove,
		Data: map[string]interface{}{
			"id": workLogID,
		},
	}

	handler.handleWorkLogRemove(c, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp models.Response
	err = json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeEntityNotFound, resp.ErrorCode)
}

func TestWorkLogHandler_Remove_Unauthorized(t *testing.T) {
	handler, db := setupWorkLogTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	// No username set

	req := &models.Request{
		Action: models.ActionWorkLogRemove,
		Data:   map[string]interface{}{},
	}

	handler.handleWorkLogRemove(c, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// ============================================================================
// ActionWorkLogList Tests
// ============================================================================

func TestWorkLogHandler_List_Success(t *testing.T) {
	handler, db := setupWorkLogTestHandler(t)
	defer db.Close()

	// Create tickets and work logs
	for i := 1; i <= 3; i++ {
		ticketID := uuid.New().String()
		createTestTicketForWorkLog(t, db, ticketID, i, "Ticket "+string(rune(i)), "TEST-"+string(rune(i)))

		workLogID := uuid.New().String()
		createTestWorkLog(t, db, workLogID, ticketID, "testuser", 60*i)
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionWorkLogList,
		Data:   map[string]interface{}{},
	}

	handler.handleWorkLogList(c, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.Equal(t, float64(3), resp.Data["count"])

	workLogs := resp.Data["workLogs"].([]interface{})
	assert.Len(t, workLogs, 3)
}

func TestWorkLogHandler_List_Empty(t *testing.T) {
	handler, db := setupWorkLogTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionWorkLogList,
		Data:   map[string]interface{}{},
	}

	handler.handleWorkLogList(c, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.Equal(t, float64(0), resp.Data["count"])

	workLogs := resp.Data["workLogs"].([]interface{})
	assert.Len(t, workLogs, 0)
}

func TestWorkLogHandler_List_ExcludesDeleted(t *testing.T) {
	handler, db := setupWorkLogTestHandler(t)
	defer db.Close()

	ticketID := uuid.New().String()
	createTestTicketForWorkLog(t, db, ticketID, 1, "Test Ticket", "TEST-1")

	// Create active work log
	workLogID1 := uuid.New().String()
	createTestWorkLog(t, db, workLogID1, ticketID, "testuser", 60)

	// Create deleted work log
	workLogID2 := uuid.New().String()
	_, err := db.Exec(context.Background(),
		"INSERT INTO work_log (id, ticket_id, user_id, time_spent, work_date, description, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		workLogID2, ticketID, "testuser", 120, time.Now().Unix(), "Deleted", time.Now().Unix(), time.Now().Unix(), true)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionWorkLogList,
		Data:   map[string]interface{}{},
	}

	handler.handleWorkLogList(c, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err = json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.Equal(t, float64(1), resp.Data["count"]) // Only 1 active work log
}

func TestWorkLogHandler_List_Unauthorized(t *testing.T) {
	handler, db := setupWorkLogTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	// No username set

	req := &models.Request{
		Action: models.ActionWorkLogList,
		Data:   map[string]interface{}{},
	}

	handler.handleWorkLogList(c, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// ============================================================================
// ActionWorkLogListByTicket Tests
// ============================================================================

func TestWorkLogHandler_ListByTicket_Success(t *testing.T) {
	handler, db := setupWorkLogTestHandler(t)
	defer db.Close()

	// Create two tickets
	ticket1ID := uuid.New().String()
	createTestTicketForWorkLog(t, db, ticket1ID, 1, "Ticket 1", "TEST-1")

	ticket2ID := uuid.New().String()
	createTestTicketForWorkLog(t, db, ticket2ID, 2, "Ticket 2", "TEST-2")

	// Create work logs for ticket1
	for i := 1; i <= 3; i++ {
		workLogID := uuid.New().String()
		createTestWorkLog(t, db, workLogID, ticket1ID, "testuser", 60*i)
	}

	// Create work log for ticket2
	workLogID := uuid.New().String()
	createTestWorkLog(t, db, workLogID, ticket2ID, "testuser", 30)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionWorkLogListByTicket,
		Data: map[string]interface{}{
			"ticketId": ticket1ID,
		},
	}

	handler.handleWorkLogListByTicket(c, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.Equal(t, ticket1ID, resp.Data["ticketId"])
	assert.Equal(t, float64(3), resp.Data["count"])
	assert.Equal(t, float64(360), resp.Data["totalMinutes"]) // 60 + 120 + 180 = 360
	assert.Equal(t, float64(6.0), resp.Data["totalHours"])   // 360 / 60 = 6
}

func TestWorkLogHandler_ListByTicket_Empty(t *testing.T) {
	handler, db := setupWorkLogTestHandler(t)
	defer db.Close()

	ticketID := uuid.New().String()
	createTestTicketForWorkLog(t, db, ticketID, 1, "Test Ticket", "TEST-1")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionWorkLogListByTicket,
		Data: map[string]interface{}{
			"ticketId": ticketID,
		},
	}

	handler.handleWorkLogListByTicket(c, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.Equal(t, float64(0), resp.Data["count"])
	assert.Equal(t, float64(0), resp.Data["totalMinutes"])
}

func TestWorkLogHandler_ListByTicket_MissingTicketID(t *testing.T) {
	handler, db := setupWorkLogTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionWorkLogListByTicket,
		Data:   map[string]interface{}{},
	}

	handler.handleWorkLogListByTicket(c, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeMissingData, resp.ErrorCode)
}

func TestWorkLogHandler_ListByTicket_Unauthorized(t *testing.T) {
	handler, db := setupWorkLogTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	// No username set

	req := &models.Request{
		Action: models.ActionWorkLogListByTicket,
		Data:   map[string]interface{}{},
	}

	handler.handleWorkLogListByTicket(c, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// ============================================================================
// ActionWorkLogListByUser Tests
// ============================================================================

func TestWorkLogHandler_ListByUser_Success(t *testing.T) {
	handler, db := setupWorkLogTestHandler(t)
	defer db.Close()

	// Create tickets
	ticket1ID := uuid.New().String()
	createTestTicketForWorkLog(t, db, ticket1ID, 1, "Ticket 1", "TEST-1")

	ticket2ID := uuid.New().String()
	createTestTicketForWorkLog(t, db, ticket2ID, 2, "Ticket 2", "TEST-2")

	// Create work logs for testuser
	for i := 1; i <= 3; i++ {
		workLogID := uuid.New().String()
		createTestWorkLog(t, db, workLogID, ticket1ID, "testuser", 60*i)
	}

	// Create work log for different user
	workLogID := uuid.New().String()
	_, err := db.Exec(context.Background(),
		"INSERT INTO work_log (id, ticket_id, user_id, time_spent, work_date, description, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		workLogID, ticket2ID, "otheruser", 120, time.Now().Unix(), "Other user work", time.Now().Unix(), time.Now().Unix(), false)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionWorkLogListByUser,
		Data: map[string]interface{}{
			"userId": "testuser",
		},
	}

	handler.handleWorkLogListByUser(c, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err = json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.Equal(t, "testuser", resp.Data["userId"])
	assert.Equal(t, float64(3), resp.Data["count"])
	assert.Equal(t, float64(360), resp.Data["totalMinutes"]) // 60 + 120 + 180 = 360
}

func TestWorkLogHandler_ListByUser_DefaultsToCurrentUser(t *testing.T) {
	handler, db := setupWorkLogTestHandler(t)
	defer db.Close()

	ticketID := uuid.New().String()
	createTestTicketForWorkLog(t, db, ticketID, 1, "Test Ticket", "TEST-1")

	// Create work logs for testuser
	for i := 1; i <= 2; i++ {
		workLogID := uuid.New().String()
		createTestWorkLog(t, db, workLogID, ticketID, "testuser", 60*i)
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionWorkLogListByUser,
		Data:   map[string]interface{}{
			// No userId specified - should default to current user
		},
	}

	handler.handleWorkLogListByUser(c, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.Equal(t, "testuser", resp.Data["userId"])
	assert.Equal(t, float64(2), resp.Data["count"])
}

func TestWorkLogHandler_ListByUser_Empty(t *testing.T) {
	handler, db := setupWorkLogTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionWorkLogListByUser,
		Data: map[string]interface{}{
			"userId": "testuser",
		},
	}

	handler.handleWorkLogListByUser(c, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.Equal(t, float64(0), resp.Data["count"])
	assert.Equal(t, float64(0), resp.Data["totalMinutes"])
}

func TestWorkLogHandler_ListByUser_Unauthorized(t *testing.T) {
	handler, db := setupWorkLogTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	// No username set

	req := &models.Request{
		Action: models.ActionWorkLogListByUser,
		Data:   map[string]interface{}{},
	}

	handler.handleWorkLogListByUser(c, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// ============================================================================
// ActionWorkLogGetTotalTime Tests
// ============================================================================

func TestWorkLogHandler_GetTotalTime_Success(t *testing.T) {
	handler, db := setupWorkLogTestHandler(t)
	defer db.Close()

	ticketID := uuid.New().String()
	createTestTicketForWorkLog(t, db, ticketID, 1, "Test Ticket", "TEST-1")

	// Create multiple work logs
	workLogID1 := uuid.New().String()
	createTestWorkLog(t, db, workLogID1, ticketID, "testuser", 60)

	workLogID2 := uuid.New().String()
	createTestWorkLog(t, db, workLogID2, ticketID, "testuser", 120)

	workLogID3 := uuid.New().String()
	createTestWorkLog(t, db, workLogID3, ticketID, "otheruser", 90)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionWorkLogGetTotalTime,
		Data: map[string]interface{}{
			"ticketId": ticketID,
		},
	}

	handler.handleWorkLogGetTotalTime(c, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.Equal(t, ticketID, resp.Data["ticketId"])
	assert.Equal(t, float64(270), resp.Data["totalMinutes"])     // 60 + 120 + 90 = 270
	assert.Equal(t, float64(4.5), resp.Data["totalHours"])       // 270 / 60 = 4.5
	assert.Equal(t, float64(0.5625), resp.Data["totalDays"])     // 270 / 480 = 0.5625
}

func TestWorkLogHandler_GetTotalTime_NoWorkLogs(t *testing.T) {
	handler, db := setupWorkLogTestHandler(t)
	defer db.Close()

	ticketID := uuid.New().String()
	createTestTicketForWorkLog(t, db, ticketID, 1, "Test Ticket", "TEST-1")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionWorkLogGetTotalTime,
		Data: map[string]interface{}{
			"ticketId": ticketID,
		},
	}

	handler.handleWorkLogGetTotalTime(c, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.Equal(t, float64(0), resp.Data["totalMinutes"])
	assert.Equal(t, float64(0), resp.Data["totalHours"])
	assert.Equal(t, float64(0), resp.Data["totalDays"])
}

func TestWorkLogHandler_GetTotalTime_ExcludesDeleted(t *testing.T) {
	handler, db := setupWorkLogTestHandler(t)
	defer db.Close()

	ticketID := uuid.New().String()
	createTestTicketForWorkLog(t, db, ticketID, 1, "Test Ticket", "TEST-1")

	// Create active work log
	workLogID1 := uuid.New().String()
	createTestWorkLog(t, db, workLogID1, ticketID, "testuser", 60)

	// Create deleted work log
	workLogID2 := uuid.New().String()
	_, err := db.Exec(context.Background(),
		"INSERT INTO work_log (id, ticket_id, user_id, time_spent, work_date, description, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		workLogID2, ticketID, "testuser", 120, time.Now().Unix(), "Deleted", time.Now().Unix(), time.Now().Unix(), true)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionWorkLogGetTotalTime,
		Data: map[string]interface{}{
			"ticketId": ticketID,
		},
	}

	handler.handleWorkLogGetTotalTime(c, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err = json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.Equal(t, float64(60), resp.Data["totalMinutes"]) // Only active work log
}

func TestWorkLogHandler_GetTotalTime_MissingTicketID(t *testing.T) {
	handler, db := setupWorkLogTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionWorkLogGetTotalTime,
		Data:   map[string]interface{}{},
	}

	handler.handleWorkLogGetTotalTime(c, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeMissingData, resp.ErrorCode)
}

func TestWorkLogHandler_GetTotalTime_Unauthorized(t *testing.T) {
	handler, db := setupWorkLogTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	// No username set

	req := &models.Request{
		Action: models.ActionWorkLogGetTotalTime,
		Data:   map[string]interface{}{},
	}

	handler.handleWorkLogGetTotalTime(c, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
