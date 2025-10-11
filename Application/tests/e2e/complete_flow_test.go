package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"helixtrack.ru/core/internal/cache"
	"helixtrack.ru/core/internal/config"
	"helixtrack.ru/core/internal/database"
	"helixtrack.ru/core/internal/handlers"
	"helixtrack.ru/core/internal/models"
	"helixtrack.ru/core/internal/security"
	"helixtrack.ru/core/internal/services"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// TestE2E_CompleteUserJourney tests a complete user journey through the system
func TestE2E_CompleteUserJourney(t *testing.T) {
	// Setup complete application stack
	app := setupCompleteApplication(t)
	defer app.cleanup()

	// Scenario: User checks system health
	t.Run("Check System Health", func(t *testing.T) {
		resp := app.makeRequest("POST", "/do", models.Request{
			Action: models.ActionHealth,
		}, "")

		assert.Equal(t, http.StatusOK, resp.Code)

		var healthResp models.Response
		err := json.Unmarshal(resp.Body.Bytes(), &healthResp)
		require.NoError(t, err)
		assert.Equal(t, -1, healthResp.ErrorCode)
		assert.Equal(t, "healthy", healthResp.Data["status"])
	})

	// Scenario: User checks API version
	t.Run("Check API Version", func(t *testing.T) {
		resp := app.makeRequest("POST", "/do", models.Request{
			Action: models.ActionVersion,
		}, "")

		assert.Equal(t, http.StatusOK, resp.Code)

		var versionResp models.Response
		err := json.Unmarshal(resp.Body.Bytes(), &versionResp)
		require.NoError(t, err)
		assert.NotEmpty(t, versionResp.Data["version"])
	})

	// Scenario: User authenticates
	t.Run("User Authentication", func(t *testing.T) {
		resp := app.makeRequest("POST", "/do", models.Request{
			Action: models.ActionAuthenticate,
			Data: map[string]interface{}{
				"username": "testuser",
				"password": "testpass",
			},
		}, "")

		assert.Equal(t, http.StatusOK, resp.Code)

		var authResp models.Response
		err := json.Unmarshal(resp.Body.Bytes(), &authResp)
		require.NoError(t, err)
		assert.Equal(t, -1, authResp.ErrorCode)
		assert.Equal(t, "testuser", authResp.Data["username"])
	})

	// Scenario: Authenticated user creates a ticket
	t.Run("Create Ticket", func(t *testing.T) {
		token := "valid-test-token"

		resp := app.makeRequest("POST", "/do", models.Request{
			Action: models.ActionCreate,
			Object: "ticket",
			Data: map[string]interface{}{
				"title":       "Test Ticket",
				"description": "This is a test ticket",
				"priority":    "high",
			},
		}, token)

		// Should pass authentication (200 or 403 depending on permissions)
		assert.NotEqual(t, http.StatusUnauthorized, resp.Code)
	})

	// Scenario: User attempts action without authentication
	t.Run("Unauthorized Access Denied", func(t *testing.T) {
		resp := app.makeRequest("POST", "/do", models.Request{
			Action: models.ActionList,
			Object: "ticket",
		}, "")

		assert.Equal(t, http.StatusUnauthorized, resp.Code)
	})
}

// TestE2E_SecurityFullStack tests complete security stack end-to-end
func TestE2E_SecurityFullStack(t *testing.T) {
	app := setupCompleteApplication(t)
	defer app.cleanup()

	// Test 1: Malicious request is blocked
	t.Run("SQL Injection Blocked", func(t *testing.T) {
		token := "valid-test-token"

		resp := app.makeRequest("POST", "/do", models.Request{
			Action: models.ActionCreate,
			Object: "ticket",
			Data: map[string]interface{}{
				"title": "'; DROP TABLE users; --",
			},
		}, token)

		// Should be blocked by input validation
		assert.Equal(t, http.StatusBadRequest, resp.Code)
	})

	// Test 2: XSS attempt is blocked
	t.Run("XSS Attack Blocked", func(t *testing.T) {
		token := "valid-test-token"

		resp := app.makeRequest("POST", "/do", models.Request{
			Action: models.ActionCreate,
			Object: "comment",
			Data: map[string]interface{}{
				"text": "<script>alert('xss')</script>",
			},
		}, token)

		// Should be blocked by input validation
		assert.Equal(t, http.StatusBadRequest, resp.Code)
	})

	// Test 3: CSRF protection
	t.Run("CSRF Protection", func(t *testing.T) {
		// Skip: CSRF protection is tested separately in security package tests
		// E2E tests have CSRF disabled to focus on application logic
		t.Skip("CSRF protection tested in security package")
	})

	// Test 4: Rate limiting
	t.Run("Rate Limiting Works", func(t *testing.T) {
		// Send many requests quickly
		successCount := 0
		rateLimitedCount := 0

		for i := 0; i < 20; i++ {
			resp := app.makeRequest("POST", "/do", models.Request{
				Action: models.ActionVersion,
			}, "")

			if resp.Code == http.StatusOK {
				successCount++
			} else if resp.Code == http.StatusTooManyRequests {
				rateLimitedCount++
			}
		}

		// Some requests should be rate limited
		assert.Greater(t, rateLimitedCount, 0, "Rate limiting should block some requests")
	})
}

// TestE2E_DatabaseOperations tests complete database operations end-to-end
func TestE2E_DatabaseOperations(t *testing.T) {
	app := setupCompleteApplication(t)
	defer app.cleanup()

	ctx := context.Background()

	// Setup: Create test table
	_, err := app.db.Exec(ctx, `
		CREATE TABLE e2e_tickets (
			id TEXT PRIMARY KEY,
			title TEXT NOT NULL,
			status TEXT NOT NULL,
			created INTEGER NOT NULL
		)
	`)
	require.NoError(t, err)

	// Scenario: Create, Read, Update, Delete workflow
	t.Run("Complete CRUD Workflow", func(t *testing.T) {
		// Create
		ticketID := "ticket-e2e-1"
		_, err := app.db.Exec(ctx,
			"INSERT INTO e2e_tickets (id, title, status, created) VALUES (?, ?, ?, ?)",
			ticketID, "E2E Test Ticket", "open", time.Now().Unix(),
		)
		require.NoError(t, err)

		// Read
		row := app.db.QueryRow(ctx, "SELECT title, status FROM e2e_tickets WHERE id = ?", ticketID)
		var title, status string
		err = row.Scan(&title, &status)
		require.NoError(t, err)
		assert.Equal(t, "E2E Test Ticket", title)
		assert.Equal(t, "open", status)

		// Update
		_, err = app.db.Exec(ctx, "UPDATE e2e_tickets SET status = ? WHERE id = ?", "closed", ticketID)
		require.NoError(t, err)

		// Verify update
		row = app.db.QueryRow(ctx, "SELECT status FROM e2e_tickets WHERE id = ?", ticketID)
		err = row.Scan(&status)
		require.NoError(t, err)
		assert.Equal(t, "closed", status)

		// Delete
		_, err = app.db.Exec(ctx, "DELETE FROM e2e_tickets WHERE id = ?", ticketID)
		require.NoError(t, err)

		// Verify deletion
		row = app.db.QueryRow(ctx, "SELECT COUNT(*) FROM e2e_tickets WHERE id = ?", ticketID)
		var count int
		err = row.Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 0, count)
	})
}

// TestE2E_CachingLayer tests caching throughout the application
func TestE2E_CachingLayer(t *testing.T) {
	app := setupCompleteApplication(t)
	defer app.cleanup()

	ctx := context.Background()

	// Setup test data
	_, err := app.db.Exec(ctx, "CREATE TABLE e2e_cache_test (id TEXT PRIMARY KEY, data TEXT)")
	require.NoError(t, err)

	_, err = app.db.Exec(ctx, "INSERT INTO e2e_cache_test (id, data) VALUES (?, ?)", "item-1", "cached data")
	require.NoError(t, err)

	// Test caching flow
	t.Run("Cache Miss and Hit", func(t *testing.T) {
		cacheKey := "e2e:item-1"

		// First access - cache miss
		_, found := app.cache.Get(context.Background(), cacheKey)
		assert.False(t, found)

		// Read from database
		row := app.db.QueryRow(ctx, "SELECT data FROM e2e_cache_test WHERE id = ?", "item-1")
		var data string
		err := row.Scan(&data)
		require.NoError(t, err)

		// Store in cache
		app.cache.Set(context.Background(), cacheKey, data, 5*time.Minute)

		// Second access - cache hit
		cached, found := app.cache.Get(context.Background(), cacheKey)
		assert.True(t, found)
		assert.Equal(t, data, cached)
	})
}

// TestE2E_PerformanceUnderLoad tests system performance under load
func TestE2E_PerformanceUnderLoad(t *testing.T) {
	app := setupCompleteApplication(t)
	defer app.cleanup()

	// Scenario: Handle concurrent users
	t.Run("Concurrent User Requests", func(t *testing.T) {
		done := make(chan bool)
		numUsers := 50
		requestsPerUser := 10

		startTime := time.Now()

		for i := 0; i < numUsers; i++ {
			go func(userID int) {
				defer func() { done <- true }()

				for j := 0; j < requestsPerUser; j++ {
					resp := app.makeRequest("POST", "/do", models.Request{
						Action: models.ActionVersion,
					}, "")

					assert.Equal(t, http.StatusOK, resp.Code)
				}
			}(i)
		}

		// Wait for all users
		for i := 0; i < numUsers; i++ {
			<-done
		}

		duration := time.Since(startTime)
		totalRequests := numUsers * requestsPerUser

		// Calculate requests per second
		rps := float64(totalRequests) / duration.Seconds()

		// Should handle at least 100 requests per second
		assert.Greater(t, rps, 100.0, "System should handle at least 100 req/s")

		t.Logf("Performance: %d requests in %v (%.2f req/s)", totalRequests, duration, rps)
	})
}

// TestE2E_ErrorHandling tests error handling throughout the stack
func TestE2E_ErrorHandling(t *testing.T) {
	app := setupCompleteApplication(t)
	defer app.cleanup()

	scenarios := []struct {
		name           string
		request        models.Request
		token          string
		expectedStatus int
	}{
		{
			name: "Invalid JSON",
			request: models.Request{
				Action: "invalid-action",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Missing required fields",
			request: models.Request{
				Action: models.ActionCreate,
				Object: "", // Missing object
			},
			token:          "valid-test-token",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Unauthorized access",
			request: models.Request{
				Action: models.ActionList,
			},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			resp := app.makeRequest("POST", "/do", scenario.request, scenario.token)
			assert.Equal(t, scenario.expectedStatus, resp.Code)

			var errorResp models.Response
			err := json.Unmarshal(resp.Body.Bytes(), &errorResp)
			require.NoError(t, err)
			assert.NotEqual(t, -1, errorResp.ErrorCode, "Should have error code")
		})
	}
}

// Application represents the complete application for E2E testing
type Application struct {
	router      *gin.Engine
	db          database.Database
	cache       cache.Cache
	authService services.AuthService
	permService services.PermissionService
	tmpDir      string
}

func setupCompleteApplication(t *testing.T) *Application {
	// Setup database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "e2e_test.db")

	dbCfg := config.DatabaseConfig{
		Type:       "sqlite",
		SQLitePath: dbPath,
	}

	db, err := database.NewDatabase(dbCfg)
	require.NoError(t, err)

	// Setup cache
	cacheCfg := cache.DefaultCacheConfig()
	c := cache.NewInMemoryCache(cacheCfg)

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
			return username == "testuser", nil
		},
		IsEnabledFunc: func() bool {
			return true
		},
	}

	// Setup router with complete middleware stack
	router := gin.New()

	// Add security middleware
	router.Use(security.SecurityHeadersMiddleware(security.DefaultSecurityHeadersConfig()))

	// CSRF protection is tested separately in security tests
	// Disabled here to allow e2e tests to focus on application logic
	// router.Use(security.CSRFProtectionMiddleware(security.DefaultCSRFProtectionConfig()))

	rateCfg := security.DefaultDDoSProtectionConfig()
	rateCfg.MaxRequestsPerSecond = 10
	rateCfg.BurstSize = 10
	router.Use(security.DDoSProtectionMiddleware(rateCfg))

	router.Use(security.InputValidationMiddleware(security.DefaultInputValidationConfig()))

	// Setup handler
	handler := handlers.NewHandler(db, authService, permService, "1.0.0-e2e")

	// Add routes - DoAction handles all validation including JWT
	router.POST("/do", func(c *gin.Context) {
		// Parse request and set in context - let handler do all validation
		bodyBytes, _ := c.GetRawData()
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		var req models.Request
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, models.NewErrorResponse(
				models.ErrorCodeInvalidRequest,
				"Invalid request format",
				"",
			))
			return
		}

		c.Set("request", &req)
		handler.DoAction(c)
	})

	return &Application{
		router:      router,
		db:          db,
		cache:       c,
		authService: authService,
		permService: permService,
		tmpDir:      tmpDir,
	}
}

func (app *Application) cleanup() {
	app.db.Close()
	app.cache.Close()
}

func (app *Application) makeRequest(method, path string, body interface{}, token string) *httptest.ResponseRecorder {
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest(method, path, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-CSRF-Token", "valid-token")
	req.Header.Set("Cookie", "csrf_token=valid-token")

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	w := httptest.NewRecorder()
	app.router.ServeHTTP(w, req)

	return w
}

func (app *Application) makeRequestWithoutCSRF(method, path string, body interface{}) *httptest.ResponseRecorder {
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest(method, path, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	// Intentionally omit CSRF token

	w := httptest.NewRecorder()
	app.router.ServeHTTP(w, req)

	return w
}
