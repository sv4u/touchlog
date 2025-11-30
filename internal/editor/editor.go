package editor

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/sv4u/touchlog/internal/config"
	"github.com/sv4u/touchlog/internal/template"
	"github.com/sv4u/touchlog/internal/xdg"
)

// state represents the current application state
// Using a custom type for type safety
type state int

const (
	stateSelectTemplate state = iota // iota auto-increments: 0, 1, 2, 3...
	stateLoadingTemplate
	stateEditNote
	stateSaving
	stateError
)

// vimState represents the vim mode state (normal or insert)
type vimState int

const (
	vimNormal vimState = iota
	vimInsert
)

// model represents the application state
// In Bubble Tea, the model holds all state
type model struct {
	state             state             // Current application state
	config            *config.Config    // Loaded configuration
	templateList      list.Model        // Bubbles list component
	textarea          textarea.Model    // Bubbles textarea component
	selectedTemplate  *config.Template  // Currently selected template
	noteContent       string            // Current note content
	variables         map[string]string // Template variables
	err               error             // Error state
	width             int               // Terminal width
	height            int               // Terminal height
	vimMode           bool              // Whether vim mode is enabled
	vimState          vimState          // Current vim state (normal/insert)
	commandBuffer     string            // Buffer for vim commands (e.g., :w, :q)
	outputDirOverride string            // Override for output directory (from CLI or API)
}

// modelConfig holds configuration options for NewModel
type modelConfig struct {
	outputDirectory string
}

// ModelOption is a function that configures a modelConfig
type ModelOption func(*modelConfig)

// WithOutputDirectory sets the output directory override
func WithOutputDirectory(dir string) ModelOption {
	return func(cfg *modelConfig) {
		cfg.outputDirectory = dir
	}
}

// NewModel creates and initializes a new editor model
// This is called when the application starts
// It accepts optional ModelOption functions to configure the model
func NewModel(opts ...ModelOption) (tea.Model, error) {
	// Apply options to default config
	cfg := &modelConfig{}
	for _, opt := range opts {
		opt(cfg)
	}

	// Load configuration
	configPath, err := xdg.ConfigFilePath()
	if err != nil {
		return nil, fmt.Errorf("failed to get config path: %w", err)
	}

	// Load or create config (auto-creates default config if missing)
	configCfg, err := config.LoadOrCreateConfig(configPath)
	if err != nil {
		// Only non-existence errors are handled by LoadOrCreateConfig
		// Other errors (permission, invalid YAML, etc.) are returned as-is
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Optionally create example templates if templates directory is empty
	templatesDir, err := xdg.TemplatesDir()
	if err != nil {
		// Log warning but don't fail - templates directory creation is optional
		// Could use a logger here, but for now we'll just continue
		// In a future enhancement, we could add logging
	} else {
		// Try to create templates (non-fatal if it fails)
		if err := template.CreateExampleTemplates(templatesDir); err != nil {
			// Template creation failed, but don't fail the entire application
			// This is a nice-to-have feature, not critical
			// Could log this error in the future
		}
	}

	// Create template list items
	// In Go, we need to convert []config.Template to []list.Item
	// list.Item is an interface that requires Item() methods
	items := make([]list.Item, 0)
	for _, tmpl := range configCfg.GetTemplates() {
		items = append(items, templateItem{
			title:       tmpl.Name,
			description: tmpl.File,
			file:        tmpl.File,
		})
	}

	// Validate that at least one template is configured
	if len(items) == 0 {
		return nil, fmt.Errorf("no templates configured. Please add at least one template to your config file at '%s'", configPath)
	}

	// Initialize list component
	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Select a Template"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)

	// Initialize textarea component
	ta := textarea.New()
	ta.Placeholder = "Start typing your note..."
	ta.CharLimit = 0 // No limit
	ta.SetWidth(80)
	ta.SetHeight(20)

	// Return initial model
	return model{
		state:             stateSelectTemplate,
		config:            configCfg,
		templateList:      l,
		textarea:          ta,
		variables:         template.GetDefaultVariables(configCfg),
		vimMode:           configCfg.GetVimMode(),
		vimState:          vimNormal, // Start in normal mode when vim mode is enabled
		outputDirOverride: cfg.outputDirectory,
	}, nil
}

// Init is called when the model is first created
// It can return an initial command to run
func (m model) Init() tea.Cmd {
	return nil
}

// templateItem implements the list.Item interface
// In Go, interfaces are satisfied implicitly
type templateItem struct {
	title       string
	description string
	file        string
}

// These methods satisfy the list.Item interface
func (i templateItem) FilterValue() string { return i.title }
func (i templateItem) Title() string       { return i.title }
func (i templateItem) Description() string { return i.description }

// Update handles messages and updates the model
// This is the core of the MVU pattern
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Type switch: Go's way to handle different message types
	switch msg := msg.(type) {

	// Handle window resize
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.templateList.SetWidth(msg.Width - 4)
		m.templateList.SetHeight(msg.Height - 4)
		m.textarea.SetWidth(msg.Width - 4)
		m.textarea.SetHeight(msg.Height - 6)
		return m, nil

	// Handle custom messages
	case templateLoadedMsg:
		// Template loaded successfully, transition to edit state
		m.state = stateEditNote
		m.noteContent = msg.content
		m.textarea.SetValue(msg.content)
		// Set vim state: if vim mode is enabled, start in normal mode, otherwise focus textarea
		if m.vimMode {
			m.vimState = vimNormal
			m.textarea.Blur()
		} else {
			m.textarea.Focus()
		}
		return m, nil

	case noteSavedMsg:
		// Note saved successfully, quit the application
		return m, tea.Quit

	case errMsg:
		// Error occurred, transition to error state
		m.state = stateError
		m.err = msg.err
		return m, nil

	// Handle keyboard input
	case tea.KeyMsg:
		key := msg.String()

		// Handle vim mode keybindings
		if m.vimMode {
			// Template selection state vim keybindings
			if m.state == stateSelectTemplate {
				switch key {
				case "j":
					// Move down
					m.templateList.CursorDown()
					return m, nil
				case "k":
					// Move up
					m.templateList.CursorUp()
					return m, nil
				case "enter":
					// Select template
					selected := m.templateList.SelectedItem()
					if selected != nil {
						item := selected.(templateItem)
						m.selectedTemplate = &config.Template{
							Name: item.title,
							File: item.file,
						}
						m.state = stateLoadingTemplate
						return m, m.loadTemplateCmd()
					}
				case "q":
					// Quit
					return m, tea.Quit
				}
			}

			// Note editing state vim keybindings
			if m.state == stateEditNote {
				// Handle command mode (when typing :)
				if m.commandBuffer != "" {
					if key == "enter" {
						// Execute command
						switch m.commandBuffer {
						case ":w":
							// Save
							m.state = stateSaving
							m.commandBuffer = ""
							return m, m.saveNoteCmd()
						case ":q":
							// Quit
							m.commandBuffer = ""
							return m, tea.Quit
						default:
							// Unknown command, clear buffer
							m.commandBuffer = ""
						}
					} else if key == "esc" {
						// Cancel command
						m.commandBuffer = ""
					} else if len(key) == 1 && key[0] >= 32 && key[0] <= 126 {
						// Add character to command buffer
						m.commandBuffer += key
					}
					return m, nil
				}

				// Normal mode keybindings
				if m.vimState == vimNormal {
					switch key {
					case "i", "a":
						// Enter insert mode
						m.vimState = vimInsert
						m.textarea.Focus()
						return m, nil
					case "h":
						// Move left (handled by textarea in insert mode, but we can move cursor in normal mode)
						// For now, just enter insert mode
						m.vimState = vimInsert
						m.textarea.Focus()
						return m, nil
					case "j":
						// Move down - simplified: enter insert mode
						m.vimState = vimInsert
						m.textarea.Focus()
						return m, nil
					case "k":
						// Move up
						m.vimState = vimInsert
						m.textarea.Focus()
						return m, nil
					case "l":
						// Move right
						m.vimState = vimInsert
						m.textarea.Focus()
						return m, nil
					case "w", "b", "0", "$":
						// Word movement and line start/end - simplified: enter insert mode
						m.vimState = vimInsert
						m.textarea.Focus()
						return m, nil
					case "dd":
						// Delete current line based on cursor position
						lines := strings.Split(m.noteContent, "\n")
						if len(lines) > 0 {
							// Get current line from cursor position
							// Line() returns 0-indexed line number
							currentLine := m.textarea.Line()
							if currentLine < 0 {
								currentLine = 0
							}
							if currentLine >= len(lines) {
								currentLine = len(lines) - 1
							}
							// Remove the current line
							if len(lines) > 1 {
								newLines := make([]string, 0, len(lines)-1)
								newLines = append(newLines, lines[:currentLine]...)
								newLines = append(newLines, lines[currentLine+1:]...)
								m.noteContent = strings.Join(newLines, "\n")
								m.textarea.SetValue(m.noteContent)
								// Move cursor to the same line (or last line if we deleted the last line)
								if currentLine >= len(newLines) && len(newLines) > 0 {
									// Deleted the last line, move to new last line
									// Calculate character position for the start of the new last line
									charPos := 0
									for i := 0; i < len(newLines)-1; i++ {
										charPos += len(newLines[i]) + 1 // +1 for newline
									}
									m.textarea.SetCursor(charPos)
								} else if len(newLines) > 0 {
									// Calculate character position for the start of current line
									charPos := 0
									for i := 0; i < currentLine; i++ {
										charPos += len(newLines[i]) + 1 // +1 for newline
									}
									m.textarea.SetCursor(charPos)
								}
							} else {
								m.noteContent = ""
								m.textarea.SetValue(m.noteContent)
								m.textarea.SetCursor(0)
							}
						}
						return m, nil
					case "yy":
						// Yank line - simplified: copy current line (not implemented fully)
						return m, nil
					case "p", "P":
						// Paste - simplified: not implemented
						return m, nil
					case "u":
						// Undo - not implemented (would need undo stack)
						return m, nil
					case ":":
						// Enter command mode
						m.commandBuffer = ":"
						return m, nil
					case "ctrl+s":
						// Save (works in both modes)
						m.state = stateSaving
						return m, m.saveNoteCmd()
					}
				} else {
					// Insert mode
					if key == "esc" {
						// Exit insert mode
						m.vimState = vimNormal
						m.textarea.Blur()
						return m, nil
					}
					// In insert mode, delegate to textarea (handled below)
				}
			}
		}

		// Non-vim mode or fallback keybindings
		switch key {
		case "ctrl+c":
			// Quit the application
			return m, tea.Quit
		case "q":
			// Quit (only if not in vim mode or in vim normal mode)
			if !m.vimMode || (m.state == stateEditNote && m.vimState == vimNormal) {
				return m, tea.Quit
			}

		case "enter":
			// Handle Enter key based on current state
			if m.state == stateSelectTemplate && !m.vimMode {
				// Get selected item from list
				selected := m.templateList.SelectedItem()
				if selected != nil {
					item := selected.(templateItem) // Type assertion
					m.selectedTemplate = &config.Template{
						Name: item.title,
						File: item.file,
					}
					// Transition to loading state to prevent concurrent loads
					m.state = stateLoadingTemplate
					// Return a command to load the template
					return m, m.loadTemplateCmd()
				}
			}

		case "ctrl+s":
			// Save note (only in edit state)
			if m.state == stateEditNote {
				m.state = stateSaving
				return m, m.saveNoteCmd()
			}
		}
	}

	// Delegate to component updates based on state
	switch m.state {
	case stateSelectTemplate:
		// Update the list component
		var cmd tea.Cmd
		m.templateList, cmd = m.templateList.Update(msg)
		return m, cmd

	case stateLoadingTemplate:
		// Ignore input while loading template
		return m, nil

	case stateEditNote:
		// Update the textarea component only in insert mode when vim mode is enabled
		if m.vimMode && m.vimState == vimNormal {
			// In normal mode, don't delegate to textarea
			return m, nil
		}
		// In insert mode or non-vim mode, delegate to textarea
		var cmd tea.Cmd
		m.textarea, cmd = m.textarea.Update(msg)
		m.noteContent = m.textarea.Value()
		return m, cmd

	case stateSaving:
		// Handle save completion
		// (Implementation details for save messages)
		return m, nil

	case stateError:
		// Handle error state
		return m, nil
	}

	return m, nil
}

// View renders the UI
func (m model) View() string {
	switch m.state {
	case stateSelectTemplate:
		return m.templateList.View()

	case stateLoadingTemplate:
		return "Loading template..."

	case stateEditNote:
		helpText := "Press Ctrl+S to save, Ctrl+C to quit"
		if m.vimMode {
			// Show vim mode indicator
			modeIndicator := "-- NORMAL --"
			if m.vimState == vimInsert {
				modeIndicator = "-- INSERT --"
			}
			if m.commandBuffer != "" {
				modeIndicator = m.commandBuffer
			}
			helpText = fmt.Sprintf("%s | %s", modeIndicator, helpText)
		}
		return fmt.Sprintf(
			"%s\n\n%s\n\n%s",
			m.textarea.View(),
			helpStyle.Render("Editing note..."),
			helpStyle.Render(helpText),
		)

	case stateSaving:
		return "Saving note..."

	case stateError:
		return fmt.Sprintf("Error: %v\n\nPress Ctrl+C to quit", m.err)
	}

	return ""
}

// loadTemplateCmd returns a command that loads a template asynchronously
// Commands in Bubble Tea are functions that return tea.Cmd
func (m model) loadTemplateCmd() tea.Cmd {
	return func() tea.Msg {
		// This runs in a goroutine (async)
		templatesDir, err := xdg.TemplatesDir()
		if err != nil {
			return errMsg{err}
		}

		content, err := template.LoadTemplate(templatesDir, m.selectedTemplate.File)
		if err != nil {
			return errMsg{err}
		}

		// Process template with variables
		processed := template.ProcessTemplate(content, m.variables)

		return templateLoadedMsg{content: processed}
	}
}

// expandPath expands ~ to the user's home directory
func expandPath(path string) (string, error) {
	if strings.HasPrefix(path, "~") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get home directory: %w", err)
		}
		if path == "~" {
			return homeDir, nil
		}
		// Skip the leading ~ and the following /
		remaining := path[2:]
		// Strip leading / if present to avoid issues with filepath.Join
		// This handles cases like "~~/path" where path[2:] would be "/path"
		remaining = strings.TrimPrefix(remaining, "/")
		return filepath.Join(homeDir, remaining), nil
	}
	return path, nil
}

// saveNoteCmd returns a command that saves the note
func (m model) saveNoteCmd() tea.Cmd {
	return func() tea.Msg {
		// Priority: CLI/API override > config file
		var notesDir string
		if m.outputDirOverride != "" {
			notesDir = m.outputDirOverride
		} else {
			notesDir = m.config.GetNotesDirectory()
		}

		if notesDir == "" {
			return errMsg{fmt.Errorf("notes directory not configured")}
		}

		// Expand ~ in path if present
		expandedDir, err := expandPath(notesDir)
		if err != nil {
			return errMsg{fmt.Errorf("failed to expand notes directory path: %w", err)}
		}

		// Generate filename from timestamp
		filename := fmt.Sprintf("%s.md", time.Now().Format("2006-01-02-150405"))
		fullPath := filepath.Join(expandedDir, filename)

		// Create directory if it doesn't exist
		if _, err := os.Stat(expandedDir); os.IsNotExist(err) {
			// Directory doesn't exist, create it recursively
			if err := os.MkdirAll(expandedDir, 0755); err != nil {
				return errMsg{fmt.Errorf("failed to create notes directory '%s': %w", expandedDir, err)}
			}
		} else if err != nil {
			// Some other error checking directory
			return errMsg{fmt.Errorf("failed to check notes directory '%s': %w", expandedDir, err)}
		}

		// Write the file
		if err := os.WriteFile(fullPath, []byte(m.noteContent), 0644); err != nil {
			return errMsg{fmt.Errorf("failed to save note to '%s': %w", fullPath, err)}
		}

		return noteSavedMsg{filepath: fullPath}
	}
}

// Custom message types for state transitions
type templateLoadedMsg struct {
	content string
}

type noteSavedMsg struct {
	filepath string
}

type errMsg struct {
	err error
}

// Styling
var helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
