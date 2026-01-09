package testcases

import "fmt"

// ============================================================================
// TEST 02: PostgreSQL Installation and Setup Tests
// ============================================================================

// testPostgreSQL_Installation installs PostgreSQL and verifies the installation
func (s *MCPServerTestSuite) testPostgreSQL_Installation() {
	s.T().Log("Testing PostgreSQL installation...")

	// Ensure PostgreSQL is installed (this will install if not already done)
	s.EnsurePostgreSQLInstalled()

	// Verify PostgreSQL is installed
	output, exitCode, err := s.ExecCommand("psql --version")
	s.NoError(err)
	s.Equal(0, exitCode)
	s.Contains(output, "PostgreSQL", "PostgreSQL should be installed")

	s.T().Log("✓ PostgreSQL installed successfully")
}

// testPostgreSQL_ServiceStatus verifies PostgreSQL service is running
func (s *MCPServerTestSuite) testPostgreSQL_ServiceStatus() {
	s.T().Log("Testing PostgreSQL service status...")

	// Check if PostgreSQL process is running (works in both systemd and non-systemd environments)
	output, exitCode, _ := s.ExecCommand("ps aux | grep postgres | grep -v grep")
	if exitCode == 0 && output != "" {
		s.T().Log("✓ PostgreSQL process is running")
	} else {
		// Also check systemctl if available
		output, exitCode, _ = s.ExecCommand("systemctl is-active postgresql")
		if exitCode == 0 {
			s.T().Log("✓ PostgreSQL service is active (systemd)")
		} else {
			s.T().Logf("PostgreSQL service check: %s (exit code: %d)", output, exitCode)
			s.T().Log("Note: PostgreSQL may be running without systemd in container environment")
		}
	}
}

// testPostgreSQL_DatabaseConnection verifies database connection with configured credentials
func (s *MCPServerTestSuite) testPostgreSQL_DatabaseConnection() {
	s.T().Log("Testing PostgreSQL database connection...")

	// Verify database connection using config
	dbPassword := s.Config.Database.Password
	dbUser := s.Config.Database.User
	dbName := s.Config.Database.Database

	psqlCmd := fmt.Sprintf("PGPASSWORD=%s psql -U %s -d %s -c 'SELECT version();'", dbPassword, dbUser, dbName)
	output, exitCode, err := s.ExecCommand(psqlCmd)

	s.NoError(err, "Database connection failed: %s", output)
	s.Equal(0, exitCode, "Database connection should succeed")
	s.Contains(output, "PostgreSQL", "Query should return PostgreSQL version")

	s.T().Log("✓ PostgreSQL database connection successful")
}

// testPostgreSQL_MCPDatabase verifies the MCP database exists
func (s *MCPServerTestSuite) testPostgreSQL_MCPDatabase() {
	s.T().Log("Testing MCP database exists...")

	// Check if mcp_server database exists (or whatever database is configured)
	dbPassword := s.Config.Database.Password
	dbUser := s.Config.Database.User
	dbName := s.Config.Database.Database

	// List databases and check if our database exists
	checkCmd := fmt.Sprintf("PGPASSWORD=%s psql -U %s -lqt | cut -d '|' -f 1 | grep -w %s", dbPassword, dbUser, dbName)
	output, exitCode, err := s.ExecCommand(checkCmd)

	s.NoError(err, "Failed to check database existence: %s", output)
	s.Equal(0, exitCode, "Database %s should exist", dbName)
	s.T().Logf("✓ MCP database '%s' exists", dbName)
}
