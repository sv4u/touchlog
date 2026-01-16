package model

import "time"

// Type aliases for canonical contracts
type NoteID string
type EdgeType string
type TypeName string
type Key string

// TypeKey represents a qualified note reference (type:key)
type TypeKey struct {
	Type TypeName
	Key  Key
}

// Frontmatter represents the YAML frontmatter of a note
type Frontmatter struct {
	ID      NoteID
	Type    TypeName
	Key     Key
	Created time.Time
	Updated time.Time
	Title   string
	Tags    []string
	State   string
	// Extra preserves unknown fields from YAML frontmatter
	Extra map[string]any
}

// Note represents a complete parsed note
type Note struct {
	FM       Frontmatter
	Path     string
	Body     string
	RawLinks []RawLink
	Diags    []Diagnostic
}

// RawLink represents a link extracted from note body
type RawLink struct {
	Source       TypeKey
	Target       RawTarget
	EdgeType     EdgeType
	Span         Span
	ResolvedToID *NoteID // Set during indexing when link is resolved (Phase 2+)
}

// RawTarget represents a link target (may be unresolved)
type RawTarget struct {
	// Type is nil if unqualified (e.g., [[key]] vs [[type:key]])
	Type *TypeName
	Key  Key
	// Label is reserved for future use
	Label *string
}

// Diagnostic represents a parse error or warning
type Diagnostic struct {
	Level   string // "info", "warn", or "error"
	Code    string
	Message string
	Span    Span
}

// Span represents a location in source code
type Span struct {
	Path      string
	StartByte int
	EndByte   int
	StartLine int // optional in v0
	StartCol  int // optional in v0
}

// Constants

// DefaultEdgeType is used when no edge type is specified in a link
const DefaultEdgeType EdgeType = "related-to"

// Diagnostic levels
const (
	DiagnosticLevelInfo  = "info"
	DiagnosticLevelWarn  = "warn"
	DiagnosticLevelError = "error"
)

// Version constants
const (
	ConfigSchemaVersion = 1
	IndexSchemaVersion  = 1
	ProtocolVersion     = 1
)
