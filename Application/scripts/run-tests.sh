#!/bin/bash
# Run all tests and generate coverage badges

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
COVERAGE_DIR="$PROJECT_ROOT/coverage"
BADGES_DIR="$PROJECT_ROOT/docs/badges"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}HelixTrack Core - Test Execution${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Create coverage directory
mkdir -p "$COVERAGE_DIR"
mkdir -p "$BADGES_DIR"

cd "$PROJECT_ROOT"

# Check if go is installed
if ! command -v go &> /dev/null; then
    echo -e "${RED}Error: Go is not installed${NC}"
    exit 1
fi

echo -e "${YELLOW}Running tests...${NC}"

# Run tests with coverage
go test -v -coverprofile="$COVERAGE_DIR/coverage.out" -covermode=atomic ./... 2>&1 | tee "$COVERAGE_DIR/test-output.txt"

# Check if tests passed
if [ ${PIPESTATUS[0]} -ne 0 ]; then
    echo -e "${RED}Tests failed!${NC}"

    # Generate failed badge
    cat > "$BADGES_DIR/tests.svg" << 'EOF'
<svg xmlns="http://www.w3.org/2000/svg" width="80" height="20">
  <linearGradient id="b" x2="0" y2="100%">
    <stop offset="0" stop-color="#bbb" stop-opacity=".1"/>
    <stop offset="1" stop-opacity=".1"/>
  </linearGradient>
  <mask id="a"><rect width="80" height="20" rx="3" fill="#fff"/></mask>
  <g mask="url(#a)">
    <path fill="#555" d="M0 0h37v20H0z"/>
    <path fill="#e05d44" d="M37 0h43v20H37z"/>
    <path fill="url(#b)" d="M0 0h80v20H0z"/>
  </g>
  <g fill="#fff" text-anchor="middle" font-family="DejaVu Sans,Verdana,Geneva,sans-serif" font-size="11">
    <text x="18.5" y="15" fill="#010101" fill-opacity=".3">tests</text>
    <text x="18.5" y="14">tests</text>
    <text x="57.5" y="15" fill="#010101" fill-opacity=".3">failing</text>
    <text x="57.5" y="14">failing</text>
  </g>
</svg>
EOF

    exit 1
fi

echo -e "${GREEN}All tests passed!${NC}"
echo ""

# Calculate coverage
echo -e "${YELLOW}Calculating coverage...${NC}"
COVERAGE=$(go tool cover -func="$COVERAGE_DIR/coverage.out" | grep total | awk '{print $3}' | sed 's/%//')

echo -e "${GREEN}Total coverage: ${COVERAGE}%${NC}"
echo ""

# Generate coverage HTML report
echo -e "${YELLOW}Generating coverage HTML report...${NC}"
go tool cover -html="$COVERAGE_DIR/coverage.out" -o "$COVERAGE_DIR/coverage.html"
echo -e "${GREEN}Coverage report saved to: $COVERAGE_DIR/coverage.html${NC}"
echo ""

# Determine badge color based on coverage
COVERAGE_INT=$(printf "%.0f" "$COVERAGE")
if [ "$COVERAGE_INT" -ge 90 ]; then
    COLOR="#4c1"
    STATUS="excellent"
elif [ "$COVERAGE_INT" -ge 80 ]; then
    COLOR="#97ca00"
    STATUS="good"
elif [ "$COVERAGE_INT" -ge 70 ]; then
    COLOR="#dfb317"
    STATUS="acceptable"
else
    COLOR="#e05d44"
    STATUS="poor"
fi

# Generate coverage badge
cat > "$BADGES_DIR/coverage.svg" << EOF
<svg xmlns="http://www.w3.org/2000/svg" width="106" height="20">
  <linearGradient id="b" x2="0" y2="100%">
    <stop offset="0" stop-color="#bbb" stop-opacity=".1"/>
    <stop offset="1" stop-opacity=".1"/>
  </linearGradient>
  <mask id="a"><rect width="106" height="20" rx="3" fill="#fff"/></mask>
  <g mask="url(#a)">
    <path fill="#555" d="M0 0h61v20H0z"/>
    <path fill="$COLOR" d="M61 0h45v20H61z"/>
    <path fill="url(#b)" d="M0 0h106v20H0z"/>
  </g>
  <g fill="#fff" text-anchor="middle" font-family="DejaVu Sans,Verdana,Geneva,sans-serif" font-size="11">
    <text x="30.5" y="15" fill="#010101" fill-opacity=".3">coverage</text>
    <text x="30.5" y="14">coverage</text>
    <text x="82.5" y="15" fill="#010101" fill-opacity=".3">${COVERAGE}%</text>
    <text x="82.5" y="14">${COVERAGE}%</text>
  </g>
</svg>
EOF

# Generate tests passing badge
cat > "$BADGES_DIR/tests.svg" << 'EOF'
<svg xmlns="http://www.w3.org/2000/svg" width="88" height="20">
  <linearGradient id="b" x2="0" y2="100%">
    <stop offset="0" stop-color="#bbb" stop-opacity=".1"/>
    <stop offset="1" stop-opacity=".1"/>
  </linearGradient>
  <mask id="a"><rect width="88" height="20" rx="3" fill="#fff"/></mask>
  <g mask="url(#a)">
    <path fill="#555" d="M0 0h37v20H0z"/>
    <path fill="#4c1" d="M37 0h51v20H37z"/>
    <path fill="url(#b)" d="M0 0h88v20H0z"/>
  </g>
  <g fill="#fff" text-anchor="middle" font-family="DejaVu Sans,Verdana,Geneva,sans-serif" font-size="11">
    <text x="18.5" y="15" fill="#010101" fill-opacity=".3">tests</text>
    <text x="18.5" y="14">tests</text>
    <text x="61.5" y="15" fill="#010101" fill-opacity=".3">passing</text>
    <text x="61.5" y="14">passing</text>
  </g>
</svg>
EOF

# Generate build badge
cat > "$BADGES_DIR/build.svg" << 'EOF'
<svg xmlns="http://www.w3.org/2000/svg" width="88" height="20">
  <linearGradient id="b" x2="0" y2="100%">
    <stop offset="0" stop-color="#bbb" stop-opacity=".1"/>
    <stop offset="1" stop-opacity=".1"/>
  </linearGradient>
  <mask id="a"><rect width="88" height="20" rx="3" fill="#fff"/></mask>
  <g mask="url(#a)">
    <path fill="#555" d="M0 0h37v20H0z"/>
    <path fill="#4c1" d="M37 0h51v20H37z"/>
    <path fill="url(#b)" d="M0 0h88v20H0z"/>
  </g>
  <g fill="#fff" text-anchor="middle" font-family="DejaVu Sans,Verdana,Geneva,sans-serif" font-size="11">
    <text x="18.5" y="15" fill="#010101" fill-opacity=".3">build</text>
    <text x="18.5" y="14">build</text>
    <text x="61.5" y="15" fill="#010101" fill-opacity=".3">passing</text>
    <text x="61.5" y="14">passing</text>
  </g>
</svg>
EOF

# Generate Go version badge
GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
cat > "$BADGES_DIR/go-version.svg" << EOF
<svg xmlns="http://www.w3.org/2000/svg" width="78" height="20">
  <linearGradient id="b" x2="0" y2="100%">
    <stop offset="0" stop-color="#bbb" stop-opacity=".1"/>
    <stop offset="1" stop-opacity=".1"/>
  </linearGradient>
  <mask id="a"><rect width="78" height="20" rx="3" fill="#fff"/></mask>
  <g mask="url(#a)">
    <path fill="#555" d="M0 0h25v20H0z"/>
    <path fill="#007d9c" d="M25 0h53v20H25z"/>
    <path fill="url(#b)" d="M0 0h78v20H0z"/>
  </g>
  <g fill="#fff" text-anchor="middle" font-family="DejaVu Sans,Verdana,Geneva,sans-serif" font-size="11">
    <text x="12.5" y="15" fill="#010101" fill-opacity=".3">Go</text>
    <text x="12.5" y="14">Go</text>
    <text x="50.5" y="15" fill="#010101" fill-opacity=".3">$GO_VERSION</text>
    <text x="50.5" y="14">$GO_VERSION</text>
  </g>
</svg>
EOF

echo -e "${GREEN}Badges generated in: $BADGES_DIR${NC}"
echo ""

# Test summary
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Test Summary${NC}"
echo -e "${BLUE}========================================${NC}"
echo -e "Status: ${GREEN}PASSED${NC}"
echo -e "Coverage: ${GREEN}${COVERAGE}%${NC} ($STATUS)"
echo -e "Go Version: ${BLUE}${GO_VERSION}${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Save test results
cat > "$COVERAGE_DIR/test-summary.json" << EOF
{
  "status": "passed",
  "coverage": "$COVERAGE",
  "timestamp": "$(date -u +"%Y-%m-%dT%H:%M:%SZ")",
  "go_version": "$GO_VERSION"
}
EOF

echo -e "${GREEN}Test execution completed successfully!${NC}"
