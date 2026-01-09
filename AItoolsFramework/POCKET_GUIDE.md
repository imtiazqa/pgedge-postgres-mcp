# AItoolsFramework Pocket Guide

Quick reference for common tasks and commands.

## Quick Start

```bash
cd AItoolsFramework/mcp-server

# Run tests locally
make test-local

# Run tests in container
make test-container
```

## Configuration Files

```
config/local.yaml       # Local execution (development)
config/container.yaml   # Container execution (CI/CD)
```

**Switch between live/staging:**
```yaml
execution:
  server_env: live      # Change to 'staging' for staging repos
```

## Makefile Commands

```bash
# Main test targets
make test-local          # Run tests on local machine
make test-container      # Run tests in Docker container

# Run specific categories
make test-installation   # Installation tests only
make test-database       # Database tests only
make test-service        # Service tests only
make test-mcp            # MCP protocol tests only
make test-kb             # Knowledge base tests only

# Utilities
make clean               # Clean test cache
make cleanup             # Full cleanup (packages, containers)
make help                # Show all targets
```

## Logging Levels

Edit `logging.level` in config files:

```yaml
logging:
  level: minimal         # Summary only (for CI)
  level: detailed        # Full logs (for debugging)
  level: verbose         # Maximum detail
```

**Output:**
- **minimal**: Clean summary on terminal, full logs in file
- **detailed**: Everything on terminal and in file
- **verbose**: Maximum debugging information

## Test Suite Types

### E2ESuite
For installation, service, file tests:
```go
type MyTestSuite struct {
    suite.E2ESuite
}

func (s *MyTestSuite) TestExample() {
    s.EnsureMCPPackagesInstalled()  // Auto-installs dependencies
    s.AssertFileExists("/usr/bin/pgedge-postgres-mcp")
}
```

### APISuite
For MCP protocol tests:
```go
type MyMCPSuite struct {
    suite.APISuite
}

func (s *MyMCPSuite) TestMCP() {
    s.StartMCPServer(binary, config, mcp.ModeStdio)
    defer s.StopMCPServer()

    resp, _ := s.MCPServer.Initialize(s.Ctx)
    s.MCPAssertions.AssertValidInitializeResponse(resp)
}
```

### DatabaseSuite
For database tests:
```go
type MyDBSuite struct {
    suite.DatabaseSuite
}

func (s *MyDBSuite) TestDB() {
    db := s.GetDB()
    s.SeedTable("users", []string{"name"}, [][]interface{}{{"Alice"}})
    s.AssertRowCount("users", 1)
}
```

## Common Assertions

```go
// File assertions
s.AssertFileExists("/path/to/file")
s.AssertFileNotExists("/path/to/file")
s.AssertFileContains("/path", "content")

// Command execution
output, exitCode, err := s.ExecCommand("ls -la")
s.NoError(err)
s.Equal(0, exitCode)

// Database assertions (DatabaseSuite)
s.AssertRowCount("table_name", 5)
s.AssertTableExists("table_name")

// MCP assertions (APISuite)
s.MCPAssertions.AssertValidInitializeResponse(resp)
s.MCPAssertions.AssertToolExists("query_database")
```

## Configuration Access

```go
// Get config values
version := s.Config.PostgreSQL.Version
binary := s.Config.GetBinaryPath("mcp_server")
configDir := s.Config.GetConfigDir()

// Database config
host := s.Config.Database.Host
port := s.Config.Database.Port
```

## Environment Variables

Use in config files:
```yaml
database:
  password: ${DB_PASSWORD:-postgres123}  # Default if not set
  host: ${DB_HOST:-localhost}
```

## Test Structure

```
testcases/
├── installation/    # Package installation, files, users
├── database/        # PostgreSQL, database operations
├── service/         # systemd, service management
├── mcp/             # MCP protocol, stdio, tokens
├── kb/              # Knowledge base tests
└── examples/        # Framework usage examples
```

## Running Specific Tests

```bash
# By category
go test ./testcases/installation/

# By test name
go test ./testcases/installation/ -run TestBinaryInstallation

# Verbose output
go test -v ./testcases/mcp/

# All tests
go test -v ./testcases/...
```

## Execution Modes

### Local Mode
```yaml
execution:
  mode: local
  skip_sudo_check: true    # For tests without sudo
```
- Runs on your machine
- Fast
- Uses existing installation

### Container Mode
```yaml
execution:
  mode: container-systemd
  container:
    os_image: "jrei/systemd-ubuntu:22.04"
    use_systemd: true
    skip_sudo_check: true
```
- Runs in Docker
- Fresh environment
- Auto-installs everything
- Perfect for CI/CD

## Troubleshooting

**Tests failing?**
1. Check config file path: `TESTFW_CONFIG=config/local.yaml`
2. Check logging level: Set to `detailed` for more info
3. Check logs: `cat test-results/test-local.log`

**Installation issues?**
1. Ensure `sudo` access (local mode)
2. Check `server_env` setting (live vs staging)
3. Check PostgreSQL version in config
4. Check repository URLs

**Container issues?**
1. Ensure Docker is running
2. Check systemd image: `jrei/systemd-ubuntu:22.04`
3. Set `skip_sudo_check: true` in container config

## File Locations

```
# Test results
test-results/test-local.log      # Local mode logs
test-results/test-container.log  # Container mode logs

# Configuration
config/local.yaml                # Local config
config/container.yaml            # Container config

# Framework
common/suite/                    # Base suites
common/config/types.go           # Config structure
common/executor/                 # Execution modes
```

## Quick Tips

1. **Always use `Ensure*` methods** - Never install directly
2. **Use minimal logging in CI** - Cleaner output
3. **Use container mode for isolation** - Fresh environment every time
4. **Check full logs** - `test-results/*.log` has everything
5. **One suite installation** - Global state optimizes speed

## Common Patterns

### Setup Suite
```go
func (s *MyTestSuite) SetupSuite() {
    s.E2ESuite.SetupSuite()  // Always call parent
    s.EnsureMCPPackagesInstalled()
}
```

### Test Method
```go
func (s *MyTestSuite) TestExample() {
    // 1. Setup (if needed)

    // 2. Execute
    output, exitCode, err := s.ExecCommand("command")

    // 3. Assert
    s.NoError(err)
    s.Equal(0, exitCode)
    s.Contains(output, "expected")
}
```

### Cleanup
```go
func (s *MyTestSuite) TearDownSuite() {
    s.E2ESuite.TearDownSuite()  // Always call parent
    // Custom cleanup if needed
}
```

## Need More Help?

- [README.md](README.md) - Full documentation
- [ARCHITECTURE.md](ARCHITECTURE.md) - Framework design
- [OPTIMIZATION.md](OPTIMIZATION.md) - Performance details
- [testcases/README.md](mcp-server/testcases/README.md) - Test guide
- Examples: `testcases/examples/`

---

**Quick Reference Card**

| Task | Command |
|------|---------|
| Run local tests | `make test-local` |
| Run container tests | `make test-container` |
| Run specific category | `make test-<category>` |
| Clean cache | `make clean` |
| Switch to staging | Edit config: `server_env: staging` |
| More logging | Edit config: `level: detailed` |
| Check logs | `cat test-results/*.log` |
| Help | `make help` |
