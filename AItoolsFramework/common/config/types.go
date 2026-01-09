package config

import (
	"time"
)

// TestConfig is the root configuration for the test framework
type TestConfig struct {
	// Environment name (dev, staging, prod)
	Environment string `yaml:"environment"`

	// Project information
	Project ProjectInfo `yaml:"project"`

	// Execution configuration
	Execution ExecutionConfig `yaml:"execution"`

	// Database configuration
	Database DatabaseConfig `yaml:"database"`

	// PostgreSQL configuration
	PostgreSQL PostgreSQLConfig `yaml:"postgresql"`

	// Repository configuration
	Repository RepositoryConfig `yaml:"repository"`

	// Binary paths
	Binaries BinariesConfig `yaml:"binaries"`

	// Configuration directory
	ConfigDir string `yaml:"config_dir"`

	// Package names
	Packages PackagesConfig `yaml:"packages"`

	// Reporting configuration
	Reporting ReportingConfig `yaml:"reporting"`

	// Timeout configuration
	Timeouts TimeoutConfig `yaml:"timeouts"`

	// Retry configuration
	Retry RetryConfig `yaml:"retry"`

	// Fixtures configuration
	Fixtures FixturesConfig `yaml:"fixtures"`

	// Logging configuration
	Logging LoggingConfig `yaml:"logging"`

	// Service name (MCP server systemd service)
	ServiceName string `yaml:"service_name"`

	// Server configuration
	Server ServerConfig `yaml:"server"`

	// Log directory
	LogDir string `yaml:"log_dir"`
}

// BinariesConfig defines binary paths
type BinariesConfig struct {
	MCPServer string `yaml:"mcp_server"`
	KBBuilder string `yaml:"kb_builder"`
	CLI       string `yaml:"cli"`
}

// PackagesConfig defines package names for installation
type PackagesConfig struct {
	MCPServer []string `yaml:"mcp_server"`
	CLI       []string `yaml:"cli"`
	Web       []string `yaml:"web"`
	KB        []string `yaml:"kb"`
}

// ProjectInfo contains project metadata
type ProjectInfo struct {
	Name        string `yaml:"name"`
	Version     string `yaml:"version"`
	Description string `yaml:"description"`
}

// ContainerConfig defines container execution settings
type ContainerConfig struct {
	// Docker image to use
	OSImage string `yaml:"os_image"`

	// Enable systemd in container
	UseSystemd bool `yaml:"use_systemd"`

	// Skip sudo check (containers often run as root)
	SkipSudoCheck bool `yaml:"skip_sudo_check"`
}

// ExecutionConfig defines how tests execute
type ExecutionConfig struct {
	// Mode: local, container-systemd
	Mode string `yaml:"mode"`

	// Container configuration (for container modes)
	Container ContainerConfig `yaml:"container"`

	// Container image (for container modes) - legacy, use Container.OSImage
	OSImage string `yaml:"os_image"`

	// PostgreSQL version
	PGVersion string `yaml:"pg_version"`

	// Server environment (live, staging)
	ServerEnv string `yaml:"server_env"`

	// Parallel execution
	Parallel   bool `yaml:"parallel"`
	MaxWorkers int  `yaml:"max_workers"`

	// Skip sudo check for local execution (useful for simple tests)
	SkipSudoCheck bool `yaml:"skip_sudo_check"`
}

// DatabaseConfig defines database test settings
type DatabaseConfig struct {
	// Connection details
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Database string `yaml:"database"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	SSLMode  string `yaml:"ssl_mode"`

	// Test database settings
	CreateTestDB  bool   `yaml:"create_test_db"`
	TestDBPrefix  string `yaml:"test_db_prefix"`
	DropAfterTest bool   `yaml:"drop_after_test"`

	// Pooling
	MaxConnections    int    `yaml:"max_connections"`
	IdleTimeout       string `yaml:"idle_timeout"`
	ConnectionTimeout string `yaml:"connection_timeout"`
}

// PostgreSQLConfig defines PostgreSQL installation settings
type PostgreSQLConfig struct {
	// PostgreSQL version to install (16, 17, 18)
	Version string `yaml:"version"`
}

// RepositoryConfig defines repository URLs for different environments
type RepositoryConfig struct {
	// Debian/Ubuntu repository configuration
	Debian DebianRepoConfig `yaml:"debian"`

	// RHEL/Rocky/Alma repository configuration
	RHEL RHELRepoConfig `yaml:"rhel"`
}

// DebianRepoConfig defines Debian/Ubuntu repository settings
type DebianRepoConfig struct {
	// Base URL for apt repository
	BaseURL string `yaml:"base_url"`

	// Release package URL (will be auto-selected based on server_env if not provided)
	ReleasePackageURL string `yaml:"release_package_url"`

	// Live repository release package URL
	LiveReleaseURL string `yaml:"live_release_url"`

	// Staging repository release package URL
	StagingReleaseURL string `yaml:"staging_release_url"`
}

// RHELRepoConfig defines RHEL/Rocky/Alma repository settings
type RHELRepoConfig struct {
	// Base URL for DNF repository
	BaseURL string `yaml:"base_url"`

	// Release package URL (will be auto-selected based on server_env if not provided)
	ReleasePackageURL string `yaml:"release_package_url"`

	// Live repository release package URL
	LiveReleaseURL string `yaml:"live_release_url"`

	// Staging repository release package URL
	StagingReleaseURL string `yaml:"staging_release_url"`
}

// ReportingConfig defines test reporting settings
type ReportingConfig struct {
	// Output formats
	Console  bool `yaml:"console"`
	JSON     bool `yaml:"json"`
	JUnit    bool `yaml:"junit"`
	Markdown bool `yaml:"markdown"`

	// Output paths
	OutputPaths OutputPathsConfig `yaml:"output_paths"`

	// Console settings
	ConsoleSettings ConsoleSettingsConfig `yaml:"console_settings"`
}

// OutputPathsConfig defines output file paths
type OutputPathsConfig struct {
	JSON     string `yaml:"json"`
	JUnit    string `yaml:"junit"`
	Markdown string `yaml:"markdown"`
	LogFile  string `yaml:"log_file"` // Detailed test execution log
}

// ConsoleSettingsConfig defines console output settings
type ConsoleSettingsConfig struct {
	ColorOutput     bool `yaml:"color_output"`
	ShowTimings     bool `yaml:"show_timings"`
	ShowElephant    bool `yaml:"show_elephant"`
	VerboseFailures bool `yaml:"verbose_failures"`
}

// TimeoutConfig defines timeout settings
type TimeoutConfig struct {
	Default        time.Duration `yaml:"default"`
	Suite          time.Duration `yaml:"suite"`
	Test           time.Duration `yaml:"test"`
	Command        time.Duration `yaml:"command"`
	DatabaseQuery  time.Duration `yaml:"database_query"`
	HTTPRequest    time.Duration `yaml:"http_request"`
	ServiceStartup time.Duration `yaml:"service_startup"`
	PackageInstall time.Duration `yaml:"package_install"`
}

// RetryConfig defines retry settings
type RetryConfig struct {
	DefaultAttempts    int           `yaml:"default_attempts"`
	DefaultDelay       time.Duration `yaml:"default_delay"`
	ExponentialBackoff bool          `yaml:"exponential_backoff"`
	MaxDelay           time.Duration `yaml:"max_delay"`
}

// FixturesConfig defines test fixture settings
type FixturesConfig struct {
	Database DatabaseFixturesConfig `yaml:"database"`
	Configs  ConfigFixturesConfig   `yaml:"configs"`
	Responses ResponseFixturesConfig `yaml:"responses"`
}

// DatabaseFixturesConfig defines database fixture settings
type DatabaseFixturesConfig struct {
	SchemaFile   string   `yaml:"schema_file"`
	TestdataFile string   `yaml:"testdata_file"`
	CleanupOrder []string `yaml:"cleanup_order"`
}

// ConfigFixturesConfig defines config fixture settings
type ConfigFixturesConfig struct {
	ServerConfig string `yaml:"server_config"`
}

// ResponseFixturesConfig defines response fixture settings
type ResponseFixturesConfig struct {
	ToolsResponse     string `yaml:"tools_response"`
	ResourcesResponse string `yaml:"resources_response"`
}

// LoggingConfig defines logging settings
type LoggingConfig struct {
	Level       string `yaml:"level"`       // minimal, detailed, verbose
	LogCommands bool   `yaml:"log_commands"`
	LogOutput   bool   `yaml:"log_output"`
}

// ServerConfig defines MCP server settings
type ServerConfig struct {
	Port string `yaml:"port"` // Server port (e.g., "8080")
}

// GetFixturePath is a helper to get fixture paths (can be enhanced)
func (c *TestConfig) GetFixturePath(key string) (string, error) {
	switch key {
	case "database.schema_file":
		return c.Fixtures.Database.SchemaFile, nil
	case "database.testdata_file":
		return c.Fixtures.Database.TestdataFile, nil
	default:
		return "", nil
	}
}

// GetAPIBaseURL returns the API base URL from config (stub for now)
func (c *TestConfig) GetAPIBaseURL() string {
	// This will be enhanced when we add API config
	return ""
}

// GetBinaryPath returns the binary path for a binary name
func (c *TestConfig) GetBinaryPath(binaryName string) string {
	switch binaryName {
	case "mcp_server":
		return c.Binaries.MCPServer
	case "kb_builder":
		return c.Binaries.KBBuilder
	case "cli":
		return c.Binaries.CLI
	default:
		return ""
	}
}

// GetConfigDir returns the configuration directory
func (c *TestConfig) GetConfigDir() string {
	return c.ConfigDir
}
