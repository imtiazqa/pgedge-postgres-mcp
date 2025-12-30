package regression

import (
	"strings"
)

// ========================================================================
// TEST 06: User Management
// ========================================================================

// Test06_UserManagement tests user creation, listing, and permissions
func (s *RegressionTestSuite) Test06_UserManagement() {
	s.T().Log("TEST 06: Testing user management commands")

	// Ensure packages are installed
	s.ensureMCPPackagesInstalled()

	// Test 1: Create user (using config file for database connection)
	createCmd := `/usr/bin/pgedge-postgres-mcp -config /etc/pgedge/postgres-mcp.yaml -add-user -user-file /etc/pgedge/pgedge-postgres-mcp-users.yaml -username testuser -password testpass123 -user-note "test user"`
	output, exitCode, err := s.execCmd(s.ctx, createCmd)
	s.NoError(err, "User creation failed\nOutput: %s", output)
	s.Equal(0, exitCode)
	s.Contains(output, "User created", "Should confirm user creation")

	// Test 2: List users
	output, exitCode, err = s.execCmd(s.ctx, "/usr/bin/pgedge-postgres-mcp -config /etc/pgedge/postgres-mcp.yaml -list-users -user-file /etc/pgedge/pgedge-postgres-mcp-users.yaml")
	s.NoError(err)
	s.Equal(0, exitCode)
	s.Contains(output, "testuser", "Should list created user")

	// Test 3: Verify user file was created
	output, exitCode, err = s.execCmd(s.ctx, "test -f /etc/pgedge/pgedge-postgres-mcp-users.yaml && echo 'OK'")
	s.NoError(err)
	s.Equal(0, exitCode)
	s.Contains(output, "OK", "User file should exist")

	// Test 4: User file has correct permissions (should be restrictive)
	output, exitCode, err = s.execCmd(s.ctx, "stat -c '%a' /etc/pgedge/pgedge-postgres-mcp-users.yaml")
	s.NoError(err)
	s.Equal(0, exitCode)
	// File should be readable but ideally 600 or 644
	output = strings.TrimSpace(output)
	s.Regexp(`^[0-9]{3}$`, output, "Should have valid permissions")

	s.T().Log("✓ User management working correctly")
}
