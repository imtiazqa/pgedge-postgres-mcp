package testcases

import (
	"fmt"
	"strings"
)

// ============================================================================
// TEST 07: Package Files Verification Tests
// ============================================================================

func (s *MCPServerTestSuite) testFiles_BinariesExist() {
	s.T().Log("Verifying installed binaries...")

	// Ensure packages are installed
	s.EnsureMCPPackagesInstalled()

	binaries := []struct {
		path        string
		permissions string
	}{
		{s.Config.Binaries.MCPServer, "755"},
		{s.Config.Binaries.KBBuilder, "755"},
		{s.Config.Binaries.CLI, "755"},
	}

	for _, bin := range binaries {
		// Check if binary exists
		s.AssertFileExists(bin.path)

		// Check permissions
		output, exitCode, err := s.ExecCommand(fmt.Sprintf("stat -c '%%a' %s", bin.path))
		s.NoError(err, "Failed to check permissions for %s", bin.path)
		s.Equal(0, exitCode, "Should get permissions for %s", bin.path)
		s.Contains(output, bin.permissions, "%s should have %s permissions", bin.path, bin.permissions)

		// Check ownership (should be root:root)
		output, exitCode, err = s.ExecCommand(fmt.Sprintf("stat -c '%%U:%%G' %s", bin.path))
		s.NoError(err, "Failed to check ownership for %s", bin.path)
		s.Equal(0, exitCode, "Should get ownership for %s", bin.path)
		s.Contains(output, "root:root", "%s should be owned by root:root", bin.path)

		s.T().Logf("  ✓ %s exists with correct permissions (%s) and ownership (root:root)", bin.path, bin.permissions)
	}
}

func (s *MCPServerTestSuite) testFiles_SystemdService() {
	s.T().Log("Verifying systemd service file...")

	serviceFile := "/usr/lib/systemd/system/pgedge-postgres-mcp.service"

	// Check if service file exists
	s.AssertFileExists(serviceFile)

	// Check permissions (should be 644)
	output, exitCode, err := s.ExecCommand(fmt.Sprintf("stat -c '%%a' %s", serviceFile))
	s.NoError(err, "Failed to check service file permissions")
	s.Equal(0, exitCode, "Should get service file permissions")
	s.Contains(output, "644", "Service file should have 644 permissions")

	s.T().Logf("  ✓ %s exists with correct permissions (644)", serviceFile)

	// Check /usr/share/pgedge/nla-web directory
	s.T().Log("Verifying /usr/share/pgedge/nla-web directory...")
	nlaWebPath := "/usr/share/pgedge/nla-web"

	// Check if directory exists
	output, exitCode, err = s.ExecCommand(fmt.Sprintf("test -d %s && echo 'exists'", nlaWebPath))
	s.NoError(err, "Failed to check if %s exists", nlaWebPath)
	s.Equal(0, exitCode, "%s should exist", nlaWebPath)
	s.Contains(output, "exists", "%s should exist", nlaWebPath)

	// Check permissions (should be 755)
	output, exitCode, err = s.ExecCommand(fmt.Sprintf("stat -c '%%a' %s", nlaWebPath))
	s.NoError(err, "Failed to check permissions for %s", nlaWebPath)
	s.Equal(0, exitCode, "Should get permissions for %s", nlaWebPath)
	s.Contains(output, "755", "%s should have 755 permissions", nlaWebPath)

	// Check ownership (should be root:root)
	output, exitCode, err = s.ExecCommand(fmt.Sprintf("stat -c '%%U:%%G' %s", nlaWebPath))
	s.NoError(err, "Failed to check ownership for %s", nlaWebPath)
	s.Equal(0, exitCode, "Should get ownership for %s", nlaWebPath)
	s.Contains(output, "root:root", "%s should be owned by root:root", nlaWebPath)

	s.T().Logf("  ✓ %s exists with correct permissions (755) and ownership (root:root)", nlaWebPath)

	// List files in nla-web directory
	output, exitCode, err = s.ExecCommand(fmt.Sprintf("ls -lrt %s", nlaWebPath))
	s.NoError(err, "Failed to list files in %s", nlaWebPath)
	s.Equal(0, exitCode, "Should be able to list files in %s", nlaWebPath)

	s.T().Logf("    Files in %s:", nlaWebPath)
	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		if line != "" && !strings.HasPrefix(line, "total") {
			s.T().Logf("      %s", line)
		}
	}
}

func (s *MCPServerTestSuite) testFiles_ConfigFiles() {
	s.T().Log("Verifying configuration files...")

	configDir := s.Config.ConfigDir
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
		path := fmt.Sprintf("%s/%s", configDir, cfg.name)

		// Check if file exists
		s.AssertFileExists(path)

		// Check permissions
		output, exitCode, err := s.ExecCommand(fmt.Sprintf("stat -c '%%a' %s", path))
		s.NoError(err, "Failed to check permissions for %s", path)
		s.Equal(0, exitCode, "Should get permissions for %s", path)
		s.Contains(output, cfg.permissions, "%s should have %s permissions", path, cfg.permissions)

		// Check ownership (should be root:root)
		output, exitCode, err = s.ExecCommand(fmt.Sprintf("stat -c '%%U:%%G' %s", path))
		s.NoError(err, "Failed to check ownership for %s", path)
		s.Equal(0, exitCode, "Should get ownership for %s", path)
		s.Contains(output, "root:root", "%s should be owned by root:root", path)

		s.T().Logf("  ✓ %s exists with correct permissions (%s) and ownership (root:root)", path, cfg.permissions)
	}
}

func (s *MCPServerTestSuite) testFiles_DataDirectory() {
	s.T().Log("Verifying data directory...")

	dataDir := "/var/lib/pgedge/postgres-mcp"

	// Check if directory exists
	s.AssertDirectoryExists(dataDir)

	// Check permissions
	output, exitCode, err := s.ExecCommand(fmt.Sprintf("stat -c '%%a' %s", dataDir))
	s.NoError(err, "Failed to check data directory permissions")
	s.Equal(0, exitCode, "Should get data directory permissions")
	s.Contains(output, "755", "Data directory should have 755 permissions")

	// Verify ownership is pgedge:pgedge
	output, exitCode, err = s.ExecCommand(fmt.Sprintf("stat -c '%%U:%%G' %s", dataDir))
	s.NoError(err, "Failed to check data directory ownership")
	s.Equal(0, exitCode, "Should get data directory ownership")
	s.Contains(output, "pgedge:pgedge", "%s should be owned by pgedge:pgedge", dataDir)

	s.T().Logf("  ✓ %s exists with correct permissions (755) and ownership (pgedge:pgedge)", dataDir)
}

func (s *MCPServerTestSuite) testFiles_LogDirectories() {
	s.T().Log("Verifying log directories...")

	logDirectories := []struct {
		path        string
		owner       string
		permissions string
	}{
		{"/var/log/pgedge/postgres-mcp", "pgedge:pgedge", "755"},
		{"/var/log/pgedge/nla-web", "pgedge:pgedge", "755"},
	}

	for _, logDir := range logDirectories {
		// Check if directory exists (might not exist until service runs)
		output, exitCode, _ := s.ExecCommand(fmt.Sprintf("test -d %s && echo 'exists' || echo 'missing'", logDir.path))

		if exitCode == 0 && strings.Contains(output, "exists") {
			// Directory exists, verify permissions and ownership

			// Check permissions
			output, exitCode, err := s.ExecCommand(fmt.Sprintf("stat -c '%%a' %s", logDir.path))
			s.NoError(err, "Failed to check permissions for %s", logDir.path)
			s.Equal(0, exitCode, "Should get permissions for %s", logDir.path)
			s.Contains(output, logDir.permissions, "%s should have %s permissions", logDir.path, logDir.permissions)

			// Check ownership
			output, exitCode, err = s.ExecCommand(fmt.Sprintf("stat -c '%%U:%%G' %s", logDir.path))
			s.NoError(err, "Failed to check ownership for %s", logDir.path)
			s.Equal(0, exitCode, "Should get ownership for %s", logDir.path)
			s.Contains(output, logDir.owner, "%s should be owned by %s", logDir.path, logDir.owner)

			s.T().Logf("  ✓ %s exists with correct permissions (%s) and ownership (%s)", logDir.path, logDir.permissions, logDir.owner)

			// List log files if any
			output, exitCode, _ = s.ExecCommand(fmt.Sprintf("ls -lh %s 2>/dev/null | tail -n +2 || echo 'empty'", logDir.path))
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
			s.T().Logf("  ℹ %s not yet created (will be created on first service run)", logDir.path)
		}
	}

	// Check parent log directory
	s.T().Log("Verifying parent log directory...")
	parentLogDir := "/var/log/pgedge"

	output, exitCode, err := s.ExecCommand(fmt.Sprintf("test -d %s && echo 'exists'", parentLogDir))
	s.NoError(err, "Failed to check if parent log directory exists")
	s.Equal(0, exitCode, "Parent log directory should exist")
	s.Contains(output, "exists", "%s should exist", parentLogDir)

	// Check permissions
	output, exitCode, err = s.ExecCommand(fmt.Sprintf("stat -c '%%a' %s", parentLogDir))
	s.NoError(err, "Failed to check parent log directory permissions")
	s.Equal(0, exitCode, "Should get parent log directory permissions")
	s.Contains(output, "755", "Parent log directory should have 755 permissions")

	s.T().Logf("  ✓ %s exists with correct permissions (755)", parentLogDir)
}
