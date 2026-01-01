# Makefile for touchlog
# Provides targets to build with and without version injection

# Variables
BINARY_NAME := touchlog
CMD_PATH := ./cmd/touchlog
MODULE_PATH := github.com/sv4u/touchlog
VERSION_PKG := $(MODULE_PATH)/internal/version

# Default target
.DEFAULT_GOAL := build

# Build without version injection (development build)
# This is useful for local development where you don't need version info
.PHONY: build
build:
	@echo "Building $(BINARY_NAME) (development build, no version injection)..."
	@go build -o $(BINARY_NAME) $(CMD_PATH)
	@echo "Build complete: $(BINARY_NAME)"

# Build with version injection (release build)
# Injects version, git commit, and build date via -ldflags
.PHONY: build-release
build-release:
	@echo "Building $(BINARY_NAME) with version injection..."
	@VERSION=$$(git describe --tags --always 2>/dev/null || echo "dev"); \
	COMMIT=$$(git rev-parse --short HEAD 2>/dev/null || echo "unknown"); \
	DATE=$$(date -u +%Y-%m-%dT%H:%M:%SZ); \
	go build -ldflags "-X $(VERSION_PKG).Version=$$VERSION -X $(VERSION_PKG).GitCommit=$$COMMIT -X $(VERSION_PKG).BuildDate=$$DATE" \
		-o $(BINARY_NAME) $(CMD_PATH)
	@echo "Release build complete: $(BINARY_NAME)"
	@echo "Version: $$(git describe --tags --always 2>/dev/null || echo 'dev')"

# Install the binary to $GOPATH/bin or $GOBIN
.PHONY: install
install:
	@echo "Installing $(BINARY_NAME)..."
	@go install $(CMD_PATH)
	@echo "Installation complete"

# Install with version injection
.PHONY: install-release
install-release:
	@echo "Installing $(BINARY_NAME) with version injection..."
	@VERSION=$$(git describe --tags --always 2>/dev/null || echo "dev"); \
	COMMIT=$$(git rev-parse --short HEAD 2>/dev/null || echo "unknown"); \
	DATE=$$(date -u +%Y-%m-%dT%H:%M:%SZ); \
	go install -ldflags "-X $(VERSION_PKG).Version=$$VERSION -X $(VERSION_PKG).GitCommit=$$COMMIT -X $(VERSION_PKG).BuildDate=$$DATE" \
		$(CMD_PATH)
	@echo "Release installation complete"

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	@go test -v ./...

# Run tests with coverage
.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	@rm -f $(BINARY_NAME)
	@rm -f coverage.out coverage.html
	@echo "Clean complete"

# Format code
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@echo "Formatting complete"

# Run linter (requires golangci-lint)
.PHONY: lint
lint:
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not found, skipping..."; \
		go vet ./...; \
	fi

# Run go vet
.PHONY: vet
vet:
	@echo "Running go vet..."
	@go vet ./...

# Show help
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build          - Build without version injection (development build)"
	@echo "  build-release  - Build with version injection (release build)"
	@echo "  install        - Install binary without version injection"
	@echo "  install-release - Install binary with version injection"
	@echo "  test           - Run tests"
	@echo "  test-coverage  - Run tests with coverage report"
	@echo "  clean          - Remove build artifacts"
	@echo "  fmt            - Format code"
	@echo "  lint           - Run linter (golangci-lint if available)"
	@echo "  vet            - Run go vet"
	@echo "  help           - Show this help message"

