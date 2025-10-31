/*-------------------------------------------------------------------------
 *
 * pgEdge Postgres MCP Server
 *
 * Copyright (c) 2025, pgEdge, Inc.
 * This software is released under The PostgreSQL License
 *
 *-------------------------------------------------------------------------
 */

package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"pgedge-postgres-mcp/internal/auth"
	"pgedge-postgres-mcp/internal/config"
	"pgedge-postgres-mcp/internal/database"
	"pgedge-postgres-mcp/internal/llm"
	"pgedge-postgres-mcp/internal/mcp"
	"pgedge-postgres-mcp/internal/resources"
	"pgedge-postgres-mcp/internal/tools"
)

func main() {
	// Get executable path for default config location
	execPath, err := os.Executable()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Failed to get executable path: %v\n", err)
		os.Exit(1)
	}
	defaultConfigPath := config.GetDefaultConfigPath(execPath)

	// Command line flags
	configFile := flag.String("config", defaultConfigPath, "Path to configuration file")
	dbConnString := flag.String("db", "", "PostgreSQL connection string (overrides config file)")
	apiKey := flag.String("api-key", "", "Anthropic API key (overrides config file)")
	model := flag.String("model", "", "Anthropic model to use (overrides config file)")
	httpMode := flag.Bool("http", false, "Enable HTTP transport mode (default: stdio)")
	httpAddr := flag.String("addr", "", "HTTP server address")
	tlsMode := flag.Bool("tls", false, "Enable TLS/HTTPS (requires -http)")
	certFile := flag.String("cert", "", "Path to TLS certificate file")
	keyFile := flag.String("key", "", "Path to TLS key file")
	chainFile := flag.String("chain", "", "Path to TLS certificate chain file (optional)")
	noAuth := flag.Bool("no-auth", false, "Disable API token authentication in HTTP mode")
	tokenFilePath := flag.String("token-file", "", "Path to API token file")

	// Token management commands
	addTokenCmd := flag.Bool("add-token", false, "Add a new API token")
	removeTokenCmd := flag.String("remove-token", "", "Remove an API token by ID or hash prefix")
	listTokensCmd := flag.Bool("list-tokens", false, "List all API tokens")
	tokenNote := flag.String("token-note", "", "Annotation for the new token (used with -add-token)")
	tokenExpiry := flag.String("token-expiry", "", "Token expiry duration: '30d', '1y', '2w', '12h', 'never' (used with -add-token)")

	flag.Parse()

	// Handle token management commands
	if *addTokenCmd || *removeTokenCmd != "" || *listTokensCmd {
		defaultTokenPath := auth.GetDefaultTokenPath(execPath)
		tokenFile := *tokenFilePath
		if tokenFile == "" {
			tokenFile = defaultTokenPath
		}

		if *addTokenCmd {
			var expiry time.Duration
			switch {
			case *tokenExpiry != "" && *tokenExpiry != "never":
				var err error
				expiry, err = parseDuration(*tokenExpiry)
				if err != nil {
					fmt.Fprintf(os.Stderr, "ERROR: Invalid expiry duration: %v\n", err)
					os.Exit(1)
				}
			case *tokenExpiry == "":
				expiry = 0 // Will prompt user
			default:
				expiry = -1 // Never expires
			}

			if err := addTokenCommand(tokenFile, *tokenNote, expiry); err != nil {
				fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
				os.Exit(1)
			}
			return
		}

		if *removeTokenCmd != "" {
			if err := removeTokenCommand(tokenFile, *removeTokenCmd); err != nil {
				fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
				os.Exit(1)
			}
			return
		}

		if *listTokensCmd {
			if err := listTokensCommand(tokenFile); err != nil {
				fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
				os.Exit(1)
			}
			return
		}
	}

	// Track which flags were explicitly set
	cliFlags := config.CLIFlags{}
	flag.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "config":
			cliFlags.ConfigFileSet = true
			cliFlags.ConfigFile = *configFile
		case "db":
			cliFlags.ConnectionStringSet = true
			cliFlags.ConnectionString = *dbConnString
		case "api-key":
			cliFlags.APIKeySet = true
			cliFlags.APIKey = *apiKey
		case "model":
			cliFlags.ModelSet = true
			cliFlags.Model = *model
		case "http":
			cliFlags.HTTPEnabledSet = true
			cliFlags.HTTPEnabled = *httpMode
		case "addr":
			cliFlags.HTTPAddrSet = true
			cliFlags.HTTPAddr = *httpAddr
		case "tls":
			cliFlags.TLSEnabledSet = true
			cliFlags.TLSEnabled = *tlsMode
		case "cert":
			cliFlags.TLSCertSet = true
			cliFlags.TLSCertFile = *certFile
		case "key":
			cliFlags.TLSKeySet = true
			cliFlags.TLSKeyFile = *keyFile
		case "chain":
			cliFlags.TLSChainSet = true
			cliFlags.TLSChainFile = *chainFile
		case "no-auth":
			cliFlags.AuthEnabledSet = true
			cliFlags.AuthEnabled = !*noAuth // Invert because it's "no-auth"
		case "token-file":
			cliFlags.AuthTokenSet = true
			cliFlags.AuthTokenFile = *tokenFilePath
		}
	})

	// Validate basic flag dependencies before loading full config
	if !*httpMode && (*tlsMode || *certFile != "" || *keyFile != "" || *chainFile != "") {
		fmt.Fprintf(os.Stderr, "ERROR: TLS options (-tls, -cert, -key, -chain) require -http flag\n")
		flag.Usage()
		os.Exit(1)
	}

	// Determine which config file to load
	configPath := *configFile
	if !cliFlags.ConfigFileSet {
		// Check if default config exists
		if config.ConfigFileExists(defaultConfigPath) {
			configPath = defaultConfigPath
		} else {
			// No config file, rely on env vars and CLI flags
			configPath = ""
		}
	}

	// Load configuration
	cfg, err := config.LoadConfig(configPath, cliFlags)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(1)
	}

	// Set default token file path if not specified and HTTP is enabled
	if cfg.HTTP.Enabled && cfg.HTTP.Auth.TokenFile == "" {
		cfg.HTTP.Auth.TokenFile = auth.GetDefaultTokenPath(execPath)
	}

	// Set environment variables for clients that read them directly
	// This ensures backward compatibility
	if err := os.Setenv("POSTGRES_CONNECTION_STRING", cfg.Database.ConnectionString); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Failed to set environment variable: %v\n", err)
		os.Exit(1)
	}
	if cfg.Anthropic.APIKey != "" {
		if err := os.Setenv("ANTHROPIC_API_KEY", cfg.Anthropic.APIKey); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: Failed to set environment variable: %v\n", err)
			os.Exit(1)
		}
	}
	if cfg.Anthropic.Model != "" {
		if err := os.Setenv("ANTHROPIC_MODEL", cfg.Anthropic.Model); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: Failed to set environment variable: %v\n", err)
			os.Exit(1)
		}
	}

	// Verify TLS files exist if HTTPS is enabled
	if cfg.HTTP.TLS.Enabled {
		if _, err := os.Stat(cfg.HTTP.TLS.CertFile); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: Certificate file not found: %s\n", cfg.HTTP.TLS.CertFile)
			os.Exit(1)
		}
		if _, err := os.Stat(cfg.HTTP.TLS.KeyFile); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: Key file not found: %s\n", cfg.HTTP.TLS.KeyFile)
			os.Exit(1)
		}
		if cfg.HTTP.TLS.ChainFile != "" {
			if _, err := os.Stat(cfg.HTTP.TLS.ChainFile); err != nil {
				fmt.Fprintf(os.Stderr, "ERROR: Chain file not found: %s\n", cfg.HTTP.TLS.ChainFile)
				os.Exit(1)
			}
		}
	}

	// Load token store if HTTP auth is enabled
	var tokenStore *auth.TokenStore
	if cfg.HTTP.Enabled && cfg.HTTP.Auth.Enabled {
		if _, err := os.Stat(cfg.HTTP.Auth.TokenFile); os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "ERROR: Token file not found: %s\n", cfg.HTTP.Auth.TokenFile)
			fmt.Fprintf(os.Stderr, "Create tokens with: %s -add-token\n", os.Args[0])
			fmt.Fprintf(os.Stderr, "Or disable authentication with: -no-auth\n")
			os.Exit(1)
		}

		tokenStore, err = auth.LoadTokenStore(cfg.HTTP.Auth.TokenFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: Failed to load token file: %v\n", err)
			os.Exit(1)
		}

		fmt.Fprintf(os.Stderr, "Loaded %d API token(s) from %s\n", len(tokenStore.Tokens), cfg.HTTP.Auth.TokenFile)
	}

	// Create LLM client (shared across all modes)
	llmClient := llm.NewClient()

	// Register resources first (so they can be used by tools)
	resourceRegistry := resources.NewRegistry()

	var server *mcp.Server
	var clientManager *database.ClientManager

	// Choose tool provider based on mode
	if cfg.HTTP.Enabled && cfg.HTTP.Auth.Enabled {
		// HTTP mode with authentication: Use per-token connection isolation
		clientManager = database.NewClientManager()

		// Clean up expired tokens on startup (no connections exist yet)
		if removed, _ := tokenStore.CleanupExpiredTokens(); removed > 0 {
			fmt.Fprintf(os.Stderr, "Removed %d expired token(s)\n", removed)
			// Save the cleaned store
			if err := auth.SaveTokenStore(cfg.HTTP.Auth.TokenFile, tokenStore); err != nil {
				fmt.Fprintf(os.Stderr, "WARNING: Failed to save cleaned token file: %v\n", err)
			}
		}

		// Start periodic cleanup of expired tokens and their connections
		go func() {
			ticker := time.NewTicker(5 * time.Minute)
			defer ticker.Stop()
			for range ticker.C {
				if removed, hashes := tokenStore.CleanupExpiredTokens(); removed > 0 {
					fmt.Fprintf(os.Stderr, "Removed %d expired token(s)\n", removed)
					// Clean up database connections for expired tokens
					if err := clientManager.RemoveClients(hashes); err != nil {
						fmt.Fprintf(os.Stderr, "WARNING: Failed to cleanup connections: %v\n", err)
					}
					// Save the cleaned store
					if err := auth.SaveTokenStore(cfg.HTTP.Auth.TokenFile, tokenStore); err != nil {
						fmt.Fprintf(os.Stderr, "WARNING: Failed to save cleaned token file: %v\n", err)
					}
				}
			}
		}()

		// Create a fallback client for initialization (not used for actual requests)
		fallbackClient := database.NewClient()

		// Initialize fallback database connection in background
		go func() {
			if err := fallbackClient.Connect(); err != nil {
				fmt.Fprintf(os.Stderr, "ERROR: Failed to connect to database: %v\n", err)
				return
			}

			if err := fallbackClient.LoadMetadata(); err != nil {
				fmt.Fprintf(os.Stderr, "ERROR: Failed to load database metadata: %v\n", err)
				return
			}

			fmt.Fprintf(os.Stderr, "Database ready: %d tables/views loaded\n", len(fallbackClient.GetMetadata()))
		}()

		// Set up resources with fallback client
		resourceRegistry.Register(resources.URISettings, resources.PGSettingsResource(fallbackClient))
		resourceRegistry.Register(resources.URISystemInfo, resources.PGSystemInfoResource(fallbackClient))
		resourceRegistry.Register(resources.URIStatActivity, resources.PGStatActivityResource(fallbackClient))
		resourceRegistry.Register(resources.URIStatDatabase, resources.PGStatDatabaseResource(fallbackClient))
		resourceRegistry.Register(resources.URIStatUserTables, resources.PGStatUserTablesResource(fallbackClient))
		resourceRegistry.Register(resources.URIStatUserIndexes, resources.PGStatUserIndexesResource(fallbackClient))
		resourceRegistry.Register(resources.URIStatReplication, resources.PGStatReplicationResource(fallbackClient))
		resourceRegistry.Register(resources.URIStatBgwriter, resources.PGStatBgwriterResource(fallbackClient))
		resourceRegistry.Register(resources.URIStatWAL, resources.PGStatWALResource(fallbackClient))

		// Use context-aware provider for per-token connection isolation
		contextAwareProvider := tools.NewContextAwareProvider(clientManager, llmClient, resourceRegistry, true, fallbackClient)
		if err := contextAwareProvider.RegisterTools(nil); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: Failed to register tools: %v\n", err)
			os.Exit(1)
		}

		server = mcp.NewServer(contextAwareProvider)
		server.SetResourceProvider(resourceRegistry)

		fmt.Fprintf(os.Stderr, "Connection isolation: ENABLED (per-token database connections)\n")
	} else {
		// Stdio mode or HTTP without auth: Use shared database connection
		dbClient := database.NewClient()

		// Initialize database in background
		go func() {
			if err := dbClient.Connect(); err != nil {
				fmt.Fprintf(os.Stderr, "ERROR: Failed to connect to database: %v\n", err)
				return
			}

			if err := dbClient.LoadMetadata(); err != nil {
				fmt.Fprintf(os.Stderr, "ERROR: Failed to load database metadata: %v\n", err)
				return
			}

			fmt.Fprintf(os.Stderr, "Database ready: %d tables/views loaded\n", len(dbClient.GetMetadata()))
		}()

		// Set up resources with shared client
		resourceRegistry.Register(resources.URISettings, resources.PGSettingsResource(dbClient))
		resourceRegistry.Register(resources.URISystemInfo, resources.PGSystemInfoResource(dbClient))
		resourceRegistry.Register(resources.URIStatActivity, resources.PGStatActivityResource(dbClient))
		resourceRegistry.Register(resources.URIStatDatabase, resources.PGStatDatabaseResource(dbClient))
		resourceRegistry.Register(resources.URIStatUserTables, resources.PGStatUserTablesResource(dbClient))
		resourceRegistry.Register(resources.URIStatUserIndexes, resources.PGStatUserIndexesResource(dbClient))
		resourceRegistry.Register(resources.URIStatReplication, resources.PGStatReplicationResource(dbClient))
		resourceRegistry.Register(resources.URIStatBgwriter, resources.PGStatBgwriterResource(dbClient))
		resourceRegistry.Register(resources.URIStatWAL, resources.PGStatWALResource(dbClient))

		// Register tools with shared client
		toolRegistry := tools.NewRegistry()
		toolRegistry.Register("query_database", tools.QueryDatabaseTool(dbClient, llmClient))
		toolRegistry.Register("get_schema_info", tools.GetSchemaInfoTool(dbClient))
		toolRegistry.Register("set_pg_configuration", tools.SetPGConfigurationTool(dbClient))
		toolRegistry.Register("recommend_pg_configuration", tools.RecommendPGConfigurationTool())
		toolRegistry.Register("analyze_bloat", tools.AnalyzeBloatTool(dbClient))
		toolRegistry.Register("read_server_log", tools.ReadServerLogTool(dbClient))
		toolRegistry.Register("read_postgresql_conf", tools.ReadPostgresqlConfTool(dbClient))
		toolRegistry.Register("read_pg_hba_conf", tools.ReadPgHbaConfTool(dbClient))
		toolRegistry.Register("read_pg_ident_conf", tools.ReadPgIdentConfTool(dbClient))
		toolRegistry.Register("read_resource", tools.ReadResourceTool(resourceRegistry))

		server = mcp.NewServer(toolRegistry)
		server.SetResourceProvider(resourceRegistry)
	}

	if cfg.HTTP.Enabled {
		// HTTP/HTTPS mode
		httpConfig := &mcp.HTTPConfig{
			Addr:        cfg.HTTP.Address,
			TLSEnable:   cfg.HTTP.TLS.Enabled,
			CertFile:    cfg.HTTP.TLS.CertFile,
			KeyFile:     cfg.HTTP.TLS.KeyFile,
			ChainFile:   cfg.HTTP.TLS.ChainFile,
			AuthEnabled: cfg.HTTP.Auth.Enabled,
			TokenStore:  tokenStore,
		}

		if cfg.HTTP.TLS.Enabled {
			fmt.Fprintf(os.Stderr, "Starting MCP server in HTTPS mode on %s\n", cfg.HTTP.Address)
			fmt.Fprintf(os.Stderr, "Certificate: %s\n", cfg.HTTP.TLS.CertFile)
			fmt.Fprintf(os.Stderr, "Key: %s\n", cfg.HTTP.TLS.KeyFile)
			if cfg.HTTP.TLS.ChainFile != "" {
				fmt.Fprintf(os.Stderr, "Chain: %s\n", cfg.HTTP.TLS.ChainFile)
			}
		} else {
			fmt.Fprintf(os.Stderr, "Starting MCP server in HTTP mode on %s\n", cfg.HTTP.Address)
		}

		if cfg.HTTP.Auth.Enabled {
			fmt.Fprintf(os.Stderr, "Authentication: ENABLED\n")
		} else {
			fmt.Fprintf(os.Stderr, "Authentication: DISABLED (warning: server is not secured)\n")
		}

		err = server.RunHTTP(httpConfig)
	} else {
		// Default stdio mode
		err = server.Run()
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(1)
	}

	// Cleanup
	if clientManager != nil {
		// Close all per-token connections
		if err := clientManager.CloseAll(); err != nil {
			fmt.Fprintf(os.Stderr, "WARNING: Error closing database connections: %v\n", err)
		}
	}
}
