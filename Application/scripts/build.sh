#!/bin/bash
#
# HelixTrack Core - Build Script
# Builds all modules with verification
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

# Build configuration
BUILD_OUTPUT="htCore"
BUILD_FLAGS="-v"
RELEASE_FLAGS="-ldflags='-s -w'"

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

# Check if Go is installed
check_go() {
    if ! command -v go &> /dev/null; then
        log_error "Go is not installed"
        log_info "Run: ./scripts/setup-environment.sh"
        exit 1
    fi

    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    log_info "Using Go version: ${GO_VERSION}"
}

# Clean previous builds
clean_build() {
    log_info "Cleaning previous builds..."

    cd "$PROJECT_ROOT"

    # Remove binary
    if [ -f "$BUILD_OUTPUT" ]; then
        rm -f "$BUILD_OUTPUT"
        log_success "Removed old binary"
    fi

    # Remove test binaries
    find . -name "*.test" -type f -delete

    log_success "Build artifacts cleaned"
}

# Verify dependencies
verify_dependencies() {
    log_section "Verifying Dependencies"

    cd "$PROJECT_ROOT"

    log_info "Downloading dependencies..."
    go mod download

    log_info "Verifying module dependencies..."
    if go mod verify; then
        log_success "All dependencies verified"
    else
        log_error "Dependency verification failed"
        exit 1
    fi

    log_info "Tidying module dependencies..."
    go mod tidy

    log_success "Dependencies are up to date"
}

# Run pre-build checks
pre_build_checks() {
    log_section "Running Pre-Build Checks"

    cd "$PROJECT_ROOT"

    # Check for go vet issues
    log_info "Running go vet..."
    if go vet ./...; then
        log_success "go vet passed"
    else
        log_error "go vet found issues"
        exit 1
    fi

    # Check formatting
    log_info "Checking code formatting..."
    UNFORMATTED=$(gofmt -l . | grep -v vendor || true)

    if [ -z "$UNFORMATTED" ]; then
        log_success "All code is properly formatted"
    else
        log_warning "The following files are not formatted:"
        echo "$UNFORMATTED"
        log_info "Run 'go fmt ./...' to fix formatting"

        read -p "Continue anyway? (y/N) " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            exit 1
        fi
    fi
}

# Build the application
build_app() {
    log_section "Building Application"

    cd "$PROJECT_ROOT"

    local build_type=${1:-"debug"}

    if [ "$build_type" = "release" ]; then
        log_info "Building release binary..."
        log_info "Build flags: $RELEASE_FLAGS"

        # Build with optimizations
        if go build $RELEASE_FLAGS -o "$BUILD_OUTPUT" main.go; then
            log_success "Release build successful"
        else
            log_error "Release build failed"
            exit 1
        fi
    else
        log_info "Building debug binary..."
        log_info "Build flags: $BUILD_FLAGS"

        # Build with debug info
        if go build $BUILD_FLAGS -o "$BUILD_OUTPUT" main.go; then
            log_success "Debug build successful"
        else
            log_error "Debug build failed"
            exit 1
        fi
    fi

    # Make executable
    chmod +x "$BUILD_OUTPUT"
}

# Verify build
verify_build() {
    log_section "Verifying Build"

    cd "$PROJECT_ROOT"

    # Check if binary exists
    if [ ! -f "$BUILD_OUTPUT" ]; then
        log_error "Binary not found: $BUILD_OUTPUT"
        exit 1
    fi

    log_success "Binary exists: $BUILD_OUTPUT"

    # Check if executable
    if [ ! -x "$BUILD_OUTPUT" ]; then
        log_error "Binary is not executable"
        exit 1
    fi

    log_success "Binary is executable"

    # Get binary size
    BINARY_SIZE=$(du -h "$BUILD_OUTPUT" | cut -f1)
    log_info "Binary size: $BINARY_SIZE"

    # Test version command
    log_info "Testing version command..."
    if ./"$BUILD_OUTPUT" --version 2>&1 | grep -q "HelixTrack Core"; then
        log_success "Version command works"
        ./"$BUILD_OUTPUT" --version
    else
        log_warning "Version command did not return expected output"
    fi
}

# Run quick smoke test
smoke_test() {
    log_section "Running Smoke Test"

    cd "$PROJECT_ROOT"

    log_info "Starting server for smoke test..."

    # Start server in background
    ./"$BUILD_OUTPUT" > smoke-test.log 2>&1 &
    SERVER_PID=$!

    log_info "Server PID: $SERVER_PID"

    # Wait for server to be ready
    WAIT_COUNT=0
    TIMEOUT=30

    while [ $WAIT_COUNT -lt $TIMEOUT ]; do
        if curl -s "http://localhost:8080/health" >/dev/null 2>&1; then
            log_success "Server started successfully"
            break
        fi

        sleep 1
        WAIT_COUNT=$((WAIT_COUNT + 1))
    done

    if [ $WAIT_COUNT -ge $TIMEOUT ]; then
        log_error "Server failed to start within ${TIMEOUT}s"
        log_info "Log output:"
        cat smoke-test.log
        kill $SERVER_PID 2>/dev/null || true
        exit 1
    fi

    # Test health endpoint
    log_info "Testing health endpoint..."
    RESPONSE=$(curl -s -X POST "http://localhost:8080/do" \
        -H "Content-Type: application/json" \
        -d '{"action":"health"}')

    if echo "$RESPONSE" | grep -q '"errorCode":-1'; then
        log_success "Health endpoint working"
    else
        log_error "Health endpoint failed"
        echo "Response: $RESPONSE"
        kill $SERVER_PID 2>/dev/null || true
        exit 1
    fi

    # Stop server
    log_info "Stopping server..."
    kill $SERVER_PID 2>/dev/null || true
    sleep 1

    # Force kill if still running
    if ps -p $SERVER_PID > /dev/null 2>&1; then
        kill -9 $SERVER_PID 2>/dev/null || true
    fi

    log_success "Smoke test passed"

    # Clean up log
    rm -f smoke-test.log
}

# Generate build info
generate_build_info() {
    log_section "Generating Build Info"

    cd "$PROJECT_ROOT"

    BUILD_INFO_FILE="BUILD_INFO.txt"

    cat > "$BUILD_INFO_FILE" << EOF
HelixTrack Core - Build Information
====================================

Build Date:     $(date '+%Y-%m-%d %H:%M:%S')
Go Version:     $(go version | awk '{print $3}')
OS:             $(uname -s)
Architecture:   $(uname -m)
Binary:         $BUILD_OUTPUT
Binary Size:    $(du -h "$BUILD_OUTPUT" | cut -f1)
Git Commit:     $(git rev-parse --short HEAD 2>/dev/null || echo "N/A")
Git Branch:     $(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "N/A")

Dependencies:
$(go list -m all)

Build Command:  go build -o $BUILD_OUTPUT main.go

====================================
EOF

    log_success "Build info saved: $BUILD_INFO_FILE"
}

# Print build summary
print_summary() {
    log_section "Build Summary"

    cd "$PROJECT_ROOT"

    echo -e "${CYAN}Binary:${NC}          $BUILD_OUTPUT"
    echo -e "${CYAN}Size:${NC}            $(du -h "$BUILD_OUTPUT" | cut -f1)"
    echo -e "${CYAN}Location:${NC}        $PROJECT_ROOT/$BUILD_OUTPUT"
    echo

    log_info "To run the application:"
    echo "  ./$BUILD_OUTPUT"
    echo

    log_info "To run with custom config:"
    echo "  ./$BUILD_OUTPUT --config=path/to/config.json"
    echo

    log_info "To show version:"
    echo "  ./$BUILD_OUTPUT --version"
    echo

    log_success "Build completed successfully!"
}

# Main execution
main() {
    log_section "HelixTrack Core - Build Script"

    START_TIME=$(date +%s)

    # Parse arguments
    BUILD_TYPE="debug"
    RUN_TESTS=false
    RUN_SMOKE_TEST=false
    SKIP_CHECKS=false

    while [[ $# -gt 0 ]]; do
        case $1 in
            --release)
                BUILD_TYPE="release"
                shift
                ;;
            --with-tests)
                RUN_TESTS=true
                shift
                ;;
            --smoke-test)
                RUN_SMOKE_TEST=true
                shift
                ;;
            --skip-checks)
                SKIP_CHECKS=true
                shift
                ;;
            *)
                log_error "Unknown option: $1"
                echo "Usage: $0 [--release] [--with-tests] [--smoke-test] [--skip-checks]"
                exit 1
                ;;
        esac
    done

    check_go
    clean_build
    verify_dependencies

    if [ "$SKIP_CHECKS" = false ]; then
        pre_build_checks
    fi

    build_app "$BUILD_TYPE"
    verify_build

    if [ "$RUN_SMOKE_TEST" = true ]; then
        smoke_test
    fi

    if [ "$RUN_TESTS" = true ]; then
        log_section "Running Tests"
        bash "$SCRIPT_DIR/run-all-tests.sh"
    fi

    generate_build_info
    print_summary

    END_TIME=$(date +%s)
    DURATION=$((END_TIME - START_TIME))

    log_info "Total build time: ${DURATION}s"

    exit 0
}

# Run main
main "$@"
