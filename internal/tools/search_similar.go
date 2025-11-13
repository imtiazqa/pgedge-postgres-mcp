/*-------------------------------------------------------------------------
 *
 * pgEdge Postgres MCP Server - Simplified Semantic Search
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
	"strings"

	"pgedge-postgres-mcp/internal/config"
	"pgedge-postgres-mcp/internal/database"
	"pgedge-postgres-mcp/internal/embedding"
	"pgedge-postgres-mcp/internal/mcp"
)

// SearchSimilarTool creates a simplified semantic search tool that auto-discovers vector tables
func SearchSimilarTool(dbClient *database.Client, cfg *config.Config) Tool {
	return Tool{
		Definition: mcp.Tool{
			Name:        "search_similar",
			Description: "PRIMARY TOOL for semantic search queries. Automatically discovers vector tables, generates embeddings, and returns results. Just provide natural language text - no table/column names needed. Use this instead of generate_embedding + semantic_search.",
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"text_query": map[string]interface{}{
						"type":        "string",
						"description": "Natural language search query",
					},
					"top_k": map[string]interface{}{
						"type":        "integer",
						"description": "Number of results (default: 3)",
						"default":     3,
					},
				},
				Required: []string{"text_query"},
			},
		},
		Handler: func(args map[string]interface{}) (mcp.ToolResponse, error) {
			textQuery, ok := args["text_query"].(string)
			if !ok || textQuery == "" {
				return mcp.NewToolError("'text_query' is required")
			}

			topK := 3
			if topKRaw, ok := args["top_k"]; ok {
				switch v := topKRaw.(type) {
				case float64:
					topK = int(v)
				case int:
					topK = v
				}
			}

			// Check if metadata is loaded
			if !dbClient.IsMetadataLoaded() {
				return mcp.NewToolError(mcp.DatabaseNotReadyError)
			}

			// Step 1: Auto-discover vector tables
			metadata := dbClient.GetMetadata()
			var vectorTables []struct {
				TableName    string
				SchemaName   string
				VectorColumn string
				Dimensions   int
			}

			for _, table := range metadata {
				for _, col := range table.Columns {
					if col.IsVectorColumn {
						vectorTables = append(vectorTables, struct {
							TableName    string
							SchemaName   string
							VectorColumn string
							Dimensions   int
						}{
							TableName:    table.TableName,
							SchemaName:   table.SchemaName,
							VectorColumn: col.ColumnName,
							Dimensions:   col.VectorDimensions,
						})
						break // Only one vector column per table needed
					}
				}
			}

			if len(vectorTables) == 0 {
				return mcp.NewToolError("No tables with vector columns found in database. Cannot perform semantic search.")
			}

			// If multiple tables, use the first one (could be enhanced to search all)
			targetTable := vectorTables[0]

			var sb strings.Builder
			sb.WriteString(fmt.Sprintf("Auto-discovered vector table: %s.%s (column: %s, dimensions: %d)\n\n",
				targetTable.SchemaName, targetTable.TableName, targetTable.VectorColumn, targetTable.Dimensions))

			// Step 2: Generate embedding
			if !cfg.Embedding.Enabled {
				return mcp.NewToolError("Embedding generation not enabled in server configuration")
			}

			embCfg := embedding.Config{
				Provider:        cfg.Embedding.Provider,
				Model:           cfg.Embedding.Model,
				AnthropicAPIKey: cfg.Embedding.AnthropicAPIKey,
				OpenAIAPIKey:    cfg.Embedding.OpenAIAPIKey,
				OllamaURL:       cfg.Embedding.OllamaURL,
			}

			provider, err := embedding.NewProvider(embCfg)
			if err != nil {
				return mcp.NewToolError(fmt.Sprintf("Failed to initialize embedding provider: %v", err))
			}

			ctx := context.Background()
			queryVector, err := provider.Embed(ctx, textQuery)
			if err != nil {
				return mcp.NewToolError(fmt.Sprintf("Failed to generate embedding: %v", err))
			}

			if len(queryVector) != targetTable.Dimensions {
				return mcp.NewToolError(fmt.Sprintf("Embedding dimensions (%d) don't match table vector column dimensions (%d)",
					len(queryVector), targetTable.Dimensions))
			}

			// Step 3: Execute semantic search using the existing semantic_search tool logic
			// Build the search arguments
			searchArgs := map[string]interface{}{
				"table_name":      fmt.Sprintf("%s.%s", targetTable.SchemaName, targetTable.TableName),
				"vector_column":   targetTable.VectorColumn,
				"query_vector":    queryVector,
				"top_k":           topK,
				"distance_metric": "cosine",
			}

			// Get the semantic_search tool and execute it
			semanticSearchTool := SemanticSearchTool(dbClient, cfg)
			result, err := semanticSearchTool.Handler(searchArgs)
			if err != nil {
				return mcp.NewToolError(fmt.Sprintf("Semantic search failed: %v", err))
			}

			// Prepend our discovery info to the result
			if !result.IsError && len(result.Content) > 0 {
				originalText := result.Content[0].Text
				result.Content[0].Text = sb.String() + originalText
			}

			return result, nil
		},
	}
}
