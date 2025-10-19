# Docker Infrastructure Failure Scenarios and Recovery

This document catalogs all possible failure scenarios in the HelixTrack Docker infrastructure and their recovery mechanisms.

## Critical Failure Scenarios

### 1. Complete System Lock (Deadlock)

**Scenario:** All services waiting on each other, no progress possible

**Possible Causes:**
- Circular dependency in service startup
- Database connection pool exhaustion
- Consul leader election deadlock
- All ports in range exhausted

**Prevention Mechanisms:**
- ✅ Services start independently (no circular dependencies)
- ✅ Database connection pooling with timeout
- ✅ Consul bootstrap_expect=1 prevents election deadlock
- ✅ Automatic port selection with fallback

**Recovery:**
```bash
# Force restart all services
docker-compose down --timeout 5
docker-compose up -d

# Or nuclear option
docker system prune -af --volumes
./scripts/start-production.sh
```

**Test:**
```bash
# Test is included in test-infrastructure.sh
# Simulates port exhaustion scenario
```

---

### 2. Database Total Failure

**Scenario:** PostgreSQL database completely fails or becomes corrupted

**Possible Causes:**
- Disk full
- Data corruption
- OOM (Out of Memory) killer
- Catastrophic hardware failure

**Detection:**
- Service health checks fail
- Connection timeouts
- pg_isready returns error

**Prevention:**
- ✅ Volume mounts for data persistence
- ✅ Health checks with restart policy
- ✅ Connection retry logic in services
- ✅ SSL/TLS encryption prevents data corruption in transit

**Recovery:**
```bash
# 1. Check database status
docker-compose exec core-db pg_isready

# 2. View logs
docker logs core-db

# 3. Restart database
docker-compose restart core-db

# 4. If corrupted, restore from backup
docker-compose down
docker volume rm helixtrack_core_db_data
# Restore backup volume
docker-compose up -d core-db

# 5. Re-run migrations
docker-compose exec core-db psql -U helixtrack -d helixtrack_core -f /migrations/latest.sql
```

**Mitigation:**
- Services continue running (fail open)
- Requests fail gracefully with proper error messages
- Auto-restart on failure

---

### 3. Network Partition

**Scenario:** Services can't communicate due to network failure

**Possible Causes:**
- Docker network failure
- Firewall blocking traffic
- Network bridge down
- DNS resolution failure

**Detection:**
- Services can't register with Consul
- Health checks timeout
- Connection refused errors

**Prevention:**
- ✅ All services on same Docker network
- ✅ Service discovery with fallback to direct addressing
- ✅ DNS resolution via Consul
- ✅ Multiple health check types (HTTP + TCP)

**Recovery:**
```bash
# 1. Check network
docker network inspect helixtrack-network

# 2. Restart network
docker-compose down
docker network rm helixtrack-network
docker network create helixtrack-network
docker-compose up -d

# 3. Re-register services
# Services auto-register on restart
```

**Mitigation:**
- HAProxy continues routing to known healthy instances
- Services retry connections with exponential backoff

---

### 4. Service Discovery Failure (Consul Down)

**Scenario:** Consul service registry becomes unavailable

**Possible Causes:**
- Consul crash
- Data corruption
- Disk full
- Network partition

**Detection:**
- Services can't register
- Discovery queries fail
- Consul health check critical

**Prevention:**
- ✅ Consul data persistence via volume
- ✅ Health checks with auto-restart
- ✅ Services can operate without Consul (degraded mode)

**Recovery:**
```bash
# 1. Restart Consul
docker-compose restart service-registry

# 2. Check Consul status
curl http://localhost:8500/v1/status/leader

# 3. Re-register all services
# Services auto-register on restart
docker-compose restart core-service auth-service perm-service

# 4. If corrupted, restore snapshot
docker-compose exec service-registry consul snapshot restore /backups/consul-backup.snap
```

**Mitigation:**
- Services continue running with last known configuration
- HAProxy uses static backend configuration as fallback
- Direct service addressing still works

---

### 5. Load Balancer Failure (HAProxy Down)

**Scenario:** HAProxy stops responding

**Possible Causes:**
- HAProxy crash
- Configuration error
- Memory exhaustion
- Process killed

**Detection:**
- Port 80/443 not responding
- Health check endpoint down
- Stats page unavailable

**Prevention:**
- ✅ HAProxy health checks
- ✅ Docker restart policy
- ✅ Configuration validation before reload
- ✅ Resource limits to prevent OOM

**Recovery:**
```bash
# 1. Restart HAProxy
docker-compose restart load-balancer

# 2. Validate configuration
docker-compose exec load-balancer haproxy -c -f /usr/local/etc/haproxy/haproxy.cfg

# 3. Check logs
docker logs load-balancer

# 4. Access services directly
curl http://core-service:8080/health
```

**Mitigation:**
- Services still accessible via direct ports
- Consul DNS provides service discovery
- Multiple HAProxy instances can run (scale up)

---

### 6. Port Exhaustion

**Scenario:** All available ports in range (8080-8089) are used

**Possible Causes:**
- Too many service instances
- Ports not released on shutdown
- Port conflict with external services

**Detection:**
- New instances fail to start
- "Address already in use" errors
- Service registration fails

**Prevention:**
- ✅ Automatic port selection scans for available ports
- ✅ Graceful shutdown releases ports
- ✅ Consul deregistration on service stop
- ✅ Configurable port range

**Recovery:**
```bash
# 1. Stop excess instances
docker-compose down --scale core-service=5

# 2. Check what's using ports
lsof -i :8080-8089

# 3. Expand port range
# Edit docker-compose.yml
SERVER_PORT_RANGE_END=8099  # Increase from 8089 to 8099

# 4. Restart services
docker-compose up -d
```

**Mitigation:**
- Services fail to start but don't crash existing instances
- Error logged clearly
- Manual intervention required

---

### 7. Memory Exhaustion (OOM)

**Scenario:** System runs out of memory

**Possible Causes:**
- Memory leak in application
- Too many containers running
- Large database queries
- Insufficient host memory

**Detection:**
- OOM killer logs in system logs
- Containers suddenly restarting
- Performance degradation

**Prevention:**
- ✅ Resource limits in docker-compose.yml
- ✅ PostgreSQL shared_buffers configured
- ✅ Go garbage collection
- ✅ Connection pooling prevents unbounded connections

**Recovery:**
```bash
# 1. Check memory usage
docker stats

# 2. Identify memory hog
docker stats --no-stream | sort -k4 -h

# 3. Restart problematic service
docker-compose restart <service>

# 4. Reduce scale
docker-compose up -d --scale core-service=2

# 5. Increase host memory or reduce limits
# Edit docker-compose.yml
resources:
  limits:
    memory: 512M  # Reduce from 1G
```

**Mitigation:**
- Docker restart policy brings services back up
- Other services continue running
- Requests fail gracefully

---

### 8. Disk Space Exhaustion

**Scenario:** Host disk fills up completely

**Possible Causes:**
- Log files growing unbounded
- Database data growth
- Docker images accumulating
- Volume data not cleaned

**Detection:**
- "No space left on device" errors
- Write operations fail
- Services can't start

**Prevention:**
- ✅ Log rotation configured in PostgreSQL
- ✅ Zap logger with size limits
- ✅ Docker log driver limits
- ✅ Regular cleanup of old images

**Recovery:**
```bash
# 1. Check disk space
df -h

# 2. Clean Docker system
docker system prune -af --volumes

# 3. Clean old logs
find /var/lib/docker/containers -name "*.log" -mtime +7 -delete

# 4. Resize volume or add disk
# Depends on host system

# 5. Restart services
docker-compose up -d
```

**Mitigation:**
- Regular automated cleanup
- Monitoring alerts before critical
- Database vacuum to reclaim space

---

### 9. Database Connection Pool Exhaustion

**Scenario:** All database connections in use, no new connections possible

**Possible Causes:**
- Long-running queries
- Connection leaks
- Too many concurrent requests
- max_connections too low

**Detection:**
- "Too many connections" errors
- Service hangs on database operations
- Timeouts

**Prevention:**
- ✅ Connection pooling with limits
- ✅ Connection timeouts
- ✅ Automatic connection cleanup
- ✅ PostgreSQL max_connections=100

**Recovery:**
```bash
# 1. Check active connections
docker-compose exec core-db psql -U helixtrack -c "SELECT count(*) FROM pg_stat_activity;"

# 2. Kill long-running queries
docker-compose exec core-db psql -U helixtrack -c "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE state = 'active' AND query_start < now() - interval '5 minutes';"

# 3. Restart services to reset connections
docker-compose restart core-service

# 4. Increase max_connections
# Edit docker/postgres/postgresql.conf
max_connections = 200

# Reload PostgreSQL
docker-compose exec core-db pg_ctl reload
```

**Mitigation:**
- Connection timeout prevents indefinite waiting
- Exponential backoff on retry
- Circuit breaker pattern

---

### 10. SSL Certificate Expiration

**Scenario:** SSL certificates expire

**Possible Causes:**
- Certificates not renewed
- Renewal automation failed
- Manual certificates forgotten

**Detection:**
- SSL handshake failures
- Browser warnings
- "Certificate expired" errors

**Prevention:**
- ✅ Self-signed certs valid for 10 years (dev)
- ✅ Automated renewal for Let's Encrypt (production)
- ✅ Monitoring alerts before expiration

**Recovery:**
```bash
# 1. Generate new certificates
cd docker/postgres
./docker-entrypoint-initdb.d/00-generate-ssl-certs.sh

# 2. Replace certificates in running containers
docker cp server.crt core-db:/var/lib/postgresql/
docker cp server.key core-db:/var/lib/postgresql/

# 3. Reload services
docker-compose restart core-db

# 4. For HAProxy
docker cp helixtrack.pem load-balancer:/usr/local/etc/haproxy/certs/
docker-compose exec load-balancer kill -USR2 1
```

**Mitigation:**
- Grace period before expiration
- Monitoring alerts
- Fallback to non-SSL (not recommended for production)

---

### 11. Configuration Corruption

**Scenario:** Configuration files become corrupted or invalid

**Possible Causes:**
- Manual editing errors
- File system corruption
- Incomplete file write
- Version mismatch

**Detection:**
- Services fail to start
- Validation errors in logs
- Docker Compose config fails

**Prevention:**
- ✅ Configuration validation before use
- ✅ Version control for all configs
- ✅ Atomic writes
- ✅ Default fallback configurations

**Recovery:**
```bash
# 1. Validate configuration
docker-compose config

# 2. Restore from git
git checkout docker-compose-production.yml

# 3. Restore from backup
cp docker-compose-production.yml.backup docker-compose-production.yml

# 4. Use default configuration
cp Configurations/default.json Configurations/production.json

# 5. Restart services
docker-compose up -d
```

**Mitigation:**
- Configuration validation on startup
- Fail fast with clear error messages
- Version control provides recovery point

---

### 12. Cascading Failures

**Scenario:** One failure triggers multiple failures

**Possible Causes:**
- Shared resource exhaustion
- Retry storms
- Circuit breaker not configured
- No backpressure

**Detection:**
- Multiple services failing simultaneously
- Exponentially increasing error rates
- System-wide performance degradation

**Prevention:**
- ✅ Circuit breaker pattern
- ✅ Retry with exponential backoff
- ✅ Request timeouts
- ✅ Bulkhead pattern (resource isolation)

**Recovery:**
```bash
# 1. Stop all services
docker-compose down

# 2. Start core dependencies first
docker-compose up -d core-db service-registry

# 3. Wait for stability
sleep 30

# 4. Start services gradually
docker-compose up -d core-service
sleep 10
docker-compose up -d auth-service perm-service
sleep 10
docker-compose up -d load-balancer

# 5. Monitor health
watch -n 1 'docker-compose ps'
```

**Mitigation:**
- Graceful degradation
- Fail fast where appropriate
- Isolated failure domains
- Rate limiting

---

## Recovery Procedures Summary

### Quick Recovery Checklist

1. **Identify the failing component**
   ```bash
   docker-compose ps
   docker-compose logs --tail=100
   ```

2. **Check resource utilization**
   ```bash
   docker stats
   df -h
   free -h
   ```

3. **Restart specific service**
   ```bash
   docker-compose restart <service>
   ```

4. **Full system restart**
   ```bash
   ./scripts/stop-production.sh
   ./scripts/start-production.sh
   ```

5. **Nuclear option (data loss!)**
   ```bash
   ./scripts/stop-production.sh --cleanup
   docker system prune -af --volumes
   ./scripts/start-production.sh
   ```

### Health Check Commands

```bash
# Overall system health
./tests/docker-infrastructure/test-infrastructure.sh

# Individual service health
curl http://localhost:8080/health  # Core
curl http://localhost:8081/health  # Auth
curl http://localhost:8082/health  # Permissions

# Infrastructure health
curl http://localhost:8500/v1/status/leader  # Consul
curl http://localhost:8404/stats  # HAProxy

# Database health
docker-compose exec core-db pg_isready
```

### Monitoring and Alerting

**Key Metrics to Monitor:**

1. **Service Health**
   - Health check status (Consul)
   - Response time
   - Error rate

2. **Resource Usage**
   - CPU utilization
   - Memory usage
   - Disk space
   - Network I/O

3. **Database**
   - Connection count
   - Query time
   - Replication lag
   - Lock wait time

4. **Infrastructure**
   - Container restart count
   - Service registration count
   - Load balancer request rate
   - SSL certificate expiration

**Alerting Thresholds:**

- Memory > 80%: Warning
- Memory > 95%: Critical
- Disk > 80%: Warning
- Disk > 95%: Critical
- Health check failures > 3: Critical
- Response time > 5s: Warning
- Response time > 10s: Critical

## Conclusion

The HelixTrack Docker infrastructure is designed with multiple layers of redundancy and graceful degradation:

1. **Service Independence**: Services can run independently
2. **Auto-Recovery**: Docker restart policies
3. **Health Monitoring**: Comprehensive health checks
4. **Graceful Degradation**: System continues in degraded mode
5. **Clear Error Messages**: Easy troubleshooting
6. **Multiple Recovery Paths**: From service restart to full rebuild

**No single point of total failure** - even if Consul, HAProxy, or database fails, services continue running and can be accessed directly.

**Testing:** Run `./tests/docker-infrastructure/test-infrastructure.sh` regularly to verify all recovery mechanisms work correctly.
