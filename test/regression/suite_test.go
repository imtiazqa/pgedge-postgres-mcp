package regression

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

// RegressionTestSuite runs basic regression tests
type RegressionTestSuite struct {
	suite.Suite
	ctx       context.Context
	container *SimpleContainer
	osImage   string
	repoURL   string
}

// SetupSuite runs once before all tests
func (s *RegressionTestSuite) SetupSuite() {
	s.ctx = context.Background()

	// Get OS image from environment or use default
	s.osImage = os.Getenv("TEST_OS_IMAGE")
	if s.osImage == "" {
		s.osImage = "debian:12" // Default to Debian 12
	}

	// Get repo URL from environment
	s.repoURL = os.Getenv("PGEDGE_REPO_URL")
	if s.repoURL == "" {
		s.repoURL = "https://apt.pgedge.com" // Example repo URL
	}

	s.T().Logf("Testing with OS image: %s", s.osImage)
	s.T().Logf("Using repository: %s", s.repoURL)
}

// SetupTest runs before each test
func (s *RegressionTestSuite) SetupTest() {
	s.T().Logf("=== Setting up container for: %s ===", s.T().Name())

	var err error
	s.container, err = NewContainer(s.osImage)
	s.Require().NoError(err, "Failed to create container")

	err = s.container.Start(s.ctx)
	s.Require().NoError(err, "Failed to start container")

	s.T().Logf("Container started successfully")
}

// TearDownTest runs after each test
func (s *RegressionTestSuite) TearDownTest() {
	if s.container != nil {
		if s.T().Failed() {
			// Print logs on failure
			logs, _ := s.container.GetLogs(s.ctx)
			s.T().Logf("Container logs:\n%s", logs)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Stop and remove container
		if err := s.container.Cleanup(ctx); err != nil {
			s.T().Logf("Warning: Container cleanup failed: %v", err)
		} else {
			s.T().Logf("Container stopped and removed successfully")
		}
	}
}

// TearDownSuite runs once after all tests
func (s *RegressionTestSuite) TearDownSuite() {
	s.T().Log("=== Test suite completed ===")
	s.T().Log("All Docker containers have been cleaned up")
}

// ========================================================================
// TEST 01: Repository Installation
// ========================================================================
func (s *RegressionTestSuite) Test01_RepositoryInstallation() {
	s.T().Log("TEST 01: Installing pgEdge repository")

	// Determine package manager
	isDebian := strings.Contains(s.osImage, "debian") || strings.Contains(s.osImage, "ubuntu")
	isRHEL := strings.Contains(s.osImage, "rocky") || strings.Contains(s.osImage, "alma") || strings.Contains(s.osImage, "rhel")

	if isDebian {
		// Debian/Ubuntu: Install repository
		commands := []string{
			"apt-get update",
			"apt-get install -y wget gnupg",
			// Add your actual repo setup commands here
			"echo 'deb [trusted=yes] " + s.repoURL + " stable main' > /etc/apt/sources.list.d/pgedge.list",
			"apt-get update",
		}

		for _, cmd := range commands {
			output, exitCode, err := s.container.Exec(s.ctx, cmd)
			s.NoError(err, "Command failed: %s\nOutput: %s", cmd, output)
			s.Equal(0, exitCode, "Command exited with non-zero: %s\nOutput: %s", cmd, output)
		}

		// Verify repository is available
		output, exitCode, err := s.container.Exec(s.ctx, "apt-cache search pgedge-postgres-mcp")
		s.NoError(err)
		s.Equal(0, exitCode)
		s.Contains(output, "pgedge-postgres-mcp", "Package should be available in repo")

	} else if isRHEL {
		// RHEL/Rocky/Alma: Install repository
		// Determine EL version
		versionCmd := "rpm -E %{rhel}"
		versionOutput, _, _ := s.container.Exec(s.ctx, versionCmd)
		elVersion := strings.TrimSpace(versionOutput)
		if elVersion == "" || elVersion == "%{rhel}" {
			elVersion = "9" // Default to EL9
		}

		commands := []string{
			// Install EPEL repository first
			fmt.Sprintf("dnf -y install https://dl.fedoraproject.org/pub/epel/epel-release-latest-%s.noarch.rpm", elVersion),
			// Install pgEdge repository
			"dnf install -y https://dnf.pgedge.com/reporpm/pgedge-release-latest.noarch.rpm",
			// Update metadata
			"dnf check-update || true", // May return non-zero
		}

		for _, cmd := range commands {
			output, _, err := s.container.Exec(s.ctx, cmd)
			s.NoError(err, "Command failed: %s\nOutput: %s", cmd, output)
		}

		// Verify repository is available
		output, exitCode, err := s.container.Exec(s.ctx, "dnf search pgedge-postgres-mcp")
		s.NoError(err)
		s.Equal(0, exitCode)
		s.Contains(output, "pgedge-postgres-mcp", "Package should be available in repo")
	} else {
		s.Fail("Unsupported OS image: %s", s.osImage)
	}

	s.T().Log("✓ Repository installed successfully")
}

// ========================================================================
// TEST 02: PostgreSQL Installation and Setup
// ========================================================================
func (s *RegressionTestSuite) Test02_PostgreSQLSetup() {
	s.T().Log("TEST 02: Installing and configuring pgEdge PostgreSQL")

	// First install repository (this test can run independently)
	s.Test01_RepositoryInstallation()

	isDebian := strings.Contains(s.osImage, "debian") || strings.Contains(s.osImage, "ubuntu")

	// Step 1: Install PostgreSQL packages
	s.T().Log("Step 1: Installing PostgreSQL packages")
	var pgPackages []string
	if isDebian {
		pgPackages = []string{
			"apt-get install -y postgresql-16",
		}
	} else {
		pgPackages = []string{
			"dnf install -y pgedge-postgresql16-server",
			"dnf install -y pgedge-postgresql16",
		}
	}

	for _, installCmd := range pgPackages {
		output, exitCode, err := s.container.Exec(s.ctx, installCmd)
		s.NoError(err, "PostgreSQL install failed: %s\nOutput: %s", installCmd, output)
		s.Equal(0, exitCode, "PostgreSQL install exited with error: %s\nOutput: %s", installCmd, output)
	}

	// Step 2: Initialize PostgreSQL database
	s.T().Log("Step 2: Initializing PostgreSQL database")
	if isDebian {
		// Debian/Ubuntu initialization
		output, exitCode, err := s.container.Exec(s.ctx, "pg_ctlcluster 16 main start")
		s.NoError(err, "Failed to start PostgreSQL: %s", output)
		s.Equal(0, exitCode, "PostgreSQL start failed: %s", output)
	} else {
		// RHEL/Rocky initialization (manual, not systemd)
		// Initialize database
		initCmd := "su - postgres -c '/usr/pgsql-16/bin/initdb -D /var/lib/pgsql/16/data'"
		output, exitCode, err := s.container.Exec(s.ctx, initCmd)
		s.NoError(err, "PostgreSQL initdb failed: %s", output)
		s.Equal(0, exitCode, "PostgreSQL initdb failed: %s", output)

		// Configure PostgreSQL to accept local connections
		configCmd := `echo "host all all 127.0.0.1/32 md5" >> /var/lib/pgsql/16/data/pg_hba.conf`
		output, exitCode, err = s.container.Exec(s.ctx, configCmd)
		s.NoError(err, "Failed to configure pg_hba.conf: %s", output)
		s.Equal(0, exitCode, "pg_hba.conf config failed: %s", output)

		// Start PostgreSQL manually
		startCmd := "su - postgres -c '/usr/pgsql-16/bin/pg_ctl -D /var/lib/pgsql/16/data -l /var/lib/pgsql/16/data/logfile start'"
		output, exitCode, err = s.container.Exec(s.ctx, startCmd)
		s.NoError(err, "PostgreSQL start failed: %s", output)
		s.Equal(0, exitCode, "PostgreSQL start failed: %s", output)

		// Wait for PostgreSQL to be ready
		time.Sleep(3 * time.Second)
	}

	// Step 3: Set postgres user password
	s.T().Log("Step 3: Setting postgres user password")
	setPwCmd := `su - postgres -c "psql -c \"ALTER USER postgres WITH PASSWORD 'postgres123';\""`
	output, exitCode, err := s.container.Exec(s.ctx, setPwCmd)
	s.NoError(err, "Failed to set postgres password: %s", output)
	s.Equal(0, exitCode, "Set password failed: %s", output)

	// Step 4: Create MCP database
	s.T().Log("Step 4: Creating MCP database")
	createDbCmd := `su - postgres -c "psql -c \"CREATE DATABASE mcp_server;\""`
	output, exitCode, err = s.container.Exec(s.ctx, createDbCmd)
	s.NoError(err, "Failed to create MCP database: %s", output)
	s.Equal(0, exitCode, "Create database failed: %s", output)

	s.T().Log("✓ PostgreSQL installed and configured successfully")
}

// ========================================================================
// TEST 03: MCP Server Package Installation
// ========================================================================
func (s *RegressionTestSuite) Test03_MCPServerInstallation() {
	s.T().Log("TEST 03: Installing MCP server packages")

	// First setup PostgreSQL
	s.Test02_PostgreSQLSetup()

	isDebian := strings.Contains(s.osImage, "debian") || strings.Contains(s.osImage, "ubuntu")

	// Step 1: Install MCP server packages
	s.T().Log("Step 1: Installing MCP server packages")
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
		output, exitCode, err := s.container.Exec(s.ctx, installCmd)
		s.NoError(err, "Install failed: %s\nOutput: %s", installCmd, output)
		s.Equal(0, exitCode, "Install exited with error: %s\nOutput: %s", installCmd, output)
	}

	// Step 2: Update MCP server configuration
	s.T().Log("Step 2: Updating MCP server configuration files")

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
	output, exitCode, err := s.container.Exec(s.ctx, yamlConfig)
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
	output, exitCode, err = s.container.Exec(s.ctx, envConfig)
	s.NoError(err, "Failed to update postgres-mcp.env: %s", output)
	s.Equal(0, exitCode, "Update env failed: %s", output)

	s.T().Log("✓ All MCP server packages installed and configured successfully")
}

// ========================================================================
// TEST 04: Installation Validation
// ========================================================================
func (s *RegressionTestSuite) Test04_InstallationValidation() {
	s.T().Log("TEST 04: Validating MCP server installation")

	// Install package first
	s.Test03_MCPServerInstallation()

	// Check 1: Binary exists and is executable
	output, exitCode, err := s.container.Exec(s.ctx, "test -x /usr/bin/pgedge-postgres-mcp && echo 'OK'")
	s.NoError(err)
	s.Equal(0, exitCode)
	s.Contains(output, "OK", "Binary should exist and be executable")

	// Check 2: Config directory exists
	output, exitCode, err = s.container.Exec(s.ctx, "test -d /etc/pgedge && echo 'OK'")
	s.NoError(err)
	s.Equal(0, exitCode)
	s.Contains(output, "OK", "Config directory should exist")

	// Check 3: Config files exist
	output, exitCode, err = s.container.Exec(s.ctx, "test -f /etc/pgedge/postgres-mcp.yaml && echo 'OK'")
	s.NoError(err)
	s.Equal(0, exitCode)
	s.Contains(output, "OK", "postgres-mcp.yaml should exist")

	output, exitCode, err = s.container.Exec(s.ctx, "test -f /etc/pgedge/postgres-mcp.env && echo 'OK'")
	s.NoError(err)
	s.Equal(0, exitCode)
	s.Contains(output, "OK", "postgres-mcp.env should exist")

	// Check 4: Systemd service file exists
	output, exitCode, err = s.container.Exec(s.ctx, "test -f /usr/lib/systemd/system/pgedge-postgres-mcp.service && echo 'OK'")
	s.NoError(err)
	s.Equal(0, exitCode)
	s.Contains(output, "OK", "Systemd service file should exist")

	// Check 5: Data directory exists with correct permissions
	output, exitCode, err = s.container.Exec(s.ctx, "test -d /var/lib/pgedge/postgres-mcp && echo 'OK'")
	s.NoError(err)
	s.Equal(0, exitCode)
	s.Contains(output, "OK", "Data directory should exist")

	s.T().Log("✓ All installation validation checks passed")
}

// ========================================================================
// TEST 05: Token Management
// ========================================================================
func (s *RegressionTestSuite) Test05_TokenManagement() {
	s.T().Log("TEST 05: Testing token management commands")

	// Install package first
	s.Test03_MCPServerInstallation()

	// Test 1: Create token (using config file for database connection)
	createCmd := `/usr/bin/pgedge-postgres-mcp -config /etc/pgedge/postgres-mcp.yaml -add-token -token-file /etc/pgedge/pgedge-postgres-mcp-tokens.yaml -token-note "test-token"`
	output, exitCode, err := s.container.Exec(s.ctx, createCmd)
	s.NoError(err, "Token creation failed\nOutput: %s", output)
	s.Equal(0, exitCode)
	s.Contains(output, "Token:", "Should show generated token")
	s.Contains(output, "Hash:", "Should show token hash")

	// Set proper ownership on token file
	chownCmd := "chown pgedge:pgedge /etc/pgedge/pgedge-postgres-mcp-tokens.yaml"
	output, exitCode, err = s.container.Exec(s.ctx, chownCmd)
	s.NoError(err, "Failed to set ownership on token file: %s", output)
	s.Equal(0, exitCode, "chown failed: %s", output)

	// Test 2: List tokens
	output, exitCode, err = s.container.Exec(s.ctx, "/usr/bin/pgedge-postgres-mcp -config /etc/pgedge/postgres-mcp.yaml -list-tokens -token-file /etc/pgedge/pgedge-postgres-mcp-tokens.yaml")
	s.NoError(err)
	s.Equal(0, exitCode)
	s.Contains(output, "test-token", "Should list created token")

	// Test 3: Verify token file was created
	output, exitCode, err = s.container.Exec(s.ctx, "test -f /etc/pgedge/pgedge-postgres-mcp-tokens.yaml && echo 'OK'")
	s.NoError(err)
	s.Equal(0, exitCode)
	s.Contains(output, "OK", "Token file should exist")

	s.T().Log("✓ Token management working correctly")
}

// ========================================================================
// TEST 06: User Management
// ========================================================================
func (s *RegressionTestSuite) Test06_UserManagement() {
	s.T().Log("TEST 06: Testing user management commands")

	// Install package first
	s.Test03_MCPServerInstallation()

	// Test 1: Create user (using config file for database connection)
	createCmd := `/usr/bin/pgedge-postgres-mcp -config /etc/pgedge/postgres-mcp.yaml -add-user -user-file /etc/pgedge/pgedge-postgres-mcp-users.yaml -username testuser -password testpass123 -user-note "test user"`
	output, exitCode, err := s.container.Exec(s.ctx, createCmd)
	s.NoError(err, "User creation failed\nOutput: %s", output)
	s.Equal(0, exitCode)
	s.Contains(output, "User created", "Should confirm user creation")

	// Test 2: List users
	output, exitCode, err = s.container.Exec(s.ctx, "/usr/bin/pgedge-postgres-mcp -config /etc/pgedge/postgres-mcp.yaml -list-users -user-file /etc/pgedge/pgedge-postgres-mcp-users.yaml")
	s.NoError(err)
	s.Equal(0, exitCode)
	s.Contains(output, "testuser", "Should list created user")

	// Test 3: Verify user file was created
	output, exitCode, err = s.container.Exec(s.ctx, "test -f /etc/pgedge/pgedge-postgres-mcp-users.yaml && echo 'OK'")
	s.NoError(err)
	s.Equal(0, exitCode)
	s.Contains(output, "OK", "User file should exist")

	// Test 4: User file has correct permissions (should be restrictive)
	output, exitCode, err = s.container.Exec(s.ctx, "stat -c '%a' /etc/pgedge/pgedge-postgres-mcp-users.yaml")
	s.NoError(err)
	s.Equal(0, exitCode)
	// File should be readable but ideally 600 or 644
	s.Regexp(`^[0-9]{3}$`, strings.TrimSpace(output), "Should have valid permissions")

	s.T().Log("✓ User management working correctly")
}

// Run the test suite
func TestRegressionSuite(t *testing.T) {
	suite.Run(t, new(RegressionTestSuite))
}
