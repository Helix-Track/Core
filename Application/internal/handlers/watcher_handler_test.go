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

// setupWatcherTestHandler creates a test handler with watcher table and dependencies
func setupWatcherTestHandler(t *testing.T) *Handler {
	handler := setupTestHandler(t)

	// Create ticket_watcher_mapping table
	_, err := handler.db.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS ticket_watcher_mapping (
			id TEXT PRIMARY KEY,
			ticket_id TEXT NOT NULL,
			user_id TEXT NOT NULL,
			created INTEGER NOT NULL,
			deleted INTEGER NOT NULL DEFAULT 0
		)
	`)
	require.NoError(t, err)

	// Insert test ticket for watcher tests
	_, err = handler.db.Exec(context.Background(),
		"INSERT INTO ticket (id, title, description, status, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?)",
		"test-ticket-id", "Test Ticket", "Test ticket description", "open", 1000, 1000, 0)
	require.NoError(t, err)

	return handler
}

// ============================================================================
// Watcher Add Tests
// ============================================================================

func TestWatcherHandler_Add_Success(t *testing.T) {
	handler := setupWatcherTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionWatcherAdd,
		Data: map[string]interface{}{
			"ticketId": "test-ticket-id",
			"userId":   "user123",
		},
	}

	w := performRequest(handler, "POST", "/do", reqBody)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response models.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, models.ErrorCodeNoError, response.ErrorCode)

	watcherData, ok := response.Data["watcher"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "test-ticket-id", watcherData["TicketID"])
	assert.Equal(t, "user123", watcherData["UserID"])
	assert.NotEmpty(t, watcherData["ID"])

	// Verify in database
	var count int
	err = handler.db.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM ticket_watcher_mapping WHERE ticket_id = ? AND user_id = ? AND deleted = 0",
		"test-ticket-id", "user123").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestWatcherHandler_Add_DefaultUserId(t *testing.T) {
	handler := setupWatcherTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionWatcherAdd,
		Data: map[string]interface{}{
			"ticketId": "test-ticket-id",
			// No userId - should default to current username "test-user"
		},
	}

	w := performRequest(handler, "POST", "/do", reqBody)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response models.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	watcherData, ok := response.Data["watcher"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "test-user", watcherData["UserID"]) // Default username from setupTestHandler

	// Verify in database
	var count int
	err = handler.db.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM ticket_watcher_mapping WHERE ticket_id = ? AND user_id = ? AND deleted = 0",
		"test-ticket-id", "test-user").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestWatcherHandler_Add_AlreadyWatching(t *testing.T) {
	handler := setupWatcherTestHandler(t)

	// Insert existing watcher
	_, err := handler.db.Exec(context.Background(),
		"INSERT INTO ticket_watcher_mapping (id, ticket_id, user_id, created, deleted) VALUES (?, ?, ?, ?, ?)",
		"watcher-id", "test-ticket-id", "user123", 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionWatcherAdd,
		Data: map[string]interface{}{
			"ticketId": "test-ticket-id",
			"userId":   "user123",
		},
	}

	w := performRequest(handler, "POST", "/do", reqBody)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response models.Response
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, models.ErrorCodeEntityAlreadyExists, response.ErrorCode)
	assert.Contains(t, response.ErrorMessage, "Already watching")
}

func TestWatcherHandler_Add_MissingTicketId(t *testing.T) {
	handler := setupWatcherTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionWatcherAdd,
		Data: map[string]interface{}{
			"userId": "user123",
			// Missing ticketId
		},
	}

	w := performRequest(handler, "POST", "/do", reqBody)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response models.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, models.ErrorCodeMissingData, response.ErrorCode)
	assert.Contains(t, response.ErrorMessage, "Missing ticket ID")
}

func TestWatcherHandler_Add_MultipleWatchers(t *testing.T) {
	handler := setupWatcherTestHandler(t)

	// Add multiple watchers to the same ticket
	users := []string{"user1", "user2", "user3", "user4", "user5"}

	for _, userID := range users {
		reqBody := models.Request{
			Action: models.ActionWatcherAdd,
			Data: map[string]interface{}{
				"ticketId": "test-ticket-id",
				"userId":   userID,
			},
		}

		w := performRequest(handler, "POST", "/do", reqBody)
		assert.Equal(t, http.StatusCreated, w.Code)
	}

	// Verify all watchers were added
	var count int
	err := handler.db.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM ticket_watcher_mapping WHERE ticket_id = ? AND deleted = 0",
		"test-ticket-id").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, len(users), count)
}

func TestWatcherHandler_Add_CanReaddAfterRemoval(t *testing.T) {
	handler := setupWatcherTestHandler(t)

	// Add watcher
	_, err := handler.db.Exec(context.Background(),
		"INSERT INTO ticket_watcher_mapping (id, ticket_id, user_id, created, deleted) VALUES (?, ?, ?, ?, ?)",
		"watcher-id", "test-ticket-id", "user123", 1000, 1)
	require.NoError(t, err)

	// Try to add again (deleted = 1, so should allow)
	reqBody := models.Request{
		Action: models.ActionWatcherAdd,
		Data: map[string]interface{}{
			"ticketId": "test-ticket-id",
			"userId":   "user123",
		},
	}

	w := performRequest(handler, "POST", "/do", reqBody)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response models.Response
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, models.ErrorCodeNoError, response.ErrorCode)

	// Verify there are now 2 records: one deleted, one active
	var count int
	err = handler.db.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM ticket_watcher_mapping WHERE ticket_id = ? AND user_id = ?",
		"test-ticket-id", "user123").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 2, count)

	// Verify only one is active
	err = handler.db.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM ticket_watcher_mapping WHERE ticket_id = ? AND user_id = ? AND deleted = 0",
		"test-ticket-id", "user123").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

// ============================================================================
// Watcher Remove Tests
// ============================================================================

func TestWatcherHandler_Remove_Success(t *testing.T) {
	handler := setupWatcherTestHandler(t)

	// Insert watcher
	_, err := handler.db.Exec(context.Background(),
		"INSERT INTO ticket_watcher_mapping (id, ticket_id, user_id, created, deleted) VALUES (?, ?, ?, ?, ?)",
		"watcher-id", "test-ticket-id", "user123", 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionWatcherRemove,
		Data: map[string]interface{}{
			"ticketId": "test-ticket-id",
			"userId":   "user123",
		},
	}

	w := performRequest(handler, "POST", "/do", reqBody)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Response
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, models.ErrorCodeNoError, response.ErrorCode)
	assert.True(t, response.Data["removed"].(bool))

	// Verify soft delete in database
	var deleted bool
	err = handler.db.QueryRow(context.Background(),
		"SELECT deleted FROM ticket_watcher_mapping WHERE ticket_id = ? AND user_id = ?",
		"test-ticket-id", "user123").Scan(&deleted)
	require.NoError(t, err)
	assert.True(t, deleted)
}

func TestWatcherHandler_Remove_DefaultUserId(t *testing.T) {
	handler := setupWatcherTestHandler(t)

	// Insert watcher with default username
	_, err := handler.db.Exec(context.Background(),
		"INSERT INTO ticket_watcher_mapping (id, ticket_id, user_id, created, deleted) VALUES (?, ?, ?, ?, ?)",
		"watcher-id", "test-ticket-id", "test-user", 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionWatcherRemove,
		Data: map[string]interface{}{
			"ticketId": "test-ticket-id",
			// No userId - should default to current username "test-user"
		},
	}

	w := performRequest(handler, "POST", "/do", reqBody)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Response
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, models.ErrorCodeNoError, response.ErrorCode)
	assert.True(t, response.Data["removed"].(bool))
	assert.Equal(t, "test-user", response.Data["userId"].(string))

	// Verify soft delete
	var deleted bool
	err = handler.db.QueryRow(context.Background(),
		"SELECT deleted FROM ticket_watcher_mapping WHERE ticket_id = ? AND user_id = ?",
		"test-ticket-id", "test-user").Scan(&deleted)
	require.NoError(t, err)
	assert.True(t, deleted)
}

func TestWatcherHandler_Remove_NotFound(t *testing.T) {
	handler := setupWatcherTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionWatcherRemove,
		Data: map[string]interface{}{
			"ticketId": "test-ticket-id",
			"userId":   "non-existent-user",
		},
	}

	w := performRequest(handler, "POST", "/do", reqBody)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response models.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, models.ErrorCodeEntityNotFound, response.ErrorCode)
	assert.Contains(t, response.ErrorMessage, "Watcher not found")
}

func TestWatcherHandler_Remove_MissingTicketId(t *testing.T) {
	handler := setupWatcherTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionWatcherRemove,
		Data: map[string]interface{}{
			"userId": "user123",
			// Missing ticketId
		},
	}

	w := performRequest(handler, "POST", "/do", reqBody)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response models.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, models.ErrorCodeMissingData, response.ErrorCode)
	assert.Contains(t, response.ErrorMessage, "Missing ticket ID")
}

// ============================================================================
// Watcher List Tests
// ============================================================================

func TestWatcherHandler_List_Success(t *testing.T) {
	handler := setupWatcherTestHandler(t)

	// Insert multiple watchers
	users := []string{"user1", "user2", "user3"}
	for i, userID := range users {
		_, err := handler.db.Exec(context.Background(),
			"INSERT INTO ticket_watcher_mapping (id, ticket_id, user_id, created, deleted) VALUES (?, ?, ?, ?, ?)",
			"watcher-"+string(rune('1'+i)), "test-ticket-id", userID, 1000+int64(i), 0)
		require.NoError(t, err)
	}

	reqBody := models.Request{
		Action: models.ActionWatcherList,
		Data: map[string]interface{}{
			"ticketId": "test-ticket-id",
		},
	}

	w := performRequest(handler, "POST", "/do", reqBody)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, models.ErrorCodeNoError, response.ErrorCode)

	watchers, ok := response.Data["watchers"].([]interface{})
	require.True(t, ok)
	assert.Len(t, watchers, 3)

	count, ok := response.Data["count"].(float64)
	require.True(t, ok)
	assert.Equal(t, float64(3), count)

	ticketID, ok := response.Data["ticketId"].(string)
	require.True(t, ok)
	assert.Equal(t, "test-ticket-id", ticketID)
}

func TestWatcherHandler_List_Empty(t *testing.T) {
	handler := setupWatcherTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionWatcherList,
		Data: map[string]interface{}{
			"ticketId": "test-ticket-id",
		},
	}

	w := performRequest(handler, "POST", "/do", reqBody)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, models.ErrorCodeNoError, response.ErrorCode)

	watchers, ok := response.Data["watchers"].([]interface{})
	require.True(t, ok)
	assert.Empty(t, watchers)

	count, ok := response.Data["count"].(float64)
	require.True(t, ok)
	assert.Equal(t, float64(0), count)
}

func TestWatcherHandler_List_OrderedByCreated(t *testing.T) {
	handler := setupWatcherTestHandler(t)

	// Insert watchers in reverse order
	users := []struct {
		userID  string
		created int64
	}{
		{"user3", 3000},
		{"user1", 1000},
		{"user2", 2000},
	}

	for i, user := range users {
		_, err := handler.db.Exec(context.Background(),
			"INSERT INTO ticket_watcher_mapping (id, ticket_id, user_id, created, deleted) VALUES (?, ?, ?, ?, ?)",
			"watcher-"+string(rune('1'+i)), "test-ticket-id", user.userID, user.created, 0)
		require.NoError(t, err)
	}

	reqBody := models.Request{
		Action: models.ActionWatcherList,
		Data: map[string]interface{}{
			"ticketId": "test-ticket-id",
		},
	}

	w := performRequest(handler, "POST", "/do", reqBody)

	var response models.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	watchers, ok := response.Data["watchers"].([]interface{})
	require.True(t, ok)

	// Verify chronological order (oldest first)
	userIDs := make([]string, len(watchers))
	for i, watcher := range watchers {
		watcherMap := watcher.(map[string]interface{})
		userIDs[i] = watcherMap["UserID"].(string)
	}

	assert.Equal(t, "user1", userIDs[0])
	assert.Equal(t, "user2", userIDs[1])
	assert.Equal(t, "user3", userIDs[2])
}

func TestWatcherHandler_List_ExcludesDeleted(t *testing.T) {
	handler := setupWatcherTestHandler(t)

	// Insert active and deleted watchers
	_, err := handler.db.Exec(context.Background(),
		"INSERT INTO ticket_watcher_mapping (id, ticket_id, user_id, created, deleted) VALUES (?, ?, ?, ?, ?)",
		"watcher-1", "test-ticket-id", "active-user", 1000, 0)
	require.NoError(t, err)

	_, err = handler.db.Exec(context.Background(),
		"INSERT INTO ticket_watcher_mapping (id, ticket_id, user_id, created, deleted) VALUES (?, ?, ?, ?, ?)",
		"watcher-2", "test-ticket-id", "deleted-user", 1000, 1)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionWatcherList,
		Data: map[string]interface{}{
			"ticketId": "test-ticket-id",
		},
	}

	w := performRequest(handler, "POST", "/do", reqBody)

	var response models.Response
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	watchers, ok := response.Data["watchers"].([]interface{})
	require.True(t, ok)
	assert.Len(t, watchers, 1)

	watcherMap := watchers[0].(map[string]interface{})
	assert.Equal(t, "active-user", watcherMap["UserID"])
}

func TestWatcherHandler_List_MissingTicketId(t *testing.T) {
	handler := setupWatcherTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionWatcherList,
		Data:   map[string]interface{}{
			// Missing ticketId
		},
	}

	w := performRequest(handler, "POST", "/do", reqBody)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response models.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, models.ErrorCodeMissingData, response.ErrorCode)
	assert.Contains(t, response.ErrorMessage, "Missing ticket ID")
}

// ============================================================================
// Full Watcher Cycle Test
// ============================================================================

func TestWatcherHandler_FullCycle(t *testing.T) {
	handler := setupWatcherTestHandler(t)

	// 1. Add watcher
	addReq := models.Request{
		Action: models.ActionWatcherAdd,
		Data: map[string]interface{}{
			"ticketId": "test-ticket-id",
			"userId":   "cycle-user",
		},
	}

	w := performRequest(handler, "POST", "/do", addReq)
	assert.Equal(t, http.StatusCreated, w.Code)

	var addResp models.Response
	err := json.Unmarshal(w.Body.Bytes(), &addResp)
	require.NoError(t, err)

	watcherData := addResp.Data["watcher"].(map[string]interface{})
	assert.Equal(t, "cycle-user", watcherData["UserID"])

	// 2. List watchers (should have 1)
	listReq := models.Request{
		Action: models.ActionWatcherList,
		Data: map[string]interface{}{
			"ticketId": "test-ticket-id",
		},
	}

	w = performRequest(handler, "POST", "/do", listReq)
	var listResp models.Response
	err = json.Unmarshal(w.Body.Bytes(), &listResp)
	require.NoError(t, err)

	watchers, ok := listResp.Data["watchers"].([]interface{})
	require.True(t, ok)
	assert.Len(t, watchers, 1)

	// 3. Try to add again (should fail with already exists)
	w = performRequest(handler, "POST", "/do", addReq)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var dupResp models.Response
	err = json.Unmarshal(w.Body.Bytes(), &dupResp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeEntityAlreadyExists, dupResp.ErrorCode)

	// 4. Remove watcher
	removeReq := models.Request{
		Action: models.ActionWatcherRemove,
		Data: map[string]interface{}{
			"ticketId": "test-ticket-id",
			"userId":   "cycle-user",
		},
	}

	w = performRequest(handler, "POST", "/do", removeReq)
	assert.Equal(t, http.StatusOK, w.Code)

	var removeResp models.Response
	err = json.Unmarshal(w.Body.Bytes(), &removeResp)
	require.NoError(t, err)
	assert.True(t, removeResp.Data["removed"].(bool))

	// 5. List watchers (should be empty)
	w = performRequest(handler, "POST", "/do", listReq)
	err = json.Unmarshal(w.Body.Bytes(), &listResp)
	require.NoError(t, err)

	watchers, ok = listResp.Data["watchers"].([]interface{})
	require.True(t, ok)
	assert.Empty(t, watchers)

	// 6. Add again after removal (should succeed)
	w = performRequest(handler, "POST", "/do", addReq)
	assert.Equal(t, http.StatusCreated, w.Code)

	// 7. Verify re-added
	w = performRequest(handler, "POST", "/do", listReq)
	err = json.Unmarshal(w.Body.Bytes(), &listResp)
	require.NoError(t, err)

	watchers, ok = listResp.Data["watchers"].([]interface{})
	require.True(t, ok)
	assert.Len(t, watchers, 1)
}

// ============================================================================
// Event Publishing Tests
// ============================================================================

// TestWatcherHandler_Add_PublishesEvent tests that watcher addition publishes an event
func TestWatcherHandler_Add_PublishesEvent(t *testing.T) {
	handler, mockPublisher := setupTestHandlerWithPublisher(t)

	// Create ticket_watcher_mapping table
	_, err := handler.db.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS ticket_watcher_mapping (
			id TEXT PRIMARY KEY,
			ticket_id TEXT NOT NULL,
			user_id TEXT NOT NULL,
			created INTEGER NOT NULL,
			deleted INTEGER NOT NULL DEFAULT 0
		)
	`)
	require.NoError(t, err)

	// Insert test ticket with project_id for hierarchical context
	_, err = handler.db.Exec(context.Background(),
		"INSERT INTO ticket (id, title, description, status, project_id, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		"test-ticket-id", "Test Ticket", "Test ticket description", "open", "project-123", 1000, 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionWatcherAdd,
		Data: map[string]interface{}{
			"ticketId": "test-ticket-id",
			"userId":   "user123",
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

	// Verify event was published
	assert.Equal(t, 1, mockPublisher.GetEventCount())
	lastCall := mockPublisher.GetLastEntityCall()
	require.NotNil(t, lastCall)

	// Verify event details
	assert.Equal(t, models.ActionCreate, lastCall.Action)
	assert.Equal(t, "watcher", lastCall.Object)
	assert.Equal(t, "testuser", lastCall.Username)
	assert.NotEmpty(t, lastCall.EntityID)

	// Verify event data
	assert.Equal(t, "test-ticket-id", lastCall.Data["ticket_id"])
	assert.Equal(t, "user123", lastCall.Data["user_id"])
	assert.NotEmpty(t, lastCall.Data["id"])

	// Verify hierarchical context (project ID from parent ticket)
	assert.Equal(t, "project-123", lastCall.Context.ProjectID)
	assert.Contains(t, lastCall.Context.Permissions, "READ")
}

// TestWatcherHandler_Remove_PublishesEvent tests that watcher removal publishes an event
func TestWatcherHandler_Remove_PublishesEvent(t *testing.T) {
	handler, mockPublisher := setupTestHandlerWithPublisher(t)

	// Create ticket_watcher_mapping table
	_, err := handler.db.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS ticket_watcher_mapping (
			id TEXT PRIMARY KEY,
			ticket_id TEXT NOT NULL,
			user_id TEXT NOT NULL,
			created INTEGER NOT NULL,
			deleted INTEGER NOT NULL DEFAULT 0
		)
	`)
	require.NoError(t, err)

	// Insert test ticket with project_id for hierarchical context
	_, err = handler.db.Exec(context.Background(),
		"INSERT INTO ticket (id, title, description, status, project_id, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		"test-ticket-id", "Test Ticket", "Test ticket description", "open", "project-456", 1000, 1000, 0)
	require.NoError(t, err)

	// Insert watcher
	_, err = handler.db.Exec(context.Background(),
		"INSERT INTO ticket_watcher_mapping (id, ticket_id, user_id, created, deleted) VALUES (?, ?, ?, ?, ?)",
		"watcher-id", "test-ticket-id", "user123", 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionWatcherRemove,
		Data: map[string]interface{}{
			"ticketId": "test-ticket-id",
			"userId":   "user123",
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
	assert.Equal(t, "watcher", lastCall.Object)
	assert.Equal(t, "test-ticket-id:user123", lastCall.EntityID) // Composite ID
	assert.Equal(t, "testuser", lastCall.Username)

	// Verify event data
	assert.Equal(t, "test-ticket-id", lastCall.Data["ticket_id"])
	assert.Equal(t, "user123", lastCall.Data["user_id"])

	// Verify hierarchical context (project ID from parent ticket)
	assert.Equal(t, "project-456", lastCall.Context.ProjectID)
	assert.Contains(t, lastCall.Context.Permissions, "READ")
}

// TestWatcherHandler_Add_NoEventOnFailure tests that no event is published on add failure
func TestWatcherHandler_Add_NoEventOnFailure(t *testing.T) {
	handler, mockPublisher := setupTestHandlerWithPublisher(t)

	// Create ticket_watcher_mapping table
	_, err := handler.db.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS ticket_watcher_mapping (
			id TEXT PRIMARY KEY,
			ticket_id TEXT NOT NULL,
			user_id TEXT NOT NULL,
			created INTEGER NOT NULL,
			deleted INTEGER NOT NULL DEFAULT 0
		)
	`)
	require.NoError(t, err)

	// Insert test ticket
	_, err = handler.db.Exec(context.Background(),
		"INSERT INTO ticket (id, title, description, status, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?)",
		"test-ticket-id", "Test Ticket", "Test ticket description", "open", 1000, 1000, 0)
	require.NoError(t, err)

	// Insert existing watcher
	_, err = handler.db.Exec(context.Background(),
		"INSERT INTO ticket_watcher_mapping (id, ticket_id, user_id, created, deleted) VALUES (?, ?, ?, ?, ?)",
		"watcher-id", "test-ticket-id", "user123", 1000, 0)
	require.NoError(t, err)

	// Try to add the same watcher again (should fail with already exists)
	reqBody := models.Request{
		Action: models.ActionWatcherAdd,
		Data: map[string]interface{}{
			"ticketId": "test-ticket-id",
			"userId":   "user123",
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

// TestWatcherHandler_Remove_NoEventOnFailure tests that no event is published on remove failure
func TestWatcherHandler_Remove_NoEventOnFailure(t *testing.T) {
	handler, mockPublisher := setupTestHandlerWithPublisher(t)

	// Create ticket_watcher_mapping table
	_, err := handler.db.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS ticket_watcher_mapping (
			id TEXT PRIMARY KEY,
			ticket_id TEXT NOT NULL,
			user_id TEXT NOT NULL,
			created INTEGER NOT NULL,
			deleted INTEGER NOT NULL DEFAULT 0
		)
	`)
	require.NoError(t, err)

	// Insert test ticket (no watcher)
	_, err = handler.db.Exec(context.Background(),
		"INSERT INTO ticket (id, title, description, status, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?)",
		"test-ticket-id", "Test Ticket", "Test ticket description", "open", 1000, 1000, 0)
	require.NoError(t, err)

	// Try to remove non-existent watcher
	reqBody := models.Request{
		Action: models.ActionWatcherRemove,
		Data: map[string]interface{}{
			"ticketId": "test-ticket-id",
			"userId":   "non-existent-user",
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
