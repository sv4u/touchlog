.PHONY: help build install test test-race test-coverage lint fmt fmt-check deps deps-verify deps-tidy clean all install-tools docker-test docker-test-linux docker-test-linux-basic docker-test-linux-race docker-test-linux-coverage docker-test-macos docker-build-test docker-clean-test

# Variables
BINARY_NAME=touchlog
MAIN_PATH=./cmd/touchlog
COVERAGE_OUT=coverage.out
COVERAGE_DIR=coverage
GOLANGCI_LINT_VERSION=v2.6.2

# Version detection
# Try to get version from git, fallback to "dev" if git is not available
# Users can override by setting VERSION and COMMIT manually:
#   make build VERSION=1.0.0 COMMIT=abc123
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null | sed 's/^v//' || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "")

VERSION_PKG=github.com/sv4u/touchlog/v2/internal/version
LDFLAGS=-X $(VERSION_PKG).Version=$(VERSION) -X $(VERSION_PKG).Commit=$(COMMIT)

# Default target
help: ## Show this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Build targets
build: ## Build the binary
	@if [ "$(VERSION)" = "dev" ] && [ -z "$(COMMIT)" ]; then \
		if ! command -v git >/dev/null 2>&1; then \
			echo "⚠ Warning: git is not installed. Version will be 'dev'."; \
			echo "  Install git or set VERSION and COMMIT manually: make build VERSION=1.0.0 COMMIT=abc123"; \
		elif ! git rev-parse --git-dir >/dev/null 2>&1; then \
			echo "⚠ Warning: Not in a git repository. Version will be 'dev'."; \
			echo "  Set VERSION and COMMIT manually: make build VERSION=1.0.0 COMMIT=abc123"; \
		fi; \
	fi
	@echo "Building $(BINARY_NAME) (version: $(VERSION)$(if $(COMMIT),-$(COMMIT),))..."
	@CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o $(BINARY_NAME) $(MAIN_PATH)

install: ## Install the binary with version information (from local source)
	@if [ "$(VERSION)" = "dev" ] && [ -z "$(COMMIT)" ]; then \
		if ! command -v git >/dev/null 2>&1; then \
			echo "⚠ Warning: git is not installed. Version will be 'dev'."; \
			echo "  Install git or set VERSION and COMMIT manually: make install VERSION=1.0.0 COMMIT=abc123"; \
		elif ! git rev-parse --git-dir >/dev/null 2>&1; then \
			echo "⚠ Warning: Not in a git repository. Version will be 'dev'."; \
			echo "  Set VERSION and COMMIT manually: make install VERSION=1.0.0 COMMIT=abc123"; \
		fi; \
	fi
	@echo "Installing $(BINARY_NAME) (version: $(VERSION)$(if $(COMMIT),-$(COMMIT),))..."
	@go install -ldflags "$(LDFLAGS)" $(MAIN_PATH)
	@if [ "$(VERSION)" = "dev" ] && [ -z "$(COMMIT)" ]; then \
		echo "⚠ Installed $(BINARY_NAME) without version information (shows 'dev')"; \
	else \
		echo "✓ Installed $(BINARY_NAME) with version information"; \
	fi

# Test targets
test: ## Run all tests
	@echo "Running tests..."
	@go test ./... -v

test-race: ## Run tests with race detector
	@echo "Running tests with race detector..."
	@go test ./... -race -v

test-coverage: ## Generate test coverage reports
	@echo "Generating coverage reports..."
	@mkdir -p $(COVERAGE_DIR)
	@go test ./... -coverprofile=$(COVERAGE_OUT) -covermode=atomic
	@if [ ! -f $(COVERAGE_OUT) ]; then \
		echo "Warning: $(COVERAGE_OUT) was not created, creating empty file"; \
		touch $(COVERAGE_OUT); \
	fi
	@if [ -f $(COVERAGE_OUT) ] && [ -s $(COVERAGE_OUT) ]; then \
		echo "Generating HTML coverage report..."; \
		go tool cover -html=$(COVERAGE_OUT) -o $(COVERAGE_DIR)/coverage.html; \
		echo "✓ HTML coverage report generated"; \
		echo ""; \
		echo "Coverage by package:"; \
		go tool cover -func=$(COVERAGE_OUT) | grep -E "^github.com" | awk '{printf "  %-60s %s\n", $$1, $$3}' || echo "  (no coverage data found)"; \
		echo ""; \
		COVERAGE=$$(go tool cover -func=$(COVERAGE_OUT) | tail -1 | awk '{print $$3}'); \
		echo "Total Coverage: $$COVERAGE"; \
	else \
		echo "⚠ Skipping HTML coverage report generation ($(COVERAGE_OUT) missing or empty)"; \
		touch $(COVERAGE_DIR)/coverage.html; \
	fi

test-coverage-xml: test-coverage ## Generate XML coverage report (requires gocover-cobertura)
	@if [ -f $(COVERAGE_OUT) ] && [ -s $(COVERAGE_OUT) ]; then \
		echo "Installing gocover-cobertura if needed..."; \
		which gocover-cobertura > /dev/null 2>&1 || go install github.com/boumenot/gocover-cobertura@latest; \
		echo "Generating XML coverage report..."; \
		gocover-cobertura < $(COVERAGE_OUT) > $(COVERAGE_DIR)/coverage.xml; \
		echo "✓ XML coverage report generated"; \
	else \
		echo "⚠ Skipping XML coverage report generation ($(COVERAGE_OUT) missing or empty)"; \
		touch $(COVERAGE_DIR)/coverage.xml; \
	fi

# Linting and code quality targets
lint: ## Run all linters (golangci-lint, go vet, staticcheck)
	@echo "Running golangci-lint..."
	@golangci-lint --version || (echo "golangci-lint not found. Install with: make install-tools" && exit 1)
	@golangci-lint run
	@echo "Running go vet..."
	@go vet ./...
	@echo "Running staticcheck..."
	@staticcheck ./... || (echo "staticcheck not found. Install with: go install honnef.co/go/tools/cmd/staticcheck@latest" && exit 1)

fmt: ## Format code
	@echo "Formatting code..."
	@go fmt ./...

fmt-check: ## Check code formatting
	@echo "Checking code formatting..."
	@unformatted=$$(gofmt -l .); \
	if [ -n "$$unformatted" ]; then \
		echo "The following files are not formatted:"; \
		echo "$$unformatted"; \
		echo "Run 'make fmt' to format the code"; \
		exit 1; \
	fi
	@echo "✓ All files are properly formatted"

# Dependency management targets
deps: ## Download dependencies
	@echo "Downloading dependencies..."
	@go mod download

deps-verify: ## Verify dependencies
	@echo "Verifying dependencies..."
	@go mod verify

deps-tidy: ## Tidy dependencies (run go mod tidy)
	@echo "Tidying dependencies..."
	@go mod tidy
	@if ! git diff --exit-code go.mod go.sum 2>/dev/null; then \
		echo "go.mod or go.sum have been modified. Please commit the changes."; \
		exit 1; \
	fi
	@echo "✓ Dependencies are tidy"

deps-check: deps-tidy ## Check if dependencies are in sync (download + tidy + verify)
	@echo "Checking dependencies..."
	@go list ./... > /dev/null 2>&1 || true
	@$(MAKE) deps-verify

# Tool installation
install-tools: ## Install development tools
	@echo "Installing development tools..."
	@echo "Installing golangci-lint..."
	@which golangci-lint > /dev/null 2>&1 || (curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin $(GOLANGCI_LINT_VERSION))
	@echo "Installing staticcheck..."
	@go install honnef.co/go/tools/cmd/staticcheck@latest
	@echo "Installing gocover-cobertura..."
	@go install github.com/boumenot/gocover-cobertura@latest
	@echo "✓ Development tools installed"

# Cleanup targets
clean: ## Clean build artifacts and coverage files
	@echo "Cleaning build artifacts..."
	@rm -f $(BINARY_NAME)
	@rm -f $(COVERAGE_OUT)
	@rm -rf $(COVERAGE_DIR)
	@echo "✓ Clean complete"

# Composite targets
all: fmt-check lint test ## Run all checks (format check, lint, test)

ci: deps-verify fmt-check lint test test-race test-coverage ## Run CI checks (mirrors GitHub Actions CI workflow)

test-full: test test-race test-coverage ## Run all test variants

# Docker test targets
docker-build-test: ## Build Docker test image
	@echo "Building Docker test image..."
	@docker build -f Dockerfile.test -t touchlog-test:linux .
	@echo "✓ Docker test image built"

docker-test-linux: docker-build-test ## Run all tests in Linux Docker container
	@echo "Running all tests in Linux Docker container..."
	@mkdir -p $(COVERAGE_DIR)
	@docker run --rm \
		-v $$(pwd):/app \
		-v $$(pwd)/$(COVERAGE_DIR):/app/$(COVERAGE_DIR) \
		-v $$(pwd)/$(COVERAGE_OUT):/app/$(COVERAGE_OUT) \
		-e CGO_ENABLED=1 \
		touchlog-test:linux make test-full
	@echo "✓ Tests completed. Coverage reports saved to $(COVERAGE_DIR)/"

docker-test-linux-basic: docker-build-test ## Run basic tests in Linux Docker container
	@echo "Running basic tests in Linux Docker container..."
	@docker run --rm \
		-v $$(pwd):/app \
		-v $$(pwd)/$(COVERAGE_DIR):/app/$(COVERAGE_DIR) \
		-e CGO_ENABLED=1 \
		touchlog-test:linux make test

docker-test-linux-race: docker-build-test ## Run race detector tests in Linux Docker container
	@echo "Running race detector tests in Linux Docker container..."
	@docker run --rm \
		-v $$(pwd):/app \
		-v $$(pwd)/$(COVERAGE_DIR):/app/$(COVERAGE_DIR) \
		-e CGO_ENABLED=1 \
		touchlog-test:linux make test-race

docker-test-linux-coverage: docker-build-test ## Generate coverage reports in Linux Docker container
	@echo "Generating coverage reports in Linux Docker container..."
	@mkdir -p $(COVERAGE_DIR)
	@docker run --rm \
		-v $$(pwd):/app \
		-v $$(pwd)/$(COVERAGE_DIR):/app/$(COVERAGE_DIR) \
		-v $$(pwd)/$(COVERAGE_OUT):/app/$(COVERAGE_OUT) \
		-e CGO_ENABLED=1 \
		touchlog-test:linux make test-coverage-xml
	@echo "✓ Coverage reports saved to $(COVERAGE_DIR)/"

docker-test: docker-test-linux ## Run all tests in Docker (alias for docker-test-linux)

docker-test-macos: ## Run tests natively on macOS (requires macOS host)
	@echo "Running tests natively on macOS..."
	@if [ "$$(uname)" != "Darwin" ]; then \
		echo "Error: This target requires macOS host"; \
		exit 1; \
	fi
	@$(MAKE) test-full

docker-clean-test: ## Clean Docker test containers and images
	@echo "Cleaning Docker test resources..."
	@docker-compose -f docker-compose.test.yml down --rmi local 2>/dev/null || true
	@docker rmi touchlog-test:linux 2>/dev/null || true
	@echo "✓ Docker test resources cleaned"

# Release targets (for local testing)
goreleaser-check: ## Validate GoReleaser configuration
	@goreleaser check || (echo "goreleaser not found. Install with: go install github.com/goreleaser/goreleaser/v2@latest" && exit 1)

goreleaser-snapshot: ## Run GoReleaser in snapshot mode (local testing)
	@goreleaser release --snapshot --clean || (echo "goreleaser not found. Install with: go install github.com/goreleaser/goreleaser/v2@latest" && exit 1)
