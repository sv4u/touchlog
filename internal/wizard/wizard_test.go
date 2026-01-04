package wizard

import (
	"testing"

	"github.com/sv4u/touchlog/internal/config"
)

func TestNewWizard(t *testing.T) {
	cfg := config.CreateDefaultConfig()
	
	w, err := NewWizard(cfg, false)
	if err != nil {
		t.Fatalf("NewWizard() error = %v, want nil", err)
	}
	
	if w == nil {
		t.Fatal("NewWizard() returned nil wizard")
	}
	
	if w.GetState() != StateMainMenu {
		t.Errorf("NewWizard() initial state = %v, want %v", w.GetState(), StateMainMenu)
	}
	
	if w.GetConfig() != cfg {
		t.Errorf("NewWizard() config = %v, want %v", w.GetConfig(), cfg)
	}
}

func TestNewWizard_NilConfig(t *testing.T) {
	_, err := NewWizard(nil, false)
	if err == nil {
		t.Error("NewWizard(nil) error = nil, want error")
	}
}

func TestWizard_TransitionTo(t *testing.T) {
	cfg := config.CreateDefaultConfig()
	w, _ := NewWizard(cfg, false)
	
	tests := []struct {
		name    string
		setup   func(*Wizard) error
		newState State
		wantErr bool
	}{
		{
			"Valid: MainMenu to TemplateSelection",
			func(w *Wizard) error { return nil }, // Already in MainMenu
			StateTemplateSelection,
			false,
		},
		{
			"Valid: TemplateSelection to OutputDir",
			func(w *Wizard) error { return w.TransitionTo(StateTemplateSelection) },
			StateOutputDir,
			false,
		},
		{
			"Invalid: MainMenu to OutputDir (skip)",
			func(w *Wizard) error { return nil }, // Already in MainMenu
			StateOutputDir,
			true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset to MainMenu for each test
			w.Reset()
			
			// Setup: navigate to starting state
			if err := tt.setup(w); err != nil {
				t.Fatalf("Setup failed: %v", err)
			}
			
			err := w.TransitionTo(tt.newState)
			if (err != nil) != tt.wantErr {
				t.Errorf("Wizard.TransitionTo(%v) error = %v, wantErr %v", tt.newState, err, tt.wantErr)
			}
			
			if !tt.wantErr {
				if w.GetState() != tt.newState {
					t.Errorf("Wizard.TransitionTo() state = %v, want %v", w.GetState(), tt.newState)
				}
			}
		})
	}
}

func TestWizard_CanGoBack(t *testing.T) {
	cfg := config.CreateDefaultConfig()
	w, _ := NewWizard(cfg, false)
	
	tests := []struct {
		name    string
		setup   func(*Wizard) error
		state   State
		want    bool
	}{
		{"TemplateSelection", func(w *Wizard) error { return w.TransitionTo(StateTemplateSelection) }, StateTemplateSelection, true},
		{"OutputDir", func(w *Wizard) error {
			if err := w.TransitionTo(StateTemplateSelection); err != nil {
				return err
			}
			return w.TransitionTo(StateOutputDir)
		}, StateOutputDir, true},
		{"Title", func(w *Wizard) error {
			if err := w.TransitionTo(StateTemplateSelection); err != nil {
				return err
			}
			if err := w.TransitionTo(StateOutputDir); err != nil {
				return err
			}
			return w.TransitionTo(StateTitle)
		}, StateTitle, true},
		{"Tags", func(w *Wizard) error {
			if err := w.TransitionTo(StateTemplateSelection); err != nil {
				return err
			}
			if err := w.TransitionTo(StateOutputDir); err != nil {
				return err
			}
			if err := w.TransitionTo(StateTitle); err != nil {
				return err
			}
			return w.TransitionTo(StateTags)
		}, StateTags, true},
		{"Message", func(w *Wizard) error {
			if err := w.TransitionTo(StateTemplateSelection); err != nil {
				return err
			}
			if err := w.TransitionTo(StateOutputDir); err != nil {
				return err
			}
			if err := w.TransitionTo(StateTitle); err != nil {
				return err
			}
			if err := w.TransitionTo(StateTags); err != nil {
				return err
			}
			return w.TransitionTo(StateMessage)
		}, StateMessage, true},
		{"MainMenu", func(w *Wizard) error { return nil }, StateMainMenu, false},
		{"FileCreated", func(w *Wizard) error {
			if err := w.TransitionTo(StateTemplateSelection); err != nil {
				return err
			}
			if err := w.TransitionTo(StateOutputDir); err != nil {
				return err
			}
			if err := w.TransitionTo(StateTitle); err != nil {
				return err
			}
			if err := w.TransitionTo(StateTags); err != nil {
				return err
			}
			if err := w.TransitionTo(StateMessage); err != nil {
				return err
			}
			return w.TransitionTo(StateFileCreated)
		}, StateFileCreated, false},
		{"ReviewScreen", func(w *Wizard) error {
			if err := w.TransitionTo(StateTemplateSelection); err != nil {
				return err
			}
			if err := w.TransitionTo(StateOutputDir); err != nil {
				return err
			}
			if err := w.TransitionTo(StateTitle); err != nil {
				return err
			}
			if err := w.TransitionTo(StateTags); err != nil {
				return err
			}
			if err := w.TransitionTo(StateMessage); err != nil {
				return err
			}
			if err := w.TransitionTo(StateFileCreated); err != nil {
				return err
			}
			if err := w.TransitionTo(StateEditorLaunch); err != nil {
				return err
			}
			return w.TransitionTo(StateReviewScreen)
		}, StateReviewScreen, false},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w.Reset()
			if err := tt.setup(w); err != nil {
				t.Fatalf("Failed to setup state %v: %v", tt.state, err)
			}
			
			if w.GetState() != tt.state {
				t.Fatalf("Setup did not reach state %v, got %v", tt.state, w.GetState())
			}
			
			if got := w.CanGoBack(); got != tt.want {
				t.Errorf("Wizard.CanGoBack() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWizard_GoBack(t *testing.T) {
	cfg := config.CreateDefaultConfig()
	w, _ := NewWizard(cfg, false)
	
	// Navigate forward
	if err := w.TransitionTo(StateTemplateSelection); err != nil {
		t.Fatalf("Failed to transition: %v", err)
	}
	if err := w.TransitionTo(StateOutputDir); err != nil {
		t.Fatalf("Failed to transition: %v", err)
	}
	
	// Go back
	if err := w.GoBack(); err != nil {
		t.Fatalf("Wizard.GoBack() error = %v, want nil", err)
	}
	
	if w.GetState() != StateTemplateSelection {
		t.Errorf("Wizard.GoBack() state = %v, want %v", w.GetState(), StateTemplateSelection)
	}
}

func TestWizard_GoBack_NotAllowed(t *testing.T) {
	cfg := config.CreateDefaultConfig()
	w, _ := NewWizard(cfg, false)
	
	// Try to go back from MainMenu (not allowed)
	err := w.GoBack()
	if err == nil {
		t.Error("Wizard.GoBack() from MainMenu error = nil, want error")
	}
}

func TestWizard_SettersAndGetters(t *testing.T) {
	cfg := config.CreateDefaultConfig()
	w, _ := NewWizard(cfg, false)
	
	// Test setters and getters
	w.SetOutputDir("/tmp/notes")
	if w.GetOutputDir() != "/tmp/notes" {
		t.Errorf("GetOutputDir() = %v, want /tmp/notes", w.GetOutputDir())
	}
	
	w.SetTemplateName("daily")
	if w.GetTemplateName() != "daily" {
		t.Errorf("GetTemplateName() = %v, want daily", w.GetTemplateName())
	}
	
	w.SetTitle("Test Title")
	if w.GetTitle() != "Test Title" {
		t.Errorf("GetTitle() = %v, want Test Title", w.GetTitle())
	}
	
	w.SetTags([]string{"work", "meeting"})
	tags := w.GetTags()
	if len(tags) != 2 || tags[0] != "work" || tags[1] != "meeting" {
		t.Errorf("GetTags() = %v, want [work, meeting]", tags)
	}
	
	w.SetMessage("Test message")
	if w.GetMessage() != "Test message" {
		t.Errorf("GetMessage() = %v, want Test message", w.GetMessage())
	}
}

func TestWizard_Reset(t *testing.T) {
	cfg := config.CreateDefaultConfig()
	w, _ := NewWizard(cfg, false)
	
	// Set some values
	w.SetOutputDir("/tmp/notes")
	w.SetTitle("Test")
	w.SetTags([]string{"tag1"})
	
	// Navigate to a different state
	if err := w.TransitionTo(StateTemplateSelection); err != nil {
		t.Fatalf("Failed to transition: %v", err)
	}
	
	// Reset
	w.Reset()
	
	// Check that everything is reset
	if w.GetState() != StateMainMenu {
		t.Errorf("Reset() state = %v, want %v", w.GetState(), StateMainMenu)
	}
	if w.GetOutputDir() != "" {
		t.Errorf("Reset() outputDir = %v, want empty", w.GetOutputDir())
	}
	if w.GetTitle() != "" {
		t.Errorf("Reset() title = %v, want empty", w.GetTitle())
	}
	if len(w.GetTags()) != 0 {
		t.Errorf("Reset() tags = %v, want empty", w.GetTags())
	}
}

func TestWizard_ValidateState(t *testing.T) {
	cfg := config.CreateDefaultConfig()
	w, _ := NewWizard(cfg, false)
	
	// Navigate to OutputDir state
	if err := w.TransitionTo(StateTemplateSelection); err != nil {
		t.Fatalf("Failed to transition: %v", err)
	}
	if err := w.TransitionTo(StateOutputDir); err != nil {
		t.Fatalf("Failed to transition: %v", err)
	}
	
	// Should fail validation (output dir is empty)
	if err := w.ValidateState(); err == nil {
		t.Error("ValidateState() with empty outputDir error = nil, want error")
	}
	
	// Set output dir
	w.SetOutputDir("/tmp/notes")
	
	// Validation should pass (actual path validation is done in flow.go)
	// ValidateState only checks if required fields are set
	if err := w.ValidateState(); err != nil {
		t.Errorf("ValidateState() with outputDir set error = %v, want nil", err)
	}
}

