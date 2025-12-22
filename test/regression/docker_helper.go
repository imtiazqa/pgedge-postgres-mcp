package regression

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

// SimpleContainer manages a test Docker container
type SimpleContainer struct {
	cli         *client.Client
	containerID string
	image       string
	name        string
}

// NewContainer creates a new test container
func NewContainer(image string) (*SimpleContainer, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client: %w", err)
	}

	return &SimpleContainer{
		cli:   cli,
		image: image,
		name:  fmt.Sprintf("mcp-test-%d", time.Now().Unix()),
	}, nil
}

// Start pulls image and starts container
func (c *SimpleContainer) Start(ctx context.Context) error {
	// Pull image
	reader, err := c.cli.ImagePull(ctx, c.image, types.ImagePullOptions{})
	if err != nil {
		return fmt.Errorf("failed to pull image: %w", err)
	}
	io.Copy(io.Discard, reader)
	reader.Close()

	// Create and start container
	resp, err := c.cli.ContainerCreate(ctx,
		&container.Config{
			Image: c.image,
			Cmd:   []string{"/bin/bash", "-c", "sleep infinity"},
			Tty:   true,
		},
		&container.HostConfig{
			Privileged: true, // For systemd
		},
		nil, nil, c.name)
	if err != nil {
		return fmt.Errorf("failed to create container: %w", err)
	}

	c.containerID = resp.ID

	if err := c.cli.ContainerStart(ctx, c.containerID, types.ContainerStartOptions{}); err != nil {
		return fmt.Errorf("failed to start container: %w", err)
	}

	// Wait a bit for container to be ready
	time.Sleep(2 * time.Second)
	return nil
}

// Exec runs a command in the container
func (c *SimpleContainer) Exec(ctx context.Context, cmd string) (string, int, error) {
	execConfig := types.ExecConfig{
		AttachStdout: true,
		AttachStderr: true,
		Cmd:          []string{"/bin/bash", "-c", cmd},
	}

	execID, err := c.cli.ContainerExecCreate(ctx, c.containerID, execConfig)
	if err != nil {
		return "", 0, fmt.Errorf("failed to create exec: %w", err)
	}

	resp, err := c.cli.ContainerExecAttach(ctx, execID.ID, types.ExecStartCheck{})
	if err != nil {
		return "", 0, fmt.Errorf("failed to attach exec: %w", err)
	}
	defer resp.Close()

	output, err := io.ReadAll(resp.Reader)
	if err != nil {
		return "", 0, fmt.Errorf("failed to read output: %w", err)
	}

	// Get exit code
	inspect, err := c.cli.ContainerExecInspect(ctx, execID.ID)
	if err != nil {
		return string(output), 0, err
	}

	return string(output), inspect.ExitCode, nil
}

// Cleanup stops and removes the container
func (c *SimpleContainer) Cleanup(ctx context.Context) error {
	timeout := 5
	if err := c.cli.ContainerStop(ctx, c.containerID, container.StopOptions{Timeout: &timeout}); err != nil {
		// Ignore stop errors
	}
	return c.cli.ContainerRemove(ctx, c.containerID, types.ContainerRemoveOptions{Force: true})
}

// GetLogs retrieves container logs
func (c *SimpleContainer) GetLogs(ctx context.Context) (string, error) {
	reader, err := c.cli.ContainerLogs(ctx, c.containerID, types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Tail:       "100",
	})
	if err != nil {
		return "", err
	}
	defer reader.Close()

	logs, err := io.ReadAll(reader)
	return string(logs), err
}
