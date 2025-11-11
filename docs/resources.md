# MCP Resources

Resources provide read-only access to PostgreSQL system information and statistics. All resources are accessed via the `read_resource` tool or through MCP protocol resource methods.

## System Information Resources

### pg://settings

Returns PostgreSQL server configuration parameters including current values, default values, pending changes, and descriptions.

**Access**: Read the resource to view all PostgreSQL configuration settings from pg_settings.

**Output**: JSON array with detailed information about each configuration parameter:

```json
[
  {
    "name": "max_connections",
    "current_value": "100",
    "category": "Connections and Authentication / Connection Settings",
    "description": "Sets the maximum number of concurrent connections.",
    "context": "postmaster",
    "type": "integer",
    "source": "configuration file",
    "min_value": "1",
    "max_value": "262143",
    "default_value": "100",
    "reset_value": "100",
    "pending_restart": false
  },
  ...
]
```

### pg://system_info

Returns PostgreSQL version, operating system, and build architecture information. Provides a quick and efficient way to check server version and platform details without executing natural language queries.

**Access**: Read the resource to view PostgreSQL system information.

**Output**: JSON object with detailed system information:

```json
{
  "postgresql_version": "15.4",
  "version_number": "150004",
  "full_version": "PostgreSQL 15.4 on x86_64-pc-linux-gnu, compiled by gcc (GCC) 11.2.0, 64-bit",
  "operating_system": "linux",
  "architecture": "x86_64-pc-linux-gnu",
  "compiler": "gcc (GCC) 11.2.0",
  "bit_version": "64-bit"
}
```

**Fields:**

- `postgresql_version`: Short version string (e.g., "15.4")
- `version_number`: Numeric version identifier (e.g., "150004")
- `full_version`: Complete version string from PostgreSQL version() function
- `operating_system`: Operating system (e.g., "linux", "darwin", "mingw32")
- `architecture`: Full architecture string (e.g., "x86_64-pc-linux-gnu", "aarch64-apple-darwin")
- `compiler`: Compiler used to build PostgreSQL (e.g., "gcc (GCC) 11.2.0")
- `bit_version`: Architecture bit version (e.g., "64-bit", "32-bit")

**Use Cases:**

- Quickly check PostgreSQL version without natural language queries
- Verify server platform and architecture
- Audit server build information
- Troubleshoot compatibility issues

## Statistics Resources

All statistics resources are compatible with PostgreSQL 14 and later. They provide real-time monitoring data from PostgreSQL's `pg_stat_*` system views.

### pg://stat/activity

Shows information about currently executing queries and connections. Essential for monitoring active database sessions and identifying long-running queries.

**Output**: JSON with current database activity:

```json
{
  "activity_count": 5,
  "activities": [
    {
      "datname": "mydb",
      "pid": 12345,
      "usename": "myuser",
      "application_name": "psql",
      "client_addr": "127.0.0.1",
      "backend_start": "2024-10-30T10:00:00",
      "state": "active",
      "query": "SELECT * FROM users"
    }
  ],
  "description": "Current database activity showing all non-idle connections and their queries."
}
```

**Use Cases:**

- Monitor currently executing queries
- Identify long-running queries
- Track connection counts
- Troubleshoot performance issues

### pg://stat/replication

Shows the status of replication connections from this primary server including WAL sender processes, replication lag, and sync state. Empty if the server is not a replication primary or has no active replicas.

**Output**: JSON with replication status:

```json
{
  "replica_count": 2,
  "replicas": [
    {
      "pid": 12345,
      "usename": "replicator",
      "application_name": "walreceiver",
      "client_addr": "192.168.1.100",
      "client_hostname": "replica1",
      "client_port": 5432,
      "backend_start": "2024-10-30T10:00:00",
      "state": "streaming",
      "sync_state": "async",
      "replay_lag": "00:00:02"
    }
  ],
  "status": "Primary server with 2 active replica(s)",
  "description": "Replication status for all connected standby servers. Monitor replay_lag to detect replication delays."
}
```

**Key Fields:**

- `state`: Replication state (startup, catchup, streaming, backup, stopping)
- `sync_state`: Synchronization state (sync, async, quorum, potential)
- `replay_lag`: Time delay between primary and replica

**Use Cases:**

- Monitor replication health
- Identify replication lag issues
- Verify replica connections
- Track synchronous vs asynchronous replicas

## Accessing Resources

Resources can be accessed in two ways:

### 1. Via read_resource Tool

```json
{
  "uri": "pg://system_info"
}
```

Or list all resources:
```json
{
  "list": true
}
```

### 2. Via Natural Language (Claude Desktop)

Simply ask Claude to read a resource:

- "Show me the output from pg://system_info"
- "Read the pg://settings resource"
- "What's the current PostgreSQL version?" (uses pg://system_info)
- "Show me current database activity" (uses pg://stat/activity)
