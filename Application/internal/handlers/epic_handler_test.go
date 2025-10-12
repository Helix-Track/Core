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

func setupEpicTestHandler(t *testing.T) (*Handler, database.Database) {
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

// Helper to create a test ticket with required fields
func createTestTicket(t *testing.T, db database.Database, ticketID string, ticketNumber int, title, key string, isEpic bool, epicName, epicColor string) {
	var projectID, typeID, statusID string
	db.QueryRow(context.Background(), "SELECT id FROM workflow LIMIT 1").Scan(&projectID)
	db.QueryRow(context.Background(), "SELECT id FROM ticket_type WHERE title = 'task' LIMIT 1").Scan(&typeID)
	db.QueryRow(context.Background(), "SELECT id FROM ticket_status WHERE title = 'open' LIMIT 1").Scan(&statusID)

	_, err := db.Exec(context.Background(),
		"INSERT INTO ticket (id, ticket_number, title, ticket_key, ticket_type_id, ticket_status_id, project_id, creator, is_epic, epic_name, epic_color, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		ticketID, ticketNumber, title, key, typeID, statusID, projectID, "testuser", isEpic, epicName, epicColor, time.Now().Unix(), time.Now().Unix(), false)
	require.NoError(t, err)
}

func TestEpicHandler_Create_Success(t *testing.T) {
	handler, db := setupEpicTestHandler(t)
	defer db.Close()

	// Create a test ticket first
	ticketID := uuid.New().String()
	createTestTicket(t, db, ticketID, 1, "Test Ticket", "TEST-1", false, "", "")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionEpicCreate,
		Data: map[string]interface{}{
			"ticketId":  ticketID,
			"epicName":  "My Epic",
			"epicColor": models.EpicColorGhola,
		},
	}

	handler.handleEpicCreate(c, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.Equal(t, ticketID, resp.Data["ticketId"])
	assert.Equal(t, "My Epic", resp.Data["epicName"])
	assert.True(t, resp.Data["isEpic"].(bool))
}

func TestEpicHandler_Create_MissingTicketID(t *testing.T) {
	handler, db := setupEpicTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionEpicCreate,
		Data:   map[string]interface{}{},
	}

	handler.handleEpicCreate(c, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeMissingData, resp.ErrorCode)
}

func TestEpicHandler_Create_TicketNotFound(t *testing.T) {
	handler, db := setupEpicTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionEpicCreate,
		Data: map[string]interface{}{
			"ticketId": "nonexistent-id",
		},
	}

	handler.handleEpicCreate(c, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeEntityNotFound, resp.ErrorCode)
}

func TestEpicHandler_Create_DefaultColor(t *testing.T) {
	handler, db := setupEpicTestHandler(t)
	defer db.Close()

	// Create a test ticket
	ticketID := uuid.New().String()
	createTestTicket(t, db, ticketID, 1, "Test Ticket", "TEST-1", false, "", "")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionEpicCreate,
		Data: map[string]interface{}{
			"ticketId": ticketID,
			"epicName": "My Epic",
			// No epicColor specified
		},
	}

	handler.handleEpicCreate(c, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.Equal(t, models.EpicColorGhola, resp.Data["epicColor"]) // Default color
}

func TestEpicHandler_Read_Success(t *testing.T) {
	handler, db := setupEpicTestHandler(t)
	defer db.Close()

	// Create an epic
	ticketID := uuid.New().String()
	createTestTicket(t, db, ticketID, 1, "Epic Ticket", "EPIC-1", true, "My Epic", models.EpicColorWestar)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionEpicRead,
		Data: map[string]interface{}{
			"ticketId": ticketID,
		},
	}

	handler.handleEpicRead(c, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.Equal(t, "My Epic", resp.Data["epicName"])
	assert.Equal(t, models.EpicColorWestar, resp.Data["epicColor"])
}

func TestEpicHandler_Read_NotFound(t *testing.T) {
	handler, db := setupEpicTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionEpicRead,
		Data: map[string]interface{}{
			"ticketId": "nonexistent-id",
		},
	}

	handler.handleEpicRead(c, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestEpicHandler_List_Success(t *testing.T) {
	handler, db := setupEpicTestHandler(t)
	defer db.Close()

	// Create multiple epics
	for i := 0; i < 3; i++ {
		ticketID := uuid.New().String()
		createTestTicket(t, db, ticketID, i+1, "Epic "+string(rune(i+1)), "EPIC-"+string(rune(i+1)), true, "Epic Name "+string(rune(i+1)), models.EpicColorGhola)
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionEpicList,
		Data:   map[string]interface{}{},
	}

	handler.handleEpicList(c, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.Equal(t, float64(3), resp.Data["count"])
}

func TestEpicHandler_Modify_Success(t *testing.T) {
	handler, db := setupEpicTestHandler(t)
	defer db.Close()

	// Create an epic
	ticketID := uuid.New().String()
	createTestTicket(t, db, ticketID, 1, "Epic Ticket", "EPIC-1", true, "Old Name", models.EpicColorGhola)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionEpicModify,
		Data: map[string]interface{}{
			"ticketId":  ticketID,
			"epicName":  "New Name",
			"epicColor": models.EpicColorJungle,
		},
	}

	handler.handleEpicModify(c, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.True(t, resp.Data["updated"].(bool))
}

func TestEpicHandler_Remove_Success(t *testing.T) {
	handler, db := setupEpicTestHandler(t)
	defer db.Close()

	// Create an epic
	ticketID := uuid.New().String()
	createTestTicket(t, db, ticketID, 1, "Epic Ticket", "EPIC-1", true, "My Epic", models.EpicColorGhola)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionEpicRemove,
		Data: map[string]interface{}{
			"ticketId": ticketID,
		},
	}

	handler.handleEpicRemove(c, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.True(t, resp.Data["removed"].(bool))
}

func TestEpicHandler_AddStory_Success(t *testing.T) {
	handler, db := setupEpicTestHandler(t)
	defer db.Close()

	// Create an epic
	epicID := uuid.New().String()
	createTestTicket(t, db, epicID, 1, "Epic Ticket", "EPIC-1", true, "My Epic", models.EpicColorGhola)

	// Create a story ticket
	storyID := uuid.New().String()
	createTestTicket(t, db, storyID, 2, "Story Ticket", "STORY-1", false, "", "")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionEpicAddStory,
		Data: map[string]interface{}{
			"epicId":  epicID,
			"storyId": storyID,
		},
	}

	handler.handleEpicAddStory(c, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.True(t, resp.Data["added"].(bool))
}

func TestEpicHandler_RemoveStory_Success(t *testing.T) {
	handler, db := setupEpicTestHandler(t)
	defer db.Close()

	// Create an epic
	epicID := uuid.New().String()
	createTestTicket(t, db, epicID, 1, "Epic Ticket", "EPIC-1", true, "My Epic", models.EpicColorGhola)

	// Create a story linked to epic
	storyID := uuid.New().String()
	var projectID, typeID, statusID string
	db.QueryRow(context.Background(), "SELECT id FROM workflow LIMIT 1").Scan(&projectID)
	db.QueryRow(context.Background(), "SELECT id FROM ticket_type WHERE title = 'task' LIMIT 1").Scan(&typeID)
	db.QueryRow(context.Background(), "SELECT id FROM ticket_status WHERE title = 'open' LIMIT 1").Scan(&statusID)

	_, err := db.Exec(context.Background(),
		"INSERT INTO ticket (id, ticket_number, title, ticket_key, ticket_type_id, ticket_status_id, project_id, creator, epic_id, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		storyID, 2, "Story Ticket", "STORY-1", typeID, statusID, projectID, "testuser", epicID, time.Now().Unix(), time.Now().Unix(), false)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionEpicRemoveStory,
		Data: map[string]interface{}{
			"storyId": storyID,
		},
	}

	handler.handleEpicRemoveStory(c, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err = json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.True(t, resp.Data["removed"].(bool))
}

func TestEpicHandler_ListStories_Success(t *testing.T) {
	handler, db := setupEpicTestHandler(t)
	defer db.Close()

	// Create an epic
	epicID := uuid.New().String()
	createTestTicket(t, db, epicID, 1, "Epic Ticket", "EPIC-1", true, "My Epic", models.EpicColorGhola)

	// Get required IDs for creating stories
	var projectID, typeID, statusID string
	db.QueryRow(context.Background(), "SELECT id FROM workflow LIMIT 1").Scan(&projectID)
	db.QueryRow(context.Background(), "SELECT id FROM ticket_type WHERE title = 'task' LIMIT 1").Scan(&typeID)
	db.QueryRow(context.Background(), "SELECT id FROM ticket_status WHERE title = 'open' LIMIT 1").Scan(&statusID)

	// Create stories linked to epic (starting from ticket_number=2 to avoid conflict with epic)
	for i := 0; i < 3; i++ {
		storyID := uuid.New().String()
		_, err := db.Exec(context.Background(),
			"INSERT INTO ticket (id, ticket_number, title, ticket_key, ticket_type_id, ticket_status_id, project_id, creator, epic_id, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
			storyID, i+2, "Story "+string(rune(i+1)), "STORY-"+string(rune(i+1)), typeID, statusID, projectID, "testuser", epicID, time.Now().Unix(), time.Now().Unix(), false)
		require.NoError(t, err)
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionEpicListStories,
		Data: map[string]interface{}{
			"epicId": epicID,
		},
	}

	handler.handleEpicListStories(c, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.Equal(t, float64(3), resp.Data["count"])
}

func TestEpicHandler_Unauthorized(t *testing.T) {
	handler, db := setupEpicTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	// No username set

	req := &models.Request{
		Action: models.ActionEpicCreate,
		Data:   map[string]interface{}{},
	}

	handler.handleEpicCreate(c, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
