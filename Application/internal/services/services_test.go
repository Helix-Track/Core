package services

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"helixtrack.ru/core/internal/models"
)

func TestAuthService_Authenticate(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/authenticate", r.URL.Path)

		var req AuthRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		if req.Username == "testuser" && req.Password == "testpass" {
			resp := AuthResponse{
				Token: "test-token",
				Claims: &models.JWTClaims{
					RegisteredClaims: jwt.RegisteredClaims{
						Subject: "authentication",
					},
					Name:     "Test User",
					Username: "testuser",
					Role:     "admin",
				},
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(resp)
		} else {
			w.WriteHeader(http.StatusUnauthorized)
		}
	}))
	defer server.Close()

	authService := NewAuthService(server.URL, 30, true)

	t.Run("Successful authentication", func(t *testing.T) {
		claims, err := authService.Authenticate(context.Background(), "testuser", "testpass")
		require.NoError(t, err)
		assert.NotNil(t, claims)
		assert.Equal(t, "testuser", claims.Username)
		assert.Equal(t, "Test User", claims.Name)
		assert.Equal(t, "admin", claims.Role)
	})

	t.Run("Failed authentication", func(t *testing.T) {
		claims, err := authService.Authenticate(context.Background(), "wronguser", "wrongpass")
		assert.Error(t, err)
		assert.Nil(t, claims)
	})
}

func TestAuthService_ValidateToken(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/validate", r.URL.Path)

		authHeader := r.Header.Get("Authorization")
		if authHeader == "Bearer valid-token" {
			claims := models.JWTClaims{
				RegisteredClaims: jwt.RegisteredClaims{
					Subject: "authentication",
				},
				Username: "testuser",
				Name:     "Test User",
				Role:     "user",
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(claims)
		} else {
			w.WriteHeader(http.StatusUnauthorized)
		}
	}))
	defer server.Close()

	authService := NewAuthService(server.URL, 30, true)

	t.Run("Valid token", func(t *testing.T) {
		claims, err := authService.ValidateToken(context.Background(), "valid-token")
		require.NoError(t, err)
		assert.NotNil(t, claims)
		assert.Equal(t, "testuser", claims.Username)
	})

	t.Run("Invalid token", func(t *testing.T) {
		claims, err := authService.ValidateToken(context.Background(), "invalid-token")
		assert.Error(t, err)
		assert.Nil(t, claims)
	})
}

func TestAuthService_IsEnabled(t *testing.T) {
	t.Run("Enabled service", func(t *testing.T) {
		authService := NewAuthService("http://localhost:8081", 30, true)
		assert.True(t, authService.IsEnabled())
	})

	t.Run("Disabled service", func(t *testing.T) {
		authService := NewAuthService("http://localhost:8081", 30, false)
		assert.False(t, authService.IsEnabled())
	})
}

func TestAuthService_Disabled(t *testing.T) {
	authService := NewAuthService("http://localhost:8081", 30, false)

	t.Run("Authenticate when disabled", func(t *testing.T) {
		claims, err := authService.Authenticate(context.Background(), "user", "pass")
		assert.Error(t, err)
		assert.Nil(t, claims)
		assert.Contains(t, err.Error(), "not enabled")
	})

	t.Run("ValidateToken when disabled", func(t *testing.T) {
		claims, err := authService.ValidateToken(context.Background(), "token")
		assert.Error(t, err)
		assert.Nil(t, claims)
		assert.Contains(t, err.Error(), "not enabled")
	})
}

func TestAuthService_ContextTimeout(t *testing.T) {
	// Create a slow server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	authService := NewAuthService(server.URL, 1, true) // 1 second timeout

	ctx := context.Background()
	_, err := authService.Authenticate(ctx, "user", "pass")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "deadline exceeded")
}

func TestPermissionService_CheckPermission(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/check", r.URL.Path)

		var req PermissionCheckRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		allowed := req.Username == "admin" || req.RequiredLevel <= models.PermissionRead
		resp := PermissionCheckResponse{
			Allowed: allowed,
			Reason:  "test reason",
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	permService := NewPermissionService(server.URL, 30, true)

	t.Run("Permission allowed", func(t *testing.T) {
		allowed, err := permService.CheckPermission(context.Background(), "admin", "project/test", models.PermissionDelete)
		require.NoError(t, err)
		assert.True(t, allowed)
	})

	t.Run("Permission denied", func(t *testing.T) {
		allowed, err := permService.CheckPermission(context.Background(), "user", "project/test", models.PermissionDelete)
		require.NoError(t, err)
		assert.False(t, allowed)
	})
}

func TestPermissionService_GetUserPermissions(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/permissions/testuser", r.URL.Path)

		resp := UserPermissionsResponse{
			Permissions: []models.Permission{
				{Context: "org/team1", Level: models.PermissionRead},
				{Context: "org/team2", Level: models.PermissionUpdate},
			},
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	permService := NewPermissionService(server.URL, 30, true)

	perms, err := permService.GetUserPermissions(context.Background(), "testuser")
	require.NoError(t, err)
	assert.Len(t, perms, 2)
	assert.Equal(t, "org/team1", perms[0].Context)
	assert.Equal(t, models.PermissionRead, perms[0].Level)
}

func TestPermissionService_IsEnabled(t *testing.T) {
	t.Run("Enabled service", func(t *testing.T) {
		permService := NewPermissionService("http://localhost:8082", 30, true)
		assert.True(t, permService.IsEnabled())
	})

	t.Run("Disabled service", func(t *testing.T) {
		permService := NewPermissionService("http://localhost:8082", 30, false)
		assert.False(t, permService.IsEnabled())
	})
}

func TestPermissionService_Disabled(t *testing.T) {
	permService := NewPermissionService("http://localhost:8082", 30, false)

	t.Run("CheckPermission when disabled allows all", func(t *testing.T) {
		allowed, err := permService.CheckPermission(context.Background(), "user", "context", models.PermissionDelete)
		require.NoError(t, err)
		assert.True(t, allowed, "Should allow when disabled")
	})

	t.Run("GetUserPermissions when disabled returns empty", func(t *testing.T) {
		perms, err := permService.GetUserPermissions(context.Background(), "user")
		require.NoError(t, err)
		assert.Empty(t, perms)
	})
}

func TestMockAuthService(t *testing.T) {
	mock := &MockAuthService{
		AuthenticateFunc: func(ctx context.Context, username, password string) (*models.JWTClaims, error) {
			return &models.JWTClaims{Username: username}, nil
		},
		ValidateTokenFunc: func(ctx context.Context, token string) (*models.JWTClaims, error) {
			return &models.JWTClaims{Username: "mockuser"}, nil
		},
		IsEnabledFunc: func() bool {
			return true
		},
	}

	claims, err := mock.Authenticate(context.Background(), "test", "pass")
	require.NoError(t, err)
	assert.Equal(t, "test", claims.Username)

	claims, err = mock.ValidateToken(context.Background(), "token")
	require.NoError(t, err)
	assert.Equal(t, "mockuser", claims.Username)

	assert.True(t, mock.IsEnabled())
}

func TestMockPermissionService(t *testing.T) {
	mock := &MockPermissionService{
		CheckPermissionFunc: func(ctx context.Context, username, permissionContext string, requiredLevel models.PermissionLevel) (bool, error) {
			return username == "admin", nil
		},
		GetUserPermissionsFunc: func(ctx context.Context, username string) ([]models.Permission, error) {
			return []models.Permission{{Context: "test", Level: models.PermissionRead}}, nil
		},
		IsEnabledFunc: func() bool {
			return true
		},
	}

	allowed, err := mock.CheckPermission(context.Background(), "admin", "context", models.PermissionDelete)
	require.NoError(t, err)
	assert.True(t, allowed)

	perms, err := mock.GetUserPermissions(context.Background(), "user")
	require.NoError(t, err)
	assert.Len(t, perms, 1)

	assert.True(t, mock.IsEnabled())
}
