package testcases

import "fmt"

// ============================================================================
// TEST 01: Repository Installation Tests
// ============================================================================

// testRepository_Installation installs and verifies the pgEdge repository
func (s *MCPServerTestSuite) testRepository_Installation() {
	s.T().Log("Testing repository installation...")

	// Ensure repository is installed (this will install if not already done)
	s.EnsureRepositoryInstalled()

	// Verify repository is properly configured
	isDebian := s.isDebianBased()

	if isDebian {
		// Debian/Ubuntu - verify apt repository
		s.T().Log("Verifying apt repository configuration...")
		output, exitCode, err := s.ExecCommand("apt-cache policy | grep -i pgedge")
		s.NoError(err)
		s.Equal(0, exitCode, "pgEdge repository should be in apt sources")
		s.Contains(output, "pgedge", "Repository should contain pgedge packages")
		s.T().Log("✓ pgEdge apt repository is installed and configured")
	} else {
		// RHEL/Rocky/Alma - verify dnf repository
		s.T().Log("Verifying dnf repository configuration...")
		output, exitCode, err := s.ExecCommand("dnf repolist | grep -i pgedge")
		s.NoError(err)
		s.Equal(0, exitCode, "pgEdge repository should be in dnf repos")
		s.Contains(output, "pgedge", "Repository should contain pgedge packages")
		s.T().Log("✓ pgEdge dnf repository is installed and configured")
	}
}

// testRepository_PackageAvailability verifies all required packages are available in the repository
func (s *MCPServerTestSuite) testRepository_PackageAvailability() {
	s.T().Log("Testing package availability in repository...")

	isDebian := s.isDebianBased()

	// Get package names from config
	allPackages := append(s.Config.Packages.MCPServer, s.Config.Packages.CLI...)
	allPackages = append(allPackages, s.Config.Packages.Web...)
	allPackages = append(allPackages, s.Config.Packages.KB...)

	s.T().Logf("Checking %d packages in repository...", len(allPackages))

	for _, pkg := range allPackages {
		var searchCmd string
		if isDebian {
			searchCmd = fmt.Sprintf("apt-cache search %s", pkg)
		} else {
			searchCmd = fmt.Sprintf("dnf search %s", pkg)
		}

		output, exitCode, err := s.ExecCommand(searchCmd)
		s.NoError(err, "Failed to search for package %s", pkg)
		s.Equal(0, exitCode, "Package search should succeed for %s", pkg)
		s.Contains(output, pkg, "Package %s should be available in repository", pkg)

		s.T().Logf("  ✓ Package %s is available in repository", pkg)
	}

	s.T().Log("✓ All required packages are available in repository")
}
