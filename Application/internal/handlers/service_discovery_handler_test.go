package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"helixtrack.ru/core/internal/config"
	"helixtrack.ru/core/internal/database"
	"helixtrack.ru/core/internal/models"
)

// setupServiceDiscoveryTestHandler creates test handler with service registry tables
func setupServiceDiscoveryTestHandler(t *testing.T) *ServiceDiscoveryHandler {
	db, err := database.NewDatabase(config.DatabaseConfig{Type: "sqlite", SQLitePath: ":memory:"})
	require.NoError(t, err)

	// Create service_registry table
	_, err = db.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS service_registry (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			type TEXT NOT NULL,
			version TEXT,
			url TEXT NOT NULL,
			health_check_url TEXT,
			public_key TEXT,
			signature TEXT,
			certificate TEXT,
			status TEXT,
			priority INTEGER DEFAULT 0,
			metadata TEXT,
			registered_by TEXT,
			registered_at INTEGER,
			last_health_check INTEGER,
			health_check_count INTEGER DEFAULT 0,
			failed_health_count INTEGER DEFAULT 0,
			deleted INTEGER DEFAULT 0
		)
	`)
	require.NoError(t, err)

	// Create service_rotation_audit table
	_, err = db.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS service_rotation_audit (
			id TEXT PRIMARY KEY,
			old_service_id TEXT,
			new_service_id TEXT,
			reason TEXT,
			requested_by TEXT,
			rotation_time INTEGER,
			verification_hash TEXT,
			success INTEGER,
			error_message TEXT
		)
	`)
	require.NoError(t, err)

	// Create service_health_check table
	_, err = db.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS service_health_check (
			id TEXT PRIMARY KEY,
			service_id TEXT NOT NULL,
			check_time INTEGER NOT NULL,
			status TEXT NOT NULL,
			response_time_ms INTEGER,
			error_message TEXT
		)
	`)
	require.NoError(t, err)

	handler, err := NewServiceDiscoveryHandler(db)
	require.NoError(t, err)

	return handler
}

// TestServiceDiscoveryHandler_RegisterService_MissingAdminToken tests registration without admin token
func TestServiceDiscoveryHandler_RegisterService_MissingAdminToken(t *testing.T) {
	handler := setupServiceDiscoveryTestHandler(t)
	gin.SetMode(gin.TestMode)

	reqBody := models.ServiceRegistrationRequest{
		Name:           "Test Service",
		Type:           "authentication",
		Version:        "1.0.0",
		URL:            "http://localhost:8080",
		HealthCheckURL: "http://localhost:8080/health",
		AdminToken:     "", // Missing
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/services/register", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req

	handler.RegisterService(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestServiceDiscoveryHandler_DiscoverServices_EmptyList tests discovery when no services exist
func TestServiceDiscoveryHandler_DiscoverServices_EmptyList(t *testing.T) {
	handler := setupServiceDiscoveryTestHandler(t)
	gin.SetMode(gin.TestMode)

	reqBody := models.ServiceDiscoveryRequest{
		Type:        "authentication",
		OnlyHealthy: false,
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/services/discover", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req

	handler.DiscoverServices(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

// TestServiceDiscoveryHandler_DiscoverServices_FilterByType tests discovery with type filter
func TestServiceDiscoveryHandler_DiscoverServices_FilterByType(t *testing.T) {
	handler := setupServiceDiscoveryTestHandler(t)
	gin.SetMode(gin.TestMode)

	// Insert test services
	now := time.Now().Unix()
	_, err := handler.db.Exec(context.Background(), `
		INSERT INTO service_registry (id, name, type, version, url, status, priority, registered_by, registered_at, deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, generateTestID(), "Auth Service", "authentication", "1.0.0", "http://localhost:8080", "healthy", 10, "admin", now, 0)
	require.NoError(t, err)

	_, err = handler.db.Exec(context.Background(), `
		INSERT INTO service_registry (id, name, type, version, url, status, priority, registered_by, registered_at, deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, generateTestID(), "Perm Service", "permissions", "1.0.0", "http://localhost:8081", "healthy", 10, "admin", now, 0)
	require.NoError(t, err)

	reqBody := models.ServiceDiscoveryRequest{
		Type:        "authentication",
		OnlyHealthy: false,
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/services/discover", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req

	handler.DiscoverServices(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Response
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	dataMap := response.Data
	discovery := dataMap["discovery"].(map[string]interface{})
	totalCount := int(discovery["TotalCount"].(float64))
	assert.Equal(t, 1, totalCount) // Only authentication service
}

// TestServiceDiscoveryHandler_DecommissionService_Success tests decommissioning a service
func TestServiceDiscoveryHandler_DecommissionService_Success(t *testing.T) {
	handler := setupServiceDiscoveryTestHandler(t)
	gin.SetMode(gin.TestMode)

	// Insert test service
	serviceID := "test-service-id"
	now := time.Now().Unix()
	_, err := handler.db.Exec(context.Background(), `
		INSERT INTO service_registry (id, name, type, version, url, status, priority, registered_by, registered_at, deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, serviceID, "Test Service", "authentication", "1.0.0", "http://localhost:8080", "healthy", 10, "admin", now, 0)
	require.NoError(t, err)

	reqBody := models.ServiceDecommissionRequest{
		ServiceID:  serviceID,
		Reason:     "Testing decommission",
		AdminToken: strings.Repeat("a", 32), // Valid length
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/services/decommission", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req

	handler.DecommissionService(c)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify service is decommissioned
	var deleted int
	err = handler.db.QueryRow(context.Background(), "SELECT deleted FROM service_registry WHERE id = ?", serviceID).Scan(&deleted)
	require.NoError(t, err)
	assert.Equal(t, 1, deleted)
}

// TestServiceDiscoveryHandler_DecommissionService_NotFound tests decommissioning non-existent service
func TestServiceDiscoveryHandler_DecommissionService_NotFound(t *testing.T) {
	handler := setupServiceDiscoveryTestHandler(t)
	gin.SetMode(gin.TestMode)

	reqBody := models.ServiceDecommissionRequest{
		ServiceID:  "non-existent-id",
		Reason:     "Testing",
		AdminToken: strings.Repeat("a", 32),
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/services/decommission", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req

	handler.DecommissionService(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// TestServiceDiscoveryHandler_GetServiceHealth_Success tests getting service health
func TestServiceDiscoveryHandler_GetServiceHealth_Success(t *testing.T) {
	handler := setupServiceDiscoveryTestHandler(t)
	gin.SetMode(gin.TestMode)

	// Insert test service
	serviceID := "test-service-id"
	now := time.Now().Unix()
	_, err := handler.db.Exec(context.Background(), `
		INSERT INTO service_registry (id, name, type, version, url, status, priority, registered_by, registered_at, health_check_count, failed_health_count, deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, serviceID, "Test Service", "authentication", "1.0.0", "http://localhost:8080", "healthy", 10, "admin", now, 5, 0, 0)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/services/"+serviceID+"/health", nil)
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = []gin.Param{{Key: "id", Value: serviceID}}

	handler.GetServiceHealth(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

// TestServiceDiscoveryHandler_GetServiceHealth_NotFound tests getting health for non-existent service
func TestServiceDiscoveryHandler_GetServiceHealth_NotFound(t *testing.T) {
	handler := setupServiceDiscoveryTestHandler(t)
	gin.SetMode(gin.TestMode)

	req := httptest.NewRequest(http.MethodGet, "/services/non-existent/health", nil)
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = []gin.Param{{Key: "id", Value: "non-existent-id"}}

	handler.GetServiceHealth(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// TestServiceDiscoveryHandler_ListServices_EmptyList tests listing when no services exist
func TestServiceDiscoveryHandler_ListServices_EmptyList(t *testing.T) {
	handler := setupServiceDiscoveryTestHandler(t)
	gin.SetMode(gin.TestMode)

	req := httptest.NewRequest(http.MethodGet, "/services", nil)
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req

	handler.ListServices(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	dataMap := response.Data
	total := int(dataMap["total"].(float64))
	assert.Equal(t, 0, total)
}

// TestServiceDiscoveryHandler_ListServices_MultipleServices tests listing multiple services
func TestServiceDiscoveryHandler_ListServices_MultipleServices(t *testing.T) {
	handler := setupServiceDiscoveryTestHandler(t)
	gin.SetMode(gin.TestMode)

	// Insert test services
	now := time.Now().Unix()
	for i := 0; i < 3; i++ {
		_, err := handler.db.Exec(context.Background(), `
			INSERT INTO service_registry (id, name, type, version, url, status, priority, registered_at, last_health_check, health_check_count, failed_health_count, deleted)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, generateTestID(), "Service "+string(rune('A'+i)), "authentication", "1.0.0", "http://localhost:808"+string(rune('0'+i)), "healthy", 10, now, now, 5, 0, 0)
		require.NoError(t, err)
	}

	req := httptest.NewRequest(http.MethodGet, "/services", nil)
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req

	handler.ListServices(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	dataMap := response.Data
	total := int(dataMap["total"].(float64))
	assert.Equal(t, 3, total)
}

// TestServiceDiscoveryHandler_UpdateService_NotFound tests updating non-existent service
func TestServiceDiscoveryHandler_UpdateService_NotFound(t *testing.T) {
	handler := setupServiceDiscoveryTestHandler(t)
	gin.SetMode(gin.TestMode)

	reqBody := models.ServiceUpdateRequest{
		ServiceID:  "non-existent-id",
		Version:    "2.0.0",
		AdminToken: strings.Repeat("a", 32),
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/services/update", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req

	handler.UpdateService(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// TestServiceDiscoveryHandler_UpdateService_NoFields tests update with no fields
func TestServiceDiscoveryHandler_UpdateService_NoFields(t *testing.T) {
	handler := setupServiceDiscoveryTestHandler(t)
	gin.SetMode(gin.TestMode)

	reqBody := models.ServiceUpdateRequest{
		ServiceID:  "test-id",
		AdminToken: strings.Repeat("a", 32),
		// No fields to update
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/services/update", bytes.NewReader(body))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req

	handler.UpdateService(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
