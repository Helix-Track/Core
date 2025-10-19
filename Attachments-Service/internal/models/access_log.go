package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// AccessLog represents an audit log entry for file operations
type AccessLog struct {
	ID           string  `json:"id" db:"id"`
	ReferenceID  *string `json:"reference_id,omitempty" db:"reference_id"`
	FileHash     *string `json:"file_hash,omitempty" db:"file_hash"`
	UserID       *string `json:"user_id,omitempty" db:"user_id"`
	IPAddress    *string `json:"ip_address,omitempty" db:"ip_address"`
	Action       string  `json:"action" db:"action"`
	StatusCode   *int    `json:"status_code,omitempty" db:"status_code"`
	ErrorMessage *string `json:"error_message,omitempty" db:"error_message"`
	UserAgent    *string `json:"user_agent,omitempty" db:"user_agent"`
	Timestamp    int64   `json:"timestamp" db:"timestamp"`
}

// Access log action constants
const (
	ActionUpload         = "upload"
	ActionDownload       = "download"
	ActionDelete         = "delete"
	ActionMetadataRead   = "metadata_read"
	ActionMetadataUpdate = "metadata_update"
)

// NewAccessLog creates a new access log entry
func NewAccessLog(action string) *AccessLog {
	return &AccessLog{
		ID:        uuid.New().String(),
		Action:    action,
		Timestamp: time.Now().Unix(),
	}
}

// Validate validates the access log
func (l *AccessLog) Validate() error {
	if l.ID == "" {
		return fmt.Errorf("id is required")
	}
	if !isValidAction(l.Action) {
		return fmt.Errorf("invalid action: %s", l.Action)
	}
	if l.Timestamp == 0 {
		return fmt.Errorf("timestamp is required")
	}
	if l.StatusCode != nil && (*l.StatusCode < 100 || *l.StatusCode > 599) {
		return fmt.Errorf("invalid status_code: %d", *l.StatusCode)
	}
	return nil
}

// SetReferenceID sets the reference ID
func (l *AccessLog) SetReferenceID(id string) {
	l.ReferenceID = &id
}

// SetFileHash sets the file hash
func (l *AccessLog) SetFileHash(hash string) {
	l.FileHash = &hash
}

// SetUserID sets the user ID
func (l *AccessLog) SetUserID(userID string) {
	l.UserID = &userID
}

// SetIPAddress sets the IP address
func (l *AccessLog) SetIPAddress(ip string) {
	l.IPAddress = &ip
}

// SetStatusCode sets the HTTP status code
func (l *AccessLog) SetStatusCode(code int) {
	l.StatusCode = &code
}

// SetError sets the error message
func (l *AccessLog) SetError(err error) {
	if err != nil {
		msg := err.Error()
		l.ErrorMessage = &msg
	}
}

// SetUserAgent sets the user agent
func (l *AccessLog) SetUserAgent(ua string) {
	l.UserAgent = &ua
}

// IsSuccess checks if the log represents a successful operation
func (l *AccessLog) IsSuccess() bool {
	return l.StatusCode != nil && *l.StatusCode >= 200 && *l.StatusCode < 300
}

// IsError checks if the log represents an error
func (l *AccessLog) IsError() bool {
	return l.StatusCode != nil && *l.StatusCode >= 400
}

// PresignedURL represents a temporary access token for file downloads
type PresignedURL struct {
	Token         string  `json:"token" db:"token"`
	ReferenceID   string  `json:"reference_id" db:"reference_id"`
	UserID        *string `json:"user_id,omitempty" db:"user_id"`
	IPAddress     *string `json:"ip_address,omitempty" db:"ip_address"`
	ExpiresAt     int64   `json:"expires_at" db:"expires_at"`
	MaxDownloads  int     `json:"max_downloads" db:"max_downloads"`
	DownloadCount int     `json:"download_count" db:"download_count"`
	Created       int64   `json:"created" db:"created"`
}

// NewPresignedURL creates a new presigned URL
func NewPresignedURL(referenceID string, expiresIn int) *PresignedURL {
	now := time.Now().Unix()
	return &PresignedURL{
		Token:         uuid.New().String(),
		ReferenceID:   referenceID,
		ExpiresAt:     now + int64(expiresIn),
		MaxDownloads:  1,
		DownloadCount: 0,
		Created:       now,
	}
}

// Validate validates the presigned URL
func (u *PresignedURL) Validate() error {
	if u.Token == "" {
		return fmt.Errorf("token is required")
	}
	if u.ReferenceID == "" {
		return fmt.Errorf("reference_id is required")
	}
	if u.ExpiresAt == 0 {
		return fmt.Errorf("expires_at is required")
	}
	if u.MaxDownloads <= 0 {
		return fmt.Errorf("max_downloads must be positive")
	}
	if u.DownloadCount < 0 {
		return fmt.Errorf("download_count must be non-negative")
	}
	if u.DownloadCount > u.MaxDownloads {
		return fmt.Errorf("download_count exceeds max_downloads")
	}
	if u.Created == 0 {
		return fmt.Errorf("created timestamp is required")
	}
	return nil
}

// IsExpired checks if the URL has expired
func (u *PresignedURL) IsExpired() bool {
	return time.Now().Unix() > u.ExpiresAt
}

// IsExhausted checks if all downloads have been used
func (u *PresignedURL) IsExhausted() bool {
	return u.DownloadCount >= u.MaxDownloads
}

// IsValid checks if the URL is still valid (not expired and not exhausted)
func (u *PresignedURL) IsValid() bool {
	return !u.IsExpired() && !u.IsExhausted()
}

// IncrementDownloadCount increments the download counter
func (u *PresignedURL) IncrementDownloadCount() error {
	if u.IsExpired() {
		return fmt.Errorf("presigned URL has expired")
	}
	if u.IsExhausted() {
		return fmt.Errorf("presigned URL download limit reached")
	}
	u.DownloadCount++
	return nil
}

// GetRemainingDownloads returns the number of remaining downloads
func (u *PresignedURL) GetRemainingDownloads() int {
	remaining := u.MaxDownloads - u.DownloadCount
	if remaining < 0 {
		return 0
	}
	return remaining
}

// GetTimeUntilExpiry returns seconds until expiry
func (u *PresignedURL) GetTimeUntilExpiry() int64 {
	remaining := u.ExpiresAt - time.Now().Unix()
	if remaining < 0 {
		return 0
	}
	return remaining
}

// CleanupJob represents a periodic cleanup job
type CleanupJob struct {
	ID             string  `json:"id" db:"id"`
	JobType        string  `json:"job_type" db:"job_type"`
	Started        int64   `json:"started" db:"started"`
	Completed      *int64  `json:"completed,omitempty" db:"completed"`
	Status         string  `json:"status" db:"status"`
	ItemsProcessed int     `json:"items_processed" db:"items_processed"`
	ItemsDeleted   int     `json:"items_deleted" db:"items_deleted"`
	ErrorMessage   *string `json:"error_message,omitempty" db:"error_message"`
}

// Cleanup job type constants
const (
	JobTypeOrphanFiles      = "orphan_files"
	JobTypeDanglingRefs     = "dangling_refs"
	JobTypeExpiredPresigned = "expired_presigned"
	JobTypeOldHealthData    = "old_health_data"
	JobTypeOldAccessLogs    = "old_access_logs"
)

// Cleanup job status constants
const (
	JobStatusRunning   = "running"
	JobStatusCompleted = "completed"
	JobStatusFailed    = "failed"
)

// NewCleanupJob creates a new cleanup job
func NewCleanupJob(jobType string) *CleanupJob {
	return &CleanupJob{
		ID:             uuid.New().String(),
		JobType:        jobType,
		Started:        time.Now().Unix(),
		Status:         JobStatusRunning,
		ItemsProcessed: 0,
		ItemsDeleted:   0,
	}
}

// Validate validates the cleanup job
func (j *CleanupJob) Validate() error {
	if j.ID == "" {
		return fmt.Errorf("id is required")
	}
	if !isValidJobType(j.JobType) {
		return fmt.Errorf("invalid job_type: %s", j.JobType)
	}
	if !isValidJobStatus(j.Status) {
		return fmt.Errorf("invalid status: %s", j.Status)
	}
	if j.Started == 0 {
		return fmt.Errorf("started timestamp is required")
	}
	if j.ItemsProcessed < 0 {
		return fmt.Errorf("items_processed must be non-negative")
	}
	if j.ItemsDeleted < 0 {
		return fmt.Errorf("items_deleted must be non-negative")
	}
	if j.ItemsDeleted > j.ItemsProcessed {
		return fmt.Errorf("items_deleted cannot exceed items_processed")
	}
	return nil
}

// Complete marks the job as completed
func (j *CleanupJob) Complete() {
	now := time.Now().Unix()
	j.Completed = &now
	j.Status = JobStatusCompleted
}

// Fail marks the job as failed
func (j *CleanupJob) Fail(err error) {
	now := time.Now().Unix()
	j.Completed = &now
	j.Status = JobStatusFailed
	if err != nil {
		msg := err.Error()
		j.ErrorMessage = &msg
	}
}

// IncrementProcessed increments the processed counter
func (j *CleanupJob) IncrementProcessed() {
	j.ItemsProcessed++
}

// IncrementDeleted increments both processed and deleted counters
func (j *CleanupJob) IncrementDeleted() {
	j.ItemsProcessed++
	j.ItemsDeleted++
}

// GetDuration returns the job duration in seconds
func (j *CleanupJob) GetDuration() int64 {
	if j.Completed == nil {
		return time.Now().Unix() - j.Started
	}
	return *j.Completed - j.Started
}

// IsRunning checks if the job is still running
func (j *CleanupJob) IsRunning() bool {
	return j.Status == JobStatusRunning
}

// StorageStats represents overall storage statistics
type StorageStats struct {
	TotalFiles        int64   `json:"total_files"`
	TotalSizeBytes    int64   `json:"total_size_bytes"`
	TotalReferences   int64   `json:"total_references"`
	DeduplicationRate float64 `json:"deduplication_rate"` // Percentage of storage saved
	UniqueFiles       int64   `json:"unique_files"`       // Files with ref_count = 1
	SharedFiles       int64   `json:"shared_files"`       // Files with ref_count > 1
	OrphanedFiles     int64   `json:"orphaned_files"`     // Files with ref_count = 0
	PendingScans      int64   `json:"pending_scans"`      // Files awaiting virus scan
	InfectedFiles     int64   `json:"infected_files"`     // Files marked as infected
}

// Helper functions

func isValidAction(action string) bool {
	validActions := []string{
		ActionUpload, ActionDownload, ActionDelete,
		ActionMetadataRead, ActionMetadataUpdate,
	}
	for _, valid := range validActions {
		if action == valid {
			return true
		}
	}
	return false
}

func isValidJobType(jobType string) bool {
	validTypes := []string{
		JobTypeOrphanFiles, JobTypeDanglingRefs, JobTypeExpiredPresigned,
		JobTypeOldHealthData, JobTypeOldAccessLogs,
	}
	for _, valid := range validTypes {
		if jobType == valid {
			return true
		}
	}
	return false
}

func isValidJobStatus(status string) bool {
	validStatuses := []string{JobStatusRunning, JobStatusCompleted, JobStatusFailed}
	for _, valid := range validStatuses {
		if status == valid {
			return true
		}
	}
	return false
}
