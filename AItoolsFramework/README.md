# AItoolsFramework - Enterprise Test Automation Framework

A comprehensive, reusable, multi-project test automation framework in Go using Testify.

## Overview

AItoolsFramework is designed to test multiple AI tools/projects with:
- **Zero hardcoded values** - all configuration in YAML
- **Multi-project support** - each project has its own tests and config
- **Reusable utilities** - shared test framework in `/common`
- **Future-proof** - easy to add new projects and test types

## Directory Structure

```
/AItoolsFramework/
├── common/              # Shared framework utilities
│   ├── config/          # Configuration management
│   ├── executor/        # Command execution (local, container)
│   ├── suite/           # Base test suites
│   ├── fixtures/        # Fixture management
│   ├── assertions/      # Custom assertions
│   ├── reporters/       # Test reporting
│   └── utils/           # Shared utilities
│
├── mcp-server/          # MCP Server tests
│   ├── config/          # MCP-specific configuration
│   ├── suites/          # Test suites
│   ├── fixtures/        # Test data
│   └── utils/           # MCP-specific utilities
│
└── future-project/      # Template for new projects
```

## Quick Start

### 1. Configure Your Tests

Edit `mcp-server/config/dev.yaml`:

```yaml
environment: dev

execution:
  mode: local              # or container-systemd

database:
  host: localhost
  port: 5432
  user: postgres
  password: ${DB_PASSWORD}  # Environment variable

reporting:
  log_level: detailed
  console: true
```

### 2. Run Tests

```bash
cd mcp-server
TESTFW_CONFIG=config/dev.yaml go test ./suites/...
```

Or use the Makefile:

```bash
make test
```

## Key Features

### Fully Configurable

ALL values are in YAML configuration:
- Repository URLs and credentials
- Package names and install commands
- API endpoints
- Database connections
- Test data and fixtures
- Timeouts and retries

### Multi-Project Architecture

Each project has:
- Own configuration files (dev.yaml, staging.yaml, prod.yaml)
- Own test suites
- Own fixtures and test data
- Own project-specific utilities

But shares:
- Common framework utilities
- Base test suites
- Executors
- Reporters

### Environment Variable Overrides

Use `${VAR}` or `${VAR:-default}` syntax:

```yaml
database:
  host: ${DB_HOST:-localhost}
  password: ${DB_PASSWORD}
```

### Reusable Base Suites

```go
type MyTestSuite struct {
    suite.E2ESuite  // Inherits common functionality
}

func (s *MyTestSuite) TestExample() {
    // Use framework helpers
    output, exitCode, err := s.ExecCommand("echo test")
    s.NoError(err)
    s.Equal(0, exitCode)
}
```

## Adding a New Project

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
mkdir -p {config,suites,fixtures,utils}
```

4. Create `config/dev.yaml` with your configuration

5. Write tests using the framework suites

See [mcp-server/](mcp-server/) for a complete example.

## Documentation

- [Common Framework](common/README.md) - Shared utilities documentation
- [MCP Server Tests](mcp-server/README.md) - MCP server test guide
- [Configuration Guide](docs/configuration.md) - Configuration reference
- [Best Practices](docs/best-practices.md) - Testing best practices

## Requirements

- Go 1.21 or later
- Local mode: sudo access (for package installation)
- Container mode: Docker (not yet implemented in Week 1)

## Current Status

### Week 1 ✅ Complete
- Core framework structure
- Configuration management with YAML + env vars
- Local executor implementation
- Base and E2E test suites
- Example tests running

### Week 2 ✅ Complete
- Fixture management system with dependency resolution
- Database utilities (PostgreSQL helper, seeder)
- DatabaseSuite base with assertions
- Database-specific assertions
- Example database tests
- SQL fixtures (schema.sql, testdata.sql)

### Week 3 ✅ Complete
- MCP protocol utilities (message builders, validators)
- MCP server helper for managing server lifecycle
- MCP-specific assertions (tools, resources, prompts)
- HTTP test utilities (client with fluent API)
- HTTP-specific assertions
- APISuite base for API/MCP testing
- Example MCP protocol tests

### In Progress / Future
⏳ Container executor (Week 4)
⏳ Console reporter enhancements (Week 4)
⏳ Additional reporters (JSON, JUnit - Week 4)
⏳ Complete integration examples (Week 4)

## License

Copyright © 2026 pgEdge. All rights reserved.
