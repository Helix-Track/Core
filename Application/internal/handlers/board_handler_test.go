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

// setupBoardTestHandler creates a test handler for board tests with database schema
func setupBoardTestHandler(t *testing.T) *Handler {
	handler := setupTestHandler(t)

	// Create board-related tables in the test database
	ctx := context.Background()

	// Create board table
	_, err := handler.db.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS board (
			id          TEXT    NOT NULL PRIMARY KEY UNIQUE,
			title       TEXT,
			description TEXT,
			created     INTEGER NOT NULL,
			modified    INTEGER NOT NULL,
			deleted     BOOLEAN NOT NULL
		)
	`)
	require.NoError(t, err)

	// Create board_meta_data table
	_, err = handler.db.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS board_meta_data (
			id       TEXT    NOT NULL PRIMARY KEY UNIQUE,
			board_id TEXT    NOT NULL,
			property TEXT    NOT NULL,
			value    TEXT,
			created  INTEGER NOT NULL,
			modified INTEGER NOT NULL,
			deleted  BOOLEAN NOT NULL
		)
	`)
	require.NoError(t, err)

	// Create ticket_board_mapping table
	_, err = handler.db.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS ticket_board_mapping (
			id        TEXT    NOT NULL PRIMARY KEY UNIQUE,
			ticket_id TEXT    NOT NULL,
			board_id  TEXT    NOT NULL,
			created   INTEGER NOT NULL,
			modified  INTEGER NOT NULL,
			deleted   BOOLEAN NOT NULL,
			UNIQUE (ticket_id, board_id)
		)
	`)
	require.NoError(t, err)

	return handler
}

// =============================================================================
// handleBoardCreate Tests
// =============================================================================

func TestBoardHandler_Create_Success(t *testing.T) {
	handler := setupBoardTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionBoardCreate,
		Data: map[string]interface{}{
			"title":       "Test Board",
			"description": "Test board description",
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

	board, ok := resp.Data["board"].(map[string]interface{})
	require.True(t, ok)
	assert.NotEmpty(t, board["id"])
	assert.Equal(t, "Test Board", board["title"])
	assert.Equal(t, "Test board description", board["description"])
}

func TestBoardHandler_Create_MinimalFields(t *testing.T) {
	handler := setupBoardTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionBoardCreate,
		Data: map[string]interface{}{
			"title": "Minimal Board",
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
}

func TestBoardHandler_Create_MissingTitle(t *testing.T) {
	handler := setupBoardTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionBoardCreate,
		Data: map[string]interface{}{
			"description": "Description without title",
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

func TestBoardHandler_Create_Unauthorized(t *testing.T) {
	handler := setupBoardTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionBoardCreate,
		Data: map[string]interface{}{
			"title": "Test Board",
		},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("request", &reqBody)
	// No username set - testing unauthorized access

	handler.DoAction(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// =============================================================================
// handleBoardRead Tests
// =============================================================================

func TestBoardHandler_Read_Success(t *testing.T) {
	handler := setupBoardTestHandler(t)

	// Create board first
	createReq := models.Request{
		Action: models.ActionBoardCreate,
		Data: map[string]interface{}{
			"title":       "Read Test Board",
			"description": "Description for read test",
		},
	}
	createBody, _ := json.Marshal(createReq)
	createHttpReq := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(createBody))
	createHttpReq.Header.Set("Content-Type", "application/json")
	wCreate := httptest.NewRecorder()
	cCreate, _ := gin.CreateTestContext(wCreate)
	cCreate.Request = createHttpReq
	cCreate.Set("username", "testuser")
	cCreate.Set("request", &createReq)
	handler.DoAction(cCreate)

	var createResp models.Response
	json.NewDecoder(wCreate.Body).Decode(&createResp)
	boardData := createResp.Data["board"].(map[string]interface{})
	boardID := boardData["id"].(string)

	// Read the board
	readReq := models.Request{
		Action: models.ActionBoardRead,
		Data: map[string]interface{}{
			"id": boardID,
		},
	}
	readBody, _ := json.Marshal(readReq)
	readHttpReq := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(readBody))
	readHttpReq.Header.Set("Content-Type", "application/json")
	wRead := httptest.NewRecorder()
	cRead, _ := gin.CreateTestContext(wRead)
	cRead.Request = readHttpReq
	cRead.Set("username", "testuser")
	cRead.Set("request", &readReq)
	handler.DoAction(cRead)

	assert.Equal(t, http.StatusOK, wRead.Code)

	var readResp models.Response
	err := json.NewDecoder(wRead.Body).Decode(&readResp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, readResp.ErrorCode)

	board := readResp.Data["board"].(map[string]interface{})
	assert.Equal(t, boardID, board["id"])
	assert.Equal(t, "Read Test Board", board["title"])
}

func TestBoardHandler_Read_MissingID(t *testing.T) {
	handler := setupBoardTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionBoardRead,
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

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestBoardHandler_Read_NotFound(t *testing.T) {
	handler := setupBoardTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionBoardRead,
		Data: map[string]interface{}{
			"id": "non-existent-board-id",
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
}

// =============================================================================
// handleBoardList Tests
// =============================================================================

func TestBoardHandler_List_Empty(t *testing.T) {
	handler := setupBoardTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionBoardList,
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

	boards, ok := resp.Data["boards"].([]interface{})
	require.True(t, ok)
	assert.Equal(t, 0, len(boards))
}

func TestBoardHandler_List_Multiple(t *testing.T) {
	handler := setupBoardTestHandler(t)

	// Create 3 boards
	boardNames := []string{"Board 1", "Board 2", "Board 3"}
	for _, name := range boardNames {
		createReq := models.Request{
			Action: models.ActionBoardCreate,
			Data: map[string]interface{}{
				"title": name,
			},
		}
		createBody, _ := json.Marshal(createReq)
		createHttpReq := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(createBody))
		createHttpReq.Header.Set("Content-Type", "application/json")
		wCreate := httptest.NewRecorder()
		cCreate, _ := gin.CreateTestContext(wCreate)
		cCreate.Request = createHttpReq
		cCreate.Set("username", "testuser")
		cCreate.Set("request", &createReq)
		handler.DoAction(cCreate)
	}

	// List boards
	reqBody := models.Request{
		Action: models.ActionBoardList,
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

	boards, ok := resp.Data["boards"].([]interface{})
	require.True(t, ok)
	assert.Equal(t, 3, len(boards))
	assert.Equal(t, float64(3), resp.Data["count"])
}

func TestBoardHandler_List_ExcludesDeleted(t *testing.T) {
	handler := setupBoardTestHandler(t)

	// Create 2 boards
	var boardID string
	for i := 1; i <= 2; i++ {
		createReq := models.Request{
			Action: models.ActionBoardCreate,
			Data: map[string]interface{}{
				"title": "Test Board",
			},
		}
		createBody, _ := json.Marshal(createReq)
		createHttpReq := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(createBody))
		createHttpReq.Header.Set("Content-Type", "application/json")
		wCreate := httptest.NewRecorder()
		cCreate, _ := gin.CreateTestContext(wCreate)
		cCreate.Request = createHttpReq
		cCreate.Set("username", "testuser")
		cCreate.Set("request", &createReq)
		handler.DoAction(cCreate)

		if i == 1 {
			var createResp models.Response
			json.NewDecoder(wCreate.Body).Decode(&createResp)
			boardData := createResp.Data["board"].(map[string]interface{})
			boardID = boardData["id"].(string)
		}
	}

	// Delete first board
	_, err := handler.db.Exec(context.Background(),
		"UPDATE board SET deleted = 1 WHERE id = ?", boardID)
	require.NoError(t, err)

	// List boards
	reqBody := models.Request{
		Action: models.ActionBoardList,
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

	var resp models.Response
	json.NewDecoder(w.Body).Decode(&resp)
	boards := resp.Data["boards"].([]interface{})

	// Should have only 1 board
	assert.Equal(t, 1, len(boards))
}

// =============================================================================
// handleBoardModify Tests
// =============================================================================

func TestBoardHandler_Modify_Success(t *testing.T) {
	handler := setupBoardTestHandler(t)

	// Create board
	createReq := models.Request{
		Action: models.ActionBoardCreate,
		Data: map[string]interface{}{
			"title":       "Original Title",
			"description": "Original Description",
		},
	}
	createBody, _ := json.Marshal(createReq)
	createHttpReq := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(createBody))
	createHttpReq.Header.Set("Content-Type", "application/json")
	wCreate := httptest.NewRecorder()
	cCreate, _ := gin.CreateTestContext(wCreate)
	cCreate.Request = createHttpReq
	cCreate.Set("username", "testuser")
	cCreate.Set("request", &createReq)
	handler.DoAction(cCreate)

	var createResp models.Response
	json.NewDecoder(wCreate.Body).Decode(&createResp)
	boardData := createResp.Data["board"].(map[string]interface{})
	boardID := boardData["id"].(string)

	// Modify board
	modifyReq := models.Request{
		Action: models.ActionBoardModify,
		Data: map[string]interface{}{
			"id":          boardID,
			"title":       "Updated Title",
			"description": "Updated Description",
		},
	}
	modifyBody, _ := json.Marshal(modifyReq)
	modifyHttpReq := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(modifyBody))
	modifyHttpReq.Header.Set("Content-Type", "application/json")
	wModify := httptest.NewRecorder()
	cModify, _ := gin.CreateTestContext(wModify)
	cModify.Request = modifyHttpReq
	cModify.Set("username", "testuser")
	cModify.Set("request", &modifyReq)
	handler.DoAction(cModify)

	assert.Equal(t, http.StatusOK, wModify.Code)

	var modifyResp models.Response
	err := json.NewDecoder(wModify.Body).Decode(&modifyResp)
	require.NoError(t, err)
	assert.True(t, modifyResp.Data["updated"].(bool))
}

func TestBoardHandler_Modify_MissingID(t *testing.T) {
	handler := setupBoardTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionBoardModify,
		Data: map[string]interface{}{
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

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestBoardHandler_Modify_NotFound(t *testing.T) {
	handler := setupBoardTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionBoardModify,
		Data: map[string]interface{}{
			"id":    "non-existent-id",
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
}

// =============================================================================
// handleBoardRemove Tests
// =============================================================================

func TestBoardHandler_Remove_Success(t *testing.T) {
	handler := setupBoardTestHandler(t)

	// Create board
	createReq := models.Request{
		Action: models.ActionBoardCreate,
		Data: map[string]interface{}{
			"title": "To Delete",
		},
	}
	createBody, _ := json.Marshal(createReq)
	createHttpReq := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(createBody))
	createHttpReq.Header.Set("Content-Type", "application/json")
	wCreate := httptest.NewRecorder()
	cCreate, _ := gin.CreateTestContext(wCreate)
	cCreate.Request = createHttpReq
	cCreate.Set("username", "testuser")
	cCreate.Set("request", &createReq)
	handler.DoAction(cCreate)

	var createResp models.Response
	json.NewDecoder(wCreate.Body).Decode(&createResp)
	boardData := createResp.Data["board"].(map[string]interface{})
	boardID := boardData["id"].(string)

	// Remove board
	removeReq := models.Request{
		Action: models.ActionBoardRemove,
		Data: map[string]interface{}{
			"id": boardID,
		},
	}
	removeBody, _ := json.Marshal(removeReq)
	removeHttpReq := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(removeBody))
	removeHttpReq.Header.Set("Content-Type", "application/json")
	wRemove := httptest.NewRecorder()
	cRemove, _ := gin.CreateTestContext(wRemove)
	cRemove.Request = removeHttpReq
	cRemove.Set("username", "testuser")
	cRemove.Set("request", &removeReq)
	handler.DoAction(cRemove)

	assert.Equal(t, http.StatusOK, wRemove.Code)

	var removeResp models.Response
	err := json.NewDecoder(wRemove.Body).Decode(&removeResp)
	require.NoError(t, err)
	assert.True(t, removeResp.Data["deleted"].(bool))
}

func TestBoardHandler_Remove_MissingID(t *testing.T) {
	handler := setupBoardTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionBoardRemove,
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

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestBoardHandler_Remove_NotFound(t *testing.T) {
	handler := setupBoardTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionBoardRemove,
		Data: map[string]interface{}{
			"id": "non-existent-id",
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
}

// =============================================================================
// Helper Functions Tests
// =============================================================================

func TestBoardHandler_CRUD_FullCycle(t *testing.T) {
	handler := setupBoardTestHandler(t)

	// Create
	createReq := models.Request{
		Action: models.ActionBoardCreate,
		Data: map[string]interface{}{
			"title":       "Cycle Board",
			"description": "Full cycle test",
		},
	}
	createBody, _ := json.Marshal(createReq)
	createHttpReq := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(createBody))
	createHttpReq.Header.Set("Content-Type", "application/json")
	wCreate := httptest.NewRecorder()
	cCreate, _ := gin.CreateTestContext(wCreate)
	cCreate.Request = createHttpReq
	cCreate.Set("username", "testuser")
	cCreate.Set("request", &createReq)
	handler.DoAction(cCreate)
	assert.Equal(t, http.StatusCreated, wCreate.Code)

	var createResp models.Response
	json.NewDecoder(wCreate.Body).Decode(&createResp)
	boardData := createResp.Data["board"].(map[string]interface{})
	boardID := boardData["id"].(string)

	// Read
	readReq := models.Request{
		Action: models.ActionBoardRead,
		Data: map[string]interface{}{
			"id": boardID,
		},
	}
	readBody, _ := json.Marshal(readReq)
	readHttpReq := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(readBody))
	readHttpReq.Header.Set("Content-Type", "application/json")
	wRead := httptest.NewRecorder()
	cRead, _ := gin.CreateTestContext(wRead)
	cRead.Request = readHttpReq
	cRead.Set("username", "testuser")
	cRead.Set("request", &readReq)
	handler.DoAction(cRead)
	assert.Equal(t, http.StatusOK, wRead.Code)

	// Modify
	modifyReq := models.Request{
		Action: models.ActionBoardModify,
		Data: map[string]interface{}{
			"id":    boardID,
			"title": "Updated Cycle Board",
		},
	}
	modifyBody, _ := json.Marshal(modifyReq)
	modifyHttpReq := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(modifyBody))
	modifyHttpReq.Header.Set("Content-Type", "application/json")
	wModify := httptest.NewRecorder()
	cModify, _ := gin.CreateTestContext(wModify)
	cModify.Request = modifyHttpReq
	cModify.Set("username", "testuser")
	cModify.Set("request", &modifyReq)
	handler.DoAction(cModify)
	assert.Equal(t, http.StatusOK, wModify.Code)

	// Delete
	removeReq := models.Request{
		Action: models.ActionBoardRemove,
		Data: map[string]interface{}{
			"id": boardID,
		},
	}
	removeBody, _ := json.Marshal(removeReq)
	removeHttpReq := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(removeBody))
	removeHttpReq.Header.Set("Content-Type", "application/json")
	wRemove := httptest.NewRecorder()
	cRemove, _ := gin.CreateTestContext(wRemove)
	cRemove.Request = removeHttpReq
	cRemove.Set("username", "testuser")
	cRemove.Set("request", &removeReq)
	handler.DoAction(cRemove)
	assert.Equal(t, http.StatusOK, wRemove.Code)

	// Verify deleted
	readReq2 := models.Request{
		Action: models.ActionBoardRead,
		Data: map[string]interface{}{
			"id": boardID,
		},
	}
	readBody2, _ := json.Marshal(readReq2)
	readHttpReq2 := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(readBody2))
	readHttpReq2.Header.Set("Content-Type", "application/json")
	wRead2 := httptest.NewRecorder()
	cRead2, _ := gin.CreateTestContext(wRead2)
	cRead2.Request = readHttpReq2
	cRead2.Set("username", "testuser")
	cRead2.Set("request", &readReq2)
	handler.DoAction(cRead2)
	assert.Equal(t, http.StatusNotFound, wRead2.Code)
}
