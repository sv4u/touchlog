package entry

import (
	"testing"
)

func TestGenerateSlug(t *testing.T) {
	tests := []struct {
		name    string
		title   string
		message string
		want    string
	}{
		{
			name:    "simple title",
			title:   "Standup Notes",
			message: "",
			want:    "standup-notes",
		},
		{
			name:    "title with special characters",
			title:   "Meeting: Q4 Planning!",
			message: "",
			want:    "meeting-q4-planning",
		},
		{
			name:    "title with multiple spaces",
			title:   "Daily  Standup   Notes",
			message: "",
			want:    "daily-standup-notes",
		},
		{
			name:    "empty title, use first line of message",
			title:   "",
			message: "Quick note about the project",
			want:    "quick-note-about-the-project",
		},
		{
			name:    "empty title and message",
			title:   "",
			message: "",
			want:    "untitled",
		},
		{
			name:    "title with unicode characters",
			title:   "Café Meeting",
			message: "",
			want:    "caf-meeting",
		},
		{
			name:    "very long title",
			title:   "This is a very long title that should be truncated to fifty characters maximum",
			message: "",
			want:    "this-is-a-very-long-title-that-should-be-truncated",
		},
		{
			name:    "title with leading/trailing hyphens",
			title:   "-Important Note-",
			message: "",
			want:    "important-note",
		},
		{
			name:    "title with only special characters",
			title:   "!!!@@@###",
			message: "",
			want:    "untitled",
		},
		{
			name:    "title with numbers",
			title:   "Project 2024 Q1",
			message: "",
			want:    "project-2024-q1",
		},
		{
			name:    "message with newlines",
			title:   "",
			message: "First line\nSecond line\nThird line",
			want:    "first-line",
		},
		{
			name:    "title takes precedence over message",
			title:   "My Title",
			message: "First line of message",
			want:    "my-title",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateSlug(tt.title, tt.message)
			if got != tt.want {
				t.Errorf("GenerateSlug(%q, %q) = %q, want %q", tt.title, tt.message, got, tt.want)
			}
		})
	}
}

func TestGenerateSlugMaxLength(t *testing.T) {
	// Test that slug is truncated to MaxSlugLength
	longTitle := "a"
	for i := 0; i < 100; i++ {
		longTitle += "b"
	}

	slug := GenerateSlug(longTitle, "")
	if len(slug) > MaxSlugLength {
		t.Errorf("GenerateSlug() returned slug of length %d, expected <= %d", len(slug), MaxSlugLength)
	}
}

func TestGenerateSlugDeterministic(t *testing.T) {
	// Test that same input produces same output
	title := "Test Title"
	message := "Test message"

	slug1 := GenerateSlug(title, message)
	slug2 := GenerateSlug(title, message)

	if slug1 != slug2 {
		t.Errorf("GenerateSlug() is not deterministic: got %q and %q", slug1, slug2)
	}
}
