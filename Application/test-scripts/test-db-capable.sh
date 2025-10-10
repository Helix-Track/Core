#!/bin/bash
# Test /do endpoint - Database Capable action

BASE_URL="${BASE_URL:-http://localhost:8080}"

echo "Testing Database Capable endpoint..."
curl -X POST "${BASE_URL}/do" \
  -H "Content-Type: application/json" \
  -d '{
    "action": "dbCapable"
  }' | jq .

echo ""
