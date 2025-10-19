#!/bin/bash
###############################################################################
# HelixTrack Core - Production Start Script
#
# Usage:
#   ./scripts/start-production.sh [options]
#
# Options:
#   --with-monitoring    Start with Prometheus and Grafana
#   --with-extensions    Start with optional extensions
#   --detached          Run in detached mode (default)
#   --logs              Follow logs after starting
#   --build             Force rebuild before starting
#
# Examples:
#   ./scripts/start-production.sh
#   ./scripts/start-production.sh --with-monitoring --logs
#   ./scripts/start-production.sh --with-extensions --build
###############################################################################

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
COMPOSE_FILE="$PROJECT_DIR/docker-compose-production.yml"
ENV_FILE="$PROJECT_DIR/.env.production"

# Default options
DETACHED=true
FOLLOW_LOGS=false
WITH_MONITORING=false
WITH_EXTENSIONS=false
FORCE_BUILD=false

###############################################################################
# Functions
###############################################################################

print_header() {
    echo -e "${BLUE}"
    echo "========================================="
    echo "  HelixTrack Production Startup"
    echo "========================================="
    echo -e "${NC}"
}

print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠ $1${NC}"
}

print_info() {
    echo -e "${BLUE}ℹ $1${NC}"
}

check_prerequisites() {
    print_info "Checking prerequisites..."

    # Check Docker
    if ! command -v docker &> /dev/null; then
        print_error "Docker is not installed"
        exit 1
    fi
    print_success "Docker found"

    # Check Docker Compose
    if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
        print_error "Docker Compose is not installed"
        exit 1
    fi
    print_success "Docker Compose found"

    # Check if Docker daemon is running
    if ! docker info &> /dev/null; then
        print_error "Docker daemon is not running"
        exit 1
    fi
    print_success "Docker daemon running"
}

check_environment() {
    print_info "Checking environment configuration..."

    if [ ! -f "$ENV_FILE" ]; then
        print_warning "Environment file not found: $ENV_FILE"
        print_info "Creating default environment file..."

        cat > "$ENV_FILE" <<'EOF'
# HelixTrack Production Environment
# IMPORTANT: Change all passwords before production deployment!

# Build Configuration
BUILD_VERSION=1.0.0
BUILD_DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
BUILD_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Core Database
CORE_DB_NAME=helixtrack_core
CORE_DB_USER=helixtrack
CORE_DB_PASSWORD=helixtrack_secure_password_change_me
CORE_DB_PORT=5432

# Core Service
CORE_PORT=8080
CORE_METRICS_PORT=9090
CORE_REPLICAS=1

# Authentication Service
AUTH_DB_NAME=helixtrack_auth
AUTH_DB_USER=auth_user
AUTH_DB_PASSWORD=auth_secure_password_change_me
AUTH_PORT=8081
AUTH_REPLICAS=1
AUTH_SERVICE_URL=http://auth-service:8081

# Permissions Service
PERM_DB_NAME=helixtrack_perm
PERM_DB_USER=perm_user
PERM_DB_PASSWORD=perm_secure_password_change_me
PERM_PORT=8082
PERM_REPLICAS=1
PERM_SERVICE_URL=http://perm-service:8082

# Documents Extension
DOCS_DB_NAME=helixtrack_documents
DOCS_DB_USER=docs_user
DOCS_DB_PASSWORD=docs_secure_password_change_me
DOCS_PORT=8083

# Security
JWT_SECRET=your-jwt-secret-key-change-in-production-minimum-32-characters
ENCRYPTION_KEY=your-encryption-key-change-in-production-minimum-32-chars

# Monitoring
GRAFANA_USER=admin
GRAFANA_PASSWORD=admin_change_me
GRAFANA_DB_NAME=grafana
GRAFANA_DB_USER=grafana
GRAFANA_DB_PASSWORD=grafana_password_change_me
EOF

        print_success "Created default environment file"
        print_warning "IMPORTANT: Edit $ENV_FILE and change all passwords!"
        print_info "Press Enter to continue or Ctrl+C to abort..."
        read
    else
        print_success "Environment file found"
    fi

    # Source environment
    set -a
    source "$ENV_FILE"
    set +a
}

stop_existing() {
    print_info "Stopping existing containers..."

    if docker-compose -f "$COMPOSE_FILE" ps -q 2>/dev/null | grep -q .; then
        docker-compose -f "$COMPOSE_FILE" down
        print_success "Stopped existing containers"
    else
        print_info "No existing containers running"
    fi
}

build_images() {
    if [ "$FORCE_BUILD" = true ]; then
        print_info "Building Docker images..."
        docker-compose -f "$COMPOSE_FILE" build --no-cache
        print_success "Images built"
    fi
}

start_services() {
    print_info "Starting services..."

    # Build compose command
    CMD="docker-compose -f $COMPOSE_FILE"

    # Add profiles
    PROFILES=""
    if [ "$WITH_MONITORING" = true ]; then
        PROFILES="$PROFILES --profile monitoring"
    fi
    if [ "$WITH_EXTENSIONS" = true ]; then
        PROFILES="$PROFILES --profile extensions"
    fi

    # Start command
    START_CMD="$CMD $PROFILES up"
    if [ "$DETACHED" = true ]; then
        START_CMD="$START_CMD -d"
    fi

    # Execute
    eval $START_CMD

    print_success "Services started"
}

wait_for_health() {
    print_info "Waiting for services to be healthy..."

    MAX_WAIT=120
    WAITED=0

    while [ $WAITED -lt $MAX_WAIT ]; do
        if docker-compose -f "$COMPOSE_FILE" ps | grep -q "unhealthy"; then
            echo -n "."
            sleep 2
            WAITED=$((WAITED + 2))
        else
            echo ""
            print_success "All services are healthy"
            return 0
        fi
    done

    echo ""
    print_warning "Some services may not be healthy yet"
}

show_status() {
    print_info "Service Status:"
    echo ""
    docker-compose -f "$COMPOSE_FILE" ps
    echo ""

    print_info "Network Information:"
    docker network inspect helixtrack-network --format '{{range .Containers}}{{.Name}}: {{.IPv4Address}}{{"\n"}}{{end}}' || true
    echo ""

    print_info "Accessing Services:"
    echo "  Core API:       http://localhost:${CORE_PORT:-8080}"
    echo "  Auth Service:   http://localhost:${AUTH_PORT:-8081}"
    echo "  Perm Service:   http://localhost:${PERM_PORT:-8082}"
    echo "  Service Registry: http://localhost:8500"
    if [ "$WITH_MONITORING" = true ]; then
        echo "  Prometheus:     http://localhost:9091"
        echo "  Grafana:        http://localhost:3000"
    fi
    echo ""
}

follow_logs() {
    if [ "$FOLLOW_LOGS" = true ]; then
        print_info "Following logs (Ctrl+C to exit)..."
        docker-compose -f "$COMPOSE_FILE" logs -f
    fi
}

###############################################################################
# Parse Arguments
###############################################################################

while [[ $# -gt 0 ]]; do
    case $1 in
        --with-monitoring)
            WITH_MONITORING=true
            shift
            ;;
        --with-extensions)
            WITH_EXTENSIONS=true
            shift
            ;;
        --detached)
            DETACHED=true
            shift
            ;;
        --logs)
            FOLLOW_LOGS=true
            shift
            ;;
        --build)
            FORCE_BUILD=true
            shift
            ;;
        --help)
            grep "^#" "$0" | grep -v "#!/bin/bash" | sed 's/^# //' | sed 's/^#//'
            exit 0
            ;;
        *)
            print_error "Unknown option: $1"
            echo "Use --help for usage information"
            exit 1
            ;;
    esac
done

###############################################################################
# Main Execution
###############################################################################

print_header

# Execute startup sequence
check_prerequisites
check_environment
stop_existing
build_images
start_services
wait_for_health
show_status
follow_logs

print_success "HelixTrack started successfully!"
echo ""
print_info "To view logs: docker-compose -f $COMPOSE_FILE logs -f"
print_info "To stop:      ./scripts/stop-production.sh"
echo ""
