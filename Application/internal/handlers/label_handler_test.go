package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"helixtrack.ru/core/internal/models"
)

// setupLabelTestHandler creates a test handler with label tables and dependencies
func setupLabelTestHandler(t *testing.T) *Handler {
	handler := setupTestHandler(t)

	// Create label table
	_, err := handler.db.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS label (
			id TEXT PRIMARY KEY,
			title TEXT NOT NULL,
			description TEXT,
			created INTEGER NOT NULL,
			modified INTEGER NOT NULL,
			deleted INTEGER NOT NULL DEFAULT 0
		)
	`)
	require.NoError(t, err)

	// Create label_category table
	_, err = handler.db.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS label_category (
			id TEXT PRIMARY KEY,
			title TEXT NOT NULL,
			description TEXT,
			created INTEGER NOT NULL,
			modified INTEGER NOT NULL,
			deleted INTEGER NOT NULL DEFAULT 0
		)
	`)
	require.NoError(t, err)

	// Create label_ticket_mapping table
	_, err = handler.db.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS label_ticket_mapping (
			id TEXT PRIMARY KEY,
			label_id TEXT NOT NULL,
			ticket_id TEXT NOT NULL,
			created INTEGER NOT NULL,
			modified INTEGER NOT NULL,
			deleted INTEGER NOT NULL DEFAULT 0
		)
	`)
	require.NoError(t, err)

	// Create label_label_category_mapping table
	_, err = handler.db.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS label_label_category_mapping (
			id TEXT PRIMARY KEY,
			label_id TEXT NOT NULL,
			label_category_id TEXT NOT NULL,
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

// ============================================================================
// Label CRUD Tests
// ============================================================================

func TestLabelHandler_Create_Success(t *testing.T) {
	handler := setupLabelTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionLabelCreate,
		Data: map[string]interface{}{
			"title":       "bug",
			"description": "Bug-related issues",
			"color":       "#FF0000",
		},
	}

	w := performRequest(handler, "POST", "/do", reqBody)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response models.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, models.ErrorCodeNoError, response.ErrorCode)

	labelData, ok := response.Data["label"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "bug", labelData["Title"])
	assert.Equal(t, "Bug-related issues", labelData["Description"])
	assert.NotEmpty(t, labelData["ID"])

	// Verify in database
	var count int
	err = handler.db.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM label WHERE title = ?", "bug").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestLabelHandler_Create_MinimalFields(t *testing.T) {
	handler := setupLabelTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionLabelCreate,
		Data: map[string]interface{}{
			"title": "feature",
		},
	}

	w := performRequest(handler, "POST", "/do", reqBody)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response models.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, models.ErrorCodeNoError, response.ErrorCode)

	labelData, ok := response.Data["label"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "feature", labelData["Title"])
}

func TestLabelHandler_Create_MissingTitle(t *testing.T) {
	handler := setupLabelTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionLabelCreate,
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

func TestLabelHandler_Create_MultipleCommonLabels(t *testing.T) {
	handler := setupLabelTestHandler(t)

	labels := []struct {
		title       string
		description string
		color       string
	}{
		{"bug", "Bug reports", "#FF0000"},
		{"feature", "New features", "#00FF00"},
		{"documentation", "Documentation updates", "#0000FF"},
		{"enhancement", "Enhancements", "#FFAA00"},
		{"urgent", "Urgent issues", "#FF00FF"},
	}

	for _, label := range labels {
		reqBody := models.Request{
			Action: models.ActionLabelCreate,
			Data: map[string]interface{}{
				"title":       label.title,
				"description": label.description,
				"color":       label.color,
			},
		}

		w := performRequest(handler, "POST", "/do", reqBody)
		assert.Equal(t, http.StatusCreated, w.Code)
	}

	// Verify all labels were created
	var count int
	err := handler.db.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM label WHERE deleted = 0").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, len(labels), count)
}

func TestLabelHandler_Read_Success(t *testing.T) {
	handler := setupLabelTestHandler(t)

	// Insert test label
	labelID := "test-label-id"
	_, err := handler.db.Exec(context.Background(),
		"INSERT INTO label (id, title, description, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)",
		labelID, "bug", "Bug label", 1000, 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionLabelRead,
		Data: map[string]interface{}{
			"id": labelID,
		},
	}

	w := performRequest(handler, "POST", "/do", reqBody)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Response
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, models.ErrorCodeNoError, response.ErrorCode)

	labelData, ok := response.Data["label"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, labelID, labelData["ID"])
	assert.Equal(t, "bug", labelData["Title"])
}

func TestLabelHandler_Read_NotFound(t *testing.T) {
	handler := setupLabelTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionLabelRead,
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
	assert.Contains(t, response.ErrorMessage, "Label not found")
}

func TestLabelHandler_List_Empty(t *testing.T) {
	handler := setupLabelTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionLabelList,
		Data:   map[string]interface{}{},
	}

	w := performRequest(handler, "POST", "/do", reqBody)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, models.ErrorCodeNoError, response.ErrorCode)

	labels, ok := response.Data["labels"].([]interface{})
	require.True(t, ok)
	assert.Empty(t, labels)

	count, ok := response.Data["count"].(float64)
	require.True(t, ok)
	assert.Equal(t, float64(0), count)
}

func TestLabelHandler_List_Multiple(t *testing.T) {
	handler := setupLabelTestHandler(t)

	// Insert multiple labels
	labels := []struct {
		id    string
		title string
	}{
		{"label-1", "bug"},
		{"label-2", "feature"},
		{"label-3", "documentation"},
	}

	for _, label := range labels {
		_, err := handler.db.Exec(context.Background(),
			"INSERT INTO label (id, title, description, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)",
			label.id, label.title, "Description", 1000, 1000, 0)
		require.NoError(t, err)
	}

	reqBody := models.Request{
		Action: models.ActionLabelList,
		Data:   map[string]interface{}{},
	}

	w := performRequest(handler, "POST", "/do", reqBody)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, models.ErrorCodeNoError, response.ErrorCode)

	labelsList, ok := response.Data["labels"].([]interface{})
	require.True(t, ok)
	assert.Len(t, labelsList, 3)

	count, ok := response.Data["count"].(float64)
	require.True(t, ok)
	assert.Equal(t, float64(3), count)
}

func TestLabelHandler_List_OrderedByTitle(t *testing.T) {
	handler := setupLabelTestHandler(t)

	// Insert labels in random order
	labels := []string{"zebra", "apple", "mango"}
	for i, title := range labels {
		_, err := handler.db.Exec(context.Background(),
			"INSERT INTO label (id, title, description, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)",
			"label-"+string(rune('1'+i)), title, "Description", 1000, 1000, 0)
		require.NoError(t, err)
	}

	reqBody := models.Request{
		Action: models.ActionLabelList,
		Data:   map[string]interface{}{},
	}

	w := performRequest(handler, "POST", "/do", reqBody)

	var response models.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	labelsList, ok := response.Data["labels"].([]interface{})
	require.True(t, ok)

	// Verify alphabetical order
	titles := make([]string, len(labelsList))
	for i, label := range labelsList {
		labelMap := label.(map[string]interface{})
		titles[i] = labelMap["Title"].(string)
	}

	assert.Equal(t, "apple", titles[0])
	assert.Equal(t, "mango", titles[1])
	assert.Equal(t, "zebra", titles[2])
}

func TestLabelHandler_List_ExcludesDeleted(t *testing.T) {
	handler := setupLabelTestHandler(t)

	// Insert active and deleted labels
	_, err := handler.db.Exec(context.Background(),
		"INSERT INTO label (id, title, description, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)",
		"active-1", "bug", "Active label", 1000, 1000, 0)
	require.NoError(t, err)

	_, err = handler.db.Exec(context.Background(),
		"INSERT INTO label (id, title, description, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)",
		"deleted-1", "obsolete", "Deleted label", 1000, 1000, 1)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionLabelList,
		Data:   map[string]interface{}{},
	}

	w := performRequest(handler, "POST", "/do", reqBody)

	var response models.Response
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	labelsList, ok := response.Data["labels"].([]interface{})
	require.True(t, ok)
	assert.Len(t, labelsList, 1)

	labelMap := labelsList[0].(map[string]interface{})
	assert.Equal(t, "bug", labelMap["Title"])
}

func TestLabelHandler_Modify_Success(t *testing.T) {
	handler := setupLabelTestHandler(t)

	// Insert test label
	labelID := "test-label-id"
	_, err := handler.db.Exec(context.Background(),
		"INSERT INTO label (id, title, description, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)",
		labelID, "Old Title", "Old description", 1000, 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionLabelModify,
		Data: map[string]interface{}{
			"id":          labelID,
			"title":       "New Title",
			"description": "New description",
			"color":       "#00FF00",
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
		"SELECT title, description FROM label WHERE id = ?", labelID).Scan(&title, &description)
	require.NoError(t, err)
	assert.Equal(t, "New Title", title)
	assert.Equal(t, "New description", description)
}

func TestLabelHandler_Modify_PartialUpdate(t *testing.T) {
	handler := setupLabelTestHandler(t)

	// Insert test label
	labelID := "test-label-id"
	_, err := handler.db.Exec(context.Background(),
		"INSERT INTO label (id, title, description, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)",
		labelID, "Original Title", "Original description", 1000, 1000, 0)
	require.NoError(t, err)

	// Update only title
	reqBody := models.Request{
		Action: models.ActionLabelModify,
		Data: map[string]interface{}{
			"id":    labelID,
			"title": "Updated Title",
		},
	}

	w := performRequest(handler, "POST", "/do", reqBody)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify title changed but description stayed the same
	var title, description string
	err = handler.db.QueryRow(context.Background(),
		"SELECT title, description FROM label WHERE id = ?", labelID).Scan(&title, &description)
	require.NoError(t, err)
	assert.Equal(t, "Updated Title", title)
	assert.Equal(t, "Original description", description)
}

func TestLabelHandler_Modify_NotFound(t *testing.T) {
	handler := setupLabelTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionLabelModify,
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

func TestLabelHandler_Modify_NoFieldsToUpdate(t *testing.T) {
	handler := setupLabelTestHandler(t)

	// Insert test label
	labelID := "test-label-id"
	_, err := handler.db.Exec(context.Background(),
		"INSERT INTO label (id, title, description, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)",
		labelID, "Title", "Description", 1000, 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionLabelModify,
		Data: map[string]interface{}{
			"id": labelID,
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

func TestLabelHandler_Remove_Success(t *testing.T) {
	handler := setupLabelTestHandler(t)

	// Insert test label
	labelID := "test-label-id"
	_, err := handler.db.Exec(context.Background(),
		"INSERT INTO label (id, title, description, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)",
		labelID, "bug", "Bug label", 1000, 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionLabelRemove,
		Data: map[string]interface{}{
			"id": labelID,
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
		"SELECT deleted FROM label WHERE id = ?", labelID).Scan(&deleted)
	require.NoError(t, err)
	assert.True(t, deleted)
}

func TestLabelHandler_Remove_NotFound(t *testing.T) {
	handler := setupLabelTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionLabelRemove,
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
// Label Category CRUD Tests
// ============================================================================

func TestLabelHandler_CategoryCreate_Success(t *testing.T) {
	handler := setupLabelTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionLabelCategoryCreate,
		Data: map[string]interface{}{
			"title":       "Priority Labels",
			"description": "Labels for priority classification",
		},
	}

	w := performRequest(handler, "POST", "/do", reqBody)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response models.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, models.ErrorCodeNoError, response.ErrorCode)

	categoryData, ok := response.Data["category"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "Priority Labels", categoryData["Title"])
	assert.NotEmpty(t, categoryData["ID"])

	// Verify in database
	var count int
	err = handler.db.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM label_category WHERE title = ?", "Priority Labels").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestLabelHandler_CategoryCreate_MinimalFields(t *testing.T) {
	handler := setupLabelTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionLabelCategoryCreate,
		Data: map[string]interface{}{
			"title": "Status Labels",
		},
	}

	w := performRequest(handler, "POST", "/do", reqBody)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response models.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	categoryData, ok := response.Data["category"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "Status Labels", categoryData["Title"])
}

func TestLabelHandler_CategoryRead_Success(t *testing.T) {
	handler := setupLabelTestHandler(t)

	// Insert test category
	categoryID := "test-category-id"
	_, err := handler.db.Exec(context.Background(),
		"INSERT INTO label_category (id, title, description, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)",
		categoryID, "Priority Labels", "Priority category", 1000, 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionLabelCategoryRead,
		Data: map[string]interface{}{
			"id": categoryID,
		},
	}

	w := performRequest(handler, "POST", "/do", reqBody)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Response
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	categoryData, ok := response.Data["category"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, categoryID, categoryData["ID"])
	assert.Equal(t, "Priority Labels", categoryData["Title"])
}

func TestLabelHandler_CategoryList_Multiple(t *testing.T) {
	handler := setupLabelTestHandler(t)

	// Insert multiple categories
	categories := []string{"Priority", "Status", "Type"}
	for i, title := range categories {
		_, err := handler.db.Exec(context.Background(),
			"INSERT INTO label_category (id, title, description, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)",
			"cat-"+string(rune('1'+i)), title, "Description", 1000, 1000, 0)
		require.NoError(t, err)
	}

	reqBody := models.Request{
		Action: models.ActionLabelCategoryList,
		Data:   map[string]interface{}{},
	}

	w := performRequest(handler, "POST", "/do", reqBody)

	var response models.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	categoriesList, ok := response.Data["categories"].([]interface{})
	require.True(t, ok)
	assert.Len(t, categoriesList, 3)
}

func TestLabelHandler_CategoryModify_Success(t *testing.T) {
	handler := setupLabelTestHandler(t)

	// Insert test category
	categoryID := "test-category-id"
	_, err := handler.db.Exec(context.Background(),
		"INSERT INTO label_category (id, title, description, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)",
		categoryID, "Old Title", "Old description", 1000, 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionLabelCategoryModify,
		Data: map[string]interface{}{
			"id":          categoryID,
			"title":       "New Title",
			"description": "New description",
		},
	}

	w := performRequest(handler, "POST", "/do", reqBody)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Response
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response.Data["updated"].(bool))

	// Verify in database
	var title string
	err = handler.db.QueryRow(context.Background(),
		"SELECT title FROM label_category WHERE id = ?", categoryID).Scan(&title)
	require.NoError(t, err)
	assert.Equal(t, "New Title", title)
}

func TestLabelHandler_CategoryRemove_Success(t *testing.T) {
	handler := setupLabelTestHandler(t)

	// Insert test category
	categoryID := "test-category-id"
	_, err := handler.db.Exec(context.Background(),
		"INSERT INTO label_category (id, title, description, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)",
		categoryID, "Priority", "Priority category", 1000, 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionLabelCategoryRemove,
		Data: map[string]interface{}{
			"id": categoryID,
		},
	}

	w := performRequest(handler, "POST", "/do", reqBody)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Response
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response.Data["deleted"].(bool))

	// Verify soft delete
	var deleted bool
	err = handler.db.QueryRow(context.Background(),
		"SELECT deleted FROM label_category WHERE id = ?", categoryID).Scan(&deleted)
	require.NoError(t, err)
	assert.True(t, deleted)
}

// ============================================================================
// Label-Ticket Mapping Tests
// ============================================================================

func TestLabelHandler_AddTicket_Success(t *testing.T) {
	handler := setupLabelTestHandler(t)

	// Insert test label
	labelID := "test-label-id"
	_, err := handler.db.Exec(context.Background(),
		"INSERT INTO label (id, title, description, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)",
		labelID, "bug", "Bug label", 1000, 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionLabelAddTicket,
		Data: map[string]interface{}{
			"labelId":  labelID,
			"ticketId": "test-ticket-id",
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
		"SELECT COUNT(*) FROM label_ticket_mapping WHERE label_id = ? AND ticket_id = ? AND deleted = 0",
		labelID, "test-ticket-id").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestLabelHandler_AddTicket_MissingLabelId(t *testing.T) {
	handler := setupLabelTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionLabelAddTicket,
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
	assert.Contains(t, response.ErrorMessage, "Missing label ID")
}

func TestLabelHandler_RemoveTicket_Success(t *testing.T) {
	handler := setupLabelTestHandler(t)

	// Insert test label and mapping
	labelID := "test-label-id"
	_, err := handler.db.Exec(context.Background(),
		"INSERT INTO label (id, title, description, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)",
		labelID, "bug", "Bug label", 1000, 1000, 0)
	require.NoError(t, err)

	_, err = handler.db.Exec(context.Background(),
		"INSERT INTO label_ticket_mapping (id, label_id, ticket_id, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)",
		"mapping-id", labelID, "test-ticket-id", 1000, 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionLabelRemoveTicket,
		Data: map[string]interface{}{
			"labelId":  labelID,
			"ticketId": "test-ticket-id",
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
		"SELECT deleted FROM label_ticket_mapping WHERE label_id = ? AND ticket_id = ?",
		labelID, "test-ticket-id").Scan(&deleted)
	require.NoError(t, err)
	assert.True(t, deleted)
}

func TestLabelHandler_RemoveTicket_NotFound(t *testing.T) {
	handler := setupLabelTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionLabelRemoveTicket,
		Data: map[string]interface{}{
			"labelId":  "non-existent-label",
			"ticketId": "non-existent-ticket",
		},
	}

	w := performRequest(handler, "POST", "/do", reqBody)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response models.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, models.ErrorCodeEntityNotFound, response.ErrorCode)
}

func TestLabelHandler_ListTickets_Success(t *testing.T) {
	handler := setupLabelTestHandler(t)

	// Insert test label
	labelID := "test-label-id"
	_, err := handler.db.Exec(context.Background(),
		"INSERT INTO label (id, title, description, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)",
		labelID, "bug", "Bug label", 1000, 1000, 0)
	require.NoError(t, err)

	// Insert mappings
	ticketIDs := []string{"ticket-1", "ticket-2", "ticket-3"}
	for i, ticketID := range ticketIDs {
		_, err = handler.db.Exec(context.Background(),
			"INSERT INTO label_ticket_mapping (id, label_id, ticket_id, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)",
			"mapping-"+string(rune('1'+i)), labelID, ticketID, 1000+int64(i), 1000, 0)
		require.NoError(t, err)
	}

	reqBody := models.Request{
		Action: models.ActionLabelListTickets,
		Data: map[string]interface{}{
			"labelId": labelID,
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

func TestLabelHandler_ListTickets_Empty(t *testing.T) {
	handler := setupLabelTestHandler(t)

	// Insert test label with no mappings
	labelID := "test-label-id"
	_, err := handler.db.Exec(context.Background(),
		"INSERT INTO label (id, title, description, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)",
		labelID, "bug", "Bug label", 1000, 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionLabelListTickets,
		Data: map[string]interface{}{
			"labelId": labelID,
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
// Label-Category Mapping Tests
// ============================================================================

func TestLabelHandler_AssignCategory_Success(t *testing.T) {
	handler := setupLabelTestHandler(t)

	// Insert test label and category
	labelID := "test-label-id"
	categoryID := "test-category-id"

	_, err := handler.db.Exec(context.Background(),
		"INSERT INTO label (id, title, description, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)",
		labelID, "bug", "Bug label", 1000, 1000, 0)
	require.NoError(t, err)

	_, err = handler.db.Exec(context.Background(),
		"INSERT INTO label_category (id, title, description, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)",
		categoryID, "Priority", "Priority category", 1000, 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionLabelAssignCategory,
		Data: map[string]interface{}{
			"labelId":    labelID,
			"categoryId": categoryID,
		},
	}

	w := performRequest(handler, "POST", "/do", reqBody)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Response
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, models.ErrorCodeNoError, response.ErrorCode)
	assert.True(t, response.Data["assigned"].(bool))

	// Verify mapping in database
	var count int
	err = handler.db.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM label_label_category_mapping WHERE label_id = ? AND label_category_id = ? AND deleted = 0",
		labelID, categoryID).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestLabelHandler_AssignCategory_MissingLabelId(t *testing.T) {
	handler := setupLabelTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionLabelAssignCategory,
		Data: map[string]interface{}{
			"categoryId": "test-category-id",
		},
	}

	w := performRequest(handler, "POST", "/do", reqBody)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response models.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, models.ErrorCodeMissingData, response.ErrorCode)
	assert.Contains(t, response.ErrorMessage, "Missing label ID")
}

func TestLabelHandler_UnassignCategory_Success(t *testing.T) {
	handler := setupLabelTestHandler(t)

	// Insert test label, category, and mapping
	labelID := "test-label-id"
	categoryID := "test-category-id"

	_, err := handler.db.Exec(context.Background(),
		"INSERT INTO label (id, title, description, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)",
		labelID, "bug", "Bug label", 1000, 1000, 0)
	require.NoError(t, err)

	_, err = handler.db.Exec(context.Background(),
		"INSERT INTO label_category (id, title, description, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)",
		categoryID, "Priority", "Priority category", 1000, 1000, 0)
	require.NoError(t, err)

	_, err = handler.db.Exec(context.Background(),
		"INSERT INTO label_label_category_mapping (id, label_id, label_category_id, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)",
		"mapping-id", labelID, categoryID, 1000, 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionLabelUnassignCategory,
		Data: map[string]interface{}{
			"labelId":    labelID,
			"categoryId": categoryID,
		},
	}

	w := performRequest(handler, "POST", "/do", reqBody)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Response
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, models.ErrorCodeNoError, response.ErrorCode)
	assert.True(t, response.Data["unassigned"].(bool))

	// Verify soft delete
	var deleted bool
	err = handler.db.QueryRow(context.Background(),
		"SELECT deleted FROM label_label_category_mapping WHERE label_id = ? AND label_category_id = ?",
		labelID, categoryID).Scan(&deleted)
	require.NoError(t, err)
	assert.True(t, deleted)
}

func TestLabelHandler_UnassignCategory_NotFound(t *testing.T) {
	handler := setupLabelTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionLabelUnassignCategory,
		Data: map[string]interface{}{
			"labelId":    "non-existent-label",
			"categoryId": "non-existent-category",
		},
	}

	w := performRequest(handler, "POST", "/do", reqBody)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response models.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, models.ErrorCodeEntityNotFound, response.ErrorCode)
}

func TestLabelHandler_ListCategories_Success(t *testing.T) {
	handler := setupLabelTestHandler(t)

	// Insert test label
	labelID := "test-label-id"
	_, err := handler.db.Exec(context.Background(),
		"INSERT INTO label (id, title, description, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)",
		labelID, "bug", "Bug label", 1000, 1000, 0)
	require.NoError(t, err)

	// Insert categories and mappings
	categoryIDs := []string{"cat-1", "cat-2", "cat-3"}
	for i, categoryID := range categoryIDs {
		_, err = handler.db.Exec(context.Background(),
			"INSERT INTO label_category (id, title, description, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)",
			categoryID, "Category "+string(rune('A'+i)), "Description", 1000, 1000, 0)
		require.NoError(t, err)

		_, err = handler.db.Exec(context.Background(),
			"INSERT INTO label_label_category_mapping (id, label_id, label_category_id, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)",
			"mapping-"+string(rune('1'+i)), labelID, categoryID, 1000+int64(i), 1000, 0)
		require.NoError(t, err)
	}

	reqBody := models.Request{
		Action: models.ActionLabelListCategories,
		Data: map[string]interface{}{
			"labelId": labelID,
		},
	}

	w := performRequest(handler, "POST", "/do", reqBody)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Response
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, models.ErrorCodeNoError, response.ErrorCode)

	categoriesList, ok := response.Data["categoryIds"].([]interface{})
	require.True(t, ok)
	assert.Len(t, categoriesList, 3)

	count, ok := response.Data["count"].(float64)
	require.True(t, ok)
	assert.Equal(t, float64(3), count)
}

func TestLabelHandler_ListCategories_Empty(t *testing.T) {
	handler := setupLabelTestHandler(t)

	// Insert test label with no category mappings
	labelID := "test-label-id"
	_, err := handler.db.Exec(context.Background(),
		"INSERT INTO label (id, title, description, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)",
		labelID, "bug", "Bug label", 1000, 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionLabelListCategories,
		Data: map[string]interface{}{
			"labelId": labelID,
		},
	}

	w := performRequest(handler, "POST", "/do", reqBody)

	var response models.Response
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	categoriesList, ok := response.Data["categoryIds"].([]interface{})
	require.True(t, ok)
	assert.Empty(t, categoriesList)
}

// ============================================================================
// Full Label CRUD Cycle Test
// ============================================================================

func TestLabelHandler_FullCRUDCycle(t *testing.T) {
	handler := setupLabelTestHandler(t)

	// 1. Create label
	createReq := models.Request{
		Action: models.ActionLabelCreate,
		Data: map[string]interface{}{
			"title":       "Full Cycle Label",
			"description": "Testing full CRUD cycle",
			"color":       "#FF00FF",
		},
	}

	w := performRequest(handler, "POST", "/do", createReq)
	assert.Equal(t, http.StatusCreated, w.Code)

	var createResp models.Response
	err := json.Unmarshal(w.Body.Bytes(), &createResp)
	require.NoError(t, err)

	labelData := createResp.Data["label"].(map[string]interface{})
	labelID := labelData["ID"].(string)

	// 2. Read label
	readReq := models.Request{
		Action: models.ActionLabelRead,
		Data: map[string]interface{}{
			"id": labelID,
		},
	}

	w = performRequest(handler, "POST", "/do", readReq)
	assert.Equal(t, http.StatusOK, w.Code)

	// 3. Modify label
	modifyReq := models.Request{
		Action: models.ActionLabelModify,
		Data: map[string]interface{}{
			"id":          labelID,
			"title":       "Modified Label",
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

	modifiedData := readResp.Data["label"].(map[string]interface{})
	assert.Equal(t, "Modified Label", modifiedData["Title"])

	// 5. Add to ticket
	addTicketReq := models.Request{
		Action: models.ActionLabelAddTicket,
		Data: map[string]interface{}{
			"labelId":  labelID,
			"ticketId": "test-ticket-id",
		},
	}

	w = performRequest(handler, "POST", "/do", addTicketReq)
	assert.Equal(t, http.StatusOK, w.Code)

	// 6. Create category and assign
	createCatReq := models.Request{
		Action: models.ActionLabelCategoryCreate,
		Data: map[string]interface{}{
			"title": "Test Category",
		},
	}

	w = performRequest(handler, "POST", "/do", createCatReq)
	var catResp models.Response
	err = json.Unmarshal(w.Body.Bytes(), &catResp)
	require.NoError(t, err)

	categoryData := catResp.Data["category"].(map[string]interface{})
	categoryID := categoryData["ID"].(string)

	assignReq := models.Request{
		Action: models.ActionLabelAssignCategory,
		Data: map[string]interface{}{
			"labelId":    labelID,
			"categoryId": categoryID,
		},
	}

	w = performRequest(handler, "POST", "/do", assignReq)
	assert.Equal(t, http.StatusOK, w.Code)

	// 7. Remove label
	removeReq := models.Request{
		Action: models.ActionLabelRemove,
		Data: map[string]interface{}{
			"id": labelID,
		},
	}

	w = performRequest(handler, "POST", "/do", removeReq)
	assert.Equal(t, http.StatusOK, w.Code)

	// 8. Verify label is deleted
	w = performRequest(handler, "POST", "/do", readReq)
	assert.Equal(t, http.StatusNotFound, w.Code)
}
