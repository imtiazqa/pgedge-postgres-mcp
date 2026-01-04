package suite

import (
	"fmt"
	"strings"
	"time"
)

// ============================================================================
// E2ESuite Installation Implementations
//
// This file contains the private implementation methods for E2ESuite's
// installation functionality. The public API methods (Ensure*) are in e2e.go.
//
// File Organization:
//   - e2e.go: E2ESuite public API and helper methods
//   - install.go: E2ESuite installation implementations (this file)
//
// Usage:
//   Tests call s.EnsureMCPPackagesInstalled() which chains through:
//   EnsureMCPPackagesInstalled() → installMCPPackages()
//   EnsurePostgreSQLInstalled() → installPostgreSQL()
//   EnsureRepositoryInstalled() → installRepository()
// ============================================================================

// ============================================================================
// Repository Installation
// ============================================================================

// installRepository performs the actual repository installation
func (s *E2ESuite) installRepository() {
	isDebian := s.isDebianBased()

	if isDebian {
		s.installDebianRepository()
	} else {
		s.installRHELRepository()
	}
}

// installDebianRepository installs repository on Debian/Ubuntu systems
func (s *E2ESuite) installDebianRepository() {
	s.T().Log("Installing repository on Debian/Ubuntu system...")

	// Install prerequisites
	commands := []string{
		s.getPkgManagerUpdate(),
		s.getPkgManagerInstall("curl", "gnupg2", "lsb-release"),
	}

	for _, cmd := range commands {
		output, exitCode, err := s.ExecCommand(cmd)
		s.NoError(err, "Command failed: %s\nOutput: %s", cmd, output)
		s.Equal(0, exitCode, "Command exited with non-zero: %s\nOutput: %s", cmd, output)
	}

	// Download and install pgedge-release package
	s.T().Log("Downloading and installing pgedge-release package...")
	repoURL := s.getRepositoryURL()
	s.T().Logf("Using repository URL: %s", repoURL)
	commands = []string{
		fmt.Sprintf("curl -sSL %s -o /tmp/pgedge-release.deb", repoURL),
		"dpkg -i /tmp/pgedge-release.deb",
		"rm -f /tmp/pgedge-release.deb",
	}

	for _, cmd := range commands {
		output, exitCode, err := s.ExecCommand(cmd)
		s.NoError(err, "Command failed: %s\nOutput: %s", cmd, output)
		s.Equal(0, exitCode, "Command exited with non-zero: %s\nOutput: %s", cmd, output)
	}

	// Update package lists
	s.T().Log("Updating package lists...")
	output, exitCode, err := s.ExecCommand(s.getPkgManagerUpdate())
	s.NoError(err, "package manager update failed: %s", output)
	s.Equal(0, exitCode, "package manager update exited with non-zero: %s", output)

	// Verify repository is available
	s.T().Log("Verifying repository is available...")
	output, exitCode, err = s.ExecCommand(s.getPkgManagerSearch("pgedge-postgres-mcp"))
	s.NoError(err)
	s.Equal(0, exitCode)
	s.Contains(output, "pgedge-postgres-mcp", "Package should be available in repo")

	s.T().Log("✓ Debian repository installed successfully")
}

// installRHELRepository installs repository on RHEL/Rocky/Alma systems
func (s *E2ESuite) installRHELRepository() {
	s.T().Log("Installing repository on RHEL/Rocky/Alma system...")

	// Determine EL version
	versionCmd := "rpm -E %{rhel}"
	versionOutput, _, _ := s.ExecCommand(versionCmd)
	elVersion := strings.TrimSpace(versionOutput)
	if elVersion == "" || elVersion == "%{rhel}" {
		elVersion = "9" // Default to EL9
	}

	s.T().Logf("Detected EL version: %s", elVersion)

	// Install EPEL repository first
	s.T().Log("Installing EPEL repository...")
	epelCmd := fmt.Sprintf("dnf -y install https://dl.fedoraproject.org/pub/epel/epel-release-latest-%s.noarch.rpm", elVersion)
	output, exitCode, err := s.ExecCommand(epelCmd)
	s.NoError(err, "EPEL installation failed: %s", output)
	s.Equal(0, exitCode, "EPEL installation exited with non-zero: %s", output)

	// Install pgEdge repository
	s.T().Log("Installing pgEdge repository...")
	repoURL := s.getRepositoryURL()
	s.T().Logf("Using repository URL: %s", repoURL)
	pgedgeCmd := fmt.Sprintf("dnf install -y %s", repoURL)
	output, exitCode, err = s.ExecCommand(pgedgeCmd)
	s.NoError(err, "pgEdge repo installation failed: %s", output)
	s.Equal(0, exitCode, "pgEdge repo installation exited with non-zero: %s", output)

	// Update metadata
	s.T().Log("Updating package manager metadata...")
	_, _, err = s.ExecCommand(s.getPkgManagerUpdate())
	s.NoError(err, "Failed to update package manager metadata")

	// Verify repository is available
	s.T().Log("Verifying repository is available...")
	output, exitCode, err = s.ExecCommand(s.getPkgManagerSearch("pgedge-postgres-mcp"))
	s.NoError(err)
	s.Equal(0, exitCode)
	s.Contains(output, "pgedge-postgres-mcp", "Package should be available in repo")

	s.T().Log("✓ RHEL repository installed successfully")
}

// ============================================================================
// PostgreSQL Installation
// ============================================================================

// installPostgreSQL performs the actual PostgreSQL installation
func (s *E2ESuite) installPostgreSQL() {
	isDebian := s.isDebianBased()
	pgVersion := s.getPostgreSQLVersion()

	s.T().Logf("Installing PostgreSQL %s...", pgVersion)

	// Step 1: Install PostgreSQL packages
	s.T().Logf("Step 1: Installing PostgreSQL %s packages", pgVersion)
	if isDebian {
		s.installPostgreSQLDebian(pgVersion)
	} else {
		s.installPostgreSQLRHEL(pgVersion)
	}

	// Step 2: Initialize and start PostgreSQL
	s.T().Logf("Step 2: Initializing PostgreSQL %s database", pgVersion)
	if isDebian {
		s.initializePostgreSQLDebian(pgVersion)
	} else {
		s.initializePostgreSQLRHEL(pgVersion)
	}

	// Step 3: Set postgres user password
	s.T().Log("Step 3: Setting postgres user password")
	dbPassword := s.Config.Database.Password
	dbUser := s.Config.Database.User
	setPwCmd := fmt.Sprintf(`su - postgres -c "psql -c \"ALTER USER %s WITH PASSWORD '%s';\""`, dbUser, dbPassword)
	output, exitCode, err := s.ExecCommand(setPwCmd)
	s.NoError(err, "Failed to set postgres password: %s", output)
	s.Equal(0, exitCode, "Set password failed: %s", output)

	// Step 4: Create MCP database (skip if using default 'postgres' database)
	s.T().Log("Step 4: Creating MCP database")
	dbName := s.Config.Database.Database

	// Skip database creation if using the default 'postgres' database
	if dbName != "postgres" && dbName != "template0" && dbName != "template1" {
		// First, try to drop the database if it exists
		dropDbCmd := fmt.Sprintf(`su - postgres -c "psql -c \"DROP DATABASE IF EXISTS %s;\""`, dbName)
		s.ExecCommand(dropDbCmd) // Ignore errors

		createDbCmd := fmt.Sprintf(`su - postgres -c "psql -c \"CREATE DATABASE %s;\""`, dbName)
		output, exitCode, err = s.ExecCommand(createDbCmd)
		s.NoError(err, "Failed to create MCP database: %s", output)
		s.Equal(0, exitCode, "Create database failed: %s", output)
		s.T().Logf("✓ Database %s created successfully", dbName)
	} else {
		s.T().Logf("ℹ Using existing system database: %s", dbName)
	}

	s.T().Log("✓ PostgreSQL installed and configured successfully")
}

// installPostgreSQLDebian installs PostgreSQL on Debian/Ubuntu
func (s *E2ESuite) installPostgreSQLDebian(pgVersion string) {
	packageName := fmt.Sprintf("postgresql-%s", pgVersion)
	installCmd := s.getPkgManagerInstall(packageName)
	output, exitCode, err := s.ExecCommand(installCmd)
	s.NoError(err, "PostgreSQL install failed: %s\nOutput: %s", installCmd, output)
	s.Equal(0, exitCode, "PostgreSQL install exited with error: %s\nOutput: %s", installCmd, output)
}

// installPostgreSQLRHEL installs PostgreSQL on RHEL/Rocky/Alma
func (s *E2ESuite) installPostgreSQLRHEL(pgVersion string) {
	serverPkg := fmt.Sprintf("pgedge-postgresql%s-server", pgVersion)
	clientPkg := fmt.Sprintf("pgedge-postgresql%s", pgVersion)

	installCmd := s.getPkgManagerInstall(serverPkg, clientPkg)
	output, exitCode, err := s.ExecCommand(installCmd)
	s.NoError(err, "PostgreSQL install failed: %s\nOutput: %s", installCmd, output)
	s.Equal(0, exitCode, "PostgreSQL install exited with error: %s\nOutput: %s", installCmd, output)
}

// initializePostgreSQLDebian initializes PostgreSQL on Debian/Ubuntu
func (s *E2ESuite) initializePostgreSQLDebian(pgVersion string) {
	// Configure pg_hba.conf for password authentication
	s.T().Log("  Configuring PostgreSQL authentication...")
	hbaFile := fmt.Sprintf("/etc/postgresql/%s/main/pg_hba.conf", pgVersion)

	// Backup original and configure for md5/password authentication
	backupCmd := fmt.Sprintf("cp %s %s.bak", hbaFile, hbaFile)
	s.ExecCommand(backupCmd) // Ignore errors

	// Replace peer/ident with md5 for local connections
	sedCmd := fmt.Sprintf("sed -i 's/local.*all.*all.*peer/local   all             all                                     md5/' %s", hbaFile)
	s.ExecCommand(sedCmd)

	sedCmd2 := fmt.Sprintf("sed -i 's/host.*all.*all.*127.0.0.1.*ident/host    all             all             127.0.0.1\\/32            md5/' %s", hbaFile)
	s.ExecCommand(sedCmd2)

	// Use pg_ctlcluster with --skip-systemctl-redirect to bypass systemd
	startCmd := fmt.Sprintf("pg_ctlcluster --skip-systemctl-redirect %s main start", pgVersion)
	output, exitCode, err := s.ExecCommand(startCmd)
	s.NoError(err, "Failed to start PostgreSQL: %s", output)
	// Exit code 0 = started successfully, Exit code 2 = already running (both are OK)
	s.True(exitCode == 0 || exitCode == 2, "PostgreSQL start failed with unexpected exit code %d: %s", exitCode, output)

	// Reload to apply pg_hba.conf changes
	reloadCmd := fmt.Sprintf("pg_ctlcluster --skip-systemctl-redirect %s main reload", pgVersion)
	s.ExecCommand(reloadCmd)
}

// initializePostgreSQLRHEL initializes PostgreSQL on RHEL/Rocky/Alma
func (s *E2ESuite) initializePostgreSQLRHEL(pgVersion string) {
	// Stop any existing PostgreSQL instances for cleanup
	s.T().Log("  Stopping any existing PostgreSQL instances...")
	for _, ver := range []string{"16", "17", "18"} {
		stopCmd := fmt.Sprintf("su - postgres -c '/usr/pgsql-%s/bin/pg_ctl -D /var/lib/pgsql/%s/data stop' 2>/dev/null || true", ver, ver)
		s.ExecCommand(stopCmd) // Ignore errors
	}

	time.Sleep(2 * time.Second)

	// Remove existing data directory if it exists
	s.T().Log("  Cleaning up existing data directory...")
	cleanupCmd := fmt.Sprintf("rm -rf /var/lib/pgsql/%s/data", pgVersion)
	s.ExecCommand(cleanupCmd)

	// Initialize database
	s.T().Logf("  Initializing new PostgreSQL %s database...", pgVersion)
	initCmd := fmt.Sprintf("su - postgres -c '/usr/pgsql-%s/bin/initdb -D /var/lib/pgsql/%s/data'", pgVersion, pgVersion)
	output, exitCode, err := s.ExecCommand(initCmd)
	s.NoError(err, "PostgreSQL initdb failed: %s", output)
	s.Equal(0, exitCode, "PostgreSQL initdb failed: %s", output)

	// Configure PostgreSQL to accept local connections
	configCmd := fmt.Sprintf(`echo "host all all 127.0.0.1/32 md5" >> /var/lib/pgsql/%s/data/pg_hba.conf`, pgVersion)
	output, exitCode, err = s.ExecCommand(configCmd)
	s.NoError(err, "Failed to configure pg_hba.conf: %s", output)
	s.Equal(0, exitCode, "pg_hba.conf config failed: %s", output)

	// Start PostgreSQL manually
	s.T().Logf("  Starting PostgreSQL %s...", pgVersion)
	startCmd := fmt.Sprintf("su - postgres -c '/usr/pgsql-%s/bin/pg_ctl -D /var/lib/pgsql/%s/data -l /var/lib/pgsql/%s/data/logfile start'", pgVersion, pgVersion, pgVersion)
	output, exitCode, err = s.ExecCommand(startCmd)
	s.NoError(err, "PostgreSQL start failed: %s", output)
	s.Equal(0, exitCode, "PostgreSQL start failed: %s", output)

	// Wait for PostgreSQL to be ready
	time.Sleep(3 * time.Second)
}

// ============================================================================
// MCP Packages Installation
// ============================================================================

// installMCPPackages performs the actual MCP package installation
func (s *E2ESuite) installMCPPackages() {
	s.T().Log("Installing MCP server packages...")

	// Get all packages from config and install them
	var packageLists = [][]string{
		s.Config.Packages.MCPServer,
		s.Config.Packages.CLI,
		s.Config.Packages.Web,
		s.Config.Packages.KB,
	}

	for _, pkgList := range packageLists {
		for _, pkg := range pkgList {
			installCmd := s.getPkgManagerInstall(pkg)
			output, exitCode, err := s.ExecCommand(installCmd)
			s.NoError(err, "Package install failed: %s\nOutput: %s", installCmd, output)
			s.Equal(0, exitCode, "Package install exited with error: %s\nOutput: %s", installCmd, output)
		}
	}

	// Configure database password in env file
	s.T().Log("Configuring database password...")
	envFile := fmt.Sprintf("%s/postgres-mcp.env", s.Config.ConfigDir)
	password := s.Config.Database.Password

	configCmd := fmt.Sprintf(`grep -q "^PGEDGE_DB_PASSWORD=" %s 2>/dev/null && \
		sed -i 's/^PGEDGE_DB_PASSWORD=.*/PGEDGE_DB_PASSWORD=%s/' %s || \
		echo 'PGEDGE_DB_PASSWORD=%s' >> %s`, envFile, password, envFile, password, envFile)

	output, exitCode, err := s.ExecCommand(configCmd)
	s.NoError(err, "Failed to set database password: %s", output)
	s.Equal(0, exitCode, "Set database password failed: %s", output)

	s.T().Log("✓ All MCP server packages installed and configured successfully")
}
