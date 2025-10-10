package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"helixtrack.ru/core/internal/models"
)

// AuthService defines the interface for authentication operations
type AuthService interface {
	// Authenticate validates credentials and returns a JWT token
	Authenticate(ctx context.Context, username, password string) (*models.JWTClaims, error)

	// ValidateToken validates a JWT token and returns the claims
	ValidateToken(ctx context.Context, token string) (*models.JWTClaims, error)

	// IsEnabled returns whether the authentication service is enabled
	IsEnabled() bool
}

// httpAuthService is the HTTP-based implementation of AuthService
type httpAuthService struct {
	baseURL    string
	httpClient *http.Client
	enabled    bool
}

// NewAuthService creates a new authentication service client
func NewAuthService(baseURL string, timeout int, enabled bool) AuthService {
	return &httpAuthService{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: time.Duration(timeout) * time.Second,
		},
		enabled: enabled,
	}
}

// AuthRequest represents an authentication request
type AuthRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// AuthResponse represents an authentication response
type AuthResponse struct {
	Token  string           `json:"token"`
	Claims *models.JWTClaims `json:"claims,omitempty"`
}

// Authenticate validates credentials and returns JWT claims
func (s *httpAuthService) Authenticate(ctx context.Context, username, password string) (*models.JWTClaims, error) {
	if !s.enabled {
		return nil, fmt.Errorf("authentication service is not enabled")
	}

	reqBody := AuthRequest{
		Username: username,
		Password: password,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.baseURL+"/authenticate", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("authentication failed: status %d, body: %s", resp.StatusCode, string(body))
	}

	var authResp AuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return authResp.Claims, nil
}

// ValidateToken validates a JWT token and returns the claims
func (s *httpAuthService) ValidateToken(ctx context.Context, token string) (*models.JWTClaims, error) {
	if !s.enabled {
		return nil, fmt.Errorf("authentication service is not enabled")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.baseURL+"/validate", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("token validation failed: status %d, body: %s", resp.StatusCode, string(body))
	}

	var claims models.JWTClaims
	if err := json.NewDecoder(resp.Body).Decode(&claims); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &claims, nil
}

// IsEnabled returns whether the authentication service is enabled
func (s *httpAuthService) IsEnabled() bool {
	return s.enabled
}

// MockAuthService is a mock implementation for testing
type MockAuthService struct {
	AuthenticateFunc   func(ctx context.Context, username, password string) (*models.JWTClaims, error)
	ValidateTokenFunc  func(ctx context.Context, token string) (*models.JWTClaims, error)
	IsEnabledFunc      func() bool
}

func (m *MockAuthService) Authenticate(ctx context.Context, username, password string) (*models.JWTClaims, error) {
	if m.AuthenticateFunc != nil {
		return m.AuthenticateFunc(ctx, username, password)
	}
	return nil, fmt.Errorf("not implemented")
}

func (m *MockAuthService) ValidateToken(ctx context.Context, token string) (*models.JWTClaims, error) {
	if m.ValidateTokenFunc != nil {
		return m.ValidateTokenFunc(ctx, token)
	}
	return nil, fmt.Errorf("not implemented")
}

func (m *MockAuthService) IsEnabled() bool {
	if m.IsEnabledFunc != nil {
		return m.IsEnabledFunc()
	}
	return false
}
