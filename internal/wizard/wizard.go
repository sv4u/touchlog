package wizard

import (
	"fmt"
	"time"

	"github.com/sv4u/touchlog/internal/config"
	"github.com/sv4u/touchlog/internal/entry"
)

// Wizard represents the wizard state machine
type Wizard struct {
	// State management
	state     State
	prevState State
	history   []State

	// Entry data
	outputDir    string
	templateName string
	title        string
	tags         []string
	message      string

	// File management
	tempFilePath  string
	finalFilePath string
	fileContent   string

	// Configuration
	config *config.Config

	// Timestamp for entry creation
	timestamp time.Time

	// Metadata
	metadata   *entry.Metadata
	includeGit bool
}

// NewWizard creates a new wizard instance
func NewWizard(cfg *config.Config, includeGit bool) (*Wizard, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	return &Wizard{
		state:      StateMainMenu,
		prevState:  -1,
		history:    []State{StateMainMenu},
		config:     cfg,
		timestamp:  time.Now(),
		metadata:   nil, // Will be collected when needed
		includeGit: includeGit,
	}, nil
}

// GetState returns the current state
func (w *Wizard) GetState() State {
	return w.state
}

// GetOutputDir returns the output directory
func (w *Wizard) GetOutputDir() string {
	return w.outputDir
}

// GetTemplateName returns the selected template name
func (w *Wizard) GetTemplateName() string {
	return w.templateName
}

// GetTitle returns the title
func (w *Wizard) GetTitle() string {
	return w.title
}

// GetTags returns the tags
func (w *Wizard) GetTags() []string {
	return w.tags
}

// GetMessage returns the message
func (w *Wizard) GetMessage() string {
	return w.message
}

// GetTempFilePath returns the temporary file path
func (w *Wizard) GetTempFilePath() string {
	return w.tempFilePath
}

// GetFinalFilePath returns the final file path (after confirmation)
func (w *Wizard) GetFinalFilePath() string {
	return w.finalFilePath
}

// GetFileContent returns the file content
func (w *Wizard) GetFileContent() string {
	return w.fileContent
}

// GetTimestamp returns the timestamp for entry creation
func (w *Wizard) GetTimestamp() time.Time {
	return w.timestamp
}

// GetConfig returns the configuration
func (w *Wizard) GetConfig() *config.Config {
	return w.config
}

// SetOutputDir sets the output directory
// Resets metadata so it will be recollected with the new directory
func (w *Wizard) SetOutputDir(dir string) {
	w.outputDir = dir
	w.metadata = nil // Reset metadata so it will be recollected with new outputDir
}

// SetTemplateName sets the template name
func (w *Wizard) SetTemplateName(name string) {
	w.templateName = name
}

// SetTitle sets the title
func (w *Wizard) SetTitle(title string) {
	w.title = title
}

// SetTags sets the tags
func (w *Wizard) SetTags(tags []string) {
	w.tags = tags
}

// SetMessage sets the message
func (w *Wizard) SetMessage(message string) {
	w.message = message
}

// SetTempFilePath sets the temporary file path
func (w *Wizard) SetTempFilePath(path string) {
	w.tempFilePath = path
}

// SetFinalFilePath sets the final file path
func (w *Wizard) SetFinalFilePath(path string) {
	w.finalFilePath = path
}

// SetFileContent sets the file content
func (w *Wizard) SetFileContent(content string) {
	w.fileContent = content
}

// TransitionTo transitions to a new state
func (w *Wizard) TransitionTo(newState State) error {
	if !CanTransitionTo(w.state, newState) {
		return fmt.Errorf("invalid transition from %s to %s", w.state, newState)
	}

	// Update state history
	w.prevState = w.state
	w.history = append(w.history, newState)
	w.state = newState

	return nil
}

// CanGoBack checks if back navigation is allowed
func (w *Wizard) CanGoBack() bool {
	return CanGoBack(w.state)
}

// GoBack navigates back to the previous state
func (w *Wizard) GoBack() error {
	if !w.CanGoBack() {
		return fmt.Errorf("back navigation not allowed from state %s", w.state)
	}

	prevState := GetPreviousState(w.state)
	if prevState == -1 {
		return fmt.Errorf("no previous state for %s", w.state)
	}

	// Update state
	w.prevState = w.state
	w.history = append(w.history, prevState)
	w.state = prevState

	return nil
}

// ValidateState validates the current state
func (w *Wizard) ValidateState() error {
	// Each state can have its own validation rules
	switch w.state {
	case StateOutputDir:
		if w.outputDir == "" {
			return fmt.Errorf("output directory is required")
		}
		// Additional validation will be done in flow.go
	case StateTemplateSelection:
		// Template selection is optional (can use default)
		// No validation needed
	case StateTitle, StateTags, StateMessage:
		// These fields are optional (can be empty)
		// No validation needed
	default:
		// Other states don't need validation
	}

	return nil
}

// Reset clears all wizard data (useful for starting over)
func (w *Wizard) Reset() {
	w.state = StateMainMenu
	w.prevState = -1
	w.history = []State{StateMainMenu}
	w.outputDir = ""
	w.templateName = ""
	w.title = ""
	w.tags = nil
	w.message = ""
	w.tempFilePath = ""
	w.finalFilePath = ""
	w.fileContent = ""
	w.timestamp = time.Now()
	w.metadata = nil // Reset metadata so it will be recollected with new outputDir
}
