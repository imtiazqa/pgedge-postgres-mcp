/*-------------------------------------------------------------------------
 *
 * pgEdge Postgres MCP Server - Consolidated Connection Management Tool
 *
 * Copyright (c) 2025, pgEdge, Inc.
 * This software is released under The PostgreSQL License
 *
 *-------------------------------------------------------------------------
 */

package tools

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"pgedge-postgres-mcp/internal/auth"
	"pgedge-postgres-mcp/internal/database"
	"pgedge-postgres-mcp/internal/mcp"
)

// ManageConnectionsTool creates a consolidated tool for all connection management operations
func ManageConnectionsTool(clientManager *database.ClientManager, connMgr *ConnectionManager, configPath string) Tool {
	return Tool{
		Definition: mcp.Tool{
			Name:        "manage_connections",
			Description: "Manage database connections. Operations: connect (set active), add (save new), edit (update), remove (delete), list (show all).",
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"operation": map[string]interface{}{
						"type":        "string",
						"enum":        []string{"connect", "add", "edit", "remove", "list"},
						"description": "Operation type",
					},
					// For "connect" operation
					"connection_string": map[string]interface{}{
						"type":        "string",
						"description": "Connection string/alias (for connect)",
					},
					// For "add"/"edit"/"remove" operations
					"alias": map[string]interface{}{
						"type":        "string",
						"description": "Connection alias",
					},
					// Connection parameters for add/edit
					"host": map[string]interface{}{
						"type":        "string",
						"description": "Hostname/IP",
					},
					"port": map[string]interface{}{
						"type":        "number",
						"description": "Port",
					},
					"user": map[string]interface{}{
						"type":        "string",
						"description": "Username",
					},
					"password": map[string]interface{}{
						"type":        "string",
						"description": "Password",
					},
					"dbname": map[string]interface{}{
						"type":        "string",
						"description": "Database",
					},
					"sslmode": map[string]interface{}{
						"type":        "string",
						"description": "SSL mode",
					},
					"sslcert": map[string]interface{}{
						"type":        "string",
						"description": "SSL cert path",
					},
					"sslkey": map[string]interface{}{
						"type":        "string",
						"description": "SSL key path",
					},
					"sslrootcert": map[string]interface{}{
						"type":        "string",
						"description": "Root CA path",
					},
					"sslpassword": map[string]interface{}{
						"type":        "string",
						"description": "SSL key password",
					},
					"connect_timeout": map[string]interface{}{
						"type":        "number",
						"description": "Timeout (sec)",
					},
					"application_name": map[string]interface{}{
						"type":        "string",
						"description": "App name",
					},
					"description": map[string]interface{}{
						"type":        "string",
						"description": "Notes",
					},
				},
				Required: []string{"operation"},
			},
		},
		Handler: func(args map[string]interface{}) (mcp.ToolResponse, error) {
			operation, ok := args["operation"].(string)
			if !ok || operation == "" {
				return mcp.NewToolError("'operation' parameter is required")
			}

			switch operation {
			case "connect":
				return handleConnect(args, clientManager, connMgr, configPath)
			case "add":
				return handleAdd(args, connMgr, configPath)
			case "edit":
				return handleEdit(args, connMgr, configPath)
			case "remove":
				return handleRemove(args, connMgr, configPath)
			case "list":
				return handleList(connMgr)
			default:
				return mcp.NewToolError(fmt.Sprintf("Unknown operation: %s. Valid operations: connect, add, edit, remove, list", operation))
			}
		},
	}
}

// handleConnect sets the active database connection
func handleConnect(args map[string]interface{}, clientManager *database.ClientManager, connMgr *ConnectionManager, configPath string) (mcp.ToolResponse, error) {
	connStrOrAlias, ok := args["connection_string"].(string)
	if !ok {
		return mcp.NewToolError("'connection_string' is required for connect operation")
	}

	// This is the same logic as SetDatabaseConnectionTool - try merge, then attempt connection
	ctx := context.Background()
	mergedConnStr, matchedAlias := tryMergeSavedConnection(ctx, connStrOrAlias, connMgr, configPath)

	// Create a new client with the connection string
	client := database.NewClientWithConnectionString(mergedConnStr)

	// Test the connection
	if err := client.Connect(); err != nil {
		return mcp.NewToolError(fmt.Sprintf("Failed to connect to database: %v", err))
	}

	// Load metadata
	if err := client.LoadMetadata(); err != nil {
		client.Close()
		return mcp.NewToolError(fmt.Sprintf("Failed to load database metadata: %v", err))
	}

	// Set as the default client for this session
	if err := clientManager.SetClient("default", client); err != nil {
		client.Close()
		return mcp.NewToolError(fmt.Sprintf("Failed to set database connection: %v", err))
	}

	var message string
	if matchedAlias != "" {
		message = fmt.Sprintf("Connected to saved connection '%s'", matchedAlias)
	} else {
		message = "Connected successfully"
	}

	metadata := client.GetMetadata()
	message = fmt.Sprintf("%s. Loaded metadata for %d tables/views.", message, len(metadata))

	return mcp.NewToolSuccess(message)
}

// handleAdd saves a new connection
func handleAdd(args map[string]interface{}, connMgr *ConnectionManager, configPath string) (mcp.ToolResponse, error) {
	// Parse required parameters
	alias, ok := args["alias"].(string)
	if !ok || alias == "" {
		return mcp.NewToolError("'alias' is required for add operation")
	}

	host, ok := args["host"].(string)
	if !ok || host == "" {
		return mcp.NewToolError("'host' is required for add operation")
	}

	user, ok := args["user"].(string)
	if !ok || user == "" {
		return mcp.NewToolError("'user' is required for add operation")
	}

	// Parse optional parameters
	password, _ := args["password"].(string)      //nolint:errcheck // Optional
	dbname, _ := args["dbname"].(string)          //nolint:errcheck // Optional
	sslmode, _ := args["sslmode"].(string)        //nolint:errcheck // Optional
	sslcert, _ := args["sslcert"].(string)        //nolint:errcheck // Optional
	sslkey, _ := args["sslkey"].(string)          //nolint:errcheck // Optional
	sslrootcert, _ := args["sslrootcert"].(string) //nolint:errcheck // Optional
	sslpassword, _ := args["sslpassword"].(string) //nolint:errcheck // Optional
	appname, _ := args["application_name"].(string) //nolint:errcheck // Optional
	desc, _ := args["description"].(string)        //nolint:errcheck // Optional

	port := 5432
	if portRaw, ok := args["port"]; ok {
		switch v := portRaw.(type) {
		case float64:
			port = int(v)
		case int:
			port = v
		}
	}

	connectTimeout := 0
	if timeoutRaw, ok := args["connect_timeout"]; ok {
		switch v := timeoutRaw.(type) {
		case float64:
			connectTimeout = int(v)
		case int:
			connectTimeout = v
		}
	}

	// Get connection store
	ctx := context.Background()
	store, err := connMgr.GetConnectionStore(ctx)
	if err != nil {
		return mcp.NewToolError(fmt.Sprintf("Error: %v", err))
	}

	// Check if alias already exists
	if _, exists := store.Connections[alias]; exists {
		return mcp.NewToolError(fmt.Sprintf("Connection '%s' already exists. Use edit operation to modify it", alias))
	}

	// Create connection
	conn := &auth.SavedConnection{
		Alias:          alias,
		Host:           host,
		Port:           port,
		User:           user,
		DBName:         dbname,
		SSLMode:        sslmode,
		SSLCert:        sslcert,
		SSLKey:         sslkey,
		SSLRootCert:    sslrootcert,
		ConnectTimeout: connectTimeout,
		ApplicationName: appname,
		Description:    desc,
	}

	// Encrypt passwords
	if password != "" {
		encryptedPassword, err := connMgr.encryptionKey.Encrypt(password)
		if err != nil {
			return mcp.NewToolError(fmt.Sprintf("Error encrypting password: %v", err))
		}
		conn.Password = encryptedPassword
	}

	if sslpassword != "" {
		encryptedSSLPassword, err := connMgr.encryptionKey.Encrypt(sslpassword)
		if err != nil {
			return mcp.NewToolError(fmt.Sprintf("Error encrypting SSL password: %v", err))
		}
		conn.SSLPassword = encryptedSSLPassword
	}

	// Add connection
	if err := store.Add(conn); err != nil {
		return mcp.NewToolError(fmt.Sprintf("Error adding connection: %v", err))
	}

	// Save changes
	if err := connMgr.saveChanges(configPath); err != nil {
		return mcp.NewToolError(fmt.Sprintf("Error saving connection: %v", err))
	}

	return mcp.NewToolSuccess(fmt.Sprintf("Connection '%s' saved successfully", alias))
}

// handleEdit updates an existing connection
func handleEdit(args map[string]interface{}, connMgr *ConnectionManager, configPath string) (mcp.ToolResponse, error) {
	alias, ok := args["alias"].(string)
	if !ok || alias == "" {
		return mcp.NewToolError("'alias' is required for edit operation")
	}

	// Get connection store
	ctx := context.Background()
	store, err := connMgr.GetConnectionStore(ctx)
	if err != nil {
		return mcp.NewToolError(fmt.Sprintf("Error: %v", err))
	}

	// Check if connection exists
	existing, exists := store.Connections[alias]
	if !exists {
		return mcp.NewToolError(fmt.Sprintf("Connection '%s' not found", alias))
	}

	// Build updates from provided fields
	hasUpdates := false

	if host, ok := args["host"].(string); ok && host != "" {
		existing.Host = host
		hasUpdates = true
	}

	if portRaw, ok := args["port"]; ok {
		switch v := portRaw.(type) {
		case float64:
			existing.Port = int(v)
		case int:
			existing.Port = v
		}
		hasUpdates = true
	}

	if user, ok := args["user"].(string); ok && user != "" {
		existing.User = user
		hasUpdates = true
	}

	if password, ok := args["password"].(string); ok && password != "" {
		encryptedPassword, err := connMgr.encryptionKey.Encrypt(password)
		if err != nil {
			return mcp.NewToolError(fmt.Sprintf("Error encrypting password: %v", err))
		}
		existing.Password = encryptedPassword
		hasUpdates = true
	}

	if dbname, ok := args["dbname"].(string); ok {
		existing.DBName = dbname
		hasUpdates = true
	}

	if sslmode, ok := args["sslmode"].(string); ok {
		existing.SSLMode = sslmode
		hasUpdates = true
	}

	if sslcert, ok := args["sslcert"].(string); ok {
		existing.SSLCert = sslcert
		hasUpdates = true
	}

	if sslkey, ok := args["sslkey"].(string); ok {
		existing.SSLKey = sslkey
		hasUpdates = true
	}

	if sslrootcert, ok := args["sslrootcert"].(string); ok {
		existing.SSLRootCert = sslrootcert
		hasUpdates = true
	}

	if sslpassword, ok := args["sslpassword"].(string); ok && sslpassword != "" {
		encryptedSSLPassword, err := connMgr.encryptionKey.Encrypt(sslpassword)
		if err != nil {
			return mcp.NewToolError(fmt.Sprintf("Error encrypting SSL password: %v", err))
		}
		existing.SSLPassword = encryptedSSLPassword
		hasUpdates = true
	}

	if timeoutRaw, ok := args["connect_timeout"]; ok {
		switch v := timeoutRaw.(type) {
		case float64:
			existing.ConnectTimeout = int(v)
		case int:
			existing.ConnectTimeout = v
		}
		hasUpdates = true
	}

	if appname, ok := args["application_name"].(string); ok {
		existing.ApplicationName = appname
		hasUpdates = true
	}

	if desc, ok := args["description"].(string); ok {
		existing.Description = desc
		hasUpdates = true
	}

	if !hasUpdates {
		return mcp.NewToolError("No fields provided to update")
	}

	// Update connection
	if err := store.Update(alias, existing); err != nil {
		return mcp.NewToolError(fmt.Sprintf("Error updating connection: %v", err))
	}

	// Save changes
	if err := connMgr.saveChanges(configPath); err != nil {
		return mcp.NewToolError(fmt.Sprintf("Error saving changes: %v", err))
	}

	return mcp.NewToolSuccess(fmt.Sprintf("Connection '%s' updated successfully", alias))
}

// handleRemove deletes a saved connection
func handleRemove(args map[string]interface{}, connMgr *ConnectionManager, configPath string) (mcp.ToolResponse, error) {
	alias, ok := args["alias"].(string)
	if !ok || alias == "" {
		return mcp.NewToolError("'alias' is required for remove operation")
	}

	// Get connection store
	ctx := context.Background()
	store, err := connMgr.GetConnectionStore(ctx)
	if err != nil {
		return mcp.NewToolError(fmt.Sprintf("Error: %v", err))
	}

	// Check if connection exists
	if _, exists := store.Connections[alias]; !exists {
		return mcp.NewToolError(fmt.Sprintf("Connection '%s' not found", alias))
	}

	// Remove connection
	if err := store.Remove(alias); err != nil {
		return mcp.NewToolError(fmt.Sprintf("Error removing connection: %v", err))
	}

	// Save changes
	if err := connMgr.saveChanges(configPath); err != nil {
		return mcp.NewToolError(fmt.Sprintf("Error saving changes: %v", err))
	}

	return mcp.NewToolSuccess(fmt.Sprintf("Connection '%s' removed successfully", alias))
}

// handleList shows all saved connections
func handleList(connMgr *ConnectionManager) (mcp.ToolResponse, error) {
	// Get connection store
	ctx := context.Background()
	store, err := connMgr.GetConnectionStore(ctx)
	if err != nil {
		return mcp.NewToolError(fmt.Sprintf("Error: %v", err))
	}

	connections := store.List()
	if len(connections) == 0 {
		return mcp.NewToolSuccess("No saved connections found.\n\nUse manage_connections with operation='add' to save a connection.")
	}

	// Sort by alias
	sort.Slice(connections, func(i, j int) bool {
		return connections[i].Alias < connections[j].Alias
	})

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Saved Database Connections (%d total):\n", len(connections)))
	result.WriteString("==========================================\n\n")

	for _, conn := range connections {
		result.WriteString(fmt.Sprintf("Alias: %s\n", conn.Alias))
		result.WriteString(fmt.Sprintf("  Host: %s", conn.Host))
		if conn.Port != 0 && conn.Port != 5432 {
			result.WriteString(fmt.Sprintf(":%d", conn.Port))
		}
		result.WriteString("\n")
		result.WriteString(fmt.Sprintf("  User: %s\n", conn.User))
		if conn.DBName != "" {
			result.WriteString(fmt.Sprintf("  Database: %s\n", conn.DBName))
		}
		if conn.SSLMode != "" {
			result.WriteString(fmt.Sprintf("  SSL Mode: %s\n", conn.SSLMode))
		}
		if conn.SSLCert != "" {
			result.WriteString(fmt.Sprintf("  SSL Cert: %s\n", conn.SSLCert))
		}
		if conn.Description != "" {
			result.WriteString(fmt.Sprintf("  Description: %s\n", conn.Description))
		}
		result.WriteString(fmt.Sprintf("  Created: %s\n", conn.CreatedAt.Format("2006-01-02 15:04:05")))
		if !conn.LastUsedAt.IsZero() {
			result.WriteString(fmt.Sprintf("  Last Used: %s\n", conn.LastUsedAt.Format("2006-01-02 15:04:05")))
		}
		result.WriteString("\n")
	}

	return mcp.NewToolSuccess(result.String())
}
