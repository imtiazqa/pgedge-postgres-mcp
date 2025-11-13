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
    "encoding/json"
    "fmt"
    "strings"

    "pgedge-postgres-mcp/internal/config"
    "pgedge-postgres-mcp/internal/database"
    "pgedge-postgres-mcp/internal/embedding"
    "pgedge-postgres-mcp/internal/mcp"
)

// SemanticSearchTool creates the semantic_search tool for pgvector similarity search
func SemanticSearchTool(dbClient *database.Client, cfg *config.Config) Tool {
    return Tool{
        Definition: mcp.Tool{
            Name:        "semantic_search",
            Description: "SEMANTIC SEARCH REQUIRES 2 STEPS: (1) MUST call get_schema_info(vector_tables_only=true) to discover tables with vector columns, (2) call this tool with table_name, vector_column, and text_query. The text_query parameter auto-generates embeddings. DO NOT ask user for table/column names - discover them in step 1.",
            InputSchema: mcp.InputSchema{
                Type: "object",
                Properties: map[string]interface{}{
                    "table_name": map[string]interface{}{
                        "type":        "string",
                        "description": "Table name from get_schema_info with vector(N) column",
                    },
                    "vector_column": map[string]interface{}{
                        "type":        "string",
                        "description": "Vector column name from schema (type shows as vector(dimensions))",
                    },
                    "query_vector": map[string]interface{}{
                        "type":        "array",
                        "items":       map[string]interface{}{"type": "number"},
                        "description": "Pre-computed vector",
                    },
                    "text_query": map[string]interface{}{
                        "type":        "string",
                        "description": "User's search text (auto-generates embedding) - PREFERRED over query_vector",
                    },
                    "top_k": map[string]interface{}{
                        "type":        "integer",
                        "description": "Results to return (default: 3)",
                        "default":     3,
                    },
                    "distance_metric": map[string]interface{}{
                        "type":        "string",
                        "description": "Metric: cosine/l2/inner_product",
                        "enum":        []string{"cosine", "l2", "inner_product"},
                        "default":     "cosine",
                    },
                    "filter_conditions": map[string]interface{}{
                        "type":        "string",
                        "description": "SQL WHERE clause",
                    },
                },
                Required: []string{"table_name", "vector_column"},
            },
        },
        Handler: func(args map[string]interface{}) (mcp.ToolResponse, error) {
            // Extract and validate required parameters
            tableName, ok := args["table_name"].(string)
            if !ok || tableName == "" {
                return mcp.NewToolError("Missing or invalid 'table_name' parameter")
            }

            vectorColumn, ok := args["vector_column"].(string)
            if !ok || vectorColumn == "" {
                return mcp.NewToolError("Missing or invalid 'vector_column' parameter")
            }

            // Check for either query_vector or text_query (but not both)
            queryVectorRaw, hasQueryVector := args["query_vector"]
            textQuery, hasTextQuery := args["text_query"].(string)

            if !hasQueryVector && !hasTextQuery {
                return mcp.NewToolError("Either 'query_vector' or 'text_query' parameter must be provided")
            }

            if hasQueryVector && hasTextQuery {
                return mcp.NewToolError("Only one of 'query_vector' or 'text_query' can be provided, not both")
            }

            var queryVector []float64
            var err error

            // Generate embedding from text if text_query is provided
            if hasTextQuery {
                textQuery = strings.TrimSpace(textQuery)
                if textQuery == "" {
                    return mcp.NewToolError("'text_query' parameter cannot be empty or whitespace-only")
                }

                // Check if embedding generation is enabled
                if !cfg.Embedding.Enabled {
                    return mcp.NewToolError("'text_query' parameter requires embedding generation to be enabled. Please enable it in the server configuration or provide 'query_vector' instead.")
                }

                // Create embedding provider from config
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

                // Generate embedding
                ctx := context.Background()
                queryVector, err = provider.Embed(ctx, textQuery)
                if err != nil {
                    return mcp.NewToolError(fmt.Sprintf("Failed to generate embedding from text: %v", err))
                }

                if len(queryVector) == 0 {
                    return mcp.NewToolError("Received empty embedding vector from provider")
                }
            } else {
                // Parse provided query_vector
                queryVector, err = parseQueryVector(queryVectorRaw)
                if err != nil {
                    return mcp.NewToolError(fmt.Sprintf("Invalid 'query_vector' parameter: %v", err))
                }

                if len(queryVector) == 0 {
                    return mcp.NewToolError("'query_vector' cannot be empty")
                }
            }

            // Extract optional parameters
            topK := 3
            if topKRaw, ok := args["top_k"]; ok {
                switch v := topKRaw.(type) {
                case float64:
                    topK = int(v)
                case int:
                    topK = v
                }
            }
            if topK < 1 {
                return mcp.NewToolError("'top_k' must be greater than 0")
            }

            distanceMetric := "cosine"
            if metricRaw, ok := args["distance_metric"]; ok {
                if metric, ok := metricRaw.(string); ok {
                    distanceMetric = metric
                }
            }

            // Validate distance metric
            distanceOp, metricName, err := getDistanceOperator(distanceMetric)
            if err != nil {
                return mcp.NewToolError(err.Error())
            }

            filterConditions := ""
            if filterRaw, ok := args["filter_conditions"]; ok {
                if filter, ok := filterRaw.(string); ok {
                    filterConditions = strings.TrimSpace(filter)
                }
            }

            // Check if metadata is loaded
            connStr := dbClient.GetDefaultConnection()
            if !dbClient.IsMetadataLoadedFor(connStr) {
                return mcp.NewToolError(mcp.DatabaseNotReadyError)
            }

            // Normalize table name (add schema if not present)
            normalizedTableName := tableName
            if !strings.Contains(tableName, ".") {
                normalizedTableName = "public." + tableName
            }

            // Verify table exists in metadata
            metadata := dbClient.GetMetadata()
            tableInfo, exists := metadata[normalizedTableName]
            if !exists {
                return mcp.NewToolError(fmt.Sprintf("Table '%s' not found in database metadata. Available tables: %s",
                    tableName, formatAvailableTables(metadata)))
            }

            // Verify column exists and is a vector column
            var vectorColInfo *database.ColumnInfo
            for i := range tableInfo.Columns {
                if tableInfo.Columns[i].ColumnName == vectorColumn {
                    vectorColInfo = &tableInfo.Columns[i]
                    break
                }
            }

            if vectorColInfo == nil {
                return mcp.NewToolError(fmt.Sprintf("Column '%s' not found in table '%s'. Available columns: %s",
                    vectorColumn, tableName, formatAvailableColumns(tableInfo)))
            }

            if !vectorColInfo.IsVectorColumn {
                return mcp.NewToolError(fmt.Sprintf("Column '%s' in table '%s' is not a pgvector column (type: %s)",
                    vectorColumn, tableName, vectorColInfo.DataType))
            }

            // Verify dimensions match
            if vectorColInfo.VectorDimensions != len(queryVector) {
                return mcp.NewToolError(fmt.Sprintf("Query vector dimensions (%d) don't match column dimensions (%d) for '%s.%s'",
                    len(queryVector), vectorColInfo.VectorDimensions, tableName, vectorColumn))
            }

            // Build the SQL query
            vectorLiteral := formatVectorLiteral(queryVector)

            // Build column list excluding vector columns to reduce token usage
            var selectColumns []string
            for _, col := range tableInfo.Columns {
                if !col.IsVectorColumn {
                    selectColumns = append(selectColumns, col.ColumnName)
                }
            }
            columnList := strings.Join(selectColumns, ", ")
            if columnList == "" {
                // Fallback to * if no non-vector columns (unlikely)
                columnList = "*"
            }

            var whereClause string
            if filterConditions != "" {
                whereClause = fmt.Sprintf("WHERE %s", filterConditions)
            }

            sqlQuery := fmt.Sprintf(
                "SELECT %s, (%s %s '%s'::vector) AS distance FROM %s %s ORDER BY %s %s '%s'::vector LIMIT %d",
                columnList, vectorColumn, distanceOp, vectorLiteral, normalizedTableName, whereClause,
                vectorColumn, distanceOp, vectorLiteral, topK,
            )

            // Execute the query
            ctx := context.Background()
            pool := dbClient.GetPoolFor(connStr)
            if pool == nil {
                return mcp.NewToolError(fmt.Sprintf("Connection pool not found for: %s", connStr))
            }

            // Begin a read-only transaction
            tx, err := pool.Begin(ctx)
            if err != nil {
                return mcp.NewToolError(fmt.Sprintf("Failed to begin transaction: %v", err))
            }

            // Track whether transaction was committed
            committed := false
            defer func() {
                if !committed {
                    // Only rollback if not committed - prevents idle transactions
                    _ = tx.Rollback(ctx) //nolint:errcheck // rollback in defer after commit is expected to fail
                }
            }()

            // Set transaction to read-only
            _, err = tx.Exec(ctx, "SET TRANSACTION READ ONLY")
            if err != nil {
                return mcp.NewToolError(fmt.Sprintf("Failed to set transaction read-only: %v", err))
            }

            rows, err := tx.Query(ctx, sqlQuery)
            if err != nil {
                return mcp.NewToolError(fmt.Sprintf("Error executing semantic search:\n%s\n\nError: %v", sqlQuery, err))
            }
            defer rows.Close()

            // Get column names
            fieldDescriptions := rows.FieldDescriptions()
            var columnNames []string
            for _, fd := range fieldDescriptions {
                columnNames = append(columnNames, string(fd.Name))
            }

            // Collect results
            var results []map[string]interface{}
            for rows.Next() {
                values, err := rows.Values()
                if err != nil {
                    return mcp.NewToolError(fmt.Sprintf("Error reading row: %v", err))
                }

                row := make(map[string]interface{})
                for i, colName := range columnNames {
                    row[colName] = values[i]
                }
                results = append(results, row)
            }

            if err := rows.Err(); err != nil {
                return mcp.NewToolError(fmt.Sprintf("Error iterating rows: %v", err))
            }

            // Commit the read-only transaction
            if err := tx.Commit(ctx); err != nil {
                return mcp.NewToolError(fmt.Sprintf("Failed to commit transaction: %v", err))
            }
            committed = true

            // Format results (compact JSON to reduce token usage)
            resultsJSON, err := json.Marshal(results)
            if err != nil {
                return mcp.NewToolError(fmt.Sprintf("Error formatting results: %v", err))
            }

            var sb strings.Builder
            sb.WriteString("Semantic Search Results:\n")
            sb.WriteString(fmt.Sprintf("Table: %s\n", tableName))
            sb.WriteString(fmt.Sprintf("Vector Column: %s (dimensions: %d)\n", vectorColumn, vectorColInfo.VectorDimensions))
            sb.WriteString(fmt.Sprintf("Distance Metric: %s\n", metricName))
            sb.WriteString(fmt.Sprintf("Top K: %d\n", topK))
            if filterConditions != "" {
                sb.WriteString(fmt.Sprintf("Filter: %s\n", filterConditions))
            }
            sb.WriteString(fmt.Sprintf("\nResults (%d rows):\n%s", len(results), string(resultsJSON)))

            return mcp.NewToolSuccess(sb.String())
        },
    }
}

// parseQueryVector converts various input formats to []float64
func parseQueryVector(raw interface{}) ([]float64, error) {
    switch v := raw.(type) {
    case []interface{}:
        result := make([]float64, len(v))
        for i, val := range v {
            switch num := val.(type) {
            case float64:
                result[i] = num
            case int:
                result[i] = float64(num)
            case int64:
                result[i] = float64(num)
            default:
                return nil, fmt.Errorf("element at index %d is not a number: %v", i, val)
            }
        }
        return result, nil
    case []float64:
        return v, nil
    case []float32:
        result := make([]float64, len(v))
        for i, val := range v {
            result[i] = float64(val)
        }
        return result, nil
    default:
        return nil, fmt.Errorf("expected array of numbers, got %T", raw)
    }
}

// getDistanceOperator returns the SQL operator and human-readable name for a distance metric
func getDistanceOperator(metric string) (operator string, name string, err error) {
    switch strings.ToLower(metric) {
    case "cosine":
        return "<=>", "Cosine Distance", nil
    case "l2", "euclidean":
        return "<->", "L2 (Euclidean) Distance", nil
    case "inner_product", "inner":
        return "<#>", "Inner Product (Negative)", nil
    default:
        return "", "", fmt.Errorf("invalid distance metric '%s'. Valid options: cosine, l2, inner_product", metric)
    }
}

// formatVectorLiteral converts a float slice to a PostgreSQL vector literal
func formatVectorLiteral(vec []float64) string {
    strNums := make([]string, len(vec))
    for i, v := range vec {
        strNums[i] = fmt.Sprintf("%f", v)
    }
    return "[" + strings.Join(strNums, ",") + "]"
}

// formatAvailableTables returns a comma-separated list of available tables
func formatAvailableTables(metadata map[string]database.TableInfo) string {
    var tables []string
    for key := range metadata {
        tables = append(tables, key)
    }
    if len(tables) == 0 {
        return "(none)"
    }
    return strings.Join(tables, ", ")
}

// formatAvailableColumns returns a comma-separated list of available columns
func formatAvailableColumns(tableInfo database.TableInfo) string {
    var columns []string
    for _, col := range tableInfo.Columns {
        colDesc := col.ColumnName
        if col.IsVectorColumn {
            colDesc += fmt.Sprintf(" (vector(%d))", col.VectorDimensions)
        }
        columns = append(columns, colDesc)
    }
    if len(columns) == 0 {
        return "(none)"
    }
    return strings.Join(columns, ", ")
}
