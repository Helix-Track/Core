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

func setupNotificationTestHandler(t *testing.T) (*Handler, database.Database) {
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

// ============================================================================
// ActionNotificationSchemeCreate Tests
// ============================================================================

func TestNotificationHandler_SchemeCreate_Success(t *testing.T) {
	handler, db := setupNotificationTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionNotificationSchemeCreate,
		Data: map[string]interface{}{
			"title":       "Default Scheme",
			"description": "Default notification scheme",
		},
	}

	handler.handleNotificationSchemeCreate(c, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.Equal(t, "Default Scheme", resp.Data["title"])
	assert.NotEmpty(t, resp.Data["id"])
}

func TestNotificationHandler_SchemeCreate_WithProject(t *testing.T) {
	handler, db := setupNotificationTestHandler(t)
	defer db.Close()

	// Create a project
	var workflowID string
	db.QueryRow(context.Background(), "SELECT id FROM workflow LIMIT 1").Scan(&workflowID)

	projectID := uuid.New().String()
	_, err := db.Exec(context.Background(),
		"INSERT INTO project (id, identifier, title, workflow_id, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?)",
		projectID, "PROJ", "Test Project", workflowID, time.Now().Unix(), time.Now().Unix(), false)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionNotificationSchemeCreate,
		Data: map[string]interface{}{
			"title":       "Project Scheme",
			"description": "Project-specific scheme",
			"projectId":   projectID,
		},
	}

	handler.handleNotificationSchemeCreate(c, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp models.Response
	err = json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.Equal(t, projectID, resp.Data["projectId"])
}

func TestNotificationHandler_SchemeCreate_MissingTitle(t *testing.T) {
	handler, db := setupNotificationTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionNotificationSchemeCreate,
		Data:   map[string]interface{}{},
	}

	handler.handleNotificationSchemeCreate(c, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeMissingData, resp.ErrorCode)
}

// ============================================================================
// ActionNotificationSchemeRead Tests
// ============================================================================

func TestNotificationHandler_SchemeRead_Success(t *testing.T) {
	handler, db := setupNotificationTestHandler(t)
	defer db.Close()

	// Create a scheme
	schemeID := uuid.New().String()
	now := time.Now().Unix()
	_, err := db.Exec(context.Background(),
		"INSERT INTO notification_scheme (id, title, description, project_id, created, modified, deleted) VALUES (?, ?, ?, NULL, ?, ?, ?)",
		schemeID, "Test Scheme", "Test Description", now, now, false)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionNotificationSchemeRead,
		Data: map[string]interface{}{
			"schemeId": schemeID,
		},
	}

	handler.handleNotificationSchemeRead(c, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err = json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.Equal(t, schemeID, resp.Data["id"])
	assert.Equal(t, "Test Scheme", resp.Data["title"])
}

func TestNotificationHandler_SchemeRead_NotFound(t *testing.T) {
	handler, db := setupNotificationTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionNotificationSchemeRead,
		Data: map[string]interface{}{
			"schemeId": "nonexistent-id",
		},
	}

	handler.handleNotificationSchemeRead(c, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// ============================================================================
// ActionNotificationSchemeList Tests
// ============================================================================

func TestNotificationHandler_SchemeList_Success(t *testing.T) {
	handler, db := setupNotificationTestHandler(t)
	defer db.Close()

	// Create multiple schemes
	now := time.Now().Unix()
	for i := 1; i <= 3; i++ {
		schemeID := uuid.New().String()
		_, err := db.Exec(context.Background(),
			"INSERT INTO notification_scheme (id, title, description, project_id, created, modified, deleted) VALUES (?, ?, ?, NULL, ?, ?, ?)",
			schemeID, "Scheme "+string(rune('0'+i)), "Description "+string(rune('0'+i)), now, now, false)
		require.NoError(t, err)
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionNotificationSchemeList,
		Data:   map[string]interface{}{},
	}

	handler.handleNotificationSchemeList(c, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.Equal(t, float64(3), resp.Data["count"])
}

// ============================================================================
// ActionNotificationSchemeModify Tests
// ============================================================================

func TestNotificationHandler_SchemeModify_Success(t *testing.T) {
	handler, db := setupNotificationTestHandler(t)
	defer db.Close()

	// Create a scheme
	schemeID := uuid.New().String()
	now := time.Now().Unix()
	_, err := db.Exec(context.Background(),
		"INSERT INTO notification_scheme (id, title, description, project_id, created, modified, deleted) VALUES (?, ?, ?, NULL, ?, ?, ?)",
		schemeID, "Old Title", "Old Description", now, now, false)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionNotificationSchemeModify,
		Data: map[string]interface{}{
			"schemeId":    schemeID,
			"title":       "New Title",
			"description": "New Description",
		},
	}

	handler.handleNotificationSchemeModify(c, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err = json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.True(t, resp.Data["updated"].(bool))
}

// ============================================================================
// ActionNotificationSchemeRemove Tests
// ============================================================================

func TestNotificationHandler_SchemeRemove_Success(t *testing.T) {
	handler, db := setupNotificationTestHandler(t)
	defer db.Close()

	// Create a scheme
	schemeID := uuid.New().String()
	now := time.Now().Unix()
	_, err := db.Exec(context.Background(),
		"INSERT INTO notification_scheme (id, title, description, project_id, created, modified, deleted) VALUES (?, ?, ?, NULL, ?, ?, ?)",
		schemeID, "To Delete", "Will be deleted", now, now, false)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionNotificationSchemeRemove,
		Data: map[string]interface{}{
			"schemeId": schemeID,
		},
	}

	handler.handleNotificationSchemeRemove(c, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err = json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.True(t, resp.Data["removed"].(bool))
}

// ============================================================================
// ActionNotificationSchemeAddRule Tests
// ============================================================================

func TestNotificationHandler_SchemeAddRule_Success(t *testing.T) {
	handler, db := setupNotificationTestHandler(t)
	defer db.Close()

	// Create a scheme
	schemeID := uuid.New().String()
	now := time.Now().Unix()
	_, err := db.Exec(context.Background(),
		"INSERT INTO notification_scheme (id, title, description, project_id, created, modified, deleted) VALUES (?, ?, ?, NULL, ?, ?, ?)",
		schemeID, "Test Scheme", "Test", now, now, false)
	require.NoError(t, err)

	// Create a notification event
	eventID := uuid.New().String()
	_, err = db.Exec(context.Background(),
		"INSERT INTO notification_event (id, event_type, title, description, created, deleted) VALUES (?, ?, ?, ?, ?, ?)",
		eventID, "issue_created", "Issue Created", "Triggered when issue is created", now, false)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionNotificationSchemeAddRule,
		Data: map[string]interface{}{
			"schemeId":      schemeID,
			"eventId":       eventID,
			"recipientType": "assignee",
		},
	}

	handler.handleNotificationSchemeAddRule(c, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp models.Response
	err = json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.True(t, resp.Data["added"].(bool))
	assert.NotEmpty(t, resp.Data["ruleId"])
}

// ============================================================================
// ActionNotificationSchemeRemoveRule Tests
// ============================================================================

func TestNotificationHandler_SchemeRemoveRule_Success(t *testing.T) {
	handler, db := setupNotificationTestHandler(t)
	defer db.Close()

	// Create a rule
	ruleID := uuid.New().String()
	schemeID := uuid.New().String()
	eventID := uuid.New().String()
	now := time.Now().Unix()

	// Create scheme and event first
	_, err := db.Exec(context.Background(),
		"INSERT INTO notification_scheme (id, title, description, project_id, created, modified, deleted) VALUES (?, ?, ?, NULL, ?, ?, ?)",
		schemeID, "Test Scheme", "Test", now, now, false)
	require.NoError(t, err)

	_, err = db.Exec(context.Background(),
		"INSERT INTO notification_event (id, event_type, title, description, created, deleted) VALUES (?, ?, ?, ?, ?, ?)",
		eventID, "issue_created", "Issue Created", "Test", now, false)
	require.NoError(t, err)

	_, err = db.Exec(context.Background(),
		"INSERT INTO notification_rule (id, notification_scheme_id, notification_event_id, recipient_type, recipient_id, created, deleted) VALUES (?, ?, ?, ?, NULL, ?, ?)",
		ruleID, schemeID, eventID, "assignee", now, false)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionNotificationSchemeRemoveRule,
		Data: map[string]interface{}{
			"ruleId": ruleID,
		},
	}

	handler.handleNotificationSchemeRemoveRule(c, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err = json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.True(t, resp.Data["removed"].(bool))
}

// ============================================================================
// ActionNotificationSchemeListRules Tests
// ============================================================================

func TestNotificationHandler_SchemeListRules_Success(t *testing.T) {
	handler, db := setupNotificationTestHandler(t)
	defer db.Close()

	schemeID := uuid.New().String()
	eventID := uuid.New().String()
	now := time.Now().Unix()

	// Create scheme and event
	_, err := db.Exec(context.Background(),
		"INSERT INTO notification_scheme (id, title, description, project_id, created, modified, deleted) VALUES (?, ?, ?, NULL, ?, ?, ?)",
		schemeID, "Test Scheme", "Test", now, now, false)
	require.NoError(t, err)

	_, err = db.Exec(context.Background(),
		"INSERT INTO notification_event (id, event_type, title, description, created, deleted) VALUES (?, ?, ?, ?, ?, ?)",
		eventID, "issue_created", "Issue Created", "Test", now, false)
	require.NoError(t, err)

	// Create multiple rules
	for i := 1; i <= 2; i++ {
		ruleID := uuid.New().String()
		_, err = db.Exec(context.Background(),
			"INSERT INTO notification_rule (id, notification_scheme_id, notification_event_id, recipient_type, recipient_id, created, deleted) VALUES (?, ?, ?, ?, NULL, ?, ?)",
			ruleID, schemeID, eventID, "assignee", now, false)
		require.NoError(t, err)
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionNotificationSchemeListRules,
		Data: map[string]interface{}{
			"schemeId": schemeID,
		},
	}

	handler.handleNotificationSchemeListRules(c, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err = json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.Equal(t, float64(2), resp.Data["count"])
}

// ============================================================================
// ActionNotificationEventList Tests
// ============================================================================

func TestNotificationHandler_EventList_Success(t *testing.T) {
	handler, db := setupNotificationTestHandler(t)
	defer db.Close()

	// Create notification events
	now := time.Now().Unix()
	events := []string{"issue_created", "issue_updated", "comment_added"}
	for _, eventType := range events {
		eventID := uuid.New().String()
		_, err := db.Exec(context.Background(),
			"INSERT INTO notification_event (id, event_type, title, description, created, deleted) VALUES (?, ?, ?, ?, ?, ?)",
			eventID, eventType, eventType, "Test event", now, false)
		require.NoError(t, err)
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionNotificationEventList,
		Data:   map[string]interface{}{},
	}

	handler.handleNotificationEventList(c, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.Equal(t, float64(3), resp.Data["count"])
}

// ============================================================================
// ActionNotificationSend Tests
// ============================================================================

func TestNotificationHandler_Send_Success(t *testing.T) {
	handler, db := setupNotificationTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionNotificationSend,
		Data: map[string]interface{}{
			"recipientId": "user123",
			"message":     "Test notification message",
		},
	}

	handler.handleNotificationSend(c, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.True(t, resp.Data["sent"].(bool))
	assert.Equal(t, "user123", resp.Data["recipientId"])
}

func TestNotificationHandler_Send_MissingRecipient(t *testing.T) {
	handler, db := setupNotificationTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionNotificationSend,
		Data: map[string]interface{}{
			"message": "Test message",
		},
	}

	handler.handleNotificationSend(c, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeMissingData, resp.ErrorCode)
}
