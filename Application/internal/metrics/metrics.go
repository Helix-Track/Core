package metrics

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
)

// Metrics collects application performance metrics
type Metrics struct {
	// Request metrics
	totalRequests       int64
	successfulRequests  int64
	failedRequests      int64
	totalRequestTime    int64 // nanoseconds
	minRequestTime      int64 // nanoseconds
	maxRequestTime      int64 // nanoseconds

	// Status code counts
	status2xx           int64
	status3xx           int64
	status4xx           int64
	status5xx           int64

	// Endpoint metrics
	endpointMetrics     map[string]*EndpointMetrics
	endpointMu          sync.RWMutex

	// System metrics
	startTime           time.Time
	lastRequestTime     time.Time
	lastRequestMu       sync.RWMutex
}

// EndpointMetrics contains metrics for a specific endpoint
type EndpointMetrics struct {
	Path              string
	Method            string
	Count             int64
	TotalTime         int64 // nanoseconds
	MinTime           int64 // nanoseconds
	MaxTime           int64 // nanoseconds
	Status2xx         int64
	Status3xx         int64
	Status4xx         int64
	Status5xx         int64
}

// MetricsSummary contains summarized metrics
type MetricsSummary struct {
	TotalRequests      int64         `json:"total_requests"`
	SuccessfulRequests int64         `json:"successful_requests"`
	FailedRequests     int64         `json:"failed_requests"`
	AvgRequestTime     time.Duration `json:"avg_request_time"`
	MinRequestTime     time.Duration `json:"min_request_time"`
	MaxRequestTime     time.Duration `json:"max_request_time"`
	RequestsPerSecond  float64       `json:"requests_per_second"`
	Status2xx          int64         `json:"status_2xx"`
	Status3xx          int64         `json:"status_3xx"`
	Status4xx          int64         `json:"status_4xx"`
	Status5xx          int64         `json:"status_5xx"`
	Uptime             time.Duration `json:"uptime"`
	LastRequestTime    time.Time     `json:"last_request_time"`
	Endpoints          []*EndpointMetrics `json:"endpoints,omitempty"`
}

// NewMetrics creates a new metrics collector
func NewMetrics() *Metrics {
	return &Metrics{
		endpointMetrics: make(map[string]*EndpointMetrics),
		startTime:       time.Now(),
		minRequestTime:  int64(^uint64(0) >> 1), // Max int64
	}
}

// RecordRequest records a request
func (m *Metrics) RecordRequest(duration time.Duration, statusCode int, path, method string) {
	durationNs := duration.Nanoseconds()

	// Update total counts
	atomic.AddInt64(&m.totalRequests, 1)

	if statusCode >= 200 && statusCode < 300 {
		atomic.AddInt64(&m.successfulRequests, 1)
		atomic.AddInt64(&m.status2xx, 1)
	} else if statusCode >= 300 && statusCode < 400 {
		atomic.AddInt64(&m.status3xx, 1)
	} else if statusCode >= 400 && statusCode < 500 {
		atomic.AddInt64(&m.failedRequests, 1)
		atomic.AddInt64(&m.status4xx, 1)
	} else if statusCode >= 500 {
		atomic.AddInt64(&m.failedRequests, 1)
		atomic.AddInt64(&m.status5xx, 1)
	}

	// Update timing
	atomic.AddInt64(&m.totalRequestTime, durationNs)

	// Update min time
	for {
		oldMin := atomic.LoadInt64(&m.minRequestTime)
		if durationNs >= oldMin {
			break
		}
		if atomic.CompareAndSwapInt64(&m.minRequestTime, oldMin, durationNs) {
			break
		}
	}

	// Update max time
	for {
		oldMax := atomic.LoadInt64(&m.maxRequestTime)
		if durationNs <= oldMax {
			break
		}
		if atomic.CompareAndSwapInt64(&m.maxRequestTime, oldMax, durationNs) {
			break
		}
	}

	// Update last request time
	m.lastRequestMu.Lock()
	m.lastRequestTime = time.Now()
	m.lastRequestMu.Unlock()

	// Update endpoint metrics
	m.recordEndpointMetrics(path, method, durationNs, statusCode)
}

// recordEndpointMetrics records metrics for a specific endpoint
func (m *Metrics) recordEndpointMetrics(path, method string, durationNs int64, statusCode int) {
	key := method + ":" + path

	m.endpointMu.RLock()
	endpoint, exists := m.endpointMetrics[key]
	m.endpointMu.RUnlock()

	if !exists {
		m.endpointMu.Lock()
		// Double-check after acquiring write lock
		endpoint, exists = m.endpointMetrics[key]
		if !exists {
			endpoint = &EndpointMetrics{
				Path:    path,
				Method:  method,
				MinTime: int64(^uint64(0) >> 1), // Max int64
			}
			m.endpointMetrics[key] = endpoint
		}
		m.endpointMu.Unlock()
	}

	// Update endpoint metrics
	atomic.AddInt64(&endpoint.Count, 1)
	atomic.AddInt64(&endpoint.TotalTime, durationNs)

	// Update endpoint min time
	for {
		oldMin := atomic.LoadInt64(&endpoint.MinTime)
		if durationNs >= oldMin {
			break
		}
		if atomic.CompareAndSwapInt64(&endpoint.MinTime, oldMin, durationNs) {
			break
		}
	}

	// Update endpoint max time
	for {
		oldMax := atomic.LoadInt64(&endpoint.MaxTime)
		if durationNs <= oldMax {
			break
		}
		if atomic.CompareAndSwapInt64(&endpoint.MaxTime, oldMax, durationNs) {
			break
		}
	}

	// Update endpoint status codes
	if statusCode >= 200 && statusCode < 300 {
		atomic.AddInt64(&endpoint.Status2xx, 1)
	} else if statusCode >= 300 && statusCode < 400 {
		atomic.AddInt64(&endpoint.Status3xx, 1)
	} else if statusCode >= 400 && statusCode < 500 {
		atomic.AddInt64(&endpoint.Status4xx, 1)
	} else if statusCode >= 500 {
		atomic.AddInt64(&endpoint.Status5xx, 1)
	}
}

// GetSummary returns a summary of metrics
func (m *Metrics) GetSummary(includeEndpoints bool) *MetricsSummary {
	totalRequests := atomic.LoadInt64(&m.totalRequests)
	successfulRequests := atomic.LoadInt64(&m.successfulRequests)
	failedRequests := atomic.LoadInt64(&m.failedRequests)
	totalRequestTime := atomic.LoadInt64(&m.totalRequestTime)
	minRequestTime := atomic.LoadInt64(&m.minRequestTime)
	maxRequestTime := atomic.LoadInt64(&m.maxRequestTime)
	status2xx := atomic.LoadInt64(&m.status2xx)
	status3xx := atomic.LoadInt64(&m.status3xx)
	status4xx := atomic.LoadInt64(&m.status4xx)
	status5xx := atomic.LoadInt64(&m.status5xx)

	var avgRequestTime time.Duration
	if totalRequests > 0 {
		avgRequestTime = time.Duration(totalRequestTime / totalRequests)
	}

	uptime := time.Since(m.startTime)
	var requestsPerSecond float64
	if uptime.Seconds() > 0 {
		requestsPerSecond = float64(totalRequests) / uptime.Seconds()
	}

	m.lastRequestMu.RLock()
	lastRequestTime := m.lastRequestTime
	m.lastRequestMu.RUnlock()

	summary := &MetricsSummary{
		TotalRequests:      totalRequests,
		SuccessfulRequests: successfulRequests,
		FailedRequests:     failedRequests,
		AvgRequestTime:     avgRequestTime,
		MinRequestTime:     time.Duration(minRequestTime),
		MaxRequestTime:     time.Duration(maxRequestTime),
		RequestsPerSecond:  requestsPerSecond,
		Status2xx:          status2xx,
		Status3xx:          status3xx,
		Status4xx:          status4xx,
		Status5xx:          status5xx,
		Uptime:             uptime,
		LastRequestTime:    lastRequestTime,
	}

	if includeEndpoints {
		m.endpointMu.RLock()
		summary.Endpoints = make([]*EndpointMetrics, 0, len(m.endpointMetrics))
		for _, endpoint := range m.endpointMetrics {
			// Create a copy
			endpointCopy := &EndpointMetrics{
				Path:      endpoint.Path,
				Method:    endpoint.Method,
				Count:     atomic.LoadInt64(&endpoint.Count),
				TotalTime: atomic.LoadInt64(&endpoint.TotalTime),
				MinTime:   atomic.LoadInt64(&endpoint.MinTime),
				MaxTime:   atomic.LoadInt64(&endpoint.MaxTime),
				Status2xx: atomic.LoadInt64(&endpoint.Status2xx),
				Status3xx: atomic.LoadInt64(&endpoint.Status3xx),
				Status4xx: atomic.LoadInt64(&endpoint.Status4xx),
				Status5xx: atomic.LoadInt64(&endpoint.Status5xx),
			}
			summary.Endpoints = append(summary.Endpoints, endpointCopy)
		}
		m.endpointMu.RUnlock()
	}

	return summary
}

// Reset resets all metrics
func (m *Metrics) Reset() {
	atomic.StoreInt64(&m.totalRequests, 0)
	atomic.StoreInt64(&m.successfulRequests, 0)
	atomic.StoreInt64(&m.failedRequests, 0)
	atomic.StoreInt64(&m.totalRequestTime, 0)
	atomic.StoreInt64(&m.minRequestTime, int64(^uint64(0)>>1))
	atomic.StoreInt64(&m.maxRequestTime, 0)
	atomic.StoreInt64(&m.status2xx, 0)
	atomic.StoreInt64(&m.status3xx, 0)
	atomic.StoreInt64(&m.status4xx, 0)
	atomic.StoreInt64(&m.status5xx, 0)

	m.endpointMu.Lock()
	m.endpointMetrics = make(map[string]*EndpointMetrics)
	m.endpointMu.Unlock()

	m.startTime = time.Now()

	m.lastRequestMu.Lock()
	m.lastRequestTime = time.Time{}
	m.lastRequestMu.Unlock()
}

// MetricsMiddleware creates middleware for metrics collection
func MetricsMiddleware(metrics *Metrics) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Process request
		c.Next()

		// Record metrics
		duration := time.Since(start)
		metrics.RecordRequest(duration, c.Writer.Status(), c.Request.URL.Path, c.Request.Method)
	}
}

// Global metrics instance
var globalMetrics = NewMetrics()

// GetGlobalMetrics returns the global metrics instance
func GetGlobalMetrics() *Metrics {
	return globalMetrics
}

// HealthStatus represents application health status
type HealthStatus struct {
	Status      string        `json:"status"`
	Uptime      time.Duration `json:"uptime"`
	Version     string        `json:"version"`
	Database    string        `json:"database"`
	Cache       string        `json:"cache,omitempty"`
	Metrics     *MetricsSummary `json:"metrics,omitempty"`
	Timestamp   time.Time     `json:"timestamp"`
}

// HealthCheck performs a health check
type HealthCheck struct {
	Version string
	DBPing  func() error
}

// Check performs health check
func (hc *HealthCheck) Check(includeMetrics bool) *HealthStatus {
	status := "healthy"
	dbStatus := "connected"

	// Check database
	if hc.DBPing != nil {
		if err := hc.DBPing(); err != nil {
			status = "unhealthy"
			dbStatus = "disconnected"
		}
	}

	health := &HealthStatus{
		Status:    status,
		Uptime:    time.Since(globalMetrics.startTime),
		Version:   hc.Version,
		Database:  dbStatus,
		Timestamp: time.Now(),
	}

	if includeMetrics {
		health.Metrics = globalMetrics.GetSummary(true)
	}

	return health
}
