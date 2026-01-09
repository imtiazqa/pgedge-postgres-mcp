package testcases

import "fmt"

// ============================================================================
// TEST 05: Token Management Tests
// ============================================================================

// testToken_FileExists verifies token file location is accessible (runs first)
func (s *MCPServerTestSuite) testToken_FileExists() {
	s.T().Log("Verifying token file location accessibility...")

	// Ensure packages are installed (this runs first, so install here)
	s.EnsureMCPPackagesInstalled()

	// Verify the config directory exists
	configDir := s.Config.ConfigDir
	s.AssertDirectoryExists(configDir)

	s.T().Logf("✓ Config directory %s is accessible", configDir)
}

// testToken_CreateToken creates a new token and verifies creation
func (s *MCPServerTestSuite) testToken_CreateToken() {
	s.T().Log("Testing token creation...")

	mcpBinary := s.Config.Binaries.MCPServer
	configFile := fmt.Sprintf("%s/postgres-mcp.yaml", s.Config.ConfigDir)
	tokenFile := fmt.Sprintf("%s/pgedge-postgres-mcp-tokens.yaml", s.Config.ConfigDir)

	// Create token
	createCmd := fmt.Sprintf("%s -config %s -add-token -token-file %s -token-note \"test-token\"",
		mcpBinary, configFile, tokenFile)

	output, exitCode, err := s.ExecCommand(createCmd)
	s.NoError(err, "Token creation failed\nOutput: %s", output)
	s.Equal(0, exitCode, "Token creation should succeed")
	s.Contains(output, "Token:", "Should show generated token")
	s.Contains(output, "Hash:", "Should show token hash")

	// Set proper ownership on token file
	chownCmd := fmt.Sprintf("chown pgedge:pgedge %s", tokenFile)
	output, exitCode, err = s.ExecCommand(chownCmd)
	s.NoError(err, "Failed to set ownership on token file: %s", output)
	s.Equal(0, exitCode, "chown should succeed: %s", output)

	s.T().Log("✓ Token created successfully")
}

// testToken_ListTokens lists tokens and verifies the created token exists
func (s *MCPServerTestSuite) testToken_ListTokens() {
	s.T().Log("Testing token listing...")

	mcpBinary := s.Config.Binaries.MCPServer
	configFile := fmt.Sprintf("%s/postgres-mcp.yaml", s.Config.ConfigDir)
	tokenFile := fmt.Sprintf("%s/pgedge-postgres-mcp-tokens.yaml", s.Config.ConfigDir)

	// List tokens
	listCmd := fmt.Sprintf("%s -config %s -list-tokens -token-file %s", mcpBinary, configFile, tokenFile)
	output, exitCode, err := s.ExecCommand(listCmd)
	s.NoError(err, "Token listing failed")
	s.Equal(0, exitCode, "Token listing should succeed")
	s.Contains(output, "test-token", "Should list created token")

	// Verify token file was created
	output, exitCode, err = s.ExecCommand(fmt.Sprintf("test -f %s && echo 'OK'", tokenFile))
	s.NoError(err, "Failed to check token file")
	s.Equal(0, exitCode, "Token file check should succeed")
	s.Contains(output, "OK", "Token file should exist")

	s.T().Log("✓ Token management working correctly")
}
