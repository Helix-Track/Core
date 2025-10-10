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
	"helixtrack.ru/core/internal/middleware"
	"helixtrack.ru/core/internal/models"
)

// setupTicketRelationshipTestHandler creates test handler with relationship tables
func setupTicketRelationshipTestHandler(t *testing.T) *Handler {
	handler := setupTestHandler(t)

	// Create ticket_relationship_type table
	_, err := handler.db.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS ticket_relationship_type (
			id TEXT PRIMARY KEY,
			title TEXT NOT NULL,
			description TEXT,
			created INTEGER NOT NULL,
			modified INTEGER NOT NULL,
			deleted INTEGER DEFAULT 0
		)
	`)
	require.NoError(t, err)

	// Create ticket_relationship table
	_, err = handler.db.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS ticket_relationship (
			id TEXT PRIMARY KEY,
			ticket_id TEXT NOT NULL,
			child_ticket_id TEXT NOT NULL,
			ticket_relationship_type_id TEXT NOT NULL,
			created INTEGER NOT NULL,
			modified INTEGER NOT NULL,
			deleted INTEGER DEFAULT 0
		)
	`)
	require.NoError(t, err)

	return handler
}

// ====================
// Relationship Type CRUD Tests
// ====================

// TestTicketRelationshipHandler_TypeCreate_Success tests successful relationship type creation
func TestTicketRelationshipHandler_TypeCreate_Success(t *testing.T) {
	handler := setupTicketRelationshipTestHandler(t)
	gin.SetMode(gin.TestMode)

	reqBody := models.Request{
		Action: models.ActionTicketRelationshipTypeCreate,
		Data: map[string]interface{}{
			"title":       "Blocks",
			"description": "This ticket blocks another ticket",
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set(middleware.UsernameKey, "testuser")

	handler.handleTicketRelationshipTypeCreate(c, &reqBody)

	assert.Equal(t, http.StatusCreated, w.Code)

	// Verify database insertion
	var count int
	err := handler.db.QueryRow(context.Background(), "SELECT COUNT(*) FROM ticket_relationship_type WHERE title = ?", "Blocks").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

// TestTicketRelationshipHandler_TypeCreate_MissingTitle tests creation with missing title
func TestTicketRelationshipHandler_TypeCreate_MissingTitle(t *testing.T) {
	handler := setupTicketRelationshipTestHandler(t)
	gin.SetMode(gin.TestMode)

	reqBody := models.Request{
		Action: models.ActionTicketRelationshipTypeCreate,
		Data: map[string]interface{}{
			"description": "Missing title",
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set(middleware.UsernameKey, "testuser")

	handler.handleTicketRelationshipTypeCreate(c, &reqBody)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestTicketRelationshipHandler_TypeRead_Success tests reading a relationship type
func TestTicketRelationshipHandler_TypeRead_Success(t *testing.T) {
	handler := setupTicketRelationshipTestHandler(t)
	gin.SetMode(gin.TestMode)

	// Insert test relationship type
	typeID := "test-type-id"
	now := time.Now().Unix()
	_, err := handler.db.Exec(context.Background(), `
		INSERT INTO ticket_relationship_type (id, title, description, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?)
	`, typeID, "Blocks", "Blocking relationship", now, now, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionTicketRelationshipTypeRead,
		Data: map[string]interface{}{
			"id": typeID,
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set(middleware.UsernameKey, "testuser")

	handler.handleTicketRelationshipTypeRead(c, &reqBody)

	assert.Equal(t, http.StatusOK, w.Code)
}

// TestTicketRelationshipHandler_TypeRead_NotFound tests reading non-existent type
func TestTicketRelationshipHandler_TypeRead_NotFound(t *testing.T) {
	handler := setupTicketRelationshipTestHandler(t)
	gin.SetMode(gin.TestMode)

	reqBody := models.Request{
		Action: models.ActionTicketRelationshipTypeRead,
		Data: map[string]interface{}{
			"id": "non-existent-id",
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set(middleware.UsernameKey, "testuser")

	handler.handleTicketRelationshipTypeRead(c, &reqBody)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// TestTicketRelationshipHandler_TypeList_EmptyList tests listing when no types exist
func TestTicketRelationshipHandler_TypeList_EmptyList(t *testing.T) {
	handler := setupTicketRelationshipTestHandler(t)
	gin.SetMode(gin.TestMode)

	reqBody := models.Request{
		Action: models.ActionTicketRelationshipTypeList,
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set(middleware.UsernameKey, "testuser")

	handler.handleTicketRelationshipTypeList(c, &reqBody)

	assert.Equal(t, http.StatusOK, w.Code)
}

// TestTicketRelationshipHandler_TypeList_MultipleTypes tests listing multiple types
func TestTicketRelationshipHandler_TypeList_MultipleTypes(t *testing.T) {
	handler := setupTicketRelationshipTestHandler(t)
	gin.SetMode(gin.TestMode)

	// Insert test types
	now := time.Now().Unix()
	types := []string{"Blocks", "Depends on", "Relates to"}
	for _, title := range types {
		_, err := handler.db.Exec(context.Background(), `
			INSERT INTO ticket_relationship_type (id, title, created, modified, deleted)
			VALUES (?, ?, ?, ?, ?)
		`, generateTestID(), title, now, now, 0)
		require.NoError(t, err)
	}

	reqBody := models.Request{
		Action: models.ActionTicketRelationshipTypeList,
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set(middleware.UsernameKey, "testuser")

	handler.handleTicketRelationshipTypeList(c, &reqBody)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	responseData := response.Data.(map[string]interface{})
	count := int(responseData["count"].(float64))
	assert.Equal(t, 3, count)
}

// TestTicketRelationshipHandler_TypeModify_Success tests modifying a relationship type
func TestTicketRelationshipHandler_TypeModify_Success(t *testing.T) {
	handler := setupTicketRelationshipTestHandler(t)
	gin.SetMode(gin.TestMode)

	// Insert test type
	typeID := "test-type-id"
	now := time.Now().Unix()
	_, err := handler.db.Exec(context.Background(), `
		INSERT INTO ticket_relationship_type (id, title, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?)
	`, typeID, "Blocks", now, now, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionTicketRelationshipTypeModify,
		Data: map[string]interface{}{
			"id":          typeID,
			"title":       "Blocks (Updated)",
			"description": "Updated description",
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set(middleware.UsernameKey, "testuser")

	handler.handleTicketRelationshipTypeModify(c, &reqBody)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify update
	var title string
	err = handler.db.QueryRow(context.Background(), "SELECT title FROM ticket_relationship_type WHERE id = ?", typeID).Scan(&title)
	require.NoError(t, err)
	assert.Equal(t, "Blocks (Updated)", title)
}

// TestTicketRelationshipHandler_TypeModify_NotFound tests modifying non-existent type
func TestTicketRelationshipHandler_TypeModify_NotFound(t *testing.T) {
	handler := setupTicketRelationshipTestHandler(t)
	gin.SetMode(gin.TestMode)

	reqBody := models.Request{
		Action: models.ActionTicketRelationshipTypeModify,
		Data: map[string]interface{}{
			"id":    "non-existent-id",
			"title": "New Title",
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set(middleware.UsernameKey, "testuser")

	handler.handleTicketRelationshipTypeModify(c, &reqBody)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// TestTicketRelationshipHandler_TypeRemove_Success tests removing a relationship type
func TestTicketRelationshipHandler_TypeRemove_Success(t *testing.T) {
	handler := setupTicketRelationshipTestHandler(t)
	gin.SetMode(gin.TestMode)

	// Insert test type
	typeID := "test-type-id"
	now := time.Now().Unix()
	_, err := handler.db.Exec(context.Background(), `
		INSERT INTO ticket_relationship_type (id, title, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?)
	`, typeID, "Blocks", now, now, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionTicketRelationshipTypeRemove,
		Data: map[string]interface{}{
			"id": typeID,
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set(middleware.UsernameKey, "testuser")

	handler.handleTicketRelationshipTypeRemove(c, &reqBody)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify soft delete
	var deleted int
	err = handler.db.QueryRow(context.Background(), "SELECT deleted FROM ticket_relationship_type WHERE id = ?", typeID).Scan(&deleted)
	require.NoError(t, err)
	assert.Equal(t, 1, deleted)
}

// ====================
// Relationship Management Tests
// ====================

// TestTicketRelationshipHandler_Create_Success tests creating a ticket relationship
func TestTicketRelationshipHandler_Create_Success(t *testing.T) {
	handler := setupTicketRelationshipTestHandler(t)
	gin.SetMode(gin.TestMode)

	// Insert test relationship type
	typeID := "test-type-id"
	now := time.Now().Unix()
	_, err := handler.db.Exec(context.Background(), `
		INSERT INTO ticket_relationship_type (id, title, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?)
	`, typeID, "Blocks", now, now, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionTicketRelationshipCreate,
		Data: map[string]interface{}{
			"ticket_id":                      "ticket-1",
			"child_ticket_id":                "ticket-2",
			"ticket_relationship_type_id": typeID,
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set(middleware.UsernameKey, "testuser")

	handler.handleTicketRelationshipCreate(c, &reqBody)

	assert.Equal(t, http.StatusCreated, w.Code)

	// Verify database insertion
	var count int
	err = handler.db.QueryRow(context.Background(), "SELECT COUNT(*) FROM ticket_relationship WHERE ticket_id = ? AND child_ticket_id = ?", "ticket-1", "ticket-2").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

// TestTicketRelationshipHandler_Create_MissingTicketID tests creation with missing ticket ID
func TestTicketRelationshipHandler_Create_MissingTicketID(t *testing.T) {
	handler := setupTicketRelationshipTestHandler(t)
	gin.SetMode(gin.TestMode)

	reqBody := models.Request{
		Action: models.ActionTicketRelationshipCreate,
		Data: map[string]interface{}{
			"child_ticket_id":                "ticket-2",
			"ticket_relationship_type_id": "type-id",
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set(middleware.UsernameKey, "testuser")

	handler.handleTicketRelationshipCreate(c, &reqBody)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestTicketRelationshipHandler_Remove_Success tests removing a ticket relationship
func TestTicketRelationshipHandler_Remove_Success(t *testing.T) {
	handler := setupTicketRelationshipTestHandler(t)
	gin.SetMode(gin.TestMode)

	// Insert test relationship
	relationshipID := "test-rel-id"
	now := time.Now().Unix()
	_, err := handler.db.Exec(context.Background(), `
		INSERT INTO ticket_relationship (id, ticket_id, child_ticket_id, ticket_relationship_type_id, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, relationshipID, "ticket-1", "ticket-2", "type-id", now, now, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionTicketRelationshipRemove,
		Data: map[string]interface{}{
			"id": relationshipID,
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set(middleware.UsernameKey, "testuser")

	handler.handleTicketRelationshipRemove(c, &reqBody)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify soft delete
	var deleted int
	err = handler.db.QueryRow(context.Background(), "SELECT deleted FROM ticket_relationship WHERE id = ?", relationshipID).Scan(&deleted)
	require.NoError(t, err)
	assert.Equal(t, 1, deleted)
}

// TestTicketRelationshipHandler_Remove_NotFound tests removing non-existent relationship
func TestTicketRelationshipHandler_Remove_NotFound(t *testing.T) {
	handler := setupTicketRelationshipTestHandler(t)
	gin.SetMode(gin.TestMode)

	reqBody := models.Request{
		Action: models.ActionTicketRelationshipRemove,
		Data: map[string]interface{}{
			"id": "non-existent-id",
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set(middleware.UsernameKey, "testuser")

	handler.handleTicketRelationshipRemove(c, &reqBody)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// TestTicketRelationshipHandler_List_EmptyList tests listing when no relationships exist
func TestTicketRelationshipHandler_List_EmptyList(t *testing.T) {
	handler := setupTicketRelationshipTestHandler(t)
	gin.SetMode(gin.TestMode)

	reqBody := models.Request{
		Action: models.ActionTicketRelationshipList,
		Data: map[string]interface{}{
			"ticket_id": "ticket-1",
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set(middleware.UsernameKey, "testuser")

	handler.handleTicketRelationshipList(c, &reqBody)

	assert.Equal(t, http.StatusOK, w.Code)
}

// TestTicketRelationshipHandler_List_MultipleRelationships tests listing multiple relationships
func TestTicketRelationshipHandler_List_MultipleRelationships(t *testing.T) {
	handler := setupTicketRelationshipTestHandler(t)
	gin.SetMode(gin.TestMode)

	// Insert test relationships
	now := time.Now().Unix()
	_, err := handler.db.Exec(context.Background(), `
		INSERT INTO ticket_relationship (id, ticket_id, child_ticket_id, ticket_relationship_type_id, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, generateTestID(), "ticket-1", "ticket-2", "type-id", now, now, 0)
	require.NoError(t, err)

	_, err = handler.db.Exec(context.Background(), `
		INSERT INTO ticket_relationship (id, ticket_id, child_ticket_id, ticket_relationship_type_id, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, generateTestID(), "ticket-1", "ticket-3", "type-id", now, now, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionTicketRelationshipList,
		Data: map[string]interface{}{
			"ticket_id": "ticket-1",
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set(middleware.UsernameKey, "testuser")

	handler.handleTicketRelationshipList(c, &reqBody)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Response
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	responseData := response.Data.(map[string]interface{})
	count := int(responseData["count"].(float64))
	assert.Equal(t, 2, count)
}

// TestTicketRelationshipHandler_List_BidirectionalSearch tests bidirectional relationship search
func TestTicketRelationshipHandler_List_BidirectionalSearch(t *testing.T) {
	handler := setupTicketRelationshipTestHandler(t)
	gin.SetMode(gin.TestMode)

	// Insert relationships where ticket-1 is both parent and child
	now := time.Now().Unix()
	_, err := handler.db.Exec(context.Background(), `
		INSERT INTO ticket_relationship (id, ticket_id, child_ticket_id, ticket_relationship_type_id, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, generateTestID(), "ticket-1", "ticket-2", "type-id", now, now, 0)
	require.NoError(t, err)

	_, err = handler.db.Exec(context.Background(), `
		INSERT INTO ticket_relationship (id, ticket_id, child_ticket_id, ticket_relationship_type_id, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, generateTestID(), "ticket-3", "ticket-1", "type-id", now, now, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionTicketRelationshipList,
		Data: map[string]interface{}{
			"ticket_id": "ticket-1",
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set(middleware.UsernameKey, "testuser")

	handler.handleTicketRelationshipList(c, &reqBody)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Response
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	responseData := response.Data.(map[string]interface{})
	count := int(responseData["count"].(float64))
	assert.Equal(t, 2, count) // Both relationships where ticket-1 is involved
}
