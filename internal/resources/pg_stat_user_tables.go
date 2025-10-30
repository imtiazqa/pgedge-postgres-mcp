package resources

import (
	"fmt"
	"os"

	"pgedge-postgres-mcp/internal/database"
	"pgedge-postgres-mcp/internal/mcp"

	"github.com/jackc/pgx/v5"
)

// PGStatUserTablesResource provides table access statistics
func PGStatUserTablesResource(dbClient *database.Client) Resource {
	return Resource{
		Definition: mcp.Resource{
			URI:         URIStatUserTables,
			Name:        "PostgreSQL Table Statistics",
			Description: "Shows statistics for user tables including sequential and index scans, tuple operations (inserts/updates/deletes), and vacuum/analyze activity. Critical for identifying tables that need optimization or indexing improvements.",
			MimeType:    "application/json",
		},
		Handler: func() (mcp.ResourceContent, error) {
			query := fmt.Sprintf(`
				SELECT
					schemaname, relname, seq_scan, seq_tup_read,
					idx_scan, idx_tup_fetch, n_tup_ins, n_tup_upd,
					n_tup_del, n_live_tup, n_dead_tup
				FROM pg_stat_user_tables
				ORDER BY schemaname, relname
				LIMIT %d`, DefaultQueryLimit)

			processor := func(rows pgx.Rows) (interface{}, error) {
				var tables []map[string]interface{}

				for rows.Next() {
					var schemaname, relname string
					var seqScan, seqTupRead, nTupIns, nTupUpd, nTupDel int64
					var idxScan, idxTupFetch *int64
					var nLiveTup, nDeadTup int64

					if err := rows.Scan(&schemaname, &relname, &seqScan, &seqTupRead,
						&idxScan, &idxTupFetch, &nTupIns, &nTupUpd, &nTupDel,
						&nLiveTup, &nDeadTup); err != nil {
						fmt.Fprintf(os.Stderr, "WARNING: Failed to scan row in pg_stat_user_tables: %v\n", err)
						continue
					}

					tables = append(tables, map[string]interface{}{
						"schemaname":    schemaname,
						"relname":       relname,
						"seq_scan":      seqScan,
						"seq_tup_read":  seqTupRead,
						"idx_scan":      idxScan,
						"idx_tup_fetch": idxTupFetch,
						"n_tup_ins":     nTupIns,
						"n_tup_upd":     nTupUpd,
						"n_tup_del":     nTupDel,
						"n_live_tup":    nLiveTup,
						"n_dead_tup":    nDeadTup,
					})
				}

				return map[string]interface{}{
					"table_count": len(tables),
					"tables":      tables,
				}, nil
			}

			return database.ExecuteResourceQuery(dbClient, URIStatUserTables, query, processor)
		},
	}
}
