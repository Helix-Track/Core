#!/bin/bash
#
# HelixTrack Core - AI QA Test Runner
# Runs AI-powered QA tests against running instances
#

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Configuration
SQLITE_PORT=8080
POSTGRES_PORT=8081
SERVER_STARTUP_TIMEOUT=30
TEST_TIMEOUT=300

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_section() {
    echo
    echo -e "${CYAN}============================================${NC}"
    echo -e "${CYAN}$1${NC}"
    echo -e "${CYAN}============================================${NC}"
    echo
}

# Check Python installation
check_python() {
    if ! command -v python3 &> /dev/null; then
        log_error "Python 3 is not installed"
        log_info "Run: ./scripts/setup-environment.sh"
        exit 1
    fi

    PYTHON_VERSION=$(python3 --version 2>&1 | awk '{print $2}')
    log_info "Using Python: ${PYTHON_VERSION}"
}

# Check Python dependencies
check_python_deps() {
    log_info "Checking Python dependencies..."

    MISSING_DEPS=()

    # Check for requests
    if ! python3 -c "import requests" 2>/dev/null; then
        MISSING_DEPS+=("requests")
    fi

    # Check for colorama
    if ! python3 -c "import colorama" 2>/dev/null; then
        MISSING_DEPS+=("colorama")
    fi

    if [ ${#MISSING_DEPS[@]} -gt 0 ]; then
        log_error "Missing Python dependencies: ${MISSING_DEPS[*]}"
        log_info "Installing dependencies..."
        pip3 install --user "${MISSING_DEPS[@]}"
    else
        log_success "All Python dependencies are installed"
    fi
}

# Build the application
build_app() {
    log_section "Building Application"

    cd "$PROJECT_ROOT"

    if [ ! -f "go.mod" ]; then
        log_error "go.mod not found"
        exit 1
    fi

    log_info "Building HelixTrack Core..."

    if go build -o htCore main.go; then
        log_success "Build successful: htCore"
    else
        log_error "Build failed"
        exit 1
    fi
}

# Start test server
start_test_server() {
    local config=$1
    local port=$2
    local name=$3

    log_info "Starting $name server on port $port..."

    # Check if port is already in use
    if lsof -Pi :$port -sTCP:LISTEN -t >/dev/null 2>&1; then
        log_warning "Port $port is already in use"
        log_info "Attempting to stop existing process..."
        kill $(lsof -t -i:$port) 2>/dev/null || true
        sleep 2
    fi

    # Start server in background
    cd "$PROJECT_ROOT"

    if [ ! -f "./htCore" ]; then
        log_error "htCore binary not found. Building first..."
        build_app
    fi

    # Create log directory
    mkdir -p "$PROJECT_ROOT/test-logs"

    LOG_FILE="$PROJECT_ROOT/test-logs/${name}-server.log"

    if [ -f "$config" ]; then
        ./htCore --config="$config" > "$LOG_FILE" 2>&1 &
    else
        ./htCore > "$LOG_FILE" 2>&1 &
    fi

    SERVER_PID=$!

    log_info "Server PID: $SERVER_PID"

    # Wait for server to be ready
    log_info "Waiting for server to be ready..."

    WAIT_COUNT=0
    while [ $WAIT_COUNT -lt $SERVER_STARTUP_TIMEOUT ]; do
        if curl -s "http://localhost:$port/health" >/dev/null 2>&1; then
            log_success "$name server is ready"
            return 0
        fi

        sleep 1
        WAIT_COUNT=$((WAIT_COUNT + 1))
    done

    log_error "$name server failed to start within ${SERVER_STARTUP_TIMEOUT}s"
    log_info "Check log file: $LOG_FILE"
    cat "$LOG_FILE"
    return 1
}

# Stop test server
stop_test_server() {
    local pid=$1
    local name=$2

    if [ -n "$pid" ] && ps -p $pid > /dev/null 2>&1; then
        log_info "Stopping $name server (PID: $pid)..."
        kill $pid 2>/dev/null || true
        sleep 2

        # Force kill if still running
        if ps -p $pid > /dev/null 2>&1; then
            log_warning "Force killing $name server..."
            kill -9 $pid 2>/dev/null || true
        fi

        log_success "$name server stopped"
    fi
}

# Run AI QA tests
run_ai_qa_tests() {
    log_section "Running AI QA Tests"

    cd "$PROJECT_ROOT/tests/ai-qa"

    if [ ! -f "run_all_tests.py" ]; then
        log_error "AI QA test script not found: run_all_tests.py"
        exit 1
    fi

    log_info "Executing AI QA test suite..."

    # Set environment variables
    export SQLITE_API_URL="http://localhost:$SQLITE_PORT"
    export POSTGRES_API_URL="http://localhost:$POSTGRES_PORT"
    export TEST_TIMEOUT="$TEST_TIMEOUT"
    export VERBOSE="true"

    # Run tests
    if python3 run_all_tests.py; then
        log_success "AI QA tests passed"
        return 0
    else
        log_error "AI QA tests failed"
        return 1
    fi
}

# Run quick API tests
run_quick_api_tests() {
    log_section "Running Quick API Tests"

    cd "$PROJECT_ROOT"

    API_URL="http://localhost:$SQLITE_PORT"

    log_info "Testing /do endpoint with version action..."

    # Test version endpoint
    RESPONSE=$(curl -s -X POST "$API_URL/do" \
        -H "Content-Type: application/json" \
        -d '{"action":"version"}')

    if echo "$RESPONSE" | grep -q '"errorCode":-1'; then
        log_success "Version endpoint working"
        echo "Response: $RESPONSE"
    else
        log_error "Version endpoint failed"
        echo "Response: $RESPONSE"
        return 1
    fi

    log_info "Testing /do endpoint with health action..."

    # Test health endpoint
    RESPONSE=$(curl -s -X POST "$API_URL/do" \
        -H "Content-Type: application/json" \
        -d '{"action":"health"}')

    if echo "$RESPONSE" | grep -q '"errorCode":-1'; then
        log_success "Health endpoint working"
        echo "Response: $RESPONSE"
    else
        log_error "Health endpoint failed"
        echo "Response: $RESPONSE"
        return 1
    fi

    log_info "Testing JWT capability..."

    # Test JWT capable
    RESPONSE=$(curl -s -X POST "$API_URL/do" \
        -H "Content-Type: application/json" \
        -d '{"action":"jwtCapable"}')

    if echo "$RESPONSE" | grep -q '"errorCode":-1'; then
        log_success "JWT capable endpoint working"
        echo "Response: $RESPONSE"
    else
        log_error "JWT capable endpoint failed"
        echo "Response: $RESPONSE"
        return 1
    fi

    log_success "All quick API tests passed"
}

# Cleanup function
cleanup() {
    log_section "Cleaning Up"

    # Stop SQLite server
    if [ -n "$SQLITE_PID" ]; then
        stop_test_server "$SQLITE_PID" "SQLite"
    fi

    # Stop PostgreSQL server
    if [ -n "$POSTGRES_PID" ]; then
        stop_test_server "$POSTGRES_PID" "PostgreSQL"
    fi

    # Kill any remaining htCore processes
    pkill -f "htCore" 2>/dev/null || true

    log_success "Cleanup completed"
}

# Trap to ensure cleanup on exit
trap cleanup EXIT INT TERM

# Main execution
main() {
    log_section "HelixTrack Core - AI QA Test Runner"

    START_TIME=$(date +%s)

    check_python
    check_python_deps
    build_app

    EXIT_CODE=0

    # Start SQLite server
    SQLITE_CONFIG="$PROJECT_ROOT/../Configurations/dev.json"
    if [ ! -f "$SQLITE_CONFIG" ]; then
        log_warning "SQLite config not found, using defaults"
        SQLITE_CONFIG=""
    fi

    if start_test_server "$SQLITE_CONFIG" "$SQLITE_PORT" "SQLite"; then
        SQLITE_PID=$SERVER_PID

        # Run quick API tests
        run_quick_api_tests || EXIT_CODE=$?

        # Note: Full AI QA tests would require both SQLite and PostgreSQL servers
        # and the full test suite. For now, we run quick API tests.

        log_info "Full AI QA test suite requires:"
        log_info "  1. SQLite server (running)"
        log_info "  2. PostgreSQL server (not started)"
        log_info "  3. Both databases initialized"
        log_info ""
        log_info "To run full AI QA tests:"
        log_info "  cd tests/ai-qa && python3 run_all_tests.py"

    else
        log_error "Failed to start SQLite server"
        EXIT_CODE=1
    fi

    END_TIME=$(date +%s)
    DURATION=$((END_TIME - START_TIME))

    echo
    log_info "Total execution time: ${DURATION}s"
    echo

    if [ $EXIT_CODE -eq 0 ]; then
        log_success "🎉 All API tests passed! 🎉"
    else
        log_error "❌ Some tests failed"
    fi

    exit $EXIT_CODE
}

# Run main
main "$@"
