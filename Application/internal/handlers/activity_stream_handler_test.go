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

func setupActivityStreamTestHandler(t *testing.T) (*Handler, database.Database) {
	db, err := database.NewDatabase(config.DatabaseConfig{
		Type:       "sqlite",
		SQLitePath: ":memory:",
	})
	require.NoError(t, err)

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

// Helper to create audit entries for testing
func createTestAuditEntry(t *testing.T, db database.Database, userID, entityID, entityType, action, activityType string, isPublic bool) string {
	auditID := uuid.New().String()
	now := time.Now().Unix()
	_, err := db.Exec(context.Background(),
		"INSERT INTO audit (id, action, user_id, entity_id, entity_type, details, is_public, activity_type, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		auditID, action, userID, entityID, entityType, "{}", isPublic, activityType, now, now, false)
	require.NoError(t, err)
	return auditID
}

// ============================================================================
// ActionActivityStreamGet Tests
// ============================================================================

func TestActivityStreamHandler_Get_Success(t *testing.T) {
	handler, db := setupActivityStreamTestHandler(t)
	defer db.Close()

	// Create multiple audit entries
	for i := 1; i <= 5; i++ {
		entityID := uuid.New().String()
		createTestAuditEntry(t, db, "user"+string(rune('0'+i)), entityID, "ticket", "create", "ticket_created", true)
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionActivityStreamGet,
		Data:   map[string]interface{}{},
	}

	handler.handleActivityStreamGet(c, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.Equal(t, float64(5), resp.Data["count"])

	activities := resp.Data["activities"].([]interface{})
	assert.Len(t, activities, 5)
}

func TestActivityStreamHandler_Get_WithPagination(t *testing.T) {
	handler, db := setupActivityStreamTestHandler(t)
	defer db.Close()

	// Create multiple audit entries
	for i := 1; i <= 10; i++ {
		entityID := uuid.New().String()
		createTestAuditEntry(t, db, "user1", entityID, "ticket", "create", "ticket_created", true)
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionActivityStreamGet,
		Data: map[string]interface{}{
			"limit":  float64(5),
			"offset": float64(0),
		},
	}

	handler.handleActivityStreamGet(c, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.Equal(t, float64(5), resp.Data["count"])
	assert.Equal(t, float64(5), resp.Data["limit"])
	assert.Equal(t, float64(0), resp.Data["offset"])
}

func TestActivityStreamHandler_Get_OnlyPublic(t *testing.T) {
	handler, db := setupActivityStreamTestHandler(t)
	defer db.Close()

	// Create public and private audit entries
	createTestAuditEntry(t, db, "user1", uuid.New().String(), "ticket", "create", "ticket_created", true)
	createTestAuditEntry(t, db, "user1", uuid.New().String(), "ticket", "create", "ticket_created", true)
	createTestAuditEntry(t, db, "user1", uuid.New().String(), "ticket", "internal_note", "note_added", false)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionActivityStreamGet,
		Data:   map[string]interface{}{},
	}

	handler.handleActivityStreamGet(c, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.Equal(t, float64(2), resp.Data["count"]) // Only 2 public entries
}

func TestActivityStreamHandler_Get_Unauthorized(t *testing.T) {
	handler, db := setupActivityStreamTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	// No username set

	req := &models.Request{
		Action: models.ActionActivityStreamGet,
		Data:   map[string]interface{}{},
	}

	handler.handleActivityStreamGet(c, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// ============================================================================
// ActionActivityStreamGetByProject Tests
// ============================================================================

func TestActivityStreamHandler_GetByProject_Success(t *testing.T) {
	handler, db := setupActivityStreamTestHandler(t)
	defer db.Close()

	// Create a project and ticket
	var workflowID string
	db.QueryRow(context.Background(), "SELECT id FROM workflow LIMIT 1").Scan(&workflowID)

	projectID := uuid.New().String()
	_, err := db.Exec(context.Background(),
		"INSERT INTO project (id, identifier, title, workflow_id, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?)",
		projectID, "PROJ", "Test Project", workflowID, time.Now().Unix(), time.Now().Unix(), false)
	require.NoError(t, err)

	// Create audit entry with project entity
	createTestAuditEntry(t, db, "user1", projectID, "project", "create", "project_created", true)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionActivityStreamGetByProject,
		Data: map[string]interface{}{
			"projectId": projectID,
		},
	}

	handler.handleActivityStreamGetByProject(c, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err = json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.Equal(t, projectID, resp.Data["projectId"])
}

func TestActivityStreamHandler_GetByProject_MissingProjectID(t *testing.T) {
	handler, db := setupActivityStreamTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionActivityStreamGetByProject,
		Data:   map[string]interface{}{},
	}

	handler.handleActivityStreamGetByProject(c, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeMissingData, resp.ErrorCode)
}

// ============================================================================
// ActionActivityStreamGetByUser Tests
// ============================================================================

func TestActivityStreamHandler_GetByUser_Success(t *testing.T) {
	handler, db := setupActivityStreamTestHandler(t)
	defer db.Close()

	// Create audit entries for specific user
	userID := "testuser123"
	for i := 1; i <= 3; i++ {
		entityID := uuid.New().String()
		createTestAuditEntry(t, db, userID, entityID, "ticket", "create", "ticket_created", true)
	}

	// Create audit entry for different user
	createTestAuditEntry(t, db, "otheruser", uuid.New().String(), "ticket", "create", "ticket_created", true)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionActivityStreamGetByUser,
		Data: map[string]interface{}{
			"userId": userID,
		},
	}

	handler.handleActivityStreamGetByUser(c, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.Equal(t, userID, resp.Data["userId"])
	assert.Equal(t, float64(3), resp.Data["count"])
}

func TestActivityStreamHandler_GetByUser_MissingUserID(t *testing.T) {
	handler, db := setupActivityStreamTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionActivityStreamGetByUser,
		Data:   map[string]interface{}{},
	}

	handler.handleActivityStreamGetByUser(c, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeMissingData, resp.ErrorCode)
}

// ============================================================================
// ActionActivityStreamGetByTicket Tests
// ============================================================================

func TestActivityStreamHandler_GetByTicket_Success(t *testing.T) {
	handler, db := setupActivityStreamTestHandler(t)
	defer db.Close()

	ticketID := uuid.New().String()

	// Create audit entries for specific ticket
	for i := 1; i <= 3; i++ {
		createTestAuditEntry(t, db, "user"+string(rune('0'+i)), ticketID, "ticket", "modify", "ticket_updated", true)
	}

	// Create audit entry for different ticket
	createTestAuditEntry(t, db, "user1", uuid.New().String(), "ticket", "modify", "ticket_updated", true)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionActivityStreamGetByTicket,
		Data: map[string]interface{}{
			"ticketId": ticketID,
		},
	}

	handler.handleActivityStreamGetByTicket(c, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.Equal(t, ticketID, resp.Data["ticketId"])
	assert.Equal(t, float64(3), resp.Data["count"])
}

func TestActivityStreamHandler_GetByTicket_MissingTicketID(t *testing.T) {
	handler, db := setupActivityStreamTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionActivityStreamGetByTicket,
		Data:   map[string]interface{}{},
	}

	handler.handleActivityStreamGetByTicket(c, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeMissingData, resp.ErrorCode)
}

// ============================================================================
// ActionActivityStreamFilter Tests
// ============================================================================

func TestActivityStreamHandler_Filter_Success(t *testing.T) {
	handler, db := setupActivityStreamTestHandler(t)
	defer db.Close()

	// Create audit entries with different activity types
	createTestAuditEntry(t, db, "user1", uuid.New().String(), "ticket", "create", "ticket_created", true)
	createTestAuditEntry(t, db, "user1", uuid.New().String(), "ticket", "create", "ticket_created", true)
	createTestAuditEntry(t, db, "user1", uuid.New().String(), "comment", "create", "comment_added", true)
	createTestAuditEntry(t, db, "user1", uuid.New().String(), "comment", "create", "comment_added", true)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionActivityStreamFilter,
		Data: map[string]interface{}{
			"activityType": "ticket_created",
		},
	}

	handler.handleActivityStreamFilter(c, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.Equal(t, "ticket_created", resp.Data["activityType"])
	assert.Equal(t, float64(2), resp.Data["count"])
}

func TestActivityStreamHandler_Filter_MissingActivityType(t *testing.T) {
	handler, db := setupActivityStreamTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionActivityStreamFilter,
		Data:   map[string]interface{}{},
	}

	handler.handleActivityStreamFilter(c, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeMissingData, resp.ErrorCode)
}

func TestActivityStreamHandler_Filter_WithUserFilter(t *testing.T) {
	handler, db := setupActivityStreamTestHandler(t)
	defer db.Close()

	// Create audit entries
	createTestAuditEntry(t, db, "user1", uuid.New().String(), "ticket", "create", "ticket_created", true)
	createTestAuditEntry(t, db, "user2", uuid.New().String(), "ticket", "create", "ticket_created", true)
	createTestAuditEntry(t, db, "user1", uuid.New().String(), "comment", "create", "comment_added", true)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionActivityStreamFilter,
		Data: map[string]interface{}{
			"activityType": "ticket_created",
			"userId":       "user1",
		},
	}

	handler.handleActivityStreamFilter(c, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.Equal(t, float64(1), resp.Data["count"]) // Only user1's ticket_created
}

func TestActivityStreamHandler_Filter_WithEntityTypeFilter(t *testing.T) {
	handler, db := setupActivityStreamTestHandler(t)
	defer db.Close()

	// Create audit entries
	createTestAuditEntry(t, db, "user1", uuid.New().String(), "ticket", "create", "ticket_created", true)
	createTestAuditEntry(t, db, "user1", uuid.New().String(), "ticket", "create", "ticket_created", true)
	createTestAuditEntry(t, db, "user1", uuid.New().String(), "project", "create", "ticket_created", true)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionActivityStreamFilter,
		Data: map[string]interface{}{
			"activityType": "ticket_created",
			"entityType":   "ticket",
		},
	}

	handler.handleActivityStreamFilter(c, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.Equal(t, float64(2), resp.Data["count"]) // Only ticket entity types
}
