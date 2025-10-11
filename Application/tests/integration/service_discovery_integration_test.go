package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"helixtrack.ru/core/internal/config"
	"helixtrack.ru/core/internal/database"
	"helixtrack.ru/core/internal/handlers"
	"helixtrack.ru/core/internal/models"
	"helixtrack.ru/core/internal/security"
)

func setupServiceDiscoveryTest(t *testing.T) (*gin.Engine, database.Database, *security.ServiceSigner) {
	// Create in-memory database
	dbCfg := config.DatabaseConfig{
		Type:       "sqlite",
		SQLitePath: ":memory:",
	}
	db, err := database.NewDatabase(dbCfg)
	require.NoError(t, err)

	// Initialize tables
	err = handlers.InitializeServiceDiscoveryTables(db)
	require.NoError(t, err)

	// Create service signer
	signer, err := security.NewServiceSigner()
	require.NoError(t, err)

	// Create handler
	handler, err := handlers.NewServiceDiscoveryHandler(db)
	require.NoError(t, err)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Register routes
	api := router.Group("/api/services")
	{
		api.POST("/register", handler.RegisterService)
		api.POST("/discover", handler.DiscoverServices)
		api.POST("/rotate", handler.RotateService)
		api.POST("/decommission", handler.DecommissionService)
		api.POST("/update", handler.UpdateService)
		api.GET("/list", handler.ListServices)
		api.GET("/health/:id", handler.GetServiceHealth)
	}

	return router, db, signer
}

func TestServiceDiscovery_RegisterService(t *testing.T) {
	router, db, signer := setupServiceDiscoveryTest(t)
	defer db.Close()

	t.Run("Successful service registration", func(t *testing.T) {
		service := &models.ServiceRegistration{
			Name:           "Auth Service",
			Type:           models.ServiceTypeAuthentication,
			Version:        "1.0.0",
			URL:            "http://localhost:8081",
			HealthCheckURL: "http://localhost:8081/health",
			Role:           models.ServiceRolePrimary,
			FailoverGroup:  "auth-group-1",
			Priority:       10,
			Metadata:       "{}",
			RegisteredAt:   time.Now(),
		}

		err := signer.SignServiceRegistration(service)
		require.NoError(t, err)

		reqBody := models.ServiceRegistrationRequest{
			Name:           service.Name,
			Type:           service.Type,
			Version:        service.Version,
			URL:            service.URL,
			HealthCheckURL: service.HealthCheckURL,
			PublicKey:      service.PublicKey,
			Certificate:    "",
			Role:           service.Role,
			FailoverGroup:  service.FailoverGroup,
			Priority:       service.Priority,
			Metadata:       service.Metadata,
			AdminToken:     "valid-admin-token-with-32-characters-minimum-length",
		}

		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/services/register", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response models.Response
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, -1, response.ErrorCode)
	})

	t.Run("Reject registration with short admin token", func(t *testing.T) {
		reqBody := models.ServiceRegistrationRequest{
			Name:           "Test Service",
			Type:           models.ServiceTypeAuthentication,
			Version:        "1.0.0",
			URL:            "http://localhost:8082",
			HealthCheckURL: "http://localhost:8082/health",
			PublicKey:      "key",
			Role:           models.ServiceRolePrimary,
			Priority:       5,
			Metadata:       "{}",
			AdminToken:     "short", // Too short
		}

		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/services/register", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Reject duplicate service registration", func(t *testing.T) {
		service := &models.ServiceRegistration{
			Name:           "Duplicate Service",
			Type:           models.ServiceTypePermissions,
			Version:        "1.0.0",
			URL:            "http://localhost:8083",
			HealthCheckURL: "http://localhost:8083/health",
			Role:           models.ServiceRolePrimary,
			Priority:       5,
			Metadata:       "{}",
			RegisteredAt:   time.Now(),
		}

		err := signer.SignServiceRegistration(service)
		require.NoError(t, err)

		reqBody := models.ServiceRegistrationRequest{
			Name:           service.Name,
			Type:           service.Type,
			Version:        service.Version,
			URL:            service.URL,
			HealthCheckURL: service.HealthCheckURL,
			PublicKey:      service.PublicKey,
			Role:           service.Role,
			Priority:       service.Priority,
			Metadata:       service.Metadata,
			AdminToken:     "valid-admin-token-with-32-characters-minimum",
		}

		// First registration
		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/services/register", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusCreated, w.Code)

		// Duplicate registration
		req = httptest.NewRequest(http.MethodPost, "/api/services/register", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusConflict, w.Code)
	})
}

func TestServiceDiscovery_DiscoverServices(t *testing.T) {
	router, db, signer := setupServiceDiscoveryTest(t)
	defer db.Close()

	// Register a service first
	service := &models.ServiceRegistration{
		Name:           "Discoverable Service",
		Type:           models.ServiceTypeLokalization,
		Version:        "1.0.0",
		URL:            "http://localhost:8084",
		HealthCheckURL: "http://localhost:8084/health",
		Role:           models.ServiceRolePrimary,
		Status:         models.ServiceStatusHealthy,
		Priority:       10,
		Metadata:       "{}",
		RegisteredAt:   time.Now(),
	}

	err := signer.SignServiceRegistration(service)
	require.NoError(t, err)

	// Insert directly into database
	ctx := context.Background()
	_, err = db.Exec(ctx, `
		INSERT INTO service_registry (id, name, type, version, url, health_check_url, public_key, signature,
			status, role, is_active, priority, metadata, registered_by, registered_at, deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, "svc-1", service.Name, service.Type, service.Version, service.URL, service.HealthCheckURL,
		service.PublicKey, service.Signature, service.Status, service.Role, 1, service.Priority,
		service.Metadata, "system", time.Now().Unix(), 0)
	require.NoError(t, err)

	t.Run("Discover services by type", func(t *testing.T) {
		reqBody := models.ServiceDiscoveryRequest{
			Type:        models.ServiceTypeLokalization,
			OnlyHealthy: true,
		}

		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/services/discover", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response models.ServiceDiscoveryResponse
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Greater(t, response.TotalCount, 0)
		assert.NotEmpty(t, response.Services)
	})

	t.Run("No services found for unknown type", func(t *testing.T) {
		reqBody := models.ServiceDiscoveryRequest{
			Type:        models.ServiceTypeExtension,
			OnlyHealthy: true,
		}

		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/services/discover", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response models.ServiceDiscoveryResponse
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, 0, response.TotalCount)
	})
}

func TestServiceDiscovery_RotateService(t *testing.T) {
	router, db, signer := setupServiceDiscoveryTest(t)
	defer db.Close()

	// Register old service
	oldService := &models.ServiceRegistration{
		ID:             "old-svc",
		Name:           "Old Service",
		Type:           models.ServiceTypeAuthentication,
		Version:        "1.0.0",
		URL:            "http://localhost:8085",
		HealthCheckURL: "http://localhost:8085/health",
		Role:           models.ServiceRolePrimary,
		Status:         models.ServiceStatusHealthy,
		Priority:       5,
		Metadata:       "{}",
		RegisteredAt:   time.Now().Add(-10 * time.Minute),
	}

	err := signer.SignServiceRegistration(oldService)
	require.NoError(t, err)

	ctx := context.Background()
	_, err = db.Exec(ctx, `
		INSERT INTO service_registry (id, name, type, version, url, health_check_url, public_key, signature,
			status, role, is_active, priority, metadata, registered_by, registered_at, deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, oldService.ID, oldService.Name, oldService.Type, oldService.Version, oldService.URL,
		oldService.HealthCheckURL, oldService.PublicKey, oldService.Signature, oldService.Status,
		oldService.Role, 1, oldService.Priority, oldService.Metadata, "system",
		oldService.RegisteredAt.Unix(), 0)
	require.NoError(t, err)

	t.Run("Successful service rotation", func(t *testing.T) {
		newService := models.ServiceRegistration{
			Name:           "New Service",
			Type:           models.ServiceTypeAuthentication,
			Version:        "1.1.0",
			URL:            "http://localhost:8086",
			HealthCheckURL: "http://localhost:8086/health",
			Role:           models.ServiceRolePrimary,
			Status:         models.ServiceStatusHealthy,
			Priority:       10,
			Metadata:       "{}",
			RegisteredAt:   time.Now().Add(-10 * time.Minute),
		}

		err := signer.SignServiceRegistration(&newService)
		require.NoError(t, err)

		reqBody := models.ServiceRotationRequest{
			CurrentServiceID: oldService.ID,
			NewService:       newService,
			Reason:           "Upgrade to version 1.1.0",
			RequestedBy:      "admin",
			AdminToken:       "valid-admin-token-with-32-characters-minimum-length",
			VerificationCode: "verification-code-123",
		}

		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/services/rotate", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response models.ServiceRotationResponse
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.True(t, response.Success)
		assert.NotEmpty(t, response.NewServiceID)
	})

	t.Run("Reject rotation with mismatched types", func(t *testing.T) {
		newService := models.ServiceRegistration{
			Name:           "Wrong Type Service",
			Type:           models.ServiceTypePermissions, // Different type
			Version:        "1.1.0",
			URL:            "http://localhost:8087",
			HealthCheckURL: "http://localhost:8087/health",
			Role:           models.ServiceRolePrimary,
			Status:         models.ServiceStatusHealthy,
			Priority:       10,
			Metadata:       "{}",
			RegisteredAt:   time.Now().Add(-10 * time.Minute),
		}

		err := signer.SignServiceRegistration(&newService)
		require.NoError(t, err)

		reqBody := models.ServiceRotationRequest{
			CurrentServiceID: oldService.ID,
			NewService:       newService,
			Reason:           "Invalid rotation",
			RequestedBy:      "admin",
			AdminToken:       "valid-admin-token-with-32-characters-minimum",
			VerificationCode: "code",
		}

		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/services/rotate", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestServiceDiscovery_DecommissionService(t *testing.T) {
	router, db, signer := setupServiceDiscoveryTest(t)
	defer db.Close()

	// Register service
	service := &models.ServiceRegistration{
		ID:             "decomm-svc",
		Name:           "Decommission Service",
		Type:           models.ServiceTypeAuthentication,
		Version:        "1.0.0",
		URL:            "http://localhost:8088",
		HealthCheckURL: "http://localhost:8088/health",
		Role:           models.ServiceRolePrimary,
		Status:         models.ServiceStatusHealthy,
		Priority:       5,
		Metadata:       "{}",
		RegisteredAt:   time.Now(),
	}

	err := signer.SignServiceRegistration(service)
	require.NoError(t, err)

	ctx := context.Background()
	_, err = db.Exec(ctx, `
		INSERT INTO service_registry (id, name, type, version, url, health_check_url, public_key, signature,
			status, role, is_active, priority, metadata, registered_by, registered_at, deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, service.ID, service.Name, service.Type, service.Version, service.URL, service.HealthCheckURL,
		service.PublicKey, service.Signature, service.Status, service.Role, 1, service.Priority,
		service.Metadata, "system", service.RegisteredAt.Unix(), 0)
	require.NoError(t, err)

	t.Run("Successful decommission", func(t *testing.T) {
		reqBody := models.ServiceDecommissionRequest{
			ServiceID:  service.ID,
			Reason:     "End of life",
			AdminToken: "valid-admin-token-with-32-characters-minimum",
		}

		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/services/decommission", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response models.Response
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, -1, response.ErrorCode)
	})

	t.Run("Decommission nonexistent service", func(t *testing.T) {
		reqBody := models.ServiceDecommissionRequest{
			ServiceID:  "nonexistent",
			Reason:     "Test",
			AdminToken: "valid-admin-token-with-32-characters-minimum",
		}

		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/services/decommission", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestServiceDiscovery_ListServices(t *testing.T) {
	router, db, signer := setupServiceDiscoveryTest(t)
	defer db.Close()

	// Register multiple services
	for i := 0; i < 3; i++ {
		service := &models.ServiceRegistration{
			ID:             "list-svc-" + string(rune('a'+i)),
			Name:           "List Service " + string(rune('A'+i)),
			Type:           models.ServiceTypeAuthentication,
			Version:        "1.0.0",
			URL:            "http://localhost:808" + string(rune('0'+i)),
			HealthCheckURL: "http://localhost:808" + string(rune('0'+i)) + "/health",
			Role:           models.ServiceRolePrimary,
			Status:         models.ServiceStatusHealthy,
			Priority:       i,
			Metadata:       "{}",
			RegisteredAt:   time.Now(),
		}

		err := signer.SignServiceRegistration(service)
		require.NoError(t, err)

		ctx := context.Background()
		_, err = db.Exec(ctx, `
			INSERT INTO service_registry (id, name, type, version, url, health_check_url, public_key, signature,
				status, role, is_active, priority, metadata, registered_by, registered_at, deleted)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, service.ID, service.Name, service.Type, service.Version, service.URL, service.HealthCheckURL,
			service.PublicKey, service.Signature, service.Status, service.Role, 1, service.Priority,
			service.Metadata, "system", service.RegisteredAt.Unix(), 0)
		require.NoError(t, err)
	}

	t.Run("List all services", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/services/list", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response models.ServiceDiscoveryResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, response.TotalCount, 3)
	})
}

func TestServiceDiscovery_GetServiceHealth(t *testing.T) {
	router, db, signer := setupServiceDiscoveryTest(t)
	defer db.Close()

	// Register service
	service := &models.ServiceRegistration{
		ID:             "health-svc",
		Name:           "Health Service",
		Type:           models.ServiceTypeAuthentication,
		Version:        "1.0.0",
		URL:            "http://localhost:8089",
		HealthCheckURL: "http://localhost:8089/health",
		Role:           models.ServiceRolePrimary,
		Status:         models.ServiceStatusHealthy,
		Priority:       5,
		Metadata:       "{}",
		RegisteredAt:   time.Now(),
	}

	err := signer.SignServiceRegistration(service)
	require.NoError(t, err)

	ctx := context.Background()
	_, err = db.Exec(ctx, `
		INSERT INTO service_registry (id, name, type, version, url, health_check_url, public_key, signature,
			status, role, is_active, priority, metadata, registered_by, registered_at, deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, service.ID, service.Name, service.Type, service.Version, service.URL, service.HealthCheckURL,
		service.PublicKey, service.Signature, service.Status, service.Role, 1, service.Priority,
		service.Metadata, "system", service.RegisteredAt.Unix(), 0)
	require.NoError(t, err)

	t.Run("Get service health", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/services/health/"+service.ID, nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Get health for nonexistent service", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/services/health/nonexistent", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}
