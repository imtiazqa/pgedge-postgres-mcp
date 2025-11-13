/*-------------------------------------------------------------------------
 *
 * pgEdge Postgres MCP Server
 *
 * Copyright (c) 2025, pgEdge, Inc.
 * This software is released under The PostgreSQL License
 *
 *-------------------------------------------------------------------------
 */

package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the complete server configuration
type Config struct {
	// HTTP server configuration
	HTTP HTTPConfig `yaml:"http"`

	// Embedding configuration
	Embedding EmbeddingConfig `yaml:"embedding"`

	// Preferences file path
	PreferencesFile string `yaml:"preferences_file"`

	// Secret file path (for encryption key)
	SecretFile string `yaml:"secret_file"`
}

// HTTPConfig holds HTTP/HTTPS server settings
type HTTPConfig struct {
	Enabled bool       `yaml:"enabled"`
	Address string     `yaml:"address"`
	TLS     TLSConfig  `yaml:"tls"`
	Auth    AuthConfig `yaml:"auth"`
}

// AuthConfig holds authentication settings
type AuthConfig struct {
	Enabled   bool   `yaml:"enabled"`    // Whether authentication is required
	TokenFile string `yaml:"token_file"` // Path to token configuration file
}

// TLSConfig holds TLS/HTTPS settings
type TLSConfig struct {
	Enabled   bool   `yaml:"enabled"`
	CertFile  string `yaml:"cert_file"`
	KeyFile   string `yaml:"key_file"`
	ChainFile string `yaml:"chain_file"`
}

// EmbeddingConfig holds embedding generation settings
type EmbeddingConfig struct {
	Enabled         bool   `yaml:"enabled"`           // Whether embedding generation is enabled (default: false)
	Provider        string `yaml:"provider"`          // "anthropic", "openai", or "ollama"
	Model           string `yaml:"model"`             // Provider-specific model name
	AnthropicAPIKey string `yaml:"anthropic_api_key"` // API key for Anthropic (required if provider is "anthropic")
	OpenAIAPIKey    string `yaml:"openai_api_key"`    // API key for OpenAI (required if provider is "openai")
	OllamaURL       string `yaml:"ollama_url"`        // URL for Ollama service (default: http://localhost:11434)
}

// LoadConfig loads configuration with proper priority:
// 1. Command line flags (highest priority)
// 2. Environment variables
// 3. Configuration file
// 4. Hard-coded defaults (lowest priority)
func LoadConfig(configPath string, cliFlags CLIFlags) (*Config, error) {
	// Start with defaults
	cfg := defaultConfig()

	// Load config file if it exists
	if configPath != "" {
		fileCfg, err := loadConfigFile(configPath)
		if err != nil {
			// If file was explicitly specified, error out
			if cliFlags.ConfigFileSet {
				return nil, fmt.Errorf("failed to load config file %s: %w", configPath, err)
			}
			// Otherwise just use defaults (file may not exist and that's ok)
		} else {
			// Merge file config into defaults
			mergeConfig(cfg, fileCfg)
		}
	}

	// Override with environment variables
	applyEnvironmentVariables(cfg)

	// Override with command line flags (highest priority)
	applyCLIFlags(cfg, cliFlags)

	// Validate final configuration
	if err := validateConfig(cfg); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

// CLIFlags represents command line flag values and whether they were explicitly set
type CLIFlags struct {
	ConfigFileSet bool
	ConfigFile    string

	// HTTP flags
	HTTPEnabled    bool
	HTTPEnabledSet bool
	HTTPAddr       string
	HTTPAddrSet    bool

	// TLS flags
	TLSEnabled    bool
	TLSEnabledSet bool
	TLSCertFile   string
	TLSCertSet    bool
	TLSKeyFile    string
	TLSKeySet     bool
	TLSChainFile  string
	TLSChainSet   bool

	// Auth flags
	AuthEnabled    bool
	AuthEnabledSet bool
	AuthTokenFile  string
	AuthTokenSet   bool

	// Preferences flags
	PreferencesFile    string
	PreferencesFileSet bool

	// Secret file flags
	SecretFile    string
	SecretFileSet bool
}

// defaultConfig returns configuration with hard-coded defaults
func defaultConfig() *Config {
	return &Config{
		HTTP: HTTPConfig{
			Enabled: false,
			Address: ":8080",
			TLS: TLSConfig{
				Enabled:   false,
				CertFile:  "./server.crt",
				KeyFile:   "./server.key",
				ChainFile: "",
			},
			Auth: AuthConfig{
				Enabled:   true, // Authentication enabled by default
				TokenFile: "",   // Will be set to default path if not specified
			},
		},
		Embedding: EmbeddingConfig{
			Enabled:         false,                      // Disabled by default (opt-in)
			Provider:        "ollama",                   // Default provider
			Model:           "nomic-embed-text",         // Default Ollama model
			AnthropicAPIKey: "",                         // Must be provided if using Anthropic
			OllamaURL:       "http://localhost:11434",   // Default Ollama URL
		},
		PreferencesFile: "", // Will be set to default path if not specified
		SecretFile:      "", // Will be set to default path if not specified
	}
}

// loadConfigFile loads configuration from a YAML file
func loadConfigFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	return &cfg, nil
}

// mergeConfig merges source config into dest, only overriding non-zero values
func mergeConfig(dest, src *Config) {
	// HTTP
	if src.HTTP.Enabled {
		dest.HTTP.Enabled = src.HTTP.Enabled
	}
	if src.HTTP.Address != "" {
		dest.HTTP.Address = src.HTTP.Address
	}

	// TLS
	if src.HTTP.TLS.Enabled {
		dest.HTTP.TLS.Enabled = src.HTTP.TLS.Enabled
	}
	if src.HTTP.TLS.CertFile != "" {
		dest.HTTP.TLS.CertFile = src.HTTP.TLS.CertFile
	}
	if src.HTTP.TLS.KeyFile != "" {
		dest.HTTP.TLS.KeyFile = src.HTTP.TLS.KeyFile
	}
	if src.HTTP.TLS.ChainFile != "" {
		dest.HTTP.TLS.ChainFile = src.HTTP.TLS.ChainFile
	}

	// Auth - note: we need to preserve false values, so check if src differs from default
	// Use a simple heuristic: if token file is set, assume auth config is intentional
	if src.HTTP.Auth.TokenFile != "" || !src.HTTP.Auth.Enabled {
		dest.HTTP.Auth.Enabled = src.HTTP.Auth.Enabled
		dest.HTTP.Auth.TokenFile = src.HTTP.Auth.TokenFile
	}

	// Embedding - merge if any embedding fields are set
	if src.Embedding.Provider != "" || src.Embedding.Enabled {
		dest.Embedding.Enabled = src.Embedding.Enabled
		if src.Embedding.Provider != "" {
			dest.Embedding.Provider = src.Embedding.Provider
		}
		if src.Embedding.Model != "" {
			dest.Embedding.Model = src.Embedding.Model
		}
		if src.Embedding.AnthropicAPIKey != "" {
			dest.Embedding.AnthropicAPIKey = src.Embedding.AnthropicAPIKey
		}
		if src.Embedding.OllamaURL != "" {
			dest.Embedding.OllamaURL = src.Embedding.OllamaURL
		}
	}

	// Preferences
	if src.PreferencesFile != "" {
		dest.PreferencesFile = src.PreferencesFile
	}

	// Secret file
	if src.SecretFile != "" {
		dest.SecretFile = src.SecretFile
	}
}

// setStringFromEnv sets a string config value from an environment variable if it exists
func setStringFromEnv(dest *string, key string) {
	if val := os.Getenv(key); val != "" {
		*dest = val
	}
}

// setBoolFromEnv sets a boolean config value from an environment variable if it exists
// Accepts "true", "1", or "yes" as true values
func setBoolFromEnv(dest *bool, key string) {
	if val := os.Getenv(key); val != "" {
		*dest = val == "true" || val == "1" || val == "yes"
	}
}

// applyEnvironmentVariables overrides config with environment variables if they exist
// All environment variables use the PGEDGE_ prefix to avoid collisions
func applyEnvironmentVariables(cfg *Config) {
	// HTTP
	setBoolFromEnv(&cfg.HTTP.Enabled, "PGEDGE_HTTP_ENABLED")
	setStringFromEnv(&cfg.HTTP.Address, "PGEDGE_HTTP_ADDRESS")

	// TLS
	setBoolFromEnv(&cfg.HTTP.TLS.Enabled, "PGEDGE_TLS_ENABLED")
	setStringFromEnv(&cfg.HTTP.TLS.CertFile, "PGEDGE_TLS_CERT_FILE")
	setStringFromEnv(&cfg.HTTP.TLS.KeyFile, "PGEDGE_TLS_KEY_FILE")
	setStringFromEnv(&cfg.HTTP.TLS.ChainFile, "PGEDGE_TLS_CHAIN_FILE")

	// Auth
	setBoolFromEnv(&cfg.HTTP.Auth.Enabled, "PGEDGE_AUTH_ENABLED")
	setStringFromEnv(&cfg.HTTP.Auth.TokenFile, "PGEDGE_AUTH_TOKEN_FILE")

	// Embedding
	setBoolFromEnv(&cfg.Embedding.Enabled, "PGEDGE_EMBEDDING_ENABLED")
	setStringFromEnv(&cfg.Embedding.Provider, "PGEDGE_EMBEDDING_PROVIDER")
	setStringFromEnv(&cfg.Embedding.Model, "PGEDGE_EMBEDDING_MODEL")
	setStringFromEnv(&cfg.Embedding.AnthropicAPIKey, "PGEDGE_ANTHROPIC_API_KEY")
	setStringFromEnv(&cfg.Embedding.OpenAIAPIKey, "PGEDGE_OPENAI_API_KEY")
	setStringFromEnv(&cfg.Embedding.OllamaURL, "PGEDGE_OLLAMA_URL")

	// Also support standard OpenAI environment variable for convenience
	if cfg.Embedding.OpenAIAPIKey == "" {
		setStringFromEnv(&cfg.Embedding.OpenAIAPIKey, "OPENAI_API_KEY")
	}

	// Preferences
	setStringFromEnv(&cfg.PreferencesFile, "PGEDGE_PREFERENCES_FILE")

	// Secret file
	setStringFromEnv(&cfg.SecretFile, "PGEDGE_SECRET_FILE")
}

// applyCLIFlags overrides config with CLI flags if they were explicitly set
func applyCLIFlags(cfg *Config, flags CLIFlags) {
	// HTTP
	if flags.HTTPEnabledSet {
		cfg.HTTP.Enabled = flags.HTTPEnabled
	}
	if flags.HTTPAddrSet {
		cfg.HTTP.Address = flags.HTTPAddr
	}

	// TLS
	if flags.TLSEnabledSet {
		cfg.HTTP.TLS.Enabled = flags.TLSEnabled
	}
	if flags.TLSCertSet {
		cfg.HTTP.TLS.CertFile = flags.TLSCertFile
	}
	if flags.TLSKeySet {
		cfg.HTTP.TLS.KeyFile = flags.TLSKeyFile
	}
	if flags.TLSChainSet {
		cfg.HTTP.TLS.ChainFile = flags.TLSChainFile
	}

	// Auth
	if flags.AuthEnabledSet {
		cfg.HTTP.Auth.Enabled = flags.AuthEnabled
	}
	if flags.AuthTokenSet {
		cfg.HTTP.Auth.TokenFile = flags.AuthTokenFile
	}

	// Preferences
	if flags.PreferencesFileSet {
		cfg.PreferencesFile = flags.PreferencesFile
	}

	// Secret file
	if flags.SecretFileSet {
		cfg.SecretFile = flags.SecretFile
	}
}

// validateConfig checks if the configuration is valid
func validateConfig(cfg *Config) error {
	// TLS requires HTTP to be enabled
	if cfg.HTTP.TLS.Enabled && !cfg.HTTP.Enabled {
		return fmt.Errorf("TLS requires HTTP mode to be enabled")
	}

	// If HTTPS is enabled, cert and key are required
	if cfg.HTTP.TLS.Enabled {
		if cfg.HTTP.TLS.CertFile == "" {
			return fmt.Errorf("TLS certificate file is required when HTTPS is enabled")
		}
		if cfg.HTTP.TLS.KeyFile == "" {
			return fmt.Errorf("TLS key file is required when HTTPS is enabled")
		}
	}

	// If HTTP is enabled and auth is enabled, token file is required
	if cfg.HTTP.Enabled && cfg.HTTP.Auth.Enabled {
		if cfg.HTTP.Auth.TokenFile == "" {
			return fmt.Errorf("authentication token file is required when HTTP auth is enabled (use -no-auth to disable)")
		}
	}

	return nil
}

// GetDefaultConfigPath returns the default config file path (same directory as binary)
func GetDefaultConfigPath(binaryPath string) string {
	dir := filepath.Dir(binaryPath)
	return filepath.Join(dir, "pgedge-postgres-mcp.yaml")
}

// GetDefaultSecretPath returns the default secret file path (same directory as binary)
func GetDefaultSecretPath(binaryPath string) string {
	dir := filepath.Dir(binaryPath)
	return filepath.Join(dir, "pgedge-postgres-mcp.secret")
}

// ConfigFileExists checks if a config file exists at the given path
func ConfigFileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// SaveConfig saves the configuration to a YAML file
func SaveConfig(path string, cfg *Config) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Write with appropriate permissions
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
