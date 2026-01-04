package wizard

// State represents the current state of the wizard
type State int

const (
	// StateMainMenu is the initial state showing the main menu
	StateMainMenu State = iota
	// StateTemplateSelection is the template selection state
	StateTemplateSelection
	// StateOutputDir is the output directory input state
	StateOutputDir
	// StateTitle is the title input state
	StateTitle
	// StateTags is the tags input state
	StateTags
	// StateMessage is the message input state
	StateMessage
	// StateFileCreated is the state after file is created (temp file)
	StateFileCreated
	// StateEditorLaunch is the state when editor is being launched
	StateEditorLaunch
	// StateReviewScreen is the review screen state
	StateReviewScreen
)

// String returns a string representation of the state
func (s State) String() string {
	switch s {
	case StateMainMenu:
		return "MainMenu"
	case StateTemplateSelection:
		return "TemplateSelection"
	case StateOutputDir:
		return "OutputDir"
	case StateTitle:
		return "Title"
	case StateTags:
		return "Tags"
	case StateMessage:
		return "Message"
	case StateFileCreated:
		return "FileCreated"
	case StateEditorLaunch:
		return "EditorLaunch"
	case StateReviewScreen:
		return "ReviewScreen"
	default:
		return "Unknown"
	}
}

// CanTransitionTo checks if a transition from current state to new state is valid
func CanTransitionTo(current State, new State) bool {
	switch current {
	case StateMainMenu:
		return new == StateTemplateSelection
	case StateTemplateSelection:
		return new == StateOutputDir || new == StateMainMenu // Can go back to main menu
	case StateOutputDir:
		return new == StateTitle || new == StateTemplateSelection // Can go back
	case StateTitle:
		return new == StateTags || new == StateOutputDir // Can go back
	case StateTags:
		return new == StateMessage || new == StateTitle // Can go back
	case StateMessage:
		return new == StateFileCreated || new == StateTags // Can go back
	case StateFileCreated:
		return new == StateEditorLaunch
	case StateEditorLaunch:
		return new == StateReviewScreen
	case StateReviewScreen:
		return new == StateEditorLaunch || new == StateMainMenu // Can re-launch editor or exit
	default:
		return false
	}
}

// CanGoBack checks if back navigation is allowed from the current state
func CanGoBack(s State) bool {
	switch s {
	case StateTemplateSelection, StateOutputDir, StateTitle, StateTags, StateMessage:
		return true
	case StateMainMenu, StateFileCreated, StateEditorLaunch, StateReviewScreen:
		return false
	default:
		return false
	}
}

// GetPreviousState returns the expected previous state for back navigation
// Returns -1 if back navigation is not allowed
func GetPreviousState(s State) State {
	switch s {
	case StateTemplateSelection:
		return StateMainMenu
	case StateOutputDir:
		return StateTemplateSelection
	case StateTitle:
		return StateOutputDir
	case StateTags:
		return StateTitle
	case StateMessage:
		return StateTags
	default:
		return -1 // Invalid state for back navigation
	}
}

