#!/bin/bash

# Event Publishing Test Runner
# Runs all event publishing tests and generates reports

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
TEST_DIR="./internal/handlers"
WEBSOCKET_TEST_DIR="./internal/websocket"
REPORT_DIR="./test-reports/event-publishing"
COVERAGE_DIR="${REPORT_DIR}/coverage"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")

# Create report directories
mkdir -p "${REPORT_DIR}"
mkdir -p "${COVERAGE_DIR}"

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Event Publishing Test Runner${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Function to print section header
print_header() {
    echo -e "${YELLOW}>>> $1${NC}"
}

# Function to print success
print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

# Function to print error
print_error() {
    echo -e "${RED}✗ $1${NC}"
}

# Test counters
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# Run handler event publishing tests
print_header "Running Handler Event Publishing Tests"
echo ""

HANDLERS=(
    "Priority"
    "Resolution"
    "Watcher"
    "Ticket"
    "Project"
    "Comment"
    "Version"
    "Filter"
    "CustomField"
)

for HANDLER in "${HANDLERS[@]}"; do
    echo -e "${BLUE}Testing ${HANDLER} Handler...${NC}"

    # Run tests
    TEST_OUTPUT=$(go test -v -run "Test${HANDLER}Handler.*Event" ${TEST_DIR} 2>&1)
    TEST_RESULT=$?

    # Count tests
    TEST_COUNT=$(echo "${TEST_OUTPUT}" | grep -c "^=== RUN" || true)
    PASS_COUNT=$(echo "${TEST_OUTPUT}" | grep -c "^--- PASS:" || true)
    FAIL_COUNT=$(echo "${TEST_OUTPUT}" | grep -c "^--- FAIL:" || true)

    TOTAL_TESTS=$((TOTAL_TESTS + TEST_COUNT))
    PASSED_TESTS=$((PASSED_TESTS + PASS_COUNT))
    FAILED_TESTS=$((FAILED_TESTS + FAIL_COUNT))

    # Save output
    echo "${TEST_OUTPUT}" > "${REPORT_DIR}/${HANDLER}_handler_tests_${TIMESTAMP}.log"

    if [ ${TEST_RESULT} -eq 0 ]; then
        print_success "${HANDLER} Handler: ${PASS_COUNT}/${TEST_COUNT} tests passed"
    else
        print_error "${HANDLER} Handler: ${FAIL_COUNT}/${TEST_COUNT} tests failed"
        echo "${TEST_OUTPUT}" | grep "FAIL:" || true
    fi
    echo ""
done

# Run WebSocket integration tests
print_header "Running WebSocket Integration Tests"
echo ""

echo -e "${BLUE}Testing WebSocket Manager Integration...${NC}"
WS_OUTPUT=$(go test -v -run "TestWebSocket.*Integration" ${WEBSOCKET_TEST_DIR} 2>&1)
WS_RESULT=$?

WS_TEST_COUNT=$(echo "${WS_OUTPUT}" | grep -c "^=== RUN" || true)
WS_PASS_COUNT=$(echo "${WS_OUTPUT}" | grep -c "^--- PASS:" || true)
WS_FAIL_COUNT=$(echo "${WS_OUTPUT}" | grep -c "^--- FAIL:" || true)

TOTAL_TESTS=$((TOTAL_TESTS + WS_TEST_COUNT))
PASSED_TESTS=$((PASSED_TESTS + WS_PASS_COUNT))
FAILED_TESTS=$((FAILED_TESTS + WS_FAIL_COUNT))

# Save output
echo "${WS_OUTPUT}" > "${REPORT_DIR}/websocket_integration_tests_${TIMESTAMP}.log"

if [ ${WS_RESULT} -eq 0 ]; then
    print_success "WebSocket Integration: ${WS_PASS_COUNT}/${WS_TEST_COUNT} tests passed"
else
    print_error "WebSocket Integration: ${WS_FAIL_COUNT}/${WS_TEST_COUNT} tests failed"
    echo "${WS_OUTPUT}" | grep "FAIL:" || true
fi
echo ""

# Run all event tests with coverage
print_header "Generating Coverage Report"
echo ""

go test -coverprofile="${COVERAGE_DIR}/coverage_${TIMESTAMP}.out" \
    -covermode=atomic \
    ${TEST_DIR} ${WEBSOCKET_TEST_DIR} > /dev/null 2>&1

# Generate coverage percentage
COVERAGE=$(go tool cover -func="${COVERAGE_DIR}/coverage_${TIMESTAMP}.out" | grep total | awk '{print $3}')
echo -e "Overall Coverage: ${BLUE}${COVERAGE}${NC}"

# Generate HTML coverage report
go tool cover -html="${COVERAGE_DIR}/coverage_${TIMESTAMP}.out" \
    -o "${COVERAGE_DIR}/coverage_${TIMESTAMP}.html"

print_success "Coverage report generated: ${COVERAGE_DIR}/coverage_${TIMESTAMP}.html"
echo ""

# Generate summary report
print_header "Test Summary"
echo ""

echo -e "Total Tests:  ${BLUE}${TOTAL_TESTS}${NC}"
echo -e "Passed:       ${GREEN}${PASSED_TESTS}${NC}"
echo -e "Failed:       ${RED}${FAILED_TESTS}${NC}"
echo -e "Coverage:     ${BLUE}${COVERAGE}${NC}"
echo ""

# Calculate success rate
if [ ${TOTAL_TESTS} -gt 0 ]; then
    SUCCESS_RATE=$((PASSED_TESTS * 100 / TOTAL_TESTS))
    echo -e "Success Rate: ${BLUE}${SUCCESS_RATE}%${NC}"
else
    echo -e "Success Rate: ${RED}N/A${NC}"
fi
echo ""

# Save summary to file
cat > "${REPORT_DIR}/summary_${TIMESTAMP}.txt" <<EOF
Event Publishing Test Summary
Generated: $(date)

Total Tests:  ${TOTAL_TESTS}
Passed:       ${PASSED_TESTS}
Failed:       ${FAILED_TESTS}
Coverage:     ${COVERAGE}
Success Rate: ${SUCCESS_RATE}%

Handler Tests:
$(for HANDLER in "${HANDLERS[@]}"; do echo "  - ${HANDLER}"; done)

WebSocket Integration Tests: ${WS_TEST_COUNT} tests

Reports saved in: ${REPORT_DIR}
Coverage reports: ${COVERAGE_DIR}
EOF

print_success "Summary report saved: ${REPORT_DIR}/summary_${TIMESTAMP}.txt"
echo ""

# Exit with appropriate code
if [ ${FAILED_TESTS} -gt 0 ]; then
    print_error "Some tests failed. Please review the logs."
    exit 1
else
    print_success "All tests passed!"
    exit 0
fi
