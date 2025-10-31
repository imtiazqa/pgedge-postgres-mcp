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

// Resource URI constants
// These constants define the URIs for all available MCP resources
const (
	// System Information Resources
	URISettings   = "pg://settings"
	URISystemInfo = "pg://system_info"

	// Statistics Resources
	URIStatActivity     = "pg://stat/activity"
	URIStatDatabase     = "pg://stat/database"
	URIStatUserTables   = "pg://stat/user_tables"
	URIStatUserIndexes  = "pg://stat/user_indexes"
	URIStatReplication  = "pg://stat/replication"
	URIStatBgwriter     = "pg://stat/bgwriter"
	URIStatWAL          = "pg://stat/wal"

	// I/O Statistics Resources
	URIStatIOUserTables   = "pg://statio/user_tables"
	URIStatIOUserIndexes  = "pg://statio/user_indexes"
	URIStatIOUserSequences = "pg://statio/user_sequences"
)
