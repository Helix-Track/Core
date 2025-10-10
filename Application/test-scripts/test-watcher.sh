#!/bin/bash
# Watcher Management API Test Script
BASE_URL="${BASE_URL:-http://localhost:8080}"
JWT_TOKEN="${JWT_TOKEN:-}"

echo "Testing Watcher Management API"
echo "=================================================="

# Note: Requires valid ticket ID
TICKET_ID="${TICKET_ID:-test-ticket-id}"

# Add Watcher
curl -s -X POST "$BASE_URL/do" -H "Content-Type: application/json" \
  -d "{\"action\": \"watcherAdd\", \"jwt\": \"$JWT_TOKEN\", \"data\": {\"ticketId\": \"$TICKET_ID\"}}" | jq '.'

# List Watchers
curl -s -X POST "$BASE_URL/do" -H "Content-Type: application/json" \
  -d "{\"action\": \"watcherList\", \"jwt\": \"$JWT_TOKEN\", \"data\": {\"ticketId\": \"$TICKET_ID\"}}" | jq '.'

echo "Watcher tests completed"
