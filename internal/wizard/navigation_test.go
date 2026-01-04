package wizard

import (
	"testing"
)

func TestNavigationHistory(t *testing.T) {
	h := NewNavigationHistory()
	
	// Test initial state
	if h.CanNavigateBack() {
		t.Error("NewNavigationHistory().CanNavigateBack() = true, want false")
	}
	
	// Push states
	h.Push(StateMainMenu)
	h.Push(StateTemplateSelection)
	h.Push(StateOutputDir)
	
	// Should be able to go back now
	if !h.CanNavigateBack() {
		t.Error("CanNavigateBack() after pushing states = false, want true")
	}
	
	// Test Peek
	last, ok := h.Peek()
	if !ok {
		t.Error("Peek() ok = false, want true")
	}
	if last != StateOutputDir {
		t.Errorf("Peek() = %v, want %v", last, StateOutputDir)
	}
	
	// Test Pop
	popped, ok := h.Pop()
	if !ok {
		t.Error("Pop() ok = false, want true")
	}
	if popped != StateOutputDir {
		t.Errorf("Pop() = %v, want %v", popped, StateOutputDir)
	}
	
	// Peek should now return previous state
	last, ok = h.Peek()
	if !ok {
		t.Error("Peek() after Pop ok = false, want true")
	}
	if last != StateTemplateSelection {
		t.Errorf("Peek() after Pop = %v, want %v", last, StateTemplateSelection)
	}
}

func TestNavigationHistory_Empty(t *testing.T) {
	h := NewNavigationHistory()
	
	// Test Pop on empty history
	_, ok := h.Pop()
	if ok {
		t.Error("Pop() on empty history ok = true, want false")
	}
	
	// Test Peek on empty history
	_, ok = h.Peek()
	if ok {
		t.Error("Peek() on empty history ok = true, want false")
	}
}

func TestNavigationHistory_Clear(t *testing.T) {
	h := NewNavigationHistory()
	
	// Push some states
	h.Push(StateMainMenu)
	h.Push(StateTemplateSelection)
	
	// Clear
	h.Clear()
	
	// Should not be able to go back
	if h.CanNavigateBack() {
		t.Error("CanNavigateBack() after Clear = true, want false")
	}
	
	// History should be empty
	history := h.GetHistory()
	if len(history) != 0 {
		t.Errorf("GetHistory() after Clear length = %d, want 0", len(history))
	}
}

func TestNavigationHistory_GetHistory(t *testing.T) {
	h := NewNavigationHistory()
	
	// Push states
	states := []State{StateMainMenu, StateTemplateSelection, StateOutputDir}
	for _, s := range states {
		h.Push(s)
	}
	
	// Get history
	history := h.GetHistory()
	
	// Check length
	if len(history) != len(states) {
		t.Errorf("GetHistory() length = %d, want %d", len(history), len(states))
	}
	
	// Check values
	for i, want := range states {
		if history[i] != want {
			t.Errorf("GetHistory()[%d] = %v, want %v", i, history[i], want)
		}
	}
	
	// Modify returned slice should not affect original
	history[0] = StateReviewScreen
	history2 := h.GetHistory()
	if history2[0] == StateReviewScreen {
		t.Error("GetHistory() returned mutable slice")
	}
}

