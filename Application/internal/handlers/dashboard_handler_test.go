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

func setupDashboardTestHandler(t *testing.T) (*Handler, database.Database) {
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

// Helper to create a test dashboard
func createTestDashboard(t *testing.T, db database.Database, dashboardID, title, ownerID string, isPublic, isFavorite bool) {
	_, err := db.Exec(context.Background(),
		"INSERT INTO dashboard (id, title, description, owner_id, is_public, is_favorite, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		dashboardID, title, "Test description", ownerID, isPublic, isFavorite, time.Now().Unix(), time.Now().Unix(), false)
	require.NoError(t, err)
}

// Helper to create a test widget
func createTestWidget(t *testing.T, db database.Database, widgetID, dashboardID, widgetType string) {
	_, err := db.Exec(context.Background(),
		"INSERT INTO dashboard_widget (id, dashboard_id, widget_type, title, position_x, position_y, width, height, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		widgetID, dashboardID, widgetType, "Test Widget", 0, 0, 4, 4, time.Now().Unix(), time.Now().Unix(), false)
	require.NoError(t, err)
}

// Helper to create a test share mapping
func createTestShare(t *testing.T, db database.Database, shareID, dashboardID, userID string) {
	_, err := db.Exec(context.Background(),
		"INSERT INTO dashboard_share_mapping (id, dashboard_id, user_id, team_id, project_id, created, deleted) VALUES (?, ?, ?, NULL, NULL, ?, ?)",
		shareID, dashboardID, userID, time.Now().Unix(), false)
	require.NoError(t, err)
}

// ============================================================================
// ActionDashboardCreate Tests
// ============================================================================

func TestDashboardHandler_Create_Success(t *testing.T) {
	handler, db := setupDashboardTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionDashboardCreate,
		Data: map[string]interface{}{
			"title":       "My Dashboard",
			"description": "A test dashboard",
			"isPublic":    true,
			"isFavorite":  false,
		},
	}

	handler.handleDashboardCreate(c, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.Equal(t, "My Dashboard", resp.Data["title"])
	assert.Equal(t, "testuser", resp.Data["ownerId"])
	assert.Equal(t, true, resp.Data["isPublic"])
	assert.Equal(t, false, resp.Data["isFavorite"])
}

func TestDashboardHandler_Create_Success_MinimalFields(t *testing.T) {
	handler, db := setupDashboardTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionDashboardCreate,
		Data: map[string]interface{}{
			"title": "Minimal Dashboard",
		},
	}

	handler.handleDashboardCreate(c, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.Equal(t, "Minimal Dashboard", resp.Data["title"])
}

func TestDashboardHandler_Create_Success_WithLayout(t *testing.T) {
	handler, db := setupDashboardTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionDashboardCreate,
		Data: map[string]interface{}{
			"title":  "Dashboard with Layout",
			"layout": `{"columns": 12, "rows": 8}`,
		},
	}

	handler.handleDashboardCreate(c, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
}

func TestDashboardHandler_Create_MissingTitle(t *testing.T) {
	handler, db := setupDashboardTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionDashboardCreate,
		Data:   map[string]interface{}{},
	}

	handler.handleDashboardCreate(c, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeMissingData, resp.ErrorCode)
	assert.Contains(t, resp.ErrorMessage, "title")
}

func TestDashboardHandler_Create_EmptyTitle(t *testing.T) {
	handler, db := setupDashboardTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionDashboardCreate,
		Data: map[string]interface{}{
			"title": "",
		},
	}

	handler.handleDashboardCreate(c, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeMissingData, resp.ErrorCode)
}

func TestDashboardHandler_Create_Unauthorized(t *testing.T) {
	handler, db := setupDashboardTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	// No username set

	req := &models.Request{
		Action: models.ActionDashboardCreate,
		Data:   map[string]interface{}{},
	}

	handler.handleDashboardCreate(c, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestDashboardHandler_Create_PermissionDenied(t *testing.T) {
	db, err := database.NewDatabase(config.DatabaseConfig{
		Type:       "sqlite",
		SQLitePath: ":memory:",
	})
	require.NoError(t, err)
	defer db.Close()

	err = InitializeProjectTables(db)
	require.NoError(t, err)

	mockAuth := &services.MockAuthService{
		IsEnabledFunc: func() bool { return true },
	}

	mockPerm := &services.MockPermissionService{
		IsEnabledFunc: func() bool { return true },
		CheckPermissionFunc: func(ctx context.Context, username, permissionContext string, requiredLevel models.PermissionLevel) (bool, error) {
			return false, nil // Deny permission
		},
	}

	handler := NewHandler(db, mockAuth, mockPerm, "1.0.0-test")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionDashboardCreate,
		Data: map[string]interface{}{
			"title": "Test",
		},
	}

	handler.handleDashboardCreate(c, req)

	assert.Equal(t, http.StatusForbidden, w.Code)

	var resp models.Response
	err = json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeForbidden, resp.ErrorCode)
}

// ============================================================================
// ActionDashboardRead Tests
// ============================================================================

func TestDashboardHandler_Read_Success_Owner(t *testing.T) {
	handler, db := setupDashboardTestHandler(t)
	defer db.Close()

	dashboardID := uuid.New().String()
	createTestDashboard(t, db, dashboardID, "Test Dashboard", "testuser", false, false)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionDashboardRead,
		Data: map[string]interface{}{
			"dashboardId": dashboardID,
		},
	}

	handler.handleDashboardRead(c, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.Equal(t, dashboardID, resp.Data["id"])
	assert.Equal(t, "Test Dashboard", resp.Data["title"])
}

func TestDashboardHandler_Read_Success_Public(t *testing.T) {
	handler, db := setupDashboardTestHandler(t)
	defer db.Close()

	dashboardID := uuid.New().String()
	createTestDashboard(t, db, dashboardID, "Public Dashboard", "otheruser", true, false)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionDashboardRead,
		Data: map[string]interface{}{
			"dashboardId": dashboardID,
		},
	}

	handler.handleDashboardRead(c, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
}

func TestDashboardHandler_Read_Success_Shared(t *testing.T) {
	handler, db := setupDashboardTestHandler(t)
	defer db.Close()

	dashboardID := uuid.New().String()
	createTestDashboard(t, db, dashboardID, "Shared Dashboard", "otheruser", false, false)

	shareID := uuid.New().String()
	createTestShare(t, db, shareID, dashboardID, "testuser")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionDashboardRead,
		Data: map[string]interface{}{
			"dashboardId": dashboardID,
		},
	}

	handler.handleDashboardRead(c, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
}

func TestDashboardHandler_Read_MissingDashboardID(t *testing.T) {
	handler, db := setupDashboardTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionDashboardRead,
		Data:   map[string]interface{}{},
	}

	handler.handleDashboardRead(c, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeMissingData, resp.ErrorCode)
}

func TestDashboardHandler_Read_NotFound(t *testing.T) {
	handler, db := setupDashboardTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionDashboardRead,
		Data: map[string]interface{}{
			"dashboardId": "nonexistent-id",
		},
	}

	handler.handleDashboardRead(c, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeEntityNotFound, resp.ErrorCode)
}

func TestDashboardHandler_Read_AccessDenied(t *testing.T) {
	handler, db := setupDashboardTestHandler(t)
	defer db.Close()

	dashboardID := uuid.New().String()
	createTestDashboard(t, db, dashboardID, "Private Dashboard", "otheruser", false, false)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionDashboardRead,
		Data: map[string]interface{}{
			"dashboardId": dashboardID,
		},
	}

	handler.handleDashboardRead(c, req)

	assert.Equal(t, http.StatusForbidden, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeForbidden, resp.ErrorCode)
}

func TestDashboardHandler_Read_Unauthorized(t *testing.T) {
	handler, db := setupDashboardTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	// No username set

	req := &models.Request{
		Action: models.ActionDashboardRead,
		Data:   map[string]interface{}{},
	}

	handler.handleDashboardRead(c, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// ============================================================================
// ActionDashboardList Tests
// ============================================================================

func TestDashboardHandler_List_Success_Owned(t *testing.T) {
	handler, db := setupDashboardTestHandler(t)
	defer db.Close()

	// Create owned dashboards
	for i := 0; i < 3; i++ {
		dashboardID := uuid.New().String()
		createTestDashboard(t, db, dashboardID, "Dashboard "+string(rune(i+1)), "testuser", false, false)
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionDashboardList,
		Data:   map[string]interface{}{},
	}

	handler.handleDashboardList(c, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.Equal(t, float64(3), resp.Data["count"])
}

func TestDashboardHandler_List_Success_Mixed(t *testing.T) {
	handler, db := setupDashboardTestHandler(t)
	defer db.Close()

	// Create owned dashboard
	dashboard1ID := uuid.New().String()
	createTestDashboard(t, db, dashboard1ID, "My Dashboard", "testuser", false, false)

	// Create public dashboard
	dashboard2ID := uuid.New().String()
	createTestDashboard(t, db, dashboard2ID, "Public Dashboard", "otheruser", true, false)

	// Create shared dashboard
	dashboard3ID := uuid.New().String()
	createTestDashboard(t, db, dashboard3ID, "Shared Dashboard", "otheruser", false, false)
	shareID := uuid.New().String()
	createTestShare(t, db, shareID, dashboard3ID, "testuser")

	// Create private dashboard from another user (should not appear)
	dashboard4ID := uuid.New().String()
	createTestDashboard(t, db, dashboard4ID, "Private Dashboard", "otheruser", false, false)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionDashboardList,
		Data:   map[string]interface{}{},
	}

	handler.handleDashboardList(c, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.Equal(t, float64(3), resp.Data["count"]) // Own + Public + Shared
}

func TestDashboardHandler_List_Empty(t *testing.T) {
	handler, db := setupDashboardTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionDashboardList,
		Data:   map[string]interface{}{},
	}

	handler.handleDashboardList(c, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.Equal(t, float64(0), resp.Data["count"])
}

func TestDashboardHandler_List_Unauthorized(t *testing.T) {
	handler, db := setupDashboardTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	// No username set

	req := &models.Request{
		Action: models.ActionDashboardList,
		Data:   map[string]interface{}{},
	}

	handler.handleDashboardList(c, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// ============================================================================
// ActionDashboardModify Tests
// ============================================================================

func TestDashboardHandler_Modify_Success(t *testing.T) {
	handler, db := setupDashboardTestHandler(t)
	defer db.Close()

	dashboardID := uuid.New().String()
	createTestDashboard(t, db, dashboardID, "Original Title", "testuser", false, false)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionDashboardModify,
		Data: map[string]interface{}{
			"dashboardId": dashboardID,
			"title":       "Updated Title",
			"description": "Updated description",
			"isPublic":    true,
		},
	}

	handler.handleDashboardModify(c, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.True(t, resp.Data["updated"].(bool))
	assert.Equal(t, dashboardID, resp.Data["dashboardId"])
}

func TestDashboardHandler_Modify_Success_OnlyTitle(t *testing.T) {
	handler, db := setupDashboardTestHandler(t)
	defer db.Close()

	dashboardID := uuid.New().String()
	createTestDashboard(t, db, dashboardID, "Original Title", "testuser", false, false)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionDashboardModify,
		Data: map[string]interface{}{
			"dashboardId": dashboardID,
			"title":       "New Title",
		},
	}

	handler.handleDashboardModify(c, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
}

func TestDashboardHandler_Modify_MissingDashboardID(t *testing.T) {
	handler, db := setupDashboardTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionDashboardModify,
		Data: map[string]interface{}{
			"title": "New Title",
		},
	}

	handler.handleDashboardModify(c, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeMissingData, resp.ErrorCode)
}

func TestDashboardHandler_Modify_NotFound(t *testing.T) {
	handler, db := setupDashboardTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionDashboardModify,
		Data: map[string]interface{}{
			"dashboardId": "nonexistent-id",
			"title":       "New Title",
		},
	}

	handler.handleDashboardModify(c, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeEntityNotFound, resp.ErrorCode)
}

func TestDashboardHandler_Modify_NotOwner(t *testing.T) {
	handler, db := setupDashboardTestHandler(t)
	defer db.Close()

	dashboardID := uuid.New().String()
	createTestDashboard(t, db, dashboardID, "Other's Dashboard", "otheruser", false, false)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionDashboardModify,
		Data: map[string]interface{}{
			"dashboardId": dashboardID,
			"title":       "New Title",
		},
	}

	handler.handleDashboardModify(c, req)

	assert.Equal(t, http.StatusForbidden, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeForbidden, resp.ErrorCode)
}

func TestDashboardHandler_Modify_NoFieldsToUpdate(t *testing.T) {
	handler, db := setupDashboardTestHandler(t)
	defer db.Close()

	dashboardID := uuid.New().String()
	createTestDashboard(t, db, dashboardID, "Original Title", "testuser", false, false)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionDashboardModify,
		Data: map[string]interface{}{
			"dashboardId": dashboardID,
			// No fields to update
		},
	}

	handler.handleDashboardModify(c, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeMissingData, resp.ErrorCode)
}

func TestDashboardHandler_Modify_Unauthorized(t *testing.T) {
	handler, db := setupDashboardTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	// No username set

	req := &models.Request{
		Action: models.ActionDashboardModify,
		Data:   map[string]interface{}{},
	}

	handler.handleDashboardModify(c, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// ============================================================================
// ActionDashboardRemove Tests
// ============================================================================

func TestDashboardHandler_Remove_Success(t *testing.T) {
	handler, db := setupDashboardTestHandler(t)
	defer db.Close()

	dashboardID := uuid.New().String()
	createTestDashboard(t, db, dashboardID, "Test Dashboard", "testuser", false, false)

	// Add a widget
	widgetID := uuid.New().String()
	createTestWidget(t, db, widgetID, dashboardID, models.WidgetTypeFilterResults)

	// Add a share
	shareID := uuid.New().String()
	createTestShare(t, db, shareID, dashboardID, "otheruser")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionDashboardRemove,
		Data: map[string]interface{}{
			"dashboardId": dashboardID,
		},
	}

	handler.handleDashboardRemove(c, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.True(t, resp.Data["removed"].(bool))
	assert.Equal(t, dashboardID, resp.Data["dashboardId"])

	// Verify dashboard is soft-deleted
	var deleted bool
	err = db.QueryRow(context.Background(), "SELECT deleted FROM dashboard WHERE id = ?", dashboardID).Scan(&deleted)
	require.NoError(t, err)
	assert.True(t, deleted)
}

func TestDashboardHandler_Remove_MissingDashboardID(t *testing.T) {
	handler, db := setupDashboardTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionDashboardRemove,
		Data:   map[string]interface{}{},
	}

	handler.handleDashboardRemove(c, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeMissingData, resp.ErrorCode)
}

func TestDashboardHandler_Remove_NotFound(t *testing.T) {
	handler, db := setupDashboardTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionDashboardRemove,
		Data: map[string]interface{}{
			"dashboardId": "nonexistent-id",
		},
	}

	handler.handleDashboardRemove(c, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeEntityNotFound, resp.ErrorCode)
}

func TestDashboardHandler_Remove_NotOwner(t *testing.T) {
	handler, db := setupDashboardTestHandler(t)
	defer db.Close()

	dashboardID := uuid.New().String()
	createTestDashboard(t, db, dashboardID, "Other's Dashboard", "otheruser", false, false)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionDashboardRemove,
		Data: map[string]interface{}{
			"dashboardId": dashboardID,
		},
	}

	handler.handleDashboardRemove(c, req)

	assert.Equal(t, http.StatusForbidden, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeForbidden, resp.ErrorCode)
}

func TestDashboardHandler_Remove_Unauthorized(t *testing.T) {
	handler, db := setupDashboardTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	// No username set

	req := &models.Request{
		Action: models.ActionDashboardRemove,
		Data:   map[string]interface{}{},
	}

	handler.handleDashboardRemove(c, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// ============================================================================
// ActionDashboardShare Tests
// ============================================================================

func TestDashboardHandler_Share_Success_WithUser(t *testing.T) {
	handler, db := setupDashboardTestHandler(t)
	defer db.Close()

	dashboardID := uuid.New().String()
	createTestDashboard(t, db, dashboardID, "Test Dashboard", "testuser", false, false)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionDashboardShare,
		Data: map[string]interface{}{
			"dashboardId": dashboardID,
			"userId":      "otheruser",
		},
	}

	handler.handleDashboardShare(c, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.True(t, resp.Data["shared"].(bool))
	assert.Equal(t, "otheruser", resp.Data["userId"])
}

func TestDashboardHandler_Share_Success_WithTeam(t *testing.T) {
	handler, db := setupDashboardTestHandler(t)
	defer db.Close()

	dashboardID := uuid.New().String()
	createTestDashboard(t, db, dashboardID, "Test Dashboard", "testuser", false, false)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionDashboardShare,
		Data: map[string]interface{}{
			"dashboardId": dashboardID,
			"teamId":      "team123",
		},
	}

	handler.handleDashboardShare(c, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.Equal(t, "team123", resp.Data["teamId"])
}

func TestDashboardHandler_Share_Success_WithProject(t *testing.T) {
	handler, db := setupDashboardTestHandler(t)
	defer db.Close()

	dashboardID := uuid.New().String()
	createTestDashboard(t, db, dashboardID, "Test Dashboard", "testuser", false, false)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionDashboardShare,
		Data: map[string]interface{}{
			"dashboardId": dashboardID,
			"projectId":   "project123",
		},
	}

	handler.handleDashboardShare(c, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.Equal(t, "project123", resp.Data["projectId"])
}

func TestDashboardHandler_Share_MissingDashboardID(t *testing.T) {
	handler, db := setupDashboardTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionDashboardShare,
		Data: map[string]interface{}{
			"userId": "otheruser",
		},
	}

	handler.handleDashboardShare(c, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeMissingData, resp.ErrorCode)
}

func TestDashboardHandler_Share_MissingRecipient(t *testing.T) {
	handler, db := setupDashboardTestHandler(t)
	defer db.Close()

	dashboardID := uuid.New().String()
	createTestDashboard(t, db, dashboardID, "Test Dashboard", "testuser", false, false)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionDashboardShare,
		Data: map[string]interface{}{
			"dashboardId": dashboardID,
			// No userId, teamId, or projectId
		},
	}

	handler.handleDashboardShare(c, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeMissingData, resp.ErrorCode)
}

func TestDashboardHandler_Share_MultipleRecipients(t *testing.T) {
	handler, db := setupDashboardTestHandler(t)
	defer db.Close()

	dashboardID := uuid.New().String()
	createTestDashboard(t, db, dashboardID, "Test Dashboard", "testuser", false, false)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionDashboardShare,
		Data: map[string]interface{}{
			"dashboardId": dashboardID,
			"userId":      "otheruser",
			"teamId":      "team123",
		},
	}

	handler.handleDashboardShare(c, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeInvalidData, resp.ErrorCode)
}

func TestDashboardHandler_Share_NotOwner(t *testing.T) {
	handler, db := setupDashboardTestHandler(t)
	defer db.Close()

	dashboardID := uuid.New().String()
	createTestDashboard(t, db, dashboardID, "Other's Dashboard", "otheruser", false, false)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionDashboardShare,
		Data: map[string]interface{}{
			"dashboardId": dashboardID,
			"userId":      "anotheruser",
		},
	}

	handler.handleDashboardShare(c, req)

	assert.Equal(t, http.StatusForbidden, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeForbidden, resp.ErrorCode)
}

func TestDashboardHandler_Share_DashboardNotFound(t *testing.T) {
	handler, db := setupDashboardTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionDashboardShare,
		Data: map[string]interface{}{
			"dashboardId": "nonexistent-id",
			"userId":      "otheruser",
		},
	}

	handler.handleDashboardShare(c, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeEntityNotFound, resp.ErrorCode)
}

func TestDashboardHandler_Share_Unauthorized(t *testing.T) {
	handler, db := setupDashboardTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	// No username set

	req := &models.Request{
		Action: models.ActionDashboardShare,
		Data:   map[string]interface{}{},
	}

	handler.handleDashboardShare(c, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// ============================================================================
// ActionDashboardUnshare Tests
// ============================================================================

func TestDashboardHandler_Unshare_Success_WithUser(t *testing.T) {
	handler, db := setupDashboardTestHandler(t)
	defer db.Close()

	dashboardID := uuid.New().String()
	createTestDashboard(t, db, dashboardID, "Test Dashboard", "testuser", false, false)

	shareID := uuid.New().String()
	createTestShare(t, db, shareID, dashboardID, "otheruser")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionDashboardUnshare,
		Data: map[string]interface{}{
			"dashboardId": dashboardID,
			"userId":      "otheruser",
		},
	}

	handler.handleDashboardUnshare(c, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.True(t, resp.Data["unshared"].(bool))
}

func TestDashboardHandler_Unshare_Success_WithTeam(t *testing.T) {
	handler, db := setupDashboardTestHandler(t)
	defer db.Close()

	dashboardID := uuid.New().String()
	createTestDashboard(t, db, dashboardID, "Test Dashboard", "testuser", false, false)

	// Create share with team
	shareID := uuid.New().String()
	_, err := db.Exec(context.Background(),
		"INSERT INTO dashboard_share_mapping (id, dashboard_id, user_id, team_id, project_id, created, deleted) VALUES (?, ?, NULL, ?, NULL, ?, ?)",
		shareID, dashboardID, "team123", time.Now().Unix(), false)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionDashboardUnshare,
		Data: map[string]interface{}{
			"dashboardId": dashboardID,
			"teamId":      "team123",
		},
	}

	handler.handleDashboardUnshare(c, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err = json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
}

func TestDashboardHandler_Unshare_MissingDashboardID(t *testing.T) {
	handler, db := setupDashboardTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionDashboardUnshare,
		Data: map[string]interface{}{
			"userId": "otheruser",
		},
	}

	handler.handleDashboardUnshare(c, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeMissingData, resp.ErrorCode)
}

func TestDashboardHandler_Unshare_MissingRecipient(t *testing.T) {
	handler, db := setupDashboardTestHandler(t)
	defer db.Close()

	dashboardID := uuid.New().String()
	createTestDashboard(t, db, dashboardID, "Test Dashboard", "testuser", false, false)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionDashboardUnshare,
		Data: map[string]interface{}{
			"dashboardId": dashboardID,
		},
	}

	handler.handleDashboardUnshare(c, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeMissingData, resp.ErrorCode)
}

func TestDashboardHandler_Unshare_ShareNotFound(t *testing.T) {
	handler, db := setupDashboardTestHandler(t)
	defer db.Close()

	dashboardID := uuid.New().String()
	createTestDashboard(t, db, dashboardID, "Test Dashboard", "testuser", false, false)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionDashboardUnshare,
		Data: map[string]interface{}{
			"dashboardId": dashboardID,
			"userId":      "nonexistentuser",
		},
	}

	handler.handleDashboardUnshare(c, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeEntityNotFound, resp.ErrorCode)
}

func TestDashboardHandler_Unshare_NotOwner(t *testing.T) {
	handler, db := setupDashboardTestHandler(t)
	defer db.Close()

	dashboardID := uuid.New().String()
	createTestDashboard(t, db, dashboardID, "Other's Dashboard", "otheruser", false, false)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionDashboardUnshare,
		Data: map[string]interface{}{
			"dashboardId": dashboardID,
			"userId":      "anotheruser",
		},
	}

	handler.handleDashboardUnshare(c, req)

	assert.Equal(t, http.StatusForbidden, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeForbidden, resp.ErrorCode)
}

func TestDashboardHandler_Unshare_Unauthorized(t *testing.T) {
	handler, db := setupDashboardTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	// No username set

	req := &models.Request{
		Action: models.ActionDashboardUnshare,
		Data:   map[string]interface{}{},
	}

	handler.handleDashboardUnshare(c, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// Continue in next part due to length...

// ============================================================================
// Widget Actions Tests  
// ============================================================================

func TestDashboardHandler_AddWidget_Success(t *testing.T) {
	handler, db := setupDashboardTestHandler(t)
	defer db.Close()

	dashboardID := uuid.New().String()
	createTestDashboard(t, db, dashboardID, "Test Dashboard", "testuser", false, false)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionDashboardAddWidget,
		Data: map[string]interface{}{
			"dashboardId": dashboardID,
			"widgetType":  models.WidgetTypeFilterResults,
			"title":       "My Widget",
			"positionX":   float64(0),
			"positionY":   float64(0),
			"width":       float64(4),
			"height":      float64(4),
		},
	}

	handler.handleDashboardAddWidget(c, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.True(t, resp.Data["added"].(bool))
	assert.NotEmpty(t, resp.Data["widgetId"])
}

func TestDashboardHandler_AddWidget_MissingDashboardID(t *testing.T) {
	handler, db := setupDashboardTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionDashboardAddWidget,
		Data: map[string]interface{}{
			"widgetType": models.WidgetTypeFilterResults,
		},
	}

	handler.handleDashboardAddWidget(c, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDashboardHandler_AddWidget_MissingWidgetType(t *testing.T) {
	handler, db := setupDashboardTestHandler(t)
	defer db.Close()

	dashboardID := uuid.New().String()
	createTestDashboard(t, db, dashboardID, "Test Dashboard", "testuser", false, false)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionDashboardAddWidget,
		Data: map[string]interface{}{
			"dashboardId": dashboardID,
		},
	}

	handler.handleDashboardAddWidget(c, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDashboardHandler_AddWidget_NotOwner(t *testing.T) {
	handler, db := setupDashboardTestHandler(t)
	defer db.Close()

	dashboardID := uuid.New().String()
	createTestDashboard(t, db, dashboardID, "Other's Dashboard", "otheruser", false, false)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionDashboardAddWidget,
		Data: map[string]interface{}{
			"dashboardId": dashboardID,
			"widgetType":  models.WidgetTypeFilterResults,
		},
	}

	handler.handleDashboardAddWidget(c, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestDashboardHandler_RemoveWidget_Success(t *testing.T) {
	handler, db := setupDashboardTestHandler(t)
	defer db.Close()

	dashboardID := uuid.New().String()
	createTestDashboard(t, db, dashboardID, "Test Dashboard", "testuser", false, false)

	widgetID := uuid.New().String()
	createTestWidget(t, db, widgetID, dashboardID, models.WidgetTypeFilterResults)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionDashboardRemoveWidget,
		Data: map[string]interface{}{
			"widgetId": widgetID,
		},
	}

	handler.handleDashboardRemoveWidget(c, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.True(t, resp.Data["removed"].(bool))
}

func TestDashboardHandler_RemoveWidget_NotFound(t *testing.T) {
	handler, db := setupDashboardTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionDashboardRemoveWidget,
		Data: map[string]interface{}{
			"widgetId": "nonexistent-id",
		},
	}

	handler.handleDashboardRemoveWidget(c, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestDashboardHandler_ModifyWidget_Success(t *testing.T) {
	handler, db := setupDashboardTestHandler(t)
	defer db.Close()

	dashboardID := uuid.New().String()
	createTestDashboard(t, db, dashboardID, "Test Dashboard", "testuser", false, false)

	widgetID := uuid.New().String()
	createTestWidget(t, db, widgetID, dashboardID, models.WidgetTypeFilterResults)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionDashboardModifyWidget,
		Data: map[string]interface{}{
			"widgetId": widgetID,
			"title":    "Updated Widget Title",
			"width":    float64(6),
		},
	}

	handler.handleDashboardModifyWidget(c, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.True(t, resp.Data["updated"].(bool))
}

func TestDashboardHandler_ListWidgets_Success(t *testing.T) {
	handler, db := setupDashboardTestHandler(t)
	defer db.Close()

	dashboardID := uuid.New().String()
	createTestDashboard(t, db, dashboardID, "Test Dashboard", "testuser", false, false)

	// Create multiple widgets
	for i := 0; i < 3; i++ {
		widgetID := uuid.New().String()
		createTestWidget(t, db, widgetID, dashboardID, models.WidgetTypeFilterResults)
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionDashboardListWidgets,
		Data: map[string]interface{}{
			"dashboardId": dashboardID,
		},
	}

	handler.handleDashboardListWidgets(c, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.Equal(t, float64(3), resp.Data["count"])
}

func TestDashboardHandler_SetLayout_Success(t *testing.T) {
	handler, db := setupDashboardTestHandler(t)
	defer db.Close()

	dashboardID := uuid.New().String()
	createTestDashboard(t, db, dashboardID, "Test Dashboard", "testuser", false, false)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionDashboardSetLayout,
		Data: map[string]interface{}{
			"dashboardId": dashboardID,
			"layout":      `{"columns": 12, "rows": 10}`,
		},
	}

	handler.handleDashboardSetLayout(c, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.True(t, resp.Data["updated"].(bool))
}

func TestDashboardHandler_SetLayout_MissingLayout(t *testing.T) {
	handler, db := setupDashboardTestHandler(t)
	defer db.Close()

	dashboardID := uuid.New().String()
	createTestDashboard(t, db, dashboardID, "Test Dashboard", "testuser", false, false)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionDashboardSetLayout,
		Data: map[string]interface{}{
			"dashboardId": dashboardID,
		},
	}

	handler.handleDashboardSetLayout(c, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDashboardHandler_SetLayout_NotOwner(t *testing.T) {
	handler, db := setupDashboardTestHandler(t)
	defer db.Close()

	dashboardID := uuid.New().String()
	createTestDashboard(t, db, dashboardID, "Other's Dashboard", "otheruser", false, false)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionDashboardSetLayout,
		Data: map[string]interface{}{
			"dashboardId": dashboardID,
			"layout":      `{"columns": 12}`,
		},
	}

	handler.handleDashboardSetLayout(c, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}
