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

// TestPriorityHandler_Create_Success tests successful priority creation with all fields
func TestPriorityHandler_Create_Success(t *testing.T) {
	handler := setupTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionPriorityCreate,
		Data: map[string]interface{}{
			"title":       "Critical",
			"description": "Critical priority issues",
			"level":       float64(5),
			"icon":        "alert-triangle",
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

	handler.DoAction(c)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)

	priority, ok := resp.Data["priority"].(map[string]interface{})
	require.True(t, ok)
	assert.NotEmpty(t, priority["id"])
	assert.Equal(t, "Critical", priority["title"])
	assert.Equal(t, "Critical priority issues", priority["description"])
	assert.Equal(t, float64(5), priority["level"])
	assert.Equal(t, "alert-triangle", priority["icon"])
	assert.Equal(t, "#FF0000", priority["color"])
}

// TestPriorityHandler_Create_MinimalFields tests priority creation with only required fields
func TestPriorityHandler_Create_MinimalFields(t *testing.T) {
	handler := setupTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionPriorityCreate,
		Data: map[string]interface{}{
			"title": "Medium",
			"level": float64(3),
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

	priority, ok := resp.Data["priority"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "Medium", priority["title"])
	assert.Equal(t, float64(3), priority["level"])
}

// TestPriorityHandler_Create_AllLevels tests creating priorities at all valid levels
func TestPriorityHandler_Create_AllLevels(t *testing.T) {
	handler := setupTestHandler(t)

	levels := []struct {
		title string
		level int
	}{
		{"Lowest", 1},
		{"Low", 2},
		{"Medium", 3},
		{"High", 4},
		{"Critical", 5},
	}

	for _, lvl := range levels {
		reqBody := models.Request{
			Action: models.ActionPriorityCreate,
			Data: map[string]interface{}{
				"title": lvl.title,
				"level": float64(lvl.level),
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

		priority, ok := resp.Data["priority"].(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, lvl.title, priority["title"])
		assert.Equal(t, float64(lvl.level), priority["level"])
	}
}

// TestPriorityHandler_Create_MissingTitle tests priority creation without title
func TestPriorityHandler_Create_MissingTitle(t *testing.T) {
	handler := setupTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionPriorityCreate,
		Data: map[string]interface{}{
			"level": float64(3),
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

// TestPriorityHandler_Create_MissingLevel tests priority creation without level
func TestPriorityHandler_Create_MissingLevel(t *testing.T) {
	handler := setupTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionPriorityCreate,
		Data: map[string]interface{}{
			"title": "Medium",
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

// TestPriorityHandler_Create_InvalidLevel tests priority creation with invalid level
func TestPriorityHandler_Create_InvalidLevel(t *testing.T) {
	handler := setupTestHandler(t)

	invalidLevels := []int{0, 6, 10, -1}

	for _, level := range invalidLevels {
		reqBody := models.Request{
			Action: models.ActionPriorityCreate,
			Data: map[string]interface{}{
				"title": "Invalid Priority",
				"level": float64(level),
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
		assert.Equal(t, models.ErrorCodeInvalidData, resp.ErrorCode)
	}
}

// TestPriorityHandler_Read_Success tests successful priority read
func TestPriorityHandler_Read_Success(t *testing.T) {
	handler := setupTestHandler(t)

	// Insert test priority
	_, err := handler.db.Exec(context.Background(),
		"INSERT INTO priority (id, title, description, level, icon, color, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		"test-priority-id", "High", "High priority", 4, "arrow-up", "#FF6600", 1000, 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionPriorityRead,
		Data: map[string]interface{}{
			"id": "test-priority-id",
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

	priority, ok := resp.Data["priority"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "test-priority-id", priority["id"])
	assert.Equal(t, "High", priority["title"])
	assert.Equal(t, float64(4), priority["level"])
}

// TestPriorityHandler_Read_NotFound tests reading non-existent priority
func TestPriorityHandler_Read_NotFound(t *testing.T) {
	handler := setupTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionPriorityRead,
		Data: map[string]interface{}{
			"id": "non-existent-priority",
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

// TestPriorityHandler_List_Empty tests listing priorities when none exist
func TestPriorityHandler_List_Empty(t *testing.T) {
	handler := setupTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionPriorityList,
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

	priorities, ok := resp.Data["priorities"].([]interface{})
	require.True(t, ok)
	assert.Empty(t, priorities)
}

// TestPriorityHandler_List_Multiple tests listing multiple priorities
func TestPriorityHandler_List_Multiple(t *testing.T) {
	handler := setupTestHandler(t)

	// Insert multiple priorities
	priorities := []struct {
		id    string
		title string
		level int
	}{
		{"prio-1", "Low", 2},
		{"prio-2", "High", 4},
		{"prio-3", "Medium", 3},
	}

	for _, prio := range priorities {
		_, err := handler.db.Exec(context.Background(),
			"INSERT INTO priority (id, title, description, level, icon, color, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
			prio.id, prio.title, "Description", prio.level, "icon", "#000000", 1000, 1000, 0)
		require.NoError(t, err)
	}

	reqBody := models.Request{
		Action: models.ActionPriorityList,
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

	prioList, ok := resp.Data["priorities"].([]interface{})
	require.True(t, ok)
	assert.Len(t, prioList, 3)
}

// TestPriorityHandler_List_OrderedByLevel tests that priorities are ordered by level
func TestPriorityHandler_List_OrderedByLevel(t *testing.T) {
	handler := setupTestHandler(t)

	// Insert priorities in random level order
	priorities := []struct {
		id    string
		title string
		level int
	}{
		{"prio-1", "Critical", 5},
		{"prio-2", "Lowest", 1},
		{"prio-3", "Medium", 3},
	}

	for _, prio := range priorities {
		_, err := handler.db.Exec(context.Background(),
			"INSERT INTO priority (id, title, description, level, icon, color, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
			prio.id, prio.title, "Description", prio.level, "icon", "#000000", 1000, 1000, 0)
		require.NoError(t, err)
	}

	reqBody := models.Request{
		Action: models.ActionPriorityList,
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

	prioList, ok := resp.Data["priorities"].([]interface{})
	require.True(t, ok)
	assert.Len(t, prioList, 3)

	// Verify ordering by level (ascending: 1, 3, 5)
	assert.Equal(t, "Lowest", prioList[0].(map[string]interface{})["title"])
	assert.Equal(t, float64(1), prioList[0].(map[string]interface{})["level"])
	assert.Equal(t, "Medium", prioList[1].(map[string]interface{})["title"])
	assert.Equal(t, float64(3), prioList[1].(map[string]interface{})["level"])
	assert.Equal(t, "Critical", prioList[2].(map[string]interface{})["title"])
	assert.Equal(t, float64(5), prioList[2].(map[string]interface{})["level"])
}

// TestPriorityHandler_Modify_Success tests successful priority modification
func TestPriorityHandler_Modify_Success(t *testing.T) {
	handler := setupTestHandler(t)

	// Insert test priority
	_, err := handler.db.Exec(context.Background(),
		"INSERT INTO priority (id, title, description, level, icon, color, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		"test-priority-id", "Medium", "Old description", 3, "old-icon", "#000000", 1000, 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionPriorityModify,
		Data: map[string]interface{}{
			"id":          "test-priority-id",
			"title":       "High",
			"description": "Updated description",
			"level":       float64(4),
			"icon":        "arrow-up",
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

	handler.DoAction(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err = json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)

	// Verify update in database
	var title, description, icon, color string
	var level int
	err = handler.db.QueryRow(context.Background(),
		"SELECT title, description, level, icon, color FROM priority WHERE id = ?",
		"test-priority-id").Scan(&title, &description, &level, &icon, &color)
	require.NoError(t, err)
	assert.Equal(t, "High", title)
	assert.Equal(t, "Updated description", description)
	assert.Equal(t, 4, level)
	assert.Equal(t, "arrow-up", icon)
	assert.Equal(t, "#FF6600", color)
}

// TestPriorityHandler_Modify_LevelOnly tests modifying only the level
func TestPriorityHandler_Modify_LevelOnly(t *testing.T) {
	handler := setupTestHandler(t)

	// Insert test priority
	_, err := handler.db.Exec(context.Background(),
		"INSERT INTO priority (id, title, description, level, icon, color, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		"test-priority-id", "Medium", "Description", 3, "icon", "#000000", 1000, 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionPriorityModify,
		Data: map[string]interface{}{
			"id":    "test-priority-id",
			"level": float64(5),
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

	// Verify level updated, other fields unchanged
	var title string
	var level int
	err = handler.db.QueryRow(context.Background(),
		"SELECT title, level FROM priority WHERE id = ?",
		"test-priority-id").Scan(&title, &level)
	require.NoError(t, err)
	assert.Equal(t, "Medium", title)
	assert.Equal(t, 5, level)
}

// TestPriorityHandler_Modify_InvalidLevel tests modifying with invalid level
func TestPriorityHandler_Modify_InvalidLevel(t *testing.T) {
	handler := setupTestHandler(t)

	// Insert test priority
	_, err := handler.db.Exec(context.Background(),
		"INSERT INTO priority (id, title, description, level, icon, color, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		"test-priority-id", "Medium", "Description", 3, "icon", "#000000", 1000, 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionPriorityModify,
		Data: map[string]interface{}{
			"id":    "test-priority-id",
			"level": float64(10), // Invalid level
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
	err = json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeInvalidData, resp.ErrorCode)
}

// TestPriorityHandler_Modify_NotFound tests modifying non-existent priority
func TestPriorityHandler_Modify_NotFound(t *testing.T) {
	handler := setupTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionPriorityModify,
		Data: map[string]interface{}{
			"id":    "non-existent-priority",
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

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeEntityNotFound, resp.ErrorCode)
}

// TestPriorityHandler_Remove_Success tests successful priority deletion
func TestPriorityHandler_Remove_Success(t *testing.T) {
	handler := setupTestHandler(t)

	// Insert test priority
	_, err := handler.db.Exec(context.Background(),
		"INSERT INTO priority (id, title, description, level, icon, color, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		"test-priority-id", "Deprecated", "Old priority", 2, "old", "#CCCCCC", 1000, 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionPriorityRemove,
		Data: map[string]interface{}{
			"id": "test-priority-id",
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
		"SELECT deleted FROM priority WHERE id = ?",
		"test-priority-id").Scan(&deleted)
	require.NoError(t, err)
	assert.True(t, deleted)
}

// TestPriorityHandler_Remove_NotFound tests deleting non-existent priority
func TestPriorityHandler_Remove_NotFound(t *testing.T) {
	handler := setupTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionPriorityRemove,
		Data: map[string]interface{}{
			"id": "non-existent-priority",
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

// TestPriorityHandler_CRUD_FullCycle tests complete priority lifecycle
func TestPriorityHandler_CRUD_FullCycle(t *testing.T) {
	handler := setupTestHandler(t)

	// 1. Create priority
	createReq := models.Request{
		Action: models.ActionPriorityCreate,
		Data: map[string]interface{}{
			"title":       "Urgent",
			"description": "Urgent issues",
			"level":       float64(4),
			"icon":        "zap",
			"color":       "#FF9900",
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
	priorityData := createResp.Data["priority"].(map[string]interface{})
	priorityID := priorityData["id"].(string)

	// 2. Read priority
	readReq := models.Request{
		Action: models.ActionPriorityRead,
		Data:   map[string]interface{}{"id": priorityID},
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

	// 3. Modify priority
	modifyReq := models.Request{
		Action: models.ActionPriorityModify,
		Data: map[string]interface{}{
			"id":    priorityID,
			"level": float64(5),
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

	// 4. Delete priority
	deleteReq := models.Request{
		Action: models.ActionPriorityRemove,
		Data:   map[string]interface{}{"id": priorityID},
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

	// 5. Verify deletion - priority should not be found
	readReq = models.Request{
		Action: models.ActionPriorityRead,
		Data:   map[string]interface{}{"id": priorityID},
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

// Event Publishing Tests

// TestPriorityHandler_Create_PublishesEvent tests that priority creation publishes an event
func TestPriorityHandler_Create_PublishesEvent(t *testing.T) {
	handler, mockPublisher := setupTestHandlerWithPublisher(t)

	reqBody := models.Request{
		Action: models.ActionPriorityCreate,
		Data: map[string]interface{}{
			"title":       "Critical",
			"description": "Critical priority issues",
			"level":       float64(5),
			"icon":        "alert-triangle",
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

	handler.DoAction(c)

	assert.Equal(t, http.StatusCreated, w.Code)

	// Verify event was published
	assert.Equal(t, 1, mockPublisher.GetEventCount())
	lastCall := mockPublisher.GetLastEntityCall()
	require.NotNil(t, lastCall)

	// Verify event details
	assert.Equal(t, models.ActionCreate, lastCall.Action)
	assert.Equal(t, "priority", lastCall.Object)
	assert.Equal(t, "testuser", lastCall.Username)
	assert.NotEmpty(t, lastCall.EntityID)

	// Verify event data
	assert.Equal(t, "Critical", lastCall.Data["title"])
	assert.Equal(t, "Critical priority issues", lastCall.Data["description"])
	assert.Equal(t, float64(5), lastCall.Data["level"])
	assert.Equal(t, "alert-triangle", lastCall.Data["icon"])
	assert.Equal(t, "#FF0000", lastCall.Data["color"])

	// Verify system-wide context (empty project ID)
	assert.Equal(t, "", lastCall.Context.ProjectID)
	assert.Contains(t, lastCall.Context.Permissions, "READ")
}

// TestPriorityHandler_Modify_PublishesEvent tests that priority modification publishes an event
func TestPriorityHandler_Modify_PublishesEvent(t *testing.T) {
	handler, mockPublisher := setupTestHandlerWithPublisher(t)

	// Insert test priority
	_, err := handler.db.Exec(context.Background(),
		"INSERT INTO priority (id, title, description, level, icon, color, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		"test-priority-id", "Medium", "Old description", 3, "old-icon", "#000000", 1000, 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionPriorityModify,
		Data: map[string]interface{}{
			"id":          "test-priority-id",
			"title":       "High",
			"description": "Updated description",
			"level":       float64(4),
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
	assert.Equal(t, "priority", lastCall.Object)
	assert.Equal(t, "test-priority-id", lastCall.EntityID)
	assert.Equal(t, "testuser", lastCall.Username)

	// Verify event data
	assert.Equal(t, "High", lastCall.Data["title"])
	assert.Equal(t, "Updated description", lastCall.Data["description"])
	assert.Equal(t, float64(4), lastCall.Data["level"])

	// Verify system-wide context
	assert.Equal(t, "", lastCall.Context.ProjectID)
	assert.Contains(t, lastCall.Context.Permissions, "READ")
}

// TestPriorityHandler_Remove_PublishesEvent tests that priority deletion publishes an event
func TestPriorityHandler_Remove_PublishesEvent(t *testing.T) {
	handler, mockPublisher := setupTestHandlerWithPublisher(t)

	// Insert test priority
	_, err := handler.db.Exec(context.Background(),
		"INSERT INTO priority (id, title, description, level, icon, color, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		"test-priority-id", "Deprecated", "Old priority", 2, "old", "#CCCCCC", 1000, 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionPriorityRemove,
		Data: map[string]interface{}{
			"id": "test-priority-id",
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
	assert.Equal(t, "priority", lastCall.Object)
	assert.Equal(t, "test-priority-id", lastCall.EntityID)
	assert.Equal(t, "testuser", lastCall.Username)

	// Verify event data
	assert.Equal(t, "test-priority-id", lastCall.Data["id"])
	assert.Equal(t, "Deprecated", lastCall.Data["title"])

	// Verify system-wide context
	assert.Equal(t, "", lastCall.Context.ProjectID)
	assert.Contains(t, lastCall.Context.Permissions, "READ")
}

// TestPriorityHandler_Create_NoEventOnFailure tests that no event is published on create failure
func TestPriorityHandler_Create_NoEventOnFailure(t *testing.T) {
	handler, mockPublisher := setupTestHandlerWithPublisher(t)

	reqBody := models.Request{
		Action: models.ActionPriorityCreate,
		Data: map[string]interface{}{
			// Missing required field 'title'
			"level": float64(3),
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

// TestPriorityHandler_Modify_NoEventOnFailure tests that no event is published on modify failure
func TestPriorityHandler_Modify_NoEventOnFailure(t *testing.T) {
	handler, mockPublisher := setupTestHandlerWithPublisher(t)

	reqBody := models.Request{
		Action: models.ActionPriorityModify,
		Data: map[string]interface{}{
			"id":    "non-existent-priority",
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

// TestPriorityHandler_Remove_NoEventOnFailure tests that no event is published on remove failure
func TestPriorityHandler_Remove_NoEventOnFailure(t *testing.T) {
	handler, mockPublisher := setupTestHandlerWithPublisher(t)

	reqBody := models.Request{
		Action: models.ActionPriorityRemove,
		Data: map[string]interface{}{
			"id": "non-existent-priority",
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
