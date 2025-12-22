# MCP Server Regression Tests

Automated regression testing for pgEdge Postgres MCP Server package installation.

## Overview

This test suite validates:

1. ✅ Repository installation
2. ✅ PostgreSQL installation and configuration
3. ✅ MCP server package installation from repository
4. ✅ Installation validation (files, permissions, services)
5. ✅ Token management commands
6. ✅ User management commands

## Prerequisites

- Docker installed and running
- Go 1.21+ installed
- Internet connection (for pulling Docker images)

## Quick Start

```bash
# Navigate to test directory
cd test/regression

# Initialize Go module
go mod init pgedge-postgres-mcp/test/regression
go mod tidy

# Run tests on Debian 12 (default)
make test

# Run tests on all platforms
make test-all
```

## Running Tests

### Test on Specific OS

```bash
# Debian 12
make test-debian

# Rocky Linux 9
make test-rocky

# Ubuntu 22.04
make test-ubuntu
```

### Test Individual Cases

```bash
# Run specific test
make test-one TEST=Test01_RepositoryInstallation
make test-one TEST=Test02_PackageInstallation
make test-one TEST=Test03_InstallationValidation
make test-one TEST=Test04_TokenManagement
make test-one TEST=Test05_UserManagement
```

### Custom OS Image

```bash
# Test on custom Docker image
TEST_OS_IMAGE=debian:11 go test -v -timeout 20m
TEST_OS_IMAGE=almalinux:9 go test -v -timeout 20m
```

## Test Workflow

Each test:

1. **Pulls** Docker image (debian:12, rockylinux:9, etc.)
2. **Starts** clean container
3. **Installs** pgEdge repository
4. **Installs** MCP server package from repo
5. **Validates** installation
6. **Runs** test cases
7. **Removes** container

Total time per OS: ~5-10 minutes

## Test Cases

### Test 01: Repository Installation

- Adds pgEdge package repository
- Installs EPEL repository (RHEL/Rocky)
- Verifies repository configuration
- Confirms packages are available

### Test 02: PostgreSQL Installation and Configuration

- Installs pgEdge PostgreSQL 16 packages
- Initializes PostgreSQL database
- Starts PostgreSQL service
- Sets postgres user password
- Creates MCP database

### Test 03: MCP Server Package Installation

- Installs `pgedge-postgres-mcp` from repository
- Installs `pgedge-nla-cli` package
- Installs `pgedge-nla-web` package
- Installs `pgedge-postgres-mcp-kb` package
- Updates `postgres-mcp.yaml` configuration
- Updates `postgres-mcp.env` configuration

### Test 04: Installation Validation

- Binary exists at `/usr/bin/pgedge-postgres-mcp`
- Config directory exists at `/etc/pgedge`
- Config files exist (`postgres-mcp.yaml`, `postgres-mcp.env`)
- Systemd service file installed at `/usr/lib/systemd/system/`
- Data directory exists at `/var/lib/pgedge/postgres-mcp`

### Test 05: Token Management

- Creates API token with `add-token` command
- Lists tokens with `list-tokens` command
- Verifies token file created correctly
- Validates token format

### Test 06: User Management

- Creates user with `add-user` command
- Lists users with `list-users` command
- Verifies user file created correctly
- Checks file permissions

## Configuration

### Environment Variables

```bash
# Docker image to test
export TEST_OS_IMAGE=debian:12

# Repository URL
export PGEDGE_REPO_URL=https://apt.pgedge.com
```

## Cleanup

```bash
# Remove any orphaned test containers
make clean

# Or manually
docker ps -a | grep mcp-test | awk '{print $1}' | xargs docker rm -f
```

## Troubleshooting

### Container fails to start

```bash
# Check Docker is running
docker ps

# Check image exists
docker images | grep debian
```

### Tests timeout

```bash
# Increase timeout
go test -v -timeout 30m
```

### See container logs on failure

Test automatically prints container logs when a test fails.

### Manual debugging

```bash
# Start container manually
docker run -it --rm debian:12 /bin/bash

# Run commands from test manually
apt-get update
apt-get install -y pgedge-postgres-mcp
```

## Adding New Tests

```go
func (s *RegressionTestSuite) Test06_YourNewTest() {
    s.T().Log("TEST 06: Your test description")

    // Install package first
    s.Test02_PackageInstallation()

    // Your test logic
    output, exitCode, err := s.container.Exec(s.ctx, "your-command")
    s.NoError(err)
    s.Equal(0, exitCode)

    s.T().Log("✓ Test passed")
}
```

## CI/CD Integration

```yaml
# .github/workflows/regression.yml
name: Regression Tests
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      - name: Run regression tests
        run: cd test/regression && make test-all
```

## Expected Output

```
=== RUN   TestRegressionSuite
=== RUN   TestRegressionSuite/Test01_RepositoryInstallation
    suite_test.go:XX: TEST 01: Installing pgEdge repository
    suite_test.go:XX: ✓ Repository installed successfully
=== RUN   TestRegressionSuite/Test02_PackageInstallation
    suite_test.go:XX: TEST 02: Installing MCP server package
    suite_test.go:XX: ✓ Package installed successfully
=== RUN   TestRegressionSuite/Test03_InstallationValidation
    suite_test.go:XX: TEST 03: Validating MCP server installation
    suite_test.go:XX: ✓ All installation validation checks passed
=== RUN   TestRegressionSuite/Test04_TokenManagement
    suite_test.go:XX: TEST 04: Testing token management commands
    suite_test.go:XX: ✓ Token management working correctly
=== RUN   TestRegressionSuite/Test05_UserManagement
    suite_test.go:XX: TEST 05: Testing user management commands
    suite_test.go:XX: ✓ User management working correctly
--- PASS: TestRegressionSuite (127.45s)
    --- PASS: TestRegressionSuite/Test01_RepositoryInstallation (23.12s)
    --- PASS: TestRegressionSuite/Test02_PackageInstallation (28.34s)
    --- PASS: TestRegressionSuite/Test03_InstallationValidation (31.45s)
    --- PASS: TestRegressionSuite/Test04_TokenManagement (22.11s)
    --- PASS: TestRegressionSuite/Test05_UserManagement (22.43s)
PASS
ok      pgedge-postgres-mcp/test/regression    127.456s
```

## File Structure

```
test/regression/
├── docker_helper.go    # Docker container management
├── suite_test.go       # 5 basic test cases
├── Makefile            # Test execution commands
└── README.md           # This file
```

## Next Steps

To expand the test suite, you can add:

- Service management tests (systemd start/stop/restart)
- HTTP mode functionality tests
- Database connection tests
- Configuration file validation
- Package upgrade/downgrade tests
- Package removal tests
- Security/permission tests
- Performance benchmarks

See the "Adding New Tests" section above for examples.
