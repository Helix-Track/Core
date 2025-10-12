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

func setupProjectRoleTestHandler(t *testing.T) (*Handler, database.Database) {
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

// Helper to create a test project for project role tests
func createTestProjectForRole(t *testing.T, db database.Database, projectID, title string) {
	_, err := db.Exec(context.Background(),
		"INSERT INTO project (id, title, created, modified, deleted) VALUES (?, ?, ?, ?, ?)",
		projectID, title, time.Now().Unix(), time.Now().Unix(), false)
	require.NoError(t, err)
}

// Helper to create a test project role with full parameters
func createTestProjectRoleWithDetails(t *testing.T, db database.Database, roleID, title, description string, projectID *string) {
	var query string
	var args []interface{}

	if projectID != nil && *projectID != "" {
		query = `INSERT INTO project_role (id, title, description, project_id, created, modified, deleted)
		         VALUES (?, ?, ?, ?, ?, ?, ?)`
		args = []interface{}{roleID, title, description, *projectID, time.Now().Unix(), time.Now().Unix(), false}
	} else {
		query = `INSERT INTO project_role (id, title, description, project_id, created, modified, deleted)
		         VALUES (?, ?, ?, NULL, ?, ?, ?)`
		args = []interface{}{roleID, title, description, time.Now().Unix(), time.Now().Unix(), false}
	}

	_, err := db.Exec(context.Background(), query, args...)
	require.NoError(t, err)
}

// Helper to create a project role user mapping with project_id
func createTestProjectRoleUserMappingWithProject(t *testing.T, db database.Database, roleID, projectID, userID string) string {
	mappingID := uuid.New().String()
	_, err := db.Exec(context.Background(),
		"INSERT INTO project_role_user_mapping (id, project_role_id, project_id, user_id, created, deleted) VALUES (?, ?, ?, ?, ?, ?)",
		mappingID, roleID, projectID, userID, time.Now().Unix(), false)
	require.NoError(t, err)
	return mappingID
}

func TestProjectRoleHandler_Create_Success_GlobalRole(t *testing.T) {
	handler, db := setupProjectRoleTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionProjectRoleCreate,
		Data: map[string]interface{}{
			"title":       "Global Admin",
			"description": "Global administrator role",
		},
	}

	handler.handleProjectRoleCreate(c, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.Equal(t, "Global Admin", resp.Data["title"])
	assert.Equal(t, "Global administrator role", resp.Data["description"])
	assert.True(t, resp.Data["isGlobal"].(bool))
	assert.NotNil(t, resp.Data["id"])
}

func TestProjectRoleHandler_Create_Success_ProjectSpecificRole(t *testing.T) {
	handler, db := setupProjectRoleTestHandler(t)
	defer db.Close()

	// Create a test project
	projectID := uuid.New().String()
	createTestProjectForRole(t, db, projectID, "Test Project")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionProjectRoleCreate,
		Data: map[string]interface{}{
			"title":       "Project Manager",
			"description": "Manages the project",
			"projectId":   projectID,
		},
	}

	handler.handleProjectRoleCreate(c, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.Equal(t, "Project Manager", resp.Data["title"])
	assert.Equal(t, projectID, resp.Data["projectId"])
	assert.False(t, resp.Data["isGlobal"].(bool))
}

func TestProjectRoleHandler_Create_MissingTitle(t *testing.T) {
	handler, db := setupProjectRoleTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionProjectRoleCreate,
		Data: map[string]interface{}{
			"description": "No title provided",
		},
	}

	handler.handleProjectRoleCreate(c, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeMissingData, resp.ErrorCode)
}

func TestProjectRoleHandler_Create_ProjectNotFound(t *testing.T) {
	handler, db := setupProjectRoleTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionProjectRoleCreate,
		Data: map[string]interface{}{
			"title":     "Test Role",
			"projectId": "nonexistent-project-id",
		},
	}

	handler.handleProjectRoleCreate(c, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeEntityNotFound, resp.ErrorCode)
}

func TestProjectRoleHandler_Create_Unauthorized(t *testing.T) {
	handler, db := setupProjectRoleTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	// No username set

	req := &models.Request{
		Action: models.ActionProjectRoleCreate,
		Data: map[string]interface{}{
			"title": "Test Role",
		},
	}

	handler.handleProjectRoleCreate(c, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestProjectRoleHandler_Create_Forbidden(t *testing.T) {
	handler, db := setupProjectRoleTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "unauthorizeduser")

	req := &models.Request{
		Action: models.ActionProjectRoleCreate,
		Data: map[string]interface{}{
			"title": "Test Role",
		},
	}

	handler.handleProjectRoleCreate(c, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestProjectRoleHandler_Read_Success(t *testing.T) {
	handler, db := setupProjectRoleTestHandler(t)
	defer db.Close()

	// Create a test role
	roleID := uuid.New().String()
	createTestProjectRoleWithDetails(t, db, roleID, "Test Role", "Test Description", nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionProjectRoleRead,
		Data: map[string]interface{}{
			"roleId": roleID,
		},
	}

	handler.handleProjectRoleRead(c, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.Equal(t, roleID, resp.Data["id"])
	assert.Equal(t, "Test Role", resp.Data["title"])
	assert.Equal(t, "Test Description", resp.Data["description"])
	assert.True(t, resp.Data["isGlobal"].(bool))
}

func TestProjectRoleHandler_Read_MissingRoleID(t *testing.T) {
	handler, db := setupProjectRoleTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionProjectRoleRead,
		Data:   map[string]interface{}{},
	}

	handler.handleProjectRoleRead(c, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeMissingData, resp.ErrorCode)
}

func TestProjectRoleHandler_Read_NotFound(t *testing.T) {
	handler, db := setupProjectRoleTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionProjectRoleRead,
		Data: map[string]interface{}{
			"roleId": "nonexistent-role-id",
		},
	}

	handler.handleProjectRoleRead(c, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeEntityNotFound, resp.ErrorCode)
}

func TestProjectRoleHandler_List_Success_AllRoles(t *testing.T) {
	handler, db := setupProjectRoleTestHandler(t)
	defer db.Close()

	// Create global roles
	for i := 0; i < 2; i++ {
		roleID := uuid.New().String()
		createTestProjectRoleWithDetails(t, db, roleID, "Global Role "+string(rune('A'+i)), "Description", nil)
	}

	// Create project-specific roles
	projectID := uuid.New().String()
	createTestProjectForRole(t, db, projectID, "Test Project")
	projIDPtr := &projectID
	for i := 0; i < 2; i++ {
		roleID := uuid.New().String()
		createTestProjectRoleWithDetails(t, db, roleID, "Project Role "+string(rune('A'+i)), "Description", projIDPtr)
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionProjectRoleList,
		Data:   map[string]interface{}{},
	}

	handler.handleProjectRoleList(c, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.Equal(t, float64(4), resp.Data["count"])
}

func TestProjectRoleHandler_List_Success_FilterByProject(t *testing.T) {
	handler, db := setupProjectRoleTestHandler(t)
	defer db.Close()

	// Create a project
	projectID := uuid.New().String()
	createTestProjectForRole(t, db, projectID, "Test Project")

	// Create global roles (should be included in project filter)
	for i := 0; i < 2; i++ {
		roleID := uuid.New().String()
		createTestProjectRoleWithDetails(t, db, roleID, "Global Role "+string(rune('A'+i)), "Description", nil)
	}

	// Create project-specific roles
	projIDPtr := &projectID
	for i := 0; i < 2; i++ {
		roleID := uuid.New().String()
		createTestProjectRoleWithDetails(t, db, roleID, "Project Role "+string(rune('A'+i)), "Description", projIDPtr)
	}

	// Create roles for another project (should not be included)
	otherProjectID := uuid.New().String()
	createTestProjectForRole(t, db, otherProjectID, "Other Project")
	otherProjIDPtr := &otherProjectID
	roleID := uuid.New().String()
	createTestProjectRoleWithDetails(t, db, roleID, "Other Project Role", "Description", otherProjIDPtr)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionProjectRoleList,
		Data: map[string]interface{}{
			"projectId": projectID,
		},
	}

	handler.handleProjectRoleList(c, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	// Should include 2 global roles + 2 project-specific roles = 4 total
	assert.Equal(t, float64(4), resp.Data["count"])
}

func TestProjectRoleHandler_Modify_Success(t *testing.T) {
	handler, db := setupProjectRoleTestHandler(t)
	defer db.Close()

	// Create a test role
	roleID := uuid.New().String()
	createTestProjectRoleWithDetails(t, db, roleID, "Old Title", "Old Description", nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionProjectRoleModify,
		Data: map[string]interface{}{
			"roleId":      roleID,
			"title":       "New Title",
			"description": "New Description",
		},
	}

	handler.handleProjectRoleModify(c, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.True(t, resp.Data["updated"].(bool))
	assert.Equal(t, roleID, resp.Data["roleId"])
}

func TestProjectRoleHandler_Modify_MissingRoleID(t *testing.T) {
	handler, db := setupProjectRoleTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionProjectRoleModify,
		Data: map[string]interface{}{
			"title": "New Title",
		},
	}

	handler.handleProjectRoleModify(c, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeMissingData, resp.ErrorCode)
}

func TestProjectRoleHandler_Modify_NotFound(t *testing.T) {
	handler, db := setupProjectRoleTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionProjectRoleModify,
		Data: map[string]interface{}{
			"roleId": "nonexistent-role-id",
			"title":  "New Title",
		},
	}

	handler.handleProjectRoleModify(c, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeEntityNotFound, resp.ErrorCode)
}

func TestProjectRoleHandler_Modify_NoFieldsToUpdate(t *testing.T) {
	handler, db := setupProjectRoleTestHandler(t)
	defer db.Close()

	// Create a test role
	roleID := uuid.New().String()
	createTestProjectRoleWithDetails(t, db, roleID, "Test Role", "Description", nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionProjectRoleModify,
		Data: map[string]interface{}{
			"roleId": roleID,
			// No fields to update
		},
	}

	handler.handleProjectRoleModify(c, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeMissingData, resp.ErrorCode)
}

func TestProjectRoleHandler_Remove_Success(t *testing.T) {
	handler, db := setupProjectRoleTestHandler(t)
	defer db.Close()

	// Create a test role
	roleID := uuid.New().String()
	createTestProjectRoleWithDetails(t, db, roleID, "Test Role", "Description", nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionProjectRoleRemove,
		Data: map[string]interface{}{
			"roleId": roleID,
		},
	}

	handler.handleProjectRoleRemove(c, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.True(t, resp.Data["removed"].(bool))
	assert.Equal(t, roleID, resp.Data["roleId"])
}

func TestProjectRoleHandler_Remove_NotFound(t *testing.T) {
	handler, db := setupProjectRoleTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionProjectRoleRemove,
		Data: map[string]interface{}{
			"roleId": "nonexistent-role-id",
		},
	}

	handler.handleProjectRoleRemove(c, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeEntityNotFound, resp.ErrorCode)
}

func TestProjectRoleHandler_Remove_AlsoRemovesUserMappings(t *testing.T) {
	handler, db := setupProjectRoleTestHandler(t)
	defer db.Close()

	// Create project and role
	projectID := uuid.New().String()
	createTestProjectForRole(t, db, projectID, "Test Project")
	roleID := uuid.New().String()
	createTestProjectRoleWithDetails(t, db, roleID, "Test Role", "Description", nil)

	// Create user mappings
	createTestProjectRoleUserMappingWithProject(t, db, roleID, projectID, "user1")
	createTestProjectRoleUserMappingWithProject(t, db, roleID, projectID, "user2")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionProjectRoleRemove,
		Data: map[string]interface{}{
			"roleId": roleID,
		},
	}

	handler.handleProjectRoleRemove(c, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify mappings are soft-deleted
	var count int
	err := db.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM project_role_user_mapping WHERE project_role_id = ? AND deleted = 0",
		roleID).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestProjectRoleHandler_AssignUser_Success(t *testing.T) {
	handler, db := setupProjectRoleTestHandler(t)
	defer db.Close()

	// Create project and role
	projectID := uuid.New().String()
	createTestProjectForRole(t, db, projectID, "Test Project")
	roleID := uuid.New().String()
	createTestProjectRoleWithDetails(t, db, roleID, "Test Role", "Description", nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionProjectRoleAssignUser,
		Data: map[string]interface{}{
			"roleId":    roleID,
			"projectId": projectID,
			"userId":    "user123",
		},
	}

	handler.handleProjectRoleAssignUser(c, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.True(t, resp.Data["assigned"].(bool))
	assert.Equal(t, roleID, resp.Data["roleId"])
	assert.Equal(t, projectID, resp.Data["projectId"])
	assert.Equal(t, "user123", resp.Data["userId"])
	assert.NotNil(t, resp.Data["mappingId"])
}

func TestProjectRoleHandler_AssignUser_MissingRoleID(t *testing.T) {
	handler, db := setupProjectRoleTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionProjectRoleAssignUser,
		Data: map[string]interface{}{
			"projectId": "project123",
			"userId":    "user123",
		},
	}

	handler.handleProjectRoleAssignUser(c, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeMissingData, resp.ErrorCode)
}

func TestProjectRoleHandler_AssignUser_MissingUserID(t *testing.T) {
	handler, db := setupProjectRoleTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionProjectRoleAssignUser,
		Data: map[string]interface{}{
			"roleId":    "role123",
			"projectId": "project123",
		},
	}

	handler.handleProjectRoleAssignUser(c, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeMissingData, resp.ErrorCode)
}

func TestProjectRoleHandler_AssignUser_MissingProjectID(t *testing.T) {
	handler, db := setupProjectRoleTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionProjectRoleAssignUser,
		Data: map[string]interface{}{
			"roleId": "role123",
			"userId": "user123",
		},
	}

	handler.handleProjectRoleAssignUser(c, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeMissingData, resp.ErrorCode)
}

func TestProjectRoleHandler_AssignUser_RoleNotFound(t *testing.T) {
	handler, db := setupProjectRoleTestHandler(t)
	defer db.Close()

	// Create project only
	projectID := uuid.New().String()
	createTestProjectForRole(t, db, projectID, "Test Project")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionProjectRoleAssignUser,
		Data: map[string]interface{}{
			"roleId":    "nonexistent-role-id",
			"projectId": projectID,
			"userId":    "user123",
		},
	}

	handler.handleProjectRoleAssignUser(c, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeEntityNotFound, resp.ErrorCode)
}

func TestProjectRoleHandler_AssignUser_ProjectNotFound(t *testing.T) {
	handler, db := setupProjectRoleTestHandler(t)
	defer db.Close()

	// Create role only
	roleID := uuid.New().String()
	createTestProjectRoleWithDetails(t, db, roleID, "Test Role", "Description", nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionProjectRoleAssignUser,
		Data: map[string]interface{}{
			"roleId":    roleID,
			"projectId": "nonexistent-project-id",
			"userId":    "user123",
		},
	}

	handler.handleProjectRoleAssignUser(c, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeEntityNotFound, resp.ErrorCode)
}

func TestProjectRoleHandler_AssignUser_AlreadyAssigned(t *testing.T) {
	handler, db := setupProjectRoleTestHandler(t)
	defer db.Close()

	// Create project and role
	projectID := uuid.New().String()
	createTestProjectForRole(t, db, projectID, "Test Project")
	roleID := uuid.New().String()
	createTestProjectRoleWithDetails(t, db, roleID, "Test Role", "Description", nil)

	// Create existing mapping
	createTestProjectRoleUserMappingWithProject(t, db, roleID, projectID, "user123")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionProjectRoleAssignUser,
		Data: map[string]interface{}{
			"roleId":    roleID,
			"projectId": projectID,
			"userId":    "user123",
		},
	}

	handler.handleProjectRoleAssignUser(c, req)

	assert.Equal(t, http.StatusConflict, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeEntityAlreadyExists, resp.ErrorCode)
}

func TestProjectRoleHandler_UnassignUser_Success(t *testing.T) {
	handler, db := setupProjectRoleTestHandler(t)
	defer db.Close()

	// Create project and role
	projectID := uuid.New().String()
	createTestProjectForRole(t, db, projectID, "Test Project")
	roleID := uuid.New().String()
	createTestProjectRoleWithDetails(t, db, roleID, "Test Role", "Description", nil)

	// Create mapping
	createTestProjectRoleUserMappingWithProject(t, db, roleID, projectID, "user123")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionProjectRoleUnassignUser,
		Data: map[string]interface{}{
			"roleId":    roleID,
			"projectId": projectID,
			"userId":    "user123",
		},
	}

	handler.handleProjectRoleUnassignUser(c, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.True(t, resp.Data["unassigned"].(bool))
	assert.Equal(t, roleID, resp.Data["roleId"])
	assert.Equal(t, projectID, resp.Data["projectId"])
	assert.Equal(t, "user123", resp.Data["userId"])
}

func TestProjectRoleHandler_UnassignUser_NotFound(t *testing.T) {
	handler, db := setupProjectRoleTestHandler(t)
	defer db.Close()

	// Create project and role but no mapping
	projectID := uuid.New().String()
	createTestProjectForRole(t, db, projectID, "Test Project")
	roleID := uuid.New().String()
	createTestProjectRoleWithDetails(t, db, roleID, "Test Role", "Description", nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionProjectRoleUnassignUser,
		Data: map[string]interface{}{
			"roleId":    roleID,
			"projectId": projectID,
			"userId":    "user123",
		},
	}

	handler.handleProjectRoleUnassignUser(c, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeEntityNotFound, resp.ErrorCode)
}

func TestProjectRoleHandler_ListUsers_Success_AllProjects(t *testing.T) {
	handler, db := setupProjectRoleTestHandler(t)
	defer db.Close()

	// Create role
	roleID := uuid.New().String()
	createTestProjectRoleWithDetails(t, db, roleID, "Test Role", "Description", nil)

	// Create multiple projects
	project1ID := uuid.New().String()
	createTestProjectForRole(t, db, project1ID, "Project 1")
	project2ID := uuid.New().String()
	createTestProjectForRole(t, db, project2ID, "Project 2")

	// Create mappings for different users in different projects
	createTestProjectRoleUserMappingWithProject(t, db, roleID, project1ID, "user1")
	createTestProjectRoleUserMappingWithProject(t, db, roleID, project1ID, "user2")
	createTestProjectRoleUserMappingWithProject(t, db, roleID, project2ID, "user3")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionProjectRoleListUsers,
		Data: map[string]interface{}{
			"roleId": roleID,
		},
	}

	handler.handleProjectRoleListUsers(c, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.Equal(t, roleID, resp.Data["roleId"])
	assert.Equal(t, float64(3), resp.Data["count"])
}

func TestProjectRoleHandler_ListUsers_Success_FilterByProject(t *testing.T) {
	handler, db := setupProjectRoleTestHandler(t)
	defer db.Close()

	// Create role
	roleID := uuid.New().String()
	createTestProjectRoleWithDetails(t, db, roleID, "Test Role", "Description", nil)

	// Create multiple projects
	project1ID := uuid.New().String()
	createTestProjectForRole(t, db, project1ID, "Project 1")
	project2ID := uuid.New().String()
	createTestProjectForRole(t, db, project2ID, "Project 2")

	// Create mappings
	createTestProjectRoleUserMappingWithProject(t, db, roleID, project1ID, "user1")
	createTestProjectRoleUserMappingWithProject(t, db, roleID, project1ID, "user2")
	createTestProjectRoleUserMappingWithProject(t, db, roleID, project2ID, "user3")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionProjectRoleListUsers,
		Data: map[string]interface{}{
			"roleId":    roleID,
			"projectId": project1ID,
		},
	}

	handler.handleProjectRoleListUsers(c, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.Equal(t, roleID, resp.Data["roleId"])
	assert.Equal(t, float64(2), resp.Data["count"]) // Only project1 users
}

func TestProjectRoleHandler_ListUsers_MissingRoleID(t *testing.T) {
	handler, db := setupProjectRoleTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionProjectRoleListUsers,
		Data:   map[string]interface{}{},
	}

	handler.handleProjectRoleListUsers(c, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeMissingData, resp.ErrorCode)
}

func TestProjectRoleHandler_ListUsers_EmptyList(t *testing.T) {
	handler, db := setupProjectRoleTestHandler(t)
	defer db.Close()

	// Create role with no user mappings
	roleID := uuid.New().String()
	createTestProjectRoleWithDetails(t, db, roleID, "Test Role", "Description", nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionProjectRoleListUsers,
		Data: map[string]interface{}{
			"roleId": roleID,
		},
	}

	handler.handleProjectRoleListUsers(c, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.Equal(t, float64(0), resp.Data["count"])
}
