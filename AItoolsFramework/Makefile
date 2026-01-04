.PHONY: test test-all test-mcp-server clean help

# Test all projects
test-all:
	@echo "Running tests for all projects..."
	$(MAKE) test-mcp-server

# Test individual projects
test-mcp-server:
	@echo "Testing MCP Server..."
	cd mcp-server && $(MAKE) test

test-mcp-server-verbose:
	@echo "Testing MCP Server (verbose)..."
	cd mcp-server && $(MAKE) test-verbose

test-mcp-server-staging:
	@echo "Testing MCP Server (staging)..."
	cd mcp-server && $(MAKE) test-staging

# Clean all projects
clean:
	@echo "Cleaning all projects..."
	cd mcp-server && $(MAKE) clean
	go clean -testcache

# Tidy all Go modules
tidy:
	@echo "Tidying root module..."
	go mod tidy
	@echo "Tidying mcp-server module..."
	cd mcp-server && go mod tidy

# Format all code
fmt:
	@echo "Formatting all Go code..."
	go fmt ./...
	cd mcp-server && go fmt ./...

# Help
help:
	@echo "AItoolsFramework - Multi-Project Test Framework"
	@echo ""
	@echo "Available targets:"
	@echo "  test-all              - Run tests for all projects"
	@echo "  test-mcp-server       - Run MCP server tests"
	@echo "  test-mcp-server-verbose - Run MCP server tests (verbose)"
	@echo "  test-mcp-server-staging - Run MCP server staging tests"
	@echo "  clean                 - Clean all test caches"
	@echo "  tidy                  - Tidy all Go modules"
	@echo "  fmt                   - Format all Go code"
	@echo "  help                  - Show this help message"
	@echo ""
	@echo "Project structure:"
	@echo "  /common          - Shared framework utilities"
	@echo "  /mcp-server      - MCP Server tests"
	@echo ""
	@echo "Quick start:"
	@echo "  make test-mcp-server"
	@echo "  cd mcp-server && make test-verbose"
