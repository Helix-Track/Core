#!/bin/bash
# Permission Management API Test Script
BASE_URL="${BASE_URL:-http://localhost:8080}"
JWT_TOKEN="${JWT_TOKEN:-}"

echo "Testing Permission Management API"
echo "=================================================="

# List Permissions
curl -s -X POST "$BASE_URL/do" -H "Content-Type: application/json" \
  -d "{\"action\": \"permissionList\", \"jwt\": \"$JWT_TOKEN\", \"data\": {}}" | jq '.'

# Check Permission
curl -s -X POST "$BASE_URL/do" -H "Content-Type: application/json" \
  -d "{\"action\": \"permissionCheck\", \"jwt\": \"$JWT_TOKEN\", \"data\": {\"resource\": \"ticket\", \"action\": \"create\"}}" | jq '.'

echo "Permission tests completed"
