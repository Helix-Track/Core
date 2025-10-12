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

func setupSecurityLevelTestHandler(t *testing.T) (*Handler, database.Database) {
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

// Helper to create a test project for security level tests
func createTestProjectSecLevel(t *testing.T, db database.Database, projectID string) {
	_, err := db.Exec(context.Background(),
		"INSERT INTO project (id, title, created, modified, deleted) VALUES (?, ?, ?, ?, ?)",
		projectID, "Test Project", time.Now().Unix(), time.Now().Unix(), false)
	require.NoError(t, err)
}

// Helper to create a test user
func createTestUserSecLevel(t *testing.T, db database.Database, userID, username string) {
	_, err := db.Exec(context.Background(),
		"INSERT INTO user (id, username, email, created, deleted) VALUES (?, ?, ?, ?, ?)",
		userID, username, username+"@test.com", time.Now().Unix(), false)
	require.NoError(t, err)
}

// Helper to create a test team
func createTestTeamSecLevel(t *testing.T, db database.Database, teamID string, projectID string) {
	_, err := db.Exec(context.Background(),
		"INSERT INTO team (id, title, project_id, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)",
		teamID, "Test Team", projectID, time.Now().Unix(), time.Now().Unix(), false)
	require.NoError(t, err)
}

// Helper to create a test team user
func createTestTeamUserSecLevel(t *testing.T, db database.Database, teamID, userID string) {
	_, err := db.Exec(context.Background(),
		"INSERT INTO team_user (id, team_id, user_id, created, deleted) VALUES (?, ?, ?, ?, ?)",
		uuid.New().String(), teamID, userID, time.Now().Unix(), false)
	require.NoError(t, err)
}

// Helper to create a test project role
func createTestProjectRoleSecLevel(t *testing.T, db database.Database, roleID, title, projectID string) {
	_, err := db.Exec(context.Background(),
		"INSERT INTO project_role (id, title, project_id, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)",
		roleID, title, projectID, time.Now().Unix(), time.Now().Unix(), false)
	require.NoError(t, err)
}

// Helper to create a test project role user mapping
func createTestProjectRoleUserMappingSecLevel(t *testing.T, db database.Database, roleID, projectID, userID string) {
	_, err := db.Exec(context.Background(),
		"INSERT INTO project_role_user_mapping (id, project_role_id, project_id, user_id, created, deleted) VALUES (?, ?, ?, ?, ?, ?)",
		uuid.New().String(), roleID, projectID, userID, time.Now().Unix(), false)
	require.NoError(t, err)
}

func TestSecurityLevelHandler_Create_Success(t *testing.T) {
	handler, db := setupSecurityLevelTestHandler(t)
	defer db.Close()

	// Create a test project first
	projectID := uuid.New().String()
	createTestProjectSecLevel(t, db, projectID)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionSecurityLevelCreate,
		Data: map[string]interface{}{
			"title":       "Confidential",
			"description": "Confidential documents",
			"projectId":   projectID,
			"level":       float64(models.SecurityLevelConfidential),
		},
	}

	handler.handleSecurityLevelCreate(c, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.Equal(t, "Confidential", resp.Data["title"])
	assert.Equal(t, projectID, resp.Data["projectId"])
	assert.Equal(t, float64(models.SecurityLevelConfidential), resp.Data["level"])
}

func TestSecurityLevelHandler_Create_MissingTitle(t *testing.T) {
	handler, db := setupSecurityLevelTestHandler(t)
	defer db.Close()

	projectID := uuid.New().String()
	createTestProjectSecLevel(t, db, projectID)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionSecurityLevelCreate,
		Data: map[string]interface{}{
			"projectId": projectID,
			"level":     float64(models.SecurityLevelPublic),
		},
	}

	handler.handleSecurityLevelCreate(c, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeMissingData, resp.ErrorCode)
}

func TestSecurityLevelHandler_Create_MissingProjectID(t *testing.T) {
	handler, db := setupSecurityLevelTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionSecurityLevelCreate,
		Data: map[string]interface{}{
			"title": "Confidential",
			"level": float64(models.SecurityLevelConfidential),
		},
	}

	handler.handleSecurityLevelCreate(c, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeMissingData, resp.ErrorCode)
}

func TestSecurityLevelHandler_Create_MissingLevel(t *testing.T) {
	handler, db := setupSecurityLevelTestHandler(t)
	defer db.Close()

	projectID := uuid.New().String()
	createTestProjectSecLevel(t, db, projectID)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionSecurityLevelCreate,
		Data: map[string]interface{}{
			"title":     "Confidential",
			"projectId": projectID,
		},
	}

	handler.handleSecurityLevelCreate(c, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeMissingData, resp.ErrorCode)
}

func TestSecurityLevelHandler_Create_InvalidLevel(t *testing.T) {
	handler, db := setupSecurityLevelTestHandler(t)
	defer db.Close()

	projectID := uuid.New().String()
	createTestProjectSecLevel(t, db, projectID)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionSecurityLevelCreate,
		Data: map[string]interface{}{
			"title":     "Invalid Level",
			"projectId": projectID,
			"level":     float64(10), // Invalid level
		},
	}

	handler.handleSecurityLevelCreate(c, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeInvalidData, resp.ErrorCode)
}

func TestSecurityLevelHandler_Create_ProjectNotFound(t *testing.T) {
	handler, db := setupSecurityLevelTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionSecurityLevelCreate,
		Data: map[string]interface{}{
			"title":     "Confidential",
			"projectId": "nonexistent-id",
			"level":     float64(models.SecurityLevelConfidential),
		},
	}

	handler.handleSecurityLevelCreate(c, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeEntityNotFound, resp.ErrorCode)
}

func TestSecurityLevelHandler_Read_Success(t *testing.T) {
	handler, db := setupSecurityLevelTestHandler(t)
	defer db.Close()

	// Create a project and security level
	projectID := uuid.New().String()
	createTestProjectSecLevel(t, db, projectID)

	securityLevelID := uuid.New().String()
	_, err := db.Exec(context.Background(),
		"INSERT INTO security_level (id, title, description, project_id, level, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		securityLevelID, "Confidential", "Test security level", projectID, models.SecurityLevelConfidential, time.Now().Unix(), time.Now().Unix(), false)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionSecurityLevelRead,
		Data: map[string]interface{}{
			"securityLevelId": securityLevelID,
		},
	}

	handler.handleSecurityLevelRead(c, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err = json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.Equal(t, "Confidential", resp.Data["title"])
	assert.Equal(t, "Test security level", resp.Data["description"])
	assert.Equal(t, float64(models.SecurityLevelConfidential), resp.Data["level"])
}

func TestSecurityLevelHandler_Read_MissingID(t *testing.T) {
	handler, db := setupSecurityLevelTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionSecurityLevelRead,
		Data:   map[string]interface{}{},
	}

	handler.handleSecurityLevelRead(c, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeMissingData, resp.ErrorCode)
}

func TestSecurityLevelHandler_Read_NotFound(t *testing.T) {
	handler, db := setupSecurityLevelTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionSecurityLevelRead,
		Data: map[string]interface{}{
			"securityLevelId": "nonexistent-id",
		},
	}

	handler.handleSecurityLevelRead(c, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeEntityNotFound, resp.ErrorCode)
}

func TestSecurityLevelHandler_List_Success(t *testing.T) {
	handler, db := setupSecurityLevelTestHandler(t)
	defer db.Close()

	// Create a project
	projectID := uuid.New().String()
	createTestProjectSecLevel(t, db, projectID)

	// Create multiple security levels
	for i := 0; i < 3; i++ {
		securityLevelID := uuid.New().String()
		_, err := db.Exec(context.Background(),
			"INSERT INTO security_level (id, title, description, project_id, level, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
			securityLevelID, "Level "+string(rune(i+1)), "Description", projectID, i+1, time.Now().Unix(), time.Now().Unix(), false)
		require.NoError(t, err)
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionSecurityLevelList,
		Data:   map[string]interface{}{},
	}

	handler.handleSecurityLevelList(c, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.Equal(t, float64(3), resp.Data["count"])
}

func TestSecurityLevelHandler_List_FilterByProject(t *testing.T) {
	handler, db := setupSecurityLevelTestHandler(t)
	defer db.Close()

	// Create two projects
	projectID1 := uuid.New().String()
	projectID2 := uuid.New().String()
	createTestProjectSecLevel(t, db, projectID1)
	createTestProjectSecLevel(t, db, projectID2)

	// Create security levels for project 1
	for i := 0; i < 2; i++ {
		securityLevelID := uuid.New().String()
		_, err := db.Exec(context.Background(),
			"INSERT INTO security_level (id, title, description, project_id, level, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
			securityLevelID, "Level "+string(rune(i+1)), "Description", projectID1, i+1, time.Now().Unix(), time.Now().Unix(), false)
		require.NoError(t, err)
	}

	// Create security level for project 2
	securityLevelID := uuid.New().String()
	_, err := db.Exec(context.Background(),
		"INSERT INTO security_level (id, title, description, project_id, level, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		securityLevelID, "Level 3", "Description", projectID2, 3, time.Now().Unix(), time.Now().Unix(), false)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionSecurityLevelList,
		Data: map[string]interface{}{
			"projectId": projectID1,
		},
	}

	handler.handleSecurityLevelList(c, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err = json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.Equal(t, float64(2), resp.Data["count"])
}

func TestSecurityLevelHandler_Modify_Success(t *testing.T) {
	handler, db := setupSecurityLevelTestHandler(t)
	defer db.Close()

	// Create a project and security level
	projectID := uuid.New().String()
	createTestProjectSecLevel(t, db, projectID)

	securityLevelID := uuid.New().String()
	_, err := db.Exec(context.Background(),
		"INSERT INTO security_level (id, title, description, project_id, level, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		securityLevelID, "Old Title", "Old Description", projectID, models.SecurityLevelPublic, time.Now().Unix(), time.Now().Unix(), false)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionSecurityLevelModify,
		Data: map[string]interface{}{
			"securityLevelId": securityLevelID,
			"title":           "New Title",
			"description":     "New Description",
			"level":           float64(models.SecurityLevelConfidential),
		},
	}

	handler.handleSecurityLevelModify(c, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err = json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.True(t, resp.Data["updated"].(bool))
	assert.Equal(t, securityLevelID, resp.Data["securityLevelId"])
}

func TestSecurityLevelHandler_Modify_MissingID(t *testing.T) {
	handler, db := setupSecurityLevelTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionSecurityLevelModify,
		Data: map[string]interface{}{
			"title": "New Title",
		},
	}

	handler.handleSecurityLevelModify(c, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeMissingData, resp.ErrorCode)
}

func TestSecurityLevelHandler_Modify_NotFound(t *testing.T) {
	handler, db := setupSecurityLevelTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionSecurityLevelModify,
		Data: map[string]interface{}{
			"securityLevelId": "nonexistent-id",
			"title":           "New Title",
		},
	}

	handler.handleSecurityLevelModify(c, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeEntityNotFound, resp.ErrorCode)
}

func TestSecurityLevelHandler_Modify_NoFieldsToUpdate(t *testing.T) {
	handler, db := setupSecurityLevelTestHandler(t)
	defer db.Close()

	// Create a project and security level
	projectID := uuid.New().String()
	createTestProjectSecLevel(t, db, projectID)

	securityLevelID := uuid.New().String()
	_, err := db.Exec(context.Background(),
		"INSERT INTO security_level (id, title, description, project_id, level, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		securityLevelID, "Title", "Description", projectID, models.SecurityLevelPublic, time.Now().Unix(), time.Now().Unix(), false)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionSecurityLevelModify,
		Data: map[string]interface{}{
			"securityLevelId": securityLevelID,
		},
	}

	handler.handleSecurityLevelModify(c, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp models.Response
	err = json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeMissingData, resp.ErrorCode)
}

func TestSecurityLevelHandler_Modify_InvalidLevel(t *testing.T) {
	handler, db := setupSecurityLevelTestHandler(t)
	defer db.Close()

	// Create a project and security level
	projectID := uuid.New().String()
	createTestProjectSecLevel(t, db, projectID)

	securityLevelID := uuid.New().String()
	_, err := db.Exec(context.Background(),
		"INSERT INTO security_level (id, title, description, project_id, level, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		securityLevelID, "Title", "Description", projectID, models.SecurityLevelPublic, time.Now().Unix(), time.Now().Unix(), false)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionSecurityLevelModify,
		Data: map[string]interface{}{
			"securityLevelId": securityLevelID,
			"level":           float64(10), // Invalid level
		},
	}

	handler.handleSecurityLevelModify(c, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp models.Response
	err = json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeInvalidData, resp.ErrorCode)
}

func TestSecurityLevelHandler_Remove_Success(t *testing.T) {
	handler, db := setupSecurityLevelTestHandler(t)
	defer db.Close()

	// Create a project and security level
	projectID := uuid.New().String()
	createTestProjectSecLevel(t, db, projectID)

	securityLevelID := uuid.New().String()
	_, err := db.Exec(context.Background(),
		"INSERT INTO security_level (id, title, description, project_id, level, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		securityLevelID, "Confidential", "Test security level", projectID, models.SecurityLevelConfidential, time.Now().Unix(), time.Now().Unix(), false)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionSecurityLevelRemove,
		Data: map[string]interface{}{
			"securityLevelId": securityLevelID,
		},
	}

	handler.handleSecurityLevelRemove(c, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err = json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.True(t, resp.Data["removed"].(bool))
	assert.Equal(t, securityLevelID, resp.Data["securityLevelId"])
}

func TestSecurityLevelHandler_Remove_MissingID(t *testing.T) {
	handler, db := setupSecurityLevelTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionSecurityLevelRemove,
		Data:   map[string]interface{}{},
	}

	handler.handleSecurityLevelRemove(c, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeMissingData, resp.ErrorCode)
}

func TestSecurityLevelHandler_Remove_NotFound(t *testing.T) {
	handler, db := setupSecurityLevelTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionSecurityLevelRemove,
		Data: map[string]interface{}{
			"securityLevelId": "nonexistent-id",
		},
	}

	handler.handleSecurityLevelRemove(c, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeEntityNotFound, resp.ErrorCode)
}

func TestSecurityLevelHandler_Grant_UserSuccess(t *testing.T) {
	handler, db := setupSecurityLevelTestHandler(t)
	defer db.Close()

	// Create a project and security level
	projectID := uuid.New().String()
	createTestProjectSecLevel(t, db, projectID)

	securityLevelID := uuid.New().String()
	_, err := db.Exec(context.Background(),
		"INSERT INTO security_level (id, title, description, project_id, level, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		securityLevelID, "Confidential", "Test security level", projectID, models.SecurityLevelConfidential, time.Now().Unix(), time.Now().Unix(), false)
	require.NoError(t, err)

	// Create a test user
	userID := uuid.New().String()
	createTestUserSecLevel(t, db, userID, "user1")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionSecurityLevelGrant,
		Data: map[string]interface{}{
			"securityLevelId": securityLevelID,
			"userId":          userID,
		},
	}

	handler.handleSecurityLevelGrant(c, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp models.Response
	err = json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.True(t, resp.Data["granted"].(bool))
	assert.Equal(t, userID, resp.Data["userId"])
}

func TestSecurityLevelHandler_Grant_TeamSuccess(t *testing.T) {
	handler, db := setupSecurityLevelTestHandler(t)
	defer db.Close()

	// Create a project and security level
	projectID := uuid.New().String()
	createTestProjectSecLevel(t, db, projectID)

	securityLevelID := uuid.New().String()
	_, err := db.Exec(context.Background(),
		"INSERT INTO security_level (id, title, description, project_id, level, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		securityLevelID, "Confidential", "Test security level", projectID, models.SecurityLevelConfidential, time.Now().Unix(), time.Now().Unix(), false)
	require.NoError(t, err)

	// Create a test team
	teamID := uuid.New().String()
	createTestTeamSecLevel(t, db, teamID, projectID)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionSecurityLevelGrant,
		Data: map[string]interface{}{
			"securityLevelId": securityLevelID,
			"teamId":          teamID,
		},
	}

	handler.handleSecurityLevelGrant(c, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp models.Response
	err = json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.True(t, resp.Data["granted"].(bool))
	assert.Equal(t, teamID, resp.Data["teamId"])
}

func TestSecurityLevelHandler_Grant_ProjectRoleSuccess(t *testing.T) {
	handler, db := setupSecurityLevelTestHandler(t)
	defer db.Close()

	// Create a project and security level
	projectID := uuid.New().String()
	createTestProjectSecLevel(t, db, projectID)

	securityLevelID := uuid.New().String()
	_, err := db.Exec(context.Background(),
		"INSERT INTO security_level (id, title, description, project_id, level, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		securityLevelID, "Confidential", "Test security level", projectID, models.SecurityLevelConfidential, time.Now().Unix(), time.Now().Unix(), false)
	require.NoError(t, err)

	// Create a test project role
	roleID := uuid.New().String()
	createTestProjectRoleSecLevel(t, db, roleID, "Developer", projectID)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionSecurityLevelGrant,
		Data: map[string]interface{}{
			"securityLevelId": securityLevelID,
			"projectRoleId":   roleID,
		},
	}

	handler.handleSecurityLevelGrant(c, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp models.Response
	err = json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.True(t, resp.Data["granted"].(bool))
	assert.Equal(t, roleID, resp.Data["projectRoleId"])
}

func TestSecurityLevelHandler_Grant_MissingSecurityLevelID(t *testing.T) {
	handler, db := setupSecurityLevelTestHandler(t)
	defer db.Close()

	userID := uuid.New().String()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionSecurityLevelGrant,
		Data: map[string]interface{}{
			"userId": userID,
		},
	}

	handler.handleSecurityLevelGrant(c, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeMissingData, resp.ErrorCode)
}

func TestSecurityLevelHandler_Grant_SecurityLevelNotFound(t *testing.T) {
	handler, db := setupSecurityLevelTestHandler(t)
	defer db.Close()

	userID := uuid.New().String()
	createTestUserSecLevel(t, db, userID, "user1")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionSecurityLevelGrant,
		Data: map[string]interface{}{
			"securityLevelId": "nonexistent-id",
			"userId":          userID,
		},
	}

	handler.handleSecurityLevelGrant(c, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeEntityNotFound, resp.ErrorCode)
}

func TestSecurityLevelHandler_Grant_NoRecipient(t *testing.T) {
	handler, db := setupSecurityLevelTestHandler(t)
	defer db.Close()

	// Create a project and security level
	projectID := uuid.New().String()
	createTestProjectSecLevel(t, db, projectID)

	securityLevelID := uuid.New().String()
	_, err := db.Exec(context.Background(),
		"INSERT INTO security_level (id, title, description, project_id, level, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		securityLevelID, "Confidential", "Test security level", projectID, models.SecurityLevelConfidential, time.Now().Unix(), time.Now().Unix(), false)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionSecurityLevelGrant,
		Data: map[string]interface{}{
			"securityLevelId": securityLevelID,
		},
	}

	handler.handleSecurityLevelGrant(c, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp models.Response
	err = json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeMissingData, resp.ErrorCode)
}

func TestSecurityLevelHandler_Grant_MultipleRecipients(t *testing.T) {
	handler, db := setupSecurityLevelTestHandler(t)
	defer db.Close()

	// Create a project and security level
	projectID := uuid.New().String()
	createTestProjectSecLevel(t, db, projectID)

	securityLevelID := uuid.New().String()
	_, err := db.Exec(context.Background(),
		"INSERT INTO security_level (id, title, description, project_id, level, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		securityLevelID, "Confidential", "Test security level", projectID, models.SecurityLevelConfidential, time.Now().Unix(), time.Now().Unix(), false)
	require.NoError(t, err)

	userID := uuid.New().String()
	teamID := uuid.New().String()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionSecurityLevelGrant,
		Data: map[string]interface{}{
			"securityLevelId": securityLevelID,
			"userId":          userID,
			"teamId":          teamID,
		},
	}

	handler.handleSecurityLevelGrant(c, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp models.Response
	err = json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeInvalidData, resp.ErrorCode)
}

func TestSecurityLevelHandler_Grant_Duplicate(t *testing.T) {
	handler, db := setupSecurityLevelTestHandler(t)
	defer db.Close()

	// Create a project and security level
	projectID := uuid.New().String()
	createTestProjectSecLevel(t, db, projectID)

	securityLevelID := uuid.New().String()
	_, err := db.Exec(context.Background(),
		"INSERT INTO security_level (id, title, description, project_id, level, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		securityLevelID, "Confidential", "Test security level", projectID, models.SecurityLevelConfidential, time.Now().Unix(), time.Now().Unix(), false)
	require.NoError(t, err)

	// Create a test user
	userID := uuid.New().String()
	createTestUserSecLevel(t, db, userID, "user1")

	// Grant access once
	_, err = db.Exec(context.Background(),
		"INSERT INTO security_level_permission_mapping (id, security_level_id, user_id, created, deleted) VALUES (?, ?, ?, ?, ?)",
		uuid.New().String(), securityLevelID, userID, time.Now().Unix(), false)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionSecurityLevelGrant,
		Data: map[string]interface{}{
			"securityLevelId": securityLevelID,
			"userId":          userID,
		},
	}

	handler.handleSecurityLevelGrant(c, req)

	assert.Equal(t, http.StatusConflict, w.Code)

	var resp models.Response
	err = json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeEntityAlreadyExists, resp.ErrorCode)
}

func TestSecurityLevelHandler_Revoke_UserSuccess(t *testing.T) {
	handler, db := setupSecurityLevelTestHandler(t)
	defer db.Close()

	// Create a project and security level
	projectID := uuid.New().String()
	createTestProjectSecLevel(t, db, projectID)

	securityLevelID := uuid.New().String()
	_, err := db.Exec(context.Background(),
		"INSERT INTO security_level (id, title, description, project_id, level, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		securityLevelID, "Confidential", "Test security level", projectID, models.SecurityLevelConfidential, time.Now().Unix(), time.Now().Unix(), false)
	require.NoError(t, err)

	// Create a test user and grant access
	userID := uuid.New().String()
	createTestUserSecLevel(t, db, userID, "user1")

	_, err = db.Exec(context.Background(),
		"INSERT INTO security_level_permission_mapping (id, security_level_id, user_id, created, deleted) VALUES (?, ?, ?, ?, ?)",
		uuid.New().String(), securityLevelID, userID, time.Now().Unix(), false)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionSecurityLevelRevoke,
		Data: map[string]interface{}{
			"securityLevelId": securityLevelID,
			"userId":          userID,
		},
	}

	handler.handleSecurityLevelRevoke(c, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err = json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.True(t, resp.Data["revoked"].(bool))
	assert.Equal(t, userID, resp.Data["userId"])
}

func TestSecurityLevelHandler_Revoke_TeamSuccess(t *testing.T) {
	handler, db := setupSecurityLevelTestHandler(t)
	defer db.Close()

	// Create a project and security level
	projectID := uuid.New().String()
	createTestProjectSecLevel(t, db, projectID)

	securityLevelID := uuid.New().String()
	_, err := db.Exec(context.Background(),
		"INSERT INTO security_level (id, title, description, project_id, level, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		securityLevelID, "Confidential", "Test security level", projectID, models.SecurityLevelConfidential, time.Now().Unix(), time.Now().Unix(), false)
	require.NoError(t, err)

	// Create a test team and grant access
	teamID := uuid.New().String()
	createTestTeamSecLevel(t, db, teamID, projectID)

	_, err = db.Exec(context.Background(),
		"INSERT INTO security_level_permission_mapping (id, security_level_id, team_id, created, deleted) VALUES (?, ?, ?, ?, ?)",
		uuid.New().String(), securityLevelID, teamID, time.Now().Unix(), false)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionSecurityLevelRevoke,
		Data: map[string]interface{}{
			"securityLevelId": securityLevelID,
			"teamId":          teamID,
		},
	}

	handler.handleSecurityLevelRevoke(c, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err = json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.True(t, resp.Data["revoked"].(bool))
	assert.Equal(t, teamID, resp.Data["teamId"])
}

func TestSecurityLevelHandler_Revoke_MissingSecurityLevelID(t *testing.T) {
	handler, db := setupSecurityLevelTestHandler(t)
	defer db.Close()

	userID := uuid.New().String()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionSecurityLevelRevoke,
		Data: map[string]interface{}{
			"userId": userID,
		},
	}

	handler.handleSecurityLevelRevoke(c, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeMissingData, resp.ErrorCode)
}

func TestSecurityLevelHandler_Revoke_NoRecipient(t *testing.T) {
	handler, db := setupSecurityLevelTestHandler(t)
	defer db.Close()

	securityLevelID := uuid.New().String()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionSecurityLevelRevoke,
		Data: map[string]interface{}{
			"securityLevelId": securityLevelID,
		},
	}

	handler.handleSecurityLevelRevoke(c, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeMissingData, resp.ErrorCode)
}

func TestSecurityLevelHandler_Revoke_NotFound(t *testing.T) {
	handler, db := setupSecurityLevelTestHandler(t)
	defer db.Close()

	securityLevelID := uuid.New().String()
	userID := uuid.New().String()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionSecurityLevelRevoke,
		Data: map[string]interface{}{
			"securityLevelId": securityLevelID,
			"userId":          userID,
		},
	}

	handler.handleSecurityLevelRevoke(c, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeEntityNotFound, resp.ErrorCode)
}

func TestSecurityLevelHandler_Check_DirectUserAccess(t *testing.T) {
	handler, db := setupSecurityLevelTestHandler(t)
	defer db.Close()

	// Create a project and security level
	projectID := uuid.New().String()
	createTestProjectSecLevel(t, db, projectID)

	securityLevelID := uuid.New().String()
	_, err := db.Exec(context.Background(),
		"INSERT INTO security_level (id, title, description, project_id, level, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		securityLevelID, "Confidential", "Test security level", projectID, models.SecurityLevelConfidential, time.Now().Unix(), time.Now().Unix(), false)
	require.NoError(t, err)

	// Create a test user and grant access
	userID := uuid.New().String()
	createTestUserSecLevel(t, db, userID, "user1")

	_, err = db.Exec(context.Background(),
		"INSERT INTO security_level_permission_mapping (id, security_level_id, user_id, created, deleted) VALUES (?, ?, ?, ?, ?)",
		uuid.New().String(), securityLevelID, userID, time.Now().Unix(), false)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionSecurityLevelCheck,
		Data: map[string]interface{}{
			"securityLevelId": securityLevelID,
			"userId":          userID,
		},
	}

	handler.handleSecurityLevelCheck(c, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err = json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.True(t, resp.Data["hasAccess"].(bool))
	assert.Equal(t, userID, resp.Data["userId"])
}

func TestSecurityLevelHandler_Check_TeamAccess(t *testing.T) {
	handler, db := setupSecurityLevelTestHandler(t)
	defer db.Close()

	// Create a project and security level
	projectID := uuid.New().String()
	createTestProjectSecLevel(t, db, projectID)

	securityLevelID := uuid.New().String()
	_, err := db.Exec(context.Background(),
		"INSERT INTO security_level (id, title, description, project_id, level, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		securityLevelID, "Confidential", "Test security level", projectID, models.SecurityLevelConfidential, time.Now().Unix(), time.Now().Unix(), false)
	require.NoError(t, err)

	// Create a test user and team
	userID := uuid.New().String()
	createTestUserSecLevel(t, db, userID, "user1")

	teamID := uuid.New().String()
	createTestTeamSecLevel(t, db, teamID, projectID)
	createTestTeamUserSecLevel(t, db, teamID, userID)

	// Grant access to team
	_, err = db.Exec(context.Background(),
		"INSERT INTO security_level_permission_mapping (id, security_level_id, team_id, created, deleted) VALUES (?, ?, ?, ?, ?)",
		uuid.New().String(), securityLevelID, teamID, time.Now().Unix(), false)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionSecurityLevelCheck,
		Data: map[string]interface{}{
			"securityLevelId": securityLevelID,
			"userId":          userID,
		},
	}

	handler.handleSecurityLevelCheck(c, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err = json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.True(t, resp.Data["hasAccess"].(bool))
}

func TestSecurityLevelHandler_Check_RoleAccess(t *testing.T) {
	handler, db := setupSecurityLevelTestHandler(t)
	defer db.Close()

	// Create a project and security level
	projectID := uuid.New().String()
	createTestProjectSecLevel(t, db, projectID)

	securityLevelID := uuid.New().String()
	_, err := db.Exec(context.Background(),
		"INSERT INTO security_level (id, title, description, project_id, level, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		securityLevelID, "Confidential", "Test security level", projectID, models.SecurityLevelConfidential, time.Now().Unix(), time.Now().Unix(), false)
	require.NoError(t, err)

	// Create a test user and project role
	userID := uuid.New().String()
	createTestUserSecLevel(t, db, userID, "user1")

	roleID := uuid.New().String()
	createTestProjectRoleSecLevel(t, db, roleID, "Developer", projectID)
	createTestProjectRoleUserMappingSecLevel(t, db, roleID, projectID, userID)

	// Grant access to role
	_, err = db.Exec(context.Background(),
		"INSERT INTO security_level_permission_mapping (id, security_level_id, project_role_id, created, deleted) VALUES (?, ?, ?, ?, ?)",
		uuid.New().String(), securityLevelID, roleID, time.Now().Unix(), false)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionSecurityLevelCheck,
		Data: map[string]interface{}{
			"securityLevelId": securityLevelID,
			"userId":          userID,
		},
	}

	handler.handleSecurityLevelCheck(c, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err = json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.True(t, resp.Data["hasAccess"].(bool))
}

func TestSecurityLevelHandler_Check_NoAccess(t *testing.T) {
	handler, db := setupSecurityLevelTestHandler(t)
	defer db.Close()

	// Create a project and security level
	projectID := uuid.New().String()
	createTestProjectSecLevel(t, db, projectID)

	securityLevelID := uuid.New().String()
	_, err := db.Exec(context.Background(),
		"INSERT INTO security_level (id, title, description, project_id, level, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		securityLevelID, "Confidential", "Test security level", projectID, models.SecurityLevelConfidential, time.Now().Unix(), time.Now().Unix(), false)
	require.NoError(t, err)

	// Create a test user (no access granted)
	userID := uuid.New().String()
	createTestUserSecLevel(t, db, userID, "user1")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionSecurityLevelCheck,
		Data: map[string]interface{}{
			"securityLevelId": securityLevelID,
			"userId":          userID,
		},
	}

	handler.handleSecurityLevelCheck(c, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err = json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.False(t, resp.Data["hasAccess"].(bool))
}

func TestSecurityLevelHandler_Check_MissingSecurityLevelID(t *testing.T) {
	handler, db := setupSecurityLevelTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionSecurityLevelCheck,
		Data:   map[string]interface{}{},
	}

	handler.handleSecurityLevelCheck(c, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeMissingData, resp.ErrorCode)
}

func TestSecurityLevelHandler_Unauthorized(t *testing.T) {
	handler, db := setupSecurityLevelTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	// No username set

	req := &models.Request{
		Action: models.ActionSecurityLevelCreate,
		Data:   map[string]interface{}{},
	}

	handler.handleSecurityLevelCreate(c, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestSecurityLevelHandler_PermissionDenied(t *testing.T) {
	handler, db := setupSecurityLevelTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "unauthorized_user")

	req := &models.Request{
		Action: models.ActionSecurityLevelCreate,
		Data: map[string]interface{}{
			"title":     "Confidential",
			"projectId": uuid.New().String(),
			"level":     float64(models.SecurityLevelConfidential),
		},
	}

	handler.handleSecurityLevelCreate(c, req)

	assert.Equal(t, http.StatusForbidden, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeForbidden, resp.ErrorCode)
}
