#!/bin/bash
###############################################################################
# HelixTrack AI QA Automation Runner
#
# Intelligent test automation with AI-powered analysis
###############################################################################

set -e

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m'

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$(dirname "$SCRIPT_DIR")")"

echo -e "${BLUE}=========================================${NC}"
echo -e "${BLUE}  HelixTrack AI QA Automation${NC}"
echo -e "${BLUE}=========================================${NC}"
echo ""

# Check Python
if ! command -v python3 &> /dev/null; then
    echo -e "${YELLOW}Python 3 not found. Please install Python 3.8+${NC}"
    exit 1
fi

# Check if virtual environment exists
if [ ! -d "$SCRIPT_DIR/venv" ]; then
    echo -e "${BLUE}Creating virtual environment...${NC}"
    python3 -m venv "$SCRIPT_DIR/venv"
fi

# Activate virtual environment
source "$SCRIPT_DIR/venv/bin/activate"

# Install dependencies
echo -e "${BLUE}Installing/updating dependencies...${NC}"
pip install -q --upgrade pip
pip install -q -r "$SCRIPT_DIR/requirements.txt"

# Run AI QA tests
echo ""
echo -e "${BLUE}Running AI-powered tests...${NC}"
echo ""

# Set Python path
export PYTHONPATH="$SCRIPT_DIR:$PYTHONPATH"

# Run tests with pytest
pytest "$SCRIPT_DIR/tests/" \
    -v \
    --tb=short \
    --html="$SCRIPT_DIR/reports/latest.html" \
    --self-contained-html \
    "$@"

TEST_RESULT=$?

echo ""
if [ $TEST_RESULT -eq 0 ]; then
    echo -e "${GREEN}✓ All AI QA tests passed!${NC}"
    echo -e "${BLUE}Report: $SCRIPT_DIR/reports/latest.html${NC}"
else
    echo -e "${YELLOW}✗ Some tests failed. Check report for details.${NC}"
    echo -e "${BLUE}Report: $SCRIPT_DIR/reports/latest.html${NC}"
fi

# Deactivate virtual environment
deactivate

exit $TEST_RESULT
