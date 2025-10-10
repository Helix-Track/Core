#!/bin/bash
# Report Management API Test Script
BASE_URL="${BASE_URL:-http://localhost:8080}"
JWT_TOKEN="${JWT_TOKEN:-}"

echo "Testing Report Management API"
echo "=================================================="

# Create Report
curl -s -X POST "$BASE_URL/do" -H "Content-Type: application/json" \
  -d "{\"action\": \"reportCreate\", \"jwt\": \"$JWT_TOKEN\", \"data\": {\"name\": \"Ticket Report\", \"query\": \"SELECT * FROM tickets\"}}" | jq '.'

# List Reports
curl -s -X POST "$BASE_URL/do" -H "Content-Type: application/json" \
  -d "{\"action\": \"reportList\", \"jwt\": \"$JWT_TOKEN\", \"data\": {}}" | jq '.'

echo "Report tests completed"
