#!/bin/bash
# Cleanup script for pgEdge MCP Server Test Suite
# This script removes all test artifacts, services, and configurations
# Works with both local and container execution modes

set -e  # Exit on error

echo "========================================================================="
echo "pgEdge MCP Server Test Suite Cleanup"
echo "========================================================================="
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;36m'
NC='\033[0m' # No Color

# Function to print status
print_status() {
    echo -e "${GREEN}✓${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

print_info() {
    echo -e "${BLUE}ℹ${NC} $1"
}

# Check if running with sudo/root for local mode
if [ "$EUID" -eq 0 ]; then
    SUDO=""
else
    SUDO="sudo"
fi

# Detect execution mode
EXECUTION_MODE="unknown"
if [ -f "config/container.yaml" ]; then
    EXECUTION_MODE="container"
    print_info "Container mode detected"
elif [ -f "config/local.yaml" ]; then
    EXECUTION_MODE="local"
    print_info "Local mode detected"
fi

echo "This script will clean up:"
echo "  • PostgreSQL server and data"
echo "  • MCP server service and packages"
echo "  • pgEdge packages and repositories"
echo "  • Test configuration files"
echo "  • Test result logs"
if [ "$EXECUTION_MODE" = "container" ]; then
    echo "  • Docker containers used for testing"
fi
echo ""
read -p "Continue? (y/N) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Cleanup cancelled."
    exit 0
fi

echo ""
echo "========================================================================="
echo "Step 1: Stopping Services"
echo "========================================================================="

# Stop MCP server
if systemctl is-active --quiet pgedge-postgres-mcp.service 2>/dev/null; then
    echo "Stopping MCP server service..."
    $SUDO systemctl stop pgedge-postgres-mcp.service || true
    $SUDO systemctl disable pgedge-postgres-mcp.service || true
    print_status "MCP server stopped"
else
    print_warning "MCP server service not running"
fi

# Stop PostgreSQL (check multiple versions)
for pg_service in postgresql-18 postgresql-17 postgresql-16 postgresql; do
    if systemctl is-active --quiet $pg_service 2>/dev/null; then
        echo "Stopping $pg_service..."
        $SUDO systemctl stop $pg_service || true
        $SUDO systemctl disable $pg_service || true
        print_status "$pg_service stopped"
    fi
done

# For Debian/Ubuntu systems using pg_ctlcluster
if command -v pg_ctlcluster &> /dev/null; then
    for version in 18 17 16; do
        if [ -d "/etc/postgresql/$version/main" ]; then
            echo "Stopping PostgreSQL $version using pg_ctlcluster..."
            $SUDO pg_ctlcluster --skip-systemctl-redirect $version main stop 2>/dev/null || true
            print_status "PostgreSQL $version stopped"
        fi
    done
fi

# Kill any remaining MCP server processes
if pgrep -f "pgedge-postgres-mcp" > /dev/null 2>&1; then
    echo "Killing remaining MCP server processes..."
    $SUDO pkill -f "pgedge-postgres-mcp" || true
    print_status "MCP server processes terminated"
fi

# Kill any remaining PostgreSQL processes
if pgrep -f "postgres" > /dev/null 2>&1; then
    echo "Killing remaining PostgreSQL processes..."
    $SUDO pkill -f "postgres" || true
    print_status "PostgreSQL processes terminated"
fi

echo ""
echo "========================================================================="
echo "Step 2: Removing Packages"
echo "========================================================================="

# Detect package manager
if command -v apt-get &> /dev/null; then
    PKG_MGR="apt"
    print_info "Detected Debian/Ubuntu system"

    # Remove pgEdge packages
    echo "Removing pgEdge packages..."
    $SUDO apt-get remove -y pgedge-postgres-mcp pgedge-nla-cli pgedge-nla-web pgedge-nla-kb-builder pgedge-postgres-mcp-kb 2>/dev/null || true

    # Remove PostgreSQL
    echo "Removing PostgreSQL packages..."
    $SUDO apt-get remove -y 'postgresql-*' 'pgedge-postgresql-*' 2>/dev/null || true

    # Remove pgedge-release package
    echo "Removing pgedge-release package..."
    $SUDO apt-get remove -y pgedge-release 2>/dev/null || true

    $SUDO apt-get autoremove -y || true

elif command -v dnf &> /dev/null; then
    PKG_MGR="dnf"
    print_info "Detected RHEL/Rocky/Alma system"

    # Remove pgEdge packages
    echo "Removing pgEdge packages..."
    $SUDO dnf remove -y pgedge-postgres-mcp pgedge-nla-cli pgedge-nla-web pgedge-nla-kb-builder pgedge-postgres-mcp-kb 2>/dev/null || true

    # Remove PostgreSQL
    echo "Removing PostgreSQL packages..."
    $SUDO dnf remove -y 'postgresql*' 'pgedge-postgresql*' 2>/dev/null || true

    # Remove pgedge-release package
    echo "Removing pgedge-release package..."
    $SUDO dnf remove -y pgedge-release 2>/dev/null || true

    $SUDO dnf autoremove -y || true

elif command -v yum &> /dev/null; then
    PKG_MGR="yum"
    print_info "Detected older RHEL system"

    # Remove pgEdge packages
    echo "Removing pgEdge packages..."
    $SUDO yum remove -y pgedge-postgres-mcp pgedge-nla-cli pgedge-nla-web pgedge-nla-kb-builder pgedge-postgres-mcp-kb 2>/dev/null || true

    # Remove PostgreSQL
    echo "Removing PostgreSQL packages..."
    $SUDO yum remove -y 'postgresql*' 'pgedge-postgresql*' 2>/dev/null || true

    # Remove pgedge-release package
    echo "Removing pgedge-release package..."
    $SUDO yum remove -y pgedge-release 2>/dev/null || true

else
    print_warning "Unknown package manager, skipping package removal"
fi

print_status "Packages removed"

echo ""
echo "========================================================================="
echo "Step 3: Removing Data and Configuration"
echo "========================================================================="

# Remove PostgreSQL data directories
echo "Removing PostgreSQL data directories..."
$SUDO rm -rf /var/lib/postgresql/* || true
$SUDO rm -rf /var/lib/pgsql/* || true
$SUDO rm -rf /etc/postgresql || true
print_status "PostgreSQL data removed"

# Remove pgEdge directories
echo "Removing pgEdge directories..."
$SUDO rm -rf /etc/pgedge || true
$SUDO rm -rf /usr/share/pgedge || true
$SUDO rm -rf /var/log/pgedge || true
$SUDO rm -rf /var/lib/pgedge || true
rm -rf ~/.pgedge || true
print_status "pgEdge directories removed"

# Remove test files
echo "Removing test files and logs..."
rm -rf test-results/* || true
rm -rf /tmp/test_kb_database || true
rm -rf /tmp/test-mcp-server-config.yaml || true
rm -rf /tmp/mcp-server-test.log || true
rm -rf /tmp/pgedge-release.deb || true
print_status "Test files removed"

# Remove repository configurations
echo "Removing repository configurations..."
$SUDO rm -f /etc/yum.repos.d/pgedge*.repo || true
$SUDO rm -f /etc/apt/sources.list.d/pgedge*.list || true
$SUDO rm -f /etc/apt/trusted.gpg.d/pgedge*.gpg || true
print_status "Repository configurations removed"

echo ""
echo "========================================================================="
echo "Step 4: Cleaning System"
echo "========================================================================="

# Reload systemd
echo "Reloading systemd..."
$SUDO systemctl daemon-reload || true
print_status "Systemd reloaded"

# Clean package cache
if [ "$PKG_MGR" = "apt" ]; then
    echo "Cleaning apt cache..."
    $SUDO apt-get clean || true
elif [ "$PKG_MGR" = "dnf" ]; then
    echo "Cleaning dnf cache..."
    $SUDO dnf clean all || true
elif [ "$PKG_MGR" = "yum" ]; then
    echo "Cleaning yum cache..."
    $SUDO yum clean all || true
fi
print_status "Package cache cleaned"

# Clean Go test cache
if command -v go &> /dev/null; then
    echo "Cleaning Go test cache..."
    go clean -testcache || true
    print_status "Go test cache cleaned"
fi

if [ "$EXECUTION_MODE" = "container" ]; then
    echo ""
    echo "========================================================================="
    echo "Step 5: Docker Cleanup"
    echo "========================================================================="

    if command -v docker &> /dev/null; then
        print_info "Found Docker installation"

        # List all test containers
        TEST_CONTAINERS=$(docker ps -a --filter "label=ai-tools-framework" --format "{{.ID}}" 2>/dev/null || true)
        if [ -n "$TEST_CONTAINERS" ]; then
            echo "Found test containers: $TEST_CONTAINERS"
            read -p "Remove test containers? (y/N) " -n 1 -r
            echo
            if [[ $REPLY =~ ^[Yy]$ ]]; then
                docker stop $TEST_CONTAINERS 2>/dev/null || true
                docker rm $TEST_CONTAINERS 2>/dev/null || true
                print_status "Test containers removed"
            fi
        else
            print_warning "No test containers found"
        fi

        # Optional: Clean up systemd-based containers
        SYSTEMD_CONTAINERS=$(docker ps -a --filter "ancestor=jrei/systemd-ubuntu:22.04" --format "{{.ID}}" 2>/dev/null || true)
        if [ -n "$SYSTEMD_CONTAINERS" ]; then
            echo "Found systemd test containers"
            read -p "Remove systemd containers? (y/N) " -n 1 -r
            echo
            if [[ $REPLY =~ ^[Yy]$ ]]; then
                docker stop $SYSTEMD_CONTAINERS 2>/dev/null || true
                docker rm $SYSTEMD_CONTAINERS 2>/dev/null || true
                print_status "Systemd containers removed"
            fi
        fi

        # Optional: Clean up all stopped containers
        read -p "Remove ALL stopped Docker containers? (y/N) " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            docker container prune -f || true
            print_status "Docker containers pruned"
        fi

        # Optional: Clean up unused images
        read -p "Remove unused Docker images? (y/N) " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            docker image prune -f || true
            print_status "Docker images pruned"
        fi
    else
        print_warning "Docker not found, skipping container cleanup"
    fi
fi

echo ""
echo "========================================================================="
echo "Cleanup Summary"
echo "========================================================================="
echo ""
print_status "All services stopped"
print_status "All packages removed"
print_status "All data and configurations cleaned"
print_status "Test logs cleared"
if [ "$EXECUTION_MODE" = "container" ]; then
    print_status "Container cleanup completed"
fi
print_status "System ready for fresh testing"
echo ""
echo "Note: To run tests again, execute:"
if [ "$EXECUTION_MODE" = "container" ]; then
    echo "  make test-container    # For container mode"
else
    echo "  make test-local        # For local mode"
fi
echo "  make test              # For default mode"
echo ""
echo "========================================================================="
echo "Cleanup completed successfully!"
echo "========================================================================="
