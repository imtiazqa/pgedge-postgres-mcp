/*-------------------------------------------------------------------------
 *
 * pgEdge Postgres MCP Server
 *
 * Copyright (c) 2025, pgEdge, Inc.
 * This software is released under The PostgreSQL License
 *
 *-------------------------------------------------------------------------
 */

package resources

import (
	"context"
	"fmt"

	"pgedge-postgres-mcp/internal/auth"
	"pgedge-postgres-mcp/internal/database"
	"pgedge-postgres-mcp/internal/mcp"
)

// ContextAwareRegistry wraps a resource registry and provides per-token database clients
// This ensures connection isolation in HTTP/HTTPS mode with authentication
type ContextAwareRegistry struct {
	clientManager *database.ClientManager
	authEnabled   bool
}

// NewContextAwareRegistry creates a new context-aware resource registry
func NewContextAwareRegistry(clientManager *database.ClientManager, authEnabled bool) *ContextAwareRegistry {
	return &ContextAwareRegistry{
		clientManager: clientManager,
		authEnabled:   authEnabled,
	}
}

// List returns all available resource definitions
func (r *ContextAwareRegistry) List() []mcp.Resource {
	// Return static list of all resources
	return []mcp.Resource{
		{
			URI:         URISystemInfo,
			Name:        "PostgreSQL System Information",
			Description: "Returns PostgreSQL version, operating system, and build architecture information. Provides a quick way to check server version and platform details.",
			MimeType:    "application/json",
		},
		{
			URI:         URIStatActivity,
			Name:        "PostgreSQL Current Activity",
			Description: "Shows information about currently executing queries and connections. Useful for monitoring active sessions, identifying long-running queries, and understanding current database load. Each row represents one server process with details about its current activity.",
			MimeType:    "application/json",
		},
		{
			URI:         URIStatReplication,
			Name:        "PostgreSQL Replication Status",
			Description: "Shows the status of replication connections from this primary server including WAL sender processes, replication lag, and sync state. Empty if the server is not a replication primary or has no active replicas. Critical for monitoring replication health and identifying lag issues.",
			MimeType:    "application/json",
		},
	}
}

// Read retrieves a resource by URI with the appropriate database client
func (r *ContextAwareRegistry) Read(ctx context.Context, uri string) (mcp.ResourceContent, error) {
	// Get the appropriate database client for this request
	dbClient, err := r.getClient(ctx)
	if err != nil {
		return mcp.ResourceContent{
			URI: uri,
			Contents: []mcp.ContentItem{
				{
					Type: "text",
					Text: fmt.Sprintf("Failed to get database client: %v\nPlease call set_database_connection first to configure the database connection.", err),
				},
			},
		}, nil
	}

	// Create resource handler with the correct client
	var resource Resource
	switch uri {
	case URISystemInfo:
		resource = PGSystemInfoResource(dbClient)
	case URIStatActivity:
		resource = PGStatActivityResource(dbClient)
	case URIStatReplication:
		resource = PGStatReplicationResource(dbClient)
	default:
		return mcp.ResourceContent{
			URI: uri,
			Contents: []mcp.ContentItem{
				{
					Type: "text",
					Text: "Resource not found: " + uri,
				},
			},
		}, nil
	}

	return resource.Handler()
}

// getClient returns the appropriate database client based on authentication state
func (r *ContextAwareRegistry) getClient(ctx context.Context) (*database.Client, error) {
	if !r.authEnabled {
		// Authentication disabled - use "default" key in ClientManager
		client, err := r.clientManager.GetOrCreateClient("default", false)
		if err != nil {
			return nil, fmt.Errorf("no database connection configured: %w", err)
		}
		return client, nil
	}

	// Authentication enabled - get per-token client
	tokenHash := auth.GetTokenHashFromContext(ctx)
	if tokenHash == "" {
		return nil, fmt.Errorf("no authentication token found in request context")
	}

	// Get or create client for this token (don't auto-connect)
	client, err := r.clientManager.GetOrCreateClient(tokenHash, false)
	if err != nil {
		return nil, fmt.Errorf("no database connection configured for this token: %w", err)
	}

	return client, nil
}
