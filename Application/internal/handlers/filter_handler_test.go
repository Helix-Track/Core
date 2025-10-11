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

// setupFilterTestHandler creates a test handler with filter tables and dependencies
func setupFilterTestHandler(t *testing.T) *Handler {
	handler := setupTestHandler(t)

	// Create filter table
	_, err := handler.db.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS filter (
			id TEXT PRIMARY KEY,
			title TEXT NOT NULL,
			description TEXT,
			owner_id TEXT NOT NULL,
			query TEXT NOT NULL,
			is_public INTEGER NOT NULL DEFAULT 0,
			is_favorite INTEGER NOT NULL DEFAULT 0,
			created INTEGER NOT NULL,
			modified INTEGER NOT NULL,
			deleted INTEGER NOT NULL DEFAULT 0
		)
	`)
	require.NoError(t, err)

	// Create filter_share_mapping table
	_, err = handler.db.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS filter_share_mapping (
			id TEXT PRIMARY KEY,
			filter_id TEXT NOT NULL,
			user_id TEXT,
			team_id TEXT,
			project_id TEXT,
			created INTEGER NOT NULL,
			deleted INTEGER NOT NULL DEFAULT 0
		)
	`)
	require.NoError(t, err)

	return handler
}

// ============================================================================
// Filter Save (Create) Tests
// ============================================================================

func TestFilterHandler_Save_CreateSuccess(t *testing.T) {
	handler := setupFilterTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionFilterSave,
		Data: map[string]interface{}{
			"title":       "My Bugs Filter",
			"description": "All my assigned bugs",
			"query":       `{"status": "open", "type": "bug"}`,
			"isPublic":    false,
			"isFavorite":  true,
		},
	}

	w := performRequest(handler, "POST", "/do", reqBody)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response models.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, models.ErrorCodeNoError, response.ErrorCode)

	filterData, ok := response.Data["filter"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "My Bugs Filter", filterData["Title"])
	assert.Equal(t, `{"status": "open", "type": "bug"}`, filterData["Query"])
	assert.Equal(t, true, filterData["IsFavorite"])
	assert.NotEmpty(t, filterData["ID"])

	// Verify in database
	var count int
	err = handler.db.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM filter WHERE title = ?", "My Bugs Filter").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestFilterHandler_Save_CreateMinimalFields(t *testing.T) {
	handler := setupFilterTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionFilterSave,
		Data: map[string]interface{}{
			"title": "Minimal Filter",
			"query": `{"status": "open"}`,
		},
	}

	w := performRequest(handler, "POST", "/do", reqBody)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response models.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	filterData, ok := response.Data["filter"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "Minimal Filter", filterData["Title"])
	assert.Empty(t, filterData["Description"])
	assert.Equal(t, false, filterData["IsPublic"])
	assert.Equal(t, false, filterData["IsFavorite"])
}

func TestFilterHandler_Save_CreateMissingTitle(t *testing.T) {
	handler := setupFilterTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionFilterSave,
		Data: map[string]interface{}{
			"query": `{"status": "open"}`,
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

func TestFilterHandler_Save_CreateMissingQuery(t *testing.T) {
	handler := setupFilterTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionFilterSave,
		Data: map[string]interface{}{
			"title": "My Filter",
		},
	}

	w := performRequest(handler, "POST", "/do", reqBody)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response models.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, models.ErrorCodeMissingData, response.ErrorCode)
	assert.Contains(t, response.ErrorMessage, "Missing query")
}

func TestFilterHandler_Save_CreateInvalidQueryJSON(t *testing.T) {
	handler := setupFilterTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionFilterSave,
		Data: map[string]interface{}{
			"title": "My Filter",
			"query": `{invalid json}`,
		},
	}

	w := performRequest(handler, "POST", "/do", reqBody)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response models.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, models.ErrorCodeInvalidData, response.ErrorCode)
	assert.Contains(t, response.ErrorMessage, "Invalid query format")
}

func TestFilterHandler_Save_CreateComplexQuery(t *testing.T) {
	handler := setupFilterTestHandler(t)

	complexQuery := `{
		"AND": [
			{"status": "open"},
			{"assignee": "test-user"},
			{"priority": {"$gte": 3}},
			{"labels": {"$in": ["bug", "critical"]}}
		]
	}`

	reqBody := models.Request{
		Action: models.ActionFilterSave,
		Data: map[string]interface{}{
			"title": "Complex Filter",
			"query": complexQuery,
		},
	}

	w := performRequest(handler, "POST", "/do", reqBody)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response models.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	filterData, ok := response.Data["filter"].(map[string]interface{})
	require.True(t, ok)
	// Verify query was stored (exact match might differ due to JSON formatting)
	assert.Contains(t, filterData["Query"], "status")
	assert.Contains(t, filterData["Query"], "priority")
}

// ============================================================================
// Filter Save (Update) Tests
// ============================================================================

func TestFilterHandler_Save_UpdateSuccess(t *testing.T) {
	handler := setupFilterTestHandler(t)

	// Insert existing filter
	filterID := "test-filter-id"
	_, err := handler.db.Exec(context.Background(),
		`INSERT INTO filter (id, title, description, owner_id, query, is_public, is_favorite, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		filterID, "Old Title", "Old description", "test-user", `{"status": "old"}`, 0, 0, 1000, 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionFilterSave,
		Data: map[string]interface{}{
			"id":          filterID,
			"title":       "New Title",
			"description": "New description",
			"query":       `{"status": "new"}`,
			"isPublic":    true,
			"isFavorite":  true,
		},
	}

	w := performRequest(handler, "POST", "/do", reqBody)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Response
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, models.ErrorCodeNoError, response.ErrorCode)

	// Verify in database
	var title, query string
	var isPublic, isFavorite bool
	err = handler.db.QueryRow(context.Background(),
		"SELECT title, query, is_public, is_favorite FROM filter WHERE id = ?", filterID).
		Scan(&title, &query, &isPublic, &isFavorite)
	require.NoError(t, err)
	assert.Equal(t, "New Title", title)
	assert.Equal(t, `{"status": "new"}`, query)
	assert.True(t, isPublic)
	assert.True(t, isFavorite)
}

func TestFilterHandler_Save_UpdateNotFound(t *testing.T) {
	handler := setupFilterTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionFilterSave,
		Data: map[string]interface{}{
			"id":    "non-existent-id",
			"title": "New Title",
			"query": `{"status": "open"}`,
		},
	}

	w := performRequest(handler, "POST", "/do", reqBody)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response models.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, models.ErrorCodeEntityNotFound, response.ErrorCode)
}

func TestFilterHandler_Save_UpdateNotOwner(t *testing.T) {
	handler := setupFilterTestHandler(t)

	// Insert filter owned by different user
	filterID := "test-filter-id"
	_, err := handler.db.Exec(context.Background(),
		`INSERT INTO filter (id, title, description, owner_id, query, is_public, is_favorite, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		filterID, "Title", "Description", "other-user", `{"status": "open"}`, 0, 0, 1000, 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionFilterSave,
		Data: map[string]interface{}{
			"id":    filterID,
			"title": "New Title",
			"query": `{"status": "closed"}`,
		},
	}

	w := performRequest(handler, "POST", "/do", reqBody)

	assert.Equal(t, http.StatusForbidden, w.Code)

	var response models.Response
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, models.ErrorCodeForbidden, response.ErrorCode)
	assert.Contains(t, response.ErrorMessage, "your own filters")
}

// ============================================================================
// Filter Load Tests
// ============================================================================

func TestFilterHandler_Load_Success(t *testing.T) {
	handler := setupFilterTestHandler(t)

	// Insert test filter
	filterID := "test-filter-id"
	_, err := handler.db.Exec(context.Background(),
		`INSERT INTO filter (id, title, description, owner_id, query, is_public, is_favorite, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		filterID, "My Filter", "Description", "test-user", `{"status": "open"}`, 0, 1, 1000, 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionFilterLoad,
		Data: map[string]interface{}{
			"id": filterID,
		},
	}

	w := performRequest(handler, "POST", "/do", reqBody)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Response
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	filterData, ok := response.Data["filter"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, filterID, filterData["ID"])
	assert.Equal(t, "My Filter", filterData["Title"])
}

func TestFilterHandler_Load_NotFound(t *testing.T) {
	handler := setupFilterTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionFilterLoad,
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

func TestFilterHandler_Load_PublicFilter(t *testing.T) {
	handler := setupFilterTestHandler(t)

	// Insert public filter owned by different user
	filterID := "public-filter-id"
	_, err := handler.db.Exec(context.Background(),
		`INSERT INTO filter (id, title, description, owner_id, query, is_public, is_favorite, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		filterID, "Public Filter", "Public description", "other-user", `{"status": "open"}`, 1, 0, 1000, 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionFilterLoad,
		Data: map[string]interface{}{
			"id": filterID,
		},
	}

	w := performRequest(handler, "POST", "/do", reqBody)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Response
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	filterData, ok := response.Data["filter"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "Public Filter", filterData["Title"])
	assert.Equal(t, true, filterData["IsPublic"])
}

func TestFilterHandler_Load_SharedFilter(t *testing.T) {
	handler := setupFilterTestHandler(t)

	// Insert filter owned by different user
	filterID := "shared-filter-id"
	_, err := handler.db.Exec(context.Background(),
		`INSERT INTO filter (id, title, description, owner_id, query, is_public, is_favorite, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		filterID, "Shared Filter", "Shared description", "other-user", `{"status": "open"}`, 0, 0, 1000, 1000, 0)
	require.NoError(t, err)

	// Create share mapping
	testUser := "test-user"
	_, err = handler.db.Exec(context.Background(),
		`INSERT INTO filter_share_mapping (id, filter_id, user_id, team_id, project_id, created, deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		"share-id", filterID, &testUser, nil, nil, 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionFilterLoad,
		Data: map[string]interface{}{
			"id": filterID,
		},
	}

	w := performRequest(handler, "POST", "/do", reqBody)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Response
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	filterData, ok := response.Data["filter"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "Shared Filter", filterData["Title"])
}

func TestFilterHandler_Load_NoAccess(t *testing.T) {
	handler := setupFilterTestHandler(t)

	// Insert private filter owned by different user
	filterID := "private-filter-id"
	_, err := handler.db.Exec(context.Background(),
		`INSERT INTO filter (id, title, description, owner_id, query, is_public, is_favorite, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		filterID, "Private Filter", "Private description", "other-user", `{"status": "open"}`, 0, 0, 1000, 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionFilterLoad,
		Data: map[string]interface{}{
			"id": filterID,
		},
	}

	w := performRequest(handler, "POST", "/do", reqBody)

	assert.Equal(t, http.StatusForbidden, w.Code)

	var response models.Response
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, models.ErrorCodeForbidden, response.ErrorCode)
	assert.Contains(t, response.ErrorMessage, "do not have access")
}

// ============================================================================
// Filter List Tests
// ============================================================================

func TestFilterHandler_List_Empty(t *testing.T) {
	handler := setupFilterTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionFilterList,
		Data:   map[string]interface{}{},
	}

	w := performRequest(handler, "POST", "/do", reqBody)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	filters, ok := response.Data["filters"].([]interface{})
	require.True(t, ok)
	assert.Empty(t, filters)
}

func TestFilterHandler_List_OwnFilters(t *testing.T) {
	handler := setupFilterTestHandler(t)

	// Insert own filters
	_, err := handler.db.Exec(context.Background(),
		`INSERT INTO filter (id, title, description, owner_id, query, is_public, is_favorite, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		"filter-1", "My Filter 1", "Desc", "test-user", `{"status": "open"}`, 0, 0, 1000, 1000, 0)
	require.NoError(t, err)

	_, err = handler.db.Exec(context.Background(),
		`INSERT INTO filter (id, title, description, owner_id, query, is_public, is_favorite, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		"filter-2", "My Filter 2", "Desc", "test-user", `{"status": "closed"}`, 0, 1, 2000, 2000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionFilterList,
		Data:   map[string]interface{}{},
	}

	w := performRequest(handler, "POST", "/do", reqBody)

	var response models.Response
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	filters, ok := response.Data["filters"].([]interface{})
	require.True(t, ok)
	assert.Len(t, filters, 2)
}

func TestFilterHandler_List_IncludesPublicFilters(t *testing.T) {
	handler := setupFilterTestHandler(t)

	// Insert own filter
	_, err := handler.db.Exec(context.Background(),
		`INSERT INTO filter (id, title, description, owner_id, query, is_public, is_favorite, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		"own-filter", "My Filter", "Desc", "test-user", `{"status": "open"}`, 0, 0, 1000, 1000, 0)
	require.NoError(t, err)

	// Insert public filter from another user
	_, err = handler.db.Exec(context.Background(),
		`INSERT INTO filter (id, title, description, owner_id, query, is_public, is_favorite, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		"public-filter", "Public Filter", "Desc", "other-user", `{"status": "closed"}`, 1, 0, 2000, 2000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionFilterList,
		Data:   map[string]interface{}{},
	}

	w := performRequest(handler, "POST", "/do", reqBody)

	var response models.Response
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	filters, ok := response.Data["filters"].([]interface{})
	require.True(t, ok)
	assert.Len(t, filters, 2)
}

func TestFilterHandler_List_OrderedByFavoriteThenModified(t *testing.T) {
	handler := setupFilterTestHandler(t)

	// Insert filters with different favorite/modified times
	_, err := handler.db.Exec(context.Background(),
		`INSERT INTO filter (id, title, description, owner_id, query, is_public, is_favorite, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		"filter-1", "Non-favorite old", "Desc", "test-user", `{}`, 0, 0, 1000, 1000, 0)
	require.NoError(t, err)

	_, err = handler.db.Exec(context.Background(),
		`INSERT INTO filter (id, title, description, owner_id, query, is_public, is_favorite, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		"filter-2", "Favorite old", "Desc", "test-user", `{}`, 0, 1, 2000, 2000, 0)
	require.NoError(t, err)

	_, err = handler.db.Exec(context.Background(),
		`INSERT INTO filter (id, title, description, owner_id, query, is_public, is_favorite, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		"filter-3", "Favorite new", "Desc", "test-user", `{}`, 0, 1, 3000, 3000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionFilterList,
		Data:   map[string]interface{}{},
	}

	w := performRequest(handler, "POST", "/do", reqBody)

	var response models.Response
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	filters, ok := response.Data["filters"].([]interface{})
	require.True(t, ok)

	// Verify order: favorites first (newest first), then non-favorites
	titles := make([]string, len(filters))
	for i, filter := range filters {
		filterMap := filter.(map[string]interface{})
		titles[i] = filterMap["Title"].(string)
	}

	assert.Equal(t, "Favorite new", titles[0])    // Favorite, most recent
	assert.Equal(t, "Favorite old", titles[1])    // Favorite, older
	assert.Equal(t, "Non-favorite old", titles[2]) // Not favorite
}

// ============================================================================
// Filter Share Tests
// ============================================================================

func TestFilterHandler_Share_Public(t *testing.T) {
	handler := setupFilterTestHandler(t)

	// Insert filter
	filterID := "test-filter-id"
	_, err := handler.db.Exec(context.Background(),
		`INSERT INTO filter (id, title, description, owner_id, query, is_public, is_favorite, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		filterID, "My Filter", "Desc", "test-user", `{}`, 0, 0, 1000, 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionFilterShare,
		Data: map[string]interface{}{
			"filterId":  filterID,
			"shareType": "public",
		},
	}

	w := performRequest(handler, "POST", "/do", reqBody)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Response
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response.Data["shared"].(bool))
	assert.True(t, response.Data["isPublic"].(bool))

	// Verify is_public flag is set
	var isPublic bool
	err = handler.db.QueryRow(context.Background(),
		"SELECT is_public FROM filter WHERE id = ?", filterID).Scan(&isPublic)
	require.NoError(t, err)
	assert.True(t, isPublic)
}

func TestFilterHandler_Share_WithUser(t *testing.T) {
	handler := setupFilterTestHandler(t)

	// Insert filter
	filterID := "test-filter-id"
	_, err := handler.db.Exec(context.Background(),
		`INSERT INTO filter (id, title, description, owner_id, query, is_public, is_favorite, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		filterID, "My Filter", "Desc", "test-user", `{}`, 0, 0, 1000, 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionFilterShare,
		Data: map[string]interface{}{
			"filterId":  filterID,
			"shareType": "user",
			"userId":    "other-user",
		},
	}

	w := performRequest(handler, "POST", "/do", reqBody)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Response
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response.Data["shared"].(bool))
	assert.NotEmpty(t, response.Data["shareId"])

	// Verify share mapping
	var count int
	err = handler.db.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM filter_share_mapping WHERE filter_id = ? AND user_id = ?",
		filterID, "other-user").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestFilterHandler_Share_WithTeam(t *testing.T) {
	handler := setupFilterTestHandler(t)

	// Insert filter
	filterID := "test-filter-id"
	_, err := handler.db.Exec(context.Background(),
		`INSERT INTO filter (id, title, description, owner_id, query, is_public, is_favorite, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		filterID, "My Filter", "Desc", "test-user", `{}`, 0, 0, 1000, 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionFilterShare,
		Data: map[string]interface{}{
			"filterId":  filterID,
			"shareType": "team",
			"teamId":    "team-123",
		},
	}

	w := performRequest(handler, "POST", "/do", reqBody)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Response
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response.Data["shared"].(bool))

	// Verify share mapping
	var count int
	err = handler.db.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM filter_share_mapping WHERE filter_id = ? AND team_id = ?",
		filterID, "team-123").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestFilterHandler_Share_NotOwner(t *testing.T) {
	handler := setupFilterTestHandler(t)

	// Insert filter owned by different user
	filterID := "test-filter-id"
	_, err := handler.db.Exec(context.Background(),
		`INSERT INTO filter (id, title, description, owner_id, query, is_public, is_favorite, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		filterID, "My Filter", "Desc", "other-user", `{}`, 0, 0, 1000, 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionFilterShare,
		Data: map[string]interface{}{
			"filterId":  filterID,
			"shareType": "public",
		},
	}

	w := performRequest(handler, "POST", "/do", reqBody)

	assert.Equal(t, http.StatusForbidden, w.Code)

	var response models.Response
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, models.ErrorCodeForbidden, response.ErrorCode)
}

func TestFilterHandler_Share_MissingUserId(t *testing.T) {
	handler := setupFilterTestHandler(t)

	// Insert filter
	filterID := "test-filter-id"
	_, err := handler.db.Exec(context.Background(),
		`INSERT INTO filter (id, title, description, owner_id, query, is_public, is_favorite, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		filterID, "My Filter", "Desc", "test-user", `{}`, 0, 0, 1000, 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionFilterShare,
		Data: map[string]interface{}{
			"filterId":  filterID,
			"shareType": "user",
			// Missing userId
		},
	}

	w := performRequest(handler, "POST", "/do", reqBody)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response models.Response
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, models.ErrorCodeMissingData, response.ErrorCode)
}

// ============================================================================
// Filter Modify Tests
// ============================================================================

func TestFilterHandler_Modify_Success(t *testing.T) {
	handler := setupFilterTestHandler(t)

	// Insert filter
	filterID := "test-filter-id"
	_, err := handler.db.Exec(context.Background(),
		`INSERT INTO filter (id, title, description, owner_id, query, is_public, is_favorite, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		filterID, "Old Title", "Old desc", "test-user", `{"old": true}`, 0, 0, 1000, 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionFilterModify,
		Data: map[string]interface{}{
			"id":    filterID,
			"title": "New Title",
			"query": `{"new": true}`,
		},
	}

	w := performRequest(handler, "POST", "/do", reqBody)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Response
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response.Data["updated"].(bool))

	// Verify in database
	var title, query string
	err = handler.db.QueryRow(context.Background(),
		"SELECT title, query FROM filter WHERE id = ?", filterID).Scan(&title, &query)
	require.NoError(t, err)
	assert.Equal(t, "New Title", title)
	assert.Equal(t, `{"new": true}`, query)
}

func TestFilterHandler_Modify_InvalidQuery(t *testing.T) {
	handler := setupFilterTestHandler(t)

	// Insert filter
	filterID := "test-filter-id"
	_, err := handler.db.Exec(context.Background(),
		`INSERT INTO filter (id, title, description, owner_id, query, is_public, is_favorite, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		filterID, "Title", "Desc", "test-user", `{}`, 0, 0, 1000, 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionFilterModify,
		Data: map[string]interface{}{
			"id":    filterID,
			"query": `{invalid json}`,
		},
	}

	w := performRequest(handler, "POST", "/do", reqBody)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response models.Response
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, models.ErrorCodeInvalidData, response.ErrorCode)
}

func TestFilterHandler_Modify_NoFieldsToUpdate(t *testing.T) {
	handler := setupFilterTestHandler(t)

	// Insert filter
	filterID := "test-filter-id"
	_, err := handler.db.Exec(context.Background(),
		`INSERT INTO filter (id, title, description, owner_id, query, is_public, is_favorite, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		filterID, "Title", "Desc", "test-user", `{}`, 0, 0, 1000, 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionFilterModify,
		Data: map[string]interface{}{
			"id": filterID,
			// No fields to update
		},
	}

	w := performRequest(handler, "POST", "/do", reqBody)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response models.Response
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, models.ErrorCodeMissingData, response.ErrorCode)
}

// ============================================================================
// Filter Remove Tests
// ============================================================================

func TestFilterHandler_Remove_Success(t *testing.T) {
	handler := setupFilterTestHandler(t)

	// Insert filter
	filterID := "test-filter-id"
	_, err := handler.db.Exec(context.Background(),
		`INSERT INTO filter (id, title, description, owner_id, query, is_public, is_favorite, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		filterID, "My Filter", "Desc", "test-user", `{}`, 0, 0, 1000, 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionFilterRemove,
		Data: map[string]interface{}{
			"id": filterID,
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
		"SELECT deleted FROM filter WHERE id = ?", filterID).Scan(&deleted)
	require.NoError(t, err)
	assert.True(t, deleted)
}

func TestFilterHandler_Remove_CascadesShares(t *testing.T) {
	handler := setupFilterTestHandler(t)

	// Insert filter
	filterID := "test-filter-id"
	_, err := handler.db.Exec(context.Background(),
		`INSERT INTO filter (id, title, description, owner_id, query, is_public, is_favorite, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		filterID, "My Filter", "Desc", "test-user", `{}`, 0, 0, 1000, 1000, 0)
	require.NoError(t, err)

	// Create share mappings
	testUser := "user1"
	_, err = handler.db.Exec(context.Background(),
		`INSERT INTO filter_share_mapping (id, filter_id, user_id, team_id, project_id, created, deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		"share-1", filterID, &testUser, nil, nil, 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionFilterRemove,
		Data: map[string]interface{}{
			"id": filterID,
		},
	}

	w := performRequest(handler, "POST", "/do", reqBody)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify share mappings are also deleted
	var shareDeleted bool
	err = handler.db.QueryRow(context.Background(),
		"SELECT deleted FROM filter_share_mapping WHERE filter_id = ?", filterID).Scan(&shareDeleted)
	require.NoError(t, err)
	assert.True(t, shareDeleted)
}

func TestFilterHandler_Remove_NotOwner(t *testing.T) {
	handler := setupFilterTestHandler(t)

	// Insert filter owned by different user
	filterID := "test-filter-id"
	_, err := handler.db.Exec(context.Background(),
		`INSERT INTO filter (id, title, description, owner_id, query, is_public, is_favorite, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		filterID, "My Filter", "Desc", "other-user", `{}`, 0, 0, 1000, 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionFilterRemove,
		Data: map[string]interface{}{
			"id": filterID,
		},
	}

	w := performRequest(handler, "POST", "/do", reqBody)

	assert.Equal(t, http.StatusForbidden, w.Code)

	var response models.Response
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, models.ErrorCodeForbidden, response.ErrorCode)
}

// ============================================================================
// Full Filter Lifecycle Test
// ============================================================================

func TestFilterHandler_FullLifecycle(t *testing.T) {
	handler := setupFilterTestHandler(t)

	// 1. Create filter
	createReq := models.Request{
		Action: models.ActionFilterSave,
		Data: map[string]interface{}{
			"title":       "Lifecycle Filter",
			"description": "Testing full lifecycle",
			"query":       `{"status": "open", "priority": {"$gte": 3}}`,
			"isFavorite":  true,
		},
	}

	w := performRequest(handler, "POST", "/do", createReq)
	assert.Equal(t, http.StatusCreated, w.Code)

	var createResp models.Response
	err := json.Unmarshal(w.Body.Bytes(), &createResp)
	require.NoError(t, err)

	filterData := createResp.Data["filter"].(map[string]interface{})
	filterID := filterData["ID"].(string)

	// 2. Load filter
	loadReq := models.Request{
		Action: models.ActionFilterLoad,
		Data: map[string]interface{}{
			"id": filterID,
		},
	}

	w = performRequest(handler, "POST", "/do", loadReq)
	assert.Equal(t, http.StatusOK, w.Code)

	// 3. Modify filter
	modifyReq := models.Request{
		Action: models.ActionFilterModify,
		Data: map[string]interface{}{
			"id":    filterID,
			"title": "Updated Lifecycle Filter",
			"query": `{"status": "closed"}`,
		},
	}

	w = performRequest(handler, "POST", "/do", modifyReq)
	assert.Equal(t, http.StatusOK, w.Code)

	// 4. Share filter
	shareReq := models.Request{
		Action: models.ActionFilterShare,
		Data: map[string]interface{}{
			"filterId":  filterID,
			"shareType": "public",
		},
	}

	w = performRequest(handler, "POST", "/do", shareReq)
	assert.Equal(t, http.StatusOK, w.Code)

	// 5. List filters (should include our filter)
	listReq := models.Request{
		Action: models.ActionFilterList,
		Data:   map[string]interface{}{},
	}

	w = performRequest(handler, "POST", "/do", listReq)
	var listResp models.Response
	err = json.Unmarshal(w.Body.Bytes(), &listResp)
	require.NoError(t, err)

	filters, ok := listResp.Data["filters"].([]interface{})
	require.True(t, ok)
	assert.Len(t, filters, 1)

	// 6. Remove filter
	removeReq := models.Request{
		Action: models.ActionFilterRemove,
		Data: map[string]interface{}{
			"id": filterID,
		},
	}

	w = performRequest(handler, "POST", "/do", removeReq)
	assert.Equal(t, http.StatusOK, w.Code)

	// 7. Verify filter is deleted
	w = performRequest(handler, "POST", "/do", loadReq)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

// ============================================================================
// Event Publishing Tests
// ============================================================================

// TestFilterHandler_Save_Create_PublishesEvent tests that filter creation publishes an event
func TestFilterHandler_Save_Create_PublishesEvent(t *testing.T) {
	handler, mockPublisher := setupTestHandlerWithPublisher(t)

	// Create filter tables
	_, err := handler.db.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS filter (
			id TEXT PRIMARY KEY,
			title TEXT NOT NULL,
			description TEXT,
			owner_id TEXT NOT NULL,
			query TEXT NOT NULL,
			is_public INTEGER NOT NULL DEFAULT 0,
			is_favorite INTEGER NOT NULL DEFAULT 0,
			created INTEGER NOT NULL,
			modified INTEGER NOT NULL,
			deleted INTEGER NOT NULL DEFAULT 0
		)
	`)
	require.NoError(t, err)

	_, err = handler.db.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS filter_share_mapping (
			id TEXT PRIMARY KEY,
			filter_id TEXT NOT NULL,
			user_id TEXT,
			team_id TEXT,
			project_id TEXT,
			created INTEGER NOT NULL,
			deleted INTEGER NOT NULL DEFAULT 0
		)
	`)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionFilterSave,
		Data: map[string]interface{}{
			"title":       "Critical Bugs",
			"description": "All critical bugs assigned to me",
			"query":       `{"status": "open", "priority": 5}`,
			"isPublic":    false,
			"isFavorite":  true,
		},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")

	handler.DoAction(c)

	assert.Equal(t, http.StatusCreated, w.Code)

	// Verify event was published
	assert.Equal(t, 1, mockPublisher.GetEventCount())
	lastCall := mockPublisher.GetLastEntityCall()
	require.NotNil(t, lastCall)

	// Verify event details
	assert.Equal(t, models.ActionCreate, lastCall.Action)
	assert.Equal(t, "filter", lastCall.Object)
	assert.Equal(t, "testuser", lastCall.Username)
	assert.NotEmpty(t, lastCall.EntityID)

	// Verify event data
	assert.Equal(t, "Critical Bugs", lastCall.Data["title"])
	assert.Equal(t, "All critical bugs assigned to me", lastCall.Data["description"])
	assert.Equal(t, "testuser", lastCall.Data["owner_id"])
	assert.Equal(t, false, lastCall.Data["is_public"])
	assert.Equal(t, true, lastCall.Data["is_favorite"])

	// Verify system-wide context (empty project ID)
	assert.Equal(t, "", lastCall.Context.ProjectID)
	assert.Contains(t, lastCall.Context.Permissions, "READ")
}

// TestFilterHandler_Save_Update_PublishesEvent tests that filter update publishes an event
func TestFilterHandler_Save_Update_PublishesEvent(t *testing.T) {
	handler, mockPublisher := setupTestHandlerWithPublisher(t)

	// Create filter tables
	_, err := handler.db.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS filter (
			id TEXT PRIMARY KEY,
			title TEXT NOT NULL,
			description TEXT,
			owner_id TEXT NOT NULL,
			query TEXT NOT NULL,
			is_public INTEGER NOT NULL DEFAULT 0,
			is_favorite INTEGER NOT NULL DEFAULT 0,
			created INTEGER NOT NULL,
			modified INTEGER NOT NULL,
			deleted INTEGER NOT NULL DEFAULT 0
		)
	`)
	require.NoError(t, err)

	_, err = handler.db.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS filter_share_mapping (
			id TEXT PRIMARY KEY,
			filter_id TEXT NOT NULL,
			user_id TEXT,
			team_id TEXT,
			project_id TEXT,
			created INTEGER NOT NULL,
			deleted INTEGER NOT NULL DEFAULT 0
		)
	`)
	require.NoError(t, err)

	// Insert existing filter
	filterID := "test-filter-id"
	_, err = handler.db.Exec(context.Background(),
		`INSERT INTO filter (id, title, description, owner_id, query, is_public, is_favorite, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		filterID, "Old Title", "Old description", "testuser", `{"status": "old"}`, 0, 0, 1000, 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionFilterSave,
		Data: map[string]interface{}{
			"id":          filterID,
			"title":       "Updated Title",
			"description": "Updated description",
			"query":       `{"status": "new"}`,
			"isPublic":    true,
			"isFavorite":  true,
		},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")

	handler.DoAction(c)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify event was published
	assert.Equal(t, 1, mockPublisher.GetEventCount())
	lastCall := mockPublisher.GetLastEntityCall()
	require.NotNil(t, lastCall)

	// Verify event details
	assert.Equal(t, models.ActionModify, lastCall.Action)
	assert.Equal(t, "filter", lastCall.Object)
	assert.Equal(t, filterID, lastCall.EntityID)
	assert.Equal(t, "testuser", lastCall.Username)

	// Verify event data
	assert.Equal(t, "Updated Title", lastCall.Data["title"])
	assert.Equal(t, "Updated description", lastCall.Data["description"])
	assert.Equal(t, "testuser", lastCall.Data["owner_id"])
	assert.Equal(t, true, lastCall.Data["is_public"])
	assert.Equal(t, true, lastCall.Data["is_favorite"])

	// Verify system-wide context
	assert.Equal(t, "", lastCall.Context.ProjectID)
	assert.Contains(t, lastCall.Context.Permissions, "READ")
}

// TestFilterHandler_Modify_PublishesEvent tests that filter modification publishes an event
func TestFilterHandler_Modify_PublishesEvent(t *testing.T) {
	handler, mockPublisher := setupTestHandlerWithPublisher(t)

	// Create filter tables
	_, err := handler.db.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS filter (
			id TEXT PRIMARY KEY,
			title TEXT NOT NULL,
			description TEXT,
			owner_id TEXT NOT NULL,
			query TEXT NOT NULL,
			is_public INTEGER NOT NULL DEFAULT 0,
			is_favorite INTEGER NOT NULL DEFAULT 0,
			created INTEGER NOT NULL,
			modified INTEGER NOT NULL,
			deleted INTEGER NOT NULL DEFAULT 0
		)
	`)
	require.NoError(t, err)

	_, err = handler.db.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS filter_share_mapping (
			id TEXT PRIMARY KEY,
			filter_id TEXT NOT NULL,
			user_id TEXT,
			team_id TEXT,
			project_id TEXT,
			created INTEGER NOT NULL,
			deleted INTEGER NOT NULL DEFAULT 0
		)
	`)
	require.NoError(t, err)

	// Insert test filter
	filterID := "test-filter-id"
	_, err = handler.db.Exec(context.Background(),
		`INSERT INTO filter (id, title, description, owner_id, query, is_public, is_favorite, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		filterID, "Old Title", "Old description", "testuser", `{"old": true}`, 0, 0, 1000, 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionFilterModify,
		Data: map[string]interface{}{
			"id":    filterID,
			"title": "Modified Title",
			"query": `{"new": true}`,
		},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")

	handler.DoAction(c)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify event was published
	assert.Equal(t, 1, mockPublisher.GetEventCount())
	lastCall := mockPublisher.GetLastEntityCall()
	require.NotNil(t, lastCall)

	// Verify event details
	assert.Equal(t, models.ActionModify, lastCall.Action)
	assert.Equal(t, "filter", lastCall.Object)
	assert.Equal(t, filterID, lastCall.EntityID)
	assert.Equal(t, "testuser", lastCall.Username)

	// Verify event data contains the updated fields
	assert.Equal(t, "Modified Title", lastCall.Data["title"])
	assert.Equal(t, `{"new": true}`, lastCall.Data["query"])

	// Verify system-wide context
	assert.Equal(t, "", lastCall.Context.ProjectID)
	assert.Contains(t, lastCall.Context.Permissions, "READ")
}

// TestFilterHandler_Remove_PublishesEvent tests that filter deletion publishes an event
func TestFilterHandler_Remove_PublishesEvent(t *testing.T) {
	handler, mockPublisher := setupTestHandlerWithPublisher(t)

	// Create filter tables
	_, err := handler.db.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS filter (
			id TEXT PRIMARY KEY,
			title TEXT NOT NULL,
			description TEXT,
			owner_id TEXT NOT NULL,
			query TEXT NOT NULL,
			is_public INTEGER NOT NULL DEFAULT 0,
			is_favorite INTEGER NOT NULL DEFAULT 0,
			created INTEGER NOT NULL,
			modified INTEGER NOT NULL,
			deleted INTEGER NOT NULL DEFAULT 0
		)
	`)
	require.NoError(t, err)

	_, err = handler.db.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS filter_share_mapping (
			id TEXT PRIMARY KEY,
			filter_id TEXT NOT NULL,
			user_id TEXT,
			team_id TEXT,
			project_id TEXT,
			created INTEGER NOT NULL,
			deleted INTEGER NOT NULL DEFAULT 0
		)
	`)
	require.NoError(t, err)

	// Insert test filter
	filterID := "test-filter-id"
	_, err = handler.db.Exec(context.Background(),
		`INSERT INTO filter (id, title, description, owner_id, query, is_public, is_favorite, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		filterID, "Deprecated Filter", "Old filter", "testuser", `{}`, 0, 0, 1000, 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionFilterRemove,
		Data: map[string]interface{}{
			"id": filterID,
		},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")

	handler.DoAction(c)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify event was published
	assert.Equal(t, 1, mockPublisher.GetEventCount())
	lastCall := mockPublisher.GetLastEntityCall()
	require.NotNil(t, lastCall)

	// Verify event details
	assert.Equal(t, models.ActionRemove, lastCall.Action)
	assert.Equal(t, "filter", lastCall.Object)
	assert.Equal(t, filterID, lastCall.EntityID)
	assert.Equal(t, "testuser", lastCall.Username)

	// Verify event data
	assert.Equal(t, filterID, lastCall.Data["id"])
	assert.Equal(t, "testuser", lastCall.Data["owner_id"])

	// Verify system-wide context
	assert.Equal(t, "", lastCall.Context.ProjectID)
	assert.Contains(t, lastCall.Context.Permissions, "READ")
}

// TestFilterHandler_Share_PublishesEvent tests that filter sharing publishes an event
func TestFilterHandler_Share_PublishesEvent(t *testing.T) {
	handler, mockPublisher := setupTestHandlerWithPublisher(t)

	// Create filter tables
	_, err := handler.db.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS filter (
			id TEXT PRIMARY KEY,
			title TEXT NOT NULL,
			description TEXT,
			owner_id TEXT NOT NULL,
			query TEXT NOT NULL,
			is_public INTEGER NOT NULL DEFAULT 0,
			is_favorite INTEGER NOT NULL DEFAULT 0,
			created INTEGER NOT NULL,
			modified INTEGER NOT NULL,
			deleted INTEGER NOT NULL DEFAULT 0
		)
	`)
	require.NoError(t, err)

	_, err = handler.db.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS filter_share_mapping (
			id TEXT PRIMARY KEY,
			filter_id TEXT NOT NULL,
			user_id TEXT,
			team_id TEXT,
			project_id TEXT,
			created INTEGER NOT NULL,
			deleted INTEGER NOT NULL DEFAULT 0
		)
	`)
	require.NoError(t, err)

	// Insert test filter
	filterID := "test-filter-id"
	_, err = handler.db.Exec(context.Background(),
		`INSERT INTO filter (id, title, description, owner_id, query, is_public, is_favorite, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		filterID, "My Filter", "Description", "testuser", `{}`, 0, 0, 1000, 1000, 0)
	require.NoError(t, err)

	// Test public share
	reqBody := models.Request{
		Action: models.ActionFilterShare,
		Data: map[string]interface{}{
			"filterId":  filterID,
			"shareType": "public",
		},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")

	handler.DoAction(c)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify event was published
	assert.Equal(t, 1, mockPublisher.GetEventCount())
	lastCall := mockPublisher.GetLastEntityCall()
	require.NotNil(t, lastCall)

	// Verify event details
	assert.Equal(t, models.ActionModify, lastCall.Action)
	assert.Equal(t, "filter", lastCall.Object)
	assert.Equal(t, filterID, lastCall.EntityID)
	assert.Equal(t, "testuser", lastCall.Username)

	// Verify event data
	assert.Equal(t, filterID, lastCall.Data["id"])
	assert.Equal(t, "testuser", lastCall.Data["owner_id"])
	assert.Equal(t, "public", lastCall.Data["share_type"])
	assert.Equal(t, true, lastCall.Data["is_public"])

	// Verify system-wide context
	assert.Equal(t, "", lastCall.Context.ProjectID)
	assert.Contains(t, lastCall.Context.Permissions, "READ")
}

// TestFilterHandler_Save_NoEventOnFailure tests that no event is published on save failure
func TestFilterHandler_Save_NoEventOnFailure(t *testing.T) {
	handler, mockPublisher := setupTestHandlerWithPublisher(t)

	// Create filter tables
	_, err := handler.db.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS filter (
			id TEXT PRIMARY KEY,
			title TEXT NOT NULL,
			description TEXT,
			owner_id TEXT NOT NULL,
			query TEXT NOT NULL,
			is_public INTEGER NOT NULL DEFAULT 0,
			is_favorite INTEGER NOT NULL DEFAULT 0,
			created INTEGER NOT NULL,
			modified INTEGER NOT NULL,
			deleted INTEGER NOT NULL DEFAULT 0
		)
	`)
	require.NoError(t, err)

	_, err = handler.db.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS filter_share_mapping (
			id TEXT PRIMARY KEY,
			filter_id TEXT NOT NULL,
			user_id TEXT,
			team_id TEXT,
			project_id TEXT,
			created INTEGER NOT NULL,
			deleted INTEGER NOT NULL DEFAULT 0
		)
	`)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionFilterSave,
		Data: map[string]interface{}{
			// Missing required field 'title'
			"query": `{"status": "open"}`,
		},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")

	handler.DoAction(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Verify no event was published
	assert.Equal(t, 0, mockPublisher.GetEventCount())
}

// TestFilterHandler_Modify_NoEventOnFailure tests that no event is published on modify failure
func TestFilterHandler_Modify_NoEventOnFailure(t *testing.T) {
	handler, mockPublisher := setupTestHandlerWithPublisher(t)

	// Create filter tables
	_, err := handler.db.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS filter (
			id TEXT PRIMARY KEY,
			title TEXT NOT NULL,
			description TEXT,
			owner_id TEXT NOT NULL,
			query TEXT NOT NULL,
			is_public INTEGER NOT NULL DEFAULT 0,
			is_favorite INTEGER NOT NULL DEFAULT 0,
			created INTEGER NOT NULL,
			modified INTEGER NOT NULL,
			deleted INTEGER NOT NULL DEFAULT 0
		)
	`)
	require.NoError(t, err)

	_, err = handler.db.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS filter_share_mapping (
			id TEXT PRIMARY KEY,
			filter_id TEXT NOT NULL,
			user_id TEXT,
			team_id TEXT,
			project_id TEXT,
			created INTEGER NOT NULL,
			deleted INTEGER NOT NULL DEFAULT 0
		)
	`)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionFilterModify,
		Data: map[string]interface{}{
			"id":    "non-existent-filter",
			"title": "Updated",
		},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")

	handler.DoAction(c)

	assert.Equal(t, http.StatusNotFound, w.Code)

	// Verify no event was published
	assert.Equal(t, 0, mockPublisher.GetEventCount())
}

// TestFilterHandler_Remove_NoEventOnFailure tests that no event is published on remove failure
func TestFilterHandler_Remove_NoEventOnFailure(t *testing.T) {
	handler, mockPublisher := setupTestHandlerWithPublisher(t)

	// Create filter tables
	_, err := handler.db.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS filter (
			id TEXT PRIMARY KEY,
			title TEXT NOT NULL,
			description TEXT,
			owner_id TEXT NOT NULL,
			query TEXT NOT NULL,
			is_public INTEGER NOT NULL DEFAULT 0,
			is_favorite INTEGER NOT NULL DEFAULT 0,
			created INTEGER NOT NULL,
			modified INTEGER NOT NULL,
			deleted INTEGER NOT NULL DEFAULT 0
		)
	`)
	require.NoError(t, err)

	_, err = handler.db.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS filter_share_mapping (
			id TEXT PRIMARY KEY,
			filter_id TEXT NOT NULL,
			user_id TEXT,
			team_id TEXT,
			project_id TEXT,
			created INTEGER NOT NULL,
			deleted INTEGER NOT NULL DEFAULT 0
		)
	`)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionFilterRemove,
		Data: map[string]interface{}{
			"id": "non-existent-filter",
		},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")

	handler.DoAction(c)

	assert.Equal(t, http.StatusNotFound, w.Code)

	// Verify no event was published
	assert.Equal(t, 0, mockPublisher.GetEventCount())
}

// TestFilterHandler_Share_NoEventOnFailure tests that no event is published on share failure
func TestFilterHandler_Share_NoEventOnFailure(t *testing.T) {
	handler, mockPublisher := setupTestHandlerWithPublisher(t)

	// Create filter tables
	_, err := handler.db.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS filter (
			id TEXT PRIMARY KEY,
			title TEXT NOT NULL,
			description TEXT,
			owner_id TEXT NOT NULL,
			query TEXT NOT NULL,
			is_public INTEGER NOT NULL DEFAULT 0,
			is_favorite INTEGER NOT NULL DEFAULT 0,
			created INTEGER NOT NULL,
			modified INTEGER NOT NULL,
			deleted INTEGER NOT NULL DEFAULT 0
		)
	`)
	require.NoError(t, err)

	_, err = handler.db.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS filter_share_mapping (
			id TEXT PRIMARY KEY,
			filter_id TEXT NOT NULL,
			user_id TEXT,
			team_id TEXT,
			project_id TEXT,
			created INTEGER NOT NULL,
			deleted INTEGER NOT NULL DEFAULT 0
		)
	`)
	require.NoError(t, err)

	// Insert filter owned by different user
	filterID := "test-filter-id"
	_, err = handler.db.Exec(context.Background(),
		`INSERT INTO filter (id, title, description, owner_id, query, is_public, is_favorite, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		filterID, "Other User Filter", "Description", "otheruser", `{}`, 0, 0, 1000, 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionFilterShare,
		Data: map[string]interface{}{
			"filterId":  filterID,
			"shareType": "public",
		},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")

	handler.DoAction(c)

	assert.Equal(t, http.StatusForbidden, w.Code)

	// Verify no event was published
	assert.Equal(t, 0, mockPublisher.GetEventCount())
}
