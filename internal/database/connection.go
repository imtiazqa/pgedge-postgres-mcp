package database

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Client manages the PostgreSQL connection and metadata
type Client struct {
	pool           *pgxpool.Pool
	metadata       map[string]TableInfo
	metadataLoaded bool
	mu             sync.RWMutex
}

// NewClient creates a new database client
func NewClient() *Client {
	return &Client{
		metadata: make(map[string]TableInfo),
	}
}

// Connect establishes a connection to the PostgreSQL database
func (c *Client) Connect() error {
	connStr := os.Getenv("POSTGRES_CONNECTION_STRING")
	if connStr == "" {
		connStr = "postgres://localhost/postgres?sslmode=disable"
	}

	var err error
	c.pool, err = pgxpool.New(context.Background(), connStr)
	if err != nil {
		return fmt.Errorf("unable to create connection pool: %w", err)
	}

	if err := c.pool.Ping(context.Background()); err != nil {
		return fmt.Errorf("unable to ping database: %w", err)
	}

	return nil
}

// Close closes the database connection
func (c *Client) Close() {
	if c.pool != nil {
		c.pool.Close()
	}
}

// LoadMetadata loads table and column metadata from the database
func (c *Client) LoadMetadata() error {
	ctx := context.Background()

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

	rows, err := c.pool.Query(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to query metadata: %w", err)
	}
	defer rows.Close()

	newMetadata := make(map[string]TableInfo)
	for rows.Next() {
		var schemaName, tableName, tableType, tableDesc, columnName, dataType, isNullable, columnDesc string

		err := rows.Scan(&schemaName, &tableName, &tableType, &tableDesc, &columnName, &dataType, &isNullable, &columnDesc)
		if err != nil {
			return fmt.Errorf("failed to scan row: %w", err)
		}

		key := schemaName + "." + tableName

		table, exists := newMetadata[key]
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

		newMetadata[key] = table
	}

	if err := rows.Err(); err != nil {
		return err
	}

	// Update metadata atomically
	c.mu.Lock()
	c.metadata = newMetadata
	c.metadataLoaded = true
	c.mu.Unlock()

	return nil
}

// GetMetadata returns a copy of the metadata map
func (c *Client) GetMetadata() map[string]TableInfo {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make(map[string]TableInfo, len(c.metadata))
	for k, v := range c.metadata {
		result[k] = v
	}
	return result
}

// IsMetadataLoaded returns whether metadata has been loaded
func (c *Client) IsMetadataLoaded() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.metadataLoaded
}

// GetPool returns the connection pool for direct queries
func (c *Client) GetPool() *pgxpool.Pool {
	return c.pool
}
