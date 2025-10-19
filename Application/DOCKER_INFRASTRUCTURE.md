# HelixTrack Docker Infrastructure

Complete guide to HelixTrack's production-ready Docker infrastructure with automatic port selection, service discovery, load balancing, and encryption.

## Table of Contents

1. [Overview](#overview)
2. [Architecture](#architecture)
3. [Components](#components)
4. [Quick Start](#quick-start)
5. [Configuration](#configuration)
6. [Service Discovery](#service-discovery)
7. [Load Balancing](#load-balancing)
8. [Database Encryption](#database-encryption)
9. [Automatic Port Selection](#automatic-port-selection)
10. [Scaling and Rotation](#scaling-and-rotation)
11. [Monitoring](#monitoring)
12. [Testing](#testing)
13. [Troubleshooting](#troubleshooting)
14. [Production Deployment](#production-deployment)

## Overview

HelixTrack's Docker infrastructure provides:

- **🐳 Docker Compose Orchestration** - Single command to start entire stack
- **🔄 Automatic Port Selection** - Services automatically find available ports (8080-8089)
- **🔍 Service Discovery** - Consul-based service registry
- **⚖️ Load Balancing** - HAProxy with health checks and failover
- **🔐 Database Encryption** - PostgreSQL with SSL/TLS and pgcrypto
- **📈 Horizontal Scaling** - Scale services with `docker-compose scale`
- **🔄 Service Rotation** - Zero-downtime deployments
- **💚 Health Monitoring** - Comprehensive health checks
- **📊 Metrics & Monitoring** - Prometheus + Grafana
- **🧪 Comprehensive Testing** - 35+ automated infrastructure tests

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                         External Traffic                         │
└──────────────────────────┬──────────────────────────────────────┘
                           │
                           ▼
               ┌───────────────────────┐
               │    HAProxy (80/443)    │
               │   Load Balancer        │
               └───────────┬───────────┘
                           │
          ┌────────────────┼────────────────┐
          │                │                │
          ▼                ▼                ▼
    ┌─────────┐      ┌─────────┐      ┌─────────┐
    │  Core   │      │  Core   │      │  Core   │
    │ :8080   │      │ :8081   │      │ :8082   │
    └────┬────┘      └────┬────┘      └────┬────┘
         │                │                │
         └────────────────┼────────────────┘
                          │
         ┌────────────────┼────────────────┐
         │                │                │
         ▼                ▼                ▼
    ┌─────────┐      ┌─────────┐      ┌─────────┐
    │  Auth   │      │  Perm   │      │  Docs   │
    │ Service │      │ Service │      │ Service │
    └────┬────┘      └────┬────┘      └────┬────┘
         │                │                │
         └────────────────┼────────────────┘
                          │
         ┌────────────────┴────────────────┐
         │                                  │
         ▼                                  ▼
    ┌─────────┐                       ┌─────────┐
    │ Consul  │                       │PostgreSQL│
    │ Service │                       │ Cluster │
    │Discovery│                       │ with SSL│
    └─────────┘                       └─────────┘
         │                                  │
         └──────────────┬───────────────────┘
                        │
                        ▼
              ┌──────────────────┐
              │   Monitoring     │
              │ Prometheus       │
              │ + Grafana        │
              └──────────────────┘
```

## Components

### Core Services

1. **Core Service** (`core-service`)
   - Main HelixTrack API
   - Ports: 8080-8089 (automatic selection)
   - Health check: `/health`
   - Scalable: Yes

2. **Authentication Service** (`auth-service`)
   - JWT authentication
   - Port: 8081
   - Health check: `/health`
   - Scalable: Yes

3. **Permissions Service** (`perm-service`)
   - RBAC authorization
   - Port: 8082
   - Health check: `/health`
   - Scalable: Yes

4. **Documents Service** (`documents-service`)
   - Confluence-style documents
   - Port: 8083
   - Health check: `/health`
   - Scalable: Yes
   - Profile: `extensions`

### Infrastructure Services

5. **PostgreSQL Databases**
   - Core DB, Auth DB, Perm DB, Docs DB
   - SSL/TLS encryption required
   - pgcrypto extension for data encryption
   - Automatic backups

6. **Consul** (`service-registry`)
   - Service discovery
   - Health monitoring
   - Configuration management
   - UI: `http://localhost:8500/ui`

7. **HAProxy** (`load-balancer`)
   - Load balancing
   - SSL termination
   - Health checks
   - Stats: `http://localhost:8404/stats`

8. **Prometheus** (`prometheus`)
   - Metrics collection
   - UI: `http://localhost:9091`
   - Profile: `monitoring`

9. **Grafana** (`grafana`)
   - Metrics visualization
   - Dashboards
   - UI: `http://localhost:3000`
   - Profile: `monitoring`

## Quick Start

### Start All Services

```bash
cd /path/to/HelixTrack/Core/Application

# Start core services only
./scripts/start-production.sh

# Start with monitoring
./scripts/start-production.sh --with-monitoring

# Start with all extensions
./scripts/start-production.sh --with-extensions

# Start with monitoring and extensions
./scripts/start-production.sh --with-monitoring --with-extensions --logs
```

### Stop All Services

```bash
# Graceful shutdown
./scripts/stop-production.sh

# Remove volumes (WARNING: deletes all data!)
./scripts/stop-production.sh --remove-volumes

# Full cleanup
./scripts/stop-production.sh --cleanup
```

### Check Status

```bash
# View all services
docker-compose -f docker-compose-production.yml ps

# Check health
curl http://localhost:8080/health  # Core
curl http://localhost:8500/v1/status/leader  # Consul
curl http://localhost:8404/stats  # HAProxy

# View logs
docker-compose -f docker-compose-production.yml logs -f core-service
```

## Configuration

### Environment Variables

Create `.env.production`:

```bash
# Build Configuration
BUILD_VERSION=1.0.0
BUILD_DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
BUILD_COMMIT=$(git rev-parse --short HEAD)

# Core Database
CORE_DB_NAME=helixtrack_core
CORE_DB_USER=helixtrack
CORE_DB_PASSWORD=CHANGE_ME_IN_PRODUCTION
CORE_DB_PORT=5432

# Core Service
CORE_PORT=8080
CORE_METRICS_PORT=9090
CORE_REPLICAS=3

# Automatic Port Selection
AUTO_PORT_SELECTION=true
SERVER_PORT_RANGE_START=8080
SERVER_PORT_RANGE_END=8089

# Service Discovery
SERVICE_DISCOVERY_ENABLED=true
SERVICE_REGISTRY_URL=http://service-registry:8500

# Security
JWT_SECRET=your-jwt-secret-key-change-in-production-minimum-32-characters
ENCRYPTION_KEY=your-encryption-key-change-in-production-minimum-32-chars

# Monitoring
GRAFANA_USER=admin
GRAFANA_PASSWORD=CHANGE_ME
```

⚠️ **IMPORTANT:** Change all passwords before production deployment!

### Docker Compose Profiles

```bash
# Core services only (default)
docker-compose up -d

# With monitoring
docker-compose --profile monitoring up -d

# With extensions
docker-compose --profile extensions up -d

# Everything
docker-compose --profile monitoring --profile extensions up -d
```

## Service Discovery

### How It Works

1. **Service Starts**
   - Finds available port (8080-8089)
   - Starts on selected port
   - Registers with Consul

2. **Registration**
   ```json
   {
     "ID": "hostname-8080",
     "Name": "helixtrack-core",
     "Address": "172.20.0.5",
     "Port": 8080,
     "Check": {
       "HTTP": "http://172.20.0.5:8080/health",
       "Interval": "30s"
     }
   }
   ```

3. **Discovery**
   - Other services query Consul
   - Get list of healthy instances
   - Connect directly or via load balancer

### Query Services

```bash
# List all services
curl http://localhost:8500/v1/catalog/services

# Get core service instances
curl http://localhost:8500/v1/catalog/service/helixtrack-core | jq

# Get only healthy instances
curl http://localhost:8500/v1/health/service/helixtrack-core?passing | jq

# DNS query
dig @localhost -p 8600 helixtrack-core.service.consul
```

### Manual Registration

```bash
# Register service
curl -X PUT \
  -d '{
    "ID": "core-8080",
    "Name": "helixtrack-core",
    "Port": 8080,
    "Check": {
      "HTTP": "http://localhost:8080/health",
      "Interval": "30s"
    }
  }' \
  http://localhost:8500/v1/agent/service/register

# Deregister service
curl -X PUT http://localhost:8500/v1/agent/service/deregister/core-8080
```

## Load Balancing

### HAProxy Configuration

Load balancer distributes traffic using round-robin:

```
Client Request
     ↓
HAProxy (:80/:443)
     ↓
Health Checks
     ↓
Route to healthy backend
     ↓
Core Service (8080, 8081, or 8082)
```

### Access Points

- **HTTP**: `http://localhost/`
- **HTTPS**: `https://localhost/` (requires SSL cert)
- **Stats**: `http://localhost:8404/stats` (user: `admin`, pass: `admin`)
- **Health**: `http://localhost:8405/health`

### Adding Backends

Edit `docker/haproxy/haproxy.cfg`:

```haproxy
backend core_services
    server core-1 core-service:8080 check
    server core-2 core-service:8081 check
    server core-3 core-service:8082 check
```

Reload HAProxy:

```bash
docker-compose exec load-balancer kill -USR2 1
```

## Database Encryption

### SSL/TLS Connections

All database connections REQUIRE SSL:

```bash
# Connection string
postgresql://user:pass@host:5432/db?sslmode=require

# Environment variable
DATABASE_SSL_MODE=require
```

### Column Encryption (pgcrypto)

```sql
-- Encrypt data
INSERT INTO users (email, encrypted_ssn)
VALUES ('user@example.com', encrypt_text('123-45-6789', 'key'));

-- Decrypt data
SELECT decrypt_text(encrypted_ssn, 'key') AS ssn FROM users;

-- Hash password
INSERT INTO users (username, password_hash)
VALUES ('alice', hash_password('secret123'));

-- Verify password
SELECT verify_password('secret123', password_hash) FROM users;
```

See [docker/postgres/README.md](docker/postgres/README.md) for complete encryption documentation.

## Automatic Port Selection

### How It Works

```bash
# In entrypoint.sh

find_available_port() {
    for port in $(seq $SERVER_PORT_RANGE_START $SERVER_PORT_RANGE_END); do
        if ! nc -z localhost $port 2>/dev/null; then
            echo "$port"
            return 0
        fi
    done
    return 1
}

SELECTED_PORT=$(find_available_port)
```

### Configuration

```yaml
# docker-compose-production.yml
environment:
  - AUTO_PORT_SELECTION=true
  - SERVER_PORT_RANGE_START=8080
  - SERVER_PORT_RANGE_END=8089
```

### Port Exhaustion

If all ports are in use:

1. **Error logged**: "No available ports found in range"
2. **Service fails to start** (doesn't crash existing instances)
3. **Solution**: Increase range or scale down

```bash
# Expand port range
SERVER_PORT_RANGE_END=8099  # 20 ports instead of 10

# Or reduce instances
docker-compose up -d --scale core-service=5
```

## Scaling and Rotation

### Horizontal Scaling

```bash
# Scale core service to 5 instances
docker-compose up -d --scale core-service=5

# Scale authentication service
docker-compose up -d --scale auth-service=3

# Each instance gets unique port automatically
curl http://localhost:8500/v1/catalog/service/helixtrack-core | jq '.[].ServicePort'
# Output: 8080, 8081, 8082, 8083, 8084
```

### Zero-Downtime Deployment

```bash
# 1. Scale up with new version
docker-compose up -d --scale core-service=6 --build

# 2. Wait for health checks
sleep 30

# 3. Scale down old instances
docker-compose up -d --scale core-service=3

# Traffic seamlessly moves to new instances
```

### Service Rotation

Services automatically:

1. **Register** with Consul on startup
2. **Health checked** every 30 seconds
3. **Deregister** on graceful shutdown
4. **Auto-removed** if critical for 90+ seconds

```bash
# Stop one instance
docker stop <container-id>

# HAProxy automatically stops routing to it
# Other instances continue serving traffic
```

## Monitoring

### Prometheus Metrics

```bash
# Core service metrics
curl http://localhost:9090/metrics

# Consul metrics
curl http://localhost:8500/v1/agent/metrics?format=prometheus

# HAProxy metrics
curl http://localhost:8406/metrics
```

### Grafana Dashboards

1. Access: `http://localhost:3000`
2. Login: `admin` / `admin`
3. Pre-configured dashboards:
   - Service health
   - Request rates
   - Response times
   - Resource usage

### Health Checks

```bash
# Check all services
for service in core auth perm docs; do
  echo "=== $service ==="
  curl -s http://localhost:${port}/health | jq
done

# Consul health
curl http://localhost:8500/v1/health/state/any | jq

# HAProxy backend health
curl -u admin:admin http://localhost:8404/stats | grep backend
```

## Testing

### Run All Tests

```bash
# Docker infrastructure tests (35 tests)
./tests/docker-infrastructure/test-infrastructure.sh

# AI QA automation
./tests/ai-qa/run-ai-qa.sh

# Go unit tests
go test ./...

# API tests
cd test-scripts && ./test-all.sh
```

### Test Categories

1. **Infrastructure Tests** (35 tests)
   - Prerequisites (Docker, configs)
   - Service startup
   - Service discovery
   - Load balancing
   - Scaling and rotation
   - Failure scenarios

2. **AI QA Tests**
   - API discovery
   - Performance analysis
   - Anomaly detection
   - Security testing
   - Load testing

See [tests/docker-infrastructure/README.md](tests/docker-infrastructure/README.md) for details.

## Troubleshooting

### Common Issues

#### 1. Service Won't Start

```bash
# Check logs
docker-compose logs core-service

# Check resources
docker stats

# Validate configuration
docker-compose config
```

#### 2. Database Connection Failed

```bash
# Check database health
docker-compose exec core-db pg_isready

# Check network
docker network inspect helixtrack-network

# Verify credentials
docker-compose exec core-db psql -U helixtrack -d helixtrack_core -c "SELECT 1"
```

#### 3. Port Already in Use

```bash
# Find what's using the port
lsof -i :8080

# Stop conflicting service
docker stop <container>

# Or use different port range
export SERVER_PORT_RANGE_START=9080
export SERVER_PORT_RANGE_END=9089
```

#### 4. Consul Service Not Registering

```bash
# Check Consul
curl http://localhost:8500/v1/status/leader

# Check service logs
docker logs core-service | grep -i consul

# Manual registration
curl -X PUT \
  -d '{"Name":"helixtrack-core","Port":8080}' \
  http://localhost:8500/v1/agent/service/register
```

#### 5. HAProxy 503 Error

```bash
# Check backend health
curl http://localhost:8404/stats

# Check backends directly
curl http://core-service:8080/health

# Restart HAProxy
docker-compose restart load-balancer
```

See [tests/docker-infrastructure/FAILURE_SCENARIOS.md](tests/docker-infrastructure/FAILURE_SCENARIOS.md) for comprehensive failure analysis.

## Production Deployment

### Pre-Deployment Checklist

- [ ] Change all default passwords in `.env.production`
- [ ] Replace self-signed SSL certificates with CA-signed
- [ ] Configure firewall rules
- [ ] Set up automated backups
- [ ] Configure monitoring and alerting
- [ ] Test disaster recovery procedures
- [ ] Document custom configurations
- [ ] Set up log aggregation
- [ ] Configure resource limits appropriately
- [ ] Enable security features (ACLs, encryption)

### Production Configuration

```bash
# 1. Generate production environment file
cp .env.production.example .env.production
vi .env.production  # Edit all passwords and secrets

# 2. Generate SSL certificates
cd docker/postgres
./docker-entrypoint-initdb.d/00-generate-ssl-certs.sh

# For HAProxy (use real CA certificates in production)
openssl req -x509 -newkey rsa:4096 -nodes \
  -keyout helixtrack.key -out helixtrack.crt -days 365
cat helixtrack.crt helixtrack.key > docker/haproxy/certs/helixtrack.pem

# 3. Build production images
docker-compose -f docker-compose-production.yml build

# 4. Start services
./scripts/start-production.sh --with-monitoring

# 5. Verify health
./tests/docker-infrastructure/test-infrastructure.sh
```

### High Availability Setup

```yaml
# docker-compose-production.yml
services:
  core-service:
    deploy:
      replicas: 5
      restart_policy:
        condition: on-failure
        max_attempts: 3
      resources:
        limits:
          cpus: '1.0'
          memory: 1G
        reservations:
          cpus: '0.5'
          memory: 512M
```

### Backup Strategy

```bash
# Database backups
docker-compose exec core-db pg_dump -U helixtrack helixtrack_core > backup.sql

# Consul snapshots
docker-compose exec service-registry consul snapshot save consul-backup.snap

# Automated daily backups
0 2 * * * /path/to/backup-script.sh
```

### Security Hardening

1. **Enable ACLs in Consul**
2. **Use secrets management** (Vault, Docker Secrets)
3. **Enable audit logging**
4. **Restrict network access**
5. **Regular security updates**
6. **Use least privilege principles**

See [docs/DEPLOYMENT.md](docs/DEPLOYMENT.md) for complete production deployment guide.

## Resources

- [Docker Compose File](docker-compose-production.yml)
- [Start Script](scripts/start-production.sh)
- [Stop Script](scripts/stop-production.sh)
- [Entrypoint Script](docker/scripts/entrypoint.sh)
- [HAProxy Configuration](docker/haproxy/README.md)
- [Consul Configuration](docker/consul/README.md)
- [PostgreSQL Encryption](docker/postgres/README.md)
- [Infrastructure Tests](tests/docker-infrastructure/README.md)
- [Failure Scenarios](tests/docker-infrastructure/FAILURE_SCENARIOS.md)
- [AI QA Framework](tests/ai-qa/README.md)

## Support

For issues or questions:

1. Check troubleshooting section above
2. Review [FAILURE_SCENARIOS.md](tests/docker-infrastructure/FAILURE_SCENARIOS.md)
3. Run infrastructure tests
4. Check Docker logs
5. Open GitHub issue

## License

MIT License - Same as HelixTrack Core
