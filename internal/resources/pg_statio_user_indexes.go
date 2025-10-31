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

// PGStatIOUserIndexesResource provides I/O statistics for user indexes
func PGStatIOUserIndexesResource(dbClient *database.Client) Resource {
	return Resource{
		Definition: mcp.Resource{
			URI:         URIStatIOUserIndexes,
			Name:        "PostgreSQL Index I/O Statistics",
			Description: "Shows disk block I/O statistics for user indexes. Tracks blocks read from disk vs. cache hits for each index. Essential for identifying indexes causing high I/O load and evaluating cache effectiveness. Helps determine if shared_buffers should be increased or if indexes need optimization.",
			MimeType:    "application/json",
		},
		Handler: func() (mcp.ResourceContent, error) {
			query := `
				SELECT
					schemaname,
					relname,
					indexrelname,
					idx_blks_read,
					idx_blks_hit
				FROM pg_statio_user_indexes
				ORDER BY idx_blks_read DESC`

			processor := func(rows pgx.Rows) (interface{}, error) {
				var indexes []map[string]interface{}

				for rows.Next() {
					var schemaname, relname, indexrelname string
					var idxBlksRead, idxBlksHit int64

					if err := rows.Scan(&schemaname, &relname, &indexrelname,
						&idxBlksRead, &idxBlksHit); err != nil {
						fmt.Fprintf(os.Stderr, "WARNING: Failed to scan row in pg_statio_user_indexes: %v\n", err)
						continue
					}

					// Calculate cache hit ratio
					totalBlks := idxBlksRead + idxBlksHit
					var hitRatio *float64
					if totalBlks > 0 {
						ratio := float64(idxBlksHit) / float64(totalBlks) * 100
						hitRatio = &ratio
					}

					// Assess I/O performance
					var ioStatus string
					switch {
					case totalBlks == 0:
						ioStatus = "no_activity"
					case hitRatio != nil && *hitRatio >= 95:
						ioStatus = "excellent"
					case hitRatio != nil && *hitRatio >= 80:
						ioStatus = "good"
					case hitRatio != nil && *hitRatio >= 50:
						ioStatus = "needs_attention"
					default:
						ioStatus = "poor"
					}

					indexes = append(indexes, map[string]interface{}{
						"schemaname":    schemaname,
						"relname":       relname,
						"indexrelname":  indexrelname,
						"idx_blks_read": idxBlksRead,
						"idx_blks_hit":  idxBlksHit,
						"hit_ratio":     hitRatio,
						"io_status":     ioStatus,
					})
				}

				return map[string]interface{}{
					"index_count": len(indexes),
					"indexes":     indexes,
					"description": "Per-index I/O statistics showing disk reads vs cache hits. Indexes ordered by disk reads (highest first). Hit ratios above 95% are excellent, below 80% may indicate cache pressure.",
				}, nil
			}

			return database.ExecuteResourceQuery(dbClient, URIStatIOUserIndexes, query, processor)
		},
	}
}
