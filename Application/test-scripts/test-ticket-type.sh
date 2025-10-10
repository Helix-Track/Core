#!/bin/bash
# Ticket Type Management API Test Script
BASE_URL="${BASE_URL:-http://localhost:8080}"
JWT_TOKEN="${JWT_TOKEN:-}"

echo "Testing Ticket Type Management API"
echo "=================================================="

# Create Type
curl -s -X POST "$BASE_URL/do" -H "Content-Type: application/json" \
  -d "{\"action\": \"ticketTypeCreate\", \"jwt\": \"$JWT_TOKEN\", \"data\": {\"name\": \"Bug\", \"icon\": \"bug\", \"color\": \"#FF0000\", \"description\": \"Software bug\"}}" | jq '.'

# List Types
curl -s -X POST "$BASE_URL/do" -H "Content-Type: application/json" \
  -d "{\"action\": \"ticketTypeList\", \"jwt\": \"$JWT_TOKEN\", \"data\": {}}" | jq '.'

echo "Ticket Type tests completed"
