package models

import (
	"fmt"
	"time"
)

// UploadQuota represents per-user upload quotas and usage
type UploadQuota struct {
	UserID    string `json:"user_id" db:"user_id"`
	MaxBytes  int64  `json:"max_bytes" db:"max_bytes"`
	UsedBytes int64  `json:"used_bytes" db:"used_bytes"`
	MaxFiles  int    `json:"max_files" db:"max_files"`
	UsedFiles int    `json:"used_files" db:"used_files"`
	Created   int64  `json:"created" db:"created"`
	Modified  int64  `json:"modified" db:"modified"`
}

// Default quota values
const (
	DefaultMaxBytes int64 = 10737418240 // 10 GB
	DefaultMaxFiles int   = 10000
)

// NewUploadQuota creates a new upload quota with defaults
func NewUploadQuota(userID string) *UploadQuota {
	now := time.Now().Unix()
	return &UploadQuota{
		UserID:    userID,
		MaxBytes:  DefaultMaxBytes,
		UsedBytes: 0,
		MaxFiles:  DefaultMaxFiles,
		UsedFiles: 0,
		Created:   now,
		Modified:  now,
	}
}

// Validate validates the upload quota
func (q *UploadQuota) Validate() error {
	if q.UserID == "" {
		return fmt.Errorf("user_id is required")
	}
	if q.MaxBytes <= 0 {
		return fmt.Errorf("max_bytes must be positive")
	}
	if q.UsedBytes < 0 {
		return fmt.Errorf("used_bytes must be non-negative")
	}
	if q.MaxFiles <= 0 {
		return fmt.Errorf("max_files must be positive")
	}
	if q.UsedFiles < 0 {
		return fmt.Errorf("used_files must be non-negative")
	}
	if q.UsedBytes > q.MaxBytes {
		return fmt.Errorf("used_bytes exceeds max_bytes")
	}
	if q.UsedFiles > q.MaxFiles {
		return fmt.Errorf("used_files exceeds max_files")
	}
	if q.Created == 0 {
		return fmt.Errorf("created timestamp is required")
	}
	return nil
}

// CanUpload checks if the user can upload a file of given size
func (q *UploadQuota) CanUpload(sizeBytes int64) bool {
	return (q.UsedBytes+sizeBytes <= q.MaxBytes) && (q.UsedFiles+1 <= q.MaxFiles)
}

// IncrementUsage increments the usage counters
func (q *UploadQuota) IncrementUsage(sizeBytes int64, fileCount int) error {
	newBytes := q.UsedBytes + sizeBytes
	newFiles := q.UsedFiles + fileCount

	if newBytes > q.MaxBytes {
		return fmt.Errorf("quota exceeded: bytes")
	}
	if newFiles > q.MaxFiles {
		return fmt.Errorf("quota exceeded: files")
	}

	q.UsedBytes = newBytes
	q.UsedFiles = newFiles
	q.Modified = time.Now().Unix()
	return nil
}

// DecrementUsage decrements the usage counters
func (q *UploadQuota) DecrementUsage(sizeBytes int64, fileCount int) {
	q.UsedBytes -= sizeBytes
	q.UsedFiles -= fileCount

	// Ensure non-negative
	if q.UsedBytes < 0 {
		q.UsedBytes = 0
	}
	if q.UsedFiles < 0 {
		q.UsedFiles = 0
	}

	q.Modified = time.Now().Unix()
}

// GetUsagePercent returns the usage percentage (bytes)
func (q *UploadQuota) GetUsagePercent() float64 {
	if q.MaxBytes == 0 {
		return 0
	}
	return (float64(q.UsedBytes) / float64(q.MaxBytes)) * 100
}

// GetFilesUsagePercent returns the files usage percentage
func (q *UploadQuota) GetFilesUsagePercent() float64 {
	if q.MaxFiles == 0 {
		return 0
	}
	return (float64(q.UsedFiles) / float64(q.MaxFiles)) * 100
}

// GetRemainingBytes returns remaining bytes available
func (q *UploadQuota) GetRemainingBytes() int64 {
	remaining := q.MaxBytes - q.UsedBytes
	if remaining < 0 {
		return 0
	}
	return remaining
}

// GetRemainingFiles returns remaining file slots available
func (q *UploadQuota) GetRemainingFiles() int {
	remaining := q.MaxFiles - q.UsedFiles
	if remaining < 0 {
		return 0
	}
	return remaining
}

// IsNearLimit checks if quota is near limit (>90%)
func (q *UploadQuota) IsNearLimit() bool {
	return q.GetUsagePercent() >= 90 || q.GetFilesUsagePercent() >= 90
}

// UserStorageUsage represents aggregated storage usage for a user
type UserStorageUsage struct {
	UserID        string  `json:"user_id"`
	TotalBytes    int64   `json:"total_bytes"`
	TotalFiles    int64   `json:"total_files"`
	QuotaBytes    int64   `json:"quota_bytes"`
	QuotaFiles    int     `json:"quota_files"`
	UsagePercent  float64 `json:"usage_percent"`
	RemainingBytes int64  `json:"remaining_bytes"`
	RemainingFiles int    `json:"remaining_files"`
}

// NewUserStorageUsage creates a UserStorageUsage from UploadQuota
func NewUserStorageUsage(quota *UploadQuota) *UserStorageUsage {
	return &UserStorageUsage{
		UserID:         quota.UserID,
		TotalBytes:     quota.UsedBytes,
		TotalFiles:     int64(quota.UsedFiles),
		QuotaBytes:     quota.MaxBytes,
		QuotaFiles:     quota.MaxFiles,
		UsagePercent:   quota.GetUsagePercent(),
		RemainingBytes: quota.GetRemainingBytes(),
		RemainingFiles: quota.GetRemainingFiles(),
	}
}
