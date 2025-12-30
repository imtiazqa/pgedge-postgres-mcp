package regression

import (
	"fmt"
	"strings"
)

// ========================================================================
// TEST 07: Package Files Verification
// ========================================================================

// Test07_PackageFilesVerification verifies installed package files and permissions
func (s *RegressionTestSuite) Test07_PackageFilesVerification() {
	s.T().Log("TEST 07: Verifying installed package files and permissions")

	// Ensure packages are installed
	s.ensureMCPPackagesInstalled()

	// ====================================================================
	// 1. Verify binaries in /usr/bin with executable permissions
	// ====================================================================
	s.T().Log("Checking binaries in /usr/bin...")

	binaries := []struct {
		name        string
		permissions string // Expected permissions pattern
	}{
		{"pgedge-postgres-mcp", "755"},
		{"pgedge-nla-kb-builder", "755"},
		{"pgedge-nla-cli", "755"},
	}

	for _, bin := range binaries {
		// Check if binary exists
		output, exitCode, err := s.execCmd(s.ctx, fmt.Sprintf("test -f /usr/bin/%s && echo 'exists'", bin.name))
		s.NoError(err, "Failed to check if %s exists", bin.name)
		s.Equal(0, exitCode, "%s should exist in /usr/bin", bin.name)
		s.Contains(output, "exists", "%s should exist in /usr/bin", bin.name)

		// Check permissions (should be executable: 755 or 775)
		output, exitCode, err = s.execCmd(s.ctx, fmt.Sprintf("stat -c '%%a' /usr/bin/%s", bin.name))
		s.NoError(err, "Failed to check permissions for %s", bin.name)
		s.Equal(0, exitCode, "Should get permissions for %s", bin.name)
		s.Contains(output, bin.permissions, "%s should have %s permissions", bin.name, bin.permissions)

		// Verify it's owned by root
		output, exitCode, err = s.execCmd(s.ctx, fmt.Sprintf("stat -c '%%U:%%G' /usr/bin/%s", bin.name))
		s.NoError(err, "Failed to check ownership for %s", bin.name)
		s.Equal(0, exitCode, "Should get ownership for %s", bin.name)
		s.Contains(output, "root:root", "%s should be owned by root:root", bin.name)

		s.T().Logf("  ✓ /usr/bin/%s exists with correct permissions (%s)", bin.name, bin.permissions)
	}

	// ====================================================================
	// 2. Verify systemd service file
	// ====================================================================
	s.T().Log("Checking systemd service file...")

	serviceFile := "/usr/lib/systemd/system/pgedge-postgres-mcp.service"

	// Check if service file exists
	output, exitCode, err := s.execCmd(s.ctx, fmt.Sprintf("test -f %s && echo 'exists'", serviceFile))
	s.NoError(err, "Failed to check if service file exists")
	s.Equal(0, exitCode, "Service file should exist")
	s.Contains(output, "exists", "Service file should exist at %s", serviceFile)

	// Check permissions (should be 644)
	output, exitCode, err = s.execCmd(s.ctx, fmt.Sprintf("stat -c '%%a' %s", serviceFile))
	s.NoError(err, "Failed to check service file permissions")
	s.Equal(0, exitCode, "Should get service file permissions")
	s.Contains(output, "644", "Service file should have 644 permissions")

	s.T().Logf("  ✓ %s exists with correct permissions (644)", serviceFile)

	// ====================================================================
	// 3. Verify /usr/share directories
	// ====================================================================
	s.T().Log("Checking /usr/share/pgedge/nla-web directory...")

	nlaWebPath := "/usr/share/pgedge/nla-web"

	// Check if directory exists
	output, exitCode, err = s.execCmd(s.ctx, fmt.Sprintf("test -d %s && echo 'exists'", nlaWebPath))
	s.NoError(err, "Failed to check if %s exists", nlaWebPath)
	s.Equal(0, exitCode, "%s should exist", nlaWebPath)
	s.Contains(output, "exists", "%s should exist", nlaWebPath)

	// Check permissions (should be readable: 755)
	output, exitCode, err = s.execCmd(s.ctx, fmt.Sprintf("stat -c '%%a' %s", nlaWebPath))
	s.NoError(err, "Failed to check permissions for %s", nlaWebPath)
	s.Equal(0, exitCode, "Should get permissions for %s", nlaWebPath)
	s.Contains(output, "755", "%s should have 755 permissions", nlaWebPath)

	// Verify it's owned by root
	output, exitCode, err = s.execCmd(s.ctx, fmt.Sprintf("stat -c '%%U:%%G' %s", nlaWebPath))
	s.NoError(err, "Failed to check ownership for %s", nlaWebPath)
	s.Equal(0, exitCode, "Should get ownership for %s", nlaWebPath)
	s.Contains(output, "root:root", "%s should be owned by root:root", nlaWebPath)

	s.T().Logf("  ✓ %s exists with correct permissions (755)", nlaWebPath)

	// List files inside nla-web directory
	output, exitCode, err = s.execCmd(s.ctx, fmt.Sprintf("ls -lrt %s", nlaWebPath))
	s.NoError(err, "Failed to list files in %s", nlaWebPath)
	s.Equal(0, exitCode, "Should be able to list files in %s", nlaWebPath)

	s.T().Logf("    Files in %s:", nlaWebPath)
	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		if line != "" && !strings.HasPrefix(line, "total") {
			s.T().Logf("      %s", line)
		}
	}

	// ====================================================================
	// 4. Verify /etc/pgedge configuration files
	// ====================================================================
	s.T().Log("Checking /etc/pgedge configuration files...")

	configFiles := []struct {
		name        string
		permissions string
	}{
		{"postgres-mcp.env", "644"},
		{"pgedge-nla-kb-builder.yaml", "644"},
		{"nla-cli.yaml", "644"},
		{"postgres-mcp.yaml", "644"},
	}

	for _, cfg := range configFiles {
		path := fmt.Sprintf("/etc/pgedge/%s", cfg.name)

		// Check if file exists
		output, exitCode, err := s.execCmd(s.ctx, fmt.Sprintf("test -f %s && echo 'exists'", path))
		s.NoError(err, "Failed to check if %s exists", path)
		s.Equal(0, exitCode, "%s should exist", path)
		s.Contains(output, "exists", "%s should exist", path)

		// Check permissions (should be readable: 644)
		output, exitCode, err = s.execCmd(s.ctx, fmt.Sprintf("stat -c '%%a' %s", path))
		s.NoError(err, "Failed to check permissions for %s", path)
		s.Equal(0, exitCode, "Should get permissions for %s", path)
		s.Contains(output, cfg.permissions, "%s should have %s permissions", path, cfg.permissions)

		// Verify it's owned by root
		output, exitCode, err = s.execCmd(s.ctx, fmt.Sprintf("stat -c '%%U:%%G' %s", path))
		s.NoError(err, "Failed to check ownership for %s", path)
		s.Equal(0, exitCode, "Should get ownership for %s", path)
		s.Contains(output, "root:root", "%s should be owned by root:root", path)

		s.T().Logf("  ✓ %s exists with correct permissions (%s)", path, cfg.permissions)
	}

	// ====================================================================
	// 5. Verify /var/lib/pgedge/postgres-mcp directory
	// ====================================================================
	s.T().Log("Checking /var/lib/pgedge/postgres-mcp directory...")

	dataDir := "/var/lib/pgedge/postgres-mcp"

	// Check if directory exists
	output, exitCode, err = s.execCmd(s.ctx, fmt.Sprintf("test -d %s && echo 'exists'", dataDir))
	s.NoError(err, "Failed to check if data directory exists")
	s.Equal(0, exitCode, "Data directory should exist")
	s.Contains(output, "exists", "%s should exist", dataDir)

	// Check permissions (should be 755 for directory)
	output, exitCode, err = s.execCmd(s.ctx, fmt.Sprintf("stat -c '%%a' %s", dataDir))
	s.NoError(err, "Failed to check data directory permissions")
	s.Equal(0, exitCode, "Should get data directory permissions")
	s.Contains(output, "755", "Data directory should have 755 permissions")

	// Verify it's owned by pgedge:pgedge
	output, exitCode, err = s.execCmd(s.ctx, fmt.Sprintf("stat -c '%%U:%%G' %s", dataDir))
	s.NoError(err, "Failed to check data directory ownership")
	s.Equal(0, exitCode, "Should get data directory ownership")
	s.Contains(output, "pgedge:pgedge", "%s should be owned by pgedge:pgedge", dataDir)

	s.T().Logf("  ✓ %s exists with correct ownership (pgedge:pgedge)", dataDir)

	// ====================================================================
	// 6. Verify log directories exist
	// ====================================================================
	s.T().Log("Checking log directories...")

	logDirectories := []struct {
		path        string
		owner       string
		permissions string
	}{
		{"/var/log/pgedge/postgres-mcp", "pgedge:pgedge", "755"},
		{"/var/log/pgedge/nla-web", "pgedge:pgedge", "755"},
	}

	for _, logDir := range logDirectories {
		// Check if directory exists (it might not exist until service runs)
		output, exitCode, err := s.execCmd(s.ctx, fmt.Sprintf("test -d %s && echo 'exists' || echo 'missing'", logDir.path))
		s.NoError(err, "Failed to check if %s exists", logDir.path)

		if strings.Contains(output, "exists") {
			// Directory exists, verify permissions and ownership

			// Check permissions
			output, exitCode, err = s.execCmd(s.ctx, fmt.Sprintf("stat -c '%%a' %s", logDir.path))
			s.NoError(err, "Failed to check permissions for %s", logDir.path)
			s.Equal(0, exitCode, "Should get permissions for %s", logDir.path)
			s.Contains(output, logDir.permissions, "%s should have %s permissions", logDir.path, logDir.permissions)

			// Check ownership
			output, exitCode, err = s.execCmd(s.ctx, fmt.Sprintf("stat -c '%%U:%%G' %s", logDir.path))
			s.NoError(err, "Failed to check ownership for %s", logDir.path)
			s.Equal(0, exitCode, "Should get ownership for %s", logDir.path)
			s.Contains(output, logDir.owner, "%s should be owned by %s", logDir.path, logDir.owner)

			s.T().Logf("  ✓ %s exists with correct ownership (%s)", logDir.path, logDir.owner)

			// List log files inside the directory
			output, exitCode, err = s.execCmd(s.ctx, fmt.Sprintf("ls -lh %s 2>/dev/null || echo 'empty'", logDir.path))
			s.NoError(err, "Failed to list log files in %s", logDir.path)

			if strings.Contains(output, "empty") || strings.TrimSpace(output) == "" {
				s.T().Logf("    ℹ %s is empty (no log files yet)", logDir.path)
			} else {
				s.T().Logf("    Log files in %s:", logDir.path)
				for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
					if line != "" && !strings.HasPrefix(line, "total") {
						s.T().Logf("      %s", line)
					}
				}
			}
		} else {
			// Directory doesn't exist yet - this is acceptable (created on first run)
			s.T().Logf("  ℹ %s not yet created (will be created on first service run)", logDir.path)
		}
	}

	// ====================================================================
	// 7. Verify parent /var/log/pgedge directory
	// ====================================================================
	s.T().Log("Checking parent log directory...")

	parentLogDir := "/var/log/pgedge"

	// Check if parent directory exists
	output, exitCode, err = s.execCmd(s.ctx, fmt.Sprintf("test -d %s && echo 'exists'", parentLogDir))
	s.NoError(err, "Failed to check if parent log directory exists")
	s.Equal(0, exitCode, "Parent log directory should exist")
	s.Contains(output, "exists", "%s should exist", parentLogDir)

	// Check permissions
	output, exitCode, err = s.execCmd(s.ctx, fmt.Sprintf("stat -c '%%a' %s", parentLogDir))
	s.NoError(err, "Failed to check parent log directory permissions")
	s.Equal(0, exitCode, "Should get parent log directory permissions")
	s.Contains(output, "755", "Parent log directory should have 755 permissions")

	s.T().Logf("  ✓ %s exists with correct permissions (755)", parentLogDir)

	s.T().Log("✓ All package files and permissions verified successfully")
}
