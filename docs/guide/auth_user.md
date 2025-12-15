# Authentication - User Management

User accounts provide interactive authentication with session-based access. Users authenticate with username and password to receive a 24-hour session token.

### When to Use Users vs API Tokens

- **API Tokens**: Direct machine-to-machine access, long-lived, managed by administrators
- **User Accounts**: Interactive applications, session-based, users manage own passwords

### Adding Users

#### Interactive Mode

```bash
# Add user with prompts
./bin/pgedge-postgres-mcp -add-user
```

You'll be prompted for:

- **Username**: Unique username for the account
- **Password**: Password (hidden, with confirmation)
- **Note**: Optional description (e.g., "Alice Smith - Developer")

#### Command Line Mode

```bash
# Add user with all details specified
./bin/pgedge-postgres-mcp -add-user \
  -username alice \
  -password "SecurePassword123!" \
  -user-note "Alice Smith - Developer"
```

### Listing Users

```bash
./bin/pgedge-postgres-mcp -list-users
```

Output:
```
Users:
==========================================================================================
Username             Created                   Last Login           Status      Annotation
------------------------------------------------------------------------------------------
alice                2024-10-30 10:15          2024-11-14 09:30     Enabled     Developer
bob                  2024-10-15 14:20          Never                Enabled     Admin
charlie              2024-09-01 08:00          2024-10-10 16:45     DISABLED    Former emp
==========================================================================================
```

### Updating Users

```bash
# Update password
./bin/pgedge-postgres-mcp -update-user -username alice

# Update with new password from command line (less secure)
./bin/pgedge-postgres-mcp -update-user \
  -username alice \
  -password "NewPassword456!"

# Update annotation only
./bin/pgedge-postgres-mcp -update-user \
  -username alice \
  -user-note "Alice Smith - Senior Developer"
```

### Managing User Status

```bash
# Disable a user account (prevents login)
./bin/pgedge-postgres-mcp -disable-user -username charlie

# Re-enable a user account
./bin/pgedge-postgres-mcp -enable-user -username charlie
```

### Deleting Users

```bash
# Delete user (with confirmation prompt)
./bin/pgedge-postgres-mcp -delete-user -username charlie
```

### Custom User File Location

```bash
# Specify custom user file path
./bin/pgedge-postgres-mcp -user-file /etc/pgedge/pgedge-postgres-mcp-users.yaml -list-users
```

### User Storage

- **Default location**: `pgedge-postgres-mcp-users.yaml` in the same directory as the binary
- **Storage format**: YAML with bcrypt-hashed passwords (cost factor 12)
- **File permissions**: Automatically set to 0600 (owner read/write only)
- **Session tokens**: Generated with crypto/rand (32 bytes, 24-hour validity)
