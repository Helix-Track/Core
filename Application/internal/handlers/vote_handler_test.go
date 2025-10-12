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

func setupVoteTestHandler(t *testing.T) (*Handler, database.Database) {
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

// Helper to create a test ticket for vote tests
func createTestTicketForVote(t *testing.T, db database.Database, ticketID string, ticketNumber int, title, key string) {
	var projectID, typeID, statusID string
	db.QueryRow(context.Background(), "SELECT id FROM workflow LIMIT 1").Scan(&projectID)
	db.QueryRow(context.Background(), "SELECT id FROM ticket_type WHERE title = 'task' LIMIT 1").Scan(&typeID)
	db.QueryRow(context.Background(), "SELECT id FROM ticket_status WHERE title = 'open' LIMIT 1").Scan(&statusID)

	_, err := db.Exec(context.Background(),
		"INSERT INTO ticket (id, ticket_number, title, ticket_key, ticket_type_id, ticket_status_id, project_id, creator, vote_count, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		ticketID, ticketNumber, title, key, typeID, statusID, projectID, "testuser", 0, time.Now().Unix(), time.Now().Unix(), false)
	require.NoError(t, err)
}

// ============================================================================
// ActionVoteAdd Tests
// ============================================================================

func TestVoteHandler_Add_Success(t *testing.T) {
	handler, db := setupVoteTestHandler(t)
	defer db.Close()

	// Create a test ticket first
	ticketID := uuid.New().String()
	createTestTicketForVote(t, db, ticketID, 1, "Test Ticket", "TEST-1")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionVoteAdd,
		Data: map[string]interface{}{
			"ticketId": ticketID,
		},
	}

	handler.handleVoteAdd(c, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.Equal(t, ticketID, resp.Data["ticketId"])
	assert.Equal(t, "testuser", resp.Data["userId"])
	assert.NotEmpty(t, resp.Data["id"])

	// Verify vote count was updated
	var voteCount int
	db.QueryRow(context.Background(), "SELECT vote_count FROM ticket WHERE id = ?", ticketID).Scan(&voteCount)
	assert.Equal(t, 1, voteCount)
}

func TestVoteHandler_Add_MissingTicketID(t *testing.T) {
	handler, db := setupVoteTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionVoteAdd,
		Data:   map[string]interface{}{},
	}

	handler.handleVoteAdd(c, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeMissingData, resp.ErrorCode)
}

func TestVoteHandler_Add_TicketNotFound(t *testing.T) {
	handler, db := setupVoteTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionVoteAdd,
		Data: map[string]interface{}{
			"ticketId": "nonexistent-id",
		},
	}

	handler.handleVoteAdd(c, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeEntityNotFound, resp.ErrorCode)
}

func TestVoteHandler_Add_AlreadyVoted(t *testing.T) {
	handler, db := setupVoteTestHandler(t)
	defer db.Close()

	ticketID := uuid.New().String()
	createTestTicketForVote(t, db, ticketID, 1, "Test Ticket", "TEST-1")

	// Add first vote
	voteID := uuid.New().String()
	_, err := db.Exec(context.Background(),
		"INSERT INTO ticket_vote_mapping (id, ticket_id, user_id, created, deleted) VALUES (?, ?, ?, ?, ?)",
		voteID, ticketID, "testuser", time.Now().Unix(), false)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionVoteAdd,
		Data: map[string]interface{}{
			"ticketId": ticketID,
		},
	}

	handler.handleVoteAdd(c, req)

	assert.Equal(t, http.StatusConflict, w.Code)

	var resp models.Response
	err = json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeEntityAlreadyExists, resp.ErrorCode)
}

func TestVoteHandler_Add_Unauthorized(t *testing.T) {
	handler, db := setupVoteTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	// No username set

	req := &models.Request{
		Action: models.ActionVoteAdd,
		Data:   map[string]interface{}{},
	}

	handler.handleVoteAdd(c, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// ============================================================================
// ActionVoteRemove Tests
// ============================================================================

func TestVoteHandler_Remove_Success(t *testing.T) {
	handler, db := setupVoteTestHandler(t)
	defer db.Close()

	ticketID := uuid.New().String()
	createTestTicketForVote(t, db, ticketID, 1, "Test Ticket", "TEST-1")

	// Add a vote first
	voteID := uuid.New().String()
	_, err := db.Exec(context.Background(),
		"INSERT INTO ticket_vote_mapping (id, ticket_id, user_id, created, deleted) VALUES (?, ?, ?, ?, ?)",
		voteID, ticketID, "testuser", time.Now().Unix(), false)
	require.NoError(t, err)

	// Update vote count
	_, err = db.Exec(context.Background(), "UPDATE ticket SET vote_count = 1 WHERE id = ?", ticketID)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionVoteRemove,
		Data: map[string]interface{}{
			"ticketId": ticketID,
		},
	}

	handler.handleVoteRemove(c, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err = json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.True(t, resp.Data["deleted"].(bool))

	// Verify vote count was updated
	var voteCount int
	db.QueryRow(context.Background(), "SELECT vote_count FROM ticket WHERE id = ?", ticketID).Scan(&voteCount)
	assert.Equal(t, 0, voteCount)
}

func TestVoteHandler_Remove_MissingTicketID(t *testing.T) {
	handler, db := setupVoteTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionVoteRemove,
		Data:   map[string]interface{}{},
	}

	handler.handleVoteRemove(c, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeMissingData, resp.ErrorCode)
}

func TestVoteHandler_Remove_VoteNotFound(t *testing.T) {
	handler, db := setupVoteTestHandler(t)
	defer db.Close()

	ticketID := uuid.New().String()
	createTestTicketForVote(t, db, ticketID, 1, "Test Ticket", "TEST-1")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionVoteRemove,
		Data: map[string]interface{}{
			"ticketId": ticketID,
		},
	}

	handler.handleVoteRemove(c, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeEntityNotFound, resp.ErrorCode)
}

func TestVoteHandler_Remove_Unauthorized(t *testing.T) {
	handler, db := setupVoteTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	// No username set

	req := &models.Request{
		Action: models.ActionVoteRemove,
		Data:   map[string]interface{}{},
	}

	handler.handleVoteRemove(c, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// ============================================================================
// ActionVoteCount Tests
// ============================================================================

func TestVoteHandler_Count_Success(t *testing.T) {
	handler, db := setupVoteTestHandler(t)
	defer db.Close()

	ticketID := uuid.New().String()
	createTestTicketForVote(t, db, ticketID, 1, "Test Ticket", "TEST-1")

	// Set vote count
	_, err := db.Exec(context.Background(), "UPDATE ticket SET vote_count = 5 WHERE id = ?", ticketID)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionVoteCount,
		Data: map[string]interface{}{
			"ticketId": ticketID,
		},
	}

	handler.handleVoteCount(c, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err = json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.Equal(t, ticketID, resp.Data["ticketId"])
	assert.Equal(t, float64(5), resp.Data["voteCount"])
}

func TestVoteHandler_Count_TicketNotFound(t *testing.T) {
	handler, db := setupVoteTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionVoteCount,
		Data: map[string]interface{}{
			"ticketId": "nonexistent-id",
		},
	}

	handler.handleVoteCount(c, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeEntityNotFound, resp.ErrorCode)
}

// ============================================================================
// ActionVoteList Tests
// ============================================================================

func TestVoteHandler_List_Success(t *testing.T) {
	handler, db := setupVoteTestHandler(t)
	defer db.Close()

	ticketID := uuid.New().String()
	createTestTicketForVote(t, db, ticketID, 1, "Test Ticket", "TEST-1")

	// Add multiple votes
	for i := 1; i <= 3; i++ {
		voteID := uuid.New().String()
		userID := "user" + string(rune('0'+i))
		_, err := db.Exec(context.Background(),
			"INSERT INTO ticket_vote_mapping (id, ticket_id, user_id, created, deleted) VALUES (?, ?, ?, ?, ?)",
			voteID, ticketID, userID, time.Now().Unix(), false)
		require.NoError(t, err)
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionVoteList,
		Data: map[string]interface{}{
			"ticketId": ticketID,
		},
	}

	handler.handleVoteList(c, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.Equal(t, ticketID, resp.Data["ticketId"])
	assert.Equal(t, float64(3), resp.Data["count"])

	votes := resp.Data["votes"].([]interface{})
	assert.Len(t, votes, 3)
}

func TestVoteHandler_List_Empty(t *testing.T) {
	handler, db := setupVoteTestHandler(t)
	defer db.Close()

	ticketID := uuid.New().String()
	createTestTicketForVote(t, db, ticketID, 1, "Test Ticket", "TEST-1")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionVoteList,
		Data: map[string]interface{}{
			"ticketId": ticketID,
		},
	}

	handler.handleVoteList(c, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.Equal(t, float64(0), resp.Data["count"])
}

// ============================================================================
// ActionVoteCheck Tests
// ============================================================================

func TestVoteHandler_Check_HasVoted(t *testing.T) {
	handler, db := setupVoteTestHandler(t)
	defer db.Close()

	ticketID := uuid.New().String()
	createTestTicketForVote(t, db, ticketID, 1, "Test Ticket", "TEST-1")

	// Add vote for testuser
	voteID := uuid.New().String()
	_, err := db.Exec(context.Background(),
		"INSERT INTO ticket_vote_mapping (id, ticket_id, user_id, created, deleted) VALUES (?, ?, ?, ?, ?)",
		voteID, ticketID, "testuser", time.Now().Unix(), false)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionVoteCheck,
		Data: map[string]interface{}{
			"ticketId": ticketID,
		},
	}

	handler.handleVoteCheck(c, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err = json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.Equal(t, ticketID, resp.Data["ticketId"])
	assert.Equal(t, "testuser", resp.Data["userId"])
	assert.True(t, resp.Data["hasVoted"].(bool))
}

func TestVoteHandler_Check_NotVoted(t *testing.T) {
	handler, db := setupVoteTestHandler(t)
	defer db.Close()

	ticketID := uuid.New().String()
	createTestTicketForVote(t, db, ticketID, 1, "Test Ticket", "TEST-1")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionVoteCheck,
		Data: map[string]interface{}{
			"ticketId": ticketID,
		},
	}

	handler.handleVoteCheck(c, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.False(t, resp.Data["hasVoted"].(bool))
}
