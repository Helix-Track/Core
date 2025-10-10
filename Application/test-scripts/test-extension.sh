#!/bin/bash
# Extension Management API Test Script
BASE_URL="${BASE_URL:-http://localhost:8080}"
JWT_TOKEN="${JWT_TOKEN:-}"

echo "Testing Extension Management API"
echo "=================================================="

# List Extensions
curl -s -X POST "$BASE_URL/do" -H "Content-Type: application/json" \
  -d "{\"action\": \"extensionList\", \"jwt\": \"$JWT_TOKEN\", \"data\": {}}" | jq '.'

echo "Extension tests completed"
