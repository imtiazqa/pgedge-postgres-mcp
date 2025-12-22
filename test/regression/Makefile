.PHONY: test test-debian test-rocky test-ubuntu test-all clean help

# Default test
test: test-debian

# Test on Debian 12
test-debian:
	@echo "=== Testing on Debian 12 ==="
	TEST_OS_IMAGE=debian:12 go test -v -timeout 20m

# Test on Rocky Linux 9
test-rocky:
	@echo "=== Testing on Rocky Linux 9 ==="
	TEST_OS_IMAGE=rockylinux:9 go test -v -timeout 20m

# Test on Ubuntu 22.04
test-ubuntu:
	@echo "=== Testing on Ubuntu 22.04 ==="
	TEST_OS_IMAGE=ubuntu:22.04 go test -v -timeout 20m

# Test on all platforms
test-all: test-debian test-rocky test-ubuntu

# Run specific test
test-one:
	@echo "=== Running specific test ==="
	@echo "Usage: make test-one TEST=Test01_RepositoryInstallation"
	TEST_OS_IMAGE=debian:12 go test -v -run $(TEST) -timeout 10m

# Clean up any orphaned containers
clean:
	@echo "Cleaning up test containers..."
	docker ps -a | grep mcp-test | awk '{print $$1}' | xargs -r docker rm -f || true

# Show help
help:
	@echo "Available targets:"
	@echo "  make test         - Run tests on Debian 12 (default)"
	@echo "  make test-debian  - Run tests on Debian 12"
	@echo "  make test-rocky   - Run tests on Rocky Linux 9"
	@echo "  make test-ubuntu  - Run tests on Ubuntu 22.04"
	@echo "  make test-all     - Run tests on all platforms"
	@echo "  make test-one     - Run specific test (set TEST=TestName)"
	@echo "  make clean        - Clean up orphaned containers"
	@echo ""
	@echo "Environment variables:"
	@echo "  TEST_OS_IMAGE     - Docker image to test (e.g., debian:12)"
	@echo "  PGEDGE_REPO_URL   - Repository URL for packages"
