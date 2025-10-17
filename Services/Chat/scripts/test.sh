#!/bin/bash

# Test script for HelixTrack Chat Service

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Get script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

cd "$PROJECT_DIR"

echo -e "${BLUE}======================================${NC}"
echo -e "${BLUE}  HelixTrack Chat Service Tests${NC}"
echo -e "${BLUE}======================================${NC}"
echo ""

# Run tests with coverage
echo -e "${YELLOW}Running unit tests with coverage...${NC}"
go test ./... -v -cover -coverprofile=coverage.out

echo ""
echo -e "${YELLOW}Generating coverage report...${NC}"
go tool cover -func=coverage.out

# Optional: Generate HTML coverage report
if command -v go &> /dev/null; then
    echo ""
    echo -e "${YELLOW}Generating HTML coverage report...${NC}"
    go tool cover -html=coverage.out -o coverage.html
    echo -e "${GREEN}HTML coverage report: coverage.html${NC}"
fi

# Run race detection
echo ""
echo -e "${YELLOW}Running race detection...${NC}"
go test ./... -race -short

# Run linting if golangci-lint is available
if command -v golangci-lint &> /dev/null; then
    echo ""
    echo -e "${YELLOW}Running linter...${NC}"
    golangci-lint run
else
    echo ""
    echo -e "${YELLOW}golangci-lint not found, skipping lint check${NC}"
fi

echo ""
echo -e "${GREEN}======================================${NC}"
echo -e "${GREEN}  All Tests Passed Successfully!${NC}"
echo -e "${GREEN}======================================${NC}"
echo ""
