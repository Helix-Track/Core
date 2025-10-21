# End-to-End Tests

This directory contains end-to-end (E2E) tests for the Localization Service. These tests validate the complete system by running against a live HTTP/3 QUIC service instance.

## Overview

**Purpose:** Validate complete system behavior including HTTP/3 QUIC protocol, TLS, authentication, and real database/cache interactions.

**Test Count:** 9 comprehensive tests

**Location:** `/e2e/e2e_test.go`

## Test Cases

### 1. TestHealthCheck
Validates service health endpoint without authentication.

**Checks:**
- Service is running and responsive
- Database connectivity
- Cache connectivity
- Health response format

**Authentication:** Not required

---

### 2. TestGetCatalog
Tests complete catalog retrieval for a language.

**Checks:**
- JWT authentication
- Catalog data structure
- Language-specific catalog loading
- Success response format

**Authentication:** Required

---

### 3. TestGetSingleLocalization
Tests single localization key retrieval.

**Checks:**
- Key-based localization fetching
- Variable interpolation (if applicable)
- Fallback behavior
- Response data format

**Authentication:** Required

---

### 4. TestBatchLocalization
Tests batch retrieval of multiple localization keys.

**Checks:**
- Batch request format
- Multiple key handling
- Fallback support
- Performance for bulk operations

**Authentication:** Required

---

### 5. TestGetLanguages
Tests language enumeration endpoint.

**Checks:**
- All available languages returned
- Active language filtering
- Language metadata completeness

**Authentication:** Required

---

### 6. TestCompleteWorkflow
Simulates a complete user workflow through multiple endpoints.

**Steps:**
1. Health check
2. Get available languages
3. Load catalog for specific language
4. Batch fetch localizations

**Validates:** State consistency across multiple API calls

**Authentication:** Required

---

### 7. TestCachePerformance
Measures and validates caching effectiveness.

**Checks:**
- First request latency (uncached)
- Second request latency (cached)
- Performance improvement percentage
- Cache hit behavior

**Expected:** 30-90% performance improvement for cached requests

**Authentication:** Required

---

### 8. TestHTTP3Protocol
Verifies HTTP/3 QUIC protocol usage.

**Checks:**
- Protocol version in response
- Alt-Svc header presence
- QUIC transport layer

**Note:** Some clients may negotiate HTTP/2 over TLS; Alt-Svc header indicates HTTP/3 availability.

**Authentication:** Not required

---

### 9. TestErrorHandling
Tests error scenarios and response formats.

**Scenarios:**
- Invalid endpoint (404 Not Found)
- Missing authentication (401 Unauthorized)

**Validates:** Proper error codes and response formats

**Authentication:** Mixed (tests both authenticated and unauthenticated scenarios)

---

## Prerequisites

Before running E2E tests, ensure:

1. **Service is Running**
   ```bash
   cd /path/to/localization-service
   ./htLoc --config=configs/default.json
   ```

2. **TLS Certificates Generated**
   ```bash
   ./scripts/generate-certs.sh
   ```

3. **Database Initialized**
   - PostgreSQL with test data, or
   - SQLite with seed data

4. **Configuration Matches**
   - JWT secret in environment matches server config
   - Service URL is accessible
   - Port 8085 (default) is not blocked by firewall

## Running Tests

### Quick Start

```bash
# From service root directory
export SERVICE_URL="https://localhost:8085"
export JWT_SECRET="your-jwt-secret-from-config"
go test ./e2e/ -v
```

### Using Test Runner Script

```bash
# Run all tests (unit + integration + E2E)
./scripts/run-all-tests.sh

# Skip E2E tests
SKIP_E2E=true ./scripts/run-all-tests.sh
```

### Manual Test Run

```bash
# Terminal 1: Start service
./htLoc --config=configs/default.json

# Terminal 2: Run E2E tests
cd e2e
go test -v
```

## Configuration

E2E tests are configured via environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `SERVICE_URL` | Service endpoint | `https://localhost:8085` |
| `JWT_SECRET` | JWT signing secret | `test-secret-key-for-e2e-testing` |
| `JWT_TOKEN` | Pre-generated JWT token | Auto-generated if not set |

**Note:** If `JWT_TOKEN` is not provided, tests will generate a token using `JWT_SECRET`. The secret must match the server configuration.

## Expected Output

### Successful Run

```
========================================
E2E Tests for Localization Service
========================================

Configuration:
  Service URL: https://localhost:8085
  JWT Secret: test-secret-key-for-e2e-testing

Prerequisites:
  1. Service must be running with HTTP/3 QUIC enabled
  2. TLS certificates must be valid (or self-signed)
  3. Database must be initialized with test data
  4. JWT secret must match server configuration

Note: Tests requiring JWT_TOKEN will be skipped if not set.
      Health check test will always run.

=== RUN   TestHealthCheck
    e2e_test.go:61: Testing health check endpoint...
    e2e_test.go:81: ✓ Health check passed
--- PASS: TestHealthCheck (0.12s)

=== RUN   TestGetCatalog
    e2e_test.go:92: Testing catalog retrieval...
    e2e_test.go:130: ✓ Catalog retrieval passed
--- PASS: TestGetCatalog (0.08s)

...

PASS
ok      github.com/helixtrack/localization-service/e2e  0.978s

========================================
E2E Tests Completed
========================================
```

### Service Not Running

If the service is not running, tests requiring connectivity will fail:

```
=== RUN   TestHealthCheck
    e2e_test.go:64: Failed to call health endpoint: dial tcp [::1]:8085: connect: connection refused
--- FAIL: TestHealthCheck (0.01s)
```

**Solution:** Start the service before running E2E tests.

## Test Patterns

### Reused from Integration Tests

The E2E tests reuse patterns from the integration test suite:

- **JWT Token Generation:** `createTestJWT(username, role, secret)`
  - Matches the pattern from `internal/handlers/integration_test.go`
  - Generates HS256-signed tokens with 1-hour expiration

- **Claims Structure:** Identical to production JWT claims
  - Username and Role fields
  - Standard JWT registered claims (exp, iat)

- **Test Organization:** Similar naming and structure
  - Test functions follow `TestXxxYyy` pattern
  - Descriptive test names
  - Comprehensive logging

### E2E-Specific Features

- **Real HTTP Client:** Uses `net/http` with TLS transport
- **Self-Signed Cert Support:** `InsecureSkipVerify: true` for testing
- **Network Timeouts:** 30-second timeout for all requests
- **Performance Measurement:** Tracks request latency
- **Protocol Verification:** Checks HTTP/3 protocol usage

## Troubleshooting

### Issue: Connection Refused

**Cause:** Service not running or wrong URL

**Solution:**
```bash
# Check if service is running
curl -k https://localhost:8085/health

# Start service if not running
./htLoc --config=configs/default.json
```

---

### Issue: Certificate Errors

**Cause:** TLS certificates not generated or invalid

**Solution:**
```bash
# Generate self-signed certificates
./scripts/generate-certs.sh

# Verify certificates exist
ls -la certs/
```

---

### Issue: 401 Unauthorized

**Cause:** JWT secret mismatch

**Solution:**
```bash
# Ensure JWT_SECRET matches server config
export JWT_SECRET=$(grep jwt_secret configs/default.json | cut -d'"' -f4)

# Or use the same secret as server
export JWT_SECRET="your-production-secret"
```

---

### Issue: Tests Skipped

**Cause:** `JWT_TOKEN` environment variable not set and tests auto-generate tokens

**Note:** This is expected behavior. Tests that require authentication will use auto-generated tokens.

**Verify:**
```bash
# Check test output for:
# "Generated test JWT token for user 'e2euser'"
```

---

### Issue: HTTP/3 Not Detected

**Cause:** Client negotiated HTTP/2 instead of HTTP/3

**Note:** This is normal. HTTP/3 support depends on:
- Server configuration
- Client library support
- Network conditions

**Verify:** Check for `Alt-Svc` header indicating HTTP/3 availability

---

## Integration with CI/CD

### GitHub Actions Example

```yaml
name: E2E Tests

on: [push, pull_request]

jobs:
  e2e:
    runs-on: ubuntu-latest

    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_PASSWORD: testpass
        options: >-
          --health-cmd pg_isready
          --health-interval 10s

    steps:
      - uses: actions/checkout@v3

      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.22'

      - name: Generate Certificates
        run: ./scripts/generate-certs.sh

      - name: Start Service
        run: |
          ./htLoc --config=configs/ci.json &
          sleep 5

      - name: Run E2E Tests
        env:
          SERVICE_URL: https://localhost:8085
          JWT_SECRET: ci-test-secret
        run: go test ./e2e/ -v

      - name: Stop Service
        run: killall htLoc
```

## Best Practices

1. **Always Check Service Status**
   - Verify service is running before E2E tests
   - Use health check endpoint to confirm

2. **Use Proper Secrets**
   - Never commit JWT secrets to version control
   - Use environment variables or secure vaults

3. **Clean Test Data**
   - Use dedicated test database
   - Reset state between test runs if needed

4. **Monitor Performance**
   - Track cache performance improvements
   - Set alerts for regression

5. **Separate Test Levels**
   - Run unit/integration tests first (fast feedback)
   - Run E2E tests last (complete validation)
   - Use `SKIP_E2E=true` for quick iterations

## Documentation

- **Main Test Report:** `../TEST_RESULTS.md`
- **Service Manual:** `../USER_MANUAL.md`
- **API Reference:** `../USER_MANUAL.md` (API section)
- **Integration Tests:** `../internal/handlers/integration_test.go`

## Support

For issues or questions:
1. Check logs: `tail -f /tmp/htLocLogs/*.log`
2. Verify configuration: `cat configs/default.json`
3. Review test output for specific error messages
4. Consult USER_MANUAL.md for API details

---

**Last Updated:** October 21, 2025
**Test Framework:** Go testing + testify
**Protocol:** HTTP/3 QUIC with TLS 1.3
