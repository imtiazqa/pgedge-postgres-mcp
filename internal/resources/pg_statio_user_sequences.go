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

// PGStatIOUserSequencesResource provides I/O statistics for user sequences
func PGStatIOUserSequencesResource(dbClient *database.Client) Resource {
	return Resource{
		Definition: mcp.Resource{
			URI:         URIStatIOUserSequences,
			Name:        "PostgreSQL Sequence I/O Statistics",
			Description: "Shows disk block I/O statistics for user sequences. Tracks blocks read from disk vs. cache hits for sequence objects. Sequences should typically have very high cache hit ratios since they're frequently accessed. Low hit ratios may indicate cache pressure or excessive sequence usage patterns.",
			MimeType:    "application/json",
		},
		Handler: func() (mcp.ResourceContent, error) {
			query := `
				SELECT
					schemaname,
					relname,
					blks_read,
					blks_hit
				FROM pg_statio_user_sequences
				ORDER BY blks_read DESC`

			processor := func(rows pgx.Rows) (interface{}, error) {
				var sequences []map[string]interface{}

				for rows.Next() {
					var schemaname, relname string
					var blksRead, blksHit int64

					if err := rows.Scan(&schemaname, &relname,
						&blksRead, &blksHit); err != nil {
						fmt.Fprintf(os.Stderr, "WARNING: Failed to scan row in pg_statio_user_sequences: %v\n", err)
						continue
					}

					// Calculate cache hit ratio
					totalBlks := blksRead + blksHit
					var hitRatio *float64
					if totalBlks > 0 {
						ratio := float64(blksHit) / float64(totalBlks) * 100
						hitRatio = &ratio
					}

					// Assess cache performance
					var cacheStatus string
					switch {
					case totalBlks == 0:
						cacheStatus = "no_activity"
					case hitRatio != nil && *hitRatio >= 99:
						cacheStatus = "excellent"
					case hitRatio != nil && *hitRatio >= 95:
						cacheStatus = "good"
					case hitRatio != nil && *hitRatio >= 80:
						cacheStatus = "needs_attention"
					default:
						cacheStatus = "poor"
					}

					sequences = append(sequences, map[string]interface{}{
						"schemaname":   schemaname,
						"relname":      relname,
						"blks_read":    blksRead,
						"blks_hit":     blksHit,
						"hit_ratio":    hitRatio,
						"cache_status": cacheStatus,
					})
				}

				return map[string]interface{}{
					"sequence_count": len(sequences),
					"sequences":      sequences,
					"description":    "Per-sequence I/O statistics showing disk reads vs cache hits. Sequences ordered by disk reads (highest first). Hit ratios above 99% are expected; lower values may indicate cache issues.",
				}, nil
			}

			return database.ExecuteResourceQuery(dbClient, URIStatIOUserSequences, query, processor)
		},
	}
}
