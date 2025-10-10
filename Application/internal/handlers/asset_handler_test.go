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
	"helixtrack.ru/core/internal/database"
	"helixtrack.ru/core/internal/models"
)

// setupAssetTestHandler creates test handler with asset and mapping tables
func setupAssetTestHandler(t *testing.T) *Handler {
	handler := setupTestHandler(t)

	// Create asset table
	_, err := handler.db.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS asset (
			id TEXT PRIMARY KEY,
			url TEXT NOT NULL,
			description TEXT,
			created INTEGER NOT NULL,
			modified INTEGER NOT NULL,
			deleted INTEGER DEFAULT 0
		)
	`)
	require.NoError(t, err)

	// Create asset_ticket_mapping table
	_, err = handler.db.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS asset_ticket_mapping (
			id TEXT PRIMARY KEY,
			asset_id TEXT NOT NULL,
			ticket_id TEXT NOT NULL,
			created INTEGER NOT NULL,
			modified INTEGER NOT NULL,
			deleted INTEGER DEFAULT 0
		)
	`)
	require.NoError(t, err)

	// Create asset_comment_mapping table
	_, err = handler.db.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS asset_comment_mapping (
			id TEXT PRIMARY KEY,
			asset_id TEXT NOT NULL,
			comment_id TEXT NOT NULL,
			created INTEGER NOT NULL,
			modified INTEGER NOT NULL,
			deleted INTEGER DEFAULT 0
		)
	`)
	require.NoError(t, err)

	// Create asset_project_mapping table
	_, err = handler.db.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS asset_project_mapping (
			id TEXT PRIMARY KEY,
			asset_id TEXT NOT NULL,
			project_id TEXT NOT NULL,
			created INTEGER NOT NULL,
			modified INTEGER NOT NULL,
			deleted INTEGER DEFAULT 0
		)
	`)
	require.NoError(t, err)

	return handler
}

// generateAssetID generates a unique asset ID for testing
func generateAssetID() string {
	return generateTestID()
}

// TestAssetHandler_Create_Success tests creating an asset with full fields
func TestAssetHandler_Create_Success(t *testing.T) {
	handler := setupAssetTestHandler(t)
	gin.SetMode(gin.TestMode)

	reqBody := models.Request{
		Action: models.ActionAssetCreate,
		Data: map[string]interface{}{
			"url":         "https://example.com/file.pdf",
			"description": "Important document",
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")

	handler.handleAssetCreate(c, &reqBody)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response models.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeSuccess, response.ErrorCode)

	dataMap := response.Data.(map[string]interface{})
	asset := dataMap["asset"].(map[string]interface{})
	assert.NotEmpty(t, asset["id"])
	assert.Equal(t, "https://example.com/file.pdf", asset["url"])
	assert.Equal(t, "Important document", asset["description"])
}

// TestAssetHandler_Create_MinimalFields tests creating an asset with only required fields
func TestAssetHandler_Create_MinimalFields(t *testing.T) {
	handler := setupAssetTestHandler(t)
	gin.SetMode(gin.TestMode)

	reqBody := models.Request{
		Action: models.ActionAssetCreate,
		Data: map[string]interface{}{
			"url": "https://example.com/image.png",
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")

	handler.handleAssetCreate(c, &reqBody)

	assert.Equal(t, http.StatusCreated, w.Code)
}

// TestAssetHandler_Create_MissingURL tests creating without required URL field
func TestAssetHandler_Create_MissingURL(t *testing.T) {
	handler := setupAssetTestHandler(t)
	gin.SetMode(gin.TestMode)

	reqBody := models.Request{
		Action: models.ActionAssetCreate,
		Data: map[string]interface{}{
			"description": "Missing URL",
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")

	handler.handleAssetCreate(c, &reqBody)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestAssetHandler_Create_Unauthorized tests creating without authentication
func TestAssetHandler_Create_Unauthorized(t *testing.T) {
	handler := setupAssetTestHandler(t)
	gin.SetMode(gin.TestMode)

	reqBody := models.Request{
		Action: models.ActionAssetCreate,
		Data: map[string]interface{}{
			"url": "https://example.com/file.pdf",
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req

	handler.handleAssetCreate(c, &reqBody)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestAssetHandler_Read_Success tests reading an existing asset
func TestAssetHandler_Read_Success(t *testing.T) {
	handler := setupAssetTestHandler(t)
	gin.SetMode(gin.TestMode)

	// Insert test asset
	assetID := generateAssetID()
	now := time.Now().Unix()
	_, err := handler.db.Exec(context.Background(),
		`INSERT INTO asset (id, url, description, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)`,
		assetID, "https://example.com/test.pdf", "Test Asset", now, now, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionAssetRead,
		Data: map[string]interface{}{
			"id": assetID,
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")

	handler.handleAssetRead(c, &reqBody)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Response
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	dataMap := response.Data.(map[string]interface{})
	asset := dataMap["asset"].(map[string]interface{})
	assert.Equal(t, assetID, asset["id"])
	assert.Equal(t, "https://example.com/test.pdf", asset["url"])
}

// TestAssetHandler_Read_NotFound tests reading non-existent asset
func TestAssetHandler_Read_NotFound(t *testing.T) {
	handler := setupAssetTestHandler(t)
	gin.SetMode(gin.TestMode)

	reqBody := models.Request{
		Action: models.ActionAssetRead,
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

	handler.handleAssetRead(c, &reqBody)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// TestAssetHandler_Read_MissingID tests reading without asset ID
func TestAssetHandler_Read_MissingID(t *testing.T) {
	handler := setupAssetTestHandler(t)
	gin.SetMode(gin.TestMode)

	reqBody := models.Request{
		Action: models.ActionAssetRead,
		Data:   map[string]interface{}{},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")

	handler.handleAssetRead(c, &reqBody)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestAssetHandler_List_EmptyList tests listing when no assets exist
func TestAssetHandler_List_EmptyList(t *testing.T) {
	handler := setupAssetTestHandler(t)
	gin.SetMode(gin.TestMode)

	reqBody := models.Request{
		Action: models.ActionAssetList,
		Data:   map[string]interface{}{},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")

	handler.handleAssetList(c, &reqBody)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	dataMap := response.Data.(map[string]interface{})
	count := int(dataMap["count"].(float64))
	assert.Equal(t, 0, count)
}

// TestAssetHandler_List_MultipleAssets tests listing multiple assets
func TestAssetHandler_List_MultipleAssets(t *testing.T) {
	handler := setupAssetTestHandler(t)
	gin.SetMode(gin.TestMode)

	// Insert test assets
	now := time.Now().Unix()
	for i := 0; i < 3; i++ {
		assetID := generateAssetID()
		_, err := handler.db.Exec(context.Background(),
			`INSERT INTO asset (id, url, description, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)`,
			assetID, "https://example.com/file"+string(rune('1'+i))+".pdf", "Asset "+string(rune('A'+i)), now, now, 0)
		require.NoError(t, err)
	}

	reqBody := models.Request{
		Action: models.ActionAssetList,
		Data:   map[string]interface{}{},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")

	handler.handleAssetList(c, &reqBody)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	dataMap := response.Data.(map[string]interface{})
	count := int(dataMap["count"].(float64))
	assert.Equal(t, 3, count)
}

// TestAssetHandler_List_ExcludesDeleted tests that deleted assets are excluded
func TestAssetHandler_List_ExcludesDeleted(t *testing.T) {
	handler := setupAssetTestHandler(t)
	gin.SetMode(gin.TestMode)

	// Insert active and deleted assets
	now := time.Now().Unix()
	activeID := generateAssetID()
	deletedID := generateAssetID()

	_, err := handler.db.Exec(context.Background(),
		`INSERT INTO asset (id, url, description, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)`,
		activeID, "https://example.com/active.pdf", "Active", now, now, 0)
	require.NoError(t, err)

	_, err = handler.db.Exec(context.Background(),
		`INSERT INTO asset (id, url, description, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)`,
		deletedID, "https://example.com/deleted.pdf", "Deleted", now, now, 1)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionAssetList,
		Data:   map[string]interface{}{},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")

	handler.handleAssetList(c, &reqBody)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Response
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	dataMap := response.Data.(map[string]interface{})
	count := int(dataMap["count"].(float64))
	assert.Equal(t, 1, count)
}

// TestAssetHandler_Modify_Success tests modifying an asset
func TestAssetHandler_Modify_Success(t *testing.T) {
	handler := setupAssetTestHandler(t)
	gin.SetMode(gin.TestMode)

	// Insert test asset
	assetID := generateAssetID()
	now := time.Now().Unix()
	_, err := handler.db.Exec(context.Background(),
		`INSERT INTO asset (id, url, description, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)`,
		assetID, "https://example.com/old.pdf", "Old Description", now, now, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionAssetModify,
		Data: map[string]interface{}{
			"id":          assetID,
			"url":         "https://example.com/new.pdf",
			"description": "New Description",
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")

	handler.handleAssetModify(c, &reqBody)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify changes in database
	var url, description string
	err = handler.db.QueryRow(context.Background(), "SELECT url, description FROM asset WHERE id = ?", assetID).Scan(&url, &description)
	require.NoError(t, err)
	assert.Equal(t, "https://example.com/new.pdf", url)
	assert.Equal(t, "New Description", description)
}

// TestAssetHandler_Modify_PartialUpdate tests updating only some fields
func TestAssetHandler_Modify_PartialUpdate(t *testing.T) {
	handler := setupAssetTestHandler(t)
	gin.SetMode(gin.TestMode)

	// Insert test asset
	assetID := generateAssetID()
	now := time.Now().Unix()
	_, err := handler.db.Exec(context.Background(),
		`INSERT INTO asset (id, url, description, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)`,
		assetID, "https://example.com/file.pdf", "Original", now, now, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionAssetModify,
		Data: map[string]interface{}{
			"id":          assetID,
			"description": "Updated Description",
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")

	handler.handleAssetModify(c, &reqBody)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify URL unchanged, description updated
	var url, description string
	err = handler.db.QueryRow(context.Background(), "SELECT url, description FROM asset WHERE id = ?", assetID).Scan(&url, &description)
	require.NoError(t, err)
	assert.Equal(t, "https://example.com/file.pdf", url)
	assert.Equal(t, "Updated Description", description)
}

// TestAssetHandler_Modify_NotFound tests modifying non-existent asset
func TestAssetHandler_Modify_NotFound(t *testing.T) {
	handler := setupAssetTestHandler(t)
	gin.SetMode(gin.TestMode)

	reqBody := models.Request{
		Action: models.ActionAssetModify,
		Data: map[string]interface{}{
			"id":  "non-existent-id",
			"url": "https://example.com/new.pdf",
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")

	handler.handleAssetModify(c, &reqBody)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// TestAssetHandler_Modify_NoFields tests modifying without any fields
func TestAssetHandler_Modify_NoFields(t *testing.T) {
	handler := setupAssetTestHandler(t)
	gin.SetMode(gin.TestMode)

	// Insert test asset
	assetID := generateAssetID()
	now := time.Now().Unix()
	_, err := handler.db.Exec(context.Background(),
		`INSERT INTO asset (id, url, description, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)`,
		assetID, "https://example.com/file.pdf", "Test", now, now, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionAssetModify,
		Data: map[string]interface{}{
			"id": assetID,
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")

	handler.handleAssetModify(c, &reqBody)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestAssetHandler_Remove_Success tests soft-deleting an asset
func TestAssetHandler_Remove_Success(t *testing.T) {
	handler := setupAssetTestHandler(t)
	gin.SetMode(gin.TestMode)

	// Insert test asset
	assetID := generateAssetID()
	now := time.Now().Unix()
	_, err := handler.db.Exec(context.Background(),
		`INSERT INTO asset (id, url, description, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)`,
		assetID, "https://example.com/file.pdf", "Test", now, now, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionAssetRemove,
		Data: map[string]interface{}{
			"id": assetID,
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")

	handler.handleAssetRemove(c, &reqBody)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify soft delete
	var deleted int
	err = handler.db.QueryRow(context.Background(), "SELECT deleted FROM asset WHERE id = ?", assetID).Scan(&deleted)
	require.NoError(t, err)
	assert.Equal(t, 1, deleted)
}

// TestAssetHandler_Remove_NotFound tests deleting non-existent asset
func TestAssetHandler_Remove_NotFound(t *testing.T) {
	handler := setupAssetTestHandler(t)
	gin.SetMode(gin.TestMode)

	reqBody := models.Request{
		Action: models.ActionAssetRemove,
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

	handler.handleAssetRemove(c, &reqBody)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// TestAssetHandler_AddTicket_Success tests adding asset to ticket
func TestAssetHandler_AddTicket_Success(t *testing.T) {
	handler := setupAssetTestHandler(t)
	gin.SetMode(gin.TestMode)

	assetID := generateAssetID()
	ticketID := generateTestID()

	reqBody := models.Request{
		Action: models.ActionAssetAddTicket,
		Data: map[string]interface{}{
			"assetId":  assetID,
			"ticketId": ticketID,
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")

	handler.handleAssetAddTicket(c, &reqBody)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify mapping created
	var count int
	err := handler.db.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM asset_ticket_mapping WHERE asset_id = ? AND ticket_id = ? AND deleted = 0",
		assetID, ticketID).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

// TestAssetHandler_RemoveTicket_Success tests removing asset from ticket
func TestAssetHandler_RemoveTicket_Success(t *testing.T) {
	handler := setupAssetTestHandler(t)
	gin.SetMode(gin.TestMode)

	assetID := generateAssetID()
	ticketID := generateTestID()

	// Create mapping
	mappingID := generateTestID()
	now := time.Now().Unix()
	_, err := handler.db.Exec(context.Background(),
		`INSERT INTO asset_ticket_mapping (id, asset_id, ticket_id, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)`,
		mappingID, assetID, ticketID, now, now, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionAssetRemoveTicket,
		Data: map[string]interface{}{
			"assetId":  assetID,
			"ticketId": ticketID,
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")

	handler.handleAssetRemoveTicket(c, &reqBody)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify mapping soft-deleted
	var deleted int
	err = handler.db.QueryRow(context.Background(),
		"SELECT deleted FROM asset_ticket_mapping WHERE asset_id = ? AND ticket_id = ?",
		assetID, ticketID).Scan(&deleted)
	require.NoError(t, err)
	assert.Equal(t, 1, deleted)
}

// TestAssetHandler_ListTickets_Success tests listing tickets for asset
func TestAssetHandler_ListTickets_Success(t *testing.T) {
	handler := setupAssetTestHandler(t)
	gin.SetMode(gin.TestMode)

	assetID := generateAssetID()
	now := time.Now().Unix()

	// Create multiple ticket mappings
	for i := 0; i < 3; i++ {
		ticketID := generateTestID()
		mappingID := generateTestID()
		_, err := handler.db.Exec(context.Background(),
			`INSERT INTO asset_ticket_mapping (id, asset_id, ticket_id, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)`,
			mappingID, assetID, ticketID, now, now, 0)
		require.NoError(t, err)
	}

	reqBody := models.Request{
		Action: models.ActionAssetListTickets,
		Data: map[string]interface{}{
			"assetId": assetID,
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")

	handler.handleAssetListTickets(c, &reqBody)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	dataMap := response.Data.(map[string]interface{})
	count := int(dataMap["count"].(float64))
	assert.Equal(t, 3, count)
}

// TestAssetHandler_AddComment_Success tests adding asset to comment
func TestAssetHandler_AddComment_Success(t *testing.T) {
	handler := setupAssetTestHandler(t)
	gin.SetMode(gin.TestMode)

	assetID := generateAssetID()
	commentID := generateTestID()

	reqBody := models.Request{
		Action: models.ActionAssetAddComment,
		Data: map[string]interface{}{
			"assetId":   assetID,
			"commentId": commentID,
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")

	handler.handleAssetAddComment(c, &reqBody)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify mapping created
	var count int
	err := handler.db.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM asset_comment_mapping WHERE asset_id = ? AND comment_id = ? AND deleted = 0",
		assetID, commentID).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

// TestAssetHandler_RemoveComment_Success tests removing asset from comment
func TestAssetHandler_RemoveComment_Success(t *testing.T) {
	handler := setupAssetTestHandler(t)
	gin.SetMode(gin.TestMode)

	assetID := generateAssetID()
	commentID := generateTestID()

	// Create mapping
	mappingID := generateTestID()
	now := time.Now().Unix()
	_, err := handler.db.Exec(context.Background(),
		`INSERT INTO asset_comment_mapping (id, asset_id, comment_id, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)`,
		mappingID, assetID, commentID, now, now, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionAssetRemoveComment,
		Data: map[string]interface{}{
			"assetId":   assetID,
			"commentId": commentID,
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")

	handler.handleAssetRemoveComment(c, &reqBody)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify mapping soft-deleted
	var deleted int
	err = handler.db.QueryRow(context.Background(),
		"SELECT deleted FROM asset_comment_mapping WHERE asset_id = ? AND comment_id = ?",
		assetID, commentID).Scan(&deleted)
	require.NoError(t, err)
	assert.Equal(t, 1, deleted)
}

// TestAssetHandler_ListComments_Success tests listing comments for asset
func TestAssetHandler_ListComments_Success(t *testing.T) {
	handler := setupAssetTestHandler(t)
	gin.SetMode(gin.TestMode)

	assetID := generateAssetID()
	now := time.Now().Unix()

	// Create multiple comment mappings
	for i := 0; i < 2; i++ {
		commentID := generateTestID()
		mappingID := generateTestID()
		_, err := handler.db.Exec(context.Background(),
			`INSERT INTO asset_comment_mapping (id, asset_id, comment_id, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)`,
			mappingID, assetID, commentID, now, now, 0)
		require.NoError(t, err)
	}

	reqBody := models.Request{
		Action: models.ActionAssetListComments,
		Data: map[string]interface{}{
			"assetId": assetID,
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")

	handler.handleAssetListComments(c, &reqBody)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	dataMap := response.Data.(map[string]interface{})
	count := int(dataMap["count"].(float64))
	assert.Equal(t, 2, count)
}

// TestAssetHandler_AddProject_Success tests adding asset to project
func TestAssetHandler_AddProject_Success(t *testing.T) {
	handler := setupAssetTestHandler(t)
	gin.SetMode(gin.TestMode)

	assetID := generateAssetID()
	projectID := generateTestID()

	reqBody := models.Request{
		Action: models.ActionAssetAddProject,
		Data: map[string]interface{}{
			"assetId":   assetID,
			"projectId": projectID,
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")

	handler.handleAssetAddProject(c, &reqBody)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify mapping created
	var count int
	err := handler.db.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM asset_project_mapping WHERE asset_id = ? AND project_id = ? AND deleted = 0",
		assetID, projectID).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

// TestAssetHandler_RemoveProject_Success tests removing asset from project
func TestAssetHandler_RemoveProject_Success(t *testing.T) {
	handler := setupAssetTestHandler(t)
	gin.SetMode(gin.TestMode)

	assetID := generateAssetID()
	projectID := generateTestID()

	// Create mapping
	mappingID := generateTestID()
	now := time.Now().Unix()
	_, err := handler.db.Exec(context.Background(),
		`INSERT INTO asset_project_mapping (id, asset_id, project_id, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)`,
		mappingID, assetID, projectID, now, now, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionAssetRemoveProject,
		Data: map[string]interface{}{
			"assetId":   assetID,
			"projectId": projectID,
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")

	handler.handleAssetRemoveProject(c, &reqBody)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify mapping soft-deleted
	var deleted int
	err = handler.db.QueryRow(context.Background(),
		"SELECT deleted FROM asset_project_mapping WHERE asset_id = ? AND project_id = ?",
		assetID, projectID).Scan(&deleted)
	require.NoError(t, err)
	assert.Equal(t, 1, deleted)
}

// TestAssetHandler_ListProjects_Success tests listing projects for asset
func TestAssetHandler_ListProjects_Success(t *testing.T) {
	handler := setupAssetTestHandler(t)
	gin.SetMode(gin.TestMode)

	assetID := generateAssetID()
	now := time.Now().Unix()

	// Create multiple project mappings
	for i := 0; i < 4; i++ {
		projectID := generateTestID()
		mappingID := generateTestID()
		_, err := handler.db.Exec(context.Background(),
			`INSERT INTO asset_project_mapping (id, asset_id, project_id, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)`,
			mappingID, assetID, projectID, now, now, 0)
		require.NoError(t, err)
	}

	reqBody := models.Request{
		Action: models.ActionAssetListProjects,
		Data: map[string]interface{}{
			"assetId": assetID,
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")

	handler.handleAssetListProjects(c, &reqBody)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	dataMap := response.Data.(map[string]interface{})
	count := int(dataMap["count"].(float64))
	assert.Equal(t, 4, count)
}

// TestAssetHandler_FullCRUDCycle tests complete asset lifecycle
func TestAssetHandler_FullCRUDCycle(t *testing.T) {
	handler := setupAssetTestHandler(t)
	gin.SetMode(gin.TestMode)

	// 1. Create
	reqBody := models.Request{
		Action: models.ActionAssetCreate,
		Data: map[string]interface{}{
			"url":         "https://example.com/lifecycle.pdf",
			"description": "Lifecycle Test",
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")

	handler.handleAssetCreate(c, &reqBody)
	assert.Equal(t, http.StatusCreated, w.Code)

	var createResponse models.Response
	json.Unmarshal(w.Body.Bytes(), &createResponse)
	dataMap := createResponse.Data.(map[string]interface{})
	asset := dataMap["asset"].(map[string]interface{})
	assetID := asset["id"].(string)

	// 2. Read
	reqBody.Action = models.ActionAssetRead
	reqBody.Data = map[string]interface{}{"id": assetID}
	body, _ = json.Marshal(reqBody)
	req = httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")

	handler.handleAssetRead(c, &reqBody)
	assert.Equal(t, http.StatusOK, w.Code)

	// 3. Modify
	reqBody.Action = models.ActionAssetModify
	reqBody.Data = map[string]interface{}{
		"id":          assetID,
		"description": "Modified Lifecycle",
	}
	body, _ = json.Marshal(reqBody)
	req = httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")

	handler.handleAssetModify(c, &reqBody)
	assert.Equal(t, http.StatusOK, w.Code)

	// 4. Remove
	reqBody.Action = models.ActionAssetRemove
	reqBody.Data = map[string]interface{}{"id": assetID}
	body, _ = json.Marshal(reqBody)
	req = httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")

	handler.handleAssetRemove(c, &reqBody)
	assert.Equal(t, http.StatusOK, w.Code)

	// 5. Verify deleted
	var deleted int
	err := handler.db.QueryRow(context.Background(), "SELECT deleted FROM asset WHERE id = ?", assetID).Scan(&deleted)
	require.NoError(t, err)
	assert.Equal(t, 1, deleted)
}
