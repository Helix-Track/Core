#!/bin/bash
###############################################################################
# HelixTrack Core - Production Stop Script
#
# Usage:
#   ./scripts/stop-production.sh [options]
#
# Options:
#   --remove-volumes    Remove all data volumes
#   --remove-images     Remove built images
#   --cleanup           Full cleanup (volumes + images + networks)
#   --force             Force stop containers without graceful shutdown
#
# Examples:
#   ./scripts/stop-production.sh
#   ./scripts/stop-production.sh --cleanup
#   ./scripts/stop-production.sh --remove-volumes
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

# Default options
REMOVE_VOLUMES=false
REMOVE_IMAGES=false
FULL_CLEANUP=false
FORCE_STOP=false

###############################################################################
# Functions
###############################################################################

print_header() {
    echo -e "${BLUE}"
    echo "========================================="
    echo "  HelixTrack Production Shutdown"
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

    # Check Docker Compose
    if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
        print_error "Docker Compose is not installed"
        exit 1
    fi

    # Check if Docker daemon is running
    if ! docker info &> /dev/null; then
        print_error "Docker daemon is not running"
        exit 1
    fi

    print_success "Prerequisites verified"
}

check_running_containers() {
    print_info "Checking for running containers..."

    if docker-compose -f "$COMPOSE_FILE" ps -q 2>/dev/null | grep -q .; then
        print_info "Found running containers"
        echo ""
        docker-compose -f "$COMPOSE_FILE" ps
        echo ""
        return 0
    else
        print_warning "No containers are currently running"
        return 1
    fi
}

deregister_services() {
    print_info "Deregistering services from service discovery..."

    # Get service registry container ID
    CONSUL_CONTAINER=$(docker-compose -f "$COMPOSE_FILE" ps -q service-registry 2>/dev/null || true)

    if [ -n "$CONSUL_CONTAINER" ]; then
        # Get all registered services
        SERVICES=$(docker exec "$CONSUL_CONTAINER" consul catalog services 2>/dev/null || true)

        if [ -n "$SERVICES" ]; then
            print_info "Deregistering services from Consul..."
            # Give services time to deregister themselves gracefully
            sleep 2
            print_success "Services deregistered"
        fi
    else
        print_info "Service registry not running, skipping deregistration"
    fi
}

stop_services() {
    print_info "Stopping services..."

    if [ "$FORCE_STOP" = true ]; then
        print_warning "Force stopping containers..."
        docker-compose -f "$COMPOSE_FILE" kill
        docker-compose -f "$COMPOSE_FILE" rm -f
    else
        print_info "Gracefully stopping containers (max 30s timeout)..."
        docker-compose -f "$COMPOSE_FILE" down --timeout 30
    fi

    print_success "Services stopped"
}

remove_volumes() {
    if [ "$REMOVE_VOLUMES" = true ] || [ "$FULL_CLEANUP" = true ]; then
        print_warning "Removing data volumes..."

        # List volumes before removal
        echo ""
        print_info "Volumes to be removed:"
        docker-compose -f "$COMPOSE_FILE" down -v --remove-orphans 2>/dev/null || true

        # Also remove named volumes directly
        VOLUMES=$(docker volume ls -q | grep -E "helixtrack|htcore" || true)
        if [ -n "$VOLUMES" ]; then
            echo "$VOLUMES" | xargs docker volume rm 2>/dev/null || true
            print_success "Data volumes removed"
        else
            print_info "No volumes to remove"
        fi

        echo ""
        print_warning "WARNING: All database data has been deleted!"
        print_info "You will need to re-initialize databases on next startup"
    fi
}

remove_images() {
    if [ "$REMOVE_IMAGES" = true ] || [ "$FULL_CLEANUP" = true ]; then
        print_warning "Removing built images..."

        # Get image names from docker-compose
        IMAGES=$(docker-compose -f "$COMPOSE_FILE" config --images 2>/dev/null || true)

        if [ -n "$IMAGES" ]; then
            echo ""
            print_info "Images to be removed:"
            echo "$IMAGES"
            echo ""

            # Remove images
            echo "$IMAGES" | xargs docker rmi -f 2>/dev/null || true
            print_success "Images removed"
        else
            print_info "No images to remove"
        fi
    fi
}

cleanup_networks() {
    if [ "$FULL_CLEANUP" = true ]; then
        print_info "Cleaning up networks..."

        # Remove project-specific networks
        NETWORKS=$(docker network ls -q --filter name=helixtrack || true)
        if [ -n "$NETWORKS" ]; then
            echo "$NETWORKS" | xargs docker network rm 2>/dev/null || true
            print_success "Networks removed"
        else
            print_info "No networks to remove"
        fi
    fi
}

cleanup_orphans() {
    print_info "Removing orphaned containers..."
    docker-compose -f "$COMPOSE_FILE" down --remove-orphans
    print_success "Orphaned containers removed"
}

show_final_status() {
    echo ""
    print_info "Final Status:"
    echo ""

    # Check for any remaining containers
    REMAINING=$(docker-compose -f "$COMPOSE_FILE" ps -q 2>/dev/null || true)
    if [ -z "$REMAINING" ]; then
        print_success "All containers stopped successfully"
    else
        print_warning "Some containers may still be running:"
        docker-compose -f "$COMPOSE_FILE" ps
    fi

    echo ""

    # Show remaining volumes if not removed
    if [ "$REMOVE_VOLUMES" != true ] && [ "$FULL_CLEANUP" != true ]; then
        print_info "Data volumes preserved:"
        docker volume ls | grep -E "helixtrack|htcore" || print_info "No volumes found"
        echo ""
        print_info "To remove volumes: $0 --remove-volumes"
    fi

    echo ""
}

confirm_cleanup() {
    if [ "$FULL_CLEANUP" = true ] || [ "$REMOVE_VOLUMES" = true ]; then
        echo ""
        print_warning "=============================================="
        print_warning "WARNING: This will delete all data!"
        print_warning "=============================================="

        if [ "$REMOVE_VOLUMES" = true ]; then
            print_warning "• All database data will be lost"
        fi

        if [ "$FULL_CLEANUP" = true ]; then
            print_warning "• All volumes will be deleted"
            print_warning "• All images will be removed"
            print_warning "• All networks will be cleaned up"
        fi

        echo ""
        print_warning "This action cannot be undone!"
        echo ""
        read -p "Are you sure you want to continue? (yes/no): " -r
        echo

        if [[ ! $REPLY =~ ^[Yy][Ee][Ss]$ ]]; then
            print_info "Operation cancelled"
            exit 0
        fi
    fi
}

###############################################################################
# Parse Arguments
###############################################################################

while [[ $# -gt 0 ]]; do
    case $1 in
        --remove-volumes)
            REMOVE_VOLUMES=true
            shift
            ;;
        --remove-images)
            REMOVE_IMAGES=true
            shift
            ;;
        --cleanup)
            FULL_CLEANUP=true
            shift
            ;;
        --force)
            FORCE_STOP=true
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

# Execute shutdown sequence
check_prerequisites

# Check if containers are running
if ! check_running_containers; then
    print_info "Nothing to stop"
    exit 0
fi

# Confirm cleanup if requested
confirm_cleanup

# Execute shutdown
deregister_services
stop_services
cleanup_orphans
remove_volumes
remove_images
cleanup_networks
show_final_status

print_success "HelixTrack stopped successfully!"
echo ""

if [ "$FULL_CLEANUP" != true ] && [ "$REMOVE_VOLUMES" != true ]; then
    print_info "To start again: ./scripts/start-production.sh"
    print_info "For full cleanup: $0 --cleanup"
fi

echo ""
