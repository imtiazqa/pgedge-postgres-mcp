package testcases

// ============================================================================
// Repository Installation Tests
// ============================================================================

func (s *MCPServerTestSuite) testRepository_Installation() {
	s.T().Log("Testing repository installation...")

	isDebian := s.isDebianBased()

	if isDebian {
		// Debian/Ubuntu - check apt repository
		output, exitCode, err := s.ExecCommand("apt-cache policy | grep -i pgedge")
		s.NoError(err)
		s.Equal(0, exitCode, "pgEdge repository should be in apt sources")
		s.Contains(output, "pgedge", "Repository should contain pgedge packages")
		s.T().Log("✓ pgEdge apt repository is installed")
	} else {
		// RHEL/Rocky/Alma - check dnf repository
		output, exitCode, err := s.ExecCommand("dnf repolist | grep -i pgedge")
		s.NoError(err)
		s.Equal(0, exitCode, "pgEdge repository should be in dnf repos")
		s.Contains(output, "pgedge", "Repository should contain pgedge packages")
		s.T().Log("✓ pgEdge dnf repository is installed")
	}
}

func (s *MCPServerTestSuite) testRepository_PackageAvailability() {
	s.T().Log("Testing package availability in repository...")

	isDebian := s.isDebianBased()

	// Get package names from config
	allPackages := append(s.Config.Packages.MCPServer, s.Config.Packages.CLI...)
	allPackages = append(allPackages, s.Config.Packages.Web...)
	allPackages = append(allPackages, s.Config.Packages.KB...)

	for _, pkg := range allPackages {
		var searchCmd string
		if isDebian {
			searchCmd = "apt-cache search " + pkg
		} else {
			searchCmd = "dnf search " + pkg
		}

		output, exitCode, err := s.ExecCommand(searchCmd)
		s.NoError(err, "Failed to search for package %s", pkg)
		s.Equal(0, exitCode, "Package search should succeed for %s", pkg)
		s.Contains(output, pkg, "Package %s should be available in repository", pkg)

		s.T().Logf("  ✓ Package %s is available in repository", pkg)
	}
}
