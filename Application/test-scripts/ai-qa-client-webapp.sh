#!/bin/bash

# AI QA - Web Application Client Simulation
# Simulates a web application user interacting with HelixTrack Core

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
CLIENT_LOG="$OUTPUT_DIR/webapp-client.log"

# Simulation parameters
SIMULATION_DURATION="${SIMULATION_DURATION:-300}"  # 5 minutes default
ACTION_INTERVAL="${ACTION_INTERVAL:-10}"            # 10 seconds between actions

# Counters
actions_performed=0
successful_actions=0
failed_actions=0

echo -e "${MAGENTA}========================================${NC}"
echo -e "${MAGENTA}Web App Client Simulation${NC}"
echo -e "${MAGENTA}========================================${NC}"
echo ""

# Check prerequisites
if [ ! -f "$USERS_FILE" ] || [ ! -f "$PROJECTS_FILE" ]; then
    echo -e "${RED}Error: Setup not completed!${NC}"
    echo -e "${YELLOW}Run ai-qa-setup-organization.sh and ai-qa-setup-projects.sh first${NC}"
    exit 1
fi

# Initialize log
echo "Web App Client Simulation - $(date)" > "$CLIENT_LOG"
echo "======================================" >> "$CLIENT_LOG"
echo "" >> "$CLIENT_LOG"

# Select a random user for this session
USER_COUNT=$(jq '.users | length' "$USERS_FILE")
RANDOM_USER_INDEX=$((RANDOM % USER_COUNT))
USER_NAME=$(jq -r ".users[$RANDOM_USER_INDEX].name" "$USERS_FILE")
USER_TOKEN=$(jq -r ".users[$RANDOM_USER_INDEX].token" "$USERS_FILE")
USER_EMAIL=$(jq -r ".users[$RANDOM_USER_INDEX].email" "$USERS_FILE")

echo -e "${CYAN}Simulating User:${NC} $USER_NAME ($USER_EMAIL)"
echo -e "${CYAN}Session Duration:${NC} $SIMULATION_DURATION seconds"
echo -e "${CYAN}Action Interval:${NC} $ACTION_INTERVAL seconds"
echo ""

log_action() {
    local message="$1"
    echo -e "[$(date '+%Y-%m-%d %H:%M:%S')] $message" >> "$CLIENT_LOG"
    echo -e "$message"
}

# Function to make API call
api_call() {
    local action="$1"
    local data="$2"
    local description="$3"

    actions_performed=$((actions_performed + 1))
    log_action "${YELLOW}[Web App] $description${NC}"

    local body="{\"action\": \"$action\", \"jwt\": \"$USER_TOKEN\", \"data\": $data}"

    local response
    response=$(curl -s -X POST "$BASE_URL/do" \
        -H "Content-Type: application/json" \
        -d "$body")

    local error_code
    error_code=$(echo "$response" | grep -o '"errorCode":[^,}]*' | head -1 | cut -d':' -f2 | tr -d ' ')

    if [ "$error_code" = "-1" ]; then
        log_action "  ${GREEN}✓ SUCCESS${NC}"
        successful_actions=$((successful_actions + 1))
        echo "$response"
        return 0
    else
        log_action "  ${RED}✗ FAILED (errorCode: $error_code)${NC}"
        failed_actions=$((failed_actions + 1))
        return 1
    fi
}

# Function to simulate typical web app actions
simulate_web_actions() {
    local action_type=$((RANDOM % 15))

    case $action_type in
        0|1|2)
            # View dashboard / list projects (most common action)
            log_action "${BLUE}→ Dashboard: Loading projects list${NC}"
            api_call "projectList" "{}" "Fetching projects list" > /dev/null
            ;;
        3|4)
            # View tickets in a project
            local project_count=$(jq '.projects | length' "$PROJECTS_FILE")
            if [ "$project_count" -gt 0 ]; then
                local proj_index=$((RANDOM % project_count))
                local project_id=$(jq -r ".projects[$proj_index].id" "$PROJECTS_FILE")
                local project_key=$(jq -r ".projects[$proj_index].key" "$PROJECTS_FILE")

                log_action "${BLUE}→ Ticket Board: Viewing tickets for $project_key${NC}"
                api_call "ticketList" "{\"projectId\": \"$project_id\"}" "Loading tickets for $project_key" > /dev/null
            fi
            ;;
        5)
            # Create a new ticket (comment)
            local project_count=$(jq '.projects | length' "$PROJECTS_FILE")
            if [ "$project_count" -gt 0 ]; then
                local proj_index=$((RANDOM % project_count))
                local project_id=$(jq -r ".projects[$proj_index].id" "$PROJECTS_FILE")
                local project_key=$(jq -r ".projects[$proj_index].key" "$PROJECTS_FILE")
                local type_id=$(jq -r ".projects[$proj_index].types[] | select(.name == \"Task\") | .id" "$PROJECTS_FILE")
                local status_id=$(jq -r ".projects[$proj_index].statuses[0].id" "$PROJECTS_FILE")
                local priority_id=$(jq -r ".projects[$proj_index].priorities[1].id" "$PROJECTS_FILE")

                local ticket_title="Web App Task - $(date +%s)"
                local ticket_desc="Task created via web application at $(date)"

                log_action "${BLUE}→ Create Ticket: New task in $project_key${NC}"
                TICKET_RESPONSE=$(api_call "ticketCreate" "{\"projectId\": \"$project_id\", \"title\": \"$ticket_title\", \"description\": \"$ticket_desc\", \"typeId\": \"$type_id\", \"statusId\": \"$status_id\", \"priorityId\": \"$priority_id\"}" "Creating new ticket")
                TICKET_ID=$(echo "$TICKET_RESPONSE" | jq -r '.data.id // empty')

                if [ -n "$TICKET_ID" ]; then
                    log_action "  ${GREEN}Created ticket ID: $TICKET_ID${NC}"
                fi
            fi
            ;;
        6)
            # Add comment to a ticket
            log_action "${BLUE}→ Comment: Adding comment to ticket${NC}"
            local comment_text="Web app comment at $(date): This looks good, proceeding with implementation."
            # Note: Would need a ticket ID - simplified for simulation
            ;;
        7)
            # Update ticket status (drag-and-drop simulation)
            log_action "${BLUE}→ Status Update: Moving ticket to 'In Progress'${NC}"
            # Note: Would need a ticket ID - simplified for simulation
            ;;
        8)
            # View filters
            log_action "${BLUE}→ Filters: Loading saved filters${NC}"
            api_call "filterList" "{}" "Fetching saved filters" > /dev/null
            ;;
        9)
            # Search tickets
            local search_query="bug"
            log_action "${BLUE}→ Search: Searching for '$search_query'${NC}"
            api_call "ticketSearch" "{\"query\": \"$search_query\"}" "Searching tickets" > /dev/null
            ;;
        10)
            # View user profile
            log_action "${BLUE}→ Profile: Viewing user settings${NC}"
            api_call "userProfile" "{}" "Loading user profile" > /dev/null
            ;;
        11)
            # View project settings
            local project_count=$(jq '.projects | length' "$PROJECTS_FILE")
            if [ "$project_count" -gt 0 ]; then
                local proj_index=$((RANDOM % project_count))
                local project_id=$(jq -r ".projects[$proj_index].id" "$PROJECTS_FILE")
                local project_key=$(jq -r ".projects[$proj_index].key" "$PROJECTS_FILE")

                log_action "${BLUE}→ Settings: Viewing $project_key settings${NC}"
                api_call "projectRead" "{\"id\": \"$project_id\"}" "Loading project settings" > /dev/null
            fi
            ;;
        12)
            # View reports/analytics
            log_action "${BLUE}→ Reports: Loading project analytics${NC}"
            api_call "reportList" "{}" "Fetching reports" > /dev/null
            ;;
        13)
            # Check notifications
            log_action "${BLUE}→ Notifications: Checking new notifications${NC}"
            api_call "notificationList" "{}" "Fetching notifications" > /dev/null
            ;;
        14)
            # Refresh activity feed
            log_action "${BLUE}→ Activity: Loading recent activity${NC}"
            api_call "activityList" "{}" "Fetching recent activity" > /dev/null
            ;;
    esac
}

# Main simulation loop
start_time=$(date +%s)
end_time=$((start_time + SIMULATION_DURATION))

log_action "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
log_action "${GREEN}Web App Session Started${NC}"
log_action "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

# Initial page load - dashboard
log_action "${BLUE}→ Page Load: Loading dashboard${NC}"
api_call "health" "{}" "Health check" > /dev/null
api_call "projectList" "{}" "Initial projects load" > /dev/null
echo ""

# Simulation loop
while [ $(date +%s) -lt $end_time ]; do
    simulate_web_actions
    echo ""
    sleep $ACTION_INTERVAL
done

# Summary
echo ""
log_action "${MAGENTA}========================================${NC}"
log_action "${MAGENTA}Web App Session Complete${NC}"
log_action "${MAGENTA}========================================${NC}"
echo ""
log_action "${CYAN}Session Summary:${NC}"
log_action "  User: $USER_NAME"
log_action "  Duration: $SIMULATION_DURATION seconds"
log_action "  Actions Performed: $actions_performed"
log_action "  ${GREEN}Successful: $successful_actions${NC}"
log_action "  ${RED}Failed: $failed_actions${NC}"
log_action "  Success Rate: $(awk "BEGIN {printf \"%.1f\", ($successful_actions/$actions_performed)*100}")%"
echo ""
log_action "${CYAN}Log File:${NC} $CLIENT_LOG"
echo ""

if [ $failed_actions -eq 0 ]; then
    log_action "${GREEN}✓ All web app actions completed successfully!${NC}"
    exit 0
else
    log_action "${YELLOW}⚠ Some actions failed. Check log for details.${NC}"
    exit 1
fi
