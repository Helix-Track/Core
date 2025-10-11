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

// setupCycleTestHandler creates a test handler with cycle test data
func setupCycleTestHandler(t *testing.T) *Handler {
	handler := setupTestHandler(t)
	ctx := context.Background()

	// Create cycle table
	_, err := handler.db.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS cycle (
			id          TEXT    NOT NULL PRIMARY KEY UNIQUE,
			title       TEXT    NOT NULL,
			description TEXT,
			type        INTEGER NOT NULL,
			cycle_id    TEXT,
			started     INTEGER,
			ended       INTEGER,
			created     INTEGER NOT NULL,
			modified    INTEGER NOT NULL,
			deleted     BOOLEAN NOT NULL
		)
	`)
	require.NoError(t, err)

	// Create project table
	_, err = handler.db.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS project (
			id          TEXT    NOT NULL PRIMARY KEY UNIQUE,
			identifier  TEXT    NOT NULL UNIQUE,
			title       TEXT    NOT NULL,
			description TEXT,
			workflow_id TEXT    NOT NULL,
			created     INTEGER NOT NULL,
			modified    INTEGER NOT NULL,
			deleted     BOOLEAN NOT NULL
		)
	`)
	require.NoError(t, err)

	// Create ticket table
	_, err = handler.db.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS ticket (
			id               TEXT    NOT NULL PRIMARY KEY UNIQUE,
			ticket_number    INTEGER NOT NULL,
			position         INTEGER NOT NULL,
			title            TEXT,
			description      TEXT,
			created          INTEGER NOT NULL,
			modified         INTEGER NOT NULL,
			ticket_type_id   TEXT    NOT NULL,
			ticket_status_id TEXT    NOT NULL,
			project_id       TEXT    NOT NULL,
			user_id          TEXT,
			estimation       REAL    NOT NULL,
			story_points     INTEGER NOT NULL,
			creator          TEXT    NOT NULL,
			deleted          BOOLEAN NOT NULL,
			UNIQUE (ticket_number, project_id)
		)
	`)
	require.NoError(t, err)

	// Create cycle_project_mapping table
	_, err = handler.db.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS cycle_project_mapping (
			id         TEXT    NOT NULL PRIMARY KEY UNIQUE,
			cycle_id   TEXT    NOT NULL,
			project_id TEXT    NOT NULL,
			created    INTEGER NOT NULL,
			modified   INTEGER NOT NULL,
			deleted    BOOLEAN NOT NULL,
			UNIQUE (cycle_id, project_id)
		)
	`)
	require.NoError(t, err)

	// Create ticket_cycle_mapping table
	_, err = handler.db.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS ticket_cycle_mapping (
			id        TEXT    NOT NULL PRIMARY KEY UNIQUE,
			ticket_id TEXT    NOT NULL,
			cycle_id  TEXT    NOT NULL,
			created   INTEGER NOT NULL,
			modified  INTEGER NOT NULL,
			deleted   BOOLEAN NOT NULL,
			UNIQUE (ticket_id, cycle_id)
		)
	`)
	require.NoError(t, err)

	// Insert test project for cycle-project mappings
	_, err = handler.db.Exec(ctx,
		"INSERT INTO project (id, identifier, title, description, workflow_id, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		"test-project-id", "TEST", "Test Project", "Test project description", "workflow-1", 1000, 1000, 0)
	require.NoError(t, err)

	// Insert test ticket for cycle-ticket mappings
	_, err = handler.db.Exec(ctx,
		"INSERT INTO ticket (id, ticket_number, position, title, description, ticket_type_id, ticket_status_id, project_id, user_id, estimation, story_points, creator, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		"test-ticket-id", 1, 0, "Test Ticket", "Test ticket description", "type-1", "status-1", "test-project-id", "user-1", 0.0, 0, "testuser", 1000, 1000, 0)
	require.NoError(t, err)

	return handler
}

// TestCycleHandler_Create_Success tests successful cycle creation
func TestCycleHandler_Create_Success(t *testing.T) {
	handler := setupCycleTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionCycleCreate,
		Data: map[string]interface{}{
			"title":       "Sprint 1",
			"description": "First sprint",
			"type":        float64(models.CycleTypeSprint), // 10
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

	cycle, ok := resp.Data["cycle"].(map[string]interface{})
	require.True(t, ok)
	assert.NotEmpty(t, cycle["id"])
	assert.Equal(t, "Sprint 1", cycle["title"])
	assert.Equal(t, "First sprint", cycle["description"])
	assert.Equal(t, float64(models.CycleTypeSprint), cycle["type"])
}

// TestCycleHandler_Create_AllTypes tests creation of all cycle types
func TestCycleHandler_Create_AllTypes(t *testing.T) {
	handler := setupCycleTestHandler(t)

	testCases := []struct {
		name      string
		cycleType int
		title     string
	}{
		{"Sprint", models.CycleTypeSprint, "Sprint 1"},
		{"Milestone", models.CycleTypeMilestone, "Milestone 1"},
		{"Release", models.CycleTypeRelease, "Release 1.0"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			reqBody := models.Request{
				Action: models.ActionCycleCreate,
				Data: map[string]interface{}{
					"title": tc.title,
					"type":  float64(tc.cycleType),
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

			cycle, ok := resp.Data["cycle"].(map[string]interface{})
			require.True(t, ok)
			assert.Equal(t, tc.title, cycle["title"])
			assert.Equal(t, float64(tc.cycleType), cycle["type"])
		})
	}
}

// TestCycleHandler_Create_WithParent tests cycle creation with parent cycle
func TestCycleHandler_Create_WithParent(t *testing.T) {
	handler := setupCycleTestHandler(t)

	// Create parent cycle (Release)
	_, err := handler.db.Exec(context.Background(),
		"INSERT INTO cycle (id, title, type, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)",
		"parent-cycle-id", "Release 1.0", models.CycleTypeRelease, 1000, 1000, 0)
	require.NoError(t, err)

	// Create child cycle (Sprint) with parent
	reqBody := models.Request{
		Action: models.ActionCycleCreate,
		Data: map[string]interface{}{
			"title":   "Sprint 1",
			"type":    float64(models.CycleTypeSprint),
			"cycleId": "parent-cycle-id",
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

	cycle, ok := resp.Data["cycle"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "parent-cycle-id", cycle["cycleId"])
}

// TestCycleHandler_Create_InvalidParentHierarchy tests invalid parent-child type hierarchy
func TestCycleHandler_Create_InvalidParentHierarchy(t *testing.T) {
	handler := setupCycleTestHandler(t)

	// Create parent cycle (Sprint - type 10)
	_, err := handler.db.Exec(context.Background(),
		"INSERT INTO cycle (id, title, type, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)",
		"parent-cycle-id", "Sprint 1", models.CycleTypeSprint, 1000, 1000, 0)
	require.NoError(t, err)

	// Try to create child cycle (Milestone - type 100) with Sprint parent - should fail
	// (Parent type must be greater than child type)
	reqBody := models.Request{
		Action: models.ActionCycleCreate,
		Data: map[string]interface{}{
			"title":   "Milestone 1",
			"type":    float64(models.CycleTypeMilestone),
			"cycleId": "parent-cycle-id",
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
	assert.Equal(t, models.ErrorCodeInvalidData, resp.ErrorCode)
}

// TestCycleHandler_Create_MissingTitle tests cycle creation with missing title
func TestCycleHandler_Create_MissingTitle(t *testing.T) {
	handler := setupCycleTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionCycleCreate,
		Data: map[string]interface{}{
			"type": float64(models.CycleTypeSprint),
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

// TestCycleHandler_Create_InvalidType tests cycle creation with invalid type
func TestCycleHandler_Create_InvalidType(t *testing.T) {
	handler := setupCycleTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionCycleCreate,
		Data: map[string]interface{}{
			"title": "Invalid Cycle",
			"type":  float64(999), // Invalid type
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
	assert.Equal(t, models.ErrorCodeInvalidData, resp.ErrorCode)
}

// TestCycleHandler_Read_Success tests successful cycle read
func TestCycleHandler_Read_Success(t *testing.T) {
	handler := setupCycleTestHandler(t)

	// Insert test cycle
	_, err := handler.db.Exec(context.Background(),
		"INSERT INTO cycle (id, title, description, type, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?)",
		"test-cycle-id", "Sprint 1", "First sprint", models.CycleTypeSprint, 1000, 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionCycleRead,
		Data: map[string]interface{}{
			"id": "test-cycle-id",
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

	cycle, ok := resp.Data["cycle"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "test-cycle-id", cycle["id"])
	assert.Equal(t, "Sprint 1", cycle["title"])
}

// TestCycleHandler_Read_NotFound tests reading non-existent cycle
func TestCycleHandler_Read_NotFound(t *testing.T) {
	handler := setupCycleTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionCycleRead,
		Data: map[string]interface{}{
			"id": "non-existent-cycle",
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

// TestCycleHandler_List_Empty tests listing cycles when none exist
func TestCycleHandler_List_Empty(t *testing.T) {
	handler := setupCycleTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionCycleList,
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

	cycles, ok := resp.Data["cycles"].([]interface{})
	require.True(t, ok)
	assert.Empty(t, cycles)
}

// TestCycleHandler_List_Multiple tests listing multiple cycles
func TestCycleHandler_List_Multiple(t *testing.T) {
	handler := setupCycleTestHandler(t)

	// Insert multiple cycles
	cycles := []struct {
		id    string
		title string
		typ   int
	}{
		{"cycle-1", "Sprint 1", models.CycleTypeSprint},
		{"cycle-2", "Milestone 1", models.CycleTypeMilestone},
		{"cycle-3", "Release 1.0", models.CycleTypeRelease},
	}

	for _, cycle := range cycles {
		_, err := handler.db.Exec(context.Background(),
			"INSERT INTO cycle (id, title, type, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)",
			cycle.id, cycle.title, cycle.typ, 1000, 1000, 0)
		require.NoError(t, err)
	}

	reqBody := models.Request{
		Action: models.ActionCycleList,
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

	cyclesList, ok := resp.Data["cycles"].([]interface{})
	require.True(t, ok)
	assert.Len(t, cyclesList, 3)
}

// TestCycleHandler_List_FilterByType tests listing cycles filtered by type
func TestCycleHandler_List_FilterByType(t *testing.T) {
	handler := setupCycleTestHandler(t)

	// Insert multiple cycles of different types
	_, err := handler.db.Exec(context.Background(),
		"INSERT INTO cycle (id, title, type, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)",
		"sprint-1", "Sprint 1", models.CycleTypeSprint, 1000, 1000, 0)
	require.NoError(t, err)

	_, err = handler.db.Exec(context.Background(),
		"INSERT INTO cycle (id, title, type, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)",
		"milestone-1", "Milestone 1", models.CycleTypeMilestone, 1000, 1000, 0)
	require.NoError(t, err)

	// Filter for sprints only
	reqBody := models.Request{
		Action: models.ActionCycleList,
		Data: map[string]interface{}{
			"type": float64(models.CycleTypeSprint),
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

	cycles, ok := resp.Data["cycles"].([]interface{})
	require.True(t, ok)
	assert.Len(t, cycles, 1)

	cycle := cycles[0].(map[string]interface{})
	assert.Equal(t, "Sprint 1", cycle["title"])
}

// TestCycleHandler_Modify_Success tests successful cycle modification
func TestCycleHandler_Modify_Success(t *testing.T) {
	handler := setupCycleTestHandler(t)

	// Insert test cycle
	_, err := handler.db.Exec(context.Background(),
		"INSERT INTO cycle (id, title, description, type, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?)",
		"test-cycle-id", "Sprint 1", "First sprint", models.CycleTypeSprint, 1000, 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionCycleModify,
		Data: map[string]interface{}{
			"id":          "test-cycle-id",
			"title":       "Sprint 1 - Updated",
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
	c.Set("request", &reqBody)

	handler.DoAction(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err = json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)

	// Verify update in database
	var title, description string
	err = handler.db.QueryRow(context.Background(),
		"SELECT title, description FROM cycle WHERE id = ?",
		"test-cycle-id").Scan(&title, &description)
	require.NoError(t, err)
	assert.Equal(t, "Sprint 1 - Updated", title)
	assert.Equal(t, "Updated description", description)
}

// TestCycleHandler_Modify_NotFound tests modifying non-existent cycle
func TestCycleHandler_Modify_NotFound(t *testing.T) {
	handler := setupCycleTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionCycleModify,
		Data: map[string]interface{}{
			"id":    "non-existent-cycle",
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

// TestCycleHandler_Remove_Success tests successful cycle deletion
func TestCycleHandler_Remove_Success(t *testing.T) {
	handler := setupCycleTestHandler(t)

	// Insert test cycle
	_, err := handler.db.Exec(context.Background(),
		"INSERT INTO cycle (id, title, type, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)",
		"test-cycle-id", "Sprint 1", models.CycleTypeSprint, 1000, 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionCycleRemove,
		Data: map[string]interface{}{
			"id": "test-cycle-id",
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
		"SELECT deleted FROM cycle WHERE id = ?",
		"test-cycle-id").Scan(&deleted)
	require.NoError(t, err)
	assert.True(t, deleted)
}

// TestCycleHandler_Remove_NotFound tests deleting non-existent cycle
func TestCycleHandler_Remove_NotFound(t *testing.T) {
	handler := setupCycleTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionCycleRemove,
		Data: map[string]interface{}{
			"id": "non-existent-cycle",
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

// TestCycleHandler_AssignProject_Success tests successful cycle-project assignment
func TestCycleHandler_AssignProject_Success(t *testing.T) {
	handler := setupCycleTestHandler(t)

	// Insert test cycle
	_, err := handler.db.Exec(context.Background(),
		"INSERT INTO cycle (id, title, type, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)",
		"test-cycle-id", "Sprint 1", models.CycleTypeSprint, 1000, 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionCycleAssignProject,
		Data: map[string]interface{}{
			"cycleId":   "test-cycle-id",
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

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp models.Response
	err = json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)

	mapping, ok := resp.Data["mapping"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "test-cycle-id", mapping["cycleId"])
	assert.Equal(t, "test-project-id", mapping["projectId"])
}

// TestCycleHandler_AssignProject_AlreadyAssigned tests assigning already assigned project
func TestCycleHandler_AssignProject_AlreadyAssigned(t *testing.T) {
	handler := setupCycleTestHandler(t)

	// Insert test cycle
	_, err := handler.db.Exec(context.Background(),
		"INSERT INTO cycle (id, title, type, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)",
		"test-cycle-id", "Sprint 1", models.CycleTypeSprint, 1000, 1000, 0)
	require.NoError(t, err)

	// Insert existing mapping
	_, err = handler.db.Exec(context.Background(),
		"INSERT INTO cycle_project_mapping (id, cycle_id, project_id, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)",
		"mapping-id", "test-cycle-id", "test-project-id", 1000, 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionCycleAssignProject,
		Data: map[string]interface{}{
			"cycleId":   "test-cycle-id",
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

	assert.Equal(t, http.StatusConflict, w.Code)

	var resp models.Response
	err = json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeEntityAlreadyExists, resp.ErrorCode)
}

// TestCycleHandler_UnassignProject_Success tests successful cycle-project unassignment
func TestCycleHandler_UnassignProject_Success(t *testing.T) {
	handler := setupCycleTestHandler(t)

	// Insert test cycle
	_, err := handler.db.Exec(context.Background(),
		"INSERT INTO cycle (id, title, type, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)",
		"test-cycle-id", "Sprint 1", models.CycleTypeSprint, 1000, 1000, 0)
	require.NoError(t, err)

	// Insert mapping
	_, err = handler.db.Exec(context.Background(),
		"INSERT INTO cycle_project_mapping (id, cycle_id, project_id, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)",
		"mapping-id", "test-cycle-id", "test-project-id", 1000, 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionCycleUnassignProject,
		Data: map[string]interface{}{
			"cycleId":   "test-cycle-id",
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
	err = json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)

	// Verify soft delete in database
	var deleted bool
	err = handler.db.QueryRow(context.Background(),
		"SELECT deleted FROM cycle_project_mapping WHERE cycle_id = ? AND project_id = ?",
		"test-cycle-id", "test-project-id").Scan(&deleted)
	require.NoError(t, err)
	assert.True(t, deleted)
}

// TestCycleHandler_ListProjects_Success tests listing projects for a cycle
func TestCycleHandler_ListProjects_Success(t *testing.T) {
	handler := setupCycleTestHandler(t)

	// Insert test cycle
	_, err := handler.db.Exec(context.Background(),
		"INSERT INTO cycle (id, title, type, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)",
		"test-cycle-id", "Sprint 1", models.CycleTypeSprint, 1000, 1000, 0)
	require.NoError(t, err)

	// Insert mapping
	_, err = handler.db.Exec(context.Background(),
		"INSERT INTO cycle_project_mapping (id, cycle_id, project_id, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)",
		"mapping-id", "test-cycle-id", "test-project-id", 1000, 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionCycleListProjects,
		Data: map[string]interface{}{
			"cycleId": "test-cycle-id",
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

	projects, ok := resp.Data["projects"].([]interface{})
	require.True(t, ok)
	assert.Len(t, projects, 1)

	project := projects[0].(map[string]interface{})
	assert.Equal(t, "test-project-id", project["id"])
	assert.Equal(t, "Test Project", project["title"])
}

// TestCycleHandler_AddTicket_Success tests successful ticket addition to cycle
func TestCycleHandler_AddTicket_Success(t *testing.T) {
	handler := setupCycleTestHandler(t)

	// Insert test cycle
	_, err := handler.db.Exec(context.Background(),
		"INSERT INTO cycle (id, title, type, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)",
		"test-cycle-id", "Sprint 1", models.CycleTypeSprint, 1000, 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionCycleAddTicket,
		Data: map[string]interface{}{
			"cycleId":  "test-cycle-id",
			"ticketId": "test-ticket-id",
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
	assert.Equal(t, "test-cycle-id", mapping["cycleId"])
	assert.Equal(t, "test-ticket-id", mapping["ticketId"])
}

// TestCycleHandler_RemoveTicket_Success tests successful ticket removal from cycle
func TestCycleHandler_RemoveTicket_Success(t *testing.T) {
	handler := setupCycleTestHandler(t)

	// Insert test cycle
	_, err := handler.db.Exec(context.Background(),
		"INSERT INTO cycle (id, title, type, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)",
		"test-cycle-id", "Sprint 1", models.CycleTypeSprint, 1000, 1000, 0)
	require.NoError(t, err)

	// Insert mapping
	_, err = handler.db.Exec(context.Background(),
		"INSERT INTO ticket_cycle_mapping (id, ticket_id, cycle_id, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)",
		"mapping-id", "test-ticket-id", "test-cycle-id", 1000, 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionCycleRemoveTicket,
		Data: map[string]interface{}{
			"cycleId":  "test-cycle-id",
			"ticketId": "test-ticket-id",
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
		"SELECT deleted FROM ticket_cycle_mapping WHERE ticket_id = ? AND cycle_id = ?",
		"test-ticket-id", "test-cycle-id").Scan(&deleted)
	require.NoError(t, err)
	assert.True(t, deleted)
}

// TestCycleHandler_ListTickets_Success tests listing tickets in a cycle
func TestCycleHandler_ListTickets_Success(t *testing.T) {
	handler := setupCycleTestHandler(t)

	// Insert test cycle
	_, err := handler.db.Exec(context.Background(),
		"INSERT INTO cycle (id, title, type, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)",
		"test-cycle-id", "Sprint 1", models.CycleTypeSprint, 1000, 1000, 0)
	require.NoError(t, err)

	// Insert mapping
	_, err = handler.db.Exec(context.Background(),
		"INSERT INTO ticket_cycle_mapping (id, ticket_id, cycle_id, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?)",
		"mapping-id", "test-ticket-id", "test-cycle-id", 1000, 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionCycleListTickets,
		Data: map[string]interface{}{
			"cycleId": "test-cycle-id",
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

	tickets, ok := resp.Data["tickets"].([]interface{})
	require.True(t, ok)
	assert.Len(t, tickets, 1)

	ticket := tickets[0].(map[string]interface{})
	assert.Equal(t, "test-ticket-id", ticket["id"])
	assert.Equal(t, "Test Ticket", ticket["title"])
}

// TestCycleHandler_CRUD_FullCycle tests complete cycle lifecycle
func TestCycleHandler_CRUD_FullCycle(t *testing.T) {
	handler := setupCycleTestHandler(t)

	// 1. Create cycle
	createReq := models.Request{
		Action: models.ActionCycleCreate,
		Data: map[string]interface{}{
			"title":       "Sprint 1",
			"description": "First sprint",
			"type":        float64(models.CycleTypeSprint),
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
	cycleData := createResp.Data["cycle"].(map[string]interface{})
	cycleID := cycleData["id"].(string)

	// 2. Read cycle
	readReq := models.Request{
		Action: models.ActionCycleRead,
		Data:   map[string]interface{}{"id": cycleID},
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

	// 3. Modify cycle
	modifyReq := models.Request{
		Action: models.ActionCycleModify,
		Data: map[string]interface{}{
			"id":    cycleID,
			"title": "Sprint 1 - Updated",
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

	// 4. Delete cycle
	deleteReq := models.Request{
		Action: models.ActionCycleRemove,
		Data:   map[string]interface{}{"id": cycleID},
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

	// 5. Verify deletion - cycle should not be found
	readReq = models.Request{
		Action: models.ActionCycleRead,
		Data:   map[string]interface{}{"id": cycleID},
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
