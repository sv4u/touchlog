package wizard

import (
	"testing"
	"time"

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

func TestWizard_FilePathSettersAndGetters(t *testing.T) {
	cfg := config.CreateDefaultConfig()
	w, _ := NewWizard(cfg, false)
	
	// Test temp file path
	w.SetTempFilePath("/tmp/temp.md")
	if w.GetTempFilePath() != "/tmp/temp.md" {
		t.Errorf("GetTempFilePath() = %v, want /tmp/temp.md", w.GetTempFilePath())
	}
	
	// Test final file path
	w.SetFinalFilePath("/tmp/final.md")
	if w.GetFinalFilePath() != "/tmp/final.md" {
		t.Errorf("GetFinalFilePath() = %v, want /tmp/final.md", w.GetFinalFilePath())
	}
	
	// Test file content
	content := "# Test\n\nContent here"
	w.SetFileContent(content)
	if w.GetFileContent() != content {
		t.Errorf("GetFileContent() = %v, want %v", w.GetFileContent(), content)
	}
}

func TestWizard_Timestamp(t *testing.T) {
	cfg := config.CreateDefaultConfig()
	w, _ := NewWizard(cfg, false)
	
	// Timestamp should be set on creation
	timestamp := w.GetTimestamp()
	if timestamp.IsZero() {
		t.Error("GetTimestamp() returned zero time")
	}
	
	// Timestamp should be recent (within last second)
	now := time.Now()
	diff := now.Sub(timestamp)
	if diff < 0 {
		diff = -diff
	}
	if diff > time.Second {
		t.Errorf("GetTimestamp() = %v, want time within last second", timestamp)
	}
}

func TestWizard_ValidateState_AllStates(t *testing.T) {
	cfg := config.CreateDefaultConfig()
	w, _ := NewWizard(cfg, false)
	
	// Test validation for all states
	states := []State{
		StateMainMenu,
		StateTemplateSelection,
		StateOutputDir,
		StateTitle,
		StateTags,
		StateMessage,
		StateFileCreated,
		StateEditorLaunch,
		StateReviewScreen,
	}
	
	for _, state := range states {
		t.Run(state.String(), func(t *testing.T) {
			w.Reset()
			
			// Navigate to state (if possible)
			// For states that require navigation, we'll try to get there
			// For MainMenu, we're already there
			if state != StateMainMenu {
				// Try to navigate - may fail for some states, that's okay
				_ = w.TransitionTo(state)
			}
			
			// ValidateState should not panic
			// Some states may return errors (like OutputDir with empty dir)
			// but that's expected behavior
			err := w.ValidateState()
			_ = err // Use error to avoid unused variable
		})
	}
}

func TestWizard_Reset_AllFields(t *testing.T) {
	cfg := config.CreateDefaultConfig()
	w, _ := NewWizard(cfg, false)
	
	// Set all fields
	w.SetOutputDir("/tmp/notes")
	w.SetTemplateName("daily")
	w.SetTitle("Test Title")
	w.SetTags([]string{"tag1", "tag2"})
	w.SetMessage("Test message")
	w.SetTempFilePath("/tmp/temp.md")
	w.SetFinalFilePath("/tmp/final.md")
	w.SetFileContent("# Content")
	
	// Navigate to a different state
	if err := w.TransitionTo(StateTemplateSelection); err != nil {
		t.Fatalf("Failed to transition: %v", err)
	}
	if err := w.TransitionTo(StateOutputDir); err != nil {
		t.Fatalf("Failed to transition: %v", err)
	}
	
	// Reset
	w.Reset()
	
	// Verify all fields are reset
	if w.GetState() != StateMainMenu {
		t.Errorf("Reset() state = %v, want %v", w.GetState(), StateMainMenu)
	}
	if w.GetOutputDir() != "" {
		t.Errorf("Reset() outputDir = %v, want empty", w.GetOutputDir())
	}
	if w.GetTemplateName() != "" {
		t.Errorf("Reset() templateName = %v, want empty", w.GetTemplateName())
	}
	if w.GetTitle() != "" {
		t.Errorf("Reset() title = %v, want empty", w.GetTitle())
	}
	if len(w.GetTags()) != 0 {
		t.Errorf("Reset() tags = %v, want empty", w.GetTags())
	}
	if w.GetMessage() != "" {
		t.Errorf("Reset() message = %v, want empty", w.GetMessage())
	}
	if w.GetTempFilePath() != "" {
		t.Errorf("Reset() tempFilePath = %v, want empty", w.GetTempFilePath())
	}
	if w.GetFinalFilePath() != "" {
		t.Errorf("Reset() finalFilePath = %v, want empty", w.GetFinalFilePath())
	}
	if w.GetFileContent() != "" {
		t.Errorf("Reset() fileContent = %v, want empty", w.GetFileContent())
	}
	
	// Timestamp should be updated (recent)
	timestamp := w.GetTimestamp()
	now := time.Now()
	diff := now.Sub(timestamp)
	if diff < 0 {
		diff = -diff
	}
	if diff > time.Second {
		t.Errorf("Reset() timestamp = %v, want time within last second", timestamp)
	}
}

func TestWizard_NewWizard_IncludeGit(t *testing.T) {
	cfg := config.CreateDefaultConfig()
	
	// Test with includeGit = true
	w1, err := NewWizard(cfg, true)
	if err != nil {
		t.Fatalf("NewWizard() error = %v, want nil", err)
	}
	// We can't easily test that includeGit is set without accessing private fields
	// But we can verify the wizard was created successfully
	if w1 == nil {
		t.Fatal("NewWizard() returned nil wizard")
	}
	
	// Test with includeGit = false
	w2, err := NewWizard(cfg, false)
	if err != nil {
		t.Fatalf("NewWizard() error = %v, want nil", err)
	}
	if w2 == nil {
		t.Fatal("NewWizard() returned nil wizard")
	}
}

func TestWizard_TransitionTo_HistoryTracking(t *testing.T) {
	cfg := config.CreateDefaultConfig()
	w, _ := NewWizard(cfg, false)
	
	// Initial state should be in history
	if w.GetState() != StateMainMenu {
		t.Fatalf("Initial state = %v, want %v", w.GetState(), StateMainMenu)
	}
	
	// Transition to TemplateSelection
	if err := w.TransitionTo(StateTemplateSelection); err != nil {
		t.Fatalf("Failed to transition: %v", err)
	}
	
	// State should be updated
	if w.GetState() != StateTemplateSelection {
		t.Errorf("State after transition = %v, want %v", w.GetState(), StateTemplateSelection)
	}
	
	// Transition to OutputDir
	if err := w.TransitionTo(StateOutputDir); err != nil {
		t.Fatalf("Failed to transition: %v", err)
	}
	
	// State should be updated
	if w.GetState() != StateOutputDir {
		t.Errorf("State after transition = %v, want %v", w.GetState(), StateOutputDir)
	}
}

func TestWizard_GoBack_MultipleSteps(t *testing.T) {
	cfg := config.CreateDefaultConfig()
	w, _ := NewWizard(cfg, false)
	
	// Navigate forward multiple steps
	if err := w.TransitionTo(StateTemplateSelection); err != nil {
		t.Fatalf("Failed to transition: %v", err)
	}
	if err := w.TransitionTo(StateOutputDir); err != nil {
		t.Fatalf("Failed to transition: %v", err)
	}
	if err := w.TransitionTo(StateTitle); err != nil {
		t.Fatalf("Failed to transition: %v", err)
	}
	
	// Go back once
	if err := w.GoBack(); err != nil {
		t.Fatalf("Wizard.GoBack() error = %v, want nil", err)
	}
	if w.GetState() != StateOutputDir {
		t.Errorf("State after GoBack() = %v, want %v", w.GetState(), StateOutputDir)
	}
	
	// Go back again
	if err := w.GoBack(); err != nil {
		t.Fatalf("Wizard.GoBack() error = %v, want nil", err)
	}
	if w.GetState() != StateTemplateSelection {
		t.Errorf("State after second GoBack() = %v, want %v", w.GetState(), StateTemplateSelection)
	}
}

