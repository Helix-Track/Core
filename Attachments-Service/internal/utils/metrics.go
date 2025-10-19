package utils

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metrics holds all Prometheus metrics for the Attachments Service
type PrometheusMetrics struct {
	// Upload metrics
	UploadTotal    *prometheus.CounterVec
	UploadDuration *prometheus.HistogramVec
	UploadSize     *prometheus.HistogramVec

	// Download metrics
	DownloadTotal    *prometheus.CounterVec
	DownloadDuration *prometheus.HistogramVec
	DownloadSize     *prometheus.HistogramVec

	// Delete metrics
	DeleteTotal *prometheus.CounterVec

	// Storage metrics
	StorageBytesTotal    prometheus.Gauge
	StorageFilesTotal    prometheus.Gauge
	StorageReferencesTotal prometheus.Gauge
	DeduplicationRate    prometheus.Gauge
	OrphanedFilesTotal   prometheus.Gauge

	// Deduplication metrics
	DeduplicationTotal *prometheus.CounterVec
	DeduplicationSavedBytes prometheus.Counter

	// Security metrics
	VirusScanTotal    *prometheus.CounterVec
	VirusDetectedTotal prometheus.Counter
	RateLimitHitsTotal *prometheus.CounterVec

	// Error metrics
	ErrorsTotal *prometheus.CounterVec

	// Request metrics
	RequestTotal    *prometheus.CounterVec
	RequestDuration *prometheus.HistogramVec

	// Storage endpoint metrics
	StorageEndpointHealth *prometheus.GaugeVec
	StorageEndpointLatency *prometheus.GaugeVec

	// Quota metrics
	QuotaUsageBytes *prometheus.GaugeVec
	QuotaUsageFiles *prometheus.GaugeVec
	QuotaExceeded   *prometheus.CounterVec
}

// NewPrometheusMetrics creates and registers all Prometheus metrics
func NewPrometheusMetrics() *PrometheusMetrics {
	return &PrometheusMetrics{
		// Upload metrics
		UploadTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "attachments_uploads_total",
				Help: "Total number of file uploads",
			},
			[]string{"status", "mime_type"},
		),
		UploadDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "attachments_upload_duration_seconds",
				Help:    "Upload duration in seconds",
				Buckets: prometheus.ExponentialBuckets(0.001, 2, 15), // 1ms to ~32s
			},
			[]string{"mime_type"},
		),
		UploadSize: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "attachments_upload_size_bytes",
				Help:    "Upload size in bytes",
				Buckets: prometheus.ExponentialBuckets(1024, 2, 20), // 1KB to ~1GB
			},
			[]string{"mime_type"},
		),

		// Download metrics
		DownloadTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "attachments_downloads_total",
				Help: "Total number of file downloads",
			},
			[]string{"status"},
		),
		DownloadDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "attachments_download_duration_seconds",
				Help:    "Download duration in seconds",
				Buckets: prometheus.ExponentialBuckets(0.001, 2, 15),
			},
			[]string{"cache_hit"},
		),
		DownloadSize: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "attachments_download_size_bytes",
				Help:    "Download size in bytes",
				Buckets: prometheus.ExponentialBuckets(1024, 2, 20),
			},
			[]string{"cache_hit"},
		),

		// Delete metrics
		DeleteTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "attachments_deletes_total",
				Help: "Total number of file deletes",
			},
			[]string{"status"},
		),

		// Storage metrics
		StorageBytesTotal: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "attachments_storage_bytes_total",
				Help: "Total storage used in bytes",
			},
		),
		StorageFilesTotal: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "attachments_storage_files_total",
				Help: "Total number of unique files stored",
			},
		),
		StorageReferencesTotal: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "attachments_storage_references_total",
				Help: "Total number of file references",
			},
		),
		DeduplicationRate: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "attachments_deduplication_rate",
				Help: "Deduplication rate as percentage",
			},
		),
		OrphanedFilesTotal: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "attachments_orphaned_files_total",
				Help: "Total number of orphaned files",
			},
		),

		// Deduplication metrics
		DeduplicationTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "attachments_deduplication_total",
				Help: "Total number of deduplicated uploads",
			},
			[]string{"result"}, // "duplicate" or "unique"
		),
		DeduplicationSavedBytes: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "attachments_deduplication_saved_bytes_total",
				Help: "Total bytes saved through deduplication",
			},
		),

		// Security metrics
		VirusScanTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "attachments_virus_scans_total",
				Help: "Total number of virus scans",
			},
			[]string{"result"}, // "clean", "infected", "failed", "skipped"
		),
		VirusDetectedTotal: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "attachments_virus_detected_total",
				Help: "Total number of viruses detected",
			},
		),
		RateLimitHitsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "attachments_rate_limit_hits_total",
				Help: "Total number of rate limit hits",
			},
			[]string{"type"}, // "ip", "user", "global"
		),

		// Error metrics
		ErrorsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "attachments_errors_total",
				Help: "Total number of errors",
			},
			[]string{"type", "operation"},
		),

		// Request metrics
		RequestTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "attachments_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "path", "status"},
		),
		RequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "attachments_request_duration_seconds",
				Help:    "HTTP request duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "path"},
		),

		// Storage endpoint metrics
		StorageEndpointHealth: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "attachments_storage_endpoint_health",
				Help: "Storage endpoint health status (1=healthy, 0=unhealthy)",
			},
			[]string{"endpoint_id", "endpoint_type", "role"},
		),
		StorageEndpointLatency: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "attachments_storage_endpoint_latency_ms",
				Help: "Storage endpoint latency in milliseconds",
			},
			[]string{"endpoint_id", "endpoint_type", "role"},
		),

		// Quota metrics
		QuotaUsageBytes: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "attachments_quota_usage_bytes",
				Help: "User quota usage in bytes",
			},
			[]string{"user_id"},
		),
		QuotaUsageFiles: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "attachments_quota_usage_files",
				Help: "User quota usage in number of files",
			},
			[]string{"user_id"},
		),
		QuotaExceeded: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "attachments_quota_exceeded_total",
				Help: "Total number of quota exceeded events",
			},
			[]string{"user_id", "type"}, // type: "bytes" or "files"
		),
	}
}

// RecordUpload records an upload operation
func (m *PrometheusMetrics) RecordUpload(status, mimeType string, size int64, duration time.Duration) {
	m.UploadTotal.WithLabelValues(status, mimeType).Inc()
	m.UploadDuration.WithLabelValues(mimeType).Observe(duration.Seconds())
	m.UploadSize.WithLabelValues(mimeType).Observe(float64(size))
}

// RecordDownload records a download operation
func (m *PrometheusMetrics) RecordDownload(status string, size int64, duration time.Duration, cacheHit bool) {
	m.DownloadTotal.WithLabelValues(status).Inc()

	cacheHitStr := "false"
	if cacheHit {
		cacheHitStr = "true"
	}

	m.DownloadDuration.WithLabelValues(cacheHitStr).Observe(duration.Seconds())
	m.DownloadSize.WithLabelValues(cacheHitStr).Observe(float64(size))
}

// RecordDelete records a delete operation
func (m *PrometheusMetrics) RecordDelete(status string) {
	m.DeleteTotal.WithLabelValues(status).Inc()
}

// RecordDeduplication records a deduplication event
func (m *PrometheusMetrics) RecordDeduplication(isDuplicate bool, savedBytes int64) {
	result := "unique"
	if isDuplicate {
		result = "duplicate"
		m.DeduplicationSavedBytes.Add(float64(savedBytes))
	}
	m.DeduplicationTotal.WithLabelValues(result).Inc()
}

// RecordVirusScan records a virus scan result
func (m *PrometheusMetrics) RecordVirusScan(result string) {
	m.VirusScanTotal.WithLabelValues(result).Inc()
	if result == "infected" {
		m.VirusDetectedTotal.Inc()
	}
}

// RecordRateLimitHit records a rate limit hit
func (m *PrometheusMetrics) RecordRateLimitHit(limitType string) {
	m.RateLimitHitsTotal.WithLabelValues(limitType).Inc()
}

// RecordError records an error
func (m *PrometheusMetrics) RecordError(errorType, operation string) {
	m.ErrorsTotal.WithLabelValues(errorType, operation).Inc()
}

// RecordRequest records an HTTP request
func (m *PrometheusMetrics) RecordRequest(method, path, status string, duration time.Duration) {
	m.RequestTotal.WithLabelValues(method, path, status).Inc()
	m.RequestDuration.WithLabelValues(method, path).Observe(duration.Seconds())
}

// UpdateStorageStats updates storage statistics
func (m *PrometheusMetrics) UpdateStorageStats(totalBytes, totalFiles, totalReferences, orphaned int64, deduplicationRate float64) {
	m.StorageBytesTotal.Set(float64(totalBytes))
	m.StorageFilesTotal.Set(float64(totalFiles))
	m.StorageReferencesTotal.Set(float64(totalReferences))
	m.OrphanedFilesTotal.Set(float64(orphaned))
	m.DeduplicationRate.Set(deduplicationRate)
}

// UpdateEndpointHealth updates storage endpoint health
func (m *PrometheusMetrics) UpdateEndpointHealth(endpointID, endpointType, role string, healthy bool, latencyMs int) {
	healthValue := 0.0
	if healthy {
		healthValue = 1.0
	}
	m.StorageEndpointHealth.WithLabelValues(endpointID, endpointType, role).Set(healthValue)
	m.StorageEndpointLatency.WithLabelValues(endpointID, endpointType, role).Set(float64(latencyMs))
}

// UpdateQuotaUsage updates user quota usage
func (m *PrometheusMetrics) UpdateQuotaUsage(userID string, usedBytes int64, usedFiles int) {
	m.QuotaUsageBytes.WithLabelValues(userID).Set(float64(usedBytes))
	m.QuotaUsageFiles.WithLabelValues(userID).Set(float64(usedFiles))
}

// RecordQuotaExceeded records a quota exceeded event
func (m *PrometheusMetrics) RecordQuotaExceeded(userID, quotaType string) {
	m.QuotaExceeded.WithLabelValues(userID, quotaType).Inc()
}

// PrometheusHandler returns an HTTP handler for Prometheus metrics
func PrometheusHandler() http.Handler {
	return promhttp.Handler()
}

// MetricsMiddleware creates middleware for recording HTTP metrics
func MetricsMiddleware(metrics *PrometheusMetrics) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Wrap response writer to capture status code
			wrapped := &responseWriter{ResponseWriter: w, statusCode: 200}

			next.ServeHTTP(wrapped, r)

			duration := time.Since(start)
			metrics.RecordRequest(r.Method, r.URL.Path, http.StatusText(wrapped.statusCode), duration)
		})
	}
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
