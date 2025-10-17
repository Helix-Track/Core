package services

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"helixtrack.ru/chat/internal/models"
)

func TestHTTPCoreService_GetUserInfo(t *testing.T) {
	userID := uuid.New()

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/do", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		// Parse request
		var req map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		assert.Equal(t, "userRead", req["action"])

		// Send response
		response := models.APIResponse{
			ErrorCode: models.ErrorCodeSuccess,
			Data: map[string]interface{}{
				"id":        userID.String(),
				"username":  "testuser",
				"full_name": "Test User",
				"email":     "test@example.com",
			},
		}

		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	service := NewCoreService(server.URL)
	userInfo, err := service.GetUserInfo(context.Background(), userID, "test-jwt")

	assert.NoError(t, err)
	assert.NotNil(t, userInfo)
	assert.Equal(t, "testuser", userInfo.Username)
	assert.Equal(t, "Test User", userInfo.FullName)
}

func TestHTTPCoreService_ValidateEntityAccess(t *testing.T) {
	userID := uuid.New()
	entityID := uuid.New()

	tests := []struct {
		name           string
		hasAccess      bool
		expectAccess   bool
	}{
		{
			name:         "has access",
			hasAccess:    true,
			expectAccess: true,
		},
		{
			name:         "no access",
			hasAccess:    false,
			expectAccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				response := models.APIResponse{
					ErrorCode: models.ErrorCodeSuccess,
					Data: map[string]interface{}{
						"has_access": tt.hasAccess,
					},
				}
				json.NewEncoder(w).Encode(response)
			}))
			defer server.Close()

			service := NewCoreService(server.URL)
			hasAccess, err := service.ValidateEntityAccess(
				context.Background(),
				userID,
				entityID,
				"ticket",
				"test-jwt",
			)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectAccess, hasAccess)
		})
	}
}

func TestHTTPCoreService_GetEntityDetails(t *testing.T) {
	entityID := uuid.New()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]interface{}
		json.NewDecoder(r.Body).Decode(&req)

		assert.Equal(t, "read", req["action"])
		assert.Equal(t, "ticket", req["object"])

		response := models.APIResponse{
			ErrorCode: models.ErrorCodeSuccess,
			Data: map[string]interface{}{
				"id":     entityID.String(),
				"title":  "Test Ticket",
				"status": "open",
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	service := NewCoreService(server.URL)
	details, err := service.GetEntityDetails(
		context.Background(),
		entityID,
		"ticket",
		"test-jwt",
	)

	assert.NoError(t, err)
	assert.NotNil(t, details)
	assert.Equal(t, "Test Ticket", details["title"])
	assert.Equal(t, "open", details["status"])
}

func TestHTTPCoreService_ErrorHandling(t *testing.T) {
	userID := uuid.New()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := models.APIResponse{
			ErrorCode:    models.ErrorCodeNotFound,
			ErrorMessage: "User not found",
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	service := NewCoreService(server.URL)
	userInfo, err := service.GetUserInfo(context.Background(), userID, "test-jwt")

	assert.Error(t, err)
	assert.Nil(t, userInfo)
	assert.Contains(t, err.Error(), "User not found")
}

func TestMockCoreService(t *testing.T) {
	userID := uuid.New()
	entityID := uuid.New()

	mock := &MockCoreService{
		GetUserInfoFunc: func(ctx context.Context, id uuid.UUID, jwt string) (*models.UserInfo, error) {
			return &models.UserInfo{
				ID:       id,
				Username: "mockuser",
				FullName: "Mock User",
			}, nil
		},
		ValidateEntityAccessFunc: func(ctx context.Context, uid, eid uuid.UUID, et, jwt string) (bool, error) {
			return true, nil
		},
		GetEntityDetailsFunc: func(ctx context.Context, eid uuid.UUID, et, jwt string) (map[string]interface{}, error) {
			return map[string]interface{}{
				"id":   eid.String(),
				"name": "Mock Entity",
			}, nil
		},
	}

	// Test GetUserInfo
	userInfo, err := mock.GetUserInfo(context.Background(), userID, "test-jwt")
	assert.NoError(t, err)
	assert.Equal(t, "mockuser", userInfo.Username)

	// Test ValidateEntityAccess
	hasAccess, err := mock.ValidateEntityAccess(context.Background(), userID, entityID, "ticket", "test-jwt")
	assert.NoError(t, err)
	assert.True(t, hasAccess)

	// Test GetEntityDetails
	details, err := mock.GetEntityDetails(context.Background(), entityID, "ticket", "test-jwt")
	assert.NoError(t, err)
	assert.Equal(t, "Mock Entity", details["name"])
}

func TestMockCoreService_Defaults(t *testing.T) {
	userID := uuid.New()
	entityID := uuid.New()

	mock := &MockCoreService{}

	// Test default implementations
	userInfo, err := mock.GetUserInfo(context.Background(), userID, "test-jwt")
	assert.NoError(t, err)
	assert.NotNil(t, userInfo)

	hasAccess, err := mock.ValidateEntityAccess(context.Background(), userID, entityID, "ticket", "test-jwt")
	assert.NoError(t, err)
	assert.True(t, hasAccess)

	details, err := mock.GetEntityDetails(context.Background(), entityID, "ticket", "test-jwt")
	assert.NoError(t, err)
	assert.NotNil(t, details)
}
