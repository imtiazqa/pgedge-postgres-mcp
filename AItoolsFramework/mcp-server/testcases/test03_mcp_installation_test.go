package testcases

import "fmt"

// ============================================================================
// TEST 03: MCP Server Installation Tests
// ============================================================================

// testMCPServer_PackagesInstalled installs and verifies MCP server packages
func (s *MCPServerTestSuite) testMCPServer_PackagesInstalled() {
	s.T().Log("Testing MCP server packages installation...")

	// Ensure MCP packages are installed (this will install if not already done)
	s.EnsureMCPPackagesInstalled()

	// Verify all packages are installed
	isDebian := s.isDebianBased()

	// Get all packages from config
	allPackages := append(s.Config.Packages.MCPServer, s.Config.Packages.CLI...)
	allPackages = append(allPackages, s.Config.Packages.Web...)
	allPackages = append(allPackages, s.Config.Packages.KB...)

	s.T().Logf("Verifying %d installed packages...", len(allPackages))

	for _, pkg := range allPackages {
		var checkCmd string
		if isDebian {
			checkCmd = fmt.Sprintf("dpkg -l | grep %s", pkg)
		} else {
			checkCmd = fmt.Sprintf("rpm -qa | grep %s", pkg)
		}

		output, exitCode, err := s.ExecCommand(checkCmd)
		s.NoError(err, "Failed to check package %s: %s", pkg, output)
		s.Equal(0, exitCode, "Package %s should be installed", pkg)
		s.T().Logf("  ✓ Package installed: %s", pkg)
	}

	s.T().Log("✓ All MCP server packages installed successfully")
}

// testInstallation_MCPPackages verifies MCP binaries and config files
func (s *MCPServerTestSuite) testInstallation_MCPPackages() {
	s.T().Log("Testing MCP package files...")

	// Verify binaries exist (from config)
	s.T().Log("Checking binaries...")
	s.AssertFileExists(s.Config.Binaries.MCPServer)
	s.T().Logf("  ✓ %s exists", s.Config.Binaries.MCPServer)

	s.AssertFileExists(s.Config.Binaries.CLI)
	s.T().Logf("  ✓ %s exists", s.Config.Binaries.CLI)

	s.AssertFileExists(s.Config.Binaries.KBBuilder)
	s.T().Logf("  ✓ %s exists", s.Config.Binaries.KBBuilder)

	// Verify config files exist (from config)
	s.T().Log("Checking configuration files...")
	configDir := s.Config.ConfigDir

	mcpYaml := fmt.Sprintf("%s/postgres-mcp.yaml", configDir)
	s.AssertFileExists(mcpYaml)
	s.T().Logf("  ✓ %s exists", mcpYaml)

	mcpEnv := fmt.Sprintf("%s/postgres-mcp.env", configDir)
	s.AssertFileExists(mcpEnv)
	s.T().Logf("  ✓ %s exists", mcpEnv)

	s.T().Log("✓ All MCP package files verified")
}

// testInstallation_Repository verifies repository configuration
func (s *MCPServerTestSuite) testInstallation_Repository() {
	s.T().Log("Testing repository configuration...")

	isDebian := s.isDebianBased()

	if isDebian {
		// Debian/Ubuntu - verify apt repository
		output, exitCode, err := s.ExecCommand("apt-cache policy | grep -i pgedge")
		s.NoError(err)
		s.Equal(0, exitCode, "pgEdge repository should be in apt sources")
		s.Contains(output, "pgedge", "Repository should contain pgedge packages")
		s.T().Log("✓ pgEdge apt repository is configured")
	} else {
		// RHEL/Rocky/Alma - verify dnf repository
		output, exitCode, err := s.ExecCommand("dnf repolist | grep -i pgedge")
		s.NoError(err)
		s.Equal(0, exitCode, "pgEdge repository should be in dnf repos")
		s.Contains(output, "pgedge", "Repository should contain pgedge packages")
		s.T().Log("✓ pgEdge dnf repository is configured")
	}
}
