.PHONY: test test-verbose test-local test-container test-container-examples test-installation test-database test-service test-mcp test-kb test-examples config clean cleanup help

# Default configuration file
CONFIG_FILE ?= config/local.yaml
# Get absolute path to Makefile directory
MAKEFILE_DIR := $(shell pwd)
# Log file for detailed output
LOG_FILE ?= test-results/test-execution.log

# Test targets - Run all tests
test:
	@echo "Running MCP Server tests with $(CONFIG_FILE)..."
	@mkdir -p test-results
	@go clean -cache -testcache
	@touch testcases/*.go
	@LOG_LEVEL=$$(grep -A2 "^logging:" $(CONFIG_FILE) | grep "level:" | awk '{print $$2}' | tr -d '"' | tr -d "'"); \
	if [ "$$LOG_LEVEL" = "minimal" ]; then \
		echo "Logging level: minimal (terminal shows summary only, full logs in file)"; \
		TESTFW_CONFIG=$(MAKEFILE_DIR)/$(CONFIG_FILE) go test -v -count=1 -timeout 30m ./testcases 2>&1 | tee $(LOG_FILE) | sed -n '/🧪 Test Suite Summary/,/Total Duration:/p; /^ok[[:space:]]/p'; \
	else \
		echo "Logging level: $$LOG_LEVEL (full logs on terminal and file)"; \
		TESTFW_CONFIG=$(MAKEFILE_DIR)/$(CONFIG_FILE) go test -v -count=1 -timeout 30m ./testcases 2>&1 | tee $(LOG_FILE); \
	fi

test-verbose:
	@echo "Running MCP Server tests (verbose) with $(CONFIG_FILE)..."
	@mkdir -p test-results
	@go clean -cache -testcache
	@touch testcases/*.go
	TESTFW_CONFIG=$(MAKEFILE_DIR)/$(CONFIG_FILE) go test -v -count=1 -timeout 30m ./testcases 2>&1 | tee $(LOG_FILE)

test-local:
	@echo "Running tests in local mode (uses server_env from local.yaml)..."
	@go clean -cache -testcache
	$(MAKE) test CONFIG_FILE=config/local.yaml LOG_FILE=test-results/test-local.log

test-container:
	@echo "Running tests in container mode (uses server_env from container.yaml)..."
	@echo "Note: Single package = single container (fast!)"
	@echo "CONFIG_FILE=config/container.yaml LOG_FILE=test-results/test-container.log"
	@echo "Checking logging configuration..."
	@mkdir -p test-results
	@go clean -cache -testcache
	@touch testcases/*.go
	@LOG_LEVEL=$$(grep -A2 "^logging:" config/container.yaml | grep "level:" | awk '{print $$2}' | tr -d '"' | tr -d "'"); \
	if [ "$$LOG_LEVEL" = "minimal" ]; then \
		echo "Logging level: minimal (terminal shows summary only, full logs in file)"; \
		TESTFW_CONFIG=$(MAKEFILE_DIR)/config/container.yaml go test -v -count=1 -timeout 30m ./testcases 2>&1 | tee test-results/test-container.log | sed -n '/🧪 Test Suite Summary/,/Total Duration:/p; /^ok[[:space:]]/p'; \
	else \
		echo "Logging level: $$LOG_LEVEL (full logs on terminal and file)"; \
		TESTFW_CONFIG=$(MAKEFILE_DIR)/config/container.yaml go test -v -count=1 -timeout 30m ./testcases 2>&1 | tee test-results/test-container.log; \
	fi

test-container-examples:
	@echo "Running example tests in container mode..."
	TESTFW_CONFIG=$(MAKEFILE_DIR)/config/container.yaml go test -v ./testcases/examples/

# Run tests by category
test-installation:
	@echo "Running installation tests..."
	TESTFW_CONFIG=$(MAKEFILE_DIR)/$(CONFIG_FILE) go test -v ./testcases/installation/

test-database:
	@echo "Running database tests..."
	TESTFW_CONFIG=$(MAKEFILE_DIR)/$(CONFIG_FILE) go test -v ./testcases/database/

test-service:
	@echo "Running service tests..."
	TESTFW_CONFIG=$(MAKEFILE_DIR)/$(CONFIG_FILE) go test -v ./testcases/service/

test-mcp:
	@echo "Running MCP protocol tests..."
	TESTFW_CONFIG=$(MAKEFILE_DIR)/$(CONFIG_FILE) go test -v ./testcases/mcp/

test-kb:
	@echo "Running knowledge base tests..."
	TESTFW_CONFIG=$(MAKEFILE_DIR)/$(CONFIG_FILE) go test -v ./testcases/kb/

test-examples:
	@echo "Running example tests..."
	TESTFW_CONFIG=$(MAKEFILE_DIR)/$(CONFIG_FILE) go test -v ./testcases/examples/

# Run specific test
test-run:
	@echo "Running specific test: $(TEST)..."
	TESTFW_CONFIG=$(MAKEFILE_DIR)/$(CONFIG_FILE) go test -v ./testcases/... -run $(TEST)

# Run tests sequentially by category
test-sequential:
	@echo "Running all test categories sequentially..."
	@for category in installation database service mcp kb examples; do \
		echo ""; \
		echo "==========================================="; \
		echo "Running $$category tests..."; \
		echo "==========================================="; \
		TESTFW_CONFIG=$(MAKEFILE_DIR)/$(CONFIG_FILE) go test -v ./testcases/$$category/ || exit 1; \
	done

# Show current configuration
config:
	@echo "=== Current Test Configuration ==="
	@echo "Config file: $(CONFIG_FILE)"
	@echo ""
	@cat $(CONFIG_FILE)

# List all available tests
list-tests:
	@echo "Available test categories:"
	@echo ""
	@for dir in testcases/*/; do \
		category=$$(basename $$dir); \
		echo "📁 $$category/"; \
		find $$dir -name "*_test.go" -exec basename {} \; | sed 's/^/  - /'; \
		echo ""; \
	done

# Clean test cache and results
clean:
	@echo "Cleaning test cache and results..."
	go clean -testcache
	rm -rf test-results/

# Full cleanup including packages and services
cleanup:
	@echo "Running full cleanup script..."
	@bash ../common/utils/cleanup.sh

# Tidy dependencies
tidy:
	@echo "Tidying Go modules..."
	go mod tidy

# Format code
fmt:
	@echo "Formatting Go code..."
	go fmt ./...

# Run linter (if golangci-lint is installed)
lint:
	@if command -v golangci-lint >/dev/null 2>&1; then \
		echo "Running linter..."; \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed, skipping..."; \
	fi

# Help
help:
	@echo "MCP Server Test Makefile"
	@echo ""
	@echo "Main test targets:"
	@echo "  test-local          - Run tests locally (edit local.yaml to switch live/staging)"
	@echo "  test-container      - Run tests in container (edit container.yaml to switch live/staging)"
	@echo ""
	@echo "General test targets:"
	@echo "  test                - Run all tests (uses CONFIG_FILE, default: local.yaml)"
	@echo "  test-verbose        - Run all tests with verbose output"
	@echo "  test-container-examples - Run example tests in container mode"
	@echo ""
	@echo "Test by category:"
	@echo "  test-installation   - Run installation tests"
	@echo "  test-database       - Run database tests"
	@echo "  test-service        - Run service management tests"
	@echo "  test-mcp            - Run MCP protocol tests"
	@echo "  test-kb             - Run knowledge base tests"
	@echo "  test-examples       - Run example tests"
	@echo ""
	@echo "Other targets:"
	@echo "  test-run TEST=<name> - Run specific test"
	@echo "  test-sequential     - Run all categories sequentially"
	@echo "  list-tests          - List all available tests"
	@echo "  config              - Display current configuration"
	@echo "  clean               - Clean test cache and results"
	@echo "  cleanup             - Full cleanup (packages, services, data, containers)"
	@echo "  tidy                - Tidy Go module dependencies"
	@echo "  fmt                 - Format Go code"
	@echo "  lint                - Run linter (if available)"
	@echo "  help                - Show this help message"
	@echo ""
	@echo "Environment variables:"
	@echo "  CONFIG_FILE         - Configuration file to use (default: config/local.yaml)"
	@echo ""
	@echo "Examples:"
	@echo "  make test-local                                  # Test locally"
	@echo "  make test-container                              # Test in container"
	@echo "  make test-verbose"
	@echo "  make test-installation"
	@echo "  make test-run TEST=Test01_RepositoryInstallation"
	@echo "  make list-tests"
	@echo ""
	@echo "To switch between live and staging repositories:"
	@echo "  For local tests:     Edit config/local.yaml and change 'server_env: live' to 'server_env: staging'"
	@echo "  For container tests: Edit config/container.yaml and change 'server_env: live' to 'server_env: staging'"
