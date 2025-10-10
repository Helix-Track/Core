#!/bin/bash
# Ticket Relationship API Test Script
BASE_URL="${BASE_URL:-http://localhost:8080}"
JWT_TOKEN="${JWT_TOKEN:-}"

echo "Testing Ticket Relationship API"
echo "=================================================="

# Create Relationship Type
curl -s -X POST "$BASE_URL/do" -H "Content-Type: application/json" \
  -d "{\"action\": \"ticketRelationshipTypeCreate\", \"jwt\": \"$JWT_TOKEN\", \"data\": {\"name\": \"Blocks\", \"reverseType\": \"Is Blocked By\"}}" | jq '.'

# List Relationship Types
curl -s -X POST "$BASE_URL/do" -H "Content-Type: application/json" \
  -d "{\"action\": \"ticketRelationshipTypeList\", \"jwt\": \"$JWT_TOKEN\", \"data\": {}}" | jq '.'

echo "Ticket Relationship tests completed"
