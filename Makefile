.PHONY: build clean test run install help

# Binary name
BINARY_NAME=pgedge-mcp

# Build variables
GO=go
GOFLAGS=-v

# Default target
all: build

# Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	$(GO) build $(GOFLAGS) -o $(BINARY_NAME)
	@echo "Build complete: $(BINARY_NAME)"

# Build for multiple platforms
build-all: build-linux build-darwin build-windows

build-linux:
	@echo "Building for Linux..."
	GOOS=linux GOARCH=amd64 $(GO) build $(GOFLAGS) -o $(BINARY_NAME)-linux-amd64
	@echo "Linux build complete: $(BINARY_NAME)-linux-amd64"

build-darwin:
	@echo "Building for macOS..."
	GOOS=darwin GOARCH=amd64 $(GO) build $(GOFLAGS) -o $(BINARY_NAME)-darwin-amd64
	GOOS=darwin GOARCH=arm64 $(GO) build $(GOFLAGS) -o $(BINARY_NAME)-darwin-arm64
	@echo "macOS builds complete: $(BINARY_NAME)-darwin-{amd64,arm64}"

build-windows:
	@echo "Building for Windows..."
	GOOS=windows GOARCH=amd64 $(GO) build $(GOFLAGS) -o $(BINARY_NAME)-windows-amd64.exe
	@echo "Windows build complete: $(BINARY_NAME)-windows-amd64.exe"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_NAME)-*
	$(GO) clean
	@echo "Clean complete"

# Run tests
test:
	@echo "Running tests..."
	$(GO) test -v ./...

# Run with example environment (requires .env file)
run:
	@if [ ! -f .env ]; then \
		echo "Error: .env file not found. Copy .env.example to .env and configure it."; \
		exit 1; \
	fi
	@echo "Running $(BINARY_NAME)..."
	@export $$(cat .env | xargs) && ./$(BINARY_NAME)

# Install dependencies
deps:
	@echo "Installing dependencies..."
	$(GO) mod download
	$(GO) mod tidy
	@echo "Dependencies installed"

# Install the binary to GOPATH/bin
install: build
	@echo "Installing $(BINARY_NAME) to $(GOPATH)/bin..."
	$(GO) install
	@echo "Install complete"

# Format code
fmt:
	@echo "Formatting code..."
	$(GO) fmt ./...
	@echo "Format complete"

# Run linter (requires golangci-lint)
lint:
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not found. Install it with:"; \
		echo "  brew install golangci-lint  # macOS"; \
		echo "  or visit https://golangci-lint.run/usage/install/"; \
	fi

# Show help
help:
	@echo "pgEdge MCP Server - Makefile commands:"
	@echo ""
	@echo "  make build         - Build the binary"
	@echo "  make build-all     - Build for all platforms"
	@echo "  make build-linux   - Build for Linux (amd64)"
	@echo "  make build-darwin  - Build for macOS (amd64 and arm64)"
	@echo "  make build-windows - Build for Windows (amd64)"
	@echo "  make clean         - Remove build artifacts"
	@echo "  make test          - Run tests"
	@echo "  make run           - Run with environment from .env file"
	@echo "  make deps          - Install/update dependencies"
	@echo "  make install       - Install binary to GOPATH/bin"
	@echo "  make fmt           - Format Go code"
	@echo "  make lint          - Run linter (requires golangci-lint)"
	@echo "  make help          - Show this help message"
	@echo ""
