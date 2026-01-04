# Test Suite Framework

This directory contains the base test suite implementations for the MCP server test framework.

## File Organization

```
suite/
├── README.md        - This file
├── base.go          - BaseSuite (foundation for all test suites)
├── e2e.go           - E2ESuite (end-to-end testing with installation API)
├── install.go       - E2ESuite installation implementations
├── api.go           - APISuite (API/MCP protocol testing)
└── database.go      - DatabaseSuite (database testing)
```

## Suite Hierarchy

```
BaseSuite (base.go)
    ├── E2ESuite (e2e.go + install.go)
    │   │   └── Used for: installation, service, files tests
    │   │
    │   └── APISuite (api.go)
    │       └── Used for: MCP protocol, stdio, KB tests
    │       └── Note: Extends E2ESuite to inherit installation methods
    │
    └── DatabaseSuite (database.go)
        └── Used for: database-specific tests
```

## E2ESuite - File Split Explanation

The E2ESuite is split across two files for clarity and maintainability:

### e2e.go
- **Purpose**: Public API and interface
- **Contains**:
  - Suite struct definition
  - SetupSuite/TearDownSuite
  - Assertion helpers (`AssertFileExists`, etc.)
  - Public installation API (`EnsureRepositoryInstalled`, etc.)
  - Helper utilities

### install.go
- **Purpose**: Installation implementation details
- **Contains**:
  - Private installation methods (`installRepository`, etc.)
  - OS-specific installation logic (Debian vs RHEL)
  - PostgreSQL setup procedures
  - MCP package installation

**Why Split?**
1. Prevents e2e.go from becoming too large (300+ lines of install logic)
2. Clear separation: interface (e2e.go) vs implementation (install.go)
3. Easier to maintain: installation changes isolated to install.go
4. Go allows methods on same type across multiple files in same package

## Usage Examples

### Basic E2E Test

```go
type MyTestSuite struct {
    suite.E2ESuite
}

func (s *MyTestSuite) TestSomething() {
    // Automatically install dependencies
    s.EnsureMCPPackagesInstalled()

    // Now test the installed software
    s.AssertFileExists("/usr/bin/pgedge-postgres-mcp")
}
```

### Database Test

```go
type MyDatabaseTestSuite struct {
    suite.DatabaseSuite
}

func (s *MyDatabaseTestSuite) TestQuery() {
    // Database connection already established
    db := s.GetDB()
    // Run queries...
}
```

### MCP Protocol Test

```go
type MyMCPTestSuite struct {
    suite.APISuite
}

func (s *MyMCPTestSuite) TestMCPProtocol() {
    s.StartMCPServer("/usr/bin/pgedge-postgres-mcp", "/etc/pgedge/postgres-mcp.yaml", mcp.ModeStdio)
    defer s.StopMCPServer()

    resp, err := s.MCPServer.Initialize(s.Ctx)
    s.MCPAssertions.AssertValidInitializeResponse(resp)
}
```

## Installation Workflow

When a test calls `s.EnsureMCPPackagesInstalled()`:

```
1. EnsureMCPPackagesInstalled() (e2e.go)
   ↓
2. Check: Already installed? → Return
   ↓
3. EnsurePostgreSQLInstalled() (e2e.go)
   ↓
4. Check: PostgreSQL installed? → Return
   ↓
5. EnsureRepositoryInstalled() (e2e.go)
   ↓
6. Check: Repository installed? → Return
   ↓
7. installRepository() (install.go)
   ├── Detect OS type (Debian vs RHEL)
   ├── installDebianRepository() OR installRHELRepository()
   └── Mark as installed
   ↓
8. installPostgreSQL() (install.go)
   ├── Install packages
   ├── Initialize database
   ├── Set password
   ├── Create MCP database
   └── Mark as installed
   ↓
9. installMCPPackages() (install.go)
   ├── Install all MCP packages
   ├── Configure database password
   └── Mark as installed
```

## Key Features

### Idempotent Installation
Each `Ensure*` method only installs once per test run, tracked by `setupState`:
```go
setupState struct {
    repoInstalled        bool
    postgresqlInstalled  bool
    mcpPackagesInstalled bool
}
```

### Dependency Chain
- MCP Packages require PostgreSQL
- PostgreSQL requires Repository
- Automatically handled by `Ensure*` methods

### OS Detection
Automatically detects Debian vs RHEL and uses appropriate package manager:
- Debian/Ubuntu: `apt-get`
- RHEL/Rocky/Alma: `dnf`

## Configuration

PostgreSQL version and other settings come from config files:

```yaml
# config/container.yaml
postgresql:
  version: "17"  # 16, 17, or 18

database:
  host: localhost
  port: 5432
  user: postgres
  password: postgres123
```

## Future Enhancements

Potential improvements for future maintainers:

1. **Parallel Installation**: Install independent components concurrently
2. **Installation Cache**: Cache installed packages across test runs
3. **Version Pinning**: Support specific package versions
4. **Cleanup Helpers**: Uninstall methods for test isolation
5. **Installation Profiles**: Pre-configured installation sets (minimal, full, etc.)
