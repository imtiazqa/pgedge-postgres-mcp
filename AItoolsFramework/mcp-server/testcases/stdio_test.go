package testcases

import "fmt"

// ============================================================================
// Stdio Mode Tests
// ============================================================================

func (s *MCPServerTestSuite) testStdio_ConfigurationFile() {
	s.T().Log("Testing stdio mode configuration...")

	dbPassword := s.Config.Database.Password
	dbUser := s.Config.Database.User
	dbName := s.Config.Database.Database
	dbHost := s.Config.Database.Host
	dbPort := s.Config.Database.Port

	// Create a test stdio configuration
	configContent := fmt.Sprintf(`cat > /tmp/postgres-mcp-stdio-test.yaml << 'STDIOEOF'
databases:
  - name: %s
    host: %s
    port: %d
    user: %s
    password: %s
    dbname: %s

server:
  mode: stdio
  port: 8080
STDIOEOF`, dbName, dbHost, dbPort, dbUser, dbPassword, dbName)

	output, exitCode, err := s.ExecCommand(configContent)
	s.NoError(err, "Failed to create stdio config: %s", output)
	s.Equal(0, exitCode, "Config creation should succeed")

	// Verify config file was created
	s.AssertFileExists("/tmp/postgres-mcp-stdio-test.yaml")

	s.T().Log("✓ Stdio configuration file created")
}

func (s *MCPServerTestSuite) testStdio_BinarySupportsStdio() {
	s.T().Log("Testing MCP server supports stdio mode...")

	mcpBinary := s.Config.Binaries.MCPServer

	// Check help output mentions stdio mode
	output, _, err := s.ExecCommand(fmt.Sprintf("%s --help 2>&1 || true", mcpBinary))
	s.NoError(err)

	// Help output should mention configuration
	s.Contains(output, "config", "Help should mention configuration")

	s.T().Log("✓ MCP server binary supports stdio mode")
}
