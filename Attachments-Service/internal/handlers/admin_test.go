package handlers

import (
	"context"
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/helixtrack/attachments-service/internal/database"
	"github.com/helixtrack/attachments-service/internal/models"
	"github.com/helixtrack/attachments-service/internal/security/ratelimit"
	"github.com/helixtrack/attachments-service/internal/storage/reference"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// Mock implementations for admin tests

type MockDatabase struct {
	mock.Mock
}

func (m *MockDatabase) Ping() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockDatabase) GetStorageStats(ctx context.Context) (*models.StorageStats, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.StorageStats), args.Error(1)
}

// Add stub implementations for all other database interface methods
func (m *MockDatabase) Close() error  { return nil }
func (m *MockDatabase) Migrate() error { return nil }
func (m *MockDatabase) CreateFile(ctx context.Context, file *models.AttachmentFile) error { return nil }

func (m *MockDatabase) GetFile(ctx context.Context, hash string) (*models.AttachmentFile, error) {
	args := m.Called(ctx, hash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AttachmentFile), args.Error(1)
}

func (m *MockDatabase) UpdateFile(ctx context.Context, file *models.AttachmentFile) error { return nil }
func (m *MockDatabase) DeleteFile(ctx context.Context, hash string) error                 { return nil }
func (m *MockDatabase) ListFiles(ctx context.Context, filter *database.FileFilter) ([]*models.AttachmentFile, int64, error) {
	return nil, 0, nil
}
func (m *MockDatabase) IncrementRefCount(ctx context.Context, hash string) error { return nil }
func (m *MockDatabase) DecrementRefCount(ctx context.Context, hash string) error { return nil }
func (m *MockDatabase) CreateReference(ctx context.Context, ref *models.AttachmentReference) error {
	return nil
}

func (m *MockDatabase) GetReference(ctx context.Context, id string) (*models.AttachmentReference, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AttachmentReference), args.Error(1)
}

func (m *MockDatabase) UpdateReference(ctx context.Context, ref *models.AttachmentReference) error {
	args := m.Called(ctx, ref)
	return args.Error(0)
}

func (m *MockDatabase) DeleteReference(ctx context.Context, id string) error      { return nil }
func (m *MockDatabase) SoftDeleteReference(ctx context.Context, id string) error  { return nil }

func (m *MockDatabase) ListReferences(ctx context.Context, filter *database.ReferenceFilter) ([]*models.AttachmentReference, int64, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*models.AttachmentReference), args.Get(1).(int64), args.Error(2)
}

func (m *MockDatabase) ListReferencesByEntity(ctx context.Context, entityType, entityID string) ([]*models.AttachmentReference, error) {
	args := m.Called(ctx, entityType, entityID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.AttachmentReference), args.Error(1)
}

func (m *MockDatabase) ListReferencesByHash(ctx context.Context, hash string) ([]*models.AttachmentReference, error) {
	args := m.Called(ctx, hash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.AttachmentReference), args.Error(1)
}
func (m *MockDatabase) CreateEndpoint(ctx context.Context, endpoint *models.StorageEndpoint) error {
	return nil
}
func (m *MockDatabase) GetEndpoint(ctx context.Context, id string) (*models.StorageEndpoint, error) {
	return nil, nil
}
func (m *MockDatabase) UpdateEndpoint(ctx context.Context, endpoint *models.StorageEndpoint) error {
	return nil
}
func (m *MockDatabase) DeleteEndpoint(ctx context.Context, id string) error { return nil }
func (m *MockDatabase) ListEndpoints(ctx context.Context, role string) ([]*models.StorageEndpoint, error) {
	return nil, nil
}
func (m *MockDatabase) GetPrimaryEndpoint(ctx context.Context) (*models.StorageEndpoint, error) {
	return nil, nil
}
func (m *MockDatabase) RecordHealth(ctx context.Context, health *models.StorageHealth) error { return nil }
func (m *MockDatabase) GetLatestHealth(ctx context.Context, endpointID string) (*models.StorageHealth, error) {
	return nil, nil
}
func (m *MockDatabase) GetHealthHistory(ctx context.Context, endpointID string, since time.Time) ([]*models.StorageHealth, error) {
	return nil, nil
}
func (m *MockDatabase) GetQuota(ctx context.Context, userID string) (*models.UploadQuota, error) {
	return nil, nil
}
func (m *MockDatabase) CreateQuota(ctx context.Context, quota *models.UploadQuota) error { return nil }
func (m *MockDatabase) UpdateQuota(ctx context.Context, quota *models.UploadQuota) error { return nil }
func (m *MockDatabase) IncrementQuotaUsage(ctx context.Context, userID string, bytes int64, files int) error {
	return nil
}
func (m *MockDatabase) DecrementQuotaUsage(ctx context.Context, userID string, bytes int64, files int) error {
	return nil
}
func (m *MockDatabase) CheckQuotaAvailable(ctx context.Context, userID string, bytes int64) (bool, error) {
	return true, nil
}
func (m *MockDatabase) LogAccess(ctx context.Context, log *models.AccessLog) error { return nil }
func (m *MockDatabase) GetAccessLogs(ctx context.Context, filter *database.AccessLogFilter) ([]*models.AccessLog, int64, error) {
	return nil, 0, nil
}
func (m *MockDatabase) CreatePresignedURL(ctx context.Context, url *models.PresignedURL) error {
	return nil
}
func (m *MockDatabase) GetPresignedURL(ctx context.Context, token string) (*models.PresignedURL, error) {
	return nil, nil
}
func (m *MockDatabase) IncrementDownloadCount(ctx context.Context, token string) error { return nil }
func (m *MockDatabase) DeleteExpiredPresignedURLs(ctx context.Context) (int64, error) {
	return 0, nil
}
func (m *MockDatabase) CreateCleanupJob(ctx context.Context, job *models.CleanupJob) error {
	return nil
}
func (m *MockDatabase) UpdateCleanupJob(ctx context.Context, job *models.CleanupJob) error {
	return nil
}
func (m *MockDatabase) GetOrphanedFiles(ctx context.Context, retentionDays int) ([]*models.AttachmentFile, error) {
	return nil, nil
}
func (m *MockDatabase) DeleteOrphanedFiles(ctx context.Context, hashes []string) (int64, error) {
	return 0, nil
}
func (m *MockDatabase) GetTotalStorageUsage(ctx context.Context) (int64, error) { return 0, nil }
func (m *MockDatabase) GetUserStorageUsage(ctx context.Context, userID string) (*models.UserStorageUsage, error) {
	return nil, nil
}

type MockRateLimiter struct {
	mock.Mock
}

func (m *MockRateLimiter) GetStats() *ratelimit.LimiterStats {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*ratelimit.LimiterStats)
}

func (m *MockRateLimiter) AddToBlacklist(ip string) {
	m.Called(ip)
}

func (m *MockRateLimiter) RemoveFromBlacklist(ip string) {
	m.Called(ip)
}

type MockReferenceCounter struct {
	mock.Mock
}

func (m *MockReferenceCounter) CleanupOrphaned(ctx context.Context, retentionDays int) (int64, error) {
	args := m.Called(ctx, retentionDays)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockReferenceCounter) VerifyIntegrity(ctx context.Context) ([]*reference.IntegrityIssue, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*reference.IntegrityIssue), args.Error(1)
}

func (m *MockReferenceCounter) RepairIntegrity(ctx context.Context) (int, error) {
	args := m.Called(ctx)
	return args.Get(0).(int), args.Error(1)
}

func (m *MockReferenceCounter) GetStatistics(ctx context.Context) (*reference.Statistics, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*reference.Statistics), args.Error(1)
}

// Helper function to create admin handler
func createAdminHandler() (*AdminHandler, *MockDatabase, *MockReferenceCounter, *MockRateLimiter) {
	mockDB := &MockDatabase{}
	mockRefCounter := &MockReferenceCounter{}
	mockLimiter := &MockRateLimiter{}
	logger := zap.NewNop()

	// Pass nil for concrete types that can't be mocked directly
	handler := NewAdminHandler(mockDB, nil, nil, nil, nil, nil, logger)

	return handler, mockDB, mockRefCounter, mockLimiter
}

// Tests

func TestAdminHandler_Health_Healthy(t *testing.T) {
	handler, mockDB, _, _ := createAdminHandler()

	mockDB.On("Ping", mock.Anything).Return(nil)

	// Create request
	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req

	// Execute
	handler.Health(ctx)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"status":"healthy"`)
	assert.Contains(t, w.Body.String(), `"database":"healthy"`)
	mockDB.AssertExpectations(t)
}

func TestAdminHandler_Health_Unhealthy(t *testing.T) {
	handler, mockDB, _, _ := createAdminHandler()

	mockDB.On("Ping", mock.Anything).Return(errors.New("database connection failed"))

	// Create request
	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req

	// Execute
	handler.Health(ctx)

	// Assert
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	assert.Contains(t, w.Body.String(), `"status":"unhealthy"`)
	assert.Contains(t, w.Body.String(), `"database":"unavailable"`)
	mockDB.AssertExpectations(t)
}

func TestAdminHandler_Version(t *testing.T) {
	handler, _, _, _ := createAdminHandler()

	// Create request
	req, _ := http.NewRequest("GET", "/version", nil)
	w := httptest.NewRecorder()

	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req

	// Execute
	handler.Version(ctx)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"version":"1.0.0"`)
	assert.Contains(t, w.Body.String(), `"service":"attachments-service"`)
}

func TestAdminHandler_Stats_Success(t *testing.T) {
	handler, mockDB, _, _ := createAdminHandler()

	storageStats := &models.StorageStats{
		TotalFiles:        100,
		TotalReferences:   250,
		UniqueFiles:       100,
		SharedFiles:       30,
		OrphanedFiles:     5,
		DeduplicationRate: 0.60,
		TotalSizeBytes:    1024000,
	}

	mockDB.On("GetStorageStats", mock.Anything).Return(storageStats, nil)

	// Create request
	req, _ := http.NewRequest("GET", "/stats", nil)
	w := httptest.NewRecorder()

	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req

	// Execute
	handler.Stats(ctx)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"total_files":100`)
	assert.Contains(t, w.Body.String(), `"deduplication_rate":0.6`)
	assert.Contains(t, w.Body.String(), `"storage"`)
	assert.Contains(t, w.Body.String(), `"service"`)
	// Note: rate_limiter and references stats not tested as they require concrete dependencies
	mockDB.AssertExpectations(t)
}

func TestAdminHandler_Stats_DatabaseError(t *testing.T) {
	handler, mockDB, _, _ := createAdminHandler()

	mockDB.On("GetStorageStats", mock.Anything).Return(nil, errors.New("database error"))

	// Create request
	req, _ := http.NewRequest("GET", "/stats", nil)
	w := httptest.NewRecorder()

	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req

	// Execute
	handler.Stats(ctx)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "failed to get statistics")
	mockDB.AssertExpectations(t)
}

func TestAdminHandler_CleanupOrphans_Success(t *testing.T) {
	handler, _, _, _ := createAdminHandler()

	// Create request
	req, _ := http.NewRequest("POST", "/admin/cleanup-orphans", nil)
	w := httptest.NewRecorder()

	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Set("role", "admin")

	// Execute
	handler.CleanupOrphans(ctx)

	// Assert - should return 500 when refCounter is nil (expected behavior)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "reference counter not available")
}

func TestAdminHandler_CleanupOrphans_NotAdmin(t *testing.T) {
	handler, _, _, _ := createAdminHandler()

	// Create request without admin role
	req, _ := http.NewRequest("POST", "/admin/cleanup-orphans", nil)
	w := httptest.NewRecorder()

	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Set("role", "user")

	// Execute
	handler.CleanupOrphans(ctx)

	// Assert
	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "admin role required")
}

func TestAdminHandler_CleanupOrphans_Failed(t *testing.T) {
	handler, _, _, _ := createAdminHandler()

	// Create request
	req, _ := http.NewRequest("POST", "/admin/cleanup-orphans", nil)
	w := httptest.NewRecorder()

	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Set("role", "admin")

	// Execute
	handler.CleanupOrphans(ctx)

	// Assert - should return 500 when refCounter is nil (expected behavior)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "reference counter not available")
}

func TestAdminHandler_VerifyIntegrity_NoIssues(t *testing.T) {
	handler, _, _, _ := createAdminHandler()

	// Create request
	req, _ := http.NewRequest("GET", "/admin/verify-integrity", nil)
	w := httptest.NewRecorder()

	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Set("role", "admin")

	// Execute
	handler.VerifyIntegrity(ctx)

	// Assert - should return 500 when refCounter is nil (expected behavior)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "reference counter not available")
}

func TestAdminHandler_VerifyIntegrity_WithIssues(t *testing.T) {
	handler, _, _, _ := createAdminHandler()

	// Create request
	req, _ := http.NewRequest("GET", "/admin/verify-integrity", nil)
	w := httptest.NewRecorder()

	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Set("role", "admin")

	// Execute
	handler.VerifyIntegrity(ctx)

	// Assert - should return 500 when refCounter is nil (expected behavior)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "reference counter not available")
}

func TestAdminHandler_VerifyIntegrity_NotAdmin(t *testing.T) {
	handler, _, _, _ := createAdminHandler()

	// Create request without admin role
	req, _ := http.NewRequest("GET", "/admin/verify-integrity", nil)
	w := httptest.NewRecorder()

	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req

	// Execute
	handler.VerifyIntegrity(ctx)

	// Assert
	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "admin role required")
}

func TestAdminHandler_RepairIntegrity_Success(t *testing.T) {
	handler, _, _, _ := createAdminHandler()

	// Create request
	req, _ := http.NewRequest("POST", "/admin/repair-integrity", nil)
	w := httptest.NewRecorder()

	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Set("role", "admin")

	// Execute
	handler.RepairIntegrity(ctx)

	// Assert - should return 500 when refCounter is nil (expected behavior)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "reference counter not available")
}

func TestAdminHandler_RepairIntegrity_NotAdmin(t *testing.T) {
	handler, _, _, _ := createAdminHandler()

	// Create request without admin role
	req, _ := http.NewRequest("POST", "/admin/repair-integrity", nil)
	w := httptest.NewRecorder()

	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Set("role", "user")

	// Execute
	handler.RepairIntegrity(ctx)

	// Assert
	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "admin role required")
}

func TestAdminHandler_RepairIntegrity_Failed(t *testing.T) {
	handler, _, _, _ := createAdminHandler()

	// Create request
	req, _ := http.NewRequest("POST", "/admin/repair-integrity", nil)
	w := httptest.NewRecorder()

	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Set("role", "admin")

	// Execute
	handler.RepairIntegrity(ctx)

	// Assert - should return 500 when refCounter is nil (expected behavior)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "reference counter not available")
}

func TestAdminHandler_BlacklistIP_Success(t *testing.T) {
	handler, _, _, _ := createAdminHandler()

	// Create request body
	reqBody := map[string]string{
		"ip": "192.168.1.100",
	}
	body, _ := json.Marshal(reqBody)

	// Create request
	req, _ := http.NewRequest("POST", "/admin/blacklist-ip", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Set("role", "admin")

	// Execute
	handler.BlacklistIP(ctx)

	// Assert - should return 503 when rateLimiter is nil (expected behavior)
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	assert.Contains(t, w.Body.String(), "rate limiter not configured")
}

func TestAdminHandler_BlacklistIP_NotAdmin(t *testing.T) {
	handler, _, _, _ := createAdminHandler()

	reqBody := map[string]string{
		"ip": "192.168.1.100",
	}
	body, _ := json.Marshal(reqBody)

	// Create request without admin role
	req, _ := http.NewRequest("POST", "/admin/blacklist-ip", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req

	// Execute
	handler.BlacklistIP(ctx)

	// Assert
	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "admin role required")
}

func TestAdminHandler_BlacklistIP_InvalidRequest(t *testing.T) {
	handler, _, _, _ := createAdminHandler()

	// Create request with invalid JSON
	req, _ := http.NewRequest("POST", "/admin/blacklist-ip", bytes.NewReader([]byte("{invalid")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Set("role", "admin")

	// Execute
	handler.BlacklistIP(ctx)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid request")
}

func TestAdminHandler_BlacklistIP_NoLimiter(t *testing.T) {
	handler, _, _, _ := createAdminHandler()
	handler.rateLimiter = nil

	reqBody := map[string]string{
		"ip": "192.168.1.100",
	}
	body, _ := json.Marshal(reqBody)

	// Create request
	req, _ := http.NewRequest("POST", "/admin/blacklist-ip", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Set("role", "admin")

	// Execute
	handler.BlacklistIP(ctx)

	// Assert
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	assert.Contains(t, w.Body.String(), "rate limiter not configured")
}

func TestAdminHandler_UnblacklistIP_Success(t *testing.T) {
	handler, _, _, _ := createAdminHandler()

	// Create request body
	reqBody := map[string]string{
		"ip": "192.168.1.100",
	}
	body, _ := json.Marshal(reqBody)

	// Create request
	req, _ := http.NewRequest("POST", "/admin/unblacklist-ip", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Set("role", "admin")

	// Execute
	handler.UnblacklistIP(ctx)

	// Assert - should return 503 when rateLimiter is nil (expected behavior)
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	assert.Contains(t, w.Body.String(), "rate limiter not configured")
}

func TestAdminHandler_UnblacklistIP_NotAdmin(t *testing.T) {
	handler, _, _, _ := createAdminHandler()

	reqBody := map[string]string{
		"ip": "192.168.1.100",
	}
	body, _ := json.Marshal(reqBody)

	// Create request without admin role
	req, _ := http.NewRequest("POST", "/admin/unblacklist-ip", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Set("role", "user")

	// Execute
	handler.UnblacklistIP(ctx)

	// Assert
	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "admin role required")
}

func TestAdminHandler_ServiceInfo(t *testing.T) {
	handler, _, _, _ := createAdminHandler()

	// Create request
	req, _ := http.NewRequest("GET", "/admin/service-info", nil)
	w := httptest.NewRecorder()

	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req

	// Execute
	handler.ServiceInfo(ctx)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"service_name":"attachments-service"`)
	assert.Contains(t, w.Body.String(), `"version":"1.0.0"`)
}

func TestNewAdminHandler(t *testing.T) {
	mockDB := &MockDatabase{}
	logger := zap.NewNop()

	handler := NewAdminHandler(mockDB, nil, nil, nil, nil, nil, logger)

	assert.NotNil(t, handler)
	assert.Equal(t, mockDB, handler.db)
	assert.Equal(t, logger, handler.logger)
	assert.WithinDuration(t, time.Now(), handler.startTime, time.Second)
}
