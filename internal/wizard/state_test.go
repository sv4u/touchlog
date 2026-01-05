package wizard

import (
	"testing"
)

func TestStateString(t *testing.T) {
	tests := []struct {
		state State
		want  string
	}{
		{StateMainMenu, "MainMenu"},
		{StateTemplateSelection, "TemplateSelection"},
		{StateOutputDir, "OutputDir"},
		{StateTitle, "Title"},
		{StateTags, "Tags"},
		{StateMessage, "Message"},
		{StateFileCreated, "FileCreated"},
		{StateEditorLaunch, "EditorLaunch"},
		{StateReviewScreen, "ReviewScreen"},
		{State(999), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.state.String(); got != tt.want {
				t.Errorf("State.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCanTransitionTo(t *testing.T) {
	tests := []struct {
		name    string
		current State
		new     State
		want    bool
	}{
		// Valid transitions
		{"MainMenu to TemplateSelection", StateMainMenu, StateTemplateSelection, true},
		{"TemplateSelection to OutputDir", StateTemplateSelection, StateOutputDir, true},
		{"TemplateSelection to MainMenu (back)", StateTemplateSelection, StateMainMenu, true},
		{"OutputDir to Title", StateOutputDir, StateTitle, true},
		{"OutputDir to TemplateSelection (back)", StateOutputDir, StateTemplateSelection, true},
		{"Title to Tags", StateTitle, StateTags, true},
		{"Title to OutputDir (back)", StateTitle, StateOutputDir, true},
		{"Tags to Message", StateTags, StateMessage, true},
		{"Tags to Title (back)", StateTags, StateTitle, true},
		{"Message to FileCreated", StateMessage, StateFileCreated, true},
		{"Message to Tags (back)", StateMessage, StateTags, true},
		{"FileCreated to EditorLaunch", StateFileCreated, StateEditorLaunch, true},
		{"EditorLaunch to ReviewScreen", StateEditorLaunch, StateReviewScreen, true},
		{"ReviewScreen to EditorLaunch (re-edit)", StateReviewScreen, StateEditorLaunch, true},
		{"ReviewScreen to MainMenu (exit)", StateReviewScreen, StateMainMenu, true},

		// Invalid transitions
		{"MainMenu to OutputDir (skip)", StateMainMenu, StateOutputDir, false},
		{"OutputDir to Tags (skip)", StateOutputDir, StateTags, false},
		{"Title to Message (skip)", StateTitle, StateMessage, false},
		{"FileCreated to ReviewScreen (skip)", StateFileCreated, StateReviewScreen, false},
		{"EditorLaunch to FileCreated (back)", StateEditorLaunch, StateFileCreated, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CanTransitionTo(tt.current, tt.new); got != tt.want {
				t.Errorf("CanTransitionTo(%v, %v) = %v, want %v", tt.current, tt.new, got, tt.want)
			}
		})
	}
}

func TestCanGoBack(t *testing.T) {
	tests := []struct {
		name  string
		state State
		want  bool
	}{
		{"MainMenu", StateMainMenu, false},
		{"TemplateSelection", StateTemplateSelection, true},
		{"OutputDir", StateOutputDir, true},
		{"Title", StateTitle, true},
		{"Tags", StateTags, true},
		{"Message", StateMessage, true},
		{"FileCreated", StateFileCreated, false},
		{"EditorLaunch", StateEditorLaunch, false},
		{"ReviewScreen", StateReviewScreen, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CanGoBack(tt.state); got != tt.want {
				t.Errorf("CanGoBack(%v) = %v, want %v", tt.state, got, tt.want)
			}
		})
	}
}

func TestGetPreviousState(t *testing.T) {
	tests := []struct {
		name    string
		state   State
		want    State
		wantErr bool
	}{
		{"TemplateSelection", StateTemplateSelection, StateMainMenu, false},
		{"OutputDir", StateOutputDir, StateTemplateSelection, false},
		{"Title", StateTitle, StateOutputDir, false},
		{"Tags", StateTags, StateTitle, false},
		{"Message", StateMessage, StateTags, false},
		{"MainMenu (no back)", StateMainMenu, -1, true},
		{"FileCreated (no back)", StateFileCreated, -1, true},
		{"EditorLaunch (no back)", StateEditorLaunch, -1, true},
		{"ReviewScreen (no back)", StateReviewScreen, -1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetPreviousState(tt.state)
			if got != tt.want {
				t.Errorf("GetPreviousState(%v) = %v, want %v", tt.state, got, tt.want)
			}
			if (got == -1) != tt.wantErr {
				t.Errorf("GetPreviousState(%v) error = %v, wantErr %v", tt.state, got == -1, tt.wantErr)
			}
		})
	}
}
