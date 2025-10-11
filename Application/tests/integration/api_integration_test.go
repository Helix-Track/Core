package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"helixtrack.ru/core/internal/config"
	"helixtrack.ru/core/internal/database"
	"helixtrack.ru/core/internal/handlers"
	"helixtrack.ru/core/internal/middleware"
	"helixtrack.ru/core/internal/models"
	"helixtrack.ru/core/internal/services"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// TestAPI_FullAuthenticationFlow tests the complete authentication flow
// from login to authenticated requests
func TestAPI_FullAuthenticationFlow(t *testing.T) {
	// Setup
	db, authService, permService := setupIntegrationTest(t)
	defer db.Close()

	handler := handlers.NewHandler(db, authService, permService, "1.0.0-test")

	router := gin.New()
	router.POST("/do", handler.DoAction)

	// Step 1: Check if JWT is capable
	reqBody := models.Request{
		Action: models.ActionJWTCapable,
	}
	w := performRequest(router, "POST", "/do", reqBody)
	assert.Equal(t, http.StatusOK, w.Code)

	var jwtResp models.Response
	err := json.Unmarshal(w.Body.Bytes(), &jwtResp)
	require.NoError(t, err)
	assert.Equal(t, -1, jwtResp.ErrorCode)

	// Step 2: Authenticate
	authReq := models.Request{
		Action: models.ActionAuthenticate,
		Data: map[string]interface{}{
			"username": "testuser",
			"password": "testpass",
		},
	}
	w = performRequest(router, "POST", "/do", authReq)
	assert.Equal(t, http.StatusOK, w.Code)

	var authResp models.Response
	err = json.Unmarshal(w.Body.Bytes(), &authResp)
	require.NoError(t, err)
	assert.Equal(t, -1, authResp.ErrorCode)
}

// TestAPI_HandlerWithDatabase tests handler operations with real database
func TestAPI_HandlerWithDatabase(t *testing.T) {
	db, authService, permService := setupIntegrationTest(t)
	defer db.Close()

	handler := handlers.NewHandler(db, authService, permService, "1.0.0-test")

	router := gin.New()
	router.POST("/do", handler.DoAction)

	// Test version endpoint
	versionReq := models.Request{
		Action: models.ActionVersion,
	}
	w := performRequest(router, "POST", "/do", versionReq)
	assert.Equal(t, http.StatusOK, w.Code)

	var versionResp models.Response
	err := json.Unmarshal(w.Body.Bytes(), &versionResp)
	require.NoError(t, err)
	assert.Equal(t, -1, versionResp.ErrorCode)
	assert.Equal(t, "1.0.0-test", versionResp.Data["version"])

	// Test database capability
	dbReq := models.Request{
		Action: models.ActionDBCapable,
	}
	w = performRequest(router, "POST", "/do", dbReq)
	assert.Equal(t, http.StatusOK, w.Code)

	var dbResp models.Response
	err = json.Unmarshal(w.Body.Bytes(), &dbResp)
	require.NoError(t, err)
	assert.Equal(t, -1, dbResp.ErrorCode)
	assert.True(t, dbResp.Data["dbCapable"].(bool))
	assert.Equal(t, "sqlite", dbResp.Data["type"])
}

// TestAPI_HandlerWithJWTMiddleware tests handler with JWT authentication middleware
func TestAPI_HandlerWithJWTMiddleware(t *testing.T) {
	db, authService, permService := setupIntegrationTest(t)
	defer db.Close()

	handler := handlers.NewHandler(db, authService, permService, "1.0.0-test")

	router := gin.New()

	// Add JWT middleware
	jwtMiddleware := middleware.NewJWTMiddleware(authService, "test-secret")

	// Protected endpoint
	router.POST("/do", jwtMiddleware.Validate(), handler.DoAction)

	// Test without JWT (should fail)
	createReq := models.Request{
		Action: models.ActionCreate,
		Object: "ticket",
	}
	w := performRequest(router, "POST", "/do", createReq)
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// Test with valid JWT
	validToken := "valid-test-token"
	w = performRequestWithAuth(router, "POST", "/do", createReq, validToken)
	// Should succeed authentication, may fail on permissions
	assert.NotEqual(t, http.StatusUnauthorized, w.Code)
}

// TestAPI_HandlerWithPermissionCheck tests permission middleware integration
func TestAPI_HandlerWithPermissionCheck(t *testing.T) {
	db, authService, permService := setupIntegrationTest(t)
	defer db.Close()

	handler := handlers.NewHandler(db, authService, permService, "1.0.0-test")

	router := gin.New()

	// Add JWT middleware
	jwtMiddleware := middleware.NewJWTMiddleware(authService, "test-secret")

	router.POST("/do", jwtMiddleware.Validate(), handler.DoAction)

	// Test create operation (requires CREATE permission)
	createReq := models.Request{
		Action: models.ActionCreate,
		Object: "ticket",
		Data: map[string]interface{}{
			"title": "Test Ticket",
		},
	}

	validToken := "valid-test-token"
	w := performRequestWithAuth(router, "POST", "/do", createReq, validToken)
	// Response depends on permission service mock
	assert.NotEqual(t, http.StatusInternalServerError, w.Code)
}

// TestAPI_HealthEndpoint tests health check with all dependencies
func TestAPI_HealthEndpoint(t *testing.T) {
	db, authService, permService := setupIntegrationTest(t)
	defer db.Close()

	handler := handlers.NewHandler(db, authService, permService, "1.0.0-test")

	router := gin.New()
	router.POST("/do", handler.DoAction)

	healthReq := models.Request{
		Action: models.ActionHealth,
	}
	w := performRequest(router, "POST", "/do", healthReq)
	assert.Equal(t, http.StatusOK, w.Code)

	var healthResp models.Response
	err := json.Unmarshal(w.Body.Bytes(), &healthResp)
	require.NoError(t, err)
	assert.Equal(t, -1, healthResp.ErrorCode)
	assert.Equal(t, "healthy", healthResp.Data["status"])

	checks := healthResp.Data["checks"].(map[string]interface{})
	assert.Equal(t, "healthy", checks["database"])
}

// TestAPI_InvalidRequests tests error handling throughout the stack
func TestAPI_InvalidRequests(t *testing.T) {
	db, authService, permService := setupIntegrationTest(t)
	defer db.Close()

	handler := handlers.NewHandler(db, authService, permService, "1.0.0-test")

	router := gin.New()
	router.POST("/do", handler.DoAction)

	tests := []struct {
		name           string
		request        models.Request
		expectedStatus int
		expectedError  int
	}{
		{
			name: "Invalid action",
			request: models.Request{
				Action: "invalid-action",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  models.ErrorCodeInvalidAction,
		},
		{
			name: "Missing object for create",
			request: models.Request{
				Action: models.ActionCreate,
				Object: "",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  models.ErrorCodeMissingObject,
		},
		{
			name: "Missing username for authenticate",
			request: models.Request{
				Action: models.ActionAuthenticate,
				Data: map[string]interface{}{
					"password": "test",
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  models.ErrorCodeMissingData,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := performRequest(router, "POST", "/do", tt.request)
			assert.Equal(t, tt.expectedStatus, w.Code)

			var resp models.Response
			err := json.Unmarshal(w.Body.Bytes(), &resp)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedError, resp.ErrorCode)
		})
	}
}

// TestAPI_DatabaseOperations tests database integration with handlers
func TestAPI_DatabaseOperations(t *testing.T) {
	db, authService, permService := setupIntegrationTest(t)
	defer db.Close()

	ctx := context.Background()

	// Create test table
	_, err := db.Exec(ctx, `
		CREATE TABLE test_tickets (
			id TEXT PRIMARY KEY,
			title TEXT NOT NULL,
			description TEXT,
			created INTEGER NOT NULL
		)
	`)
	require.NoError(t, err)

	// Insert test data
	_, err = db.Exec(ctx,
		"INSERT INTO test_tickets (id, title, description, created) VALUES (?, ?, ?, ?)",
		"ticket-1", "Test Ticket", "Test Description", 1234567890,
	)
	require.NoError(t, err)

	// Query test data
	row := db.QueryRow(ctx, "SELECT title FROM test_tickets WHERE id = ?", "ticket-1")
	var title string
	err = row.Scan(&title)
	require.NoError(t, err)
	assert.Equal(t, "Test Ticket", title)

	// Verify handler can access same database
	handler := handlers.NewHandler(db, authService, permService, "1.0.0-test")
	assert.NotNil(t, handler)
}

// TestAPI_ConcurrentRequests tests concurrent request handling
func TestAPI_ConcurrentRequests(t *testing.T) {
	db, authService, permService := setupIntegrationTest(t)
	defer db.Close()

	handler := handlers.NewHandler(db, authService, permService, "1.0.0-test")

	router := gin.New()
	router.POST("/do", handler.DoAction)

	// Send concurrent requests
	done := make(chan bool)
	numRequests := 50

	for i := 0; i < numRequests; i++ {
		go func(index int) {
			defer func() { done <- true }()

			req := models.Request{
				Action: models.ActionVersion,
			}
			w := performRequest(router, "POST", "/do", req)
			assert.Equal(t, http.StatusOK, w.Code)
		}(i)
	}

	// Wait for all requests to complete
	for i := 0; i < numRequests; i++ {
		<-done
	}
}

// TestAPI_MiddlewareChain tests complete middleware chain integration
func TestAPI_MiddlewareChain(t *testing.T) {
	db, authService, permService := setupIntegrationTest(t)
	defer db.Close()

	handler := handlers.NewHandler(db, authService, permService, "1.0.0-test")

	router := gin.New()

	// Add middleware chain
	router.Use(gin.Recovery())
	jwtMiddleware := middleware.NewJWTMiddleware(authService, "test-secret")
	router.Use(jwtMiddleware.Validate())

	router.POST("/do", handler.DoAction)

	// Test request goes through entire middleware chain
	req := models.Request{
		Action: models.ActionVersion,
	}

	validToken := "valid-test-token"
	w := performRequestWithAuth(router, "POST", "/do", req, validToken)

	// Should pass through middleware and reach handler
	assert.Equal(t, http.StatusOK, w.Code)
}

// Helper Functions

func setupIntegrationTest(t *testing.T) (database.Database, services.AuthService, services.PermissionService) {
	// Setup test database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "integration_test.db")

	dbCfg := config.DatabaseConfig{
		Type:       "sqlite",
		SQLitePath: dbPath,
	}

	db, err := database.NewDatabase(dbCfg)
	require.NoError(t, err)

	// Setup mock auth service
	authService := &services.MockAuthService{
		AuthenticateFunc: func(ctx context.Context, username, password string) (*models.JWTClaims, error) {
			if username == "testuser" && password == "testpass" {
				return &models.JWTClaims{
					Username: username,
					Role:     "user",
					Name:     "Test User",
				}, nil
			}
			return nil, assert.AnError
		},
		ValidateTokenFunc: func(ctx context.Context, token string) (*models.JWTClaims, error) {
			if token == "valid-test-token" {
				return &models.JWTClaims{
					Username: "testuser",
					Role:     "user",
					Name:     "Test User",
				}, nil
			}
			return nil, assert.AnError
		},
		IsEnabledFunc: func() bool {
			return true
		},
	}

	// Setup mock permission service
	permService := &services.MockPermissionService{
		CheckPermissionFunc: func(ctx context.Context, username, object string, permission models.PermissionLevel) (bool, error) {
			// Allow all permissions for test user
			if username == "testuser" {
				return true, nil
			}
			return false, nil
		},
		IsEnabledFunc: func() bool {
			return true
		},
	}

	return db, authService, permService
}

func performRequest(router *gin.Engine, method, path string, body interface{}) *httptest.ResponseRecorder {
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest(method, path, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	return w
}

func performRequestWithAuth(router *gin.Engine, method, path string, body interface{}, token string) *httptest.ResponseRecorder {
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest(method, path, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	return w
}
