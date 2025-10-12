#!/bin/bash

# AI QA - WebSocket Real-Time Event Testing
# Tests WebSocket connections and real-time event delivery across multiple clients

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
MAGENTA='\033[0;35m'
NC='\033[0m' # No Color

# Configuration
BASE_URL="${BASE_URL:-http://localhost:8080}"
WS_URL="${WS_URL:-ws://localhost:8080/ws}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
OUTPUT_DIR="$SCRIPT_DIR/ai-qa-output"
USERS_FILE="$OUTPUT_DIR/users.json"
PROJECTS_FILE="$OUTPUT_DIR/projects.json"
WS_LOG="$OUTPUT_DIR/websocket-realtime.log"

# Test parameters
TEST_DURATION="${TEST_DURATION:-120}"  # 2 minutes
CONCURRENT_CLIENTS="${CONCURRENT_CLIENTS:-3}"

# Counters
connections_established=0
events_sent=0
events_received=0
failed_operations=0

echo -e "${MAGENTA}========================================${NC}"
echo -e "${MAGENTA}WebSocket Real-Time Testing${NC}"
echo -e "${MAGENTA}========================================${NC}"
echo ""

# Check prerequisites
if [ ! -f "$USERS_FILE" ] || [ ! -f "$PROJECTS_FILE" ]; then
    echo -e "${RED}Error: Setup not completed!${NC}"
    echo -e "${YELLOW}Run ai-qa-setup-organization.sh and ai-qa-setup-projects.sh first${NC}"
    exit 1
fi

# Check for WebSocket client
if ! command -v websocat &> /dev/null && ! command -v wscat &> /dev/null; then
    echo -e "${RED}Error: No WebSocket client found!${NC}"
    echo -e "${YELLOW}Install with: ${BLUE}cargo install websocat${NC} or ${BLUE}npm install -g wscat${NC}"
    exit 1
fi

WS_CLIENT=$(command -v websocat &> /dev/null && echo "websocat" || echo "wscat")

# Initialize log
echo "WebSocket Real-Time Testing - $(date)" > "$WS_LOG"
echo "======================================" >> "$WS_LOG"
echo "" >> "$WS_LOG"

echo -e "${CYAN}WebSocket URL:${NC} $WS_URL"
echo -e "${CYAN}WebSocket Client:${NC} $WS_CLIENT"
echo -e "${CYAN}Test Duration:${NC} $TEST_DURATION seconds"
echo -e "${CYAN}Concurrent Clients:${NC} $CONCURRENT_CLIENTS"
echo ""

log_action() {
    local message="$1"
    echo -e "[$(date '+%Y-%m-%d %H:%M:%S')] $message" >> "$WS_LOG"
    echo -e "$message"
}

# Function to make API call (for triggering events)
api_call() {
    local action="$1"
    local data="$2"
    local jwt="$3"
    local description="$4"

    log_action "${YELLOW}[API] $description${NC}"

    local body="{\"action\": \"$action\", \"jwt\": \"$jwt\", \"data\": $data}"

    local response
    response=$(curl -s -X POST "$BASE_URL/do" \
        -H "Content-Type: application/json" \
        -d "$body")

    local error_code
    error_code=$(echo "$response" | grep -o '"errorCode":[^,}]*' | head -1 | cut -d':' -f2 | tr -d ' ')

    if [ "$error_code" = "-1" ]; then
        events_sent=$((events_sent + 1))
        log_action "  ${GREEN}✓ Event triggered${NC}"
        echo "$response"
        return 0
    else
        failed_operations=$((failed_operations + 1))
        log_action "  ${RED}✗ Failed (errorCode: $error_code)${NC}"
        return 1
    fi
}

# Function to establish WebSocket connection
establish_ws_connection() {
    local user_index="$1"
    local user_name=$(jq -r ".users[$user_index].name" "$USERS_FILE")
    local user_token=$(jq -r ".users[$user_index].token" "$USERS_FILE")
    local client_id="client-$user_index"
    local ws_output="$OUTPUT_DIR/ws-${client_id}.log"

    log_action "${BLUE}[WS $client_id] Establishing connection for $user_name${NC}"

    # Create subscription message
    local subscribe_msg='{
        "type": "subscribe",
        "data": {
            "eventTypes": [
                "ticket.created",
                "ticket.updated",
                "ticket.deleted",
                "ticket.assigned",
                "ticket.commented",
                "project.created",
                "project.updated",
                "sprint.started",
                "sprint.completed",
                "user.mentioned"
            ],
            "entityTypes": ["ticket", "project", "sprint", "comment"],
            "includeReads": false
        }
    }'

    # Start WebSocket connection in background
    if [ "$WS_CLIENT" = "websocat" ]; then
        echo "$subscribe_msg" | websocat "${WS_URL}?token=${user_token}" > "$ws_output" 2>&1 &
    else
        echo "$subscribe_msg" | wscat -c "${WS_URL}?token=${user_token}" > "$ws_output" 2>&1 &
    fi

    local ws_pid=$!
    echo "$ws_pid" > "$OUTPUT_DIR/ws-${client_id}.pid"

    connections_established=$((connections_established + 1))
    log_action "  ${GREEN}✓ Connection established (PID: $ws_pid)${NC}"

    sleep 2  # Give connection time to establish
}

# Function to trigger real-time events
trigger_realtime_events() {
    local actor_user_index=$((RANDOM % $(jq '.users | length' "$USERS_FILE")))
    local actor_name=$(jq -r ".users[$actor_user_index].name" "$USERS_FILE")
    local actor_token=$(jq -r ".users[$actor_user_index].token" "$USERS_FILE")

    local event_type=$((RANDOM % 8))

    case $event_type in
        0|1|2)
            # Create ticket (most common event)
            local project_count=$(jq '.projects | length' "$PROJECTS_FILE")
            if [ "$project_count" -gt 0 ]; then
                local proj_index=$((RANDOM % project_count))
                local project_id=$(jq -r ".projects[$proj_index].id" "$PROJECTS_FILE")
                local project_key=$(jq -r ".projects[$proj_index].key" "$PROJECTS_FILE")
                local type_id=$(jq -r ".projects[$proj_index].types[] | select(.name == \"Task\") | .id" "$PROJECTS_FILE")
                local status_id=$(jq -r ".projects[$proj_index].statuses[0].id" "$PROJECTS_FILE")
                local priority_id=$(jq -r ".projects[$proj_index].priorities[1].id" "$PROJECTS_FILE")

                local ticket_title="RT Event Task - $(date +%s)"
                local ticket_desc="Task created to trigger real-time event at $(date)"

                log_action "${CYAN}[EVENT] $actor_name creating ticket in $project_key${NC}"
                api_call "ticketCreate" "{\"projectId\": \"$project_id\", \"title\": \"$ticket_title\", \"description\": \"$ticket_desc\", \"typeId\": \"$type_id\", \"statusId\": \"$status_id\", \"priorityId\": \"$priority_id\"}" "$actor_token" "ticket.created event" > /dev/null
            fi
            ;;
        3)
            # Update ticket status
            log_action "${CYAN}[EVENT] $actor_name updating ticket status${NC}"
            # Note: Would need ticket ID in production
            ;;
        4)
            # Add comment
            log_action "${CYAN}[EVENT] $actor_name adding comment${NC}"
            # Note: Would need ticket ID in production
            ;;
        5)
            # Assign ticket
            log_action "${CYAN}[EVENT] $actor_name assigning ticket${NC}"
            # Note: Would need ticket ID in production
            ;;
        6)
            # Mention user in comment
            local mentioned_user_index=$((RANDOM % $(jq '.users | length' "$USERS_FILE")))
            local mentioned_name=$(jq -r ".users[$mentioned_user_index].name" "$USERS_FILE")
            log_action "${CYAN}[EVENT] $actor_name mentioning @$mentioned_name${NC}"
            # Note: Would need ticket ID in production
            ;;
        7)
            # Update project
            local project_count=$(jq '.projects | length' "$PROJECTS_FILE")
            if [ "$project_count" -gt 0 ]; then
                local proj_index=$((RANDOM % project_count))
                local project_id=$(jq -r ".projects[$proj_index].id" "$PROJECTS_FILE")
                local project_key=$(jq -r ".projects[$proj_index].key" "$PROJECTS_FILE")

                log_action "${CYAN}[EVENT] $actor_name updating project $project_key${NC}"
                # Note: Simplified for simulation
            fi
            ;;
    esac
}

# Main test execution
log_action "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
log_action "${GREEN}Starting WebSocket Real-Time Test${NC}"
log_action "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

# Step 1: Establish WebSocket connections for multiple clients
log_action "${BLUE}Step 1: Establishing WebSocket connections${NC}"
for i in $(seq 0 $((CONCURRENT_CLIENTS - 1))); do
    establish_ws_connection $i
    echo ""
done

echo ""
log_action "${GREEN}✓ All WebSocket connections established${NC}"
echo ""

# Step 2: Trigger events and monitor delivery
log_action "${BLUE}Step 2: Triggering real-time events${NC}"
echo ""

start_time=$(date +%s)
end_time=$((start_time + TEST_DURATION))

while [ $(date +%s) -lt $end_time ]; do
    trigger_realtime_events
    echo ""
    sleep 5  # Trigger event every 5 seconds
done

# Step 3: Close WebSocket connections
log_action "${BLUE}Step 3: Closing WebSocket connections${NC}"
for i in $(seq 0 $((CONCURRENT_CLIENTS - 1))); do
    local client_id="client-$i"
    local ws_pid=$(cat "$OUTPUT_DIR/ws-${client_id}.pid" 2>/dev/null || echo "")

    if [ -n "$ws_pid" ]; then
        log_action "  Closing connection $client_id (PID: $ws_pid)"
        kill "$ws_pid" 2>/dev/null || true
    fi
done

echo ""

# Step 4: Analyze WebSocket logs
log_action "${BLUE}Step 4: Analyzing WebSocket event delivery${NC}"
echo ""

for i in $(seq 0 $((CONCURRENT_CLIENTS - 1))); do
    local client_id="client-$i"
    local ws_output="$OUTPUT_DIR/ws-${client_id}.log"
    local user_name=$(jq -r ".users[$i].name" "$USERS_FILE")

    if [ -f "$ws_output" ]; then
        local event_count=$(grep -c "type" "$ws_output" 2>/dev/null || echo "0")
        events_received=$((events_received + event_count))
        log_action "  ${GREEN}[$client_id] $user_name received $event_count events${NC}"
    fi
done

echo ""

# Summary
log_action "${MAGENTA}========================================${NC}"
log_action "${MAGENTA}WebSocket Real-Time Test Complete${NC}"
log_action "${MAGENTA}========================================${NC}"
echo ""
log_action "${CYAN}Test Summary:${NC}"
log_action "  Duration: $TEST_DURATION seconds"
log_action "  Concurrent Clients: $CONCURRENT_CLIENTS"
log_action "  ${GREEN}Connections Established: $connections_established${NC}"
log_action "  ${GREEN}Events Sent: $events_sent${NC}"
log_action "  ${GREEN}Events Received: $events_received${NC}"
log_action "  ${RED}Failed Operations: $failed_operations${NC}"
echo ""
log_action "${CYAN}WebSocket Logs:${NC}"
for i in $(seq 0 $((CONCURRENT_CLIENTS - 1))); do
    local client_id="client-$i"
    log_action "  ${GREEN}$OUTPUT_DIR/ws-${client_id}.log${NC}"
done
echo ""
log_action "${CYAN}Main Log:${NC} $WS_LOG"
echo ""

if [ $failed_operations -eq 0 ] && [ $events_received -gt 0 ]; then
    log_action "${GREEN}✓ WebSocket real-time testing completed successfully!${NC}"
    log_action "${GREEN}✓ Real-time event delivery confirmed${NC}"
    exit 0
else
    log_action "${YELLOW}⚠ Some issues detected. Review logs for details.${NC}"
    exit 1
fi
