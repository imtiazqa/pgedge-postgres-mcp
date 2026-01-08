# AItoolsFramework - Enterprise Test Automation Framework

A comprehensive, reusable, multi-project test automation framework in Go using
Testify.

## Overview

AItoolsFramework is designed to test multiple AI tools/projects with:

- **Zero hardcoded values** - all configuration in YAML
- **Multi-project support** - each project has its own tests and config
- **Reusable utilities** - shared test framework in `/common`
- **Multiple execution modes** - local or containerized testing
- **Future-proof** - easy to add new projects and test types

## Directory Structure

```
/AItoolsFramework/
├── common/              # Shared framework utilities (Go module)
│   ├── config/          # Configuration management
│   ├── executor/        # Command execution (local, container-systemd)
│   ├── suite/           # Base test suites
│   ├── database/        # Database helpers
│   ├── http/            # HTTP client utilities
│   ├── mcp/             # MCP protocol helpers
│   ├── assertions/      # Custom assertions
│   └── utils/           # Shared utilities
│
└── mcp-server/          # MCP Server test suite (Go module)
    ├── config/          # Test configurations
    │   ├── local.yaml       # Local execution
    │   └── container.yaml   # Container execution
    ├── testcases/       # Test implementations
    │   ├── installation/    # Installation tests
    │   ├── database/        # Database tests
    │   ├── service/         # Service tests
    │   ├── mcp/             # MCP protocol tests
    │   ├── kb/              # Knowledge base tests
    │   └── examples/        # Example tests
    └── Makefile         # Test runner targets
```

## Quick Start

### 1. Configure Your Tests

Choose a configuration file:

- **Local mode** - `mcp-server/config/local.yaml`
- **Container mode** - `mcp-server/config/container.yaml`

Example configuration:

```yaml
environment: local

execution:
  mode: local              # or container-systemd
  server_env: live         # or staging

postgresql:
  version: "17"

logging:
  level: minimal           # minimal, detailed, verbose
  log_commands: true
  log_output: false

database:
  host: localhost
  port: 5432
  user: postgres
  password: ${DB_PASSWORD:-postgres123}
```

### 2. Run Tests

Using Makefile (recommended):

```bash
cd mcp-server

# Run in local mode
make test-local

# Run in container mode
make test-container

# Run specific test category
make test-installation
make test-database
make test-mcp
```

Or directly with go test:

```bash
cd mcp-server
TESTFW_CONFIG=config/local.yaml go test -v ./testcases/...
```

## Key Features

### Fully Configurable

ALL values are in YAML configuration:

- Repository URLs and credentials
- Package names and install commands
- PostgreSQL version
- Database connections
- Timeouts and retry settings
- Logging and reporting options

### Multi-Module Architecture

The framework uses Go's multi-module workspace pattern:

- **`common/`** - Reusable framework library (independent Go module)
- **`mcp-server/`** - MCP server test suite (uses `common` via replace
  directive)
- Future projects can reuse the same `common` framework

### Multiple Execution Modes

#### Local Mode
- Runs tests on your local machine
- Faster execution
- Requires manual setup or uses existing installation

#### Container Mode (systemd)
- Runs tests in Docker container with systemd support
- Fresh, isolated environment
- Automatic dependency installation
- Perfect for CI/CD pipelines

### Environment Variable Overrides

Use `${VAR}` or `${VAR:-default}` syntax in configuration:

```yaml
database:
  host: ${DB_HOST:-localhost}
  password: ${DB_PASSWORD}
  port: ${DB_PORT:-5432}
```

### Flexible Logging

Control output verbosity via `logging.level`:

- **minimal** - Summary only (clean output for CI)
- **detailed** - Full command logs and output
- **verbose** - Maximum debugging information

### Reusable Base Suites

```go
type MyTestSuite struct {
    suite.E2ESuite  // Inherits common functionality
}

func (s *MyTestSuite) SetupSuite() {
    s.E2ESuite.SetupSuite()
}

func (s *MyTestSuite) TestExample() {
    // Use framework helpers
    s.EnsureMCPPackagesInstalled()

    output, exitCode, err := s.ExecCommand("echo test")
    s.NoError(err)
    s.Equal(0, exitCode)

    s.AssertFileExists("/usr/bin/pgedge-postgres-mcp")
}
```

## Switching Between Live and Staging

Edit the configuration file and change `server_env`:

```yaml
# For live repositories
execution:
  server_env: live

# For staging repositories
execution:
  server_env: staging
```

The framework automatically selects the correct repository URLs based on this
setting.

## Test Suite Types

### E2ESuite
- **Purpose**: End-to-end system testing
- **Features**: Installation helpers, file assertions, command execution
- **Used by**: Installation, service, file verification tests

### APISuite (extends E2ESuite)
- **Purpose**: API and MCP protocol testing
- **Features**: MCP server lifecycle, protocol assertions, HTTP helpers
- **Used by**: MCP protocol, stdio, knowledge base tests

### DatabaseSuite
- **Purpose**: Database-specific testing
- **Features**: Connection management, seeding, schema operations
- **Used by**: Database tests

## Adding a New Test Project

1. Create project directory:
```bash
mkdir new-project
cd new-project
```

2. Initialize Go module:
```bash
go mod init github.com/pgedge/AItoolsFramework/new-project
go mod edit -replace github.com/pgedge/AItoolsFramework/common=../common
```

3. Create directory structure:
```bash
mkdir -p {config,testcases,docs}
```

4. Create configuration files in `config/`

5. Write tests using the framework suites

See [mcp-server/](mcp-server/) for a complete example.

## Documentation

- [Architecture](ARCHITECTURE.md) - Framework architecture and design
- [Optimization](OPTIMIZATION.md) - Performance optimizations
- [Common Framework](common/suite/README.md) - Suite documentation
- [Test Cases](mcp-server/testcases/README.md) - MCP server test guide

## Requirements

- Go 1.21 or later
- Local mode: sudo access (for package installation)
- Container mode: Docker with systemd support

## Makefile Targets

From `mcp-server/` directory:

```bash
make test-local          # Run tests locally
make test-container      # Run tests in container
make test-installation   # Run installation tests only
make test-database       # Run database tests only
make test-service        # Run service tests only
make test-mcp            # Run MCP protocol tests only
make test-kb             # Run knowledge base tests only
make clean               # Clean test cache and results
make cleanup             # Full cleanup (packages, containers)
make help                # Show all available targets
```

## Current Status

### ✅ Complete
- Core framework structure with multi-module support
- Configuration management with YAML + environment variables
- Local and container-systemd executor implementations
- Base test suites (E2ESuite, APISuite, DatabaseSuite)
- Installation automation with dependency management
- Global state optimization for faster test runs
- MCP protocol utilities and helpers
- Database utilities and assertions
- HTTP client and API testing support
- Comprehensive test coverage for MCP server
- Logging and reporting configuration

### Framework Features
- **Configuration**: YAML-based with environment variable support
- **Execution Modes**: Local and container-systemd
- **Test Suites**: E2ESuite, APISuite, DatabaseSuite
- **Utilities**: Database, HTTP, MCP protocol helpers
- **Assertions**: File, database, MCP-specific assertions
- **Reporting**: Configurable logging levels and output formats
- **Optimization**: Global installation state for performance

## License

Copyright © 2026 pgEdge. All rights reserved.
