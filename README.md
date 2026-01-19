# touchlog

[![CI](https://github.com/sv4u/touchlog/workflows/CI/badge.svg)](https://github.com/sv4u/touchlog/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/sv4u/touchlog)](https://goreportcard.com/report/github.com/sv4u/touchlog)
[![codecov](https://codecov.io/gh/sv4u/touchlog/branch/master/graph/badge.svg?token=IOT1S6CPGY)](https://codecov.io/gh/sv4u/touchlog/branch/master)
[![License](https://img.shields.io/github/license/sv4u/touchlog.svg)](https://github.com/sv4u/touchlog/blob/main/LICENSE)
[![Go Version](https://img.shields.io/github/go-mod/go-version/sv4u/touchlog)](https://github.com/sv4u/touchlog/blob/main/go.mod)
[![GitHub release](https://img.shields.io/github/release/sv4u/touchlog.svg)](https://github.com/sv4u/touchlog/releases)

A knowledge graph note-taking system built with Go. touchlog provides a powerful CLI for managing structured notes with automatic link resolution, graph queries, and real-time indexing.

## Features

- **Interactive wizard**: Beautiful TUI wizard for creating notes with step-by-step guidance
- **Vault-based organization**: Organize notes in type-specific directories with structured frontmatter
- **Automatic indexing**: SQLite-based index with automatic link resolution and incremental updates
- **Graph queries**: Find backlinks, neighbors, and paths between notes using graph traversal
- **Graph visualization**: Export knowledge graphs to Graphviz DOT format
- **Real-time updates**: Daemon mode with filesystem watching for automatic index updates
- **Structured notes**: YAML frontmatter with validation and diagnostics
- **Wiki-style links**: Support for `[[note:key]]`, `[[key]]`, and edge-type annotations
- **Deterministic exports**: Stable, diffable JSON and DOT exports

## Installation

### Prerequisites

- Go 1.22 or later

### Build from Source

```bash
git clone https://github.com/sv4u/touchlog.git
cd touchlog
go build ./cmd/touchlog
```

The binary will be created in the current directory as `touchlog`.

### Install via Go

**Recommended: Install with version information (from local source):**

```bash
git clone https://github.com/sv4u/touchlog.git
cd touchlog
make install
```

This installs touchlog to `$GOPATH/bin` (or `$GOBIN` if set) with proper version information.

**Alternative: Install from remote (without version information):**

```bash
go install github.com/sv4u/touchlog/v2/cmd/touchlog@latest
```

Note: Installing directly from remote will show "dev" as the version. See [Building with Version Information](#building-with-version-information) for details.

## Quick Start

1. **Initialize a vault**:

   ```bash
   touchlog init
   ```

   This creates a `.touchlog` directory with a default configuration file.

2. **Create a note**:

   ```bash
   touchlog new
   ```

   This launches an interactive wizard that guides you through:
   - Selecting the note type
   - Entering a unique key/name (with validation)
   - Setting the title
   - Adding optional tags (comma-separated)
   - Setting the state (optional, defaults to type's default)
   - Reviewing all details before creation

3. **Build the index**:

   ```bash
   touchlog index rebuild
   ```

   This scans all `.Rmd` files in your vault and builds the index.

4. **Query your notes**:

   ```bash
   touchlog query search --type note
   touchlog query backlinks --target note:my-note
   touchlog query neighbors --root note:my-note --max-depth 2
   touchlog query paths --source note:start --destination note:end --max-depth 5
   ```

5. **Export the graph**:

   ```bash
   touchlog graph export dot --out graph.dot
   dot -Tpng graph.dot -o graph.png
   ```

## Vault Structure

A touchlog vault is a directory containing:

```
vault/
├── .touchlog/
│   ├── config.yaml          # Vault configuration
│   └── index.db             # SQLite index (auto-generated)
├── note/                    # Type-specific directories
│   ├── my-note.Rmd
│   └── another-note.Rmd
├── article/
│   └── my-article.Rmd
└── ...
```

### Note Files

Notes are stored as `.Rmd` files with YAML frontmatter:

```markdown
---
id: note-001
type: note
key: my-note
title: My Note
state: draft
tags: [important, todo]
created: 2024-01-01T00:00:00Z
updated: 2024-01-01T00:00:00Z
---
# My Note

This note links to [[note:another-note]] and [[article:my-article]].

You can also use unqualified links: [[another-note]] if the key is unique.
```

### Configuration

The vault configuration file (`.touchlog/config.yaml`) defines:

- **Types**: Note types with validation rules (key patterns, default states)
- **Tags**: Tag configuration and preferred tags
- **Edges**: Edge type definitions for relationships
- **Templates**: Template configuration (future)

Example configuration:

```yaml
version: 1
types:
  note:
    description: A note
    default_state: draft
    key_pattern: ^[a-z0-9]+(-[a-z0-9]+)*$
    key_max_len: 64
  article:
    description: An article
    default_state: draft
    key_pattern: ^[a-z0-9]+(-[a-z0-9]+)*$
    key_max_len: 64
tags:
  preferred: [important, todo, reference]
edges:
  related-to:
    description: General relationship
  references:
    description: Reference relationship
templates:
  root: templates
```

## Commands

### Vault Management

- `touchlog init` - Initialize a new vault in the current directory
- `touchlog new` - Create a new note using an interactive TUI wizard
  - The wizard guides you through type selection, key input (with validation), title, tags, and state
  - Automatically falls back to non-interactive mode in CI/test environments

### Index Management

- `touchlog index rebuild` - Rebuild the entire index from scratch
- `touchlog index export --out <file>` - Export the index to JSON

### Querying

- `touchlog query search [options]` - Search notes with filters
  - `--type <types>` - Filter by types (comma-separated)
  - `--state <states>` - Filter by states (comma-separated)
  - `--tag <tags>` - Filter by tags (comma-separated)
  - `--match-any-tag` - Match any tag (default: match all)
  - `--limit <n>` - Limit results
  - `--format table|json` - Output format

- `touchlog query backlinks --target <node> [options]` - Find backlinks to a node
  - `--direction in|out|both` - Link direction (default: in)
  - `--edge-type <types>` - Filter by edge types
  - `--format table|json` - Output format

- `touchlog query neighbors --root <node> --max-depth <n> [options]` - Find neighbors
  - `--direction in|out|both` - Link direction (default: both)
  - `--edge-type <types>` - Filter by edge types
  - `--format table|json` - Output format

- `touchlog query paths --source <node> --destination <node> [--destination <node> ...] --max-depth <n> [options]` - Find paths
  - `--max-paths <n>` - Maximum paths per destination (default: 10)
  - `--direction in|out|both` - Link direction (default: both)
  - `--edge-type <types>` - Filter by edge types
  - `--format table|json` - Output format

### Interactive Wizard

The `touchlog new` command launches an interactive TUI (Terminal User Interface) wizard built with [bubbletea](https://github.com/charmbracelet/bubbletea). The wizard guides you through creating a new note step-by-step:

1. **Type Selection**: Choose from available note types configured in your vault
2. **Key Input**: Enter a unique key/name for the note
   - Validates against the type's key pattern
   - Checks maximum length
   - Ensures uniqueness within the type directory
3. **Title Input**: Enter a descriptive title for the note
4. **Tags Input** (optional): Add comma-separated tags
5. **State Input** (optional): Set the note state (defaults to the type's default state)
6. **Verification**: Review all details before creating the note

**Navigation**:

- Use arrow keys (`↑`/`↓` or `j`/`k`) to navigate selections
- Press `Enter` to confirm/continue
- Press `Esc` to go back to the previous step
- Press `q` or `Ctrl+C` to quit

**Automatic Mode Detection**: The wizard automatically detects if it's running in an interactive terminal. In non-interactive environments (tests, CI), it falls back to default values without starting the TUI.

### Graph Operations

- `touchlog graph export dot --out <file> [options]` - Export graph to DOT format
  - `--root <node> [--root <node> ...]` - Root nodes (can specify multiple)
  - `--type <types>` - Filter by types
  - `--state <states>` - Filter by states
  - `--tag <tags>` - Filter by tags
  - `--edge-type <types>` - Filter by edge types
  - `--depth <n>` - Maximum depth (default: 10)
  - `--force` - Overwrite existing file

### Daemon

- `touchlog daemon start` - Start the daemon for real-time indexing
- `touchlog daemon stop` - Stop the daemon
- `touchlog daemon status` - Check daemon status

## Architecture

touchlog follows a **contracts-first** architecture with clear separation of concerns:

### Core Components

- **Models** (`internal/model/`): Canonical data structures (Note, Frontmatter, RawLink, etc.)
- **Config** (`internal/config/`): Configuration loading and validation
- **Note Parser** (`internal/note/`): Frontmatter parsing and wiki-link extraction
- **Store** (`internal/store/`): SQLite persistence layer with migrations
- **Index** (`internal/index/`): Full-scan indexing with atomic rebuilds
- **Query** (`internal/query/`): Search and graph query execution
- **Graph** (`internal/graph/`): Graph loading and export
- **Daemon** (`internal/daemon/`): IPC server and lifecycle management
- **Watch** (`internal/watch/`): Filesystem watching and incremental indexing

### Design Principles

1. **Contracts First**: Models and schemas defined before behavior
2. **Derived State Only**: Index and exports are always rebuildable
3. **Determinism**: All outputs are stable and diffable
4. **Explicit Errors**: Never silently fail or coerce invalid input
5. **Diagnostics**: Validation produces diagnostics, not hard failures

### Indexing

The index is built in two passes:

1. **Pass 1**: Parse all notes, build `(type, key) -> id` map
2. **Pass 2**: Resolve all links using the type/key map

The index is stored in SQLite with the following schema:

- `meta`: Metadata (schema version, etc.)
- `nodes`: Note nodes with frontmatter
- `edges`: Links between notes (with resolved `to_id`)
- `tags`: Tags associated with notes
- `diagnostics`: Parse errors and warnings

### Graph Queries

All graph queries (backlinks, neighbors, paths) execute in-memory after loading the relevant subgraph from SQLite. This approach:

- Avoids complex recursive SQL
- Simplifies BFS correctness and determinism
- Aligns with future TUI graph viewer needs

All traversals maintain visited sets for cycle detection and guarantee termination.

## Examples

### Creating and Linking Notes

```bash
# Create a note using the interactive wizard
touchlog new
# The wizard will guide you through:
# 1. Select type (e.g., "note")
# 2. Enter key (e.g., "introduction")
# 3. Enter title (e.g., "Introduction to touchlog")
# 4. Add tags (optional, e.g., "getting-started, tutorial")
# 5. Set state (optional, defaults to type's default)
# 6. Review and confirm

# Create another note that links to it
touchlog new
# Create note with key "getting-started" and add link: [[note:introduction]]

# Rebuild index
touchlog index rebuild

# Find backlinks to introduction
touchlog query backlinks --target note:introduction
```

### Finding Related Notes

```bash
# Find all neighbors within 2 hops
touchlog query neighbors --root note:my-note --max-depth 2

# Find paths between notes
touchlog query paths --source note:start --destination note:end --max-depth 5

# Export graph for visualization
touchlog graph export dot --out graph.dot
dot -Tsvg graph.dot -o graph.svg
```

### Using the Daemon

```bash
# Start daemon for real-time indexing
touchlog daemon start

# Create/edit notes - index updates automatically
# Query works immediately
touchlog query search --type note

# Stop daemon
touchlog daemon stop
```

## Development

### Project Structure

```
touchlog/
├── cmd/
│   └── touchlog/
│       └── main.go              # CLI entry point
├── internal/
│   ├── model/                   # Core data models
│   ├── config/                  # Configuration loading
│   ├── note/                    # Note parsing
│   ├── store/                   # SQLite persistence
│   ├── index/                   # Indexing logic
│   ├── query/                   # Query execution
│   ├── graph/                   # Graph operations
│   ├── daemon/                  # Daemon and IPC
│   ├── watch/                   # Filesystem watching
│   └── cli/                     # CLI command definitions
├── testdata/
│   ├── vaults/                  # Test vault fixtures
│   └── golden/                  # Golden test fixtures
├── go.mod
└── README.md
```

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test ./... -cover

# Run specific package tests
go test ./internal/query -v
```

#### Using Makefile

```bash
# Run all test variants (test, test-race, test-coverage)
make test-full

# Run basic tests
make test

# Run tests with race detector
make test-race

# Generate coverage reports
make test-coverage
```

### Docker Testing

You can run tests in isolated Linux and macOS environments using Docker. This ensures consistent test execution across different platforms.

#### Prerequisites

- Docker and Docker Compose installed
- For macOS testing: macOS host (tests run natively on macOS)

#### Quick Start

```bash
# Run all tests in Linux container
make docker-test-linux

# Run specific test types
make docker-test-linux-basic      # Basic tests
make docker-test-linux-race       # Race detector
make docker-test-linux-coverage   # Coverage reports

# Run tests natively on macOS
make docker-test-macos
```

#### Available Methods

**1. Using Makefile (Recommended)**:

```bash
# Build Docker image
make docker-build-test

# Run all test variants
make docker-test-linux

# Run specific variants
make docker-test-linux-basic
make docker-test-linux-race
make docker-test-linux-coverage

# Clean up
make docker-clean-test
```

**2. Using Docker Compose**:

```bash
# Run all tests
docker-compose -f docker-compose.test.yml run --rm test-linux

# Run specific variants
docker-compose -f docker-compose.test.yml run --rm test-linux-basic
docker-compose -f docker-compose.test.yml run --rm test-linux-race
docker-compose -f docker-compose.test.yml run --rm test-linux-coverage
```

**3. Using Docker Directly**:

```bash
# Build image
docker build -f Dockerfile.test -t touchlog-test:linux .

# Run tests
docker run --rm \
  -v $(pwd):/app \
  -v $(pwd)/coverage:/app/coverage \
  -v $(pwd)/coverage.out:/app/coverage.out \
  -e CGO_ENABLED=1 \
  touchlog-test:linux make test-full
```

**4. Using Helper Script**:

```bash
# Run all tests
./scripts/docker-test.sh

# Run with options
./scripts/docker-test.sh -p linux -t coverage
./scripts/docker-test.sh -p macos -t full

# Build only
./scripts/docker-test.sh build

# Clean up
./scripts/docker-test.sh clean
```

#### Coverage Reports

All coverage reports are automatically saved to the host:

- **HTML**: `coverage/coverage.html` - Open in browser
- **XML**: `coverage/coverage.xml` - For CI/CD
- **Profile**: `coverage.out` - Raw data

#### Platform-Specific Notes

**Linux**: The Linux Docker container uses Alpine Linux with Go 1.25. All dependencies (including SQLite) are pre-installed.

**macOS**: macOS testing runs natively on the macOS host (not in a container). This is because Docker on macOS runs Linux containers, not macOS containers.

```bash
make docker-test-macos
```

**WSL (Windows Subsystem for Linux)**: For WSL testing, run the Linux Docker container from within WSL:

```bash
# In WSL terminal
make docker-test-linux
```

The container will run in the WSL2 environment, providing Linux-like testing on Windows.

#### Testing on Multiple Platforms

The GitHub Actions workflow automatically runs tests on:

- **Linux** (Ubuntu latest)
- **macOS** (macOS latest and macOS 14)

See `.github/workflows/test-and-coverage.yml` for details.

#### Troubleshooting

**Permission Issues**: If you encounter permission issues with coverage files:

```bash
# Fix permissions
sudo chown -R $USER:$USER coverage/ coverage.out
```

**Container Build Fails**: If the Docker build fails:

```bash
# Clean and rebuild
make docker-clean-test
make docker-build-test
```

**Coverage Files Not Saved**: Ensure the coverage directory exists and is writable:

```bash
mkdir -p coverage
chmod 755 coverage
```

#### File Structure

```
touchlog/
├── Dockerfile.test              # Docker image for testing
├── docker-compose.test.yml     # Docker Compose configuration
├── .dockerignore               # Files to exclude from Docker build
├── scripts/
│   └── docker-test.sh          # Helper script for Docker testing
└── coverage/                   # Coverage reports (generated)
    ├── coverage.html
    ├── coverage.xml
    └── coverage.out
```

### Building

```bash
# Build binary
go build ./cmd/touchlog

# Build for specific platform
GOOS=linux GOARCH=amd64 go build ./cmd/touchlog
```

### Building with Version Information

Touchlog embeds version information at build time using Go's ldflags mechanism. The version displayed depends on how the binary was built:

#### Using Makefile (Recommended for Development)

The Makefile automatically injects version information:

```bash
make build
./touchlog version
# Output: touchlog version v1.2.3-5-gabc123-abc123
```

The Makefile uses:

- `git describe --tags --always --dirty` for the version string
- `git rev-parse --short HEAD` for the commit hash

#### Using GoReleaser (Releases)

Official releases built with GoReleaser include:

- Full semantic version from git tags (e.g., "1.2.3")
- Full 40-character commit hash

#### Using `go install`

**Installing from local source (recommended for development):**

```bash
# Clone the repository first
git clone https://github.com/sv4u/touchlog.git
cd touchlog

# Install with version information using Makefile
make install

# Or install manually with ldflags
VERSION=$(git describe --tags --always --dirty | sed 's/^v//')
COMMIT=$(git rev-parse --short HEAD)
go install -ldflags "-X github.com/sv4u/touchlog/v2/internal/version.Version=$VERSION -X github.com/sv4u/touchlog/v2/internal/version.Commit=$COMMIT" ./cmd/touchlog
```

**Installing from remote (with automatic version detection):**

When installing directly from GitHub, touchlog automatically extracts version information from Go's build metadata:

```bash
go install github.com/sv4u/touchlog/v2/cmd/touchlog@latest
touchlog version
# Output: touchlog version 2.1.1
# (Version extracted from module version in build info)
```

**Important:** When installing from the module proxy (via `go install @latest`), Go only includes the module version in build metadata, not the VCS commit hash. This is a limitation of the Go module proxy system - it doesn't have access to the git repository.

**To get full version information (including commit hash) like GoReleaser builds:**

- Clone the repository and use `make install` (recommended)
- Or build manually with ldflags as shown in the "Manual Build with ldflags" section

The automatic version detection from BuildInfo provides the module version, which is better than "dev" but doesn't include the commit hash when installing from remote.

#### Manual Build with ldflags

To build manually with version information:

```bash
VERSION=$(git describe --tags --always --dirty | sed 's/^v//')
COMMIT=$(git rev-parse --short HEAD)
go build -ldflags "-X github.com/sv4u/touchlog/v2/internal/version.Version=$VERSION -X github.com/sv4u/touchlog/v2/internal/version.Commit=$COMMIT" -o touchlog ./cmd/touchlog
```

#### When Git is Not Available

If git is not installed or you're not in a git repository, the Makefile will automatically fall back to `VERSION=dev` and `COMMIT=""`. You'll see a warning message, and the binary will show "dev" as the version.

**To manually specify version when git is unavailable:**

```bash
# Build with custom version
make build VERSION=1.0.0 COMMIT=abc123

# Install with custom version
make install VERSION=1.0.0 COMMIT=abc123

# Or build directly with go
go build -ldflags "-X github.com/sv4u/touchlog/v2/internal/version.Version=1.0.0 -X github.com/sv4u/touchlog/v2/internal/version.Commit=abc123" -o touchlog ./cmd/touchlog
```

**Scenarios:**

1. **`go install` from remote (GitHub)**: No git repository available, version will be "dev" with a warning. This is expected and cannot be avoided without cloning the repository.

2. **`make install` without git installed**: Falls back to "dev", shows warning. Install git or manually specify VERSION/COMMIT.

3. **`make install` in non-git directory**: Falls back to "dev", shows warning. Initialize git repo or manually specify VERSION/COMMIT.

4. **`go install` from local source without git**: Same as #2 or #3. Use `make install` or manually specify ldflags.

## Contributing

### Commit Message Guidelines

This project follows the [Conventional Commits](https://www.conventionalcommits.org/) specification.

**Commit Types**:

- `feat`: A new feature
- `fix`: A bug fix
- `docs`: Documentation only changes
- `test`: Adding or updating tests
- `refactor`: Code refactoring without changing functionality
- `perf`: Performance improvements
- `ci`: Changes to CI/CD configuration
- `chore`: Other changes that don't modify src or test files

**Examples**:

```text
feat(query): add backlinks command
fix(index): resolve link resolution bug
docs(readme): update architecture section
test(query): add edge case tests for cycles
```

## License

See the [LICENSE](LICENSE) file for details.
