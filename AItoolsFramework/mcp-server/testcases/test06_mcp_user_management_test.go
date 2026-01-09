package testcases

import (
	"fmt"
	"strings"
)

// ============================================================================
// User Management Tests
// ============================================================================

func (s *MCPServerTestSuite) testUser_CreateUser() {
	s.T().Log("Testing user creation...")

	// Ensure MCP packages are installed (this will install if not already done)
	s.EnsureMCPPackagesInstalled()

	// Create a test user using config paths
	mcpBinary := s.Config.Binaries.MCPServer
	configFile := fmt.Sprintf("%s/postgres-mcp.yaml", s.Config.ConfigDir)
	userFile := fmt.Sprintf("%s/pgedge-postgres-mcp-users.yaml", s.Config.ConfigDir)

	// Clean up any existing test user from previous runs
	s.T().Log("Cleaning up any existing test user...")
	deleteCmd := fmt.Sprintf("%s -config %s -delete-user -user-file %s -username testuser 2>/dev/null || true",
		mcpBinary, configFile, userFile)
	s.ExecCommand(deleteCmd)

	createCmd := fmt.Sprintf("%s -config %s -add-user -user-file %s -username testuser -password testpass123 -user-note \"test user\"",
		mcpBinary, configFile, userFile)

	output, exitCode, err := s.ExecCommand(createCmd)
	s.NoError(err, "User creation failed\nOutput: %s", output)
	s.Equal(0, exitCode)
	s.Contains(output, "User created", "Should confirm user creation")

	s.T().Log("✓ User created successfully")
}

func (s *MCPServerTestSuite) testUser_ListUsers() {
	s.T().Log("Testing user listing...")

	mcpBinary := s.Config.Binaries.MCPServer
	configFile := fmt.Sprintf("%s/postgres-mcp.yaml", s.Config.ConfigDir)
	userFile := fmt.Sprintf("%s/pgedge-postgres-mcp-users.yaml", s.Config.ConfigDir)

	listCmd := fmt.Sprintf("%s -config %s -list-users -user-file %s", mcpBinary, configFile, userFile)
	output, exitCode, err := s.ExecCommand(listCmd)
	s.NoError(err)
	s.Equal(0, exitCode)
	s.Contains(output, "testuser", "Should list created user")

	s.T().Log("✓ User listing successful")
}

func (s *MCPServerTestSuite) testUser_FilePermissions() {
	s.T().Log("Testing user file permissions...")

	userFile := fmt.Sprintf("%s/pgedge-postgres-mcp-users.yaml", s.Config.ConfigDir)

	// Verify user file was created
	s.AssertFileExists(userFile)

	// Check file permissions
	output, exitCode, err := s.ExecCommand(fmt.Sprintf("stat -c '%%a' %s", userFile))
	s.NoError(err)
	s.Equal(0, exitCode)
	// File should have valid permissions (trim whitespace)
	s.Regexp(`^[0-9]{3}$`, strings.TrimSpace(output), "Should have valid permissions")

	s.T().Log("✓ User file has correct permissions")
}
