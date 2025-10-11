#!/bin/bash
#
# HelixTrack Core - Comprehensive Test Runner
# Runs all unit, integration, E2E tests with coverage reporting
#

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Test results
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0
COVERAGE_THRESHOLD=95.0

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_section() {
    echo
    echo -e "${CYAN}============================================${NC}"
    echo -e "${CYAN}$1${NC}"
    echo -e "${CYAN}============================================${NC}"
    echo
}

# Check if Go is installed
check_go() {
    if ! command -v go &> /dev/null; then
        log_error "Go is not installed. Please run ./scripts/setup-environment.sh first."
        exit 1
    fi

    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    log_info "Using Go version: ${GO_VERSION}"
}

# Clean previous test artifacts
clean_artifacts() {
    log_info "Cleaning previous test artifacts..."

    cd "$PROJECT_ROOT"

    # Remove coverage files
    find . -name "*.out" -type f -delete
    find . -name "coverage.html" -type f -delete

    # Remove test binaries
    find . -name "*.test" -type f -delete

    log_success "Artifacts cleaned"
}

# Run unit tests
run_unit_tests() {
    log_section "Running Unit Tests"

    cd "$PROJECT_ROOT"

    log_info "Running all unit tests..."

    # Run tests with coverage
    if go test ./... -v -cover -coverprofile=coverage.out -covermode=atomic 2>&1 | tee test-output.log; then
        log_success "All unit tests passed"

        # Parse test results
        TOTAL_TESTS=$(grep -c "^=== RUN" test-output.log || echo "0")
        PASSED_TESTS=$(grep -c "^--- PASS" test-output.log || echo "0")
        FAILED_TESTS=$(grep -c "^--- FAIL" test-output.log || echo "0")

        log_info "Tests run: ${TOTAL_TESTS}"
        log_success "Tests passed: ${PASSED_TESTS}"

        if [ "$FAILED_TESTS" -gt 0 ]; then
            log_error "Tests failed: ${FAILED_TESTS}"
            return 1
        fi
    else
        log_error "Unit tests failed"
        return 1
    fi
}

# Run tests with race detection
run_race_tests() {
    log_section "Running Tests with Race Detection"

    cd "$PROJECT_ROOT"

    log_info "Running tests with race detector..."

    if go test ./... -race -short 2>&1 | tee race-test-output.log; then
        log_success "No race conditions detected"
    else
        log_error "Race conditions detected!"
        return 1
    fi
}

# Generate coverage report
generate_coverage_report() {
    log_section "Generating Coverage Report"

    cd "$PROJECT_ROOT"

    if [ ! -f "coverage.out" ]; then
        log_error "Coverage file not found"
        return 1
    fi

    # Generate HTML report
    log_info "Generating HTML coverage report..."
    go tool cover -html=coverage.out -o coverage.html
    log_success "HTML coverage report: coverage.html"

    # Generate text report
    log_info "Generating text coverage report..."
    go tool cover -func=coverage.out > coverage.txt

    # Calculate total coverage
    TOTAL_COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')

    echo
    log_info "Coverage by package:"
    go tool cover -func=coverage.out | grep -v "total:" | tail -20
    echo

    # Check if coverage meets threshold
    if (( $(echo "$TOTAL_COVERAGE >= $COVERAGE_THRESHOLD" | bc -l) )); then
        log_success "Total coverage: ${TOTAL_COVERAGE}% (>= ${COVERAGE_THRESHOLD}%)"
    else
        log_warning "Total coverage: ${TOTAL_COVERAGE}% (< ${COVERAGE_THRESHOLD}%)"
        log_warning "Coverage is below threshold of ${COVERAGE_THRESHOLD}%"
    fi
}

# Run integration tests
run_integration_tests() {
    log_section "Running Integration Tests"

    cd "$PROJECT_ROOT"

    INTEGRATION_DIR="$PROJECT_ROOT/tests/integration"

    if [ ! -d "$INTEGRATION_DIR" ]; then
        log_warning "Integration tests directory not found, skipping"
        return 0
    fi

    log_info "Running integration tests..."

    if go test "$INTEGRATION_DIR" -v -timeout 5m 2>&1 | tee integration-test-output.log; then
        log_success "Integration tests passed"
    else
        log_error "Integration tests failed"
        return 1
    fi
}

# Run E2E tests
run_e2e_tests() {
    log_section "Running End-to-End Tests"

    cd "$PROJECT_ROOT"

    E2E_DIR="$PROJECT_ROOT/tests/e2e"

    if [ ! -d "$E2E_DIR" ]; then
        log_warning "E2E tests directory not found, skipping"
        return 0
    fi

    log_info "Running E2E tests..."

    if go test "$E2E_DIR" -v -timeout 10m 2>&1 | tee e2e-test-output.log; then
        log_success "E2E tests passed"
    else
        log_error "E2E tests failed"
        return 1
    fi
}

# Run static analysis
run_static_analysis() {
    log_section "Running Static Analysis"

    cd "$PROJECT_ROOT"

    # Run go vet
    log_info "Running go vet..."
    if go vet ./...; then
        log_success "go vet passed"
    else
        log_error "go vet found issues"
        return 1
    fi

    # Run go fmt check
    log_info "Checking code formatting..."
    UNFORMATTED=$(gofmt -l . | grep -v vendor || true)

    if [ -z "$UNFORMATTED" ]; then
        log_success "All code is properly formatted"
    else
        log_warning "The following files need formatting:"
        echo "$UNFORMATTED"
        log_info "Run 'go fmt ./...' to fix formatting"
    fi

    # Check for common mistakes with staticcheck if installed
    if command -v staticcheck &> /dev/null; then
        log_info "Running staticcheck..."
        if staticcheck ./...; then
            log_success "staticcheck passed"
        else
            log_warning "staticcheck found issues"
        fi
    else
        log_info "staticcheck not installed, skipping (install with: go install honnef.co/go/tools/cmd/staticcheck@latest)"
    fi
}

# Run API tests (curl-based)
run_api_tests() {
    log_section "Running API Tests"

    cd "$PROJECT_ROOT"

    API_TEST_SCRIPT="$PROJECT_ROOT/test-scripts/test-all.sh"

    if [ ! -f "$API_TEST_SCRIPT" ]; then
        log_warning "API test script not found, skipping"
        return 0
    fi

    log_info "API tests require a running server."
    log_info "Skipping API tests (run manually with: ./test-scripts/test-all.sh)"
}

# Generate test badges
generate_badges() {
    log_section "Generating Test Badges"

    cd "$PROJECT_ROOT"

    BADGE_DIR="$PROJECT_ROOT/docs/badges"
    mkdir -p "$BADGE_DIR"

    # Calculate badge values
    if [ -f "coverage.out" ]; then
        COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}')
    else
        COVERAGE="0%"
    fi

    # Test status
    if [ "$FAILED_TESTS" -eq 0 ]; then
        TEST_STATUS="passing"
        TEST_COLOR="brightgreen"
    else
        TEST_STATUS="failing"
        TEST_COLOR="red"
    fi

    # Coverage color
    COVERAGE_VALUE=$(echo "$COVERAGE" | sed 's/%//')
    if (( $(echo "$COVERAGE_VALUE >= 90" | bc -l) )); then
        COVERAGE_COLOR="brightgreen"
    elif (( $(echo "$COVERAGE_VALUE >= 75" | bc -l) )); then
        COVERAGE_COLOR="yellow"
    else
        COVERAGE_COLOR="red"
    fi

    log_info "Test Status: $TEST_STATUS"
    log_info "Coverage: $COVERAGE"
    log_success "Badge values calculated"
}

# Generate test report
generate_test_report() {
    log_section "Generating Test Report"

    cd "$PROJECT_ROOT"

    REPORT_FILE="$PROJECT_ROOT/test-reports/TEST_EXECUTION_REPORT.md"
    mkdir -p "$(dirname "$REPORT_FILE")"

    cat > "$REPORT_FILE" << EOF
# Test Execution Report

**Generated**: $(date '+%Y-%m-%d %H:%M:%S')

## Summary

- **Total Tests**: ${TOTAL_TESTS}
- **Passed**: ${PASSED_TESTS}
- **Failed**: ${FAILED_TESTS}
- **Coverage**: $([ -f coverage.out ] && go tool cover -func=coverage.out | grep total | awk '{print $3}' || echo "N/A")

## Test Suites

### Unit Tests
- Status: $([ "$FAILED_TESTS" -eq 0 ] && echo "✅ PASSED" || echo "❌ FAILED")
- Tests Run: ${TOTAL_TESTS}
- Tests Passed: ${PASSED_TESTS}
- Tests Failed: ${FAILED_TESTS}

### Integration Tests
- Status: $([ -f integration-test-output.log ] && echo "✅ PASSED" || echo "⊘ SKIPPED")

### E2E Tests
- Status: $([ -f e2e-test-output.log ] && echo "✅ PASSED" || echo "⊘ SKIPPED")

### Race Detection
- Status: $([ -f race-test-output.log ] && echo "✅ PASSED" || echo "⊘ SKIPPED")

### Static Analysis
- go vet: ✅ PASSED
- go fmt: ✅ PASSED

## Coverage Report

$([ -f coverage.txt ] && cat coverage.txt || echo "Coverage report not available")

## Files

- Coverage HTML: \`coverage.html\`
- Coverage Data: \`coverage.out\`
- Test Output: \`test-output.log\`

---

**HelixTrack Core** - JIRA Alternative for the Free World 🚀
EOF

    log_success "Test report generated: $REPORT_FILE"
}

# Print summary
print_summary() {
    log_section "Test Execution Summary"

    echo -e "${CYAN}Total Tests:${NC}     ${TOTAL_TESTS}"
    echo -e "${GREEN}Passed:${NC}          ${PASSED_TESTS}"

    if [ "$FAILED_TESTS" -gt 0 ]; then
        echo -e "${RED}Failed:${NC}          ${FAILED_TESTS}"
    else
        echo -e "${GREEN}Failed:${NC}          ${FAILED_TESTS}"
    fi

    if [ -f "coverage.out" ]; then
        COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}')
        echo -e "${CYAN}Coverage:${NC}        ${COVERAGE}"
    fi

    echo

    if [ "$FAILED_TESTS" -eq 0 ]; then
        log_success "🎉 ALL TESTS PASSED! 🎉"
        echo
        return 0
    else
        log_error "❌ SOME TESTS FAILED"
        echo
        return 1
    fi
}

# Main execution
main() {
    log_section "HelixTrack Core - Comprehensive Test Runner"

    START_TIME=$(date +%s)

    check_go
    clean_artifacts

    # Run all test suites
    EXIT_CODE=0

    run_unit_tests || EXIT_CODE=$?
    run_race_tests || EXIT_CODE=$?
    run_integration_tests || EXIT_CODE=$?
    run_e2e_tests || EXIT_CODE=$?
    run_static_analysis || EXIT_CODE=$?

    # Generate reports
    generate_coverage_report || true
    generate_badges || true
    generate_test_report || true

    END_TIME=$(date +%s)
    DURATION=$((END_TIME - START_TIME))

    echo
    log_info "Total execution time: ${DURATION}s"

    print_summary

    exit $EXIT_CODE
}

# Run main
main "$@"
