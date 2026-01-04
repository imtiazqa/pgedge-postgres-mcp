package testcases

import "fmt"

// ============================================================================
// Package Files Verification Tests
// ============================================================================

func (s *MCPServerTestSuite) testFiles_BinariesExist() {
	s.T().Log("Testing binaries...")

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

		s.T().Logf("  ✓ %s exists with correct permissions (%s)", bin.path, bin.permissions)
	}
}

func (s *MCPServerTestSuite) testFiles_SystemdService() {
	s.T().Log("Testing systemd service file...")

	serviceFile := "/usr/lib/systemd/system/pgedge-postgres-mcp.service"

	// Check if service file exists
	s.AssertFileExists(serviceFile)

	// Check permissions (should be 644)
	output, exitCode, err := s.ExecCommand(fmt.Sprintf("stat -c '%%a' %s", serviceFile))
	s.NoError(err, "Failed to check service file permissions")
	s.Equal(0, exitCode, "Should get service file permissions")
	s.Contains(output, "644", "Service file should have 644 permissions")

	s.T().Logf("  ✓ %s exists with correct permissions (644)", serviceFile)
}

func (s *MCPServerTestSuite) testFiles_ConfigFiles() {
	s.T().Log("Testing configuration files...")

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

		s.T().Logf("  ✓ %s exists with correct permissions (%s)", path, cfg.permissions)
	}
}

func (s *MCPServerTestSuite) testFiles_DataDirectory() {
	s.T().Log("Testing data directory...")

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

	s.T().Logf("  ✓ %s exists with correct ownership (pgedge:pgedge)", dataDir)
}

func (s *MCPServerTestSuite) testFiles_LogDirectories() {
	s.T().Log("Testing log directories...")

	logDirectories := []struct {
		path  string
		owner string
	}{
		{"/var/log/pgedge/postgres-mcp", "pgedge:pgedge"},
		{"/var/log/pgedge/nla-web", "pgedge:pgedge"},
	}

	for _, logDir := range logDirectories {
		// Check if directory exists (might not exist until service runs)
		output, exitCode, _ := s.ExecCommand(fmt.Sprintf("test -d %s && echo 'exists' || echo 'missing'", logDir.path))

		if exitCode == 0 && output == "exists\n" {
			s.T().Logf("  ✓ %s exists", logDir.path)
		} else {
			s.T().Logf("  ℹ %s not yet created (will be created on first service run)", logDir.path)
		}
	}
}
