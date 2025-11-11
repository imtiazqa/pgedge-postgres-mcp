# Connection Migration Tool

This tool migrates saved database connections from the old connection string format to the new individual parameter format with encrypted passwords.

## Purpose

The pgEdge MCP Server now stores database connections using individual parameters (host, port, user, password, etc.) instead of full connection strings. This provides:

- Better security through encrypted password storage
- Support for all PostgreSQL connection parameters including SSL/TLS settings
- Easier connection management and editing

## Usage

```bash
# Build the migration tool (if not already built)
go build -o bin/migrate-connections ./cmd/migrate-connections

# Run the migration
./bin/migrate-connections \
  -prefs <path-to-preferences-file> \
  -secret <path-to-secret-file> \
  [-output <output-file>]
```

### Arguments

- `-prefs` (required): Path to the preferences file to migrate
- `-secret` (required): Path to the encryption secret file
- `-output` (optional): Output file path (defaults to input file with `.migrated` extension)

### Example

```bash
./bin/migrate-connections \
  -prefs bin/pgedge-postgres-mcp-prefs.yaml \
  -secret bin/pgedge-postgres-mcp.secret
```

This will create `bin/pgedge-postgres-mcp-prefs.yaml.migrated` with the converted connections.

## What It Does

The migration tool:

1. Reads the old preferences file with connection strings
2. Parses each connection string into individual parameters:
   - User, password, host, port, database name
   - SSL/TLS parameters (sslmode, sslcert, sslkey, sslrootcert, etc.)
   - Connection parameters (connect_timeout, application_name)
3. Encrypts all passwords using the encryption key from the secret file
4. Writes the new format to the output file
5. Preserves all metadata (aliases, descriptions, timestamps)

## After Migration

Once you've reviewed the migrated file and verified it's correct:

```bash
# Backup the original
mv bin/pgedge-postgres-mcp-prefs.yaml bin/pgedge-postgres-mcp-prefs.yaml.backup

# Use the migrated version
mv bin/pgedge-postgres-mcp-prefs.yaml.migrated bin/pgedge-postgres-mcp-prefs.yaml
```

## Example Migration

**Before (old format):**
```yaml
connections:
    connections:
        production:
            alias: production
            connection_string: postgres://user:pass@host.example.com:5432/mydb?sslmode=require
            maintenance_db: postgres
            description: Production database
```

**After (new format):**
```yaml
connections:
    connections:
        production:
            alias: production
            description: Production database
            host: host.example.com
            port: 5432
            user: user
            password: "zEcYRaNr7f7yVwpjWDLS7o2ud1Td..."  # Encrypted
            dbname: mydb
            maintenance_db: postgres
            sslmode: require
```

## Notes

- The tool requires the encryption secret file to encrypt passwords correctly
- Passwords are encrypted using AES-256-GCM with the server's encryption key
- Empty passwords (for trust-based authentication) are handled correctly
- All SSL/TLS parameters from the connection string are preserved
- The original file is not modified - review the output before replacing it
