package testcases

import "fmt"

// ============================================================================
// MCP Knowledge Base Integration Tests
// ============================================================================

func (s *MCPServerTestSuite) testMCPKB_BuilderBinary() {
	s.T().Log("Testing KB builder binary...")

	kbBinary := s.Config.Binaries.KBBuilder

	// Check if KB builder binary exists
	s.AssertFileExists(kbBinary)

	// Test binary is executable
	output, exitCode, err := s.ExecCommand(fmt.Sprintf("test -x %s && echo 'OK'", kbBinary))
	s.NoError(err)
	s.Equal(0, exitCode)
	s.Contains(output, "OK", "KB builder should be executable")

	s.T().Log("✓ KB builder binary is available and executable")
}

func (s *MCPServerTestSuite) testMCPKB_BuilderHelp() {
	s.T().Log("Testing KB builder help command...")

	kbBinary := s.Config.Binaries.KBBuilder

	// Run help command
	output, _, err := s.ExecCommand(fmt.Sprintf("%s --help 2>&1 || true", kbBinary))
	s.NoError(err)

	// Verify help output contains expected information
	s.Contains(output, "Usage:", "Help should contain usage information")
	s.Contains(output, "Flags:", "Help should contain flags information")

	s.T().Log("✓ KB builder help command works")
}

func (s *MCPServerTestSuite) testMCPKB_ConfigFile() {
	s.T().Log("Testing KB builder configuration file...")

	configFile := fmt.Sprintf("%s/pgedge-nla-kb-builder.yaml", s.Config.ConfigDir)

	// Check if config file exists
	s.AssertFileExists(configFile)

	// Read and verify config content
	output, exitCode, err := s.ExecCommand(fmt.Sprintf("cat %s", configFile))
	s.NoError(err)
	s.Equal(0, exitCode)
	s.NotEmpty(output, "Config file should not be empty")

	s.T().Log("✓ KB builder configuration file exists")
}

func (s *MCPServerTestSuite) testMCPKB_DefaultDatabaseLocation() {
	s.T().Log("Testing KB default database location...")

	// Get home directory for default KB location
	homeOutput, exitCode, err := s.ExecCommand("echo $HOME")
	if exitCode == 0 && err == nil {
		homeDir := homeOutput
		defaultKBPath := fmt.Sprintf("%s/.pgedge/pgedge-nla-kb.db", homeDir)

		// Check if default KB database exists
		output, exitCode, _ := s.ExecCommand(fmt.Sprintf("test -f %s && echo 'exists' || echo 'missing'", defaultKBPath))

		if exitCode == 0 && output == "exists\n" {
			s.T().Logf("  ✓ Default KB database found at %s", defaultKBPath)
		} else {
			s.T().Log("  ℹ No default KB database found (this is normal for fresh installation)")
		}
	}
}
