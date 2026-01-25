package cli

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sv4u/touchlog/v2/internal/note"
	"github.com/sv4u/touchlog/v2/internal/store"
)

// noteItem represents a note in the selection list
type noteItem struct {
	id    string
	typ   string
	key   string
	title string
	tags  []string
	path  string
	body  string // First 500 chars for search
}

// FilterValue returns the searchable string for fuzzy matching
func (n noteItem) FilterValue() string {
	// Concatenate all searchable fields
	return fmt.Sprintf("%s %s %s %s %s",
		n.title, n.key, n.typ,
		strings.Join(n.tags, " "),
		n.body)
}

// Title returns the title for list display
func (n noteItem) Title() string {
	return fmt.Sprintf("%s:%s", n.typ, n.key)
}

// Description returns the description for list display
func (n noteItem) Description() string {
	tagStr := ""
	if len(n.tags) > 0 {
		tagStr = " [" + strings.Join(n.tags, ", ") + "]"
	}
	return n.title + tagStr
}

// editWizardModel is the bubbletea model for the edit wizard
type editWizardModel struct {
	list      list.Model
	selected  *noteItem
	vaultRoot string
	quitting  bool
	err       error
}

// initialEditModel creates the initial edit wizard model
func initialEditModel(notes []noteItem, vaultRoot string) editWizardModel {
	// Convert notes to list items
	items := make([]list.Item, len(notes))
	for i, n := range notes {
		items[i] = n
	}

	// Create the list
	delegate := newNoteItemDelegate()
	l := list.New(items, delegate, 0, 0)
	l.Title = "Select a note to edit"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.SetShowHelp(true)
	l.Styles.Title = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("170")).
		MarginLeft(2)

	return editWizardModel{
		list:      l,
		vaultRoot: vaultRoot,
	}
}

// Init initializes the model
func (m editWizardModel) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m editWizardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Don't handle keys when filtering
		if m.list.FilterState() == list.Filtering {
			break
		}

		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		case "enter":
			if item, ok := m.list.SelectedItem().(noteItem); ok {
				m.selected = &item
				return m, tea.Quit
			}
		}

	case tea.WindowSizeMsg:
		h, v := lipgloss.NewStyle().Margin(1, 2).GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// View renders the UI
func (m editWizardModel) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\n\nPress q to quit.", m.err)
	}

	if m.quitting {
		return ""
	}

	return lipgloss.NewStyle().Margin(1, 2).Render(m.list.View())
}

// noteItemDelegate is a custom delegate for rendering note items
type noteItemDelegate struct {
	styles noteItemStyles
}

type noteItemStyles struct {
	normalTitle   lipgloss.Style
	normalDesc    lipgloss.Style
	selectedTitle lipgloss.Style
	selectedDesc  lipgloss.Style
	dimmedTitle   lipgloss.Style
	dimmedDesc    lipgloss.Style
	filterMatch   lipgloss.Style
}

func newNoteItemDelegate() noteItemDelegate {
	return noteItemDelegate{
		styles: noteItemStyles{
			normalTitle: lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Light: "#1a1a1a", Dark: "#dddddd"}).
				Padding(0, 0, 0, 2),
			normalDesc: lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Light: "#a49fa5", Dark: "#777777"}).
				Padding(0, 0, 0, 2),
			selectedTitle: lipgloss.NewStyle().
				Border(lipgloss.NormalBorder(), false, false, false, true).
				BorderForeground(lipgloss.AdaptiveColor{Light: "#f793ff", Dark: "#ad58b4"}).
				Foreground(lipgloss.AdaptiveColor{Light: "#ee6ff8", Dark: "#ee6ff8"}).
				Padding(0, 0, 0, 1),
			selectedDesc: lipgloss.NewStyle().
				Border(lipgloss.NormalBorder(), false, false, false, true).
				BorderForeground(lipgloss.AdaptiveColor{Light: "#f793ff", Dark: "#ad58b4"}).
				Foreground(lipgloss.AdaptiveColor{Light: "#f793ff", Dark: "#ad58b4"}).
				Padding(0, 0, 0, 1),
			dimmedTitle: lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Light: "#a49fa5", Dark: "#777777"}).
				Padding(0, 0, 0, 2),
			dimmedDesc: lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Light: "#c2b8c2", Dark: "#4d4d4d"}).
				Padding(0, 0, 0, 2),
			filterMatch: lipgloss.NewStyle().
				Underline(true),
		},
	}
}

// Height returns the height of a single item
func (d noteItemDelegate) Height() int {
	return 2
}

// Spacing returns the spacing between items
func (d noteItemDelegate) Spacing() int {
	return 1
}

// Update handles item-level updates
func (d noteItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd {
	return nil
}

// Render renders an item
func (d noteItemDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	n, ok := item.(noteItem)
	if !ok {
		return
	}

	title := n.Title()
	desc := n.Description()

	if m.Index() == index {
		// Selected item
		title = d.styles.selectedTitle.Render(title)
		desc = d.styles.selectedDesc.Render(desc)
	} else if m.FilterState() == list.Filtering && index != m.Index() {
		// Dimmed during filtering
		title = d.styles.dimmedTitle.Render(title)
		desc = d.styles.dimmedDesc.Render(desc)
	} else {
		// Normal item
		title = d.styles.normalTitle.Render(title)
		desc = d.styles.normalDesc.Render(desc)
	}

	fmt.Fprintf(w, "%s\n%s", title, desc)
}

// loadNotesForEdit loads all notes from the vault with optional filters
func loadNotesForEdit(vaultRoot string, typeFilter string, tagFilters []string) ([]noteItem, error) {
	// Open database
	db, err := store.OpenOrCreateDB(vaultRoot)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}
	defer func() {
		_ = db.Close()
	}()

	// Build query
	query := `
		SELECT 
			n.id,
			n.type,
			n.key,
			n.title,
			n.path,
			COALESCE(json_group_array(t.tag), '[]') as tags
		FROM nodes n
		LEFT JOIN tags t ON n.id = t.node_id
	`

	var args []interface{}
	var whereParts []string

	if typeFilter != "" {
		whereParts = append(whereParts, "n.type = ?")
		args = append(args, typeFilter)
	}

	if len(tagFilters) > 0 {
		// Match notes that have any of the specified tags
		placeholders := make([]string, len(tagFilters))
		for i, tag := range tagFilters {
			placeholders[i] = "?"
			args = append(args, tag)
		}
		whereParts = append(whereParts, fmt.Sprintf(
			"EXISTS (SELECT 1 FROM tags t2 WHERE t2.node_id = n.id AND t2.tag IN (%s))",
			strings.Join(placeholders, ","),
		))
	}

	if len(whereParts) > 0 {
		query += " WHERE " + strings.Join(whereParts, " AND ")
	}

	query += " GROUP BY n.id, n.type, n.key, n.title, n.path"
	query += " ORDER BY n.updated DESC"

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying notes: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	var notes []noteItem
	for rows.Next() {
		var n noteItem
		var tagsJSON sql.NullString
		if err := rows.Scan(&n.id, &n.typ, &n.key, &n.title, &n.path, &tagsJSON); err != nil {
			return nil, fmt.Errorf("scanning row: %w", err)
		}

		// Parse tags
		if tagsJSON.Valid {
			n.tags = parseTagsFromJSON(tagsJSON.String)
		}

		notes = append(notes, n)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Load body content for each note (truncated for search)
	// Paths in database are already absolute
	for i := range notes {
		body, err := loadNoteBody(notes[i].path)
		if err != nil {
			// Skip notes with missing files
			continue
		}
		notes[i].body = truncateString(body, 500)
	}

	return notes, nil
}

// loadNoteBody reads and parses a note file to extract the body
// The path is expected to be absolute (as stored in the database)
func loadNoteBody(absPath string) (string, error) {
	content, err := os.ReadFile(absPath)
	if err != nil {
		return "", err
	}

	parsed := note.Parse(absPath, content)
	return parsed.Body, nil
}

// truncateString truncates a string to maxLen runes (characters), preserving valid UTF-8.
// This handles multi-byte characters correctly (e.g., CJK, emoji, accented letters).
func truncateString(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen])
}

// parseTagsFromJSON parses tags from a JSON array string
// This uses json.Unmarshal to correctly handle all JSON escaping, including tags with commas.
// It also filters out empty strings that result from JSON null values (e.g., when SQLite's
// json_group_array returns [null] for notes with no tags via LEFT JOIN).
func parseTagsFromJSON(jsonStr string) []string {
	var rawTags []string
	if err := json.Unmarshal([]byte(jsonStr), &rawTags); err != nil {
		// If parsing fails, return empty slice
		return []string{}
	}

	// Filter out empty strings (which result from JSON null values)
	tags := make([]string, 0, len(rawTags))
	for _, tag := range rawTags {
		if tag != "" {
			tags = append(tags, tag)
		}
	}
	return tags
}
