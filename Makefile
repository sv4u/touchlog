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

# Cross-platform build (platform agnostic)
# Accepts GOOS and GOARCH as variables (e.g., make build-cross GOOS=linux GOARCH=amd64)
# Uses CGO_ENABLED=0 for static builds
.PHONY: build-cross
build-cross:
	@echo "Building $(BINARY_NAME) for $(GOOS)/$(GOARCH)..."
	@if [ -z "$(GOOS)" ] || [ -z "$(GOARCH)" ]; then \
		echo "Error: GOOS and GOARCH must be set (e.g., make build-cross GOOS=linux GOARCH=amd64)"; \
		exit 1; \
	fi
	@CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -v -o $(BINARY_NAME) $(CMD_PATH)
	@echo "Cross-platform build complete: $(BINARY_NAME) ($(GOOS)/$(GOARCH))"

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

# Run tests with coverage (matches CI requirements)
.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test ./... -v -coverprofile=coverage.out -covermode=atomic
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Check coverage threshold
# Accepts THRESHOLD as variable (default: 20)
# Example: make check-coverage THRESHOLD=30
.PHONY: check-coverage
check-coverage:
	@echo "Checking coverage threshold..."
	@if [ ! -f coverage.out ]; then \
		echo "Error: coverage.out not found. Run 'make test-coverage' first."; \
		exit 1; \
	fi
	@THRESHOLD=$${THRESHOLD:-20}; \
	COVERAGE=$$(go tool cover -func=coverage.out | grep total | awk '{print $$3}' | sed 's/%//'); \
	if awk "BEGIN {if ($$COVERAGE < $$THRESHOLD) exit 0; else exit 1}"; then \
		echo "Coverage $$COVERAGE% is below $$THRESHOLD% threshold"; \
		exit 1; \
	fi; \
	echo "Coverage: $$COVERAGE%"

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

# Check code formatting (fails if files need formatting)
.PHONY: fmt-check
fmt-check:
	@echo "Checking code formatting..."
	@UNFORMATTED=$$(gofmt -l .); \
	if [ -n "$$UNFORMATTED" ]; then \
		echo "Unformatted files:"; \
		echo "$$UNFORMATTED"; \
		exit 1; \
	fi
	@echo "All files are properly formatted"

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

# Run staticcheck
.PHONY: staticcheck
staticcheck:
	@echo "Running staticcheck..."
	@if ! command -v staticcheck >/dev/null 2>&1; then \
		echo "Installing staticcheck..."; \
		go install honnef.co/go/tools/cmd/staticcheck@latest; \
	fi
	@staticcheck ./...

# Validate Go module
.PHONY: validate-module
validate-module:
	@echo "Validating Go module..."
	@if [ ! -f go.mod ]; then \
		echo "Error: go.mod file not found"; \
		exit 1; \
	fi
	@if [ ! -f go.sum ]; then \
		echo "Error: go.sum file not found"; \
		exit 1; \
	fi
	@EXPECTED_MODULE="$(MODULE_PATH)"; \
	ACTUAL_MODULE=$$(grep -E '^module ' go.mod | sed 's/^module //' | tr -d '[:space:]'); \
	if [ "$$ACTUAL_MODULE" != "$$EXPECTED_MODULE" ]; then \
		echo "Error: Module path mismatch"; \
		echo "   Expected: $$EXPECTED_MODULE"; \
		echo "   Actual:   $$ACTUAL_MODULE"; \
		exit 1; \
	fi
	@echo "Running go mod verify..."
	@go mod verify
	@echo "Checking if go.mod is tidy..."
	@go mod tidy
	@if ! git diff --quiet go.mod go.sum 2>/dev/null; then \
		echo "Error: go.mod or go.sum has uncommitted changes"; \
		echo "   Please run 'go mod tidy' and commit the changes"; \
		git diff go.mod go.sum; \
		exit 1; \
	fi
	@ACTUAL_MODULE=$$(grep -E '^module ' go.mod | sed 's/^module //' | tr -d '[:space:]'); \
	echo "Module validation passed"; \
	echo "   Module path: $$ACTUAL_MODULE"; \
	echo "   Module integrity: verified"

# Show help
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build          - Build without version injection (development build)"
	@echo "  build-release  - Build with version injection (release build)"
	@echo "  build-cross    - Cross-platform build (requires GOOS and GOARCH, e.g., make build-cross GOOS=linux GOARCH=amd64)"
	@echo "  install        - Install binary without version injection"
	@echo "  install-release - Install binary with version injection"
	@echo "  test           - Run tests"
	@echo "  test-coverage  - Run tests with coverage report (matches CI)"
	@echo "  check-coverage - Check coverage threshold (default: 20%, use THRESHOLD=30 to override)"
	@echo "  clean          - Remove build artifacts"
	@echo "  fmt            - Format code"
	@echo "  fmt-check      - Check code formatting (fails if files need formatting)"
	@echo "  lint           - Run linter (golangci-lint if available)"
	@echo "  vet            - Run go vet"
	@echo "  staticcheck    - Run staticcheck"
	@echo "  validate-module - Validate Go module (checks go.mod, go.sum, module path, and tidiness)"
	@echo "  help           - Show this help message"

