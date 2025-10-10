package services

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"helixtrack.ru/core/internal/models"
)

func TestNewAuthService(t *testing.T) {
	service := NewAuthService("http://localhost:8080", 30, true)
	assert.NotNil(t, service)

	httpService, ok := service.(*httpAuthService)
	assert.True(t, ok)
	assert.Equal(t, "http://localhost:8080", httpService.baseURL)
	assert.True(t, httpService.enabled)
	assert.NotNil(t, httpService.httpClient)
}

func TestAuthService_IsEnabled(t *testing.T) {
	tests := []struct {
		name    string
		enabled bool
	}{
		{"Enabled service", true},
		{"Disabled service", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewAuthService("http://localhost:8080", 30, tt.enabled)
			assert.Equal(t, tt.enabled, service.IsEnabled())
		})
	}
}

func TestAuthService_Authenticate_Success(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/authenticate", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		var req AuthRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		assert.NoError(t, err)
		assert.Equal(t, "testuser", req.Username)
		assert.Equal(t, "testpass", req.Password)

		resp := AuthResponse{
			Token: "test-token",
			Claims: &models.JWTClaims{
				Username: "testuser",
				Role:     "user",
			},
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	service := NewAuthService(server.URL, 30, true)
	claims, err := service.Authenticate(context.Background(), "testuser", "testpass")

	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, "testuser", claims.Username)
	assert.Equal(t, "user", claims.Role)
}

func TestAuthService_Authenticate_Disabled(t *testing.T) {
	service := NewAuthService("http://localhost:8080", 30, false)
	claims, err := service.Authenticate(context.Background(), "testuser", "testpass")

	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.Contains(t, err.Error(), "not enabled")
}

func TestAuthService_Authenticate_InvalidCredentials(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": "invalid credentials"}`))
	}))
	defer server.Close()

	service := NewAuthService(server.URL, 30, true)
	claims, err := service.Authenticate(context.Background(), "testuser", "wrongpass")

	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.Contains(t, err.Error(), "authentication failed")
}

func TestAuthService_Authenticate_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "server error"}`))
	}))
	defer server.Close()

	service := NewAuthService(server.URL, 30, true)
	claims, err := service.Authenticate(context.Background(), "testuser", "testpass")

	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestAuthService_Authenticate_InvalidResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`invalid json`))
	}))
	defer server.Close()

	service := NewAuthService(server.URL, 30, true)
	claims, err := service.Authenticate(context.Background(), "testuser", "testpass")

	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.Contains(t, err.Error(), "failed to decode")
}

func TestAuthService_Authenticate_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	service := NewAuthService(server.URL, 1, true) // 1 second timeout
	ctx := context.Background()
	claims, err := service.Authenticate(ctx, "testuser", "testpass")

	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestAuthService_ValidateToken_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/validate", r.URL.Path)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		claims := models.JWTClaims{
			Username: "testuser",
			Role:     "user",
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(claims)
	}))
	defer server.Close()

	service := NewAuthService(server.URL, 30, true)
	claims, err := service.ValidateToken(context.Background(), "test-token")

	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, "testuser", claims.Username)
	assert.Equal(t, "user", claims.Role)
}

func TestAuthService_ValidateToken_Disabled(t *testing.T) {
	service := NewAuthService("http://localhost:8080", 30, false)
	claims, err := service.ValidateToken(context.Background(), "test-token")

	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.Contains(t, err.Error(), "not enabled")
}

func TestAuthService_ValidateToken_Invalid(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": "invalid token"}`))
	}))
	defer server.Close()

	service := NewAuthService(server.URL, 30, true)
	claims, err := service.ValidateToken(context.Background(), "invalid-token")

	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.Contains(t, err.Error(), "validation failed")
}

func TestAuthService_ValidateToken_InvalidResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`invalid json`))
	}))
	defer server.Close()

	service := NewAuthService(server.URL, 30, true)
	claims, err := service.ValidateToken(context.Background(), "test-token")

	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.Contains(t, err.Error(), "failed to decode")
}

func TestMockAuthService_Authenticate(t *testing.T) {
	mock := &MockAuthService{
		AuthenticateFunc: func(ctx context.Context, username, password string) (*models.JWTClaims, error) {
			if username == "testuser" && password == "testpass" {
				return &models.JWTClaims{
					Username: username,
					Role:     "user",
				}, nil
			}
			return nil, assert.AnError
		},
	}

	// Valid credentials
	claims, err := mock.Authenticate(context.Background(), "testuser", "testpass")
	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, "testuser", claims.Username)

	// Invalid credentials
	claims, err = mock.Authenticate(context.Background(), "wronguser", "wrongpass")
	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestMockAuthService_ValidateToken(t *testing.T) {
	mock := &MockAuthService{
		ValidateTokenFunc: func(ctx context.Context, token string) (*models.JWTClaims, error) {
			if token == "valid-token" {
				return &models.JWTClaims{
					Username: "testuser",
					Role:     "user",
				}, nil
			}
			return nil, assert.AnError
		},
	}

	// Valid token
	claims, err := mock.ValidateToken(context.Background(), "valid-token")
	assert.NoError(t, err)
	assert.NotNil(t, claims)

	// Invalid token
	claims, err = mock.ValidateToken(context.Background(), "invalid-token")
	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestMockAuthService_IsEnabled(t *testing.T) {
	mock := &MockAuthService{
		IsEnabledFunc: func() bool {
			return true
		},
	}

	assert.True(t, mock.IsEnabled())
}

func TestMockAuthService_Defaults(t *testing.T) {
	// Mock with no functions set
	mock := &MockAuthService{}

	claims, err := mock.Authenticate(context.Background(), "test", "test")
	assert.Error(t, err)
	assert.Nil(t, claims)

	claims, err = mock.ValidateToken(context.Background(), "token")
	assert.Error(t, err)
	assert.Nil(t, claims)

	assert.False(t, mock.IsEnabled())
}

func BenchmarkAuthService_Authenticate(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := AuthResponse{
			Token: "test-token",
			Claims: &models.JWTClaims{
				Username: "testuser",
				Role:     "user",
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	service := NewAuthService(server.URL, 30, true)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.Authenticate(context.Background(), "testuser", "testpass")
	}
}

func BenchmarkAuthService_ValidateToken(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims := models.JWTClaims{
			Username: "testuser",
			Role:     "user",
		}
		json.NewEncoder(w).Encode(claims)
	}))
	defer server.Close()

	service := NewAuthService(server.URL, 30, true)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.ValidateToken(context.Background(), "test-token")
	}
}
