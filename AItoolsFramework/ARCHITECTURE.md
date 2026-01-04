# Test Framework Architecture

This document describes the organization and structure of the AItoolsFramework test framework.

## Directory Structure

```
AItoolsFramework/
├── common/                 # Shared framework components
│   ├── assertions/        # Custom test assertions
│   ├── config/            # Configuration management
│   ├── database/          # Database helpers
│   ├── executor/          # Test execution (local, container)
│   ├── fixtures/          # Test data management
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
└── mcp-server/            # MCP server test project
    ├── config/            # Test configurations
    │   ├── dev.yaml       # Local development
    │   └── container.yaml # Container mode
    └── testcases/         # Actual tests
        ├── database/      # Database tests
        ├── installation/  # Installation tests
        ├── service/       # Service management tests
        ├── mcp/           # MCP protocol tests
        ├── kb/            # Knowledge base tests
        └── examples/      # Example tests
```

## Key Design Decisions

### 1. Suite File Organization (common/suite)

**Decision**: Split E2ESuite across two files (e2e.go + install.go)

**Rationale**:
- Prevents single file from becoming too large
- Clear separation: public API vs implementation
- Installation logic (~300 lines) isolated for easier maintenance
- Follows Go convention of multi-file types in same package

**Files**:
- `e2e.go`: Public methods (`EnsureMCPPackagesInstalled()`, assertion helpers)
- `install.go`: Private implementation (`installRepository()`, OS-specific logic)

### 2. Installation State Tracking

**Decision**: Use `setupState` struct to track installed components

**Rationale**:
- Ensures idempotent installations (install only once)
- Avoids redundant package installations
- Tracks dependency chain (repo → PostgreSQL → MCP packages)

**Implementation**:
```go
setupState struct {
    repoInstalled        bool
    postgresqlInstalled  bool
    mcpPackagesInstalled bool
}
```

### 3. Configuration-Driven Installation

**Decision**: PostgreSQL version and settings from config files

**Rationale**:
- No hardcoded values in test code
- Easy to test different PostgreSQL versions
- Supports different environments (dev, staging, container)

**Example**:
```yaml
postgresql:
  version: "17"  # Configurable per environment
```

### 4. Dependency Chain Management

**Decision**: Automatic dependency resolution in `Ensure*` methods

**Rationale**:
- Tests don't need to know installation order
- Simple API: just call `EnsureMCPPackagesInstalled()`
- Framework handles: repo → PostgreSQL → MCP packages

**Flow**:
```
Test calls: s.EnsureMCPPackagesInstalled()
   ↓
Framework ensures: Repository → PostgreSQL → MCP Packages
```

## Suite Hierarchy

```
BaseSuite (base.go)
    │
    ├─→ E2ESuite (e2e.go + install.go)
    │   │   - Purpose: End-to-end testing with system dependencies
    │   │   - Features: File assertions, installation helpers
    │   │   - Used by: Installation, service, files tests
    │   │
    │   └─→ APISuite (api.go)
    │       - Purpose: API/MCP protocol testing
    │       - Features: MCP server lifecycle, protocol assertions, installation helpers
    │       - Used by: MCP protocol, stdio, KB tests
    │       - Note: Extends E2ESuite to inherit installation methods
    │
    └─→ DatabaseSuite (database.go)
        - Purpose: Database-specific testing
        - Features: Connection management, seeding, cleanup
        - Used by: Database tests
```

## Test Execution Modes

Tests can run in different execution modes configured via YAML:

### 1. Local Mode (dev.yaml)
- Runs on local machine
- Requires pre-installed dependencies
- Fast execution

### 2. Container Mode (container.yaml)
- Runs in Docker container with systemd
- Fresh environment per test
- Auto-installs dependencies
- Slower but isolated

### 3. Configuration Example

```yaml
# container.yaml
execution:
  mode: container-systemd
  container:
    os_image: "jrei/systemd-ubuntu:22.04"
    use_systemd: true
    skip_sudo_check: true

postgresql:
  version: "17"

database:
  host: localhost
  user: postgres
  password: postgres123
```

## Adding New Tests

### For Installation/Service Tests (use E2ESuite):

```go
type MyTestSuite struct {
    suite.E2ESuite
}

func (s *MyTestSuite) SetupSuite() {
    s.E2ESuite.SetupSuite()
}

func (s *MyTestSuite) TestSomething() {
    // Install dependencies automatically
    s.EnsureMCPPackagesInstalled()

    // Test...
    s.AssertFileExists("/usr/bin/pgedge-postgres-mcp")
}
```

### For MCP Protocol Tests (use APISuite):

```go
type MyMCPTestSuite struct {
    suite.APISuite
}

func (s *MyMCPTestSuite) TestProtocol() {
    s.StartMCPServer(binary, config, mcp.ModeStdio)
    defer s.StopMCPServer()

    resp, _ := s.MCPServer.Initialize(s.Ctx)
    s.MCPAssertions.AssertValidInitializeResponse(resp)
}
```

### For Database Tests (use DatabaseSuite):

```go
type MyDatabaseTestSuite struct {
    suite.DatabaseSuite
}

func (s *MyDatabaseTestSuite) TestQuery() {
    db := s.GetDB()
    s.SeedTable("users", []string{"name"}, [][]interface{}{{"Alice"}})
    s.AssertRowCount("users", 1)
}
```

## Best Practices

### 1. File Placement
- Suite base classes → `common/suite/`
- Shared utilities → `common/{assertions,database,http,etc}/`
- Actual tests → `mcp-server/testcases/{category}/`
- Configuration → `mcp-server/config/`

### 2. Suite Organization
- Split large suites across files when logical (like E2ESuite)
- Keep public API in main file
- Keep implementation details in separate files
- Document file relationships clearly

### 3. Installation Patterns
- Always use `Ensure*` methods (never direct installation)
- Call appropriate `Ensure*` at start of each test method
- Let framework handle dependency chains
- Don't assume pre-installed software in container mode

### 4. Documentation
- Document file purpose at top of each file
- Explain file relationships (see e2e.go ↔ install.go)
- Provide usage examples in comments
- Maintain README files for complex directories

## Future Enhancements

Consider these improvements for future maintainers:

1. **Parallel Installations**: Speed up by installing independent components concurrently
2. **Installation Cache**: Reuse installed packages across test runs
3. **Snapshot/Restore**: Save container state after installation for faster reruns
4. **Installation Profiles**: Pre-configured sets (minimal, standard, full)
5. **Cleanup Helpers**: Automatic uninstall for test isolation
6. **Version Matrix Testing**: Test against multiple PostgreSQL versions

## Questions?

For questions about the framework design or organization:
1. Check the relevant README in the component directory
2. Look for inline comments in the code
3. Review test examples in `testcases/examples/`
