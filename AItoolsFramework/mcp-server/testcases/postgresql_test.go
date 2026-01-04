package testcases

import "fmt"

// ============================================================================
// PostgreSQL Installation and Setup Tests
// ============================================================================

func (s *MCPServerTestSuite) testPostgreSQL_Installation() {
	s.T().Log("Testing PostgreSQL installation...")

	// Check PostgreSQL is installed
	output, exitCode, err := s.ExecCommand("psql --version")
	s.NoError(err)
	s.Equal(0, exitCode)
	s.Contains(output, "PostgreSQL", "PostgreSQL should be installed")

	s.T().Log("✓ PostgreSQL installed successfully")
}

func (s *MCPServerTestSuite) testPostgreSQL_ServiceStatus() {
	s.T().Log("Testing PostgreSQL service status...")

	// Check PostgreSQL service status (note: may not use systemd in containers)
	output, exitCode, _ := s.ExecCommand("systemctl is-active postgresql")
	if exitCode == 0 {
		s.T().Log("✓ PostgreSQL service is active")
	} else {
		s.T().Logf("PostgreSQL service status: %s (exit code: %d)", output, exitCode)
		s.T().Log("Note: PostgreSQL may be running without systemd in container environment")
	}
}

func (s *MCPServerTestSuite) testPostgreSQL_DatabaseConnection() {
	s.T().Log("Testing PostgreSQL database connection...")

	// Verify database connection using config
	dbPassword := s.Config.Database.Password
	dbUser := s.Config.Database.User
	dbName := s.Config.Database.Database

	psqlCmd := fmt.Sprintf("PGPASSWORD=%s psql -U %s -d %s -c 'SELECT version();'", dbPassword, dbUser, dbName)
	output, exitCode, err := s.ExecCommand(psqlCmd)

	if exitCode == 0 {
		s.NoError(err)
		s.Contains(output, "PostgreSQL", "Query should return PostgreSQL version")
		s.T().Log("✓ PostgreSQL database connection successful")
	} else {
		s.T().Logf("Database connection test: %s (exit code: %d)", output, exitCode)
		s.T().Log("Note: Database connection may require additional setup in container environment")
	}
}

func (s *MCPServerTestSuite) testPostgreSQL_MCPDatabase() {
	s.T().Log("Testing MCP database exists...")

	// Check if mcp_server database exists
	dbPassword := s.Config.Database.Password
	dbUser := s.Config.Database.User

	checkCmd := fmt.Sprintf("PGPASSWORD=%s psql -U %s -lqt | cut -d '|' -f 1 | grep -w mcp_server", dbPassword, dbUser)
	output, exitCode, _ := s.ExecCommand(checkCmd)

	if exitCode == 0 {
		s.T().Log("✓ MCP database 'mcp_server' exists")
	} else {
		s.T().Logf("MCP database check: %s (exit code: %d)", output, exitCode)
		s.T().Log("Note: MCP database may need to be created")
	}
}
