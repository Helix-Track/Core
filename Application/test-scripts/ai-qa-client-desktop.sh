#!/bin/bash

# AI QA - Desktop Application Client Simulation
# Simulates a desktop application user interacting with HelixTrack Core

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
CLIENT_LOG="$OUTPUT_DIR/desktop-client.log"

# Simulation parameters (desktop users are power users)
SIMULATION_DURATION="${SIMULATION_DURATION:-300}"  # 5 minutes default
ACTION_INTERVAL="${ACTION_INTERVAL:-8}"             # 8 seconds between actions (faster than web/mobile)

# Counters
actions_performed=0
successful_actions=0
failed_actions=0

echo -e "${MAGENTA}========================================${NC}"
echo -e "${MAGENTA}Desktop App Client Simulation${NC}"
echo -e "${MAGENTA}========================================${NC}"
echo ""

# Check prerequisites
if [ ! -f "$USERS_FILE" ] || [ ! -f "$PROJECTS_FILE" ]; then
    echo -e "${RED}Error: Setup not completed!${NC}"
    echo -e "${YELLOW}Run ai-qa-setup-organization.sh and ai-qa-setup-projects.sh first${NC}"
    exit 1
fi

# Initialize log
echo "Desktop App Client Simulation - $(date)" > "$CLIENT_LOG"
echo "======================================" >> "$CLIENT_LOG"
echo "" >> "$CLIENT_LOG"

# Select a random user for this session
USER_COUNT=$(jq '.users | length' "$USERS_FILE")
RANDOM_USER_INDEX=$((RANDOM % USER_COUNT))
USER_NAME=$(jq -r ".users[$RANDOM_USER_INDEX].name" "$USERS_FILE")
USER_TOKEN=$(jq -r ".users[$RANDOM_USER_INDEX].token" "$USERS_FILE")
USER_EMAIL=$(jq -r ".users[$RANDOM_USER_INDEX].email" "$USERS_FILE")

echo -e "${CYAN}Simulating User:${NC} $USER_NAME ($USER_EMAIL)"
echo -e "${CYAN}Platform:${NC} Desktop App (Electron, Linux x64)"
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
    log_action "${YELLOW}[Desktop] $description${NC}"

    local body="{\"action\": \"$action\", \"jwt\": \"$USER_TOKEN\", \"data\": $data, \"clientType\": \"desktop\", \"clientVersion\": \"3.1.2\"}"

    local response
    response=$(curl -s -X POST "$BASE_URL/do" \
        -H "Content-Type: application/json" \
        -H "User-Agent: HelixTrack-Desktop/3.1.2 (Linux x64; Electron 27.0)" \
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

# Function to simulate typical Desktop app actions (power user features)
simulate_desktop_actions() {
    local action_type=$((RANDOM % 18))

    case $action_type in
        0)
            # Bulk operations (desktop power feature)
            log_action "${BLUE}🖥️ Bulk Update: Processing multiple tickets${NC}"
            # Note: Would need ticket IDs in production
            ;;
        1)
            # Advanced filtering and queries
            log_action "${BLUE}🖥️ Advanced Filter: Complex JQL query${NC}"
            local filter_query='status IN ("In Progress", "Ready for QA") AND assignee = currentUser() ORDER BY priority DESC'
            api_call "ticketSearch" "{\"jql\": \"$filter_query\"}" "Advanced search" > /dev/null
            ;;
        2)
            # Export data (desktop feature)
            log_action "${BLUE}🖥️ Export: Generating CSV report${NC}"
            api_call "reportExport" "{\"format\": \"csv\", \"type\": \"tickets\"}" "Exporting data" > /dev/null
            ;;
        3)
            # Create multiple tickets from template
            local project_count=$(jq '.projects | length' "$PROJECTS_FILE")
            if [ "$project_count" -gt 0 ]; then
                local proj_index=$((RANDOM % project_count))
                local project_id=$(jq -r ".projects[$proj_index].id" "$PROJECTS_FILE")
                local project_key=$(jq -r ".projects[$proj_index].key" "$PROJECTS_FILE")
                local type_id=$(jq -r ".projects[$proj_index].types[] | select(.name == \"Task\") | .id" "$PROJECTS_FILE")
                local status_id=$(jq -r ".projects[$proj_index].statuses[0].id" "$PROJECTS_FILE")
                local priority_id=$(jq -r ".projects[$proj_index].priorities[2].id" "$PROJECTS_FILE")

                log_action "${BLUE}🖥️ Batch Create: Creating tickets from template${NC}"

                for i in {1..3}; do
                    local ticket_title="Desktop Batch Task $i - $(date +%s)"
                    local ticket_desc="Task created via desktop app batch operation"

                    api_call "ticketCreate" "{\"projectId\": \"$project_id\", \"title\": \"$ticket_title\", \"description\": \"$ticket_desc\", \"typeId\": \"$type_id\", \"statusId\": \"$status_id\", \"priorityId\": \"$priority_id\"}" "Creating batch ticket $i" > /dev/null
                done
            fi
            ;;
        4)
            # Keyboard shortcuts - rapid ticket navigation
            log_action "${BLUE}🖥️ Keyboard Nav: Quick ticket switching (Ctrl+K)${NC}"
            api_call "ticketSearch" "{\"query\": \"\"}" "Quick search" > /dev/null
            ;;
        5)
            # Git integration (desktop feature)
            log_action "${BLUE}🖥️ Git Integration: Linking commits to tickets${NC}"
            # Note: Would need ticket ID and commit hash in production
            ;;
        6)
            # Local caching and offline mode
            log_action "${BLUE}🖥️ Cache: Syncing local database${NC}"
            api_call "syncStatus" "{}" "Checking sync status" > /dev/null
            ;;
        7)
            # Advanced reporting and charts
            log_action "${BLUE}🖥️ Analytics: Generating burndown chart${NC}"
            api_call "reportBurndown" "{\"sprintId\": \"sprint-1\"}" "Generating burndown" > /dev/null
            ;;
        8)
            # Multi-window support - view multiple projects
            local project_count=$(jq '.projects | length' "$PROJECTS_FILE")
            if [ "$project_count" -gt 1 ]; then
                log_action "${BLUE}🖥️ Multi-Window: Opening multiple projects${NC}"
                for p in $(seq 0 $((project_count - 1))); do
                    local project_id=$(jq -r ".projects[$p].id" "$PROJECTS_FILE")
                    api_call "projectRead" "{\"id\": \"$project_id\"}" "Loading project in window $((p+1))" > /dev/null
                done
            fi
            ;;
        9)
            # Code review integration
            log_action "${BLUE}🖥️ Code Review: Viewing PR linked to ticket${NC}"
            api_call "repositoryPullRequestList" "{}" "Fetching PRs" > /dev/null
            ;;
        10)
            # Time tracking (desktop feature)
            log_action "${BLUE}🖥️ Time Tracking: Logging work hours${NC}"
            api_call "worklogCreate" "{\"hours\": 2.5, \"description\": \"Desktop work session\"}" "Logging work time" > /dev/null
            ;;
        11)
            # Advanced notifications with system tray
            log_action "${BLUE}🖥️ System Tray: Checking background notifications${NC}"
            api_call "notificationList" "{\"unreadOnly\": true}" "System tray notifications" > /dev/null
            ;;
        12)
            # Drag and drop file attachments
            log_action "${BLUE}🖥️ File Upload: Attaching local files${NC}"
            # Note: Would need ticket ID and file upload in production
            ;;
        13)
            # Admin operations (desktop power users)
            log_action "${BLUE}🖥️ Admin: Managing project settings${NC}"
            api_call "projectList" "{}" "Admin project view" > /dev/null
            ;;
        14)
            # Custom fields bulk edit
            log_action "${BLUE}🖥️ Custom Fields: Bulk updating fields${NC}"
            api_call "customFieldList" "{}" "Loading custom fields" > /dev/null
            ;;
        15)
            # Advanced workflow transitions
            log_action "${BLUE}🖥️ Workflow: Viewing transition rules${NC}"
            api_call "workflowList" "{}" "Loading workflows" > /dev/null
            ;;
        16)
            # Desktop notifications for mentions
            log_action "${BLUE}🖥️ Mentions: Checking @mentions${NC}"
            api_call "commentMentionsList" "{}" "Loading mentions" > /dev/null
            ;;
        17)
            # Hotkey operations - quick actions
            log_action "${BLUE}🖥️ Hotkey: Quick create ticket (Ctrl+Shift+N)${NC}"
            local project_count=$(jq '.projects | length' "$PROJECTS_FILE")
            if [ "$project_count" -gt 0 ]; then
                local proj_index=$((RANDOM % project_count))
                local project_id=$(jq -r ".projects[$proj_index].id" "$PROJECTS_FILE")
                local type_id=$(jq -r ".projects[$proj_index].types[] | select(.name == \"Task\") | .id" "$PROJECTS_FILE")
                local status_id=$(jq -r ".projects[$proj_index].statuses[0].id" "$PROJECTS_FILE")
                local priority_id=$(jq -r ".projects[$proj_index].priorities[1].id" "$PROJECTS_FILE")

                local ticket_title="Hotkey Task - $(date +%s)"
                api_call "ticketCreate" "{\"projectId\": \"$project_id\", \"title\": \"$ticket_title\", \"typeId\": \"$type_id\", \"statusId\": \"$status_id\", \"priorityId\": \"$priority_id\"}" "Quick ticket creation" > /dev/null
            fi
            ;;
    esac
}

# Main simulation loop
start_time=$(date +%s)
end_time=$((start_time + SIMULATION_DURATION))

log_action "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
log_action "${GREEN}Desktop App Session Started${NC}"
log_action "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

# App launch - comprehensive initialization
log_action "${BLUE}🖥️ App Launch: Starting desktop app${NC}"
api_call "health" "{}" "Health check" > /dev/null
api_call "userProfile" "{}" "Loading user profile" > /dev/null
api_call "projectList" "{}" "Initial project load" > /dev/null
api_call "notificationList" "{}" "Initial notifications" > /dev/null
api_call "activityList" "{\"limit\": 50}" "Recent activity" > /dev/null
echo ""

# Simulation loop
while [ $(date +%s) -lt $end_time ]; do
    simulate_desktop_actions
    echo ""
    sleep $ACTION_INTERVAL
done

# Summary
echo ""
log_action "${MAGENTA}========================================${NC}"
log_action "${MAGENTA}Desktop App Session Complete${NC}"
log_action "${MAGENTA}========================================${NC}"
echo ""
log_action "${CYAN}Session Summary:${NC}"
log_action "  User: $USER_NAME"
log_action "  Platform: Desktop (Electron, Linux x64)"
log_action "  Duration: $SIMULATION_DURATION seconds"
log_action "  Actions Performed: $actions_performed"
log_action "  ${GREEN}Successful: $successful_actions${NC}"
log_action "  ${RED}Failed: $failed_actions${NC}"
log_action "  Success Rate: $(awk "BEGIN {printf \"%.1f\", ($successful_actions/$actions_performed)*100}")%"
echo ""
log_action "${CYAN}Log File:${NC} $CLIENT_LOG"
echo ""

if [ $failed_actions -eq 0 ]; then
    log_action "${GREEN}✓ All desktop app actions completed successfully!${NC}"
    exit 0
else
    log_action "${YELLOW}⚠ Some actions failed. Check log for details.${NC}"
    exit 1
fi
