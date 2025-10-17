package handlers

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"helixtrack.ru/chat/internal/models"
)

func TestMessageSend(t *testing.T) {
	roomID := uuid.New()
	userID := uuid.New()
	mockRoom := CreateMockChatRoom(roomID.String(), "Test Room", userID)

	tests := []struct {
		name           string
		requestData    map[string]interface{}
		setupMock      func(*TestHelpers)
		expectedStatus int
		expectedError  int
	}{
		{
			name: "successful message send",
			requestData: map[string]interface{}{
				"chat_room_id": roomID.String(),
				"content":      "Hello, world!",
				"type":         "text",
			},
			setupMock: func(th *TestHelpers) {
				th.db.ChatRoomReadFunc = func(ctx context.Context, id string) (*models.ChatRoom, error) {
					return mockRoom, nil
				}
				th.db.ParticipantGetFunc = func(ctx context.Context, chatRoomID, uid string) (*models.ChatParticipant, error) {
					return CreateMockParticipant(chatRoomID, userID, models.ParticipantRoleMember), nil
				}
				th.db.MessageCreateFunc = func(ctx context.Context, message *models.Message) error {
					message.ID = uuid.New()
					assert.Equal(t, "Hello, world!", message.Content)
					assert.Equal(t, models.MessageTypeText, message.Type)
					return nil
				}
			},
			expectedStatus: 200,
			expectedError:  -1,
		},
		{
			name: "missing chat room ID",
			requestData: map[string]interface{}{
				"content": "Hello",
			},
			setupMock:      func(th *TestHelpers) {},
			expectedStatus: 400,
			expectedError:  models.ErrorCodeInvalidParameter,
		},
		{
			name: "missing content",
			requestData: map[string]interface{}{
				"chat_room_id": roomID.String(),
			},
			setupMock:      func(th *TestHelpers) {},
			expectedStatus: 400,
			expectedError:  models.ErrorCodeInvalidParameter,
		},
		{
			name: "not a participant",
			requestData: map[string]interface{}{
				"chat_room_id": roomID.String(),
				"content":      "Forbidden message",
			},
			setupMock: func(th *TestHelpers) {
				th.db.ChatRoomReadFunc = func(ctx context.Context, id string) (*models.ChatRoom, error) {
					return mockRoom, nil
				}
				th.db.ParticipantGetFunc = func(ctx context.Context, chatRoomID, uid string) (*models.ChatParticipant, error) {
					return nil, errors.New("not a participant")
				}
			},
			expectedStatus: 403,
			expectedError:  models.ErrorCodeForbidden,
		},
		{
			name: "database error on create",
			requestData: map[string]interface{}{
				"chat_room_id": roomID.String(),
				"content":      "Test message",
			},
			setupMock: func(th *TestHelpers) {
				th.db.ChatRoomReadFunc = func(ctx context.Context, id string) (*models.ChatRoom, error) {
					return mockRoom, nil
				}
				th.db.ParticipantGetFunc = func(ctx context.Context, chatRoomID, uid string) (*models.ChatParticipant, error) {
					return CreateMockParticipant(chatRoomID, userID, models.ParticipantRoleMember), nil
				}
				th.db.MessageCreateFunc = func(ctx context.Context, message *models.Message) error {
					return errors.New("database error")
				}
			},
			expectedStatus: 500,
			expectedError:  models.ErrorCodeDatabaseError,
		},
		{
			name: "send with markdown formatting",
			requestData: map[string]interface{}{
				"chat_room_id":   roomID.String(),
				"content":        "# Heading\n\nSome **bold** text",
				"content_format": "markdown",
			},
			setupMock: func(th *TestHelpers) {
				th.db.ChatRoomReadFunc = func(ctx context.Context, id string) (*models.ChatRoom, error) {
					return mockRoom, nil
				}
				th.db.ParticipantGetFunc = func(ctx context.Context, chatRoomID, uid string) (*models.ChatParticipant, error) {
					return CreateMockParticipant(chatRoomID, userID, models.ParticipantRoleMember), nil
				}
				th.db.MessageCreateFunc = func(ctx context.Context, message *models.Message) error {
					message.ID = uuid.New()
					assert.Equal(t, models.ContentFormatMarkdown, message.ContentFormat)
					return nil
				}
			},
			expectedStatus: 200,
			expectedError:  -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			th := NewTestHelpers(t)
			tt.setupMock(th)

			request := map[string]interface{}{
			"action": "messageSend",
			"data":   tt.requestData,
		}

			c, w := th.CreateTestContext("POST", "/api/do", request)
			th.SetClaims(c, userID, "testuser", "user")

			th.h.DoAction(c)

			response := th.AssertJSONResponse(w, tt.expectedStatus, tt.expectedError)

			if tt.expectedError == -1 {
				assert.NotNil(t, response.Data, "Expected data in successful response")
			}
		})
	}
}

func TestMessageReply(t *testing.T) {
	roomID := uuid.New()
	userID := uuid.New()
	parentMsgID := uuid.New()
	mockRoom := CreateMockChatRoom(roomID.String(), "Test Room", userID)
	parentMessage := CreateMockMessage(parentMsgID.String(), roomID.String(), userID, "Parent message")

	tests := []struct {
		name           string
		requestData    map[string]interface{}
		setupMock      func(*TestHelpers)
		expectedStatus int
		expectedError  int
	}{
		{
			name: "successful reply",
			requestData: map[string]interface{}{
				"chat_room_id": roomID.String(),
				"parent_id":    parentMsgID.String(),
				"content":      "This is a reply",
			},
			setupMock: func(th *TestHelpers) {
				th.db.ChatRoomReadFunc = func(ctx context.Context, id string) (*models.ChatRoom, error) {
					return mockRoom, nil
				}
				th.db.ParticipantGetFunc = func(ctx context.Context, chatRoomID, uid string) (*models.ChatParticipant, error) {
					return CreateMockParticipant(chatRoomID, userID, models.ParticipantRoleMember), nil
				}
				th.db.MessageReadFunc = func(ctx context.Context, id string) (*models.Message, error) {
					return parentMessage, nil
				}
				th.db.MessageCreateFunc = func(ctx context.Context, message *models.Message) error {
					message.ID = uuid.New()
					assert.Equal(t, parentMsgID, *message.ParentID)
					assert.Equal(t, "This is a reply", message.Content)
					return nil
				}
			},
			expectedStatus: 200,
			expectedError:  -1,
		},
		{
			name: "missing parent ID",
			requestData: map[string]interface{}{
				"chat_room_id": roomID.String(),
				"content":      "Reply without parent",
			},
			setupMock:      func(th *TestHelpers) {},
			expectedStatus: 400,
			expectedError:  models.ErrorCodeInvalidParameter,
		},
		{
			name: "parent message not found",
			requestData: map[string]interface{}{
				"chat_room_id": roomID.String(),
				"parent_id":    parentMsgID.String(),
				"content":      "Reply",
			},
			setupMock: func(th *TestHelpers) {
				th.db.ChatRoomReadFunc = func(ctx context.Context, id string) (*models.ChatRoom, error) {
					return mockRoom, nil
				}
				th.db.ParticipantGetFunc = func(ctx context.Context, chatRoomID, uid string) (*models.ChatParticipant, error) {
					return CreateMockParticipant(chatRoomID, userID, models.ParticipantRoleMember), nil
				}
				th.db.MessageReadFunc = func(ctx context.Context, id string) (*models.Message, error) {
					return nil, errors.New("not found")
				}
			},
			expectedStatus: 500,
			expectedError:  models.ErrorCodeDatabaseError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			th := NewTestHelpers(t)
			tt.setupMock(th)

			request := map[string]interface{}{
			"action": "messageReply",
			"data":   tt.requestData,
		}

			c, w := th.CreateTestContext("POST", "/api/do", request)
			th.SetClaims(c, userID, "testuser", "user")

			th.h.DoAction(c)

			response := th.AssertJSONResponse(w, tt.expectedStatus, tt.expectedError)

			if tt.expectedError == -1 {
				assert.NotNil(t, response.Data, "Expected data in successful response")
			}
		})
	}
}

func TestMessageList(t *testing.T) {
	roomID := uuid.New()
	userID := uuid.New()
	mockRoom := CreateMockChatRoom(roomID.String(), "Test Room", userID)
	msg1 := CreateMockMessage(uuid.New().String(), roomID.String(), userID, "Message 1")
	msg2 := CreateMockMessage(uuid.New().String(), roomID.String(), userID, "Message 2")

	tests := []struct {
		name           string
		requestData    map[string]interface{}
		setupMock      func(*TestHelpers)
		expectedStatus int
		expectedError  int
		expectedCount  int
	}{
		{
			name: "successful list",
			requestData: map[string]interface{}{
				"chat_room_id": roomID.String(),
				"limit":        50,
				"offset":       0,
			},
			setupMock: func(th *TestHelpers) {
				th.db.ChatRoomReadFunc = func(ctx context.Context, id string) (*models.ChatRoom, error) {
					return mockRoom, nil
				}
				th.db.ParticipantGetFunc = func(ctx context.Context, chatRoomID, uid string) (*models.ChatParticipant, error) {
					return CreateMockParticipant(chatRoomID, userID, models.ParticipantRoleMember), nil
				}
				th.db.MessageListFunc = func(ctx context.Context, req *models.MessageListRequest) ([]*models.Message, int, error) {
					return []*models.Message{msg1, msg2}, 2, nil
				}
			},
			expectedStatus: 200,
			expectedError:  -1,
			expectedCount:  2,
		},
		{
			name: "default pagination",
			requestData: map[string]interface{}{
				"chat_room_id": roomID.String(),
			},
			setupMock: func(th *TestHelpers) {
				th.db.ChatRoomReadFunc = func(ctx context.Context, id string) (*models.ChatRoom, error) {
					return mockRoom, nil
				}
				th.db.ParticipantGetFunc = func(ctx context.Context, chatRoomID, uid string) (*models.ChatParticipant, error) {
					return CreateMockParticipant(chatRoomID, userID, models.ParticipantRoleMember), nil
				}
				th.db.MessageListFunc = func(ctx context.Context, req *models.MessageListRequest) ([]*models.Message, int, error) {
					assert.Equal(t, 50, req.Limit, "Expected default limit of 50")
					assert.Equal(t, 0, req.Offset, "Expected default offset of 0")
					return []*models.Message{msg1}, 1, nil
				}
			},
			expectedStatus: 200,
			expectedError:  -1,
			expectedCount:  1,
		},
		{
			name: "empty list",
			requestData: map[string]interface{}{
				"chat_room_id": roomID.String(),
			},
			setupMock: func(th *TestHelpers) {
				th.db.ChatRoomReadFunc = func(ctx context.Context, id string) (*models.ChatRoom, error) {
					return mockRoom, nil
				}
				th.db.ParticipantGetFunc = func(ctx context.Context, chatRoomID, uid string) (*models.ChatParticipant, error) {
					return CreateMockParticipant(chatRoomID, userID, models.ParticipantRoleMember), nil
				}
				th.db.MessageListFunc = func(ctx context.Context, req *models.MessageListRequest) ([]*models.Message, int, error) {
					return []*models.Message{}, 0, nil
				}
			},
			expectedStatus: 200,
			expectedError:  -1,
			expectedCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			th := NewTestHelpers(t)
			tt.setupMock(th)

			request := map[string]interface{}{
			"action": "messageList",
			"data":   tt.requestData,
		}

			c, w := th.CreateTestContext("POST", "/api/do", request)
			th.SetClaims(c, userID, "testuser", "user")

			th.h.DoAction(c)

			response := th.AssertJSONResponse(w, tt.expectedStatus, tt.expectedError)

			if tt.expectedError == -1 {
				listData, ok := response.Data.(map[string]interface{})
				assert.True(t, ok, "Expected list response structure")

				items, ok := listData["items"].([]interface{})
				assert.True(t, ok, "Expected items array")
				assert.Equal(t, tt.expectedCount, len(items), "Item count mismatch")
			}
		})
	}
}

func TestMessageSearch(t *testing.T) {
	roomID := uuid.New()
	userID := uuid.New()
	mockRoom := CreateMockChatRoom(roomID.String(), "Test Room", userID)
	msg1 := CreateMockMessage(uuid.New().String(), roomID.String(), userID, "Important meeting tomorrow")
	msg2 := CreateMockMessage(uuid.New().String(), roomID.String(), userID, "Meeting notes attached")

	tests := []struct {
		name           string
		requestData    map[string]interface{}
		setupMock      func(*TestHelpers)
		expectedStatus int
		expectedError  int
		expectedCount  int
	}{
		{
			name: "successful search",
			requestData: map[string]interface{}{
				"chat_room_id": roomID.String(),
				"query":        "meeting",
				"limit":        20,
			},
			setupMock: func(th *TestHelpers) {
				th.db.ChatRoomReadFunc = func(ctx context.Context, id string) (*models.ChatRoom, error) {
					return mockRoom, nil
				}
				th.db.ParticipantGetFunc = func(ctx context.Context, chatRoomID, uid string) (*models.ChatParticipant, error) {
					return CreateMockParticipant(chatRoomID, userID, models.ParticipantRoleMember), nil
				}
				th.db.MessageSearchFunc = func(ctx context.Context, chatRoomID, query string, limit, offset int) ([]*models.Message, int, error) {
					assert.Equal(t, "meeting", query)
					return []*models.Message{msg1, msg2}, 2, nil
				}
			},
			expectedStatus: 200,
			expectedError:  -1,
			expectedCount:  2,
		},
		{
			name: "missing search query",
			requestData: map[string]interface{}{
				"chat_room_id": roomID.String(),
			},
			setupMock:      func(th *TestHelpers) {},
			expectedStatus: 400,
			expectedError:  models.ErrorCodeInvalidParameter,
		},
		{
			name: "no results",
			requestData: map[string]interface{}{
				"chat_room_id": roomID.String(),
				"query":        "nonexistent",
			},
			setupMock: func(th *TestHelpers) {
				th.db.ChatRoomReadFunc = func(ctx context.Context, id string) (*models.ChatRoom, error) {
					return mockRoom, nil
				}
				th.db.ParticipantGetFunc = func(ctx context.Context, chatRoomID, uid string) (*models.ChatParticipant, error) {
					return CreateMockParticipant(chatRoomID, userID, models.ParticipantRoleMember), nil
				}
				th.db.MessageSearchFunc = func(ctx context.Context, chatRoomID, query string, limit, offset int) ([]*models.Message, int, error) {
					return []*models.Message{}, 0, nil
				}
			},
			expectedStatus: 200,
			expectedError:  -1,
			expectedCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			th := NewTestHelpers(t)
			tt.setupMock(th)

			request := map[string]interface{}{
			"action": "messageSearch",
			"data":   tt.requestData,
		}

			c, w := th.CreateTestContext("POST", "/api/do", request)
			th.SetClaims(c, userID, "testuser", "user")

			th.h.DoAction(c)

			response := th.AssertJSONResponse(w, tt.expectedStatus, tt.expectedError)

			if tt.expectedError == -1 {
				listData, ok := response.Data.(map[string]interface{})
				assert.True(t, ok, "Expected list response structure")

				items, ok := listData["items"].([]interface{})
				assert.True(t, ok, "Expected items array")
				assert.Equal(t, tt.expectedCount, len(items), "Item count mismatch")
			}
		})
	}
}

func TestMessageUpdate(t *testing.T) {
	roomID := uuid.New()
	userID := uuid.New()
	otherUserID := uuid.New()
	msgID := uuid.New()
	mockMessage := CreateMockMessage(msgID.String(), roomID.String(), userID, "Original content")

	tests := []struct {
		name           string
		senderID       uuid.UUID
		requestData    map[string]interface{}
		setupMock      func(*TestHelpers)
		expectedStatus int
		expectedError  int
	}{
		{
			name:     "successful update by author",
			senderID: userID,
			requestData: map[string]interface{}{
				"id":      msgID.String(),
				"content": "Updated content",
			},
			setupMock: func(th *TestHelpers) {
				th.db.MessageReadFunc = func(ctx context.Context, id string) (*models.Message, error) {
					return mockMessage, nil
				}
				th.db.MessageUpdateFunc = func(ctx context.Context, message *models.Message) error {
					assert.Equal(t, "Updated content", message.Content)
					assert.True(t, message.IsEdited)
					return nil
				}
			},
			expectedStatus: 200,
			expectedError:  -1,
		},
		{
			name:     "forbidden - cannot edit others' messages",
			senderID: otherUserID,
			requestData: map[string]interface{}{
				"id":      msgID.String(),
				"content": "Forbidden edit",
			},
			setupMock: func(th *TestHelpers) {
				th.db.MessageReadFunc = func(ctx context.Context, id string) (*models.Message, error) {
					return mockMessage, nil
				}
			},
			expectedStatus: 403,
			expectedError:  models.ErrorCodeForbidden,
		},
		{
			name:     "missing message ID",
			senderID: userID,
			requestData: map[string]interface{}{
				"content": "Updated",
			},
			setupMock:      func(th *TestHelpers) {},
			expectedStatus: 400,
			expectedError:  models.ErrorCodeInvalidParameter,
		},
		{
			name:     "missing content",
			senderID: userID,
			requestData: map[string]interface{}{
				"id": msgID.String(),
			},
			setupMock:      func(th *TestHelpers) {},
			expectedStatus: 400,
			expectedError:  models.ErrorCodeInvalidParameter,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			th := NewTestHelpers(t)
			tt.setupMock(th)

			request := map[string]interface{}{
			"action": "messageUpdate",
			"data":   tt.requestData,
		}

			c, w := th.CreateTestContext("POST", "/api/do", request)
			th.SetClaims(c, tt.senderID, "testuser", "user")

			th.h.DoAction(c)

			th.AssertJSONResponse(w, tt.expectedStatus, tt.expectedError)
		})
	}
}

func TestMessageDelete(t *testing.T) {
	roomID := uuid.New()
	userID := uuid.New()
	adminID := uuid.New()
	otherUserID := uuid.New()
	msgID := uuid.New()
	mockMessage := CreateMockMessage(msgID.String(), roomID.String(), userID, "Test message")

	tests := []struct {
		name           string
		deleterID      uuid.UUID
		deleterRole    models.ParticipantRole
		requestData    map[string]interface{}
		setupMock      func(*TestHelpers)
		expectedStatus int
		expectedError  int
	}{
		{
			name:        "successful delete by author",
			deleterID:   userID,
			deleterRole: models.ParticipantRoleMember,
			requestData: map[string]interface{}{
				"id": msgID.String(),
			},
			setupMock: func(th *TestHelpers) {
				th.db.MessageReadFunc = func(ctx context.Context, id string) (*models.Message, error) {
					return mockMessage, nil
				}
				th.db.MessageDeleteFunc = func(ctx context.Context, id string) error {
					return nil
				}
			},
			expectedStatus: 200,
			expectedError:  -1,
		},
		{
			name:        "successful delete by admin",
			deleterID:   adminID,
			deleterRole: models.ParticipantRoleAdmin,
			requestData: map[string]interface{}{
				"id": msgID.String(),
			},
			setupMock: func(th *TestHelpers) {
				th.db.MessageReadFunc = func(ctx context.Context, id string) (*models.Message, error) {
					return mockMessage, nil
				}
				th.db.ParticipantGetFunc = func(ctx context.Context, chatRoomID, uid string) (*models.ChatParticipant, error) {
					return CreateMockParticipant(chatRoomID, adminID, models.ParticipantRoleAdmin), nil
				}
				th.db.MessageDeleteFunc = func(ctx context.Context, id string) error {
					return nil
				}
			},
			expectedStatus: 200,
			expectedError:  -1,
		},
		{
			name:        "forbidden - non-admin cannot delete others' messages",
			deleterID:   otherUserID,
			deleterRole: models.ParticipantRoleMember,
			requestData: map[string]interface{}{
				"id": msgID.String(),
			},
			setupMock: func(th *TestHelpers) {
				th.db.MessageReadFunc = func(ctx context.Context, id string) (*models.Message, error) {
					return mockMessage, nil
				}
				th.db.ParticipantGetFunc = func(ctx context.Context, chatRoomID, uid string) (*models.ChatParticipant, error) {
					return CreateMockParticipant(chatRoomID, otherUserID, models.ParticipantRoleMember), nil
				}
			},
			expectedStatus: 403,
			expectedError:  models.ErrorCodeForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			th := NewTestHelpers(t)
			tt.setupMock(th)

			request := map[string]interface{}{
			"action": "messageDelete",
			"data":   tt.requestData,
		}

			c, w := th.CreateTestContext("POST", "/api/do", request)
			th.SetClaims(c, tt.deleterID, "testuser", "user")

			th.h.DoAction(c)

			th.AssertJSONResponse(w, tt.expectedStatus, tt.expectedError)
		})
	}
}

func TestMessagePin(t *testing.T) {
	roomID := uuid.New()
	adminID := uuid.New()
	memberID := uuid.New()
	msgID := uuid.New()
	mockMessage := CreateMockMessage(msgID.String(), roomID.String(), memberID, "Test message")

	tests := []struct {
		name           string
		userID         uuid.UUID
		userRole       models.ParticipantRole
		requestData    map[string]interface{}
		setupMock      func(*TestHelpers)
		expectedStatus int
		expectedError  int
	}{
		{
			name:     "successful pin by admin",
			userID:   adminID,
			userRole: models.ParticipantRoleAdmin,
			requestData: map[string]interface{}{
				"id": msgID.String(),
			},
			setupMock: func(th *TestHelpers) {
				th.db.MessageReadFunc = func(ctx context.Context, id string) (*models.Message, error) {
					return mockMessage, nil
				}
				th.db.ParticipantGetFunc = func(ctx context.Context, chatRoomID, uid string) (*models.ChatParticipant, error) {
					return CreateMockParticipant(chatRoomID, adminID, models.ParticipantRoleAdmin), nil
				}
				th.db.MessageUpdateFunc = func(ctx context.Context, message *models.Message) error {
					assert.True(t, message.IsPinned)
					assert.NotNil(t, message.PinnedBy)
					return nil
				}
			},
			expectedStatus: 200,
			expectedError:  -1,
		},
		{
			name:     "successful pin by moderator",
			userID:   adminID,
			userRole: models.ParticipantRoleModerator,
			requestData: map[string]interface{}{
				"id": msgID.String(),
			},
			setupMock: func(th *TestHelpers) {
				th.db.MessageReadFunc = func(ctx context.Context, id string) (*models.Message, error) {
					return mockMessage, nil
				}
				th.db.ParticipantGetFunc = func(ctx context.Context, chatRoomID, uid string) (*models.ChatParticipant, error) {
					return CreateMockParticipant(chatRoomID, adminID, models.ParticipantRoleModerator), nil
				}
				th.db.MessageUpdateFunc = func(ctx context.Context, message *models.Message) error {
					return nil
				}
			},
			expectedStatus: 200,
			expectedError:  -1,
		},
		{
			name:     "forbidden - member cannot pin",
			userID:   memberID,
			userRole: models.ParticipantRoleMember,
			requestData: map[string]interface{}{
				"id": msgID.String(),
			},
			setupMock: func(th *TestHelpers) {
				th.db.MessageReadFunc = func(ctx context.Context, id string) (*models.Message, error) {
					return mockMessage, nil
				}
				th.db.ParticipantGetFunc = func(ctx context.Context, chatRoomID, uid string) (*models.ChatParticipant, error) {
					return CreateMockParticipant(chatRoomID, memberID, models.ParticipantRoleMember), nil
				}
			},
			expectedStatus: 403,
			expectedError:  models.ErrorCodeForbidden,
		},
		{
			name:   "missing message ID",
			userID: adminID,
			requestData: map[string]interface{}{},
			setupMock:      func(th *TestHelpers) {},
			expectedStatus: 400,
			expectedError:  models.ErrorCodeInvalidParameter,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			th := NewTestHelpers(t)
			tt.setupMock(th)

			request := map[string]interface{}{
			"action": "messagePin",
			"data":   tt.requestData,
		}

			c, w := th.CreateTestContext("POST", "/api/do", request)
			th.SetClaims(c, tt.userID, "testuser", "user")

			th.h.DoAction(c)

			th.AssertJSONResponse(w, tt.expectedStatus, tt.expectedError)
		})
	}
}

func TestMessageUnpin(t *testing.T) {
	roomID := uuid.New()
	adminID := uuid.New()
	memberID := uuid.New()
	msgID := uuid.New()
	mockMessage := CreateMockMessage(msgID.String(), roomID.String(), memberID, "Test message")
	mockMessage.IsPinned = true
	mockMessage.PinnedBy = &adminID

	tests := []struct {
		name           string
		userID         uuid.UUID
		userRole       models.ParticipantRole
		requestData    map[string]interface{}
		setupMock      func(*TestHelpers)
		expectedStatus int
		expectedError  int
	}{
		{
			name:     "successful unpin by admin",
			userID:   adminID,
			userRole: models.ParticipantRoleAdmin,
			requestData: map[string]interface{}{
				"id": msgID.String(),
			},
			setupMock: func(th *TestHelpers) {
				th.db.MessageReadFunc = func(ctx context.Context, id string) (*models.Message, error) {
					return mockMessage, nil
				}
				th.db.ParticipantGetFunc = func(ctx context.Context, chatRoomID, uid string) (*models.ChatParticipant, error) {
					return CreateMockParticipant(chatRoomID, adminID, models.ParticipantRoleAdmin), nil
				}
				th.db.MessageUpdateFunc = func(ctx context.Context, message *models.Message) error {
					assert.False(t, message.IsPinned)
					return nil
				}
			},
			expectedStatus: 200,
			expectedError:  -1,
		},
		{
			name:     "forbidden - member cannot unpin",
			userID:   memberID,
			userRole: models.ParticipantRoleMember,
			requestData: map[string]interface{}{
				"id": msgID.String(),
			},
			setupMock: func(th *TestHelpers) {
				th.db.MessageReadFunc = func(ctx context.Context, id string) (*models.Message, error) {
					return mockMessage, nil
				}
				th.db.ParticipantGetFunc = func(ctx context.Context, chatRoomID, uid string) (*models.ChatParticipant, error) {
					return CreateMockParticipant(chatRoomID, memberID, models.ParticipantRoleMember), nil
				}
			},
			expectedStatus: 403,
			expectedError:  models.ErrorCodeForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			th := NewTestHelpers(t)
			tt.setupMock(th)

			request := map[string]interface{}{
			"action": "messageUnpin",
			"data":   tt.requestData,
		}

			c, w := th.CreateTestContext("POST", "/api/do", request)
			th.SetClaims(c, tt.userID, "testuser", "user")

			th.h.DoAction(c)

			th.AssertJSONResponse(w, tt.expectedStatus, tt.expectedError)
		})
	}
}

func TestMessageUpdate_CreatesEditHistory(t *testing.T) {
	roomID := uuid.New()
	userID := uuid.New()
	msgID := uuid.New()
	mockMessage := CreateMockMessage(msgID.String(), roomID.String(), userID, "Original content")

	tests := []struct {
		name              string
		requestData       map[string]interface{}
		setupMock         func(*TestHelpers)
		expectedStatus    int
		expectedError     int
		expectHistoryCall bool
	}{
		{
			name: "successful update creates edit history",
			requestData: map[string]interface{}{
				"id":      msgID.String(),
				"content": "Updated content",
			},
			setupMock: func(th *TestHelpers) {
				historyCalled := false
				th.db.MessageReadFunc = func(ctx context.Context, id string) (*models.Message, error) {
					return mockMessage, nil
				}
				th.db.MessageEditHistoryCountFunc = func(ctx context.Context, messageID string) (int, error) {
					return 0, nil // First edit
				}
				th.db.MessageEditHistoryCreateFunc = func(ctx context.Context, history *models.MessageEditHistory) error {
					historyCalled = true
					assert.Equal(t, mockMessage.ID, history.MessageID)
					assert.Equal(t, userID, history.EditorID)
					assert.Equal(t, "Original content", history.PreviousContent)
					assert.Equal(t, models.ContentFormatPlain, history.PreviousContentFormat)
					assert.Equal(t, 1, history.EditNumber)
					return nil
				}
				th.db.MessageUpdateFunc = func(ctx context.Context, message *models.Message) error {
					assert.True(t, historyCalled, "Edit history should be created before message update")
					assert.Equal(t, "Updated content", message.Content)
					assert.True(t, message.IsEdited)
					return nil
				}
			},
			expectedStatus:    200,
			expectedError:     -1,
			expectHistoryCall: true,
		},
		{
			name: "second edit increments edit number",
			requestData: map[string]interface{}{
				"id":      msgID.String(),
				"content": "Second update",
			},
			setupMock: func(th *TestHelpers) {
				th.db.MessageReadFunc = func(ctx context.Context, id string) (*models.Message, error) {
					return mockMessage, nil
				}
				th.db.MessageEditHistoryCountFunc = func(ctx context.Context, messageID string) (int, error) {
					return 1, nil // Second edit
				}
				th.db.MessageEditHistoryCreateFunc = func(ctx context.Context, history *models.MessageEditHistory) error {
					assert.Equal(t, 2, history.EditNumber, "Second edit should have edit number 2")
					return nil
				}
				th.db.MessageUpdateFunc = func(ctx context.Context, message *models.Message) error {
					return nil
				}
			},
			expectedStatus:    200,
			expectedError:     -1,
			expectHistoryCall: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			th := NewTestHelpers(t)
			tt.setupMock(th)

			request := map[string]interface{}{
				"action": "messageUpdate",
				"data":   tt.requestData,
			}

			c, w := th.CreateTestContext("POST", "/api/do", request)
			th.SetClaims(c, userID, "testuser", "user")

			th.h.DoAction(c)

			th.AssertJSONResponse(w, tt.expectedStatus, tt.expectedError)
		})
	}
}

func TestMessageGetEditHistory(t *testing.T) {
	roomID := uuid.New()
	userID := uuid.New()
	otherUserID := uuid.New()
	msgID := uuid.New()
	mockMessage := CreateMockMessage(msgID.String(), roomID.String(), userID, "Current content")
	mockMessage.IsEdited = true

	// Create mock edit history
	history1 := &models.MessageEditHistory{
		ID:                    uuid.New(),
		MessageID:             msgID,
		EditorID:              userID,
		PreviousContent:       "Original content",
		PreviousContentFormat: models.ContentFormatPlain,
		EditNumber:            1,
		EditedAt:              1234567890,
		CreatedAt:             1234567890,
	}
	history2 := &models.MessageEditHistory{
		ID:                    uuid.New(),
		MessageID:             msgID,
		EditorID:              userID,
		PreviousContent:       "First edit",
		PreviousContentFormat: models.ContentFormatPlain,
		EditNumber:            2,
		EditedAt:              1234567900,
		CreatedAt:             1234567900,
	}

	tests := []struct {
		name           string
		userID         uuid.UUID
		requestData    map[string]interface{}
		setupMock      func(*TestHelpers)
		expectedStatus int
		expectedError  int
		expectedEdits  int
	}{
		{
			name:   "successful retrieval of edit history",
			userID: userID,
			requestData: map[string]interface{}{
				"id": msgID.String(),
			},
			setupMock: func(th *TestHelpers) {
				th.db.MessageReadFunc = func(ctx context.Context, id string) (*models.Message, error) {
					return mockMessage, nil
				}
				th.db.ParticipantGetFunc = func(ctx context.Context, chatRoomID, uid string) (*models.ChatParticipant, error) {
					return CreateMockParticipant(chatRoomID, userID, models.ParticipantRoleMember), nil
				}
				th.db.MessageEditHistoryListFunc = func(ctx context.Context, messageID string) ([]*models.MessageEditHistory, error) {
					assert.Equal(t, msgID.String(), messageID)
					return []*models.MessageEditHistory{history1, history2}, nil
				}
				th.db.MessageEditHistoryCountFunc = func(ctx context.Context, messageID string) (int, error) {
					return 2, nil
				}
			},
			expectedStatus: 200,
			expectedError:  -1,
			expectedEdits:  2,
		},
		{
			name:   "multiple edits in correct order",
			userID: userID,
			requestData: map[string]interface{}{
				"id": msgID.String(),
			},
			setupMock: func(th *TestHelpers) {
				th.db.MessageReadFunc = func(ctx context.Context, id string) (*models.Message, error) {
					return mockMessage, nil
				}
				th.db.ParticipantGetFunc = func(ctx context.Context, chatRoomID, uid string) (*models.ChatParticipant, error) {
					return CreateMockParticipant(chatRoomID, userID, models.ParticipantRoleMember), nil
				}
				th.db.MessageEditHistoryListFunc = func(ctx context.Context, messageID string) ([]*models.MessageEditHistory, error) {
					return []*models.MessageEditHistory{history1, history2}, nil
				}
				th.db.MessageEditHistoryCountFunc = func(ctx context.Context, messageID string) (int, error) {
					return 2, nil
				}
			},
			expectedStatus: 200,
			expectedError:  -1,
			expectedEdits:  2,
		},
		{
			name:   "empty history for unedited message",
			userID: userID,
			requestData: map[string]interface{}{
				"id": msgID.String(),
			},
			setupMock: func(th *TestHelpers) {
				uneditedMessage := CreateMockMessage(msgID.String(), roomID.String(), userID, "Never edited")
				uneditedMessage.IsEdited = false
				th.db.MessageReadFunc = func(ctx context.Context, id string) (*models.Message, error) {
					return uneditedMessage, nil
				}
				th.db.ParticipantGetFunc = func(ctx context.Context, chatRoomID, uid string) (*models.ChatParticipant, error) {
					return CreateMockParticipant(chatRoomID, userID, models.ParticipantRoleMember), nil
				}
				th.db.MessageEditHistoryListFunc = func(ctx context.Context, messageID string) ([]*models.MessageEditHistory, error) {
					return []*models.MessageEditHistory{}, nil
				}
				th.db.MessageEditHistoryCountFunc = func(ctx context.Context, messageID string) (int, error) {
					return 0, nil
				}
			},
			expectedStatus: 200,
			expectedError:  -1,
			expectedEdits:  0,
		},
		{
			name:   "forbidden - not a participant",
			userID: otherUserID,
			requestData: map[string]interface{}{
				"id": msgID.String(),
			},
			setupMock: func(th *TestHelpers) {
				th.db.MessageReadFunc = func(ctx context.Context, id string) (*models.Message, error) {
					return mockMessage, nil
				}
				th.db.ParticipantGetFunc = func(ctx context.Context, chatRoomID, uid string) (*models.ChatParticipant, error) {
					return nil, errors.New("not a participant")
				}
			},
			expectedStatus: 403,
			expectedError:  models.ErrorCodeForbidden,
		},
		{
			name:   "message not found",
			userID: userID,
			requestData: map[string]interface{}{
				"id": msgID.String(),
			},
			setupMock: func(th *TestHelpers) {
				th.db.MessageReadFunc = func(ctx context.Context, id string) (*models.Message, error) {
					return nil, models.ErrNotFound
				}
			},
			expectedStatus: 404,
			expectedError:  models.ErrorCodeNotFound,
		},
		{
			name:   "missing message ID",
			userID: userID,
			requestData: map[string]interface{}{
				// No ID provided
			},
			setupMock:      func(th *TestHelpers) {},
			expectedStatus: 400,
			expectedError:  models.ErrorCodeInvalidParameter,
		},
		{
			name:   "database error on history fetch",
			userID: userID,
			requestData: map[string]interface{}{
				"id": msgID.String(),
			},
			setupMock: func(th *TestHelpers) {
				th.db.MessageReadFunc = func(ctx context.Context, id string) (*models.Message, error) {
					return mockMessage, nil
				}
				th.db.ParticipantGetFunc = func(ctx context.Context, chatRoomID, uid string) (*models.ChatParticipant, error) {
					return CreateMockParticipant(chatRoomID, userID, models.ParticipantRoleMember), nil
				}
				th.db.MessageEditHistoryListFunc = func(ctx context.Context, messageID string) ([]*models.MessageEditHistory, error) {
					return nil, errors.New("database error")
				}
			},
			expectedStatus: 500,
			expectedError:  models.ErrorCodeDatabaseError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			th := NewTestHelpers(t)
			tt.setupMock(th)

			request := map[string]interface{}{
				"action": "messageGetEditHistory",
				"data":   tt.requestData,
			}

			c, w := th.CreateTestContext("POST", "/api/do", request)
			th.SetClaims(c, tt.userID, "testuser", "user")

			th.h.DoAction(c)

			response := th.AssertJSONResponse(w, tt.expectedStatus, tt.expectedError)

			if tt.expectedError == -1 {
				data, ok := response.Data.(map[string]interface{})
				assert.True(t, ok, "Expected edit history response structure")

				editHistory, ok := data["edit_history"].([]interface{})
				assert.True(t, ok, "Expected edit_history array")
				assert.Equal(t, tt.expectedEdits, len(editHistory), "Edit count mismatch")

				totalEdits, ok := data["total_edits"].(float64)
				assert.True(t, ok, "Expected total_edits field")
				assert.Equal(t, float64(tt.expectedEdits), totalEdits, "Total edits mismatch")
			}
		})
	}
}
