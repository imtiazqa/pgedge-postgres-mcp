package suite

import (
	"fmt"
	"strings"
	"sync"
)

// Global installation state shared across all test suite instances
// This ensures packages are installed only ONCE per test run, not once per suite
var globalInstallState struct {
	sync.Mutex
	repoInstalled        bool
	postgresqlInstalled  bool
	mcpPackagesInstalled bool
}

// E2ESuite provides End-to-End test suite functionality
type E2ESuite struct {
	BaseSuite
}

// SetupSuite runs once before all tests
func (s *E2ESuite) SetupSuite() {
	// Call base setup
	s.BaseSuite.SetupSuite()

	// E2E-specific setup
	s.T().Log("E2E Suite initialized")
}

// Helper methods for E2E testing

// AssertFileExists asserts that a file exists
func (s *E2ESuite) AssertFileExists(path string) {
	output, exitCode, err := s.ExecCommand(fmt.Sprintf("test -f %s && echo exists || echo missing", path))
	s.NoError(err)
	s.Equal(0, exitCode, "File check command should succeed")
	s.Contains(output, "exists", "File %s should exist", path)
}

// AssertDirectoryExists asserts that a directory exists
func (s *E2ESuite) AssertDirectoryExists(path string) {
	output, exitCode, err := s.ExecCommand(fmt.Sprintf("test -d %s && echo exists || echo missing", path))
	s.NoError(err)
	s.Equal(0, exitCode, "Directory check command should succeed")
	s.Contains(output, "exists", "Directory %s should exist", path)
}

// AssertServiceRunning asserts that a service is running (using systemctl)
func (s *E2ESuite) AssertServiceRunning(serviceName string) {
	output, exitCode, err := s.ExecCommand(fmt.Sprintf("systemctl is-active %s", serviceName))
	s.NoError(err)
	outputTrimmed := strings.TrimSpace(output)
	if exitCode == 0 && outputTrimmed == "active" {
		// Service is running
		return
	}

	// If systemctl doesn't work, try alternative methods
	s.T().Logf("Service %s may not be running or systemctl not available. Exit code: %d, output: %s",
		serviceName, exitCode, output)
}

// AssertCommandSucceeds asserts that a command exits with code 0
func (s *E2ESuite) AssertCommandSucceeds(cmd string) {
	output, exitCode, err := s.ExecCommand(cmd)
	s.NoError(err, "Command should not error: %s", cmd)
	s.Equal(0, exitCode, "Command should exit with code 0: %s\nOutput: %s", cmd, output)
}

// AssertCommandFails asserts that a command exits with non-zero code
func (s *E2ESuite) AssertCommandFails(cmd string) {
	_, exitCode, _ := s.ExecCommand(cmd)
	s.NotEqual(0, exitCode, "Command should exit with non-zero code: %s", cmd)
}

// AssertCommandOutput asserts that a command produces expected output
func (s *E2ESuite) AssertCommandOutput(cmd string, expectedOutput string) {
	output, exitCode, err := s.ExecCommand(cmd)
	s.NoError(err)
	s.Equal(0, exitCode)
	s.Contains(output, expectedOutput, "Command output should contain expected string")
}

// RunCommandWithRetry runs a command with retry logic
func (s *E2ESuite) RunCommandWithRetry(cmd string, maxRetries int) (string, int, error) {
	var output string
	var exitCode int
	var err error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			s.T().Logf("Retry attempt %d/%d for command: %s", attempt, maxRetries, cmd)
		}

		output, exitCode, err = s.ExecCommand(cmd)
		if err == nil && exitCode == 0 {
			return output, exitCode, nil
		}
	}

	return output, exitCode, fmt.Errorf("command failed after %d retries", maxRetries)
}

// ============================================================================
// Installation Helpers - for dynamic dependency installation during tests
//
// These methods provide automatic installation of dependencies during test
// execution. Each method is idempotent (only installs once per test run).
//
// Implementation details are in install.go
// ============================================================================

// EnsureRepositoryInstalled ensures pgEdge repository is installed (runs only once per test run)
func (s *E2ESuite) EnsureRepositoryInstalled() {
	globalInstallState.Lock()
	defer globalInstallState.Unlock()

	if globalInstallState.repoInstalled {
		s.T().Log("pgEdge repository already installed (skipping)")
		return
	}

	s.T().Log("Installing pgEdge repository...")
	s.installRepository()
	globalInstallState.repoInstalled = true
}

// EnsurePostgreSQLInstalled ensures PostgreSQL is installed (runs only once per test run)
func (s *E2ESuite) EnsurePostgreSQLInstalled() {
	globalInstallState.Lock()
	defer globalInstallState.Unlock()

	if globalInstallState.postgresqlInstalled {
		s.T().Log("PostgreSQL already installed (skipping)")
		return
	}

	// Unlock before calling EnsureRepositoryInstalled to avoid deadlock
	globalInstallState.Unlock()
	s.EnsureRepositoryInstalled()
	globalInstallState.Lock()

	s.T().Log("Installing PostgreSQL...")
	s.installPostgreSQL()
	globalInstallState.postgresqlInstalled = true
}

// EnsureMCPPackagesInstalled ensures MCP packages are installed (runs only once per test run)
func (s *E2ESuite) EnsureMCPPackagesInstalled() {
	globalInstallState.Lock()
	defer globalInstallState.Unlock()

	if globalInstallState.mcpPackagesInstalled {
		s.T().Log("MCP packages already installed (skipping)")
		return
	}

	// Unlock before calling EnsurePostgreSQLInstalled to avoid deadlock
	globalInstallState.Unlock()
	s.EnsurePostgreSQLInstalled()
	globalInstallState.Lock()

	s.T().Log("Installing MCP server packages...")
	s.installMCPPackages()
	globalInstallState.mcpPackagesInstalled = true
}

// Helper: Determine OS type
func (s *E2ESuite) isDebianBased() bool {
	output, exitCode, _ := s.ExecCommand("test -f /etc/debian_version && echo 'debian' || echo 'redhat'")
	return exitCode == 0 && strings.TrimSpace(output) == "debian"
}

// Helper: Get PostgreSQL version from config or environment
func (s *E2ESuite) getPostgreSQLVersion() string {
	// Try to get from config first
	if s.Config.PostgreSQL.Version != "" {
		return s.Config.PostgreSQL.Version
	}

	// Default to version 17
	return "17"
}

// ============================================================================
// Package Manager Helper Methods
// ============================================================================

// getPkgManagerUpdate returns the update command for the current OS
func (s *E2ESuite) getPkgManagerUpdate() string {
	if s.isDebianBased() {
		return "DEBIAN_FRONTEND=noninteractive apt-get update"
	}
	return "dnf check-update || true"
}

// getPkgManagerInstall returns the install command for the current OS
func (s *E2ESuite) getPkgManagerInstall(packages ...string) string {
	pkgList := strings.Join(packages, " ")
	if s.isDebianBased() {
		return fmt.Sprintf("DEBIAN_FRONTEND=noninteractive apt-get install -y %s", pkgList)
	}
	return fmt.Sprintf("dnf install -y %s", pkgList)
}

// getPkgManagerSearch returns the search command for the current OS
func (s *E2ESuite) getPkgManagerSearch(packageName string) string {
	if s.isDebianBased() {
		return fmt.Sprintf("apt-cache search %s", packageName)
	}
	return fmt.Sprintf("dnf search %s", packageName)
}

// getRepositoryURL returns the appropriate repository release package URL based on OS and server environment
func (s *E2ESuite) getRepositoryURL() string {
	serverEnv := strings.ToLower(s.Config.Execution.ServerEnv)
	isDebian := s.isDebianBased()

	if isDebian {
		// Debian/Ubuntu repository
		if serverEnv == "staging" {
			if s.Config.Repository.Debian.StagingReleaseURL != "" {
				return s.Config.Repository.Debian.StagingReleaseURL
			}
			// Fallback to default staging URL
			return "https://apt-staging.pgedge.com/repodeb/pgedge-release_latest_all.deb"
		}
		// Default to live
		if s.Config.Repository.Debian.LiveReleaseURL != "" {
			return s.Config.Repository.Debian.LiveReleaseURL
		}
		// Fallback to default live URL
		return "https://apt.pgedge.com/repodeb/pgedge-release_latest_all.deb"
	} else {
		// RHEL/Rocky/Alma repository
		if serverEnv == "staging" {
			if s.Config.Repository.RHEL.StagingReleaseURL != "" {
				return s.Config.Repository.RHEL.StagingReleaseURL
			}
			// Fallback to default staging URL
			return "https://dnf-staging.pgedge.com/reporpm/pgedge-release-latest.noarch.rpm"
		}
		// Default to live
		if s.Config.Repository.RHEL.LiveReleaseURL != "" {
			return s.Config.Repository.RHEL.LiveReleaseURL
		}
		// Fallback to default live URL
		return "https://dnf.pgedge.com/reporpm/pgedge-release-latest.noarch.rpm"
	}
}
