package wizard

// NavigationHistory provides utilities for managing navigation history
type NavigationHistory struct {
	states []State
}

// NewNavigationHistory creates a new navigation history
func NewNavigationHistory() *NavigationHistory {
	return &NavigationHistory{
		states: make([]State, 0),
	}
}

// Push adds a state to the history
func (h *NavigationHistory) Push(state State) {
	h.states = append(h.states, state)
}

// Pop removes and returns the last state from history
func (h *NavigationHistory) Pop() (State, bool) {
	if len(h.states) == 0 {
		return -1, false
	}
	last := h.states[len(h.states)-1]
	h.states = h.states[:len(h.states)-1]
	return last, true
}

// Peek returns the last state without removing it
func (h *NavigationHistory) Peek() (State, bool) {
	if len(h.states) == 0 {
		return -1, false
	}
	return h.states[len(h.states)-1], true
}

// Clear clears the navigation history
func (h *NavigationHistory) Clear() {
	h.states = h.states[:0]
}

// GetHistory returns a copy of the navigation history
func (h *NavigationHistory) GetHistory() []State {
	result := make([]State, len(h.states))
	copy(result, h.states)
	return result
}

// CanNavigateBack checks if back navigation is possible
func (h *NavigationHistory) CanNavigateBack() bool {
	return len(h.states) > 1 // Need at least 2 states to go back
}
