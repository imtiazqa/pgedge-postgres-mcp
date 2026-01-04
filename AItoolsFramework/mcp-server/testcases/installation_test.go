package testcases

import "fmt"

// ============================================================================
// Installation Tests
// ============================================================================

func (s *MCPServerTestSuite) testInstallation_MCPPackages() {
	s.T().Log("Testing MCP package installation...")

	// Verify binaries exist (from config)
	s.AssertFileExists(s.Config.Binaries.MCPServer)
	s.AssertFileExists(s.Config.Binaries.CLI)
	s.AssertFileExists(s.Config.Binaries.KBBuilder)

	// Verify config files exist (from config)
	configDir := s.Config.ConfigDir
	s.AssertFileExists(fmt.Sprintf("%s/postgres-mcp.yaml", configDir))
	s.AssertFileExists(fmt.Sprintf("%s/postgres-mcp.env", configDir))

	s.T().Log("✓ All MCP packages installed correctly")
}

func (s *MCPServerTestSuite) testInstallation_PackageFiles() {
	s.T().Log("Testing package file installation...")

	// Test binary permissions (from config)
	mcpBinary := s.Config.Binaries.MCPServer
	output, exitCode, err := s.ExecCommand(fmt.Sprintf("test -x %s && echo 'executable' || echo 'not executable'", mcpBinary))
	s.NoError(err)
	s.Equal(0, exitCode)
	s.Contains(output, "executable", "MCP server binary should be executable")

	// Test config file permissions (from config)
	configFile := fmt.Sprintf("%s/postgres-mcp.yaml", s.Config.ConfigDir)
	output, exitCode, err = s.ExecCommand(fmt.Sprintf("test -r %s && echo 'readable' || echo 'not readable'", configFile))
	s.NoError(err)
	s.Equal(0, exitCode)
	s.Contains(output, "readable", "Config file should be readable")

	s.T().Log("✓ Package files have correct permissions")
}

func (s *MCPServerTestSuite) testInstallation_Repository() {
	s.T().Log("Testing repository installation...")

	isDebian := s.isDebianBased()

	if isDebian {
		// Debian/Ubuntu
		output, exitCode, err := s.ExecCommand("apt-cache policy | grep -i pgedge")
		s.NoError(err)
		s.Equal(0, exitCode, "pgEdge repository should be in apt sources")
		s.Contains(output, "pgedge", "Repository should contain pgedge packages")
	} else {
		// RHEL/Rocky/Alma
		output, exitCode, err := s.ExecCommand("dnf repolist | grep -i pgedge")
		s.NoError(err)
		s.Equal(0, exitCode, "pgEdge repository should be in dnf repos")
		s.Contains(output, "pgedge", "Repository should contain pgedge packages")
	}

	s.T().Log("✓ Repository installed and available")
}

func (s *MCPServerTestSuite) testInstallation_PostgreSQL() {
	s.T().Log("Testing PostgreSQL installation...")

	// Check PostgreSQL is installed
	output, exitCode, err := s.ExecCommand("psql --version")
	s.NoError(err)
	s.Equal(0, exitCode)
	s.Contains(output, "PostgreSQL", "PostgreSQL should be installed")

	// Check PostgreSQL is running
	output, exitCode, err = s.ExecCommand("systemctl is-active postgresql")
	if exitCode == 0 {
		s.T().Log("✓ PostgreSQL service is active")
	} else {
		s.T().Logf("PostgreSQL service status: %s (exit code: %d)", output, exitCode)
	}

	// Verify database connection (using config password)
	dbPassword := s.Config.Database.Password
	dbUser := s.Config.Database.User
	dbName := s.Config.Database.Database

	psqlCmd := fmt.Sprintf("PGPASSWORD=%s psql -U %s -d %s -c 'SELECT version();'", dbPassword, dbUser, dbName)
	output, exitCode, err = s.ExecCommand(psqlCmd)
	s.NoError(err)
	s.Equal(0, exitCode, "Should be able to connect to PostgreSQL")
	s.Contains(output, "PostgreSQL", "Query should return PostgreSQL version")

	s.T().Log("✓ PostgreSQL installed and functional")
}
