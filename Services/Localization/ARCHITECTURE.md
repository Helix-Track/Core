# HelixTrack Localization Service - Architecture

**Version:** 1.0.0
**Status:** Implementation
**License:** MIT

## Overview

The HelixTrack Localization (Lokalisation) Service is a microservice that provides centralized localization management for the entire HelixTrack ecosystem. It stores all localizable strings for all services and client applications, eliminating hardcoded messages and enabling multi-language support.

### Key Features

- ✅ **Centralized Localization** - Single source of truth for all localizations
- ✅ **Multi-Language Support** - Support for unlimited languages and locales
- ✅ **JWT-Based Security** - Only authenticated users can access localizations
- ✅ **Service Discovery** - Automatic registration with Consul/etcd
- ✅ **Auto Port Selection** - Automatic port binding with fallback
- ✅ **PostgreSQL with Encryption** - SQL Cipher for encrypted storage
- ✅ **High Performance Caching** - In-memory and distributed caching
- ✅ **Versioning** - Localization catalog versioning
- ✅ **Fallback Support** - Automatic fallback to default language
- ✅ **100% Test Coverage** - Comprehensive testing strategy

## System Architecture

```
┌────────────────────────────────────────────────────────┐
│              Localization Service                       │
├────────────────────────────────────────────────────────┤
│                                                         │
│  ┌─────────────┐  ┌──────────────┐  ┌──────────────┐ │
│  │ API Gateway │  │ JWT Auth     │  │ Rate Limiter │ │
│  └─────────────┘  └──────────────┘  └──────────────┘ │
│                                                         │
│  ┌──────────────────────────────────────────────────┐ │
│  │           Business Logic Layer                    │ │
│  │  - Localization Catalog Manager                   │ │
│  │  - Language Manager                               │ │
│  │  - Fallback Resolution                            │ │
│  │  - Version Control                                │ │
│  └──────────────────────────────────────────────────┘ │
│                                                         │
│  ┌──────────────────────────────────────────────────┐ │
│  │              Caching Layer                        │ │
│  │  - In-Memory Cache (LRU)                          │ │
│  │  - Redis Cache (Distributed)                      │ │
│  │  - Cache Invalidation                             │ │
│  └──────────────────────────────────────────────────┘ │
│                                                         │
│  ┌──────────────────────────────────────────────────┐ │
│  │           Database Layer                          │ │
│  │  - PostgreSQL with SQL Cipher                     │ │
│  │  - Encrypted at Rest                              │ │
│  │  - Connection Pooling                             │ │
│  └──────────────────────────────────────────────────┘ │
└────────────────────────────────────────────────────────┘
           │                    │                    │
           ▼                    ▼                    ▼
    ┌────────────┐      ┌────────────┐      ┌────────────┐
    │    Core    │      │ Web Client │      │   Mobile   │
    │  Service   │      │   Angular  │      │   Clients  │
    └────────────┘      └────────────┘      └────────────┘
```

## Database Schema

### Tables

#### 1. `languages`
Supported languages in the system.

```sql
CREATE TABLE languages (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code            VARCHAR(10) NOT NULL UNIQUE,  -- ISO 639-1 (e.g., 'en', 'de', 'fr')
    name            VARCHAR(100) NOT NULL,         -- English name (e.g., 'English')
    native_name     VARCHAR(100),                  -- Native name (e.g., 'Deutsch')
    is_rtl          BOOLEAN DEFAULT FALSE,         -- Right-to-left language
    is_active       BOOLEAN DEFAULT TRUE,          -- Language enabled
    is_default      BOOLEAN DEFAULT FALSE,         -- Default fallback language
    created_at      BIGINT NOT NULL,               -- Unix timestamp
    modified_at     BIGINT NOT NULL,               -- Unix timestamp
    deleted         BOOLEAN DEFAULT FALSE          -- Soft delete
);

CREATE INDEX idx_languages_code ON languages(code);
CREATE INDEX idx_languages_is_default ON languages(is_default);
CREATE INDEX idx_languages_is_active ON languages(is_active);
```

#### 2. `localization_keys`
Master list of all localization keys.

```sql
CREATE TABLE localization_keys (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key             VARCHAR(255) NOT NULL UNIQUE,  -- e.g., 'error.auth.invalid_token'
    category        VARCHAR(100),                  -- e.g., 'error', 'ui', 'message'
    description     TEXT,                          -- Developer notes
    context         VARCHAR(255),                  -- Usage context
    created_at      BIGINT NOT NULL,
    modified_at     BIGINT NOT NULL,
    deleted         BOOLEAN DEFAULT FALSE
);

CREATE INDEX idx_localization_keys_key ON localization_keys(key);
CREATE INDEX idx_localization_keys_category ON localization_keys(category);
```

#### 3. `localizations`
Actual localized strings.

```sql
CREATE TABLE localizations (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key_id          UUID NOT NULL REFERENCES localization_keys(id),
    language_id     UUID NOT NULL REFERENCES languages(id),
    value           TEXT NOT NULL,                 -- Localized string
    plural_forms    JSONB,                         -- Plural forms (optional)
    variables       JSONB,                         -- Variable placeholders
    version         INTEGER DEFAULT 1,             -- Version number
    approved        BOOLEAN DEFAULT FALSE,         -- Reviewed and approved
    approved_by     VARCHAR(255),                  -- Username of approver
    approved_at     BIGINT,                        -- Approval timestamp
    created_at      BIGINT NOT NULL,
    modified_at     BIGINT NOT NULL,
    deleted         BOOLEAN DEFAULT FALSE,

    UNIQUE(key_id, language_id)
);

CREATE INDEX idx_localizations_key_id ON localizations(key_id);
CREATE INDEX idx_localizations_language_id ON localizations(language_id);
CREATE INDEX idx_localizations_approved ON localizations(approved);
CREATE INDEX idx_localizations_version ON localizations(version);
```

#### 4. `localization_catalogs`
Pre-built catalogs for fast retrieval.

```sql
CREATE TABLE localization_catalogs (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    language_id     UUID NOT NULL REFERENCES languages(id),
    category        VARCHAR(100),                  -- Optional category filter
    catalog_data    JSONB NOT NULL,                -- Complete catalog as JSON
    version         INTEGER NOT NULL,              -- Catalog version
    checksum        VARCHAR(64) NOT NULL,          -- SHA-256 checksum
    created_at      BIGINT NOT NULL,
    modified_at     BIGINT NOT NULL,

    UNIQUE(language_id, category, version)
);

CREATE INDEX idx_localization_catalogs_language_id ON localization_catalogs(language_id);
CREATE INDEX idx_localization_catalogs_version ON localization_catalogs(version);
CREATE INDEX idx_localization_catalogs_checksum ON localization_catalogs(checksum);
```

#### 5. `localization_cache_keys`
Cache key tracking for invalidation.

```sql
CREATE TABLE localization_cache_keys (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cache_key       VARCHAR(255) NOT NULL UNIQUE,
    language_code   VARCHAR(10),
    category        VARCHAR(100),
    ttl             INTEGER DEFAULT 3600,          -- TTL in seconds
    expires_at      BIGINT NOT NULL,
    created_at      BIGINT NOT NULL
);

CREATE INDEX idx_localization_cache_keys_cache_key ON localization_cache_keys(cache_key);
CREATE INDEX idx_localization_cache_keys_expires_at ON localization_cache_keys(expires_at);
```

#### 6. `localization_audit_log`
Audit trail for all changes.

```sql
CREATE TABLE localization_audit_log (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    action          VARCHAR(50) NOT NULL,          -- CREATE, UPDATE, DELETE, APPROVE
    entity_type     VARCHAR(50) NOT NULL,          -- LANGUAGE, KEY, LOCALIZATION
    entity_id       UUID NOT NULL,
    username        VARCHAR(255) NOT NULL,
    changes         JSONB,                         -- Before/after values
    ip_address      VARCHAR(45),                   -- IPv4/IPv6
    user_agent      TEXT,
    created_at      BIGINT NOT NULL
);

CREATE INDEX idx_localization_audit_log_entity_id ON localization_audit_log(entity_id);
CREATE INDEX idx_localization_audit_log_username ON localization_audit_log(username);
CREATE INDEX idx_localization_audit_log_created_at ON localization_audit_log(created_at);
```

### Encryption

PostgreSQL data encryption using SQL Cipher:

- **At-Rest Encryption**: All sensitive data encrypted using AES-256
- **Column-Level Encryption**: `localizations.value` column encrypted
- **Transparent Decryption**: Automatic decryption on read with proper credentials

## API Specification

### Base URL

```
http://localhost:8085
```

### Authentication

All endpoints (except health check) require JWT authentication via the `Authorization` header:

```
Authorization: Bearer <jwt_token>
```

### Endpoints

#### 1. Health Check

```
GET /health
```

Response:
```json
{
  "status": "healthy",
  "version": "1.0.0",
  "checks": {
    "database": {"status": "healthy", "latency_ms": 5},
    "cache": {"status": "healthy"}
  }
}
```

#### 2. Get Catalog

Get complete localization catalog for a language.

```
GET /v1/catalog/:language
```

Query Parameters:
- `category` (optional): Filter by category
- `version` (optional): Specific version (default: latest)

Response:
```json
{
  "language": "en",
  "version": 1,
  "checksum": "sha256...",
  "catalog": {
    "error.auth.invalid_token": "Invalid authentication token",
    "ui.button.submit": "Submit",
    "message.welcome": "Welcome to HelixTrack"
  }
}
```

#### 3. Get Single Localization

```
GET /v1/localize/:key
```

Query Parameters:
- `language` (required): Language code
- `fallback` (optional): Enable fallback to default language (default: true)

Response:
```json
{
  "key": "error.auth.invalid_token",
  "language": "en",
  "value": "Invalid authentication token",
  "variables": {},
  "approved": true
}
```

#### 4. Get Multiple Localizations

```
POST /v1/localize/batch
```

Request:
```json
{
  "keys": ["error.auth.invalid_token", "ui.button.submit"],
  "language": "en",
  "fallback": true
}
```

Response:
```json
{
  "language": "en",
  "localizations": {
    "error.auth.invalid_token": "Invalid authentication token",
    "ui.button.submit": "Submit"
  }
}
```

#### 5. List Languages

```
GET /v1/languages
```

Response:
```json
{
  "languages": [
    {
      "id": "uuid...",
      "code": "en",
      "name": "English",
      "native_name": "English",
      "is_rtl": false,
      "is_default": true,
      "is_active": true
    },
    {
      "id": "uuid...",
      "code": "de",
      "name": "German",
      "native_name": "Deutsch",
      "is_rtl": false,
      "is_default": false,
      "is_active": true
    }
  ]
}
```

#### 6. Admin: Create/Update Localization

```
POST /v1/admin/localizations
```

Request:
```json
{
  "key": "error.new.message",
  "language": "en",
  "value": "This is a new error message",
  "category": "error",
  "description": "New error message",
  "approved": false
}
```

#### 7. Admin: Approve Localization

```
POST /v1/admin/localizations/:id/approve
```

#### 8. Admin: Invalidate Cache

```
POST /v1/admin/cache/invalidate
```

Request:
```json
{
  "language": "en",      // Optional
  "category": "error"    // Optional
}
```

## Security

### JWT Validation

- All requests validated against Authentication service
- Token must contain valid user information
- Expired tokens rejected
- Admin endpoints require admin role

### Rate Limiting

- Per-IP: 1000 requests/minute
- Per-User: 5000 requests/minute
- Global: 100,000 requests/minute

### Database Encryption

- PostgreSQL with SQL Cipher
- AES-256 encryption at rest
- Column-level encryption for sensitive data
- Encrypted backups

## Performance & Caching

### Multi-Layer Caching Strategy

#### 1. In-Memory Cache (Service Level)
- LRU cache with 1GB limit
- TTL: 1 hour
- Stores frequently accessed catalogs
- Fast lookups (<1ms)

#### 2. Redis Cache (Distributed)
- Shared across service instances
- TTL: 4 hours
- Enables horizontal scaling
- Cache warming on startup

#### 3. Database-Level Caching
- Materialized views for catalogs
- Automatic refresh on changes
- PostgreSQL query cache

### Cache Invalidation

- Automatic invalidation on updates
- Manual invalidation via admin API
- Time-based expiration
- Version-based invalidation

### Performance Targets

- Catalog retrieval: <50ms (with cache)
- Single key lookup: <10ms (with cache)
- Batch lookup (100 keys): <100ms (with cache)
- Throughput: 10,000 requests/second per instance

## Service Discovery & Port Management

### Service Discovery

- **Provider**: Consul (primary), etcd (alternative)
- **Service Name**: `localization-service`
- **Health Check**: HTTP GET /health every 10s
- **TTL**: 30s
- **Deregistration**: Automatic on shutdown

### Port Configuration

- **Preferred Port**: 8085
- **Port Range**: 8085-8095
- **Auto Selection**: Automatic fallback to next available port
- **Configuration**: JSON-based configuration file

## High Availability

### Replication

- **Database**: PostgreSQL streaming replication
- **Read Replicas**: Support for read-only replicas
- **Failover**: Automatic failover with 99.9% uptime SLA

### Horizontal Scaling

- Stateless service design
- Load balancer compatible
- Session-less (JWT-based auth)
- Redis for distributed caching

### Disaster Recovery

- Automated backups every 6 hours
- Point-in-time recovery
- Cross-region replication (optional)

## Monitoring & Observability

### Metrics (Prometheus)

- `localization_requests_total` - Total requests
- `localization_cache_hits` - Cache hit rate
- `localization_cache_misses` - Cache miss rate
- `localization_db_latency` - Database latency
- `localization_errors_total` - Error count

### Logging

- Structured JSON logging (Uber Zap)
- Log levels: DEBUG, INFO, WARN, ERROR
- Request/response logging
- Audit logging for all changes

### Health Checks

- Database connectivity
- Cache connectivity (Redis)
- Service discovery status
- Disk space
- Memory usage

## Deployment

### Docker

```bash
docker build -t helixtrack/localization-service:1.0.0 .
docker run -p 8085:8085 \
  -e DB_HOST=postgres \
  -e DB_PASSWORD=secret \
  -e JWT_SECRET=your-secret \
  helixtrack/localization-service:1.0.0
```

### Kubernetes

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: localization-service
spec:
  replicas: 3
  selector:
    matchLabels:
      app: localization-service
  template:
    metadata:
      labels:
        app: localization-service
    spec:
      containers:
      - name: localization-service
        image: helixtrack/localization-service:1.0.0
        ports:
        - containerPort: 8085
        env:
        - name: DB_HOST
          value: "postgres-service"
        - name: REDIS_HOST
          value: "redis-service"
```

## Client Integration

### Core Service Integration

```go
// Core/Application/internal/services/localization_service.go
type LocalizationService struct {
    baseURL string
    cache   *Cache
    client  *http.Client
}

func (s *LocalizationService) GetCatalog(language string) (*Catalog, error) {
    // Check cache first
    if cached := s.cache.Get(language); cached != nil {
        return cached, nil
    }

    // Fetch from service
    catalog, err := s.fetchCatalog(language)
    if err != nil {
        return nil, err
    }

    // Cache for 1 hour
    s.cache.Set(language, catalog, 3600)

    return catalog, nil
}
```

### Web Client Integration (Angular)

```typescript
// Web-Client/src/app/core/services/localization.service.ts
@Injectable({providedIn: 'root'})
export class LocalizationService {
  private catalog: Map<string, string> = new Map();
  private currentLanguage = 'en';

  async loadCatalog(language: string): Promise<void> {
    const cached = localStorage.getItem(`l10n_${language}`);
    if (cached) {
      this.catalog = new Map(JSON.parse(cached));
      return;
    }

    const response = await this.http.get<Catalog>(
      `${this.baseURL}/v1/catalog/${language}`
    ).toPromise();

    this.catalog = new Map(Object.entries(response.catalog));
    localStorage.setItem(`l10n_${language}`, JSON.stringify([...this.catalog]));
  }

  t(key: string, variables?: Record<string, any>): string {
    let value = this.catalog.get(key) || key;
    if (variables) {
      Object.entries(variables).forEach(([k, v]) => {
        value = value.replace(`{${k}}`, String(v));
      });
    }
    return value;
  }
}
```

## Testing Strategy

### Unit Tests

- All models: 100% coverage
- All handlers: 100% coverage
- All database operations: 100% coverage
- Cache operations: 100% coverage

### Integration Tests

- Service-to-service communication
- Database integration
- Cache integration
- JWT validation

### E2E Tests

- Complete user workflows
- Multi-language scenarios
- Cache invalidation scenarios
- Failover scenarios

### AI QA Automation

- Intelligent test generation
- Performance regression detection
- Security vulnerability scanning
- Load testing

## Roadmap

- [x] Architecture design
- [ ] Database schema implementation
- [ ] Core service implementation
- [ ] API endpoints
- [ ] JWT authentication
- [ ] Caching layer
- [ ] Service discovery
- [ ] Client integration (Core)
- [ ] Client integration (Web)
- [ ] Client integration (Desktop)
- [ ] Client integration (Mobile)
- [ ] Unit tests (100% coverage)
- [ ] Integration tests
- [ ] E2E tests with AI QA
- [ ] Documentation
- [ ] Production deployment

---

**HelixTrack Localization Service** - Centralized localization for the free world.
