#!/bin/bash

# AI QA - Comprehensive Test Orchestrator
# Master script that runs the complete AI QA test suite

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
MAGENTA='\033[0;35m'
WHITE='\033[1;37m'
NC='\033[0m' # No Color

# Configuration
BASE_URL="${BASE_URL:-http://localhost:8080}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
OUTPUT_DIR="$SCRIPT_DIR/ai-qa-output"
REPORT_FILE="$OUTPUT_DIR/AI_QA_COMPREHENSIVE_REPORT.md"

# Test parameters
CLIENT_DURATION="${CLIENT_DURATION:-300}"    # 5 minutes per client
WS_DURATION="${WS_DURATION:-120}"            # 2 minutes WebSocket test
CONCURRENT_CLIENTS="${CONCURRENT_CLIENTS:-3}"

# Tracking
start_time=$(date +%s)
phase_results=()

echo -e "${WHITE}╔════════════════════════════════════════════════════════╗${NC}"
echo -e "${WHITE}║                                                        ║${NC}"
echo -e "${WHITE}║     ${MAGENTA}HelixTrack Core - AI QA Comprehensive Test${WHITE}      ║${NC}"
echo -e "${WHITE}║                                                        ║${NC}"
echo -e "${WHITE}╚════════════════════════════════════════════════════════╝${NC}"
echo ""
echo -e "${CYAN}Test Configuration:${NC}"
echo -e "  Target URL: ${GREEN}$BASE_URL${NC}"
echo -e "  Output Directory: ${GREEN}$OUTPUT_DIR${NC}"
echo -e "  Client Test Duration: ${GREEN}$CLIENT_DURATION seconds${NC}"
echo -e "  WebSocket Test Duration: ${GREEN}$WS_DURATION seconds${NC}"
echo -e "  Concurrent WS Clients: ${GREEN}$CONCURRENT_CLIENTS${NC}"
echo ""

# Create output directory
mkdir -p "$OUTPUT_DIR"

# Initialize report
cat > "$REPORT_FILE" << 'EOF'
# HelixTrack Core - AI QA Comprehensive Test Report

**Generated:** $(date '+%Y-%m-%d %H:%M:%S')

## Executive Summary

This report documents the comprehensive AI QA testing of HelixTrack Core V3.0, simulating real-world enterprise usage across multiple client applications with extensive real-time WebSocket testing.

---

## Test Execution

EOF

# Function to log phase result
log_phase() {
    local phase_name="$1"
    local status="$2"
    local details="$3"

    phase_results+=("$phase_name|$status|$details")

    if [ "$status" = "SUCCESS" ]; then
        echo -e "${GREEN}✓ $phase_name: PASSED${NC}"
    elif [ "$status" = "FAILED" ]; then
        echo -e "${RED}✗ $phase_name: FAILED${NC}"
    else
        echo -e "${YELLOW}⚠ $phase_name: $status${NC}"
    fi

    if [ -n "$details" ]; then
        echo -e "  ${CYAN}$details${NC}"
    fi
    echo ""
}

# Function to check if server is running
check_server() {
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BLUE}Pre-Flight Check${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""

    echo -e "${YELLOW}Checking if HelixTrack Core is running...${NC}"

    if curl -s "$BASE_URL/do" -X POST -H "Content-Type: application/json" -d '{"action":"version"}' | grep -q '"errorCode":-1'; then
        echo -e "${GREEN}✓ Server is running${NC}"

        # Get version info
        VERSION_RESPONSE=$(curl -s "$BASE_URL/do" -X POST -H "Content-Type: application/json" -d '{"action":"version"}')
        VERSION=$(echo "$VERSION_RESPONSE" | jq -r '.data.version // "unknown"')
        BUILD=$(echo "$VERSION_RESPONSE" | jq -r '.data.build // "unknown"')

        echo -e "${CYAN}  Version: $VERSION${NC}"
        echo -e "${CYAN}  Build: $BUILD${NC}"
        echo ""
        return 0
    else
        echo -e "${RED}✗ Server is not responding${NC}"
        echo -e "${YELLOW}Please start HelixTrack Core:${NC}"
        echo -e "${CYAN}  cd Application${NC}"
        echo -e "${CYAN}  ./htCore${NC}"
        echo ""
        exit 1
    fi
}

# Phase 1: Organization Setup
run_phase1() {
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BLUE}Phase 1: Organization Setup${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""

    if [ -f "$OUTPUT_DIR/users.json" ] && [ -f "$OUTPUT_DIR/teams.json" ]; then
        echo -e "${YELLOW}Organization already exists. Skipping setup...${NC}"
        echo -e "${CYAN}To re-run setup, delete: $OUTPUT_DIR/*.json${NC}"
        log_phase "Phase 1: Organization Setup" "SKIPPED" "Organization already configured"
        echo ""
        return 0
    fi

    echo -e "${YELLOW}Creating organization structure...${NC}"
    if "$SCRIPT_DIR/ai-qa-setup-organization.sh"; then
        log_phase "Phase 1: Organization Setup" "SUCCESS" "Account, organization, 3 teams, 11 users created"

        # Add to report
        echo "### Phase 1: Organization Setup ✅" >> "$REPORT_FILE"
        echo "" >> "$REPORT_FILE"
        echo "**Status:** SUCCESS" >> "$REPORT_FILE"
        echo "" >> "$REPORT_FILE"
        echo "**Created:**" >> "$REPORT_FILE"
        echo "- 1 Account (TechCorp Global)" >> "$REPORT_FILE"
        echo "- 1 Organization (TechCorp Engineering)" >> "$REPORT_FILE"
        echo "- 3 Teams (Frontend, Backend, QA & DevOps)" >> "$REPORT_FILE"
        echo "- 11 Users (authenticated and ready)" >> "$REPORT_FILE"
        echo "" >> "$REPORT_FILE"
    else
        log_phase "Phase 1: Organization Setup" "FAILED" "Check logs for details"
        echo "### Phase 1: Organization Setup ❌" >> "$REPORT_FILE"
        echo "" >> "$REPORT_FILE"
        echo "**Status:** FAILED" >> "$REPORT_FILE"
        echo "" >> "$REPORT_FILE"
        return 1
    fi
}

# Phase 2: Project Workflows Setup
run_phase2() {
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BLUE}Phase 2: Project Workflows Setup${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""

    if [ -f "$OUTPUT_DIR/projects.json" ]; then
        echo -e "${YELLOW}Projects already exist. Skipping setup...${NC}"
        echo -e "${CYAN}To re-run setup, delete: $OUTPUT_DIR/projects.json${NC}"
        log_phase "Phase 2: Project Workflows" "SKIPPED" "Projects already configured"
        echo ""
        return 0
    fi

    echo -e "${YELLOW}Creating projects with workflows...${NC}"
    if "$SCRIPT_DIR/ai-qa-setup-projects.sh"; then
        log_phase "Phase 2: Project Workflows" "SUCCESS" "4 projects with epics, stories, tasks created"

        # Add to report
        echo "### Phase 2: Project Workflows Setup ✅" >> "$REPORT_FILE"
        echo "" >> "$REPORT_FILE"
        echo "**Status:** SUCCESS" >> "$REPORT_FILE"
        echo "" >> "$REPORT_FILE"
        echo "**Created:**" >> "$REPORT_FILE"
        echo "- 4 Projects (BANK, UNI, CHAT, SHOP)" >> "$REPORT_FILE"
        echo "- Multiple epics per project" >> "$REPORT_FILE"
        echo "- Stories, tasks, and subtasks" >> "$REPORT_FILE"
        echo "- Custom priorities, statuses, and ticket types" >> "$REPORT_FILE"
        echo "" >> "$REPORT_FILE"
    else
        log_phase "Phase 2: Project Workflows" "FAILED" "Check logs for details"
        echo "### Phase 2: Project Workflows Setup ❌" >> "$REPORT_FILE"
        echo "" >> "$REPORT_FILE"
        echo "**Status:** FAILED" >> "$REPORT_FILE"
        echo "" >> "$REPORT_FILE"
        return 1
    fi
}

# Phase 3: Client Simulations (parallel execution)
run_phase3() {
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BLUE}Phase 3: Client Application Simulations${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""

    echo -e "${YELLOW}Starting client simulations (running in parallel)...${NC}"
    echo ""

    # Start all clients in background
    SIMULATION_DURATION=$CLIENT_DURATION "$SCRIPT_DIR/ai-qa-client-webapp.sh" > "$OUTPUT_DIR/webapp-run.log" 2>&1 &
    WEBAPP_PID=$!
    echo -e "${CYAN}  Started Web App client (PID: $WEBAPP_PID)${NC}"

    SIMULATION_DURATION=$CLIENT_DURATION "$SCRIPT_DIR/ai-qa-client-android.sh" > "$OUTPUT_DIR/android-run.log" 2>&1 &
    ANDROID_PID=$!
    echo -e "${CYAN}  Started Android client (PID: $ANDROID_PID)${NC}"

    SIMULATION_DURATION=$CLIENT_DURATION "$SCRIPT_DIR/ai-qa-client-desktop.sh" > "$OUTPUT_DIR/desktop-run.log" 2>&1 &
    DESKTOP_PID=$!
    echo -e "${CYAN}  Started Desktop client (PID: $DESKTOP_PID)${NC}"

    echo ""
    echo -e "${YELLOW}Clients are running... (Duration: $CLIENT_DURATION seconds)${NC}"
    echo ""

    # Wait for all clients to complete
    webapp_success=0
    android_success=0
    desktop_success=0

    wait $WEBAPP_PID && webapp_success=1 || webapp_success=0
    wait $ANDROID_PID && android_success=1 || android_success=0
    wait $DESKTOP_PID && desktop_success=1 || desktop_success=0

    # Report results
    echo ""
    if [ $webapp_success -eq 1 ]; then
        echo -e "${GREEN}  ✓ Web App client: SUCCESS${NC}"
    else
        echo -e "${RED}  ✗ Web App client: FAILED${NC}"
    fi

    if [ $android_success -eq 1 ]; then
        echo -e "${GREEN}  ✓ Android client: SUCCESS${NC}"
    else
        echo -e "${RED}  ✗ Android client: FAILED${NC}"
    fi

    if [ $desktop_success -eq 1 ]; then
        echo -e "${GREEN}  ✓ Desktop client: SUCCESS${NC}"
    else
        echo -e "${RED}  ✗ Desktop client: FAILED${NC}"
    fi

    echo ""

    if [ $webapp_success -eq 1 ] && [ $android_success -eq 1 ] && [ $desktop_success -eq 1 ]; then
        log_phase "Phase 3: Client Simulations" "SUCCESS" "All 3 clients completed successfully"

        # Add to report
        echo "### Phase 3: Client Application Simulations ✅" >> "$REPORT_FILE"
        echo "" >> "$REPORT_FILE"
        echo "**Status:** SUCCESS" >> "$REPORT_FILE"
        echo "" >> "$REPORT_FILE"
        echo "**Clients Tested:**" >> "$REPORT_FILE"
        echo "- ✅ Web Application" >> "$REPORT_FILE"
        echo "- ✅ Android Mobile App" >> "$REPORT_FILE"
        echo "- ✅ Desktop Application" >> "$REPORT_FILE"
        echo "" >> "$REPORT_FILE"
        echo "**Duration:** $CLIENT_DURATION seconds per client" >> "$REPORT_FILE"
        echo "" >> "$REPORT_FILE"
    else
        log_phase "Phase 3: Client Simulations" "PARTIAL" "Some clients failed"
        echo "### Phase 3: Client Application Simulations ⚠️" >> "$REPORT_FILE"
        echo "" >> "$REPORT_FILE"
        echo "**Status:** PARTIAL SUCCESS" >> "$REPORT_FILE"
        echo "" >> "$REPORT_FILE"
        return 1
    fi
}

# Phase 4: WebSocket Real-Time Testing
run_phase4() {
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BLUE}Phase 4: WebSocket Real-Time Testing${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""

    echo -e "${YELLOW}Starting WebSocket real-time event testing...${NC}"
    echo ""

    if TEST_DURATION=$WS_DURATION CONCURRENT_CLIENTS=$CONCURRENT_CLIENTS "$SCRIPT_DIR/ai-qa-websocket-realtime.sh"; then
        log_phase "Phase 4: WebSocket Testing" "SUCCESS" "$CONCURRENT_CLIENTS concurrent clients, real-time events verified"

        # Add to report
        echo "### Phase 4: WebSocket Real-Time Testing ✅" >> "$REPORT_FILE"
        echo "" >> "$REPORT_FILE"
        echo "**Status:** SUCCESS" >> "$REPORT_FILE"
        echo "" >> "$REPORT_FILE"
        echo "**Testing:**" >> "$REPORT_FILE"
        echo "- Concurrent WebSocket connections: $CONCURRENT_CLIENTS" >> "$REPORT_FILE"
        echo "- Real-time event delivery verified" >> "$REPORT_FILE"
        echo "- Event types: ticket.created, ticket.updated, project.created, etc." >> "$REPORT_FILE"
        echo "" >> "$REPORT_FILE"
    else
        log_phase "Phase 4: WebSocket Testing" "FAILED" "Check logs for details"
        echo "### Phase 4: WebSocket Real-Time Testing ❌" >> "$REPORT_FILE"
        echo "" >> "$REPORT_FILE"
        echo "**Status:** FAILED" >> "$REPORT_FILE"
        echo "" >> "$REPORT_FILE"
        return 1
    fi
}

# Main execution
main() {
    check_server

    # Run all phases
    run_phase1
    run_phase2
    run_phase3
    run_phase4

    # Calculate total duration
    end_time=$(date +%s)
    total_duration=$((end_time - start_time))
    minutes=$((total_duration / 60))
    seconds=$((total_duration % 60))

    # Final report
    echo -e "${WHITE}╔════════════════════════════════════════════════════════╗${NC}"
    echo -e "${WHITE}║                                                        ║${NC}"
    echo -e "${WHITE}║              ${GREEN}Comprehensive Test Complete${WHITE}              ║${NC}"
    echo -e "${WHITE}║                                                        ║${NC}"
    echo -e "${WHITE}╚════════════════════════════════════════════════════════╝${NC}"
    echo ""
    echo -e "${CYAN}Test Summary:${NC}"
    echo -e "  Total Duration: ${GREEN}${minutes}m ${seconds}s${NC}"
    echo -e "  Phases Completed: ${GREEN}${#phase_results[@]}${NC}"
    echo ""

    # Count successes and failures
    success_count=0
    failed_count=0
    skipped_count=0

    for result in "${phase_results[@]}"; do
        IFS='|' read -r phase status details <<< "$result"
        if [ "$status" = "SUCCESS" ]; then
            success_count=$((success_count + 1))
            echo -e "  ${GREEN}✓ $phase${NC}"
        elif [ "$status" = "FAILED" ]; then
            failed_count=$((failed_count + 1))
            echo -e "  ${RED}✗ $phase${NC}"
        elif [ "$status" = "PARTIAL" ]; then
            failed_count=$((failed_count + 1))
            echo -e "  ${YELLOW}⚠ $phase${NC}"
        else
            skipped_count=$((skipped_count + 1))
            echo -e "  ${CYAN}○ $phase ($status)${NC}"
        fi
    done

    echo ""
    echo -e "${CYAN}Results:${NC}"
    echo -e "  ${GREEN}Success: $success_count${NC}"
    echo -e "  ${RED}Failed: $failed_count${NC}"
    echo -e "  ${CYAN}Skipped: $skipped_count${NC}"
    echo ""

    # Finalize report
    cat >> "$REPORT_FILE" << EOF

---

## Final Summary

**Total Test Duration:** ${minutes}m ${seconds}s
**Phases Executed:** ${#phase_results[@]}
**Success Rate:** $(awk "BEGIN {printf \"%.1f\", (($success_count/$((success_count + failed_count)))*100)}")%

### Results

- ✅ **Successful:** $success_count
- ❌ **Failed:** $failed_count
- ⏭️ **Skipped:** $skipped_count

---

## Output Files

All test outputs are located in: \`$OUTPUT_DIR/\`

- Organization data: \`tokens.json\`, \`users.json\`, \`teams.json\`
- Project data: \`projects.json\`
- Client logs: \`webapp-client.log\`, \`android-client.log\`, \`desktop-client.log\`
- WebSocket logs: \`websocket-realtime.log\`, \`ws-client-*.log\`

---

**Report Generated:** $(date '+%Y-%m-%d %H:%M:%S')
**HelixTrack Core Version:** 3.0.0
**Test Suite:** AI QA Comprehensive Test

EOF

    echo -e "${CYAN}Report Generated:${NC} ${GREEN}$REPORT_FILE${NC}"
    echo ""

    if [ $failed_count -eq 0 ]; then
        echo -e "${GREEN}✓ All tests completed successfully!${NC}"
        echo ""
        exit 0
    else
        echo -e "${YELLOW}⚠ Some tests failed. Review the report and logs for details.${NC}"
        echo ""
        exit 1
    fi
}

# Run main
main
