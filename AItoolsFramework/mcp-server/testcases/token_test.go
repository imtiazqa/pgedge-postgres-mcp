package testcases

import (
	"fmt"
	"strings"
)

// ============================================================================
// Token Management Tests
// ============================================================================

func (s *MCPServerTestSuite) testToken_CreateToken() {
	s.T().Log("Testing token creation...")

	mcpBinary := s.Config.Binaries.MCPServer
	configFile := fmt.Sprintf("%s/postgres-mcp.yaml", s.Config.ConfigDir)
	tokenFile := fmt.Sprintf("%s/pgedge-postgres-mcp-tokens.yaml", s.Config.ConfigDir)

	createCmd := fmt.Sprintf("%s -config %s -add-token -token-file %s -token-note \"test-token\"",
		mcpBinary, configFile, tokenFile)

	output, exitCode, err := s.ExecCommand(createCmd)
	s.NoError(err, "Token creation failed\nOutput: %s", output)
	s.Equal(0, exitCode)
	s.Contains(output, "Token:", "Should show generated token")
	s.Contains(output, "Hash:", "Should show token hash")

	// Set proper ownership on token file
	chownCmd := fmt.Sprintf("chown pgedge:pgedge %s", tokenFile)
	output, exitCode, err = s.ExecCommand(chownCmd)
	s.NoError(err, "Failed to set ownership on token file: %s", output)
	s.Equal(0, exitCode, "chown failed: %s", output)

	s.T().Log("✓ Token created successfully")
}

func (s *MCPServerTestSuite) testToken_ListTokens() {
	s.T().Log("Testing token listing...")

	mcpBinary := s.Config.Binaries.MCPServer
	configFile := fmt.Sprintf("%s/postgres-mcp.yaml", s.Config.ConfigDir)
	tokenFile := fmt.Sprintf("%s/pgedge-postgres-mcp-tokens.yaml", s.Config.ConfigDir)

	listCmd := fmt.Sprintf("%s -config %s -list-tokens -token-file %s", mcpBinary, configFile, tokenFile)
	output, exitCode, err := s.ExecCommand(listCmd)
	s.NoError(err)
	s.Equal(0, exitCode)
	s.Contains(output, "test-token", "Should list created token")

	s.T().Log("✓ Token listing successful")
}

func (s *MCPServerTestSuite) testToken_FileExists() {
	s.T().Log("Testing token file can be created...")

	tokenFile := fmt.Sprintf("%s/pgedge-postgres-mcp-tokens.yaml", s.Config.ConfigDir)

	// Token file may not exist initially - that's okay
	// It will be created when the first token is created
	output, _, _ := s.ExecCommand(fmt.Sprintf("test -f %s && echo exists || echo missing", tokenFile))

	if strings.Contains(output, "missing") {
		s.T().Logf("  ℹ Token file doesn't exist yet (will be created on first token creation)")
	} else {
		s.T().Logf("  ✓ Token file already exists")
	}

	// Verify the directory exists
	configDir := s.Config.ConfigDir
	s.AssertDirectoryExists(configDir)

	s.T().Log("✓ Token file location is accessible")
}
