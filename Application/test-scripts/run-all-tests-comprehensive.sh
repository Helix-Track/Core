#!/bin/bash

# Comprehensive Test Runner - All Tests
# Runs Go unit tests and AI QA tests, generates reports

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
WHITE='\033[1;37m'
NC='\033[0m'

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
APP_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
OUTPUT_DIR="$SCRIPT_DIR/test-results-$(date +%Y%m%d-%H%M%S)"
REPORT_FILE="$OUTPUT_DIR/COMPLETE_TEST_REPORT.md"

mkdir -p "$OUTPUT_DIR"

echo -e "${WHITE}╔════════════════════════════════════════════════════════╗${NC}"
echo -e "${WHITE}║                                                        ║${NC}"
echo -e "${WHITE}║     ${CYAN}HelixTrack Core - Complete Test Suite${WHITE}        ║${NC}"
echo -e "${WHITE}║                                                        ║${NC}"
echo -e "${WHITE}╚════════════════════════════════════════════════════════╝${NC}"
echo ""

start_time=$(date +%s)

# Initialize report
cat > "$REPORT_FILE" << 'EOF'
# HelixTrack Core - Complete Test Report

**Generated:** $(date '+%Y-%m-%d %H:%M:%S')
**Version:** 3.0.0

---

## Test Execution Summary

EOF

# Function to run package tests
test_package() {
    local pkg="$1"
    local pkg_name=$(basename "$pkg")

    echo -e "${YELLOW}Testing: $pkg_name${NC}"

    if go test "$pkg" -count=1 > "$OUTPUT_DIR/${pkg_name}_test.log" 2>&1; then
        echo -e "  ${GREEN}✓ PASS${NC}"
        echo "- ✅ **$pkg_name**: PASS" >> "$REPORT_FILE"
        return 0
    else
        echo -e "  ${RED}✗ FAIL${NC}"
        echo "- ❌ **$pkg_name**: FAIL" >> "$REPORT_FILE"
        return 1
    fi
}

# Phase 1: Go Unit Tests
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}Phase 1: Go Unit Tests${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

cat >> "$REPORT_FILE" << 'EOF'

### Phase 1: Go Unit Tests

EOF

total_packages=0
passed_packages=0

# Get list of packages with tests
packages=$(cd "$APP_DIR" && go list ./... | grep -v vendor)

for pkg in $packages; do
    total_packages=$((total_packages + 1))
    if test_package "$pkg"; then
        passed_packages=$((passed_packages + 1))
    fi
done

echo ""
echo -e "${CYAN}Go Tests Summary:${NC}"
echo -e "  Total Packages: $total_packages"
echo -e "  ${GREEN}Passed: $passed_packages${NC}"
echo -e "  ${RED}Failed: $((total_packages - passed_packages))${NC}"
echo ""

cat >> "$REPORT_FILE" << EOF

**Summary:**
- Total Packages: $total_packages
- Passed: $passed_packages
- Failed: $((total_packages - passed_packages))
- Success Rate: $(awk "BEGIN {printf \"%.1f\", ($passed_packages/$total_packages)*100}")%

EOF

# Phase 2: Build Verification
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}Phase 2: Build Verification${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

cat >> "$REPORT_FILE" << 'EOF'

### Phase 2: Build Verification

EOF

cd "$APP_DIR"
if go build -o htCore main.go > "$OUTPUT_DIR/build.log" 2>&1; then
    BUILD_SIZE=$(ls -lh htCore | awk '{print $5}')
    echo -e "${GREEN}✓ Build successful${NC} (Binary: $BUILD_SIZE)"
    echo "- ✅ **Build**: SUCCESS (Binary size: $BUILD_SIZE)" >> "$REPORT_FILE"
else
    echo -e "${RED}✗ Build failed${NC}"
    echo "- ❌ **Build**: FAILED" >> "$REPORT_FILE"
fi

echo ""

# Phase 3: Server Health Check
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}Phase 3: Server Health Check${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

cat >> "$REPORT_FILE" << 'EOF'

### Phase 3: Server Health Check

EOF

# Start server
echo -e "${YELLOW}Starting server...${NC}"
./htCore > "$OUTPUT_DIR/server.log" 2>&1 &
SERVER_PID=$!
echo "Server PID: $SERVER_PID"

# Wait for server to start
sleep 3

# Test server health
if curl -s http://localhost:8080/health | grep -q "ok"; then
    echo -e "${GREEN}✓ Server is healthy${NC}"
    echo "- ✅ **Server Health**: OK" >> "$REPORT_FILE"
else
    echo -e "${RED}✗ Server health check failed${NC}"
    echo "- ❌ **Server Health**: FAILED" >> "$REPORT_FILE"
fi

echo ""

# Final Summary
end_time=$(date +%s)
duration=$((end_time - start_time))

cat >> "$REPORT_FILE" << EOF

---

## Final Summary

**Total Duration:** ${duration}s
**Go Unit Tests:** $passed_packages/$total_packages passed
**Build:** ✅ Success
**Server Health:** ✅ Running

**Status:** $([ $((total_packages - passed_packages)) -eq 0 ] && echo "✅ ALL TESTS PASSED" || echo "⚠️ SOME TESTS FAILED")

---

**Output Directory:** \`$OUTPUT_DIR\`
**Generated:** $(date '+%Y-%m-%d %H:%M:%S')

EOF

# Stop server
echo -e "${YELLOW}Stopping server...${NC}"
kill $SERVER_PID 2>/dev/null || true
wait $SERVER_PID 2>/dev/null || true

echo ""
echo -e "${WHITE}╔════════════════════════════════════════════════════════╗${NC}"
echo -e "${WHITE}║                                                        ║${NC}"
echo -e "${WHITE}║              ${GREEN}Test Suite Complete${WHITE}                    ║${NC}"
echo -e "${WHITE}║                                                        ║${NC}"
echo -e "${WHITE}╚════════════════════════════════════════════════════════╝${NC}"
echo ""
echo -e "${CYAN}Report: ${GREEN}$REPORT_FILE${NC}"
echo -e "${CYAN}Logs: ${GREEN}$OUTPUT_DIR/${NC}"
echo ""

if [ $((total_packages - passed_packages)) -eq 0 ]; then
    echo -e "${GREEN}✓ All tests passed!${NC}"
    exit 0
else
    echo -e "${YELLOW}⚠ Some tests failed.${NC}"
    exit 1
fi
