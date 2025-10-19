package deduplication

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	"github.com/helixtrack/attachments-service/internal/database"
	"github.com/helixtrack/attachments-service/internal/models"
	"github.com/helixtrack/attachments-service/internal/storage/adapters"
	"github.com/helixtrack/attachments-service/internal/utils"
	"go.uber.org/zap"
)

// Engine handles file deduplication logic
type Engine struct {
	db      database.Database
	storage adapters.StorageAdapter
	hasher  *utils.FileHasher
	logger  *zap.Logger
}

// NewEngine creates a new deduplication engine
func NewEngine(db database.Database, storage adapters.StorageAdapter, logger *zap.Logger) *Engine {
	return &Engine{
		db:      db,
		storage: storage,
		hasher:  utils.NewFileHasher(),
		logger:  logger,
	}
}

// UploadResult contains the result of an upload operation
type UploadResult struct {
	FileHash      string
	ReferenceID   string
	SizeBytes     int64
	Deduplicated  bool
	SavedBytes    int64
	File          *models.AttachmentFile
	Reference     *models.AttachmentReference
	StoragePath   string
}

// ProcessUpload processes a file upload with deduplication
func (e *Engine) ProcessUpload(ctx context.Context, reader io.Reader, metadata *UploadMetadata) (*UploadResult, error) {
	startTime := time.Now()

	// Step 1: Calculate hash while reading file into buffer
	// We need to buffer the file because we might need to read it twice
	// (once for hash, once for storage if it's a new file)
	var buffer bytes.Buffer
	hashReader := utils.NewHashReader(io.TeeReader(reader, &buffer))

	// Read all data (calculating hash in the process)
	if _, err := io.Copy(io.Discard, hashReader); err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	hash, size := hashReader.GetHash()

	e.logger.Info("file hash calculated",
		zap.String("hash", hash),
		zap.Int64("size", size),
		zap.Duration("duration", time.Since(startTime)),
	)

	// Step 2: Check if file already exists in database
	existingFile, err := e.db.GetFile(ctx, hash)

	var file *models.AttachmentFile
	var deduplicated bool
	var savedBytes int64

	if err == nil {
		// File exists - deduplication!
		file = existingFile
		deduplicated = true
		savedBytes = size

		e.logger.Info("file deduplicated",
			zap.String("hash", hash),
			zap.Int64("size", size),
			zap.Int("existing_ref_count", file.RefCount),
		)
	} else {
		// File doesn't exist - store it
		storagePath, err := e.storage.Store(ctx, hash, &buffer, size)
		if err != nil {
			return nil, fmt.Errorf("failed to store file: %w", err)
		}

		// Create file record in database
		file = models.NewAttachmentFile(hash, size, metadata.MimeType, storagePath)
		file.Extension = metadata.Extension

		if err := e.db.CreateFile(ctx, file); err != nil {
			// Try to cleanup storage on database error
			e.storage.Delete(ctx, storagePath)
			return nil, fmt.Errorf("failed to create file record: %w", err)
		}

		deduplicated = false
		savedBytes = 0

		e.logger.Info("new file stored",
			zap.String("hash", hash),
			zap.Int64("size", size),
			zap.String("storage_path", storagePath),
		)
	}

	// Step 3: Create reference
	reference := models.NewAttachmentReference(
		hash,
		metadata.EntityType,
		metadata.EntityID,
		metadata.Filename,
		metadata.UploaderID,
	)

	if metadata.Description != "" {
		reference.Description = &metadata.Description
	}

	if len(metadata.Tags) > 0 {
		reference.Tags = metadata.Tags
	}

	if err := e.db.CreateReference(ctx, reference); err != nil {
		// If this is a new file, delete it
		if !deduplicated {
			e.storage.Delete(ctx, file.StoragePrimary)
			e.db.DeleteFile(ctx, hash)
		}
		return nil, fmt.Errorf("failed to create reference: %w", err)
	}

	e.logger.Info("reference created",
		zap.String("reference_id", reference.ID),
		zap.String("file_hash", hash),
		zap.String("entity_type", metadata.EntityType),
		zap.String("entity_id", metadata.EntityID),
	)

	// Step 4: Update file's last accessed time
	file.UpdateLastAccessed()
	if err := e.db.UpdateFile(ctx, file); err != nil {
		e.logger.Warn("failed to update file last accessed time",
			zap.Error(err),
		)
	}

	return &UploadResult{
		FileHash:     hash,
		ReferenceID:  reference.ID,
		SizeBytes:    size,
		Deduplicated: deduplicated,
		SavedBytes:   savedBytes,
		File:         file,
		Reference:    reference,
		StoragePath:  file.StoragePrimary,
	}, nil
}

// ProcessUploadFromPath processes a file upload from a local path
func (e *Engine) ProcessUploadFromPath(ctx context.Context, filePath string, metadata *UploadMetadata) (*UploadResult, error) {
	// Calculate hash from file
	hash, size, err := e.hasher.CalculateHashFromFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate hash: %w", err)
	}

	// Check if file already exists
	existingFile, err := e.db.GetFile(ctx, hash)

	var file *models.AttachmentFile
	var deduplicated bool
	var savedBytes int64

	if err == nil {
		// File exists - deduplication!
		file = existingFile
		deduplicated = true
		savedBytes = size
	} else {
		// Open file for storage
		fileReader, err := adapters.OpenFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to open file: %w", err)
		}
		defer fileReader.Close()

		// Store file
		storagePath, err := e.storage.Store(ctx, hash, fileReader, size)
		if err != nil {
			return nil, fmt.Errorf("failed to store file: %w", err)
		}

		// Create file record
		file = models.NewAttachmentFile(hash, size, metadata.MimeType, storagePath)
		file.Extension = metadata.Extension

		if err := e.db.CreateFile(ctx, file); err != nil {
			e.storage.Delete(ctx, storagePath)
			return nil, fmt.Errorf("failed to create file record: %w", err)
		}

		deduplicated = false
		savedBytes = 0
	}

	// Create reference
	reference := models.NewAttachmentReference(
		hash,
		metadata.EntityType,
		metadata.EntityID,
		metadata.Filename,
		metadata.UploaderID,
	)

	if metadata.Description != "" {
		reference.Description = &metadata.Description
	}

	if len(metadata.Tags) > 0 {
		reference.Tags = metadata.Tags
	}

	if err := e.db.CreateReference(ctx, reference); err != nil {
		if !deduplicated {
			e.storage.Delete(ctx, file.StoragePrimary)
			e.db.DeleteFile(ctx, hash)
		}
		return nil, fmt.Errorf("failed to create reference: %w", err)
	}

	return &UploadResult{
		FileHash:     hash,
		ReferenceID:  reference.ID,
		SizeBytes:    size,
		Deduplicated: deduplicated,
		SavedBytes:   savedBytes,
		File:         file,
		Reference:    reference,
		StoragePath:  file.StoragePrimary,
	}, nil
}

// DownloadFile retrieves a file for download
func (e *Engine) DownloadFile(ctx context.Context, referenceID string) (io.ReadCloser, *models.AttachmentReference, *models.AttachmentFile, error) {
	// Get reference
	reference, err := e.db.GetReference(ctx, referenceID)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("reference not found: %w", err)
	}

	// Get file
	file, err := e.db.GetFile(ctx, reference.FileHash)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("file not found: %w", err)
	}

	// Retrieve from storage
	reader, err := e.storage.Retrieve(ctx, file.StoragePrimary)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to retrieve file from storage: %w", err)
	}

	// Update last accessed time
	file.UpdateLastAccessed()
	go e.db.UpdateFile(context.Background(), file)

	e.logger.Info("file downloaded",
		zap.String("reference_id", referenceID),
		zap.String("file_hash", file.Hash),
		zap.Int64("size", file.SizeBytes),
	)

	return reader, reference, file, nil
}

// DeleteReference deletes a reference and manages file cleanup
func (e *Engine) DeleteReference(ctx context.Context, referenceID string) error {
	// Get reference to find the file hash
	reference, err := e.db.GetReference(ctx, referenceID)
	if err != nil {
		return fmt.Errorf("reference not found: %w", err)
	}

	fileHash := reference.FileHash

	// Delete the reference (this triggers ref_count decrement via database trigger)
	if err := e.db.DeleteReference(ctx, referenceID); err != nil {
		return fmt.Errorf("failed to delete reference: %w", err)
	}

	e.logger.Info("reference deleted",
		zap.String("reference_id", referenceID),
		zap.String("file_hash", fileHash),
	)

	// Check if file is now orphaned (ref_count = 0)
	file, err := e.db.GetFile(ctx, fileHash)
	if err != nil {
		// File might have been deleted by cleanup job
		return nil
	}

	if file.RefCount == 0 {
		// File is orphaned - mark for deletion
		if err := e.db.DeleteFile(ctx, fileHash); err != nil {
			e.logger.Error("failed to mark orphaned file for deletion",
				zap.String("file_hash", fileHash),
				zap.Error(err),
			)
		} else {
			e.logger.Info("orphaned file marked for deletion",
				zap.String("file_hash", fileHash),
			)
		}
	}

	return nil
}

// CheckDeduplication checks if a file would be deduplicated
func (e *Engine) CheckDeduplication(ctx context.Context, hash string) (bool, *models.AttachmentFile, error) {
	file, err := e.db.GetFile(ctx, hash)
	if err != nil {
		// File doesn't exist - not deduplicated
		return false, nil, nil
	}

	// File exists - would be deduplicated
	return true, file, nil
}

// GetDeduplicationStats returns deduplication statistics
func (e *Engine) GetDeduplicationStats(ctx context.Context) (*DeduplicationStats, error) {
	stats, err := e.db.GetStorageStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get storage stats: %w", err)
	}

	return &DeduplicationStats{
		TotalFiles:        stats.TotalFiles,
		TotalReferences:   stats.TotalReferences,
		UniqueFiles:       stats.UniqueFiles,
		SharedFiles:       stats.SharedFiles,
		DeduplicationRate: stats.DeduplicationRate,
		SavedFiles:        stats.TotalReferences - stats.TotalFiles,
	}, nil
}

// UploadMetadata contains metadata for file upload
type UploadMetadata struct {
	EntityType  string
	EntityID    string
	Filename    string
	UploaderID  string
	MimeType    string
	Extension   string
	Description string
	Tags        []string
}

// DeduplicationStats contains deduplication statistics
type DeduplicationStats struct {
	TotalFiles        int64
	TotalReferences   int64
	UniqueFiles       int64
	SharedFiles       int64
	DeduplicationRate float64
	SavedFiles        int64
}

// OpenFile opens a file for reading (helper function)
func OpenFile(path string) (io.ReadCloser, error) {
	return adapters.OpenFile(path)
}
