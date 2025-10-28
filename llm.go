package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

// ClaudeRequest represents a request to the Claude API
type ClaudeRequest struct {
	Model     string          `json:"model"`
	MaxTokens int             `json:"max_tokens"`
	Messages  []ClaudeMessage `json:"messages"`
}

type ClaudeMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ClaudeResponse struct {
	ID      string               `json:"id"`
	Type    string               `json:"type"`
	Role    string               `json:"role"`
	Content []ClaudeContentBlock `json:"content"`
	Usage   ClaudeUsage          `json:"usage"`
}

type ClaudeContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type ClaudeUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// LLMClient handles interactions with Claude API
type LLMClient struct {
	apiKey  string
	baseURL string
	model   string
}

func NewLLMClient() *LLMClient {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	model := os.Getenv("ANTHROPIC_MODEL")
	if model == "" {
		model = "claude-sonnet-4-5" // Note to Claude - do not change this!
	}

	return &LLMClient{
		apiKey:  apiKey,
		baseURL: "https://api.anthropic.com/v1",
		model:   model,
	}
}

func (c *LLMClient) IsConfigured() bool {
	return c.apiKey != ""
}

// ConvertNLToSQL converts a natural language query to SQL using Claude
func (c *LLMClient) ConvertNLToSQL(nlQuery string, schemaContext string) (string, error) {
	if !c.IsConfigured() {
		return "", fmt.Errorf("ANTHROPIC_API_KEY not set")
	}

	prompt := fmt.Sprintf(`You are a PostgreSQL expert. Given the following database schema and a natural language query, generate a SQL query that answers the question.

Database Schema:
%s

Natural Language Query: %s

Requirements:
1. Generate ONLY the SQL query, no explanations or markdown formatting
2. Use proper PostgreSQL syntax
3. Consider the column descriptions and table relationships
4. Use appropriate JOINs when needed
5. Include proper WHERE clauses, GROUP BY, ORDER BY as needed
6. Use meaningful column aliases
7. Make the query efficient and optimized
8. Do NOT include semicolons at the end
9. Return ONLY the SQL query text, nothing else

SQL Query:`, schemaContext, nlQuery)

	reqBody := ClaudeRequest{
		Model:     c.model,
		MaxTokens: 2048,
		Messages: []ClaudeMessage{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", c.baseURL+"/messages", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var claudeResp ClaudeResponse
	if err := json.Unmarshal(body, &claudeResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(claudeResp.Content) == 0 {
		return "", fmt.Errorf("no content in response")
	}

	sqlQuery := strings.TrimSpace(claudeResp.Content[0].Text)

	// Clean up the SQL query
	sqlQuery = strings.TrimPrefix(sqlQuery, "```sql")
	sqlQuery = strings.TrimPrefix(sqlQuery, "```")
	sqlQuery = strings.TrimSuffix(sqlQuery, "```")
	sqlQuery = strings.TrimSpace(sqlQuery)
	sqlQuery = strings.TrimSuffix(sqlQuery, ";")

	return sqlQuery, nil
}

// ExplainQuery generates an explanation of a SQL query
func (c *LLMClient) ExplainQuery(sqlQuery string, schemaContext string) (string, error) {
	if !c.IsConfigured() {
		return "", fmt.Errorf("ANTHROPIC_API_KEY not set")
	}

	prompt := fmt.Sprintf(`You are a PostgreSQL expert. Given the following SQL query and database schema, explain what the query does in simple terms.

Database Schema:
%s

SQL Query:
%s

Provide a clear, concise explanation of:
1. What data the query retrieves
2. Which tables/views it uses
3. Any filtering, grouping, or sorting logic
4. The purpose and meaning of the results

Explanation:`, schemaContext, sqlQuery)

	reqBody := ClaudeRequest{
		Model:     c.model,
		MaxTokens: 1024,
		Messages: []ClaudeMessage{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", c.baseURL+"/messages", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var claudeResp ClaudeResponse
	if err := json.Unmarshal(body, &claudeResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(claudeResp.Content) == 0 {
		return "", fmt.Errorf("no content in response")
	}

	return strings.TrimSpace(claudeResp.Content[0].Text), nil
}
