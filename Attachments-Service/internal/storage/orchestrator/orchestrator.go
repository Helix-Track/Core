package orchestrator

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/helixtrack/attachments-service/internal/database"
	"github.com/helixtrack/attachments-service/internal/models"
	"github.com/helixtrack/attachments-service/internal/storage/adapters"
	"go.uber.org/zap"
)

// Orchestrator manages multiple storage endpoints with failover
type Orchestrator struct {
	db             database.Database
	logger         *zap.Logger
	endpoints      map[string]*EndpointWrapper
	primaryID      string
	backupIDs      []string
	mirrorIDs      []string
	config         *OrchestratorConfig
	mu             sync.RWMutex
	circuitBreaker *CircuitBreaker
}

// OrchestratorConfig contains orchestrator configuration
type OrchestratorConfig struct {
	// Failover settings
	EnableFailover            bool
	FailoverTimeout           time.Duration
	MaxRetries                int

	// Mirroring settings
	EnableMirroring           bool
	MirrorAsync               bool // Write to mirrors asynchronously
	RequireAllMirrorsSuccess  bool

	// Health check settings
	HealthCheckInterval       time.Duration
	HealthCheckTimeout        time.Duration
	UnhealthyThreshold        int // Consecutive failures before marking unhealthy
	HealthyThreshold          int // Consecutive successes before marking healthy

	// Circuit breaker settings
	CircuitBreakerThreshold   int
	CircuitBreakerTimeout     time.Duration
}

// DefaultOrchestratorConfig returns default orchestrator configuration
func DefaultOrchestratorConfig() *OrchestratorConfig {
	return &OrchestratorConfig{
		EnableFailover: true,
		FailoverTimeout: 30 * time.Second,
		MaxRetries: 3,

		EnableMirroring: true,
		MirrorAsync: true,
		RequireAllMirrorsSuccess: false,

		HealthCheckInterval: 1 * time.Minute,
		HealthCheckTimeout: 10 * time.Second,
		UnhealthyThreshold: 3,
		HealthyThreshold: 2,

		CircuitBreakerThreshold: 5,
		CircuitBreakerTimeout: 1 * time.Minute,
	}
}

// EndpointWrapper wraps a storage adapter with health tracking
type EndpointWrapper struct {
	ID                string
	Adapter           adapters.StorageAdapter
	Role              string // "primary", "backup", "mirror"
	Healthy           bool
	ConsecutiveFailures int
	ConsecutiveSuccesses int
	LastHealthCheck   time.Time
	CircuitBreaker    *CircuitBreaker
	mu                sync.RWMutex
}

// NewOrchestrator creates a new storage orchestrator
func NewOrchestrator(db database.Database, config *OrchestratorConfig, logger *zap.Logger) *Orchestrator {
	if config == nil {
		config = DefaultOrchestratorConfig()
	}

	o := &Orchestrator{
		db:             db,
		logger:         logger,
		endpoints:      make(map[string]*EndpointWrapper),
		backupIDs:      []string{},
		mirrorIDs:      []string{},
		config:         config,
		circuitBreaker: NewCircuitBreaker(config.CircuitBreakerThreshold, config.CircuitBreakerTimeout),
	}

	// Start health check goroutine
	go o.healthCheckLoop()

	return o
}

// RegisterEndpoint registers a storage endpoint
func (o *Orchestrator) RegisterEndpoint(id string, adapter adapters.StorageAdapter, role string) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	wrapper := &EndpointWrapper{
		ID:              id,
		Adapter:         adapter,
		Role:            role,
		Healthy:         true,
		CircuitBreaker:  NewCircuitBreaker(o.config.CircuitBreakerThreshold, o.config.CircuitBreakerTimeout),
		LastHealthCheck: time.Now(),
	}

	o.endpoints[id] = wrapper

	switch role {
	case "primary":
		o.primaryID = id
	case "backup":
		o.backupIDs = append(o.backupIDs, id)
	case "mirror":
		o.mirrorIDs = append(o.mirrorIDs, id)
	}

	o.logger.Info("storage endpoint registered",
		zap.String("id", id),
		zap.String("role", role),
		zap.String("type", adapter.GetType()),
	)

	return nil
}

// Store stores a file across configured endpoints
func (o *Orchestrator) Store(ctx context.Context, hash string, data io.Reader, size int64) (string, error) {
	// Get primary endpoint
	primary := o.getPrimaryEndpoint()
	if primary == nil {
		return "", fmt.Errorf("no primary storage endpoint configured")
	}

	// Read data into buffer (we may need to write to multiple endpoints)
	buf, err := io.ReadAll(data)
	if err != nil {
		return "", fmt.Errorf("failed to read data: %w", err)
	}

	// Store to primary with failover
	primaryPath, err := o.storeWithFailover(ctx, hash, buf, size)
	if err != nil {
		return "", fmt.Errorf("failed to store to primary and backups: %w", err)
	}

	// Store to mirrors (if enabled) - fire and forget for async
	if o.config.EnableMirroring && len(o.mirrorIDs) > 0 {
		result := &StoreResult{
			Hash: hash,
			Size: size,
			PrimaryPath: primaryPath,
			PrimaryEndpoint: o.primaryID,
			Endpoints: map[string]string{o.primaryID: primaryPath},
			Errors: make(map[string]error),
		}

		if o.config.MirrorAsync {
			// Asynchronous mirroring
			go o.storeToMirrors(context.Background(), hash, buf, size, result)
		} else {
			// Synchronous mirroring
			o.storeToMirrors(ctx, hash, buf, size, result)
		}
	}

	return primaryPath, nil
}

// storeWithFailover stores to primary with automatic failover to backup
func (o *Orchestrator) storeWithFailover(ctx context.Context, hash string, data []byte, size int64) (string, error) {
	// Try primary first
	primary := o.getPrimaryEndpoint()
	if primary != nil && primary.Healthy && primary.CircuitBreaker.CanExecute() {
		path, err := o.storeToEndpoint(ctx, primary, hash, data, size)
		if err == nil {
			primary.RecordSuccess()
			primary.CircuitBreaker.RecordSuccess()
			return path, nil
		}

		o.logger.Warn("primary storage failed",
			zap.String("endpoint_id", primary.ID),
			zap.Error(err),
		)

		primary.RecordFailure()
		primary.CircuitBreaker.RecordFailure()
	}

	// Failover to backup endpoints
	if o.config.EnableFailover {
		for _, backupID := range o.backupIDs {
			backup := o.getEndpoint(backupID)
			if backup == nil || !backup.Healthy || !backup.CircuitBreaker.CanExecute() {
				continue
			}

			path, err := o.storeToEndpoint(ctx, backup, hash, data, size)
			if err == nil {
				backup.RecordSuccess()
				backup.CircuitBreaker.RecordSuccess()

				o.logger.Info("failover successful",
					zap.String("from", o.primaryID),
					zap.String("to", backupID),
				)

				return path, nil
			}

			o.logger.Warn("backup storage failed",
				zap.String("endpoint_id", backupID),
				zap.Error(err),
			)

			backup.RecordFailure()
			backup.CircuitBreaker.RecordFailure()
		}
	}

	return "", fmt.Errorf("all storage endpoints failed")
}

// storeToEndpoint stores data to a specific endpoint
func (o *Orchestrator) storeToEndpoint(ctx context.Context, endpoint *EndpointWrapper, hash string, data []byte, size int64) (string, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, o.config.FailoverTimeout)
	defer cancel()

	done := make(chan struct{})
	var path string
	var err error

	go func() {
		reader := &byteReader{data: data, pos: 0}
		path, err = endpoint.Adapter.Store(ctx, hash, reader, size)
		close(done)
	}()

	select {
	case <-done:
		return path, err
	case <-timeoutCtx.Done():
		return "", fmt.Errorf("storage operation timed out")
	}
}

// storeToMirrors stores data to all mirror endpoints
func (o *Orchestrator) storeToMirrors(ctx context.Context, hash string, data []byte, size int64, result *StoreResult) {
	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, mirrorID := range o.mirrorIDs {
		mirror := o.getEndpoint(mirrorID)
		if mirror == nil || !mirror.Healthy {
			continue
		}

		wg.Add(1)
		go func(m *EndpointWrapper) {
			defer wg.Done()

			path, err := o.storeToEndpoint(ctx, m, hash, data, size)

			mu.Lock()
			defer mu.Unlock()

			if err != nil {
				result.Errors[m.ID] = err
				m.RecordFailure()

				o.logger.Warn("mirror storage failed",
					zap.String("endpoint_id", m.ID),
					zap.Error(err),
				)
			} else {
				result.Endpoints[m.ID] = path
				m.RecordSuccess()

				o.logger.Debug("mirror storage successful",
					zap.String("endpoint_id", m.ID),
				)
			}
		}(mirror)
	}

	wg.Wait()

	// Check if all mirrors succeeded (if required)
	if o.config.RequireAllMirrorsSuccess && len(result.Errors) > 0 {
		o.logger.Error("not all mirrors succeeded",
			zap.Int("failed", len(result.Errors)),
			zap.Int("total", len(o.mirrorIDs)),
		)
	}
}

// Retrieve retrieves a file from storage
func (o *Orchestrator) Retrieve(ctx context.Context, path string) (io.ReadCloser, error) {
	// Try primary first
	primary := o.getPrimaryEndpoint()
	if primary != nil && primary.Healthy {
		reader, err := primary.Adapter.Retrieve(ctx, path)
		if err == nil {
			return reader, nil
		}

		o.logger.Warn("primary retrieve failed",
			zap.String("endpoint_id", primary.ID),
			zap.Error(err),
		)
	}

	// Try backups
	for _, backupID := range o.backupIDs {
		backup := o.getEndpoint(backupID)
		if backup == nil || !backup.Healthy {
			continue
		}

		reader, err := backup.Adapter.Retrieve(ctx, path)
		if err == nil {
			o.logger.Info("retrieved from backup",
				zap.String("endpoint_id", backupID),
			)
			return reader, nil
		}
	}

	// Try mirrors as last resort
	for _, mirrorID := range o.mirrorIDs {
		mirror := o.getEndpoint(mirrorID)
		if mirror == nil || !mirror.Healthy {
			continue
		}

		reader, err := mirror.Adapter.Retrieve(ctx, path)
		if err == nil {
			o.logger.Info("retrieved from mirror",
				zap.String("endpoint_id", mirrorID),
			)
			return reader, nil
		}
	}

	return nil, fmt.Errorf("file not found in any storage endpoint")
}

// Delete deletes a file from all endpoints
func (o *Orchestrator) Delete(ctx context.Context, path string) error {
	var errs []error

	// Delete from all endpoints
	o.mu.RLock()
	endpoints := make([]*EndpointWrapper, 0, len(o.endpoints))
	for _, ep := range o.endpoints {
		endpoints = append(endpoints, ep)
	}
	o.mu.RUnlock()

	for _, ep := range endpoints {
		if err := ep.Adapter.Delete(ctx, path); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", ep.ID, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to delete from some endpoints: %v", errs)
	}

	return nil
}

// getPrimaryEndpoint returns the primary endpoint
func (o *Orchestrator) getPrimaryEndpoint() *EndpointWrapper {
	o.mu.RLock()
	defer o.mu.RUnlock()

	return o.endpoints[o.primaryID]
}

// getEndpoint returns an endpoint by ID
func (o *Orchestrator) getEndpoint(id string) *EndpointWrapper {
	o.mu.RLock()
	defer o.mu.RUnlock()

	return o.endpoints[id]
}

// healthCheckLoop performs periodic health checks
func (o *Orchestrator) healthCheckLoop() {
	ticker := time.NewTicker(o.config.HealthCheckInterval)
	defer ticker.Stop()

	for range ticker.C {
		o.performHealthChecks()
	}
}

// performHealthChecks checks health of all endpoints
func (o *Orchestrator) performHealthChecks() {
	o.mu.RLock()
	endpoints := make([]*EndpointWrapper, 0, len(o.endpoints))
	for _, ep := range o.endpoints {
		endpoints = append(endpoints, ep)
	}
	o.mu.RUnlock()

	for _, ep := range endpoints {
		go o.checkEndpointHealth(ep)
	}
}

// checkEndpointHealth checks health of a single endpoint
func (o *Orchestrator) checkEndpointHealth(ep *EndpointWrapper) {
	ctx, cancel := context.WithTimeout(context.Background(), o.config.HealthCheckTimeout)
	defer cancel()

	done := make(chan error)

	go func() {
		done <- ep.Adapter.Ping(ctx)
	}()

	var err error
	select {
	case err = <-done:
	case <-ctx.Done():
		err = fmt.Errorf("health check timeout")
	}

	ep.mu.Lock()
	ep.LastHealthCheck = time.Now()

	if err != nil {
		ep.ConsecutiveFailures++
		ep.ConsecutiveSuccesses = 0

		if ep.ConsecutiveFailures >= o.config.UnhealthyThreshold {
			if ep.Healthy {
				ep.Healthy = false
				o.logger.Warn("endpoint marked unhealthy",
					zap.String("endpoint_id", ep.ID),
					zap.Int("consecutive_failures", ep.ConsecutiveFailures),
				)

				// Record health status in database
				go o.recordEndpointHealth(ep.ID, false, err.Error())
			}
		}
	} else {
		ep.ConsecutiveSuccesses++
		ep.ConsecutiveFailures = 0

		if ep.ConsecutiveSuccesses >= o.config.HealthyThreshold {
			if !ep.Healthy {
				ep.Healthy = true
				o.logger.Info("endpoint marked healthy",
					zap.String("endpoint_id", ep.ID),
				)

				// Record health status in database
				go o.recordEndpointHealth(ep.ID, true, "")
			}
		}
	}

	ep.mu.Unlock()
}

// recordEndpointHealth records endpoint health status in database
func (o *Orchestrator) recordEndpointHealth(endpointID string, healthy bool, errorMessage string) {
	ctx := context.Background()

	status := models.HealthStatusHealthy
	if !healthy {
		status = models.HealthStatusUnhealthy
	}

	health := &models.StorageHealth{
		EndpointID: endpointID,
		CheckTime:  time.Now().Unix(),
		Status:     status,
	}

	if errorMessage != "" {
		health.ErrorMessage = &errorMessage
	}

	if err := o.db.RecordHealth(ctx, health); err != nil {
		o.logger.Error("failed to record health status",
			zap.String("endpoint_id", endpointID),
			zap.Error(err),
		)
	}
}

// RecordSuccess records a successful operation
func (ew *EndpointWrapper) RecordSuccess() {
	ew.mu.Lock()
	defer ew.mu.Unlock()

	ew.ConsecutiveSuccesses++
	ew.ConsecutiveFailures = 0
}

// RecordFailure records a failed operation
func (ew *EndpointWrapper) RecordFailure() {
	ew.mu.Lock()
	defer ew.mu.Unlock()

	ew.ConsecutiveFailures++
	ew.ConsecutiveSuccesses = 0
}

// StoreResult contains the result of a store operation
type StoreResult struct {
	Hash             string
	Size             int64
	PrimaryEndpoint  string
	PrimaryPath      string
	Endpoints        map[string]string // endpoint ID -> path
	Errors           map[string]error  // endpoint ID -> error
}

// byteReader implements io.Reader for byte slices
type byteReader struct {
	data []byte
	pos  int
}

func (br *byteReader) Read(p []byte) (n int, err error) {
	if br.pos >= len(br.data) {
		return 0, io.EOF
	}

	n = copy(p, br.data[br.pos:])
	br.pos += n
	return n, nil
}

// StartHealthMonitor starts the health monitoring loop in a goroutine
func (o *Orchestrator) StartHealthMonitor(ctx context.Context, interval time.Duration) {
	// Update config with provided interval if specified
	if interval > 0 {
		o.config.HealthCheckInterval = interval
	}

	o.logger.Info("starting storage health monitor",
		zap.Duration("interval", o.config.HealthCheckInterval),
	)

	go o.healthCheckLoop()
}

// EndpointHealth contains health information for an endpoint
type EndpointHealth struct {
	EndpointID string
	Role       string
	Status     string
	LatencyMs  int64
	LastCheck  time.Time
}

// GetEndpointHealth returns health status for all endpoints
func (o *Orchestrator) GetEndpointHealth() []EndpointHealth {
	o.mu.RLock()
	defer o.mu.RUnlock()

	results := make([]EndpointHealth, 0, len(o.endpoints))
	for _, ep := range o.endpoints {
		ep.mu.RLock()
		status := "unhealthy"
		if ep.Healthy {
			status = "healthy"
		}

		results = append(results, EndpointHealth{
			EndpointID: ep.ID,
			Role:       ep.Role,
			Status:     status,
			LatencyMs:  0, // TODO: track actual latency
			LastCheck:  ep.LastHealthCheck,
		})
		ep.mu.RUnlock()
	}

	return results
}

// StorageAdapter interface implementation - delegate to primary endpoint

// Exists checks if a file exists (delegates to primary)
func (o *Orchestrator) Exists(ctx context.Context, path string) (bool, error) {
	primary := o.getPrimaryEndpoint()
	if primary == nil {
		return false, fmt.Errorf("no primary endpoint available")
	}

	return primary.Adapter.Exists(ctx, path)
}

// GetSize returns the size of a file (delegates to primary)
func (o *Orchestrator) GetSize(ctx context.Context, path string) (int64, error) {
	primary := o.getPrimaryEndpoint()
	if primary == nil {
		return 0, fmt.Errorf("no primary endpoint available")
	}

	return primary.Adapter.GetSize(ctx, path)
}

// GetMetadata returns metadata about a file (delegates to primary)
func (o *Orchestrator) GetMetadata(ctx context.Context, path string) (*adapters.FileMetadata, error) {
	primary := o.getPrimaryEndpoint()
	if primary == nil {
		return nil, fmt.Errorf("no primary endpoint available")
	}

	return primary.Adapter.GetMetadata(ctx, path)
}

// Ping checks if storage is accessible (checks primary)
func (o *Orchestrator) Ping(ctx context.Context) error {
	primary := o.getPrimaryEndpoint()
	if primary == nil {
		return fmt.Errorf("no primary endpoint available")
	}

	return primary.Adapter.Ping(ctx)
}

// GetCapacity returns storage capacity information (from primary)
func (o *Orchestrator) GetCapacity(ctx context.Context) (*adapters.CapacityInfo, error) {
	primary := o.getPrimaryEndpoint()
	if primary == nil {
		return nil, fmt.Errorf("no primary endpoint available")
	}

	return primary.Adapter.GetCapacity(ctx)
}

// GetType returns the adapter type
func (o *Orchestrator) GetType() string {
	return "orchestrator"
}
