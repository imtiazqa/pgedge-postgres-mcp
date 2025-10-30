package resources

import (
	"fmt"
	"os"

	"pgedge-postgres-mcp/internal/database"
	"pgedge-postgres-mcp/internal/mcp"

	"github.com/jackc/pgx/v5"
)

// PGStatDatabaseResource provides database-wide statistics
func PGStatDatabaseResource(dbClient *database.Client) Resource {
	return Resource{
		Definition: mcp.Resource{
			URI:         URIStatDatabase,
			Name:        "PostgreSQL Database Statistics",
			Description: "Provides cumulative statistics for each database including transaction counts, block reads/writes, tuple operations, conflicts, and deadlocks. Essential for understanding database-level performance patterns and identifying I/O bottlenecks.",
			MimeType:    "application/json",
		},
		Handler: func() (mcp.ResourceContent, error) {
			query := `
				SELECT
					datname, numbackends, xact_commit, xact_rollback,
					blks_read, blks_hit, tup_returned, tup_fetched,
					tup_inserted, tup_updated, tup_deleted, conflicts,
					temp_files, temp_bytes, deadlocks
				FROM pg_stat_database
				WHERE datname IS NOT NULL
				ORDER BY datname`

			processor := func(rows pgx.Rows) (interface{}, error) {
				var databases []map[string]interface{}

				for rows.Next() {
					var datname *string
					var numbackends int
					var xactCommit, xactRollback, blksRead, blksHit int64
					var tupReturned, tupFetched, tupInserted, tupUpdated, tupDeleted int64
					var conflicts, tempFiles, deadlocks int64
					var tempBytes int64

					if err := rows.Scan(&datname, &numbackends, &xactCommit, &xactRollback,
						&blksRead, &blksHit, &tupReturned, &tupFetched, &tupInserted,
						&tupUpdated, &tupDeleted, &conflicts, &tempFiles, &tempBytes,
						&deadlocks); err != nil {
						fmt.Fprintf(os.Stderr, "WARNING: Failed to scan row in pg_stat_database: %v\n", err)
						continue
					}

					var cacheHitRatio float64
					totalReads := blksRead + blksHit
					if totalReads > 0 {
						cacheHitRatio = float64(blksHit) / float64(totalReads) * 100
					}

					databases = append(databases, map[string]interface{}{
						"datname":         datname,
						"numbackends":     numbackends,
						"xact_commit":     xactCommit,
						"xact_rollback":   xactRollback,
						"blks_read":       blksRead,
						"blks_hit":        blksHit,
						"cache_hit_ratio": fmt.Sprintf("%.2f%%", cacheHitRatio),
						"tup_returned":    tupReturned,
						"tup_fetched":     tupFetched,
						"tup_inserted":    tupInserted,
						"tup_updated":     tupUpdated,
						"tup_deleted":     tupDeleted,
						"conflicts":       conflicts,
						"temp_files":      tempFiles,
						"temp_bytes":      tempBytes,
						"deadlocks":       deadlocks,
					})
				}

				return map[string]interface{}{
					"database_count": len(databases),
					"databases":      databases,
				}, nil
			}

			return database.ExecuteResourceQuery(dbClient, URIStatDatabase, query, processor)
		},
	}
}
