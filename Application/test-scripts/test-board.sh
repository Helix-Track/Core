#!/bin/bash
# Board Management API Test Script
BASE_URL="${BASE_URL:-http://localhost:8080}"
JWT_TOKEN="${JWT_TOKEN:-}"

echo "Testing Board Management API"
echo "=================================================="

# Create Board
curl -s -X POST "$BASE_URL/do" -H "Content-Type: application/json" \
  -d "{\"action\": \"boardCreate\", \"jwt\": \"$JWT_TOKEN\", \"data\": {\"title\": \"Sprint Board\", \"description\": \"Main development board\"}}" | jq '.'

# List Boards
curl -s -X POST "$BASE_URL/do" -H "Content-Type: application/json" \
  -d "{\"action\": \"boardList\", \"jwt\": \"$JWT_TOKEN\", \"data\": {}}" | jq '.'

echo "Board tests completed"
