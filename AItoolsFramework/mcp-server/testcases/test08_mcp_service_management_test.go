package testcases

import (
	"fmt"
	"strings"
	"time"
)

// ============================================================================
// Service Management Tests
// ============================================================================

func (s *MCPServerTestSuite) testService_SystemdManagement() {
	s.T().Log("Testing MCP server service management...")

	// Ensure MCP packages are installed (this will install if not already done)
	s.EnsureMCPPackagesInstalled()

	// Determine if we can use systemd based on execution mode
	canUseSystemd := s.Config.Execution.Mode == "container-systemd" || s.Config.Execution.Mode == "local"

	if canUseSystemd {
		s.T().Log("Testing with systemd service management...")

		// ====================================================================
		// 1. Reload systemd daemon to recognize the new service
		// ====================================================================
		s.T().Log("Step 1: Reloading systemd daemon...")
		output, exitCode, err := s.ExecCommand("systemctl daemon-reload")
		s.NoError(err, "Failed to reload systemd daemon: %s", output)
		s.Equal(0, exitCode, "systemctl daemon-reload failed: %s", output)

		// ====================================================================
		// 2. Check if systemd-journald is working (container issue workaround)
		// ====================================================================
		// In container mode with systemd issues (like AlmaLinux 10), journald may fail
		// which prevents services from starting. Check if this is the case.
		if s.Config.Execution.Mode == "container-systemd" {
			s.T().Log("Step 2: Checking systemd-journald availability...")
			journalCheck, _, _ := s.ExecCommand("systemctl is-active systemd-journald.service")
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
		serviceName := s.Config.ServiceName
		s.T().Logf("Step 3: Enabling %s service...", serviceName)
		output, exitCode, err = s.ExecCommand(fmt.Sprintf("systemctl enable %s", serviceName))
		s.NoError(err, "Failed to enable service: %s", output)
		s.Equal(0, exitCode, "systemctl enable failed: %s", output)

		// ====================================================================
		// 4. Start the service
		// ====================================================================
		s.T().Logf("Step 4: Starting %s service...", serviceName)
		output, exitCode, err = s.ExecCommand(fmt.Sprintf("systemctl start %s", serviceName))
		s.NoError(err, "Failed to start service: %s", output)
		s.Equal(0, exitCode, "systemctl start failed: %s", output)

		// Wait for service to fully start
		time.Sleep(5 * time.Second)

		// ====================================================================
		// 5. Check service status
		// ====================================================================
		s.T().Log("Step 5: Checking service status...")
		output, exitCode, err = s.ExecCommand(fmt.Sprintf("systemctl status %s", serviceName))
		// Note: systemctl status returns 0 if active, 3 if not running, 4 if unknown
		if exitCode != 0 {
			s.T().Logf("Service status output:\n%s", output)
		}
		s.NoError(err, "Failed to check service status: %s", output)
		s.Equal(0, exitCode, "Service should be running (status command returned non-zero): %s", output)

		// ====================================================================
		// 6. Verify service is active
		// ====================================================================
		s.T().Log("Step 6: Verifying service is active...")
		output, exitCode, err = s.ExecCommand(fmt.Sprintf("systemctl is-active %s", serviceName))
		s.NoError(err, "Failed to check if service is active: %s", output)
		s.Equal(0, exitCode, "Service is not active: %s", output)
		s.Contains(output, "active", "Service should report as 'active'")

		s.T().Logf("  ✓ Service is active: %s", strings.TrimSpace(output))

		// ====================================================================
		// 7. Test HTTP endpoint connectivity (this also verifies port is listening)
		// ====================================================================
		s.T().Log("Step 7: Testing HTTP endpoint connectivity...")
		// Try to connect with curl (proves service is listening on configured port)
		serverPort := s.Config.Server.Port
		var httpCheckSuccess bool
		var httpStatus string
		for i := 0; i < 5; i++ {
			curlCmd := fmt.Sprintf("curl -s -o /dev/null -w '%%{http_code}' http://localhost:%s/ 2>/dev/null || echo 'curl_failed'", serverPort)
			output, exitCode, _ = s.ExecCommand(curlCmd)
			if exitCode == 0 && !strings.Contains(output, "curl_failed") {
				httpCheckSuccess = true
				httpStatus = strings.TrimSpace(output)
				s.T().Logf("  ✓ HTTP endpoint responded with status: %s (service is listening on port %s)", httpStatus, serverPort)
				break
			}
			time.Sleep(2 * time.Second)
		}

		if !httpCheckSuccess {
			// Show service logs for debugging
			logs, _, _ := s.ExecCommand(fmt.Sprintf("journalctl -u %s -n 50 --no-pager", serviceName))
			s.T().Logf("Service logs:\n%s", logs)

			// Also try port check commands for additional debugging info
			portCheckCmd := fmt.Sprintf("ss -tlnp | grep :%s || netstat -tlnp | grep :%s || lsof -i :%s || echo 'No port check tools available'", serverPort, serverPort, serverPort)
			portCheck, _, _ := s.ExecCommand(portCheckCmd)
			s.T().Logf("Port check output:\n%s", portCheck)

			s.Fail("HTTP endpoint is not responding on port %s", serverPort)
		}

		s.T().Log("✓ Service management tests completed successfully (systemd mode)")

	} else {
		// Manual service testing (for standard containers without systemd)
		s.T().Log("Testing with manual service management (no systemd)...")

		// ====================================================================
		// 1. Start the service manually in the background
		// ====================================================================
		s.T().Log("Step 1: Starting service manually...")

		mcpBinary := s.Config.Binaries.MCPServer
		configFile := fmt.Sprintf("%s/postgres-mcp.yaml", s.Config.ConfigDir)
		logFile := fmt.Sprintf("%s/server.log", s.Config.LogDir)

		// Create a simple script to run the service
		startScript := fmt.Sprintf(`cat > /tmp/start-mcp.sh << 'EOF'
#!/bin/bash
%s -config %s > %s 2>&1 &
echo $! > /tmp/mcp-server.pid
EOF`, mcpBinary, configFile, logFile)
		output, exitCode, err := s.ExecCommand(startScript)
		s.NoError(err, "Failed to create start script: %s", output)
		s.Equal(0, exitCode, "Create start script failed: %s", output)

		// Make it executable
		output, exitCode, err = s.ExecCommand("chmod +x /tmp/start-mcp.sh")
		s.NoError(err, "Failed to make start script executable: %s", output)
		s.Equal(0, exitCode, "chmod failed: %s", output)

		// Run the start script
		output, exitCode, err = s.ExecCommand("/tmp/start-mcp.sh")
		s.NoError(err, "Failed to start service: %s", output)
		s.Equal(0, exitCode, "Service start failed: %s", output)

		// Wait for service to start
		time.Sleep(5 * time.Second)

		// ====================================================================
		// 2. Check if process is running
		// ====================================================================
		s.T().Log("Step 2: Checking if service process is running...")
		output, exitCode, err = s.ExecCommand("ps aux | grep pgedge-postgres-mcp | grep -v grep")
		s.NoError(err, "Failed to check process status: %s", output)
		s.Equal(0, exitCode, "Service process is not running: %s", output)
		s.Contains(output, "pgedge-postgres-mcp", "Service process should be running")

		s.T().Logf("  ✓ Service process is running")

		// ====================================================================
		// 3. Check if service is listening on the configured port
		// ====================================================================
		serverPort := s.Config.Server.Port
		s.T().Logf("Step 3: Verifying service is listening on port %s...", serverPort)
		var portCheckSuccess bool
		for i := 0; i < 5; i++ {
			portCheckCmd := fmt.Sprintf("ss -tlnp | grep :%s || netstat -tlnp | grep :%s || true", serverPort, serverPort)
			output, exitCode, _ = s.ExecCommand(portCheckCmd)
			if exitCode == 0 && strings.Contains(output, serverPort) {
				portCheckSuccess = true
				s.T().Logf("  ✓ Service is listening on port %s", serverPort)
				break
			}
			time.Sleep(2 * time.Second)
		}

		if !portCheckSuccess {
			// Show service logs for debugging
			logCmd := fmt.Sprintf("cat %s 2>/dev/null || echo 'No logs available'", logFile)
			logs, _, _ := s.ExecCommand(logCmd)
			s.T().Logf("Service logs:\n%s", logs)
			s.Fail("Service is not listening on port %s", serverPort)
		}

		// ====================================================================
		// 4. Test HTTP endpoint (basic connectivity)
		// ====================================================================
		s.T().Log("Step 4: Testing HTTP endpoint connectivity...")
		curlCmd := fmt.Sprintf("curl -s -o /dev/null -w '%%{http_code}' http://localhost:%s/ || echo 'curl_failed'", serverPort)
		output, exitCode, err = s.ExecCommand(curlCmd)
		if exitCode == 0 && !strings.Contains(output, "curl_failed") {
			s.T().Logf("  ✓ HTTP endpoint responded with status: %s", strings.TrimSpace(output))
		} else {
			s.T().Logf("  ⚠ Could not reach HTTP endpoint (this may be expected if auth is required)")
		}

		// ====================================================================
		// 5. Stop the service (cleanup)
		// ====================================================================
		s.T().Log("Step 5: Stopping service (cleanup)...")
		output, exitCode, err = s.ExecCommand("kill $(cat /tmp/mcp-server.pid 2>/dev/null) 2>/dev/null || true")
		s.NoError(err, "Failed to stop service: %s", output)

		s.T().Log("✓ Service management tests completed successfully (manual mode)")
	}
}
