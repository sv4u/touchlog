# touchlog

[![CI](https://github.com/sv4u/touchlog/workflows/CI/badge.svg)](https://github.com/sv4u/touchlog/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/sv4u/touchlog)](https://goreportcard.com/report/github.com/sv4u/touchlog)
[![codecov](https://codecov.io/gh/sv4u/touchlog/branch/master/graph/badge.svg?token=IOT1S6CPGY)](https://codecov.io/gh/sv4u/touchlog/branch/master)
[![License](https://img.shields.io/github/license/sv4u/touchlog.svg)](https://github.com/sv4u/touchlog/blob/main/LICENSE)
[![Go Version](https://img.shields.io/github/go-mod/go-version/sv4u/touchlog)](https://github.com/sv4u/touchlog/blob/main/go.mod)
[![GitHub release](https://img.shields.io/github/release/sv4u/touchlog.svg)](https://github.com/sv4u/touchlog/releases)

A knowledge graph note-taking system built with Go. touchlog provides a powerful CLI for managing structured notes with automatic link resolution, graph queries, and real-time indexing.

## Features

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

```bash
go install github.com/sv4u/touchlog/v2/cmd/touchlog@latest
```

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

   Follow the interactive prompts to create a new note with frontmatter.

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
- `touchlog new` - Create a new note interactively

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
# Create a note
touchlog new
# Follow prompts to create note with key "introduction"

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

#### Running Tests in Docker

**Using Makefile (Recommended)**:

```bash
# Build Docker test image
make docker-build-test

# Run all tests in Linux Docker container
make docker-test-linux

# Run specific test variants
make docker-test-linux-basic      # Basic tests only
make docker-test-linux-race       # Race detector tests
make docker-test-linux-coverage   # Coverage reports

# Run tests natively on macOS (requires macOS host)
make docker-test-macos

# Clean Docker test resources
make docker-clean-test
```

**Using Docker Compose**:

```bash
# Run all tests
docker-compose -f docker-compose.test.yml run --rm test-linux

# Run specific test variants
docker-compose -f docker-compose.test.yml run --rm test-linux-basic
docker-compose -f docker-compose.test.yml run --rm test-linux-race
docker-compose -f docker-compose.test.yml run --rm test-linux-coverage
```

**Using Docker directly**:

```bash
# Build the test image
docker build -f Dockerfile.test -t touchlog-test:linux .

# Run tests
docker run --rm \
  -v $(pwd):/app \
  -v $(pwd)/coverage:/app/coverage \
  -v $(pwd)/coverage.out:/app/coverage.out \
  -e CGO_ENABLED=1 \
  touchlog-test:linux make test-full
```

#### Coverage Reports

Coverage reports are automatically saved to the host machine:

- **HTML Report**: `coverage/coverage.html` - Open in a web browser
- **XML Report**: `coverage/coverage.xml` - For CI/CD integration
- **Coverage Profile**: `coverage.out` - Raw coverage data

#### Testing on Multiple Platforms

The GitHub Actions workflow automatically runs tests on:

- **Linux** (Ubuntu latest)
- **macOS** (macOS latest and macOS 14)

**WSL Testing**: For WSL (Windows Subsystem for Linux) testing, you can run the Linux Docker container on a Windows machine with WSL2:

```bash
# On Windows with WSL2, run from WSL terminal
make docker-test-linux
```

The Linux Docker container will run in the WSL2 environment, providing a Linux-like testing environment on Windows.

### Building

```bash
# Build binary
go build ./cmd/touchlog

# Build for specific platform
GOOS=linux GOARCH=amd64 go build ./cmd/touchlog
```

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
