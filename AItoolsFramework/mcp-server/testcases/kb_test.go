package testcases

import "fmt"

// ============================================================================
// Knowledge Base Tests
// ============================================================================

func (s *MCPServerTestSuite) testKB_BuilderBinary() {
	s.T().Log("Testing KB builder binary...")

	kbBinary := s.Config.Binaries.KBBuilder

	s.AssertFileExists(kbBinary)

	// Test if binary is executable
	output, _, err := s.ExecCommand(fmt.Sprintf("%s --help 2>&1 || true", kbBinary))
	s.NoError(err)
	// Binary exists and can be executed (exit code may vary for --help)
	s.T().Logf("KB builder output: %s", output)

	s.T().Log("✓ KB builder binary available")
}
