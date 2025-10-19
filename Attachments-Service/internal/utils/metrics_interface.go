package utils

import "time"

// MetricsRecorder defines the interface for recording metrics
// This interface allows for easy mocking in tests
type MetricsRecorder interface {
	// RecordUpload records an upload operation
	RecordUpload(status, mimeType string, size int64, duration time.Duration)

	// RecordDownload records a download operation
	RecordDownload(status string, size int64, duration time.Duration, cacheHit bool)

	// RecordDelete records a delete operation
	RecordDelete(status string)

	// RecordDeduplication records a deduplication operation
	RecordDeduplication(deduplicated bool, savedBytes int64)

	// RecordVirusScan records a virus scan operation
	RecordVirusScan(status string)

	// RecordError records an error
	RecordError(errorType string, operation string)
}

// Ensure PrometheusMetrics implements MetricsRecorder interface
var _ MetricsRecorder = (*PrometheusMetrics)(nil)
