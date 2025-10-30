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
	"fmt"
	"os"

	"pgedge-postgres-mcp/internal/database"
	"pgedge-postgres-mcp/internal/mcp"

	"github.com/jackc/pgx/v5"
)

// PGStatUserIndexesResource provides index usage statistics
func PGStatUserIndexesResource(dbClient *database.Client) Resource {
	return Resource{
		Definition: mcp.Resource{
			URI:         URIStatUserIndexes,
			Name:        "PostgreSQL Index Statistics",
			Description: "Provides statistics about index usage including scan counts and tuple operations. Essential for identifying unused indexes that can be dropped and finding tables that might benefit from additional indexes. Helps optimize query performance and reduce storage overhead.",
			MimeType:    "application/json",
		},
		Handler: func() (mcp.ResourceContent, error) {
			query := `
				SELECT
					schemaname,
					relname,
					indexrelname,
					idx_scan,
					idx_tup_read,
					idx_tup_fetch
				FROM pg_stat_user_indexes
				ORDER BY schemaname, relname, indexrelname`

			processor := func(rows pgx.Rows) (interface{}, error) {
				var indexes []map[string]interface{}

				for rows.Next() {
					var schemaname, relname, indexrelname string
					var idxScan, idxTupRead, idxTupFetch int64

					if err := rows.Scan(&schemaname, &relname, &indexrelname,
						&idxScan, &idxTupRead, &idxTupFetch); err != nil {
						fmt.Fprintf(os.Stderr, "WARNING: Failed to scan row in pg_stat_user_indexes: %v\n", err)
						continue
					}

					// Mark potentially unused indexes
					var usage string
					switch {
					case idxScan == 0:
						usage = "unused"
					case idxScan < 100:
						usage = "rarely_used"
					default:
						usage = "active"
					}

					indexes = append(indexes, map[string]interface{}{
						"schemaname":    schemaname,
						"relname":       relname,
						"indexrelname":  indexrelname,
						"idx_scan":      idxScan,
						"idx_tup_read":  idxTupRead,
						"idx_tup_fetch": idxTupFetch,
						"usage_status":  usage,
					})
				}

				return map[string]interface{}{
					"index_count": len(indexes),
					"indexes":     indexes,
					"description": "Per-index statistics showing usage patterns and effectiveness. Indexes with idx_scan=0 may be candidates for removal.",
				}, nil
			}

			return database.ExecuteResourceQuery(dbClient, URIStatUserIndexes, query, processor)
		},
	}
}
