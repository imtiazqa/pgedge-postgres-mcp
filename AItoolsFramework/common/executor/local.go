package executor

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"
)

// LocalExecutor runs tests on the local machine
type LocalExecutor struct {
	opts      *ExecutorOptions
	logs      []string
	logsMutex sync.Mutex
}

// NewLocalExecutor creates a new local machine executor
func NewLocalExecutor(opts *ExecutorOptions) (*LocalExecutor, error) {
	if opts == nil {
		opts = DefaultOptions()
	}

	return &LocalExecutor{
		opts: opts,
		logs: make([]string, 0),
	}, nil
}

// Start initializes the local executor
func (l *LocalExecutor) Start(ctx context.Context) error {
	// Check if we have necessary permissions
	output, exitCode, err := l.Exec(ctx, "whoami")
	if err != nil {
		return fmt.Errorf("failed to check user: %w", err)
	}

	user := strings.TrimSpace(output)
	l.logToBuffer(fmt.Sprintf("Running tests as user: %s (exit code: %d)", user, exitCode))

	// Check for sudo access (needed for package installation) unless explicitly skipped
	if !l.opts.SkipSudoCheck {
		output, exitCode, err = l.Exec(ctx, "sudo -n true 2>&1")
		if exitCode != 0 {
			return fmt.Errorf("sudo access required for local execution. Please configure passwordless sudo or run with appropriate privileges. Output: %s", output)
		}

		l.logToBuffer("Sudo access verified")
	} else {
		l.logToBuffer("Sudo check skipped (SkipSudoCheck=true)")
	}

	// Detect OS
	osInfo, err := l.GetOSInfo(ctx)
	if err != nil {
		return fmt.Errorf("failed to detect OS: %w", err)
	}
	l.logToBuffer(fmt.Sprintf("Detected OS: %s", osInfo))

	return nil
}

// Exec runs a command on the local machine
func (l *LocalExecutor) Exec(ctx context.Context, cmdString string) (string, int, error) {
	if l.opts.LogCommands {
		l.logToBuffer(fmt.Sprintf("Executing: %s", cmdString))
	}

	cmd := exec.CommandContext(ctx, "/bin/bash", "-c", cmdString)

	// Set working directory if specified
	if l.opts.WorkDir != "" {
		cmd.Dir = l.opts.WorkDir
	}

	// Set environment variables
	if len(l.opts.Env) > 0 {
		env := cmd.Environ()
		for key, value := range l.opts.Env {
			env = append(env, fmt.Sprintf("%s=%s", key, value))
		}
		cmd.Env = env
	}

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
			// Command failed to start
			return output, -1, fmt.Errorf("failed to execute command: %w", err)
		}
	}

	if l.opts.LogOutput {
		l.logToBuffer(fmt.Sprintf("Command completed with exit code: %d", exitCode))
	}

	return output, exitCode, nil
}

// ExecWithInput runs a command with stdin input
func (l *LocalExecutor) ExecWithInput(ctx context.Context, cmdString string, stdin io.Reader) (string, int, error) {
	if l.opts.LogCommands {
		l.logToBuffer(fmt.Sprintf("Executing with input: %s", cmdString))
	}

	cmd := exec.CommandContext(ctx, "/bin/bash", "-c", cmdString)

	// Set working directory if specified
	if l.opts.WorkDir != "" {
		cmd.Dir = l.opts.WorkDir
	}

	// Set stdin
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
			return output, -1, fmt.Errorf("failed to execute command: %w", err)
		}
	}

	return output, exitCode, nil
}

// ExecStream runs a command with streaming output
func (l *LocalExecutor) ExecStream(ctx context.Context, cmdString string, stdout, stderr io.Writer) (int, error) {
	if l.opts.LogCommands {
		l.logToBuffer(fmt.Sprintf("Executing (streaming): %s", cmdString))
	}

	cmd := exec.CommandContext(ctx, "/bin/bash", "-c", cmdString)

	// Set working directory if specified
	if l.opts.WorkDir != "" {
		cmd.Dir = l.opts.WorkDir
	}

	cmd.Stdout = stdout
	cmd.Stderr = stderr

	err := cmd.Run()

	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			return -1, fmt.Errorf("failed to execute command: %w", err)
		}
	}

	return exitCode, nil
}

// Cleanup performs cleanup operations
func (l *LocalExecutor) Cleanup(ctx context.Context) error {
	l.logToBuffer("Local executor cleanup complete")
	return nil
}

// GetLogs retrieves logs
func (l *LocalExecutor) GetLogs(ctx context.Context) (string, error) {
	l.logsMutex.Lock()
	defer l.logsMutex.Unlock()

	return strings.Join(l.logs, "\n"), nil
}

// Mode returns the execution mode
func (l *LocalExecutor) Mode() ExecutionMode {
	return ModeLocal
}

// HealthCheck verifies executor is healthy
func (l *LocalExecutor) HealthCheck(ctx context.Context) error {
	_, exitCode, err := l.Exec(ctx, "echo test")
	if err != nil || exitCode != 0 {
		return fmt.Errorf("health check failed: exit code %d, error: %v", exitCode, err)
	}
	return nil
}

// logToBuffer adds a log entry
func (l *LocalExecutor) logToBuffer(message string) {
	l.logsMutex.Lock()
	defer l.logsMutex.Unlock()
	l.logs = append(l.logs, message)
}

// GetOSInfo returns OS information from the local machine
func (l *LocalExecutor) GetOSInfo(ctx context.Context) (string, error) {
	output, exitCode, err := l.Exec(ctx, "cat /etc/os-release")
	if err != nil || exitCode != 0 {
		// Try alternative method
		output, exitCode, err = l.Exec(ctx, "uname -a")
		if err != nil || exitCode != 0 {
			return "", fmt.Errorf("failed to get OS info: %v", err)
		}
		return strings.TrimSpace(output), nil
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

// IsSudoAvailable checks if sudo is available
func (l *LocalExecutor) IsSudoAvailable(ctx context.Context) bool {
	_, exitCode, _ := l.Exec(ctx, "sudo -n true 2>&1")
	return exitCode == 0
}

// GetPackageManager detects the package manager on the local system
func (l *LocalExecutor) GetPackageManager(ctx context.Context) (string, error) {
	// Check for apt (Debian/Ubuntu)
	output, exitCode, _ := l.Exec(ctx, "which apt-get")
	if exitCode == 0 && strings.TrimSpace(output) != "" {
		return "apt", nil
	}

	// Check for dnf (RHEL/Rocky/Alma/Fedora)
	output, exitCode, _ = l.Exec(ctx, "which dnf")
	if exitCode == 0 && strings.TrimSpace(output) != "" {
		return "dnf", nil
	}

	// Check for yum (older RHEL)
	output, exitCode, _ = l.Exec(ctx, "which yum")
	if exitCode == 0 && strings.TrimSpace(output) != "" {
		return "yum", nil
	}

	return "", fmt.Errorf("no supported package manager found (apt, dnf, yum)")
}
