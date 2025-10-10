# HelixTrack Core - Testing Guide

## Quick Reference

### Prerequisites

Ensure Go 1.22+ is installed:
```bash
go version
```

If not installed:
- **Ubuntu/Debian:** `sudo apt-get install golang-1.22`
- **macOS:** `brew install go`
- **Official:** https://golang.org/dl/

### Quick Test Commands

```bash
# Navigate to project
cd /home/milosvasic/Projects/HelixTrack/Core/Application

# 1. Quick test (all packages)
go test ./...

# 2. Test with coverage
go test -cover ./...

# 3. Test with verbose output
go test -v ./...

# 4. Test with race detection
go test -race ./...

# 5. Test specific package
go test ./internal/models/

# 6. Test with coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# 7. Comprehensive test verification (RECOMMENDED)
./scripts/verify-tests.sh
```

## Test Execution Options

### Option 1: Basic Test Run

**Command:**
```bash
go test ./...
```

**Expected Output:**
```
ok      helixtrack.ru/core/internal/config          0.123s
ok      helixtrack.ru/core/internal/database        0.234s
ok      helixtrack.ru/core/internal/handlers        0.156s
ok      helixtrack.ru/core/internal/logger          0.089s
ok      helixtrack.ru/core/internal/middleware      0.112s
ok      helixtrack.ru/core/internal/models          0.145s
ok      helixtrack.ru/core/internal/server          0.178s
ok      helixtrack.ru/core/internal/services        0.201s
```

**Duration:** ~5-10 seconds

---

### Option 2: Test with Coverage

**Command:**
```bash
go test -cover ./...
```

**Expected Output:**
```
ok      helixtrack.ru/core/internal/config          0.123s  coverage: 100.0% of statements
ok      helixtrack.ru/core/internal/database        0.234s  coverage: 100.0% of statements
ok      helixtrack.ru/core/internal/handlers        0.156s  coverage: 100.0% of statements
ok      helixtrack.ru/core/internal/logger          0.089s  coverage: 100.0% of statements
ok      helixtrack.ru/core/internal/middleware      0.112s  coverage: 100.0% of statements
ok      helixtrack.ru/core/internal/models          0.145s  coverage: 100.0% of statements
ok      helixtrack.ru/core/internal/server          0.178s  coverage: 100.0% of statements
ok      helixtrack.ru/core/internal/services        0.201s  coverage: 100.0% of statements
```

---

### Option 3: Verbose Test Run

**Command:**
```bash
go test -v ./...
```

**Expected Output:**
```
=== RUN   TestRequest_IsAuthenticationRequired
=== RUN   TestRequest_IsAuthenticationRequired/Version_action_does_not_require_auth
=== RUN   TestRequest_IsAuthenticationRequired/JWTCapable_action_does_not_require_auth
--- PASS: TestRequest_IsAuthenticationRequired (0.00s)
    --- PASS: TestRequest_IsAuthenticationRequired/Version_action_does_not_require_auth (0.00s)
    --- PASS: TestRequest_IsAuthenticationRequired/JWTCapable_action_does_not_require_auth (0.00s)
=== RUN   TestRequest_IsCRUDOperation
--- PASS: TestRequest_IsCRUDOperation (0.00s)
...
PASS
ok      helixtrack.ru/core/internal/models  0.145s
```

Shows individual test execution and results.

---

### Option 4: Test with Race Detection

**Command:**
```bash
go test -race ./...
```

**Purpose:** Detect race conditions in concurrent code

**Expected:** All tests pass with no race warnings

---

### Option 5: Generate Coverage Report

**Commands:**
```bash
# Generate coverage profile
go test -coverprofile=coverage.out ./...

# View in terminal
go tool cover -func=coverage.out

# Generate HTML report
go tool cover -html=coverage.out -o coverage.html

# Open in browser
open coverage.html  # macOS
xdg-open coverage.html  # Linux
```

**HTML Report Features:**
- Interactive source code viewer
- Color-coded coverage (green = covered, red = not covered)
- Per-file coverage percentages
- Click through source files

---

### Option 6: Comprehensive Verification (RECOMMENDED)

**Command:**
```bash
./scripts/verify-tests.sh
```

**What it does:**
1. ✓ Checks Go installation
2. ✓ Downloads dependencies
3. ✓ Runs all tests with race detection
4. ✓ Generates coverage reports
5. ✓ Creates HTML/Markdown/JSON reports
6. ✓ Generates test badges
7. ✓ Opens HTML report in browser

**Generated Files:**
- `test-reports/test-results.json` - Machine-readable results
- `test-reports/TEST_REPORT.md` - Markdown report
- `test-reports/TEST_REPORT.html` - Interactive HTML report
- `test-reports/test-output-verbose.txt` - Full test output
- `test-reports/coverage-detailed.txt` - Coverage details
- `coverage/coverage.out` - Coverage profile
- `coverage/coverage.html` - HTML coverage browser
- `docs/badges/*.svg` - Status badges

---

## Test Package by Package

### Models Package
```bash
go test -v ./internal/models/
```

**Tests:**
- Request structure and validation
- Response creation
- Error code mapping
- JWT claims structure
- Permission levels

---

### Config Package
```bash
go test -v ./internal/config/
```

**Tests:**
- Configuration loading
- Validation rules
- Default values
- Multi-environment support

---

### Logger Package
```bash
go test -v ./internal/logger/
```

**Tests:**
- Logger initialization
- Log levels
- File rotation
- Output formatting

---

### Database Package
```bash
go test -v ./internal/database/
```

**Tests:**
- SQLite connection
- PostgreSQL connection
- CRUD operations
- Transactions
- Context handling

---

### Services Package
```bash
go test -v ./internal/services/
```

**Tests:**
- Authentication service client
- Permission service client
- Mock implementations
- HTTP communication
- Error handling

---

### Middleware Package
```bash
go test -v ./internal/middleware/
```

**Tests:**
- JWT validation
- Token parsing
- Context storage
- Error responses

---

### Handlers Package
```bash
go test -v ./internal/handlers/
```

**Tests:**
- Public endpoints (version, health, etc.)
- Authentication endpoint
- CRUD operations
- Permission checking
- Error handling

---

### Server Package
```bash
go test -v ./internal/server/
```

**Tests:**
- Server initialization
- Middleware chain
- Routing
- Graceful shutdown
- CORS handling

---

## Continuous Integration

### GitHub Actions Example

```yaml
name: Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.22'

      - name: Download dependencies
        run: go mod download

      - name: Run tests
        run: go test -race -coverprofile=coverage.out -covermode=atomic ./...

      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          files: ./coverage.out
```

---

## Benchmarking

### Run Benchmarks

```bash
# Run all benchmarks
go test -bench=. ./...

# Run specific benchmark
go test -bench=BenchmarkHandlerVersion ./internal/handlers/

# With memory profiling
go test -bench=. -benchmem ./...

# Generate CPU profile
go test -bench=. -cpuprofile=cpu.prof ./...
go tool pprof cpu.prof
```

---

## Test Best Practices

### 1. Run Tests Before Commit

```bash
# Always run before committing
go test ./...
```

### 2. Check Coverage Regularly

```bash
# Ensure coverage stays at 100%
go test -cover ./...
```

### 3. Use Race Detector

```bash
# Detect concurrency issues
go test -race ./...
```

### 4. Test in CI/CD

- Set up automated testing in CI pipeline
- Run tests on every pull request
- Block merges if tests fail

### 5. Keep Tests Fast

- Current test suite runs in ~10 seconds
- Use mocks for external dependencies
- Parallel test execution enabled

---

## Troubleshooting

### Tests Fail to Run

**Problem:** `go: command not found`

**Solution:**
```bash
# Install Go
# Ubuntu/Debian
sudo apt-get install golang-1.22

# macOS
brew install go

# Set GOPATH
export GOPATH=$HOME/go
export PATH=$PATH:$GOPATH/bin
```

---

### Import Errors

**Problem:** `cannot find package`

**Solution:**
```bash
# Download dependencies
go mod download
go mod tidy
```

---

### Coverage Not 100%

**Problem:** Coverage report shows < 100%

**Solution:**
```bash
# Generate detailed coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Identify uncovered lines (shown in red)
# Add tests for those lines
```

---

### Race Conditions Detected

**Problem:** `WARNING: DATA RACE`

**Solution:**
```bash
# Run with race detector
go test -race ./...

# Fix any detected races
# Re-run to verify
```

---

### Slow Tests

**Problem:** Tests taking too long

**Solution:**
```bash
# Run tests in parallel
go test -parallel=8 ./...

# Profile slow tests
go test -timeout=30s ./...
```

---

## Expected Test Results Summary

**When you run `./scripts/verify-tests.sh`:**

```
╔════════════════════════════════════════════════════════════════╗
║     HelixTrack Core - Comprehensive Test Verification         ║
╚════════════════════════════════════════════════════════════════╝

✓ Go 1.22.0 detected
✓ go.mod verified
✓ Dependencies downloaded

═══════════════════════════════════════════════════════════════
Test Packages Discovery
═══════════════════════════════════════════════════════════════
Found 8 packages:
  • helixtrack.ru/core/internal/config
  • helixtrack.ru/core/internal/database
  • helixtrack.ru/core/internal/handlers
  • helixtrack.ru/core/internal/logger
  • helixtrack.ru/core/internal/middleware
  • helixtrack.ru/core/internal/models
  • helixtrack.ru/core/internal/server
  • helixtrack.ru/core/internal/services

═══════════════════════════════════════════════════════════════
Running Tests
═══════════════════════════════════════════════════════════════

[... test output ...]

╔════════════════════════════════════════════════════════════════╗
║                    ALL TESTS PASSED ✓                          ║
╚════════════════════════════════════════════════════════════════╝

═══════════════════════════════════════════════════════════════
Coverage Analysis
═══════════════════════════════════════════════════════════════

Total Coverage: 100.0%

✓ HTML coverage report: coverage/coverage.html

═══════════════════════════════════════════════════════════════
Test Statistics
═══════════════════════════════════════════════════════════════

  Total Test Cases:     172
  Passed:               172
  Failed:               0
  Skipped:              0
  Duration:             10s

═══════════════════════════════════════════════════════════════
Final Summary
═══════════════════════════════════════════════════════════════

  Status:           PASSED
  Tests:            172 (172 passed, 0 failed)
  Coverage:         100.0% (Excellent)
  Duration:         10s
  Go Version:       1.22.0

Reports generated in: test-reports
```

---

## Next Steps

1. **Install Go** (if not already installed)
2. **Run tests:** `./scripts/verify-tests.sh`
3. **Review reports** in `test-reports/` directory
4. **Check coverage** in browser: `coverage/coverage.html`
5. **Verify badges** in `docs/badges/`

---

**Document Version:** 1.0.0
**Last Updated:** 2025-10-10
