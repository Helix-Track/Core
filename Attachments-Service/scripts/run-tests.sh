#!/bin/bash

# HelixTrack Attachments Service - Test Runner
# Runs all unit tests with coverage reporting

set -e

echo "==================================="
echo "HelixTrack Attachments Service"
echo "Test Runner"
echo "==================================="
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Navigate to project root
cd "$(dirname "$0")/.."

echo "📦 Installing dependencies..."
go mod download
echo ""

echo "🧪 Running unit tests..."
echo ""

# Run tests with coverage
go test ./... -v -cover -coverprofile=coverage.out -covermode=atomic

TEST_EXIT_CODE=$?

echo ""
echo "==================================="

if [ $TEST_EXIT_CODE -eq 0 ]; then
    echo -e "${GREEN}✅ All tests PASSED${NC}"
else
    echo -e "${RED}❌ Some tests FAILED${NC}"
    exit $TEST_EXIT_CODE
fi

echo "==================================="
echo ""

# Generate coverage report
echo "📊 Generating coverage report..."
go tool cover -func=coverage.out

echo ""
echo "📈 Overall coverage:"
go tool cover -func=coverage.out | grep total | awk '{print $3}'

echo ""
echo "💡 To view HTML coverage report:"
echo "   go tool cover -html=coverage.out"
echo ""

# Run race detection
echo "🏁 Running race detection..."
go test ./... -race -short

RACE_EXIT_CODE=$?

if [ $RACE_EXIT_CODE -eq 0 ]; then
    echo -e "${GREEN}✅ No race conditions detected${NC}"
else
    echo -e "${YELLOW}⚠️  Potential race conditions detected${NC}"
fi

echo ""
echo "==================================="
echo "✅ Test run complete!"
echo "==================================="
