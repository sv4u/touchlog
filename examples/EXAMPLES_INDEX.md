# Examples Index

This document provides a quick reference to all example files in this directory.

## Configuration Files

### YAML Examples

| File | Description | Use Case |
|------|-------------|----------|
| `config-simple.yaml` | Minimal configuration with essentials | Quick setup, minimal customization |
| `config.yaml` | Standard configuration with Phase 2 features | Balanced example with new features |
| `config-full.yaml` | Complete configuration with all features | Reference for all available options |
| `config-inline-templates.yaml` | Configuration using inline templates | Managing templates in config file |

### TOML Examples (Phase 2.2 - Coming Soon)

| File | Description | Use Case |
|------|-------------|----------|
| `config-simple.toml` | Minimal TOML configuration | Quick setup in TOML format |
| `config-full.toml` | Complete TOML configuration | Reference for all options in TOML |

**Note**: TOML support is planned for Phase 2.2 and is not yet implemented. These examples are provided for forward compatibility.

## Template Files

### Standard Templates

| File | Description |
|------|-------------|
| `daily.md` | Simple daily note template |
| `meeting.md` | Structured meeting notes |
| `book-note.md` | Book notes with quotes and reflections |
| `zettelkasten.md` | Zettelkasten-style note template |
| `project-note.md` | Project tracking template |
| `journal.md` | Personal journal entry template |
| `standup.md` | Standup meeting template |
| `code-review.md` | Code review template |
| `quick-note.md` | Minimal quick note template |

### Feature Demonstration Templates

| File | Description | Feature Demonstrated |
|------|-------------|---------------------|
| `with-timezone.md` | Timezone-aware date/time formatting | Phase 2: Timezone configuration |
| `with-metadata.md` | Metadata variables (user, host, git) | Phase 7: Metadata capture (coming soon) |
| `with-custom-vars.md` | Custom variables from config | Custom variables feature |

## Quick Start Guide

1. **Choose a configuration file** based on your needs:
   - Start with `config-simple.yaml` for minimal setup
   - Use `config-full.yaml` to see all options
   - Try `config-inline-templates.yaml` to use inline templates

2. **Copy to your config directory**:
   ```bash
   cp examples/config-simple.yaml ~/.config/touchlog/config.yaml
   ```

3. **Edit the configuration**:
   - Set `notes_directory` to your desired path
   - Customize templates, variables, and other settings

4. **Copy template files** (if using file-based templates):
   ```bash
   cp examples/templates/*.md ~/.local/share/touchlog/templates/
   ```

5. **Validate your configuration**:
   ```bash
   touchlog config --strict
   ```

## Feature Highlights

### Phase 2 Features (Implemented)

- ✅ **Inline Templates**: Define templates directly in config file
- ✅ **Timezone Configuration**: Configure timezone for date/time formatting
- ✅ **Default Template**: Specify default template to use
- ✅ **Configuration Precedence**: CLI flags > Config > Defaults
- ✅ **Strict Mode**: Validate config for unknown keys

### Future Features (Examples Provided)

- 🔜 **TOML Support**: TOML configuration format (Phase 2.2)
- 🔜 **Metadata Variables**: User, host, git context (Phase 7)

## See Also

- [README.md](./README.md) - Comprehensive documentation
- [../.cursor/plans/02-configuration.md](../.cursor/plans/02-configuration.md) - Phase 2 implementation details

