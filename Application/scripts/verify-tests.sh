#!/bin/bash
# Comprehensive test verification and reporting script

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
REPORTS_DIR="$PROJECT_ROOT/test-reports"
COVERAGE_DIR="$PROJECT_ROOT/coverage"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
MAGENTA='\033[0;35m'
NC='\033[0m'

# Create directories
mkdir -p "$REPORTS_DIR"
mkdir -p "$COVERAGE_DIR"

echo -e "${BLUE}╔════════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║     HelixTrack Core - Comprehensive Test Verification         ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════════════════════════╝${NC}"
echo ""

cd "$PROJECT_ROOT"

# Check Go installation
if ! command -v go &> /dev/null; then
    echo -e "${RED}✗ Error: Go is not installed${NC}"
    echo ""
    echo "Please install Go 1.22 or higher:"
    echo "  - Ubuntu/Debian: sudo apt-get install golang-1.22"
    echo "  - macOS: brew install go"
    echo "  - Download: https://golang.org/dl/"
    exit 1
fi

GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
echo -e "${GREEN}✓ Go ${GO_VERSION} detected${NC}"
echo ""

# Verify go.mod
if [ ! -f "go.mod" ]; then
    echo -e "${RED}✗ Error: go.mod not found${NC}"
    exit 1
fi
echo -e "${GREEN}✓ go.mod verified${NC}"

# Download dependencies
echo -e "${YELLOW}→ Downloading dependencies...${NC}"
go mod download 2>&1 | tee "$REPORTS_DIR/dependencies.log"
if [ ${PIPESTATUS[0]} -eq 0 ]; then
    echo -e "${GREEN}✓ Dependencies downloaded${NC}"
else
    echo -e "${RED}✗ Failed to download dependencies${NC}"
    exit 1
fi
echo ""

# List all test packages
echo -e "${CYAN}═══════════════════════════════════════════════════════════════${NC}"
echo -e "${CYAN}Test Packages Discovery${NC}"
echo -e "${CYAN}═══════════════════════════════════════════════════════════════${NC}"

TEST_PACKAGES=$(go list ./... 2>/dev/null)
PACKAGE_COUNT=$(echo "$TEST_PACKAGES" | wc -l)

echo -e "${BLUE}Found ${PACKAGE_COUNT} packages:${NC}"
echo "$TEST_PACKAGES" | while read pkg; do
    echo -e "  ${GREEN}•${NC} $pkg"
done
echo ""

# Run tests with verbose output and coverage
echo -e "${CYAN}═══════════════════════════════════════════════════════════════${NC}"
echo -e "${CYAN}Running Tests${NC}"
echo -e "${CYAN}═══════════════════════════════════════════════════════════════${NC}"
echo ""

START_TIME=$(date +%s)

# Run tests and capture output
go test -v -race -coverprofile="$COVERAGE_DIR/coverage.out" -covermode=atomic ./... 2>&1 | tee "$REPORTS_DIR/test-output-verbose.txt"
TEST_EXIT_CODE=${PIPESTATUS[0]}

END_TIME=$(date +%s)
DURATION=$((END_TIME - START_TIME))

echo ""

# Parse test results
if [ $TEST_EXIT_CODE -eq 0 ]; then
    echo -e "${GREEN}╔════════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║                    ALL TESTS PASSED ✓                          ║${NC}"
    echo -e "${GREEN}╚════════════════════════════════════════════════════════════════╝${NC}"
    TEST_STATUS="PASSED"
    STATUS_COLOR="${GREEN}"
else
    echo -e "${RED}╔════════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${RED}║                    TESTS FAILED ✗                              ║${NC}"
    echo -e "${RED}╚════════════════════════════════════════════════════════════════╝${NC}"
    TEST_STATUS="FAILED"
    STATUS_COLOR="${RED}"
fi
echo ""

# Calculate coverage
echo -e "${CYAN}═══════════════════════════════════════════════════════════════${NC}"
echo -e "${CYAN}Coverage Analysis${NC}"
echo -e "${CYAN}═══════════════════════════════════════════════════════════════${NC}"
echo ""

if [ -f "$COVERAGE_DIR/coverage.out" ]; then
    # Generate coverage report
    go tool cover -func="$COVERAGE_DIR/coverage.out" > "$REPORTS_DIR/coverage-detailed.txt"

    # Extract total coverage
    TOTAL_COVERAGE=$(go tool cover -func="$COVERAGE_DIR/coverage.out" | grep total | awk '{print $3}')
    COVERAGE_PERCENT=$(echo $TOTAL_COVERAGE | sed 's/%//')

    echo -e "${BLUE}Total Coverage: ${GREEN}${TOTAL_COVERAGE}${NC}"
    echo ""

    # Coverage by package
    echo -e "${YELLOW}Coverage by Package:${NC}"
    echo ""

    go tool cover -func="$COVERAGE_DIR/coverage.out" | grep -v "total:" | awk '{
        package=$1
        coverage=$3
        printf "  %-50s %s\n", package, coverage
    }' | head -20

    echo ""

    # Generate HTML coverage report
    go tool cover -html="$COVERAGE_DIR/coverage.out" -o "$COVERAGE_DIR/coverage.html"
    echo -e "${GREEN}✓ HTML coverage report: ${COVERAGE_DIR}/coverage.html${NC}"

    # Determine coverage quality
    COVERAGE_INT=$(printf "%.0f" "$COVERAGE_PERCENT")
    if [ "$COVERAGE_INT" -ge 90 ]; then
        COVERAGE_QUALITY="Excellent"
        QUALITY_COLOR="${GREEN}"
    elif [ "$COVERAGE_INT" -ge 80 ]; then
        COVERAGE_QUALITY="Good"
        QUALITY_COLOR="${GREEN}"
    elif [ "$COVERAGE_INT" -ge 70 ]; then
        COVERAGE_QUALITY="Acceptable"
        QUALITY_COLOR="${YELLOW}"
    else
        COVERAGE_QUALITY="Poor"
        QUALITY_COLOR="${RED}"
    fi
else
    TOTAL_COVERAGE="0%"
    COVERAGE_PERCENT="0"
    COVERAGE_QUALITY="Unknown"
    QUALITY_COLOR="${RED}"
fi
echo ""

# Count test cases
echo -e "${CYAN}═══════════════════════════════════════════════════════════════${NC}"
echo -e "${CYAN}Test Statistics${NC}"
echo -e "${CYAN}═══════════════════════════════════════════════════════════════${NC}"
echo ""

TOTAL_TESTS=$(grep -c "RUN" "$REPORTS_DIR/test-output-verbose.txt" || echo "0")
PASSED_TESTS=$(grep -c "PASS:" "$REPORTS_DIR/test-output-verbose.txt" || echo "0")
FAILED_TESTS=$(grep -c "FAIL:" "$REPORTS_DIR/test-output-verbose.txt" || echo "0")
SKIPPED_TESTS=$(grep -c "SKIP:" "$REPORTS_DIR/test-output-verbose.txt" || echo "0")

echo -e "  Total Test Cases:     ${BLUE}${TOTAL_TESTS}${NC}"
echo -e "  Passed:               ${GREEN}${PASSED_TESTS}${NC}"
echo -e "  Failed:               ${RED}${FAILED_TESTS}${NC}"
echo -e "  Skipped:              ${YELLOW}${SKIPPED_TESTS}${NC}"
echo -e "  Duration:             ${CYAN}${DURATION}s${NC}"
echo ""

# Generate detailed test report
echo -e "${CYAN}═══════════════════════════════════════════════════════════════${NC}"
echo -e "${CYAN}Generating Detailed Reports${NC}"
echo -e "${CYAN}═══════════════════════════════════════════════════════════════${NC}"
echo ""

# Create JSON report
cat > "$REPORTS_DIR/test-results.json" << EOF
{
  "timestamp": "$(date -u +"%Y-%m-%dT%H:%M:%SZ")",
  "status": "${TEST_STATUS}",
  "go_version": "${GO_VERSION}",
  "duration_seconds": ${DURATION},
  "statistics": {
    "total_packages": ${PACKAGE_COUNT},
    "total_tests": ${TOTAL_TESTS},
    "passed": ${PASSED_TESTS},
    "failed": ${FAILED_TESTS},
    "skipped": ${SKIPPED_TESTS}
  },
  "coverage": {
    "total": "${TOTAL_COVERAGE}",
    "percent": ${COVERAGE_PERCENT},
    "quality": "${COVERAGE_QUALITY}"
  }
}
EOF

echo -e "${GREEN}✓ JSON report: ${REPORTS_DIR}/test-results.json${NC}"

# Create Markdown report
cat > "$REPORTS_DIR/TEST_REPORT.md" << EOF
# HelixTrack Core - Test Execution Report

**Generated:** $(date)
**Status:** ${TEST_STATUS}
**Go Version:** ${GO_VERSION}
**Duration:** ${DURATION} seconds

## Summary

| Metric | Value |
|--------|-------|
| Total Packages | ${PACKAGE_COUNT} |
| Total Tests | ${TOTAL_TESTS} |
| Passed | ${PASSED_TESTS} ✓ |
| Failed | ${FAILED_TESTS} |
| Skipped | ${SKIPPED_TESTS} |
| Coverage | ${TOTAL_COVERAGE} |
| Coverage Quality | ${COVERAGE_QUALITY} |

## Test Status

\`\`\`
${TEST_STATUS}
\`\`\`

## Coverage Details

Total coverage: **${TOTAL_COVERAGE}**

Coverage quality: **${COVERAGE_QUALITY}**

See detailed coverage report: [coverage.html](../coverage/coverage.html)

## Package Coverage

\`\`\`
$(go tool cover -func="$COVERAGE_DIR/coverage.out" 2>/dev/null || echo "Coverage data not available")
\`\`\`

## Test Output

Full test output available in: [test-output-verbose.txt](test-output-verbose.txt)

## Files Generated

- \`test-results.json\` - Machine-readable test results
- \`test-output-verbose.txt\` - Complete test output
- \`coverage-detailed.txt\` - Detailed coverage by function
- \`../coverage/coverage.out\` - Coverage profile
- \`../coverage/coverage.html\` - HTML coverage report

---

**Report Generated by:** HelixTrack Core Test Verification Script
**Date:** $(date -u +"%Y-%m-%dT%H:%M:%SZ")
EOF

echo -e "${GREEN}✓ Markdown report: ${REPORTS_DIR}/TEST_REPORT.md${NC}"

# Create HTML report
cat > "$REPORTS_DIR/TEST_REPORT.html" << EOF
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>HelixTrack Core - Test Report</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
            line-height: 1.6;
            max-width: 1200px;
            margin: 0 auto;
            padding: 20px;
            background-color: #f5f5f5;
        }
        .header {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 30px;
            border-radius: 10px;
            margin-bottom: 30px;
            box-shadow: 0 4px 6px rgba(0,0,0,0.1);
        }
        .header h1 {
            margin: 0;
            font-size: 2em;
        }
        .header p {
            margin: 10px 0 0 0;
            opacity: 0.9;
        }
        .status-badge {
            display: inline-block;
            padding: 8px 16px;
            border-radius: 20px;
            font-weight: bold;
            font-size: 1.1em;
            margin-top: 15px;
        }
        .status-passed {
            background-color: #10b981;
            color: white;
        }
        .status-failed {
            background-color: #ef4444;
            color: white;
        }
        .metrics {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 20px;
            margin-bottom: 30px;
        }
        .metric-card {
            background: white;
            padding: 20px;
            border-radius: 8px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        .metric-card h3 {
            margin: 0 0 10px 0;
            color: #666;
            font-size: 0.9em;
            text-transform: uppercase;
            letter-spacing: 0.5px;
        }
        .metric-value {
            font-size: 2em;
            font-weight: bold;
            color: #333;
        }
        .metric-passed { color: #10b981; }
        .metric-failed { color: #ef4444; }
        .metric-coverage { color: #3b82f6; }
        .section {
            background: white;
            padding: 25px;
            border-radius: 8px;
            margin-bottom: 20px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        .section h2 {
            margin-top: 0;
            color: #333;
            border-bottom: 2px solid #667eea;
            padding-bottom: 10px;
        }
        table {
            width: 100%;
            border-collapse: collapse;
            margin-top: 15px;
        }
        th, td {
            padding: 12px;
            text-align: left;
            border-bottom: 1px solid #e5e7eb;
        }
        th {
            background-color: #f9fafb;
            font-weight: 600;
            color: #374151;
        }
        tr:hover {
            background-color: #f9fafb;
        }
        .footer {
            text-align: center;
            color: #666;
            margin-top: 40px;
            padding-top: 20px;
            border-top: 1px solid #e5e7eb;
        }
        .progress-bar {
            width: 100%;
            height: 30px;
            background-color: #e5e7eb;
            border-radius: 15px;
            overflow: hidden;
            margin-top: 10px;
        }
        .progress-fill {
            height: 100%;
            background: linear-gradient(90deg, #10b981 0%, #059669 100%);
            display: flex;
            align-items: center;
            justify-content: center;
            color: white;
            font-weight: bold;
            transition: width 0.3s ease;
        }
    </style>
</head>
<body>
    <div class="header">
        <h1>🧪 HelixTrack Core - Test Execution Report</h1>
        <p>Generated: $(date)</p>
        <span class="status-badge status-${TEST_STATUS,,}">${TEST_STATUS}</span>
    </div>

    <div class="metrics">
        <div class="metric-card">
            <h3>Total Tests</h3>
            <div class="metric-value">${TOTAL_TESTS}</div>
        </div>
        <div class="metric-card">
            <h3>Passed</h3>
            <div class="metric-value metric-passed">${PASSED_TESTS}</div>
        </div>
        <div class="metric-card">
            <h3>Failed</h3>
            <div class="metric-value metric-failed">${FAILED_TESTS}</div>
        </div>
        <div class="metric-card">
            <h3>Coverage</h3>
            <div class="metric-value metric-coverage">${TOTAL_COVERAGE}</div>
        </div>
        <div class="metric-card">
            <h3>Duration</h3>
            <div class="metric-value">${DURATION}s</div>
        </div>
        <div class="metric-card">
            <h3>Packages</h3>
            <div class="metric-value">${PACKAGE_COUNT}</div>
        </div>
    </div>

    <div class="section">
        <h2>📊 Coverage Analysis</h2>
        <p><strong>Total Coverage:</strong> ${TOTAL_COVERAGE}</p>
        <p><strong>Quality:</strong> ${COVERAGE_QUALITY}</p>
        <div class="progress-bar">
            <div class="progress-fill" style="width: ${COVERAGE_PERCENT}%">${TOTAL_COVERAGE}</div>
        </div>
        <p style="margin-top: 20px;">
            <a href="../coverage/coverage.html" style="color: #667eea; text-decoration: none; font-weight: bold;">
                → View Detailed Coverage Report
            </a>
        </p>
    </div>

    <div class="section">
        <h2>📦 Test Summary</h2>
        <table>
            <tr>
                <th>Metric</th>
                <th>Value</th>
            </tr>
            <tr>
                <td>Go Version</td>
                <td>${GO_VERSION}</td>
            </tr>
            <tr>
                <td>Total Packages Tested</td>
                <td>${PACKAGE_COUNT}</td>
            </tr>
            <tr>
                <td>Total Test Cases</td>
                <td>${TOTAL_TESTS}</td>
            </tr>
            <tr>
                <td>Tests Passed</td>
                <td style="color: #10b981; font-weight: bold;">${PASSED_TESTS}</td>
            </tr>
            <tr>
                <td>Tests Failed</td>
                <td style="color: #ef4444; font-weight: bold;">${FAILED_TESTS}</td>
            </tr>
            <tr>
                <td>Tests Skipped</td>
                <td>${SKIPPED_TESTS}</td>
            </tr>
            <tr>
                <td>Execution Time</td>
                <td>${DURATION} seconds</td>
            </tr>
            <tr>
                <td>Coverage</td>
                <td style="color: #3b82f6; font-weight: bold;">${TOTAL_COVERAGE}</td>
            </tr>
        </table>
    </div>

    <div class="section">
        <h2>📁 Generated Files</h2>
        <ul>
            <li><code>test-results.json</code> - Machine-readable results</li>
            <li><code>test-output-verbose.txt</code> - Complete test output</li>
            <li><code>coverage-detailed.txt</code> - Detailed coverage report</li>
            <li><code>../coverage/coverage.out</code> - Coverage profile</li>
            <li><code>../coverage/coverage.html</code> - HTML coverage visualization</li>
        </ul>
    </div>

    <div class="footer">
        <p><strong>HelixTrack Core Test Verification System</strong></p>
        <p>Report generated: $(date -u +"%Y-%m-%dT%H:%M:%SZ")</p>
    </div>
</body>
</html>
EOF

echo -e "${GREEN}✓ HTML report: ${REPORTS_DIR}/TEST_REPORT.html${NC}"
echo ""

# Final summary
echo -e "${CYAN}═══════════════════════════════════════════════════════════════${NC}"
echo -e "${CYAN}Final Summary${NC}"
echo -e "${CYAN}═══════════════════════════════════════════════════════════════${NC}"
echo ""
echo -e "  Status:           ${STATUS_COLOR}${TEST_STATUS}${NC}"
echo -e "  Tests:            ${BLUE}${TOTAL_TESTS}${NC} (${GREEN}${PASSED_TESTS} passed${NC}, ${RED}${FAILED_TESTS} failed${NC})"
echo -e "  Coverage:         ${GREEN}${TOTAL_COVERAGE}${NC} (${QUALITY_COLOR}${COVERAGE_QUALITY}${NC})"
echo -e "  Duration:         ${CYAN}${DURATION}s${NC}"
echo -e "  Go Version:       ${BLUE}${GO_VERSION}${NC}"
echo ""
echo -e "${GREEN}Reports generated in: ${REPORTS_DIR}${NC}"
echo ""

# Open HTML report (optional)
if command -v xdg-open &> /dev/null; then
    echo -e "${YELLOW}Opening HTML report...${NC}"
    xdg-open "$REPORTS_DIR/TEST_REPORT.html" 2>/dev/null || true
elif command -v open &> /dev/null; then
    echo -e "${YELLOW}Opening HTML report...${NC}"
    open "$REPORTS_DIR/TEST_REPORT.html" 2>/dev/null || true
fi

# Exit with test status
exit $TEST_EXIT_CODE
