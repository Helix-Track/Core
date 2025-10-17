package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"helixtrack.ru/chat/internal/logger"
	"helixtrack.ru/chat/internal/models"
)

// CoreService interface for Core service operations
type CoreService interface {
	GetUserInfo(ctx context.Context, userID uuid.UUID, jwt string) (*models.UserInfo, error)
	ValidateEntityAccess(ctx context.Context, userID, entityID uuid.UUID, entityType, jwt string) (bool, error)
	GetEntityDetails(ctx context.Context, entityID uuid.UUID, entityType, jwt string) (map[string]interface{}, error)
}

// HTTPCoreService implements CoreService using HTTP
type HTTPCoreService struct {
	baseURL    string
	httpClient *http.Client
}

// NewCoreService creates a new Core service client
func NewCoreService(baseURL string) *HTTPCoreService {
	return &HTTPCoreService{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GetUserInfo retrieves user information from Core service
func (s *HTTPCoreService) GetUserInfo(ctx context.Context, userID uuid.UUID, jwt string) (*models.UserInfo, error) {
	request := map[string]interface{}{
		"action": "userRead",
		"jwt":    jwt,
		"data": map[string]interface{}{
			"user_id": userID.String(),
		},
	}

	response, err := s.doRequest(ctx, request)
	if err != nil {
		return nil, err
	}

	if response.ErrorCode != models.ErrorCodeSuccess {
		return nil, fmt.Errorf("core service error: %s", response.ErrorMessage)
	}

	// Parse user info from response data
	userInfo := &models.UserInfo{}
	dataBytes, err := json.Marshal(response.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response data: %w", err)
	}

	if err := json.Unmarshal(dataBytes, userInfo); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user info: %w", err)
	}

	return userInfo, nil
}

// ValidateEntityAccess checks if user has access to an entity
func (s *HTTPCoreService) ValidateEntityAccess(ctx context.Context, userID, entityID uuid.UUID, entityType, jwt string) (bool, error) {
	request := map[string]interface{}{
		"action": "checkAccess",
		"jwt":    jwt,
		"data": map[string]interface{}{
			"user_id":     userID.String(),
			"entity_id":   entityID.String(),
			"entity_type": entityType,
		},
	}

	response, err := s.doRequest(ctx, request)
	if err != nil {
		logger.Error("Failed to validate entity access",
			zap.Error(err),
			zap.String("user_id", userID.String()),
			zap.String("entity_id", entityID.String()),
			zap.String("entity_type", entityType),
		)
		return false, err
	}

	if response.ErrorCode != models.ErrorCodeSuccess {
		return false, nil
	}

	// Check if data contains "has_access" field
	if dataMap, ok := response.Data.(map[string]interface{}); ok {
		if hasAccess, ok := dataMap["has_access"].(bool); ok {
			return hasAccess, nil
		}
	}

	return false, nil
}

// GetEntityDetails retrieves entity details from Core service
func (s *HTTPCoreService) GetEntityDetails(ctx context.Context, entityID uuid.UUID, entityType, jwt string) (map[string]interface{}, error) {
	action := "read"
	object := entityType

	request := map[string]interface{}{
		"action": action,
		"jwt":    jwt,
		"object": object,
		"data": map[string]interface{}{
			"id": entityID.String(),
		},
	}

	response, err := s.doRequest(ctx, request)
	if err != nil {
		return nil, err
	}

	if response.ErrorCode != models.ErrorCodeSuccess {
		return nil, fmt.Errorf("core service error: %s", response.ErrorMessage)
	}

	// Return data as map
	if dataMap, ok := response.Data.(map[string]interface{}); ok {
		return dataMap, nil
	}

	return nil, fmt.Errorf("invalid response data format")
}

// doRequest performs HTTP request to Core service
func (s *HTTPCoreService) doRequest(ctx context.Context, request map[string]interface{}) (*models.APIResponse, error) {
	// Marshal request
	reqBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	url := s.baseURL + "/do"
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	// Execute request
	httpResp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer httpResp.Body.Close()

	// Parse response
	var response models.APIResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &response, nil
}

// MockCoreService implements CoreService for testing
type MockCoreService struct {
	GetUserInfoFunc           func(ctx context.Context, userID uuid.UUID, jwt string) (*models.UserInfo, error)
	ValidateEntityAccessFunc  func(ctx context.Context, userID, entityID uuid.UUID, entityType, jwt string) (bool, error)
	GetEntityDetailsFunc      func(ctx context.Context, entityID uuid.UUID, entityType, jwt string) (map[string]interface{}, error)
}

func (m *MockCoreService) GetUserInfo(ctx context.Context, userID uuid.UUID, jwt string) (*models.UserInfo, error) {
	if m.GetUserInfoFunc != nil {
		return m.GetUserInfoFunc(ctx, userID, jwt)
	}
	return &models.UserInfo{
		ID:       userID,
		Username: "testuser",
		FullName: "Test User",
		Email:    "test@example.com",
	}, nil
}

func (m *MockCoreService) ValidateEntityAccess(ctx context.Context, userID, entityID uuid.UUID, entityType, jwt string) (bool, error) {
	if m.ValidateEntityAccessFunc != nil {
		return m.ValidateEntityAccessFunc(ctx, userID, entityID, entityType, jwt)
	}
	return true, nil
}

func (m *MockCoreService) GetEntityDetails(ctx context.Context, entityID uuid.UUID, entityType, jwt string) (map[string]interface{}, error) {
	if m.GetEntityDetailsFunc != nil {
		return m.GetEntityDetailsFunc(ctx, entityID, entityType, jwt)
	}
	return map[string]interface{}{
		"id":   entityID.String(),
		"type": entityType,
		"name": "Test Entity",
	}, nil
}
