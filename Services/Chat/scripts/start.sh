#!/bin/bash

# Start script for HelixTrack Chat Service with Docker Compose

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
echo -e "${BLUE}  HelixTrack Chat Service Startup${NC}"
echo -e "${BLUE}======================================${NC}"
echo ""

# Check if network exists, create if not
if ! docker network inspect helixtrack-network >/dev/null 2>&1; then
    echo -e "${YELLOW}Creating helixtrack-network...${NC}"
    docker network create helixtrack-network
fi

# Check for .env file
if [ ! -f .env ]; then
    echo -e "${YELLOW}Creating .env file with default values...${NC}"
    cat > .env <<EOF
CHAT_DB_PASSWORD=chat_secure_password_change_in_production
JWT_SECRET=your-jwt-secret-key-change-in-production
EOF
    echo -e "${RED}WARNING: Please update .env with secure passwords!${NC}"
    echo ""
fi

# Start services
echo -e "${GREEN}Starting Chat services...${NC}"
docker-compose up -d

# Wait for services to be healthy
echo ""
echo -e "${YELLOW}Waiting for services to be healthy...${NC}"
sleep 5

# Check service status
echo ""
echo -e "${GREEN}Service Status:${NC}"
docker-compose ps

# Show logs
echo ""
echo -e "${BLUE}Recent logs:${NC}"
docker-compose logs --tail=20

echo ""
echo -e "${GREEN}======================================${NC}"
echo -e "${GREEN}  Chat Service Started Successfully!${NC}"
echo -e "${GREEN}======================================${NC}"
echo ""
echo -e "Chat Service: ${GREEN}http://localhost:9090${NC}"
echo -e "Health Check: ${GREEN}http://localhost:9090/health${NC}"
echo -e "Database: ${GREEN}localhost:5433${NC}"
echo ""
echo -e "View logs: ${YELLOW}docker-compose logs -f chat-service${NC}"
echo -e "Stop services: ${YELLOW}./scripts/stop.sh${NC}"
echo ""
