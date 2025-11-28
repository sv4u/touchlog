# touchlog

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

Create `~/.config/touchlog/config.yaml`:

```yaml
templates:
  - name: "Daily Note"
    file: "daily.md"
  - name: "Meeting Notes"
    file: "meeting.md"
  - name: "Book Note"
    file: "book-note.md"
notes_directory: "/path/to/your/notes"
```

### Template Files

Create template files in `~/.local/share/touchlog/templates/`. Templates support variable substitution using `{{variable}}` syntax.

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

- `{{date}}` - Current date in format `YYYY-MM-DD`
- `{{time}}` - Current time in format `HH:MM:SS`
- `{{datetime}}` - Current date and time in format `YYYY-MM-DD HH:MM:SS`

## Usage

1. **Start the application**:

   ```bash
   touchlog
   ```

2. **Select a template**: Use arrow keys to navigate and press `Enter` to select

3. **Edit your note**: The template will be loaded with variables substituted. Edit as needed.

4. **Save the note**: Press `Ctrl+S` to save. The note will be saved as a Markdown file with a timestamp-based filename (e.g., `2024-01-15-143022.md`) in your configured notes directory.

5. **Quit**: Press `Ctrl+C` or `q` to exit

## Keyboard Shortcuts

- `↑/↓` - Navigate template list
- `Enter` - Select template
- `Ctrl+S` - Save note
- `Ctrl+C` or `q` - Quit application

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

## License

See the [LICENSE](LICENSE) file for details.
