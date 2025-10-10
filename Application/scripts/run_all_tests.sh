#!/bin/bash
# HelixTrack Core - Comprehensive Test Runner
# This script runs all unit, integration, and end-to-end tests

set -e  # Exit on error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

echo -e "${BLUE}=============================================${NC}"
echo -e "${BLUE}  HelixTrack Core - Comprehensive Testing  ${NC}"
echo -e "${BLUE}=============================================${NC}"
echo ""

# Change to project root
cd "$PROJECT_ROOT"

# Function to run tests with coverage
run_tests_with_coverage() {
    local test_type=$1
    local test_path=$2
    local coverage_file=$3

    echo -e "${YELLOW}Running ${test_type} tests...${NC}"
    echo "Path: $test_path"
    echo ""

    if go test "$test_path" -v -cover -coverprofile="$coverage_file" -race -timeout 30s; then
        echo -e "${GREEN}✓ ${test_type} tests PASSED${NC}"
        echo ""

        # Show coverage summary
        go tool cover -func="$coverage_file" | tail -n 1
        echo ""
        return 0
    else
        echo -e "${RED}✗ ${test_type} tests FAILED${NC}"
        echo ""
        return 1
    fi
}

# Create coverage directory
COVERAGE_DIR="$PROJECT_ROOT/coverage"
mkdir -p "$COVERAGE_DIR"

# Track overall status
OVERALL_STATUS=0

# 1. Run Unit Tests
echo -e "${BLUE}>>> PHASE 1: Unit Tests${NC}"
echo ""

if run_tests_with_coverage "Unit" "./..." "$COVERAGE_DIR/unit.out"; then
    echo -e "${GREEN}Unit tests completed successfully${NC}"
else
    echo -e "${RED}Unit tests failed${NC}"
    OVERALL_STATUS=1
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# 2. Run Integration Tests
echo -e "${BLUE}>>> PHASE 2: Integration Tests${NC}"
echo ""

if run_tests_with_coverage "Integration" "./tests/integration/..." "$COVERAGE_DIR/integration.out"; then
    echo -e "${GREEN}Integration tests completed successfully${NC}"
else
    echo -e "${RED}Integration tests failed${NC}"
    OVERALL_STATUS=1
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# 3. Run End-to-End Tests
echo -e "${BLUE}>>> PHASE 3: End-to-End Tests${NC}"
echo ""

if run_tests_with_coverage "E2E" "./tests/e2e/..." "$COVERAGE_DIR/e2e.out"; then
    echo -e "${GREEN}E2E tests completed successfully${NC}"
else
    echo -e "${RED}E2E tests failed${NC}"
    OVERALL_STATUS=1
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# 4. Generate Combined Coverage Report
echo -e "${BLUE}>>> Generating Combined Coverage Report${NC}"
echo ""

# Combine all coverage files
echo "mode: atomic" > "$COVERAGE_DIR/combined.out"
for file in "$COVERAGE_DIR"/*.out; do
    if [ "$file" != "$COVERAGE_DIR/combined.out" ]; then
        tail -n +2 "$file" >> "$COVERAGE_DIR/combined.out" 2>/dev/null || true
    fi
done

# Generate coverage percentage
COVERAGE=$(go tool cover -func="$COVERAGE_DIR/combined.out" | grep total | awk '{print $3}')

echo -e "${BLUE}Total Coverage: ${GREEN}${COVERAGE}${NC}"
echo ""

# Generate HTML coverage report
go tool cover -html="$COVERAGE_DIR/combined.out" -o "$COVERAGE_DIR/coverage.html"
echo -e "${GREEN}HTML coverage report: ${COVERAGE_DIR}/coverage.html${NC}"
echo ""

# 5. Run Benchmarks
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo -e "${BLUE}>>> PHASE 4: Running Benchmarks${NC}"
echo ""

go test ./... -bench=. -benchmem -run=^$ | tee "$COVERAGE_DIR/benchmarks.txt"

echo ""
echo -e "${GREEN}Benchmark results saved to: ${COVERAGE_DIR}/benchmarks.txt${NC}"
echo ""

# 6. Final Summary
echo ""
echo "=============================================="
echo ""

if [ $OVERALL_STATUS -eq 0 ]; then
    echo -e "${GREEN}✓✓✓ ALL TESTS PASSED ✓✓✓${NC}"
    echo ""
    echo -e "${GREEN}Total Coverage: ${COVERAGE}${NC}"
    echo -e "${GREEN}Status: READY FOR PRODUCTION${NC}"
else
    echo -e "${RED}✗✗✗ SOME TESTS FAILED ✗✗✗${NC}"
    echo ""
    echo -e "${RED}Please review failed tests above${NC}"
fi

echo ""
echo "=============================================="
echo ""

# Test Statistics
echo -e "${BLUE}Test Statistics:${NC}"
echo "  - Unit Test Coverage: $(grep -h 'coverage:' "$COVERAGE_DIR/unit.out" | tail -1 || echo 'N/A')"
echo "  - Integration Test Coverage: $(grep -h 'coverage:' "$COVERAGE_DIR/integration.out" | tail -1 || echo 'N/A')"
echo "  - E2E Test Coverage: $(grep -h 'coverage:' "$COVERAGE_DIR/e2e.out" | tail -1 || echo 'N/A')"
echo "  - Combined Coverage: $COVERAGE"
echo ""

exit $OVERALL_STATUS
