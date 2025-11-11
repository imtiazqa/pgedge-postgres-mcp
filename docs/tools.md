# MCP Tools

The pgEdge MCP Server provides ten tools that enable natural language database interaction, configuration management, connection management, and server information.

## Available Tools

### server_info

Get information about the MCP server itself, including server name, company, version, LLM provider, and model being used.

**Input**: None (no parameters required)

**Output**:
```
Server Information:
===================

Server Name:    pgEdge PostgreSQL MCP Server
Company:        pgEdge, Inc.
Version:        1.0.0

LLM Provider:   anthropic
LLM Model:      claude-sonnet-4-5

Description:    An MCP (Model Context Protocol) server that enables AI assistants to interact with PostgreSQL databases through natural language queries and schema exploration.

License:        PostgreSQL License
Copyright:      © 2025, pgEdge, Inc.
```

**Use Cases**:
- Check which LLM provider and model the server is configured to use
- Verify server version for compatibility and troubleshooting
- Get quick reference to server information during support requests

### query_database

Executes a natural language query against the PostgreSQL database. Supports dynamic connection strings to query different databases.

**Input Examples**:

Basic query:
```json
{
  "query": "Show me all users created in the last week"
}
```

Query with temporary connection:
```json
{
  "query": "Show me table list at postgres://localhost:5433/other_db"
}
```

Set new default connection:
```json
{
  "query": "Set default database to postgres://localhost/analytics"
}
```

**Output**:
```
Natural Language Query: Show me all users created in the last week

Generated SQL:
SELECT * FROM users WHERE created_at >= NOW() - INTERVAL '7 days' ORDER BY created_at DESC

Results (15 rows):
[
  {
    "id": 123,
    "username": "john_doe",
    "created_at": "2024-10-25T14:30:00Z",
    ...
  },
  ...
]
```

**Security**: All queries are executed in read-only transactions using `SET TRANSACTION READ ONLY`, preventing INSERT, UPDATE, DELETE, and other data modifications. Write operations will fail with "cannot execute ... in a read-only transaction".

### get_schema_info

Retrieves database schema information including tables, views, columns, data types, and comments from pg_description.

**Input** (optional):
```json
{
  "schema_name": "public"
}
```

**Output**:
```
Database Schema Information:
============================

public.users (TABLE)
  Description: User accounts and authentication
  Columns:
    - id: bigint
    - username: character varying(255)
      Description: Unique username for login
    - created_at: timestamp with time zone (nullable)
      Description: Account creation timestamp
    ...
```

### set_pg_configuration

Sets PostgreSQL server configuration parameters using ALTER SYSTEM SET. Changes persist across server restarts. Some parameters require a restart to take effect.

**Input**:
```json
{
  "parameter": "max_connections",
  "value": "200"
}
```

Use "DEFAULT" as the value to reset to default:
```json
{
  "parameter": "work_mem",
  "value": "DEFAULT"
}
```

**Output**:
```
Configuration parameter 'max_connections' updated successfully.

Parameter: max_connections
Description: Sets the maximum number of concurrent connections
Type: integer
Context: postmaster

Previous value: 100
New value: 200

⚠️  WARNING: This parameter requires a server restart to take effect.
The change has been saved to postgresql.auto.conf but will not be active until the server is restarted.

SQL executed: ALTER SYSTEM SET max_connections = '200'
```

**Security Considerations**:
- Requires PostgreSQL superuser privileges
- Changes persist across server restarts via `postgresql.auto.conf`
- Test configuration changes in development before applying to production
- Some parameters require a server restart to take effect
- Keep backups of configuration files before making changes


### read_resource

Reads MCP resources by their URI. Provides access to system information and statistics.

**Input Examples**:

List all available resources:

```json
{
  "list": true
}
```

Read a specific resource:

```json
{
  "uri": "pg://system_info"
}
```

**Available Resource URIs**:

- `pg://settings` - PostgreSQL configuration parameters
- `pg://system_info` - PostgreSQL version, OS, and build architecture
- `pg://stat/activity` - Current connections and queries
- `pg://stat/replication` - Replication status

See [resources.md](resources.md) for detailed information about each resource.
## Connection Management Tools

### add_database_connection

Save a database connection with an alias for later use. Connections are persisted and available across sessions.

**Input**:
```json
{
  "alias": "production",
  "connection_string": "postgres://user:pass@host:5432/database",
  "maintenance_db": "postgres",
  "description": "Production database server"
}
```

**Parameters**:
- `alias` (required): Friendly name for the connection (e.g., "production", "staging")
- `connection_string` (required): PostgreSQL connection string
- `maintenance_db` (optional, default: "postgres"): Initial database for connections, like pgAdmin
- `description` (optional): Notes about this connection

**Output**:
```
Successfully saved connection 'production'
Connection string: postgres://user:****@host:5432/database
Maintenance DB: postgres
Description: Production database server
```

**Storage**:
- **With authentication enabled**: Stored per-token in `api-tokens.yaml`
- **With authentication disabled**: Stored globally in config file under `database.connections`

### remove_database_connection

Remove a saved database connection by its alias.

**Input**:
```json
{
  "alias": "staging"
}
```

**Output**:
```
Successfully removed connection 'staging'
```

### list_database_connections

List all saved database connections for the current user/session.

**Input**: None (no parameters required)

**Output**:
```
Saved Database Connections:
============================

Alias: production
Connection: postgres://user:****@prod-host:5432/mydb
Maintenance DB: postgres
Description: Production database
Created: 2025-01-15 10:00:00
Last Used: 2025-01-15 14:30:00

Alias: staging
Connection: postgres://user:****@staging-host:5432/mydb
Maintenance DB: postgres
Description: Staging environment
Created: 2025-01-15 10:05:00
Last Used: Never

Total: 2 saved connection(s)
```

**Note**: Connection strings are masked for security (passwords shown as `****`)

### edit_database_connection

Update an existing saved connection. You can update any or all fields.

**Input**:
```json
{
  "alias": "production",
  "new_connection_string": "postgres://newuser:newpass@newhost:5432/newdb",
  "new_maintenance_db": "template1",
  "new_description": "Updated production server"
}
```

**Parameters**:
- `alias` (required): The alias of the connection to update
- `new_connection_string` (optional): New connection string
- `new_maintenance_db` (optional): New maintenance database
- `new_description` (optional): New description

**Output**:
```
Successfully updated connection 'production'
Updated fields: connection_string, maintenance_db, description
```

### set_database_connection (Enhanced)

Set the database connection for the current session. Now supports both connection strings and aliases.

**Input with alias**:
```json
{
  "connection_string": "production"
}
```

**Input with full connection string**:
```json
{
  "connection_string": "postgres://user:pass@host:5432/database"
}
```

**Behavior**:
- If the input looks like an alias (no `postgres://` or `postgresql://` prefix), it attempts to resolve it from saved connections
- If the alias is found, it uses the saved connection string
- If not found, it treats the input as a literal connection string
- Successfully used aliases are marked with a "last used" timestamp

**Output with alias**:
```
Successfully connected to database using alias 'production'
Loaded metadata for 142 tables/views.
```

**Output with connection string**:
```
Successfully connected to database.
Loaded metadata for 142 tables/views.
```

## Connection Management Workflow

Here's a typical workflow for managing database connections:

```
1. Save connections with friendly names:
   add_database_connection(
     alias="prod",
     connection_string="postgres://...",
     description="Production DB"
   )

2. List saved connections:
   list_database_connections()

3. Connect using an alias:
   set_database_connection(connection_string="prod")

4. Work with the database:
   query_database(query="Show me...")
   get_schema_info()

5. Update a connection if needed:
   edit_database_connection(
     alias="prod",
     new_description="Production DB - Updated"
   )

6. Remove old connections:
   remove_database_connection(alias="old_staging")
```

## Security Considerations

- **Authentication Enabled (per-token connections)**:
  - Each API token has its own isolated set of saved connections
  - Users cannot see or access connections from other tokens
  - Connections are stored in `api-tokens.yaml` with the token

- **Authentication Disabled (global connections)**:
  - All connections are stored in the config file
  - All users share the same set of saved connections
  - Suitable for single-user or trusted environments

- **Connection String Security**:
  - Connection strings are stored in plain text in YAML files
  - Use appropriate file permissions (0600 for tokens, 0644 for config)
  - Connection passwords are masked in tool outputs
  - Never commit config files with real credentials to version control
  - Consider using connection strings with IAM authentication instead of passwords

