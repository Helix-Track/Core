package adapters

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"go.uber.org/zap"
)

// LocalAdapter implements StorageAdapter for local filesystem
type LocalAdapter struct {
	basePath string
	logger   *zap.Logger
}

// NewLocalAdapter creates a new local filesystem adapter
func NewLocalAdapter(basePath string, logger *zap.Logger) (*LocalAdapter, error) {
	// Create base directory if it doesn't exist
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create base directory: %w", err)
	}

	// Verify we can write to the directory
	testFile := filepath.Join(basePath, ".write_test")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		return nil, fmt.Errorf("base directory is not writable: %w", err)
	}
	os.Remove(testFile)

	return &LocalAdapter{
		basePath: basePath,
		logger:   logger,
	}, nil
}

// Store stores a file using hash-based sharding
// Hash: abcd1234ef567890... -> /base/ab/cd/abcd1234ef567890...
func (a *LocalAdapter) Store(ctx context.Context, hash string, data io.Reader, size int64) (string, error) {
	if len(hash) < 4 {
		return "", fmt.Errorf("invalid hash length: %d", len(hash))
	}

	// Create sharded directory structure: /ab/cd/
	shard1 := hash[0:2]
	shard2 := hash[2:4]
	shardDir := filepath.Join(a.basePath, shard1, shard2)

	if err := os.MkdirAll(shardDir, 0755); err != nil {
		return "", NewStorageError("mkdir", shardDir, err)
	}

	// Full file path
	filePath := filepath.Join(shardDir, hash)

	// Check if file already exists (deduplication at storage level)
	if _, err := os.Stat(filePath); err == nil {
		a.logger.Debug("file already exists at storage path",
			zap.String("path", filePath),
			zap.String("hash", hash),
		)
		return a.getRelativePath(filePath), nil
	}

	// Write to temporary file first (atomic write)
	tempPath := filePath + ".tmp"
	tempFile, err := os.OpenFile(tempPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return "", NewStorageError("create_temp", tempPath, err)
	}

	// Copy data to temp file
	written, err := io.Copy(tempFile, data)
	tempFile.Close()

	if err != nil {
		os.Remove(tempPath) // Cleanup on error
		return "", NewStorageError("write", tempPath, err)
	}

	// Verify size
	if written != size {
		os.Remove(tempPath)
		return "", fmt.Errorf("size mismatch: expected %d, wrote %d", size, written)
	}

	// Atomic rename
	if err := os.Rename(tempPath, filePath); err != nil {
		os.Remove(tempPath)
		return "", NewStorageError("rename", filePath, err)
	}

	a.logger.Info("file stored successfully",
		zap.String("path", filePath),
		zap.String("hash", hash),
		zap.Int64("size", size),
	)

	return a.getRelativePath(filePath), nil
}

// Retrieve retrieves a file from storage
func (a *LocalAdapter) Retrieve(ctx context.Context, path string) (io.ReadCloser, error) {
	fullPath := a.getFullPath(path)

	file, err := os.Open(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("file not found: %s", path)
		}
		return nil, NewStorageError("open", fullPath, err)
	}

	return file, nil
}

// Delete deletes a file from storage
func (a *LocalAdapter) Delete(ctx context.Context, path string) error {
	fullPath := a.getFullPath(path)

	if err := os.Remove(fullPath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("file not found: %s", path)
		}
		return NewStorageError("delete", fullPath, err)
	}

	a.logger.Info("file deleted successfully",
		zap.String("path", fullPath),
	)

	// Cleanup empty directories
	a.cleanupEmptyDirs(fullPath)

	return nil
}

// Exists checks if a file exists
func (a *LocalAdapter) Exists(ctx context.Context, path string) (bool, error) {
	fullPath := a.getFullPath(path)

	_, err := os.Stat(fullPath)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, NewStorageError("stat", fullPath, err)
}

// GetSize returns the size of a file
func (a *LocalAdapter) GetSize(ctx context.Context, path string) (int64, error) {
	fullPath := a.getFullPath(path)

	info, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, fmt.Errorf("file not found: %s", path)
		}
		return 0, NewStorageError("stat", fullPath, err)
	}

	return info.Size(), nil
}

// GetMetadata returns metadata about a file
func (a *LocalAdapter) GetMetadata(ctx context.Context, path string) (*FileMetadata, error) {
	fullPath := a.getFullPath(path)

	info, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &FileMetadata{
				Path:   path,
				Exists: false,
			}, nil
		}
		return nil, NewStorageError("stat", fullPath, err)
	}

	return &FileMetadata{
		Path:         path,
		Size:         info.Size(),
		LastModified: info.ModTime().Unix(),
		Exists:       true,
	}, nil
}

// Ping checks if storage is accessible
func (a *LocalAdapter) Ping(ctx context.Context) error {
	// Try to create a test file
	testFile := filepath.Join(a.basePath, ".health_check")
	if err := os.WriteFile(testFile, []byte("ok"), 0644); err != nil {
		return NewStorageError("ping", a.basePath, err)
	}

	// Clean up
	os.Remove(testFile)
	return nil
}

// GetCapacity returns storage capacity information
func (a *LocalAdapter) GetCapacity(ctx context.Context) (*CapacityInfo, error) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(a.basePath, &stat); err != nil {
		return nil, NewStorageError("statfs", a.basePath, err)
	}

	// Calculate capacity
	totalBytes := int64(stat.Blocks * uint64(stat.Bsize))
	availableBytes := int64(stat.Bavail * uint64(stat.Bsize))
	usedBytes := totalBytes - availableBytes

	capacity := &CapacityInfo{
		TotalBytes:     totalBytes,
		UsedBytes:      usedBytes,
		AvailableBytes: availableBytes,
	}
	capacity.CalculateUsagePercent()

	return capacity, nil
}

// GetType returns the adapter type
func (a *LocalAdapter) GetType() string {
	return "local"
}

// getFullPath converts relative path to full path
func (a *LocalAdapter) getFullPath(relativePath string) string {
	// If already absolute and within basePath, return as-is
	if filepath.IsAbs(relativePath) && filepath.HasPrefix(relativePath, a.basePath) {
		return relativePath
	}

	// Otherwise join with basePath
	return filepath.Join(a.basePath, relativePath)
}

// getRelativePath converts full path to relative path
func (a *LocalAdapter) getRelativePath(fullPath string) string {
	relPath, err := filepath.Rel(a.basePath, fullPath)
	if err != nil {
		// If can't get relative path, return the full path
		return fullPath
	}
	return relPath
}

// cleanupEmptyDirs removes empty parent directories after file deletion
func (a *LocalAdapter) cleanupEmptyDirs(filePath string) {
	dir := filepath.Dir(filePath)

	// Don't delete the base directory
	if dir == a.basePath || !filepath.HasPrefix(dir, a.basePath) {
		return
	}

	// Check if directory is empty
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) > 0 {
		return
	}

	// Remove empty directory
	if err := os.Remove(dir); err != nil {
		a.logger.Debug("failed to remove empty directory",
			zap.String("dir", dir),
			zap.Error(err),
		)
		return
	}

	a.logger.Debug("removed empty directory",
		zap.String("dir", dir),
	)

	// Recursively cleanup parent directories
	a.cleanupEmptyDirs(dir)
}

// ListFiles lists all files in storage (for admin/debug purposes)
func (a *LocalAdapter) ListFiles() ([]string, error) {
	var files []string

	err := filepath.Walk(a.basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and hidden files
		if info.IsDir() || info.Name()[0] == '.' {
			return nil
		}

		relPath := a.getRelativePath(path)
		files = append(files, relPath)
		return nil
	})

	if err != nil {
		return nil, NewStorageError("list", a.basePath, err)
	}

	return files, nil
}

// GetStorageStats returns storage statistics
func (a *LocalAdapter) GetStorageStats() (*StorageStats, error) {
	stats := &StorageStats{
		FileCount:  0,
		TotalBytes: 0,
	}

	err := filepath.Walk(a.basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and hidden files
		if info.IsDir() || info.Name()[0] == '.' {
			return nil
		}

		stats.FileCount++
		stats.TotalBytes += info.Size()
		return nil
	})

	if err != nil {
		return nil, NewStorageError("stats", a.basePath, err)
	}

	return stats, nil
}

// StorageStats contains storage statistics
type StorageStats struct {
	FileCount  int64
	TotalBytes int64
}

// ValidatePath validates that a path is safe (no path traversal)
func (a *LocalAdapter) ValidatePath(path string) error {
	fullPath := a.getFullPath(path)

	// Resolve to absolute path
	absPath, err := filepath.Abs(fullPath)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}

	// Ensure path is within basePath
	if !filepath.HasPrefix(absPath, a.basePath) {
		return fmt.Errorf("path traversal detected: %s", path)
	}

	return nil
}

// Copy copies a file within storage (for internal operations)
func (a *LocalAdapter) Copy(srcPath, dstPath string) error {
	srcFullPath := a.getFullPath(srcPath)
	dstFullPath := a.getFullPath(dstPath)

	// Open source file
	srcFile, err := os.Open(srcFullPath)
	if err != nil {
		return NewStorageError("open_src", srcFullPath, err)
	}
	defer srcFile.Close()

	// Create destination directory
	dstDir := filepath.Dir(dstFullPath)
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return NewStorageError("mkdir_dst", dstDir, err)
	}

	// Create destination file
	dstFile, err := os.OpenFile(dstFullPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return NewStorageError("create_dst", dstFullPath, err)
	}
	defer dstFile.Close()

	// Copy data
	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return NewStorageError("copy", dstFullPath, err)
	}

	a.logger.Info("file copied successfully",
		zap.String("src", srcPath),
		zap.String("dst", dstPath),
	)

	return nil
}

// SetLastModified sets the last modified time of a file
func (a *LocalAdapter) SetLastModified(path string, modTime time.Time) error {
	fullPath := a.getFullPath(path)

	if err := os.Chtimes(fullPath, modTime, modTime); err != nil {
		return NewStorageError("chtimes", fullPath, err)
	}

	return nil
}
