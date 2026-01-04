package config

import (
	"context"
	"fmt"
	"os"
	"regexp"

	"gopkg.in/yaml.v3"
)

// Loader loads configuration from files and environment
type Loader struct {
	validator *Validator
}

// NewLoader creates a new configuration loader
func NewLoader() *Loader {
	return &Loader{
		validator: NewValidator(),
	}
}

// Load loads configuration from file with env overrides
func (l *Loader) Load(ctx context.Context, path string) (*TestConfig, error) {
	// Read the file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Expand environment variables
	expandedData := l.expandEnvVars(string(data))

	// Parse YAML
	var config TestConfig
	if err := yaml.Unmarshal([]byte(expandedData), &config); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Validate configuration
	if err := l.validator.Validate(&config); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &config, nil
}

// LoadWithOverrides loads with explicit overrides
func (l *Loader) LoadWithOverrides(ctx context.Context, path string, overrides map[string]interface{}) (*TestConfig, error) {
	config, err := l.Load(ctx, path)
	if err != nil {
		return nil, err
	}

	// Apply overrides (simplified - can be enhanced)
	// This is a basic implementation
	return config, nil
}

// expandEnvVars expands environment variables in the config
// Supports ${VAR} and ${VAR:-default} syntax
func (l *Loader) expandEnvVars(content string) string {
	// Regex to match ${VAR} or ${VAR:-default}
	re := regexp.MustCompile(`\$\{([^}:]+)(?::-([^}]+))?\}`)

	return re.ReplaceAllStringFunc(content, func(match string) string {
		// Extract variable name and default value
		parts := re.FindStringSubmatch(match)
		if len(parts) < 2 {
			return match
		}

		varName := parts[1]
		defaultValue := ""
		if len(parts) > 2 {
			defaultValue = parts[2]
		}

		// Get environment variable
		if value := os.Getenv(varName); value != "" {
			return value
		}

		// Return default if provided
		if defaultValue != "" {
			return defaultValue
		}

		// Return empty string if no default
		return ""
	})
}

// GetConfigPath returns the config file path from environment or default
func GetConfigPath() string {
	if path := os.Getenv("TESTFW_CONFIG"); path != "" {
		// If the path doesn't exist, try searching upwards
		if _, err := os.Stat(path); err == nil {
			return path
		}
		// Try one directory up (for when running from suites/)
		upPath := "../" + path
		if _, err := os.Stat(upPath); err == nil {
			return upPath
		}
		// Return original path and let it fail with clear error
		return path
	}

	// Default path with search
	defaultPath := "config/dev.yaml"
	if _, err := os.Stat(defaultPath); err == nil {
		return defaultPath
	}
	// Try one directory up
	upPath := "../config/dev.yaml"
	if _, err := os.Stat(upPath); err == nil {
		return upPath
	}
	return defaultPath
}

// MustLoad loads configuration and panics on error (for testing)
func MustLoad(path string) *TestConfig {
	loader := NewLoader()
	config, err := loader.Load(context.Background(), path)
	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}
	return config
}
