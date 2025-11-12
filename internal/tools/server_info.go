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
	"fmt"

	"pgedge-postgres-mcp/internal/mcp"
)

// ServerInfo contains information about the MCP server
type ServerInfo struct {
	Name    string
	Company string
	Version string
}

// ServerInfoTool creates the server_info tool
func ServerInfoTool(info ServerInfo) Tool {
	return Tool{
		Definition: mcp.Tool{
			Name:        "server_info",
			Description: "Get information about the MCP server itself, including the server name, company, and version.",
			InputSchema: mcp.InputSchema{
				Type:       "object",
				Properties: map[string]interface{}{},
			},
		},
		Handler: func(args map[string]interface{}) (mcp.ToolResponse, error) {
			output := fmt.Sprintf(`Server Information:
===================

Server Name:    %s
Company:        %s
Version:        %s

Description:    An MCP (Model Context Protocol) server that enables AI assistants to interact with PostgreSQL databases through SQL queries and schema exploration.

License:        PostgreSQL License
Copyright:      © 2025, pgEdge, Inc.
`, info.Name, info.Company, info.Version)

			return mcp.NewToolSuccess(output)
		},
	}
}
