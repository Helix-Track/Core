#!/bin/bash
# Test /do endpoint - Authenticate action

BASE_URL="${BASE_URL:-http://localhost:8080}"
USERNAME="${USERNAME:-testuser}"
PASSWORD="${PASSWORD:-testpass}"

echo "Testing Authenticate endpoint..."
curl -X POST "${BASE_URL}/do" \
  -H "Content-Type: application/json" \
  -d "{
    \"action\": \"authenticate\",
    \"data\": {
      \"username\": \"${USERNAME}\",
      \"password\": \"${PASSWORD}\"
    }
  }" | jq .

echo ""
