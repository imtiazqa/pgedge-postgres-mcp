# Tester's Guide - pgEdge Postgres MCP Test Framework

Complete guide for QA testers to understand the project structure and add
new test cases.

## 1. Project Folder Structure

### Root Directory: `/pgedge-postgres-mcp/`

Main project containing the pgEdge Postgres MCP server and its test framework.

```
pgedge-postgres-mcp/
├── AItoolsFramework/          # Test automation framework (⭐ MAIN TESTING AREA)
├── cmd/                       # Executable command-line programs
├── internal/                  # Core application source code (Go packages)
├── web/                       # React-based web user interface
├── test/                      # Integration and regression tests
├── docs/                      # MkDocs project documentation
├── examples/                  # Example implementations and usage
├── docker/                    # Docker container configurations
├── kb/                        # Knowledge base files
├── Makefile                   # Build and test automation
├── go.mod, go.sum             # Go dependency management
├── docker-compose.yml         # Container orchestration
└── mkdocs.yml                 # Documentation site configuration
```

---

### `/AItoolsFramework/` - Test Automation Framework ⭐

**Purpose**: Enterprise-grade test automation framework for MCP server testing.

```
AItoolsFramework/
│
├── common/                    # Shared test framework (Go module)
│   ├── assertions/            # Custom assertion helpers
│   ├── config/                # Configuration loading and validation
│   ├── database/              # Database test helpers
│   ├── executor/              # Test execution modes (local/container)
│   ├── fixtures/              # Test data management utilities
│   ├── http/                  # HTTP client utilities
│   ├── mcp/                   # MCP protocol helpers
│   ├── reporters/             # Test result reporters
│   ├── suite/                 # Base test suites (E2E, API, Database)
│   └── utils/                 # Shared utility functions
│
└── mcp-server/                # MCP server test project ⭐⭐
    ├── config/                # Test execution configurations
    │   ├── local.yaml         # Local machine execution config
    │   └── container.yaml     # Docker container execution config
    │
    ├── testcases/             # ⭐⭐⭐ ACTUAL TEST IMPLEMENTATIONS
    │   ├── installation_test.go    # Installation tests
    │   ├── postgresql_test.go      # PostgreSQL setup tests
    │   ├── repository_test.go      # Repository setup tests
    │   ├── files_test.go           # File verification tests
    │   ├── user_test.go            # User management tests
    │   ├── token_test.go           # Token management tests
    │   ├── mcp_server_test.go      # MCP server functionality tests
    │   ├── service_test.go         # Service management tests
    │   ├── stdio_test.go           # Stdio mode tests
    │   ├── kb_test.go              # Knowledge base tests
    │   ├── mcp_kb_test.go          # MCP+KB integration tests
    │   ├── regression_suite_test.go # Regression test suite
    │   ├── example_test.go         # Framework usage examples
    │   ├── suite_test.go           # Suite initialization
    │   └── helpers_test.go         # Test helper functions
    │
    ├── fixtures/              # Test data and fixtures
    │   ├── databases/         # SQL schemas and test data
    │   ├── configs/           # Test configuration files
    │   └── responses/         # Expected response data
    │
    ├── test-results/          # Test execution results (auto-generated)
    │   ├── test-local.log     # Local execution logs
    │   └── test-container.log # Container execution logs
    │
    ├── docs/                  # Test documentation
    ├── Makefile               # Test runner commands
    ├── README.md              # Framework documentation
    ├── ARCHITECTURE.md        # Framework design
    ├── OPTIMIZATION.md        # Performance optimizations
    └── POCKET_GUIDE.md        # Quick reference guide
```

---

### `/cmd/` - Command-Line Executables

**Purpose**: Executable programs for various tools.

```
cmd/
├── kb-builder/                # Knowledge base builder tool
├── pgedge-pg-mcp-cli/         # MCP client CLI
├── pgedge-pg-mcp-svr/         # MCP server executable
└── test-config/               # Configuration testing tool
```

---

### `/internal/` - Core Application Code

**Purpose**: Internal Go packages containing core application logic.

```
internal/
├── api/                       # API request handlers
├── auth/                      # Authentication and authorization
├── chat/                      # Chat functionality
├── database/                  # Database operations and management
├── mcp/                       # MCP protocol implementation
├── tools/                     # MCP tools implementation
├── resources/                 # MCP resources implementation
├── prompts/                   # MCP prompts implementation
├── config/                    # Application configuration
├── kbchunker/                 # Knowledge base chunking
├── kbconfig/                  # Knowledge base configuration
├── kbdatabase/                # Knowledge base database
├── kbembed/                   # Knowledge base embeddings
├── kbsource/                  # Knowledge base sources
├── embedding/                 # Embedding generation
├── search/                    # Search functionality
├── llmproxy/                  # LLM proxy service
├── logging/                   # Logging utilities
├── crypto/                    # Cryptography utilities
├── definitions/               # Type definitions
├── conversations/             # Conversation management
└── compactor/                 # Data compaction utilities
```

---

### `/web/` - Web User Interface

**Purpose**: React-based frontend application.

```
web/
├── src/
│   ├── components/            # React UI components (21 files)
│   ├── contexts/              # React context providers (6 files)
│   ├── hooks/                 # Custom React hooks (10 files)
│   ├── utils/                 # Frontend utility functions
│   ├── lib/                   # External libraries
│   ├── theme/                 # UI theming
│   ├── test/                  # Frontend test files
│   ├── test-utils/            # Frontend test utilities
│   ├── assets/                # Static assets (images, fonts)
│   └── App.jsx                # Main application component
├── public/                    # Public static files
├── package.json               # NPM dependencies
└── vite.config.js             # Vite build configuration
```

---

### `/test/` - Integration & Regression Tests

**Purpose**: Additional integration and regression test files.

```
test/
├── integration/               # Integration test scripts
└── regression/                # Regression test scripts
```

---

### `/docs/` - Project Documentation

**Purpose**: MkDocs-based project documentation.

```
docs/
├── guide/                     # User guides
├── reference/                 # API reference documentation
│   └── config-examples/       # Configuration examples
├── advanced/                  # Advanced topics
├── developers/                # Developer documentation
├── contributing/              # Contribution guidelines
│   └── internal/              # Internal documentation
├── img/                       # Images and diagrams
│   └── screenshots/           # Application screenshots
├── index.md                   # Documentation home page
├── quickstart.md              # Quick start guide
└── changelog.md               # Version changelog
```

---

### `/examples/` - Example Implementations

**Purpose**: Example code and usage demonstrations.

```
examples/
├── client/                    # Client usage examples
├── server/                    # Server usage examples
└── configs/                   # Configuration examples
```

---

### `/docker/` - Docker Configurations

**Purpose**: Docker container setup files.

```
docker/
├── Dockerfile                 # Main Dockerfile
└── docker-compose.yml         # Compose configuration
```

---

### `/kb/` - Knowledge Base

**Purpose**: Knowledge base content and data files.

```
kb/
├── sources/                   # Source documents
└── embeddings/                # Generated embeddings
```

---

## 2. How to Add a New Test Case

Follow these steps to add a new test case to the existing project.

### Step 1: Identify the Test Category

Determine which category your test belongs to:

- **Installation Tests** - Package installation, binary verification, config
  files
- **Database Tests** - PostgreSQL setup, database operations, data validation
- **Service Tests** - systemd service management, service status, restarts
- **MCP Protocol Tests** - MCP server functionality, protocol messages, stdio
  mode
- **Knowledge Base Tests** - KB builder, KB search, MCP+KB integration
- **Other** - Create a new category if needed

### Step 2: Choose the Right Test Suite Type

Based on your test category, choose the appropriate base suite:

#### **E2ESuite** - For installation, service, file tests

```go
type MyTestSuite struct {
    suite.E2ESuite
}
```

**Features:**
- Installation helpers with automatic dependency management
- File/directory assertions
- Command execution
- Service management (systemd)

**Use when**: Testing installation, files, services, system commands

#### **APISuite** - For MCP protocol tests

```go
type MyMCPSuite struct {
    suite.APISuite
}
```

**Features:**
- Extends E2ESuite (has all installation methods)
- MCP server lifecycle management
- MCP protocol assertions
- HTTP client helpers

**Use when**: Testing MCP protocol, server responses, API endpoints

#### **DatabaseSuite** - For database tests

```go
type MyDatabaseSuite struct {
    suite.DatabaseSuite
}
```

**Features:**
- Database connection management
- Data seeding utilities
- Schema operations
- Transaction handling

**Use when**: Testing database operations, SQL queries, data validation

### Step 3: Create Your Test File

**Location**: `AItoolsFramework/mcp-server/testcases/`

**Naming**: `<feature>_test.go` (e.g., `my_feature_test.go`)

### Step 4: Write Your Test Suite

Here's a complete template for a new test case:

```go
package testcases

import (
    "testing"

    "github.com/stretchr/testify/suite"
    baseSuite "github.com/pgedge/AItoolsFramework/common/suite"
)

// Step 1: Define your test suite struct
type MyFeatureTestSuite struct {
    baseSuite.E2ESuite  // Or APISuite or DatabaseSuite
}

// Step 2: SetupSuite runs once before all tests
func (s *MyFeatureTestSuite) SetupSuite() {
    // ALWAYS call parent SetupSuite first
    s.E2ESuite.SetupSuite()

    // Install dependencies if needed
    s.EnsureMCPPackagesInstalled()  // Auto-installs all dependencies

    // Additional setup (optional)
    // s.DoSomethingElse()
}

// Step 3: SetupTest runs before EACH test
func (s *MyFeatureTestSuite) SetupTest() {
    // Optional: setup before each test
}

// Step 4: Write your test methods
func (s *MyFeatureTestSuite) TestFeatureOne() {
    // Arrange
    expectedValue := "expected output"

    // Act
    output, exitCode, err := s.ExecCommand("my-command --arg")

    // Assert
    s.NoError(err, "Command should execute without error")
    s.Equal(0, exitCode, "Command should exit successfully")
    s.Contains(output, expectedValue, "Output should contain expected value")
}

func (s *MyFeatureTestSuite) TestFeatureTwo() {
    // Another test case
    s.AssertFileExists("/usr/bin/pgedge-postgres-mcp")
}

// Step 5: TearDownTest runs after EACH test
func (s *MyFeatureTestSuite) TearDownTest() {
    // Optional: cleanup after each test
}

// Step 6: TearDownSuite runs once after all tests
func (s *MyFeatureTestSuite) TearDownSuite() {
    // ALWAYS call parent TearDownSuite
    s.E2ESuite.TearDownSuite()

    // Additional cleanup (optional)
}

// Step 7: Register your suite with Go's testing framework
func TestMyFeatureTestSuite(t *testing.T) {
    suite.Run(t, new(MyFeatureTestSuite))
}
```

### Step 5: Add Test-Specific Configuration (Optional)

If your test needs custom configuration, you can:

1. Add fields to the config struct in
   `common/config/types.go`
2. Update `config/local.yaml` and `config/container.yaml` with new values
3. Access in tests via `s.Config.YourNewField`

**Example:**

```yaml
# config/local.yaml
my_feature:
  enabled: true
  timeout: 30
```

```go
// In your test
timeout := s.Config.MyFeature.Timeout
```

### Step 6: Add Test Fixtures (Optional)

If your test needs test data:

1. Create fixture files in `fixtures/` directory
2. Load in your test using fixture helpers

**Example:**

```go
// fixtures/databases/my_test_data.sql
INSERT INTO users (name, email) VALUES ('Test User', 'test@example.com');
```

```go
// In your test
s.SeedTable("users", []string{"name", "email"},
    [][]interface{}{{"Test User", "test@example.com"}})
```

### Step 7: Run Your Test

```bash
# Navigate to test directory
cd AItoolsFramework/mcp-server

# Run your specific test
go test ./testcases/ -run TestMyFeatureTestSuite -v

# Or run with Makefile
make test-local
```

### Step 8: Add to Makefile (Optional)

If you created a new test category, add a Makefile target:

```makefile
# In AItoolsFramework/mcp-server/Makefile

.PHONY: test-myfeature
test-myfeature:
    @echo "Running My Feature tests..."
    TESTFW_CONFIG=$(CONFIG_FILE) go test -v ./testcases/ \
        -run "TestMyFeatureTestSuite" 2>&1 | tee $(RESULTS_DIR)/test-myfeature.log
```

---

## 3. Complete Example: Adding a "Backup Test"

Let's add a real test for database backup functionality.

### File: `testcases/backup_test.go`

```go
package testcases

import (
    "fmt"
    "os"
    "testing"
    "time"

    "github.com/stretchr/testify/suite"
    baseSuite "github.com/pgedge/AItoolsFramework/common/suite"
)

// BackupTestSuite tests database backup functionality
type BackupTestSuite struct {
    baseSuite.E2ESuite
    backupDir string
}

func (s *BackupTestSuite) SetupSuite() {
    s.E2ESuite.SetupSuite()
    s.EnsureMCPPackagesInstalled()

    // Create backup directory
    s.backupDir = "/tmp/pgedge_backups"
    err := os.MkdirAll(s.backupDir, 0755)
    s.NoError(err, "Should create backup directory")
}

func (s *BackupTestSuite) TestBackupDatabaseCommand() {
    // Arrange
    timestamp := time.Now().Format("20060102_150405")
    backupFile := fmt.Sprintf("%s/backup_%s.sql", s.backupDir, timestamp)

    // Act
    cmd := fmt.Sprintf("pg_dump -h localhost -U postgres -d testdb > %s",
        backupFile)
    _, exitCode, err := s.ExecCommand(cmd)

    // Assert
    s.NoError(err, "Backup command should execute")
    s.Equal(0, exitCode, "Backup should succeed")
    s.AssertFileExists(backupFile)

    // Verify backup file is not empty
    info, _ := os.Stat(backupFile)
    s.Greater(info.Size(), int64(0), "Backup file should not be empty")
}

func (s *BackupTestSuite) TestRestoreFromBackup() {
    // This test verifies backup can be restored
    // Implementation details...
}

func (s *BackupTestSuite) TearDownSuite() {
    s.E2ESuite.TearDownSuite()

    // Cleanup backup directory
    os.RemoveAll(s.backupDir)
}

func TestBackupTestSuite(t *testing.T) {
    suite.Run(t, new(BackupTestSuite))
}
```

### Run the test:

```bash
cd AItoolsFramework/mcp-server
go test ./testcases/ -run TestBackupTestSuite -v
```

---

## 4. Common Assertions for Testing

### File Assertions

```go
s.AssertFileExists("/path/to/file")
s.AssertFileNotExists("/path/to/file")
s.AssertFileContains("/path/to/file", "expected content")
s.AssertDirExists("/path/to/dir")
```

### Command Execution

```go
output, exitCode, err := s.ExecCommand("ls -la")
s.NoError(err)
s.Equal(0, exitCode)
s.Contains(output, "expected")
```

### Database Assertions (DatabaseSuite)

```go
s.AssertTableExists("table_name")
s.AssertRowCount("table_name", 5)
s.AssertColumnExists("table_name", "column_name")
```

### MCP Assertions (APISuite)

```go
s.MCPAssertions.AssertValidInitializeResponse(resp)
s.MCPAssertions.AssertToolExists("query_database")
s.MCPAssertions.AssertResourceExists("kb://documents")
```

### Standard Testify Assertions

```go
s.Equal(expected, actual)
s.NotEqual(notExpected, actual)
s.True(condition)
s.False(condition)
s.Nil(value)
s.NotNil(value)
s.Contains(haystack, needle)
s.Greater(a, b)
s.Less(a, b)
s.Len(slice, expectedLength)
s.Empty(value)
s.NotEmpty(value)
```

---

## 5. Configuration for Tests

### Local Mode (`config/local.yaml`)

```yaml
execution:
  mode: local
  server_env: live              # or 'staging'
  skip_sudo_check: true         # Skip sudo checks

logging:
  level: detailed               # minimal, detailed, verbose

database:
  host: localhost
  port: 5432
  user: postgres
  password: ${DB_PASSWORD:-postgres123}  # Environment variable
  name: testdb

postgresql:
  version: "16"
  install_path: "/usr/bin"
```

### Container Mode (`config/container.yaml`)

```yaml
execution:
  mode: container-systemd
  server_env: live
  container:
    os_image: "jrei/systemd-ubuntu:22.04"
    use_systemd: true
    skip_sudo_check: true

logging:
  level: minimal                # Clean output for CI

# ... same as local.yaml for other sections
```

---

## 6. Best Practices

### DO's ✅

1. **Always call parent setup/teardown methods**
   ```go
   func (s *MySuite) SetupSuite() {
       s.E2ESuite.SetupSuite()  // MUST call parent
       // Your setup
   }
   ```

2. **Use `Ensure*` methods for dependencies**
   ```go
   s.EnsureMCPPackagesInstalled()  // Not direct install
   ```

3. **Use descriptive test names**
   ```go
   func (s *Suite) TestUserCanLoginWithValidCredentials() {}
   ```

4. **Add helpful assertion messages**
   ```go
   s.Equal(expected, actual, "User should be logged in")
   ```

5. **Clean up resources in TearDown**
   ```go
   func (s *Suite) TearDownTest() {
       s.StopMCPServer()
       s.CleanupTempFiles()
   }
   ```

### DON'Ts ❌

1. **Don't skip parent setup/teardown calls**
2. **Don't hardcode values** - use config
3. **Don't leave resources running** - clean up in TearDown
4. **Don't create dependencies between tests** - tests should be independent
5. **Don't ignore errors** - always assert on errors

---

## 7. Running Tests

### Run All Tests

```bash
cd AItoolsFramework/mcp-server

# Local mode
make test-local

# Container mode
make test-container
```

### Run Specific Category

```bash
make test-installation
make test-database
make test-service
make test-mcp
make test-kb
```

### Run Specific Test Suite

```bash
go test ./testcases/ -run TestBackupTestSuite -v
```

### Run Specific Test Method

```bash
go test ./testcases/ -run TestBackupTestSuite/TestBackupDatabaseCommand -v
```

### Verbose Output

```bash
go test -v ./testcases/...
```

---

## 8. Debugging Tests

### Enable Detailed Logging

Edit `config/local.yaml`:

```yaml
logging:
  level: verbose  # Or 'detailed'
```

### Check Log Files

```bash
# Local mode logs
cat test-results/test-local.log

# Container mode logs
cat test-results/test-container.log
```

### Debug in VS Code

Add to `.vscode/launch.json`:

```json
{
    "version": "0.2.0",
    "configurations": [
        {
            "name": "Debug Test",
            "type": "go",
            "request": "launch",
            "mode": "test",
            "program": "${workspaceFolder}/AItoolsFramework/mcp-server/testcases",
            "args": ["-test.run", "TestBackupTestSuite"],
            "env": {
                "TESTFW_CONFIG": "config/local.yaml"
            }
        }
    ]
}
```

---

## 9. Quick Reference

### Test Suite Lifecycle

```
SetupSuite()          → Runs once before all tests
  ↓
SetupTest()           → Runs before each test
  ↓
TestMethod1()         → Your test
  ↓
TearDownTest()        → Runs after each test
  ↓
SetupTest()           → Runs before next test
  ↓
TestMethod2()         → Your next test
  ↓
TearDownTest()        → Runs after each test
  ↓
TearDownSuite()       → Runs once after all tests
```

### Common Commands

| Task                     | Command                                      |
| ------------------------ | -------------------------------------------- |
| Run all tests (local)    | `make test-local`                            |
| Run all tests (container)| `make test-container`                        |
| Run specific category    | `make test-<category>`                       |
| Run specific suite       | `go test ./testcases/ -run TestMySuite -v`  |
| Clean cache              | `make clean`                                 |
| Full cleanup             | `make cleanup`                               |
| Help                     | `make help`                                  |

---

## 10. Additional Resources

- **Framework Docs**: [README.md](README.md)
- **Architecture**: [ARCHITECTURE.md](ARCHITECTURE.md)
- **Quick Guide**: [POCKET_GUIDE.md](POCKET_GUIDE.md)
- **Optimization Details**: [OPTIMIZATION.md](OPTIMIZATION.md)
- **Test Examples**: [testcases/example_test.go](mcp-server/testcases/example_test.go)
- **Config Guide**: [testcases/CONFIG_USAGE.md](mcp-server/testcases/CONFIG_USAGE.md)

---

## Need Help?

If you encounter issues:

1. Check the logs in `test-results/`
2. Review existing test files for patterns
3. Consult the framework documentation
4. Ask the development team

Happy Testing! 🧪
