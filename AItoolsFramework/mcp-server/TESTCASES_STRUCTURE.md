# Test Cases Directory Structure

## Overview

All test cases are organized in the `testcases/` directory by category for better organization and maintainability.

```
mcp-server/
├── testcases/
│   ├── installation/          # Installation & setup validation
│   ├── database/             # Database operations & PostgreSQL
│   ├── service/              # Service management (systemd/manual)
│   ├── mcp/                  # MCP protocol & stdio communication
│   ├── kb/                   # Knowledge base functionality
│   └── examples/             # Framework usage examples
├── config/
│   ├── dev.yaml              # Development configuration
│   └── staging.yaml          # Staging configuration
└── Makefile                   # Test execution targets
```

## Test Categories

### 📦 installation/
**Purpose**: Validate MCP server installation and configuration

**Tests:**
- `installation_test.go` - Binary installation, config files, systemd service, data directories

**Suite**: `E2ESuite`

**Run:**
```bash
make test-installation
```

---

### 🗄️ database/
**Purpose**: PostgreSQL database operations and schema management

**Tests:**
- `postgresql_test.go` - PostgreSQL installation, initialization, user setup, database creation
- `database_test.go` - Database connectivity, schema operations, data seeding, fixtures

**Suite**: `DatabaseSuite`

**Run:**
```bash
make test-database
```

---

### ⚙️ service/
**Purpose**: Service lifecycle management

**Tests:**
- `service_test.go` - Service start/stop/restart, systemd management, manual service control, HTTP endpoint validation

**Suite**: `E2ESuite`

**Run:**
```bash
make test-service
```

---

### �� mcp/
**Purpose**: MCP protocol functionality and communication

**Tests:**
- `stdio_test.go` - MCP stdio mode, initialize, tools/list, resources/list, query_database
- `mcp_server_test.go` - MCP package installation, configuration
- `mcp_protocol_test.go` - Message builders, validators, protocol compliance

**Suite**: `APISuite`

**Run:**
```bash
make test-mcp
```

---

### 📚 kb/
**Purpose**: Knowledge base operations and MCP+KB integration

**Tests:**
- `kb_test.go` - Knowledge base CRUD operations (TODO: to be migrated)
- `mcp_kb_test.go` - MCP + knowledge base integration (TODO: to be migrated)

**Suite**: `DatabaseSuite` + `APISuite` (combined)

**Run:**
```bash
make test-kb
```

---

### 📖 examples/
**Purpose**: Framework usage demonstrations

**Tests:**
- `example_test.go` - Configuration access, assertions, helpers, suite lifecycle

**Suite**: `E2ESuite`

**Run:**
```bash
make test-examples
```

---

## Quick Commands

### Run All Tests
```bash
cd AItoolsFramework/mcp-server
make test
```

### Run Tests by Category
```bash
make test-installation    # Installation tests only
make test-database       # Database tests only
make test-service        # Service tests only
make test-mcp            # MCP protocol tests only
make test-kb             # Knowledge base tests only
make test-examples       # Example tests only
```

### Run Specific Test
```bash
make test-run TEST=TestBinaryInstallation
make test-run TEST=TestStdioInitialize
make test-run TEST=TestPostgreSQLSetup
```

### Run Sequentially (All Categories)
```bash
make test-sequential
```

### List Available Tests
```bash
make list-tests
```

### Run with Different Config
```bash
make test-staging                           # Use staging.yaml
CONFIG_FILE=config/custom.yaml make test    # Use custom config
```

## Test Organization Benefits

### ✅ Clear Separation
- Each category has a dedicated directory
- Easy to find tests by functionality
- Logical grouping of related tests

### ✅ Scalability
- Easy to add new test categories
- No mixing of unrelated tests
- Clear boundaries between test types

### ✅ Selective Execution
- Run only the tests you need
- Faster feedback during development
- Efficient CI/CD pipelines

### ✅ Better Maintainability
- Clear ownership per category
- Easier code reviews
- Simpler test discovery

## Migration Status

### ✅ Migrated (6 tests)
| Category | Test File | Original | Status |
|----------|-----------|----------|--------|
| installation | `installation_test.go` | Test04 | ✅ Complete |
| database | `postgresql_test.go` | Test02 | ✅ Complete |
| database | `database_test.go` | New | ✅ Complete |
| service | `service_test.go` | Test08 | ✅ Complete |
| mcp | `stdio_test.go` | Test11 | ✅ Complete |
| mcp | `mcp_server_test.go` | Test03 | ✅ Complete |
| mcp | `mcp_protocol_test.go` | New | ✅ Complete |
| examples | `example_test.go` | New | ✅ Complete |

### 📋 Pending (5 tests)
| Category | Test to Migrate | Priority |
|----------|----------------|----------|
| kb | `kb_test.go` | Medium |
| kb | `mcp_kb_test.go` | Medium |
| installation | `repository_test.go` | Low |
| installation | `user_test.go` | Low |
| mcp | `token_test.go` | Medium |
| service | `files_test.go` | Low |

## Adding New Tests

1. **Choose category** (or create new one):
   ```bash
   mkdir -p testcases/new-category
   ```

2. **Create test file**:
   ```bash
   touch testcases/new-category/feature_test.go
   ```

3. **Choose appropriate suite**:
   - `E2ESuite` - System operations (files, processes, commands)
   - `DatabaseSuite` - Database operations
   - `APISuite` - MCP/HTTP API calls

4. **Write tests** following the pattern:
   ```go
   package newcategory

   import (
       "testing"
       "github.com/pgedge/AItoolsFramework/common/suite"
       testifySuite "github.com/stretchr/testify/suite"
   )

   type FeatureTestSuite struct {
       suite.E2ESuite  // or DatabaseSuite, APISuite
   }

   func (s *FeatureTestSuite) TestSomething() {
       // Your test code
   }

   func TestFeatureSuite(t *testing.T) {
       testifySuite.Run(t, new(FeatureTestSuite))
   }
   ```

5. **Add Makefile target** (optional):
   ```makefile
   test-new-category:
       @echo "Running new category tests..."
       TESTFW_CONFIG=$(CONFIG_FILE) go test -v ./testcases/new-category/
   ```

6. **Update documentation**:
   - Add to `testcases/README.md`
   - Update this file

## Configuration

All tests use YAML configuration files from `config/`:

### dev.yaml (Default)
- Local execution mode
- Development database
- Minimal timeouts
- Detailed logging

### staging.yaml
- Container execution mode (optional)
- Staging database
- Production-like timeouts
- Standard logging

### Override Config
```bash
# Via environment
TESTFW_CONFIG=config/custom.yaml go test ./testcases/...

# Via Makefile
make test CONFIG_FILE=config/custom.yaml
```

## Best Practices

### ✅ DO
- Place tests in the appropriate category
- Use descriptive test names
- Follow suite patterns
- Clean up resources in teardown
- Use configuration for paths/values

### ❌ DON'T
- Mix unrelated tests in one file
- Hardcode paths or credentials
- Skip teardown/cleanup
- Create dependencies between tests
- Modify global state

## CI/CD Integration

### GitHub Actions Example
```yaml
- name: Run Installation Tests
  run: make test-installation

- name: Run Database Tests
  run: make test-database

- name: Run MCP Tests
  run: make test-mcp
```

### Run All Sequentially
```bash
make test-sequential
```

This ensures each category passes before moving to the next.
