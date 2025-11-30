# touchlog

[![CI](https://github.com/sv4u/touchlog/workflows/CI/badge.svg)](https://github.com/sv4u/touchlog/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/sv4u/touchlog)](https://goreportcard.com/report/github.com/sv4u/touchlog)
[![codecov](https://codecov.io/gh/sv4u/touchlog/branch/master/graph/badge.svg?token=IOT1S6CPGY)](https://codecov.io/gh/sv4u/touchlog)
[![License](https://img.shields.io/github/license/sv4u/touchlog.svg)](https://github.com/sv4u/touchlog/blob/main/LICENSE)
[![Go Version](https://img.shields.io/github/go-mod/go-version/sv4u/touchlog)](https://github.com/sv4u/touchlog/blob/main/go.mod)
[![GitHub release](https://img.shields.io/github/release/sv4u/touchlog.svg)](https://github.com/sv4u/touchlog/releases)

A terminal-based note editor with template support, designed for Zettelkasten note-taking methodology. Built with Go and [Bubble Tea](https://github.com/charmbracelet/bubbletea), touchlog provides an interactive TUI for quickly creating structured notes from customizable Markdown templates.

## Features

- **Template-based note creation**: Select from configured templates to start new notes
- **Variable substitution**: Templates support `{{variable}}` syntax with automatic substitution
- **Default variables**: Pre-populated variables including `{{date}}`, `{{time}}`, and `{{datetime}}`
- **Interactive TUI**: Clean terminal interface built with Bubble Tea
- **XDG-compliant**: Follows XDG Base Directory Specification for configuration and data storage
- **Markdown output**: Saves notes as `.md` files with timestamp-based filenames

## Installation

### Prerequisites

- Go 1.21 or later

### Build from Source

```bash
git clone https://github.com/sv4u/touchlog.git
cd touchlog
go build ./cmd/touchlog
```

The binary will be created in the current directory as `touchlog`.

### Install via Go

```bash
go install github.com/sv4u/touchlog/cmd/touchlog@latest
```

## Configuration

touchlog uses XDG Base Directory Specification for storing configuration and templates:

- **Configuration**: `~/.config/touchlog/config.yaml` (or `$XDG_CONFIG_HOME/touchlog/config.yaml`)
- **Templates**: `~/.local/share/touchlog/templates/` (or `$XDG_DATA_HOME/touchlog/templates/`)

### Configuration File

**Automatic Configuration Creation**: On first run, touchlog automatically creates a default configuration file at `~/.config/touchlog/config.yaml` (or `$XDG_CONFIG_HOME/touchlog/config.yaml`) if it doesn't exist. The default configuration includes:

- Three common templates: Daily Note, Meeting Notes, and Journal
- Default notes directory: `~/notes`
- All date/time variables enabled with standard formats
- Vim mode disabled

You can customize the configuration file at any time. If the templates directory is empty, touchlog will also create minimal example template files (`daily.md`, `meeting.md`, `journal.md`) to get you started.

**Manual Configuration** (optional):

If you prefer to create the configuration file manually, create `~/.config/touchlog/config.yaml`:

```yaml
templates:
  - name: "Daily Note"
    file: "daily.md"
  - name: "Meeting Notes"
    file: "meeting.md"
  - name: "Book Note"
    file: "book-note.md"
notes_directory: "/path/to/your/notes"

# Optional: Configure date/time/datetime variable formats
datetime_vars:
  date:
    enabled: true
    format: "2006-01-02"  # Go time format (YYYY-MM-DD)
  time:
    enabled: true
    format: "15:04:05"    # Go time format (HH:MM:SS)
  datetime:
    enabled: true
    format: "2006-01-02 15:04:05"  # Go time format (YYYY-MM-DD HH:MM:SS)

# Optional: Define custom static variables
variables:
  author: "Your Name"
  project: "My Project"

# Optional: Enable vim keymap support
vim_mode: false
```

### Template Files

Template files are stored in `~/.local/share/touchlog/templates/` (or `$XDG_DATA_HOME/touchlog/templates/`). Templates support variable substitution using `{{variable}}` syntax.

**Automatic Template Creation**: If the templates directory is empty on first run, touchlog will automatically create minimal example templates (`daily.md`, `meeting.md`, `journal.md`) that match the default configuration. These templates are simple but functional and can be customized as needed.

**Manual Template Creation** (optional):

Create template files in `~/.local/share/touchlog/templates/`:

**Example template** (`daily.md`):

```markdown
# Daily Note - {{date}}

## Events
- 

## Thoughts
- 

## Tasks
- 
```

**Available default variables**:

- `{{date}}` - Current date (default format: `YYYY-MM-DD`)
- `{{time}}` - Current time (default format: `HH:MM:SS`)
- `{{datetime}}` - Current date and time (default format: `YYYY-MM-DD HH:MM:SS`)

**Customizing date/time formats**:

You can customize the format of date, time, and datetime variables in your configuration file using Go's time format syntax. The format uses Go's reference time: `Mon Jan 2 15:04:05 MST 2006` (which is `01/02 03:04:05PM '06 -0700`).

**Example formats**:

```yaml
datetime_vars:
  date:
    enabled: true
    format: "January 2, 2006"      # "January 15, 2024"
  time:
    enabled: true
    format: "3:04 PM"              # "2:30 PM"
  datetime:
    enabled: true
    format: "2006-01-02T15:04:05"  # "2024-01-15T14:30:22"
```

You can also disable specific variables by setting `enabled: false`. If `datetime_vars` is not specified in the config, all variables are enabled with default formats.

**Custom variables**:

You can define custom static variables in your configuration file that will be available in all templates:

```yaml
variables:
  author: "John Doe"
  project: "My Project"
  location: "Home Office"
```

These variables can then be used in templates using `{{author}}`, `{{project}}`, `{{location}}`, etc. Custom variables can override default date/time/datetime variables, but it's recommended to avoid using reserved names (`date`, `time`, `datetime`) for custom variables.

**Vim keymap support**:

You can enable vim-style keybindings by setting `vim_mode: true` in your configuration:

```yaml
vim_mode: true
```

When vim mode is enabled:

- **Template selection**: Use `j`/`k` to navigate, `Enter` to select, `q` to quit
- **Note editing**:
  - Press `i` or `a` to enter insert mode
  - Press `Esc` to exit insert mode (normal mode)
  - In normal mode: `h`/`j`/`k`/`l` for movement, `dd` to delete line, `:w` to save, `:q` to quit
  - `Ctrl+S` works in both modes to save

## Usage

1. **Start the application**:

   ```bash
   touchlog
   ```

   Or with a custom output directory:

   ```bash
   touchlog -output-dir ~/my-notes
   # or using the shorthand
   touchlog -o ~/my-notes
   ```

2. **Select a template**: Use arrow keys to navigate and press `Enter` to select

3. **Edit your note**: The template will be loaded with variables substituted. Edit as needed.

4. **Save the note**: Press `Ctrl+S` to save. The note will be saved as a Markdown file with a timestamp-based filename (e.g., `2024-01-15-143022.md`) in your configured notes directory. The directory will be created automatically if it doesn't exist.

5. **Quit**: Press `Ctrl+C` or `q` to exit

### Command Line Options

- `-output-dir <path>` or `-o <path>`: Override the notes directory specified in the config file

   The output directory can be an absolute path or a path starting with `~` (which will be expanded to your home directory). This option takes precedence over the `notes_directory` setting in the config file.

   **Examples**:

   ```bash
   touchlog -output-dir ~/my-notes
   touchlog -o /tmp/notes
   touchlog --output-dir ~/Documents/journal
   ```

## Keyboard Shortcuts

### Default Mode

- `↑/↓` - Navigate template list
- `Enter` - Select template
- `Ctrl+S` - Save note
- `Ctrl+C` or `q` - Quit application

### Vim Mode (when `vim_mode: true` in config)

**Template Selection**:

- `j`/`k` - Navigate up/down template list
- `Enter` - Select template
- `q` - Quit

**Note Editing**:

- `i`/`a` - Enter insert mode
- `Esc` - Exit insert mode (normal mode)
- `h`/`j`/`k`/`l` - Cursor movement (in normal mode, enters insert mode)
- `w`/`b` - Word forward/backward (simplified)
- `0`/`$` - Beginning/end of line (simplified)
- `dd` - Delete line
- `:w` - Save (vim-style)
- `:q` - Quit (vim-style)
- `Ctrl+S` - Save (works in both modes)

## Project Structure

```text
touchlog/
├── cmd/
│   └── touchlog/
│       └── main.go              # Application entry point
├── internal/
│   ├── config/
│   │   └── config.go            # Configuration file parsing
│   ├── template/
│   │   └── template.go          # Template loading and processing
│   ├── xdg/
│   │   └── xdg.go               # XDG path resolution
│   └── editor/
│       └── editor.go            # Bubble Tea model and application logic
├── go.mod                       # Go module definition
└── README.md
```

## Programmatic API

touchlog can be used programmatically by other Go applications. This is useful for external tools like a Zettelkasten daemon that needs to create notes programmatically.

### Basic Usage

```go
import "github.com/sv4u/touchlog/internal/api"

opts := &api.Options{
    OutputDirectory: "/path/to/notes",
}
err := api.Run(opts)
if err != nil {
    // Handle error
}
```

### API Reference

**`api.Options`**:

- `OutputDirectory` (string): Override the notes directory specified in the config file. This takes precedence over the config file setting. Supports `~` expansion for home directory paths.
- `ConfigPath` (string): Reserved for future use. Currently, the default config file location is always used.

**`api.Run(opts *Options) error`**:

Creates and runs a touchlog instance with the given options. Returns an error if initialization or execution fails.

**Priority Order**:

When determining which output directory to use, touchlog follows this priority:

1. CLI flag (`-output-dir` or `-o`) - highest priority
2. API parameter (`Options.OutputDirectory`)
3. Config file setting (`notes_directory`) - lowest priority

Only one source is used (no merging). The first non-empty value in the priority order is used.

## Architecture

touchlog uses the Model-View-Update (MVU) pattern implemented by Bubble Tea:

- **Model**: Application state (current view, selected template, note content)
- **Update**: Handles user input and state transitions
- **View**: Renders the UI using Bubbles components (list, textarea)

The application flow:

1. **Template Selection**: User selects from available templates
2. **Template Loading**: Template is loaded and variables are substituted
3. **Note Editing**: User edits the note content
4. **Saving**: Note is saved to the configured directory

## Dependencies

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Bubbles](https://github.com/charmbracelet/bubbles) - UI components
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) - Terminal styling
- [XDG](https://github.com/adrg/xdg) - XDG Base Directory implementation
- [YAML v3](https://github.com/go-yaml/yaml) - YAML parsing

## Contributing

### Commit Message Guidelines

This project follows the [Conventional Commits](https://www.conventionalcommits.org/) specification. All commit messages should follow this format:

```text
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

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
feat(editor): add vim mode support
fix(config): resolve path resolution issue
docs(readme): update installation instructions
test(template): add tests for variable substitution
```

**Breaking Changes**: Use `BREAKING CHANGE:` in the footer or append `!` after the type/scope to indicate breaking changes:

```text
feat(api)!: change configuration file format

BREAKING CHANGE: The config file now uses YAML instead of JSON
```

### Changelog

The project maintains an automated changelog generated from git commit history. The changelog is located in `CHANGELOG.md` and follows the [Keep a Changelog](https://keepachangelog.com/en/1.0.0/) format.

**How commits appear in the changelog**:

- Commits are automatically categorized by type (Features, Bug Fixes, Documentation, etc.)
- The commit subject (description) appears in the changelog
- Commits with scopes show the scope as a prefix (e.g., `**ci:** fix workflow`)
- Breaking changes are highlighted separately

**Generating the changelog locally**:

```bash
# Install git-chglog
go install github.com/git-chglog/git-chglog/cmd/git-chglog@latest

# Generate changelog
export PATH="$PATH:$(go env GOPATH)/bin"
git-chglog --output CHANGELOG.md
```

**Automatic changelog updates**:

- The changelog is automatically generated and updated during the release process
- When a new release is published, the changelog is regenerated and committed to the repository
- GoReleaser uses git-chglog to generate release notes from the changelog

**Release process**:

1. Create a new git tag (e.g., `v1.2.6`)
2. Push the tag to trigger the release workflow
3. The workflow will:
   - Generate the changelog from commits since the last release
   - Commit the updated changelog to the repository
   - Use GoReleaser to create the GitHub release with changelog entries

## License

See the [LICENSE](LICENSE) file for details.
