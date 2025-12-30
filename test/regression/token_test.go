package regression

// ========================================================================
// TEST 05: Token Management
// ========================================================================

// Test05_TokenManagement tests token creation, listing, and file verification
func (s *RegressionTestSuite) Test05_TokenManagement() {
	s.T().Log("TEST 05: Testing token management commands")

	// Ensure packages are installed
	s.ensureMCPPackagesInstalled()

	// Test 1: Create token (using config file for database connection)
	createCmd := `/usr/bin/pgedge-postgres-mcp -config /etc/pgedge/postgres-mcp.yaml -add-token -token-file /etc/pgedge/pgedge-postgres-mcp-tokens.yaml -token-note "test-token"`
	output, exitCode, err := s.execCmd(s.ctx, createCmd)
	s.NoError(err, "Token creation failed\nOutput: %s", output)
	s.Equal(0, exitCode)
	s.Contains(output, "Token:", "Should show generated token")
	s.Contains(output, "Hash:", "Should show token hash")

	// Set proper ownership on token file
	chownCmd := "chown pgedge:pgedge /etc/pgedge/pgedge-postgres-mcp-tokens.yaml"
	output, exitCode, err = s.execCmd(s.ctx, chownCmd)
	s.NoError(err, "Failed to set ownership on token file: %s", output)
	s.Equal(0, exitCode, "chown failed: %s", output)

	// Test 2: List tokens
	output, exitCode, err = s.execCmd(s.ctx, "/usr/bin/pgedge-postgres-mcp -config /etc/pgedge/postgres-mcp.yaml -list-tokens -token-file /etc/pgedge/pgedge-postgres-mcp-tokens.yaml")
	s.NoError(err)
	s.Equal(0, exitCode)
	s.Contains(output, "test-token", "Should list created token")

	// Test 3: Verify token file was created
	output, exitCode, err = s.execCmd(s.ctx, "test -f /etc/pgedge/pgedge-postgres-mcp-tokens.yaml && echo 'OK'")
	s.NoError(err)
	s.Equal(0, exitCode)
	s.Contains(output, "OK", "Token file should exist")

	s.T().Log("✓ Token management working correctly")
}
