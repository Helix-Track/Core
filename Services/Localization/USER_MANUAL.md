# Localization Service - User Manual

**Version:** 1.0.0
**Last Updated:** 2025-10-21
**Service:** HelixTrack Localization Service

---

## Table of Contents

1. [Overview](#overview)
2. [Features](#features)
3. [Architecture](#architecture)
4. [Installation](#installation)
5. [Configuration](#configuration)
6. [Running the Service](#running-the-service)
7. [API Reference](#api-reference)
8. [Client Integration](#client-integration)
9. [Security](#security)
10. [Deployment](#deployment)
11. [Monitoring](#monitoring)
12. [Troubleshooting](#troubleshooting)
13. [Examples](#examples)

---

## Overview

The HelixTrack Localization Service is a high-performance microservice that provides centralized localization and internationalization (i18n) capabilities for all HelixTrack client applications.

### Purpose

- **Centralized Translation Management**: Single source of truth for all translations
- **Multi-Language Support**: Support for unlimited languages and locales
- **Real-Time Updates**: Dynamic language switching without application restart
- **High Performance**: Multi-layer caching with Redis and in-memory storage
- **Scalable**: Horizontal scaling with service discovery support

### Key Specifications

- **Protocol**: HTTP/3 over QUIC (TLS 1.2/1.3)
- **Port**: 8085 (default, configurable with rotation 8085-8095)
- **Authentication**: JWT token-based with Security Engine integration
- **Database**: PostgreSQL with SQL Cipher encryption
- **Cache**: Multi-layer (In-memory LRU + Redis)
- **Performance**: Sub-10ms response time (cached), 100-300ms (uncached)

---

## Features

### Core Features

✅ **Multi-Language Support**
- Unlimited language support
- Right-to-Left (RTL) language support (Arabic, Hebrew, Farsi, Urdu)
- Locale-specific variations
- Default language fallback

✅ **Dynamic Content**
- Variable interpolation: `"Hello {name}"`
- Plural forms support (planned)
- Number/date formatting (planned)
- Currency localization (planned)

✅ **High Performance**
- Multi-layer caching (In-memory + Redis)
- HTTP/3 QUIC for low-latency communication
- Compression (gzip)
- Connection pooling

✅ **Security**
- JWT authentication
- SQL Cipher database encryption
- TLS 1.3 encryption
- Rate limiting
- RBAC (Role-Based Access Control)

✅ **Scalability**
- Horizontal scaling support
- Service discovery (Consul/etcd)
- Stateless design
- Distributed caching

✅ **Developer-Friendly**
- RESTful API
- Batch operations
- Comprehensive error messages
- Health checks
- OpenAPI documentation (planned)

---

## Architecture

### System Architecture

```
┌─────────────────────────────────────────────────────────┐
│                  Client Applications                     │
│  (Web, Desktop, Android, iOS, Core Backend)             │
└─────────────────────────────────────────────────────────┘
                           │
                    HTTP/3 QUIC (TLS)
                      JWT Auth
                           │
                           ▼
┌─────────────────────────────────────────────────────────┐
│            Localization Service (Port 8085)              │
│  ┌─────────────┐  ┌──────────┐  ┌────────────────┐    │
│  │   Handlers  │──│Middleware│──│  Rate Limiter  │    │
│  └─────────────┘  └──────────┘  └────────────────┘    │
│         │              │                                 │
│         ▼              ▼                                 │
│  ┌─────────────┐  ┌──────────┐                         │
│  │   Cache     │  │   Auth   │                         │
│  │ In-Memory   │  │  (JWT)   │                         │
│  │  + Redis    │  └──────────┘                         │
│  └─────────────┘                                        │
│         │                                                │
│         ▼                                                │
│  ┌─────────────┐                                        │
│  │  Database   │                                        │
│  │ PostgreSQL  │                                        │
│  │(SQL Cipher) │                                        │
│  └─────────────┘                                        │
└─────────────────────────────────────────────────────────┘
                           │
                           ▼
          ┌────────────────────────────────┐
          │  Service Discovery (Optional)   │
          │    Consul / etcd                │
          └────────────────────────────────┘
```

### Database Schema

The service uses PostgreSQL with 6 core tables:

1. **languages** - Supported languages and metadata
2. **localization_keys** - Translation keys with categories
3. **localizations** - Actual translations
4. **localization_catalogs** - Compiled catalogs (cached)
5. **localization_audit_log** - Audit trail
6. **localization_stats** - Usage statistics

---

## Installation

### Prerequisites

- **Go 1.22+**
- **PostgreSQL 12+** (with pgcrypto extension)
- **Redis** (optional, for distributed caching)
- **OpenSSL** (for TLS certificate generation)

### Build from Source

```bash
# Clone repository
cd /home/milosvasic/Projects/HelixTrack/Core/Services/Localization

# Download dependencies
go mod tidy
go mod download

# Build service
go build -o localization-service ./cmd/main.go

# Generate TLS certificates (development)
./scripts/generate-certs.sh

# Initialize database
psql -U postgres -f ../../Database/DDL/Services/Localization/Definition.V1.sql
```

### Binary Installation

```bash
# Copy binary to system path
sudo cp localization-service /usr/local/bin/
sudo chmod +x /usr/local/bin/localization-service

# Copy configuration
sudo mkdir -p /etc/helixtrack/localization
sudo cp configs/default.json /etc/helixtrack/localization/config.json
```

---

## Configuration

### Configuration File

The service is configured via JSON file (default: `configs/default.json`):

```json
{
  "service": {
    "name": "localization-service",
    "port": 8085,
    "port_range": [8085, 8095],
    "environment": "development",
    "read_timeout": 30,
    "write_timeout": 30,
    "max_header_bytes": 8192,
    "tls_cert_file": "certs/server.crt",
    "tls_key_file": "certs/server.key",
    "discovery": {
      "enabled": false,
      "provider": "consul",
      "consul_address": "localhost:8500",
      "etcd_endpoints": []
    }
  },
  "database": {
    "driver": "postgres",
    "host": "localhost",
    "port": 5432,
    "database": "helixtrack_localization",
    "user": "helixtrack",
    "password": "helixtrack",
    "ssl_mode": "disable",
    "max_connections": 50,
    "idle_connections": 10,
    "connection_timeout": 30,
    "connection_lifetime": 3600,
    "encryption_key": "your-32-byte-encryption-key-here!"
  },
  "cache": {
    "in_memory": {
      "enabled": true,
      "max_size_mb": 1024,
      "default_ttl": 3600,
      "cleanup_interval": 300
    },
    "redis": {
      "enabled": false,
      "addresses": ["localhost:6379"],
      "password": "",
      "database": 0,
      "max_retries": 3,
      "pool_size": 10,
      "default_ttl": 14400
    }
  },
  "security": {
    "jwt_secret": "your-jwt-secret-key-change-in-production",
    "jwt_issuer": "helixtrack-auth",
    "rate_limiting": {
      "per_ip_requests_per_minute": 1000,
      "per_user_requests_per_minute": 5000,
      "global_requests_per_minute": 100000
    },
    "admin_roles": ["admin", "superadmin"]
  },
  "logging": {
    "level": "info",
    "format": "json",
    "output": "stdout"
  }
}
```

### Environment Variables

Override configuration with environment variables:

```bash
export DB_HOST=postgres.example.com
export DB_PASSWORD=secure_password
export JWT_SECRET=your_jwt_secret
export DB_ENCRYPTION_KEY=your_32_byte_encryption_key
export REDIS_PASSWORD=redis_password
```

### Key Management

Use the HelixTrack Key Manager to generate secure keys:

```bash
# Generate JWT secret
keymanager generate -type jwt -name jwt-secret -service localization -length 64

# Generate database encryption key
keymanager generate -type db -name db-key -service localization -length 32

# Generate TLS certificate
keymanager generate -type tls -name service-tls -service localization

# Generate Redis password
keymanager generate -type redis -name redis-pwd -service localization -length 32
```

---

## Running the Service

### Development Mode

```bash
# Start with default configuration
./localization-service

# Start with custom configuration
./localization-service --config=configs/dev.json

# With environment variables
DB_PASSWORD=secure_pass ./localization-service
```

### Production Mode

```bash
# Systemd service
sudo systemctl start helixtrack-localization
sudo systemctl enable helixtrack-localization
sudo systemctl status helixtrack-localization

# View logs
sudo journalctl -u helixtrack-localization -f
```

### Docker

```bash
# Build Docker image
docker build -t helixtrack/localization:1.0.0 .

# Run container
docker run -d \
  --name localization-service \
  -p 8085:8085 \
  -e DB_HOST=postgres \
  -e DB_PASSWORD=secure_pass \
  -e JWT_SECRET=your_secret \
  -v /path/to/certs:/app/certs \
  helixtrack/localization:1.0.0
```

### Health Check

```bash
# Check service health
curl -k https://localhost:8085/health

# Expected response:
# {
#   "status": "healthy",
#   "timestamp": "2025-10-21T12:30:00Z",
#   "version": "1.0.0"
# }
```

---

## API Reference

### Base URL

```
https://localhost:8085
```

### Authentication

All endpoints (except `/health`) require JWT authentication:

```http
Authorization: Bearer <jwt_token>
```

### Common Response Format

**Success:**
```json
{
  "success": true,
  "data": { ... }
}
```

**Error:**
```json
{
  "success": false,
  "error": {
    "code": 1001,
    "message": "Error description"
  }
}
```

### Endpoints

#### 1. Health Check

**GET /health**

Check service health status.

**Request:**
```bash
curl -k https://localhost:8085/health
```

**Response:**
```json
{
  "status": "healthy",
  "timestamp": "2025-10-21T12:30:00Z",
  "version": "1.0.0"
}
```

---

#### 2. Get Catalog

**GET /v1/catalog/:language**

Retrieve complete localization catalog for a language.

**Parameters:**
- `language` (path) - Language code (e.g., "en", "de", "fr")

**Request:**
```bash
curl -k -H "Authorization: Bearer <token>" \
  https://localhost:8085/v1/catalog/en
```

**Response:**
```json
{
  "success": true,
  "data": {
    "language": "en",
    "version": 1,
    "checksum": "abc123...",
    "catalog": {
      "app.welcome": "Welcome to HelixTrack",
      "app.hello": "Hello {name}",
      "app.error": "An error occurred"
    }
  }
}
```

---

#### 3. Get Single Localization

**GET /v1/localize/:key**

Fetch a single localization with optional fallback.

**Parameters:**
- `key` (path) - Localization key
- `language` (query) - Language code
- `fallback` (query) - Enable fallback to default language (default: true)

**Request:**
```bash
curl -k -H "Authorization: Bearer <token>" \
  "https://localhost:8085/v1/localize/app.welcome?language=en&fallback=true"
```

**Response:**
```json
{
  "success": true,
  "data": {
    "key": "app.welcome",
    "language": "en",
    "value": "Welcome to HelixTrack",
    "variables": [],
    "approved": true
  }
}
```

---

#### 4. Batch Localization

**POST /v1/localize/batch**

Fetch multiple localizations in one request.

**Request Body:**
```json
{
  "language": "en",
  "keys": ["app.welcome", "app.error", "app.success"],
  "fallback": true
}
```

**Request:**
```bash
curl -k -X POST -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"language":"en","keys":["app.welcome","app.error"],"fallback":true}' \
  https://localhost:8085/v1/localize/batch
```

**Response:**
```json
{
  "success": true,
  "data": {
    "language": "en",
    "localizations": {
      "app.welcome": "Welcome to HelixTrack",
      "app.error": "An error occurred"
    }
  }
}
```

---

#### 5. List Languages

**GET /v1/languages**

Get list of available languages.

**Parameters:**
- `active_only` (query) - Return only active languages (default: false)

**Request:**
```bash
curl -k -H "Authorization: Bearer <token>" \
  "https://localhost:8085/v1/languages?active_only=true"
```

**Response:**
```json
{
  "success": true,
  "data": {
    "languages": [
      {
        "id": "lang-1",
        "code": "en",
        "name": "English",
        "native_name": "English",
        "is_rtl": false,
        "is_active": true,
        "is_default": true
      },
      {
        "id": "lang-2",
        "code": "ar",
        "name": "Arabic",
        "native_name": "العربية",
        "is_rtl": true,
        "is_active": true,
        "is_default": false
      }
    ]
  }
}
```

---

#### 6. Create Language (Admin)

**POST /v1/admin/languages**

Create a new language.

**Requires:** Admin role

**Request Body:**
```json
{
  "code": "de",
  "name": "German",
  "native_name": "Deutsch",
  "is_rtl": false,
  "is_active": true,
  "is_default": false
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "lang-3",
    "code": "de",
    "name": "German",
    "native_name": "Deutsch",
    "created_at": "2025-10-21T12:30:00Z"
  }
}
```

---

#### 7. Create Localization Key (Admin)

**POST /v1/admin/keys**

Create a new localization key.

**Request Body:**
```json
{
  "key": "app.new_feature",
  "description": "New feature message",
  "category": "app",
  "variables": ["feature_name"]
}
```

---

#### 8. Create Localization (Admin)

**POST /v1/admin/localizations**

Create a new translation.

**Request Body:**
```json
{
  "key_id": "key-123",
  "language_id": "lang-1",
  "value": "Welcome to the new feature: {feature_name}",
  "approved": false
}
```

---

## Client Integration

### Client Libraries

Official client libraries are available for all platforms:

| Platform | Location | Documentation |
|----------|----------|---------------|
| Go (Core Backend) | `Core/Application/internal/services/localization_service.go` | [Inline](../../../Application/internal/services/) |
| Angular (Web) | `Web-Client/src/app/core/services/localization.service.ts` | [README](../../../Web-Client/README.md) |
| Tauri (Desktop) | `Desktop-Client/src/app/core/services/localization.service.ts` | [README](../../../Desktop-Client/README.md) |
| Kotlin (Android) | `Android-Client/app/src/main/java/com/helixtrack/android/services/LocalizationService.kt` | [README](../../../Android-Client/README.md) |
| Swift (iOS) | `iOS-Client/Sources/Services/LocalizationService.swift` | [README](../../../iOS-Client/README.md) |

### Quick Integration Example (TypeScript)

```typescript
import { LocalizationService } from '@core/services/localization.service';

export class MyComponent {
  constructor(private localization: LocalizationService) {}

  async ngOnInit() {
    // Set service URL
    this.localization.setServiceUrl('https://localhost:8085');

    // Load catalog
    await this.localization.loadCatalog('en', this.jwtToken);

    // Get translation
    const welcome = this.localization.localize('app.welcome');
    console.log(welcome); // "Welcome to HelixTrack"

    // With variables
    const hello = this.localization.localize('app.hello', { name: 'Alice' });
    console.log(hello); // "Hello Alice"

    // Batch localization
    const batch = this.localization.localizeBatch([
      'app.welcome',
      'app.error',
      'app.success'
    ]);
  }
}
```

---

## Security

### Authentication

The service uses JWT tokens issued by the Security Engine:

**Token Format:**
```json
{
  "sub": "authentication",
  "username": "user@example.com",
  "role": "admin",
  "permissions": "READ|CREATE|UPDATE|DELETE",
  "iat": 1697891234,
  "exp": 1697977634
}
```

### Authorization

- **Public Endpoints**: `/health` (no auth required)
- **User Endpoints**: Catalog retrieval, localization fetch (requires valid JWT)
- **Admin Endpoints**: Create/update/delete operations (requires admin role)

### Rate Limiting

Default rate limits:
- Per IP: 1,000 requests/minute
- Per User: 5,000 requests/minute
- Global: 100,000 requests/minute

### Data Encryption

- **In Transit**: TLS 1.3 encryption (HTTP/3 QUIC)
- **At Rest**: SQL Cipher database encryption (AES-256)
- **Cache**: Encrypted Redis connections (optional)

---

## Deployment

### Systemd Service

Create `/etc/systemd/system/helixtrack-localization.service`:

```ini
[Unit]
Description=HelixTrack Localization Service
After=network.target postgresql.service

[Service]
Type=simple
User=helixtrack
Group=helixtrack
WorkingDirectory=/opt/helixtrack/localization
ExecStart=/usr/local/bin/localization-service --config=/etc/helixtrack/localization/config.json
Restart=always
RestartSec=5

# Environment
Environment=DB_PASSWORD=secure_password
Environment=JWT_SECRET=your_jwt_secret

# Security
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/opt/helixtrack/localization/keys

[Install]
WantedBy=multi-user.target
```

### Kubernetes

See [DEPLOYMENT.md](./DEPLOYMENT.md) for Kubernetes manifests.

### High Availability

For production deployment:

1. **Multiple Instances**: Run 3+ instances behind load balancer
2. **Distributed Cache**: Use Redis cluster for cache
3. **Database Replication**: PostgreSQL primary-replica setup
4. **Service Discovery**: Use Consul or etcd
5. **Health Checks**: Configure load balancer health checks

---

## Monitoring

### Metrics

The service exposes metrics for monitoring (future enhancement):

- Request count by endpoint
- Response time percentiles (p50, p95, p99)
- Cache hit rate
- Database connection pool usage
- Error rate by type

### Logging

Structured JSON logging with fields:

```json
{
  "timestamp": "2025-10-21T12:30:00Z",
  "level": "info",
  "service": "localization-service",
  "endpoint": "/v1/catalog/en",
  "method": "GET",
  "status": 200,
  "duration_ms": 15,
  "user": "user@example.com"
}
```

### Alerts

Recommended alerts:
- Response time > 500ms (p95)
- Error rate > 1%
- Cache hit rate < 90%
- Database connection pool > 80% usage

---

## Troubleshooting

### Common Issues

**Issue: Service fails to start**
```
Error: TLS certificate and key files are required for HTTP/3
```
**Solution**: Generate certificates with `./scripts/generate-certs.sh`

---

**Issue: Database connection failed**
```
Error: failed to connect to database: connection refused
```
**Solution**:
- Check PostgreSQL is running
- Verify database credentials
- Check firewall rules

---

**Issue: High response times**
```
Response time > 1000ms
```
**Solution**:
- Check cache hit rate (should be >90%)
- Verify Redis is running (if enabled)
- Check database query performance
- Increase cache TTL

---

**Issue: JWT authentication failures**
```
Error: invalid JWT token
```
**Solution**:
- Verify JWT secret matches Security Engine
- Check token expiration
- Validate token format

---

## Examples

### Complete Workflow Example

```bash
# 1. Start the service
./localization-service --config=configs/default.json

# 2. Create a language (admin)
curl -k -X POST -H "Authorization: Bearer <admin_token>" \
  -H "Content-Type: application/json" \
  -d '{"code":"de","name":"German","native_name":"Deutsch","is_rtl":false}' \
  https://localhost:8085/v1/admin/languages

# 3. Create a localization key
curl -k -X POST -H "Authorization: Bearer <admin_token>" \
  -H "Content-Type: application/json" \
  -d '{"key":"app.welcome","description":"Welcome message","category":"app"}' \
  https://localhost:8085/v1/admin/keys

# 4. Add translations
curl -k -X POST -H "Authorization: Bearer <admin_token>" \
  -H "Content-Type: application/json" \
  -d '{"key_id":"key-1","language_id":"lang-1","value":"Welcome to HelixTrack"}' \
  https://localhost:8085/v1/admin/localizations

# 5. Fetch catalog
curl -k -H "Authorization: Bearer <token>" \
  https://localhost:8085/v1/catalog/en

# 6. Use in application
# See client integration examples above
```

---

## Appendix

### Supported Languages

The service supports all ISO 639-1 language codes, including:

- English (en)
- German (de)
- French (fr)
- Spanish (es)
- Portuguese (pt)
- Italian (it)
- Russian (ru)
- Chinese (zh)
- Japanese (ja)
- Korean (ko)
- Arabic (ar) - RTL
- Hebrew (he) - RTL
- Farsi (fa) - RTL
- Urdu (ur) - RTL

### Error Codes

| Code | Message | Description |
|------|---------|-------------|
| 1001 | Invalid request | Malformed request body |
| 1002 | Missing parameter | Required parameter not provided |
| 2001 | Database error | Database operation failed |
| 2002 | Internal server error | Unexpected server error |
| 3001 | Entity not found | Requested entity does not exist |
| 3002 | Already exists | Entity already exists |
| 4001 | Unauthorized | Invalid or missing JWT token |
| 4002 | Forbidden | Insufficient permissions |

---

**For Support**: Contact HelixTrack Development Team
**Documentation**: https://docs.helixtrack.com
**Repository**: https://github.com/Helix-Track/Core
