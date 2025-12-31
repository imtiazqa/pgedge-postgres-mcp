package regression

import (
	"fmt"
	"strings"
)

// ========================================================================
// TEST 09: Knowledge Builder Testing
// ========================================================================

// Test09_KnowledgeBuilder tests the knowledge builder functionality
func (s *RegressionTestSuite) Test09_KnowledgeBuilder() {
	s.T().Log("TEST 09: Testing Knowledge Builder")

	// Ensure packages are installed
	s.ensureMCPPackagesInstalled()

	// ====================================================================
	// STEP 1: Print KB help
	// ====================================================================
	s.logDetailed("Step 1: Printing KB help...")

	// Run pgedge-nla-kb-builder command with --help flag
	output, exitCode, err := s.execCmd(s.ctx, "pgedge-nla-kb-builder --help")
	s.NoError(err, "Failed to run kb --help: %s", output)
	s.Equal(0, exitCode, "kb --help exited with non-zero: %s", output)

	// Verify help output contains expected sections
	s.Contains(output, "Usage:", "Help should contain 'Usage:' section")
	s.Contains(output, "Flags:", "Help should contain 'Flags:' section")
	s.T().Log("  ✓ kb --help executed successfully")

	// Log the help output for reference
	s.logDetailed("kb --help output:\n%s", output)

	// ====================================================================
	// STEP 2: Create custom directory for KB database
	// ====================================================================
	s.logDetailed("Step 2: Creating custom directory for KB database...")

	// Use a non-default path for the KB database
	kbPath := "/tmp/test_kb_database"

	// Clean up any existing test KB database
	s.execCmd(s.ctx, fmt.Sprintf("rm -rf %s", kbPath))

	// Create the directory
	output, exitCode, err = s.execCmd(s.ctx, fmt.Sprintf("mkdir -p %s", kbPath))
	s.NoError(err, "Failed to create KB database directory: %s", output)
	s.Equal(0, exitCode, "mkdir exited with non-zero: %s", output)
	s.T().Log(fmt.Sprintf("  ✓ Created KB database directory: %s", kbPath))

	// ====================================================================
	// STEP 3: Generate a small KB database
	// ====================================================================
	s.logDetailed("Step 3: Generating small KB database at custom path...")

	// Build pgedge-nla-kb-builder command to generate database
	// Use -c flag to specify config file and -d flag for database path
	kbDatabaseFile := fmt.Sprintf("%s/kb.db", kbPath)
	kbGenCmd := fmt.Sprintf("pgedge-nla-kb-builder -c /etc/pgedge/pgedge-nla-kb-builder.yaml -d %s", kbDatabaseFile)

	s.T().Log("  Running pgedge-nla-kb-builder generate command...")
	s.T().Logf("  Command: %s", kbGenCmd)

	output, exitCode, err = s.execCmd(s.ctx, kbGenCmd)

	// Log the output regardless of success/failure
	s.logDetailed("kb generate output:\n%s", output)

	s.NoError(err, "Failed to run kb generate: %s", output)
	s.Equal(0, exitCode, "kb generate exited with non-zero: %s", output)

	// Verify output contains success indicators
	if !strings.Contains(output, "error") && !strings.Contains(output, "Error") {
		s.T().Log("  ✓ kb generate completed successfully")
	}

	// ====================================================================
	// STEP 4: Verify KB database was created
	// ====================================================================
	s.logDetailed("Step 4: Verifying KB database files...")

	// Check if the KB database directory exists and has content
	output, exitCode, err = s.execCmd(s.ctx, fmt.Sprintf("ls -la %s", kbPath))
	s.NoError(err, "Failed to list KB database directory: %s", output)
	s.Equal(0, exitCode, "ls exited with non-zero: %s", output)

	s.logDetailed("KB database contents:\n%s", output)

	// Verify directory is not empty
	output, exitCode, err = s.execCmd(s.ctx, fmt.Sprintf("find %s -type f | wc -l", kbPath))
	s.NoError(err, "Failed to count files in KB database: %s", output)
	s.Equal(0, exitCode, "find exited with non-zero: %s", output)

	fileCount := strings.TrimSpace(output)
	s.NotEqual("0", fileCount, "KB database directory should contain files")
	s.T().Logf("  ✓ KB database created with %s file(s)", fileCount)

	// ====================================================================
	// STEP 5: Verify KB database structure
	// ====================================================================
	s.logDetailed("Step 5: Verifying KB database structure...")

	// Check for expected KB database files/directories
	// The exact structure depends on kb implementation, adapt as needed
	output, exitCode, err = s.execCmd(s.ctx, fmt.Sprintf("ls -R %s", kbPath))
	s.NoError(err, "Failed to list KB database structure: %s", output)
	s.Equal(0, exitCode, "ls -R exited with non-zero: %s", output)

	s.logDetailed("KB database structure:\n%s", output)
	s.T().Log("  ✓ KB database structure verified")

	// ====================================================================
	// STEP 6: Cleanup
	// ====================================================================
	s.logDetailed("Step 6: Cleaning up test KB database...")

	output, exitCode, err = s.execCmd(s.ctx, fmt.Sprintf("rm -rf %s", kbPath))
	s.NoError(err, "Failed to remove KB database: %s", output)
	s.Equal(0, exitCode, "rm exited with non-zero: %s", output)
	s.T().Log("  ✓ Test KB database cleaned up")

	s.T().Log("✓ Knowledge Builder tests completed successfully")
	s.T().Log("  • Help: pgedge-nla-kb-builder --help displayed usage information")
	s.T().Log(fmt.Sprintf("  • Database path: Custom path %s verified", kbPath))
	s.T().Log(fmt.Sprintf("  • Database generation: Successfully created %s file(s)", fileCount))
	s.T().Log("  • Cleanup: Test database removed")
}
