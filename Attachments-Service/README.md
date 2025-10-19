# HelixTrack Attachments Service

**Version:** 1.0.0
**Status:** In Development
**License:** MIT

## Overview

The HelixTrack Attachments Service is a decoupled, S3-compatible microservice designed to provide enterprise-grade file storage, retrieval, and management capabilities for the HelixTrack ecosystem.

### Key Features

- ✅ **Hash-Based Deduplication** - Store identical files once (SHA-256)
- ✅ **Reference Counting** - Automatic tracking of file usage across entities
- ✅ **Multi-Endpoint Storage** - Primary + backup + mirror storage support
- ✅ **Service Discovery** - Automatic registration with Consul/etcd
- ✅ **Auto Port Selection** - Automatic port binding if configured port is unavailable
- ✅ **Military-Grade Security** - Multi-layer security with virus scanning
- ✅ **DDoS Protection** - Rate limiting, connection limits, request size limits
- ✅ **Zero Deadlocks** - Lock-free architecture with atomic operations
- ✅ **S3-Compatible API** - RESTful API similar to AWS S3
- ✅ **100% Test Coverage Target** - Comprehensive testing strategy

## Architecture

The Attachments Service follows a layered microservices architecture:

```
┌─────────────────────────────────────────────┐
│           API Gateway Layer                  │
│  - JWT Authentication                        │
│  - Rate Limiting                             │
│  - Request Validation                        │
├─────────────────────────────────────────────┤
│         Business Logic Layer                 │
│  - Deduplication Engine                      │
│  - Reference Counter                         │
│  - Security Scanner                          │
│  - Metadata Manager                          │
├─────────────────────────────────────────────┤
│      Storage Orchestration Layer             │
│  - Multi-Endpoint Manager                    │
│  - Failover Controller                       │
│  - Replication Manager                       │
│  - Health Monitor                            │
├─────────────────────────────────────────────┤
│          Storage Adapters                    │
│  - Local Filesystem                          │
│  - AWS S3                                    │
│  - MinIO                                     │
│  - Custom Adapters                           │
└─────────────────────────────────────────────┘
```

See [docs/ATTACHMENTS_SERVICE_ARCHITECTURE.md](docs/ATTACHMENTS_SERVICE_ARCHITECTURE.md) for detailed architecture documentation.

## Quick Start

### Prerequisites

- Go 1.22+
- PostgreSQL 12+ or SQLite 3+
- (Optional) ClamAV for virus scanning
- (Optional) Consul for service discovery

### Installation

```bash
# Clone repository
cd Core/Attachments-Service

# Install dependencies
go mod download

# Initialize database
psql -U helixtrack -d helixtrack_attachments -f Database/DDL/001_initial_schema.sql

# Copy and configure
cp configs/default.json configs/production.json
# Edit configs/production.json with your settings

# Build
go build -o attachments-service cmd/main.go

# Run
./attachments-service --config=configs/production.json
```

### Development

```bash
# Run with default config (SQLite)
go run cmd/main.go --config=configs/default.json

# Run tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with race detection
go test -race ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Configuration

The service is configured via JSON files. See [configs/default.json](configs/default.json) for a complete example.

### Key Configuration Sections

#### Service Configuration

```json
{
  "service": {
    "port": 8090,
    "port_range": [8090, 8100],
    "environment": "production",
    "discovery": {
      "enabled": true,
      "provider": "consul",
      "consul_address": "localhost:8500"
    }
  }
}
```

#### Database Configuration

```json
{
  "database": {
    "driver": "postgres",
    "host": "localhost",
    "port": 5432,
    "database": "helixtrack_attachments",
    "user": "helixtrack",
    "password": "your-secure-password",
    "max_connections": 50
  }
}
```

#### Storage Configuration

```json
{
  "storage": {
    "endpoints": [
      {
        "id": "local-primary",
        "type": "local",
        "role": "primary",
        "adapter_config": {
          "path": "/var/helixtrack/attachments"
        }
      },
      {
        "id": "s3-backup",
        "type": "s3",
        "role": "backup",
        "adapter_config": {
          "bucket": "helixtrack-attachments-backup",
          "region": "us-east-1"
        }
      }
    ]
  }
}
```

#### Security Configuration

```json
{
  "security": {
    "jwt_secret": "your-secret-key",
    "allowed_mime_types": [
      "image/jpeg", "image/png", "application/pdf"
    ],
    "max_file_size_mb": 100,
    "virus_scanning": {
      "enabled": true,
      "clamd_socket": "/var/run/clamav/clamd.sock"
    },
    "rate_limiting": {
      "per_ip_requests_per_minute": 100,
      "per_user_uploads_per_minute": 10
    }
  }
}
```

## API Reference

### Upload File

```bash
POST /v1/files
Content-Type: multipart/form-data
Authorization: Bearer {jwt}

Form Fields:
  - file: (binary file data)
  - entity_type: "ticket" | "document" | "comment" | etc.
  - entity_id: "ticket-123"
  - filename: "architecture.png" (optional)
  - description: "System architecture diagram" (optional)
```

### Download File

```bash
GET /v1/files/{reference_id}/download
Authorization: Bearer {jwt}
```

### List Files for Entity

```bash
GET /v1/entities/{entity_type}/{entity_id}/files
Authorization: Bearer {jwt}
```

### Delete File

```bash
DELETE /v1/files/{reference_id}
Authorization: Bearer {jwt}
```

### Health Check

```bash
GET /health

Response:
{
  "status": "healthy",
  "version": "1.0.0",
  "checks": {
    "database": {"status": "healthy", "latency_ms": 5},
    "storage_primary": {"status": "healthy", "latency_ms": 2}
  }
}
```

## Database Schema

The service uses a sophisticated database schema with:

- **attachment_file** - Physical files (deduplicated by hash)
- **attachment_reference** - Logical references (entity-to-file mapping)
- **storage_endpoint** - Storage endpoint configuration
- **storage_health** - Health monitoring data
- **upload_quota** - Per-user quotas
- **access_log** - Audit trail
- **presigned_url** - Temporary access tokens
- **cleanup_job** - Cleanup job tracking

### Key Features

- Automatic reference counting via triggers
- Automatic quota management
- Soft delete support
- Comprehensive indexing for performance

See [Database/DDL/001_initial_schema.sql](Database/DDL/001_initial_schema.sql) for complete schema.

## Security

### Multi-Layer Security Architecture

1. **Network Layer**
   - DDoS protection (rate limiting)
   - Connection limits (per IP, global)
   - Request size limits

2. **Authentication Layer**
   - JWT token validation
   - Permission checks (RBAC)
   - Request signing

3. **Input Validation Layer**
   - MIME type validation
   - File extension validation
   - Path sanitization

4. **Content Validation Layer**
   - Magic bytes verification
   - Image decompression bomb detection
   - Virus scanning (ClamAV integration)

5. **Storage Layer**
   - Encryption at rest (AES-256)
   - Encryption in transit (TLS 1.3)
   - Integrity verification (SHA-256)

6. **Access Control Layer**
   - Presigned URLs (time-limited)
   - Access logging (audit trail)
   - IP-based access control

### Allowed File Types (Default)

- **Images:** JPEG, PNG, GIF, WebP, SVG
- **Documents:** PDF, Word, Excel, PowerPoint, Text, Markdown, CSV
- **Archives:** ZIP, TAR, GZIP
- **Videos:** MP4, WebM, QuickTime

Custom MIME types can be configured in security.allowed_mime_types.

## Performance

### Target Metrics

- Upload Latency: <1s for 10MB files
- Download Latency: <100ms (metadata only)
- Concurrent Uploads: 100+ per instance
- Concurrent Downloads: 1000+ per instance
- Throughput: 1 GB/s per instance

### Optimization Features

- Database connection pooling (50 connections default)
- In-memory caching (LRU, 1GB)
- Redis caching (shared across instances)
- CDN integration support
- Gzip compression support

## High Availability

### Failover Strategy

1. **Primary Storage Unavailable**
   - Automatic failover to backup storage
   - Health check every 30 seconds
   - Circuit breaker pattern

2. **Database Unavailable**
   - Connection retry with exponential backoff
   - Graceful degradation (read-only mode)

3. **Service Instance Failure**
   - Automatic deregistration from service discovery
   - Load balancer redirects traffic to healthy instances

### Replication Modes

- **Synchronous:** Wait for all endpoints before success (highest reliability)
- **Asynchronous:** Return immediately, replicate in background (fastest)
- **Hybrid:** Sync to primary+backup, async to mirrors (recommended)

## Monitoring & Metrics

### Prometheus Metrics

Available at `/metrics`:

- `attachments_uploads_total` - Total uploads
- `attachments_downloads_total` - Total downloads
- `attachments_storage_bytes` - Total storage used
- `attachments_ref_count` - Total file references
- `attachments_request_duration_seconds` - Request latency
- `attachments_errors_total` - Total errors

### Health Checks

- Database connectivity
- Storage endpoint health
- Virus scanner availability
- Service uptime

## Testing

### Unit Tests

```bash
go test ./internal/models -v
go test ./internal/storage -v
go test ./internal/security -v
```

### Integration Tests

```bash
go test ./tests/integration -v
```

### E2E Tests

```bash
go test ./tests/e2e -v
```

### AI QA Automation

```bash
./tests/ai-qa/ai-qa-runner.sh
```

## Deployment

### Docker

```bash
docker build -t helixtrack/attachments-service:1.0.0 .
docker run -p 8090:8090 helixtrack/attachments-service:1.0.0
```

### Docker Compose

```bash
docker-compose up -d
```

### Kubernetes

```bash
kubectl apply -f deployments/k8s/
```

See [docs/DEPLOYMENT.md](docs/DEPLOYMENT.md) for detailed deployment instructions.

## Project Structure

```
Attachments-Service/
├── cmd/
│   └── main.go                    # Service entry point
├── internal/
│   ├── config/                    # Configuration management
│   ├── database/                  # Database layer
│   ├── handlers/                  # HTTP handlers
│   ├── middleware/                # HTTP middleware
│   ├── models/                    # Data models
│   ├── security/                  # Security components
│   │   ├── scanner/               # Virus scanner
│   │   ├── ratelimit/             # Rate limiter
│   │   └── validation/            # Input validation
│   ├── storage/                   # Storage layer
│   │   ├── adapters/              # Storage adapters
│   │   ├── deduplication/         # Deduplication engine
│   │   ├── orchestrator/          # Multi-endpoint orchestrator
│   │   └── reference/             # Reference counter
│   └── utils/                     # Utilities
├── Database/
│   └── DDL/                       # Database schemas
├── configs/                       # Configuration files
├── tests/                         # Test suites
│   ├── unit/
│   ├── integration/
│   ├── e2e/
│   └── ai-qa/
├── docs/                          # Documentation
└── scripts/                       # Utility scripts
```

## Contributing

Contributions are welcome! Please follow these guidelines:

1. Fork the repository
2. Create a feature branch
3. Write tests for new functionality
4. Ensure all tests pass (100% coverage)
5. Submit a pull request

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Support

- Documentation: [docs/](docs/)
- Issues: [GitHub Issues](https://github.com/helixtrack/helixtrack/issues)
- Slack: #attachments-service

## Roadmap

- [x] Core architecture design
- [x] Database schema
- [x] Configuration system
- [x] Service discovery
- [ ] Complete implementation
  - [ ] Deduplication engine
  - [ ] Storage adapters
  - [ ] Security scanner
  - [ ] API handlers
- [ ] 100% test coverage
- [ ] AI QA automation
- [ ] Performance optimization
- [ ] Production deployment
- [ ] Documentation completion

---

**HelixTrack Attachments Service** - Enterprise-grade file storage for the free world.
