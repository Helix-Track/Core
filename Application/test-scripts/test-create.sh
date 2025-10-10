#!/bin/bash
# Test /do endpoint - Create action

BASE_URL="${BASE_URL:-http://localhost:8080}"
JWT_TOKEN="${JWT_TOKEN:-your-jwt-token-here}"
OBJECT="${OBJECT:-project}"

echo "Testing Create endpoint..."
curl -X POST "${BASE_URL}/do" \
  -H "Content-Type: application/json" \
  -d "{
    \"action\": \"create\",
    \"jwt\": \"${JWT_TOKEN}\",
    \"object\": \"${OBJECT}\",
    \"data\": {
      \"name\": \"Test Project\",
      \"description\": \"A test project\"
    }
  }" | jq .

echo ""
