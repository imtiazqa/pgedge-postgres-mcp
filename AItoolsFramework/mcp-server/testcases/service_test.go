package testcases

import "fmt"

// ============================================================================
// Service Tests
// ============================================================================

func (s *MCPServerTestSuite) testService_MCPServerBinary() {
	s.T().Log("Testing MCP server binary...")

	mcpBinary := s.Config.Binaries.MCPServer

	// Test help command
	output, _, err := s.ExecCommand(fmt.Sprintf("%s --help", mcpBinary))
	s.NoError(err)
	// Note: --help often returns exit code 2, so we just check it ran
	s.Contains(output, "Usage", "Help should show usage information")

	s.T().Log("✓ MCP server binary is functional")
}
