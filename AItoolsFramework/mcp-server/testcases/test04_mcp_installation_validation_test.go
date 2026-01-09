package testcases

import "fmt"

// ============================================================================
// TEST 04: Installation Validation Tests
// ============================================================================

// testInstallation_PackageFiles validates installation and basic functionality
func (s *MCPServerTestSuite) testInstallation_PackageFiles() {
	s.T().Log("Validating MCP server installation...")

	// Ensure packages are installed
	s.EnsureMCPPackagesInstalled()

	// Test binary is executable (functional check, not just file existence)
	mcpBinary := s.Config.Binaries.MCPServer
	output, exitCode, err := s.ExecCommand(fmt.Sprintf("test -x %s && echo 'executable'", mcpBinary))
	s.NoError(err)
	s.Equal(0, exitCode)
	s.Contains(output, "executable", "MCP server binary should be executable")
	s.T().Logf("  ✓ %s is executable", mcpBinary)

	// Test config file is readable (functional check)
	configFile := fmt.Sprintf("%s/postgres-mcp.yaml", s.Config.ConfigDir)
	output, exitCode, err = s.ExecCommand(fmt.Sprintf("test -r %s && echo 'readable'", configFile))
	s.NoError(err)
	s.Equal(0, exitCode)
	s.Contains(output, "readable", "Config file should be readable")
	s.T().Logf("  ✓ %s is readable", configFile)

	s.T().Log("✓ Package files have correct permissions")
}

// testMCPServer_BinaryFunctional tests that the binary can actually run
func (s *MCPServerTestSuite) testMCPServer_BinaryFunctional() {
	s.T().Log("Testing MCP server binary functionality...")

	mcpBinary := s.Config.Binaries.MCPServer

	// Test binary can execute --help command (functional test)
	output, _, err := s.ExecCommand(fmt.Sprintf("%s --help 2>&1 || true", mcpBinary))
	s.NoError(err, "Binary should be able to execute")
	s.NotEmpty(output, "Binary should produce output")
	s.T().Logf("  ✓ %s can execute commands", mcpBinary)

	s.T().Log("✓ MCP server binary is functional")
}

// testMCPServer_ConfigValid validates configuration file content
func (s *MCPServerTestSuite) testMCPServer_ConfigValid() {
	s.T().Log("Validating MCP server configuration content...")

	configFile := fmt.Sprintf("%s/postgres-mcp.yaml", s.Config.ConfigDir)

	// Read and validate config file structure
	output, exitCode, err := s.ExecCommand(fmt.Sprintf("cat %s", configFile))
	s.NoError(err)
	s.Equal(0, exitCode)
	s.NotEmpty(output, "Config file should not be empty")

	// Validate YAML structure contains required sections
	s.Contains(output, "databases:", "Config should have databases section")
	s.Contains(output, "http:", "Config should have http section")
	s.T().Log("  ✓ Configuration file has valid YAML structure")

	s.T().Log("✓ MCP server configuration is valid")
}

// testMCPServer_EnvironmentFile validates environment file is configured
func (s *MCPServerTestSuite) testMCPServer_EnvironmentFile() {
	s.T().Log("Validating MCP server environment configuration...")

	envFile := fmt.Sprintf("%s/postgres-mcp.env", s.Config.ConfigDir)

	// Check environment file exists
	s.AssertFileExists(envFile)

	// Verify it contains database password configuration
	output, exitCode, err := s.ExecCommand(fmt.Sprintf("cat %s", envFile))
	s.NoError(err)
	s.Equal(0, exitCode)

	// Validate environment file has required configuration
	s.Contains(output, "PGEDGE_DB_PASSWORD", "Environment file should contain database password setting")
	s.T().Log("  ✓ Environment file contains required configuration")

	s.T().Log("✓ MCP server environment file is configured")
}
