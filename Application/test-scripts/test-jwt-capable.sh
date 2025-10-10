#!/bin/bash
# Test /do endpoint - JWT Capable action

BASE_URL="${BASE_URL:-http://localhost:8080}"

echo "Testing JWT Capable endpoint..."
curl -X POST "${BASE_URL}/do" \
  -H "Content-Type: application/json" \
  -d '{
    "action": "jwtCapable"
  }' | jq .

echo ""
