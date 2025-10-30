/*-------------------------------------------------------------------------
 *
 * pgEdge Postgres MCP Server
 *
 * Copyright (c) 2025, pgEdge, Inc.
 * This software is released under The PostgreSQL License
 *
 *-------------------------------------------------------------------------
 */

package tools

import (
	"strings"
	"testing"

	"pgedge-postgres-mcp/internal/database"
)

func TestSetPGConfigurationTool(t *testing.T) {
	t.Run("missing parameter argument", func(t *testing.T) {
		client := database.NewTestClient("postgres://localhost/test", nil)
		tool := SetPGConfigurationTool(client)

		response, err := tool.Handler(map[string]interface{}{
			"value": "100",
		})

		if err != nil {
			t.Errorf("Handler returned error: %v", err)
		}
		if !response.IsError {
			t.Error("Expected IsError=true when parameter is missing")
		}
		content := response.Content[0].Text
		if !strings.Contains(content, "Missing or invalid 'parameter' argument") {
			t.Errorf("Expected parameter missing error, got: %s", content)
		}
	})

	t.Run("empty parameter string", func(t *testing.T) {
		client := database.NewTestClient("postgres://localhost/test", nil)
		tool := SetPGConfigurationTool(client)

		response, err := tool.Handler(map[string]interface{}{
			"parameter": "",
			"value":     "100",
		})

		if err != nil {
			t.Errorf("Handler returned error: %v", err)
		}
		if !response.IsError {
			t.Error("Expected IsError=true when parameter is empty")
		}
		content := response.Content[0].Text
		if !strings.Contains(content, "Missing or invalid 'parameter' argument") {
			t.Errorf("Expected parameter missing error, got: %s", content)
		}
	})

	t.Run("invalid parameter type", func(t *testing.T) {
		client := database.NewTestClient("postgres://localhost/test", nil)
		tool := SetPGConfigurationTool(client)

		response, err := tool.Handler(map[string]interface{}{
			"parameter": 123, // Invalid type - should be string
			"value":     "100",
		})

		if err != nil {
			t.Errorf("Handler returned error: %v", err)
		}
		if !response.IsError {
			t.Error("Expected IsError=true when parameter has invalid type")
		}
		content := response.Content[0].Text
		if !strings.Contains(content, "Missing or invalid 'parameter' argument") {
			t.Errorf("Expected parameter invalid error, got: %s", content)
		}
	})

	t.Run("missing value argument", func(t *testing.T) {
		client := database.NewTestClient("postgres://localhost/test", nil)
		tool := SetPGConfigurationTool(client)

		response, err := tool.Handler(map[string]interface{}{
			"parameter": "max_connections",
		})

		if err != nil {
			t.Errorf("Handler returned error: %v", err)
		}
		if !response.IsError {
			t.Error("Expected IsError=true when value is missing")
		}
		content := response.Content[0].Text
		if !strings.Contains(content, "Missing or invalid 'value' argument") {
			t.Errorf("Expected value missing error, got: %s", content)
		}
	})

	t.Run("invalid value type", func(t *testing.T) {
		client := database.NewTestClient("postgres://localhost/test", nil)
		tool := SetPGConfigurationTool(client)

		response, err := tool.Handler(map[string]interface{}{
			"parameter": "max_connections",
			"value":     123, // Invalid type - should be string
		})

		if err != nil {
			t.Errorf("Handler returned error: %v", err)
		}
		if !response.IsError {
			t.Error("Expected IsError=true when value has invalid type")
		}
		content := response.Content[0].Text
		if !strings.Contains(content, "Missing or invalid 'value' argument") {
			t.Errorf("Expected value invalid error, got: %s", content)
		}
	})

	t.Run("database not ready", func(t *testing.T) {
		client := database.NewClient()
		// Don't add any connections - database is not ready

		tool := SetPGConfigurationTool(client)
		response, err := tool.Handler(map[string]interface{}{
			"parameter": "max_connections",
			"value":     "100",
		})

		if err != nil {
			t.Errorf("Handler returned error: %v", err)
		}
		if !response.IsError {
			t.Error("Expected IsError=true when database not ready")
		}
		content := response.Content[0].Text
		if !strings.Contains(content, "Database is still initializing") {
			t.Errorf("Expected database not ready message, got: %s", content)
		}
	})

	t.Run("tool definition has required fields", func(t *testing.T) {
		client := database.NewClient()
		tool := SetPGConfigurationTool(client)

		if tool.Definition.Name != "set_pg_configuration" {
			t.Errorf("Expected name 'set_pg_configuration', got %s", tool.Definition.Name)
		}

		if tool.Definition.Description == "" {
			t.Error("Description should not be empty")
		}

		if tool.Definition.InputSchema.Type != "object" {
			t.Errorf("Expected input schema type 'object', got %s", tool.Definition.InputSchema.Type)
		}

		// Check required fields
		required := tool.Definition.InputSchema.Required
		if len(required) != 2 {
			t.Errorf("Expected 2 required fields, got %d", len(required))
		}

		hasParameter := false
		hasValue := false
		for _, field := range required {
			if field == "parameter" {
				hasParameter = true
			}
			if field == "value" {
				hasValue = true
			}
		}

		if !hasParameter {
			t.Error("'parameter' should be in required fields")
		}
		if !hasValue {
			t.Error("'value' should be in required fields")
		}

		// Check properties exist
		props := tool.Definition.InputSchema.Properties
		if _, ok := props["parameter"]; !ok {
			t.Error("'parameter' property should exist")
		}
		if _, ok := props["value"]; !ok {
			t.Error("'value' property should exist")
		}
	})

	// Note: Testing actual database operations (ALTER SYSTEM SET) would require
	// either a real database connection or extensive mocking of pgx.Pool.
	// Those tests are better suited for integration tests rather than unit tests.
	// The tests above cover input validation and error handling which don't require
	// a real database connection.
}
