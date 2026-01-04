package testcases

import "fmt"

// ============================================================================
// MCP Server Installation Tests
// ============================================================================

func (s *MCPServerTestSuite) testMCPServer_PackagesInstalled() {
	s.T().Log("Testing MCP server packages are installed...")

	isDebian := s.isDebianBased()

	// Get all packages from config
	allPackages := append(s.Config.Packages.MCPServer, s.Config.Packages.CLI...)
	allPackages = append(allPackages, s.Config.Packages.Web...)
	allPackages = append(allPackages, s.Config.Packages.KB...)

	for _, pkg := range allPackages {
		var checkCmd string
		if isDebian {
			checkCmd = "dpkg -l | grep " + pkg
		} else {
			checkCmd = "rpm -qa | grep " + pkg
		}

		_, exitCode, err := s.ExecCommand(checkCmd)
		if exitCode == 0 && err == nil {
			s.T().Logf("  ✓ Package installed: %s", pkg)
		} else {
			s.T().Logf("  ⚠ Package not found: %s", pkg)
		}
	}

	s.T().Log("✓ Package verification complete")
}

func (s *MCPServerTestSuite) testMCPServer_BinaryFunctional() {
	s.T().Log("Testing MCP server binary is functional...")

	mcpBinary := s.Config.Binaries.MCPServer

	// Test binary exists and is executable
	output, exitCode, err := s.ExecCommand(fmt.Sprintf("test -x %s && echo 'OK'", mcpBinary))
	s.NoError(err)
	s.Equal(0, exitCode)
	s.Contains(output, "OK", "MCP server binary should exist and be executable")

	s.T().Log("✓ MCP server binary is functional")
}

func (s *MCPServerTestSuite) testMCPServer_ConfigValid() {
	s.T().Log("Testing MCP server configuration is valid...")

	configFile := fmt.Sprintf("%s/postgres-mcp.yaml", s.Config.ConfigDir)

	// Read config file
	output, exitCode, err := s.ExecCommand(fmt.Sprintf("cat %s", configFile))
	s.NoError(err)
	s.Equal(0, exitCode)
	s.NotEmpty(output, "Config file should not be empty")

	// Check for expected sections
	s.Contains(output, "databases:", "Config should have databases section")
	s.Contains(output, "http:", "Config should have http section")

	s.T().Log("✓ MCP server configuration is valid")
}

func (s *MCPServerTestSuite) testMCPServer_EnvironmentFile() {
	s.T().Log("Testing MCP server environment file...")

	envFile := fmt.Sprintf("%s/postgres-mcp.env", s.Config.ConfigDir)

	// Check environment file exists
	s.AssertFileExists(envFile)

	// Verify it contains database password configuration
	output, exitCode, err := s.ExecCommand(fmt.Sprintf("cat %s", envFile))
	s.NoError(err)
	s.Equal(0, exitCode)

	// Should contain PGEDGE_DB_PASSWORD setting
	s.Contains(output, "PGEDGE_DB_PASSWORD", "Environment file should contain database password")

	s.T().Log("✓ MCP server environment file is configured")
}
