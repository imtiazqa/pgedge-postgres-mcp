package llm

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestNewClient(t *testing.T) {
	// Save original env vars
	originalAPIKey := os.Getenv("ANTHROPIC_API_KEY")
	originalModel := os.Getenv("ANTHROPIC_MODEL")
	defer func() {
		os.Setenv("ANTHROPIC_API_KEY", originalAPIKey)
		os.Setenv("ANTHROPIC_MODEL", originalModel)
	}()

	t.Run("default model when not set", func(t *testing.T) {
		os.Unsetenv("ANTHROPIC_MODEL")
		os.Setenv("ANTHROPIC_API_KEY", "test-key")

		client := NewClient()
		if client == nil {
			t.Fatal("NewClient() returned nil")
		}
		if client.model != "claude-sonnet-4-5" {
			t.Errorf("model = %q, want %q", client.model, "claude-sonnet-4-5")
		}
		if client.apiKey != "test-key" {
			t.Errorf("apiKey = %q, want %q", client.apiKey, "test-key")
		}
		if client.baseURL != "https://api.anthropic.com/v1" {
			t.Errorf("baseURL = %q, want %q", client.baseURL, "https://api.anthropic.com/v1")
		}
	})

	t.Run("custom model from env", func(t *testing.T) {
		os.Setenv("ANTHROPIC_MODEL", "claude-3-opus-20240229")
		os.Setenv("ANTHROPIC_API_KEY", "test-key-2")

		client := NewClient()
		if client.model != "claude-3-opus-20240229" {
			t.Errorf("model = %q, want %q", client.model, "claude-3-opus-20240229")
		}
	})

	t.Run("no api key", func(t *testing.T) {
		os.Unsetenv("ANTHROPIC_API_KEY")

		client := NewClient()
		if client.apiKey != "" {
			t.Errorf("apiKey = %q, want empty string", client.apiKey)
		}
	})
}

func TestIsConfigured(t *testing.T) {
	tests := []struct {
		name     string
		apiKey   string
		expected bool
	}{
		{
			name:     "with api key",
			apiKey:   "sk-ant-test-key",
			expected: true,
		},
		{
			name:     "without api key",
			apiKey:   "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{
				apiKey: tt.apiKey,
			}

			result := client.IsConfigured()
			if result != tt.expected {
				t.Errorf("IsConfigured() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestConvertNLToSQL_NotConfigured(t *testing.T) {
	client := &Client{
		apiKey: "",
	}

	_, err := client.ConvertNLToSQL("show all users", "schema context")
	if err == nil {
		t.Error("ConvertNLToSQL() expected error when not configured, got nil")
	}
	if !strings.Contains(err.Error(), "ANTHROPIC_API_KEY not set") {
		t.Errorf("ConvertNLToSQL() error = %v, want error containing 'ANTHROPIC_API_KEY not set'", err)
	}
}

func TestConvertNLToSQL_Success(t *testing.T) {
	// Create a mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method and headers
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}
		if r.Header.Get("x-api-key") != "test-api-key" {
			t.Errorf("Expected x-api-key test-api-key, got %s", r.Header.Get("x-api-key"))
		}
		if r.Header.Get("anthropic-version") != "2023-06-01" {
			t.Errorf("Expected anthropic-version 2023-06-01, got %s", r.Header.Get("anthropic-version"))
		}

		// Send mock response
		response := claudeResponse{
			ID:   "msg_123",
			Type: "message",
			Role: "assistant",
			Content: []claudeContentBlock{
				{
					Type: "text",
					Text: "SELECT * FROM users WHERE active = true",
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := &Client{
		apiKey:  "test-api-key",
		baseURL: server.URL,
		model:   "claude-sonnet-4-5",
	}

	result, err := client.ConvertNLToSQL("show active users", "public.users (TABLE)\n  Columns:\n    - id (integer)\n    - active (boolean)")
	if err != nil {
		t.Fatalf("ConvertNLToSQL() unexpected error: %v", err)
	}

	expected := "SELECT * FROM users WHERE active = true"
	if result != expected {
		t.Errorf("ConvertNLToSQL() = %q, want %q", result, expected)
	}
}

func TestConvertNLToSQL_CleanupSQL(t *testing.T) {
	tests := []struct {
		name         string
		responseText string
		expected     string
	}{
		{
			name:         "plain SQL",
			responseText: "SELECT * FROM users",
			expected:     "SELECT * FROM users",
		},
		{
			name:         "SQL with trailing semicolon",
			responseText: "SELECT * FROM users;",
			expected:     "SELECT * FROM users",
		},
		{
			name:         "SQL with markdown code block",
			responseText: "```sql\nSELECT * FROM users\n```",
			expected:     "SELECT * FROM users",
		},
		{
			name:         "SQL with generic code block",
			responseText: "```\nSELECT * FROM users\n```",
			expected:     "SELECT * FROM users",
		},
		{
			name:         "SQL with whitespace",
			responseText: "  SELECT * FROM users  ",
			expected:     "SELECT * FROM users",
		},
		{
			name:         "SQL with code block and semicolon",
			responseText: "```sql\nSELECT * FROM users;\n```",
			expected:     "SELECT * FROM users",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock HTTP server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				response := claudeResponse{
					ID:   "msg_123",
					Type: "message",
					Role: "assistant",
					Content: []claudeContentBlock{
						{
							Type: "text",
							Text: tt.responseText,
						},
					},
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(response)
			}))
			defer server.Close()

			client := &Client{
				apiKey:  "test-api-key",
				baseURL: server.URL,
				model:   "claude-sonnet-4-5",
			}

			result, err := client.ConvertNLToSQL("test query", "schema")
			if err != nil {
				t.Fatalf("ConvertNLToSQL() unexpected error: %v", err)
			}

			if result != tt.expected {
				t.Errorf("ConvertNLToSQL() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestConvertNLToSQL_APIError(t *testing.T) {
	// Create a mock HTTP server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": {"type": "invalid_request_error", "message": "Invalid request"}}`))
	}))
	defer server.Close()

	client := &Client{
		apiKey:  "test-api-key",
		baseURL: server.URL,
		model:   "claude-sonnet-4-5",
	}

	_, err := client.ConvertNLToSQL("show users", "schema")
	if err == nil {
		t.Error("ConvertNLToSQL() expected error for API error response, got nil")
	}
	if !strings.Contains(err.Error(), "API returned status 400") {
		t.Errorf("ConvertNLToSQL() error = %v, want error containing 'API returned status 400'", err)
	}
}

func TestConvertNLToSQL_EmptyResponse(t *testing.T) {
	// Create a mock HTTP server that returns empty content
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := claudeResponse{
			ID:      "msg_123",
			Type:    "message",
			Role:    "assistant",
			Content: []claudeContentBlock{},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := &Client{
		apiKey:  "test-api-key",
		baseURL: server.URL,
		model:   "claude-sonnet-4-5",
	}

	_, err := client.ConvertNLToSQL("show users", "schema")
	if err == nil {
		t.Error("ConvertNLToSQL() expected error for empty response, got nil")
	}
	if !strings.Contains(err.Error(), "no content in response") {
		t.Errorf("ConvertNLToSQL() error = %v, want error containing 'no content in response'", err)
	}
}
