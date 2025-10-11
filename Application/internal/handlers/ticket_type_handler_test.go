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

// setupTicketTypeTable creates the ticket_type table for testing
func setupTicketTypeTable(t *testing.T, handler *Handler) {
	_, err := handler.db.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS ticket_type (
			id TEXT PRIMARY KEY,
			title TEXT NOT NULL,
			description TEXT,
			icon TEXT,
			color TEXT,
			created INTEGER NOT NULL,
			modified INTEGER NOT NULL,
			deleted INTEGER NOT NULL DEFAULT 0
		)
	`)
	require.NoError(t, err)
}

// setupTicketTypeProjectMappingTable creates the ticket_type_project_mapping table for testing
func setupTicketTypeProjectMappingTable(t *testing.T, handler *Handler) {
	_, err := handler.db.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS ticket_type_project_mapping (
			id TEXT PRIMARY KEY,
			ticket_type_id TEXT NOT NULL,
			project_id TEXT NOT NULL,
			created INTEGER NOT NULL,
			deleted INTEGER NOT NULL DEFAULT 0
		)
	`)
	require.NoError(t, err)
}

// setupSimpleProjectTable creates a simplified project table for testing
func setupSimpleProjectTable(t *testing.T, handler *Handler) {
	_, err := handler.db.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS project (
			id TEXT PRIMARY KEY,
			title TEXT NOT NULL,
			description TEXT,
			created INTEGER NOT NULL,
			modified INTEGER NOT NULL,
			deleted INTEGER NOT NULL DEFAULT 0
		)
	`)
	require.NoError(t, err)
}

// setupTicketTypeTestHandler creates a test handler with ticket type test data
func setupTicketTypeTestHandler(t *testing.T) *Handler {
	handler := setupTestHandler(t)
	setupTicketTypeTable(t, handler)
	setupSimpleProjectTable(t, handler)
	setupTicketTypeProjectMappingTable(t, handler)

	// Insert test project for ticket type assignments
	_, err := handler.db.Exec(context.Background(),
		"INSERT INTO project (id, title, description, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)",
		"test-project-id", "Test Project", "Test project description", 1000, 1000, 0)
	require.NoError(t, err)

	return handler
}

// TestTicketTypeHandler_Create_Success tests successful ticket type creation with all fields
func TestTicketTypeHandler_Create_Success(t *testing.T) {
	handler := setupTicketTypeTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionTicketTypeCreate,
		Data: map[string]interface{}{
			"title":       "Bug",
			"description": "Software defect",
			"icon":        "bug",
			"color":       "#FF0000",
		},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")
	c.Set("request", &reqBody)

	handler.DoAction(c)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)

	ticketType, ok := resp.Data["ticketType"].(map[string]interface{})
	require.True(t, ok)
	assert.NotEmpty(t, ticketType["id"])
	assert.Equal(t, "Bug", ticketType["title"])
	assert.Equal(t, "Software defect", ticketType["description"])
	assert.Equal(t, "bug", ticketType["icon"])
	assert.Equal(t, "#FF0000", ticketType["color"])
}

// TestTicketTypeHandler_Create_MinimalFields tests ticket type creation with only title
func TestTicketTypeHandler_Create_MinimalFields(t *testing.T) {
	handler := setupTicketTypeTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionTicketTypeCreate,
		Data: map[string]interface{}{
			"title": "Task",
		},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")
	c.Set("request", &reqBody)

	handler.DoAction(c)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)

	ticketType, ok := resp.Data["ticketType"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "Task", ticketType["title"])
}

// TestTicketTypeHandler_Create_MissingTitle tests ticket type creation without title
func TestTicketTypeHandler_Create_MissingTitle(t *testing.T) {
	handler := setupTicketTypeTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionTicketTypeCreate,
		Data: map[string]interface{}{
			"description": "Some description",
			"icon":        "icon",
		},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")
	c.Set("request", &reqBody)

	handler.DoAction(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeMissingData, resp.ErrorCode)
}

// TestTicketTypeHandler_Create_MultipleCommonTypes tests creating multiple JIRA-style ticket types
func TestTicketTypeHandler_Create_MultipleCommonTypes(t *testing.T) {
	handler := setupTicketTypeTestHandler(t)

	types := []struct {
		title       string
		description string
		icon        string
		color       string
	}{
		{"Bug", "Software defect", "bug", "#FF0000"},
		{"Feature", "New feature request", "star", "#00FF00"},
		{"Task", "General task", "check", "#0000FF"},
		{"Epic", "Large feature set", "bolt", "#FF00FF"},
		{"Story", "User story", "book", "#00FFFF"},
	}

	for _, typ := range types {
		reqBody := models.Request{
			Action: models.ActionTicketTypeCreate,
			Data: map[string]interface{}{
				"title":       typ.title,
				"description": typ.description,
				"icon":        typ.icon,
				"color":       typ.color,
			},
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("username", "testuser")
		c.Set("request", &reqBody)

		handler.DoAction(c)

		assert.Equal(t, http.StatusCreated, w.Code)

		var resp models.Response
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	}
}

// TestTicketTypeHandler_Read_Success tests successful ticket type read
func TestTicketTypeHandler_Read_Success(t *testing.T) {
	handler := setupTicketTypeTestHandler(t)

	// Insert test ticket type
	_, err := handler.db.Exec(context.Background(),
		"INSERT INTO ticket_type (id, title, description, icon, color, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		"test-type-id", "Bug", "Software defect", "bug", "#FF0000", 1000, 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionTicketTypeRead,
		Data: map[string]interface{}{
			"id": "test-type-id",
		},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")
	c.Set("request", &reqBody)

	handler.DoAction(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err = json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)

	ticketType, ok := resp.Data["ticketType"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "test-type-id", ticketType["id"])
	assert.Equal(t, "Bug", ticketType["title"])
	assert.Equal(t, "bug", ticketType["icon"])
	assert.Equal(t, "#FF0000", ticketType["color"])
}

// TestTicketTypeHandler_Read_NotFound tests reading non-existent ticket type
func TestTicketTypeHandler_Read_NotFound(t *testing.T) {
	handler := setupTicketTypeTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionTicketTypeRead,
		Data: map[string]interface{}{
			"id": "non-existent-type",
		},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")
	c.Set("request", &reqBody)

	handler.DoAction(c)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeEntityNotFound, resp.ErrorCode)
}

// TestTicketTypeHandler_List_Empty tests listing ticket types when none exist
func TestTicketTypeHandler_List_Empty(t *testing.T) {
	handler := setupTicketTypeTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionTicketTypeList,
		Data:   map[string]interface{}{},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")
	c.Set("request", &reqBody)

	handler.DoAction(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)

	ticketTypes, ok := resp.Data["ticketTypes"].([]interface{})
	require.True(t, ok)
	assert.Empty(t, ticketTypes)
}

// TestTicketTypeHandler_List_Multiple tests listing multiple ticket types
func TestTicketTypeHandler_List_Multiple(t *testing.T) {
	handler := setupTicketTypeTestHandler(t)

	// Insert multiple ticket types
	types := []struct {
		id    string
		title string
	}{
		{"type-1", "Bug"},
		{"type-2", "Feature"},
		{"type-3", "Task"},
	}

	for _, typ := range types {
		_, err := handler.db.Exec(context.Background(),
			"INSERT INTO ticket_type (id, title, description, icon, color, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
			typ.id, typ.title, "Description", "icon", "#000000", 1000, 1000, 0)
		require.NoError(t, err)
	}

	reqBody := models.Request{
		Action: models.ActionTicketTypeList,
		Data:   map[string]interface{}{},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")
	c.Set("request", &reqBody)

	handler.DoAction(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)

	ticketTypes, ok := resp.Data["ticketTypes"].([]interface{})
	require.True(t, ok)
	assert.Len(t, ticketTypes, 3)
}

// TestTicketTypeHandler_List_OrderedByTitle tests that types are ordered by title
func TestTicketTypeHandler_List_OrderedByTitle(t *testing.T) {
	handler := setupTicketTypeTestHandler(t)

	// Insert types in non-alphabetical order
	types := []struct {
		id    string
		title string
	}{
		{"type-1", "Zebra"},
		{"type-2", "Apple"},
		{"type-3", "Mango"},
	}

	for _, typ := range types {
		_, err := handler.db.Exec(context.Background(),
			"INSERT INTO ticket_type (id, title, description, icon, color, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
			typ.id, typ.title, "Description", "icon", "#000000", 1000, 1000, 0)
		require.NoError(t, err)
	}

	reqBody := models.Request{
		Action: models.ActionTicketTypeList,
		Data:   map[string]interface{}{},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")
	c.Set("request", &reqBody)

	handler.DoAction(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)

	ticketTypes, ok := resp.Data["ticketTypes"].([]interface{})
	require.True(t, ok)
	assert.Len(t, ticketTypes, 3)

	// Verify alphabetical ordering
	assert.Equal(t, "Apple", ticketTypes[0].(map[string]interface{})["title"])
	assert.Equal(t, "Mango", ticketTypes[1].(map[string]interface{})["title"])
	assert.Equal(t, "Zebra", ticketTypes[2].(map[string]interface{})["title"])
}

// TestTicketTypeHandler_Modify_Success tests successful ticket type modification
func TestTicketTypeHandler_Modify_Success(t *testing.T) {
	handler := setupTicketTypeTestHandler(t)

	// Insert test ticket type
	_, err := handler.db.Exec(context.Background(),
		"INSERT INTO ticket_type (id, title, description, icon, color, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		"test-type-id", "Bug", "Old description", "bug", "#FF0000", 1000, 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionTicketTypeModify,
		Data: map[string]interface{}{
			"id":          "test-type-id",
			"title":       "Defect",
			"description": "Software defect",
			"icon":        "warning",
			"color":       "#FF6600",
		},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")
	c.Set("request", &reqBody)

	handler.DoAction(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err = json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)

	// Verify update in database
	var title, description, icon, color string
	err = handler.db.QueryRow(context.Background(),
		"SELECT title, description, icon, color FROM ticket_type WHERE id = ?",
		"test-type-id").Scan(&title, &description, &icon, &color)
	require.NoError(t, err)
	assert.Equal(t, "Defect", title)
	assert.Equal(t, "Software defect", description)
	assert.Equal(t, "warning", icon)
	assert.Equal(t, "#FF6600", color)
}

// TestTicketTypeHandler_Modify_IconAndColorOnly tests modifying only icon and color
func TestTicketTypeHandler_Modify_IconAndColorOnly(t *testing.T) {
	handler := setupTicketTypeTestHandler(t)

	// Insert test ticket type
	_, err := handler.db.Exec(context.Background(),
		"INSERT INTO ticket_type (id, title, description, icon, color, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		"test-type-id", "Bug", "Original description", "bug", "#FF0000", 1000, 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionTicketTypeModify,
		Data: map[string]interface{}{
			"id":    "test-type-id",
			"icon":  "alert",
			"color": "#00FF00",
		},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")
	c.Set("request", &reqBody)

	handler.DoAction(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err = json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)

	// Verify title/description unchanged, icon/color updated
	var title, description, icon, color string
	err = handler.db.QueryRow(context.Background(),
		"SELECT title, description, icon, color FROM ticket_type WHERE id = ?",
		"test-type-id").Scan(&title, &description, &icon, &color)
	require.NoError(t, err)
	assert.Equal(t, "Bug", title)
	assert.Equal(t, "Original description", description)
	assert.Equal(t, "alert", icon)
	assert.Equal(t, "#00FF00", color)
}

// TestTicketTypeHandler_Modify_NotFound tests modifying non-existent ticket type
func TestTicketTypeHandler_Modify_NotFound(t *testing.T) {
	handler := setupTicketTypeTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionTicketTypeModify,
		Data: map[string]interface{}{
			"id":    "non-existent-type",
			"title": "Updated Title",
		},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")
	c.Set("request", &reqBody)

	handler.DoAction(c)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeEntityNotFound, resp.ErrorCode)
}

// TestTicketTypeHandler_Remove_Success tests successful ticket type deletion
func TestTicketTypeHandler_Remove_Success(t *testing.T) {
	handler := setupTicketTypeTestHandler(t)

	// Insert test ticket type
	_, err := handler.db.Exec(context.Background(),
		"INSERT INTO ticket_type (id, title, description, icon, color, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		"test-type-id", "Deprecated", "Old type", "old", "#CCCCCC", 1000, 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionTicketTypeRemove,
		Data: map[string]interface{}{
			"id": "test-type-id",
		},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")
	c.Set("request", &reqBody)

	handler.DoAction(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err = json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)

	// Verify soft delete in database
	var deleted bool
	err = handler.db.QueryRow(context.Background(),
		"SELECT deleted FROM ticket_type WHERE id = ?",
		"test-type-id").Scan(&deleted)
	require.NoError(t, err)
	assert.True(t, deleted)
}

// TestTicketTypeHandler_Remove_NotFound tests deleting non-existent ticket type
func TestTicketTypeHandler_Remove_NotFound(t *testing.T) {
	handler := setupTicketTypeTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionTicketTypeRemove,
		Data: map[string]interface{}{
			"id": "non-existent-type",
		},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")
	c.Set("request", &reqBody)

	handler.DoAction(c)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeEntityNotFound, resp.ErrorCode)
}

// TestTicketTypeHandler_Assign_Success tests successful ticket type assignment to project
func TestTicketTypeHandler_Assign_Success(t *testing.T) {
	handler := setupTicketTypeTestHandler(t)

	// Insert test ticket type
	_, err := handler.db.Exec(context.Background(),
		"INSERT INTO ticket_type (id, title, description, icon, color, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		"test-type-id", "Bug", "Software defect", "bug", "#FF0000", 1000, 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionTicketTypeAssign,
		Data: map[string]interface{}{
			"ticketTypeId": "test-type-id",
			"projectId":    "test-project-id",
		},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")
	c.Set("request", &reqBody)

	handler.DoAction(c)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp models.Response
	err = json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)

	mapping, ok := resp.Data["mapping"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "test-type-id", mapping["ticketTypeId"])
	assert.Equal(t, "test-project-id", mapping["projectId"])
}

// TestTicketTypeHandler_Assign_AlreadyAssigned tests assigning already assigned ticket type
func TestTicketTypeHandler_Assign_AlreadyAssigned(t *testing.T) {
	handler := setupTicketTypeTestHandler(t)

	// Insert test ticket type
	_, err := handler.db.Exec(context.Background(),
		"INSERT INTO ticket_type (id, title, description, icon, color, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		"test-type-id", "Bug", "Software defect", "bug", "#FF0000", 1000, 1000, 0)
	require.NoError(t, err)

	// Insert existing mapping
	_, err = handler.db.Exec(context.Background(),
		"INSERT INTO ticket_type_project_mapping (id, ticket_type_id, project_id, created, deleted) VALUES (?, ?, ?, ?, ?)",
		"mapping-id", "test-type-id", "test-project-id", 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionTicketTypeAssign,
		Data: map[string]interface{}{
			"ticketTypeId": "test-type-id",
			"projectId":    "test-project-id",
		},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")
	c.Set("request", &reqBody)

	handler.DoAction(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp models.Response
	err = json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeEntityAlreadyExists, resp.ErrorCode)
}

// TestTicketTypeHandler_Unassign_Success tests successful ticket type unassignment
func TestTicketTypeHandler_Unassign_Success(t *testing.T) {
	handler := setupTicketTypeTestHandler(t)

	// Insert test ticket type
	_, err := handler.db.Exec(context.Background(),
		"INSERT INTO ticket_type (id, title, description, icon, color, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		"test-type-id", "Bug", "Software defect", "bug", "#FF0000", 1000, 1000, 0)
	require.NoError(t, err)

	// Insert mapping
	_, err = handler.db.Exec(context.Background(),
		"INSERT INTO ticket_type_project_mapping (id, ticket_type_id, project_id, created, deleted) VALUES (?, ?, ?, ?, ?)",
		"mapping-id", "test-type-id", "test-project-id", 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionTicketTypeUnassign,
		Data: map[string]interface{}{
			"ticketTypeId": "test-type-id",
			"projectId":    "test-project-id",
		},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")
	c.Set("request", &reqBody)

	handler.DoAction(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err = json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)

	// Verify soft delete in database
	var deleted bool
	err = handler.db.QueryRow(context.Background(),
		"SELECT deleted FROM ticket_type_project_mapping WHERE ticket_type_id = ? AND project_id = ?",
		"test-type-id", "test-project-id").Scan(&deleted)
	require.NoError(t, err)
	assert.True(t, deleted)
}

// TestTicketTypeHandler_Unassign_NotFound tests unassigning non-existent mapping
func TestTicketTypeHandler_Unassign_NotFound(t *testing.T) {
	handler := setupTicketTypeTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionTicketTypeUnassign,
		Data: map[string]interface{}{
			"ticketTypeId": "non-existent-type",
			"projectId":    "non-existent-project",
		},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")
	c.Set("request", &reqBody)

	handler.DoAction(c)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeEntityNotFound, resp.ErrorCode)
}

// TestTicketTypeHandler_ListByProject_Success tests listing ticket types for a project
func TestTicketTypeHandler_ListByProject_Success(t *testing.T) {
	handler := setupTicketTypeTestHandler(t)

	// Insert multiple ticket types
	types := []string{"Bug", "Feature", "Task"}
	for i, title := range types {
		typeID := "type-" + string(rune('1'+i))
		_, err := handler.db.Exec(context.Background(),
			"INSERT INTO ticket_type (id, title, description, icon, color, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
			typeID, title, "Description", "icon", "#000000", 1000, 1000, 0)
		require.NoError(t, err)

		// Assign to project
		_, err = handler.db.Exec(context.Background(),
			"INSERT INTO ticket_type_project_mapping (id, ticket_type_id, project_id, created, deleted) VALUES (?, ?, ?, ?, ?)",
			"mapping-"+typeID, typeID, "test-project-id", 1000, 0)
		require.NoError(t, err)
	}

	reqBody := models.Request{
		Action: models.ActionTicketTypeListByProject,
		Data: map[string]interface{}{
			"projectId": "test-project-id",
		},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")
	c.Set("request", &reqBody)

	handler.DoAction(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)

	ticketTypes, ok := resp.Data["ticketTypes"].([]interface{})
	require.True(t, ok)
	assert.Len(t, ticketTypes, 3)
}

// TestTicketTypeHandler_ListByProject_EmptyList tests listing when no types assigned
func TestTicketTypeHandler_ListByProject_EmptyList(t *testing.T) {
	handler := setupTicketTypeTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionTicketTypeListByProject,
		Data: map[string]interface{}{
			"projectId": "test-project-id",
		},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")
	c.Set("request", &reqBody)

	handler.DoAction(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)

	ticketTypes, ok := resp.Data["ticketTypes"].([]interface{})
	require.True(t, ok)
	assert.Empty(t, ticketTypes)
}

// TestTicketTypeHandler_CRUD_FullCycle tests complete ticket type lifecycle
func TestTicketTypeHandler_CRUD_FullCycle(t *testing.T) {
	handler := setupTicketTypeTestHandler(t)

	// 1. Create ticket type
	createReq := models.Request{
		Action: models.ActionTicketTypeCreate,
		Data: map[string]interface{}{
			"title":       "Enhancement",
			"description": "Enhancement request",
			"icon":        "lightbulb",
			"color":       "#FFFF00",
		},
	}
	body, _ := json.Marshal(createReq)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")
	c.Set("request", &createReq)
	handler.DoAction(c)

	var createResp models.Response
	json.NewDecoder(w.Body).Decode(&createResp)
	typeData := createResp.Data["ticketType"].(map[string]interface{})
	typeID := typeData["id"].(string)

	// 2. Read ticket type
	readReq := models.Request{
		Action: models.ActionTicketTypeRead,
		Data:   map[string]interface{}{"id": typeID},
	}
	body, _ = json.Marshal(readReq)
	req = httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")
	c.Set("request", &readReq)
	handler.DoAction(c)
	assert.Equal(t, http.StatusOK, w.Code)

	// 3. Modify ticket type
	modifyReq := models.Request{
		Action: models.ActionTicketTypeModify,
		Data: map[string]interface{}{
			"id":    typeID,
			"color": "#00FFFF",
		},
	}
	body, _ = json.Marshal(modifyReq)
	req = httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")
	c.Set("request", &modifyReq)
	handler.DoAction(c)
	assert.Equal(t, http.StatusOK, w.Code)

	// 4. Delete ticket type
	deleteReq := models.Request{
		Action: models.ActionTicketTypeRemove,
		Data:   map[string]interface{}{"id": typeID},
	}
	body, _ = json.Marshal(deleteReq)
	req = httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")
	c.Set("request", &deleteReq)
	handler.DoAction(c)
	assert.Equal(t, http.StatusOK, w.Code)

	// 5. Verify deletion - type should not be found
	readReq = models.Request{
		Action: models.ActionTicketTypeRead,
		Data:   map[string]interface{}{"id": typeID},
	}
	body, _ = json.Marshal(readReq)
	req = httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")
	c.Set("request", &readReq)
	handler.DoAction(c)
	assert.Equal(t, http.StatusNotFound, w.Code)
}
