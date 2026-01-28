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
	stateInputFilename
	stateVerification
	stateComplete
	stateError
)

const escDoublePressWindow = 500 * time.Millisecond

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
	typeCursor    int
	typeChoices   []model.TypeName
	keyInput      string
	keyError      string
	titleInput    string
	tagsInput     string
	stateInput    string
	filenameInput string
	filenameError string
	verifying     bool
	notePath      string
	noteID        model.NoteID

	// Quit sequence state
	quitSequence   string
	lastEscTime    time.Time
	inQuitSequence bool
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
		vaultRoot:      vaultRoot,
		cfg:            cfg,
		state:          stateSelectType,
		typeChoices:    typeChoices,
		keyInput:       "",
		titleInput:     "",
		tagsInput:      "",
		stateInput:     "",
		filenameInput:  "",
		quitSequence:   "",
		lastEscTime:    time.Time{},
		inQuitSequence: false,
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
		// 1. Handle quit sequence FIRST (before any other key handling)
		if m.inQuitSequence {
			keyStr := msg.String()
			switch m.quitSequence {
			case "waiting_for_colon":
				if keyStr == ":" {
					m.quitSequence = "waiting_for_q"
					return m, nil
				}
				// Non-matching key: reset and process normally
				m.inQuitSequence = false
				m.quitSequence = ""
				// Fall through to normal processing
			case "waiting_for_q":
				if keyStr == "q" {
					return m, tea.Quit
				}
				// Non-matching key: reset and process normally
				m.inQuitSequence = false
				m.quitSequence = ""
				// Fall through to normal processing
			}
		}

		// 2. Handle Esc key (with double-Esc detection)
		switch msg.String() {
		case "esc":
			now := time.Now()
			if m.inQuitSequence {
				// Esc cancels quit sequence
				m.inQuitSequence = false
				m.quitSequence = ""
				m.lastEscTime = time.Time{} // Reset
				// Continue with normal Esc behavior below
			} else if !m.lastEscTime.IsZero() && now.Sub(m.lastEscTime) < escDoublePressWindow {
				// Double Esc detected: enter quit sequence
				m.inQuitSequence = true
				m.quitSequence = "waiting_for_colon"
				m.lastEscTime = time.Time{} // Reset for next time
				return m, nil
			}
			// Single Esc: preserve existing behavior
			m.lastEscTime = now
			if m.state != stateSelectType {
				// Go back to previous state
				return m.goBack(), nil
			}
			return m, tea.Quit
		case "ctrl+c":
			return m, tea.Quit
		}

		// 3. Delegate to state-specific handlers
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
		case stateInputFilename:
			return m.updateFilenameInput(msg)
		case stateVerification:
			return m.updateVerification(msg)
		}
	}

	return m, nil
}

// View renders the UI
func (m wizardModel) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\n\nPress Esc Esc then :q to quit.", m.err)
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
	case stateInputFilename:
		return m.viewFilenameInput()
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

	s.WriteString("\n(Use arrow keys to navigate, Enter to select, Double Esc + :q to quit)")
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

		// Validate key using the new ValidateKey function that supports path-based keys
		if err := config.ValidateKey(keyStr, m.typeDef.KeyPattern, m.typeDef.KeyMaxLen); err != nil {
			m.keyError = err.Error()
			return m, nil
		}

		// Check uniqueness using store if it exists (for indexed notes)
		if err := checkKeyUniqueness(m.vaultRoot, m.typeName, keyStr); err != nil {
			m.keyError = err.Error()
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
	s.WriteString("\n\n(Enter to continue, Esc to go back, Double Esc + :q to quit)")
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
		// Set default filename based on title
		m.filenameInput = sanitizeTitleForFilename(title)
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
	s.WriteString("\n\n(Enter to continue, Esc to go back, Double Esc + :q to quit)")
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
	s.WriteString("\n\n(Enter to continue, Esc to go back, Double Esc + :q to quit)")
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
		m.state = stateInputFilename
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
	s.WriteString("\n\n(Enter to continue with default or custom state, Esc to go back, Double Esc + :q to quit)")
	return s.String()
}

// updateFilenameInput handles filename input
func (m wizardModel) updateFilenameInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		filenameStr := strings.TrimSpace(m.filenameInput)
		if filenameStr == "" {
			m.filenameError = "Filename cannot be empty"
			return m, nil
		}

		// Remove .Rmd extension if provided (we'll add it automatically)
		filenameStr = strings.TrimSuffix(filenameStr, ".Rmd")
		filenameStr = strings.TrimSuffix(filenameStr, ".rmd")

		// Validate filename (basic validation - no path separators, no empty)
		if filenameStr == "" {
			m.filenameError = "Filename cannot be empty"
			return m, nil
		}
		if strings.Contains(filenameStr, string(filepath.Separator)) || strings.Contains(filenameStr, "/") || strings.Contains(filenameStr, "\\") {
			m.filenameError = "Filename cannot contain path separators"
			return m, nil
		}

		// Check if file already exists using branching path logic
		var notePath string
		keyStr := string(m.key)
		if strings.Contains(keyStr, "/") {
			// Path-based key: file in subfolder
			notePath = filepath.Join(m.vaultRoot, string(m.typeName), keyStr, filenameStr+".Rmd")
		} else {
			// Flat key: file at type root (backward compatible)
			notePath = filepath.Join(m.vaultRoot, string(m.typeName), filenameStr+".Rmd")
		}
		if _, err := os.Stat(notePath); err == nil {
			m.filenameError = fmt.Sprintf("File %q already exists", filenameStr+".Rmd")
			return m, nil
		}

		// Valid filename
		m.filenameInput = filenameStr
		m.filenameError = ""
		m.state = stateVerification
	case "backspace":
		if len(m.filenameInput) > 0 {
			m.filenameInput = m.filenameInput[:len(m.filenameInput)-1]
			m.filenameError = ""
		}
	default:
		if len(msg.Runes) > 0 {
			m.filenameInput += string(msg.Runes)
			m.filenameError = ""
		}
	}
	return m, nil
}

// viewFilenameInput renders the filename input screen
func (m wizardModel) viewFilenameInput() string {
	s := strings.Builder{}
	s.WriteString("Enter output filename (without .Rmd extension):\n\n")
	s.WriteString(fmt.Sprintf("> %s", m.filenameInput))
	if m.filenameError != "" {
		s.WriteString(fmt.Sprintf("\n\nError: %s", m.filenameError))
	}
	s.WriteString("\n\n(Enter to continue, Esc to go back, Double Esc + :q to quit)")
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
	s.WriteString(fmt.Sprintf("Type:     %s\n", m.typeName))
	s.WriteString(fmt.Sprintf("Key:      %s\n", m.key))
	s.WriteString(fmt.Sprintf("Title:    %s\n", m.title))
	if len(m.tags) > 0 {
		s.WriteString(fmt.Sprintf("Tags:     %s\n", strings.Join(m.tags, ", ")))
	} else {
		s.WriteString("Tags:     (none)\n")
	}
	s.WriteString(fmt.Sprintf("State:    %s\n", m.stateVal))
	s.WriteString(fmt.Sprintf("Filename: %s.Rmd\n", m.filenameInput))
	s.WriteString("\n(Create note? [y/N], Esc to go back, Double Esc + :q to quit)")
	return s.String()
}

// createNote creates the note file and returns the note path and ID
func (m wizardModel) createNote() (string, model.NoteID, error) {
	// Determine path using branching logic for backward compatibility
	var notePath string
	keyStr := string(m.key)
	if strings.Contains(keyStr, "/") {
		// Path-based key: file in subfolder
		notePath = filepath.Join(m.vaultRoot, string(m.typeName), keyStr, m.filenameInput+".Rmd")
	} else {
		// Flat key: file at type root (backward compatible)
		notePath = filepath.Join(m.vaultRoot, string(m.typeName), m.filenameInput+".Rmd")
	}
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
	case stateInputFilename:
		m.state = stateInputState
	case stateVerification:
		m.state = stateInputFilename
	}
	return m
}

// isInteractiveTerminal checks if we're running in an interactive terminal
func isInteractiveTerminal() bool {
	return isatty.IsTerminal(os.Stdout.Fd())
}
