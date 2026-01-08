# Test Framework Architecture

This document describes the organization and structure of the AItoolsFramework
test framework.

## Directory Structure

```
AItoolsFramework/
├── common/                 # Shared framework components (Go module)
│   ├── assertions/        # Custom test assertions
│   ├── config/            # Configuration management
│   ├── database/          # Database helpers
│   ├── executor/          # Test execution (local, container-systemd)
│   ├── http/              # HTTP client helpers
│   ├── mcp/               # MCP protocol helpers
│   └── suite/             # Test suite base classes ⭐
│       ├── README.md      # Suite documentation
│       ├── base.go        # BaseSuite
│       ├── e2e.go         # E2ESuite (public API)
│       ├── install.go     # E2ESuite (installation impl)
│       ├── api.go         # APISuite
│       └── database.go    # DatabaseSuite
│
└── mcp-server/            # MCP server test project (Go module)
    ├── config/            # Test configurations
    │   ├── local.yaml       # Local development
    │   └── container.yaml   # Container mode
    └── testcases/         # Actual tests
        ├── database/      # Database tests
        ├── installation/  # Installation tests
        ├── service/       # Service management tests
        ├── mcp/           # MCP protocol tests
        ├── kb/            # Knowledge base tests
        └── examples/      # Example tests
```

## Multi-Module Architecture

### Why Two Go Modules?

The framework uses Go's multi-module workspace pattern:

**1. `common/` Module**
- Independent, reusable framework library
- Can be versioned separately
- Shareable across multiple test projects
- Contains all base functionality

**2. `mcp-server/` Module**
- Specific test suite for MCP server
- Uses `common` via `replace` directive  
- Has its own dependencies
- Focused on MCP server testing

**Benefits:**
- **Reusability**: Other projects can use the same framework
- **Isolation**: Changes in test suites don't affect the framework  
- **Local Development**: `replace` directive allows immediate code changes
- **Scalability**: Easy to add more test projects

### Module Relationship

\`\`\`go
// mcp-server/go.mod
module github.com/pgedge/AItoolsFramework/mcp-server

replace github.com/pgedge/AItoolsFramework/common => ../common

require (
    github.com/pgedge/AItoolsFramework/common v0.0.0
    github.com/stretchr/testify v1.11.1
)
\`\`\`

## Configuration System

### Logging vs Reporting Separation

The framework separates logging control from reporting configuration:

**logging**: Controls verbosity and what gets logged
- `level`: minimal, detailed, verbose
- `log_commands`: Log executed commands
- `log_output`: Log command output

**reporting**: Controls output formats and where reports are saved
- `console`, `json`, `junit`, `markdown`: Output formats
- `output_paths`: File locations for reports
- `console_settings`: Display preferences

\`\`\`yaml
logging:
  level: minimal
  log_commands: true
  log_output: false

reporting:
  console: true
  json: false
  junit: false
  output_paths:
    log_file: test-results/test-local.log
\`\`\`

## Key Design Decisions

### 1. Suite File Organization

**Decision**: Split E2ESuite across two files (e2e.go + install.go)

**Rationale**:
- Prevents single file from becoming too large  
- Clear separation: public API vs implementation
- Installation logic (~300 lines) isolated for maintenance
- Follows Go convention of multi-file types

### 2. Global Installation State

**Decision**: Use package-level `globalInstallState` instead of per-suite state

**Rationale**:
- Ensures packages are installed only once across all test suites
- Saves ~2 minutes per test run
- Thread-safe with `sync.Mutex`
- Idempotent: first suite installs, others skip instantly

See [OPTIMIZATION.md](OPTIMIZATION.md) for details.

### 3. Configuration-Driven Installation

**Decision**: All installation parameters from config files

**Rationale**:
- No hardcoded values in test code
- Easy to test different PostgreSQL versions
- Supports different environments  
- Repository URLs automatically selected based on `server_env`

### 4. Dependency Chain Management

**Decision**: Automatic dependency resolution in `Ensure*` methods

**Rationale**:
- Tests don't need to know installation order
- Simple API: just call `EnsureMCPPackagesInstalled()`
- Framework handles: repo → PostgreSQL → MCP packages

## Suite Hierarchy

\`\`\`
BaseSuite (base.go)
    │   - Purpose: Core functionality for all suites
    │   - Features: Config access, executor, logging
    │
    ├─→ E2ESuite (e2e.go + install.go)
    │   │   - Purpose: End-to-end testing with system dependencies
    │   │   - Features: File assertions, installation helpers
    │   │   - Used by: Installation, service, files tests
    │   │
    │   └─→ APISuite (api.go)
    │       - Purpose: API/MCP protocol testing
    │       - Features: MCP server lifecycle, protocol assertions
    │       - Inherits: Installation methods from E2ESuite
    │       - Used by: MCP protocol, stdio, KB tests
    │
    └─→ DatabaseSuite (database.go)
        - Purpose: Database-specific testing
        - Features: Connection management, seeding, cleanup
        - Used by: Database tests
\`\`\`

## Test Execution Modes

### 1. Local Mode (local.yaml)
- Runs on local machine
- Fast execution  
- Uses existing or installs dependencies locally
- Suitable for development

\`\`\`yaml
execution:
  mode: local
  skip_sudo_check: true
\`\`\`

### 2. Container Mode (container.yaml)
- Runs in Docker container with systemd
- Fresh environment per test run
- Auto-installs dependencies
- Isolated from host system
- Perfect for CI/CD

\`\`\`yaml
execution:
  mode: container-systemd
  container:
    os_image: "jrei/systemd-ubuntu:22.04"
    use_systemd: true
    skip_sudo_check: true
\`\`\`

## Best Practices

### 1. File Placement
- Suite base classes → `common/suite/`
- Shared utilities → `common/{assertions,database,http,mcp}/`
- Actual tests → `mcp-server/testcases/{category}/`
- Configuration → `mcp-server/config/`

### 2. Installation Patterns
- Always use `Ensure*` methods (never direct installation)
- Call appropriate `Ensure*` at start of each test method
- Let framework handle dependency chains
- Don't assume pre-installed software in container mode

### 3. Configuration Management
- Use `${VAR:-default}` for environment variables
- Keep sensitive data in environment variables, not config files
- Use `server_env: live/staging` to switch repositories

### 4. Logging
- Use `minimal` for CI/CD (clean output)
- Use `detailed` for debugging
- Use `verbose` for maximum information

## Future Enhancements

1. **Parallel Test Execution**: Run independent tests concurrently
2. **Installation Cache**: Reuse installed packages across test runs
3. **Snapshot/Restore**: Save container state after installation
4. **Version Matrix Testing**: Test against multiple PostgreSQL versions
5. **Remote Execution**: SSH-based executor for remote testing

## Documentation

- [README](README.md) - Main project documentation
- [OPTIMIZATION](OPTIMIZATION.md) - Performance optimizations
- [Test Cases](mcp-server/testcases/README.md) - Test guide
