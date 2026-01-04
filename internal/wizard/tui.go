package wizard

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/sv4u/touchlog/internal/config"
)

// WizardModel represents the Bubble Tea model for the wizard
type WizardModel struct {
	wizard *Wizard

	// Bubble Tea components
	mainMenuList   list.Model
	templateList   list.Model
	reviewList     list.Model
	textInput      textinput.Model
	textArea       textarea.Model
	loadingSpinner spinner.Model

	// UI state
	width      int
	height     int
	err        error
	errorMsg   string
	isLoading  bool
	loadingMsg string

	// Input state
	currentInput string

	// Vim-style command buffer for review screen
	commandBuffer string
}

// NewWizardModel creates a new wizard model
func NewWizardModel(cfg *config.Config, includeGit bool) (*WizardModel, error) {
	w, err := NewWizard(cfg, includeGit)
	if err != nil {
		return nil, fmt.Errorf("failed to create wizard: %w", err)
	}

	// Initialize main menu list
	mainMenuItems := []list.Item{
		menuItem{title: "Create new entry", description: "Start creating a new log entry"},
	}
	mainMenuList := list.New(mainMenuItems, list.NewDefaultDelegate(), 0, 0)
	mainMenuList.Title = "touchlog"
	mainMenuList.SetShowStatusBar(false)
	mainMenuList.SetFilteringEnabled(false)

	// Initialize template list (will be populated later)
	templateList := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	templateList.Title = "Select Template"
	templateList.SetShowStatusBar(false)
	templateList.SetFilteringEnabled(false)

	// Initialize review list (will be populated later)
	reviewList := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	reviewList.Title = "Review Entry"
	reviewList.SetShowStatusBar(false)
	reviewList.SetFilteringEnabled(false)

	// Initialize text input
	ti := textinput.New()
	ti.CharLimit = 0
	ti.Width = 50

	// Initialize textarea
	ta := textarea.New()
	ta.Placeholder = "Enter your message (optional)..."
	ta.CharLimit = 0
	ta.SetWidth(80)
	ta.SetHeight(10)

	// Initialize spinner
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	return &WizardModel{
		wizard:         w,
		mainMenuList:   mainMenuList,
		templateList:   templateList,
		reviewList:     reviewList,
		textInput:      ti,
		textArea:       ta,
		loadingSpinner: s,
	}, nil
}

// Init initializes the model
func (m *WizardModel) Init() tea.Cmd {
	// Initialize template list with available templates
	return tea.Batch(
		m.loadingSpinner.Tick,
		m.initTemplateList(),
	)
}

// Update handles messages and updates the model
func (m *WizardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Update component sizes
		m.updateComponentSizes()
		return m, nil

	case tea.KeyMsg:
		handled, newModel, cmd := m.handleKeyPress(msg)
		if handled {
			return newModel, cmd
		}
		// Key was not handled by handleKeyPress, delegate to component
		// Continue to handleStateUpdate below

	case spinner.TickMsg:
		if m.isLoading {
			var cmd tea.Cmd
			m.loadingSpinner, cmd = m.loadingSpinner.Update(msg)
			return m, cmd
		}
		return m, nil

	case fileCreatedMsg:
		m.isLoading = false
		m.wizard.SetTempFilePath(msg.tempPath)
		m.wizard.SetFileContent(msg.content)
		if err := m.wizard.TransitionTo(StateEditorLaunch); err != nil {
			m.errorMsg = err.Error()
			return m, nil
		}
		// Automatically launch editor
		return m, m.launchEditorCmd()

	case editorLaunchedMsg:
		m.isLoading = false
		m.commandBuffer = "" // Clear command buffer
		m.errorMsg = ""      // Clear error message
		if err := m.wizard.TransitionTo(StateReviewScreen); err != nil {
			m.errorMsg = err.Error()
			return m, nil
		}
		// Update review screen with file content
		return m, m.updateReviewScreen()

	case templatesLoadedMsg:
		// Update template list with loaded templates
		m.templateList.SetItems(msg.items)
		return m, nil

	case reviewScreenUpdatedMsg:
		// Update review list with actions
		m.reviewList.SetItems(msg.items)
		return m, nil

	case errorMsg:
		m.isLoading = false
		m.errorMsg = msg.err.Error()
		return m, nil
	}

	// Delegate to state-specific update handlers
	return m.handleStateUpdate(msg)
}

// View renders the UI
func (m *WizardModel) View() string {
	if m.isLoading {
		return m.renderLoading()
	}

	// Don't return early for errors - let state-specific render methods handle inline errors
	// Only show full-screen error for states that don't support inline errors
	switch m.wizard.GetState() {
	case StateMainMenu, StateTemplateSelection, StateReviewScreen, StateFileCreated, StateEditorLaunch:
		// These states don't support inline errors, show full-screen error if any
		if m.errorMsg != "" {
			return m.renderError()
		}
		// Fall through to render the state
		switch m.wizard.GetState() {
		case StateMainMenu:
			return m.renderMainMenu()
		case StateTemplateSelection:
			return m.renderTemplateSelection()
		case StateReviewScreen:
			return m.renderReviewScreen()
		case StateFileCreated, StateEditorLaunch:
			return m.renderLoading() // Show spinner while creating file/launching editor
		}
	case StateOutputDir:
		return m.renderOutputDir()
	case StateTitle:
		return m.renderTitle()
	case StateTags:
		return m.renderTags()
	case StateMessage:
		return m.renderMessage()
	default:
		return "Unknown state"
	}
	return "Unknown state"
}

// menuItem implements the list.Item interface
type menuItem struct {
	title       string
	description string
}

func (i menuItem) FilterValue() string { return i.title }
func (i menuItem) Title() string       { return i.title }
func (i menuItem) Description() string { return i.description }

// templateListItem implements the list.Item interface for templates
type templateListItem struct {
	name        string
	description string
	isDefault   bool
}

func (i templateListItem) FilterValue() string { return i.name }
func (i templateListItem) Title() string       { return i.name }
func (i templateListItem) Description() string { return i.description }

// reviewActionItem implements the list.Item interface for review actions
type reviewActionItem struct {
	title       string
	description string
	action      string
}

func (i reviewActionItem) FilterValue() string { return i.title }
func (i reviewActionItem) Title() string       { return i.title }
func (i reviewActionItem) Description() string { return i.description }

// Custom message types
type fileCreatedMsg struct {
	tempPath string
	content  string
}

type editorLaunchedMsg struct{}

type errorMsg struct {
	err error
}

// initTemplateList initializes the template list with available templates
func (m *WizardModel) initTemplateList() tea.Cmd {
	return func() tea.Msg {
		// Get available templates
		templates := GetAvailableTemplates(m.wizard.GetConfig())

		// Create list items
		items := make([]list.Item, 0, len(templates)+1)

		// Add "Use default template" option first
		defaultTemplate := m.wizard.GetConfig().GetDefaultTemplate()
		if defaultTemplate == "" {
			defaultTemplate = "daily"
		}
		items = append(items, templateListItem{
			name:        "Use default template",
			description: fmt.Sprintf("(default: %s)", defaultTemplate),
			isDefault:   true,
		})

		// Add available templates
		for _, name := range templates {
			items = append(items, templateListItem{
				name:        name,
				description: "Select this template",
				isDefault:   false,
			})
		}

		// Update template list (this will be handled in Update)
		return templatesLoadedMsg{items: items}
	}
}

type templatesLoadedMsg struct {
	items []list.Item
}

// handleKeyPress handles keyboard input
// Returns: (handled bool, model, cmd)
// If handled is false, the caller should delegate to component update
func (m *WizardModel) handleKeyPress(msg tea.KeyMsg) (bool, tea.Model, tea.Cmd) {
	key := msg.String()

	// Global keybindings
	switch key {
	case "ctrl+c":
		// Cancel and cleanup
		_ = m.wizard.Cancel()
		return true, m, tea.Quit
	}

	// State-specific keybindings
	switch m.wizard.GetState() {
	case StateMainMenu:
		handled, model, cmd := m.handleMainMenuKey(key)
		return handled, model, cmd
	case StateTemplateSelection:
		handled, model, cmd := m.handleTemplateSelectionKey(key)
		return handled, model, cmd
	case StateOutputDir, StateTitle, StateTags:
		handled, model, cmd := m.handleTextInputKey(key, msg)
		return handled, model, cmd
	case StateMessage:
		handled, model, cmd := m.handleTextAreaKey(key, msg)
		return handled, model, cmd
	case StateReviewScreen:
		handled, model, cmd := m.handleReviewScreenKey(key)
		return handled, model, cmd
	default:
		// Unknown state - don't handle the key, let it fall through
		return false, m, nil
	}
}

// handleStateUpdate handles state-specific updates
func (m *WizardModel) handleStateUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.wizard.GetState() {
	case StateMainMenu:
		var cmd tea.Cmd
		m.mainMenuList, cmd = m.mainMenuList.Update(msg)
		return m, cmd
	case StateTemplateSelection:
		var cmd tea.Cmd
		m.templateList, cmd = m.templateList.Update(msg)
		return m, cmd
	case StateOutputDir, StateTitle, StateTags:
		var cmd tea.Cmd
		m.textInput, cmd = m.textInput.Update(msg)
		return m, cmd
	case StateMessage:
		var cmd tea.Cmd
		m.textArea, cmd = m.textArea.Update(msg)
		return m, cmd
	case StateReviewScreen:
		// Don't delegate to list if in command mode
		if m.commandBuffer != "" {
			return m, nil
		}
		var cmd tea.Cmd
		m.reviewList, cmd = m.reviewList.Update(msg)
		return m, cmd
	default:
		return m, nil
	}
}

// handleMainMenuKey handles key presses in main menu state
// Returns: (handled bool, model, cmd)
// If handled is false, the key should be delegated to the list component
func (m *WizardModel) handleMainMenuKey(key string) (bool, tea.Model, tea.Cmd) {
	switch key {
	case "enter":
		selected := m.mainMenuList.SelectedItem()
		if selected != nil {
			// Transition to template selection
			if err := m.wizard.TransitionTo(StateTemplateSelection); err != nil {
				m.errorMsg = err.Error()
				return true, m, nil
			}
			return true, m, nil
		}
	case "esc", "q":
		_ = m.wizard.Cancel()
		return true, m, tea.Quit
	default:
		// Delegate all other keys to the list component (arrow keys, etc.)
		return false, m, nil
	}
	return true, m, nil
}

// handleTemplateSelectionKey handles key presses in template selection state
// Returns: (handled bool, model, cmd)
// If handled is false, the key should be delegated to the list component
func (m *WizardModel) handleTemplateSelectionKey(key string) (bool, tea.Model, tea.Cmd) {
	switch key {
	case "enter":
		selected := m.templateList.SelectedItem()
		if selected != nil {
			item := selected.(templateListItem)
			if item.isDefault {
				// Use default template (already set in wizard)
				m.wizard.SetTemplateName("")
			} else {
				m.wizard.SetTemplateName(item.name)
			}
			// Transition to output dir
			if err := m.wizard.TransitionTo(StateOutputDir); err != nil {
				m.errorMsg = err.Error()
				return true, m, nil
			}
			// Initialize text input for output dir
			m.textInput.Reset()
			m.textInput.Placeholder = "Enter output directory (e.g., ~/Documents/notes)"
			m.textInput.Focus()
			// Pre-fill with config value if available
			if notesDir := m.wizard.GetConfig().GetNotesDirectory(); notesDir != "" {
				m.textInput.SetValue(notesDir)
			}
			return true, m, nil
		}
	case "esc", "b":
		// Go back to main menu
		if err := m.wizard.GoBack(); err != nil {
			m.errorMsg = err.Error()
			return true, m, nil
		}
		return true, m, nil
	case "q":
		// Quit
		_ = m.wizard.Cancel()
		return true, m, tea.Quit
	default:
		// Delegate all other keys to the list component (arrow keys, etc.)
		return false, m, nil
	}
	return true, m, nil
}

// handleTextInputKey handles key presses in text input states (OutputDir, Title, Tags)
// Returns: (handled bool, model, cmd)
// If handled is false, the key should be delegated to the text input component
func (m *WizardModel) handleTextInputKey(key string, msg tea.Msg) (bool, tea.Model, tea.Cmd) {
	switch key {
	case "enter":
		// Validate and proceed to next state
		input := m.textInput.Value()
		currentState := m.wizard.GetState()

		switch currentState {
		case StateOutputDir:
			// Validate output directory
			if err := m.wizard.ValidateOutputDir(input); err != nil {
				m.errorMsg = err.Error()
				return true, m, nil
			}
			m.wizard.SetOutputDir(input)
			m.errorMsg = "" // Clear error
			// Transition to title
			if err := m.wizard.TransitionTo(StateTitle); err != nil {
				m.errorMsg = err.Error()
				return true, m, nil
			}
			m.textInput.Reset()
			m.textInput.Placeholder = "Enter title (optional, press Enter to skip)"
			m.textInput.Focus()
			return true, m, nil

		case StateTitle:
			m.wizard.SetTitle(input)
			// Transition to tags
			if err := m.wizard.TransitionTo(StateTags); err != nil {
				m.errorMsg = err.Error()
				return true, m, nil
			}
			m.textInput.Reset()
			m.textInput.Placeholder = "Enter tags (comma-separated, optional, press Enter to skip)"
			m.textInput.Focus()
			return true, m, nil

		case StateTags:
			tags := ParseTags(input)
			m.wizard.SetTags(tags)
			// Transition to message
			if err := m.wizard.TransitionTo(StateMessage); err != nil {
				m.errorMsg = err.Error()
				return true, m, nil
			}
			m.textArea.Reset()
			m.textArea.Placeholder = "Enter message (optional, press Ctrl+D to finish)"
			m.textArea.Focus()
			return true, m, nil
		}

	case "esc":
		// Go back (only Esc, not 'b' - 'b' should be a normal character)
		if m.wizard.CanGoBack() {
			if err := m.wizard.GoBack(); err != nil {
				m.errorMsg = err.Error()
				return true, m, nil
			}
			// Re-initialize input based on previous state
			m.initializeInputForState(m.wizard.GetState())
			m.errorMsg = "" // Clear error
			return true, m, nil
		}
	default:
		// Delegate all other keys to the text input component
		// Return false to indicate we didn't handle it
		return false, m, nil
	}
	return true, m, nil
}

// handleTextAreaKey handles key presses in message state
// Returns: (handled bool, model, cmd)
// If handled is false, the key should be delegated to the textarea component
func (m *WizardModel) handleTextAreaKey(key string, msg tea.Msg) (bool, tea.Model, tea.Cmd) {
	switch key {
	case "ctrl+d":
		// Finish message input and create file
		message := m.textArea.Value()
		m.wizard.SetMessage(message)

		// Transition to file creation
		if err := m.wizard.TransitionTo(StateFileCreated); err != nil {
			m.errorMsg = err.Error()
			return true, m, nil
		}

		// Create temp file
		m.isLoading = true
		m.loadingMsg = "Creating file..."
		return true, m, tea.Batch(
			m.loadingSpinner.Tick,
			m.createTempFileCmd(),
		)

	case "esc":
		// Go back (only Esc, not 'b' - 'b' should be a normal character)
		if m.wizard.CanGoBack() {
			if err := m.wizard.GoBack(); err != nil {
				m.errorMsg = err.Error()
				return true, m, nil
			}
			// Re-initialize input for tags
			m.textInput.Reset()
			m.textInput.Placeholder = "Enter tags (comma-separated, optional, press Enter to skip)"
			m.textInput.Focus()
			// Restore tags value
			tags := m.wizard.GetTags()
			if len(tags) > 0 {
				m.textInput.SetValue(strings.Join(tags, ", "))
			}
			m.errorMsg = "" // Clear error
			return true, m, nil
		}
	default:
		// Delegate all other keys to the textarea component
		// Return false to indicate we didn't handle it
		return false, m, nil
	}
	return true, m, nil
}

// handleReviewScreenKey handles key presses in review screen state
// Returns: (handled bool, model, cmd)
// If handled is false, the key should be delegated to the list component
func (m *WizardModel) handleReviewScreenKey(key string) (bool, tea.Model, tea.Cmd) {
	// Handle command mode (when typing :)
	if m.commandBuffer != "" {
		model, cmd := m.handleCommandMode(key)
		return true, model, cmd
	}

	switch key {
	case "enter":
		selected := m.reviewList.SelectedItem()
		if selected != nil {
			item := selected.(reviewActionItem)
			model, cmd := m.handleReviewAction(item.action)
			return true, model, cmd
		}
	case "esc":
		// Clear command buffer if in command mode, otherwise quit
		if m.commandBuffer != "" {
			m.commandBuffer = ""
			return true, m, nil
		}
		// Quit without saving
		model, cmd := m.handleReviewAction("cancel")
		return true, model, cmd
	case "q":
		// Only quit if not in command mode
		model, cmd := m.handleReviewAction("cancel")
		return true, model, cmd
	case ":":
		// Enter command mode
		m.commandBuffer = ":"
		return true, m, nil
	default:
		// Delegate all other keys to the list component (arrow keys, etc.)
		// But only if not in command mode (command mode is handled above)
		return false, m, nil
	}
	return true, m, nil
}

// handleCommandMode handles vim-style command input
func (m *WizardModel) handleCommandMode(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "enter":
		// Execute command
		cmd := strings.TrimSpace(m.commandBuffer)
		m.commandBuffer = ""

		switch cmd {
		case ":wq", ":w", ":write", ":writequit":
			// Save and exit
			return m.handleReviewAction("save")
		case ":q":
			// Quit (same as cancel - delete temp file)
			return m.handleReviewAction("cancel")
		case ":q!", ":quit!":
			// Force quit without saving (delete temp file)
			return m.handleReviewAction("cancel")
		case ":e", ":edit":
			// Re-open editor
			return m.handleReviewAction("edit")
		default:
			// Unknown command - clear buffer and show error
			if strings.HasPrefix(cmd, ":") {
				m.errorMsg = fmt.Sprintf("Unknown command: %s (try :wq, :q, :q!, or :e)", cmd)
			} else {
				m.errorMsg = fmt.Sprintf("Unknown command: %s", cmd)
			}
			m.commandBuffer = ""
			return m, nil
		}
	case "esc":
		// Cancel command mode
		m.commandBuffer = ""
		return m, nil
	case "backspace":
		// Remove last character from command buffer
		if len(m.commandBuffer) > 0 {
			m.commandBuffer = m.commandBuffer[:len(m.commandBuffer)-1]
		}
		return m, nil
	default:
		// Add character to command buffer (only printable ASCII)
		if len(key) == 1 && key[0] >= 32 && key[0] <= 126 {
			m.commandBuffer += key
		}
		return m, nil
	}
}

// handleReviewAction handles review screen actions
func (m *WizardModel) handleReviewAction(action string) (tea.Model, tea.Cmd) {
	switch action {
	case "save":
		// Confirm and save
		if err := m.wizard.Confirm(); err != nil {
			m.errorMsg = err.Error()
			return m, nil
		}
		// Exit wizard
		return m, tea.Quit

	case "edit":
		// Re-launch editor
		m.isLoading = true
		m.loadingMsg = "Opening editor..."
		return m, tea.Batch(
			m.loadingSpinner.Tick,
			m.launchEditorCmd(),
		)

	case "cancel":
		// Cancel and delete temp file
		if err := m.wizard.Cancel(); err != nil {
			m.errorMsg = err.Error()
			return m, nil
		}
		return m, tea.Quit
	}
	return m, nil
}

// initializeInputForState initializes the text input for a given state
func (m *WizardModel) initializeInputForState(state State) {
	switch state {
	case StateOutputDir:
		m.textInput.Reset()
		m.textInput.Placeholder = "Enter output directory (e.g., ~/Documents/notes)"
		m.textInput.Focus()
		// Prioritize user-entered value over config default
		if m.wizard.GetOutputDir() != "" {
			m.textInput.SetValue(m.wizard.GetOutputDir())
		} else if notesDir := m.wizard.GetConfig().GetNotesDirectory(); notesDir != "" {
			m.textInput.SetValue(notesDir)
		}
	case StateTitle:
		m.textInput.Reset()
		m.textInput.Placeholder = "Enter title (optional, press Enter to skip)"
		m.textInput.Focus()
		if title := m.wizard.GetTitle(); title != "" {
			m.textInput.SetValue(title)
		}
	case StateTags:
		m.textInput.Reset()
		m.textInput.Placeholder = "Enter tags (comma-separated, optional, press Enter to skip)"
		m.textInput.Focus()
		if tags := m.wizard.GetTags(); len(tags) > 0 {
			m.textInput.SetValue(strings.Join(tags, ", "))
		}
	}
}

// createTempFileCmd creates a command to create the temp file
func (m *WizardModel) createTempFileCmd() tea.Cmd {
	return func() tea.Msg {
		if err := m.wizard.CreateTempFile(); err != nil {
			return errorMsg{err: err}
		}
		return fileCreatedMsg{
			tempPath: m.wizard.GetTempFilePath(),
			content:  m.wizard.GetFileContent(),
		}
	}
}

// launchEditorCmd creates a command to launch the editor
func (m *WizardModel) launchEditorCmd() tea.Cmd {
	return func() tea.Msg {
		if err := m.wizard.LaunchEditor(); err != nil {
			return errorMsg{err: err}
		}
		return editorLaunchedMsg{}
	}
}

// updateReviewScreen updates the review screen with current file content
func (m *WizardModel) updateReviewScreen() tea.Cmd {
	return func() tea.Msg {
		// Read file content
		content := m.wizard.GetFileContent()

		// Create review actions
		items := []list.Item{
			reviewActionItem{
				title:       "Save and exit",
				description: "Save the file and exit (:wq)",
				action:      "save",
			},
			reviewActionItem{
				title:       "Open editor again",
				description: "Re-open the file in editor",
				action:      "edit",
			},
			reviewActionItem{
				title:       "Cancel and delete file",
				description: "Discard changes and exit (:q!)",
				action:      "cancel",
			},
		}

		// Update review list
		// This will be handled in Update method
		return reviewScreenUpdatedMsg{content: content, items: items}
	}
}

type reviewScreenUpdatedMsg struct {
	content string
	items   []list.Item
}

// updateComponentSizes updates component sizes based on window size
func (m *WizardModel) updateComponentSizes() {
	width := m.width - 4
	height := m.height - 4

	m.mainMenuList.SetWidth(width)
	m.mainMenuList.SetHeight(height - 2)

	m.templateList.SetWidth(width)
	m.templateList.SetHeight(height - 2)

	m.reviewList.SetWidth(width)
	m.reviewList.SetHeight(height - 2)

	m.textInput.Width = width - 2
	m.textArea.SetWidth(width - 2)
	m.textArea.SetHeight(height - 6)
}

// Render methods
func (m *WizardModel) renderMainMenu() string {
	return m.mainMenuList.View() + "\n\n" + m.renderHelp("↑↓: navigate | Enter: select | Esc/q: quit")
}

func (m *WizardModel) renderTemplateSelection() string {
	view := m.templateList.View()
	help := "↑↓: navigate | Enter: select | Esc/b: back | q: quit"
	return view + "\n\n" + m.renderHelp(help)
}

func (m *WizardModel) renderOutputDir() string {
	view := m.textInput.View()
	help := "Enter: continue | Esc/b: back | Ctrl+C: quit"
	if m.errorMsg != "" {
		view += "\n\n" + m.renderErrorInline(m.errorMsg)
	}
	return view + "\n\n" + m.renderHelp(help)
}

func (m *WizardModel) renderTitle() string {
	view := m.textInput.View()
	help := "Enter: continue (or skip) | Esc/b: back | Ctrl+C: quit"
	return view + "\n\n" + m.renderHelp(help)
}

func (m *WizardModel) renderTags() string {
	view := m.textInput.View()
	help := "Enter: continue (or skip) | Esc/b: back | Ctrl+C: quit"
	return view + "\n\n" + m.renderHelp(help)
}

func (m *WizardModel) renderMessage() string {
	view := m.textArea.View()
	help := "Ctrl+D: finish | Esc/b: back | Ctrl+C: quit"
	return view + "\n\n" + m.renderHelp(help)
}

func (m *WizardModel) renderReviewScreen() string {
	// Show file preview and action list
	content := m.wizard.GetFileContent()

	// Limit preview to first 20 lines
	lines := strings.Split(content, "\n")
	previewLines := lines
	if len(lines) > 20 {
		previewLines = lines[:20]
	}
	preview := strings.Join(previewLines, "\n")
	if len(lines) > 20 {
		preview += "\n... (" + fmt.Sprintf("%d", len(lines)-20) + " more lines)"
	}

	previewStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(1, 2).
		Width(m.width - 4).
		Height(m.height/2 - 2)

	reviewStyle := lipgloss.NewStyle().
		MarginTop(1)

	// Build help text
	helpText := "↑↓: navigate | Enter: select | :wq: save | :q!: cancel | Esc/q: quit"

	// Show command buffer if in command mode
	var commandDisplay string
	if m.commandBuffer != "" {
		commandStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).
			Bold(true)
		commandDisplay = "\n" + commandStyle.Render(m.commandBuffer)
		helpText = "Type command and press Enter | Esc: cancel"
	}

	// Show error if any
	var errorDisplay string
	if m.errorMsg != "" {
		errorDisplay = "\n" + m.renderErrorInline(m.errorMsg)
	}

	return previewStyle.Render(preview) + "\n" +
		reviewStyle.Render(m.reviewList.View()) +
		commandDisplay +
		errorDisplay + "\n" +
		m.renderHelp(helpText)
}

func (m *WizardModel) renderLoading() string {
	return fmt.Sprintf("%s %s\n\n%s",
		m.loadingSpinner.View(),
		m.loadingMsg,
		m.renderHelp("Please wait..."),
	)
}

func (m *WizardModel) renderError() string {
	errorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("1")).
		Bold(true).
		Padding(1, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("1"))

	return errorStyle.Render("Error: "+m.errorMsg) + "\n\n" +
		m.renderHelp("Press Ctrl+C to quit")
}

func (m *WizardModel) renderErrorInline(msg string) string {
	errorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("1")).
		Bold(true)
	return errorStyle.Render("Error: " + msg)
}

func (m *WizardModel) renderHelp(text string) string {
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Italic(true)
	return helpStyle.Render(text)
}
