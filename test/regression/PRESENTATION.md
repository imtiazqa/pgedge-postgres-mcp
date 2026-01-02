# pgEdge Postgres MCP Regression Test Suite

Framework & Libraries Overview

---

## Overview

**A production-grade Go-based regression test suite**

- 11 comprehensive test cases validating complete installation pipeline
- Multiple execution modes (container & local)
- Support for 5+ Linux distributions (AlmaLinux, Debian, Ubuntu, Rocky)
- Full database connectivity validation
- Knowledge Base (KB) functionality testing with Ollama embeddings

**Technology Stack:** Go 1.23.0 + Docker + systemd + Ollama

---

## Test Execution Workflow

```
┌─────────────────────────────────────────────────────────────┐
│                    START: Execute_Regression_suite          │
└──────────────────────────┬──────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│  SetupSuite() - Once Before All Tests                       │
│  • Read configuration (.test.env or interactive)            │
│  • Create executor (Container or Local)                     │
│  • Pull/prepare OS image (if container mode)                │
│  • Start container with systemd OR prepare local env        │
│  • Start elephant animation 🐘                              │
└──────────────────────────┬──────────────────────────────────┘
                           │
                           ▼
           ┌───────────────────────────────┐
           │   For Each Test (1-11)        │
           └───────┬───────────────────────┘
                   │
                   ▼
        ┌──────────────────────────┐
        │  SetupTest()             │
        │  • Log test start        │
        │  • Start timer           │
        └──────┬───────────────────┘
               │
               ▼
┌──────────────────────────────────────────────────────────────┐
│  Execute Test Case                                           │
│                                                               │
│  Test01: Repository Installation                             │
│    └─> Install pgEdge repo (APT/DNF) → Verify packages       │
│                                                               │
│  Test02: PostgreSQL Installation                             │
│    └─> Install PG → initdb → Create DB → Configure           │
│                                                               │
│  Test03: MCP Server Package Installation                     │
│    └─> Install pgedge-postgres-mcp + CLI + Web + KB          │
│                                                               │
│  Test04: Installation Validation                             │
│    └─> Verify binaries, configs, directories exist           │
│                                                               │
│  Test05: Token Management                                    │
│    └─> add-token → list-tokens → Verify token file           │
│                                                               │
│  Test06: User Management                                     │
│    └─> add-user → list-users → Verify user file              │
│                                                               │
│  Test07: Files and Permissions                               │
│    └─> Check all files, ownership, permissions               │
│                                                               │
│  Test08: Service Management                                  │
│    └─> Start service → Verify active → Test HTTP :8080       │
│                                                               │
│  Test09: Knowledge Builder                                   │
│    └─> Install Ollama → Load nomic-embed-text → Build KB     │
│                                                               │
│  Test10: MCP Server with Knowledge Base                      │
│    └─> Start MCP + KB → Search docs → Verify embeddings      │
│                                                               │
│  Test11: Stdio Mode & Database Connectivity                  │
│    └─> Start stdio → JSON-RPC → Query DB → Verify schema     │
│                                                               │
└──────────────────────┬───────────────────────────────────────┘
                       │
                       ▼
        ┌──────────────────────────┐
        │  TearDownTest()          │
        │  • Record duration       │
        │  • Display result        │
        │  • Log failures          │
        └──────┬───────────────────┘
               │
               ▼
           ┌───────────────────┐
           │  More tests?      │
           └────┬──────────┬───┘
           Yes  │          │ No
                │          │
        ┌───────┘          └───────┐
        │                          │
        │                          ▼
        │        ┌─────────────────────────────────────────────┐
        │        │  TearDownSuite() - Once After All Tests     │
        │        │  • Stop elephant animation                  │
        │        │  • Cleanup executor (remove container)      │
        │        │  • Print summary table with all results     │
        │        │  • Display: Pass/Fail, durations, config    │
        │        └──────────────────┬──────────────────────────┘
        │                           │
        │                           ▼
        │                ┌──────────────────────┐
        │                │  END: Exit with code │
        │                │  0 (pass) or 1 (fail)│
        │                └──────────────────────┘
        │
        └─> Loop back to "SetupTest()" for next test
```

---

## Testing Framework: Testify Suite

**Primary Framework:** `github.com/stretchr/testify v1.11.1`

**Why Testify Suite?**
- Industry-standard Go testing framework
- Provides rich assertion library
- Suite-based testing with lifecycle hooks
- Clean test organization and setup/teardown

**Key Components Used:**
```go
import "github.com/stretchr/testify/suite"

type RegressionTestSuite struct {
    suite.Suite  // Embeds testify suite
    // ... test fields
}
```

---

## Testify Suite Lifecycle Hooks

**SetupSuite()** - Runs once before all tests
- Initialize context, executor, configuration
- Start containers or prepare local environment
- Show animated elephant progress indicator

**SetupTest()** - Runs before each test
- Log test start, initialize timing

**TearDownTest()** - Runs after each test
- Record duration/status, display result, log failures

**TearDownSuite()** - Runs once after all tests
- Stop animation, cleanup executor, print summary table

---

## Container Management: Docker SDK

**Library:** `github.com/docker/docker v25.0.0`

**Capabilities:**
- Programmatic Docker container lifecycle management
- Image pulling and caching
- Container creation, start, stop, removal
- Command execution inside containers
- Log retrieval for debugging

**Key Features Used:**
- Container commit (preserve systemd installation)
- Privileged mode with cgroup mounting
- DNS configuration
- Volume mounting
- Network management

---

## Docker SDK Implementation

**Container Executor Architecture:**

```go
type ContainerExecutor struct {
    client      *client.Client  // Docker SDK client
    containerID string
    osImage     string
    // ... other fields
}

// Core Docker operations
client.ImagePull(ctx, image, options)
client.ContainerCreate(ctx, config, hostConfig, nil, "")
client.ContainerStart(ctx, containerID, options)
client.ContainerExecCreate(ctx, containerID, execConfig)
client.ContainerRemove(ctx, containerID, options)
```

**Systemd Support:** Multi-stage container setup with commit

---

## Output Formatting: go-pretty

**Library:** `github.com/jedib0t/go-pretty/v6 v6.6.4`

**Purpose:** Beautiful, formatted table output for test results

**Features:**
- Border styling and customization
- Column alignment and width
- Color support for status indicators
- Professional table rendering

**Usage in Test Suite:**
```go
import "github.com/jedib0t/go-pretty/v6/table"

t := table.NewWriter()
t.AppendHeader(table.Row{"TEST", "STATUS", "DURATION"})
t.AppendRow(table.Row{name, "✓ PASS", "12.3s"})
t.Render()
```

---

## Additional Key Dependencies

**Direct Dependencies (go.mod):**

| Library | Version | Purpose |
|---------|---------|---------|
| docker/docker | v25.0.0 | Container orchestration |
| go-pretty/v6 | v6.6.4 | Table formatting |
| testify | v1.11.1 | Testing framework |

**Notable Indirect Dependencies:**
- `go.opentelemetry.io/*` - Observability and tracing
- `docker/go-connections` - Docker networking
- `gopkg.in/yaml.v3` - YAML parsing (configs)
- `golang.org/x/sync` - Concurrency primitives

---

## Architecture Pattern: Executor Abstraction

**Design Pattern:** Strategy Pattern for execution modes

```go
type Executor interface {
    Start(ctx context.Context) error
    Exec(ctx context.Context, cmd string) (string, int, error)
    Cleanup(ctx context.Context) error
    GetLogs(ctx context.Context) (string, error)
    Mode() ExecutionMode
    GetOSInfo(ctx context.Context) (string, error)
}
```

**Implementations:**
- `ContainerExecutor` - Docker-based execution
- `LocalExecutor` - Direct host execution

**Benefit:** Same test code runs in any environment

---

## Unique Features

**1. Animated Progress Indicator**
```go
// Background goroutine with channel control
go s.animateElephant()
ticker := time.NewTicker(500 * time.Millisecond)
fmt.Printf("\rTests running... 🐘")
```

**2. State Management**
```go
setupState struct {
    repoInstalled        bool
    postgresqlInstalled  bool
    mcpPackagesInstalled bool
}
```

**3. Intelligent Caching**
- Setup operations run once
- Container state preservation via commit
- Image layer caching

---

## Summary

**Framework Stack:**
- **Testing:** Testify Suite (v1.11.1) - lifecycle hooks, assertions
- **Container:** Docker SDK (v25.0.0) - orchestration, systemd support
- **Output:** go-pretty (v6.6.4) - beautiful table formatting
- **Language:** Go 1.23.0 with modern concurrency

**Key Strengths:**
- Production-grade testing framework
- Real integration testing with actual database queries
- Beautiful output with animated progress
- Flexible execution modes (container/local)
- Robust error handling and cleanup

**Result:** Confidence in deployment quality across multiple platforms

🐘 **Happy Testing!**
