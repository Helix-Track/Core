package deduplication

import (
	"context"
	"io"

	"github.com/helixtrack/attachments-service/internal/models"
)

// DeduplicationEngine defines the interface for file deduplication operations
// This interface allows for easy mocking in tests
type DeduplicationEngine interface {
	// ProcessUpload processes a file upload with deduplication
	ProcessUpload(ctx context.Context, reader io.Reader, metadata *UploadMetadata) (*UploadResult, error)

	// ProcessUploadFromPath processes a file upload from a file path with deduplication
	ProcessUploadFromPath(ctx context.Context, filePath string, metadata *UploadMetadata) (*UploadResult, error)

	// DownloadFile retrieves a file by reference ID
	DownloadFile(ctx context.Context, referenceID string) (io.ReadCloser, *models.AttachmentReference, *models.AttachmentFile, error)

	// DeleteReference removes a reference and cleans up the file if no longer referenced
	DeleteReference(ctx context.Context, referenceID string) error

	// CheckDeduplication checks if a file with the given hash exists
	CheckDeduplication(ctx context.Context, hash string) (bool, *models.AttachmentFile, error)

	// GetDeduplicationStats returns deduplication statistics
	GetDeduplicationStats(ctx context.Context) (*DeduplicationStats, error)
}

// Ensure Engine implements DeduplicationEngine interface
var _ DeduplicationEngine = (*Engine)(nil)
