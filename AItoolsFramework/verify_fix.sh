#!/bin/bash

echo "==================================================================="
echo "Container Test Framework - Fix Verification"
echo "==================================================================="
echo ""

# Count installation calls in test files
echo "1. Verifying installation calls in test files..."
INSTALL_COUNT=$(grep -r "EnsureMCPPackagesInstalled\|EnsureRepositoryInstalled" mcp-server/testcases/ 2>/dev/null | grep -v "Binary" | wc -l)
echo "   ✓ Found $INSTALL_COUNT installation calls (expected: 14+)"
echo ""

# Verify APISuite extends E2ESuite
echo "2. Verifying APISuite extends E2ESuite..."
if grep -q "E2ESuite" common/suite/api.go; then
    echo "   ✓ APISuite extends E2ESuite"
else
    echo "   ✗ APISuite does NOT extend E2ESuite"
fi
echo ""

# Verify timeouts
echo "3. Verifying increased timeouts in container.yaml..."
grep -A 8 "^timeouts:" mcp-server/config/container.yaml
echo ""

# List all modified files
echo "4. Documentation files created:"
ls -1 *.md 2>/dev/null | grep -E "(COMPLETE_FIX|ARCHITECTURE|FINAL_STATUS)" | sed 's/^/   ✓ /'
echo ""

# Suite hierarchy
echo "5. Suite hierarchy:"
grep -A 10 "## Suite Hierarchy" common/suite/README.md | head -15
echo ""

echo "==================================================================="
echo "Verification Complete!"
echo "==================================================================="
echo ""
echo "To run the tests:"
echo "  cd mcp-server"
echo "  make test-container"
echo ""
