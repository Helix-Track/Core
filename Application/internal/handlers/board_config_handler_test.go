package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"helixtrack.ru/core/internal/config"
	"helixtrack.ru/core/internal/database"
	"helixtrack.ru/core/internal/models"
	"helixtrack.ru/core/internal/services"
)

func setupBoardConfigTestHandler(t *testing.T) (*Handler, database.Database) {
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
			return username == "testuser", nil
		},
	}

	handler := NewHandler(db, mockAuth, mockPerm, "1.0.0-test")
	mockPublisher := NewMockEventPublisher(true)
	handler.SetEventPublisher(mockPublisher)

	// Create board table
	_, err = db.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS board (
			id          TEXT    NOT NULL PRIMARY KEY UNIQUE,
			title       TEXT,
			description TEXT,
			type        TEXT,
			created     INTEGER NOT NULL,
			modified    INTEGER NOT NULL,
			deleted     BOOLEAN NOT NULL DEFAULT 0
		)
	`)
	require.NoError(t, err)

	// Create board_column table
	_, err = db.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS board_column (
			id         TEXT    NOT NULL PRIMARY KEY UNIQUE,
			board_id   TEXT    NOT NULL,
			title      TEXT    NOT NULL,
			status_id  TEXT,
			position   INTEGER NOT NULL,
			max_items  INTEGER,
			created    INTEGER NOT NULL,
			modified   INTEGER NOT NULL,
			deleted    BOOLEAN NOT NULL DEFAULT 0
		)
	`)
	require.NoError(t, err)

	// Create board_swimlane table
	_, err = db.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS board_swimlane (
			id         TEXT    NOT NULL PRIMARY KEY UNIQUE,
			board_id   TEXT    NOT NULL,
			title      TEXT    NOT NULL,
			query      TEXT,
			position   INTEGER NOT NULL,
			created    INTEGER NOT NULL,
			modified   INTEGER NOT NULL,
			deleted    BOOLEAN NOT NULL DEFAULT 0
		)
	`)
	require.NoError(t, err)

	// Create board_quick_filter table
	_, err = db.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS board_quick_filter (
			id         TEXT    NOT NULL PRIMARY KEY UNIQUE,
			board_id   TEXT    NOT NULL,
			title      TEXT    NOT NULL,
			query      TEXT,
			position   INTEGER NOT NULL,
			created    INTEGER NOT NULL,
			deleted    BOOLEAN NOT NULL DEFAULT 0
		)
	`)
	require.NoError(t, err)

	return handler, db
}

// Helper to create a test board
func createTestBoard(t *testing.T, db database.Database, boardID, title, boardType string) {
	_, err := db.Exec(context.Background(),
		"INSERT INTO board (id, title, type, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)",
		boardID, title, boardType, time.Now().Unix(), time.Now().Unix(), false)
	require.NoError(t, err)
}

// ============================================================================
// ActionBoardConfigureColumns Tests
// ============================================================================

func TestBoardConfigHandler_ConfigureColumns_Success(t *testing.T) {
	handler, db := setupBoardConfigTestHandler(t)
	defer db.Close()

	// Create a test board
	boardID := uuid.New().String()
	createTestBoard(t, db, boardID, "Test Board", "scrum")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionBoardConfigureColumns,
		Data: map[string]interface{}{
			"boardId": boardID,
			"columns": []interface{}{
				map[string]interface{}{
					"title":    "To Do",
					"statusId": "status-1",
					"maxItems": float64(10),
				},
				map[string]interface{}{
					"title":    "In Progress",
					"statusId": "status-2",
				},
				map[string]interface{}{
					"title": "Done",
				},
			},
		},
	}

	handler.handleBoardConfigureColumns(c, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.True(t, resp.Data["configured"].(bool))
	assert.Equal(t, boardID, resp.Data["boardId"])
	assert.Equal(t, float64(3), resp.Data["columnCount"])
}

func TestBoardConfigHandler_ConfigureColumns_MissingBoardID(t *testing.T) {
	handler, db := setupBoardConfigTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionBoardConfigureColumns,
		Data:   map[string]interface{}{},
	}

	handler.handleBoardConfigureColumns(c, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeMissingData, resp.ErrorCode)
}

func TestBoardConfigHandler_ConfigureColumns_BoardNotFound(t *testing.T) {
	handler, db := setupBoardConfigTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionBoardConfigureColumns,
		Data: map[string]interface{}{
			"boardId": "nonexistent-id",
			"columns": []interface{}{},
		},
	}

	handler.handleBoardConfigureColumns(c, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestBoardConfigHandler_ConfigureColumns_MissingColumns(t *testing.T) {
	handler, db := setupBoardConfigTestHandler(t)
	defer db.Close()

	boardID := uuid.New().String()
	createTestBoard(t, db, boardID, "Test Board", "scrum")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionBoardConfigureColumns,
		Data: map[string]interface{}{
			"boardId": boardID,
		},
	}

	handler.handleBoardConfigureColumns(c, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestBoardConfigHandler_ConfigureColumns_Unauthorized(t *testing.T) {
	handler, db := setupBoardConfigTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	// No username set

	req := &models.Request{
		Action: models.ActionBoardConfigureColumns,
		Data:   map[string]interface{}{},
	}

	handler.handleBoardConfigureColumns(c, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// ============================================================================
// ActionBoardAddColumn Tests
// ============================================================================

func TestBoardConfigHandler_AddColumn_Success(t *testing.T) {
	handler, db := setupBoardConfigTestHandler(t)
	defer db.Close()

	boardID := uuid.New().String()
	createTestBoard(t, db, boardID, "Test Board", "scrum")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionBoardAddColumn,
		Data: map[string]interface{}{
			"boardId":  boardID,
			"title":    "To Do",
			"statusId": "status-1",
			"position": float64(0),
			"maxItems": float64(10),
		},
	}

	handler.handleBoardAddColumn(c, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.True(t, resp.Data["added"].(bool))
	assert.NotEmpty(t, resp.Data["columnId"])
	assert.Equal(t, boardID, resp.Data["boardId"])
}

func TestBoardConfigHandler_AddColumn_MissingBoardID(t *testing.T) {
	handler, db := setupBoardConfigTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionBoardAddColumn,
		Data: map[string]interface{}{
			"title": "To Do",
		},
	}

	handler.handleBoardAddColumn(c, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestBoardConfigHandler_AddColumn_MissingTitle(t *testing.T) {
	handler, db := setupBoardConfigTestHandler(t)
	defer db.Close()

	boardID := uuid.New().String()
	createTestBoard(t, db, boardID, "Test Board", "scrum")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionBoardAddColumn,
		Data: map[string]interface{}{
			"boardId": boardID,
		},
	}

	handler.handleBoardAddColumn(c, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestBoardConfigHandler_AddColumn_Unauthorized(t *testing.T) {
	handler, db := setupBoardConfigTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	// No username set

	req := &models.Request{
		Action: models.ActionBoardAddColumn,
		Data:   map[string]interface{}{},
	}

	handler.handleBoardAddColumn(c, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// ============================================================================
// ActionBoardRemoveColumn Tests
// ============================================================================

func TestBoardConfigHandler_RemoveColumn_Success(t *testing.T) {
	handler, db := setupBoardConfigTestHandler(t)
	defer db.Close()

	boardID := uuid.New().String()
	createTestBoard(t, db, boardID, "Test Board", "scrum")

	// Create a column
	columnID := uuid.New().String()
	_, err := db.Exec(context.Background(),
		"INSERT INTO board_column (id, board_id, title, position, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?)",
		columnID, boardID, "To Do", 0, time.Now().Unix(), time.Now().Unix(), false)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionBoardRemoveColumn,
		Data: map[string]interface{}{
			"columnId": columnID,
		},
	}

	handler.handleBoardRemoveColumn(c, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err = json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.True(t, resp.Data["removed"].(bool))
	assert.Equal(t, columnID, resp.Data["columnId"])
}

func TestBoardConfigHandler_RemoveColumn_MissingColumnID(t *testing.T) {
	handler, db := setupBoardConfigTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionBoardRemoveColumn,
		Data:   map[string]interface{}{},
	}

	handler.handleBoardRemoveColumn(c, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestBoardConfigHandler_RemoveColumn_NotFound(t *testing.T) {
	handler, db := setupBoardConfigTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionBoardRemoveColumn,
		Data: map[string]interface{}{
			"columnId": "nonexistent-id",
		},
	}

	handler.handleBoardRemoveColumn(c, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestBoardConfigHandler_RemoveColumn_Unauthorized(t *testing.T) {
	handler, db := setupBoardConfigTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	// No username set

	req := &models.Request{
		Action: models.ActionBoardRemoveColumn,
		Data:   map[string]interface{}{},
	}

	handler.handleBoardRemoveColumn(c, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// ============================================================================
// ActionBoardModifyColumn Tests
// ============================================================================

func TestBoardConfigHandler_ModifyColumn_Success(t *testing.T) {
	handler, db := setupBoardConfigTestHandler(t)
	defer db.Close()

	boardID := uuid.New().String()
	createTestBoard(t, db, boardID, "Test Board", "scrum")

	// Create a column
	columnID := uuid.New().String()
	_, err := db.Exec(context.Background(),
		"INSERT INTO board_column (id, board_id, title, position, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?)",
		columnID, boardID, "Old Title", 0, time.Now().Unix(), time.Now().Unix(), false)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionBoardModifyColumn,
		Data: map[string]interface{}{
			"columnId": columnID,
			"title":    "New Title",
			"position": float64(1),
			"maxItems": float64(5),
		},
	}

	handler.handleBoardModifyColumn(c, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err = json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.True(t, resp.Data["updated"].(bool))
	assert.Equal(t, columnID, resp.Data["columnId"])
}

func TestBoardConfigHandler_ModifyColumn_OnlyTitle(t *testing.T) {
	handler, db := setupBoardConfigTestHandler(t)
	defer db.Close()

	boardID := uuid.New().String()
	createTestBoard(t, db, boardID, "Test Board", "scrum")

	columnID := uuid.New().String()
	_, err := db.Exec(context.Background(),
		"INSERT INTO board_column (id, board_id, title, position, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?)",
		columnID, boardID, "Old Title", 0, time.Now().Unix(), time.Now().Unix(), false)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionBoardModifyColumn,
		Data: map[string]interface{}{
			"columnId": columnID,
			"title":    "Updated Title",
		},
	}

	handler.handleBoardModifyColumn(c, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestBoardConfigHandler_ModifyColumn_MissingColumnID(t *testing.T) {
	handler, db := setupBoardConfigTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionBoardModifyColumn,
		Data: map[string]interface{}{
			"title": "New Title",
		},
	}

	handler.handleBoardModifyColumn(c, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestBoardConfigHandler_ModifyColumn_NotFound(t *testing.T) {
	handler, db := setupBoardConfigTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionBoardModifyColumn,
		Data: map[string]interface{}{
			"columnId": "nonexistent-id",
			"title":    "New Title",
		},
	}

	handler.handleBoardModifyColumn(c, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestBoardConfigHandler_ModifyColumn_NoFieldsToUpdate(t *testing.T) {
	handler, db := setupBoardConfigTestHandler(t)
	defer db.Close()

	boardID := uuid.New().String()
	createTestBoard(t, db, boardID, "Test Board", "scrum")

	columnID := uuid.New().String()
	_, err := db.Exec(context.Background(),
		"INSERT INTO board_column (id, board_id, title, position, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?)",
		columnID, boardID, "Title", 0, time.Now().Unix(), time.Now().Unix(), false)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionBoardModifyColumn,
		Data: map[string]interface{}{
			"columnId": columnID,
		},
	}

	handler.handleBoardModifyColumn(c, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestBoardConfigHandler_ModifyColumn_Unauthorized(t *testing.T) {
	handler, db := setupBoardConfigTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	// No username set

	req := &models.Request{
		Action: models.ActionBoardModifyColumn,
		Data:   map[string]interface{}{},
	}

	handler.handleBoardModifyColumn(c, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// ============================================================================
// ActionBoardListColumns Tests
// ============================================================================

func TestBoardConfigHandler_ListColumns_Success(t *testing.T) {
	handler, db := setupBoardConfigTestHandler(t)
	defer db.Close()

	boardID := uuid.New().String()
	createTestBoard(t, db, boardID, "Test Board", "scrum")

	// Create multiple columns
	for i := 0; i < 3; i++ {
		columnID := uuid.New().String()
		_, err := db.Exec(context.Background(),
			"INSERT INTO board_column (id, board_id, title, position, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?)",
			columnID, boardID, "Column "+string(rune(i+1)), i, time.Now().Unix(), time.Now().Unix(), false)
		require.NoError(t, err)
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionBoardListColumns,
		Data: map[string]interface{}{
			"boardId": boardID,
		},
	}

	handler.handleBoardListColumns(c, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.Equal(t, boardID, resp.Data["boardId"])
	assert.Equal(t, float64(3), resp.Data["count"])

	columns := resp.Data["columns"].([]interface{})
	assert.Len(t, columns, 3)
}

func TestBoardConfigHandler_ListColumns_Empty(t *testing.T) {
	handler, db := setupBoardConfigTestHandler(t)
	defer db.Close()

	boardID := uuid.New().String()
	createTestBoard(t, db, boardID, "Test Board", "scrum")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionBoardListColumns,
		Data: map[string]interface{}{
			"boardId": boardID,
		},
	}

	handler.handleBoardListColumns(c, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, float64(0), resp.Data["count"])
}

func TestBoardConfigHandler_ListColumns_MissingBoardID(t *testing.T) {
	handler, db := setupBoardConfigTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionBoardListColumns,
		Data:   map[string]interface{}{},
	}

	handler.handleBoardListColumns(c, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestBoardConfigHandler_ListColumns_Unauthorized(t *testing.T) {
	handler, db := setupBoardConfigTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	// No username set

	req := &models.Request{
		Action: models.ActionBoardListColumns,
		Data:   map[string]interface{}{},
	}

	handler.handleBoardListColumns(c, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// ============================================================================
// ActionBoardAddSwimlane Tests
// ============================================================================

func TestBoardConfigHandler_AddSwimlane_Success(t *testing.T) {
	handler, db := setupBoardConfigTestHandler(t)
	defer db.Close()

	boardID := uuid.New().String()
	createTestBoard(t, db, boardID, "Test Board", "scrum")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionBoardAddSwimlane,
		Data: map[string]interface{}{
			"boardId":  boardID,
			"title":    "My Swimlane",
			"query":    "status:open",
			"position": float64(0),
		},
	}

	handler.handleBoardAddSwimlane(c, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.True(t, resp.Data["added"].(bool))
	assert.NotEmpty(t, resp.Data["swimlaneId"])
	assert.Equal(t, boardID, resp.Data["boardId"])
}

func TestBoardConfigHandler_AddSwimlane_MissingBoardID(t *testing.T) {
	handler, db := setupBoardConfigTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionBoardAddSwimlane,
		Data: map[string]interface{}{
			"title": "My Swimlane",
		},
	}

	handler.handleBoardAddSwimlane(c, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestBoardConfigHandler_AddSwimlane_MissingTitle(t *testing.T) {
	handler, db := setupBoardConfigTestHandler(t)
	defer db.Close()

	boardID := uuid.New().String()
	createTestBoard(t, db, boardID, "Test Board", "scrum")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionBoardAddSwimlane,
		Data: map[string]interface{}{
			"boardId": boardID,
		},
	}

	handler.handleBoardAddSwimlane(c, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestBoardConfigHandler_AddSwimlane_Unauthorized(t *testing.T) {
	handler, db := setupBoardConfigTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	// No username set

	req := &models.Request{
		Action: models.ActionBoardAddSwimlane,
		Data:   map[string]interface{}{},
	}

	handler.handleBoardAddSwimlane(c, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// ============================================================================
// ActionBoardRemoveSwimlane Tests
// ============================================================================

func TestBoardConfigHandler_RemoveSwimlane_Success(t *testing.T) {
	handler, db := setupBoardConfigTestHandler(t)
	defer db.Close()

	boardID := uuid.New().String()
	createTestBoard(t, db, boardID, "Test Board", "scrum")

	// Create a swimlane
	swimlaneID := uuid.New().String()
	_, err := db.Exec(context.Background(),
		"INSERT INTO board_swimlane (id, board_id, title, position, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?)",
		swimlaneID, boardID, "My Swimlane", 0, time.Now().Unix(), time.Now().Unix(), false)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionBoardRemoveSwimlane,
		Data: map[string]interface{}{
			"swimlaneId": swimlaneID,
		},
	}

	handler.handleBoardRemoveSwimlane(c, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err = json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.True(t, resp.Data["removed"].(bool))
	assert.Equal(t, swimlaneID, resp.Data["swimlaneId"])
}

func TestBoardConfigHandler_RemoveSwimlane_MissingSwimlaneID(t *testing.T) {
	handler, db := setupBoardConfigTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionBoardRemoveSwimlane,
		Data:   map[string]interface{}{},
	}

	handler.handleBoardRemoveSwimlane(c, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestBoardConfigHandler_RemoveSwimlane_NotFound(t *testing.T) {
	handler, db := setupBoardConfigTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionBoardRemoveSwimlane,
		Data: map[string]interface{}{
			"swimlaneId": "nonexistent-id",
		},
	}

	handler.handleBoardRemoveSwimlane(c, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestBoardConfigHandler_RemoveSwimlane_Unauthorized(t *testing.T) {
	handler, db := setupBoardConfigTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	// No username set

	req := &models.Request{
		Action: models.ActionBoardRemoveSwimlane,
		Data:   map[string]interface{}{},
	}

	handler.handleBoardRemoveSwimlane(c, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// ============================================================================
// ActionBoardListSwimlanes Tests
// ============================================================================

func TestBoardConfigHandler_ListSwimlanes_Success(t *testing.T) {
	handler, db := setupBoardConfigTestHandler(t)
	defer db.Close()

	boardID := uuid.New().String()
	createTestBoard(t, db, boardID, "Test Board", "scrum")

	// Create multiple swimlanes
	for i := 0; i < 3; i++ {
		swimlaneID := uuid.New().String()
		_, err := db.Exec(context.Background(),
			"INSERT INTO board_swimlane (id, board_id, title, position, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?)",
			swimlaneID, boardID, "Swimlane "+string(rune(i+1)), i, time.Now().Unix(), time.Now().Unix(), false)
		require.NoError(t, err)
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionBoardListSwimlanes,
		Data: map[string]interface{}{
			"boardId": boardID,
		},
	}

	handler.handleBoardListSwimlanes(c, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.Equal(t, boardID, resp.Data["boardId"])
	assert.Equal(t, float64(3), resp.Data["count"])

	swimlanes := resp.Data["swimlanes"].([]interface{})
	assert.Len(t, swimlanes, 3)
}

func TestBoardConfigHandler_ListSwimlanes_Empty(t *testing.T) {
	handler, db := setupBoardConfigTestHandler(t)
	defer db.Close()

	boardID := uuid.New().String()
	createTestBoard(t, db, boardID, "Test Board", "scrum")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionBoardListSwimlanes,
		Data: map[string]interface{}{
			"boardId": boardID,
		},
	}

	handler.handleBoardListSwimlanes(c, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, float64(0), resp.Data["count"])
}

func TestBoardConfigHandler_ListSwimlanes_MissingBoardID(t *testing.T) {
	handler, db := setupBoardConfigTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionBoardListSwimlanes,
		Data:   map[string]interface{}{},
	}

	handler.handleBoardListSwimlanes(c, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestBoardConfigHandler_ListSwimlanes_Unauthorized(t *testing.T) {
	handler, db := setupBoardConfigTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	// No username set

	req := &models.Request{
		Action: models.ActionBoardListSwimlanes,
		Data:   map[string]interface{}{},
	}

	handler.handleBoardListSwimlanes(c, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// ============================================================================
// ActionBoardAddQuickFilter Tests
// ============================================================================

func TestBoardConfigHandler_AddQuickFilter_Success(t *testing.T) {
	handler, db := setupBoardConfigTestHandler(t)
	defer db.Close()

	boardID := uuid.New().String()
	createTestBoard(t, db, boardID, "Test Board", "scrum")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionBoardAddQuickFilter,
		Data: map[string]interface{}{
			"boardId":  boardID,
			"title":    "My Filter",
			"query":    "assignee:me",
			"position": float64(0),
		},
	}

	handler.handleBoardAddQuickFilter(c, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.True(t, resp.Data["added"].(bool))
	assert.NotEmpty(t, resp.Data["filterId"])
	assert.Equal(t, boardID, resp.Data["boardId"])
}

func TestBoardConfigHandler_AddQuickFilter_MissingBoardID(t *testing.T) {
	handler, db := setupBoardConfigTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionBoardAddQuickFilter,
		Data: map[string]interface{}{
			"title": "My Filter",
		},
	}

	handler.handleBoardAddQuickFilter(c, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestBoardConfigHandler_AddQuickFilter_MissingTitle(t *testing.T) {
	handler, db := setupBoardConfigTestHandler(t)
	defer db.Close()

	boardID := uuid.New().String()
	createTestBoard(t, db, boardID, "Test Board", "scrum")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionBoardAddQuickFilter,
		Data: map[string]interface{}{
			"boardId": boardID,
		},
	}

	handler.handleBoardAddQuickFilter(c, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestBoardConfigHandler_AddQuickFilter_Unauthorized(t *testing.T) {
	handler, db := setupBoardConfigTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	// No username set

	req := &models.Request{
		Action: models.ActionBoardAddQuickFilter,
		Data:   map[string]interface{}{},
	}

	handler.handleBoardAddQuickFilter(c, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// ============================================================================
// ActionBoardRemoveQuickFilter Tests
// ============================================================================

func TestBoardConfigHandler_RemoveQuickFilter_Success(t *testing.T) {
	handler, db := setupBoardConfigTestHandler(t)
	defer db.Close()

	boardID := uuid.New().String()
	createTestBoard(t, db, boardID, "Test Board", "scrum")

	// Create a quick filter
	filterID := uuid.New().String()
	_, err := db.Exec(context.Background(),
		"INSERT INTO board_quick_filter (id, board_id, title, position, created, deleted) VALUES (?, ?, ?, ?, ?, ?)",
		filterID, boardID, "My Filter", 0, time.Now().Unix(), false)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionBoardRemoveQuickFilter,
		Data: map[string]interface{}{
			"filterId": filterID,
		},
	}

	handler.handleBoardRemoveQuickFilter(c, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err = json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.True(t, resp.Data["removed"].(bool))
	assert.Equal(t, filterID, resp.Data["filterId"])
}

func TestBoardConfigHandler_RemoveQuickFilter_MissingFilterID(t *testing.T) {
	handler, db := setupBoardConfigTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionBoardRemoveQuickFilter,
		Data:   map[string]interface{}{},
	}

	handler.handleBoardRemoveQuickFilter(c, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestBoardConfigHandler_RemoveQuickFilter_NotFound(t *testing.T) {
	handler, db := setupBoardConfigTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionBoardRemoveQuickFilter,
		Data: map[string]interface{}{
			"filterId": "nonexistent-id",
		},
	}

	handler.handleBoardRemoveQuickFilter(c, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestBoardConfigHandler_RemoveQuickFilter_Unauthorized(t *testing.T) {
	handler, db := setupBoardConfigTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	// No username set

	req := &models.Request{
		Action: models.ActionBoardRemoveQuickFilter,
		Data:   map[string]interface{}{},
	}

	handler.handleBoardRemoveQuickFilter(c, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// ============================================================================
// ActionBoardListQuickFilters Tests
// ============================================================================

func TestBoardConfigHandler_ListQuickFilters_Success(t *testing.T) {
	handler, db := setupBoardConfigTestHandler(t)
	defer db.Close()

	boardID := uuid.New().String()
	createTestBoard(t, db, boardID, "Test Board", "scrum")

	// Create multiple quick filters
	for i := 0; i < 3; i++ {
		filterID := uuid.New().String()
		_, err := db.Exec(context.Background(),
			"INSERT INTO board_quick_filter (id, board_id, title, position, created, deleted) VALUES (?, ?, ?, ?, ?, ?)",
			filterID, boardID, "Filter "+string(rune(i+1)), i, time.Now().Unix(), false)
		require.NoError(t, err)
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionBoardListQuickFilters,
		Data: map[string]interface{}{
			"boardId": boardID,
		},
	}

	handler.handleBoardListQuickFilters(c, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.Equal(t, boardID, resp.Data["boardId"])
	assert.Equal(t, float64(3), resp.Data["count"])

	filters := resp.Data["filters"].([]interface{})
	assert.Len(t, filters, 3)
}

func TestBoardConfigHandler_ListQuickFilters_Empty(t *testing.T) {
	handler, db := setupBoardConfigTestHandler(t)
	defer db.Close()

	boardID := uuid.New().String()
	createTestBoard(t, db, boardID, "Test Board", "scrum")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionBoardListQuickFilters,
		Data: map[string]interface{}{
			"boardId": boardID,
		},
	}

	handler.handleBoardListQuickFilters(c, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, float64(0), resp.Data["count"])
}

func TestBoardConfigHandler_ListQuickFilters_MissingBoardID(t *testing.T) {
	handler, db := setupBoardConfigTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionBoardListQuickFilters,
		Data:   map[string]interface{}{},
	}

	handler.handleBoardListQuickFilters(c, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestBoardConfigHandler_ListQuickFilters_Unauthorized(t *testing.T) {
	handler, db := setupBoardConfigTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	// No username set

	req := &models.Request{
		Action: models.ActionBoardListQuickFilters,
		Data:   map[string]interface{}{},
	}

	handler.handleBoardListQuickFilters(c, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// ============================================================================
// ActionBoardSetType Tests
// ============================================================================

func TestBoardConfigHandler_SetType_Success_Scrum(t *testing.T) {
	handler, db := setupBoardConfigTestHandler(t)
	defer db.Close()

	boardID := uuid.New().String()
	createTestBoard(t, db, boardID, "Test Board", "kanban")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionBoardSetType,
		Data: map[string]interface{}{
			"boardId":   boardID,
			"boardType": "scrum",
		},
	}

	handler.handleBoardSetType(c, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.True(t, resp.Data["updated"].(bool))
	assert.Equal(t, boardID, resp.Data["boardId"])
	assert.Equal(t, "scrum", resp.Data["boardType"])
}

func TestBoardConfigHandler_SetType_Success_Kanban(t *testing.T) {
	handler, db := setupBoardConfigTestHandler(t)
	defer db.Close()

	boardID := uuid.New().String()
	createTestBoard(t, db, boardID, "Test Board", "scrum")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionBoardSetType,
		Data: map[string]interface{}{
			"boardId":   boardID,
			"boardType": "kanban",
		},
	}

	handler.handleBoardSetType(c, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.Equal(t, "kanban", resp.Data["boardType"])
}

func TestBoardConfigHandler_SetType_MissingBoardID(t *testing.T) {
	handler, db := setupBoardConfigTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionBoardSetType,
		Data: map[string]interface{}{
			"boardType": "scrum",
		},
	}

	handler.handleBoardSetType(c, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestBoardConfigHandler_SetType_InvalidType(t *testing.T) {
	handler, db := setupBoardConfigTestHandler(t)
	defer db.Close()

	boardID := uuid.New().String()
	createTestBoard(t, db, boardID, "Test Board", "scrum")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionBoardSetType,
		Data: map[string]interface{}{
			"boardId":   boardID,
			"boardType": "invalid",
		},
	}

	handler.handleBoardSetType(c, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeInvalidData, resp.ErrorCode)
}

func TestBoardConfigHandler_SetType_BoardNotFound(t *testing.T) {
	handler, db := setupBoardConfigTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionBoardSetType,
		Data: map[string]interface{}{
			"boardId":   "nonexistent-id",
			"boardType": "scrum",
		},
	}

	handler.handleBoardSetType(c, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestBoardConfigHandler_SetType_Unauthorized(t *testing.T) {
	handler, db := setupBoardConfigTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	// No username set

	req := &models.Request{
		Action: models.ActionBoardSetType,
		Data:   map[string]interface{}{},
	}

	handler.handleBoardSetType(c, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
