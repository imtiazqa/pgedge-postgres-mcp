package executor

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math/rand"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// ContainerExecutor runs tests in a Docker container with optional systemd support
type ContainerExecutor struct {
	opts         *ExecutorOptions
	osImage      string
	useSystemd   bool
	containerID  string
	containerName string
	logs         []string
	logsMutex    sync.Mutex
	isStarted    bool
}

// NewContainerExecutor creates a new container executor
func NewContainerExecutor(osImage string, useSystemd bool, opts *ExecutorOptions) (*ContainerExecutor, error) {
	if opts == nil {
		opts = DefaultOptions()
	}

	if osImage == "" {
		return nil, fmt.Errorf("osImage is required for container executor")
	}

	// Generate unique container name using nanoseconds and random number
	// This ensures uniqueness even when multiple containers are created simultaneously
	containerName := fmt.Sprintf("test-container-%d-%d", time.Now().UnixNano(), rand.Int63n(100000))

	return &ContainerExecutor{
		opts:          opts,
		osImage:       osImage,
		useSystemd:    useSystemd,
		containerName: containerName,
		logs:          make([]string, 0),
		isStarted:     false,
	}, nil
}

// Start initializes the container executor
func (c *ContainerExecutor) Start(ctx context.Context) error {
	c.logToBuffer(fmt.Sprintf("Starting container executor with image: %s (systemd: %v)", c.osImage, c.useSystemd))

	// Check if Docker is available
	if err := c.checkDockerAvailable(ctx); err != nil {
		return fmt.Errorf("Docker not available: %w", err)
	}

	// Pull the image if needed
	if err := c.pullImage(ctx); err != nil {
		return fmt.Errorf("failed to pull image: %w", err)
	}

	// Start the container
	if err := c.startContainer(ctx); err != nil {
		return fmt.Errorf("failed to start container: %w", err)
	}

	// Wait for container to be ready
	if err := c.waitForReady(ctx); err != nil {
		// Cleanup on failure
		_ = c.Cleanup(ctx)
		return fmt.Errorf("container failed to become ready: %w", err)
	}

	c.isStarted = true
	c.logToBuffer(fmt.Sprintf("Container %s started successfully", c.containerID))

	// Get OS info
	osInfo, err := c.GetOSInfo(ctx)
	if err != nil {
		c.logToBuffer(fmt.Sprintf("Warning: failed to get OS info: %v", err))
	} else {
		c.logToBuffer(fmt.Sprintf("Container OS: %s", osInfo))
	}

	return nil
}

// checkDockerAvailable verifies Docker is installed and running
func (c *ContainerExecutor) checkDockerAvailable(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "docker", "version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("docker not available: %w\nOutput: %s", err, string(output))
	}
	c.logToBuffer("Docker is available")
	return nil
}

// pullImage pulls the Docker image if not already present
func (c *ContainerExecutor) pullImage(ctx context.Context) error {
	c.logToBuffer(fmt.Sprintf("Checking for image: %s", c.osImage))

	// Check if image exists locally
	cmd := exec.CommandContext(ctx, "docker", "images", "-q", c.osImage)
	output, err := cmd.CombinedOutput()
	if err == nil && strings.TrimSpace(string(output)) != "" {
		c.logToBuffer("Image already exists locally")
		return nil
	}

	// Pull the image
	c.logToBuffer(fmt.Sprintf("Pulling image: %s", c.osImage))
	cmd = exec.CommandContext(ctx, "docker", "pull", c.osImage)
	output, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to pull image: %w\nOutput: %s", err, string(output))
	}

	c.logToBuffer("Image pulled successfully")
	return nil
}

// startContainer creates and starts the Docker container
func (c *ContainerExecutor) startContainer(ctx context.Context) error {
	// First, check if a container with this name already exists and remove it
	checkCmd := exec.CommandContext(ctx, "docker", "ps", "-a", "-q", "-f", fmt.Sprintf("name=^%s$", c.containerName))
	existingID, _ := checkCmd.CombinedOutput()
	if len(existingID) > 0 {
		c.logToBuffer(fmt.Sprintf("Removing existing container: %s", c.containerName))
		// Force remove the existing container
		stopCmd := exec.CommandContext(ctx, "docker", "rm", "-f", strings.TrimSpace(string(existingID)))
		stopCmd.Run() // Ignore errors
	}

	args := []string{"run", "-d"}

	// Add container name
	args = append(args, "--name", c.containerName)

	// Add DNS servers for proper name resolution
	args = append(args,
		"--dns", "8.8.8.8",
		"--dns", "8.8.4.4",
	)

	// Add privileged mode for systemd
	if c.useSystemd {
		args = append(args,
			"--privileged",
			"--cgroupns=host",
			"-v", "/sys/fs/cgroup:/sys/fs/cgroup:rw",
		)
	}

	// Add environment variables
	for key, value := range c.opts.Env {
		args = append(args, "-e", fmt.Sprintf("%s=%s", key, value))
	}

	// Add working directory if specified
	if c.opts.WorkDir != "" {
		args = append(args, "-w", c.opts.WorkDir)
	}

	// Add image and command
	args = append(args, c.osImage)

	if c.useSystemd {
		// For systemd, run /sbin/init
		args = append(args, "/sbin/init")
	} else {
		// For non-systemd, keep container running
		args = append(args, "tail", "-f", "/dev/null")
	}

	c.logToBuffer(fmt.Sprintf("Starting container with command: docker %s", strings.Join(args, " ")))

	cmd := exec.CommandContext(ctx, "docker", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to start container: %w\nOutput: %s", err, string(output))
	}

	c.containerID = strings.TrimSpace(string(output))
	c.logToBuffer(fmt.Sprintf("Container ID: %s", c.containerID))

	return nil
}

// waitForReady waits for the container to be ready
func (c *ContainerExecutor) waitForReady(ctx context.Context) error {
	timeout := time.After(30 * time.Second)
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			// Get container logs to help debug
			logsCmd := exec.CommandContext(ctx, "docker", "logs", c.containerID)
			logs, _ := logsCmd.CombinedOutput()
			c.logToBuffer(fmt.Sprintf("Container logs:\n%s", string(logs)))
			return fmt.Errorf("timeout waiting for container to be ready")
		case <-ticker.C:
			// Check if container is running
			cmd := exec.CommandContext(ctx, "docker", "inspect", "-f", "{{.State.Running}}", c.containerID)
			output, err := cmd.CombinedOutput()
			if err != nil {
				return fmt.Errorf("failed to inspect container: %w", err)
			}

			if strings.TrimSpace(string(output)) == "true" {
				// Container is running, try to execute a simple command directly
				// Don't use c.Exec() here as it checks c.isStarted which isn't set yet
				testCmd := exec.CommandContext(ctx, "docker", "exec", c.containerID, "/bin/bash", "-c", "echo ready")
				testOutput, testErr := testCmd.CombinedOutput()
				if testErr == nil && strings.Contains(string(testOutput), "ready") {
					c.logToBuffer("Container is ready")
					return nil
				}
				// Log the error for debugging
				c.logToBuffer(fmt.Sprintf("Container not ready yet: %v", testErr))
			}
		}
	}
}

// Exec runs a command in the container
func (c *ContainerExecutor) Exec(ctx context.Context, cmdString string) (string, int, error) {
	if !c.isStarted {
		return "", -1, fmt.Errorf("container not started")
	}

	if c.opts.LogCommands {
		c.logToBuffer(fmt.Sprintf("Executing in container: %s", cmdString))
	}

	// Check if context is already expired
	if ctx.Err() != nil {
		return "", -1, fmt.Errorf("context already expired: %w", ctx.Err())
	}

	// Build docker exec command - use /bin/sh for better compatibility
	args := []string{"exec", "-i", c.containerID, "/bin/sh", "-c", cmdString}

	cmd := exec.CommandContext(ctx, "docker", args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	// Combine stdout and stderr
	output := stdout.String() + stderr.String()

	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			return output, -1, fmt.Errorf("failed to execute command in container: %w", err)
		}
	}

	if c.opts.LogOutput {
		c.logToBuffer(fmt.Sprintf("Command completed with exit code: %d", exitCode))
	}

	return output, exitCode, nil
}

// ExecWithInput runs a command with stdin input in the container
func (c *ContainerExecutor) ExecWithInput(ctx context.Context, cmdString string, stdin io.Reader) (string, int, error) {
	if !c.isStarted {
		return "", -1, fmt.Errorf("container not started")
	}

	if c.opts.LogCommands {
		c.logToBuffer(fmt.Sprintf("Executing with input in container: %s", cmdString))
	}

	args := []string{"exec", "-i", c.containerID, "/bin/bash", "-c", cmdString}

	cmd := exec.CommandContext(ctx, "docker", args...)
	cmd.Stdin = stdin

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	output := stdout.String() + stderr.String()

	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			return output, -1, fmt.Errorf("failed to execute command in container: %w", err)
		}
	}

	return output, exitCode, nil
}

// ExecStream runs a command with streaming output in the container
func (c *ContainerExecutor) ExecStream(ctx context.Context, cmdString string, stdout, stderr io.Writer) (int, error) {
	if !c.isStarted {
		return -1, fmt.Errorf("container not started")
	}

	if c.opts.LogCommands {
		c.logToBuffer(fmt.Sprintf("Executing (streaming) in container: %s", cmdString))
	}

	args := []string{"exec", c.containerID, "/bin/bash", "-c", cmdString}

	cmd := exec.CommandContext(ctx, "docker", args...)
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	err := cmd.Run()

	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			return -1, fmt.Errorf("failed to execute command in container: %w", err)
		}
	}

	return exitCode, nil
}

// Cleanup performs cleanup operations
func (c *ContainerExecutor) Cleanup(ctx context.Context) error {
	if c.containerID == "" {
		c.logToBuffer("No container to clean up")
		return nil
	}

	c.logToBuffer(fmt.Sprintf("Cleaning up container: %s", c.containerID))

	// Stop the container
	cmd := exec.CommandContext(ctx, "docker", "stop", "-t", "5", c.containerID)
	output, err := cmd.CombinedOutput()
	if err != nil {
		c.logToBuffer(fmt.Sprintf("Warning: failed to stop container: %v\nOutput: %s", err, string(output)))
	} else {
		c.logToBuffer("Container stopped")
	}

	// Remove the container
	cmd = exec.CommandContext(ctx, "docker", "rm", "-f", c.containerID)
	output, err = cmd.CombinedOutput()
	if err != nil {
		c.logToBuffer(fmt.Sprintf("Warning: failed to remove container: %v\nOutput: %s", err, string(output)))
		return fmt.Errorf("failed to remove container: %w", err)
	}

	c.logToBuffer("Container removed")
	c.isStarted = false
	c.containerID = ""

	return nil
}

// GetLogs retrieves container logs
func (c *ContainerExecutor) GetLogs(ctx context.Context) (string, error) {
	c.logsMutex.Lock()
	executorLogs := strings.Join(c.logs, "\n")
	c.logsMutex.Unlock()

	if c.containerID == "" {
		return executorLogs, nil
	}

	// Get Docker container logs
	cmd := exec.CommandContext(ctx, "docker", "logs", c.containerID)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return executorLogs, fmt.Errorf("failed to get container logs: %w", err)
	}

	return fmt.Sprintf("=== Executor Logs ===\n%s\n\n=== Container Logs ===\n%s", executorLogs, string(output)), nil
}

// Mode returns the execution mode
func (c *ContainerExecutor) Mode() ExecutionMode {
	return ModeContainerSystemd
}

// HealthCheck verifies container is healthy
func (c *ContainerExecutor) HealthCheck(ctx context.Context) error {
	if !c.isStarted {
		return fmt.Errorf("container not started")
	}

	// Check if container is running
	cmd := exec.CommandContext(ctx, "docker", "inspect", "-f", "{{.State.Running}}", c.containerID)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to inspect container: %w", err)
	}

	if strings.TrimSpace(string(output)) != "true" {
		return fmt.Errorf("container is not running")
	}

	// Try to execute a simple command
	_, exitCode, err := c.Exec(ctx, "echo test")
	if err != nil || exitCode != 0 {
		return fmt.Errorf("health check command failed: exit code %d, error: %v", exitCode, err)
	}

	return nil
}

// GetOSInfo returns OS information from the container
func (c *ContainerExecutor) GetOSInfo(ctx context.Context) (string, error) {
	if !c.isStarted {
		return c.osImage, nil
	}

	output, exitCode, err := c.Exec(ctx, "cat /etc/os-release 2>/dev/null || uname -a")
	if err != nil || exitCode != 0 {
		return c.osImage, nil
	}

	// Parse output to get OS name
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "PRETTY_NAME=") {
			return strings.Trim(strings.TrimPrefix(line, "PRETTY_NAME="), "\""), nil
		}
	}

	return strings.TrimSpace(output), nil
}

// logToBuffer adds a log entry
func (c *ContainerExecutor) logToBuffer(message string) {
	c.logsMutex.Lock()
	defer c.logsMutex.Unlock()
	c.logs = append(c.logs, fmt.Sprintf("[%s] %s", time.Now().Format("15:04:05"), message))
}

// CopyToContainer copies a file or directory to the container
func (c *ContainerExecutor) CopyToContainer(ctx context.Context, localPath, containerPath string) error {
	if !c.isStarted {
		return fmt.Errorf("container not started")
	}

	c.logToBuffer(fmt.Sprintf("Copying %s to container:%s", localPath, containerPath))

	cmd := exec.CommandContext(ctx, "docker", "cp", localPath, fmt.Sprintf("%s:%s", c.containerID, containerPath))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to copy to container: %w\nOutput: %s", err, string(output))
	}

	c.logToBuffer("Copy completed")
	return nil
}

// CopyFromContainer copies a file or directory from the container
func (c *ContainerExecutor) CopyFromContainer(ctx context.Context, containerPath, localPath string) error {
	if !c.isStarted {
		return fmt.Errorf("container not started")
	}

	c.logToBuffer(fmt.Sprintf("Copying container:%s to %s", containerPath, localPath))

	cmd := exec.CommandContext(ctx, "docker", "cp", fmt.Sprintf("%s:%s", c.containerID, containerPath), localPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to copy from container: %w\nOutput: %s", err, string(output))
	}

	c.logToBuffer("Copy completed")
	return nil
}

// GetContainerID returns the container ID
func (c *ContainerExecutor) GetContainerID() string {
	return c.containerID
}

// GetContainerName returns the container name
func (c *ContainerExecutor) GetContainerName() string {
	return c.containerName
}
