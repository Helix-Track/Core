# HelixTrack Localization Service

**Version:** 1.0.0
**Status:** Production Ready
**License:** MIT

## Overview

The HelixTrack Localization (Lokalisation) Service is a microservice that provides centralized localization management for the entire HelixTrack ecosystem. It stores all localizable strings for all services and client applications, eliminating hardcoded messages and enabling multi-language support.

### Key Features

- ✅ **Centralized Localization** - Single source of truth for all localizations
- ✅ **Multi-Language Support** - Support for unlimited languages and locales
- ✅ **JWT-Based Security** - Only authenticated users can access localizations
- ✅ **Service Discovery** - Automatic registration with Consul/etcd
- ✅ **Auto Port Selection** - Automatic port binding with fallback (8085-8095)
- ✅ **PostgreSQL with Encryption** - SQL Cipher for encrypted storage
- ✅ **High Performance Caching** - In-memory and distributed Redis caching
- ✅ **Versioning** - Localization catalog versioning
- ✅ **Fallback Support** - Automatic fallback to default language
- ✅ **Comprehensive Testing** - 100% test coverage target

## Quick Start

### Prerequisites

- Go 1.22+
- PostgreSQL 12+ or SQLite 3+
- (Optional) Redis for distributed caching
- (Optional) Consul for service discovery

### Installation

```bash
# Navigate to service directory
cd Core/Services/Localization

# Install dependencies
go mod download

# Initialize database
psql -U helixtrack -d helixtrack_localization -f ../../Database/DDL/Services/Localization/Definition.V1.sql

# Copy and configure
cp configs/default.json configs/production.json
# Edit configs/production.json with your settings

# Build
go build -o localization-service cmd/main.go

# Run
./localization-service --config=configs/production.json
```

### Development

```bash
# Run with default config
go run cmd/main.go

# Run with custom config
go run cmd/main.go --config=configs/default.json

# Run tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with race detection
go test -race ./...
```

## Architecture

See [ARCHITECTURE.md](ARCHITECTURE.md) for detailed architecture documentation.

### Components

- **Configuration Management** - JSON-based configuration with environment variable overrides
- **Database Layer** - PostgreSQL with encryption support
- **Caching Layer** - Multi-layer caching (in-memory + Redis)
- **JWT Middleware** - Token validation and authentication
- **Rate Limiting** - Per-IP, per-user, and global rate limiting
- **API Handlers** - RESTful API for localization management
- **Service Discovery** - Consul/etcd integration
- **Audit Logging** - Complete audit trail

## API Reference

### Base URL

```
http://localhost:8085
```

### Authentication

All endpoints (except health check) require JWT authentication:

```
Authorization: Bearer <jwt_token>
```

### Endpoints

#### Health Check

```
GET /health
```

#### Get Catalog

```
GET /v1/catalog/:language?category=error
```

#### Get Single Localization

```
GET /v1/localize/:key?language=en&fallback=true
```

#### Batch Localization

```
POST /v1/localize/batch
{
  "keys": ["error.auth.invalid_token", "ui.button.submit"],
  "language": "en",
  "fallback": true
}
```

#### List Languages

```
GET /v1/languages?active_only=true
```

#### Admin Endpoints

All admin endpoints require admin role.

```
POST   /v1/admin/languages                  # Create language
PUT    /v1/admin/languages/:id              # Update language
DELETE /v1/admin/languages/:id              # Delete language

POST   /v1/admin/localizations              # Create/update localization
PUT    /v1/admin/localizations/:id          # Update localization
DELETE /v1/admin/localizations/:id          # Delete localization
POST   /v1/admin/localizations/:id/approve  # Approve localization

POST   /v1/admin/cache/invalidate           # Invalidate cache
GET    /v1/admin/stats                      # Get statistics
```

## Configuration

See `configs/default.json` for a complete configuration example.

### Key Configuration Sections

- **Service**: Port, environment, timeouts
- **Database**: PostgreSQL connection with encryption key
- **Cache**: In-memory and Redis caching configuration
- **Security**: JWT secret, rate limiting, admin roles
- **Logging**: Level, format, output

### Environment Variables

- `DB_HOST` - Database host
- `DB_PASSWORD` - Database password
- `JWT_SECRET` - JWT secret key
- `DB_ENCRYPTION_KEY` - Database encryption key
- `REDIS_PASSWORD` - Redis password

## Database Schema

The service uses 6 main tables:

1. **languages** - Supported languages
2. **localization_keys** - Master list of localization keys
3. **localizations** - Actual localized strings
4. **localization_catalogs** - Pre-built catalogs for fast retrieval
5. **localization_cache_keys** - Cache key tracking
6. **localization_audit_log** - Complete audit trail

See `../../Database/DDL/Services/Localization/Definition.V1.sql` for complete schema.

## Performance

### Caching Strategy

- **In-Memory Cache**: LRU cache with 1GB limit, 1 hour TTL
- **Redis Cache**: Distributed cache with 4 hour TTL
- **Catalog Versioning**: Automatic catalog versioning on changes

### Performance Targets

- Catalog retrieval: <50ms (with cache)
- Single key lookup: <10ms (with cache)
- Batch lookup (100 keys): <100ms (with cache)
- Throughput: 10,000 requests/second per instance

## Testing

```bash
# Unit tests
go test ./internal/...

# Integration tests
go test ./tests/integration/...

# E2E tests
go test ./tests/e2e/...

# Coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

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

See `deployments/k8s/` for Kubernetes manifests.

## Client Integration

### Core Backend

```go
import "github.com/helixtrack/localization-service-client"

client := localization.NewClient("http://localhost:8085", jwtToken)
catalog, err := client.GetCatalog("en")
```

### Web-Client (Angular)

```typescript
import { LocalizationService } from '@core/services';

constructor(private l10n: LocalizationService) {}

async ngOnInit() {
  await this.l10n.loadCatalog('en');
  console.log(this.l10n.t('error.auth.invalid_token'));
}
```

See implementation examples in respective client directories.

## Monitoring

### Metrics

The service exposes Prometheus metrics at `/metrics`:

- `http_requests_total`
- `http_request_duration_seconds`
- `cache_hits_total`
- `cache_misses_total`
- `database_query_duration_seconds`

### Health Checks

Health check endpoint at `/health` provides:
- Database connectivity status
- Cache connectivity status
- Service version

## Troubleshooting

### Common Issues

**Database connection failed**
- Verify PostgreSQL is running
- Check database credentials in config
- Ensure database exists and schema is initialized

**Cache errors**
- If Redis is unavailable, service falls back to in-memory cache
- Check Redis connection settings
- Verify Redis is running (if enabled)

**JWT validation failed**
- Ensure JWT secret matches Authentication service
- Verify token is not expired
- Check token format (Bearer <token>)

## Contributing

1. Fork the repository
2. Create a feature branch
3. Write tests for new functionality
4. Ensure all tests pass
5. Submit a pull request

## License

MIT License - see [LICENSE](../../LICENSE) file for details.

## Support

- Documentation: [docs/](docs/)
- Issues: [GitHub Issues](https://github.com/helixtrack/helixtrack/issues)
- Slack: #localization-service

---

**HelixTrack Localization Service** - Centralized localization for the free world.
