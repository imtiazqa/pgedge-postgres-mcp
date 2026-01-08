package config

import (
	"fmt"
	"strings"
)

// Validator validates configuration
type Validator struct{}

// NewValidator creates a new configuration validator
func NewValidator() *Validator {
	return &Validator{}
}

// Validate validates the configuration
func (v *Validator) Validate(config *TestConfig) error {
	var errors []string

	// Validate execution config
	if config.Execution.Mode == "" {
		errors = append(errors, "execution.mode is required")
	} else {
		validModes := []string{"local", "container-systemd"}
		if !contains(validModes, config.Execution.Mode) {
			errors = append(errors, fmt.Sprintf("execution.mode must be one of: %s", strings.Join(validModes, ", ")))
		}
	}

	// Validate database config (if specified)
	if config.Database.Host != "" {
		if config.Database.Port <= 0 || config.Database.Port > 65535 {
			errors = append(errors, "database.port must be between 1 and 65535")
		}
	}

	// Validate logging config
	if config.Logging.Level != "" {
		validLevels := []string{"minimal", "detailed", "verbose"}
		if !contains(validLevels, config.Logging.Level) {
			errors = append(errors, fmt.Sprintf("logging.level must be one of: %s", strings.Join(validLevels, ", ")))
		}
	}

	// Validate timeouts (ensure they're positive)
	if config.Timeouts.Default > 0 && config.Timeouts.Test > config.Timeouts.Suite {
		errors = append(errors, "timeouts.test should not exceed timeouts.suite")
	}

	// If there are validation errors, return them
	if len(errors) > 0 {
		return fmt.Errorf("validation errors:\n  - %s", strings.Join(errors, "\n  - "))
	}

	return nil
}

// SetDefaults sets default values for missing configuration
func (v *Validator) SetDefaults(config *TestConfig) {
	// Set default environment
	if config.Environment == "" {
		config.Environment = "dev"
	}

	// Set default execution config
	if config.Execution.Mode == "" {
		config.Execution.Mode = "local"
	}
	if config.Execution.MaxWorkers == 0 {
		config.Execution.MaxWorkers = 4
	}

	// Set default database config
	if config.Database.Port == 0 {
		config.Database.Port = 5432
	}
	if config.Database.MaxConnections == 0 {
		config.Database.MaxConnections = 10
	}

	// Set default logging config
	if config.Logging.Level == "" {
		config.Logging.Level = "detailed"
	}

	// Set default reporting config
	if !config.Reporting.Console && !config.Reporting.JSON && !config.Reporting.JUnit {
		config.Reporting.Console = true
	}

	// Set default timeouts (in seconds)
	if config.Timeouts.Default == 0 {
		config.Timeouts.Default = 30
	}
	if config.Timeouts.Suite == 0 {
		config.Timeouts.Suite = 1800 // 30 minutes
	}
	if config.Timeouts.Test == 0 {
		config.Timeouts.Test = 300 // 5 minutes
	}
	if config.Timeouts.Command == 0 {
		config.Timeouts.Command = 120 // 2 minutes
	}

	// Set default retry config
	if config.Retry.DefaultAttempts == 0 {
		config.Retry.DefaultAttempts = 3
	}
	if config.Retry.DefaultDelay == 0 {
		config.Retry.DefaultDelay = 1
	}
}

// Helper function to check if string is in slice
func contains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}
