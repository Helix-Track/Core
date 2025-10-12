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

func setupMentionTestHandler(t *testing.T) (*Handler, database.Database) {
	db, err := database.NewDatabase(config.DatabaseConfig{
		Type:       "sqlite",
		SQLitePath: ":memory:",
	})
	require.NoError(t, err)

	err = InitializeProjectTables(db)
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

	return handler, db
}

// Helper to create a comment for mention tests
func createTestCommentForMention(t *testing.T, db database.Database) string {
	commentID := uuid.New().String()
	now := time.Now().Unix()
	_, err := db.Exec(context.Background(),
		"INSERT INTO comment (id, comment, created, modified, deleted) VALUES (?, ?, ?, ?, ?)",
		commentID, "Test comment with @username mention", now, now, false)
	require.NoError(t, err)
	return commentID
}

// ============================================================================
// ActionCommentMention Tests
// ============================================================================

func TestMentionHandler_Mention_Success(t *testing.T) {
	handler, db := setupMentionTestHandler(t)
	defer db.Close()

	// Create a comment
	commentID := createTestCommentForMention(t, db)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionCommentMention,
		Data: map[string]interface{}{
			"commentId": commentID,
			"userId":    "mentioned_user",
		},
	}

	handler.handleCommentMention(c, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.Equal(t, commentID, resp.Data["commentId"])
	assert.Equal(t, "mentioned_user", resp.Data["mentionedUserId"])
	assert.NotEmpty(t, resp.Data["mentionId"])
}

func TestMentionHandler_Mention_MissingCommentID(t *testing.T) {
	handler, db := setupMentionTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionCommentMention,
		Data: map[string]interface{}{
			"userId": "mentioned_user",
		},
	}

	handler.handleCommentMention(c, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeMissingData, resp.ErrorCode)
}

func TestMentionHandler_Mention_MissingUserID(t *testing.T) {
	handler, db := setupMentionTestHandler(t)
	defer db.Close()

	commentID := createTestCommentForMention(t, db)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionCommentMention,
		Data: map[string]interface{}{
			"commentId": commentID,
		},
	}

	handler.handleCommentMention(c, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeMissingData, resp.ErrorCode)
}

func TestMentionHandler_Mention_CommentNotFound(t *testing.T) {
	handler, db := setupMentionTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionCommentMention,
		Data: map[string]interface{}{
			"commentId": "nonexistent-id",
			"userId":    "mentioned_user",
		},
	}

	handler.handleCommentMention(c, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeEntityNotFound, resp.ErrorCode)
}

func TestMentionHandler_Mention_AlreadyExists(t *testing.T) {
	handler, db := setupMentionTestHandler(t)
	defer db.Close()

	commentID := createTestCommentForMention(t, db)

	// Create existing mention
	mentionID := uuid.New().String()
	now := time.Now().Unix()
	_, err := db.Exec(context.Background(),
		"INSERT INTO comment_mention_mapping (id, comment_id, mentioned_user_id, created, deleted) VALUES (?, ?, ?, ?, ?)",
		mentionID, commentID, "mentioned_user", now, false)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionCommentMention,
		Data: map[string]interface{}{
			"commentId": commentID,
			"userId":    "mentioned_user",
		},
	}

	handler.handleCommentMention(c, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err = json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.Contains(t, resp.Data["message"].(string), "already exists")
}

func TestMentionHandler_Mention_Unauthorized(t *testing.T) {
	handler, db := setupMentionTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	// No username set

	req := &models.Request{
		Action: models.ActionCommentMention,
		Data:   map[string]interface{}{},
	}

	handler.handleCommentMention(c, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// ============================================================================
// ActionCommentUnmention Tests
// ============================================================================

func TestMentionHandler_Unmention_Success(t *testing.T) {
	handler, db := setupMentionTestHandler(t)
	defer db.Close()

	commentID := createTestCommentForMention(t, db)

	// Create mention
	mentionID := uuid.New().String()
	now := time.Now().Unix()
	_, err := db.Exec(context.Background(),
		"INSERT INTO comment_mention_mapping (id, comment_id, mentioned_user_id, created, deleted) VALUES (?, ?, ?, ?, ?)",
		mentionID, commentID, "mentioned_user", now, false)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionCommentUnmention,
		Data: map[string]interface{}{
			"commentId": commentID,
			"userId":    "mentioned_user",
		},
	}

	handler.handleCommentUnmention(c, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err = json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.Contains(t, resp.Data["message"].(string), "removed successfully")
}

func TestMentionHandler_Unmention_NotFound(t *testing.T) {
	handler, db := setupMentionTestHandler(t)
	defer db.Close()

	commentID := createTestCommentForMention(t, db)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionCommentUnmention,
		Data: map[string]interface{}{
			"commentId": commentID,
			"userId":    "nonexistent_user",
		},
	}

	handler.handleCommentUnmention(c, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeEntityNotFound, resp.ErrorCode)
}

// ============================================================================
// ActionCommentListMentions Tests
// ============================================================================

func TestMentionHandler_ListMentions_Success(t *testing.T) {
	handler, db := setupMentionTestHandler(t)
	defer db.Close()

	commentID := createTestCommentForMention(t, db)

	// Create multiple mentions
	now := time.Now().Unix()
	users := []string{"user1", "user2", "user3"}
	for _, userID := range users {
		mentionID := uuid.New().String()
		_, err := db.Exec(context.Background(),
			"INSERT INTO comment_mention_mapping (id, comment_id, mentioned_user_id, created, deleted) VALUES (?, ?, ?, ?, ?)",
			mentionID, commentID, userID, now, false)
		require.NoError(t, err)
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionCommentListMentions,
		Data: map[string]interface{}{
			"commentId": commentID,
		},
	}

	handler.handleCommentListMentions(c, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.Equal(t, commentID, resp.Data["commentId"])
	assert.Equal(t, float64(3), resp.Data["count"])

	mentions := resp.Data["mentions"].([]interface{})
	assert.Len(t, mentions, 3)
}

func TestMentionHandler_ListMentions_Empty(t *testing.T) {
	handler, db := setupMentionTestHandler(t)
	defer db.Close()

	commentID := createTestCommentForMention(t, db)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionCommentListMentions,
		Data: map[string]interface{}{
			"commentId": commentID,
		},
	}

	handler.handleCommentListMentions(c, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.Equal(t, float64(0), resp.Data["count"])
}

// ============================================================================
// ActionCommentGetMentions Tests
// ============================================================================

func TestMentionHandler_GetMentions_Success(t *testing.T) {
	handler, db := setupMentionTestHandler(t)
	defer db.Close()

	// Create comments
	comment1ID := createTestCommentForMention(t, db)
	comment2ID := createTestCommentForMention(t, db)
	comment3ID := createTestCommentForMention(t, db)

	// Create mentions for testuser
	now := time.Now().Unix()
	for _, commentID := range []string{comment1ID, comment2ID} {
		mentionID := uuid.New().String()
		_, err := db.Exec(context.Background(),
			"INSERT INTO comment_mention_mapping (id, comment_id, mentioned_user_id, created, deleted) VALUES (?, ?, ?, ?, ?)",
			mentionID, commentID, "testuser", now, false)
		require.NoError(t, err)
	}

	// Create mention for different user
	mentionID := uuid.New().String()
	_, err := db.Exec(context.Background(),
		"INSERT INTO comment_mention_mapping (id, comment_id, mentioned_user_id, created, deleted) VALUES (?, ?, ?, ?, ?)",
		mentionID, comment3ID, "otheruser", now, false)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionCommentGetMentions,
		Data: map[string]interface{}{
			"userId": "testuser",
		},
	}

	handler.handleCommentGetMentions(c, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err = json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.Equal(t, "testuser", resp.Data["userId"])
	assert.Equal(t, float64(2), resp.Data["count"])
}

func TestMentionHandler_GetMentions_DefaultsToCurrentUser(t *testing.T) {
	handler, db := setupMentionTestHandler(t)
	defer db.Close()

	commentID := createTestCommentForMention(t, db)

	// Create mention for testuser
	mentionID := uuid.New().String()
	now := time.Now().Unix()
	_, err := db.Exec(context.Background(),
		"INSERT INTO comment_mention_mapping (id, comment_id, mentioned_user_id, created, deleted) VALUES (?, ?, ?, ?, ?)",
		mentionID, commentID, "testuser", now, false)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionCommentGetMentions,
		Data:   map[string]interface{}{},
	}

	handler.handleCommentGetMentions(c, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err = json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.Equal(t, "testuser", resp.Data["userId"])
	assert.Equal(t, float64(1), resp.Data["count"])
}

// ============================================================================
// ActionCommentParseMentions Tests
// ============================================================================

func TestMentionHandler_ParseMentions_Success(t *testing.T) {
	handler, db := setupMentionTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionCommentParseMentions,
		Data: map[string]interface{}{
			"text": "Hey @john and @jane, check this out! @john mentioned again.",
		},
	}

	handler.handleCommentParseMentions(c, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.Equal(t, float64(2), resp.Data["count"]) // john and jane (john not counted twice)

	usernames := resp.Data["usernames"].([]interface{})
	assert.Len(t, usernames, 2)
	assert.Contains(t, usernames, "john")
	assert.Contains(t, usernames, "jane")
}

func TestMentionHandler_ParseMentions_NoMentions(t *testing.T) {
	handler, db := setupMentionTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionCommentParseMentions,
		Data: map[string]interface{}{
			"text": "No mentions in this text",
		},
	}

	handler.handleCommentParseMentions(c, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.Equal(t, float64(0), resp.Data["count"])
}

func TestMentionHandler_ParseMentions_MissingText(t *testing.T) {
	handler, db := setupMentionTestHandler(t)
	defer db.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionCommentParseMentions,
		Data:   map[string]interface{}{},
	}

	handler.handleCommentParseMentions(c, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeMissingData, resp.ErrorCode)
}

func TestMentionHandler_ParseMentions_WithValidation(t *testing.T) {
	handler, db := setupMentionTestHandler(t)
	defer db.Close()

	// Create users in database
	now := time.Now().Unix()
	validUsers := []string{"alice", "bob"}
	for _, username := range validUsers {
		userID := uuid.New().String()
		_, err := db.Exec(context.Background(),
			"INSERT INTO users (id, username, email, created, deleted) VALUES (?, ?, ?, ?, ?)",
			userID, username, username+"@test.com", now, false)
		require.NoError(t, err)
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Set("username", "testuser")

	req := &models.Request{
		Action: models.ActionCommentParseMentions,
		Data: map[string]interface{}{
			"text":     "Hey @alice, @bob, and @charlie!",
			"validate": true,
		},
	}

	handler.handleCommentParseMentions(c, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrorCodeNoError, resp.ErrorCode)
	assert.Equal(t, float64(3), resp.Data["count"]) // All 3 parsed
	assert.Equal(t, float64(2), resp.Data["validCount"]) // Only 2 valid (alice, bob)
}
