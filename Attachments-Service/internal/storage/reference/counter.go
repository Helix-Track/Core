package reference

import (
	"context"
	"fmt"
	"time"

	"github.com/helixtrack/attachments-service/internal/database"
	"github.com/helixtrack/attachments-service/internal/models"
	"go.uber.org/zap"
)

// Counter handles reference counting for files
type Counter struct {
	db     database.Database
	logger *zap.Logger
}

// NewCounter creates a new reference counter
func NewCounter(db database.Database, logger *zap.Logger) *Counter {
	return &Counter{
		db:     db,
		logger: logger,
	}
}

// Increment increments the reference count for a file
// This is typically called when a new reference is created
func (c *Counter) Increment(ctx context.Context, fileHash string) error {
	if err := c.db.IncrementRefCount(ctx, fileHash); err != nil {
		return fmt.Errorf("failed to increment ref count: %w", err)
	}

	c.logger.Debug("reference count incremented",
		zap.String("file_hash", fileHash),
	)

	return nil
}

// Decrement decrements the reference count for a file
// This is typically called when a reference is deleted
func (c *Counter) Decrement(ctx context.Context, fileHash string) error {
	if err := c.db.DecrementRefCount(ctx, fileHash); err != nil {
		return fmt.Errorf("failed to decrement ref count: %w", err)
	}

	c.logger.Debug("reference count decremented",
		zap.String("file_hash", fileHash),
	)

	// Check if file is now orphaned
	file, err := c.db.GetFile(ctx, fileHash)
	if err != nil {
		return nil // File might have been deleted already
	}

	if file.RefCount == 0 {
		c.logger.Info("file is now orphaned",
			zap.String("file_hash", fileHash),
			zap.Int64("size_bytes", file.SizeBytes),
		)
	}

	return nil
}

// GetCount returns the current reference count for a file
func (c *Counter) GetCount(ctx context.Context, fileHash string) (int, error) {
	file, err := c.db.GetFile(ctx, fileHash)
	if err != nil {
		return 0, fmt.Errorf("file not found: %w", err)
	}

	return file.RefCount, nil
}

// GetReferences returns all references for a file
func (c *Counter) GetReferences(ctx context.Context, fileHash string) ([]*models.AttachmentReference, error) {
	references, err := c.db.ListReferencesByHash(ctx, fileHash)
	if err != nil {
		return nil, fmt.Errorf("failed to list references: %w", err)
	}

	return references, nil
}

// FindOrphaned finds files with zero references
func (c *Counter) FindOrphaned(ctx context.Context, retentionDays int) ([]*models.AttachmentFile, error) {
	orphaned, err := c.db.GetOrphanedFiles(ctx, retentionDays)
	if err != nil {
		return nil, fmt.Errorf("failed to find orphaned files: %w", err)
	}

	c.logger.Info("found orphaned files",
		zap.Int("count", len(orphaned)),
		zap.Int("retention_days", retentionDays),
	)

	return orphaned, nil
}

// CleanupOrphaned removes orphaned files from the database
// Returns the number of files deleted
func (c *Counter) CleanupOrphaned(ctx context.Context, retentionDays int) (int64, error) {
	// Find orphaned files
	orphaned, err := c.FindOrphaned(ctx, retentionDays)
	if err != nil {
		return 0, err
	}

	if len(orphaned) == 0 {
		c.logger.Info("no orphaned files to cleanup")
		return 0, nil
	}

	// Extract hashes
	hashes := make([]string, len(orphaned))
	var totalBytes int64
	for i, file := range orphaned {
		hashes[i] = file.Hash
		totalBytes += file.SizeBytes
	}

	// Delete from database
	deleted, err := c.db.DeleteOrphanedFiles(ctx, hashes)
	if err != nil {
		return 0, fmt.Errorf("failed to delete orphaned files: %w", err)
	}

	c.logger.Info("orphaned files cleaned up",
		zap.Int64("count", deleted),
		zap.Int64("total_bytes", totalBytes),
	)

	return deleted, nil
}

// VerifyIntegrity verifies reference counting integrity
// Returns mismatches between database ref_count and actual reference count
func (c *Counter) VerifyIntegrity(ctx context.Context) ([]*IntegrityIssue, error) {
	issues := []*IntegrityIssue{}

	// Get all files
	files, _, err := c.db.ListFiles(ctx, &database.FileFilter{
		Limit:  10000, // Process in batches for large datasets
		Offset: 0,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}

	for _, file := range files {
		// Count actual references
		references, err := c.db.ListReferencesByHash(ctx, file.Hash)
		if err != nil {
			c.logger.Error("failed to count references",
				zap.String("file_hash", file.Hash),
				zap.Error(err),
			)
			continue
		}

		actualCount := len(references)

		if actualCount != file.RefCount {
			issue := &IntegrityIssue{
				FileHash:      file.Hash,
				DatabaseCount: file.RefCount,
				ActualCount:   actualCount,
				File:          file,
			}
			issues = append(issues, issue)

			c.logger.Warn("reference count mismatch",
				zap.String("file_hash", file.Hash),
				zap.Int("database_count", file.RefCount),
				zap.Int("actual_count", actualCount),
			)
		}
	}

	if len(issues) > 0 {
		c.logger.Warn("integrity issues found",
			zap.Int("count", len(issues)),
		)
	} else {
		c.logger.Info("no integrity issues found")
	}

	return issues, nil
}

// RepairIntegrity repairs reference counting integrity issues
func (c *Counter) RepairIntegrity(ctx context.Context) (int, error) {
	issues, err := c.VerifyIntegrity(ctx)
	if err != nil {
		return 0, err
	}

	repaired := 0
	for _, issue := range issues {
		// Update database ref_count to match actual count
		file := issue.File
		file.RefCount = issue.ActualCount

		if err := c.db.UpdateFile(ctx, file); err != nil {
			c.logger.Error("failed to repair ref count",
				zap.String("file_hash", file.Hash),
				zap.Error(err),
			)
			continue
		}

		c.logger.Info("reference count repaired",
			zap.String("file_hash", file.Hash),
			zap.Int("old_count", issue.DatabaseCount),
			zap.Int("new_count", issue.ActualCount),
		)

		repaired++
	}

	return repaired, nil
}

// GetStatistics returns reference counting statistics
func (c *Counter) GetStatistics(ctx context.Context) (*Statistics, error) {
	stats, err := c.db.GetStorageStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get storage stats: %w", err)
	}

	return &Statistics{
		TotalFiles:        stats.TotalFiles,
		TotalReferences:   stats.TotalReferences,
		UniqueFiles:       stats.UniqueFiles,
		SharedFiles:       stats.SharedFiles,
		OrphanedFiles:     stats.OrphanedFiles,
		DeduplicationRate: stats.DeduplicationRate,
		AverageRefsPerFile: func() float64 {
			if stats.TotalFiles > 0 {
				return float64(stats.TotalReferences) / float64(stats.TotalFiles)
			}
			return 0
		}(),
	}, nil
}

// ScheduleCleanup schedules periodic cleanup of orphaned files
func (c *Counter) ScheduleCleanup(ctx context.Context, interval time.Duration, retentionDays int) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	c.logger.Info("orphan cleanup scheduler started",
		zap.Duration("interval", interval),
		zap.Int("retention_days", retentionDays),
	)

	for {
		select {
		case <-ticker.C:
			deleted, err := c.CleanupOrphaned(ctx, retentionDays)
			if err != nil {
				c.logger.Error("orphan cleanup failed",
					zap.Error(err),
				)
			} else if deleted > 0 {
				c.logger.Info("orphan cleanup completed",
					zap.Int64("deleted", deleted),
				)
			}

		case <-ctx.Done():
			c.logger.Info("orphan cleanup scheduler stopped")
			return
		}
	}
}

// IntegrityIssue represents a reference count mismatch
type IntegrityIssue struct {
	FileHash      string
	DatabaseCount int
	ActualCount   int
	File          *models.AttachmentFile
}

// Statistics contains reference counting statistics
type Statistics struct {
	TotalFiles         int64
	TotalReferences    int64
	UniqueFiles        int64
	SharedFiles        int64
	OrphanedFiles      int64
	DeduplicationRate  float64
	AverageRefsPerFile float64
}

// GetFileUsage returns files with the most references
func (c *Counter) GetFileUsage(ctx context.Context, limit int) ([]*FileUsage, error) {
	// This would require a custom query or modification to ListFiles
	// For now, return an error indicating not implemented
	return nil, fmt.Errorf("not implemented")
}

// FileUsage represents file usage statistics
type FileUsage struct {
	FileHash   string
	RefCount   int
	SizeBytes  int64
	MimeType   string
	Created    int64
	References []*models.AttachmentReference
}

// AtomicIncrement performs an atomic increment with retry logic
func (c *Counter) AtomicIncrement(ctx context.Context, fileHash string, maxRetries int) error {
	for i := 0; i < maxRetries; i++ {
		err := c.Increment(ctx, fileHash)
		if err == nil {
			return nil
		}

		// Retry on transient errors
		if i < maxRetries-1 {
			time.Sleep(time.Millisecond * 100 * time.Duration(i+1))
			continue
		}

		return err
	}

	return fmt.Errorf("max retries exceeded")
}

// AtomicDecrement performs an atomic decrement with retry logic
func (c *Counter) AtomicDecrement(ctx context.Context, fileHash string, maxRetries int) error {
	for i := 0; i < maxRetries; i++ {
		err := c.Decrement(ctx, fileHash)
		if err == nil {
			return nil
		}

		// Retry on transient errors
		if i < maxRetries-1 {
			time.Sleep(time.Millisecond * 100 * time.Duration(i+1))
			continue
		}

		return err
	}

	return fmt.Errorf("max retries exceeded")
}
