/*-------------------------------------------------------------------------
 *
 * pgEdge Postgres MCP Server
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

	"pgedge-postgres-mcp/internal/auth"
	"pgedge-postgres-mcp/internal/database"
	"pgedge-postgres-mcp/internal/llm"
	"pgedge-postgres-mcp/internal/mcp"
	"pgedge-postgres-mcp/internal/resources"
)

// ContextAwareProvider wraps a tool registry and provides per-token database clients
// This ensures connection isolation in HTTP/HTTPS mode with authentication
type ContextAwareProvider struct {
	registry       *Registry
	clientManager  *database.ClientManager
	llmClient      *llm.Client
	resourceReg    *resources.Registry
	authEnabled    bool
	fallbackClient *database.Client // Used when auth is disabled
}

// NewContextAwareProvider creates a new context-aware tool provider
func NewContextAwareProvider(clientManager *database.ClientManager, llmClient *llm.Client, resourceReg *resources.Registry, authEnabled bool, fallbackClient *database.Client) *ContextAwareProvider {
	return &ContextAwareProvider{
		registry:       NewRegistry(),
		clientManager:  clientManager,
		llmClient:      llmClient,
		resourceReg:    resourceReg,
		authEnabled:    authEnabled,
		fallbackClient: fallbackClient,
	}
}

// RegisterTools registers all tools with their per-token client handlers
func (p *ContextAwareProvider) RegisterTools(ctx context.Context) error {
	// Get the appropriate database client
	dbClient, err := p.getClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to get database client: %w", err)
	}

	// Register tools with the per-token client
	p.registry.Register("query_database", QueryDatabaseTool(dbClient, p.llmClient))
	p.registry.Register("get_schema_info", GetSchemaInfoTool(dbClient))
	p.registry.Register("set_pg_configuration", SetPGConfigurationTool(dbClient))
	p.registry.Register("recommend_pg_configuration", RecommendPGConfigurationTool())
	p.registry.Register("analyze_bloat", AnalyzeBloatTool(dbClient))
	p.registry.Register("read_server_log", ReadServerLogTool(dbClient))
	p.registry.Register("read_postgresql_conf", ReadPostgresqlConfTool(dbClient))
	p.registry.Register("read_pg_hba_conf", ReadPgHbaConfTool(dbClient))
	p.registry.Register("read_pg_ident_conf", ReadPgIdentConfTool(dbClient))
	p.registry.Register("read_resource", ReadResourceTool(p.resourceReg))

	return nil
}

// List returns all registered tool definitions
func (p *ContextAwareProvider) List() []mcp.Tool {
	return p.registry.List()
}

// Execute runs a tool by name with the given arguments and context
func (p *ContextAwareProvider) Execute(ctx context.Context, name string, args map[string]interface{}) (mcp.ToolResponse, error) {
	// Get the appropriate database client for this request
	dbClient, err := p.getClient(ctx)
	if err != nil {
		return mcp.ToolResponse{
			Content: []mcp.ContentItem{
				{
					Type: "text",
					Text: fmt.Sprintf("Failed to get database client: %v", err),
				},
			},
			IsError: true,
		}, fmt.Errorf("failed to get database client: %w", err)
	}

	// Re-register tools with the per-token client for this specific execution
	// This ensures each request uses the correct client
	p.registry.Register("query_database", QueryDatabaseTool(dbClient, p.llmClient))
	p.registry.Register("get_schema_info", GetSchemaInfoTool(dbClient))
	p.registry.Register("set_pg_configuration", SetPGConfigurationTool(dbClient))
	p.registry.Register("analyze_bloat", AnalyzeBloatTool(dbClient))
	p.registry.Register("read_server_log", ReadServerLogTool(dbClient))
	p.registry.Register("read_postgresql_conf", ReadPostgresqlConfTool(dbClient))
	p.registry.Register("read_pg_hba_conf", ReadPgHbaConfTool(dbClient))
	p.registry.Register("read_pg_ident_conf", ReadPgIdentConfTool(dbClient))
	p.registry.Register("read_resource", ReadResourceTool(p.resourceReg))

	// Execute the tool using the registry
	return p.registry.Execute(ctx, name, args)
}

// getClient returns the appropriate database client based on authentication state
func (p *ContextAwareProvider) getClient(ctx context.Context) (*database.Client, error) {
	if !p.authEnabled {
		// Authentication disabled - use fallback client (shared)
		return p.fallbackClient, nil
	}

	// Authentication enabled - get per-token client
	tokenHash := auth.GetTokenHashFromContext(ctx)
	if tokenHash == "" {
		return nil, fmt.Errorf("no authentication token found in request context")
	}

	// Get or create client for this token
	client, err := p.clientManager.GetClient(tokenHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get client for token: %w", err)
	}

	return client, nil
}
