#!/bin/bash
# Cleanup script for pgEdge MCP regression test environment
# This script removes all test artifacts, services, and configurations

set -e  # Exit on error

echo "========================================================================="
echo "pgEdge MCP Regression Test Cleanup"
echo "========================================================================="
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
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

# Check if running with sudo/root for local mode
if [ "$EUID" -eq 0 ]; then
    SUDO=""
else
    SUDO="sudo"
fi

echo "This script will clean up:"
echo "  • PostgreSQL server and data"
echo "  • MCP server service"
echo "  • Ollama service and models"
echo "  • pgEdge packages and repositories"
echo "  • Test configuration files"
echo "  • Docker containers (if any)"
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

# Stop Ollama
if systemctl is-active --quiet ollama 2>/dev/null; then
    echo "Stopping Ollama service..."
    $SUDO systemctl stop ollama || true
    $SUDO systemctl disable ollama || true
    print_status "Ollama stopped"
else
    print_warning "Ollama service not running"
fi

# Stop PostgreSQL
for pg_service in postgresql-18 postgresql-17 postgresql-16 postgresql; do
    if systemctl is-active --quiet $pg_service 2>/dev/null; then
        echo "Stopping $pg_service..."
        $SUDO systemctl stop $pg_service || true
        $SUDO systemctl disable $pg_service || true
        print_status "$pg_service stopped"
    fi
done

# Kill any remaining MCP server processes
if pgrep -f "pgedge-postgres-mcp" > /dev/null; then
    echo "Killing remaining MCP server processes..."
    pkill -f "pgedge-postgres-mcp" || true
    print_status "MCP server processes terminated"
fi

echo ""
echo "========================================================================="
echo "Step 2: Removing Packages"
echo "========================================================================="

# Detect package manager
if command -v apt-get &> /dev/null; then
    PKG_MGR="apt"
    echo "Detected Debian/Ubuntu system"

    # Remove pgEdge packages
    echo "Removing pgEdge packages..."
    $SUDO apt-get remove -y pgedge-postgres-mcp pgedge-nla-kb-builder 2>/dev/null || true

    # Remove PostgreSQL
    echo "Removing PostgreSQL packages..."
    $SUDO apt-get remove -y 'postgresql-*' 'pgedge-postgresql-*' 2>/dev/null || true

    # Remove pgedge-release package to ensure clean reinstall
    echo "Removing pgedge-release package..."
    $SUDO apt-get remove -y pgedge-release 2>/dev/null || true

    $SUDO apt-get autoremove -y || true

elif command -v dnf &> /dev/null; then
    PKG_MGR="dnf"
    echo "Detected RHEL/Rocky/Alma system"

    # Remove pgEdge packages
    echo "Removing pgEdge packages..."
    $SUDO dnf remove -y pgedge-postgres-mcp pgedge-nla-kb-builder 2>/dev/null || true

    # Remove PostgreSQL
    echo "Removing PostgreSQL packages..."
    $SUDO dnf remove -y 'postgresql*' 'pgedge-postgresql*' 2>/dev/null || true

    # Remove pgedge-release package to ensure clean reinstall
    echo "Removing pgedge-release package..."
    $SUDO dnf remove -y pgedge-release 2>/dev/null || true

    $SUDO dnf autoremove -y || true

elif command -v yum &> /dev/null; then
    PKG_MGR="yum"
    echo "Detected older RHEL system"

    # Remove pgEdge packages
    echo "Removing pgEdge packages..."
    $SUDO yum remove -y pgedge-postgres-mcp pgedge-nla-kb-builder 2>/dev/null || true

    # Remove PostgreSQL
    echo "Removing PostgreSQL packages..."
    $SUDO yum remove -y 'postgresql*' 'pgedge-postgresql*' 2>/dev/null || true

    # Remove pgedge-release package to ensure clean reinstall
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
print_status "PostgreSQL data removed"

# Remove pgEdge directories
echo "Removing pgEdge directories..."
$SUDO rm -rf /etc/pgedge || true
$SUDO rm -rf /usr/share/pgedge || true
$SUDO rm -rf /var/log/pgedge || true
$SUDO rm -rf /var/lib/pgedge || true
rm -rf ~/.pgedge || true
print_status "pgEdge directories removed"

# Remove Ollama
echo "Removing Ollama..."
$SUDO rm -rf /usr/local/bin/ollama || true
$SUDO rm -rf /usr/share/ollama || true
$SUDO rm -rf ~/.ollama || true
$SUDO rm -f /etc/systemd/system/ollama.service || true
print_status "Ollama removed"

# Remove test files
echo "Removing test files..."
rm -rf /tmp/test_kb_database || true
rm -rf /tmp/test-mcp-server-config.yaml || true
rm -rf /tmp/mcp-server-test.log || true
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
$SUDO systemctl daemon-reload
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

echo ""
echo "========================================================================="
echo "Step 5: Docker Cleanup (Optional)"
echo "========================================================================="

if command -v docker &> /dev/null; then
    echo "Found Docker installation"

    # Stop and remove MCP test containers
    if docker ps -a | grep -q "mcp-test"; then
        echo "Stopping MCP test containers..."
        docker stop $(docker ps -a | grep "mcp-test" | awk '{print $1}') 2>/dev/null || true
        docker rm $(docker ps -a | grep "mcp-test" | awk '{print $1}') 2>/dev/null || true
        print_status "MCP test containers removed"
    else
        print_warning "No MCP test containers found"
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

echo ""
echo "========================================================================="
echo "Cleanup Summary"
echo "========================================================================="
echo ""
print_status "All services stopped"
print_status "All packages removed"
print_status "All data and configurations cleaned"
print_status "System ready for fresh installation"
echo ""
echo "Note: If you want to run tests again, simply execute:"
echo "  make Execute_Regression_suite"
echo ""
echo "========================================================================="
echo "Cleanup completed successfully!"
echo "========================================================================="
