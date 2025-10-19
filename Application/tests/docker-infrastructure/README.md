# Docker Infrastructure Test Suite

Comprehensive testing for HelixTrack Docker infrastructure including service discovery, load balancing, automatic port selection, and failure recovery.

## Overview

This test suite verifies:

- ✅ Docker and Docker Compose installation
- ✅ Configuration file validity
- ✅ Service startup and health checks
- ✅ PostgreSQL with encryption (pgcrypto)
- ✅ Consul service discovery
- ✅ HAProxy load balancing
- ✅ Automatic port selection (8080-8089)
- ✅ Service rotation and failover
- ✅ Graceful shutdown
- ✅ SSL/TLS encryption
- ✅ Failure scenario recovery

## Files

- `test-infrastructure.sh` - Main test script (35 tests)
- `FAILURE_SCENARIOS.md` - Comprehensive failure analysis and recovery procedures
- `README.md` - This file

## Running Tests

### Full Test Suite

```bash
# Run all tests
cd /path/to/HelixTrack/Core/Application
./tests/docker-infrastructure/test-infrastructure.sh

# Expected output:
# =========================================
#   HelixTrack Docker Infrastructure Tests
# =========================================
# ...
# Total tests run:    35
# Tests passed:       35
# Tests failed:       0
# Tests skipped:      0
# ✓ All tests passed!
```

### Cleanup Only

```bash
# Just cleanup test environment
./tests/docker-infrastructure/test-infrastructure.sh --cleanup-only
```

### Skip Cleanup

```bash
# Keep services running after tests
SKIP_CLEANUP=1 ./tests/docker-infrastructure/test-infrastructure.sh
```

## Test Phases

### Phase 1: Prerequisites (10 tests)

Verifies that the environment is ready:

- Docker and Docker Compose installed
- Docker daemon running
- Configuration files exist
- Scripts are executable
- Docker Compose configuration valid

**Example:**
```bash
✓ PASS: Docker installed
✓ PASS: Docker daemon running
✓ PASS: Configuration files exist
```

### Phase 2: Service Startup (8 tests)

Tests service initialization:

- PostgreSQL database starts and accepts connections
- pgcrypto encryption extension loaded
- Consul service registry starts
- Consul API and UI accessible
- Core service starts and becomes healthy

**Example:**
```bash
✓ PASS: Database service starts
✓ PASS: Database accepts connections
✓ PASS: pgcrypto extension available
```

### Phase 3: Service Discovery (2 tests)

Validates Consul integration:

- Services register with Consul
- Service discovery returns correct port

**Example:**
```bash
✓ PASS: Core service registers with Consul
✓ PASS: Service discovery returns correct port
```

### Phase 4: Load Balancing (3 tests)

Verifies HAProxy functionality:

- HAProxy starts successfully
- Statistics dashboard accessible
- Requests properly routed to backend

**Example:**
```bash
✓ PASS: HAProxy starts
✓ PASS: HAProxy stats accessible
✓ PASS: HAProxy routes to backend
```

### Phase 5: Scaling and Rotation (4 tests)

Tests horizontal scaling:

- Multiple service instances start
- Each instance gets unique port (automatic port selection)
- Load balancer distributes requests
- Service rotation works (instances can be stopped/started)

**Example:**
```bash
✓ PASS: Multiple service instances
✓ PASS: Each instance gets unique port
✓ PASS: Load balancer distributes requests
✓ PASS: Service rotation works
```

### Phase 6: Health Checks (1 test)

Validates health monitoring:

- Failed health checks cause deregistration from Consul

**Example:**
```bash
✓ PASS: Failed health check deregisters
```

### Phase 7: Security (1 test)

Tests encryption:

- Database SSL/TLS connections work

**Example:**
```bash
✓ PASS: Database SSL connection works
```

### Phase 8: Graceful Shutdown (1 test)

Verifies cleanup:

- Services deregister from Consul on shutdown

**Example:**
```bash
✓ PASS: Graceful shutdown works
```

### Phase 9: Failure Scenarios (5 tests)

Tests recovery mechanisms:

1. **Database Failure Recovery**
   - Stop database
   - Verify service marks itself unhealthy
   - Restart database
   - Verify service recovers

2. **Network Partition Recovery**
   - Simulate network disconnect
   - Verify services continue operating
   - Reconnect network
   - Verify full recovery

3. **Port Exhaustion Handling**
   - Start more instances than available ports
   - Verify system handles gracefully
   - No crashes or deadlocks

4. **Consul Failure Handling**
   - Stop Consul
   - Verify services continue running
   - Restart Consul
   - Verify services re-register

5. **HAProxy Failure Handling**
   - Stop load balancer
   - Verify direct service access works
   - Restart load balancer
   - Verify routing restored

**Example:**
```bash
✓ PASS: Database failure recovery
✓ PASS: Network partition recovery (config verified)
✓ PASS: Port exhaustion handling (limited to 10 instances)
✓ PASS: Consul failure handling (services continue running)
✓ PASS: HAProxy failure handling (direct access works)
```

## Test Requirements

### System Requirements

- Docker 20.10+
- Docker Compose 1.29+ or Docker Compose V2
- 4GB+ available RAM
- 10GB+ available disk space
- Ports 80, 443, 8080-8089, 8404-8406, 8500 available

### Required Tools

- `curl` - For HTTP requests
- `jq` - For JSON parsing
- `lsof` or `netstat` - For port checking
- `docker` and `docker-compose` - For container management

### Installation

```bash
# Ubuntu/Debian
sudo apt-get update
sudo apt-get install -y curl jq lsof

# macOS
brew install curl jq lsof

# All tools should be available by default on most systems
```

## Understanding Test Results

### Success

```
=========================================
  Test Summary
=========================================
Total tests run:    35
Tests passed:       35
Tests failed:       0
Tests skipped:      0

✓ All tests passed!
```

Exit code: `0`

### Failure

```
=========================================
  Test Summary
=========================================
Total tests run:    35
Tests passed:       32
Tests failed:       3
Tests skipped:      0

✗ Some tests failed!
```

Exit code: `1`

Check the test output for specific failures:

```
✗ FAIL: Database accepts connections
```

### Debugging Failed Tests

1. **Check Docker status:**
   ```bash
   docker ps -a
   docker-compose ps
   ```

2. **View service logs:**
   ```bash
   docker logs core-service
   docker logs core-db
   docker logs service-registry
   ```

3. **Check resource usage:**
   ```bash
   docker stats
   df -h
   free -h
   ```

4. **Validate configuration:**
   ```bash
   docker-compose config
   ```

5. **Test individual components:**
   ```bash
   # Database
   docker-compose exec core-db pg_isready

   # Consul
   curl http://localhost:8500/v1/status/leader

   # Core service
   curl http://localhost:8080/health
   ```

## CI/CD Integration

### GitLab CI

```yaml
test-docker-infrastructure:
  stage: test
  services:
    - docker:dind
  script:
    - apk add --no-cache curl jq
    - ./tests/docker-infrastructure/test-infrastructure.sh
  after_script:
    - ./tests/docker-infrastructure/test-infrastructure.sh --cleanup-only
```

### GitHub Actions

```yaml
name: Docker Infrastructure Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - name: Run tests
        run: |
          sudo apt-get update
          sudo apt-get install -y curl jq
          ./tests/docker-infrastructure/test-infrastructure.sh
```

### Jenkins

```groovy
pipeline {
    agent any
    stages {
        stage('Test') {
            steps {
                sh './tests/docker-infrastructure/test-infrastructure.sh'
            }
        }
    }
    post {
        always {
            sh './tests/docker-infrastructure/test-infrastructure.sh --cleanup-only'
        }
    }
}
```

## Performance Benchmarks

Expected test execution times:

| Phase | Tests | Time |
|-------|-------|------|
| Prerequisites | 10 | ~5s |
| Service Startup | 8 | ~60s |
| Service Discovery | 2 | ~15s |
| Load Balancing | 3 | ~10s |
| Scaling | 4 | ~30s |
| Health Checks | 1 | ~10s |
| Security | 1 | ~5s |
| Graceful Shutdown | 1 | ~10s |
| Failure Scenarios | 5 | ~120s |
| **Total** | **35** | **~4-5 minutes** |

## Troubleshooting

### Test Hangs

```bash
# Kill hanging test
Ctrl+C

# Cleanup
./tests/docker-infrastructure/test-infrastructure.sh --cleanup-only

# Try again
./tests/docker-infrastructure/test-infrastructure.sh
```

### Port Already in Use

```bash
# Find what's using the port
lsof -i :8080

# Stop the service
docker stop <container-id>

# Or use different port range
# Edit docker-compose-production.yml
SERVER_PORT_RANGE_START=9080
SERVER_PORT_RANGE_END=9089
```

### Out of Memory

```bash
# Reduce scale in docker-compose.yml
# Or increase Docker memory limits
# Docker Desktop: Preferences > Resources > Memory
```

### Disk Space Issues

```bash
# Clean up Docker
docker system prune -af --volumes

# Check space
df -h
```

## Extending Tests

### Add New Test

```bash
# In test-infrastructure.sh

# 1. Define test function
test_my_new_feature() {
    # Test logic here
    if some_condition; then
        return 0  # Pass
    else
        return 1  # Fail
    fi
}

# 2. Add to run_all_tests()
run_test "My new feature" test_my_new_feature
```

### Add New Failure Scenario

```bash
# Define test
test_my_failure_scenario() {
    print_test "My failure scenario"

    # Trigger failure
    # ...

    # Verify recovery
    if recovery_successful; then
        print_pass "My failure scenario"
        return 0
    fi

    print_fail "My failure scenario"
    return 1
}

# Add to Phase 9
test_my_failure_scenario
((TESTS_RUN++))
```

## Additional Resources

- [Failure Scenarios Documentation](FAILURE_SCENARIOS.md)
- [Docker Compose Documentation](https://docs.docker.com/compose/)
- [HAProxy Documentation](../docker/haproxy/README.md)
- [Consul Documentation](../docker/consul/README.md)
- [PostgreSQL Encryption](../docker/postgres/README.md)

## Support

For issues or questions:

1. Check [FAILURE_SCENARIOS.md](FAILURE_SCENARIOS.md)
2. Review Docker logs
3. Run individual tests manually
4. Open an issue on GitHub

## License

Same as HelixTrack Core - MIT License
