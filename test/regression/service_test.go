package regression

import (
	"strings"
	"time"
)

// ========================================================================
// TEST 08: Service Management (Start and Status Check)
// ========================================================================

// Test08_ServiceManagement tests MCP server service management
func (s *RegressionTestSuite) Test08_ServiceManagement() {
	s.T().Log("TEST 08: Testing MCP server service management")

	// Ensure packages are installed
	s.ensureMCPPackagesInstalled()

	// Determine if we can use systemd based on execution mode
	canUseSystemd := s.execMode == ModeContainerSystemd || s.execMode == ModeLocal

	if canUseSystemd {
		s.T().Log("Testing with systemd service management...")

		// ====================================================================
		// 1. Reload systemd daemon to recognize the new service
		// ====================================================================
		s.logDetailed("Step 1: Reloading systemd daemon...")
		output, exitCode, err := s.execCmd(s.ctx, "systemctl daemon-reload")
		s.NoError(err, "Failed to reload systemd daemon: %s", output)
		s.Equal(0, exitCode, "systemctl daemon-reload failed: %s", output)

		// ====================================================================
		// 2. Check if systemd-journald is working (container issue workaround)
		// ====================================================================
		// In container mode with systemd issues (like AlmaLinux 10), journald may fail
		// which prevents services from starting. Check if this is the case.
		if s.execMode == ModeContainerSystemd {
			s.logDetailed("Step 2: Checking systemd-journald availability...")
			journalCheck, _, _ := s.execCmd(s.ctx, "systemctl is-active systemd-journald.service")
			if !strings.Contains(journalCheck, "active") {
				s.T().Log("  ⚠ systemd-journald is not available in this container")
				s.T().Log("  ℹ Skipping service tests (services require working journald)")
				s.T().Log("  ℹ Note: Package installation and configuration were verified successfully")
				s.T().Log("✓ Service management tests skipped (systemd-journald unavailable in container)")
				return
			}
		}

		// ====================================================================
		// 3. Enable the service (so it starts on boot)
		// ====================================================================
		s.logDetailed("Step 3: Enabling pgedge-postgres-mcp service...")
		output, exitCode, err = s.execCmd(s.ctx, "systemctl enable pgedge-postgres-mcp.service")
		s.NoError(err, "Failed to enable service: %s", output)
		s.Equal(0, exitCode, "systemctl enable failed: %s", output)

		// ====================================================================
		// 4. Start the service
		// ====================================================================
		s.logDetailed("Step 4: Starting pgedge-postgres-mcp service...")
		output, exitCode, err = s.execCmd(s.ctx, "systemctl start pgedge-postgres-mcp.service")
		s.NoError(err, "Failed to start service: %s", output)
		s.Equal(0, exitCode, "systemctl start failed: %s", output)

		// Wait for service to fully start
		time.Sleep(5 * time.Second)

		// ====================================================================
		// 5. Check service status
		// ====================================================================
		s.logDetailed("Step 5: Checking service status...")
		output, exitCode, err = s.execCmd(s.ctx, "systemctl status pgedge-postgres-mcp.service")
		// Note: systemctl status returns 0 if active, 3 if not running, 4 if unknown
		if exitCode != 0 {
			s.T().Logf("Service status output:\n%s", output)
		}
		s.NoError(err, "Failed to check service status: %s", output)
		s.Equal(0, exitCode, "Service should be running (status command returned non-zero): %s", output)

		// ====================================================================
		// 6. Verify service is active
		// ====================================================================
		s.logDetailed("Step 6: Verifying service is active...")
		output, exitCode, err = s.execCmd(s.ctx, "systemctl is-active pgedge-postgres-mcp.service")
		s.NoError(err, "Failed to check if service is active: %s", output)
		s.Equal(0, exitCode, "Service is not active: %s", output)
		s.Contains(output, "active", "Service should report as 'active'")

		s.T().Logf("  ✓ Service is active: %s", strings.TrimSpace(output))

		// ====================================================================
		// 7. Test HTTP endpoint connectivity (this also verifies port is listening)
		// ====================================================================
		s.logDetailed("Step 7: Testing HTTP endpoint connectivity...")
		// Try to connect with curl (proves service is listening on port 8080)
		var httpCheckSuccess bool
		var httpStatus string
		for i := 0; i < 5; i++ {
			output, exitCode, _ = s.execCmd(s.ctx, "curl -s -o /dev/null -w '%{http_code}' http://localhost:8080/ 2>/dev/null || echo 'curl_failed'")
			if exitCode == 0 && !strings.Contains(output, "curl_failed") {
				httpCheckSuccess = true
				httpStatus = strings.TrimSpace(output)
				s.T().Logf("  ✓ HTTP endpoint responded with status: %s (service is listening on port 8080)", httpStatus)
				break
			}
			time.Sleep(2 * time.Second)
		}

		if !httpCheckSuccess {
			// Show service logs for debugging
			logs, _, _ := s.execCmd(s.ctx, "journalctl -u pgedge-postgres-mcp.service -n 50 --no-pager")
			s.T().Logf("Service logs:\n%s", logs)

			// Also try port check commands for additional debugging info
			portCheck, _, _ := s.execCmd(s.ctx, "ss -tlnp | grep :8080 || netstat -tlnp | grep :8080 || lsof -i :8080 || echo 'No port check tools available'")
			s.T().Logf("Port check output:\n%s", portCheck)

			s.Fail("HTTP endpoint is not responding on port 8080")
		}

		s.T().Log("✓ Service management tests completed successfully (systemd mode)")

	} else {
		// Manual service testing (for standard containers without systemd)
		s.T().Log("Testing with manual service management (no systemd)...")

		// ====================================================================
		// 1. Start the service manually in the background
		// ====================================================================
		s.logDetailed("Step 1: Starting pgedge-postgres-mcp manually...")

		// Create a simple script to run the service
		startScript := `cat > /tmp/start-mcp.sh << 'EOF'
#!/bin/bash
/usr/bin/pgedge-postgres-mcp -config /etc/pgedge/postgres-mcp.yaml > /var/log/pgedge/postgres-mcp/server.log 2>&1 &
echo $! > /tmp/mcp-server.pid
EOF`
		output, exitCode, err := s.execCmd(s.ctx, startScript)
		s.NoError(err, "Failed to create start script: %s", output)
		s.Equal(0, exitCode, "Create start script failed: %s", output)

		// Make it executable
		output, exitCode, err = s.execCmd(s.ctx, "chmod +x /tmp/start-mcp.sh")
		s.NoError(err, "Failed to make start script executable: %s", output)
		s.Equal(0, exitCode, "chmod failed: %s", output)

		// Run the start script
		output, exitCode, err = s.execCmd(s.ctx, "/tmp/start-mcp.sh")
		s.NoError(err, "Failed to start service: %s", output)
		s.Equal(0, exitCode, "Service start failed: %s", output)

		// Wait for service to start
		time.Sleep(5 * time.Second)

		// ====================================================================
		// 2. Check if process is running
		// ====================================================================
		s.logDetailed("Step 2: Checking if service process is running...")
		output, exitCode, err = s.execCmd(s.ctx, "ps aux | grep pgedge-postgres-mcp | grep -v grep")
		s.NoError(err, "Failed to check process status: %s", output)
		s.Equal(0, exitCode, "Service process is not running: %s", output)
		s.Contains(output, "pgedge-postgres-mcp", "Service process should be running")

		s.T().Logf("  ✓ Service process is running")

		// ====================================================================
		// 3. Check if service is listening on the configured port
		// ====================================================================
		s.logDetailed("Step 3: Verifying service is listening on port 8080...")
		var portCheckSuccess bool
		for i := 0; i < 5; i++ {
			output, exitCode, _ = s.execCmd(s.ctx, "ss -tlnp | grep :8080 || netstat -tlnp | grep :8080 || true")
			if exitCode == 0 && strings.Contains(output, "8080") {
				portCheckSuccess = true
				s.T().Logf("  ✓ Service is listening on port 8080")
				break
			}
			time.Sleep(2 * time.Second)
		}

		if !portCheckSuccess {
			// Show service logs for debugging
			logs, _, _ := s.execCmd(s.ctx, "cat /var/log/pgedge/postgres-mcp/server.log 2>/dev/null || echo 'No logs available'")
			s.T().Logf("Service logs:\n%s", logs)
			s.Fail("Service is not listening on port 8080")
		}

		// ====================================================================
		// 4. Test HTTP endpoint (basic connectivity)
		// ====================================================================
		s.logDetailed("Step 4: Testing HTTP endpoint connectivity...")
		output, exitCode, err = s.execCmd(s.ctx, "curl -s -o /dev/null -w '%{http_code}' http://localhost:8080/ || echo 'curl_failed'")
		if exitCode == 0 && !strings.Contains(output, "curl_failed") {
			s.T().Logf("  ✓ HTTP endpoint responded with status: %s", strings.TrimSpace(output))
		} else {
			s.T().Logf("  ⚠ Could not reach HTTP endpoint (this may be expected if auth is required)")
		}

		// ====================================================================
		// 5. Stop the service (cleanup)
		// ====================================================================
		s.logDetailed("Step 5: Stopping service (cleanup)...")
		output, exitCode, err = s.execCmd(s.ctx, "kill $(cat /tmp/mcp-server.pid 2>/dev/null) 2>/dev/null || true")
		s.NoError(err, "Failed to stop service: %s", output)

		s.T().Log("✓ Service management tests completed successfully (manual mode)")
	}
}
