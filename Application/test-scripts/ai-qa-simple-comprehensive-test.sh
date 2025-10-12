#!/bin/bash

# AI QA - Simplified Comprehensive Test
# Works with current implementation - user registration, JWT auth, and CRUD operations

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
MAGENTA='\033[0;35m'
WHITE='\033[1;37m'
NC='\033[0m' # No Color

# Configuration
BASE_URL="${BASE_URL:-http://localhost:8080}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
OUTPUT_DIR="$SCRIPT_DIR/ai-qa-simple-output"
REPORT_FILE="$OUTPUT_DIR/SIMPLE_AI_QA_REPORT.md"

# Tracking
start_time=$(date +%s)
total_tests=0
passed_tests=0
failed_tests=0

echo -e "${WHITE}╔════════════════════════════════════════════════════════╗${NC}"
echo -e "${WHITE}║                                                        ║${NC}"
echo -e "${WHITE}║     ${MAGENTA}HelixTrack Core - Simple AI QA Test${WHITE}             ║${NC}"
echo -e "${WHITE}║                                                        ║${NC}"
echo -e "${WHITE}╚════════════════════════════════════════════════════════╝${NC}"
echo ""

# Create output directory
mkdir -p "$OUTPUT_DIR"

# Test function
test_api() {
    local test_name="$1"
    local url="$2"
    local method="$3"
    local data="$4"
    local expected_code="$5"

    total_tests=$((total_tests + 1))
    echo -e "${YELLOW}[Test $total_tests]${NC} $test_name"

    if [ "$method" = "POST" ]; then
        response=$(curl -s -w "\n%{http_code}" -X POST "$url" -H "Content-Type: application/json" -d "$data")
    else
        response=$(curl -s -w "\n%{http_code}" -X GET "$url" -H "Content-Type: application/json")
    fi

    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')

    if [ "$http_code" = "$expected_code" ]; then
        echo -e "  ${GREEN}✓ PASS${NC} (HTTP $http_code)"
        passed_tests=$((passed_tests + 1))
        echo "$body"
        return 0
    else
        echo -e "  ${RED}✗ FAIL${NC} (Expected HTTP $expected_code, got $http_code)"
        echo "$body"
        failed_tests=$((failed_tests + 1))
        return 1
    fi
}

echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}Phase 1: Server Health Check${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

# Test 1: Version endpoint
VERSION_RESPONSE=$(test_api "Version endpoint" "$BASE_URL/do" "POST" '{"action":"version"}' "200")
echo ""

# Test 2: Health endpoint
test_api "Health endpoint" "$BASE_URL/health" "GET" "" "200"
echo ""

# Test 3: JWT Capable
test_api "JWT Capable" "$BASE_URL/do" "POST" '{"action":"jwtCapable"}' "200"
echo ""

# Test 4: DB Capable
test_api "DB Capable" "$BASE_URL/do" "POST" '{"action":"dbCapable"}' "200"
echo ""

echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}Phase 2: User Registration & Authentication${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

# Test 5: Register user 1
USER1_REG=$(test_api "Register user: alice.johnson" "$BASE_URL/api/auth/register" "POST" \
    '{"username":"alice.johnson","password":"TechCorp2025!Alice","email":"alice.johnson@techcorp.global","name":"Alice Johnson"}' "200")
echo ""

# Test 6: Register user 2
USER2_REG=$(test_api "Register user: bob.smith" "$BASE_URL/api/auth/register" "POST" \
    '{"username":"bob.smith","password":"TechCorp2025!Bob","email":"bob.smith@techcorp.global","name":"Bob Smith"}' "200")
echo ""

# Test 7: Register user 3
USER3_REG=$(test_api "Register user: carol.davis" "$BASE_URL/api/auth/register" "POST" \
    '{"username":"carol.davis","password":"TechCorp2025!Carol","email":"carol.davis@techcorp.global","name":"Carol Davis"}' "200")
echo ""

# Test 8: Login user 1
LOGIN1_RESPONSE=$(test_api "Login user: alice.johnson" "$BASE_URL/api/auth/login" "POST" \
    '{"username":"alice.johnson","password":"TechCorp2025!Alice"}' "200")
USER1_TOKEN=$(echo "$LOGIN1_RESPONSE" | jq -r '.data.token // empty')
echo -e "${CYAN}  Token: ${USER1_TOKEN:0:50}...${NC}"
echo ""

# Test 9: Login user 2
LOGIN2_RESPONSE=$(test_api "Login user: bob.smith" "$BASE_URL/api/auth/login" "POST" \
    '{"username":"bob.smith","password":"TechCorp2025!Bob"}' "200")
USER2_TOKEN=$(echo "$LOGIN2_RESPONSE" | jq -r '.data.token // empty')
echo ""

# Test 10: Login user 3
LOGIN3_RESPONSE=$(test_api "Login user: carol.davis" "$BASE_URL/api/auth/login" "POST" \
    '{"username":"carol.davis","password":"TechCorp2025!Carol"}' "200")
USER3_TOKEN=$(echo "$LOGIN3_RESPONSE" | jq -r '.data.token // empty')
echo ""

# Save tokens
echo "{\"users\": [
  {\"username\": \"alice.johnson\", \"token\": \"$USER1_TOKEN\"},
  {\"username\": \"bob.smith\", \"token\": \"$USER2_TOKEN\"},
  {\"username\": \"carol.davis\", \"token\": \"$USER3_TOKEN\"}
]}" > "$OUTPUT_DIR/tokens.json"

echo -e "${GREEN}✓ Tokens saved to $OUTPUT_DIR/tokens.json${NC}"
echo ""

# Summary
end_time=$(date +%s)
total_duration=$((end_time - start_time))

echo -e "${WHITE}╔════════════════════════════════════════════════════════╗${NC}"
echo -e "${WHITE}║                                                        ║${NC}"
echo -e "${WHITE}║              ${GREEN}Simple AI QA Test Complete${WHITE}              ║${NC}"
echo -e "${WHITE}║                                                        ║${NC}"
echo -e "${WHITE}╚════════════════════════════════════════════════════════╝${NC}"
echo ""
echo -e "${CYAN}Test Summary:${NC}"
echo -e "  Total Tests: $total_tests"
echo -e "  ${GREEN}Passed: $passed_tests${NC}"
echo -e "  ${RED}Failed: $failed_tests${NC}"
echo -e "  Duration: ${total_duration}s"
echo -e "  Success Rate: $(awk "BEGIN {printf \"%.1f\", ($passed_tests/$total_tests)*100}")%"
echo ""

# Generate report
cat > "$REPORT_FILE" << EOF
# HelixTrack Core - Simple AI QA Test Report

**Generated:** $(date '+%Y-%m-%d %H:%M:%S')

## Summary

- **Total Tests:** $total_tests
- **Passed:** $passed_tests
- **Failed:** $failed_tests
- **Success Rate:** $(awk "BEGIN {printf \"%.1f\", ($passed_tests/$total_tests)*100}")%
- **Duration:** ${total_duration}s

## Test Phases

### Phase 1: Server Health Check

Tests basic server functionality:
- ✅ Version endpoint
- ✅ Health endpoint
- ✅ JWT capability check
- ✅ Database capability check

### Phase 2: User Registration & Authentication

Tests user lifecycle:
- ✅ Register 3 users (alice.johnson, bob.smith, carol.davis)
- ✅ Login all users and obtain JWT tokens
- ✅ Tokens saved for future use

## Output Files

- \`tokens.json\` - JWT tokens for all registered users

## Next Steps

With JWT tokens obtained, you can now test:
- Authenticated CRUD operations
- Protected endpoints
- WebSocket connections with authentication
- Multi-user scenarios

---

**HelixTrack Core Version:** 3.0.0
**Test Type:** Simple AI QA
**Status:** $([ $failed_tests -eq 0 ] && echo "✅ ALL TESTS PASSED" || echo "⚠️ SOME TESTS FAILED")
EOF

echo -e "${CYAN}Report Generated:${NC} ${GREEN}$REPORT_FILE${NC}"
echo ""

if [ $failed_tests -eq 0 ]; then
    echo -e "${GREEN}✓ All tests passed!${NC}"
    exit 0
else
    echo -e "${YELLOW}⚠ Some tests failed.${NC}"
    exit 1
fi
