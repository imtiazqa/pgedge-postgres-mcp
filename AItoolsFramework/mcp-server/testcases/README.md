# MCP Server Test Cases

This directory contains all test cases for the MCP server, organized by category.

## Directory Structure

```
testcases/
├── installation/          # Installation and setup validation tests
├── database/             # Database-related tests (PostgreSQL, schemas, etc.)
├── service/              # Service management tests (systemd, manual start/stop)
├── mcp/                  # MCP protocol tests (stdio, tools, resources)
├── kb/                   # Knowledge base tests
└── examples/             # Example tests demonstrating framework usage
```

## Test Categories

### 📦 installation/
Tests for validating MCP server installation and configuration.

**Tests:**
- `installation_test.go` - Binary installation, config files, systemd service, data directories
- `repository_test.go` - Repository installation (Debian & RHEL)
- `user_test.go` - User management and authentication
- `files_test.go` - Package files and permissions verification

**Suite Type:** `E2ESuite`

**Run:**
```bash
go test ./testcases/installation/
```

---

### 🗄️ database/
Tests for PostgreSQL database operations and configuration.

**Tests:**
- `postgresql_test.go` - PostgreSQL installation, initialization, user setup, database creation
- `database_test.go` - Database connectivity, schema operations, data seeding

**Suite Type:** `DatabaseSuite`

**Run:**
```bash
go test ./testcases/database/
```

---

### ⚙️ service/
Tests for service management (systemd and manual).

**Tests:**
- `service_test.go` - Service start/stop/restart, status checks, HTTP endpoint validation

**Suite Type:** `E2ESuite`

**Run:**
```bash
go test ./testcases/service/
```

---

### 🔌 mcp/
Tests for MCP protocol functionality.

**Tests:**
- `stdio_test.go` - MCP server stdio mode, initialize, tools/list, resources/list
- `mcp_server_test.go` - MCP package installation and configuration
- `mcp_protocol_test.go` - MCP message builders, validators, protocol compliance
- `token_test.go` - Token management and authentication

**Suite Type:** `APISuite` (stdio, protocol), `E2ESuite` (server, token)

**Run:**
```bash
go test ./testcases/mcp/
```

---

### 📚 kb/
Tests for knowledge base functionality.

**Tests:**
- `kb_test.go` - Knowledge base builder, Ollama integration, KB generation
- `mcp_kb_test.go` - MCP + KB integration testing

**Suite Type:** `E2ESuite` (kb), `APISuite` (mcp_kb)

**Run:**
```bash
go test ./testcases/kb/
```

---

### 📖 examples/
Example tests demonstrating framework features.

**Tests:**
- `example_test.go` - Framework usage examples (config access, assertions, helpers)

**Suite Type:** `E2ESuite`

**Run:**
```bash
go test ./testcases/examples/
```

---

## Running Tests

### Run All Tests
```bash
make test
```

### Run Specific Category
```bash
go test ./testcases/installation/
go test ./testcases/database/
go test ./testcases/service/
go test ./testcases/mcp/
go test ./testcases/kb/
```

### Run Specific Test
```bash
go test ./testcases/installation/ -run TestBinaryInstallation
go test ./testcases/mcp/ -run TestStdioInitialize
```

### Run with Verbose Output
```bash
go test -v ./testcases/installation/
```

### Run All Categories Sequentially
```bash
for dir in installation database service mcp kb; do
    echo "Running $dir tests..."
    go test -v ./testcases/$dir/
done
```

## Test Naming Conventions

### File Names
- `<feature>_test.go` - Main test file for a feature
- Example: `installation_test.go`, `stdio_test.go`

### Suite Names
- `<Feature>TestSuite` - Suite struct name
- Example: `InstallationTestSuite`, `StdioTestSuite`

### Test Method Names
- `Test<Feature><Aspect>` - Descriptive test method
- Example: `TestBinaryInstallation`, `TestStdioInitialize`

## Migration Status

### ✅ Completed (100%)
- installation/ - All installation, repository, user management, and file verification tests (4 tests)
- database/ - PostgreSQL and database tests (2 tests)
- service/ - Service management tests (1 test)
- mcp/ - MCP protocol, stdio, server, and token tests (4 tests)
- kb/ - Knowledge base and MCP+KB integration tests (2 tests)
- examples/ - Framework example tests (1 test)

**Total: 12/12 regression tests migrated (100%)**

## Adding New Tests

1. **Choose the right category** (or create a new one)
2. **Create test file** in the category directory
3. **Choose appropriate suite type:**
   - `E2ESuite` - System-level tests (files, processes, services)
   - `DatabaseSuite` - Database operations
   - `APISuite` - MCP/HTTP API tests
4. **Follow naming conventions**
5. **Update this README** with test description

## Configuration

Tests use configuration from:
- `config/dev.yaml` - Development/local testing
- `config/staging.yaml` - Staging environment

Set config via environment:
```bash
TESTFW_CONFIG=config/staging.yaml go test ./testcases/...
```

Or use Makefile targets:
```bash
make test              # Uses dev.yaml
make test-staging      # Uses staging.yaml
```
