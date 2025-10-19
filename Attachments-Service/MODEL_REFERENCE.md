# Model and Interface Reference

## AttachmentReference Model

```go
type AttachmentReference struct {
    ID          string   `json:"id" db:"id"`
    FileHash    string   `json:"file_hash" db:"file_hash"`
    EntityType  string   `json:"entity_type" db:"entity_type"`
    EntityID    string   `json:"entity_id" db:"entity_id"`
    Filename    string   `json:"filename" db:"filename"`
    Description *string  `json:"description,omitempty" db:"description"`
    UploaderID  string   `json:"uploader_id" db:"uploader_id"`
    Version     int      `json:"version" db:"version"`
    Tags        []string `json:"tags,omitempty" db:"tags"`
    Created     int64    `json:"created" db:"created"`        // NOT CreatedAt
    Modified    int64    `json:"modified" db:"modified"`      // NOT ModifiedAt
    Deleted     bool     `json:"deleted" db:"deleted"`
}
```

## Database Interface

```go
type Database interface {
    Ping() error  // NO context parameter
    // ...
}
```

## PrometheusMetrics RecordDownload

```go
func (m *PrometheusMetrics) RecordDownload(
    status string,        // "success", "error", etc.
    size int64,           // bytes downloaded
    duration time.Duration, // time taken
    cacheHit bool,        // whether cache was hit
)
```

## StorageStats Model

```go
type StorageStats struct {
    TotalFiles        int64   `json:"total_files"`
    TotalSizeBytes    int64   `json:"total_size_bytes"`
    TotalReferences   int64   `json:"total_references"`
    DeduplicationRate float64 `json:"deduplication_rate"`
    UniqueFiles       int64   `json:"unique_files"`
    SharedFiles       int64   `json:"shared_files"`
    OrphanedFiles     int64   `json:"orphaned_files"`
    PendingScans      int64   `json:"pending_scans"`
    InfectedFiles     int64   `json:"infected_files"`
    // NO AverageFileSize field
}
```

## Required Fixes

1. **All handlers**: Change `CreatedAt` → `Created`
2. **admin.go**: Remove context from `Ping()` call
3. **admin.go**: Remove `AverageFileSize` from stats response
4. **download.go**: Fix `RecordDownload` arguments to `(status, size, duration, cacheHit)`
