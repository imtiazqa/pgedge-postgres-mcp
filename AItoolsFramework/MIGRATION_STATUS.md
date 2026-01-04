# Regression Test Migration Status

This document tracks the migration of tests from `/test/regression/` to the new AItoolsFramework.

## Migration Progress

| Original Test | New Test | Status | Notes |
|--------------|----------|--------|-------|
| `installation_test.go` (Test04) | `installation_test.go` | ✅ Complete | Migrated to E2ESuite |
| `stdio_test.go` (Test11) | `stdio_test.go` | ✅ Complete | Migrated to APISuite with MCP helper |
| `mcp_server_test.go` | `mcp_server_test.go` | 📋 Pending | Should use APISuite |
| `postgresql_test.go` | `postgresql_test.go` | 📋 Pending | Should use DatabaseSuite |
| `service_test.go` | `service_test.go` | 📋 Pending | Should use E2ESuite |
| `repository_test.go` | `repository_test.go` | 📋 Pending | Should use E2ESuite |
| `user_test.go` | `user_test.go` | 📋 Pending | Should use E2ESuite |
| `token_test.go` | `token_test.go` | 📋 Pending | Should use APISuite |
| `kb_test.go` | `kb_test.go` | 📋 Pending | Should use DatabaseSuite |
| `mcp_kb_test.go` | `mcp_kb_test.go` | 📋 Pending | Should use APISuite + DatabaseSuite |
| `files_test.go` | `files_test.go` | 📋 Pending | Should use E2ESuite |

## Completed Migrations

### ✅ installation_test.go

**Location**: `AItoolsFramework/mcp-server/suites/installation_test.go`

**What Changed:**
- Old: Used custom `RegressionTestSuite` with manual executor
- New: Uses `E2ESuite` with built-in assertions
- Benefits:
  - Cleaner test code
  - Built-in helpers (`AssertFileExists`, `AssertDirectoryExists`)
  - Better error messages
  - Configuration-driven paths

**Test Methods:**
- `TestBinaryInstallation` - Validates binary exists and is executable
- `TestConfigurationFiles` - Checks config files
- `TestSystemdService` - Validates systemd service
- `TestDataDirectory` - Checks data directory
- `TestBinaryVersion` - Validates version output
- `TestInstallationCompleteness` - Comprehensive check

**Run Command:**
```bash
cd AItoolsFramework/mcp-server
make test-run TEST=TestInstallationSuite
```

### ✅ stdio_test.go

**Location**: `AItoolsFramework/mcp-server/suites/stdio_test.go`

**What Changed:**
- Old: Manual stdio communication with bash scripts
- New: Uses `MCPServerHelper` for clean stdio management
- Benefits:
  - Automatic server lifecycle management
  - Built-in JSON-RPC handling
  - MCP-specific assertions
  - No manual process management

**Test Methods:**
- `TestStdioInitialize` - Tests MCP initialize
- `TestStdioToolsList` - Tests tools/list
- `TestStdioResourcesList` - Tests resources/list
- `TestStdioQueryDatabase` - Tests query_database tool
- `TestStdioGetSchemaInfo` - Tests get_schema_info tool
- `TestStdioServerLifecycle` - Tests start/stop multiple times

**Run Command:**
```bash
cd AItoolsFramework/mcp-server
make test-run TEST=TestStdioSuite
```

## Key Improvements

### Old Framework Issues
1. ❌ Manual executor management
2. ❌ Hardcoded paths and commands
3. ❌ Complex bash scripts for MCP communication
4. ❌ No fixture management
5. ❌ Repetitive assertion code
6. ❌ Difficult to maintain

### New Framework Benefits
1. ✅ Built-in executor support (local/container)
2. ✅ Configuration-driven (YAML)
3. ✅ Native MCP protocol support
4. ✅ Fixture system for test data
5. ✅ Rich assertion libraries
6. ✅ Easy to extend and maintain

## Migration Guide

### Choosing the Right Suite

| Test Type | Use Suite | Key Features |
|-----------|-----------|--------------|
| File/directory checks | `E2ESuite` | `AssertFileExists`, `AssertDirectoryExists` |
| Database operations | `DatabaseSuite` | DB helpers, fixtures, seeder |
| MCP protocol | `APISuite` | MCP server helper, protocol assertions |
| HTTP API | `APISuite` | HTTP client, status assertions |
| Mixed (DB + MCP) | Custom suite embedding both | Combine DatabaseSuite + APISuite |

### Migration Steps

1. **Identify test type** - E2E, Database, API/MCP
2. **Choose suite type** - E2ESuite, DatabaseSuite, APISuite
3. **Create new test file** in `AItoolsFramework/mcp-server/suites/`
4. **Port test logic** using framework helpers
5. **Update configuration** in `config/dev.yaml` if needed
6. **Run and verify** tests pass

### Example: Migrating a Simple E2E Test

**Old Code:**
```go
func (s *RegressionTestSuite) TestSomething() {
    output, exitCode, err := s.execCmd(s.ctx, "test -f /some/file")
    s.NoError(err)
    s.Equal(0, exitCode)
}
```

**New Code:**
```go
func (s *E2ETestSuite) TestSomething() {
    s.AssertFileExists("/some/file")
}
```

### Example: Migrating MCP Test

**Old Code:**
```go
func (s *RegressionTestSuite) TestMCP() {
    // 50 lines of bash script to start server and send JSON-RPC
    script := `...complex bash...`
    output, _, _ := s.execCmd(s.ctx, script)
    // Parse JSON manually
}
```

**New Code:**
```go
func (s *APISuite) TestMCP() {
    s.StartMCPServer(binaryPath, configPath, mcp.ModeStdio)
    defer s.StopMCPServer()

    resp, err := s.MCPServer.ListTools(s.Ctx)
    s.NoError(err)
    s.MCPAssertions.AssertToolExists(resp, "query_database")
}
```

## Next Steps

### Immediate Priority
1. Migrate `mcp_server_test.go` - Core MCP functionality
2. Migrate `postgresql_test.go` - Database operations
3. Migrate `service_test.go` - Service management

### Medium Priority
4. Migrate `token_test.go` - Authentication
5. Migrate `repository_test.go` - Repository operations
6. Migrate `user_test.go` - User management

### Lower Priority
7. Migrate `kb_test.go` - Knowledge base
8. Migrate `mcp_kb_test.go` - MCP + KB integration
9. Migrate `files_test.go` - File operations

### Final Steps
10. Update CI/CD to use new framework
11. Archive old regression tests
12. Update documentation

## Running Tests

### Run All Migrated Tests
```bash
cd AItoolsFramework/mcp-server
make test
```

### Run Specific Test Suite
```bash
make test-run TEST=TestInstallationSuite
make test-run TEST=TestStdioSuite
```

### Run with Verbose Output
```bash
make test-verbose
```

### Run Against Staging
```bash
make test-staging
```

## Configuration

Tests are configured via YAML files in `config/`:
- `dev.yaml` - Development environment (local testing)
- `staging.yaml` - Staging environment

Key configuration sections:
- `execution` - How tests run (local/container)
- `database` - Database connection settings
- `fixtures` - Test data files
- `timeouts` - Operation timeouts
- `reporting` - Test output format

## Questions?

See:
- [Framework README](README.md) - Overview
- [Example Tests](mcp-server/suites/) - Working examples
- [Configuration Guide](mcp-server/config/dev.yaml) - Full config reference
