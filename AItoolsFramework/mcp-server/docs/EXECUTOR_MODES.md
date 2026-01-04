# Executor Modes

The test framework supports multiple execution modes to run tests in different
environments. This flexibility allows you to run the same tests locally on
your machine or in isolated Docker containers.

## Available Execution Modes

### 1. Local Mode (default)

Runs tests directly on your local machine.

**Configuration**: `config/dev.yaml`

```yaml
execution:
  mode: local
```

**Features**:

- Direct execution on host machine
- Fast execution (no container overhead)
- Requires PostgreSQL and dependencies installed locally
- Performs sudo permission checks
- Detects OS and package manager automatically

**Use Cases**:

- Development and debugging
- Quick test iterations
- When you have PostgreSQL already installed
- Testing on your specific OS configuration

**Running tests**:

```bash
make test-local
# or
make test CONFIG_FILE=config/dev.yaml
```

### 2. Container Mode (systemd-enabled)

Runs tests inside Docker containers with systemd support.

**Configuration**: `config/container.yaml`

```yaml
execution:
  mode: container-systemd
  container:
    os_image: "jrei/systemd-ubuntu:22.04"
    use_systemd: true
    skip_sudo_check: true
```

**Features**:

- Isolated test environment
- Consistent across different host machines
- Supports systemd-based services
- No local PostgreSQL installation required
- Test on different Linux distributions without dual-boot/VMs

**Use Cases**:

- Running Linux tests on macOS
- CI/CD pipelines
- Testing on specific Linux distributions
- Ensuring test consistency across team members
- Testing systemd service behavior

**Popular Container Images**:

- **Ubuntu**: `jrei/systemd-ubuntu:22.04`, `jrei/systemd-ubuntu:24.04`
- **Debian**: `jrei/systemd-debian:12`
- **Rocky Linux**: `rockylinux/rockylinux:9` (with systemd setup)
- **Alma Linux**: `almalinux/9-init`

**Running tests**:

```bash
make test-container
# or
make test CONFIG_FILE=config/container.yaml
```

## Configuration Details

### Local Executor Options

```yaml
execution:
  mode: local
  skip_sudo_check: false  # Set to true to skip sudo checks

logging:
  level: detailed
  log_commands: true
  log_output: false
```

### Container Executor Options

```yaml
execution:
  mode: container-systemd
  container:
    # Docker image (must support systemd)
    os_image: "jrei/systemd-ubuntu:22.04"

    # Enable systemd in container
    use_systemd: true

    # Skip sudo check (containers often run as root)
    skip_sudo_check: true

logging:
  level: detailed
  log_commands: true
  log_output: false
```

## How It Works

### Local Executor

The local executor (`common/executor/local.go`):

1. Checks for sudo permissions (if required)
2. Detects operating system (`/etc/os-release`)
3. Identifies package manager (apt, yum, dnf)
4. Executes commands using Go's `os/exec` package
5. Returns output, exit code, and errors

### Container Executor

The container executor (`common/executor/container.go`):

1. Checks if Docker is available
2. Pulls the specified Docker image (if not cached)
3. Starts a container with systemd enabled
4. Waits for the container to be ready
5. Executes commands inside the container using `docker exec`
6. Cleans up the container after tests complete

**Container Lifecycle**:

```
Start → Pull Image → Create Container → Start Container →
Wait for Ready → Execute Tests → Cleanup
```

**Docker Flags Used**:

- `--privileged`: Required for systemd
- `--cgroupns=host`: Share cgroup namespace
- `-v /sys/fs/cgroup:/sys/fs/cgroup:rw`: Mount cgroup filesystem
- `--tmpfs /run`: Temporary filesystem for systemd
- `--tmpfs /run/lock`: Lock files
- `-d`: Detached mode

## Example Test

Here's a simple test that works with both executors:

```go
package examples

import (
    "testing"
    "github.com/pgedge/AItoolsFramework/common/suite"
    testifySuite "github.com/stretchr/testify/suite"
)

type MyTestSuite struct {
    suite.E2ESuite
}

func (s *MyTestSuite) TestSystemInfo() {
    // This command works in both local and container mode
    output, exitCode, err := s.ExecCommand("uname -a")

    s.NoError(err)
    s.Equal(0, exitCode)
    s.NotEmpty(output)

    s.T().Logf("System: %s", output)
}

func TestMySuite(t *testing.T) {
    testifySuite.Run(t, new(MyTestSuite))
}
```

**Run locally**:

```bash
TESTFW_CONFIG=config/dev.yaml go test -v ./testcases/examples/
```

**Run in container**:

```bash
TESTFW_CONFIG=config/container.yaml go test -v ./testcases/examples/
```

The test code remains identical - only the configuration changes!

## Switching Between Modes

### Method 1: Using Makefile Targets

```bash
# Run in local mode
make test-local

# Run in container mode
make test-container

# Run specific category in container mode
make test-examples CONFIG_FILE=config/container.yaml
```

### Method 2: Using Environment Variable

```bash
# Local mode
TESTFW_CONFIG=$(pwd)/config/dev.yaml go test -v ./testcases/...

# Container mode
TESTFW_CONFIG=$(pwd)/config/container.yaml go test -v ./testcases/...
```

### Method 3: Creating Custom Config Files

Create `config/custom.yaml`:

```yaml
execution:
  mode: container-systemd
  container:
    os_image: "jrei/systemd-debian:12"
    use_systemd: true
    skip_sudo_check: true
```

Run with:

```bash
make test CONFIG_FILE=config/custom.yaml
```

## Troubleshooting

### Container Mode Issues

**Docker not available**:

```
Error: Docker not available: exec: "docker": executable file not found
```

Solution: Install Docker Desktop or Docker Engine.

**Permission denied**:

```
Error: permission denied while trying to connect to Docker daemon
```

Solution: Add your user to the docker group or run with sudo.

**Container fails to start**:

Check Docker logs:

```bash
docker logs <container-name>
```

**Image pull fails**:

Check internet connectivity and Docker Hub access.

### Local Mode Issues

**PostgreSQL not running**:

```
Error: connection refused
```

Solution: Start PostgreSQL service or switch to container mode.

**Sudo permission denied**:

Set `skip_sudo_check: true` in config if tests don't require sudo.

## Performance Comparison

**Local Mode**:

- ✅ Fastest execution
- ✅ No setup time
- ❌ Requires local dependencies
- ❌ Tests may interfere with host system

**Container Mode**:

- ✅ Isolated environment
- ✅ Consistent results
- ✅ No local dependencies needed
- ❌ Slower startup (image pull + container start)
- ❌ Slightly slower command execution

## Best Practices

1. **Use container mode for CI/CD** - Ensures consistency across environments
2. **Use local mode for development** - Faster iteration during development
3. **Test on multiple OS images** - Create configs for Ubuntu, Debian, Rocky
4. **Skip sudo checks in containers** - Containers typically run as root
5. **Clean up containers** - Framework handles this automatically
6. **Use specific image tags** - Avoid `latest` tag for reproducibility

## Future Executors

The framework is designed to support additional execution modes:

- **SSH Mode**: Run tests on remote machines via SSH
- **Kubernetes Mode**: Run tests in Kubernetes pods
- **VM Mode**: Run tests in virtual machines

These executors will follow the same interface and configuration pattern.
