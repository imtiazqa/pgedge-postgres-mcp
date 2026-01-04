package testcases

import (
	"testing"

	"github.com/pgedge/AItoolsFramework/common/suite"
	testifySuite "github.com/stretchr/testify/suite"
)

// MCPServerTestSuite is the master test suite that contains all MCP server tests
// This matches the design of the original regression tests - one suite, one container
//
// All test methods are defined in separate files for better organization:
// - installation_test.go: Installation and package tests
// - service_test.go: Service management tests
// - kb_test.go: Knowledge base tests
// - example_test.go: Example/demo tests
// - helpers_test.go: Helper methods
type MCPServerTestSuite struct {
	suite.E2ESuite
}

// SetupSuite runs once before all tests in the suite
func (s *MCPServerTestSuite) SetupSuite() {
	s.E2ESuite.SetupSuite()

	// Install all required packages once for the entire test run
	// This will be used by all tests in this suite
	s.EnsureMCPPackagesInstalled()

	s.T().Log("MCP Server Test Suite initialized - all packages installed")
}

// TestMCPServerTestSuite runs the test suite
func TestMCPServerTestSuite(t *testing.T) {
	testifySuite.Run(t, new(MCPServerTestSuite))
}
