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

- `{{date}}` - Current date in format `YYYY-MM-DD`
- `{{time}}` - Current time in format `HH:MM:SS`
- `{{datetime}}` - Current date and time in format `YYYY-MM-DD HH:MM:SS`

You can create your own templates by:

1. Creating a new `.md` file in the templates directory
2. Adding it to the `templates` list in `config.yaml`
3. Using `{{variable}}` placeholders as needed

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

