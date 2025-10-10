#!/bin/bash
# Filter Management API Test Script
BASE_URL="${BASE_URL:-http://localhost:8080}"
JWT_TOKEN="${JWT_TOKEN:-}"

echo "Testing Filter Management API"
echo "=================================================="

# Save Filter
curl -s -X POST "$BASE_URL/do" -H "Content-Type: application/json" \
  -d "{\"action\": \"filterSave\", \"jwt\": \"$JWT_TOKEN\", \"data\": {\"name\": \"My Open Tickets\", \"criteria\": {\"status\": \"open\", \"assignee\": \"me\"}}}" | jq '.'

# List Filters
curl -s -X POST "$BASE_URL/do" -H "Content-Type: application/json" \
  -d "{\"action\": \"filterList\", \"jwt\": \"$JWT_TOKEN\", \"data\": {}}" | jq '.'

echo "Filter tests completed"
