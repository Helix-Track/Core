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
	"helixtrack.ru/core/internal/config"
	"helixtrack.ru/core/internal/database"
	"helixtrack.ru/core/internal/models"
	"helixtrack.ru/core/internal/services"
)

// setupDocumentTestHandler creates a handler with in-memory database and document schema
func setupDocumentTestHandler(t *testing.T) *Handler {
	db, err := database.NewDatabase(config.DatabaseConfig{
		Type:       "sqlite",
		SQLitePath: ":memory:",
	})
	require.NoError(t, err)

	// Create document tables (simplified for testing)
	ctx := context.Background()
	queries := []string{
		`CREATE TABLE document (
			id TEXT PRIMARY KEY,
			title TEXT NOT NULL,
			space_id TEXT NOT NULL,
			parent_id TEXT,
			type_id TEXT NOT NULL,
			project_id TEXT,
			creator_id TEXT NOT NULL,
			version INTEGER NOT NULL DEFAULT 1,
			position INTEGER DEFAULT 0,
			is_published INTEGER DEFAULT 0,
			is_archived INTEGER DEFAULT 0,
			publish_date INTEGER,
			created INTEGER NOT NULL,
			modified INTEGER NOT NULL,
			deleted INTEGER DEFAULT 0
		)`,
		`CREATE TABLE document_space (
			id TEXT PRIMARY KEY,
			key TEXT NOT NULL UNIQUE,
			name TEXT NOT NULL,
			description TEXT,
			owner_id TEXT NOT NULL,
			is_public INTEGER DEFAULT 0,
			created INTEGER NOT NULL,
			modified INTEGER NOT NULL,
			deleted INTEGER DEFAULT 0
		)`,
		`CREATE TABLE document_content (
			id TEXT PRIMARY KEY,
			document_id TEXT NOT NULL,
			version INTEGER NOT NULL,
			content_type TEXT NOT NULL,
			content TEXT,
			content_hash TEXT,
			size_bytes INTEGER DEFAULT 0,
			created INTEGER NOT NULL,
			modified INTEGER NOT NULL,
			deleted INTEGER DEFAULT 0
		)`,
		`CREATE TABLE document_version (
			id TEXT PRIMARY KEY,
			document_id TEXT NOT NULL,
			version_number INTEGER NOT NULL,
			editor_id TEXT NOT NULL,
			change_summary TEXT,
			created INTEGER NOT NULL,
			deleted INTEGER DEFAULT 0
		)`,
	}

	for _, query := range queries {
		_, err := db.Exec(ctx, query)
		require.NoError(t, err)
	}

	mockAuth := &services.MockAuthService{
		IsEnabledFunc: func() bool { return true },
		AuthenticateFunc: func(ctx context.Context, username, password string) (*models.JWTClaims, error) {
			return &models.JWTClaims{
				Username: "testuser",
				Role:     "admin",
				Name:     "Test User",
			}, nil
		},
	}

	mockPerm := &services.MockPermissionService{
		IsEnabledFunc: func() bool { return true },
		CheckPermissionFunc: func(ctx context.Context, username, permissionContext string, requiredLevel models.PermissionLevel) (bool, error) {
			return true, nil
		},
	}

	return NewHandler(db, mockAuth, mockPerm, "1.0.0-test")
}

// setupTestRequest creates a test request with authentication
func setupTestRequest(t *testing.T, handler *Handler, action string, data map[string]interface{}) (*httptest.ResponseRecorder, models.Response) {
	router := gin.New()

	// Add middleware to set username in context
	router.Use(func(c *gin.Context) {
		c.Set("username", "testuser")
		c.Next()
	})

	router.POST("/do", func(c *gin.Context) {
		var reqBody models.Request
		if err := c.ShouldBindJSON(&reqBody); err == nil {
			c.Set("request", &reqBody)
		}
		handler.DoAction(c)
	})

	reqBody := models.Request{
		Action: action,
		JWT:    "test-jwt-token",
		Data:   data,
	}
	body, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	var resp models.Response
	err = json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)

	return w, resp
}

// ========================================================================
// DOCUMENT SPACE TESTS
// ========================================================================

func TestDocumentSpaceCreate(t *testing.T) {
	handler := setupDocumentTestHandler(t)

	data := map[string]interface{}{
		"key":       "TEST",
		"name":      "Test Space",
		"owner_id":  "user-1",
		"is_public": true,
	}

	w, resp := setupTestRequest(t, handler, models.ActionDocumentSpaceCreate, data)

	// Document space create returns 201 Created
	assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusCreated)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.NotNil(t, resp.Data)

	// Response structure: data["space"] contains the space object
	space, ok := resp.Data["space"].(map[string]interface{})
	assert.True(t, ok, "Expected space in response data")
	assert.NotEmpty(t, space["id"])
}

func TestDocumentSpaceCreate_MissingRequired(t *testing.T) {
	handler := setupDocumentTestHandler(t)

	data := map[string]interface{}{
		"key": "TEST",
		// Missing name and owner_id
	}

	w, resp := setupTestRequest(t, handler, models.ActionDocumentSpaceCreate, data)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.NotEqual(t, models.ErrorCodeNoError, resp.ErrorCode)
}

func TestDocumentSpaceRead(t *testing.T) {
	handler := setupDocumentTestHandler(t)

	// Create space first
	createData := map[string]interface{}{
		"key":      "TEST",
		"name":     "Test Space",
		"owner_id": "user-1",
	}
	_, createResp := setupTestRequest(t, handler, models.ActionDocumentSpaceCreate, createData)
	createdSpace, ok := createResp.Data["space"].(map[string]interface{})
	require.True(t, ok, "Expected space in create response")
	spaceID := createdSpace["id"].(string)

	// Read space
	readData := map[string]interface{}{
		"space_id": spaceID,
	}
	w, resp := setupTestRequest(t, handler, models.ActionDocumentSpaceRead, readData)

	assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusCreated, "Expected 200 or 201, got %d", w.Code)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)

	space, ok := resp.Data["space"].(map[string]interface{})
	assert.True(t, ok, "Expected space in response data")
	assert.Equal(t, spaceID, space["id"])
	assert.Equal(t, "Test Space", space["name"])
}

func TestDocumentSpaceList(t *testing.T) {
	handler := setupDocumentTestHandler(t)

	// Create multiple spaces
	for i := 1; i <= 3; i++ {
		createData := map[string]interface{}{
			"key":      generateTestID(),
			"name":     "Test Space",
			"owner_id": "user-1",
		}
		setupTestRequest(t, handler, models.ActionDocumentSpaceCreate, createData)
	}

	// List spaces
	w, resp := setupTestRequest(t, handler, models.ActionDocumentSpaceList, nil)

	assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusCreated, "Expected 200 or 201, got %d", w.Code)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)

	spaces, ok := resp.Data["spaces"].([]interface{})
	assert.True(t, ok)
	assert.GreaterOrEqual(t, len(spaces), 3)
}

// ========================================================================
// DOCUMENT CRUD TESTS
// ========================================================================

func TestDocumentCreate(t *testing.T) {
	handler := setupDocumentTestHandler(t)

	// Create space first
	spaceData := map[string]interface{}{
		"key":      "TEST",
		"name":     "Test Space",
		"owner_id": "user-1",
	}
	_, spaceResp := setupTestRequest(t, handler, models.ActionDocumentSpaceCreate, spaceData)
	space, ok := spaceResp.Data["space"].(map[string]interface{})
	require.True(t, ok, "Expected space in response")
	spaceID := space["id"].(string)

	// Create document
	docData := map[string]interface{}{
		"title":    "Test Document",
		"space_id": spaceID,
		"type_id":  "type-page",
	}
	w, resp := setupTestRequest(t, handler, models.ActionDocumentCreate, docData)

	assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusCreated, "Expected 200 or 201, got %d", w.Code)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)

	document, ok := resp.Data["document"].(map[string]interface{})
	assert.True(t, ok, "Expected document in response data")
	assert.NotEmpty(t, document["id"])
	assert.Equal(t, "Test Document", document["title"])
}

func TestDocumentCreate_MissingRequired(t *testing.T) {
	handler := setupDocumentTestHandler(t)

	docData := map[string]interface{}{
		"title": "Test Document",
		// Missing space_id and type_id
	}
	w, resp := setupTestRequest(t, handler, models.ActionDocumentCreate, docData)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.NotEqual(t, models.ErrorCodeNoError, resp.ErrorCode)
}

func TestDocumentRead(t *testing.T) {
	handler := setupDocumentTestHandler(t)

	// Setup: Create space and document
	spaceData := map[string]interface{}{
		"key":      "TEST",
		"name":     "Test Space",
		"owner_id": "user-1",
	}
	_, spaceResp := setupTestRequest(t, handler, models.ActionDocumentSpaceCreate, spaceData)
	spaceID := spaceResp.Data["id"].(string)

	docData := map[string]interface{}{
		"title":    "Test Document",
		"space_id": spaceID,
		"type_id":  "type-page",
	}
	_, createResp := setupTestRequest(t, handler, models.ActionDocumentCreate, docData)
	docID := createResp.Data["id"].(string)

	// Read document
	readData := map[string]interface{}{
		"document_id": docID,
	}
	w, resp := setupTestRequest(t, handler, models.ActionDocumentRead, readData)

	assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusCreated, "Expected 200 or 201, got %d", w.Code)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.Equal(t, docID, resp.Data["id"])
	assert.Equal(t, "Test Document", resp.Data["title"])
}

func TestDocumentUpdate(t *testing.T) {
	handler := setupDocumentTestHandler(t)

	// Setup: Create space and document
	spaceData := map[string]interface{}{
		"key":      "TEST",
		"name":     "Test Space",
		"owner_id": "user-1",
	}
	_, spaceResp := setupTestRequest(t, handler, models.ActionDocumentSpaceCreate, spaceData)
	spaceID := spaceResp.Data["id"].(string)

	docData := map[string]interface{}{
		"title":    "Original Title",
		"space_id": spaceID,
		"type_id":  "type-page",
	}
	_, createResp := setupTestRequest(t, handler, models.ActionDocumentCreate, docData)
	docID := createResp.Data["id"].(string)

	// Update document
	updateData := map[string]interface{}{
		"document_id": docID,
		"title":       "Updated Title",
		"version":     1,
	}
	w, resp := setupTestRequest(t, handler, models.ActionDocumentUpdate, updateData)

	assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusCreated, "Expected 200 or 201, got %d", w.Code)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.Equal(t, "Updated Title", resp.Data["title"])
}

func TestDocumentDelete(t *testing.T) {
	handler := setupDocumentTestHandler(t)

	// Setup: Create space and document
	spaceData := map[string]interface{}{
		"key":      "TEST",
		"name":     "Test Space",
		"owner_id": "user-1",
	}
	_, spaceResp := setupTestRequest(t, handler, models.ActionDocumentSpaceCreate, spaceData)
	spaceID := spaceResp.Data["id"].(string)

	docData := map[string]interface{}{
		"title":    "Test Document",
		"space_id": spaceID,
		"type_id":  "type-page",
	}
	_, createResp := setupTestRequest(t, handler, models.ActionDocumentCreate, docData)
	docID := createResp.Data["id"].(string)

	// Delete document
	deleteData := map[string]interface{}{
		"document_id": docID,
	}
	w, resp := setupTestRequest(t, handler, models.ActionDocumentDelete, deleteData)

	assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusCreated, "Expected 200 or 201, got %d", w.Code)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)

	// Verify deletion - read should fail
	readData := map[string]interface{}{
		"document_id": docID,
	}
	_, readResp := setupTestRequest(t, handler, models.ActionDocumentRead, readData)
	assert.NotEqual(t, models.ErrorCodeNoError, readResp.ErrorCode)
}

func TestDocumentList(t *testing.T) {
	handler := setupDocumentTestHandler(t)

	// Setup: Create space
	spaceData := map[string]interface{}{
		"key":      "TEST",
		"name":     "Test Space",
		"owner_id": "user-1",
	}
	_, spaceResp := setupTestRequest(t, handler, models.ActionDocumentSpaceCreate, spaceData)
	spaceID := spaceResp.Data["id"].(string)

	// Create multiple documents
	for i := 1; i <= 5; i++ {
		docData := map[string]interface{}{
			"title":    "Test Document",
			"space_id": spaceID,
			"type_id":  "type-page",
		}
		setupTestRequest(t, handler, models.ActionDocumentCreate, docData)
	}

	// List documents
	listData := map[string]interface{}{
		"space_id": spaceID,
	}
	w, resp := setupTestRequest(t, handler, models.ActionDocumentList, listData)

	assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusCreated, "Expected 200 or 201, got %d", w.Code)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)

	documents, ok := resp.Data["documents"].([]interface{})
	assert.True(t, ok)
	assert.GreaterOrEqual(t, len(documents), 5)
}

// ========================================================================
// DOCUMENT CONTENT TESTS
// ========================================================================

func TestDocumentContentUpdate(t *testing.T) {
	handler := setupDocumentTestHandler(t)

	// Setup: Create space and document
	spaceData := map[string]interface{}{
		"key":      "TEST",
		"name":     "Test Space",
		"owner_id": "user-1",
	}
	_, spaceResp := setupTestRequest(t, handler, models.ActionDocumentSpaceCreate, spaceData)
	spaceID := spaceResp.Data["id"].(string)

	docData := map[string]interface{}{
		"title":    "Test Document",
		"space_id": spaceID,
		"type_id":  "type-page",
	}
	_, createResp := setupTestRequest(t, handler, models.ActionDocumentCreate, docData)
	docID := createResp.Data["id"].(string)

	// Update content
	contentData := map[string]interface{}{
		"document_id":  docID,
		"content":      "<p>Test content</p>",
		"content_type": "html",
	}
	w, resp := setupTestRequest(t, handler, models.ActionDocumentContentUpdate, contentData)

	assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusCreated, "Expected 200 or 201, got %d", w.Code)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
}

func TestDocumentContentGet(t *testing.T) {
	handler := setupDocumentTestHandler(t)

	// Setup: Create space, document, and content
	spaceData := map[string]interface{}{
		"key":      "TEST",
		"name":     "Test Space",
		"owner_id": "user-1",
	}
	_, spaceResp := setupTestRequest(t, handler, models.ActionDocumentSpaceCreate, spaceData)
	spaceID := spaceResp.Data["id"].(string)

	docData := map[string]interface{}{
		"title":    "Test Document",
		"space_id": spaceID,
		"type_id":  "type-page",
	}
	_, createResp := setupTestRequest(t, handler, models.ActionDocumentCreate, docData)
	docID := createResp.Data["id"].(string)

	contentData := map[string]interface{}{
		"document_id":  docID,
		"content":      "<p>Test content</p>",
		"content_type": "html",
	}
	setupTestRequest(t, handler, models.ActionDocumentContentUpdate, contentData)

	// Get content
	getData := map[string]interface{}{
		"document_id": docID,
	}
	w, resp := setupTestRequest(t, handler, models.ActionDocumentContentGet, getData)

	assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusCreated, "Expected 200 or 201, got %d", w.Code)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.Contains(t, resp.Data, "content")
}

// ========================================================================
// DOCUMENT ARCHIVE/PUBLISH TESTS
// ========================================================================

func TestDocumentArchive(t *testing.T) {
	handler := setupDocumentTestHandler(t)

	// Setup: Create space and document
	spaceData := map[string]interface{}{
		"key":      "TEST",
		"name":     "Test Space",
		"owner_id": "user-1",
	}
	_, spaceResp := setupTestRequest(t, handler, models.ActionDocumentSpaceCreate, spaceData)
	spaceID := spaceResp.Data["id"].(string)

	docData := map[string]interface{}{
		"title":    "Test Document",
		"space_id": spaceID,
		"type_id":  "type-page",
	}
	_, createResp := setupTestRequest(t, handler, models.ActionDocumentCreate, docData)
	docID := createResp.Data["id"].(string)

	// Archive document
	archiveData := map[string]interface{}{
		"document_id": docID,
	}
	w, resp := setupTestRequest(t, handler, models.ActionDocumentArchive, archiveData)

	assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusCreated, "Expected 200 or 201, got %d", w.Code)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
}

func TestDocumentPublish(t *testing.T) {
	handler := setupDocumentTestHandler(t)

	// Setup: Create space and document
	spaceData := map[string]interface{}{
		"key":      "TEST",
		"name":     "Test Space",
		"owner_id": "user-1",
	}
	_, spaceResp := setupTestRequest(t, handler, models.ActionDocumentSpaceCreate, spaceData)
	spaceID := spaceResp.Data["id"].(string)

	docData := map[string]interface{}{
		"title":    "Test Document",
		"space_id": spaceID,
		"type_id":  "type-page",
	}
	_, createResp := setupTestRequest(t, handler, models.ActionDocumentCreate, docData)
	docID := createResp.Data["id"].(string)

	// Publish document
	publishData := map[string]interface{}{
		"document_id": docID,
	}
	w, resp := setupTestRequest(t, handler, models.ActionDocumentPublish, publishData)

	assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusCreated, "Expected 200 or 201, got %d", w.Code)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
}

// ========================================================================
// AUTHENTICATION & AUTHORIZATION TESTS
// ========================================================================

func TestDocumentCreate_NoAuth(t *testing.T) {
	handler := setupDocumentTestHandler(t)
	router := gin.New()

	// No authentication middleware - should fail
	router.POST("/do", func(c *gin.Context) {
		var reqBody models.Request
		if err := c.ShouldBindJSON(&reqBody); err == nil {
			c.Set("request", &reqBody)
		}
		handler.DoAction(c)
	})

	reqBody := models.Request{
		Action: models.ActionDocumentCreate,
		Data: map[string]interface{}{
			"title":    "Test",
			"space_id": "space-1",
			"type_id":  "type-page",
		},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestDocumentCreate_NoPermission(t *testing.T) {
	db, err := database.NewDatabase(config.DatabaseConfig{
		Type:       "sqlite",
		SQLitePath: ":memory:",
	})
	require.NoError(t, err)

	mockAuth := &services.MockAuthService{
		IsEnabledFunc: func() bool { return true },
	}

	mockPerm := &services.MockPermissionService{
		IsEnabledFunc: func() bool { return true },
		CheckPermissionFunc: func(ctx context.Context, username, permissionContext string, requiredLevel models.PermissionLevel) (bool, error) {
			return false, nil // No permission
		},
	}

	handler := NewHandler(db, mockAuth, mockPerm, "1.0.0-test")
	router := gin.New()

	router.Use(func(c *gin.Context) {
		c.Set("username", "testuser")
		c.Next()
	})

	router.POST("/do", func(c *gin.Context) {
		var reqBody models.Request
		if err := c.ShouldBindJSON(&reqBody); err == nil {
			c.Set("request", &reqBody)
		}
		handler.DoAction(c)
	})

	reqBody := models.Request{
		Action: models.ActionDocumentCreate,
		JWT:    "test-jwt",
		Data: map[string]interface{}{
			"title":    "Test",
			"space_id": "space-1",
			"type_id":  "type-page",
		},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}
