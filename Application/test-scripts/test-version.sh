#!/bin/bash
# Test /do endpoint - Version action

BASE_URL="${BASE_URL:-http://localhost:8080}"

echo "Testing Version endpoint..."
curl -X POST "${BASE_URL}/do" \
  -H "Content-Type: application/json" \
  -d '{
    "action": "version"
  }' | jq .

echo ""
