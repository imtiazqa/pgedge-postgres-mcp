package testcases

import "strings"

// ============================================================================
// Helper Methods
// ============================================================================

// isDebianBased detects if OS is Debian-based
func (s *MCPServerTestSuite) isDebianBased() bool {
	output, exitCode, _ := s.ExecCommand("test -f /etc/debian_version && echo 'debian' || echo 'redhat'")
	return exitCode == 0 && strings.TrimSpace(output) == "debian"
}
