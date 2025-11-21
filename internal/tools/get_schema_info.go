/*-------------------------------------------------------------------------
 *
 * pgEdge Postgres MCP Server
 *
 * Portions copyright (c) 2025, pgEdge, Inc.
 * This software is released under The PostgreSQL License
 *
 *-------------------------------------------------------------------------
 */

package tools

import (
	"fmt"
	"strings"

	"pgedge-postgres-mcp/internal/database"
	"pgedge-postgres-mcp/internal/mcp"
)

// GetSchemaInfoTool creates the get_schema_info tool
func GetSchemaInfoTool(dbClient *database.Client) Tool {
	return Tool{
		Definition: mcp.Tool{
			Name:        "get_schema_info",
			Description: "PRIMARY TOOL for discovering database tables and schema information. Lists all tables, views, columns, data types, constraints (primary/foreign keys), and descriptions from pg_description. ALWAYS use this tool first when you need to know what tables exist in the database. Optional parameters: schema_name to filter by schema, vector_tables_only=true to reduce output for semantic search workflows.",
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"schema_name": map[string]interface{}{
						"type":        "string",
						"description": "Optional: specific schema name to get info for. If not provided, returns all schemas.",
					},
					"vector_tables_only": map[string]interface{}{
						"type":        "boolean",
						"description": "Optional: if true, only return tables with vector columns (for semantic search). Reduces output significantly.",
						"default":     false,
					},
				},
			},
		},
		Handler: func(args map[string]interface{}) (mcp.ToolResponse, error) {
			schemaName, ok := args["schema_name"].(string)
			if !ok {
				schemaName = "" // Default to empty string (all schemas)
			}

			vectorTablesOnly := false
			if vectorOnly, ok := args["vector_tables_only"].(bool); ok {
				vectorTablesOnly = vectorOnly
			}

			// Check if metadata is loaded
			if !dbClient.IsMetadataLoaded() {
				return mcp.NewToolError(mcp.DatabaseNotReadyError)
			}

			var sb strings.Builder
			sb.WriteString("Database Schema Information:\n")
			sb.WriteString("============================\n")

			metadata := dbClient.GetMetadata()
			for _, table := range metadata {
				// Filter by schema if requested
				if schemaName != "" && table.SchemaName != schemaName {
					continue
				}

				// Filter for vector tables only if requested
				if vectorTablesOnly {
					hasVectorColumn := false
					for _, col := range table.Columns {
						if col.IsVectorColumn {
							hasVectorColumn = true
							break
						}
					}
					if !hasVectorColumn {
						continue
					}
				}

				sb.WriteString(fmt.Sprintf("\n%s.%s (%s)\n", table.SchemaName, table.TableName, table.TableType))
				if table.Description != "" {
					sb.WriteString(fmt.Sprintf("  Description: %s\n", table.Description))
				}

				sb.WriteString("  Columns:\n")
				for _, col := range table.Columns {
					sb.WriteString(fmt.Sprintf("    - %s: %s", col.ColumnName, col.DataType))
					if col.IsNullable == "YES" {
						sb.WriteString(" (nullable)")
					}
					if col.Description != "" {
						sb.WriteString(fmt.Sprintf("\n      Description: %s", col.Description))
					}
					sb.WriteString("\n")
				}
			}

			return mcp.NewToolSuccess(sb.String())
		},
	}
}
