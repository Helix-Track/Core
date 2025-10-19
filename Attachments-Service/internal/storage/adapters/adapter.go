package adapters

import (
	"context"
	"io"
)

// StorageAdapter defines the interface for storage backends
type StorageAdapter interface {
	// Store stores a file and returns the storage path
	// ctx: request context for cancellation and tracing
	// hash: SHA-256 hash of the file
	// data: file content reader
	// size: file size in bytes
	// Returns: storage path, error
	Store(ctx context.Context, hash string, data io.Reader, size int64) (string, error)

	// Retrieve retrieves a file by storage path
	// ctx: request context for cancellation and tracing
	// path: storage path returned by Store
	// Returns: file content reader, error
	Retrieve(ctx context.Context, path string) (io.ReadCloser, error)

	// Delete deletes a file by storage path
	// ctx: request context for cancellation and tracing
	// path: storage path returned by Store
	// Returns: error
	Delete(ctx context.Context, path string) error

	// Exists checks if a file exists at the given path
	// ctx: request context for cancellation and tracing
	// path: storage path to check
	// Returns: true if exists, error
	Exists(ctx context.Context, path string) (bool, error)

	// GetSize returns the size of a file in bytes
	// ctx: request context for cancellation and tracing
	// path: storage path
	// Returns: size in bytes, error
	GetSize(ctx context.Context, path string) (int64, error)

	// GetMetadata returns metadata about a file
	// ctx: request context for cancellation and tracing
	// path: storage path
	// Returns: file metadata, error
	GetMetadata(ctx context.Context, path string) (*FileMetadata, error)

	// Ping checks if the storage backend is accessible
	// ctx: request context for cancellation and tracing
	// Returns: error if not accessible
	Ping(ctx context.Context) error

	// GetCapacity returns storage capacity information
	// ctx: request context for cancellation and tracing
	// Returns: capacity info, error
	GetCapacity(ctx context.Context) (*CapacityInfo, error)

	// GetType returns the adapter type (local, s3, minio, etc.)
	GetType() string
}

// FileMetadata contains metadata about a stored file
type FileMetadata struct {
	Path         string
	Size         int64
	LastModified int64
	Exists       bool
}

// CapacityInfo contains capacity information for storage
type CapacityInfo struct {
	TotalBytes     int64
	UsedBytes      int64
	AvailableBytes int64
	UsagePercent   float64
}

// CalculateUsagePercent calculates usage percentage
func (c *CapacityInfo) CalculateUsagePercent() {
	if c.TotalBytes > 0 {
		c.UsagePercent = (float64(c.UsedBytes) / float64(c.TotalBytes)) * 100
	}
}

// IsNearCapacity checks if storage is near capacity (>90%)
func (c *CapacityInfo) IsNearCapacity() bool {
	return c.UsagePercent >= 90
}

// IsFull checks if storage is full (>95%)
func (c *CapacityInfo) IsFull() bool {
	return c.UsagePercent >= 95
}

// AdapterConfig contains configuration for storage adapters
type AdapterConfig struct {
	Type   string
	Config map[string]interface{}
}

// StorageError represents a storage-specific error
type StorageError struct {
	Operation string
	Path      string
	Err       error
}

func (e *StorageError) Error() string {
	if e.Path != "" {
		return e.Operation + " failed for " + e.Path + ": " + e.Err.Error()
	}
	return e.Operation + " failed: " + e.Err.Error()
}

func (e *StorageError) Unwrap() error {
	return e.Err
}

// NewStorageError creates a new storage error
func NewStorageError(operation, path string, err error) *StorageError {
	return &StorageError{
		Operation: operation,
		Path:      path,
		Err:       err,
	}
}
