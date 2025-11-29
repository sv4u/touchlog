# Example Configuration and Templates

This directory contains example configuration and template files for touchlog.

## Setup Instructions

### 1. Copy the Configuration File

Copy `config.yaml` to your XDG configuration directory:

```bash
# Create the config directory if it doesn't exist
mkdir -p ~/.config/touchlog

# Copy the example config
cp examples/config.yaml ~/.config/touchlog/config.yaml
```

Or if you have `$XDG_CONFIG_HOME` set:

```bash
mkdir -p "$XDG_CONFIG_HOME/touchlog"
cp examples/config.yaml "$XDG_CONFIG_HOME/touchlog/config.yaml"
```

**Important**: Edit the `notes_directory` path in the config file to point to your desired notes directory.

### 2. Copy the Template Files

Copy the template files to your XDG data directory:

```bash
# Create the templates directory if it doesn't exist
mkdir -p ~/.local/share/touchlog/templates

# Copy all template files
cp examples/templates/*.md ~/.local/share/touchlog/templates/
```

Or if you have `$XDG_DATA_HOME` set:

```bash
mkdir -p "$XDG_DATA_HOME/touchlog/templates"
cp examples/templates/*.md "$XDG_DATA_HOME/touchlog/templates/"
```

## Available Templates

The example templates include:

- **daily.md** - Simple daily note template with sections for events, thoughts, and tasks
- **meeting.md** - Structured meeting notes with agenda, decisions, and action items
- **book-note.md** - Template for taking notes on books with quotes and reflections
- **zettelkasten.md** - Zettelkasten-style note template with references and related notes
- **project-note.md** - Project tracking template with status updates and next actions
- **journal.md** - Personal journal entry template with morning/evening reflections

## Customizing Templates

Templates support variable substitution using `{{variable}}` syntax. The following default variables are available:

- `{{date}}` - Current date (default format: `YYYY-MM-DD`)
- `{{time}}` - Current time (default format: `HH:MM:SS`)
- `{{datetime}}` - Current date and time (default format: `YYYY-MM-DD HH:MM:SS`)

You can create your own templates by:

1. Creating a new `.md` file in the templates directory
2. Adding it to the `templates` list in `config.yaml`
3. Using `{{variable}}` placeholders as needed

## Customizing Date/Time Formats

You can customize the format of date, time, and datetime variables in your `config.yaml` file. The format uses Go's time format syntax (reference time: `Mon Jan 2 15:04:05 MST 2006`).

**Example configuration**:

```yaml
datetime_vars:
  date:
    enabled: true
    format: "January 2, 2006"      # Results in: "January 15, 2024"
  time:
    enabled: true
    format: "3:04 PM"              # Results in: "2:30 PM"
  datetime:
    enabled: true
    format: "2006-01-02T15:04:05" # Results in: "2024-01-15T14:30:22"
```

**Common format examples**:

- `"2006-01-02"` - ISO date: `2024-01-15`
- `"01/02/2006"` - US date: `01/15/2024`
- `"2 Jan 2006"` - Short date: `15 Jan 2024`
- `"15:04:05"` - 24-hour time: `14:30:22`
- `"3:04 PM"` - 12-hour time: `2:30 PM`
- `"2006-01-02 15:04:05"` - Full datetime: `2024-01-15 14:30:22`

You can disable specific variables by setting `enabled: false`. If `datetime_vars` is not specified, all variables are enabled with default formats.

## Custom Variables

You can define custom static variables in your `config.yaml` file that will be available in all templates:

```yaml
variables:
  author: "Your Name"
  project: "My Project"
  location: "Home Office"
  team: "Engineering"
```

These variables can be used in templates using `{{author}}`, `{{project}}`, `{{location}}`, etc.

**Example template using custom variables**:

```markdown
# {{project}} - {{date}}

Author: {{author}}
Location: {{location}}

## Notes
- 
```

**Important**: Variable names cannot conflict with reserved names (`date`, `time`, `datetime`). Custom variables can override default variables, but it's recommended to use unique names for custom variables.

## Vim Keymap Support

You can enable vim-style keybindings by setting `vim_mode: true` in your `config.yaml`:

```yaml
vim_mode: true
```

When vim mode is enabled, the application uses vim keybindings:

- **Template selection**: `j`/`k` to navigate, `Enter` to select, `q` to quit
- **Note editing**:
  - Press `i` or `a` to enter insert mode
  - Press `Esc` to exit insert mode (normal mode)
  - In normal mode: basic vim commands are available
  - `:w` to save, `:q` to quit
  - The mode indicator shows "-- INSERT --" or "-- NORMAL --" at the bottom

**Note**: Vim mode provides essential vim commands. Advanced features like full cursor movement in normal mode, yank/put operations, and undo are simplified or not yet implemented.

## Example Template

Here's a simple example template:

```markdown
# Note - {{date}}

## Title
- 

## Content
- 

## Tags
#example
```

When you select this template, `{{date}}` will be automatically replaced with the current date (e.g., `2024-01-15`).
