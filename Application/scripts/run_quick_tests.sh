#!/bin/bash
# HelixTrack Core - Quick Test Runner (Unit Tests Only)
# Runs only unit tests for fast feedback during development

set -e

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

echo -e "${BLUE}Quick Test Run (Unit Tests Only)${NC}"
echo ""

cd "$PROJECT_ROOT"

# Run unit tests with race detection
go test ./internal/... -v -race -timeout 30s

echo ""
echo -e "${GREEN}✓ Quick tests completed${NC}"
