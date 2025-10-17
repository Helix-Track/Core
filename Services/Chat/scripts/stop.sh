#!/bin/bash

# Stop script for HelixTrack Chat Service

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Get script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

cd "$PROJECT_DIR"

echo -e "${YELLOW}Stopping HelixTrack Chat Service...${NC}"

# Stop and remove containers
docker-compose down

echo ""
echo -e "${GREEN}Chat Service stopped successfully!${NC}"
echo ""
echo -e "To remove volumes as well: ${YELLOW}docker-compose down -v${NC}"
echo -e "To start again: ${YELLOW}./scripts/start.sh${NC}"
echo ""
