# HelixTrack Core - Docker Deployment Guide

**Version:** 1.0.0
**Status:** Production Ready
**Last Updated:** 2025-10-10

---

## Table of Contents

1. [Overview](#overview)
2. [Quick Start](#quick-start)
3. [Configuration](#configuration)
4. [Database Options](#database-options)
5. [Running the Application](#running-the-application)
6. [Testing](#testing)
7. [Monitoring](#monitoring)
8. [Troubleshooting](#troubleshooting)
9. [Production Deployment](#production-deployment)

---

## Overview

HelixTrack Core provides comprehensive Docker support with:

- **Multiple database backends:** SQLite (default) and PostgreSQL
- **Full parametrization:** All settings configurable via environment variables
- **Docker Compose:** Complete orchestration for all services
- **Mock services:** Built-in mock authentication and permissions services
- **Monitoring:** Optional Prometheus and Grafana integration
- **Testing:** Automated AI QA test suite for both databases

### Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Docker Network                            │
│                                                              │
│  ┌────────────────┐  ┌──────────────┐  ┌─────────────────┐ │
│  │ HelixTrack Core│  │ PostgreSQL   │  │  Mock Services  │ │
│  │  (Main Service)│  │  (Optional)  │  │  Auth & Perm    │ │
│  └────────────────┘  └──────────────┘  └─────────────────┘ │
│         ↓                    ↓                   ↓          │
│  ┌────────────────┐  ┌──────────────┐  ┌─────────────────┐ │
│  │  Metrics API   │  │   Database   │  │   Health Checks │ │
│  │  (Port 9090)   │  │ (SQLite/PG)  │  │   (HTTP /health)│ │
│  └────────────────┘  └──────────────┘  └─────────────────┘ │
│                                                              │
│  Optional Monitoring Stack:                                 │
│  ┌────────────────┐  ┌──────────────┐                      │
│  │   Prometheus   │  │   Grafana    │                      │
│  │   (Port 9091)  │  │  (Port 3000) │                      │
│  └────────────────┘  └──────────────┘                      │
└─────────────────────────────────────────────────────────────┘
```

---

## Quick Start

### Prerequisites

- Docker 20.10+
- Docker Compose 2.0+
- 2GB free disk space
- Ports 8080, 8081, 8082 available

### SQLite (Easiest - No external database)

```bash
# 1. Copy environment file
cp .env.example .env.sqlite

# 2. Run with SQLite
./docker-run-sqlite.sh

# 3. Test the API
curl http://localhost:8080/health
```

### PostgreSQL (Production-ready)

```bash
# 1. Copy environment file
cp .env.example .env.postgres

# 2. Configure database password (edit .env.postgres)
# Change DB_PASSWORD to a secure password

# 3. Run with PostgreSQL
./docker-run-postgres.sh

# 4. Test the API
curl http://localhost:8080/health
```

### Using Docker Compose Directly

```bash
# SQLite
docker-compose up -d

# PostgreSQL
docker-compose -f docker-compose.postgres.yml up -d

# With monitoring
docker-compose -f docker-compose.postgres.yml --profile monitoring up -d

# With admin tools (pgAdmin)
docker-compose -f docker-compose.postgres.yml --profile admin-tools up -d
```

---

## Configuration

### Environment Variables

All configuration can be done via environment variables or .env files.

**Priority (highest to lowest):**
1. Environment variables set in shell
2. `.env.{sqlite|postgres}` file
3. Dockerfile defaults

### Configuration Files

| File | Purpose |
|------|---------|
| `.env.example` | Template with all options documented |
| `.env.sqlite` | SQLite-specific configuration |
| `.env.postgres` | PostgreSQL-specific configuration |
| `.env.test` | Testing configuration |

### Key Configuration Options

**Database:**
```bash
DB_TYPE=sqlite                    # or postgresql
DB_PATH=/app/Database/helixtrack.db
DB_HOST=postgres
DB_PORT=5432
DB_NAME=helixtrack
DB_USER=helixtrack
DB_PASSWORD=secure_password_here
```

**Server:**
```bash
SERVER_HOST=0.0.0.0
SERVER_PORT=8080
LOG_LEVEL=info                    # debug, info, warn, error
```

**Service Discovery:**
```bash
HEALTH_CHECK_INTERVAL=60s
HEALTH_CHECK_TIMEOUT=10s
HEALTH_FAILURE_THRESHOLD=3
FAILOVER_ENABLED=true
FAILOVER_STABILITY_COUNT=3
FAILBACK_DELAY=5m
```

**Security:**
```bash
ADMIN_TOKEN_MIN_LENGTH=32
SERVICE_ROTATION_MIN_AGE=5m
TLS_ENABLED=false
RATE_LIMIT_ENABLED=true
```

---

## Database Options

### SQLite (Development & Small Deployments)

**Advantages:**
- No external database required
- Zero configuration
- Perfect for development and testing
- Fast for small datasets

**Limitations:**
- Not suitable for high concurrency
- Single-server only
- Limited to ~100 concurrent connections

**Configuration:**
```yaml
environment:
  - DB_TYPE=sqlite
  - DB_PATH=/app/Database/helixtrack.db

volumes:
  - ./Database:/app/Database
```

### PostgreSQL (Production)

**Advantages:**
- High concurrency support
- ACID compliance
- Advanced features (full-text search, JSON)
- Horizontal scalability
- Battle-tested reliability

**Configuration:**
```yaml
environment:
  - DB_TYPE=postgresql
  - DB_HOST=postgres
  - DB_PORT=5432
  - DB_NAME=helixtrack
  - DB_USER=helixtrack
  - DB_PASSWORD=your_secure_password
  - DB_SSLMODE=disable  # Use 'require' in production
```

**Production Settings:**
```bash
# Connection pooling
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=5
DB_CONN_MAX_LIFETIME=5m

# SSL/TLS
DB_SSLMODE=require  # or verify-ca, verify-full
```

---

## Running the Application

### Using Helper Scripts (Recommended)

**SQLite:**
```bash
./docker-run-sqlite.sh
```

**PostgreSQL:**
```bash
./docker-run-postgres.sh
```

**Tests:**
```bash
./docker-run-tests.sh
```

### Manual Docker Compose

**Start services:**
```bash
# SQLite
docker-compose up -d

# PostgreSQL
docker-compose -f docker-compose.postgres.yml up -d

# Build from scratch
docker-compose up --build -d
```

**View logs:**
```bash
docker-compose logs -f helixtrack-core

# All services
docker-compose logs -f

# Last 100 lines
docker-compose logs --tail=100 helixtrack-core
```

**Stop services:**
```bash
docker-compose down

# Remove volumes (WARNING: deletes data)
docker-compose down -v
```

**Restart a service:**
```bash
docker-compose restart helixtrack-core
```

### Accessing Services

| Service | URL | Purpose |
|---------|-----|---------|
| Core API | http://localhost:8080 | Main API endpoint |
| Health Check | http://localhost:8080/health | Service health status |
| Metrics | http://localhost:9090 | Prometheus metrics |
| Mock Auth | http://localhost:8081 | Mock authentication service |
| Mock Perm | http://localhost:8082 | Mock permissions service |
| PostgreSQL | localhost:5432 | Database connection |
| pgAdmin | http://localhost:5050 | PostgreSQL admin (optional) |
| Prometheus | http://localhost:9091 | Metrics collection (optional) |
| Grafana | http://localhost:3000 | Metrics visualization (optional) |

### Using Optional Services

**Monitoring Stack (Prometheus + Grafana):**
```bash
docker-compose -f docker-compose.postgres.yml \
  --profile monitoring up -d
```

**Admin Tools (pgAdmin for PostgreSQL):**
```bash
docker-compose -f docker-compose.postgres.yml \
  --profile admin-tools up -d
```

**All Optional Services:**
```bash
docker-compose -f docker-compose.postgres.yml \
  --profile monitoring --profile admin-tools up -d
```

---

## Testing

### AI QA Test Suite

The AI QA test suite automatically tests both SQLite and PostgreSQL configurations.

**Run all tests:**
```bash
./docker-run-tests.sh
```

**Manual test execution:**
```bash
# Start test environment
docker-compose -f docker-compose.test.yml up -d

# Run AI QA tests
docker-compose -f docker-compose.test.yml \
  --profile ai-qa-test up ai-qa-test-runner

# View results
cat test-results/ai-qa-report.json

# Cleanup
docker-compose -f docker-compose.test.yml down -v
```

### Test Coverage

The AI QA test suite includes:
- ✅ Health check verification
- ✅ Service registration
- ✅ Service discovery
- ✅ Service listing
- ✅ Security validation (invalid registrations)
- ✅ Concurrent operations
- ✅ Both SQLite and PostgreSQL backends

### Expected Results

```
========================================
TEST SUMMARY
========================================
Total:   12
Passed:  12
Failed:  0
Skipped: 0
========================================

Success Rate: 100.00%

🎉 ALL TESTS PASSED! 🎉
```

---

## Monitoring

### Prometheus Metrics

**Access metrics:**
```bash
curl http://localhost:9090/metrics
```

**Available metrics:**
- `helixtrack_http_requests_total` - Total HTTP requests
- `helixtrack_http_request_duration_seconds` - Request duration
- `helixtrack_service_discovery_services_total` - Registered services
- `helixtrack_service_discovery_healthy_services` - Healthy services
- `helixtrack_service_discovery_failover_events_total` - Failover events

### Grafana Dashboards

**Access Grafana:**
```bash
# Start with monitoring profile
docker-compose -f docker-compose.postgres.yml \
  --profile monitoring up -d

# Open in browser
open http://localhost:3000
```

**Default credentials:**
- Username: `admin`
- Password: `admin`

### Health Checks

**Docker health checks:**
```bash
docker ps --format "table {{.Names}}\t{{.Status}}"
```

**Manual health check:**
```bash
# Core service
curl http://localhost:8080/health

# Expected response
{"status":"healthy","timestamp":"2025-10-10T10:00:00Z"}
```

---

## Troubleshooting

### Service Won't Start

**Check logs:**
```bash
docker-compose logs helixtrack-core
```

**Common issues:**

1. **Port already in use:**
```bash
# Check what's using port 8080
lsof -i :8080

# Change port in .env file
SERVER_PORT=8081
```

2. **Database connection failed:**
```bash
# Check PostgreSQL logs
docker-compose logs postgres

# Verify database is healthy
docker exec helixtrack-postgres pg_isready -U helixtrack
```

3. **Permission denied:**
```bash
# Fix directory permissions
chmod -R 755 Database logs
chown -R 1000:1000 Database logs
```

### Database Issues

**PostgreSQL won't start:**
```bash
# Check PostgreSQL logs
docker-compose -f docker-compose.postgres.yml logs postgres

# Remove corrupted data (WARNING: deletes data)
docker-compose -f docker-compose.postgres.yml down -v
docker volume rm helixtrack-postgres-data
```

**SQLite database locked:**
```bash
# Stop all services
docker-compose down

# Remove lock file
rm Database/.helixtrack.db-lock

# Restart
docker-compose up -d
```

### Performance Issues

**High memory usage:**
```bash
# Check resource usage
docker stats

# Adjust limits in docker-compose.yml
deploy:
  resources:
    limits:
      memory: 512M
```

**Slow database queries:**
```bash
# For PostgreSQL, run VACUUM
docker exec helixtrack-postgres psql -U helixtrack -d helixtrack -c "VACUUM ANALYZE;"

# For SQLite
docker exec helixtrack-core-sqlite sqlite3 /app/Database/helixtrack.db "VACUUM;"
```

### Network Issues

**Services can't communicate:**
```bash
# Check network
docker network ls
docker network inspect helixtrack-network

# Recreate network
docker-compose down
docker network rm helixtrack-network
docker-compose up -d
```

### Reset Everything

**Complete reset (WARNING: deletes all data):**
```bash
# Stop all containers
docker-compose down -v
docker-compose -f docker-compose.postgres.yml down -v

# Remove volumes
docker volume rm helixtrack-postgres-data helixtrack-prometheus-data helixtrack-grafana-data

# Remove database files
rm -rf Database/*.db Database/*.db-*

# Remove logs
rm -rf logs/*

# Rebuild and restart
docker-compose up --build -d
```

---

## Production Deployment

### Checklist

Before deploying to production:

- [ ] Change all default passwords
- [ ] Enable TLS/HTTPS (`TLS_ENABLED=true`)
- [ ] Use PostgreSQL (not SQLite)
- [ ] Configure backup strategy
- [ ] Set appropriate resource limits
- [ ] Enable monitoring (Prometheus + Grafana)
- [ ] Configure log rotation
- [ ] Set up alerts
- [ ] Test failover scenarios
- [ ] Document disaster recovery procedures

### Security Hardening

**1. Use strong passwords:**
```bash
# Generate secure password
openssl rand -base64 32

# Update .env.postgres
DB_PASSWORD=<generated_password>
```

**2. Enable TLS:**
```bash
# Generate certificates
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout certs/server.key -out certs/server.crt

# Update .env
TLS_ENABLED=true
TLS_CERT_FILE=/app/certs/server.crt
TLS_KEY_FILE=/app/certs/server.key
```

**3. Enable PostgreSQL SSL:**
```bash
DB_SSLMODE=require  # or verify-ca, verify-full
```

**4. Restrict network access:**
```yaml
# In docker-compose.yml
ports:
  - "127.0.0.1:8080:8080"  # Only local access
```

### Resource Limits

**Production settings:**
```yaml
deploy:
  resources:
    limits:
      cpus: '2.0'
      memory: 2G
    reservations:
      cpus: '1.0'
      memory: 1G
  restart_policy:
    condition: on-failure
    delay: 5s
    max_attempts: 3
```

### Backup Strategy

**PostgreSQL backups:**
```bash
# Backup
docker exec helixtrack-postgres pg_dump -U helixtrack helixtrack > backup.sql

# Restore
cat backup.sql | docker exec -i helixtrack-postgres psql -U helixtrack -d helixtrack

# Automated daily backups
0 2 * * * /path/to/backup-script.sh
```

**SQLite backups:**
```bash
# Backup
docker exec helixtrack-core-sqlite sqlite3 /app/Database/helixtrack.db ".backup '/app/Database/backup.db'"

# Copy from container
docker cp helixtrack-core-sqlite:/app/Database/backup.db ./backups/
```

### Monitoring in Production

**1. Configure Prometheus:**
```yaml
# monitoring/prometheus.yml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  - job_name: 'helixtrack-core'
    static_configs:
      - targets: ['helixtrack-core:9090']
```

**2. Set up alerts:**
```yaml
# monitoring/alerts.yml
groups:
  - name: helixtrack
    rules:
      - alert: ServiceDown
        expr: up{job="helixtrack-core"} == 0
        for: 1m
        annotations:
          summary: "HelixTrack Core is down"
```

**3. Configure Grafana dashboards:**
- Import pre-built dashboards from `monitoring/grafana/dashboards/`
- Set up alert notifications (email, Slack, PagerDuty)

---

## Additional Resources

- **Technical Documentation:** [Documentation/ServiceDiscovery_Technical.md](Documentation/ServiceDiscovery_Technical.md)
- **User Manual:** [Documentation/ServiceDiscovery_UserManual.md](Documentation/ServiceDiscovery_UserManual.md)
- **HTML Documentation:** [Documentation/html/index.html](Documentation/html/index.html)
- **Main README:** [README.md](README.md)

---

## Support

**Issues:** https://github.com/helixtrack/core/issues
**Documentation:** https://docs.helixtrack.ru
**Email:** support@helixtrack.ru

---

**Last Updated:** 2025-10-10
**Version:** 1.0.0
**Status:** ✅ Production Ready
