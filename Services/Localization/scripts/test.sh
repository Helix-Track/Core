#!/bin/bash

# HelixTrack Localization Service - Test Script

set -e

echo "🧪 Running tests for HelixTrack Localization Service..."

# Navigate to project root
cd "$(dirname "$0")/.."

# Run tests with coverage
echo "Running unit tests with coverage..."
go test -v -cover -coverprofile=coverage.out ./...

# Generate coverage report
echo ""
echo "Generating coverage report..."
go tool cover -func=coverage.out

# Show coverage percentage
echo ""
COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}')
echo "📊 Total Coverage: $COVERAGE"

# Generate HTML coverage report (optional)
if [ "$1" == "--html" ]; then
    echo "Generating HTML coverage report..."
    go tool cover -html=coverage.out -o coverage.html
    echo "✅ HTML report generated: coverage.html"
fi

# Run race detector tests
echo ""
echo "Running race detector tests..."
go test -race ./...

echo ""
echo "✅ All tests passed!"
