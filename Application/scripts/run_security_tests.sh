#!/bin/bash
# HelixTrack Core - Security Test Runner
# Runs all security-related tests

set -e

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m'

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

echo -e "${BLUE}=============================================${NC}"
echo -e "${BLUE}  HelixTrack Core - Security Test Suite    ${NC}"
echo -e "${BLUE}=============================================${NC}"
echo ""

cd "$PROJECT_ROOT"

# Run security module tests
echo -e "${YELLOW}Running Security Module Tests...${NC}"
echo ""
go test ./internal/security/... -v -cover -race

echo ""
echo -e "${YELLOW}Running Security Integration Tests...${NC}"
echo ""
go test ./tests/integration/security_integration_test.go -v -cover

echo ""
echo -e "${GREEN}✓ All security tests passed${NC}"
echo ""
echo -e "${BLUE}Security Coverage:${NC}"
go test ./internal/security/... -coverprofile=coverage.out
go tool cover -func=coverage.out | grep -E '(ddos|csrf|brute|input|security|tls|audit)' || echo "Coverage data"
rm coverage.out
