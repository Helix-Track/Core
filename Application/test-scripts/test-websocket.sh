#!/bin/bash

# WebSocket Test Script for HelixTrack Core
# Tests WebSocket connection, authentication, and event subscription

set -e

# Colors for output
RED='\033[0:31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0;m' # No Color

# Configuration
BASE_URL="${BASE_URL:-http://localhost:8080}"
WS_URL="${WS_URL:-ws://localhost:8080/ws}"
USERNAME="${USERNAME:-testuser}"
PASSWORD="${PASSWORD:-testpass}"

echo -e "${BLUE}=====================================${NC}"
echo -e "${BLUE}WebSocket Test Script${NC}"
echo -e "${BLUE}=====================================${NC}"
echo ""

# Step 1: Register user (if not exists)
echo -e "${YELLOW}Step 1: Registering user...${NC}"
REGISTER_RESPONSE=$(curl -s -X POST "${BASE_URL}/api/auth/register" \
  -H "Content-Type: application/json" \
  -d "{\"username\":\"${USERNAME}\",\"password\":\"${PASSWORD}\",\"email\":\"${USERNAME}@example.com\",\"name\":\"Test User\"}" || echo '{"errorCode":3001}')

if echo "$REGISTER_RESPONSE" | grep -q '"errorCode":-1'; then
  echo -e "${GREEN}✓ User registered successfully${NC}"
elif echo "$REGISTER_RESPONSE" | grep -q '"errorCode":3001'; then
  echo -e "${YELLOW}⚠ User already exists (OK)${NC}"
else
  echo -e "${YELLOW}⚠ Registration response: $REGISTER_RESPONSE${NC}"
fi

# Step 2: Authenticate and get JWT token
echo -e "${YELLOW}Step 2: Authenticating...${NC}"
AUTH_RESPONSE=$(curl -s -X POST "${BASE_URL}/api/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"username\":\"${USERNAME}\",\"password\":\"${PASSWORD}\"}")

if echo "$AUTH_RESPONSE" | grep -q '"errorCode":-1'; then
  JWT_TOKEN=$(echo "$AUTH_RESPONSE" | grep -o '"token":"[^"]*' | cut -d'"' -f4)
  echo -e "${GREEN}✓ Authentication successful${NC}"
  echo -e "  Token: ${JWT_TOKEN:0:50}..."
else
  echo -e "${RED}✗ Authentication failed${NC}"
  echo "  Response: $AUTH_RESPONSE"
  exit 1
fi

# Step 3: Test WebSocket Stats endpoint
echo -e "${YELLOW}Step 3: Testing WebSocket stats endpoint...${NC}"
STATS_RESPONSE=$(curl -s "${BASE_URL}/ws/stats")

if echo "$STATS_RESPONSE" | grep -q '"errorCode":-1'; then
  echo -e "${GREEN}✓ WebSocket stats endpoint accessible${NC}"
  echo "  Stats: $STATS_RESPONSE"
else
  echo -e "${RED}✗ WebSocket stats endpoint failed${NC}"
  echo "  Response: $STATS_RESPONSE"
fi

# Step 4: Check for WebSocket clients (websocat or wscat)
echo -e "${YELLOW}Step 4: Checking for WebSocket clients...${NC}"

if command -v websocat &> /dev/null; then
  WS_CLIENT="websocat"
  echo -e "${GREEN}✓ Found websocat${NC}"
elif command -v wscat &> /dev/null; then
  WS_CLIENT="wscat"
  echo -e "${GREEN}✓ Found wscat${NC}"
else
  echo -e "${YELLOW}⚠ No WebSocket client found (websocat or wscat)${NC}"
  echo -e "  Install with: ${BLUE}cargo install websocat${NC} or ${BLUE}npm install -g wscat${NC}"
  echo -e "  ${YELLOW}Skipping WebSocket connection test${NC}"
  WS_CLIENT=""
fi

# Step 5: Test WebSocket connection (if client available)
if [ -n "$WS_CLIENT" ]; then
  echo -e "${YELLOW}Step 5: Testing WebSocket connection...${NC}"
  echo -e "  Connecting to: ${WS_URL}?token=${JWT_TOKEN:0:50}..."

  # Create subscription message
  SUBSCRIBE_MSG='{"type":"subscribe","data":{"eventTypes":["ticket.created","ticket.updated","ticket.deleted","project.created"],"entityTypes":["ticket","project"],"includeReads":false}}'

  echo -e "  ${BLUE}Subscription message:${NC} $SUBSCRIBE_MSG"
  echo ""
  echo -e "${GREEN}WebSocket Test Instructions:${NC}"
  echo "  1. The WebSocket connection will open below"
  echo "  2. A subscription message will be sent automatically"
  echo "  3. You should receive an acknowledgment"
  echo "  4. Any matching events will appear in real-time"
  echo "  5. Press Ctrl+C to disconnect"
  echo ""
  echo -e "${YELLOW}Connecting...${NC}"
  echo ""

  if [ "$WS_CLIENT" = "websocat" ]; then
    # Use websocat
    echo "$SUBSCRIBE_MSG" | websocat "${WS_URL}?token=${JWT_TOKEN}" || echo -e "${RED}✗ WebSocket connection failed${NC}"
  else
    # Use wscat
    echo "$SUBSCRIBE_MSG" | wscat -c "${WS_URL}?token=${JWT_TOKEN}" || echo -e "${RED}✗ WebSocket connection failed${NC}"
  fi
else
  echo -e "${YELLOW}Step 5: Skipped (no WebSocket client available)${NC}"
fi

echo ""
echo -e "${BLUE}=====================================${NC}"
echo -e "${BLUE}WebSocket Test Complete${NC}"
echo -e "${BLUE}=====================================${NC}"
echo ""
echo -e "${GREEN}Next steps:${NC}"
echo "  1. Open the HTML client: ${BLUE}Application/test-scripts/websocket-client.html${NC}"
echo "  2. Create a ticket to trigger events"
echo "  3. Watch for real-time event notifications"
echo ""
