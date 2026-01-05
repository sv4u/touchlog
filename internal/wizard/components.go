package wizard

import (
	"github.com/charmbracelet/lipgloss"
)

// Styles for wizard components
var (
	// TitleStyle is the style for titles
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("62"))

	// HelpStyle is the style for help text
	HelpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Italic(true)

	// ErrorStyle is the style for error messages
	ErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("1")).
			Bold(true)

	// SuccessStyle is the style for success messages
	SuccessStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("2")).
			Bold(true)

	// BorderStyle is the style for borders
	BorderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(1, 2)

	// PreviewStyle is the style for file preview
	PreviewStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(1, 2)
)
