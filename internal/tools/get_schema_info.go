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
			Description: `PRIMARY TOOL for discovering database structure and available tables.

<usecase>
Use get_schema_info when you need to:
- Discover what tables exist in the database
- Understand table structure (columns, types, constraints)
- Find tables with specific capabilities (e.g., vector columns)
- Learn column names before writing queries
- Check data types and nullable constraints
- Understand primary/foreign key relationships
</usecase>

<why_use_this_first>
ALWAYS call this tool FIRST when:
- User asks to query data but doesn't specify table names
- You need to write a SQL query and don't know the schema
- User asks "what data is available?"
- Before using similarity_search (to find vector-enabled tables)
- You're unsure about column names or data types
</why_use_this_first>

<key_features>
Returns comprehensive information:
- All tables and views in the database
- Column names, data types, nullable status
- Primary keys and foreign key relationships
- Table and column descriptions from pg_description
- Vector column detection (pgvector extension)
- Schema organization
</key_features>

<filtering_options>
- No parameters: Returns ALL tables across all schemas (comprehensive)
- schema_name="public": Filter to specific schema only
- vector_tables_only=true: Show only tables with pgvector columns (reduces output 10x, perfect for similarity_search preparation)
</filtering_options>

<examples>
✓ "What tables are available?" → get_schema_info()
✓ "Show me tables with vector columns" → get_schema_info(vector_tables_only=true)
✓ "What's in the public schema?" → get_schema_info(schema_name="public")
✓ Before writing: "SELECT * FROM users..." → get_schema_info() first to confirm 'users' table exists
</examples>

<important>
This tool provides MORE detail than the pg://database-schema resource, which only shows table names and owners. Use this tool for comprehensive schema exploration.
</important>`,
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
