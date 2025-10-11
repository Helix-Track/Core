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

// TestResolutionHandler_Create_Success tests successful resolution creation
func TestResolutionHandler_Create_Success(t *testing.T) {
	handler := setupTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionResolutionCreate,
		Data: map[string]interface{}{
			"title":       "Fixed",
			"description": "Issue has been fixed",
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

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)

	resolution, ok := resp.Data["resolution"].(map[string]interface{})
	require.True(t, ok)
	assert.NotEmpty(t, resolution["id"])
	assert.Equal(t, "Fixed", resolution["title"])
	assert.Equal(t, "Issue has been fixed", resolution["description"])
}

// TestResolutionHandler_Create_MinimalFields tests resolution creation with only title
func TestResolutionHandler_Create_MinimalFields(t *testing.T) {
	handler := setupTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionResolutionCreate,
		Data: map[string]interface{}{
			"title": "Done",
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

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)

	resolution, ok := resp.Data["resolution"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "Done", resolution["title"])
}

// TestResolutionHandler_Create_MissingTitle tests resolution creation without title
func TestResolutionHandler_Create_MissingTitle(t *testing.T) {
	handler := setupTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionResolutionCreate,
		Data: map[string]interface{}{
			"description": "Some description",
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

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeMissingData, resp.ErrorCode)
}

// TestResolutionHandler_Create_MultipleCommonResolutions tests creating multiple JIRA-style resolutions
func TestResolutionHandler_Create_MultipleCommonResolutions(t *testing.T) {
	handler := setupTestHandler(t)

	resolutions := []struct {
		title       string
		description string
	}{
		{"Fixed", "Issue has been fixed"},
		{"Won't Fix", "Issue will not be fixed"},
		{"Duplicate", "Duplicate of another issue"},
		{"Cannot Reproduce", "Cannot reproduce the issue"},
		{"Done", "Work has been completed"},
		{"Incomplete", "Work is incomplete"},
	}

	for _, res := range resolutions {
		reqBody := models.Request{
			Action: models.ActionResolutionCreate,
			Data: map[string]interface{}{
				"title":       res.title,
				"description": res.description,
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

		var resp models.Response
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	}
}

// TestResolutionHandler_Read_Success tests successful resolution read
func TestResolutionHandler_Read_Success(t *testing.T) {
	handler := setupTestHandler(t)

	// Insert test resolution
	_, err := handler.db.Exec(context.Background(),
		"INSERT INTO resolution (id, title, description, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)",
		"test-resolution-id", "Fixed", "Issue fixed", 1000, 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionResolutionRead,
		Data: map[string]interface{}{
			"id": "test-resolution-id",
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

	var resp models.Response
	err = json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)

	resolution, ok := resp.Data["resolution"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "test-resolution-id", resolution["id"])
	assert.Equal(t, "Fixed", resolution["title"])
	assert.Equal(t, "Issue fixed", resolution["description"])
}

// TestResolutionHandler_Read_NotFound tests reading non-existent resolution
func TestResolutionHandler_Read_NotFound(t *testing.T) {
	handler := setupTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionResolutionRead,
		Data: map[string]interface{}{
			"id": "non-existent-resolution",
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

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeEntityNotFound, resp.ErrorCode)
}

// TestResolutionHandler_List_Empty tests listing resolutions when none exist
func TestResolutionHandler_List_Empty(t *testing.T) {
	handler := setupTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionResolutionList,
		Data:   map[string]interface{}{},
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

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)

	resolutions, ok := resp.Data["resolutions"].([]interface{})
	require.True(t, ok)
	assert.Empty(t, resolutions)
}

// TestResolutionHandler_List_Multiple tests listing multiple resolutions
func TestResolutionHandler_List_Multiple(t *testing.T) {
	handler := setupTestHandler(t)

	// Insert multiple resolutions
	resolutions := []struct {
		id    string
		title string
	}{
		{"res-1", "Fixed"},
		{"res-2", "Won't Fix"},
		{"res-3", "Duplicate"},
	}

	for _, res := range resolutions {
		_, err := handler.db.Exec(context.Background(),
			"INSERT INTO resolution (id, title, description, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)",
			res.id, res.title, "Description", 1000, 1000, 0)
		require.NoError(t, err)
	}

	reqBody := models.Request{
		Action: models.ActionResolutionList,
		Data:   map[string]interface{}{},
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

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)

	resList, ok := resp.Data["resolutions"].([]interface{})
	require.True(t, ok)
	assert.Len(t, resList, 3)
}

// TestResolutionHandler_List_OrderedByTitle tests that resolutions are ordered by title
func TestResolutionHandler_List_OrderedByTitle(t *testing.T) {
	handler := setupTestHandler(t)

	// Insert resolutions in non-alphabetical order
	resolutions := []struct {
		id    string
		title string
	}{
		{"res-1", "Zebra"},
		{"res-2", "Apple"},
		{"res-3", "Mango"},
	}

	for _, res := range resolutions {
		_, err := handler.db.Exec(context.Background(),
			"INSERT INTO resolution (id, title, description, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)",
			res.id, res.title, "Description", 1000, 1000, 0)
		require.NoError(t, err)
	}

	reqBody := models.Request{
		Action: models.ActionResolutionList,
		Data:   map[string]interface{}{},
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

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)

	resList, ok := resp.Data["resolutions"].([]interface{})
	require.True(t, ok)
	assert.Len(t, resList, 3)

	// Verify alphabetical ordering
	assert.Equal(t, "Apple", resList[0].(map[string]interface{})["title"])
	assert.Equal(t, "Mango", resList[1].(map[string]interface{})["title"])
	assert.Equal(t, "Zebra", resList[2].(map[string]interface{})["title"])
}

// TestResolutionHandler_Modify_Success tests successful resolution modification
func TestResolutionHandler_Modify_Success(t *testing.T) {
	handler := setupTestHandler(t)

	// Insert test resolution
	_, err := handler.db.Exec(context.Background(),
		"INSERT INTO resolution (id, title, description, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)",
		"test-resolution-id", "Fixed", "Old description", 1000, 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionResolutionModify,
		Data: map[string]interface{}{
			"id":          "test-resolution-id",
			"title":       "Resolved",
			"description": "Updated description",
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

	var resp models.Response
	err = json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)

	// Verify update in database
	var title, description string
	err = handler.db.QueryRow(context.Background(),
		"SELECT title, description FROM resolution WHERE id = ?",
		"test-resolution-id").Scan(&title, &description)
	require.NoError(t, err)
	assert.Equal(t, "Resolved", title)
	assert.Equal(t, "Updated description", description)
}

// TestResolutionHandler_Modify_TitleOnly tests modifying only the title
func TestResolutionHandler_Modify_TitleOnly(t *testing.T) {
	handler := setupTestHandler(t)

	// Insert test resolution
	_, err := handler.db.Exec(context.Background(),
		"INSERT INTO resolution (id, title, description, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)",
		"test-resolution-id", "Fixed", "Original description", 1000, 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionResolutionModify,
		Data: map[string]interface{}{
			"id":    "test-resolution-id",
			"title": "Completed",
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

	var resp models.Response
	err = json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)

	// Verify title updated, description unchanged
	var title, description string
	err = handler.db.QueryRow(context.Background(),
		"SELECT title, description FROM resolution WHERE id = ?",
		"test-resolution-id").Scan(&title, &description)
	require.NoError(t, err)
	assert.Equal(t, "Completed", title)
	assert.Equal(t, "Original description", description)
}

// TestResolutionHandler_Modify_NotFound tests modifying non-existent resolution
func TestResolutionHandler_Modify_NotFound(t *testing.T) {
	handler := setupTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionResolutionModify,
		Data: map[string]interface{}{
			"id":    "non-existent-resolution",
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

	handler.DoAction(c)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeEntityNotFound, resp.ErrorCode)
}

// TestResolutionHandler_Remove_Success tests successful resolution deletion
func TestResolutionHandler_Remove_Success(t *testing.T) {
	handler := setupTestHandler(t)

	// Insert test resolution
	_, err := handler.db.Exec(context.Background(),
		"INSERT INTO resolution (id, title, description, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)",
		"test-resolution-id", "Deprecated", "Old resolution", 1000, 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionResolutionRemove,
		Data: map[string]interface{}{
			"id": "test-resolution-id",
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

	var resp models.Response
	err = json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)

	// Verify soft delete in database
	var deleted bool
	err = handler.db.QueryRow(context.Background(),
		"SELECT deleted FROM resolution WHERE id = ?",
		"test-resolution-id").Scan(&deleted)
	require.NoError(t, err)
	assert.True(t, deleted)
}

// TestResolutionHandler_Remove_NotFound tests deleting non-existent resolution
func TestResolutionHandler_Remove_NotFound(t *testing.T) {
	handler := setupTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionResolutionRemove,
		Data: map[string]interface{}{
			"id": "non-existent-resolution",
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

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeEntityNotFound, resp.ErrorCode)
}

// TestResolutionHandler_CRUD_FullCycle tests complete resolution lifecycle
func TestResolutionHandler_CRUD_FullCycle(t *testing.T) {
	handler := setupTestHandler(t)

	// 1. Create resolution
	createReq := models.Request{
		Action: models.ActionResolutionCreate,
		Data: map[string]interface{}{
			"title":       "Works as Designed",
			"description": "Functionality works as designed",
		},
	}
	body, _ := json.Marshal(createReq)
	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")
	handler.DoAction(c)

	var createResp models.Response
	json.NewDecoder(w.Body).Decode(&createResp)
	resolutionData := createResp.Data["resolution"].(map[string]interface{})
	resolutionID := resolutionData["id"].(string)

	// 2. Read resolution
	readReq := models.Request{
		Action: models.ActionResolutionRead,
		Data:   map[string]interface{}{"id": resolutionID},
	}
	body, _ = json.Marshal(readReq)
	req = httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")
	handler.DoAction(c)
	assert.Equal(t, http.StatusOK, w.Code)

	// 3. Modify resolution
	modifyReq := models.Request{
		Action: models.ActionResolutionModify,
		Data: map[string]interface{}{
			"id":    resolutionID,
			"title": "Working as Intended",
		},
	}
	body, _ = json.Marshal(modifyReq)
	req = httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")
	handler.DoAction(c)
	assert.Equal(t, http.StatusOK, w.Code)

	// 4. Delete resolution
	deleteReq := models.Request{
		Action: models.ActionResolutionRemove,
		Data:   map[string]interface{}{"id": resolutionID},
	}
	body, _ = json.Marshal(deleteReq)
	req = httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")
	handler.DoAction(c)
	assert.Equal(t, http.StatusOK, w.Code)

	// 5. Verify deletion - resolution should not be found
	readReq = models.Request{
		Action: models.ActionResolutionRead,
		Data:   map[string]interface{}{"id": resolutionID},
	}
	body, _ = json.Marshal(readReq)
	req = httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")
	handler.DoAction(c)
	assert.Equal(t, http.StatusNotFound, w.Code)
}
