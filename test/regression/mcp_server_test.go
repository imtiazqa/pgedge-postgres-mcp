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

	// Step 2: Configure environment variables for test database password
	s.logDetailed("Step 2: Configuring database password in postgres-mcp.env")

	// Set PGEDGE_DB_PASSWORD to match the test database password
	envSetCmd := `grep -q "^PGEDGE_DB_PASSWORD=" /etc/pgedge/postgres-mcp.env 2>/dev/null && \
		sed -i 's/^PGEDGE_DB_PASSWORD=.*/PGEDGE_DB_PASSWORD=postgres123/' /etc/pgedge/postgres-mcp.env || \
		echo 'PGEDGE_DB_PASSWORD=postgres123' >> /etc/pgedge/postgres-mcp.env`

	output, exitCode, err = s.execCmd(s.ctx, envSetCmd)
	s.NoError(err, "Failed to set database password: %s", output)
	s.Equal(0, exitCode, "Set database password failed: %s", output)

	s.T().Log("  ✓ Database password configured in postgres-mcp.env")
	s.T().Log("  Note: Shipped default config remains intact at /etc/pgedge/postgres-mcp.yaml")
}
