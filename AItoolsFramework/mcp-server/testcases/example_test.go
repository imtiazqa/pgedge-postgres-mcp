package testcases

// ============================================================================
// Example/Demo Tests
// ============================================================================

func (s *MCPServerTestSuite) testExample_InstallationVerification() {
	s.T().Log("Testing installation verification...")

	// This test demonstrates that all packages are available
	// and can be used by other tests

	binaries := []string{
		s.Config.Binaries.MCPServer,
		s.Config.Binaries.CLI,
		s.Config.Binaries.KBBuilder,
	}

	for _, binary := range binaries {
		s.AssertFileExists(binary)
		s.T().Logf("✓ %s exists", binary)
	}

	s.T().Log("✓ All installed components verified")
}
