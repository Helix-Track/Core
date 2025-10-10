# Service Discovery and Failover - User Manual

**Version:** 1.0.0
**Date:** 2025-10-10
**Audience:** System Administrators, DevOps Engineers, Operators

## Table of Contents

1. [Introduction](#introduction)
2. [Getting Started](#getting-started)
3. [Registering a Service](#registering-a-service)
4. [Managing Services](#managing-services)
5. [Understanding Failover](#understanding-failover)
6. [Service Rotation](#service-rotation)
7. [Monitoring](#monitoring)
8. [Best Practices](#best-practices)
9. [FAQs](#faqs)
10. [Troubleshooting Guide](#troubleshooting-guide)

---

## Introduction

### What is Service Discovery?

Service Discovery is a system that allows microservices to find and communicate with each other automatically. Instead of hardcoding service locations, services can register themselves and discover other services dynamically.

### What is Failover?

Failover is the automatic process of switching to a backup service when the primary service fails. When the primary service recovers, the system automatically switches back (failback).

### Why Do We Need This?

**Without Service Discovery:**
- Manual configuration of service locations
- Difficult to scale services
- Downtime during service updates
- Manual failover procedures

**With Service Discovery:**
- Automatic service registration
- Dynamic scaling
- Zero-downtime deployments
- Automatic failover/failback
- Complete audit trail

---

## Getting Started

### Prerequisites

1. **Admin Token**: Obtain a secure admin token (minimum 32 characters) from your security team
2. **Service Details**: Know your service's URL and health check endpoint
3. **Network Access**: Ensure your service can be accessed from the core server
4. **TLS Certificate** (optional but recommended): For secure communication

### Quick Start Guide

**Step 1: Verify Core Service is Running**
```bash
curl http://localhost:8080/health
```
Expected response: `{"status": "healthy"}`

**Step 2: Register Your Service**
```bash
curl -X POST http://localhost:8080/api/services/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "My Auth Service",
    "type": "authentication",
    "version": "1.0.0",
    "url": "http://my-auth:8081",
    "health_check_url": "http://my-auth:8081/health",
    "role": "primary",
    "priority": 10,
    "admin_token": "your-secure-admin-token-here"
  }'
```

**Step 3: Verify Registration**
```bash
curl -X POST http://localhost:8080/api/services/discover \
  -H "Content-Type: application/json" \
  -d '{"type": "authentication", "only_healthy": true}'
```

---

## Registering a Service

### Basic Registration

To register a service, you need to provide:

| Field | Description | Example |
|-------|-------------|---------|
| name | Friendly name for your service | "Auth Service Primary" |
| type | Service type | "authentication" |
| version | Semantic version | "1.0.0" |
| url | Base URL of your service | "http://auth:8081" |
| health_check_url | Health check endpoint | "http://auth:8081/health" |
| role | Primary or backup | "primary" |
| priority | Higher = preferred (0-100) | 10 |
| admin_token | Your admin token | "secure-token-32-chars-min" |

### Service Types

- `authentication` - Authentication services
- `permissions` - Permission/authorization services
- `lokalisation` - Localization services
- `extension` - Extension services

### Service Roles

**Primary:**
- Main service handling production traffic
- Active by default
- Will failover to backup if unhealthy

**Backup:**
- Standby service
- Inactive by default
- Automatically activated during failover

### Priority Levels

Priority determines which service is preferred when multiple services are available:

- **High Priority (50-100):** Production services
- **Medium Priority (20-49):** Staging services
- **Low Priority (0-19):** Development services

### Complete Registration Example

```bash
curl -X POST http://localhost:8080/api/services/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Auth Service Primary",
    "type": "authentication",
    "version": "1.2.0",
    "url": "http://auth-primary:8081",
    "health_check_url": "http://auth-primary:8081/health",
    "public_key": "-----BEGIN PUBLIC KEY-----\nMIIBIjANBg...",
    "role": "primary",
    "failover_group": "auth-group-1",
    "priority": 75,
    "metadata": "{\"region\": \"us-east-1\", \"datacenter\": \"dc1\"}",
    "admin_token": "your-secure-admin-token-at-least-32-characters"
  }'
```

### Response

**Success (201 Created):**
```json
{
  "errorCode": -1,
  "errorMessage": "",
  "data": {
    "service_id": "abc-123-def-456",
    "status": "registering",
    "registered_at": "2025-10-10T10:00:00Z"
  }
}
```

**Error (400 Bad Request):**
```json
{
  "errorCode": 1001,
  "errorMessage": "Admin token must be at least 32 characters",
  "data": null
}
```

---

## Managing Services

### Discovering Services

Find services by type:

```bash
curl -X POST http://localhost:8080/api/services/discover \
  -H "Content-Type: application/json" \
  -d '{
    "type": "authentication",
    "only_healthy": true
  }'
```

**Response:**
```json
{
  "services": [
    {
      "id": "service-123",
      "name": "Auth Service Primary",
      "type": "authentication",
      "version": "1.2.0",
      "url": "http://auth-primary:8081",
      "status": "healthy",
      "role": "primary",
      "is_active": true,
      "priority": 75,
      "last_health_check": "2025-10-10T10:05:00Z"
    }
  ],
  "total_count": 1,
  "timestamp": "2025-10-10T10:06:00Z"
}
```

### Listing All Services

```bash
curl http://localhost:8080/api/services/list
```

### Checking Service Health

```bash
curl http://localhost:8080/api/services/health/service-123
```

**Response:**
```json
{
  "service_id": "service-123",
  "current_status": "healthy",
  "last_check": "2025-10-10T10:14:00Z",
  "health_check_count": 150,
  "failed_health_count": 0,
  "recent_checks": [
    {
      "timestamp": "2025-10-10T10:14:00Z",
      "status": "healthy",
      "response_time": 45,
      "status_code": 200
    }
  ]
}
```

### Updating Service Details

```bash
curl -X POST http://localhost:8080/api/services/update \
  -H "Content-Type: application/json" \
  -d '{
    "service_id": "service-123",
    "priority": 80,
    "metadata": "{\"region\": \"us-west-2\"}",
    "admin_token": "your-admin-token"
  }'
```

### Decommissioning a Service

When a service is no longer needed:

```bash
curl -X POST http://localhost:8080/api/services/decommission \
  -H "Content-Type: application/json" \
  -d '{
    "service_id": "service-123",
    "reason": "Service migrated to new platform",
    "admin_token": "your-admin-token"
  }'
```

**What happens:**
- Service status changes to "decommissioned"
- Service no longer appears in discovery results
- Health checks stop
- Service can be reactivated if needed

---

## Understanding Failover

### What is Failover?

Failover is the automatic process of switching from a failed primary service to a healthy backup service.

### How It Works

```
┌──────────────────────────────────────────────────────┐
│ Normal Operation: Primary service is active          │
│                                                       │
│  [Primary] ✓ active ────► Clients                   │
│  [Backup]  ✗ inactive                                │
└──────────────────────────────────────────────────────┘

┌──────────────────────────────────────────────────────┐
│ Failure Detected: Primary becomes unhealthy          │
│                                                       │
│  [Primary] ✗ unhealthy                               │
│  [Backup]  ✓ becoming active...                      │
└──────────────────────────────────────────────────────┘

┌──────────────────────────────────────────────────────┐
│ Failover Complete: Backup now serves traffic         │
│                                                       │
│  [Primary] ✗ inactive (recovering)                   │
│  [Backup]  ✓ active ────► Clients                   │
└──────────────────────────────────────────────────────┘

┌──────────────────────────────────────────────────────┐
│ Failback: Primary recovered and stable               │
│                                                       │
│  [Primary] ✓ active ────► Clients                   │
│  [Backup]  ✗ inactive                                │
└──────────────────────────────────────────────────────┘
```

### Failover Configuration

**Failure Threshold:** 3 consecutive failed health checks
**Stability Requirement:** 3 consecutive successful health checks
**Failback Delay:** 5 minutes minimum after failover
**Health Check Interval:** Every 1 minute

### Setting Up Failover

**Step 1: Register Primary Service**
```bash
curl -X POST http://localhost:8080/api/services/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Auth Service Primary",
    "type": "authentication",
    "url": "http://auth-primary:8081",
    "health_check_url": "http://auth-primary:8081/health",
    "role": "primary",
    "failover_group": "auth-group-1",
    "priority": 10,
    "admin_token": "your-token"
  }'
```

**Step 2: Register Backup Service**
```bash
curl -X POST http://localhost:8080/api/services/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Auth Service Backup",
    "type": "authentication",
    "url": "http://auth-backup:8082",
    "health_check_url": "http://auth-backup:8082/health",
    "role": "backup",
    "failover_group": "auth-group-1",
    "priority": 5,
    "admin_token": "your-token"
  }'
```

**Important:** Both services must have the same `failover_group` and `type`.

### Monitoring Failover Events

Check recent failover activity:

```bash
# Get failover history (via database)
sqlite3 /app/Database/service_discovery.db \
  "SELECT * FROM service_failover_events ORDER BY timestamp DESC LIMIT 10;"
```

**Sample Output:**
```
id|failover_group|service_type|old_service_id|new_service_id|failover_reason|failover_type|timestamp|automatic
evt-1|auth-group-1|authentication|primary-123|backup-456|Primary unhealthy|failover|1696934400|1
evt-2|auth-group-1|authentication|backup-456|primary-123|Primary recovered|failback|1696938000|1
```

---

## Service Rotation

### What is Service Rotation?

Service rotation is the process of replacing one service with another. Common use cases:

- **Upgrading:** Replace v1.0 with v2.0
- **Security:** Replace potentially compromised service
- **Migration:** Move to new infrastructure
- **Configuration Change:** Apply new settings

### When to Rotate

✅ **Good Reasons to Rotate:**
- New version available
- Security vulnerability fixed
- Performance improvements
- Infrastructure migration

❌ **Bad Reasons to Rotate:**
- Service temporarily slow (wait for recovery)
- Testing changes (use development environment)
- Frequent changes (causes instability)

### Rotation Process

**Step 1: Prepare New Service**
1. Deploy new service
2. Verify health check works
3. Wait at least 5 minutes after deployment
4. Ensure service is marked healthy

**Step 2: Perform Rotation**
```bash
curl -X POST http://localhost:8080/api/services/rotate \
  -H "Content-Type: application/json" \
  -d '{
    "current_service_id": "old-service-123",
    "new_service": {
      "name": "Auth Service v2",
      "type": "authentication",
      "version": "2.0.0",
      "url": "http://auth-v2:8083",
      "health_check_url": "http://auth-v2:8083/health",
      "public_key": "-----BEGIN PUBLIC KEY-----\n...",
      "status": "healthy",
      "role": "primary",
      "priority": 10,
      "metadata": "{}",
      "registered_at": "2025-10-10T09:50:00Z"
    },
    "reason": "Upgrade to version 2.0.0",
    "requested_by": "admin@example.com",
    "admin_token": "your-admin-token",
    "verification_code": "optional-code"
  }'
```

**Step 3: Verify Rotation**
```bash
# Check new service is active
curl -X POST http://localhost:8080/api/services/discover \
  -d '{"type": "authentication", "only_healthy": true}'

# Check old service is decommissioned
curl http://localhost:8080/api/services/health/old-service-123
```

### Rotation Safety Checks

The system performs these checks before allowing rotation:

1. ✓ Old service exists and can be rotated
2. ✓ New service has valid cryptographic signature
3. ✓ Admin token is valid
4. ✓ Service types match (can't replace auth with permissions)
5. ✓ New service is healthy
6. ✓ At least 5 minutes since new service registered

If any check fails, rotation is blocked.

### Rollback Plan

If rotation causes issues:

1. **Immediate:** Decommission new service
2. **Reactivate:** Register old service again
3. **Investigate:** Check logs for errors
4. **Fix:** Address issues before retrying

---

## Monitoring

### Health Check Dashboard

**Check Service Health:**
```bash
# Single service
curl http://localhost:8080/api/services/health/service-123

# All services
curl http://localhost:8080/api/services/list
```

### Key Metrics to Monitor

**1. Service Availability**
- Total registered services
- Healthy vs unhealthy count
- Average response time

**2. Failover Activity**
- Failover events in last 24 hours
- Average failover duration
- Failback success rate

**3. Health Check Status**
- Success rate
- Average response time
- Failed checks

### Setting Up Alerts

**Recommended Alerts:**

**Critical:**
- Service unhealthy for > 5 minutes
- Failover occurred
- No backup service available
- Rotation failed

**Warning:**
- Service response time > 1 second
- Failed health check (1-2 failures)
- Low number of healthy services

**Info:**
- Service registered
- Service decommissioned
- Failback completed

### Example Alert Configuration (Prometheus)

```yaml
groups:
- name: service_discovery
  rules:
  - alert: ServiceUnhealthy
    expr: service_discovery_unhealthy_services > 0
    for: 5m
    annotations:
      summary: "Service {{ $labels.service_name }} is unhealthy"

  - alert: FailoverOccurred
    expr: increase(service_discovery_failover_events_total[5m]) > 0
    annotations:
      summary: "Failover occurred in group {{ $labels.failover_group }}"
```

---

## Best Practices

### Service Registration

✅ **DO:**
- Use descriptive service names
- Set appropriate priorities
- Provide complete metadata
- Use semantic versioning
- Include health check endpoints

❌ **DON'T:**
- Register duplicate services
- Use production tokens in development
- Skip health check endpoints
- Use generic names

### Health Check Endpoints

Your service should implement a `/health` endpoint that returns:

```json
{
  "status": "healthy",
  "timestamp": "2025-10-10T10:00:00Z",
  "version": "1.0.0",
  "dependencies": {
    "database": "connected",
    "cache": "connected"
  }
}
```

**Requirements:**
- Response time < 1 second
- HTTP 200 status code when healthy
- HTTP 503 status code when unhealthy
- Include dependency status
- No authentication required

### Failover Groups

**Best Practices:**
- One primary + one or more backups per group
- Backups in different availability zones
- Similar capacity between primary and backup
- Test failover regularly
- Monitor failover events

**Example Configuration:**
```
failover_group: "auth-prod"
├── auth-primary (priority: 100, zone: us-east-1a)
├── auth-backup-1 (priority: 50, zone: us-east-1b)
└── auth-backup-2 (priority: 25, zone: us-west-2a)
```

### Security

**Admin Tokens:**
- Minimum 32 characters
- Rotate every 90 days
- Store securely (e.g., HashiCorp Vault)
- Limit access
- Audit all usage

**Service Authentication:**
- Use TLS/HTTPS in production
- Implement service-to-service authentication
- Rotate service certificates regularly
- Monitor for suspicious activity

### Operational Procedures

**Daily:**
- Check service health dashboard
- Review failed health checks
- Monitor response times

**Weekly:**
- Review failover events
- Check backup service health
- Verify monitoring alerts work

**Monthly:**
- Test failover scenarios
- Review and update priorities
- Audit admin token usage
- Update documentation

---

## FAQs

### Q: How long does failover take?

**A:** Typically 1-3 minutes:
- 3 minutes for health checks to detect failure (3 failures × 1 minute interval)
- < 1 second for failover execution
- Clients may experience brief errors during transition

### Q: Can I have multiple backup services?

**A:** Yes! Register multiple backup services in the same failover group. The system will choose the highest priority healthy backup.

### Q: What happens if both primary and backup fail?

**A:** The system will:
1. Mark both as unhealthy
2. Remove from discovery results
3. Log critical alert
4. Require manual intervention

### Q: Can I manually trigger failover?

**A:** Yes, decommission the primary service and the system will automatically failover to the backup.

### Q: How do I test failover without affecting production?

**A:**
1. Create a test failover group
2. Register test services
3. Simulate failures by stopping services
4. Observe automatic failover
5. Verify failback after recovery

### Q: Can services in different data centers failover?

**A:** Yes! As long as:
- Both services have the same failover group
- Network connectivity exists between data centers
- Health check endpoint is accessible

### Q: What happens during service rotation?

**A:**
1. Old service is decommissioned
2. New service is registered with same group
3. Active state is transferred (if applicable)
4. Clients discover new service
5. Old service is removed from discovery

### Q: How do I rollback a service rotation?

**A:** Rotate again, specifying the old service as the "new" service.

### Q: Can I pause health checks?

**A:** No, health checks run continuously. To prevent failover, set the service to maintenance mode (decommission temporarily).

### Q: How long are health check records kept?

**A:** By default, all health check records are kept indefinitely. Consider implementing data retention policies based on your requirements.

---

## Troubleshooting Guide

### Problem: Service Registration Fails

**Symptoms:**
- `400 Bad Request` response
- `401 Unauthorized` response

**Solutions:**

1. **Check admin token:**
```bash
# Token must be at least 32 characters
echo "your-token" | wc -c
```

2. **Verify JSON format:**
```bash
# Validate JSON
cat registration.json | jq .
```

3. **Check required fields:**
- name, type, version, url, health_check_url, admin_token

---

### Problem: Service Not Discovered

**Symptoms:**
- Service registered but not in discovery results
- Empty service list

**Solutions:**

1. **Check service status:**
```bash
curl http://localhost:8080/api/services/health/service-123
```

2. **Verify service is healthy:**
```bash
# Test health endpoint directly
curl http://your-service:port/health
```

3. **Check if service is active:**
```bash
# Look for is_active: true
curl http://localhost:8080/api/services/list
```

---

### Problem: Failover Doesn't Happen

**Symptoms:**
- Primary service unhealthy
- Backup service not activated
- Clients still seeing errors

**Solutions:**

1. **Check backup service exists:**
```bash
curl http://localhost:8080/api/services/list | grep backup
```

2. **Verify failover group matches:**
```bash
# Primary and backup must have same failover_group
curl http://localhost:8080/api/services/list | jq '.services[] | {name, failover_group, role}'
```

3. **Check failure count:**
```bash
# Must reach 3 failures for failover
curl http://localhost:8080/api/services/health/primary-service-123
```

4. **Verify backup is healthy:**
```bash
curl http://backup-service:port/health
```

---

### Problem: Failback Doesn't Happen

**Symptoms:**
- Primary service recovered
- Still using backup service

**Solutions:**

1. **Check stability count:**
   - Primary must be healthy for 3 consecutive checks (3 minutes)

2. **Check time since failover:**
   - Must be at least 5 minutes since failover
   ```bash
   # Check failover_events table
   sqlite3 /app/Database/service_discovery.db \
     "SELECT * FROM service_failover_events WHERE failover_group='your-group' ORDER BY timestamp DESC LIMIT 1;"
   ```

3. **Verify primary role:**
```bash
# Service must have role="primary"
curl http://localhost:8080/api/services/list | jq '.services[] | select(.id=="primary-service-123")'
```

---

### Problem: Service Rotation Fails

**Symptoms:**
- `400 Bad Request` during rotation
- Error message about validation

**Solutions:**

1. **Check new service age:**
```bash
# Must be registered for at least 5 minutes
# Check registered_at field
curl http://localhost:8080/api/services/health/new-service-123
```

2. **Verify service types match:**
```bash
# Old and new service must have same type
curl http://localhost:8080/api/services/list | \
  jq '.services[] | select(.id=="old-service-123" or .id=="new-service-123") | {id, type}'
```

3. **Check new service signature:**
   - Ensure new service was properly signed before rotation

4. **Verify admin token:**
   - Must be at least 32 characters
   - Must be valid and not expired

---

### Problem: High Response Times

**Symptoms:**
- Health checks timing out
- Slow service discovery

**Solutions:**

1. **Check network latency:**
```bash
ping your-service-hostname
```

2. **Test health endpoint directly:**
```bash
time curl http://your-service:port/health
```

3. **Check service logs** for performance issues

4. **Verify database performance:**
```bash
# Check database size
ls -lh /app/Database/service_discovery.db

# Vacuum if needed
sqlite3 /app/Database/service_discovery.db "VACUUM;"
```

---

### Getting Help

**Documentation:**
- Technical Documentation: `ServiceDiscovery_Technical.md`
- API Reference: See Technical Documentation Section 4

**Support:**
- GitHub Issues: https://github.com/helixtrack/core/issues
- Email: support@helixtrack.ru
- Documentation: https://docs.helixtrack.ru

**Emergency Contact:**
- On-call: [Your on-call rotation]
- Slack: #helixtrack-support
- Phone: [Emergency phone number]

---

**End of User Manual**
