package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"helixtrack.ru/core/internal/models"
)

// setupCommentTestHandler creates a test handler with ticket data
func setupCommentTestHandler(t *testing.T) (*Handler, string, string) {
	handler, projectID := setupTicketTestHandler(t)

	// Create a ticket for comment testing
	createReq := models.Request{
		Action: models.ActionCreate,
		Object: "ticket",
		Data: map[string]interface{}{
			"project_id": projectID,
			"title":      "Test Ticket for Comments",
		},
	}
	createBody, _ := json.Marshal(createReq)
	createHttpReq := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(createBody))
	createHttpReq.Header.Set("Content-Type", "application/json")
	wCreate := httptest.NewRecorder()
	cCreate, _ := gin.CreateTestContext(wCreate)
	cCreate.Request = createHttpReq
	cCreate.Set("username", "testuser")
	handler.DoAction(cCreate)

	var createResp models.Response
	json.NewDecoder(wCreate.Body).Decode(&createResp)
	ticketID := createResp.Data["ticket"].(map[string]interface{})["id"].(string)

	return handler, ticketID, projectID
}

// =============================================================================
// handleCreateComment Tests
// =============================================================================

func TestCommentHandler_Create_Success(t *testing.T) {
	handler, ticketID, _ := setupCommentTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionCreate,
		Object: "comment",
		Data: map[string]interface{}{
			"ticket_id": ticketID,
			"comment":   "This is a test comment",
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
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)

	comment, ok := resp.Data["comment"].(map[string]interface{})
	require.True(t, ok)
	assert.NotEmpty(t, comment["id"])
	assert.Equal(t, "This is a test comment", comment["comment"])
	assert.Equal(t, ticketID, comment["ticket_id"])
	assert.NotEmpty(t, comment["created"])
}

func TestCommentHandler_Create_MissingTicketID(t *testing.T) {
	handler, _, _ := setupCommentTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionCreate,
		Object: "comment",
		Data: map[string]interface{}{
			"comment": "Test comment",
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
	assert.Contains(t, resp.ErrorMessage, "ticket_id")
}

func TestCommentHandler_Create_MissingComment(t *testing.T) {
	handler, ticketID, _ := setupCommentTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionCreate,
		Object: "comment",
		Data: map[string]interface{}{
			"ticket_id": ticketID,
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
	assert.Contains(t, resp.ErrorMessage, "comment text")
}

func TestCommentHandler_Create_MultipleComments(t *testing.T) {
	handler, ticketID, _ := setupCommentTestHandler(t)

	// Create 3 comments
	for i := 1; i <= 3; i++ {
		reqBody := models.Request{
			Action: models.ActionCreate,
			Object: "comment",
			Data: map[string]interface{}{
				"ticket_id": ticketID,
				"comment":   fmt.Sprintf("Comment %d", i),
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
	}

	// Verify all 3 comments exist by listing them
	listReq := models.Request{
		Action: models.ActionList,
		Object: "comment",
		Data: map[string]interface{}{
			"ticket_id": ticketID,
		},
	}
	listBody, _ := json.Marshal(listReq)

	req := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(listBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("username", "testuser")

	handler.DoAction(c)

	var listResp models.Response
	json.NewDecoder(w.Body).Decode(&listResp)
	items := listResp.Data["items"].([]interface{})
	assert.Equal(t, 3, len(items))
}

// =============================================================================
// handleModifyComment Tests
// =============================================================================

func TestCommentHandler_Modify_Success(t *testing.T) {
	handler, ticketID, _ := setupCommentTestHandler(t)

	// Create comment
	createReq := models.Request{
		Action: models.ActionCreate,
		Object: "comment",
		Data: map[string]interface{}{
			"ticket_id": ticketID,
			"comment":   "Original comment",
		},
	}
	createBody, _ := json.Marshal(createReq)
	createHttpReq := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(createBody))
	createHttpReq.Header.Set("Content-Type", "application/json")
	wCreate := httptest.NewRecorder()
	cCreate, _ := gin.CreateTestContext(wCreate)
	cCreate.Request = createHttpReq
	cCreate.Set("username", "testuser")
	handler.DoAction(cCreate)

	var createResp models.Response
	json.NewDecoder(wCreate.Body).Decode(&createResp)
	commentID := createResp.Data["comment"].(map[string]interface{})["id"].(string)

	// Modify comment
	modifyReq := models.Request{
		Action: models.ActionModify,
		Object: "comment",
		Data: map[string]interface{}{
			"id":      commentID,
			"comment": "Updated comment text",
		},
	}
	modifyBody, _ := json.Marshal(modifyReq)
	modifyHttpReq := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(modifyBody))
	modifyHttpReq.Header.Set("Content-Type", "application/json")
	wModify := httptest.NewRecorder()
	cModify, _ := gin.CreateTestContext(wModify)
	cModify.Request = modifyHttpReq
	cModify.Set("username", "testuser")
	handler.DoAction(cModify)

	assert.Equal(t, http.StatusOK, wModify.Code)

	var modifyResp models.Response
	err := json.NewDecoder(wModify.Body).Decode(&modifyResp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, modifyResp.ErrorCode)

	modifiedComment := modifyResp.Data["comment"].(map[string]interface{})
	assert.Equal(t, commentID, modifiedComment["id"])
	assert.True(t, modifiedComment["updated"].(bool))

	// Verify the comment was actually updated
	readReq := models.Request{
		Action: models.ActionRead,
		Object: "comment",
		Data: map[string]interface{}{
			"id": commentID,
		},
	}
	readBody, _ := json.Marshal(readReq)
	readHttpReq := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(readBody))
	readHttpReq.Header.Set("Content-Type", "application/json")
	wRead := httptest.NewRecorder()
	cRead, _ := gin.CreateTestContext(wRead)
	cRead.Request = readHttpReq
	cRead.Set("username", "testuser")
	handler.DoAction(cRead)

	var readResp models.Response
	json.NewDecoder(wRead.Body).Decode(&readResp)
	readComment := readResp.Data["comment"].(map[string]interface{})
	assert.Equal(t, "Updated comment text", readComment["comment"])
}

func TestCommentHandler_Modify_MissingID(t *testing.T) {
	handler, _, _ := setupCommentTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionModify,
		Object: "comment",
		Data: map[string]interface{}{
			"comment": "Updated comment",
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
	assert.Contains(t, resp.ErrorMessage, "comment ID")
}

func TestCommentHandler_Modify_MissingComment(t *testing.T) {
	handler, ticketID, _ := setupCommentTestHandler(t)

	// Create comment first
	createReq := models.Request{
		Action: models.ActionCreate,
		Object: "comment",
		Data: map[string]interface{}{
			"ticket_id": ticketID,
			"comment":   "Original",
		},
	}
	createBody, _ := json.Marshal(createReq)
	createHttpReq := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(createBody))
	createHttpReq.Header.Set("Content-Type", "application/json")
	wCreate := httptest.NewRecorder()
	cCreate, _ := gin.CreateTestContext(wCreate)
	cCreate.Request = createHttpReq
	cCreate.Set("username", "testuser")
	handler.DoAction(cCreate)

	var createResp models.Response
	json.NewDecoder(wCreate.Body).Decode(&createResp)
	commentID := createResp.Data["comment"].(map[string]interface{})["id"].(string)

	// Try to modify without comment text
	reqBody := models.Request{
		Action: models.ActionModify,
		Object: "comment",
		Data: map[string]interface{}{
			"id": commentID,
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
	assert.Contains(t, resp.ErrorMessage, "comment text")
}

// =============================================================================
// handleRemoveComment Tests
// =============================================================================

func TestCommentHandler_Remove_Success(t *testing.T) {
	handler, ticketID, _ := setupCommentTestHandler(t)

	// Create comment
	createReq := models.Request{
		Action: models.ActionCreate,
		Object: "comment",
		Data: map[string]interface{}{
			"ticket_id": ticketID,
			"comment":   "To be deleted",
		},
	}
	createBody, _ := json.Marshal(createReq)
	createHttpReq := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(createBody))
	createHttpReq.Header.Set("Content-Type", "application/json")
	wCreate := httptest.NewRecorder()
	cCreate, _ := gin.CreateTestContext(wCreate)
	cCreate.Request = createHttpReq
	cCreate.Set("username", "testuser")
	handler.DoAction(cCreate)

	var createResp models.Response
	json.NewDecoder(wCreate.Body).Decode(&createResp)
	commentID := createResp.Data["comment"].(map[string]interface{})["id"].(string)

	// Remove comment
	removeReq := models.Request{
		Action: models.ActionRemove,
		Object: "comment",
		Data: map[string]interface{}{
			"id": commentID,
		},
	}
	removeBody, _ := json.Marshal(removeReq)
	removeHttpReq := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(removeBody))
	removeHttpReq.Header.Set("Content-Type", "application/json")
	wRemove := httptest.NewRecorder()
	cRemove, _ := gin.CreateTestContext(wRemove)
	cRemove.Request = removeHttpReq
	cRemove.Set("username", "testuser")
	handler.DoAction(cRemove)

	assert.Equal(t, http.StatusOK, wRemove.Code)

	var removeResp models.Response
	err := json.NewDecoder(wRemove.Body).Decode(&removeResp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, removeResp.ErrorCode)

	removedComment := removeResp.Data["comment"].(map[string]interface{})
	assert.Equal(t, commentID, removedComment["id"])
	assert.True(t, removedComment["deleted"].(bool))
}

func TestCommentHandler_Remove_MissingID(t *testing.T) {
	handler, _, _ := setupCommentTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionRemove,
		Object: "comment",
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

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeMissingData, resp.ErrorCode)
	assert.Contains(t, resp.ErrorMessage, "comment ID")
}

// =============================================================================
// handleReadComment Tests
// =============================================================================

func TestCommentHandler_Read_Success(t *testing.T) {
	handler, ticketID, _ := setupCommentTestHandler(t)

	// Create comment
	createReq := models.Request{
		Action: models.ActionCreate,
		Object: "comment",
		Data: map[string]interface{}{
			"ticket_id": ticketID,
			"comment":   "Read test comment",
		},
	}
	createBody, _ := json.Marshal(createReq)
	createHttpReq := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(createBody))
	createHttpReq.Header.Set("Content-Type", "application/json")
	wCreate := httptest.NewRecorder()
	cCreate, _ := gin.CreateTestContext(wCreate)
	cCreate.Request = createHttpReq
	cCreate.Set("username", "testuser")
	handler.DoAction(cCreate)

	var createResp models.Response
	json.NewDecoder(wCreate.Body).Decode(&createResp)
	commentID := createResp.Data["comment"].(map[string]interface{})["id"].(string)

	// Read comment
	readReq := models.Request{
		Action: models.ActionRead,
		Object: "comment",
		Data: map[string]interface{}{
			"id": commentID,
		},
	}
	readBody, _ := json.Marshal(readReq)
	readHttpReq := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(readBody))
	readHttpReq.Header.Set("Content-Type", "application/json")
	wRead := httptest.NewRecorder()
	cRead, _ := gin.CreateTestContext(wRead)
	cRead.Request = readHttpReq
	cRead.Set("username", "testuser")
	handler.DoAction(cRead)

	assert.Equal(t, http.StatusOK, wRead.Code)

	var readResp models.Response
	err := json.NewDecoder(wRead.Body).Decode(&readResp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, readResp.ErrorCode)

	comment := readResp.Data["comment"].(map[string]interface{})
	assert.Equal(t, commentID, comment["id"])
	assert.Equal(t, "Read test comment", comment["comment"])
	assert.NotEmpty(t, comment["created"])
	assert.NotEmpty(t, comment["modified"])
}

func TestCommentHandler_Read_MissingID(t *testing.T) {
	handler, _, _ := setupCommentTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionRead,
		Object: "comment",
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

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeMissingData, resp.ErrorCode)
	assert.Contains(t, resp.ErrorMessage, "comment ID")
}

func TestCommentHandler_Read_NotFound(t *testing.T) {
	handler, _, _ := setupCommentTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionRead,
		Object: "comment",
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

	handler.DoAction(c)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeEntityNotFound, resp.ErrorCode)
	assert.Contains(t, resp.ErrorMessage, "not found")
}

func TestCommentHandler_Read_DeletedComment(t *testing.T) {
	handler, ticketID, _ := setupCommentTestHandler(t)

	// Create and delete comment
	createReq := models.Request{
		Action: models.ActionCreate,
		Object: "comment",
		Data: map[string]interface{}{
			"ticket_id": ticketID,
			"comment":   "To be deleted",
		},
	}
	createBody, _ := json.Marshal(createReq)
	createHttpReq := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(createBody))
	createHttpReq.Header.Set("Content-Type", "application/json")
	wCreate := httptest.NewRecorder()
	cCreate, _ := gin.CreateTestContext(wCreate)
	cCreate.Request = createHttpReq
	cCreate.Set("username", "testuser")
	handler.DoAction(cCreate)

	var createResp models.Response
	json.NewDecoder(wCreate.Body).Decode(&createResp)
	commentID := createResp.Data["comment"].(map[string]interface{})["id"].(string)

	// Delete comment
	removeReq := models.Request{
		Action: models.ActionRemove,
		Object: "comment",
		Data: map[string]interface{}{
			"id": commentID,
		},
	}
	removeBody, _ := json.Marshal(removeReq)
	removeHttpReq := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(removeBody))
	removeHttpReq.Header.Set("Content-Type", "application/json")
	wRemove := httptest.NewRecorder()
	cRemove, _ := gin.CreateTestContext(wRemove)
	cRemove.Request = removeHttpReq
	cRemove.Set("username", "testuser")
	handler.DoAction(cRemove)

	// Try to read deleted comment
	readReq := models.Request{
		Action: models.ActionRead,
		Object: "comment",
		Data: map[string]interface{}{
			"id": commentID,
		},
	}
	readBody, _ := json.Marshal(readReq)
	readHttpReq := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(readBody))
	readHttpReq.Header.Set("Content-Type", "application/json")
	wRead := httptest.NewRecorder()
	cRead, _ := gin.CreateTestContext(wRead)
	cRead.Request = readHttpReq
	cRead.Set("username", "testuser")
	handler.DoAction(cRead)

	assert.Equal(t, http.StatusNotFound, wRead.Code)
}

// =============================================================================
// handleListComments Tests
// =============================================================================

func TestCommentHandler_List_Empty(t *testing.T) {
	handler, ticketID, _ := setupCommentTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionList,
		Object: "comment",
		Data: map[string]interface{}{
			"ticket_id": ticketID,
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
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)

	items, ok := resp.Data["items"].([]interface{})
	require.True(t, ok)
	assert.Equal(t, 0, len(items))
	assert.Equal(t, float64(0), resp.Data["total"])
}

func TestCommentHandler_List_Multiple(t *testing.T) {
	handler, ticketID, _ := setupCommentTestHandler(t)

	// Create 3 comments
	for i := 1; i <= 3; i++ {
		createReq := models.Request{
			Action: models.ActionCreate,
			Object: "comment",
			Data: map[string]interface{}{
				"ticket_id": ticketID,
				"comment":   fmt.Sprintf("Comment %d", i),
			},
		}
		createBody, _ := json.Marshal(createReq)
		createHttpReq := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(createBody))
		createHttpReq.Header.Set("Content-Type", "application/json")
		wCreate := httptest.NewRecorder()
		cCreate, _ := gin.CreateTestContext(wCreate)
		cCreate.Request = createHttpReq
		cCreate.Set("username", "testuser")
		handler.DoAction(cCreate)
	}

	// List comments
	reqBody := models.Request{
		Action: models.ActionList,
		Object: "comment",
		Data: map[string]interface{}{
			"ticket_id": ticketID,
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
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)

	items, ok := resp.Data["items"].([]interface{})
	require.True(t, ok)
	assert.Equal(t, 3, len(items))
	assert.Equal(t, float64(3), resp.Data["total"])
}

func TestCommentHandler_List_MissingTicketID(t *testing.T) {
	handler, _, _ := setupCommentTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionList,
		Object: "comment",
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

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeMissingData, resp.ErrorCode)
	assert.Contains(t, resp.ErrorMessage, "ticket_id")
}

func TestCommentHandler_List_ExcludesDeleted(t *testing.T) {
	handler, ticketID, _ := setupCommentTestHandler(t)

	// Create 2 comments
	var commentID string
	for i := 1; i <= 2; i++ {
		createReq := models.Request{
			Action: models.ActionCreate,
			Object: "comment",
			Data: map[string]interface{}{
				"ticket_id": ticketID,
				"comment":   fmt.Sprintf("Comment %d", i),
			},
		}
		createBody, _ := json.Marshal(createReq)
		createHttpReq := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(createBody))
		createHttpReq.Header.Set("Content-Type", "application/json")
		wCreate := httptest.NewRecorder()
		cCreate, _ := gin.CreateTestContext(wCreate)
		cCreate.Request = createHttpReq
		cCreate.Set("username", "testuser")
		handler.DoAction(cCreate)

		if i == 1 {
			var createResp models.Response
			json.NewDecoder(wCreate.Body).Decode(&createResp)
			commentID = createResp.Data["comment"].(map[string]interface{})["id"].(string)
		}
	}

	// Delete first comment
	removeReq := models.Request{
		Action: models.ActionRemove,
		Object: "comment",
		Data: map[string]interface{}{
			"id": commentID,
		},
	}
	removeBody, _ := json.Marshal(removeReq)
	removeHttpReq := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(removeBody))
	removeHttpReq.Header.Set("Content-Type", "application/json")
	wRemove := httptest.NewRecorder()
	cRemove, _ := gin.CreateTestContext(wRemove)
	cRemove.Request = removeHttpReq
	cRemove.Set("username", "testuser")
	handler.DoAction(cRemove)

	// List comments
	listReq := models.Request{
		Action: models.ActionList,
		Object: "comment",
		Data: map[string]interface{}{
			"ticket_id": ticketID,
		},
	}
	listBody, _ := json.Marshal(listReq)
	listHttpReq := httptest.NewRequest(http.MethodPost, "/do", bytes.NewBuffer(listBody))
	listHttpReq.Header.Set("Content-Type", "application/json")
	wList := httptest.NewRecorder()
	cList, _ := gin.CreateTestContext(wList)
	cList.Request = listHttpReq
	cList.Set("username", "testuser")
	handler.DoAction(cList)

	var listResp models.Response
	json.NewDecoder(wList.Body).Decode(&listResp)
	items := listResp.Data["items"].([]interface{})

	// Should have only 1 comment (deleted one excluded)
	assert.Equal(t, 1, len(items))
}
