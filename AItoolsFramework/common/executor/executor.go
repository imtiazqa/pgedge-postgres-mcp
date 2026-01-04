package executor

import (
	"context"
	"fmt"
	"io"
	"time"
)

// ExecutionMode defines how tests execute
type ExecutionMode int

const (
	// ModeContainerSystemd runs tests in a systemd-enabled container
	ModeContainerSystemd ExecutionMode = iota
	// ModeLocal runs tests on the local machine
	ModeLocal
	// ModeSSH runs tests on a remote machine via SSH
	ModeSSH
	// ModeKubernetes runs tests in Kubernetes pods
	ModeKubernetes
)

// String returns the string representation of the execution mode
func (m ExecutionMode) String() string {
	switch m {
	case ModeContainerSystemd:
		return "container-systemd"
	case ModeLocal:
		return "local"
	case ModeSSH:
		return "ssh"
	case ModeKubernetes:
		return "kubernetes"
	default:
		return "unknown"
	}
}

// ParseMode parses a string into an ExecutionMode
func ParseMode(mode string) (ExecutionMode, error) {
	switch mode {
	case "container-systemd":
		return ModeContainerSystemd, nil
	case "local":
		return ModeLocal, nil
	case "ssh":
		return ModeSSH, nil
	case "kubernetes":
		return ModeKubernetes, nil
	default:
		return 0, fmt.Errorf("invalid execution mode: %s", mode)
	}
}

// Executor defines the interface for command execution
type Executor interface {
	// Start initializes the executor
	Start(ctx context.Context) error

	// Exec runs a command and returns output, exit code, and error
	Exec(ctx context.Context, cmd string) (string, int, error)

	// ExecWithInput runs a command with stdin input
	ExecWithInput(ctx context.Context, cmd string, stdin io.Reader) (string, int, error)

	// ExecStream runs a command with streaming output
	ExecStream(ctx context.Context, cmd string, stdout, stderr io.Writer) (int, error)

	// Cleanup performs cleanup operations
	Cleanup(ctx context.Context) error

	// GetLogs retrieves executor logs
	GetLogs(ctx context.Context) (string, error)

	// Mode returns the execution mode
	Mode() ExecutionMode

	// GetOSInfo returns OS information
	GetOSInfo(ctx context.Context) (string, error)

	// HealthCheck verifies executor is healthy
	HealthCheck(ctx context.Context) error
}

// ExecutorOptions configures executor behavior
type ExecutorOptions struct {
	// Timeout for operations
	Timeout time.Duration

	// Working directory
	WorkDir string

	// Environment variables
	Env map[string]string

	// Logging
	LogCommands bool
	LogOutput   bool

	// Retry configuration
	RetryEnabled bool
	RetryCount   int
	RetryDelay   time.Duration

	// Skip sudo check (useful for simple tests that don't need privileged access)
	SkipSudoCheck bool
}

// DefaultOptions returns default executor options
func DefaultOptions() *ExecutorOptions {
	return &ExecutorOptions{
		Timeout:      2 * time.Minute,
		WorkDir:      "",
		Env:          make(map[string]string),
		LogCommands:  false,
		LogOutput:    false,
		RetryEnabled: false,
		RetryCount:   0,
		RetryDelay:   0,
	}
}

// NewExecutor creates an executor based on the execution mode
func NewExecutor(mode ExecutionMode, osImage string, opts *ExecutorOptions) (Executor, error) {
	if opts == nil {
		opts = DefaultOptions()
	}

	switch mode {
	case ModeContainerSystemd:
		return NewContainerExecutor(osImage, true, opts)
	case ModeLocal:
		return NewLocalExecutor(opts)
	case ModeSSH:
		return nil, fmt.Errorf("SSH executor not yet implemented")
	case ModeKubernetes:
		return nil, fmt.Errorf("Kubernetes executor not yet implemented")
	default:
		return nil, fmt.Errorf("unsupported execution mode: %v", mode)
	}
}
