package handlers

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"helixtrack.ru/chat/internal/models"
)

func TestParticipantAdd(t *testing.T) {
	roomID := uuid.New()
	ownerID := uuid.New()
	newUserID := uuid.New()
	mockRoom := CreateMockChatRoom(roomID.String(), "Test Room", ownerID)

	tests := []struct {
		name           string
		adderID        uuid.UUID
		adderRole      models.ParticipantRole
		requestData    map[string]interface{}
		setupMock      func(*TestHelpers)
		expectedStatus int
		expectedError  int
	}{
		{
			name:      "successful add by owner",
			adderID:   ownerID,
			adderRole: models.ParticipantRoleOwner,
			requestData: map[string]interface{}{
				"chat_room_id": roomID.String(),
				"user_id":      newUserID.String(),
				"role":         "member",
			},
			setupMock: func(th *TestHelpers) {
				th.db.ChatRoomReadFunc = func(ctx context.Context, id string) (*models.ChatRoom, error) {
					return mockRoom, nil
				}
				th.db.ParticipantGetFunc = func(ctx context.Context, chatRoomID, uid string) (*models.ChatParticipant, error) {
					if uid == ownerID.String() {
						return CreateMockParticipant(chatRoomID, ownerID, models.ParticipantRoleOwner), nil
					}
					return nil, errors.New("not found")
				}
				th.db.ParticipantAddFunc = func(ctx context.Context, participant *models.ChatParticipant) error {
					assert.Equal(t, newUserID, participant.UserID)
					assert.Equal(t, models.ParticipantRoleMember, participant.Role)
					return nil
				}
			},
			expectedStatus: 200,
			expectedError:  -1,
		},
		{
			name:      "successful add by admin",
			adderID:   ownerID,
			adderRole: models.ParticipantRoleAdmin,
			requestData: map[string]interface{}{
				"chat_room_id": roomID.String(),
				"user_id":      newUserID.String(),
				"role":         "member",
			},
			setupMock: func(th *TestHelpers) {
				th.db.ChatRoomReadFunc = func(ctx context.Context, id string) (*models.ChatRoom, error) {
					return mockRoom, nil
				}
				th.db.ParticipantGetFunc = func(ctx context.Context, chatRoomID, uid string) (*models.ChatParticipant, error) {
					if uid == ownerID.String() {
						return CreateMockParticipant(chatRoomID, ownerID, models.ParticipantRoleAdmin), nil
					}
					return nil, errors.New("not found")
				}
				th.db.ParticipantAddFunc = func(ctx context.Context, participant *models.ChatParticipant) error {
					return nil
				}
			},
			expectedStatus: 200,
			expectedError:  -1,
		},
		{
			name:      "forbidden - member cannot add participants",
			adderID:   ownerID,
			adderRole: models.ParticipantRoleMember,
			requestData: map[string]interface{}{
				"chat_room_id": roomID.String(),
				"user_id":      newUserID.String(),
			},
			setupMock: func(th *TestHelpers) {
				th.db.ChatRoomReadFunc = func(ctx context.Context, id string) (*models.ChatRoom, error) {
					return mockRoom, nil
				}
				th.db.ParticipantGetFunc = func(ctx context.Context, chatRoomID, uid string) (*models.ChatParticipant, error) {
					return CreateMockParticipant(chatRoomID, ownerID, models.ParticipantRoleMember), nil
				}
			},
			expectedStatus: 403,
			expectedError:  models.ErrorCodeForbidden,
		},
		{
			name:    "missing chat room ID",
			adderID: ownerID,
			requestData: map[string]interface{}{
				"user_id": newUserID.String(),
			},
			setupMock:      func(th *TestHelpers) {},
			expectedStatus: 400,
			expectedError:  models.ErrorCodeInvalidParameter,
		},
		{
			name:    "missing user ID",
			adderID: ownerID,
			requestData: map[string]interface{}{
				"chat_room_id": roomID.String(),
			},
			setupMock:      func(th *TestHelpers) {},
			expectedStatus: 400,
			expectedError:  models.ErrorCodeInvalidParameter,
		},
		{
			name:      "database error on add",
			adderID:   ownerID,
			adderRole: models.ParticipantRoleOwner,
			requestData: map[string]interface{}{
				"chat_room_id": roomID.String(),
				"user_id":      newUserID.String(),
			},
			setupMock: func(th *TestHelpers) {
				th.db.ChatRoomReadFunc = func(ctx context.Context, id string) (*models.ChatRoom, error) {
					return mockRoom, nil
				}
				th.db.ParticipantGetFunc = func(ctx context.Context, chatRoomID, uid string) (*models.ChatParticipant, error) {
					return CreateMockParticipant(chatRoomID, ownerID, models.ParticipantRoleOwner), nil
				}
				th.db.ParticipantAddFunc = func(ctx context.Context, participant *models.ChatParticipant) error {
					return errors.New("database error")
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
			"action": "participantAdd",
			"data":   tt.requestData,
		}

			c, w := th.CreateTestContext("POST", "/api/do", request)
			th.SetClaims(c, tt.adderID, "testuser", "user")

			th.h.DoAction(c)

			th.AssertJSONResponse(w, tt.expectedStatus, tt.expectedError)
		})
	}
}

func TestParticipantRemove(t *testing.T) {
	roomID := uuid.New()
	ownerID := uuid.New()
	memberID := uuid.New()
	otherUserID := uuid.New()
	mockRoom := CreateMockChatRoom(roomID.String(), "Test Room", ownerID)

	tests := []struct {
		name           string
		removerID      uuid.UUID
		removerRole    models.ParticipantRole
		targetUserID   uuid.UUID
		requestData    map[string]interface{}
		setupMock      func(*TestHelpers)
		expectedStatus int
		expectedError  int
	}{
		{
			name:         "successful remove by owner",
			removerID:    ownerID,
			removerRole:  models.ParticipantRoleOwner,
			targetUserID: memberID,
			requestData: map[string]interface{}{
				"chat_room_id": roomID.String(),
				"user_id":      memberID.String(),
			},
			setupMock: func(th *TestHelpers) {
				th.db.ChatRoomReadFunc = func(ctx context.Context, id string) (*models.ChatRoom, error) {
					return mockRoom, nil
				}
				th.db.ParticipantGetFunc = func(ctx context.Context, chatRoomID, uid string) (*models.ChatParticipant, error) {
					if uid == ownerID.String() {
						return CreateMockParticipant(chatRoomID, ownerID, models.ParticipantRoleOwner), nil
					}
					return CreateMockParticipant(chatRoomID, memberID, models.ParticipantRoleMember), nil
				}
				th.db.ParticipantRemoveFunc = func(ctx context.Context, chatRoomID, userID string) error {
					return nil
				}
			},
			expectedStatus: 200,
			expectedError:  -1,
		},
		{
			name:         "successful self-removal",
			removerID:    memberID,
			removerRole:  models.ParticipantRoleMember,
			targetUserID: memberID,
			requestData: map[string]interface{}{
				"chat_room_id": roomID.String(),
				"user_id":      memberID.String(),
			},
			setupMock: func(th *TestHelpers) {
				th.db.ChatRoomReadFunc = func(ctx context.Context, id string) (*models.ChatRoom, error) {
					return mockRoom, nil
				}
				th.db.ParticipantGetFunc = func(ctx context.Context, chatRoomID, uid string) (*models.ChatParticipant, error) {
					return CreateMockParticipant(chatRoomID, memberID, models.ParticipantRoleMember), nil
				}
				th.db.ParticipantRemoveFunc = func(ctx context.Context, chatRoomID, userID string) error {
					return nil
				}
			},
			expectedStatus: 200,
			expectedError:  -1,
		},
		{
			name:         "forbidden - member cannot remove others",
			removerID:    memberID,
			removerRole:  models.ParticipantRoleMember,
			targetUserID: otherUserID,
			requestData: map[string]interface{}{
				"chat_room_id": roomID.String(),
				"user_id":      otherUserID.String(),
			},
			setupMock: func(th *TestHelpers) {
				th.db.ChatRoomReadFunc = func(ctx context.Context, id string) (*models.ChatRoom, error) {
					return mockRoom, nil
				}
				th.db.ParticipantGetFunc = func(ctx context.Context, chatRoomID, uid string) (*models.ChatParticipant, error) {
					return CreateMockParticipant(chatRoomID, memberID, models.ParticipantRoleMember), nil
				}
			},
			expectedStatus: 403,
			expectedError:  models.ErrorCodeForbidden,
		},
		{
			name:         "forbidden - cannot remove owner",
			removerID:    memberID,
			removerRole:  models.ParticipantRoleAdmin,
			targetUserID: ownerID,
			requestData: map[string]interface{}{
				"chat_room_id": roomID.String(),
				"user_id":      ownerID.String(),
			},
			setupMock: func(th *TestHelpers) {
				th.db.ChatRoomReadFunc = func(ctx context.Context, id string) (*models.ChatRoom, error) {
					return mockRoom, nil
				}
				th.db.ParticipantGetFunc = func(ctx context.Context, chatRoomID, uid string) (*models.ChatParticipant, error) {
					if uid == memberID.String() {
						return CreateMockParticipant(chatRoomID, memberID, models.ParticipantRoleAdmin), nil
					}
					return CreateMockParticipant(chatRoomID, ownerID, models.ParticipantRoleOwner), nil
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
			"action": "participantRemove",
			"data":   tt.requestData,
		}

			c, w := th.CreateTestContext("POST", "/api/do", request)
			th.SetClaims(c, tt.removerID, "testuser", "user")

			th.h.DoAction(c)

			th.AssertJSONResponse(w, tt.expectedStatus, tt.expectedError)
		})
	}
}

func TestParticipantList(t *testing.T) {
	roomID := uuid.New()
	userID := uuid.New()
	mockRoom := CreateMockChatRoom(roomID.String(), "Test Room", userID)
	p1 := CreateMockParticipant(roomID.String(), userID, models.ParticipantRoleOwner)
	p2 := CreateMockParticipant(roomID.String(), uuid.New(), models.ParticipantRoleMember)

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
			},
			setupMock: func(th *TestHelpers) {
				th.db.ChatRoomReadFunc = func(ctx context.Context, id string) (*models.ChatRoom, error) {
					return mockRoom, nil
				}
				th.db.ParticipantGetFunc = func(ctx context.Context, chatRoomID, uid string) (*models.ChatParticipant, error) {
					return p1, nil
				}
				th.db.ParticipantListFunc = func(ctx context.Context, chatRoomID string) ([]*models.ChatParticipant, error) {
					return []*models.ChatParticipant{p1, p2}, nil
				}
			},
			expectedStatus: 200,
			expectedError:  -1,
			expectedCount:  2,
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
					return p1, nil
				}
				th.db.ParticipantListFunc = func(ctx context.Context, chatRoomID string) ([]*models.ChatParticipant, error) {
					return []*models.ChatParticipant{}, nil
				}
			},
			expectedStatus: 200,
			expectedError:  -1,
			expectedCount:  0,
		},
		{
			name:           "missing chat room ID",
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
			"action": "participantList",
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

func TestParticipantUpdateRole(t *testing.T) {
	roomID := uuid.New()
	ownerID := uuid.New()
	memberID := uuid.New()
	mockRoom := CreateMockChatRoom(roomID.String(), "Test Room", ownerID)

	tests := []struct {
		name           string
		updaterID      uuid.UUID
		updaterRole    models.ParticipantRole
		targetUserID   uuid.UUID
		newRole        string
		requestData    map[string]interface{}
		setupMock      func(*TestHelpers)
		expectedStatus int
		expectedError  int
	}{
		{
			name:         "successful role update by owner",
			updaterID:    ownerID,
			updaterRole:  models.ParticipantRoleOwner,
			targetUserID: memberID,
			newRole:      "admin",
			requestData: map[string]interface{}{
				"chat_room_id": roomID.String(),
				"user_id":      memberID.String(),
				"role":         "admin",
			},
			setupMock: func(th *TestHelpers) {
				th.db.ChatRoomReadFunc = func(ctx context.Context, id string) (*models.ChatRoom, error) {
					return mockRoom, nil
				}
				th.db.ParticipantGetFunc = func(ctx context.Context, chatRoomID, uid string) (*models.ChatParticipant, error) {
					if uid == ownerID.String() {
						return CreateMockParticipant(chatRoomID, ownerID, models.ParticipantRoleOwner), nil
					}
					return CreateMockParticipant(chatRoomID, memberID, models.ParticipantRoleMember), nil
				}
				th.db.ParticipantUpdateRoleFunc = func(ctx context.Context, chatRoomID, userID string, role models.ParticipantRole) error {
					assert.Equal(t, models.ParticipantRoleAdmin, role)
					return nil
				}
			},
			expectedStatus: 200,
			expectedError:  -1,
		},
		{
			name:         "successful role update by admin",
			updaterID:    memberID,
			updaterRole:  models.ParticipantRoleAdmin,
			targetUserID: memberID,
			newRole:      "moderator",
			requestData: map[string]interface{}{
				"chat_room_id": roomID.String(),
				"user_id":      memberID.String(),
				"role":         "moderator",
			},
			setupMock: func(th *TestHelpers) {
				th.db.ChatRoomReadFunc = func(ctx context.Context, id string) (*models.ChatRoom, error) {
					return mockRoom, nil
				}
				th.db.ParticipantGetFunc = func(ctx context.Context, chatRoomID, uid string) (*models.ChatParticipant, error) {
					return CreateMockParticipant(chatRoomID, memberID, models.ParticipantRoleAdmin), nil
				}
				th.db.ParticipantUpdateRoleFunc = func(ctx context.Context, chatRoomID, userID string, role models.ParticipantRole) error {
					return nil
				}
			},
			expectedStatus: 200,
			expectedError:  -1,
		},
		{
			name:         "forbidden - member cannot update roles",
			updaterID:    memberID,
			updaterRole:  models.ParticipantRoleMember,
			targetUserID: memberID,
			newRole:      "admin",
			requestData: map[string]interface{}{
				"chat_room_id": roomID.String(),
				"user_id":      memberID.String(),
				"role":         "admin",
			},
			setupMock: func(th *TestHelpers) {
				th.db.ChatRoomReadFunc = func(ctx context.Context, id string) (*models.ChatRoom, error) {
					return mockRoom, nil
				}
				th.db.ParticipantGetFunc = func(ctx context.Context, chatRoomID, uid string) (*models.ChatParticipant, error) {
					return CreateMockParticipant(chatRoomID, memberID, models.ParticipantRoleMember), nil
				}
			},
			expectedStatus: 403,
			expectedError:  models.ErrorCodeForbidden,
		},
		{
			name:      "missing role",
			updaterID: ownerID,
			requestData: map[string]interface{}{
				"chat_room_id": roomID.String(),
				"user_id":      memberID.String(),
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
			"action": "participantUpdateRole",
			"data":   tt.requestData,
		}

			c, w := th.CreateTestContext("POST", "/api/do", request)
			th.SetClaims(c, tt.updaterID, "testuser", "user")

			th.h.DoAction(c)

			th.AssertJSONResponse(w, tt.expectedStatus, tt.expectedError)
		})
	}
}

func TestParticipantMute(t *testing.T) {
	roomID := uuid.New()
	moderatorID := uuid.New()
	memberID := uuid.New()
	mockRoom := CreateMockChatRoom(roomID.String(), "Test Room", moderatorID)

	tests := []struct {
		name           string
		muterID        uuid.UUID
		muterRole      models.ParticipantRole
		targetUserID   uuid.UUID
		requestData    map[string]interface{}
		setupMock      func(*TestHelpers)
		expectedStatus int
		expectedError  int
	}{
		{
			name:         "successful mute by moderator",
			muterID:      moderatorID,
			muterRole:    models.ParticipantRoleModerator,
			targetUserID: memberID,
			requestData: map[string]interface{}{
				"chat_room_id": roomID.String(),
				"user_id":      memberID.String(),
			},
			setupMock: func(th *TestHelpers) {
				th.db.ChatRoomReadFunc = func(ctx context.Context, id string) (*models.ChatRoom, error) {
					return mockRoom, nil
				}
				th.db.ParticipantGetFunc = func(ctx context.Context, chatRoomID, uid string) (*models.ChatParticipant, error) {
					if uid == moderatorID.String() {
						return CreateMockParticipant(chatRoomID, moderatorID, models.ParticipantRoleModerator), nil
					}
					return CreateMockParticipant(chatRoomID, memberID, models.ParticipantRoleMember), nil
				}
				th.db.ParticipantMuteFunc = func(ctx context.Context, chatRoomID, userID string, muted bool) error {
					assert.True(t, muted)
					return nil
				}
			},
			expectedStatus: 200,
			expectedError:  -1,
		},
		{
			name:         "forbidden - member cannot mute",
			muterID:      memberID,
			muterRole:    models.ParticipantRoleMember,
			targetUserID: memberID,
			requestData: map[string]interface{}{
				"chat_room_id": roomID.String(),
				"user_id":      memberID.String(),
			},
			setupMock: func(th *TestHelpers) {
				th.db.ChatRoomReadFunc = func(ctx context.Context, id string) (*models.ChatRoom, error) {
					return mockRoom, nil
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
			"action": "participantMute",
			"data":   tt.requestData,
		}

			c, w := th.CreateTestContext("POST", "/api/do", request)
			th.SetClaims(c, tt.muterID, "testuser", "user")

			th.h.DoAction(c)

			th.AssertJSONResponse(w, tt.expectedStatus, tt.expectedError)
		})
	}
}

func TestParticipantUnmute(t *testing.T) {
	roomID := uuid.New()
	moderatorID := uuid.New()
	memberID := uuid.New()
	mockRoom := CreateMockChatRoom(roomID.String(), "Test Room", moderatorID)

	tests := []struct {
		name           string
		unmuterID      uuid.UUID
		unmuterRole    models.ParticipantRole
		targetUserID   uuid.UUID
		requestData    map[string]interface{}
		setupMock      func(*TestHelpers)
		expectedStatus int
		expectedError  int
	}{
		{
			name:         "successful unmute by moderator",
			unmuterID:    moderatorID,
			unmuterRole:  models.ParticipantRoleModerator,
			targetUserID: memberID,
			requestData: map[string]interface{}{
				"chat_room_id": roomID.String(),
				"user_id":      memberID.String(),
			},
			setupMock: func(th *TestHelpers) {
				th.db.ChatRoomReadFunc = func(ctx context.Context, id string) (*models.ChatRoom, error) {
					return mockRoom, nil
				}
				th.db.ParticipantGetFunc = func(ctx context.Context, chatRoomID, uid string) (*models.ChatParticipant, error) {
					if uid == moderatorID.String() {
						return CreateMockParticipant(chatRoomID, moderatorID, models.ParticipantRoleModerator), nil
					}
					p := CreateMockParticipant(chatRoomID, memberID, models.ParticipantRoleMember)
					p.IsMuted = true
					return p, nil
				}
				th.db.ParticipantMuteFunc = func(ctx context.Context, chatRoomID, userID string, muted bool) error {
					assert.False(t, muted)
					return nil
				}
			},
			expectedStatus: 200,
			expectedError:  -1,
		},
		{
			name:         "forbidden - member cannot unmute",
			unmuterID:    memberID,
			unmuterRole:  models.ParticipantRoleMember,
			targetUserID: memberID,
			requestData: map[string]interface{}{
				"chat_room_id": roomID.String(),
				"user_id":      memberID.String(),
			},
			setupMock: func(th *TestHelpers) {
				th.db.ChatRoomReadFunc = func(ctx context.Context, id string) (*models.ChatRoom, error) {
					return mockRoom, nil
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
			"action": "participantUnmute",
			"data":   tt.requestData,
		}

			c, w := th.CreateTestContext("POST", "/api/do", request)
			th.SetClaims(c, tt.unmuterID, "testuser", "user")

			th.h.DoAction(c)

			th.AssertJSONResponse(w, tt.expectedStatus, tt.expectedError)
		})
	}
}
