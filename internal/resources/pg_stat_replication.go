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

// PGStatReplicationResource provides replication status
func PGStatReplicationResource(dbClient *database.Client) Resource {
	return Resource{
		Definition: mcp.Resource{
			URI:         URIStatReplication,
			Name:        "PostgreSQL Replication Status",
			Description: "Shows the status of replication connections from this primary server including WAL sender processes, replication lag, and sync state. Empty if the server is not a replication primary or has no active replicas. Critical for monitoring replication health and identifying lag issues.",
			MimeType:    "application/json",
		},
		Handler: func() (mcp.ResourceContent, error) {
			// Check if we have the replay_lag column (PG 10+)
			query := `
				SELECT
					pid,
					usename,
					application_name,
					client_addr::text,
					client_hostname,
					client_port,
					backend_start::text as backend_start,
					state,
					sync_state,
					COALESCE(replay_lag::text, 'N/A') as replay_lag
				FROM pg_stat_replication
				ORDER BY backend_start`

			processor := func(rows pgx.Rows) (interface{}, error) {
				var replicas []map[string]interface{}

				for rows.Next() {
					var pid, clientPort int
					var usename, applicationName, clientAddr, clientHostname, state, syncState, replayLag *string
					var backendStart *string

					if err := rows.Scan(&pid, &usename, &applicationName, &clientAddr,
						&clientHostname, &clientPort, &backendStart, &state, &syncState, &replayLag); err != nil {
						fmt.Fprintf(os.Stderr, "WARNING: Failed to scan row in pg_stat_replication: %v\n", err)
						continue
					}

					replicas = append(replicas, map[string]interface{}{
						"pid":              pid,
						"usename":          usename,
						"application_name": applicationName,
						"client_addr":      clientAddr,
						"client_hostname":  clientHostname,
						"client_port":      clientPort,
						"backend_start":    backendStart,
						"state":            state,
						"sync_state":       syncState,
						"replay_lag":       replayLag,
					})
				}

				var statusMsg string
				if len(replicas) == 0 {
					statusMsg = "No active replicas. This server is either not a primary, or has no connected standby servers."
				} else {
					statusMsg = fmt.Sprintf("Primary server with %d active replica(s)", len(replicas))
				}

				return map[string]interface{}{
					"replica_count": len(replicas),
					"replicas":      replicas,
					"status":        statusMsg,
					"description":   "Replication status for all connected standby servers. Monitor replay_lag to detect replication delays.",
				}, nil
			}

			return database.ExecuteResourceQuery(dbClient, URIStatReplication, query, processor)
		},
	}
}
