#!/bin/bash
# Test /do endpoint - Health action

BASE_URL="${BASE_URL:-http://localhost:8080}"

echo "Testing Health endpoint (via /do)..."
curl -X POST "${BASE_URL}/do" \
  -H "Content-Type: application/json" \
  -d '{
    "action": "health"
  }' | jq .

echo ""
echo "Testing Health endpoint (dedicated)..."
curl -X GET "${BASE_URL}/health" | jq .

echo ""
