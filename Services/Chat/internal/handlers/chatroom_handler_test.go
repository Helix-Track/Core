package handlers

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"helixtrack.ru/chat/internal/models"
)

func TestChatRoomCreate(t *testing.T) {
	tests := []struct {
		name           string
		requestData    map[string]interface{}
		setupMock      func(*TestHelpers)
		expectedStatus int
		expectedError  int
	}{
		{
			name: "successful creation",
			requestData: map[string]interface{}{
				"name":        "Test Room",
				"description": "Test Description",
				"type":        "group",
				"is_private":  false,
			},
			setupMock: func(th *TestHelpers) {
				th.db.ChatRoomCreateFunc = func(ctx context.Context, room *models.ChatRoom) error {
					// Simulate ID generation
					room.ID = uuid.New()
					return nil
				}
				th.db.ParticipantAddFunc = func(ctx context.Context, participant *models.ChatParticipant) error {
					return nil
				}
			},
			expectedStatus: 200,
			expectedError:  -1,
		},
		{
			name: "missing name",
			requestData: map[string]interface{}{
				"type": "group",
			},
			setupMock:      func(th *TestHelpers) {},
			expectedStatus: 400,
			expectedError:  models.ErrorCodeInvalidParameter,
		},
		{
			name: "invalid room type",
			requestData: map[string]interface{}{
				"name": "Test Room",
				"type": "invalid_type",
			},
			setupMock:      func(th *TestHelpers) {},
			expectedStatus: 400,
			expectedError:  models.ErrorCodeInvalidParameter,
		},
		{
			name: "database error on create",
			requestData: map[string]interface{}{
				"name": "Test Room",
				"type": "group",
			},
			setupMock: func(th *TestHelpers) {
				th.db.ChatRoomCreateFunc = func(ctx context.Context, room *models.ChatRoom) error {
					return errors.New("database error")
				}
			},
			expectedStatus: 500,
			expectedError:  models.ErrorCodeDatabaseError,
		},
		{
			name: "with entity association",
			requestData: map[string]interface{}{
				"name":        "Ticket Chat",
				"type":        "group",
				"entity_type": "ticket",
				"entity_id":   uuid.New().String(),
			},
			setupMock: func(th *TestHelpers) {
				th.db.ChatRoomCreateFunc = func(ctx context.Context, room *models.ChatRoom) error {
					room.ID = uuid.New()
					return nil
				}
				th.db.ParticipantAddFunc = func(ctx context.Context, participant *models.ChatParticipant) error {
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

			userID := uuid.New()
			request := map[string]interface{}{
			"action": "chatRoomCreate",
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

func TestChatRoomRead(t *testing.T) {
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
			name: "successful read",
			requestData: map[string]interface{}{
				"id": roomID.String(),
			},
			setupMock: func(th *TestHelpers) {
				th.db.ChatRoomReadFunc = func(ctx context.Context, id string) (*models.ChatRoom, error) {
					return mockRoom, nil
				}
				th.db.ParticipantGetFunc = func(ctx context.Context, chatRoomID, userID string) (*models.ChatParticipant, error) {
					return CreateMockParticipant(chatRoomID, uuid.MustParse(userID), models.ParticipantRoleMember), nil
				}
			},
			expectedStatus: 200,
			expectedError:  -1,
		},
		{
			name:           "missing room ID",
			requestData:    map[string]interface{}{},
			setupMock:      func(th *TestHelpers) {},
			expectedStatus: 400,
			expectedError:  models.ErrorCodeInvalidParameter,
		},
		{
			name: "room not found",
			requestData: map[string]interface{}{
				"id": roomID.String(),
			},
			setupMock: func(th *TestHelpers) {
				th.db.ChatRoomReadFunc = func(ctx context.Context, id string) (*models.ChatRoom, error) {
					return nil, errors.New("not found")
				}
			},
			expectedStatus: 500,
			expectedError:  models.ErrorCodeDatabaseError,
		},
		{
			name: "not a participant",
			requestData: map[string]interface{}{
				"id": roomID.String(),
			},
			setupMock: func(th *TestHelpers) {
				th.db.ChatRoomReadFunc = func(ctx context.Context, id string) (*models.ChatRoom, error) {
					return mockRoom, nil
				}
				th.db.ParticipantGetFunc = func(ctx context.Context, chatRoomID, userID string) (*models.ChatParticipant, error) {
					return nil, errors.New("not a participant")
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
			"action": "chatRoomRead",
			"data":   tt.requestData,
		}

			c, w := th.CreateTestContext("POST", "/api/do", request)
			th.SetClaims(c, userID, "testuser", "user")

			th.h.DoAction(c)

			response := th.AssertJSONResponse(w, tt.expectedStatus, tt.expectedError)

			if tt.expectedError == -1 {
				assert.NotNil(t, response.Data, "Expected data in successful response")
				// Verify room data structure
				roomData, ok := response.Data.(map[string]interface{})
				assert.True(t, ok, "Expected room data to be a map")
				assert.Equal(t, "Test Room", roomData["name"])
			}
		})
	}
}

func TestChatRoomList(t *testing.T) {
	userID := uuid.New()
	room1 := CreateMockChatRoom(uuid.New().String(), "Room 1", userID)
	room2 := CreateMockChatRoom(uuid.New().String(), "Room 2", userID)

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
				"limit":  20,
				"offset": 0,
			},
			setupMock: func(th *TestHelpers) {
				th.db.ChatRoomListFunc = func(ctx context.Context, limit, offset int) ([]*models.ChatRoom, int, error) {
					return []*models.ChatRoom{room1, room2}, 2, nil
				}
			},
			expectedStatus: 200,
			expectedError:  -1,
			expectedCount:  2,
		},
		{
			name:        "default pagination",
			requestData: map[string]interface{}{},
			setupMock: func(th *TestHelpers) {
				th.db.ChatRoomListFunc = func(ctx context.Context, limit, offset int) ([]*models.ChatRoom, int, error) {
					assert.Equal(t, 20, limit, "Expected default limit of 20")
					assert.Equal(t, 0, offset, "Expected default offset of 0")
					return []*models.ChatRoom{room1}, 1, nil
				}
			},
			expectedStatus: 200,
			expectedError:  -1,
			expectedCount:  1,
		},
		{
			name: "empty list",
			requestData: map[string]interface{}{
				"limit": 10,
			},
			setupMock: func(th *TestHelpers) {
				th.db.ChatRoomListFunc = func(ctx context.Context, limit, offset int) ([]*models.ChatRoom, int, error) {
					return []*models.ChatRoom{}, 0, nil
				}
			},
			expectedStatus: 200,
			expectedError:  -1,
			expectedCount:  0,
		},
		{
			name: "database error",
			requestData: map[string]interface{}{
				"limit": 10,
			},
			setupMock: func(th *TestHelpers) {
				th.db.ChatRoomListFunc = func(ctx context.Context, limit, offset int) ([]*models.ChatRoom, int, error) {
					return nil, 0, errors.New("database error")
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
			"action": "chatRoomList",
			"data":   tt.requestData,
		}

			c, w := th.CreateTestContext("POST", "/api/do", request)
			th.SetClaims(c, userID, "testuser", "user")

			th.h.DoAction(c)

			response := th.AssertJSONResponse(w, tt.expectedStatus, tt.expectedError)

			if tt.expectedError == -1 {
				assert.NotNil(t, response.Data, "Expected data in successful response")
				listData, ok := response.Data.(map[string]interface{})
				assert.True(t, ok, "Expected list response structure")

				items, ok := listData["items"].([]interface{})
				assert.True(t, ok, "Expected items array")
				assert.Equal(t, tt.expectedCount, len(items), "Item count mismatch")

				pagination, ok := listData["pagination"].(map[string]interface{})
				assert.True(t, ok, "Expected pagination metadata")
				assert.NotNil(t, pagination["total"])
			}
		})
	}
}

func TestChatRoomUpdate(t *testing.T) {
	roomID := uuid.New()
	ownerID := uuid.New()
	memberID := uuid.New()
	mockRoom := CreateMockChatRoom(roomID.String(), "Original Name", ownerID)

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
			name:     "successful update by owner",
			userID:   ownerID,
			userRole: models.ParticipantRoleOwner,
			requestData: map[string]interface{}{
				"id":          roomID.String(),
				"name":        "Updated Name",
				"description": "Updated Description",
			},
			setupMock: func(th *TestHelpers) {
				th.db.ChatRoomReadFunc = func(ctx context.Context, id string) (*models.ChatRoom, error) {
					return mockRoom, nil
				}
				th.db.ParticipantGetFunc = func(ctx context.Context, chatRoomID, userID string) (*models.ChatParticipant, error) {
					return CreateMockParticipant(chatRoomID, ownerID, models.ParticipantRoleOwner), nil
				}
				th.db.ChatRoomUpdateFunc = func(ctx context.Context, room *models.ChatRoom) error {
					assert.Equal(t, "Updated Name", room.Name)
					assert.Equal(t, "Updated Description", room.Description)
					return nil
				}
			},
			expectedStatus: 200,
			expectedError:  -1,
		},
		{
			name:     "successful update by admin",
			userID:   memberID,
			userRole: models.ParticipantRoleAdmin,
			requestData: map[string]interface{}{
				"id":   roomID.String(),
				"name": "Admin Updated",
			},
			setupMock: func(th *TestHelpers) {
				th.db.ChatRoomReadFunc = func(ctx context.Context, id string) (*models.ChatRoom, error) {
					return mockRoom, nil
				}
				th.db.ParticipantGetFunc = func(ctx context.Context, chatRoomID, userID string) (*models.ChatParticipant, error) {
					return CreateMockParticipant(chatRoomID, memberID, models.ParticipantRoleAdmin), nil
				}
				th.db.ChatRoomUpdateFunc = func(ctx context.Context, room *models.ChatRoom) error {
					return nil
				}
			},
			expectedStatus: 200,
			expectedError:  -1,
		},
		{
			name:     "forbidden - member cannot update",
			userID:   memberID,
			userRole: models.ParticipantRoleMember,
			requestData: map[string]interface{}{
				"id":   roomID.String(),
				"name": "Forbidden Update",
			},
			setupMock: func(th *TestHelpers) {
				th.db.ChatRoomReadFunc = func(ctx context.Context, id string) (*models.ChatRoom, error) {
					return mockRoom, nil
				}
				th.db.ParticipantGetFunc = func(ctx context.Context, chatRoomID, userID string) (*models.ChatParticipant, error) {
					return CreateMockParticipant(chatRoomID, memberID, models.ParticipantRoleMember), nil
				}
			},
			expectedStatus: 403,
			expectedError:  models.ErrorCodeForbidden,
		},
		{
			name:   "missing room ID",
			userID: ownerID,
			requestData: map[string]interface{}{
				"name": "No ID",
			},
			setupMock:      func(th *TestHelpers) {},
			expectedStatus: 400,
			expectedError:  models.ErrorCodeInvalidParameter,
		},
		{
			name:   "room not found",
			userID: ownerID,
			requestData: map[string]interface{}{
				"id":   roomID.String(),
				"name": "Update",
			},
			setupMock: func(th *TestHelpers) {
				th.db.ChatRoomReadFunc = func(ctx context.Context, id string) (*models.ChatRoom, error) {
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
			"action": "chatRoomUpdate",
			"data":   tt.requestData,
		}

			c, w := th.CreateTestContext("POST", "/api/do", request)
			th.SetClaims(c, tt.userID, "testuser", "user")

			th.h.DoAction(c)

			th.AssertJSONResponse(w, tt.expectedStatus, tt.expectedError)
		})
	}
}

func TestChatRoomDelete(t *testing.T) {
	roomID := uuid.New()
	ownerID := uuid.New()
	memberID := uuid.New()
	mockRoom := CreateMockChatRoom(roomID.String(), "Test Room", ownerID)

	tests := []struct {
		name           string
		userID         uuid.UUID
		requestData    map[string]interface{}
		setupMock      func(*TestHelpers)
		expectedStatus int
		expectedError  int
	}{
		{
			name:   "successful delete by owner",
			userID: ownerID,
			requestData: map[string]interface{}{
				"id": roomID.String(),
			},
			setupMock: func(th *TestHelpers) {
				th.db.ChatRoomReadFunc = func(ctx context.Context, id string) (*models.ChatRoom, error) {
					return mockRoom, nil
				}
				th.db.ParticipantGetFunc = func(ctx context.Context, chatRoomID, userID string) (*models.ChatParticipant, error) {
					return CreateMockParticipant(chatRoomID, ownerID, models.ParticipantRoleOwner), nil
				}
				th.db.ChatRoomDeleteFunc = func(ctx context.Context, id string) error {
					return nil
				}
			},
			expectedStatus: 200,
			expectedError:  -1,
		},
		{
			name:   "forbidden - non-owner cannot delete",
			userID: memberID,
			requestData: map[string]interface{}{
				"id": roomID.String(),
			},
			setupMock: func(th *TestHelpers) {
				th.db.ChatRoomReadFunc = func(ctx context.Context, id string) (*models.ChatRoom, error) {
					return mockRoom, nil
				}
				th.db.ParticipantGetFunc = func(ctx context.Context, chatRoomID, userID string) (*models.ChatParticipant, error) {
					return CreateMockParticipant(chatRoomID, memberID, models.ParticipantRoleAdmin), nil
				}
			},
			expectedStatus: 403,
			expectedError:  models.ErrorCodeForbidden,
		},
		{
			name:           "missing room ID",
			userID:         ownerID,
			requestData:    map[string]interface{}{},
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
			"action": "chatRoomDelete",
			"data":   tt.requestData,
		}

			c, w := th.CreateTestContext("POST", "/api/do", request)
			th.SetClaims(c, tt.userID, "testuser", "user")

			th.h.DoAction(c)

			th.AssertJSONResponse(w, tt.expectedStatus, tt.expectedError)
		})
	}
}

func TestChatRoomGetByEntity(t *testing.T) {
	userID := uuid.New()
	entityID := uuid.New()
	roomID := uuid.New()
	mockRoom := CreateMockChatRoom(roomID.String(), "Entity Room", userID)
	mockRoom.EntityType = "ticket"
	mockRoom.EntityID = &entityID

	tests := []struct {
		name           string
		requestData    map[string]interface{}
		setupMock      func(*TestHelpers)
		expectedStatus int
		expectedError  int
	}{
		{
			name: "successful get by entity",
			requestData: map[string]interface{}{
				"entity_type": "ticket",
				"entity_id":   entityID.String(),
			},
			setupMock: func(th *TestHelpers) {
				th.db.ChatRoomGetByEntityFunc = func(ctx context.Context, entityType, entityID string) (*models.ChatRoom, error) {
					return mockRoom, nil
				}
			},
			expectedStatus: 200,
			expectedError:  -1,
		},
		{
			name: "missing entity type",
			requestData: map[string]interface{}{
				"entity_id": entityID.String(),
			},
			setupMock:      func(th *TestHelpers) {},
			expectedStatus: 400,
			expectedError:  models.ErrorCodeInvalidParameter,
		},
		{
			name: "missing entity ID",
			requestData: map[string]interface{}{
				"entity_type": "ticket",
			},
			setupMock:      func(th *TestHelpers) {},
			expectedStatus: 400,
			expectedError:  models.ErrorCodeInvalidParameter,
		},
		{
			name: "entity not found",
			requestData: map[string]interface{}{
				"entity_type": "ticket",
				"entity_id":   entityID.String(),
			},
			setupMock: func(th *TestHelpers) {
				th.db.ChatRoomGetByEntityFunc = func(ctx context.Context, entityType, entityID string) (*models.ChatRoom, error) {
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
			"action": "chatRoomGetByEntity",
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
