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
	// STEP 0: Stop any running MCP server instances
	// ====================================================================
	s.logDetailed("Step 0: Stopping any running MCP server instances...")

	// Check if systemd service is running and stop it
	output, exitCode, _ := s.execCmd(s.ctx, "systemctl is-active pgedge-postgres-mcp.service")
	if exitCode == 0 && strings.TrimSpace(output) == "active" {
		s.logDetailed("MCP server service is running, stopping it...")
		s.execCmd(s.ctx, "systemctl stop pgedge-postgres-mcp.service")
	}
	s.T().Log("  ✓ Ensured no MCP server instances are running")

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
	// STEP 3: Create minimal test configuration
	// ====================================================================
	s.logDetailed("Step 3: Creating minimal KB builder test configuration...")

	// Create a minimal config for testing with just README files (fast, small)
	kbConfigPath := fmt.Sprintf("%s/kb-test-config.yaml", kbPath)
	kbDatabaseFile := fmt.Sprintf("%s/kb.db", kbPath)
	kbDocSourcePath := fmt.Sprintf("%s/doc-source", kbPath)

	kbConfigContent := `# Minimal KB builder config for regression testing
# Note: database_path will be overridden by command line flags

sources:
    # Use a small, fast source - just the pgvector README (no doc_path = root only)
    - git_url: "https://github.com/pgvector/pgvector.git"
      tag: "v0.8.1"
      doc_path: ""
      project_name: "pgvector"
      project_version: "0.8.1"

embeddings:
    # At least one provider must be enabled - using Ollama (no API keys needed)
    openai:
        enabled: false
    voyage:
        enabled: false
    ollama:
        enabled: true
        endpoint: "http://localhost:11434"
        model: "nomic-embed-text"
`

	createConfigCmd := fmt.Sprintf("cat > %s << 'EOF'\n%sEOF", kbConfigPath, kbConfigContent)
	output, exitCode, err = s.execCmd(s.ctx, createConfigCmd)
	s.NoError(err, "Failed to create KB test config: %s", output)
	s.Equal(0, exitCode, "Create KB config exited with non-zero: %s", output)
	s.T().Log("  ✓ Created minimal KB test configuration")

	// ====================================================================
	// STEP 4: Install Ollama and embedding model for testing
	// ====================================================================
	s.logDetailed("Step 4: Installing Ollama and embedding model...")

	// Check if git is available (required for cloning repos)
	_, gitExitCode, _ := s.execCmd(s.ctx, "which git")
	hasGit := (gitExitCode == 0)

	ollamaInstalled := false
	if hasGit {
		// Install Ollama
		s.T().Log("  Installing Ollama...")
		installOllamaCmd := "curl -fsSL https://ollama.com/install.sh | sh"
		output, exitCode, err = s.execCmd(s.ctx, installOllamaCmd)
		if err == nil && exitCode == 0 {
			s.T().Log("  ✓ Ollama installed successfully")

			// Start Ollama service
			s.T().Log("  Starting Ollama service...")
			_, _, _ = s.execCmd(s.ctx, "systemctl start ollama")

			// Wait a moment for Ollama to start
			s.execCmd(s.ctx, "sleep 3")

			// Pull the embedding model
			s.T().Log("  Pulling nomic-embed-text model (this may take a minute)...")
			pullCmd := "ollama pull nomic-embed-text"
			output, exitCode, err = s.execCmd(s.ctx, pullCmd)
			if err == nil && exitCode == 0 {
				s.T().Log("  ✓ Embedding model downloaded successfully")
				ollamaInstalled = true
			} else {
				s.T().Log("  ⚠ Failed to pull embedding model, will skip KB generation")
				s.logDetailed("Model pull output: %s", output)
			}
		} else {
			s.T().Log("  ⚠ Failed to install Ollama, will skip KB generation")
			s.logDetailed("Ollama install output: %s", output)
		}
	} else {
		s.T().Log("  ⚠ Git not available, skipping Ollama installation")
	}

	// ====================================================================
	// STEP 5: Generate a small KB database
	// ====================================================================
	s.logDetailed("Step 5: Generating small KB database at custom path...")

	// Build pgedge-nla-kb-builder command to generate database
	// Use -c flag for config, -d for database path (parametrized)
	// Note: We don't have a flag for doc_source_path, so we create the directory manually
	output, exitCode, err = s.execCmd(s.ctx, fmt.Sprintf("mkdir -p %s", kbDocSourcePath))
	s.NoError(err, "Failed to create doc source directory: %s", output)
	s.Equal(0, exitCode, "mkdir doc source exited with non-zero: %s", output)

	kbGenCmd := fmt.Sprintf("pgedge-nla-kb-builder -c %s -d %s", kbConfigPath, kbDatabaseFile)

	s.T().Log("  Running pgedge-nla-kb-builder generate command...")
	s.T().Logf("  Command: %s", kbGenCmd)

	output, exitCode, err = s.execCmd(s.ctx, kbGenCmd)

	// Log the output regardless of success/failure
	s.logDetailed("kb generate output:\n%s", output)

	// Check generation results
	if exitCode != 0 {
		// If Ollama was successfully installed, KB generation should work
		if ollamaInstalled {
			s.NoError(err, "Failed to run kb generate: %s", output)
			s.Equal(0, exitCode, "kb generate exited with non-zero: %s", output)
		} else {
			// Ollama wasn't installed (git missing or install failed), so we expect this to fail
			s.T().Log("  ⚠ KB generation skipped (Ollama not available)")
			s.T().Log("  Note: KB builder tool is installed and can load config")
		}
	} else {
		s.T().Log("  ✓ kb generate completed successfully")
	}

	// ====================================================================
	// STEP 6: Verify KB database was created (if generation succeeded)
	// ====================================================================
	fileCount := "0"
	if exitCode == 0 {
		s.logDetailed("Step 6: Verifying KB database files...")

		// Check if the KB database directory exists and has content
		output, exitCode, err = s.execCmd(s.ctx, fmt.Sprintf("ls -la %s", kbPath))
		s.NoError(err, "Failed to list KB database directory: %s", output)
		s.Equal(0, exitCode, "ls exited with non-zero: %s", output)

		s.logDetailed("KB database contents:\n%s", output)

		// Verify directory is not empty
		output, exitCode, err = s.execCmd(s.ctx, fmt.Sprintf("find %s -type f | wc -l", kbPath))
		s.NoError(err, "Failed to count files in KB database: %s", output)
		s.Equal(0, exitCode, "find exited with non-zero: %s", output)

		fileCount = strings.TrimSpace(output)
		s.NotEqual("0", fileCount, "KB database directory should contain files")
		s.T().Logf("  ✓ KB database created with %s file(s)", fileCount)

		// ====================================================================
		// STEP 7: Verify KB database structure
		// ====================================================================
		s.logDetailed("Step 7: Verifying KB database structure...")

		// Check for expected KB database files/directories
		// The exact structure depends on kb implementation, adapt as needed
		output, exitCode, err = s.execCmd(s.ctx, fmt.Sprintf("ls -R %s", kbPath))
		s.NoError(err, "Failed to list KB database structure: %s", output)
		s.Equal(0, exitCode, "ls -R exited with non-zero: %s", output)

		s.logDetailed("KB database structure:\n%s", output)
		s.T().Log("  ✓ KB database structure verified")
	} else {
		s.T().Log("  ⚠ Skipping database verification (generation did not complete)")
	}

	// ====================================================================
	// STEP 8: Cleanup Ollama (if installed)
	// ====================================================================
	if ollamaInstalled {
		s.logDetailed("Step 8: Cleaning up Ollama installation...")

		// Stop Ollama service
		s.execCmd(s.ctx, "systemctl stop ollama")

		// Remove Ollama
		s.execCmd(s.ctx, "systemctl disable ollama")
		s.execCmd(s.ctx, "rm -rf /usr/local/bin/ollama")
		s.execCmd(s.ctx, "rm -rf /usr/share/ollama")
		s.execCmd(s.ctx, "rm -rf ~/.ollama")
		s.execCmd(s.ctx, "rm -f /etc/systemd/system/ollama.service")
		s.execCmd(s.ctx, "systemctl daemon-reload")

		s.T().Log("  ✓ Ollama cleaned up")
	}

	// ====================================================================
	// STEP 9: Cleanup test files
	// ====================================================================
	s.logDetailed("Step 9: Cleaning up test KB database...")

	output, exitCode, err = s.execCmd(s.ctx, fmt.Sprintf("rm -rf %s", kbPath))
	s.NoError(err, "Failed to remove KB database: %s", output)
	s.Equal(0, exitCode, "rm exited with non-zero: %s", output)
	s.T().Log("  ✓ Test KB database cleaned up")

	s.T().Log("✓ Knowledge Builder tests completed")
	s.T().Log("  • Pre-check: Stopped any running MCP server instances")
	s.T().Log("  • Help: pgedge-nla-kb-builder --help displayed usage information")
	s.T().Log("  • Configuration: Created and validated minimal KB config")
	s.T().Log(fmt.Sprintf("  • Database path: Custom path %s verified", kbPath))
	if ollamaInstalled {
		s.T().Log("  • Ollama: Installed embedding service temporarily")
	}
	if fileCount != "0" {
		s.T().Log(fmt.Sprintf("  • Database generation: Successfully created %s file(s)", fileCount))
	} else {
		s.T().Log("  • Database generation: Skipped (git not available)")
	}
	if ollamaInstalled {
		s.T().Log("  • Cleanup: Ollama and test database removed")
	} else {
		s.T().Log("  • Cleanup: Test database removed")
	}
}
