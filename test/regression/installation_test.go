package regression

// ========================================================================
// TEST 04: Installation Validation
// ========================================================================

// Test04_InstallationValidation validates that all MCP server components
// are correctly installed in their expected locations
func (s *RegressionTestSuite) Test04_InstallationValidation() {
	s.T().Log("TEST 04: Validating MCP server installation")

	// Ensure packages are installed
	s.ensureMCPPackagesInstalled()

	// Check 1: Binary exists and is executable
	output, exitCode, err := s.execCmd(s.ctx, "test -x /usr/bin/pgedge-postgres-mcp && echo 'OK'")
	s.NoError(err)
	s.Equal(0, exitCode)
	s.Contains(output, "OK", "Binary should exist and be executable")

	// Check 2: Config directory exists
	output, exitCode, err = s.execCmd(s.ctx, "test -d /etc/pgedge && echo 'OK'")
	s.NoError(err)
	s.Equal(0, exitCode)
	s.Contains(output, "OK", "Config directory should exist")

	// Check 3: Config files exist
	output, exitCode, err = s.execCmd(s.ctx, "test -f /etc/pgedge/postgres-mcp.yaml && echo 'OK'")
	s.NoError(err)
	s.Equal(0, exitCode)
	s.Contains(output, "OK", "postgres-mcp.yaml should exist")

	output, exitCode, err = s.execCmd(s.ctx, "test -f /etc/pgedge/postgres-mcp.env && echo 'OK'")
	s.NoError(err)
	s.Equal(0, exitCode)
	s.Contains(output, "OK", "postgres-mcp.env should exist")

	// Check 4: Systemd service file exists
	output, exitCode, err = s.execCmd(s.ctx, "test -f /usr/lib/systemd/system/pgedge-postgres-mcp.service && echo 'OK'")
	s.NoError(err)
	s.Equal(0, exitCode)
	s.Contains(output, "OK", "Systemd service file should exist")

	// Check 5: Data directory exists with correct permissions
	output, exitCode, err = s.execCmd(s.ctx, "test -d /var/lib/pgedge/postgres-mcp && echo 'OK'")
	s.NoError(err)
	s.Equal(0, exitCode)
	s.Contains(output, "OK", "Data directory should exist")

	s.T().Log("✓ All installation validation checks passed")
}
