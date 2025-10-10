package services

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"
	"helixtrack.ru/core/internal/database"
	"helixtrack.ru/core/internal/logger"
	"helixtrack.ru/core/internal/models"
)

// HealthChecker performs periodic health checks on registered services
type HealthChecker struct {
	db               database.Database
	httpClient       *http.Client
	checkInterval    time.Duration
	checkTimeout     time.Duration
	stopChan         chan struct{}
	wg               sync.WaitGroup
	mu               sync.RWMutex
	running          bool
	failureThreshold int // Number of failures before marking unhealthy
	failoverManager  *FailoverManager
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(db database.Database, checkInterval, checkTimeout time.Duration) *HealthChecker {
	return &HealthChecker{
		db:            db,
		httpClient: &http.Client{
			Timeout: checkTimeout,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse // Don't follow redirects
			},
		},
		checkInterval:    checkInterval,
		checkTimeout:     checkTimeout,
		stopChan:         make(chan struct{}),
		failureThreshold: 3, // Mark unhealthy after 3 consecutive failures
		failoverManager:  NewFailoverManager(db),
	}
}

// Start begins the health check loop
func (h *HealthChecker) Start() error {
	h.mu.Lock()
	if h.running {
		h.mu.Unlock()
		return fmt.Errorf("health checker is already running")
	}
	h.running = true
	h.mu.Unlock()

	logger.Info("Starting service health checker", zap.Duration("interval", h.checkInterval))

	h.wg.Add(1)
	go h.checkLoop()

	return nil
}

// Stop stops the health check loop
func (h *HealthChecker) Stop() {
	h.mu.Lock()
	if !h.running {
		h.mu.Unlock()
		return
	}
	h.running = false
	h.mu.Unlock()

	close(h.stopChan)
	h.wg.Wait()

	logger.Info("Service health checker stopped")
}

// IsRunning returns whether the health checker is currently running
func (h *HealthChecker) IsRunning() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.running
}

// checkLoop is the main health check loop
func (h *HealthChecker) checkLoop() {
	defer h.wg.Done()

	ticker := time.NewTicker(h.checkInterval)
	defer ticker.Stop()

	// Perform initial check immediately
	h.checkAllServices()

	for {
		select {
		case <-ticker.C:
			h.checkAllServices()
		case <-h.stopChan:
			return
		}
	}
}

// checkAllServices checks health of all registered services
func (h *HealthChecker) checkAllServices() {
	ctx := context.Background()

	// Get all non-deleted, non-decommissioned services
	query := `
		SELECT id, name, type, url, health_check_url, status, failed_health_count
		FROM service_registry
		WHERE deleted = 0 AND status != ?
	`

	rows, err := h.db.Query(ctx, query, models.ServiceStatusDecommission)
	if err != nil {
		logger.Error("Failed to query services for health check", zap.Error(err))
		return
	}
	defer rows.Close()

	var services []struct {
		ID               string
		Name             string
		Type             string
		URL              string
		HealthCheckURL   string
		Status           string
		FailedHealthCount int
	}

	for rows.Next() {
		var svc struct {
			ID               string
			Name             string
			Type             string
			URL              string
			HealthCheckURL   string
			Status           string
			FailedHealthCount int
		}

		err := rows.Scan(&svc.ID, &svc.Name, &svc.Type, &svc.URL, &svc.HealthCheckURL,
			&svc.Status, &svc.FailedHealthCount)
		if err != nil {
			logger.Error("Failed to scan service row", zap.Error(err))
			continue
		}

		services = append(services, svc)
	}

	// Check each service in parallel
	var wg sync.WaitGroup
	for _, svc := range services {
		wg.Add(1)
		go func(s struct {
			ID               string
			Name             string
			Type             string
			URL              string
			HealthCheckURL   string
			Status           string
			FailedHealthCount int
		}) {
			defer wg.Done()
			h.checkService(s.ID, s.Name, s.HealthCheckURL, s.FailedHealthCount)
		}(svc)
	}

	wg.Wait()
}

// checkService performs a health check on a single service
func (h *HealthChecker) checkService(serviceID, serviceName, healthCheckURL string, currentFailures int) {
	ctx := context.Background()
	startTime := time.Now()

	logger.Debug("Checking service health",
		zap.String("service_id", serviceID),
		zap.String("service_name", serviceName),
		zap.String("url", healthCheckURL),
	)

	// Create context with timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, h.checkTimeout)
	defer cancel()

	// Perform HTTP health check
	req, err := http.NewRequestWithContext(timeoutCtx, http.MethodGet, healthCheckURL, nil)
	if err != nil {
		h.recordHealthCheck(serviceID, false, 0, 0, fmt.Sprintf("Failed to create request: %v", err), currentFailures+1)
		return
	}

	resp, err := h.httpClient.Do(req)
	responseTime := time.Since(startTime).Milliseconds()

	if err != nil {
		h.recordHealthCheck(serviceID, false, responseTime, 0, fmt.Sprintf("Request failed: %v", err), currentFailures+1)
		return
	}
	defer resp.Body.Close()

	// Consider 2xx and 3xx status codes as healthy
	healthy := resp.StatusCode >= 200 && resp.StatusCode < 400

	var failureCount int
	if healthy {
		failureCount = 0 // Reset failure count on success
	} else {
		failureCount = currentFailures + 1
	}

	errorMsg := ""
	if !healthy {
		errorMsg = fmt.Sprintf("Unhealthy status code: %d", resp.StatusCode)
	}

	h.recordHealthCheck(serviceID, healthy, responseTime, resp.StatusCode, errorMsg, failureCount)
}

// recordHealthCheck records the result of a health check
func (h *HealthChecker) recordHealthCheck(
	serviceID string,
	healthy bool,
	responseTime int64,
	statusCode int,
	errorMessage string,
	failureCount int,
) {
	ctx := context.Background()
	now := time.Now()

	// Determine new status
	var newStatus models.ServiceStatus
	if healthy {
		newStatus = models.ServiceStatusHealthy
	} else {
		if failureCount >= h.failureThreshold {
			newStatus = models.ServiceStatusUnhealthy
		} else {
			// Keep current status if below threshold
			var currentStatus string
			err := h.db.QueryRow(ctx, "SELECT status FROM service_registry WHERE id = ?", serviceID).Scan(&currentStatus)
			if err == nil {
				newStatus = models.ServiceStatus(currentStatus)
			} else {
				newStatus = models.ServiceStatusHealthy // Default
			}
		}
	}

	// Update service registry
	updateQuery := `
		UPDATE service_registry
		SET status = ?,
		    last_health_check = ?,
		    health_check_count = health_check_count + 1,
		    failed_health_count = ?
		WHERE id = ?
	`

	_, err := h.db.Exec(ctx, updateQuery, newStatus, now.Unix(), failureCount, serviceID)
	if err != nil {
		logger.Error("Failed to update service health status",
			zap.String("service_id", serviceID),
			zap.Error(err),
		)
		return
	}

	// Check if failover is needed (automatic failover/failback)
	if err := h.failoverManager.CheckFailoverNeeded(serviceID, healthy, newStatus); err != nil {
		logger.Error("Failover check failed",
			zap.String("service_id", serviceID),
			zap.Error(err),
		)
		// Don't return - continue with health check recording even if failover fails
	}

	// Insert health check record
	insertQuery := `
		INSERT INTO service_health_check (id, service_id, timestamp, status, response_time, status_code, error_message, checked_by)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	checkID := fmt.Sprintf("hc_%s_%d", serviceID[:8], now.Unix())
	_, err = h.db.Exec(ctx, insertQuery,
		checkID,
		serviceID,
		now.Unix(),
		newStatus,
		responseTime,
		statusCode,
		errorMessage,
		"system",
	)

	if err != nil {
		logger.Error("Failed to insert health check record",
			zap.String("service_id", serviceID),
			zap.Error(err),
		)
		return
	}

	// Log the result
	if healthy {
		logger.Debug("Service health check passed",
			zap.String("service_id", serviceID),
			zap.Int64("response_time_ms", responseTime),
			zap.Int("status_code", statusCode),
		)
	} else {
		logger.Warn("Service health check failed",
			zap.String("service_id", serviceID),
			zap.Int("failure_count", failureCount),
			zap.String("status", string(newStatus)),
			zap.String("error", errorMessage),
		)
	}
}

// CheckServiceNow performs an immediate health check on a specific service
func (h *HealthChecker) CheckServiceNow(serviceID string) error {
	ctx := context.Background()

	// Get service details
	query := `
		SELECT name, health_check_url, failed_health_count
		FROM service_registry
		WHERE id = ? AND deleted = 0
	`

	var name, healthCheckURL string
	var failedHealthCount int

	err := h.db.QueryRow(ctx, query, serviceID).Scan(&name, &healthCheckURL, &failedHealthCount)
	if err != nil {
		return fmt.Errorf("service not found: %w", err)
	}

	// Perform check
	h.checkService(serviceID, name, healthCheckURL, failedHealthCount)

	return nil
}

// GetServiceHealthHistory returns recent health check history for a service
func (h *HealthChecker) GetServiceHealthHistory(serviceID string, limit int) ([]models.ServiceHealthCheck, error) {
	ctx := context.Background()

	query := `
		SELECT id, service_id, timestamp, status, response_time, status_code, error_message, checked_by
		FROM service_health_check
		WHERE service_id = ?
		ORDER BY timestamp DESC
		LIMIT ?
	`

	rows, err := h.db.Query(ctx, query, serviceID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query health check history: %w", err)
	}
	defer rows.Close()

	var checks []models.ServiceHealthCheck

	for rows.Next() {
		var check models.ServiceHealthCheck
		var timestamp int64

		err := rows.Scan(
			&check.ID,
			&check.ServiceID,
			&timestamp,
			&check.Status,
			&check.ResponseTime,
			&check.StatusCode,
			&check.ErrorMessage,
			&check.CheckedBy,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan health check row: %w", err)
		}

		check.Timestamp = time.Unix(timestamp, 0)
		checks = append(checks, check)
	}

	return checks, nil
}
