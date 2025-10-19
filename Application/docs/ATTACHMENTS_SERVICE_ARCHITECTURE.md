# HelixTrack Attachments Service - Architecture Design

**Version:** 1.0.0
**Status:** Design Phase
**Last Updated:** 2025-10-19

---

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [Design Goals](#design-goals)
3. [Architecture Overview](#architecture-overview)
4. [Core Components](#core-components)
5. [Storage Architecture](#storage-architecture)
6. [Security Architecture](#security-architecture)
7. [High Availability & Reliability](#high-availability--reliability)
8. [API Design](#api-design)
9. [Database Schema](#database-schema)
10. [Service Discovery & Configuration](#service-discovery--configuration)
11. [Performance & Scalability](#performance--scalability)
12. [Testing Strategy](#testing-strategy)
13. [Deployment Architecture](#deployment-architecture)
14. [Migration Plan](#migration-plan)

---

## 1. Executive Summary

The **HelixTrack Attachments Service** is a decoupled, S3-compatible microservice designed to provide enterprise-grade file storage, retrieval, and management capabilities for the HelixTrack ecosystem.

### Key Features

- **S3-Compatible API** - AWS S3-like interface for easy integration
- **Hash-Based Deduplication** - Single storage per unique file (SHA-256)
- **Reference Counting** - Track file usage across entities
- **Multi-Endpoint Storage** - Primary + fallback + mirror storage
- **Service Discovery** - Automatic port selection and registration
- **Military-Grade Security** - Anti-DDoS, penetration testing resistant
- **Zero Deadlocks** - Lock-free architecture with timeouts
- **100% Test Coverage** - Unit, integration, AI QA, and E2E tests

---

## 2. Design Goals

### Functional Requirements

1. **File Operations**
   - Upload files with automatic deduplication
   - Download files with streaming support
   - Delete files with reference counting
   - List files with pagination and filtering
   - Generate presigned URLs for temporary access

2. **Storage Management**
   - Multi-endpoint storage (local, S3, MinIO, etc.)
   - Automatic failover to backup storage
   - Configurable replication (mirrors)
   - Storage health monitoring
   - Automatic cleanup of orphaned files

3. **Security**
   - JWT authentication integration
   - MIME type validation and whitelisting
   - File content validation (magic bytes)
   - Virus scanning integration (ClamAV)
   - Rate limiting per user/IP
   - Request signing for integrity

4. **High Availability**
   - Multiple storage endpoint support
   - Automatic failover on endpoint failure
   - Read replicas for high-traffic files
   - Circuit breaker pattern
   - Graceful degradation

### Non-Functional Requirements

1. **Performance**
   - Upload: 100+ concurrent uploads
   - Download: 1000+ concurrent downloads
   - Latency: <100ms for metadata operations
   - Throughput: Support for GB-sized files

2. **Scalability**
   - Horizontal scaling (multiple service instances)
   - Storage scaling (add endpoints dynamically)
   - No single point of failure

3. **Reliability**
   - 99.9% uptime target
   - Zero data loss (multi-endpoint replication)
   - Automatic recovery from failures
   - Comprehensive error handling

4. **Security**
   - Resistant to DDoS attacks (rate limiting, connection limits)
   - Resistant to penetration attempts (input validation)
   - Encrypted storage support (AES-256)
   - Audit logging for all operations

5. **Testability**
   - 100% unit test coverage
   - Full integration test suite
   - AI QA automation
   - E2E workflow testing

---

## 3. Architecture Overview

### 3.1 Microservices Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     HelixTrack Ecosystem                     │
├─────────────────────────────────────────────────────────────┤
│                                                               │
│  ┌──────────┐      ┌──────────┐      ┌──────────┐          │
│  │   Core   │      │   Auth   │      │   Perms  │          │
│  │ Service  │──────│ Service  │──────│ Service  │          │
│  └────┬─────┘      └──────────┘      └──────────┘          │
│       │                                                       │
│       │ HTTP/REST                                            │
│       │                                                       │
│  ┌────▼──────────────────────────────────────────────┐      │
│  │        Attachments Service (NEW)                  │      │
│  │  ┌──────────────────────────────────────────┐    │      │
│  │  │         API Gateway Layer                 │    │      │
│  │  │  - Authentication                         │    │      │
│  │  │  - Rate Limiting                          │    │      │
│  │  │  - Request Validation                     │    │      │
│  │  └──────────────┬───────────────────────────┘    │      │
│  │                 │                                  │      │
│  │  ┌──────────────▼───────────────────────────┐    │      │
│  │  │      Business Logic Layer                │    │      │
│  │  │  - Deduplication Engine                  │    │      │
│  │  │  - Reference Counter                     │    │      │
│  │  │  - Security Scanner                      │    │      │
│  │  │  - Metadata Manager                      │    │      │
│  │  └──────────────┬───────────────────────────┘    │      │
│  │                 │                                  │      │
│  │  ┌──────────────▼───────────────────────────┐    │      │
│  │  │      Storage Orchestration Layer         │    │      │
│  │  │  - Multi-Endpoint Manager                │    │      │
│  │  │  - Failover Controller                   │    │      │
│  │  │  - Replication Manager                   │    │      │
│  │  │  - Health Monitor                        │    │      │
│  │  └──────────────┬───────────────────────────┘    │      │
│  │                 │                                  │      │
│  │  ┌──────────────▼───────────────────────────┐    │      │
│  │  │         Storage Adapters                 │    │      │
│  │  │  - Local Filesystem Adapter              │    │      │
│  │  │  - S3 Adapter (AWS/MinIO/etc)            │    │      │
│  │  │  - Custom Storage Adapter                │    │      │
│  │  └──────────────────────────────────────────┘    │      │
│  └────────────────┬──────────────────────────────────┘      │
│                   │                                          │
│  ┌────────────────▼──────────────────────────────────┐      │
│  │         PostgreSQL Database                       │      │
│  │  - attachment_file (hash → physical file)         │      │
│  │  - attachment_reference (entity → hash mapping)   │      │
│  │  - storage_endpoint (endpoint configurations)     │      │
│  │  - storage_health (endpoint health status)        │      │
│  └───────────────────────────────────────────────────┘      │
│                                                               │
└─────────────────────────────────────────────────────────────┘

         │                    │                    │
         ▼                    ▼                    ▼
   ┌──────────┐        ┌──────────┐        ┌──────────┐
   │ Storage  │        │ Storage  │        │ Storage  │
   │Endpoint 1│        │Endpoint 2│        │Endpoint 3│
   │(Primary) │        │(Backup)  │        │(Mirror)  │
   └──────────┘        └──────────┘        └──────────┘
   Local FS            AWS S3              MinIO
```

### 3.2 Service Communication

- **Core → Attachments**: HTTP/REST API with JWT authentication
- **Attachments → Storage**: Storage adapter abstraction layer
- **Attachments → Database**: PostgreSQL connection pool
- **Attachments → ClamAV**: Unix socket or TCP for virus scanning

---

## 4. Core Components

### 4.1 API Gateway Layer

**Responsibilities:**
- JWT token validation
- Rate limiting (per user, per IP)
- Request size validation
- Request logging and metrics
- CORS handling
- API versioning

**Technologies:**
- Gin Gonic middleware
- Token bucket rate limiter
- Prometheus metrics

---

### 4.2 Business Logic Layer

#### 4.2.1 Deduplication Engine

**Algorithm:**
1. Calculate SHA-256 hash of uploaded file
2. Check if hash exists in `attachment_file` table
3. If exists: Create reference only, increment ref count
4. If new: Store file + create reference

**Benefits:**
- Save storage space (no duplicate files)
- Faster uploads for duplicate files
- Consistent file identity

**Implementation:**
```go
type DeduplicationEngine struct {
    db      *Database
    hasher  *FileHasher
    storage *StorageOrchestrator
}

func (e *DeduplicationEngine) ProcessUpload(file io.Reader, metadata FileMetadata) (*AttachmentReference, error) {
    // Calculate hash while reading file once
    hash, size, err := e.hasher.CalculateHash(file)

    // Check if file exists
    existing, err := e.db.GetFileByHash(hash)
    if err == nil {
        // File exists, create reference only
        return e.createReference(existing, metadata)
    }

    // New file, store and create reference
    physicalFile, err := e.storage.StoreFile(hash, file, size)
    reference, err := e.createReference(physicalFile, metadata)

    return reference, nil
}
```

---

#### 4.2.2 Reference Counter

**Data Model:**
- `attachment_file`: Physical file (1 per unique hash)
- `attachment_reference`: Logical reference (N per hash)

**Reference Lifecycle:**
1. **Create**: Add reference when entity attaches file
2. **Read**: Return file via any reference
3. **Delete**: Remove reference, decrement count
4. **Cleanup**: Delete physical file when ref_count = 0

**Concurrency Safety:**
- Atomic increment/decrement operations
- Transaction-based ref count updates
- Periodic orphan cleanup job

---

#### 4.2.3 Security Scanner

**Multi-Layer Validation:**

1. **MIME Type Validation**
   ```go
   var AllowedMimeTypes = []string{
       // Images
       "image/jpeg", "image/png", "image/gif", "image/webp", "image/svg+xml",
       // Documents
       "application/pdf", "text/plain", "text/markdown",
       "application/vnd.ms-excel", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
       "application/msword", "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
       // Archives
       "application/zip", "application/x-tar", "application/gzip",
       // Video (optional)
       "video/mp4", "video/webm", "video/quicktime",
   }
   ```

2. **File Extension Validation**
   - Verify extension matches MIME type
   - Block double extensions (.pdf.exe)
   - Normalize extensions to lowercase

3. **Magic Bytes Validation**
   - Read file signature (first 512 bytes)
   - Verify matches declared MIME type
   - Use `github.com/h2non/filetype` library

4. **File Size Validation**
   - Per-file size limits (configurable)
   - Total user quota limits
   - Request size limits (DDoS protection)

5. **Virus Scanning (ClamAV Integration)**
   ```go
   func (s *SecurityScanner) ScanFile(filepath string) error {
       conn, err := clamd.NewClamd("/var/run/clamav/clamd.sock")
       result, err := conn.ScanFile(filepath)

       if result.Status == clamd.RES_FOUND {
           return fmt.Errorf("virus detected: %s", result.Description)
       }
       return nil
   }
   ```

6. **Image Validation (for images)**
   - Decode image to verify format
   - Check dimensions (max width/height)
   - Prevent decompression bombs

---

#### 4.2.4 Metadata Manager

**Metadata Storage:**
- Filename (original + sanitized)
- MIME type
- File size (bytes)
- SHA-256 hash
- Uploader ID
- Upload timestamp
- Entity references (ticket ID, document ID, etc.)
- Version number
- Tags/labels
- Access permissions

**Search Capabilities:**
- List files by entity ID
- Search by filename
- Filter by MIME type
- Filter by uploader
- Filter by date range
- Sort by size, date, name

---

### 4.3 Storage Orchestration Layer

#### 4.3.1 Multi-Endpoint Manager

**Endpoint Types:**
1. **Primary**: Main storage (fastest, most reliable)
2. **Backup**: Failover storage (activated on primary failure)
3. **Mirror**: Replication storage (all writes go here)

**Configuration Example:**
```json
{
  "storage_endpoints": [
    {
      "id": "endpoint-1",
      "type": "primary",
      "adapter": "local",
      "path": "/var/helixtrack/attachments",
      "priority": 1,
      "max_size_gb": 1000
    },
    {
      "id": "endpoint-2",
      "type": "backup",
      "adapter": "s3",
      "bucket": "helixtrack-attachments-backup",
      "region": "us-east-1",
      "priority": 2,
      "max_size_gb": 5000
    },
    {
      "id": "endpoint-3",
      "type": "mirror",
      "adapter": "minio",
      "endpoint": "https://minio.internal:9000",
      "bucket": "attachments-mirror",
      "priority": 3
    }
  ]
}
```

---

#### 4.3.2 Failover Controller

**Failover Strategy:**
1. **Write Operation**:
   - Try primary endpoint
   - On failure: Try backup endpoint
   - Log failure and alert
   - Update endpoint health status

2. **Read Operation**:
   - Try primary endpoint
   - On failure: Try backup endpoint
   - On failure: Try mirror endpoint
   - Cache successful endpoint for next read

3. **Circuit Breaker**:
   - Track failure rate per endpoint
   - Open circuit after N failures in M seconds
   - Half-open state for testing recovery
   - Close circuit when endpoint recovers

**Implementation:**
```go
type CircuitBreaker struct {
    maxFailures   int
    timeout       time.Duration
    state         CircuitState // Closed, Open, HalfOpen
    failureCount  int
    lastFailure   time.Time
    mutex         sync.RWMutex
}

func (cb *CircuitBreaker) Call(fn func() error) error {
    cb.mutex.RLock()
    state := cb.state
    cb.mutex.RUnlock()

    if state == StateOpen {
        if time.Since(cb.lastFailure) > cb.timeout {
            cb.setState(StateHalfOpen)
        } else {
            return ErrCircuitOpen
        }
    }

    err := fn()

    if err != nil {
        cb.recordFailure()
        return err
    }

    cb.recordSuccess()
    return nil
}
```

---

#### 4.3.3 Replication Manager

**Replication Modes:**
1. **Synchronous**: Wait for all mirrors before success
2. **Asynchronous**: Return immediately, replicate in background
3. **Hybrid**: Primary + backup sync, mirrors async

**Replication Strategy:**
```go
type ReplicationManager struct {
    mode          ReplicationMode
    endpoints     []*StorageEndpoint
    replicaQueue  chan *ReplicationTask
    workerPool    *WorkerPool
}

func (rm *ReplicationManager) Replicate(file *PhysicalFile) error {
    switch rm.mode {
    case Synchronous:
        return rm.replicateSync(file)
    case Asynchronous:
        rm.replicateAsync(file)
        return nil
    case Hybrid:
        // Sync to primary + backup
        err := rm.replicateSync(file, PrimaryAndBackup)
        // Async to mirrors
        rm.replicateAsync(file, MirrorsOnly)
        return err
    }
}
```

---

#### 4.3.4 Health Monitor

**Health Checks:**
- Periodic ping to each endpoint (every 30s)
- Write test file to verify write capability
- Read test file to verify read capability
- Measure latency and throughput
- Check storage capacity

**Metrics Tracked:**
- Uptime percentage
- Average latency
- Error rate
- Available capacity
- Request count

**Alerting:**
- Email/Slack notification on endpoint failure
- Auto-disable endpoint after sustained failures
- Auto-enable endpoint after recovery

---

### 4.4 Storage Adapters

#### 4.4.1 Adapter Interface

```go
type StorageAdapter interface {
    // Store file and return storage path
    Store(hash string, data io.Reader, size int64) (string, error)

    // Retrieve file by storage path
    Retrieve(path string) (io.ReadCloser, error)

    // Delete file by storage path
    Delete(path string) error

    // Check if file exists
    Exists(path string) (bool, error)

    // Get file metadata
    GetMetadata(path string) (*FileMetadata, error)

    // Health check
    Ping() error

    // Get available capacity
    GetCapacity() (*CapacityInfo, error)
}
```

---

#### 4.4.2 Local Filesystem Adapter

**Features:**
- Hierarchical directory structure (hash-based sharding)
- Atomic writes (write to temp, then rename)
- Concurrent read support
- File permissions (chmod 0644)
- Directory permissions (chmod 0755)

**Directory Structure:**
```
/var/helixtrack/attachments/
├── ab/
│   ├── cd/
│   │   └── abcd1234...hash.bin  (actual file)
├── ef/
│   ├── 12/
│   │   └── ef123456...hash.bin
```

**Benefits:**
- Fast for small deployments
- No external dependencies
- Simple backup (rsync, tar)

**Limitations:**
- Not suitable for multi-server deployments
- Requires shared filesystem for clustering

---

#### 4.4.3 S3 Adapter

**Features:**
- AWS S3 SDK integration
- Multipart upload for large files
- Server-side encryption (SSE-S3, SSE-KMS)
- S3 lifecycle policies for cost optimization
- CloudFront CDN integration

**Configuration:**
```json
{
  "adapter": "s3",
  "bucket": "helixtrack-attachments",
  "region": "us-east-1",
  "access_key_id": "AKIA...",
  "secret_access_key": "***",
  "encryption": "SSE-S3",
  "storage_class": "STANDARD_IA"
}
```

**Benefits:**
- Unlimited scalability
- 99.999999999% durability
- Built-in replication
- Pay-as-you-go pricing

---

#### 4.4.4 MinIO Adapter

**Features:**
- S3-compatible API
- Self-hosted option
- Erasure coding for data protection
- Multi-tenancy support

**Use Cases:**
- On-premises deployments
- Private cloud
- Development/testing

---

## 5. Storage Architecture

### 5.1 Hash-Based Storage

**File Storage Path:**
```
{storage_endpoint}/{hash[0:2]}/{hash[2:4]}/{full_hash}.{ext}
```

**Example:**
```
File: architecture-diagram.png
SHA-256: abcd1234ef567890...
MIME: image/png

Storage Paths:
- Primary:  /var/attachments/ab/cd/abcd1234ef567890....png
- S3:       s3://bucket/ab/cd/abcd1234ef567890....png
- MinIO:    minio://bucket/ab/cd/abcd1234ef567890....png
```

**Benefits:**
- Deduplication: Same file stored once
- Fast lookup: Hash → file in O(1)
- Even distribution: First 4 chars = 65k directories

---

### 5.2 Reference Architecture

**Database Schema:**

```sql
-- Physical file (one per unique hash)
CREATE TABLE attachment_file (
    hash            TEXT    PRIMARY KEY,        -- SHA-256 hash
    size_bytes      BIGINT  NOT NULL,
    mime_type       TEXT    NOT NULL,
    extension       TEXT,                       -- Original extension
    ref_count       INTEGER NOT NULL DEFAULT 1, -- Reference counter
    storage_primary TEXT    NOT NULL,           -- Primary storage path
    storage_backup  TEXT,                       -- Backup storage path
    storage_mirrors TEXT[],                     -- Array of mirror paths
    created         BIGINT  NOT NULL,
    last_accessed   BIGINT  NOT NULL,
    deleted         BOOLEAN NOT NULL DEFAULT false
);

-- Logical reference (many per hash)
CREATE TABLE attachment_reference (
    id              TEXT    PRIMARY KEY,        -- UUID
    file_hash       TEXT    NOT NULL,           -- FK to attachment_file
    entity_type     TEXT    NOT NULL,           -- 'ticket', 'document', 'comment', etc.
    entity_id       TEXT    NOT NULL,           -- ID of entity
    filename        TEXT    NOT NULL,           -- User-provided filename
    description     TEXT,
    uploader_id     TEXT    NOT NULL,
    version         INTEGER NOT NULL DEFAULT 1,
    tags            TEXT[],                     -- Searchable tags
    created         BIGINT  NOT NULL,
    modified        BIGINT  NOT NULL,
    deleted         BOOLEAN NOT NULL DEFAULT false,

    FOREIGN KEY (file_hash) REFERENCES attachment_file(hash)
);

CREATE INDEX idx_attachment_ref_entity ON attachment_reference(entity_type, entity_id);
CREATE INDEX idx_attachment_ref_uploader ON attachment_reference(uploader_id);
CREATE INDEX idx_attachment_ref_hash ON attachment_reference(file_hash);
```

**Example Data:**

```sql
-- User uploads "logo.png" (hash: abc123) to ticket-1
-- User uploads "diagram.png" (hash: def456) to ticket-1
-- User uploads "logo.png" (same file, hash: abc123) to ticket-2

-- attachment_file table
hash    | size_bytes | mime_type  | ref_count | storage_primary
--------|------------|------------|-----------|------------------
abc123  | 52480      | image/png  | 2         | /ab/c1/abc123.png
def456  | 104960     | image/png  | 1         | /de/f4/def456.png

-- attachment_reference table
id     | file_hash | entity_type | entity_id | filename     | uploader_id
-------|-----------|-------------|-----------|--------------|-------------
ref-1  | abc123    | ticket      | ticket-1  | logo.png     | user-1
ref-2  | def456    | ticket      | ticket-1  | diagram.png  | user-1
ref-3  | abc123    | ticket      | ticket-2  | logo.png     | user-2
```

**Storage Savings:** 2 files stored instead of 3 (33% savings)

---

### 5.3 Cleanup Strategy

**Orphan File Cleanup:**
1. Periodic job (daily at 2 AM)
2. Find files with ref_count = 0 and deleted = true
3. Grace period: 30 days before physical deletion
4. Delete from all storage endpoints
5. Remove database record

**Dangling Reference Cleanup:**
1. Periodic job (weekly)
2. Find references where entity no longer exists
3. Soft delete reference (set deleted = true)
4. Decrement file ref_count
5. Trigger orphan cleanup if ref_count = 0

---

## 6. Security Architecture

### 6.1 Multi-Layer Security

```
┌─────────────────────────────────────────────────────────┐
│                    Security Layers                       │
├─────────────────────────────────────────────────────────┤
│                                                           │
│  Layer 1: Network Security                              │
│  ┌────────────────────────────────────────────────┐    │
│  │ - DDoS Protection (rate limiting)               │    │
│  │ - IP Whitelisting/Blacklisting                  │    │
│  │ - Connection Limits (per IP)                    │    │
│  │ - Request Size Limits                           │    │
│  └────────────────────────────────────────────────┘    │
│                         │                                │
│  Layer 2: Authentication & Authorization                │
│  ┌────────────────────────────────────────────────┐    │
│  │ - JWT Token Validation                          │    │
│  │ - Permission Checks (RBAC)                      │    │
│  │ - API Key Validation (for integrations)         │    │
│  │ - Request Signing (HMAC-SHA256)                 │    │
│  └────────────────────────────────────────────────┘    │
│                         │                                │
│  Layer 3: Input Validation                              │
│  ┌────────────────────────────────────────────────┐    │
│  │ - MIME Type Validation                          │    │
│  │ - File Extension Validation                     │    │
│  │ - File Size Validation                          │    │
│  │ - Path Sanitization                             │    │
│  │ - Filename Sanitization                         │    │
│  └────────────────────────────────────────────────┘    │
│                         │                                │
│  Layer 4: Content Validation                            │
│  ┌────────────────────────────────────────────────┐    │
│  │ - Magic Bytes Verification                      │    │
│  │ - Image Decompression Bomb Detection            │    │
│  │ - Virus Scanning (ClamAV)                       │    │
│  │ - Malware Signature Detection                   │    │
│  └────────────────────────────────────────────────┘    │
│                         │                                │
│  Layer 5: Storage Security                              │
│  ┌────────────────────────────────────────────────┐    │
│  │ - Encryption at Rest (AES-256)                  │    │
│  │ - Encryption in Transit (TLS 1.3)               │    │
│  │ - Secure File Permissions (0644/0755)           │    │
│  │ - Integrity Verification (SHA-256)              │    │
│  └────────────────────────────────────────────────┘    │
│                         │                                │
│  Layer 6: Access Control                                │
│  ┌────────────────────────────────────────────────┐    │
│  │ - Presigned URLs (time-limited)                 │    │
│  │ - Access Logging (audit trail)                  │    │
│  │ - Download Limits (per user/file)               │    │
│  │ - IP-Based Access Control                       │    │
│  └────────────────────────────────────────────────┘    │
│                                                           │
└─────────────────────────────────────────────────────────┘
```

---

### 6.2 DDoS Protection

**Rate Limiting:**
```go
type RateLimiter struct {
    // Per-IP limits
    ipLimits    map[string]*TokenBucket
    // Per-User limits
    userLimits  map[string]*TokenBucket
    // Global limits
    globalLimit *TokenBucket
    mutex       sync.RWMutex
}

type TokenBucket struct {
    capacity    int       // Max tokens
    tokens      int       // Current tokens
    refillRate  int       // Tokens per second
    lastRefill  time.Time
}

// Configuration
var RateLimitConfig = RateLimitConfig{
    PerIP: RateLimit{
        Requests: 100,      // requests
        Window:   60,       // seconds
    },
    PerUser: RateLimit{
        Requests: 1000,     // requests
        Window:   60,       // seconds
    },
    PerIPUpload: RateLimit{
        Requests: 10,       // uploads
        Window:   60,       // seconds
    },
    PerUserUpload: RateLimit{
        Requests: 100,      // uploads
        Window:   60,       // seconds
    },
    Global: RateLimit{
        Requests: 10000,    // requests
        Window:   60,       // seconds
    },
}
```

**Connection Limits:**
- Max 100 concurrent connections per IP
- Max 1000 concurrent connections globally
- Connection timeout: 30 seconds
- Read timeout: 60 seconds
- Write timeout: 300 seconds (for large uploads)

**Request Size Limits:**
- Max request body: 100 MB (configurable)
- Max header size: 8 KB
- Max URI length: 4 KB
- Max multipart form size: 100 MB

---

### 6.3 Penetration Protection

**SQL Injection Prevention:**
- Parameterized queries only (no string concatenation)
- Input validation and sanitization
- ORM with built-in escaping (sqlx)

**Path Traversal Prevention:**
```go
func sanitizePath(path string) (string, error) {
    // Remove null bytes
    path = strings.ReplaceAll(path, "\x00", "")

    // Resolve to absolute path
    absPath, err := filepath.Abs(path)
    if err != nil {
        return "", err
    }

    // Clean path (remove ../, ./, etc.)
    cleanPath := filepath.Clean(absPath)

    // Verify within allowed directory
    if !strings.HasPrefix(cleanPath, allowedBasePath) {
        return "", ErrPathTraversal
    }

    return cleanPath, nil
}
```

**Command Injection Prevention:**
- No shell command execution with user input
- Validate all external tool inputs (ClamAV, etc.)
- Use libraries instead of CLI tools when possible

**XSS Prevention:**
- Sanitize filenames in responses
- Proper Content-Type headers
- Content-Disposition: attachment for downloads

**CSRF Protection:**
- Require JWT token for all mutations
- Validate JWT signature
- Check token expiration

---

### 6.4 Virus Scanning

**ClamAV Integration:**
```go
type VirusScanner struct {
    clamd       *clamd.Clamd
    enabled     bool
    timeout     time.Duration
    maxFileSize int64
}

func (vs *VirusScanner) ScanFile(filepath string) (*ScanResult, error) {
    if !vs.enabled {
        return &ScanResult{Clean: true}, nil
    }

    // Skip large files (performance)
    if fileSize > vs.maxFileSize {
        return &ScanResult{Clean: true, Skipped: true}, nil
    }

    // Scan with timeout
    ctx, cancel := context.WithTimeout(context.Background(), vs.timeout)
    defer cancel()

    resultChan := make(chan *clamd.ScanResult)
    errChan := make(chan error)

    go func() {
        result, err := vs.clamd.ScanFile(filepath)
        if err != nil {
            errChan <- err
            return
        }
        resultChan <- result
    }()

    select {
    case result := <-resultChan:
        if result.Status == clamd.RES_FOUND {
            return &ScanResult{
                Clean:       false,
                ThreatName:  result.Description,
            }, nil
        }
        return &ScanResult{Clean: true}, nil
    case err := <-errChan:
        return nil, err
    case <-ctx.Done():
        return nil, ErrScanTimeout
    }
}
```

**Scan Strategy:**
- Scan on upload (before storage)
- Periodic re-scan of existing files (weekly)
- Signature updates daily
- Quarantine infected files
- Alert administrators

---

## 7. High Availability & Reliability

### 7.1 Concurrency Safety

**Lock-Free Architecture:**
- Database transactions for atomicity
- Optimistic locking for updates
- No global locks
- Per-entity locking when needed

**Deadlock Prevention:**
```go
// Always acquire locks in same order
// Use timeout contexts
// Prefer channel-based synchronization

type SafeCounter struct {
    mu    sync.RWMutex
    count int
}

func (c *SafeCounter) Increment() {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.count++
}

// Better: Use atomic operations
type AtomicCounter struct {
    count atomic.Int64
}

func (c *AtomicCounter) Increment() {
    c.count.Add(1)
}
```

**Race Condition Prevention:**
- `go test -race` for all tests
- Atomic operations for counters
- Channel-based communication
- Immutable data structures where possible

---

### 7.2 Error Handling

**Retry Strategy:**
```go
type RetryConfig struct {
    MaxAttempts int
    InitialDelay time.Duration
    MaxDelay     time.Duration
    Multiplier   float64
}

func RetryWithExponentialBackoff(fn func() error, config RetryConfig) error {
    delay := config.InitialDelay

    for attempt := 1; attempt <= config.MaxAttempts; attempt++ {
        err := fn()
        if err == nil {
            return nil
        }

        // Don't retry on permanent errors
        if IsPermanentError(err) {
            return err
        }

        if attempt < config.MaxAttempts {
            time.Sleep(delay)
            delay = time.Duration(float64(delay) * config.Multiplier)
            if delay > config.MaxDelay {
                delay = config.MaxDelay
            }
        }
    }

    return ErrMaxRetriesExceeded
}
```

**Circuit Breaker:** (See section 4.3.2)

**Graceful Degradation:**
- Continue serving reads even if writes fail
- Return cached data if database unavailable
- Queue operations for later retry
- Inform user of degraded service

---

### 7.3 Monitoring & Alerting

**Metrics (Prometheus):**
- Request count (by endpoint, status code)
- Request latency (P50, P95, P99)
- Upload/download throughput
- Storage endpoint health
- Error rate
- Active connections
- Queue depth
- Cache hit rate

**Logging (Structured):**
```go
log.Info("file uploaded",
    zap.String("file_hash", hash),
    zap.Int64("size_bytes", size),
    zap.String("mime_type", mimeType),
    zap.String("uploader_id", uploaderID),
    zap.Duration("duration", duration),
)
```

**Alerting:**
- Error rate > 5%: Warning
- Error rate > 10%: Critical
- Endpoint down: Critical
- Disk usage > 90%: Warning
- Disk usage > 95%: Critical
- Virus detected: Critical
- Unusual upload patterns: Warning

---

## 8. API Design

### 8.1 RESTful Endpoints

**Base URL:** `https://attachments.helixtrack.com/v1`

#### Upload File
```
POST /files
Content-Type: multipart/form-data
Authorization: Bearer {jwt}

Form Fields:
  - file: (binary file data)
  - entity_type: "ticket" | "document" | "comment" | "project"
  - entity_id: "ticket-123"
  - filename: "architecture.png" (optional, defaults to uploaded filename)
  - description: "System architecture diagram" (optional)
  - tags: ["architecture", "diagram"] (optional)

Response: 201 Created
{
  "reference_id": "ref-abc-123",
  "file_hash": "abcd1234ef567890...",
  "filename": "architecture.png",
  "size_bytes": 524288,
  "mime_type": "image/png",
  "url": "https://attachments.helixtrack.com/v1/files/ref-abc-123",
  "download_url": "https://attachments.helixtrack.com/v1/files/ref-abc-123/download",
  "created": 1729353600,
  "deduplicated": false
}
```

#### Download File
```
GET /files/{reference_id}/download
Authorization: Bearer {jwt}

Response: 200 OK
Content-Type: image/png
Content-Disposition: attachment; filename="architecture.png"
Content-Length: 524288

(binary file data)
```

#### Get File Metadata
```
GET /files/{reference_id}
Authorization: Bearer {jwt}

Response: 200 OK
{
  "reference_id": "ref-abc-123",
  "file_hash": "abcd1234...",
  "filename": "architecture.png",
  "size_bytes": 524288,
  "mime_type": "image/png",
  "uploader_id": "user-789",
  "entity_type": "ticket",
  "entity_id": "ticket-123",
  "description": "System architecture diagram",
  "tags": ["architecture", "diagram"],
  "version": 1,
  "created": 1729353600,
  "modified": 1729353600
}
```

#### List Files for Entity
```
GET /entities/{entity_type}/{entity_id}/files
Authorization: Bearer {jwt}
Query Params:
  - limit: 50 (default)
  - offset: 0 (default)
  - sort: "created" | "filename" | "size"
  - order: "asc" | "desc" (default)

Response: 200 OK
{
  "files": [
    { /* file metadata */ },
    { /* file metadata */ }
  ],
  "total": 42,
  "limit": 50,
  "offset": 0
}
```

#### Delete File
```
DELETE /files/{reference_id}
Authorization: Bearer {jwt}

Response: 204 No Content
```

#### Generate Presigned URL
```
POST /files/{reference_id}/presigned-url
Authorization: Bearer {jwt}
Body:
{
  "expires_in": 3600,  // seconds
  "download": true     // force download vs inline
}

Response: 200 OK
{
  "url": "https://attachments.helixtrack.com/v1/files/ref-abc-123/download?token=xyz&expires=1729357200",
  "expires_at": 1729357200
}
```

---

### 8.2 S3-Compatible API (Optional)

**For advanced integrations:**

```
PUT /buckets/{entity_type}/{entity_id}/{filename}
Authorization: AWS4-HMAC-SHA256 ...

Response: 200 OK

GET /buckets/{entity_type}/{entity_id}/{filename}
Authorization: AWS4-HMAC-SHA256 ...

Response: 200 OK
(file content)
```

---

## 9. Database Schema

### 9.1 Complete Schema

```sql
-- ============================================================
-- ATTACHMENTS SERVICE DATABASE SCHEMA V1
-- ============================================================

-- Physical files (deduplicated by hash)
CREATE TABLE attachment_file (
    hash                TEXT    PRIMARY KEY,
    size_bytes          BIGINT  NOT NULL CHECK (size_bytes >= 0),
    mime_type           TEXT    NOT NULL,
    extension           TEXT,
    ref_count           INTEGER NOT NULL DEFAULT 1 CHECK (ref_count >= 0),
    storage_primary     TEXT    NOT NULL,
    storage_backup      TEXT,
    storage_mirrors     TEXT[], -- Array of mirror paths
    virus_scan_status   TEXT    DEFAULT 'pending', -- 'pending', 'clean', 'infected', 'failed'
    virus_scan_date     BIGINT,
    virus_scan_result   TEXT,
    created             BIGINT  NOT NULL,
    last_accessed       BIGINT  NOT NULL,
    deleted             BOOLEAN NOT NULL DEFAULT false
);

CREATE INDEX idx_attachment_file_ref_count ON attachment_file(ref_count);
CREATE INDEX idx_attachment_file_mime ON attachment_file(mime_type);
CREATE INDEX idx_attachment_file_created ON attachment_file(created DESC);
CREATE INDEX idx_attachment_file_deleted ON attachment_file(deleted) WHERE deleted = false;

-- Logical references (many-to-many: entities <-> files)
CREATE TABLE attachment_reference (
    id              TEXT    PRIMARY KEY,
    file_hash       TEXT    NOT NULL,
    entity_type     TEXT    NOT NULL, -- 'ticket', 'document', 'comment', 'project', etc.
    entity_id       TEXT    NOT NULL,
    filename        TEXT    NOT NULL,
    description     TEXT,
    uploader_id     TEXT    NOT NULL,
    version         INTEGER NOT NULL DEFAULT 1,
    tags            TEXT[], -- Searchable tags
    created         BIGINT  NOT NULL,
    modified        BIGINT  NOT NULL,
    deleted         BOOLEAN NOT NULL DEFAULT false,

    FOREIGN KEY (file_hash) REFERENCES attachment_file(hash) ON DELETE CASCADE
);

CREATE INDEX idx_attachment_ref_entity ON attachment_reference(entity_type, entity_id, deleted);
CREATE INDEX idx_attachment_ref_uploader ON attachment_reference(uploader_id);
CREATE INDEX idx_attachment_ref_hash ON attachment_reference(file_hash);
CREATE INDEX idx_attachment_ref_created ON attachment_reference(created DESC);
CREATE INDEX idx_attachment_ref_tags ON attachment_reference USING GIN(tags);

-- Storage endpoints configuration
CREATE TABLE storage_endpoint (
    id              TEXT    PRIMARY KEY,
    name            TEXT    NOT NULL,
    type            TEXT    NOT NULL, -- 'local', 's3', 'minio', 'custom'
    role            TEXT    NOT NULL, -- 'primary', 'backup', 'mirror'
    adapter_config  JSONB   NOT NULL, -- Adapter-specific configuration
    priority        INTEGER NOT NULL DEFAULT 1,
    enabled         BOOLEAN NOT NULL DEFAULT true,
    max_size_bytes  BIGINT,
    current_size    BIGINT  NOT NULL DEFAULT 0,
    created         BIGINT  NOT NULL,
    modified        BIGINT  NOT NULL
);

CREATE INDEX idx_storage_endpoint_role ON storage_endpoint(role, enabled);
CREATE INDEX idx_storage_endpoint_priority ON storage_endpoint(priority);

-- Storage endpoint health monitoring
CREATE TABLE storage_health (
    endpoint_id     TEXT    NOT NULL,
    check_time      BIGINT  NOT NULL,
    status          TEXT    NOT NULL, -- 'healthy', 'degraded', 'unhealthy'
    latency_ms      INTEGER,
    error_message   TEXT,
    available_bytes BIGINT,

    PRIMARY KEY (endpoint_id, check_time),
    FOREIGN KEY (endpoint_id) REFERENCES storage_endpoint(id) ON DELETE CASCADE
);

CREATE INDEX idx_storage_health_time ON storage_health(check_time DESC);
CREATE INDEX idx_storage_health_status ON storage_health(endpoint_id, status);

-- Upload quotas (per user)
CREATE TABLE upload_quota (
    user_id         TEXT    PRIMARY KEY,
    max_bytes       BIGINT  NOT NULL DEFAULT 10737418240, -- 10 GB default
    used_bytes      BIGINT  NOT NULL DEFAULT 0,
    max_files       INTEGER NOT NULL DEFAULT 10000,
    used_files      INTEGER NOT NULL DEFAULT 0,
    created         BIGINT  NOT NULL,
    modified        BIGINT  NOT NULL
);

-- Access logs (audit trail)
CREATE TABLE access_log (
    id              TEXT    PRIMARY KEY,
    reference_id    TEXT,
    file_hash       TEXT,
    user_id         TEXT,
    ip_address      TEXT,
    action          TEXT    NOT NULL, -- 'upload', 'download', 'delete'
    status_code     INTEGER,
    error_message   TEXT,
    user_agent      TEXT,
    timestamp       BIGINT  NOT NULL
);

CREATE INDEX idx_access_log_timestamp ON access_log(timestamp DESC);
CREATE INDEX idx_access_log_user ON access_log(user_id, timestamp DESC);
CREATE INDEX idx_access_log_action ON access_log(action, timestamp DESC);

-- Presigned URLs (temporary access tokens)
CREATE TABLE presigned_url (
    token           TEXT    PRIMARY KEY,
    reference_id    TEXT    NOT NULL,
    user_id         TEXT,
    ip_address      TEXT,
    expires_at      BIGINT  NOT NULL,
    max_downloads   INTEGER DEFAULT 1,
    download_count  INTEGER NOT NULL DEFAULT 0,
    created         BIGINT  NOT NULL,

    FOREIGN KEY (reference_id) REFERENCES attachment_reference(id) ON DELETE CASCADE
);

CREATE INDEX idx_presigned_expires ON presigned_url(expires_at);
CREATE INDEX idx_presigned_ref ON presigned_url(reference_id);

-- Cleanup jobs tracking
CREATE TABLE cleanup_job (
    id              TEXT    PRIMARY KEY,
    job_type        TEXT    NOT NULL, -- 'orphan_files', 'dangling_refs', 'expired_presigned'
    started         BIGINT  NOT NULL,
    completed       BIGINT,
    status          TEXT    NOT NULL, -- 'running', 'completed', 'failed'
    items_processed INTEGER NOT NULL DEFAULT 0,
    items_deleted   INTEGER NOT NULL DEFAULT 0,
    error_message   TEXT
);

CREATE INDEX idx_cleanup_job_started ON cleanup_job(started DESC);
```

---

### 9.2 Schema Migrations

**Migration Strategy:**
1. Create new tables in Attachments Service database
2. Migrate data from Core's `asset` and `document_attachment` tables
3. Update Core to use Attachments Service API
4. Deprecate old tables (grace period)
5. Remove old tables after migration complete

**Migration SQL:**
```sql
-- Migrate V1 assets to new schema
INSERT INTO attachment_reference (
    id, file_hash, entity_type, entity_id, filename,
    uploader_id, created, modified, deleted
)
SELECT
    'migrated-' || a.id,
    sha256(a.url), -- Generate hash from URL
    'unknown',     -- Entity type unknown in V1
    '',            -- Entity ID unknown in V1
    a.url,         -- Use URL as filename
    'system',      -- Unknown uploader
    a.created,
    a.modified,
    a.deleted
FROM asset a;

-- Migrate Documents V2 attachments
INSERT INTO attachment_file (
    hash, size_bytes, mime_type, extension, ref_count,
    storage_primary, created, last_accessed
)
SELECT
    checksum,
    size_bytes,
    mime_type,
    substring(filename from '[^.]+$'),
    1, -- Will be updated by triggers
    storage_path,
    created,
    modified
FROM document_attachment
GROUP BY checksum;

INSERT INTO attachment_reference (
    id, file_hash, entity_type, entity_id, filename,
    description, uploader_id, version, created, modified, deleted
)
SELECT
    id,
    checksum,
    'document',
    document_id,
    original_filename,
    description,
    uploader_id,
    version,
    created,
    modified,
    deleted
FROM document_attachment;
```

---

## 10. Service Discovery & Configuration

### 10.1 Service Registry

**Discovery Mechanism:**
```go
type ServiceRegistry struct {
    consul    *consul.Client
    serviceName string
    serviceID   string
    port        int
    health      *HealthChecker
}

func (sr *ServiceRegistry) Register() error {
    // Find available port
    port, err := sr.findAvailablePort(8090, 8100)

    registration := &consul.AgentServiceRegistration{
        ID:      sr.serviceID,
        Name:    sr.serviceName,
        Port:    port,
        Address: getLocalIP(),
        Tags:    []string{"attachments", "v1"},
        Check: &consul.AgentServiceCheck{
            HTTP:     fmt.Sprintf("http://localhost:%d/health", port),
            Interval: "10s",
            Timeout:  "5s",
        },
    }

    return sr.consul.Agent().ServiceRegister(registration)
}

func (sr *ServiceRegistry) findAvailablePort(start, end int) (int, error) {
    for port := start; port <= end; port++ {
        listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
        if err == nil {
            listener.Close()
            return port, nil
        }
    }
    return 0, ErrNoAvailablePort
}
```

**Service Configuration:**
```json
{
  "service": {
    "name": "attachments-service",
    "port": 8090,
    "port_range": [8090, 8100],
    "auto_port_selection": true,
    "discovery": {
      "enabled": true,
      "provider": "consul",
      "consul_address": "localhost:8500"
    }
  },
  "database": {
    "driver": "postgres",
    "host": "localhost",
    "port": 5432,
    "database": "helixtrack_attachments",
    "user": "helixtrack",
    "password": "***",
    "max_connections": 50,
    "idle_connections": 10,
    "connection_timeout": 30
  },
  "storage": {
    "endpoints": [
      {
        "id": "local-primary",
        "type": "local",
        "role": "primary",
        "path": "/var/helixtrack/attachments",
        "priority": 1
      }
    ],
    "replication_mode": "hybrid",
    "cleanup": {
      "orphan_retention_days": 30,
      "job_schedule": "0 2 * * *"
    }
  },
  "security": {
    "jwt_secret": "***",
    "jwt_issuer": "helixtrack-auth",
    "allowed_mime_types": [
      "image/jpeg", "image/png", "image/gif",
      "application/pdf", "text/plain"
    ],
    "max_file_size_mb": 100,
    "virus_scanning": {
      "enabled": true,
      "clamd_socket": "/var/run/clamav/clamd.sock",
      "max_scan_size_mb": 100
    },
    "rate_limiting": {
      "per_ip_requests_per_minute": 100,
      "per_user_uploads_per_minute": 10
    }
  },
  "logging": {
    "level": "info",
    "format": "json",
    "output": "stdout"
  },
  "metrics": {
    "enabled": true,
    "prometheus_port": 9090
  }
}
```

---

### 10.2 Health Checks

**Health Check Endpoint:**
```
GET /health

Response: 200 OK
{
  "status": "healthy",
  "version": "1.0.0",
  "uptime_seconds": 86400,
  "checks": {
    "database": {
      "status": "healthy",
      "latency_ms": 5
    },
    "storage_primary": {
      "status": "healthy",
      "latency_ms": 2
    },
    "storage_backup": {
      "status": "healthy",
      "latency_ms": 15
    },
    "virus_scanner": {
      "status": "healthy"
    }
  }
}
```

---

## 11. Performance & Scalability

### 11.1 Performance Targets

| Metric | Target | Notes |
|--------|--------|-------|
| Upload Latency | <1s for 10MB | Small files <100ms |
| Download Latency | <100ms | Excluding file transfer time |
| Metadata Query | <50ms | 95th percentile |
| Concurrent Uploads | 100+ | Per instance |
| Concurrent Downloads | 1000+ | Per instance |
| Throughput | 1 GB/s | Per instance |
| Database Connections | 50 | Pooled |

---

### 11.2 Caching Strategy

**Multi-Layer Cache:**
```
┌────────────────────────────────────────┐
│         Application Layer              │
├────────────────────────────────────────┤
│  In-Memory Cache (LRU, 1GB)            │
│  - Metadata (file info, references)    │
│  - Frequently accessed small files     │
│  - Presigned URL tokens                │
├────────────────────────────────────────┤
│  Redis Cache (100GB)                   │
│  - Shared across instances             │
│  - Metadata cache                      │
│  - Rate limiting counters              │
├────────────────────────────────────────┤
│  CDN Cache (CloudFront, etc.)          │
│  - Public files                        │
│  - Static assets                       │
│  - Large files (1 hour TTL)            │
└────────────────────────────────────────┘
```

**Cache Invalidation:**
- Write-through for uploads (cache + DB)
- Invalidate on update/delete
- TTL-based expiration (1 hour default)
- LRU eviction for memory cache

---

### 11.3 Horizontal Scaling

**Stateless Design:**
- No session state in application
- All state in database or Redis
- Any instance can serve any request

**Load Balancing:**
```
                 ┌──────────────┐
                 │ Load Balancer│
                 │  (Nginx)     │
                 └──────┬───────┘
                        │
        ┌───────────────┼───────────────┐
        │               │               │
   ┌────▼────┐     ┌────▼────┐     ┌────▼────┐
   │Instance │     │Instance │     │Instance │
   │    1    │     │    2    │     │    3    │
   └────┬────┘     └────┬────┘     └────┬────┘
        │               │               │
        └───────────────┼───────────────┘
                        │
                 ┌──────▼───────┐
                 │  PostgreSQL  │
                 │  (Primary +  │
                 │   Replicas)  │
                 └──────────────┘
```

**Scaling Strategy:**
- Add instances behind load balancer
- Database read replicas for queries
- Sticky sessions not required
- Automatic instance discovery via Consul

---

## 12. Testing Strategy

### 12.1 Unit Tests (100% Coverage Target)

**Test Structure:**
```
Core/Attachments-Service/
├── internal/
│   ├── models/
│   │   ├── attachment_file.go
│   │   ├── attachment_file_test.go      (100% coverage)
│   │   ├── attachment_reference.go
│   │   └── attachment_reference_test.go (100% coverage)
│   ├── handlers/
│   │   ├── upload_handler.go
│   │   ├── upload_handler_test.go       (100% coverage)
│   │   ├── download_handler.go
│   │   └── download_handler_test.go     (100% coverage)
│   ├── storage/
│   │   ├── deduplication.go
│   │   ├── deduplication_test.go        (100% coverage)
│   │   ├── reference_counter.go
│   │   └── reference_counter_test.go    (100% coverage)
│   └── security/
│       ├── scanner.go
│       ├── scanner_test.go              (100% coverage)
│       ├── rate_limiter.go
│       └── rate_limiter_test.go         (100% coverage)
```

**Test Coverage Requirements:**
- All public functions: 100%
- All error paths: 100%
- All edge cases: 100%
- Race conditions: `go test -race`

---

### 12.2 Integration Tests

**Test Scenarios:**
1. **Service Communication**
   - Core → Attachments service integration
   - Attachments → Storage endpoint integration
   - Attachments → ClamAV integration
   - Attachments → Database integration

2. **Multi-Endpoint Storage**
   - Primary + backup failover
   - Mirror replication
   - Endpoint health monitoring

3. **Deduplication**
   - Upload same file twice
   - Verify single storage
   - Verify ref count = 2

4. **Security**
   - MIME type validation
   - Virus scanning workflow
   - Rate limiting enforcement
   - JWT authentication

**Test Database:**
- Separate test database
- Reset between tests
- Fixtures for common data

---

### 12.3 AI QA Automation

**AI QA Test Framework:**
```bash
Core/Attachments-Service/tests/ai-qa/
├── ai-qa-runner.sh                 # Main test runner
├── test-scenarios.yaml             # AI test scenarios
├── test-generator.py               # AI test generator
└── results/
    └── ai-qa-report-2025-10-19.json
```

**AI Test Scenarios:**
```yaml
scenarios:
  - name: "Comprehensive Upload Testing"
    description: "AI generates diverse file uploads"
    ai_strategy: "generate_random_files"
    file_types:
      - images: 100 (various sizes, formats)
      - documents: 100 (PDF, Word, Excel)
      - archives: 50 (ZIP, TAR, GZ)
    edge_cases:
      - zero_byte_files
      - max_size_files
      - invalid_mime_types
      - malformed_headers
    expected_behavior: "validate_and_store_or_reject"

  - name: "Deduplication Verification"
    description: "AI verifies deduplication logic"
    ai_strategy: "upload_duplicate_files"
    test_steps:
      - upload_file_1
      - verify_storage_created
      - upload_same_file
      - verify_no_new_storage
      - verify_ref_count_incremented

  - name: "Concurrency Stress Test"
    description: "AI simulates concurrent uploads"
    ai_strategy: "concurrent_operations"
    concurrent_uploads: 100
    concurrent_downloads: 1000
    duration_seconds: 300
    metrics:
      - throughput
      - latency_p95
      - error_rate
      - deadlock_detection

  - name: "Security Penetration Testing"
    description: "AI attempts security exploits"
    ai_strategy: "adversarial_testing"
    attack_vectors:
      - path_traversal
      - sql_injection
      - xss_via_filename
      - oversized_uploads
      - malicious_mime_types
      - executable_uploads
    expected_behavior: "reject_all_attacks"

  - name: "Failover and Recovery"
    description: "AI tests high availability"
    ai_strategy: "chaos_engineering"
    failure_scenarios:
      - primary_storage_failure
      - database_connection_loss
      - clamav_service_down
      - rate_limit_exceeded
    expected_behavior: "graceful_degradation"
```

**AI QA Runner:**
```python
# test-generator.py
import anthropic
import random
import os

class AIQATestGenerator:
    def __init__(self):
        self.client = anthropic.Anthropic()

    def generate_test_files(self, file_type, count):
        """AI generates diverse test files"""
        prompt = f"""
        Generate {count} test files of type {file_type}.
        Include edge cases: empty, max size, corrupted, etc.
        Return file specifications (name, size, content pattern).
        """

        # AI generates diverse test cases
        response = self.client.messages.create(
            model="claude-3-5-sonnet-20241022",
            max_tokens=4096,
            messages=[{"role": "user", "content": prompt}]
        )

        return self.parse_test_specs(response.content)

    def generate_attack_vectors(self):
        """AI generates security attack scenarios"""
        prompt = """
        Generate creative security attack vectors for file upload API:
        - Path traversal variations
        - SQL injection in filenames
        - XSS in metadata
        - Novel attack patterns

        Return attack payloads and expected defenses.
        """

        # AI generates adversarial tests
        # ...
```

---

### 12.4 E2E Tests

**E2E Test Workflows:**
```javascript
// Core/Attachments-Service/tests/e2e/upload-workflow.spec.js

describe('Complete Upload Workflow', () => {
  it('should upload file, attach to ticket, download, and delete', async () => {
    // 1. Authenticate
    const jwt = await authenticate('user@example.com', 'password');

    // 2. Create ticket
    const ticket = await createTicket(jwt, {
      title: 'Test Ticket',
      description: 'For attachment testing'
    });

    // 3. Upload file
    const file = readFileSync('test-files/diagram.png');
    const attachment = await uploadFile(jwt, {
      file: file,
      entity_type: 'ticket',
      entity_id: ticket.id,
      filename: 'architecture-diagram.png'
    });

    expect(attachment.reference_id).toBeDefined();
    expect(attachment.file_hash).toBeDefined();

    // 4. Verify file attached to ticket
    const ticketFiles = await listTicketFiles(jwt, ticket.id);
    expect(ticketFiles.length).toBe(1);
    expect(ticketFiles[0].reference_id).toBe(attachment.reference_id);

    // 5. Download file
    const downloaded = await downloadFile(jwt, attachment.reference_id);
    expect(downloaded.length).toBe(file.length);
    expect(sha256(downloaded)).toBe(attachment.file_hash);

    // 6. Delete attachment
    await deleteFile(jwt, attachment.reference_id);

    // 7. Verify file removed from ticket
    const filesAfterDelete = await listTicketFiles(jwt, ticket.id);
    expect(filesAfterDelete.length).toBe(0);
  });

  it('should deduplicate identical uploads', async () => {
    const jwt = await authenticate('user@example.com', 'password');
    const file = readFileSync('test-files/logo.png');

    // Upload same file to two different tickets
    const ticket1 = await createTicket(jwt, {title: 'Ticket 1'});
    const ticket2 = await createTicket(jwt, {title: 'Ticket 2'});

    const attachment1 = await uploadFile(jwt, {
      file: file,
      entity_type: 'ticket',
      entity_id: ticket1.id
    });

    const attachment2 = await uploadFile(jwt, {
      file: file,
      entity_type: 'ticket',
      entity_id: ticket2.id
    });

    // Verify same hash (deduplicated)
    expect(attachment1.file_hash).toBe(attachment2.file_hash);
    expect(attachment1.deduplicated).toBe(false); // First upload
    expect(attachment2.deduplicated).toBe(true);  // Second upload

    // Verify both references exist
    const files1 = await listTicketFiles(jwt, ticket1.id);
    const files2 = await listTicketFiles(jwt, ticket2.id);

    expect(files1[0].file_hash).toBe(files2[0].file_hash);
  });
});
```

**Test Coverage:**
- All API endpoints
- All user workflows
- Error scenarios
- Performance under load
- Browser compatibility (Web Client)
- Mobile app integration (Android/iOS)

---

## 13. Deployment Architecture

### 13.1 Production Deployment

```yaml
# docker-compose.yml
version: '3.8'

services:
  attachments-service:
    build: ./Core/Attachments-Service
    image: helixtrack/attachments-service:1.0.0
    ports:
      - "8090-8100:8090" # Auto port selection
    environment:
      - DATABASE_URL=postgres://helixtrack:***@postgres:5432/attachments
      - STORAGE_PRIMARY=/var/attachments
      - CLAMAV_SOCKET=/var/run/clamav/clamd.sock
      - JWT_SECRET=${JWT_SECRET}
    volumes:
      - attachments-data:/var/attachments
      - clamav-socket:/var/run/clamav
    depends_on:
      - postgres
      - clamav
    deploy:
      replicas: 3
      resources:
        limits:
          cpus: '2'
          memory: 4G
        reservations:
          cpus: '1'
          memory: 2G

  postgres:
    image: postgres:16
    environment:
      - POSTGRES_DB=attachments
      - POSTGRES_USER=helixtrack
      - POSTGRES_PASSWORD=${DB_PASSWORD}
    volumes:
      - postgres-data:/var/lib/postgresql/data
    deploy:
      resources:
        limits:
          cpus: '2'
          memory: 4G

  clamav:
    image: clamav/clamav:latest
    volumes:
      - clamav-socket:/var/run/clamav
      - clamav-data:/var/lib/clamav
    deploy:
      resources:
        limits:
          cpus: '2'
          memory: 2G

  nginx:
    image: nginx:alpine
    ports:
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf
      - ./ssl:/etc/nginx/ssl
    depends_on:
      - attachments-service

volumes:
  attachments-data:
  postgres-data:
  clamav-socket:
  clamav-data:
```

---

### 13.2 Kubernetes Deployment

```yaml
# k8s/attachments-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: attachments-service
  namespace: helixtrack
spec:
  replicas: 3
  selector:
    matchLabels:
      app: attachments-service
  template:
    metadata:
      labels:
        app: attachments-service
    spec:
      containers:
      - name: attachments
        image: helixtrack/attachments-service:1.0.0
        ports:
        - containerPort: 8090
        env:
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: attachments-secrets
              key: database-url
        - name: JWT_SECRET
          valueFrom:
            secretKeyRef:
              name: attachments-secrets
              key: jwt-secret
        resources:
          requests:
            memory: "2Gi"
            cpu: "1"
          limits:
            memory: "4Gi"
            cpu: "2"
        volumeMounts:
        - name: attachments-storage
          mountPath: /var/attachments
        livenessProbe:
          httpGet:
            path: /health
            port: 8090
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 8090
          initialDelaySeconds: 10
          periodSeconds: 5
      volumes:
      - name: attachments-storage
        persistentVolumeClaim:
          claimName: attachments-pvc
---
apiVersion: v1
kind: Service
metadata:
  name: attachments-service
  namespace: helixtrack
spec:
  selector:
    app: attachments-service
  ports:
  - protocol: TCP
    port: 80
    targetPort: 8090
  type: LoadBalancer
```

---

## 14. Migration Plan

### 14.1 Phase 1: Development (Week 1-2)
- [ ] Create Attachments Service project structure
- [ ] Implement database schema
- [ ] Implement core models and validation
- [ ] Implement storage adapters (local, S3)
- [ ] Implement deduplication engine
- [ ] Implement reference counter
- [ ] Implement security scanner
- [ ] Write unit tests (100% coverage)

### 14.2 Phase 2: Integration (Week 3)
- [ ] Implement RESTful API handlers
- [ ] Implement service discovery
- [ ] Implement health checks
- [ ] Integrate with Core backend
- [ ] Implement ClamAV integration
- [ ] Write integration tests

### 14.3 Phase 3: Advanced Features (Week 4)
- [ ] Implement multi-endpoint storage
- [ ] Implement failover controller
- [ ] Implement replication manager
- [ ] Implement circuit breaker
- [ ] Implement rate limiting
- [ ] Implement presigned URLs

### 14.4 Phase 4: Testing (Week 5)
- [ ] Write E2E tests
- [ ] Create AI QA automation
- [ ] Execute full test suite
- [ ] Performance testing
- [ ] Security penetration testing
- [ ] Load testing

### 14.5 Phase 5: Documentation (Week 6)
- [ ] API documentation
- [ ] Deployment guides
- [ ] Migration guides
- [ ] Update Core documentation
- [ ] Update website
- [ ] Create video tutorials

### 14.6 Phase 6: Deployment (Week 7-8)
- [ ] Deploy to staging environment
- [ ] Migrate existing attachments
- [ ] Production deployment
- [ ] Monitoring setup
- [ ] Alerting configuration
- [ ] Go-live

---

## 15. Success Metrics

### 15.1 Performance Metrics
- [ ] Upload latency <1s for 10MB files
- [ ] Download latency <100ms (metadata)
- [ ] 100+ concurrent uploads supported
- [ ] 1000+ concurrent downloads supported
- [ ] 99.9% uptime achieved

### 15.2 Quality Metrics
- [ ] 100% unit test coverage
- [ ] 100% test pass rate
- [ ] 0 critical security vulnerabilities
- [ ] 0 data loss incidents
- [ ] 0 deadlocks or race conditions

### 15.3 Business Metrics
- [ ] 50% storage savings via deduplication
- [ ] 10x faster file uploads (vs current)
- [ ] 0 virus infections uploaded
- [ ] 100% audit trail completeness

---

## Conclusion

The HelixTrack Attachments Service architecture provides a robust, scalable, secure, and high-performance solution for file management. With hash-based deduplication, multi-endpoint storage, comprehensive security, and 100% test coverage, it meets enterprise-grade requirements while maintaining simplicity and ease of deployment.

**Next Steps:**
1. Review and approve architecture
2. Begin Phase 1 development
3. Set up CI/CD pipeline
4. Establish monitoring infrastructure

---

**Document Version:** 1.0.0
**Last Updated:** 2025-10-19
**Status:** Design Complete, Ready for Implementation
