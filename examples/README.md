# Example Configuration and Templates

This directory contains example configuration and template files for touchlog.

## Configuration File Examples

### YAML Configuration Files

#### `config-simple.yaml` - Minimal Configuration
A minimal configuration with just the essentials:
- Notes directory
- Basic file-based templates

**Use this when**: You want a simple setup with minimal configuration.

#### `config-full.yaml` - Complete Configuration
A comprehensive configuration demonstrating all available features:
- File-based templates
- Inline templates (Phase 2)
- Timezone configuration (Phase 2)
- Custom variables
- Date/time format customization
- Vim mode settings

**Use this when**: You want to see all available options and features.

#### `config-inline-templates.yaml` - Inline Templates Example
Showcases the inline templates feature (Phase 2):
- Templates defined directly in the config file
- No separate template files needed
- Easy to manage and version control

**Use this when**: You prefer managing templates in your config file rather than separate files.

#### `config.yaml` - Standard Configuration (Legacy)
The original example configuration file. Still valid and supported.

### TOML Configuration Files

**Note**: TOML support is planned for Phase 2.2 (not yet implemented). These examples are provided for forward compatibility.

#### `config-simple.toml` - Minimal TOML Configuration
Minimal configuration in TOML format.

#### `config-full.toml` - Complete TOML Configuration
Complete configuration in TOML format with all features.

## Setup Instructions

### 1. Choose and Copy a Configuration File

Choose one of the example configuration files and copy it to your XDG configuration directory:

```bash
# Create the config directory if it doesn't exist
mkdir -p ~/.config/touchlog

# Copy the example config (choose one)
cp examples/config-simple.yaml ~/.config/touchlog/config.yaml
# OR
cp examples/config-full.yaml ~/.config/touchlog/config.yaml
# OR
cp examples/config-inline-templates.yaml ~/.config/touchlog/config.yaml
```

Or if you have `$XDG_CONFIG_HOME` set:

```bash
mkdir -p "$XDG_CONFIG_HOME/touchlog"
cp examples/config-simple.yaml "$XDG_CONFIG_HOME/touchlog/config.yaml"
```

**Important**: Edit the `notes_directory` path in the config file to point to your desired notes directory.

### 2. Copy Template Files (if using file-based templates)

If your configuration uses file-based templates, copy the template files to your XDG data directory:

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

**Note**: If you're using inline templates (defined in the config file), you don't need to copy template files.

## Available Templates

### File-Based Templates

These templates are stored as separate `.md` files in the templates directory:

- **daily.md** - Simple daily note template with sections for events, thoughts, and tasks
- **meeting.md** - Structured meeting notes with agenda, decisions, and action items
- **book-note.md** - Template for taking notes on books with quotes and reflections
- **zettelkasten.md** - Zettelkasten-style note template with references and related notes
- **project-note.md** - Project tracking template with status updates and next actions
- **journal.md** - Personal journal entry template with morning/evening reflections
- **standup.md** - Standup meeting template with yesterday/today/blockers sections
- **code-review.md** - Code review template with feedback and action items
- **quick-note.md** - Minimal template for quick note capture

### Example Templates Demonstrating Features

- **with-timezone.md** - Demonstrates timezone-aware date/time formatting
- **with-metadata.md** - Demonstrates metadata variables (Phase 7 feature - coming soon)
- **with-custom-vars.md** - Demonstrates custom variables from config

## Inline Templates (Phase 2 Feature)

Inline templates allow you to define templates directly in your configuration file, eliminating the need for separate template files.

### Example Inline Template Configuration

```yaml
inline_templates:
  quick-note: |
    # Quick Note - {{date}} {{time}}
    
    {{message}}
    
    Tags: {{tags}}
  standup: |
    # Standup - {{date}}
    
    ## What I did yesterday
    {{message}}
    
    ## What I'm doing today
    - 
    
    ## Blockers
    - 
```

### Benefits of Inline Templates

- **Easy version control**: Templates are in the same file as your config
- **No file management**: No need to manage separate template files
- **Quick setup**: Everything in one place
- **Precedence**: Inline templates take precedence over file-based templates with the same name

### When to Use Inline Templates

- Small, simple templates
- Templates you want to version control with your config
- Quick prototyping and experimentation
- Templates that don't need to be shared across multiple config files

### When to Use File-Based Templates

- Large, complex templates
- Templates you want to reuse across multiple config files
- Templates that benefit from syntax highlighting in separate files
- Templates you want to share with others

## Timezone Configuration (Phase 2 Feature)

You can configure the timezone used for date/time formatting in your templates:

```yaml
timezone: "America/Denver"  # Optional, defaults to system timezone
```

### Supported Timezone Formats

Use IANA timezone database names:
- `"UTC"` - Coordinated Universal Time
- `"America/New_York"` - Eastern Time
- `"America/Chicago"` - Central Time
- `"America/Denver"` - Mountain Time
- `"America/Los_Angeles"` - Pacific Time
- `"Europe/London"` - British Time
- `"Europe/Paris"` - Central European Time
- `"Asia/Tokyo"` - Japan Standard Time

### Example

```yaml
timezone: "America/Denver"
datetime_vars:
  date:
    enabled: true
    format: "2006-01-02"
  time:
    enabled: true
    format: "15:04:05"
```

All date/time variables in templates will be formatted in the specified timezone.

## Customizing Templates

Templates support variable substitution using `{{variable}}` syntax.

### Default Variables

- `{{date}}` - Current date (default format: `YYYY-MM-DD`)
- `{{time}}` - Current time (default format: `HH:MM:SS`)
- `{{datetime}}` - Current date and time (default format: `YYYY-MM-DD HH:MM:SS`)

### User-Provided Variables (Phase 4)

- `{{title}}` - Title for the log entry
- `{{message}}` - Message content for the log entry
- `{{tags}}` - Tags for the log entry (comma-separated)

### Metadata Variables (Phase 7 - Coming Soon)

- `{{user}}` - Current username
- `{{host}}` - Hostname
- `{{branch}}` - Git branch (if in git repo and `--include-git` flag used)
- `{{commit}}` - Git commit hash (if in git repo and `--include-git` flag used)

### Custom Variables

Define your own static variables in the `variables` section of your config:

```yaml
variables:
  author: "Your Name"
  project: "My Project"
  location: "Home Office"
  team: "Engineering"
```

These can be used in templates as `{{author}}`, `{{project}}`, etc.

**Important**: Variable names cannot conflict with reserved names (`date`, `time`, `datetime`).

## Customizing Date/Time Formats

You can customize the format of date, time, and datetime variables in your config file. The format uses Go's time format syntax (reference time: `Mon Jan 2 15:04:05 MST 2006`).

### Example Configuration

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

### Common Format Examples

- `"2006-01-02"` - ISO date: `2024-01-15`
- `"01/02/2006"` - US date: `01/15/2024`
- `"2 Jan 2006"` - Short date: `15 Jan 2024`
- `"15:04:05"` - 24-hour time: `14:30:22`
- `"3:04 PM"` - 12-hour time: `2:30 PM`
- `"2006-01-02 15:04:05"` - Full datetime: `2024-01-15 14:30:22`

You can disable specific variables by setting `enabled: false`. If `datetime_vars` is not specified, all variables are enabled with default formats.

## Default Template

You can specify a default template to use when creating entries:

```yaml
default_template: "daily"
```

Or using the alternative format (backward compatible):

```yaml
template:
  name: "daily"
```

The `default_template` field takes precedence over `template.name` if both are specified.

## Vim Keymap Support

You can enable vim-style keybindings by setting `vim_mode: true` in your config:

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

## Configuration File Discovery

touchlog searches for configuration files in the following order:

1. **Explicit path**: `--config` flag (e.g., `touchlog --config /path/to/config.yaml`)
2. **Current directory**: `./touchlog.yaml` or `./touchlog.toml`
3. **XDG config directory**: `$XDG_CONFIG_HOME/touchlog/config.yaml` (or `~/.config/touchlog/config.yaml`)

If no config file is found, defaults are used (no error).

## Configuration Precedence

Configuration values are resolved with the following precedence (highest to lowest):

1. **CLI flags** - Command-line flags override everything
2. **Config file** - Values from the config file
3. **Defaults** - Built-in default values

Example: If you set `--output /custom/path` on the command line, it will override the `notes_directory` value in your config file.

## Strict Mode Validation

You can validate your configuration file for unknown keys using strict mode:

```bash
touchlog config --strict
```

This will reject any unknown configuration keys and help you catch typos or unsupported options.

## TOML Support (Coming Soon)

TOML configuration support is planned for Phase 2.2. Example TOML configuration files are provided in this directory for forward compatibility:

- `config-simple.toml` - Minimal TOML configuration
- `config-full.toml` - Complete TOML configuration

When TOML support is implemented, you'll be able to use `.toml` files just like `.yaml` files.

## Creating Your Own Templates

### File-Based Templates

1. Create a new `.md` file in your templates directory
2. Add it to the `templates` list in your config file
3. Use `{{variable}}` placeholders as needed

Example:

```markdown
# My Custom Template - {{date}}

## Content
{{message}}

## Tags
{{tags}}
```

### Inline Templates

Add templates directly to your config file:

```yaml
inline_templates:
  my-template: |
    # My Custom Template - {{date}}
    
    ## Content
    {{message}}
    
    ## Tags
    {{tags}}
```

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
