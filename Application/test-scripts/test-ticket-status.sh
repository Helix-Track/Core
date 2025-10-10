#!/bin/bash
# Ticket Status Management API Test Script
BASE_URL="${BASE_URL:-http://localhost:8080}"
JWT_TOKEN="${JWT_TOKEN:-}"

echo "Testing Ticket Status Management API"
echo "=================================================="

# Create Status
curl -s -X POST "$BASE_URL/do" -H "Content-Type: application/json" \
  -d "{\"action\": \"ticketStatusCreate\", \"jwt\": \"$JWT_TOKEN\", \"data\": {\"name\": \"In Review\", \"color\": \"#9900FF\", \"description\": \"Under code review\"}}" | jq '.'

# List Statuses
curl -s -X POST "$BASE_URL/do" -H "Content-Type: application/json" \
  -d "{\"action\": \"ticketStatusList\", \"jwt\": \"$JWT_TOKEN\", \"data\": {}}" | jq '.'

echo "Ticket Status tests completed"
