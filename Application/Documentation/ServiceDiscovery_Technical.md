# Service Discovery and Failover System - Technical Documentation

**Version:** 1.0.0
**Date:** 2025-10-10
**Status:** Production Ready

## Table of Contents

1. [Overview](#overview)
2. [Architecture](#architecture)
3. [Security Model](#security-model)
4. [API Reference](#api-reference)
5. [Database Schema](#database-schema)
6. [Failover Mechanism](#failover-mechanism)
7. [Health Checking](#health-checking)
8. [Service Rotation](#service-rotation)
9. [Configuration](#configuration)
10. [Deployment](#deployment)
11. [Monitoring](#monitoring)
12. [Troubleshooting](#troubleshooting)

---

## Overview

The Service Discovery and Failover System provides a production-ready solution for dynamic service registration, health monitoring, automatic failover/failback, and secure service rotation in distributed microservice architectures.

### Key Features

- **Dynamic Service Registration:** Services can register themselves at runtime with cryptographic verification
- **Automatic Health Monitoring:** Background process continuously monitors service health
- **Automatic Failover:** Primary services automatically fail over to backup services when unhealthy
- **Automatic Failback:** Primary services automatically resume operation when recovered and stable
- **Secure Service Rotation:** Multi-layer security prevents malicious service injection
- **Audit Logging:** Complete audit trail for all service operations
- **Priority-Based Selection:** Services are selected based on priority and health metrics

### Supported Service Types

- `authentication` - Authentication service
- `permissions` - Permissions/authorization service
- `lokalisation` - Localization service
- `extension` - Extension services

---

## Architecture

### Component Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                     API Layer (Gin Framework)                    │
│  /api/services/register | discover | rotate | decommission      │
└──────────────────────────┬──────────────────────────────────────┘
                           │
┌──────────────────────────▼──────────────────────────────────────┐
│              ServiceDiscoveryHandler                             │
│  - Handles all API requests                                      │
│  - Validates input                                               │
│  - Coordinates with other components                             │
└──────────────────┬───────────────────────┬──────────────────────┘
                   │                       │
        ┌──────────▼──────────┐  ┌────────▼─────────┐
        │  ServiceSigner      │  │  HealthChecker   │
        │  - RSA 2048-bit     │  │  - Background    │
        │  - Sign/Verify      │  │  - HTTP checks   │
        │  - Multi-layer      │  │  - Parallel      │
        └──────────┬──────────┘  └────────┬─────────┘
                   │                       │
                   │             ┌─────────▼──────────┐
                   │             │  FailoverManager   │
                   │             │  - Auto failover   │
                   │             │  - Auto failback   │
                   │             │  - Stability checks│
                   │             └─────────┬──────────┘
                   │                       │
        ┌──────────▼───────────────────────▼──────────┐
        │         SQLite/PostgreSQL Database           │
        │  service_registry | service_health_check     │
        │  service_failover_events | audit logs        │
        └──────────────────────────────────────────────┘
```

### Data Flow

**1. Service Registration:**
```
Client → Handler → ServiceSigner (sign) → Database → HealthChecker (immediate check)
```

**2. Service Discovery:**
```
Client → Handler → Database (query by type, health, priority) → Client
```

**3. Health Monitoring (Background):**
```
HealthChecker (every 1 min) → HTTP GET to service → Database (record)
  → FailoverManager (check if failover/failback needed) → Database (update active state)
```

**4. Service Rotation:**
```
Client → Handler → Verify old service → ServiceSigner (verify new)
  → Verify admin token → Decommission old → Register new → Database
```

---

## Security Model

### Cryptographic Signatures

All services must be cryptographically signed using RSA 2048-bit keys.

**Signature Process:**
1. Generate RSA 2048-bit key pair
2. Create signature data: `ID|Name|Type|Version|URL|PublicKey|RegisteredAt`
3. Hash data using SHA-256
4. Sign hash using RSA PKCS#1 v1.5
5. Base64 encode signature

**Verification Process:**
1. Parse public key from service registration
2. Decode base64 signature
3. Recompute data hash
4. Verify signature using RSA public key

### Admin Token Verification

Admin operations require tokens with minimum 32 characters. In production, integrate with:
- JWT service for user authentication
- Role-based access control (RBAC)
- Audit logging for all admin operations

### Service Rotation Security Layers

1. **Service State Verification:** Old service must not be rotating or decommissioned
2. **Signature Verification:** New service must have valid cryptographic signature
3. **Admin Token Verification:** Token must be valid and have sufficient permissions
4. **Type Matching:** New service must match old service type
5. **Health Verification:** New service must be healthy
6. **Time Constraints:** Service must be registered for at least 5 minutes

### Attack Prevention

- **Malicious Service Injection:** Prevented by cryptographic signatures
- **Man-in-the-Middle:** Prevented by signature verification
- **Replay Attacks:** Prevented by timestamp validation
- **Rapid Rotation:** Prevented by 5-minute minimum registration time
- **Unauthorized Operations:** Prevented by admin token verification

---

## API Reference

### 1. Register Service

**Endpoint:** `POST /api/services/register`

**Request Body:**
```json
{
  "name": "Auth Service Primary",
  "type": "authentication",
  "version": "1.0.0",
  "url": "http://auth-primary:8081",
  "health_check_url": "http://auth-primary:8081/health",
  "public_key": "-----BEGIN PUBLIC KEY-----\nMIIBIjANBg...",
  "certificate": "-----BEGIN CERTIFICATE-----\nMIIDXTCCA...",
  "role": "primary",
  "failover_group": "auth-group-1",
  "priority": 10,
  "metadata": "{}",
  "admin_token": "secure-admin-token-at-least-32-characters"
}
```

**Response (201 Created):**
```json
{
  "errorCode": -1,
  "errorMessage": "",
  "data": {
    "service_id": "uuid-generated-id",
    "status": "registering",
    "registered_at": "2025-10-10T10:00:00Z"
  }
}
```

**Security:**
- Admin token required (minimum 32 characters)
- Public key must be valid RSA public key in PEM format
- Service will receive immediate health check

---

### 2. Discover Services

**Endpoint:** `POST /api/services/discover`

**Request Body:**
```json
{
  "type": "authentication",
  "min_version": "1.0.0",
  "only_healthy": true
}
```

**Response (200 OK):**
```json
{
  "services": [
    {
      "id": "service-uuid",
      "name": "Auth Service Primary",
      "type": "authentication",
      "version": "1.1.0",
      "url": "http://auth-primary:8081",
      "health_check_url": "http://auth-primary:8081/health",
      "status": "healthy",
      "role": "primary",
      "failover_group": "auth-group-1",
      "is_active": true,
      "priority": 10,
      "last_health_check": "2025-10-10T10:05:00Z"
    }
  ],
  "total_count": 1,
  "timestamp": "2025-10-10T10:06:00Z"
}
```

**Notes:**
- Services are ordered by: priority DESC, health_check_count DESC
- Only active services are returned by default
- Version filtering uses semantic versioning

---

### 3. Rotate Service

**Endpoint:** `POST /api/services/rotate`

**Request Body:**
```json
{
  "current_service_id": "old-service-uuid",
  "new_service": {
    "name": "Auth Service v2",
    "type": "authentication",
    "version": "2.0.0",
    "url": "http://auth-v2:8082",
    "health_check_url": "http://auth-v2:8082/health",
    "public_key": "-----BEGIN PUBLIC KEY-----\n...",
    "status": "healthy",
    "role": "primary",
    "priority": 10,
    "metadata": "{}",
    "registered_at": "2025-10-10T09:50:00Z"
  },
  "reason": "Upgrade to version 2.0.0",
  "requested_by": "admin",
  "admin_token": "secure-admin-token-at-least-32-characters",
  "verification_code": "optional-verification-code"
}
```

**Response (200 OK):**
```json
{
  "success": true,
  "old_service_id": "old-service-uuid",
  "new_service_id": "new-service-uuid",
  "rotation_time": "2025-10-10T10:10:00Z",
  "verification_hash": "sha256-hash-for-audit",
  "message": "Service rotated successfully"
}
```

**Security Checks:**
1. Old service exists and can be rotated
2. New service signature is valid
3. Admin token is valid
4. Service types match
5. New service is healthy
6. At least 5 minutes since old service registration

---

### 4. Decommission Service

**Endpoint:** `POST /api/services/decommission`

**Request Body:**
```json
{
  "service_id": "service-uuid",
  "reason": "End of life",
  "admin_token": "secure-admin-token-at-least-32-characters"
}
```

**Response (200 OK):**
```json
{
  "errorCode": -1,
  "errorMessage": "",
  "message": "Service decommissioned successfully"
}
```

**Effect:**
- Service status changed to `decommissioned`
- Service removed from discovery results
- Health checks continue for audit purposes
- Service can be reactivated if needed

---

### 5. Update Service

**Endpoint:** `POST /api/services/update`

**Request Body:**
```json
{
  "service_id": "service-uuid",
  "version": "1.0.1",
  "url": "http://auth-primary-new:8081",
  "health_check_url": "http://auth-primary-new:8081/health",
  "priority": 15,
  "metadata": "{\"region\": \"us-east-1\"}",
  "admin_token": "secure-admin-token-at-least-32-characters"
}
```

**Response (200 OK):**
```json
{
  "errorCode": -1,
  "errorMessage": "",
  "message": "Service updated successfully"
}
```

**Notes:**
- Only specified fields are updated
- Triggers immediate health check
- Admin token required

---

### 6. List All Services

**Endpoint:** `GET /api/services/list`

**Query Parameters:**
- `include_deleted=true` - Include soft-deleted services
- `include_decommissioned=true` - Include decommissioned services

**Response (200 OK):**
```json
{
  "services": [...],
  "total_count": 10,
  "timestamp": "2025-10-10T10:15:00Z"
}
```

---

### 7. Get Service Health

**Endpoint:** `GET /api/services/health/:id`

**Query Parameters:**
- `limit=10` - Number of recent health checks to return

**Response (200 OK):**
```json
{
  "service_id": "service-uuid",
  "current_status": "healthy",
  "last_check": "2025-10-10T10:14:00Z",
  "health_check_count": 150,
  "failed_health_count": 0,
  "recent_checks": [
    {
      "id": "check-uuid",
      "timestamp": "2025-10-10T10:14:00Z",
      "status": "healthy",
      "response_time": 45,
      "status_code": 200,
      "error_message": ""
    }
  ]
}
```

---

## Database Schema

### service_registry

Primary table for service registration.

```sql
CREATE TABLE service_registry (
  id TEXT PRIMARY KEY,                    -- UUID
  name TEXT NOT NULL,
  type TEXT NOT NULL,                     -- authentication, permissions, etc.
  version TEXT NOT NULL,                  -- Semantic version
  url TEXT NOT NULL,                      -- Service base URL
  health_check_url TEXT NOT NULL,         -- Health check endpoint
  public_key TEXT NOT NULL,               -- RSA public key (PEM)
  signature TEXT NOT NULL,                -- Service signature (base64)
  certificate TEXT,                       -- TLS certificate (optional)
  status TEXT NOT NULL DEFAULT 'registering',  -- healthy, unhealthy, etc.
  role TEXT NOT NULL DEFAULT 'primary',   -- primary or backup
  failover_group TEXT,                    -- Failover group identifier
  is_active INTEGER DEFAULT 1,            -- Currently active (1/0)
  priority INTEGER DEFAULT 0,             -- Higher = preferred
  metadata TEXT DEFAULT '{}',             -- JSON metadata
  registered_by TEXT NOT NULL,            -- Username
  registered_at INTEGER NOT NULL,         -- Unix timestamp
  last_health_check INTEGER DEFAULT 0,
  health_check_count INTEGER DEFAULT 0,
  failed_health_count INTEGER DEFAULT 0,
  last_failover_at INTEGER DEFAULT 0,
  deleted INTEGER DEFAULT 0,              -- Soft delete flag
  UNIQUE(name, type, url)
);

CREATE INDEX idx_service_registry_type ON service_registry(type);
CREATE INDEX idx_service_registry_status ON service_registry(status);
CREATE INDEX idx_service_registry_deleted ON service_registry(deleted);
CREATE INDEX idx_service_registry_type_status ON service_registry(type, status, deleted);
CREATE INDEX idx_service_registry_priority ON service_registry(priority DESC);
CREATE INDEX idx_service_registry_failover_group ON service_registry(failover_group);
CREATE INDEX idx_service_registry_is_active ON service_registry(is_active);
CREATE INDEX idx_service_registry_role ON service_registry(role);
CREATE INDEX idx_service_registry_group_active ON service_registry(failover_group, is_active, deleted);
```

### service_health_check

Records all health check results.

```sql
CREATE TABLE service_health_check (
  id TEXT PRIMARY KEY,
  service_id TEXT NOT NULL,
  timestamp INTEGER NOT NULL,
  status TEXT NOT NULL,
  response_time INTEGER NOT NULL,    -- Milliseconds
  status_code INTEGER NOT NULL,      -- HTTP status code
  error_message TEXT,
  checked_by TEXT NOT NULL,          -- System or username
  FOREIGN KEY(service_id) REFERENCES service_registry(id)
);

CREATE INDEX idx_health_check_service ON service_health_check(service_id);
CREATE INDEX idx_health_check_timestamp ON service_health_check(timestamp DESC);
CREATE INDEX idx_health_check_status ON service_health_check(status);
```

### service_failover_events

Records all failover and failback events.

```sql
CREATE TABLE service_failover_events (
  id TEXT PRIMARY KEY,
  failover_group TEXT NOT NULL,
  service_type TEXT NOT NULL,
  old_service_id TEXT NOT NULL,
  new_service_id TEXT NOT NULL,
  failover_reason TEXT NOT NULL,
  failover_type TEXT NOT NULL,       -- "failover" or "failback"
  timestamp INTEGER NOT NULL,
  automatic INTEGER NOT NULL,        -- 1 for automatic, 0 for manual
  FOREIGN KEY(old_service_id) REFERENCES service_registry(id),
  FOREIGN KEY(new_service_id) REFERENCES service_registry(id)
);

CREATE INDEX idx_failover_events_group ON service_failover_events(failover_group);
CREATE INDEX idx_failover_events_type ON service_failover_events(service_type);
CREATE INDEX idx_failover_events_timestamp ON service_failover_events(timestamp DESC);
```

### service_rotation_audit

Audit trail for service rotations.

```sql
CREATE TABLE service_rotation_audit (
  id TEXT PRIMARY KEY,
  old_service_id TEXT NOT NULL,
  new_service_id TEXT NOT NULL,
  reason TEXT,
  requested_by TEXT NOT NULL,
  rotation_time INTEGER NOT NULL,
  verification_hash TEXT NOT NULL,
  success INTEGER NOT NULL,
  error_message TEXT,
  FOREIGN KEY(old_service_id) REFERENCES service_registry(id),
  FOREIGN KEY(new_service_id) REFERENCES service_registry(id)
);

CREATE INDEX idx_rotation_audit_old_service ON service_rotation_audit(old_service_id);
CREATE INDEX idx_rotation_audit_new_service ON service_rotation_audit(new_service_id);
CREATE INDEX idx_rotation_audit_time ON service_rotation_audit(rotation_time DESC);
```

---

## Failover Mechanism

### Overview

The failover mechanism ensures high availability by automatically switching to backup services when primary services fail and switching back when primary services recover.

### Configuration

**Stability Check Count:** 3 consecutive healthy checks required before failback
**Failback Delay:** Minimum 5 minutes after failover before attempting failback
**Failure Threshold:** 3 consecutive failures before marking service unhealthy

### Failover Workflow

1. **Detection:**
   - Health checker detects primary service is unhealthy
   - Failure count reaches threshold (3 consecutive failures)
   - Service status changed to `unhealthy`

2. **Execution:**
   - Find best healthy backup service (by priority, health check count)
   - Deactivate primary service (`is_active = 0`)
   - Activate backup service (`is_active = 1`)
   - Record failover event
   - Log operation

3. **Rollback on Failure:**
   - If backup activation fails, reactivate primary
   - Log error for manual intervention

### Failback Workflow

1. **Detection:**
   - Primary service becomes healthy
   - Stability counter increments for each consecutive healthy check
   - Stability counter reaches threshold (3 checks)
   - At least 5 minutes have passed since failover

2. **Execution:**
   - Find currently active backup
   - Deactivate backup service
   - Activate primary service
   - Record failback event
   - Reset stability counter
   - Log operation

3. **Rollback on Failure:**
   - If primary activation fails, reactivate backup
   - Log error for manual intervention

### Failover Groups

Services in the same failover group share failover responsibility:

**Example Configuration:**
```
Group: auth-group-1
├── auth-primary (role: primary, priority: 10, active: true)
├── auth-backup-1 (role: backup, priority: 5, active: false)
└── auth-backup-2 (role: backup, priority: 3, active: false)
```

**Failover Priority:**
1. Highest priority healthy backup
2. If priorities are equal, most healthy service (by health_check_count)

---

## Health Checking

### Configuration

**Check Interval:** 1 minute
**Check Timeout:** 10 seconds
**Failure Threshold:** 3 consecutive failures
**Parallelization:** All services checked concurrently

### Health Check Process

1. Query all non-deleted, non-decommissioned services from database
2. For each service, spawn goroutine to check health in parallel
3. HTTP GET request to `health_check_url` with 10-second timeout
4. Evaluate response:
   - **Healthy:** HTTP 2xx or 3xx status code
   - **Unhealthy:** HTTP 4xx, 5xx, timeout, or connection error
5. Update service status in database
6. Record health check result
7. Trigger failover check if needed

### Health Check Data

Each health check records:
- Timestamp
- Status (healthy/unhealthy)
- Response time (milliseconds)
- HTTP status code
- Error message (if unhealthy)
- Checked by (system)

### Failure Count Management

- **On Success:** Reset failure count to 0
- **On Failure:** Increment failure count
- **Threshold:** Mark unhealthy at 3 consecutive failures
- **Recovery:** Mark healthy when check succeeds

---

## Service Rotation

### Use Cases

- Upgrade service to new version
- Replace compromised service
- Change service configuration
- Migrate to new infrastructure

### Prerequisites

1. New service must be registered for at least 5 minutes
2. New service must be healthy
3. New service must have valid cryptographic signature
4. Service types must match
5. Admin token must be valid
6. Old service must not be already rotating or decommissioned

### Rotation Process

1. **Validation:**
   - Verify admin token
   - Verify old service exists and can rotate
   - Verify new service signature
   - Verify service types match
   - Verify new service health
   - Verify time constraints

2. **Execution:**
   - Decommission old service
   - Register new service with same failover group
   - Transfer active state if applicable
   - Record rotation audit event

3. **Post-Rotation:**
   - Immediate health check on new service
   - Monitor new service for stability

### Security Considerations

- Use strong admin tokens (minimum 32 characters)
- Rotate admin tokens regularly
- Audit all rotation events
- Verify service signatures before rotation
- Implement rate limiting on rotation API

---

## Configuration

### Environment Variables

```bash
# Database
DB_TYPE=sqlite
DB_PATH=/app/Database/service_discovery.db

# Health Checking
HEALTH_CHECK_INTERVAL=60s
HEALTH_CHECK_TIMEOUT=10s
HEALTH_FAILURE_THRESHOLD=3

# Failover
FAILOVER_STABILITY_COUNT=3
FAILBACK_DELAY=5m

# Security
ADMIN_TOKEN_MIN_LENGTH=32
SERVICE_ROTATION_MIN_AGE=5m
```

### Server Configuration

```go
// internal/server/server.go
type Server struct {
    serviceDiscoveryHandler *handlers.ServiceDiscoveryHandler
    // ...
}

func NewServer(cfg *config.Config) (*Server, error) {
    // Initialize service discovery
    serviceDiscoveryHandler, err := handlers.NewServiceDiscoveryHandler(db)
    if err != nil {
        return nil, err
    }

    return &Server{
        serviceDiscoveryHandler: serviceDiscoveryHandler,
        // ...
    }, nil
}
```

---

## Deployment

### Production Checklist

- [ ] Enable TLS/HTTPS for all communication
- [ ] Configure strong admin tokens
- [ ] Set up database backups
- [ ] Enable audit logging
- [ ] Configure monitoring and alerting
- [ ] Set appropriate health check intervals
- [ ] Configure failover groups correctly
- [ ] Test failover/failback scenarios
- [ ] Document service architecture
- [ ] Set up incident response procedures

### Docker Deployment

```dockerfile
FROM golang:1.22-alpine AS builder
RUN apk add --no-cache gcc musl-dev sqlite-dev
WORKDIR /app
COPY . .
RUN CGO_ENABLED=1 go build -o htCore main.go

FROM alpine:latest
RUN apk add --no-cache sqlite-libs
COPY --from=builder /app/htCore /app/
COPY Configurations/ /app/Configurations/
RUN mkdir -p /app/Database
EXPOSE 8080
CMD ["/app/htCore", "-config", "/app/Configurations/default.json"]
```

### Kubernetes Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: helixtrack-core
spec:
  replicas: 3
  selector:
    matchLabels:
      app: helixtrack-core
  template:
    metadata:
      labels:
        app: helixtrack-core
    spec:
      containers:
      - name: core
        image: helixtrack-core:latest
        ports:
        - containerPort: 8080
        env:
        - name: DB_TYPE
          value: "postgresql"
        - name: HEALTH_CHECK_INTERVAL
          value: "60s"
        volumeMounts:
        - name: config
          mountPath: /app/Configurations
      volumes:
      - name: config
        configMap:
          name: helixtrack-config
```

---

## Monitoring

### Key Metrics

**Service Health:**
- Total services registered
- Healthy services count
- Unhealthy services count
- Average response time
- Health check success rate

**Failover:**
- Failover events (last hour/day)
- Failback events (last hour/day)
- Average failover duration
- Failover success rate

**API:**
- Request rate (by endpoint)
- Error rate
- Response time
- Registration rate

### Prometheus Metrics

```
# Service counts
service_discovery_total_services{type="authentication"} 2
service_discovery_healthy_services{type="authentication"} 2
service_discovery_unhealthy_services{type="authentication"} 0

# Health checks
service_discovery_health_checks_total 1500
service_discovery_health_check_failures_total 5

# Failover
service_discovery_failover_events_total 3
service_discovery_failback_events_total 2
```

### Logging

All operations are logged with structured logging using Zap:

```go
logger.Info("Service registered",
    zap.String("service_id", serviceID),
    zap.String("name", name),
    zap.String("type", serviceType),
)

logger.Warn("Service health check failed",
    zap.String("service_id", serviceID),
    zap.Int("failure_count", failureCount),
    zap.String("error", err.Error()),
)

logger.Error("Failover failed",
    zap.String("failover_group", group),
    zap.String("old_service", oldID),
    zap.Error(err),
)
```

---

## Troubleshooting

### Common Issues

#### 1. Service Registration Fails

**Symptom:** `400 Bad Request` or `401 Unauthorized`

**Possible Causes:**
- Invalid admin token (too short)
- Invalid public key format
- Missing required fields

**Solution:**
- Verify admin token is at least 32 characters
- Verify public key is valid RSA PEM format
- Check all required fields are provided

#### 2. Service Not Discovered

**Symptom:** Service registered but not returned by discovery endpoint

**Possible Causes:**
- Service is unhealthy
- Service is not active (`is_active = 0`)
- Service is deleted or decommissioned

**Solution:**
- Check service health status: `GET /api/services/health/:id`
- Check service registration: `GET /api/services/list`
- Verify health check endpoint is accessible

#### 3. Failover Not Triggered

**Symptom:** Primary service unhealthy but failover doesn't occur

**Possible Causes:**
- No healthy backup service available
- Backup service in different failover group
- Failure threshold not reached yet (< 3 failures)

**Solution:**
- Verify backup service exists and is healthy
- Check failover group configuration
- Check failure count: `GET /api/services/health/:id`

#### 4. Failback Not Triggered

**Symptom:** Primary service healthy but doesn't resume operation

**Possible Causes:**
- Stability threshold not reached (< 3 consecutive healthy checks)
- Insufficient time since failover (< 5 minutes)
- Primary service not in `primary` role

**Solution:**
- Wait for stability checks to complete
- Check time since last failover event
- Verify service role configuration

#### 5. Service Rotation Fails

**Symptom:** `400 Bad Request` during rotation

**Possible Causes:**
- New service registered too recently (< 5 minutes)
- Service type mismatch
- Invalid signature on new service
- Admin token invalid

**Solution:**
- Wait until new service has been registered for 5 minutes
- Verify service types match exactly
- Re-sign new service registration
- Verify admin token

### Debug Commands

```bash
# Check service status
curl http://localhost:8080/api/services/health/$SERVICE_ID

# List all services
curl http://localhost:8080/api/services/list

# Check failover history
sqlite3 /app/Database/service_discovery.db \
  "SELECT * FROM service_failover_events ORDER BY timestamp DESC LIMIT 10;"

# Check health check history
sqlite3 /app/Database/service_discovery.db \
  "SELECT * FROM service_health_check WHERE service_id='$SERVICE_ID' ORDER BY timestamp DESC LIMIT 10;"

# Check current active service
sqlite3 /app/Database/service_discovery.db \
  "SELECT * FROM service_registry WHERE failover_group='auth-group-1' AND is_active=1;"
```

### Support

For additional support:
- GitHub Issues: https://github.com/helixtrack/core/issues
- Documentation: https://docs.helixtrack.ru
- Email: support@helixtrack.ru

---

**End of Technical Documentation**
