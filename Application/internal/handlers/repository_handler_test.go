package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"helixtrack.ru/core/internal/models"
)

// setupRepositoryTestHandler creates test handler with repository tables
func setupRepositoryTestHandler(t *testing.T) *Handler {
	handler := setupTestHandler(t)

	// Create repository table
	_, err := handler.db.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS repository (
			id TEXT PRIMARY KEY,
			repository TEXT NOT NULL,
			description TEXT,
			repository_type_id TEXT NOT NULL,
			created INTEGER NOT NULL,
			modified INTEGER NOT NULL,
			deleted INTEGER DEFAULT 0
		)
	`)
	require.NoError(t, err)

	// Create repository_type table
	_, err = handler.db.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS repository_type (
			id TEXT PRIMARY KEY,
			title TEXT NOT NULL,
			description TEXT,
			created INTEGER NOT NULL,
			modified INTEGER NOT NULL,
			deleted INTEGER DEFAULT 0
		)
	`)
	require.NoError(t, err)

	// Create repository_project_mapping table
	_, err = handler.db.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS repository_project_mapping (
			id TEXT PRIMARY KEY,
			repository_id TEXT NOT NULL,
			project_id TEXT NOT NULL,
			created INTEGER NOT NULL,
			modified INTEGER NOT NULL,
			deleted INTEGER DEFAULT 0
		)
	`)
	require.NoError(t, err)

	// Create repository_commit_ticket_mapping table
	_, err = handler.db.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS repository_commit_ticket_mapping (
			id TEXT PRIMARY KEY,
			repository_id TEXT NOT NULL,
			ticket_id TEXT NOT NULL,
			commit_hash TEXT NOT NULL,
			created INTEGER NOT NULL,
			modified INTEGER NOT NULL,
			deleted INTEGER DEFAULT 0
		)
	`)
	require.NoError(t, err)

	return handler
}

// ===== Repository CRUD Tests =====

// TestRepositoryHandler_Create_Success tests creating a repository
func TestRepositoryHandler_Create_Success(t *testing.T) {
	handler := setupRepositoryTestHandler(t)
	gin.SetMode(gin.TestMode)

	reqBody := models.Request{
		Action: models.ActionRepositoryCreate,
		Data: map[string]interface{}{
			"repository":         "https://github.com/user/repo.git",
			"description":        "Test Repository",
			"repository_type_id": models.RepositoryTypeGit,
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")

	handler.handleRepositoryCreate(c, &reqBody)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response models.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, response.ErrorCode)
}

// TestRepositoryHandler_Create_MissingFields tests creating without required fields
func TestRepositoryHandler_Create_MissingFields(t *testing.T) {
	handler := setupRepositoryTestHandler(t)
	gin.SetMode(gin.TestMode)

	testCases := []struct {
		name string
		data map[string]interface{}
	}{
		{
			"MissingRepository",
			map[string]interface{}{
				"repository_type_id": models.RepositoryTypeGit,
			},
		},
		{
			"MissingRepositoryTypeID",
			map[string]interface{}{
				"repository": "https://github.com/user/repo.git",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			reqBody := models.Request{
				Action: models.ActionRepositoryCreate,
				Data:   tc.data,
			}

			body, _ := json.Marshal(reqBody)
			req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
			w := httptest.NewRecorder()

			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.Set("username", "testuser")

			handler.handleRepositoryCreate(c, &reqBody)

			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	}
}

// TestRepositoryHandler_Read_Success tests reading a repository
func TestRepositoryHandler_Read_Success(t *testing.T) {
	handler := setupRepositoryTestHandler(t)
	gin.SetMode(gin.TestMode)

	// Insert test repository
	repoID := generateTestID()
	now := time.Now().Unix()
	_, err := handler.db.Exec(context.Background(),
		`INSERT INTO repository (id, repository, description, repository_type_id, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		repoID, "https://github.com/test/repo.git", "Test Repo", models.RepositoryTypeGit, now, now, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionRepositoryRead,
		Data: map[string]interface{}{
			"id": repoID,
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")

	handler.handleRepositoryRead(c, &reqBody)

	assert.Equal(t, http.StatusOK, w.Code)
}

// TestRepositoryHandler_Read_NotFound tests reading non-existent repository
func TestRepositoryHandler_Read_NotFound(t *testing.T) {
	handler := setupRepositoryTestHandler(t)
	gin.SetMode(gin.TestMode)

	reqBody := models.Request{
		Action: models.ActionRepositoryRead,
		Data: map[string]interface{}{
			"id": "non-existent-id",
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")

	handler.handleRepositoryRead(c, &reqBody)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// TestRepositoryHandler_List_Success tests listing repositories
func TestRepositoryHandler_List_Success(t *testing.T) {
	handler := setupRepositoryTestHandler(t)
	gin.SetMode(gin.TestMode)

	// Insert test repositories
	now := time.Now().Unix()
	for i := 0; i < 3; i++ {
		_, err := handler.db.Exec(context.Background(),
			`INSERT INTO repository (id, repository, description, repository_type_id, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?)`,
			generateTestID(), "https://github.com/test/repo"+string(rune('A'+i))+".git", "Repo "+string(rune('A'+i)),
			models.RepositoryTypeGit, now, now, 0)
		require.NoError(t, err)
	}

	reqBody := models.Request{
		Action: models.ActionRepositoryList,
		Data:   map[string]interface{}{},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")

	handler.handleRepositoryList(c, &reqBody)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	dataMap := response.Data
	count := int(dataMap["count"].(float64))
	assert.Equal(t, 3, count)
}

// TestRepositoryHandler_Modify_Success tests modifying a repository
func TestRepositoryHandler_Modify_Success(t *testing.T) {
	handler := setupRepositoryTestHandler(t)
	gin.SetMode(gin.TestMode)

	// Insert test repository
	repoID := generateTestID()
	now := time.Now().Unix()
	_, err := handler.db.Exec(context.Background(),
		`INSERT INTO repository (id, repository, description, repository_type_id, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		repoID, "https://github.com/test/old.git", "Old Description", models.RepositoryTypeGit, now, now, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionRepositoryModify,
		Data: map[string]interface{}{
			"id":          repoID,
			"repository":  "https://github.com/test/new.git",
			"description": "New Description",
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")

	handler.handleRepositoryModify(c, &reqBody)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify changes
	var repository, description string
	err = handler.db.QueryRow(context.Background(), "SELECT repository, description FROM repository WHERE id = ?", repoID).Scan(&repository, &description)
	require.NoError(t, err)
	assert.Equal(t, "https://github.com/test/new.git", repository)
	assert.Equal(t, "New Description", description)
}

// TestRepositoryHandler_Modify_NotFound tests modifying non-existent repository
func TestRepositoryHandler_Modify_NotFound(t *testing.T) {
	handler := setupRepositoryTestHandler(t)
	gin.SetMode(gin.TestMode)

	reqBody := models.Request{
		Action: models.ActionRepositoryModify,
		Data: map[string]interface{}{
			"id":         "non-existent-id",
			"repository": "https://github.com/test/new.git",
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")

	handler.handleRepositoryModify(c, &reqBody)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// TestRepositoryHandler_Remove_Success tests soft-deleting a repository
func TestRepositoryHandler_Remove_Success(t *testing.T) {
	handler := setupRepositoryTestHandler(t)
	gin.SetMode(gin.TestMode)

	// Insert test repository
	repoID := generateTestID()
	now := time.Now().Unix()
	_, err := handler.db.Exec(context.Background(),
		`INSERT INTO repository (id, repository, description, repository_type_id, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		repoID, "https://github.com/test/repo.git", "Test", models.RepositoryTypeGit, now, now, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionRepositoryRemove,
		Data: map[string]interface{}{
			"id": repoID,
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")

	handler.handleRepositoryRemove(c, &reqBody)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify soft delete
	var deleted int
	err = handler.db.QueryRow(context.Background(), "SELECT deleted FROM repository WHERE id = ?", repoID).Scan(&deleted)
	require.NoError(t, err)
	assert.Equal(t, 1, deleted)
}

// ===== Repository Type CRUD Tests =====

// TestRepositoryTypeHandler_Create_Success tests creating a repository type
func TestRepositoryTypeHandler_Create_Success(t *testing.T) {
	handler := setupRepositoryTestHandler(t)
	gin.SetMode(gin.TestMode)

	reqBody := models.Request{
		Action: models.ActionRepositoryTypeCreate,
		Data: map[string]interface{}{
			"title":       "Git",
			"description": "Git Version Control System",
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")

	handler.handleRepositoryTypeCreate(c, &reqBody)

	assert.Equal(t, http.StatusCreated, w.Code)
}

// TestRepositoryTypeHandler_Create_AllTypes tests creating all common repository types
func TestRepositoryTypeHandler_Create_AllTypes(t *testing.T) {
	handler := setupRepositoryTestHandler(t)
	gin.SetMode(gin.TestMode)

	types := []struct {
		title       string
		description string
	}{
		{"Git", "Git Version Control"},
		{"SVN", "Subversion"},
		{"Mercurial", "Mercurial SCM"},
		{"CVS", "Concurrent Versions System"},
		{"Perforce", "Perforce Helix"},
	}

	for _, repoType := range types {
		t.Run(repoType.title, func(t *testing.T) {
			reqBody := models.Request{
				Action: models.ActionRepositoryTypeCreate,
				Data: map[string]interface{}{
					"title":       repoType.title,
					"description": repoType.description,
				},
			}

			body, _ := json.Marshal(reqBody)
			req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
			w := httptest.NewRecorder()

			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.Set("username", "testuser")

			handler.handleRepositoryTypeCreate(c, &reqBody)

			assert.Equal(t, http.StatusCreated, w.Code)
		})
	}
}

// TestRepositoryTypeHandler_Read_Success tests reading a repository type
func TestRepositoryTypeHandler_Read_Success(t *testing.T) {
	handler := setupRepositoryTestHandler(t)
	gin.SetMode(gin.TestMode)

	// Insert test repository type
	typeID := generateTestID()
	now := time.Now().Unix()
	_, err := handler.db.Exec(context.Background(),
		`INSERT INTO repository_type (id, title, description, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)`,
		typeID, "Git", "Git VCS", now, now, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionRepositoryTypeRead,
		Data: map[string]interface{}{
			"id": typeID,
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")

	handler.handleRepositoryTypeRead(c, &reqBody)

	assert.Equal(t, http.StatusOK, w.Code)
}

// TestRepositoryTypeHandler_List_Success tests listing repository types
func TestRepositoryTypeHandler_List_Success(t *testing.T) {
	handler := setupRepositoryTestHandler(t)
	gin.SetMode(gin.TestMode)

	// Insert test repository types
	now := time.Now().Unix()
	_, err := handler.db.Exec(context.Background(),
		`INSERT INTO repository_type (id, title, description, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)`,
		generateTestID(), "Git", "Git VCS", now, now, 0)
	require.NoError(t, err)

	_, err = handler.db.Exec(context.Background(),
		`INSERT INTO repository_type (id, title, description, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)`,
		generateTestID(), "SVN", "Subversion", now, now, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionRepositoryTypeList,
		Data:   map[string]interface{}{},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")

	handler.handleRepositoryTypeList(c, &reqBody)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Response
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	dataMap := response.Data
	count := int(dataMap["count"].(float64))
	assert.Equal(t, 2, count)
}

// TestRepositoryTypeHandler_Modify_Success tests modifying a repository type
func TestRepositoryTypeHandler_Modify_Success(t *testing.T) {
	handler := setupRepositoryTestHandler(t)
	gin.SetMode(gin.TestMode)

	// Insert test repository type
	typeID := generateTestID()
	now := time.Now().Unix()
	_, err := handler.db.Exec(context.Background(),
		`INSERT INTO repository_type (id, title, description, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)`,
		typeID, "Old Title", "Old Desc", now, now, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionRepositoryTypeModify,
		Data: map[string]interface{}{
			"id":          typeID,
			"title":       "Git",
			"description": "Git Version Control System",
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")

	handler.handleRepositoryTypeModify(c, &reqBody)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify changes
	var title, description string
	err = handler.db.QueryRow(context.Background(), "SELECT title, description FROM repository_type WHERE id = ?", typeID).Scan(&title, &description)
	require.NoError(t, err)
	assert.Equal(t, "Git", title)
	assert.Equal(t, "Git Version Control System", description)
}

// TestRepositoryTypeHandler_Remove_Success tests soft-deleting a repository type
func TestRepositoryTypeHandler_Remove_Success(t *testing.T) {
	handler := setupRepositoryTestHandler(t)
	gin.SetMode(gin.TestMode)

	// Insert test repository type
	typeID := generateTestID()
	now := time.Now().Unix()
	_, err := handler.db.Exec(context.Background(),
		`INSERT INTO repository_type (id, title, description, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)`,
		typeID, "Git", "Git VCS", now, now, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionRepositoryTypeRemove,
		Data: map[string]interface{}{
			"id": typeID,
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")

	handler.handleRepositoryTypeRemove(c, &reqBody)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify soft delete
	var deleted int
	err = handler.db.QueryRow(context.Background(), "SELECT deleted FROM repository_type WHERE id = ?", typeID).Scan(&deleted)
	require.NoError(t, err)
	assert.Equal(t, 1, deleted)
}

// ===== Repository-Project Mapping Tests =====

// TestRepositoryHandler_AssignProject_Success tests assigning repository to project
func TestRepositoryHandler_AssignProject_Success(t *testing.T) {
	handler := setupRepositoryTestHandler(t)
	gin.SetMode(gin.TestMode)

	repositoryID := generateTestID()
	projectID := generateTestID()

	reqBody := models.Request{
		Action: models.ActionRepositoryAssignProject,
		Data: map[string]interface{}{
			"repository_id": repositoryID,
			"project_id":    projectID,
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")

	handler.handleRepositoryAssignProject(c, &reqBody)

	assert.Equal(t, http.StatusCreated, w.Code)

	// Verify mapping created
	var count int
	err := handler.db.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM repository_project_mapping WHERE repository_id = ? AND project_id = ? AND deleted = 0",
		repositoryID, projectID).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

// TestRepositoryHandler_UnassignProject_Success tests unassigning repository from project
func TestRepositoryHandler_UnassignProject_Success(t *testing.T) {
	handler := setupRepositoryTestHandler(t)
	gin.SetMode(gin.TestMode)

	repositoryID := generateTestID()
	projectID := generateTestID()

	// Create mapping
	now := time.Now().Unix()
	_, err := handler.db.Exec(context.Background(),
		`INSERT INTO repository_project_mapping (id, repository_id, project_id, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)`,
		generateTestID(), repositoryID, projectID, now, now, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionRepositoryUnassignProject,
		Data: map[string]interface{}{
			"repository_id": repositoryID,
			"project_id":    projectID,
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")

	handler.handleRepositoryUnassignProject(c, &reqBody)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify mapping soft-deleted
	var deleted int
	err = handler.db.QueryRow(context.Background(),
		"SELECT deleted FROM repository_project_mapping WHERE repository_id = ? AND project_id = ?",
		repositoryID, projectID).Scan(&deleted)
	require.NoError(t, err)
	assert.Equal(t, 1, deleted)
}

// TestRepositoryHandler_ListProjects_Success tests listing projects for repository
func TestRepositoryHandler_ListProjects_Success(t *testing.T) {
	handler := setupRepositoryTestHandler(t)
	gin.SetMode(gin.TestMode)

	repositoryID := generateTestID()
	now := time.Now().Unix()

	// Create multiple project mappings
	for i := 0; i < 3; i++ {
		projectID := generateTestID()
		_, err := handler.db.Exec(context.Background(),
			`INSERT INTO repository_project_mapping (id, repository_id, project_id, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)`,
			generateTestID(), repositoryID, projectID, now, now, 0)
		require.NoError(t, err)
	}

	reqBody := models.Request{
		Action: models.ActionRepositoryListProjects,
		Data: map[string]interface{}{
			"repository_id": repositoryID,
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")

	handler.handleRepositoryListProjects(c, &reqBody)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	dataMap := response.Data
	count := int(dataMap["count"].(float64))
	assert.Equal(t, 3, count)
}

// ===== Repository Commit-Ticket Mapping Tests =====

// TestRepositoryHandler_AddCommit_Success tests adding commit to ticket
func TestRepositoryHandler_AddCommit_Success(t *testing.T) {
	handler := setupRepositoryTestHandler(t)
	gin.SetMode(gin.TestMode)

	repositoryID := generateTestID()
	ticketID := generateTestID()
	commitHash := "abc123def456"

	reqBody := models.Request{
		Action: models.ActionRepositoryAddCommit,
		Data: map[string]interface{}{
			"repository_id": repositoryID,
			"ticket_id":     ticketID,
			"commit_hash":   commitHash,
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")

	handler.handleRepositoryAddCommit(c, &reqBody)

	assert.Equal(t, http.StatusCreated, w.Code)

	// Verify commit mapping created
	var count int
	err := handler.db.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM repository_commit_ticket_mapping WHERE commit_hash = ? AND deleted = 0",
		commitHash).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

// TestRepositoryHandler_AddCommit_MissingFields tests adding commit with missing fields
func TestRepositoryHandler_AddCommit_MissingFields(t *testing.T) {
	handler := setupRepositoryTestHandler(t)
	gin.SetMode(gin.TestMode)

	testCases := []struct {
		name string
		data map[string]interface{}
	}{
		{
			"MissingRepositoryID",
			map[string]interface{}{
				"ticket_id":   "ticket1",
				"commit_hash": "abc123",
			},
		},
		{
			"MissingTicketID",
			map[string]interface{}{
				"repository_id": "repo1",
				"commit_hash":   "abc123",
			},
		},
		{
			"MissingCommitHash",
			map[string]interface{}{
				"repository_id": "repo1",
				"ticket_id":     "ticket1",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			reqBody := models.Request{
				Action: models.ActionRepositoryAddCommit,
				Data:   tc.data,
			}

			body, _ := json.Marshal(reqBody)
			req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
			w := httptest.NewRecorder()

			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.Set("username", "testuser")

			handler.handleRepositoryAddCommit(c, &reqBody)

			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	}
}

// TestRepositoryHandler_RemoveCommit_Success tests removing commit from ticket
func TestRepositoryHandler_RemoveCommit_Success(t *testing.T) {
	handler := setupRepositoryTestHandler(t)
	gin.SetMode(gin.TestMode)

	commitHash := "abc123def456"
	now := time.Now().Unix()

	// Create commit mapping
	_, err := handler.db.Exec(context.Background(),
		`INSERT INTO repository_commit_ticket_mapping (id, repository_id, ticket_id, commit_hash, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		generateTestID(), generateTestID(), generateTestID(), commitHash, now, now, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionRepositoryRemoveCommit,
		Data: map[string]interface{}{
			"commit_hash": commitHash,
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")

	handler.handleRepositoryRemoveCommit(c, &reqBody)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify commit mapping soft-deleted
	var deleted int
	err = handler.db.QueryRow(context.Background(),
		"SELECT deleted FROM repository_commit_ticket_mapping WHERE commit_hash = ?",
		commitHash).Scan(&deleted)
	require.NoError(t, err)
	assert.Equal(t, 1, deleted)
}

// TestRepositoryHandler_ListCommits_Success tests listing commits for ticket
func TestRepositoryHandler_ListCommits_Success(t *testing.T) {
	handler := setupRepositoryTestHandler(t)
	gin.SetMode(gin.TestMode)

	ticketID := generateTestID()
	now := time.Now().Unix()

	// Create multiple commit mappings
	for i := 0; i < 4; i++ {
		commitHash := "abc" + string(rune('0'+i)) + "def456"
		_, err := handler.db.Exec(context.Background(),
			`INSERT INTO repository_commit_ticket_mapping (id, repository_id, ticket_id, commit_hash, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?)`,
			generateTestID(), generateTestID(), ticketID, commitHash, now, now, 0)
		require.NoError(t, err)
	}

	reqBody := models.Request{
		Action: models.ActionRepositoryListCommits,
		Data: map[string]interface{}{
			"ticket_id": ticketID,
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")

	handler.handleRepositoryListCommits(c, &reqBody)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	dataMap := response.Data
	count := int(dataMap["count"].(float64))
	assert.Equal(t, 4, count)
}

// TestRepositoryHandler_GetCommit_Success tests getting commit by hash
func TestRepositoryHandler_GetCommit_Success(t *testing.T) {
	handler := setupRepositoryTestHandler(t)
	gin.SetMode(gin.TestMode)

	commitHash := "abc123def456"
	now := time.Now().Unix()

	// Create commit mapping
	_, err := handler.db.Exec(context.Background(),
		`INSERT INTO repository_commit_ticket_mapping (id, repository_id, ticket_id, commit_hash, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		generateTestID(), generateTestID(), generateTestID(), commitHash, now, now, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionRepositoryGetCommit,
		Data: map[string]interface{}{
			"commit_hash": commitHash,
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")

	handler.handleRepositoryGetCommit(c, &reqBody)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Response
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	dataMap := response.Data
	commit := dataMap["commit"].(map[string]interface{})
	assert.Equal(t, commitHash, commit["commitHash"])
}

// TestRepositoryHandler_GetCommit_NotFound tests getting non-existent commit
func TestRepositoryHandler_GetCommit_NotFound(t *testing.T) {
	handler := setupRepositoryTestHandler(t)
	gin.SetMode(gin.TestMode)

	reqBody := models.Request{
		Action: models.ActionRepositoryGetCommit,
		Data: map[string]interface{}{
			"commit_hash": "nonexistent",
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")

	handler.handleRepositoryGetCommit(c, &reqBody)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// TestRepositoryHandler_FullCRUDCycle tests complete repository lifecycle
func TestRepositoryHandler_FullCRUDCycle(t *testing.T) {
	handler := setupRepositoryTestHandler(t)
	gin.SetMode(gin.TestMode)

	// 1. Create Repository
	reqBody := models.Request{
		Action: models.ActionRepositoryCreate,
		Data: map[string]interface{}{
			"repository":         "https://github.com/test/lifecycle.git",
			"description":        "Lifecycle Test",
			"repository_type_id": models.RepositoryTypeGit,
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")

	handler.handleRepositoryCreate(c, &reqBody)
	assert.Equal(t, http.StatusCreated, w.Code)

	var createResponse models.Response
	json.Unmarshal(w.Body.Bytes(), &createResponse)
	dataMap := createResponse.Data
	repository := dataMap["repository"].(map[string]interface{})
	repoID := repository["id"].(string)

	// 2. Read Repository
	reqBody.Action = models.ActionRepositoryRead
	reqBody.Data = map[string]interface{}{"id": repoID}
	body, _ = json.Marshal(reqBody)
	req = httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")

	handler.handleRepositoryRead(c, &reqBody)
	assert.Equal(t, http.StatusOK, w.Code)

	// 3. Modify Repository
	reqBody.Action = models.ActionRepositoryModify
	reqBody.Data = map[string]interface{}{
		"id":          repoID,
		"description": "Modified Lifecycle",
	}
	body, _ = json.Marshal(reqBody)
	req = httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")

	handler.handleRepositoryModify(c, &reqBody)
	assert.Equal(t, http.StatusOK, w.Code)

	// 4. Remove Repository
	reqBody.Action = models.ActionRepositoryRemove
	reqBody.Data = map[string]interface{}{"id": repoID}
	body, _ = json.Marshal(reqBody)
	req = httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")

	handler.handleRepositoryRemove(c, &reqBody)
	assert.Equal(t, http.StatusOK, w.Code)

	// 5. Verify deleted
	var deleted int
	err := handler.db.QueryRow(context.Background(), "SELECT deleted FROM repository WHERE id = ?", repoID).Scan(&deleted)
	require.NoError(t, err)
	assert.Equal(t, 1, deleted)
}
