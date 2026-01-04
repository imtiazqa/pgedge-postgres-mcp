# Configuration-Based Test Values

All test files have been updated to use configuration values instead of hardcoded values.

## Configuration Structure

Configuration is loaded from `config/container.yaml` and accessed via `s.Config` in tests.

### Binary Paths
```go
s.Config.Binaries.MCPServer  // /usr/bin/pgedge-postgres-mcp
s.Config.Binaries.KBBuilder  // /usr/bin/pgedge-nla-kb-builder
s.Config.Binaries.CLI        // /usr/bin/pgedge-nla-cli
```

### Configuration Directory
```go
s.Config.ConfigDir  // /etc/pgedge
```

### Package Names
```go
s.Config.Packages.MCPServer  // ["pgedge-postgres-mcp"]
s.Config.Packages.CLI        // ["pgedge-nla-cli"]
s.Config.Packages.Web        // ["pgedge-nla-web"]
s.Config.Packages.KB         // ["pgedge-postgres-mcp-kb"]
```

### Database Configuration
```go
s.Config.Database.Host      // localhost
s.Config.Database.Port      // 5432
s.Config.Database.User      // postgres
s.Config.Database.Password  // postgres123
s.Config.Database.Database  // postgres
```

## Updated Test Files

All 14 test files now use configuration values:

1. ✅ **installation_test.go** - Uses binaries, config_dir, database
2. ✅ **files_test.go** - Uses binaries, config_dir
3. ✅ **postgresql_test.go** - Uses database config
4. ✅ **repository_test.go** - Uses packages config
5. ✅ **user_test.go** - Uses binaries, config_dir
6. ✅ **token_test.go** - Uses binaries, config_dir
7. ✅ **mcp_server_test.go** - Uses binaries, config_dir, packages
8. ✅ **service_test.go** - Uses binaries
9. ✅ **kb_test.go** - Uses binaries
10. ✅ **stdio_test.go** - Uses binaries, database
11. ✅ **mcp_kb_test.go** - Uses binaries, config_dir
12. ✅ **example_test.go** - Uses binaries
13. ✅ **helpers_test.go** - No hardcoded values
14. ✅ **suite_test.go** - No hardcoded values

## Benefits

1. **Flexibility**: Change paths/packages in one place (container.yaml)
2. **Environment Support**: Different configs for different environments
3. **Maintainability**: No scattered hardcoded values
4. **Testability**: Easy to test with different configurations

## Example Usage

```go
// OLD (hardcoded):
s.AssertFileExists("/usr/bin/pgedge-postgres-mcp")

// NEW (config-based):
s.AssertFileExists(s.Config.Binaries.MCPServer)
```

```go
// OLD (hardcoded):
cmd := "PGPASSWORD=postgres123 psql -U postgres -d postgres -c 'SELECT version();'"

// NEW (config-based):
cmd := fmt.Sprintf("PGPASSWORD=%s psql -U %s -d %s -c 'SELECT version();'",
    s.Config.Database.Password,
    s.Config.Database.User,
    s.Config.Database.Database)
```

## Updating Configuration

To change any value, edit `config/container.yaml`:

```yaml
binaries:
  mcp_server: /usr/local/bin/pgedge-postgres-mcp  # Changed path

packages:
  mcp_server:
    - new-package-name  # Changed package name

database:
  password: new_password123  # Changed password
```

All tests will automatically use the new values.
