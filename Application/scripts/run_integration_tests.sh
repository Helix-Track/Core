#!/bin/bash
# HelixTrack Core - Integration Test Runner
# Runs all integration tests

set -e

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m'

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

echo -e "${BLUE}=============================================${NC}"
echo -e "${BLUE}  Integration Tests                         ${NC}"
echo -e "${BLUE}=============================================${NC}"
echo ""

cd "$PROJECT_ROOT"

mkdir -p coverage

echo -e "${YELLOW}Running API Integration Tests...${NC}"
go test ./tests/integration/api_integration_test.go -v -cover -coverprofile=coverage/api_integration.out

echo ""
echo -e "${YELLOW}Running Security Integration Tests...${NC}"
go test ./tests/integration/security_integration_test.go -v -cover -coverprofile=coverage/security_integration.out

echo ""
echo -e "${YELLOW}Running Database-Cache Integration Tests...${NC}"
go test ./tests/integration/database_cache_integration_test.go -v -cover -coverprofile=coverage/db_cache_integration.out

echo ""
echo -e "${GREEN}✓ All integration tests passed${NC}"
