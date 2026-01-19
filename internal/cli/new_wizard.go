package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mattn/go-isatty"
	"github.com/sv4u/touchlog/v2/internal/config"
	"github.com/sv4u/touchlog/v2/internal/model"
)

// wizardState represents the current step in the wizard
type wizardState int

const (
	stateSelectType wizardState = iota
	stateInputKey
	stateInputTitle
	stateInputTags
	stateInputState
	stateVerification
	stateComplete
	stateError
)

// wizardModel represents the bubbletea model for the new note wizard
type wizardModel struct {
	vaultRoot string
	cfg       *config.Config
	state     wizardState
	err       error

	// Collected data
	typeName model.TypeName
	typeDef  config.TypeDef
	key      model.Key
	title    string
	tags     []string
	stateVal string

	// UI state
	typeCursor  int
	typeChoices []model.TypeName
	keyInput    string
	keyError    string
	titleInput  string
	tagsInput   string
	stateInput  string
	verifying   bool
	notePath    string
	noteID      model.NoteID
}

// initialModel creates the initial wizard model
func initialModel(vaultRoot string, cfg *config.Config) wizardModel {
	// Build type choices
	typeChoices := make([]model.TypeName, 0, len(cfg.Types))
	for typeName := range cfg.Types {
		typeChoices = append(typeChoices, typeName)
	}

	// Build state choices (for now, just use default state from selected type)
	// We'll populate this when a type is selected

	return wizardModel{
		vaultRoot:   vaultRoot,
		cfg:         cfg,
		state:       stateSelectType,
		typeChoices: typeChoices,
		keyInput:    "",
		titleInput:  "",
		tagsInput:   "",
		stateInput:  "",
	}
}

// Init initializes the wizard model
func (m wizardModel) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model
func (m wizardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "esc":
			if m.state != stateSelectType {
				// Go back to previous state
				return m.goBack(), nil
			}
			return m, tea.Quit
		}

		switch m.state {
		case stateSelectType:
			return m.updateTypeSelection(msg)
		case stateInputKey:
			return m.updateKeyInput(msg)
		case stateInputTitle:
			return m.updateTitleInput(msg)
		case stateInputTags:
			return m.updateTagsInput(msg)
		case stateInputState:
			return m.updateStateInput(msg)
		case stateVerification:
			return m.updateVerification(msg)
		}
	}

	return m, nil
}

// View renders the UI
func (m wizardModel) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\n\nPress q to quit.", m.err)
	}

	switch m.state {
	case stateSelectType:
		return m.viewTypeSelection()
	case stateInputKey:
		return m.viewKeyInput()
	case stateInputTitle:
		return m.viewTitleInput()
	case stateInputTags:
		return m.viewTagsInput()
	case stateInputState:
		return m.viewStateInput()
	case stateVerification:
		return m.viewVerification()
	case stateComplete:
		return fmt.Sprintf("✓ Note created successfully!\n\nPath:  %s\nID:    %s\nType:  %s\nKey:   %s\nTitle: %s\n", m.notePath, m.noteID, m.typeName, m.key, m.title)
	default:
		return ""
	}
}

// updateTypeSelection handles type selection
func (m wizardModel) updateTypeSelection(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.typeCursor > 0 {
			m.typeCursor--
		}
	case "down", "j":
		if m.typeCursor < len(m.typeChoices)-1 {
			m.typeCursor++
		}
	case "enter", " ":
		selectedType := m.typeChoices[m.typeCursor]
		m.typeName = selectedType
		m.typeDef = m.cfg.Types[selectedType]
		m.stateVal = m.typeDef.DefaultState
		m.stateInput = m.typeDef.DefaultState
		m.state = stateInputKey
	}
	return m, nil
}

// viewTypeSelection renders the type selection screen
func (m wizardModel) viewTypeSelection() string {
	s := strings.Builder{}
	s.WriteString("Select note type:\n\n")

	for i, choice := range m.typeChoices {
		cursor := " "
		if i == m.typeCursor {
			cursor = ">"
		}
		typeDef := m.cfg.Types[choice]
		s.WriteString(fmt.Sprintf("%s %s - %s\n", cursor, choice, typeDef.Description))
	}

	s.WriteString("\n(Use arrow keys to navigate, Enter to select, q to quit)")
	return s.String()
}

// updateKeyInput handles key input
func (m wizardModel) updateKeyInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		// Validate and proceed
		keyStr := strings.TrimSpace(m.keyInput)
		if keyStr == "" {
			m.keyError = "Key cannot be empty"
			return m, nil
		}

		// Validate pattern
		if m.typeDef.KeyPattern != nil {
			if !m.typeDef.KeyPattern.MatchString(keyStr) {
				m.keyError = fmt.Sprintf("Key must match pattern: %s", m.typeDef.KeyPattern.String())
				return m, nil
			}
		}

		// Validate length
		if len(keyStr) > m.typeDef.KeyMaxLen {
			m.keyError = fmt.Sprintf("Key exceeds maximum length of %d", m.typeDef.KeyMaxLen)
			return m, nil
		}

		// Check uniqueness
		notePath := filepath.Join(m.vaultRoot, string(m.typeName), keyStr+".Rmd")
		if _, err := os.Stat(notePath); err == nil {
			m.keyError = fmt.Sprintf("Note with key %q already exists", keyStr)
			return m, nil
		}

		// Valid key
		m.key = model.Key(keyStr)
		m.keyError = ""
		m.state = stateInputTitle
	case "backspace":
		if len(m.keyInput) > 0 {
			m.keyInput = m.keyInput[:len(m.keyInput)-1]
			m.keyError = ""
		}
	default:
		if len(msg.Runes) > 0 {
			m.keyInput += string(msg.Runes)
			m.keyError = ""
		}
	}
	return m, nil
}

// viewKeyInput renders the key input screen
func (m wizardModel) viewKeyInput() string {
	s := strings.Builder{}
	s.WriteString(fmt.Sprintf("Enter key for %s note:\n\n", m.typeName))
	s.WriteString(fmt.Sprintf("> %s", m.keyInput))
	if m.keyError != "" {
		s.WriteString(fmt.Sprintf("\n\nError: %s", m.keyError))
	}
	s.WriteString("\n\n(Enter to continue, Esc to go back, q to quit)")
	return s.String()
}

// updateTitleInput handles title input
func (m wizardModel) updateTitleInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		title := strings.TrimSpace(m.titleInput)
		if title == "" {
			return m, nil
		}
		m.title = title
		m.state = stateInputTags
	case "backspace":
		if len(m.titleInput) > 0 {
			m.titleInput = m.titleInput[:len(m.titleInput)-1]
		}
	default:
		if len(msg.Runes) > 0 {
			m.titleInput += string(msg.Runes)
		}
	}
	return m, nil
}

// viewTitleInput renders the title input screen
func (m wizardModel) viewTitleInput() string {
	s := strings.Builder{}
	s.WriteString("Enter title:\n\n")
	s.WriteString(fmt.Sprintf("> %s", m.titleInput))
	s.WriteString("\n\n(Enter to continue, Esc to go back, q to quit)")
	return s.String()
}

// updateTagsInput handles tags input
func (m wizardModel) updateTagsInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		tagsStr := strings.TrimSpace(m.tagsInput)
		if tagsStr != "" {
			// Parse comma-separated tags
			tagList := strings.Split(tagsStr, ",")
			m.tags = make([]string, 0, len(tagList))
			for _, tag := range tagList {
				tag = strings.TrimSpace(tag)
				if tag != "" {
					m.tags = append(m.tags, tag)
				}
			}
		}
		m.state = stateInputState
	case "backspace":
		if len(m.tagsInput) > 0 {
			m.tagsInput = m.tagsInput[:len(m.tagsInput)-1]
		}
	default:
		if len(msg.Runes) > 0 {
			m.tagsInput += string(msg.Runes)
		}
	}
	return m, nil
}

// viewTagsInput renders the tags input screen
func (m wizardModel) viewTagsInput() string {
	s := strings.Builder{}
	s.WriteString("Enter tags (comma-separated, optional):\n\n")
	s.WriteString(fmt.Sprintf("> %s", m.tagsInput))
	s.WriteString("\n\n(Enter to continue, Esc to go back, q to quit)")
	return s.String()
}

// updateStateInput handles state input
func (m wizardModel) updateStateInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		// Use input or default
		stateStr := strings.TrimSpace(m.stateInput)
		if stateStr == "" {
			stateStr = m.typeDef.DefaultState
		}
		m.stateVal = stateStr
		m.state = stateVerification
	case "backspace":
		if len(m.stateInput) > 0 {
			m.stateInput = m.stateInput[:len(m.stateInput)-1]
		}
	default:
		if len(msg.Runes) > 0 {
			m.stateInput += string(msg.Runes)
		}
	}
	return m, nil
}

// viewStateInput renders the state input screen
func (m wizardModel) viewStateInput() string {
	s := strings.Builder{}
	s.WriteString(fmt.Sprintf("Enter state (optional, default: %s):\n\n", m.typeDef.DefaultState))
	s.WriteString(fmt.Sprintf("> %s", m.stateInput))
	s.WriteString("\n\n(Enter to continue with default or custom state, Esc to go back, q to quit)")
	return s.String()
}

// updateVerification handles verification screen
func (m wizardModel) updateVerification(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter", "y":
		// Create the note
		if m.verifying {
			return m, nil
		}
		m.verifying = true
		notePath, noteID, err := m.createNote()
		if err != nil {
			m.err = err
			m.state = stateError
			return m, nil
		}
		m.notePath = notePath
		m.noteID = noteID
		m.state = stateComplete
		// Show success message briefly before quitting
		return m, tea.Quit
	case "n":
		// Go back to edit
		m.state = stateSelectType
		m.verifying = false
	}
	return m, nil
}

// viewVerification renders the verification screen
func (m wizardModel) viewVerification() string {
	if m.verifying {
		return "Creating note...\n"
	}

	s := strings.Builder{}
	s.WriteString("Review your note:\n\n")
	s.WriteString(fmt.Sprintf("Type:  %s\n", m.typeName))
	s.WriteString(fmt.Sprintf("Key:   %s\n", m.key))
	s.WriteString(fmt.Sprintf("Title: %s\n", m.title))
	if len(m.tags) > 0 {
		s.WriteString(fmt.Sprintf("Tags:  %s\n", strings.Join(m.tags, ", ")))
	} else {
		s.WriteString("Tags:  (none)\n")
	}
	s.WriteString(fmt.Sprintf("State: %s\n", m.stateVal))
	s.WriteString("\n(Create note? [y/N], Esc to go back, q to quit)")
	return s.String()
}

// createNote creates the note file and returns the note path and ID
func (m wizardModel) createNote() (string, model.NoteID, error) {
	// Determine path
	notePath := filepath.Join(m.vaultRoot, string(m.typeName), string(m.key)+".Rmd")
	typeDir := filepath.Dir(notePath)

	// Create directory if missing
	if _, err := os.Stat(typeDir); os.IsNotExist(err) {
		if err := os.MkdirAll(typeDir, 0755); err != nil {
			return "", "", fmt.Errorf("creating type directory: %w", err)
		}
	}

	// Generate frontmatter and body
	now := time.Now().UTC()
	noteID := generateNoteID()
	frontmatter := generateFrontmatter(noteID, m.typeName, m.key, m.title, m.tags, m.stateVal, now)
	body := generateBody(m.title, m.cfg, m.typeName)

	// Write file atomically
	content := formatNote(frontmatter, body)
	if err := AtomicWrite(notePath, content); err != nil {
		return "", "", fmt.Errorf("writing note: %w", err)
	}

	return notePath, noteID, nil
}

// goBack returns to the previous state
func (m wizardModel) goBack() wizardModel {
	switch m.state {
	case stateInputKey:
		m.state = stateSelectType
	case stateInputTitle:
		m.state = stateInputKey
	case stateInputTags:
		m.state = stateInputTitle
	case stateInputState:
		m.state = stateInputTags
	case stateVerification:
		m.state = stateInputState
	}
	return m
}

// isInteractiveTerminal checks if we're running in an interactive terminal
func isInteractiveTerminal() bool {
	return isatty.IsTerminal(os.Stdout.Fd())
}
