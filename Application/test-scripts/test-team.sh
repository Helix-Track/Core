#!/bin/bash
# Team Management API Test Script
BASE_URL="${BASE_URL:-http://localhost:8080}"
JWT_TOKEN="${JWT_TOKEN:-}"

echo "Testing Team Management API"
echo "=================================================="

# Create Team
curl -s -X POST "$BASE_URL/do" -H "Content-Type: application/json" \
  -d "{\"action\": \"teamCreate\", \"jwt\": \"$JWT_TOKEN\", \"data\": {\"name\": \"Backend Team\", \"description\": \"Backend development team\"}}" | jq '.'

# List Teams
curl -s -X POST "$BASE_URL/do" -H "Content-Type: application/json" \
  -d "{\"action\": \"teamList\", \"jwt\": \"$JWT_TOKEN\", \"data\": {}}" | jq '.'

echo "Team tests completed"
