#!/bin/bash
# Label Management API Test Script
BASE_URL="${BASE_URL:-http://localhost:8080}"
JWT_TOKEN="${JWT_TOKEN:-}"

echo "Testing Label Management API"
echo "=================================================="

# Create Label
curl -s -X POST "$BASE_URL/do" -H "Content-Type: application/json" \
  -d "{\"action\": \"labelCreate\", \"jwt\": \"$JWT_TOKEN\", \"data\": {\"name\": \"bug\", \"color\": \"#FF0000\", \"description\": \"Bug label\"}}" | jq '.'

# List Labels
curl -s -X POST "$BASE_URL/do" -H "Content-Type: application/json" \
  -d "{\"action\": \"labelList\", \"jwt\": \"$JWT_TOKEN\", \"data\": {}}" | jq '.'

echo "Label tests completed"
