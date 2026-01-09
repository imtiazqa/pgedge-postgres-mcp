# AItoolsFramework - Folder Structure Guide

Concise two-liner description for every folder in the AItoolsFramework.

---

## Root Level

### `/AItoolsFramework/`

**Purpose**: Enterprise test automation framework for pgEdge MCP server testing.
Contains shared framework libraries and MCP server test implementation.

---

## Common Framework (`/common/`)

### `/common/`

**Purpose**: Shared test framework library (Go module) used across all test
projects. Provides reusable components like base suites, configuration
management, executors, and utilities.

---

### `/common/assertions/`

**Purpose**: Custom assertion helpers for test validation (currently empty
placeholder). Will contain specialized assertion functions beyond standard
testify assertions.

---

### `/common/config/`

**Purpose**: Configuration loading, parsing, and validation for test execution.
Contains type definitions (`types.go`), YAML loader (`loader.go`), and
validator (`validator.go`).

---

### `/common/database/`

**Purpose**: Database helper utilities for test data management (currently
empty placeholder). Will provide connection pooling, seeding, and transaction
helpers.

---

### `/common/executor/`

**Purpose**: Test execution mode implementations (local, container, SSH,
Kubernetes). Handles command execution in different environments with unified
interface (`executor.go`, `local.go`, `container.go`).

---

### `/common/fixtures/`

**Purpose**: Test data and fixtures management utilities (currently empty
placeholder). Will provide helpers for loading and managing test data files.

---

### `/common/http/`

**Purpose**: HTTP client utilities for API testing (currently empty
placeholder). Will contain reusable HTTP request/response helpers and client
configurations.

---

### `/common/mcp/`

**Purpose**: MCP protocol helper utilities (currently empty placeholder).
Will provide MCP message construction, parsing, and protocol-specific helpers.

---

### `/common/reporters/`

**Purpose**: Test result reporting and formatting utilities (currently empty
placeholder). Will generate test reports in various formats (JSON, HTML, XML).

---

### `/common/suite/`

**Purpose**: Base test suite implementations providing core testing
functionality. Contains `BaseSuite`, `E2ESuite` with installation helpers
(`install.go`), and lifecycle management.

**Files**:

- `base.go` - BaseSuite with context, executor, config, and test tracking
- `e2e.go` - E2ESuite for end-to-end tests with global installation state
- `install.go` - Installation helpers for repository, PostgreSQL, MCP packages
- `README.md` - Suite usage documentation

---

### `/common/utils/`

**Purpose**: Shared utility functions and helper scripts for common tasks.
Contains cleanup script (`cleanup.sh`) for removing test artifacts and
packages.

---

## MCP Server Test Project (`/mcp-server/`)

### `/mcp-server/`

**Purpose**: Main MCP server test implementation project (Go module).
Contains actual test cases, configurations, fixtures, and test execution
orchestration.

---

### `/mcp-server/assertions/`

**Purpose**: MCP-specific assertion helpers (currently empty placeholder).
Will contain assertions for MCP protocol responses, server states, and message
validation.

---

### `/mcp-server/config/`

**Purpose**: Test execution configuration files for different environments.
Contains YAML configs for local execution (`local.yaml`) and container
execution (`container.yaml`).

**Files**:

- `local.yaml` - Configuration for running tests on local machine
- `container.yaml` - Configuration for running tests in Docker containers

---

### `/mcp-server/docs/`

**Purpose**: Test project documentation and guides (currently minimal).
Will contain detailed test documentation, architecture diagrams, and usage
guides.

---

### `/mcp-server/fixtures/`

**Purpose**: Test data, fixtures, and expected responses for test execution.
Contains database schemas, test data, config files, and expected API
responses.

---

### `/mcp-server/fixtures/configs/`

**Purpose**: Configuration file fixtures for testing different server
configurations (currently empty). Will store sample config files for various
test scenarios.

---

### `/mcp-server/fixtures/databases/`

**Purpose**: Database schema and test data SQL files for database tests.
Contains `schema.sql` for table definitions and `testdata.sql` for seed data.

**Files**:

- `schema.sql` - Database schema definitions (tables, indexes, constraints)
- `testdata.sql` - Test data seed scripts for database validation

---

### `/mcp-server/fixtures/responses/`

**Purpose**: Expected API/MCP response fixtures for validation (currently
empty). Will contain JSON files with expected responses for protocol tests.

---

### `/mcp-server/test-results/`

**Purpose**: Auto-generated test execution results, logs, and reports.
Contains timestamped log files for local and container test runs.

**Generated Files**:

- `test-local.log` - Detailed logs from local execution mode
- `test-container.log` - Detailed logs from container execution mode

---

### `/mcp-server/testcases/`

**Purpose**: Actual test case implementations for all MCP server
functionality. Contains 15 test files covering installation, database,
service, MCP protocol, and knowledge base tests.

**Test Files**:

- `installation_test.go` - Binary installation and package verification tests
- `repository_test.go` - Repository setup and configuration tests
- `postgresql_test.go` - PostgreSQL installation and setup tests
- `files_test.go` - File and directory verification tests
- `user_test.go` - User management and permissions tests
- `token_test.go` - Authentication token management tests
- `mcp_server_test.go` - MCP server core functionality tests
- `service_test.go` - systemd service management tests
- `stdio_test.go` - MCP stdio mode communication tests
- `kb_test.go` - Knowledge base builder and functionality tests
- `mcp_kb_test.go` - MCP and knowledge base integration tests
- `regression_suite_test.go` - Consolidated regression test suite (11 tests)
- `example_test.go` - Framework usage examples and patterns
- `suite_test.go` - Test suite setup and initialization
- `helpers_test.go` - Shared test helper functions

---

### `/mcp-server/utils/`

**Purpose**: MCP-specific utility functions and helpers (currently empty
placeholder). Will contain test-specific utilities for MCP protocol handling
and data manipulation.

---

## Summary Statistics

### Common Framework

- **Total Folders**: 11
- **Populated Folders**: 3 (config, executor, suite)
- **Empty Placeholder Folders**: 8 (for future expansion)
- **Purpose**: Reusable framework library

### MCP Server Tests

- **Total Folders**: 8
- **Test Files**: 15 test files
- **Configuration Files**: 2 (local.yaml, container.yaml)
- **Purpose**: Actual test implementation

---

## Folder Purpose Categories

### Configuration & Setup (3 folders)

- `/common/config/` - Framework configuration management
- `/mcp-server/config/` - Test execution configurations
- `/common/suite/` - Base suite setup and lifecycle

### Execution & Orchestration (2 folders)

- `/common/executor/` - Multi-environment execution modes
- `/mcp-server/testcases/` - Test case implementations

### Data & Fixtures (2 folders)

- `/mcp-server/fixtures/` - Test data and expected results
- `/mcp-server/fixtures/databases/` - Database schemas and seed data

### Results & Reporting (1 folder)

- `/mcp-server/test-results/` - Test execution logs and reports

### Utilities & Helpers (3 folders)

- `/common/utils/` - Shared utility scripts
- `/mcp-server/utils/` - Test-specific utilities
- `/mcp-server/docs/` - Documentation

### Placeholders for Future Expansion (8 folders)

Currently empty, reserved for future features:

- `/common/assertions/` - Custom assertions
- `/common/database/` - Database helpers
- `/common/fixtures/` - Fixture management
- `/common/http/` - HTTP utilities
- `/common/mcp/` - MCP protocol helpers
- `/common/reporters/` - Report generators
- `/mcp-server/assertions/` - MCP assertions
- `/mcp-server/fixtures/configs/` - Config fixtures
- `/mcp-server/fixtures/responses/` - Response fixtures

---

## Quick Navigation

**Writing Tests?** → `/mcp-server/testcases/`

**Configuring Tests?** → `/mcp-server/config/`

**Adding Test Data?** → `/mcp-server/fixtures/`

**Checking Results?** → `/mcp-server/test-results/`

**Understanding Suites?** → `/common/suite/`

**Modifying Execution?** → `/common/executor/`

---

## Visual Hierarchy

```
AItoolsFramework/
│
├── common/                         # Shared framework (library)
│   ├── config/                     # [ACTIVE] Configuration system
│   ├── executor/                   # [ACTIVE] Execution modes
│   ├── suite/                      # [ACTIVE] Base test suites
│   ├── utils/                      # [ACTIVE] Utility scripts
│   ├── assertions/                 # [PLACEHOLDER] Custom assertions
│   ├── database/                   # [PLACEHOLDER] DB helpers
│   ├── fixtures/                   # [PLACEHOLDER] Fixture management
│   ├── http/                       # [PLACEHOLDER] HTTP utilities
│   ├── mcp/                        # [PLACEHOLDER] MCP helpers
│   └── reporters/                  # [PLACEHOLDER] Report generators
│
└── mcp-server/                     # Test implementation (project)
    ├── testcases/                  # [ACTIVE] Test implementations (15 files)
    ├── config/                     # [ACTIVE] Execution configs (2 files)
    ├── fixtures/                   # [ACTIVE] Test data
    │   ├── databases/              # [ACTIVE] SQL schemas and data
    │   ├── configs/                # [PLACEHOLDER] Config fixtures
    │   └── responses/              # [PLACEHOLDER] Response fixtures
    ├── test-results/               # [AUTO-GENERATED] Test logs
    ├── Makefile                    # [ACTIVE] Test runner
    ├── assertions/                 # [PLACEHOLDER] MCP assertions
    ├── utils/                      # [PLACEHOLDER] Test utilities
    └── docs/                       # [MINIMAL] Documentation
```

**Legend**:

- **[ACTIVE]** - Currently in use with files
- **[PLACEHOLDER]** - Empty, reserved for future expansion
- **[AUTO-GENERATED]** - Created during test execution
- **[MINIMAL]** - Has some content, needs expansion

---

## Total Count

- **Total Folders**: 23
- **Active Folders**: 8
- **Placeholder Folders**: 13
- **Auto-Generated Folders**: 1
- **Test Files**: 15
- **Config Files**: 5 (3 in common/config, 2 in mcp-server/config)

---

For detailed usage instructions, see:

- [TESTER_GUIDE.md](TESTER_GUIDE.md) - Complete tester guide
- [POCKET_GUIDE.md](POCKET_GUIDE.md) - Quick reference
- [README.md](README.md) - Framework overview
- [mcp-server/testcases/README.md](mcp-server/testcases/README.md) - Test
  case guide
