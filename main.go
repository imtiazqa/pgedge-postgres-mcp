package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	ProtocolVersion = "2024-11-05" // MCP protocol version we support
	ServerName      = "pgedge-mcp"
	ServerVersion   = "0.1.0"
)

// MCP Protocol Types
type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type JSONRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id,omitempty"`
	Result  interface{} `json:"result,omitempty"`
	Error   *RPCError   `json:"error,omitempty"`
}

type RPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type InitializeParams struct {
	ProtocolVersion string                 `json:"protocolVersion"`
	Capabilities    map[string]interface{} `json:"capabilities"`
	ClientInfo      ClientInfo             `json:"clientInfo"`
}

type ClientInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type Implementation struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type InitializeResult struct {
	ProtocolVersion string                 `json:"protocolVersion"`
	Capabilities    map[string]interface{} `json:"capabilities"`
	ServerInfo      Implementation         `json:"serverInfo"`
}

type Tool struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema InputSchema `json:"inputSchema"`
}

type InputSchema struct {
	Type       string                 `json:"type"`
	Properties map[string]interface{} `json:"properties"`
	Required   []string               `json:"required,omitempty"`
}

type ToolCallParams struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments,omitempty"`
}

type ToolResponse struct {
	Content []ContentItem `json:"content"`
	IsError bool          `json:"isError,omitempty"`
}

type ContentItem struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// Database metadata structures
type TableInfo struct {
	SchemaName  string
	TableName   string
	TableType   string // 'TABLE' or 'VIEW'
	Description string
	Columns     []ColumnInfo
}

type ColumnInfo struct {
	ColumnName  string
	DataType    string
	IsNullable  string
	Description string
}

// Server state
type MCPServer struct {
	dbPool         *pgxpool.Pool
	metadata       map[string]TableInfo
	llmClient      *LLMClient
	metadataLoaded bool
	mu             sync.RWMutex // Protects metadataLoaded and metadata
}

func main() {
	server := &MCPServer{
		metadata:       make(map[string]TableInfo),
		llmClient:      NewLLMClient(),
		metadataLoaded: false,
	}

	// Initialize database connection and metadata in background
	// This allows the stdio server to start immediately and respond to client requests
	go func() {
		if err := server.connectDB(); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: Failed to connect to database: %v\n", err)
			return
		}

		if err := server.loadMetadata(); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: Failed to load database metadata: %v\n", err)
			return
		}

		server.mu.Lock()
		server.metadataLoaded = true
		metadataCount := len(server.metadata)
		server.mu.Unlock()

		fmt.Fprintf(os.Stderr, "Database ready: %d tables/views loaded\n", metadataCount)
	}()

	// Start stdio server
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024) // Support large messages

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		var req JSONRPCRequest
		if err := json.Unmarshal([]byte(line), &req); err != nil {
			sendError(nil, -32700, "Parse error", err.Error())
			continue
		}

		server.handleRequest(req)
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(1)
	}

	// Cleanup
	if server.dbPool != nil {
		server.dbPool.Close()
	}
}

func (s *MCPServer) connectDB() error {
	// Get database connection string from environment
	connStr := os.Getenv("POSTGRES_CONNECTION_STRING")
	if connStr == "" {
		connStr = "postgres://localhost/postgres?sslmode=disable"
	}

	var err error
	s.dbPool, err = pgxpool.New(context.Background(), connStr)
	if err != nil {
		return fmt.Errorf("unable to create connection pool: %w", err)
	}

	// Test connection
	if err := s.dbPool.Ping(context.Background()); err != nil {
		return fmt.Errorf("unable to ping database: %w", err)
	}

	return nil
}

func (s *MCPServer) loadMetadata() error {
	ctx := context.Background()

	// Query to get tables, views, and their columns with descriptions
	query := `
		WITH table_comments AS (
			SELECT
				n.nspname AS schema_name,
				c.relname AS table_name,
				CASE c.relkind
					WHEN 'r' THEN 'TABLE'
					WHEN 'v' THEN 'VIEW'
					WHEN 'm' THEN 'MATERIALIZED VIEW'
				END AS table_type,
				obj_description(c.oid) AS table_description
			FROM pg_class c
			JOIN pg_namespace n ON n.oid = c.relnamespace
			WHERE c.relkind IN ('r', 'v', 'm')
				AND n.nspname NOT IN ('pg_catalog', 'information_schema', 'pg_toast')
			ORDER BY n.nspname, c.relname
		),
		column_info AS (
			SELECT
				n.nspname AS schema_name,
				c.relname AS table_name,
				a.attname AS column_name,
				pg_catalog.format_type(a.atttypid, a.atttypmod) AS data_type,
				CASE WHEN a.attnotnull THEN 'NO' ELSE 'YES' END AS is_nullable,
				col_description(c.oid, a.attnum) AS column_description
			FROM pg_class c
			JOIN pg_namespace n ON n.oid = c.relnamespace
			JOIN pg_attribute a ON a.attrelid = c.oid
			WHERE c.relkind IN ('r', 'v', 'm')
				AND n.nspname NOT IN ('pg_catalog', 'information_schema', 'pg_toast')
				AND a.attnum > 0
				AND NOT a.attisdropped
			ORDER BY n.nspname, c.relname, a.attnum
		)
		SELECT
			tc.schema_name,
			tc.table_name,
			tc.table_type,
			COALESCE(tc.table_description, '') AS table_description,
			ci.column_name,
			ci.data_type,
			ci.is_nullable,
			COALESCE(ci.column_description, '') AS column_description
		FROM table_comments tc
		LEFT JOIN column_info ci ON tc.schema_name = ci.schema_name AND tc.table_name = ci.table_name
		ORDER BY tc.schema_name, tc.table_name, ci.column_name
	`

	rows, err := s.dbPool.Query(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to query metadata: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var schemaName, tableName, tableType, tableDesc, columnName, dataType, isNullable, columnDesc string

		err := rows.Scan(&schemaName, &tableName, &tableType, &tableDesc, &columnName, &dataType, &isNullable, &columnDesc)
		if err != nil {
			return fmt.Errorf("failed to scan row: %w", err)
		}

		key := schemaName + "." + tableName

		table, exists := s.metadata[key]
		if !exists {
			table = TableInfo{
				SchemaName:  schemaName,
				TableName:   tableName,
				TableType:   tableType,
				Description: tableDesc,
				Columns:     []ColumnInfo{},
			}
		}

		if columnName != "" {
			table.Columns = append(table.Columns, ColumnInfo{
				ColumnName:  columnName,
				DataType:    dataType,
				IsNullable:  isNullable,
				Description: columnDesc,
			})
		}

		s.metadata[key] = table
	}

	return rows.Err()
}

func (s *MCPServer) handleRequest(req JSONRPCRequest) {
	switch req.Method {
	case "initialize":
		s.handleInitialize(req)
	case "notifications/initialized":
		// Client notification - no response needed
	case "tools/list":
		s.handleToolsList(req)
	case "tools/call":
		s.handleToolCall(req)
	default:
		if req.ID != nil {
			sendError(req.ID, -32601, "Method not found", nil)
		}
	}
}

func (s *MCPServer) handleInitialize(req JSONRPCRequest) {
	var params InitializeParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		sendError(req.ID, -32602, "Invalid params", err.Error())
		return
	}

	// Accept the client's protocol version for compatibility
	protocolVersion := params.ProtocolVersion
	if protocolVersion == "" {
		protocolVersion = ProtocolVersion
	}

	result := InitializeResult{
		ProtocolVersion: protocolVersion,
		Capabilities: map[string]interface{}{
			"tools": map[string]interface{}{},
		},
		ServerInfo: Implementation{
			Name:    ServerName,
			Version: ServerVersion,
		},
	}

	sendResponse(req.ID, result)
}

func (s *MCPServer) handleToolsList(req JSONRPCRequest) {
	tools := []Tool{
		{
			Name:        "query_database",
			Description: "Execute a natural language query against the PostgreSQL database. The system will analyze the database schema (including table names, column names, data types, and comments from pg_description) to understand the structure and convert your natural language query into SQL. Returns the query results.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"query": map[string]interface{}{
						"type":        "string",
						"description": "Natural language question about the data in the database",
					},
				},
				Required: []string{"query"},
			},
		},
		{
			Name:        "get_schema_info",
			Description: "Get detailed schema information about the database, including all tables, views, columns, data types, and descriptions from pg_description. Useful for understanding the database structure before querying.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"schema_name": map[string]interface{}{
						"type":        "string",
						"description": "Optional: specific schema name to get info for. If not provided, returns all schemas.",
					},
				},
			},
		},
	}

	result := map[string]interface{}{
		"tools": tools,
	}

	sendResponse(req.ID, result)
}

func (s *MCPServer) handleToolCall(req JSONRPCRequest) {
	var params ToolCallParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		sendError(req.ID, -32602, "Invalid params", err.Error())
		return
	}

	switch params.Name {
	case "query_database":
		s.handleQueryDatabase(req.ID, params.Arguments)
	case "get_schema_info":
		s.handleGetSchemaInfo(req.ID, params.Arguments)
	default:
		sendError(req.ID, -32602, "Unknown tool", params.Name)
	}
}

func (s *MCPServer) handleQueryDatabase(id interface{}, args map[string]interface{}) {
	query, ok := args["query"].(string)
	if !ok {
		sendError(id, -32602, "Missing or invalid 'query' parameter", nil)
		return
	}

	// Check if metadata is loaded (thread-safe)
	s.mu.RLock()
	loaded := s.metadataLoaded
	s.mu.RUnlock()

	if !loaded {
		response := ToolResponse{
			Content: []ContentItem{
				{
					Type: "text",
					Text: "Database is still initializing. Please wait a moment and try again.\n\n" +
						"The server is loading database metadata in the background. This usually takes a few seconds.",
				},
			},
			IsError: true,
		}
		sendResponse(id, response)
		return
	}

	// Generate schema context for LLM
	schemaContext := s.generateSchemaContext()

	// Check if LLM is configured
	if !s.llmClient.IsConfigured() {
		response := ToolResponse{
			Content: []ContentItem{
				{
					Type: "text",
					Text: fmt.Sprintf("Natural language query: %s\n\n"+
						"Database Schema Context:\n%s\n\n"+
						"ERROR: ANTHROPIC_API_KEY environment variable is not set.\n\n"+
						"To enable natural language to SQL conversion, please set the ANTHROPIC_API_KEY environment variable.\n"+
						"You can optionally set ANTHROPIC_MODEL to specify a different model (default: claude-sonnet-4-5).",
						query, schemaContext),
				},
			},
			IsError: true,
		}
		sendResponse(id, response)
		return
	}

	// Convert natural language to SQL using Claude
	sqlQuery, err := s.llmClient.ConvertNLToSQL(query, schemaContext)
	if err != nil {
		response := ToolResponse{
			Content: []ContentItem{
				{
					Type: "text",
					Text: fmt.Sprintf("Failed to convert natural language to SQL: %v", err),
				},
			},
			IsError: true,
		}
		sendResponse(id, response)
		return
	}

	// Execute the generated SQL query
	ctx := context.Background()
	rows, err := s.dbPool.Query(ctx, sqlQuery)
	if err != nil {
		response := ToolResponse{
			Content: []ContentItem{
				{
					Type: "text",
					Text: fmt.Sprintf("Generated SQL:\n%s\n\nError executing query: %v", sqlQuery, err),
				},
			},
			IsError: true,
		}
		sendResponse(id, response)
		return
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
			response := ToolResponse{
				Content: []ContentItem{
					{
						Type: "text",
						Text: fmt.Sprintf("Error reading row: %v", err),
					},
				},
				IsError: true,
			}
			sendResponse(id, response)
			return
		}

		row := make(map[string]interface{})
		for i, colName := range columnNames {
			row[colName] = values[i]
		}
		results = append(results, row)
	}

	if err := rows.Err(); err != nil {
		response := ToolResponse{
			Content: []ContentItem{
				{
					Type: "text",
					Text: fmt.Sprintf("Error iterating rows: %v", err),
				},
			},
			IsError: true,
		}
		sendResponse(id, response)
		return
	}

	// Format results as JSON
	resultsJSON, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		response := ToolResponse{
			Content: []ContentItem{
				{
					Type: "text",
					Text: fmt.Sprintf("Error formatting results: %v", err),
				},
			},
			IsError: true,
		}
		sendResponse(id, response)
		return
	}

	// Build response
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Natural Language Query: %s\n\n", query))
	sb.WriteString(fmt.Sprintf("Generated SQL:\n%s\n\n", sqlQuery))
	sb.WriteString(fmt.Sprintf("Results (%d rows):\n%s", len(results), string(resultsJSON)))

	response := ToolResponse{
		Content: []ContentItem{
			{
				Type: "text",
				Text: sb.String(),
			},
		},
	}

	sendResponse(id, response)
}

func (s *MCPServer) generateSchemaContext() string {
	var sb strings.Builder

	s.mu.RLock()
	defer s.mu.RUnlock()

	for key, table := range s.metadata {
		sb.WriteString(fmt.Sprintf("\n%s (%s)\n", key, table.TableType))
		if table.Description != "" {
			sb.WriteString(fmt.Sprintf("  Description: %s\n", table.Description))
		}
		sb.WriteString("  Columns:\n")

		for _, col := range table.Columns {
			sb.WriteString(fmt.Sprintf("    - %s (%s)", col.ColumnName, col.DataType))
			if col.IsNullable == "YES" {
				sb.WriteString(" NULL")
			}
			if col.Description != "" {
				sb.WriteString(fmt.Sprintf(" - %s", col.Description))
			}
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

func (s *MCPServer) handleGetSchemaInfo(id interface{}, args map[string]interface{}) {
	schemaName, _ := args["schema_name"].(string)

	// Check if metadata is loaded (thread-safe)
	s.mu.RLock()
	loaded := s.metadataLoaded
	s.mu.RUnlock()

	if !loaded {
		response := ToolResponse{
			Content: []ContentItem{
				{
					Type: "text",
					Text: "Database is still initializing. Please wait a moment and try again.\n\n" +
						"The server is loading database metadata in the background. This usually takes a few seconds.",
				},
			},
			IsError: true,
		}
		sendResponse(id, response)
		return
	}

	var sb strings.Builder
	sb.WriteString("Database Schema Information:\n")
	sb.WriteString("============================\n")

	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, table := range s.metadata {
		// Filter by schema if requested
		if schemaName != "" && table.SchemaName != schemaName {
			continue
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

	response := ToolResponse{
		Content: []ContentItem{
			{
				Type: "text",
				Text: sb.String(),
			},
		},
	}

	sendResponse(id, response)
}

func sendResponse(id interface{}, result interface{}) {
	resp := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}

	data, _ := json.Marshal(resp)
	fmt.Println(string(data))
	os.Stdout.Sync()
}

func sendError(id interface{}, code int, message string, data interface{}) {
	resp := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &RPCError{
			Code:    code,
			Message: message,
			Data:    data,
		},
	}

	respData, _ := json.Marshal(resp)
	fmt.Println(string(respData))
	os.Stdout.Sync()
}
