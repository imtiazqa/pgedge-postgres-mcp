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

// PGStatActivityResource provides current activity and connections
func PGStatActivityResource(dbClient *database.Client) Resource {
	return Resource{
		Definition: mcp.Resource{
			URI:         URIStatActivity,
			Name:        "PostgreSQL Current Activity",
			Description: "Shows information about currently executing queries and connections. Useful for monitoring active sessions, identifying long-running queries, and understanding current database load. Each row represents one server process with details about its current activity.",
			MimeType:    "application/json",
		},
		Handler: func() (mcp.ResourceContent, error) {
			query := fmt.Sprintf(`
				SELECT
					datname,
					pid,
					usename,
					application_name,
					client_addr::text,
					backend_start::text as backend_start,
					state,
					query
				FROM pg_stat_activity
				WHERE pid != pg_backend_pid()
				ORDER BY backend_start DESC
				LIMIT %d
			`, DefaultQueryLimit)

			processor := func(rows pgx.Rows) (interface{}, error) {
				var activities []map[string]interface{}

				for rows.Next() {
					var datname, usename, applicationName, clientAddr, state, query *string
					var pid int
					var backendStart *string

					err := rows.Scan(&datname, &pid, &usename, &applicationName, &clientAddr,
						&backendStart, &state, &query)
					if err != nil {
						fmt.Fprintf(os.Stderr, "WARNING: Failed to scan row in pg_stat_activity: %v\n", err)
						continue
					}

					activity := map[string]interface{}{
						"datname":          datname,
						"pid":              pid,
						"usename":          usename,
						"application_name": applicationName,
						"client_addr":      clientAddr,
						"backend_start":    backendStart,
						"state":            state,
						"query":            query,
					}

					activities = append(activities, activity)
				}

				return map[string]interface{}{
					"activity_count": len(activities),
					"activities":     activities,
					"description":    "Current database activity including active queries and connections.",
				}, nil
			}

			return database.ExecuteResourceQuery(dbClient, URIStatActivity, query, processor)
		},
	}
}
