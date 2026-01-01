package regression

// ========================================================================
// TEST 03: MCP Server Package Installation
// ========================================================================

// Test03_MCPServerInstallation tests MCP server package installation
func (s *RegressionTestSuite) Test03_MCPServerInstallation() {
	s.T().Log("TEST 03: Installing MCP server packages")
	s.ensureMCPPackagesInstalled()
	s.T().Log("✓ All MCP server packages installed and configured successfully")
}

// installMCPPackages performs the actual MCP package installation
func (s *RegressionTestSuite) installMCPPackages() {

	isDebian, _ := s.getOSType()

	// Step 1: Install MCP server packages
	s.logDetailed("Step 1: Installing MCP server packages")
	var packages []string
	if isDebian {
		packages = []string{
			"apt-get install -y pgedge-postgres-mcp",
			"apt-get install -y pgedge-nla-cli",
			"apt-get install -y pgedge-nla-web",
			"apt-get install -y pgedge-postgres-mcp-kb",
		}
	} else {
		packages = []string{
			"dnf install -y pgedge-postgres-mcp",
			"dnf install -y pgedge-nla-cli",
			"dnf install -y pgedge-nla-web",
			"dnf install -y pgedge-postgres-mcp-kb",
		}
	}

	for _, installCmd := range packages {
		output, exitCode, err := s.execCmd(s.ctx, installCmd)
		s.NoError(err, "Install failed: %s\nOutput: %s", installCmd, output)
		s.Equal(0, exitCode, "Install exited with error: %s\nOutput: %s", installCmd, output)
	}

	s.T().Log("  ✓ MCP server packages installed successfully")
	s.T().Log("  Note: Shipped default configs remain intact at /etc/pgedge/")
}
