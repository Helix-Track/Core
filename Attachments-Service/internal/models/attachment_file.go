package models

import (
	"fmt"
	"strings"
	"time"
)

// AttachmentFile represents a physical file stored once per unique hash
type AttachmentFile struct {
	Hash             string   `json:"hash" db:"hash"`
	SizeBytes        int64    `json:"size_bytes" db:"size_bytes"`
	MimeType         string   `json:"mime_type" db:"mime_type"`
	Extension        string   `json:"extension,omitempty" db:"extension"`
	RefCount         int      `json:"ref_count" db:"ref_count"`
	StoragePrimary   string   `json:"storage_primary" db:"storage_primary"`
	StorageBackup    *string  `json:"storage_backup,omitempty" db:"storage_backup"`
	StorageMirrors   []string `json:"storage_mirrors,omitempty" db:"storage_mirrors"`
	VirusScanStatus  string   `json:"virus_scan_status" db:"virus_scan_status"`
	VirusScanDate    *int64   `json:"virus_scan_date,omitempty" db:"virus_scan_date"`
	VirusScanResult  *string  `json:"virus_scan_result,omitempty" db:"virus_scan_result"`
	Created          int64    `json:"created" db:"created"`
	LastAccessed     int64    `json:"last_accessed" db:"last_accessed"`
	Deleted          bool     `json:"deleted" db:"deleted"`
}

// Virus scan status constants
const (
	VirusScanPending  = "pending"
	VirusScanClean    = "clean"
	VirusScanInfected = "infected"
	VirusScanFailed   = "failed"
	VirusScanSkipped  = "skipped"
)

// NewAttachmentFile creates a new AttachmentFile with default values
func NewAttachmentFile(hash string, sizeBytes int64, mimeType string, storagePrimary string) *AttachmentFile {
	now := time.Now().Unix()
	return &AttachmentFile{
		Hash:            hash,
		SizeBytes:       sizeBytes,
		MimeType:        mimeType,
		Extension:       extractExtension(mimeType),
		RefCount:        1,
		StoragePrimary:  storagePrimary,
		VirusScanStatus: VirusScanPending,
		Created:         now,
		LastAccessed:    now,
		Deleted:         false,
	}
}

// Validate validates the attachment file
func (f *AttachmentFile) Validate() error {
	if f.Hash == "" {
		return fmt.Errorf("hash is required")
	}
	if len(f.Hash) != 64 {
		return fmt.Errorf("hash must be 64 characters (SHA-256)")
	}
	if f.SizeBytes < 0 {
		return fmt.Errorf("size_bytes must be non-negative")
	}
	if f.MimeType == "" {
		return fmt.Errorf("mime_type is required")
	}
	if f.StoragePrimary == "" {
		return fmt.Errorf("storage_primary is required")
	}
	if f.RefCount < 0 {
		return fmt.Errorf("ref_count must be non-negative")
	}
	if f.VirusScanStatus == "" {
		return fmt.Errorf("virus_scan_status is required")
	}
	if !isValidVirusScanStatus(f.VirusScanStatus) {
		return fmt.Errorf("invalid virus_scan_status: %s", f.VirusScanStatus)
	}
	if f.Created == 0 {
		return fmt.Errorf("created timestamp is required")
	}
	if f.LastAccessed == 0 {
		return fmt.Errorf("last_accessed timestamp is required")
	}
	return nil
}

// IsImage checks if the file is an image
func (f *AttachmentFile) IsImage() bool {
	imageTypes := []string{
		"image/jpeg", "image/jpg", "image/png", "image/gif",
		"image/webp", "image/svg+xml", "image/bmp", "image/tiff",
	}
	for _, t := range imageTypes {
		if strings.EqualFold(f.MimeType, t) {
			return true
		}
	}
	return false
}

// IsDocument checks if the file is a document
func (f *AttachmentFile) IsDocument() bool {
	docTypes := []string{
		"application/pdf",
		"text/plain", "text/markdown", "text/csv",
		"application/msword",
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		"application/vnd.ms-excel",
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		"application/vnd.ms-powerpoint",
		"application/vnd.openxmlformats-officedocument.presentationml.presentation",
	}
	for _, t := range docTypes {
		if strings.EqualFold(f.MimeType, t) {
			return true
		}
	}
	return false
}

// IsVideo checks if the file is a video
func (f *AttachmentFile) IsVideo() bool {
	videoTypes := []string{
		"video/mp4", "video/mpeg", "video/webm", "video/quicktime",
		"video/x-msvideo", "video/x-matroska",
	}
	for _, t := range videoTypes {
		if strings.EqualFold(f.MimeType, t) {
			return true
		}
	}
	return false
}

// IsArchive checks if the file is an archive
func (f *AttachmentFile) IsArchive() bool {
	archiveTypes := []string{
		"application/zip", "application/x-tar", "application/gzip",
		"application/x-7z-compressed", "application/x-rar-compressed",
	}
	for _, t := range archiveTypes {
		if strings.EqualFold(f.MimeType, t) {
			return true
		}
	}
	return false
}

// GetHumanReadableSize returns a human-readable file size
func (f *AttachmentFile) GetHumanReadableSize() string {
	return FormatBytes(f.SizeBytes)
}

// MarkAsScanned marks the file as virus scanned
func (f *AttachmentFile) MarkAsScanned(status string, result *string) {
	f.VirusScanStatus = status
	now := time.Now().Unix()
	f.VirusScanDate = &now
	f.VirusScanResult = result
}

// UpdateLastAccessed updates the last accessed timestamp
func (f *AttachmentFile) UpdateLastAccessed() {
	f.LastAccessed = time.Now().Unix()
}

// isValidVirusScanStatus checks if the virus scan status is valid
func isValidVirusScanStatus(status string) bool {
	validStatuses := []string{
		VirusScanPending, VirusScanClean, VirusScanInfected,
		VirusScanFailed, VirusScanSkipped,
	}
	for _, s := range validStatuses {
		if s == status {
			return true
		}
	}
	return false
}

// extractExtension extracts file extension from MIME type
func extractExtension(mimeType string) string {
	extensions := map[string]string{
		"image/jpeg":      "jpg",
		"image/png":       "png",
		"image/gif":       "gif",
		"image/webp":      "webp",
		"image/svg+xml":   "svg",
		"application/pdf": "pdf",
		"text/plain":      "txt",
		"text/markdown":   "md",
		"text/csv":        "csv",
		"application/zip": "zip",
		"application/x-tar": "tar",
		"application/gzip": "gz",
		"video/mp4":       "mp4",
		"video/webm":      "webm",
		"video/quicktime": "mov",
		"application/msword": "doc",
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document": "docx",
		"application/vnd.ms-excel": "xls",
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet": "xlsx",
	}

	if ext, ok := extensions[strings.ToLower(mimeType)]; ok {
		return ext
	}

	// Extract from MIME type (e.g., "image/png" -> "png")
	parts := strings.Split(mimeType, "/")
	if len(parts) == 2 {
		return parts[1]
	}

	return ""
}

// FormatBytes formats bytes to human-readable format
func FormatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	units := []string{"KB", "MB", "GB", "TB", "PB", "EB"}
	return fmt.Sprintf("%.1f %s", float64(bytes)/float64(div), units[exp])
}
