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
	"helixtrack.ru/core/internal/models"
)

// setupComponentTestHandler creates a test handler with component table and dependencies
func setupComponentTestHandler(t *testing.T) *Handler {
	handler := setupTestHandler(t)

	// Create component table
	_, err := handler.db.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS component (
			id TEXT PRIMARY KEY,
			title TEXT NOT NULL,
			description TEXT,
			created INTEGER NOT NULL,
			modified INTEGER NOT NULL,
			deleted INTEGER NOT NULL DEFAULT 0
		)
	`)
	require.NoError(t, err)

	// Create component_ticket_mapping table
	_, err = handler.db.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS component_ticket_mapping (
			id TEXT PRIMARY KEY,
			component_id TEXT NOT NULL,
			ticket_id TEXT NOT NULL,
			created INTEGER NOT NULL,
			modified INTEGER NOT NULL,
			deleted INTEGER NOT NULL DEFAULT 0
		)
	`)
	require.NoError(t, err)

	// Create component_meta_data table
	_, err = handler.db.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS component_meta_data (
			id TEXT PRIMARY KEY,
			component_id TEXT NOT NULL,
			property TEXT NOT NULL,
			value TEXT,
			created INTEGER NOT NULL,
			modified INTEGER NOT NULL,
			deleted INTEGER NOT NULL DEFAULT 0
		)
	`)
	require.NoError(t, err)

	// Create ticket table for mapping tests
	_, err = handler.db.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS ticket (
			id TEXT PRIMARY KEY,
			title TEXT NOT NULL,
			description TEXT,
			status TEXT NOT NULL,
			created INTEGER NOT NULL,
			modified INTEGER NOT NULL,
			deleted INTEGER NOT NULL DEFAULT 0
		)
	`)
	require.NoError(t, err)

	// Insert test ticket for mapping tests
	_, err = handler.db.Exec(context.Background(),
		"INSERT INTO ticket (id, title, description, status, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?)",
		"test-ticket-id", "Test Ticket", "Test ticket description", "open", 1000, 1000, 0)
	require.NoError(t, err)

	return handler
}

// performRequest performs an HTTP request to the handler
func performRequest(handler *Handler, method, path string, reqBody models.Request) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(method, path, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")
	c.Set("request", &reqBody)

	handler.DoAction(c)

	return w
}

// ============================================================================
// Component CRUD Tests
// ============================================================================

func TestComponentHandler_Create_Success(t *testing.T) {
	handler := setupComponentTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionComponentCreate,
		Data: map[string]interface{}{
			"title":       "Database",
			"description": "Database-related components",
		},
	}

	w := performRequest(handler, "POST", "/do", reqBody)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response models.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, models.ErrorCodeNoError, response.ErrorCode)

	componentData, ok := response.Data["component"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "Database", componentData["title"])
	assert.Equal(t, "Database-related components", componentData["description"])
	assert.NotEmpty(t, componentData["id"])

	// Verify in database
	var count int
	err = handler.db.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM component WHERE title = ?", "Database").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestComponentHandler_Create_MinimalFields(t *testing.T) {
	handler := setupComponentTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionComponentCreate,
		Data: map[string]interface{}{
			"title": "API",
		},
	}

	w := performRequest(handler, "POST", "/do", reqBody)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response models.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, models.ErrorCodeNoError, response.ErrorCode)

	componentData, ok := response.Data["component"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "API", componentData["title"])
	assert.Empty(t, componentData["description"])
}

func TestComponentHandler_Create_MissingTitle(t *testing.T) {
	handler := setupComponentTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionComponentCreate,
		Data: map[string]interface{}{
			"description": "No title provided",
		},
	}

	w := performRequest(handler, "POST", "/do", reqBody)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response models.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, models.ErrorCodeMissingData, response.ErrorCode)
	assert.Contains(t, response.ErrorMessage, "Missing title")
}

func TestComponentHandler_Create_MultipleCommonComponents(t *testing.T) {
	handler := setupComponentTestHandler(t)

	components := []struct {
		title       string
		description string
	}{
		{"Frontend", "Frontend application components"},
		{"Backend", "Backend API components"},
		{"Database", "Database schema and migrations"},
		{"Authentication", "Auth and security components"},
		{"Reporting", "Reports and analytics"},
	}

	for _, comp := range components {
		reqBody := models.Request{
			Action: models.ActionComponentCreate,
			Data: map[string]interface{}{
				"title":       comp.title,
				"description": comp.description,
			},
		}

		w := performRequest(handler, "POST", "/do", reqBody)
		assert.Equal(t, http.StatusCreated, w.Code)
	}

	// Verify all components were created
	var count int
	err := handler.db.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM component WHERE deleted = 0").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, len(components), count)
}

func TestComponentHandler_Read_Success(t *testing.T) {
	handler := setupComponentTestHandler(t)

	// Insert test component
	componentID := "test-component-id"
	_, err := handler.db.Exec(context.Background(),
		"INSERT INTO component (id, title, description, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)",
		componentID, "Frontend", "Frontend components", 1000, 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionComponentRead,
		Data: map[string]interface{}{
			"id": componentID,
		},
	}

	w := performRequest(handler, "POST", "/do", reqBody)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Response
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, models.ErrorCodeNoError, response.ErrorCode)

	componentData, ok := response.Data["component"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, componentID, componentData["id"])
	assert.Equal(t, "Frontend", componentData["title"])
}

func TestComponentHandler_Read_NotFound(t *testing.T) {
	handler := setupComponentTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionComponentRead,
		Data: map[string]interface{}{
			"id": "non-existent-id",
		},
	}

	w := performRequest(handler, "POST", "/do", reqBody)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response models.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, models.ErrorCodeEntityNotFound, response.ErrorCode)
	assert.Contains(t, response.ErrorMessage, "Component not found")
}

func TestComponentHandler_List_Empty(t *testing.T) {
	handler := setupComponentTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionComponentList,
		Data:   map[string]interface{}{},
	}

	w := performRequest(handler, "POST", "/do", reqBody)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, models.ErrorCodeNoError, response.ErrorCode)

	components, ok := response.Data["components"].([]interface{})
	require.True(t, ok)
	assert.Empty(t, components)

	count, ok := response.Data["count"].(float64)
	require.True(t, ok)
	assert.Equal(t, float64(0), count)
}

func TestComponentHandler_List_Multiple(t *testing.T) {
	handler := setupComponentTestHandler(t)

	// Insert multiple components
	components := []struct {
		id    string
		title string
	}{
		{"comp-1", "Backend"},
		{"comp-2", "Frontend"},
		{"comp-3", "Database"},
	}

	for _, comp := range components {
		_, err := handler.db.Exec(context.Background(),
			"INSERT INTO component (id, title, description, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)",
			comp.id, comp.title, "Description", 1000, 1000, 0)
		require.NoError(t, err)
	}

	reqBody := models.Request{
		Action: models.ActionComponentList,
		Data:   map[string]interface{}{},
	}

	w := performRequest(handler, "POST", "/do", reqBody)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, models.ErrorCodeNoError, response.ErrorCode)

	componentsList, ok := response.Data["components"].([]interface{})
	require.True(t, ok)
	assert.Len(t, componentsList, 3)

	count, ok := response.Data["count"].(float64)
	require.True(t, ok)
	assert.Equal(t, float64(3), count)
}

func TestComponentHandler_List_OrderedByTitle(t *testing.T) {
	handler := setupComponentTestHandler(t)

	// Insert components in random order
	components := []string{"Zebra", "Apple", "Mango"}
	for i, title := range components {
		_, err := handler.db.Exec(context.Background(),
			"INSERT INTO component (id, title, description, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)",
			"comp-"+string(rune('1'+i)), title, "Description", 1000, 1000, 0)
		require.NoError(t, err)
	}

	reqBody := models.Request{
		Action: models.ActionComponentList,
		Data:   map[string]interface{}{},
	}

	w := performRequest(handler, "POST", "/do", reqBody)

	var response models.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	componentsList, ok := response.Data["components"].([]interface{})
	require.True(t, ok)

	// Verify alphabetical order
	titles := make([]string, len(componentsList))
	for i, comp := range componentsList {
		compMap := comp.(map[string]interface{})
		titles[i] = compMap["title"].(string)
	}

	assert.Equal(t, "Apple", titles[0])
	assert.Equal(t, "Mango", titles[1])
	assert.Equal(t, "Zebra", titles[2])
}

func TestComponentHandler_List_ExcludesDeleted(t *testing.T) {
	handler := setupComponentTestHandler(t)

	// Insert active and deleted components
	_, err := handler.db.Exec(context.Background(),
		"INSERT INTO component (id, title, description, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)",
		"active-1", "Active", "Active component", 1000, 1000, 0)
	require.NoError(t, err)

	_, err = handler.db.Exec(context.Background(),
		"INSERT INTO component (id, title, description, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)",
		"deleted-1", "Deleted", "Deleted component", 1000, 1000, 1)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionComponentList,
		Data:   map[string]interface{}{},
	}

	w := performRequest(handler, "POST", "/do", reqBody)

	var response models.Response
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	componentsList, ok := response.Data["components"].([]interface{})
	require.True(t, ok)
	assert.Len(t, componentsList, 1)

	compMap := componentsList[0].(map[string]interface{})
	assert.Equal(t, "Active", compMap["title"])
}

func TestComponentHandler_Modify_Success(t *testing.T) {
	handler := setupComponentTestHandler(t)

	// Insert test component
	componentID := "test-component-id"
	_, err := handler.db.Exec(context.Background(),
		"INSERT INTO component (id, title, description, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)",
		componentID, "Old Title", "Old description", 1000, 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionComponentModify,
		Data: map[string]interface{}{
			"id":          componentID,
			"title":       "New Title",
			"description": "New description",
		},
	}

	w := performRequest(handler, "POST", "/do", reqBody)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Response
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, models.ErrorCodeNoError, response.ErrorCode)
	assert.True(t, response.Data["updated"].(bool))

	// Verify in database
	var title, description string
	err = handler.db.QueryRow(context.Background(),
		"SELECT title, description FROM component WHERE id = ?", componentID).Scan(&title, &description)
	require.NoError(t, err)
	assert.Equal(t, "New Title", title)
	assert.Equal(t, "New description", description)
}

func TestComponentHandler_Modify_PartialUpdate(t *testing.T) {
	handler := setupComponentTestHandler(t)

	// Insert test component
	componentID := "test-component-id"
	_, err := handler.db.Exec(context.Background(),
		"INSERT INTO component (id, title, description, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)",
		componentID, "Original Title", "Original description", 1000, 1000, 0)
	require.NoError(t, err)

	// Update only title
	reqBody := models.Request{
		Action: models.ActionComponentModify,
		Data: map[string]interface{}{
			"id":    componentID,
			"title": "Updated Title",
		},
	}

	w := performRequest(handler, "POST", "/do", reqBody)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify title changed but description stayed the same
	var title, description string
	err = handler.db.QueryRow(context.Background(),
		"SELECT title, description FROM component WHERE id = ?", componentID).Scan(&title, &description)
	require.NoError(t, err)
	assert.Equal(t, "Updated Title", title)
	assert.Equal(t, "Original description", description)
}

func TestComponentHandler_Modify_NotFound(t *testing.T) {
	handler := setupComponentTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionComponentModify,
		Data: map[string]interface{}{
			"id":    "non-existent-id",
			"title": "New Title",
		},
	}

	w := performRequest(handler, "POST", "/do", reqBody)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response models.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, models.ErrorCodeEntityNotFound, response.ErrorCode)
}

func TestComponentHandler_Modify_NoFieldsToUpdate(t *testing.T) {
	handler := setupComponentTestHandler(t)

	// Insert test component
	componentID := "test-component-id"
	_, err := handler.db.Exec(context.Background(),
		"INSERT INTO component (id, title, description, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)",
		componentID, "Title", "Description", 1000, 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionComponentModify,
		Data: map[string]interface{}{
			"id": componentID,
			// No actual fields to update
		},
	}

	w := performRequest(handler, "POST", "/do", reqBody)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response models.Response
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, models.ErrorCodeMissingData, response.ErrorCode)
	assert.Contains(t, response.ErrorMessage, "No fields to update")
}

func TestComponentHandler_Remove_Success(t *testing.T) {
	handler := setupComponentTestHandler(t)

	// Insert test component
	componentID := "test-component-id"
	_, err := handler.db.Exec(context.Background(),
		"INSERT INTO component (id, title, description, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)",
		componentID, "Frontend", "Frontend components", 1000, 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionComponentRemove,
		Data: map[string]interface{}{
			"id": componentID,
		},
	}

	w := performRequest(handler, "POST", "/do", reqBody)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Response
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, models.ErrorCodeNoError, response.ErrorCode)
	assert.True(t, response.Data["deleted"].(bool))

	// Verify soft delete in database
	var deleted bool
	err = handler.db.QueryRow(context.Background(),
		"SELECT deleted FROM component WHERE id = ?", componentID).Scan(&deleted)
	require.NoError(t, err)
	assert.True(t, deleted)
}

func TestComponentHandler_Remove_NotFound(t *testing.T) {
	handler := setupComponentTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionComponentRemove,
		Data: map[string]interface{}{
			"id": "non-existent-id",
		},
	}

	w := performRequest(handler, "POST", "/do", reqBody)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response models.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, models.ErrorCodeEntityNotFound, response.ErrorCode)
}

// ============================================================================
// Component-Ticket Mapping Tests
// ============================================================================

func TestComponentHandler_AddTicket_Success(t *testing.T) {
	handler := setupComponentTestHandler(t)

	// Insert test component
	componentID := "test-component-id"
	_, err := handler.db.Exec(context.Background(),
		"INSERT INTO component (id, title, description, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)",
		componentID, "Database", "Database components", 1000, 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionComponentAddTicket,
		Data: map[string]interface{}{
			"componentId": componentID,
			"ticketId":    "test-ticket-id",
		},
	}

	w := performRequest(handler, "POST", "/do", reqBody)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Response
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, models.ErrorCodeNoError, response.ErrorCode)
	assert.True(t, response.Data["added"].(bool))

	// Verify mapping in database
	var count int
	err = handler.db.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM component_ticket_mapping WHERE component_id = ? AND ticket_id = ? AND deleted = 0",
		componentID, "test-ticket-id").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestComponentHandler_AddTicket_MissingComponentId(t *testing.T) {
	handler := setupComponentTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionComponentAddTicket,
		Data: map[string]interface{}{
			"ticketId": "test-ticket-id",
		},
	}

	w := performRequest(handler, "POST", "/do", reqBody)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response models.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, models.ErrorCodeMissingData, response.ErrorCode)
	assert.Contains(t, response.ErrorMessage, "Missing component ID")
}

func TestComponentHandler_RemoveTicket_Success(t *testing.T) {
	handler := setupComponentTestHandler(t)

	// Insert test component and mapping
	componentID := "test-component-id"
	_, err := handler.db.Exec(context.Background(),
		"INSERT INTO component (id, title, description, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)",
		componentID, "Database", "Database components", 1000, 1000, 0)
	require.NoError(t, err)

	_, err = handler.db.Exec(context.Background(),
		"INSERT INTO component_ticket_mapping (id, component_id, ticket_id, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)",
		"mapping-id", componentID, "test-ticket-id", 1000, 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionComponentRemoveTicket,
		Data: map[string]interface{}{
			"componentId": componentID,
			"ticketId":    "test-ticket-id",
		},
	}

	w := performRequest(handler, "POST", "/do", reqBody)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Response
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, models.ErrorCodeNoError, response.ErrorCode)
	assert.True(t, response.Data["removed"].(bool))

	// Verify soft delete
	var deleted bool
	err = handler.db.QueryRow(context.Background(),
		"SELECT deleted FROM component_ticket_mapping WHERE component_id = ? AND ticket_id = ?",
		componentID, "test-ticket-id").Scan(&deleted)
	require.NoError(t, err)
	assert.True(t, deleted)
}

func TestComponentHandler_RemoveTicket_NotFound(t *testing.T) {
	handler := setupComponentTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionComponentRemoveTicket,
		Data: map[string]interface{}{
			"componentId": "non-existent-component",
			"ticketId":    "non-existent-ticket",
		},
	}

	w := performRequest(handler, "POST", "/do", reqBody)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response models.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, models.ErrorCodeEntityNotFound, response.ErrorCode)
}

func TestComponentHandler_ListTickets_Success(t *testing.T) {
	handler := setupComponentTestHandler(t)

	// Insert test component
	componentID := "test-component-id"
	_, err := handler.db.Exec(context.Background(),
		"INSERT INTO component (id, title, description, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)",
		componentID, "Database", "Database components", 1000, 1000, 0)
	require.NoError(t, err)

	// Insert mappings
	ticketIDs := []string{"ticket-1", "ticket-2", "ticket-3"}
	for i, ticketID := range ticketIDs {
		_, err = handler.db.Exec(context.Background(),
			"INSERT INTO component_ticket_mapping (id, component_id, ticket_id, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)",
			"mapping-"+string(rune('1'+i)), componentID, ticketID, 1000+int64(i), 1000, 0)
		require.NoError(t, err)
	}

	reqBody := models.Request{
		Action: models.ActionComponentListTickets,
		Data: map[string]interface{}{
			"componentId": componentID,
		},
	}

	w := performRequest(handler, "POST", "/do", reqBody)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Response
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, models.ErrorCodeNoError, response.ErrorCode)

	ticketsList, ok := response.Data["ticketIds"].([]interface{})
	require.True(t, ok)
	assert.Len(t, ticketsList, 3)

	count, ok := response.Data["count"].(float64)
	require.True(t, ok)
	assert.Equal(t, float64(3), count)
}

func TestComponentHandler_ListTickets_Empty(t *testing.T) {
	handler := setupComponentTestHandler(t)

	// Insert test component with no mappings
	componentID := "test-component-id"
	_, err := handler.db.Exec(context.Background(),
		"INSERT INTO component (id, title, description, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)",
		componentID, "Database", "Database components", 1000, 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionComponentListTickets,
		Data: map[string]interface{}{
			"componentId": componentID,
		},
	}

	w := performRequest(handler, "POST", "/do", reqBody)

	var response models.Response
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	ticketsList, ok := response.Data["ticketIds"].([]interface{})
	require.True(t, ok)
	assert.Empty(t, ticketsList)
}

// ============================================================================
// Component Metadata Tests
// ============================================================================

func TestComponentHandler_SetMetadata_Success(t *testing.T) {
	handler := setupComponentTestHandler(t)

	// Insert test component
	componentID := "test-component-id"
	_, err := handler.db.Exec(context.Background(),
		"INSERT INTO component (id, title, description, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)",
		componentID, "Database", "Database components", 1000, 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionComponentSetMetadata,
		Data: map[string]interface{}{
			"componentId": componentID,
			"property":    "owner",
			"value":       "john.doe",
		},
	}

	w := performRequest(handler, "POST", "/do", reqBody)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Response
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, models.ErrorCodeNoError, response.ErrorCode)
	assert.True(t, response.Data["set"].(bool))

	// Verify in database
	var value string
	err = handler.db.QueryRow(context.Background(),
		"SELECT value FROM component_meta_data WHERE component_id = ? AND property = ? AND deleted = 0",
		componentID, "owner").Scan(&value)
	require.NoError(t, err)
	assert.Equal(t, "john.doe", value)
}

func TestComponentHandler_SetMetadata_Update(t *testing.T) {
	handler := setupComponentTestHandler(t)

	// Insert test component
	componentID := "test-component-id"
	_, err := handler.db.Exec(context.Background(),
		"INSERT INTO component (id, title, description, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)",
		componentID, "Database", "Database components", 1000, 1000, 0)
	require.NoError(t, err)

	// Insert existing metadata
	_, err = handler.db.Exec(context.Background(),
		"INSERT INTO component_meta_data (id, component_id, property, value, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?)",
		"meta-id", componentID, "owner", "old.owner", 1000, 1000, 0)
	require.NoError(t, err)

	// Update metadata
	reqBody := models.Request{
		Action: models.ActionComponentSetMetadata,
		Data: map[string]interface{}{
			"componentId": componentID,
			"property":    "owner",
			"value":       "new.owner",
		},
	}

	w := performRequest(handler, "POST", "/do", reqBody)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify updated value
	var value string
	err = handler.db.QueryRow(context.Background(),
		"SELECT value FROM component_meta_data WHERE component_id = ? AND property = ?",
		componentID, "owner").Scan(&value)
	require.NoError(t, err)
	assert.Equal(t, "new.owner", value)

	// Verify only one record exists (update, not insert)
	var count int
	err = handler.db.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM component_meta_data WHERE component_id = ? AND property = ?",
		componentID, "owner").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestComponentHandler_GetMetadata_Success(t *testing.T) {
	handler := setupComponentTestHandler(t)

	// Insert test component and metadata
	componentID := "test-component-id"
	_, err := handler.db.Exec(context.Background(),
		"INSERT INTO component (id, title, description, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)",
		componentID, "Database", "Database components", 1000, 1000, 0)
	require.NoError(t, err)

	_, err = handler.db.Exec(context.Background(),
		"INSERT INTO component_meta_data (id, component_id, property, value, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?)",
		"meta-id", componentID, "owner", "john.doe", 1000, 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionComponentGetMetadata,
		Data: map[string]interface{}{
			"componentId": componentID,
			"property":    "owner",
		},
	}

	w := performRequest(handler, "POST", "/do", reqBody)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Response
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, models.ErrorCodeNoError, response.ErrorCode)

	metadata, ok := response.Data["metadata"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "owner", metadata["property"])
	assert.Equal(t, "john.doe", metadata["value"])
}

func TestComponentHandler_GetMetadata_NotFound(t *testing.T) {
	handler := setupComponentTestHandler(t)

	// Insert test component without metadata
	componentID := "test-component-id"
	_, err := handler.db.Exec(context.Background(),
		"INSERT INTO component (id, title, description, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)",
		componentID, "Database", "Database components", 1000, 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionComponentGetMetadata,
		Data: map[string]interface{}{
			"componentId": componentID,
			"property":    "nonexistent",
		},
	}

	w := performRequest(handler, "POST", "/do", reqBody)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response models.Response
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, models.ErrorCodeEntityNotFound, response.ErrorCode)
}

func TestComponentHandler_ListMetadata_Success(t *testing.T) {
	handler := setupComponentTestHandler(t)

	// Insert test component
	componentID := "test-component-id"
	_, err := handler.db.Exec(context.Background(),
		"INSERT INTO component (id, title, description, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)",
		componentID, "Database", "Database components", 1000, 1000, 0)
	require.NoError(t, err)

	// Insert multiple metadata
	properties := []struct {
		property string
		value    string
	}{
		{"owner", "john.doe"},
		{"team", "backend-team"},
		{"version", "1.0.0"},
	}

	for i, prop := range properties {
		_, err = handler.db.Exec(context.Background(),
			"INSERT INTO component_meta_data (id, component_id, property, value, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?)",
			"meta-"+string(rune('1'+i)), componentID, prop.property, prop.value, 1000, 1000, 0)
		require.NoError(t, err)
	}

	reqBody := models.Request{
		Action: models.ActionComponentListMetadata,
		Data: map[string]interface{}{
			"componentId": componentID,
		},
	}

	w := performRequest(handler, "POST", "/do", reqBody)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Response
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, models.ErrorCodeNoError, response.ErrorCode)

	metadataList, ok := response.Data["metadata"].([]interface{})
	require.True(t, ok)
	assert.Len(t, metadataList, 3)

	count, ok := response.Data["count"].(float64)
	require.True(t, ok)
	assert.Equal(t, float64(3), count)
}

func TestComponentHandler_ListMetadata_Empty(t *testing.T) {
	handler := setupComponentTestHandler(t)

	// Insert test component without metadata
	componentID := "test-component-id"
	_, err := handler.db.Exec(context.Background(),
		"INSERT INTO component (id, title, description, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)",
		componentID, "Database", "Database components", 1000, 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionComponentListMetadata,
		Data: map[string]interface{}{
			"componentId": componentID,
		},
	}

	w := performRequest(handler, "POST", "/do", reqBody)

	var response models.Response
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	metadataList, ok := response.Data["metadata"].([]interface{})
	require.True(t, ok)
	assert.Empty(t, metadataList)
}

func TestComponentHandler_RemoveMetadata_Success(t *testing.T) {
	handler := setupComponentTestHandler(t)

	// Insert test component and metadata
	componentID := "test-component-id"
	_, err := handler.db.Exec(context.Background(),
		"INSERT INTO component (id, title, description, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)",
		componentID, "Database", "Database components", 1000, 1000, 0)
	require.NoError(t, err)

	_, err = handler.db.Exec(context.Background(),
		"INSERT INTO component_meta_data (id, component_id, property, value, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?)",
		"meta-id", componentID, "owner", "john.doe", 1000, 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionComponentRemoveMetadata,
		Data: map[string]interface{}{
			"componentId": componentID,
			"property":    "owner",
		},
	}

	w := performRequest(handler, "POST", "/do", reqBody)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Response
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, models.ErrorCodeNoError, response.ErrorCode)
	assert.True(t, response.Data["removed"].(bool))

	// Verify soft delete
	var deleted bool
	err = handler.db.QueryRow(context.Background(),
		"SELECT deleted FROM component_meta_data WHERE component_id = ? AND property = ?",
		componentID, "owner").Scan(&deleted)
	require.NoError(t, err)
	assert.True(t, deleted)
}

func TestComponentHandler_RemoveMetadata_NotFound(t *testing.T) {
	handler := setupComponentTestHandler(t)

	// Insert test component without metadata
	componentID := "test-component-id"
	_, err := handler.db.Exec(context.Background(),
		"INSERT INTO component (id, title, description, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)",
		componentID, "Database", "Database components", 1000, 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionComponentRemoveMetadata,
		Data: map[string]interface{}{
			"componentId": componentID,
			"property":    "nonexistent",
		},
	}

	w := performRequest(handler, "POST", "/do", reqBody)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response models.Response
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, models.ErrorCodeEntityNotFound, response.ErrorCode)
}

// ============================================================================
// Full Component CRUD Cycle Test
// ============================================================================

func TestComponentHandler_FullCRUDCycle(t *testing.T) {
	handler := setupComponentTestHandler(t)

	// 1. Create component
	createReq := models.Request{
		Action: models.ActionComponentCreate,
		Data: map[string]interface{}{
			"title":       "Full Cycle Component",
			"description": "Testing full CRUD cycle",
		},
	}

	w := performRequest(handler, "POST", "/do", createReq)
	assert.Equal(t, http.StatusCreated, w.Code)

	var createResp models.Response
	err := json.Unmarshal(w.Body.Bytes(), &createResp)
	require.NoError(t, err)

	componentData := createResp.Data["component"].(map[string]interface{})
	componentID := componentData["id"].(string)

	// 2. Read component
	readReq := models.Request{
		Action: models.ActionComponentRead,
		Data: map[string]interface{}{
			"id": componentID,
		},
	}

	w = performRequest(handler, "POST", "/do", readReq)
	assert.Equal(t, http.StatusOK, w.Code)

	// 3. Modify component
	modifyReq := models.Request{
		Action: models.ActionComponentModify,
		Data: map[string]interface{}{
			"id":          componentID,
			"title":       "Modified Component",
			"description": "Modified description",
		},
	}

	w = performRequest(handler, "POST", "/do", modifyReq)
	assert.Equal(t, http.StatusOK, w.Code)

	// 4. Verify modification
	w = performRequest(handler, "POST", "/do", readReq)
	var readResp models.Response
	err = json.Unmarshal(w.Body.Bytes(), &readResp)
	require.NoError(t, err)

	modifiedData := readResp.Data["component"].(map[string]interface{})
	assert.Equal(t, "Modified Component", modifiedData["title"])

	// 5. Add to ticket
	addTicketReq := models.Request{
		Action: models.ActionComponentAddTicket,
		Data: map[string]interface{}{
			"componentId": componentID,
			"ticketId":    "test-ticket-id",
		},
	}

	w = performRequest(handler, "POST", "/do", addTicketReq)
	assert.Equal(t, http.StatusOK, w.Code)

	// 6. Set metadata
	setMetaReq := models.Request{
		Action: models.ActionComponentSetMetadata,
		Data: map[string]interface{}{
			"componentId": componentID,
			"property":    "owner",
			"value":       "test.user",
		},
	}

	w = performRequest(handler, "POST", "/do", setMetaReq)
	assert.Equal(t, http.StatusOK, w.Code)

	// 7. Remove component
	removeReq := models.Request{
		Action: models.ActionComponentRemove,
		Data: map[string]interface{}{
			"id": componentID,
		},
	}

	w = performRequest(handler, "POST", "/do", removeReq)
	assert.Equal(t, http.StatusOK, w.Code)

	// 8. Verify component is deleted
	w = performRequest(handler, "POST", "/do", readReq)
	assert.Equal(t, http.StatusNotFound, w.Code)
}
