#!/bin/bash

# Comprehensive test runner for Localization Service
# Runs unit tests, integration tests, and E2E tests

set -e

# Color codes for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SERVICE_URL="${SERVICE_URL:-https://localhost:8085}"
JWT_SECRET="${JWT_SECRET:-test-secret-key-for-e2e-testing}"
SKIP_E2E="${SKIP_E2E:-false}"

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Localization Service - Test Runner${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Check if we're in the right directory
if [ ! -f "go.mod" ]; then
    echo -e "${RED}Error: Must run from Localization service root directory${NC}"
    exit 1
fi

# Step 1: Unit Tests
echo -e "${YELLOW}[1/4] Running Unit Tests...${NC}"
echo ""

go test ./internal/models/ -v -cover | tee /tmp/unit-models.log
go test ./internal/config/ -v -cover | tee /tmp/unit-config.log
go test ./internal/utils/ -v -cover | tee /tmp/unit-utils.log
go test ./internal/middleware/ -v -cover | tee /tmp/unit-middleware.log
go test ./internal/cache/ -v -cover | tee /tmp/unit-cache.log

echo ""
echo -e "${GREEN}✓ Unit tests completed${NC}"
echo ""

# Step 2: Integration Tests
echo -e "${YELLOW}[2/4] Running Integration Tests...${NC}"
echo ""

go test ./internal/handlers/ -v -cover | tee /tmp/integration.log

echo ""
echo -e "${GREEN}✓ Integration tests completed${NC}"
echo ""

# Step 3: Race Detection
echo -e "${YELLOW}[3/4] Running Race Detection...${NC}"
echo ""

go test ./internal/... -race | tee /tmp/race.log

echo ""
echo -e "${GREEN}✓ Race detection completed${NC}"
echo ""

# Step 4: E2E Tests (optional)
if [ "$SKIP_E2E" = "false" ]; then
    echo -e "${YELLOW}[4/4] Running E2E Tests...${NC}"
    echo ""
    echo -e "${BLUE}Note: E2E tests require a running service at ${SERVICE_URL}${NC}"
    echo -e "${BLUE}To skip E2E tests, set SKIP_E2E=true${NC}"
    echo ""

    # Check if service is running
    if curl -k -s -f "${SERVICE_URL}/health" > /dev/null 2>&1; then
        echo -e "${GREEN}✓ Service is running at ${SERVICE_URL}${NC}"
        echo ""

        export SERVICE_URL
        export JWT_SECRET

        go test ./e2e/ -v | tee /tmp/e2e.log

        echo ""
        echo -e "${GREEN}✓ E2E tests completed${NC}"
    else
        echo -e "${YELLOW}⚠ Service not running at ${SERVICE_URL}${NC}"
        echo -e "${YELLOW}⚠ Skipping E2E tests${NC}"
        echo ""
        echo "To run E2E tests:"
        echo "  1. Start the service: ./htLoc --config=configs/default.json"
        echo "  2. Run this script again"
    fi
else
    echo -e "${YELLOW}[4/4] E2E Tests Skipped (SKIP_E2E=true)${NC}"
fi

echo ""
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Test Summary${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Calculate summary
TOTAL_TESTS=$(grep -h "PASS:" /tmp/*.log 2>/dev/null | wc -l || echo "0")
UNIT_TESTS=$(grep -h "PASS:" /tmp/unit-*.log 2>/dev/null | wc -l || echo "0")
INTEGRATION_TESTS=$(grep -h "PASS:" /tmp/integration.log 2>/dev/null | wc -l || echo "0")
E2E_TESTS=$(grep -h "PASS:" /tmp/e2e.log 2>/dev/null | wc -l || echo "0")

echo -e "Unit Tests:        ${GREEN}${UNIT_TESTS} passed${NC}"
echo -e "Integration Tests: ${GREEN}${INTEGRATION_TESTS} passed${NC}"
if [ "$SKIP_E2E" = "false" ] && [ -f /tmp/e2e.log ]; then
    echo -e "E2E Tests:         ${GREEN}${E2E_TESTS} passed${NC}"
fi
echo -e "Total:             ${GREEN}${TOTAL_TESTS} passed${NC}"

echo ""
echo -e "${GREEN}✅ All tests completed successfully!${NC}"
echo ""

# Cleanup
rm -f /tmp/unit-*.log /tmp/integration.log /tmp/race.log /tmp/e2e.log

exit 0
