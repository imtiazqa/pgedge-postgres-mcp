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

	// Step 2: Update MCP server configuration
	s.logDetailed("Step 2: Updating MCP server configuration files")

	// Update postgres-mcp.yaml
	yamlConfig := `cat > /etc/pgedge/postgres-mcp.yaml << 'EOF'
databases:
  - host: localhost
    port: 5432
    name: mcp_server
    user: postgres
    password: postgres123
server:
  mode: http
  addr: :8080
EOF`
	output, exitCode, err := s.execCmd(s.ctx, yamlConfig)
	s.NoError(err, "Failed to update postgres-mcp.yaml: %s", output)
	s.Equal(0, exitCode, "Update config failed: %s", output)

	// Update postgres-mcp.env
	envConfig := `cat > /etc/pgedge/postgres-mcp.env << 'EOF'
DB_HOST=localhost
DB_PORT=5432
DB_NAME=mcp_server
DB_USER=postgres
DB_PASSWORD=postgres123
SERVER_MODE=http
SERVER_ADDR=:8080
EOF`
	output, exitCode, err = s.execCmd(s.ctx, envConfig)
	s.NoError(err, "Failed to update postgres-mcp.env: %s", output)
	s.Equal(0, exitCode, "Update env failed: %s", output)
}
