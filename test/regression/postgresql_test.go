package regression

import (
	"time"
)

// ========================================================================
// TEST 02: PostgreSQL Installation and Setup
// ========================================================================

// Test02_PostgreSQLSetup tests PostgreSQL installation and configuration
func (s *RegressionTestSuite) Test02_PostgreSQLSetup() {
	s.T().Log("TEST 02: Installing and configuring pgEdge PostgreSQL")
	s.ensurePostgreSQLInstalled()
	s.T().Log("✓ PostgreSQL installed and configured successfully")
}

// installPostgreSQL performs the actual PostgreSQL installation
func (s *RegressionTestSuite) installPostgreSQL() {

	isDebian, _ := s.getOSType()

	// Step 1: Install PostgreSQL packages
	s.logDetailed("Step 1: Installing PostgreSQL packages")
	var pgPackages []string
	if isDebian {
		pgPackages = []string{
			"apt-get install -y postgresql-18",
		}
	} else {
		pgPackages = []string{
			"dnf install -y pgedge-postgresql18-server",
			"dnf install -y pgedge-postgresql18",
		}
	}

	for _, installCmd := range pgPackages {
		output, exitCode, err := s.execCmd(s.ctx, installCmd)
		s.NoError(err, "PostgreSQL install failed: %s\nOutput: %s", installCmd, output)
		s.Equal(0, exitCode, "PostgreSQL install exited with error: %s\nOutput: %s", installCmd, output)
	}

	// Step 2: Initialize PostgreSQL database
	s.logDetailed("Step 2: Initializing PostgreSQL database")
	if isDebian {
		// Debian/Ubuntu initialization
		// Use pg_ctlcluster with --skip-systemctl-redirect to bypass systemd
		// This avoids the systemd PID file ownership issue on Ubuntu/Debian
		output, exitCode, err := s.execCmd(s.ctx, "pg_ctlcluster --skip-systemctl-redirect 18 main start")
		s.NoError(err, "Failed to start PostgreSQL: %s", output)
		// Exit code 0 = started successfully, Exit code 2 = already running (both are OK)
		s.True(exitCode == 0 || exitCode == 2, "PostgreSQL start failed with unexpected exit code %d: %s", exitCode, output)
	} else {
		// RHEL/Rocky initialization (manual, not systemd)

		// Stop any existing PostgreSQL instance
		s.logDetailed("  Stopping any existing PostgreSQL instances...")
		stopCmd := "su - postgres -c '/usr/pgsql-18/bin/pg_ctl -D /var/lib/pgsql/18/data stop' 2>/dev/null || true"
		s.execCmd(s.ctx, stopCmd) // Ignore errors

		// Also try older version for cleanup
		stopCmd16 := "su - postgres -c '/usr/pgsql-16/bin/pg_ctl -D /var/lib/pgsql/16/data stop' 2>/dev/null || true"
		s.execCmd(s.ctx, stopCmd16) // Ignore errors

		time.Sleep(2 * time.Second)

		// Remove existing data directory if it exists
		s.logDetailed("  Cleaning up existing data directory...")
		cleanupCmd := "rm -rf /var/lib/pgsql/18/data"
		s.execCmd(s.ctx, cleanupCmd)

		// Initialize database
		s.logDetailed("  Initializing new PostgreSQL 18 database...")
		initCmd := "su - postgres -c '/usr/pgsql-18/bin/initdb -D /var/lib/pgsql/18/data'"
		output, exitCode, err := s.execCmd(s.ctx, initCmd)
		s.NoError(err, "PostgreSQL initdb failed: %s", output)
		s.Equal(0, exitCode, "PostgreSQL initdb failed: %s", output)

		// Configure PostgreSQL to accept local connections
		configCmd := `echo "host all all 127.0.0.1/32 md5" >> /var/lib/pgsql/18/data/pg_hba.conf`
		output, exitCode, err = s.execCmd(s.ctx, configCmd)
		s.NoError(err, "Failed to configure pg_hba.conf: %s", output)
		s.Equal(0, exitCode, "pg_hba.conf config failed: %s", output)

		// Start PostgreSQL manually
		s.logDetailed("  Starting PostgreSQL 18...")
		startCmd := "su - postgres -c '/usr/pgsql-18/bin/pg_ctl -D /var/lib/pgsql/18/data -l /var/lib/pgsql/18/data/logfile start'"
		output, exitCode, err = s.execCmd(s.ctx, startCmd)
		s.NoError(err, "PostgreSQL start failed: %s", output)
		s.Equal(0, exitCode, "PostgreSQL start failed: %s", output)

		// Wait for PostgreSQL to be ready
		time.Sleep(3 * time.Second)
	}

	// Step 3: Set postgres user password
	s.logDetailed("Step 3: Setting postgres user password")
	setPwCmd := `su - postgres -c "psql -c \"ALTER USER postgres WITH PASSWORD 'postgres123';\""`
	output, exitCode, err := s.execCmd(s.ctx, setPwCmd)
	s.NoError(err, "Failed to set postgres password: %s", output)
	s.Equal(0, exitCode, "Set password failed: %s", output)

	// Step 4: Create MCP database
	s.logDetailed("Step 4: Creating MCP database")
	// First, try to drop the database if it exists
	dropDbCmd := `su - postgres -c "psql -c \"DROP DATABASE IF EXISTS mcp_server;\""`
	s.execCmd(s.ctx, dropDbCmd) // Ignore errors

	createDbCmd := `su - postgres -c "psql -c \"CREATE DATABASE mcp_server;\""`
	output, exitCode, err = s.execCmd(s.ctx, createDbCmd)
	s.NoError(err, "Failed to create MCP database: %s", output)
	s.Equal(0, exitCode, "Create database failed: %s", output)
}
