#!/bin/bash
#
# HelixTrack Core - Environment Setup Script
# Installs all dependencies required for building and testing
#

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

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

# Check if running as root
check_root() {
    if [ "$EUID" -eq 0 ]; then
        log_warning "Running as root. This is not recommended for development."
        read -p "Continue anyway? (y/N) " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            exit 1
        fi
    fi
}

# Check OS
check_os() {
    log_info "Detecting operating system..."

    if [ -f /etc/os-release ]; then
        . /etc/os-release
        OS=$ID
        VERSION=$VERSION_ID
        log_info "Detected: $PRETTY_NAME"
    else
        log_error "Cannot detect OS. /etc/os-release not found."
        exit 1
    fi
}

# Install Go
install_go() {
    GO_VERSION="1.22.0"
    GO_TAR="go${GO_VERSION}.linux-amd64.tar.gz"
    GO_URL="https://go.dev/dl/${GO_TAR}"

    log_info "Checking Go installation..."

    if command -v go &> /dev/null; then
        CURRENT_GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
        log_success "Go ${CURRENT_GO_VERSION} is already installed"

        # Check if version is sufficient
        if [ "$(printf '%s\n' "1.22" "$CURRENT_GO_VERSION" | sort -V | head -n1)" = "1.22" ]; then
            log_success "Go version is sufficient (>= 1.22)"
            return 0
        else
            log_warning "Go version is too old (< 1.22). Upgrading..."
        fi
    fi

    log_info "Installing Go ${GO_VERSION}..."

    # Download Go
    TEMP_DIR=$(mktemp -d)
    cd "$TEMP_DIR"

    log_info "Downloading Go from ${GO_URL}..."
    if ! wget -q --show-progress "$GO_URL"; then
        log_error "Failed to download Go"
        rm -rf "$TEMP_DIR"
        exit 1
    fi

    # Remove old Go installation if exists
    if [ -d "/usr/local/go" ]; then
        log_info "Removing old Go installation..."
        sudo rm -rf /usr/local/go
    fi

    # Extract Go
    log_info "Extracting Go..."
    sudo tar -C /usr/local -xzf "$GO_TAR"

    # Add to PATH
    if ! grep -q "/usr/local/go/bin" ~/.bashrc; then
        log_info "Adding Go to PATH in ~/.bashrc..."
        echo "" >> ~/.bashrc
        echo "# Go installation" >> ~/.bashrc
        echo "export PATH=\$PATH:/usr/local/go/bin" >> ~/.bashrc
        echo "export GOPATH=\$HOME/go" >> ~/.bashrc
        echo "export PATH=\$PATH:\$GOPATH/bin" >> ~/.bashrc
    fi

    # Add to current session
    export PATH=$PATH:/usr/local/go/bin
    export GOPATH=$HOME/go
    export PATH=$PATH:$GOPATH/bin

    # Cleanup
    cd -
    rm -rf "$TEMP_DIR"

    # Verify installation
    if command -v go &> /dev/null; then
        GO_VERSION_INSTALLED=$(go version)
        log_success "Go installed successfully: ${GO_VERSION_INSTALLED}"
    else
        log_error "Go installation failed"
        exit 1
    fi
}

# Install SQLite
install_sqlite() {
    log_info "Checking SQLite installation..."

    if command -v sqlite3 &> /dev/null; then
        SQLITE_VERSION=$(sqlite3 --version | awk '{print $1}')
        log_success "SQLite ${SQLITE_VERSION} is already installed"
        return 0
    fi

    log_info "Installing SQLite..."

    case "$OS" in
        ubuntu|debian)
            sudo apt-get update
            sudo apt-get install -y sqlite3 libsqlite3-dev
            ;;
        fedora|rhel|centos)
            sudo dnf install -y sqlite sqlite-devel
            ;;
        arch)
            sudo pacman -S --noconfirm sqlite
            ;;
        *)
            log_warning "Unsupported OS: $OS. Please install SQLite manually."
            return 1
            ;;
    esac

    if command -v sqlite3 &> /dev/null; then
        log_success "SQLite installed successfully"
    else
        log_error "SQLite installation failed"
        exit 1
    fi
}

# Install PostgreSQL client (optional)
install_postgresql_client() {
    log_info "Checking PostgreSQL client installation..."

    if command -v psql &> /dev/null; then
        PSQL_VERSION=$(psql --version | awk '{print $3}')
        log_success "PostgreSQL client ${PSQL_VERSION} is already installed"
        return 0
    fi

    log_info "PostgreSQL client not found. Install it? (optional, for production)"
    read -p "Install PostgreSQL client? (y/N) " -n 1 -r
    echo

    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        log_info "Skipping PostgreSQL client installation"
        return 0
    fi

    case "$OS" in
        ubuntu|debian)
            sudo apt-get update
            sudo apt-get install -y postgresql-client
            ;;
        fedora|rhel|centos)
            sudo dnf install -y postgresql
            ;;
        arch)
            sudo pacman -S --noconfirm postgresql-libs
            ;;
        *)
            log_warning "Unsupported OS: $OS. Please install PostgreSQL client manually."
            return 1
            ;;
    esac

    if command -v psql &> /dev/null; then
        log_success "PostgreSQL client installed successfully"
    else
        log_warning "PostgreSQL client installation failed (continuing without it)"
    fi
}

# Install Python dependencies (for AI QA tests)
install_python_deps() {
    log_info "Checking Python installation..."

    if ! command -v python3 &> /dev/null; then
        log_error "Python 3 is not installed. Please install Python 3.8+"
        exit 1
    fi

    PYTHON_VERSION=$(python3 --version | awk '{print $2}')
    log_success "Python ${PYTHON_VERSION} is installed"

    log_info "Installing Python dependencies for AI QA tests..."

    # Check if pip is installed
    if ! command -v pip3 &> /dev/null; then
        log_info "Installing pip..."
        case "$OS" in
            ubuntu|debian)
                sudo apt-get update
                sudo apt-get install -y python3-pip
                ;;
            fedora|rhel|centos)
                sudo dnf install -y python3-pip
                ;;
            arch)
                sudo pacman -S --noconfirm python-pip
                ;;
        esac
    fi

    # Install required Python packages
    log_info "Installing requests and colorama..."
    pip3 install --user requests colorama

    log_success "Python dependencies installed"
}

# Install Git (if not present)
install_git() {
    log_info "Checking Git installation..."

    if command -v git &> /dev/null; then
        GIT_VERSION=$(git --version | awk '{print $3}')
        log_success "Git ${GIT_VERSION} is already installed"
        return 0
    fi

    log_info "Installing Git..."

    case "$OS" in
        ubuntu|debian)
            sudo apt-get update
            sudo apt-get install -y git
            ;;
        fedora|rhel|centos)
            sudo dnf install -y git
            ;;
        arch)
            sudo pacman -S --noconfirm git
            ;;
    esac

    if command -v git &> /dev/null; then
        log_success "Git installed successfully"
    else
        log_error "Git installation failed"
        exit 1
    fi
}

# Install build essentials
install_build_tools() {
    log_info "Checking build tools..."

    case "$OS" in
        ubuntu|debian)
            if ! dpkg -l | grep -q build-essential; then
                log_info "Installing build-essential..."
                sudo apt-get update
                sudo apt-get install -y build-essential
            else
                log_success "build-essential is already installed"
            fi
            ;;
        fedora|rhel|centos)
            if ! rpm -q gcc &> /dev/null; then
                log_info "Installing development tools..."
                sudo dnf groupinstall -y "Development Tools"
            else
                log_success "Development tools are already installed"
            fi
            ;;
        arch)
            if ! pacman -Q base-devel &> /dev/null; then
                log_info "Installing base-devel..."
                sudo pacman -S --noconfirm base-devel
            else
                log_success "base-devel is already installed"
            fi
            ;;
    esac
}

# Download Go dependencies
download_go_deps() {
    log_info "Downloading Go dependencies..."

    cd "$PROJECT_ROOT"

    if [ ! -f "go.mod" ]; then
        log_error "go.mod not found in $PROJECT_ROOT"
        exit 1
    fi

    go mod download
    log_success "Go dependencies downloaded"
}

# Initialize database
init_database() {
    log_info "Initializing SQLite database..."

    cd "$PROJECT_ROOT/.."

    DB_DIR="$PROJECT_ROOT/../Database"
    if [ ! -d "$DB_DIR" ]; then
        log_error "Database directory not found: $DB_DIR"
        exit 1
    fi

    # Check if import script exists
    IMPORT_SCRIPT="$PROJECT_ROOT/../Run/Db/import_All_Definitions_to_Sqlite.sh"
    if [ -f "$IMPORT_SCRIPT" ]; then
        log_info "Running database import script..."
        bash "$IMPORT_SCRIPT"
        log_success "Database initialized"
    else
        log_warning "Database import script not found. Skipping database initialization."
        log_warning "You may need to initialize the database manually."
    fi
}

# Main installation
main() {
    log_info "HelixTrack Core - Environment Setup"
    log_info "===================================="
    echo

    check_root
    check_os

    log_info "This script will install:"
    log_info "  - Go 1.22+"
    log_info "  - SQLite 3"
    log_info "  - PostgreSQL client (optional)"
    log_info "  - Python dependencies"
    log_info "  - Build tools"
    log_info "  - Git"
    echo

    read -p "Continue with installation? (Y/n) " -n 1 -r
    echo

    if [[ $REPLY =~ ^[Nn]$ ]]; then
        log_info "Installation cancelled"
        exit 0
    fi

    echo
    log_info "Starting installation..."
    echo

    install_git
    install_build_tools
    install_go
    install_sqlite
    install_postgresql_client
    install_python_deps
    download_go_deps
    init_database

    echo
    log_success "===================================="
    log_success "Environment setup completed!"
    log_success "===================================="
    echo
    log_info "Next steps:"
    log_info "  1. Source your bashrc to update PATH: source ~/.bashrc"
    log_info "  2. Run tests: ./scripts/run-all-tests.sh"
    log_info "  3. Build application: ./scripts/build.sh"
    log_info "  4. Run application: ./htCore"
    echo
    log_info "For a new terminal session, Go will be available automatically."
    echo
}

# Run main
main "$@"
