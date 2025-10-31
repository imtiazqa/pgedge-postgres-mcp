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
	"encoding/json"
	"os"
	"strings"
	"testing"

	"pgedge-postgres-mcp/internal/database"
)

// TestPGStatIOUserTablesResource_Structure tests resource structure
func TestPGStatIOUserTablesResource_Structure(t *testing.T) {
	client := database.NewClient()
	resource := PGStatIOUserTablesResource(client)

	if resource.Definition.URI != URIStatIOUserTables {
		t.Errorf("Expected URI %s, got %s", URIStatIOUserTables, resource.Definition.URI)
	}

	if resource.Definition.Name == "" {
		t.Error("Resource name should not be empty")
	}

	if resource.Definition.Description == "" {
		t.Error("Resource description should not be empty")
	}

	if resource.Definition.MimeType != "application/json" {
		t.Errorf("Expected MimeType application/json, got %s", resource.Definition.MimeType)
	}

	if resource.Handler == nil {
		t.Error("Resource handler should not be nil")
	}

	// Verify description mentions key I/O concepts
	desc := strings.ToLower(resource.Definition.Description)
	if !strings.Contains(desc, "i/o") && !strings.Contains(desc, "disk") {
		t.Error("Description should mention I/O or disk")
	}
	if !strings.Contains(desc, "cache") && !strings.Contains(desc, "hit") {
		t.Error("Description should mention cache or hit ratio")
	}
}

// TestPGStatIOUserIndexesResource_Structure tests resource structure
func TestPGStatIOUserIndexesResource_Structure(t *testing.T) {
	client := database.NewClient()
	resource := PGStatIOUserIndexesResource(client)

	if resource.Definition.URI != URIStatIOUserIndexes {
		t.Errorf("Expected URI %s, got %s", URIStatIOUserIndexes, resource.Definition.URI)
	}

	if resource.Definition.Name == "" {
		t.Error("Resource name should not be empty")
	}

	if resource.Definition.Description == "" {
		t.Error("Resource description should not be empty")
	}

	if resource.Definition.MimeType != "application/json" {
		t.Errorf("Expected MimeType application/json, got %s", resource.Definition.MimeType)
	}

	if resource.Handler == nil {
		t.Error("Resource handler should not be nil")
	}

	// Verify description mentions indexes and I/O
	desc := strings.ToLower(resource.Definition.Description)
	if !strings.Contains(desc, "index") {
		t.Error("Description should mention indexes")
	}
	if !strings.Contains(desc, "i/o") || !strings.Contains(desc, "cache") {
		t.Error("Description should mention I/O and cache")
	}
}

// TestPGStatIOUserSequencesResource_Structure tests resource structure
func TestPGStatIOUserSequencesResource_Structure(t *testing.T) {
	client := database.NewClient()
	resource := PGStatIOUserSequencesResource(client)

	if resource.Definition.URI != URIStatIOUserSequences {
		t.Errorf("Expected URI %s, got %s", URIStatIOUserSequences, resource.Definition.URI)
	}

	if resource.Definition.Name == "" {
		t.Error("Resource name should not be empty")
	}

	if resource.Definition.Description == "" {
		t.Error("Resource description should not be empty")
	}

	if resource.Definition.MimeType != "application/json" {
		t.Errorf("Expected MimeType application/json, got %s", resource.Definition.MimeType)
	}

	if resource.Handler == nil {
		t.Error("Resource handler should not be nil")
	}

	// Verify description mentions sequences
	desc := strings.ToLower(resource.Definition.Description)
	if !strings.Contains(desc, "sequence") {
		t.Error("Description should mention sequences")
	}
	if !strings.Contains(desc, "cache") {
		t.Error("Description should mention cache")
	}
}

// TestPGStatIOUserTablesResource_Handler tests actual query execution
func TestPGStatIOUserTablesResource_Handler(t *testing.T) {
	// Skip if no database connection available
	if os.Getenv("POSTGRES_CONNECTION_STRING") == "" {
		t.Skip("POSTGRES_CONNECTION_STRING not set, skipping database test")
	}

	client := database.NewClient()
	if err := client.Connect(); err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer client.Close()

	resource := PGStatIOUserTablesResource(client)
	content, err := resource.Handler()

	if err != nil {
		t.Fatalf("Handler failed: %v", err)
	}

	if len(content.Contents) == 0 {
		t.Fatal("Expected non-empty content")
	}

	// Parse JSON response
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(content.Contents[0].Text), &data); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Verify expected fields
	if _, ok := data["table_count"]; !ok {
		t.Error("Expected table_count field in response")
	}

	if _, ok := data["tables"]; !ok {
		t.Error("Expected tables field in response")
	}

	if _, ok := data["description"]; !ok {
		t.Error("Expected description field in response")
	}

	// Verify tables is an array
	if tables, ok := data["tables"].([]interface{}); ok {
		// If there are tables, verify structure of first one
		if len(tables) > 0 {
			if table, ok := tables[0].(map[string]interface{}); ok {
				expectedFields := []string{
					"schemaname", "relname",
					"heap_blks_read", "heap_blks_hit",
					"idx_blks_read", "idx_blks_hit",
				}
				for _, field := range expectedFields {
					if _, ok := table[field]; !ok {
						t.Errorf("Expected field %s in table data", field)
					}
				}
			}
		}
	}
}

// TestPGStatIOUserIndexesResource_Handler tests actual query execution
func TestPGStatIOUserIndexesResource_Handler(t *testing.T) {
	// Skip if no database connection available
	if os.Getenv("POSTGRES_CONNECTION_STRING") == "" {
		t.Skip("POSTGRES_CONNECTION_STRING not set, skipping database test")
	}

	client := database.NewClient()
	if err := client.Connect(); err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer client.Close()

	resource := PGStatIOUserIndexesResource(client)
	content, err := resource.Handler()

	if err != nil {
		t.Fatalf("Handler failed: %v", err)
	}

	if len(content.Contents) == 0 {
		t.Fatal("Expected non-empty content")
	}

	// Parse JSON response
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(content.Contents[0].Text), &data); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Verify expected fields
	if _, ok := data["index_count"]; !ok {
		t.Error("Expected index_count field in response")
	}

	if _, ok := data["indexes"]; !ok {
		t.Error("Expected indexes field in response")
	}

	// Verify indexes is an array
	if indexes, ok := data["indexes"].([]interface{}); ok {
		// If there are indexes, verify structure
		if len(indexes) > 0 {
			if index, ok := indexes[0].(map[string]interface{}); ok {
				expectedFields := []string{
					"schemaname", "relname", "indexrelname",
					"idx_blks_read", "idx_blks_hit", "io_status",
				}
				for _, field := range expectedFields {
					if _, ok := index[field]; !ok {
						t.Errorf("Expected field %s in index data", field)
					}
				}
			}
		}
	}
}

// TestPGStatIOUserSequencesResource_Handler tests actual query execution
func TestPGStatIOUserSequencesResource_Handler(t *testing.T) {
	// Skip if no database connection available
	if os.Getenv("POSTGRES_CONNECTION_STRING") == "" {
		t.Skip("POSTGRES_CONNECTION_STRING not set, skipping database test")
	}

	client := database.NewClient()
	if err := client.Connect(); err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer client.Close()

	resource := PGStatIOUserSequencesResource(client)
	content, err := resource.Handler()

	if err != nil {
		t.Fatalf("Handler failed: %v", err)
	}

	if len(content.Contents) == 0 {
		t.Fatal("Expected non-empty content")
	}

	// Parse JSON response
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(content.Contents[0].Text), &data); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Verify expected fields
	if _, ok := data["sequence_count"]; !ok {
		t.Error("Expected sequence_count field in response")
	}

	if _, ok := data["sequences"]; !ok {
		t.Error("Expected sequences field in response")
	}

	// Verify sequences is an array
	if sequences, ok := data["sequences"].([]interface{}); ok {
		// If there are sequences, verify structure
		if len(sequences) > 0 {
			if seq, ok := sequences[0].(map[string]interface{}); ok {
				expectedFields := []string{
					"schemaname", "relname",
					"blks_read", "blks_hit", "cache_status",
				}
				for _, field := range expectedFields {
					if _, ok := seq[field]; !ok {
						t.Errorf("Expected field %s in sequence data", field)
					}
				}
			}
		}
	}
}

// TestStatIOURIConstants verifies URI constants are correctly defined
func TestStatIOURIConstants(t *testing.T) {
	// Verify URIs follow expected pattern
	expectedURIs := map[string]string{
		"URIStatIOUserTables":   "pg://statio/user_tables",
		"URIStatIOUserIndexes":  "pg://statio/user_indexes",
		"URIStatIOUserSequences": "pg://statio/user_sequences",
	}

	if URIStatIOUserTables != expectedURIs["URIStatIOUserTables"] {
		t.Errorf("Expected URIStatIOUserTables=%s, got %s", expectedURIs["URIStatIOUserTables"], URIStatIOUserTables)
	}

	if URIStatIOUserIndexes != expectedURIs["URIStatIOUserIndexes"] {
		t.Errorf("Expected URIStatIOUserIndexes=%s, got %s", expectedURIs["URIStatIOUserIndexes"], URIStatIOUserIndexes)
	}

	if URIStatIOUserSequences != expectedURIs["URIStatIOUserSequences"] {
		t.Errorf("Expected URIStatIOUserSequences=%s, got %s", expectedURIs["URIStatIOUserSequences"], URIStatIOUserSequences)
	}

	// Verify URIs are unique
	uris := []string{URIStatIOUserTables, URIStatIOUserIndexes, URIStatIOUserSequences}
	seen := make(map[string]bool)
	for _, uri := range uris {
		if seen[uri] {
			t.Errorf("Duplicate URI found: %s", uri)
		}
		seen[uri] = true
	}
}
