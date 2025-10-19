package scanner

import (
	"context"
	"io"
)

// SecurityScanner defines the interface for file security scanning
// This interface allows for easy mocking in tests
type SecurityScanner interface {
	// Scan performs a comprehensive security scan on a file
	Scan(ctx context.Context, reader io.Reader, filename string) (*ScanResult, error)

	// ScanFile performs a security scan on a file from the filesystem
	ScanFile(ctx context.Context, filePath string) (*ScanResult, error)

	// IsEnabled returns whether virus scanning is enabled
	IsEnabled() bool

	// Ping checks if the security scanner (ClamAV) is accessible
	Ping(ctx context.Context) error
}

// Ensure Scanner implements SecurityScanner interface
var _ SecurityScanner = (*Scanner)(nil)
