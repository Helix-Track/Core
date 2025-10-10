#!/bin/bash
# Comprehensive API test suite for HelixTrack Core
# Tests all 234+ API endpoints organized by feature group

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BASE_URL="${BASE_URL:-http://localhost:8080}"
export BASE_URL

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo "======================================================================="
echo -e "${BLUE}HelixTrack Core - Comprehensive API Test Suite${NC}"
echo "======================================================================="
echo "Base URL: $BASE_URL"
echo "JWT Token: ${JWT_TOKEN:+[SET]}${JWT_TOKEN:-[NOT SET]}"
echo "======================================================================="
echo ""

# Test counters
total=0
passed=0
failed=0

run_test() {
    local name="$1"
    local script="${SCRIPT_DIR}/$name"
    [ ! -f "$script" ] && return
    total=$((total + 1))
    echo -e "${BLUE}Running: $name${NC}"
    if bash "$script"; then
        passed=$((passed + 1))
        echo -e "${GREEN}✓ PASSED${NC}\n"
    else
        failed=$((failed + 1))
        echo -e "${RED}✗ FAILED${NC}\n"
    fi
}

# PUBLIC ENDPOINTS
echo -e "${YELLOW}━━━ PUBLIC ENDPOINTS ━━━${NC}"
run_test "test-version.sh"
run_test "test-jwt-capable.sh"
run_test "test-db-capable.sh"
run_test "test-health.sh"

# AUTHENTICATION
echo -e "${YELLOW}━━━ AUTHENTICATION ━━━${NC}"
run_test "test-authenticate.sh"

# GENERIC CRUD
echo -e "${YELLOW}━━━ GENERIC CRUD ━━━${NC}"
run_test "test-create.sh"

# PHASE 1 FEATURES
echo -e "${YELLOW}━━━ PHASE 1 - JIRA PARITY ━━━${NC}"
run_test "test-priority.sh"
run_test "test-resolution.sh"
run_test "test-watcher.sh"
run_test "test-filter.sh"
run_test "test-customfield.sh"

# WORKFLOW ENGINE
echo -e "${YELLOW}━━━ WORKFLOW ENGINE ━━━${NC}"
run_test "test-workflow.sh"
run_test "test-ticket-status.sh"
run_test "test-ticket-type.sh"

# AGILE/SCRUM
echo -e "${YELLOW}━━━ AGILE/SCRUM SUPPORT ━━━${NC}"
run_test "test-board.sh"
run_test "test-cycle.sh"

# MULTI-TENANCY
echo -e "${YELLOW}━━━ MULTI-TENANCY ━━━${NC}"
run_test "test-account.sh"
run_test "test-organization.sh"
run_test "test-team.sh"

# SUPPORTING SYSTEMS
echo -e "${YELLOW}━━━ SUPPORTING SYSTEMS ━━━${NC}"
run_test "test-component.sh"
run_test "test-label.sh"
run_test "test-asset.sh"

# GIT INTEGRATION
echo -e "${YELLOW}━━━ GIT INTEGRATION ━━━${NC}"
run_test "test-repository.sh"

# TICKET RELATIONSHIPS
echo -e "${YELLOW}━━━ TICKET RELATIONSHIPS ━━━${NC}"
run_test "test-ticket-relationship.sh"

# SYSTEM INFRASTRUCTURE
echo -e "${YELLOW}━━━ SYSTEM INFRASTRUCTURE ━━━${NC}"
run_test "test-permission.sh"
run_test "test-audit.sh"
run_test "test-report.sh"
run_test "test-extension.sh"

# SUMMARY
echo "======================================================================="
echo -e "${BLUE}TEST SUITE SUMMARY${NC}"
echo "======================================================================="
echo "Total: $total"
echo -e "${GREEN}Passed: $passed${NC}"
echo -e "${RED}Failed: $failed${NC}"
echo ""
[ $failed -eq 0 ] && echo -e "${GREEN}✓ ALL TESTS PASSED!${NC}" && exit 0
echo -e "${RED}✗ SOME TESTS FAILED!${NC}"
exit 1
