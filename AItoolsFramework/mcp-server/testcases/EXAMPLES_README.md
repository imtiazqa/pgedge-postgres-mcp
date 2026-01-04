# Example Tests

This directory contains example tests demonstrating the AItoolsFramework test
capabilities. These examples show how to write tests that work in both local
and container execution modes.

## Available Examples

### 1. Basic Examples ([example_test.go](example_test.go))

Demonstrates fundamental framework features:

- System information retrieval
- Command execution
- File and directory operations
- Eventually helper for async operations
- Configuration access

**Run locally**:

```bash
make test-examples
# or
TESTFW_CONFIG=$(pwd)/config/dev.yaml go test -v ./testcases/examples/
```

### 2. Container Examples ([container_example_test.go](container_example_test.go))

Demonstrates container-specific testing:

- Container mode detection
- OS detection in containers
- Systemd functionality testing
- Package manager availability
- File operations in containers
- Network connectivity
- User permissions
- Environment variables

**Run in container**:

```bash
make test-container-examples
# or
TESTFW_CONFIG=$(pwd)/config/container.yaml go test -v ./testcases/examples/
```

## Writing Your Own Tests

### Basic Test Structure

```go
package examples

import (
    "testing"
    "github.com/pgedge/AItoolsFramework/common/suite"
    testifySuite "github.com/stretchr/testify/suite"
)

type MyTestSuite struct {
    suite.E2ESuite  // Inherit framework capabilities
}

func (s *MyTestSuite) TestMyFeature() {
    // Your test code here
    output, exitCode, err := s.ExecCommand("echo 'Hello'")
    s.NoError(err)
    s.Equal(0, exitCode)
    s.Contains(output, "Hello")
}

func TestMySuite(t *testing.T) {
    testifySuite.Run(t, new(MyTestSuite))
}
```

### Available Suite Types

The framework provides three base suite types:

1. **E2ESuite** - System-level end-to-end testing
    - Command execution
    - File/directory assertions
    - System information
    - Eventually helper for async operations

2. **DatabaseSuite** - Database testing
    - All E2ESuite capabilities
    - Database connection management
    - SQL execution
    - Transaction handling

3. **APISuite** - API testing
    - All E2ESuite capabilities
    - HTTP request helpers
    - MCP protocol support
    - JSON-RPC communication

### Common Test Patterns

#### Command Execution

```go
func (s *MyTestSuite) TestCommand() {
    output, exitCode, err := s.ExecCommand("ls -la")
    s.NoError(err)
    s.Equal(0, exitCode)
}
```

#### File Assertions

```go
func (s *MyTestSuite) TestFiles() {
    s.AssertFileExists("/path/to/file")
    s.AssertDirectoryExists("/path/to/dir")
}
```

#### Async Operations

```go
func (s *MyTestSuite) TestAsync() {
    // Start background process
    s.ExecCommand("(sleep 2 && touch /tmp/delayed) &")

    // Wait for condition
    result := s.Eventually(func() bool {
        _, code, _ := s.ExecCommand("test -f /tmp/delayed")
        return code == 0
    }, 5*time.Second, 500*time.Millisecond)

    s.True(result)
}
```

#### Configuration Access

```go
func (s *MyTestSuite) TestConfig() {
    // Access any config value
    s.T().Logf("Environment: %s", s.Config.Environment)
    s.T().Logf("Mode: %s", s.Config.Execution.Mode)
}
```

## Running Examples

### Local Mode (Default)

Tests run directly on your machine:

```bash
# Run all examples
make test-examples

# Run specific test
make test-run TEST=TestSystemInfo

# Run with verbose output
TESTFW_CONFIG=$(pwd)/config/dev.yaml go test -v ./testcases/examples/
```

### Container Mode

Tests run inside Docker containers:

```bash
# Run all examples in container
make test-container-examples

# Run with specific container image
# Edit config/container.yaml to change image
make test-container-examples

# Run specific test in container
TESTFW_CONFIG=$(pwd)/config/container.yaml go test -v ./testcases/examples/ -run TestContainerMode
```

### Different OS Images

To test on different Linux distributions, edit [config/container.yaml](../../config/container.yaml):

```yaml
execution:
  mode: container-systemd
  container:
    os_image: "jrei/systemd-ubuntu:24.04"  # Change this line
    use_systemd: true
    skip_sudo_check: true
```

Available systemd-enabled images:

- Ubuntu: `jrei/systemd-ubuntu:22.04`, `jrei/systemd-ubuntu:24.04`
- Debian: `jrei/systemd-debian:12`
- Rocky Linux: `rockylinux/rockylinux:9`
- Alma Linux: `almalinux/9-init`

## Tips and Best Practices

1. **Write mode-agnostic tests** - Tests should work in both local and
   container modes
2. **Use config values** - Don't hardcode paths or values
3. **Clean up resources** - Use `TearDownTest()` to clean up
4. **Use descriptive names** - Test names should clearly describe what they test
5. **Log useful information** - Use `s.T().Logf()` for debugging
6. **Handle both Debian and RHEL** - Tests should work on both OS families

## Troubleshooting

### Test Fails in Local Mode

```
Error: connection refused
```

Check if PostgreSQL is running or switch to container mode:

```bash
make test-container-examples
```

### Container Mode Slow

First run pulls Docker image (slow), subsequent runs are faster.

Pre-pull the image:

```bash
docker pull jrei/systemd-ubuntu:22.04
```

### Permission Denied

In container.yaml, ensure:

```yaml
execution:
  container:
    skip_sudo_check: true
```

## Further Reading

- [Executor Modes Documentation](../../docs/EXECUTOR_MODES.md)
- [Framework Overview](../../docs/FRAMEWORK_OVERVIEW.md)
- [Writing Tests Guide](../../docs/WRITING_TESTS.md)
