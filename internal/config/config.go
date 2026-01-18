package config

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/sv4u/touchlog/v2/internal/model"
	"gopkg.in/yaml.v3"
)

// Config represents the complete touchlog configuration
type Config struct {
	Version   int
	Types     map[model.TypeName]TypeDef
	Tags      TagConfig
	Edges     map[model.EdgeType]EdgeDef
	Templates TemplateConfig
}

// TypeDef defines a note type
type TypeDef struct {
	Description       string
	DefaultState      string
	RequiredFields    []string // dot paths
	RecommendedFields []string
	KeyPattern        *regexp.Regexp
	KeyMaxLen         int
}

// TagConfig defines tag-related configuration
type TagConfig struct {
	Preferred []string
}

// EdgeDef defines an edge type
type EdgeDef struct {
	Description string
	AllowedFrom []model.TypeName
	AllowedTo   []model.TypeName
}

// TemplateConfig defines template configuration
type TemplateConfig struct {
	Root string
}

// DefaultKeyPattern is the default regex pattern for keys
const DefaultKeyPattern = `^[a-z0-9]+(-[a-z0-9]+)*$`

// DefaultKeyMaxLen is the default maximum length for keys
const DefaultKeyMaxLen = 64

// LoadConfig loads and merges configuration from built-in, global, and repo sources
// Merge precedence: built-in < global < repo
func LoadConfig(vaultRoot string) (*Config, error) {
	cfg := &Config{
		Types:     make(map[model.TypeName]TypeDef),
		Edges:     make(map[model.EdgeType]EdgeDef),
		Tags:      TagConfig{Preferred: []string{}},
		Templates: TemplateConfig{Root: ""},
	}

	// Start with built-in defaults
	if err := applyBuiltInDefaults(cfg); err != nil {
		return nil, fmt.Errorf("applying built-in defaults: %w", err)
	}

	// Merge global config if it exists
	globalPath := filepath.Join(os.Getenv("HOME"), ".config", "touchlog", "config.yaml")
	if _, err := os.Stat(globalPath); err == nil {
		if err := mergeConfigFile(cfg, globalPath); err != nil {
			return nil, fmt.Errorf("loading global config: %w", err)
		}
	}

	// Merge repo config if it exists
	repoPath := filepath.Join(vaultRoot, ".touchlog", "config.yaml")
	if _, err := os.Stat(repoPath); err == nil {
		if err := mergeConfigFile(cfg, repoPath); err != nil {
			return nil, fmt.Errorf("loading repo config: %w", err)
		}
	}

	// Validate the merged config
	if err := validateConfig(cfg); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	return cfg, nil
}

// applyBuiltInDefaults applies built-in default configuration
func applyBuiltInDefaults(cfg *Config) error {
	cfg.Version = model.ConfigSchemaVersion
	// Built-in defaults can be extended here
	return nil
}

// mergeConfigFile merges a YAML config file into the existing config
func mergeConfigFile(cfg *Config, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading config file: %w", err)
	}

	var raw map[string]any
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("parsing YAML: %w", err)
	}

	// Validate unknown top-level keys
	allowedKeys := map[string]bool{
		"version":   true,
		"types":     true,
		"tags":      true,
		"edges":     true,
		"templates": true,
	}

	for key := range raw {
		if !allowedKeys[key] {
			return fmt.Errorf("unknown top-level key: %q", key)
		}
	}

	// Check version
	if version, ok := raw["version"].(int); ok {
		if version != model.ConfigSchemaVersion {
			return fmt.Errorf("unsupported config version: %d (expected %d)", version, model.ConfigSchemaVersion)
		}
		cfg.Version = version
	} else if _, exists := raw["version"]; !exists {
		return fmt.Errorf("missing required field: version")
	}

	// Parse types
	if typesRaw, ok := raw["types"].(map[string]any); ok {
		for typeName, typeDefRaw := range typesRaw {
			typeDef, err := parseTypeDef(typeDefRaw)
			if err != nil {
				return fmt.Errorf("parsing type %q: %w", typeName, err)
			}
			cfg.Types[model.TypeName(typeName)] = typeDef
		}
	}

	// Parse tags
	if tagsRaw, ok := raw["tags"].(map[string]any); ok {
		if preferred, ok := tagsRaw["preferred"].([]any); ok {
			cfg.Tags.Preferred = make([]string, 0, len(preferred))
			for _, p := range preferred {
				if s, ok := p.(string); ok {
					cfg.Tags.Preferred = append(cfg.Tags.Preferred, s)
				}
			}
		}
	}

	// Parse edges
	if edgesRaw, ok := raw["edges"].(map[string]any); ok {
		for edgeType, edgeDefRaw := range edgesRaw {
			edgeDef, err := parseEdgeDef(edgeDefRaw)
			if err != nil {
				return fmt.Errorf("parsing edge %q: %w", edgeType, err)
			}
			cfg.Edges[model.EdgeType(edgeType)] = edgeDef
		}
	}

	// Parse templates
	if templatesRaw, ok := raw["templates"].(map[string]any); ok {
		if root, ok := templatesRaw["root"].(string); ok {
			cfg.Templates.Root = root
		}
	}

	return nil
}

// parseTypeDef parses a TypeDef from raw YAML data
func parseTypeDef(raw any) (TypeDef, error) {
	td := TypeDef{
		KeyPattern: regexp.MustCompile(DefaultKeyPattern),
		KeyMaxLen:  DefaultKeyMaxLen,
	}

	m, ok := raw.(map[string]any)
	if !ok {
		return td, fmt.Errorf("expected map, got %T", raw)
	}

	if desc, ok := m["description"].(string); ok {
		td.Description = desc
	}

	if defaultState, ok := m["default_state"].(string); ok {
		td.DefaultState = defaultState
	} else {
		// DefaultState is required
		return td, fmt.Errorf("missing required field: default_state")
	}

	if required, ok := m["required_fields"].([]any); ok {
		td.RequiredFields = make([]string, 0, len(required))
		for _, r := range required {
			if s, ok := r.(string); ok {
				td.RequiredFields = append(td.RequiredFields, s)
			}
		}
	}

	if recommended, ok := m["recommended_fields"].([]any); ok {
		td.RecommendedFields = make([]string, 0, len(recommended))
		for _, r := range recommended {
			if s, ok := r.(string); ok {
				td.RecommendedFields = append(td.RecommendedFields, s)
			}
		}
	}

	if pattern, ok := m["key_pattern"].(string); ok {
		re, err := regexp.Compile(pattern)
		if err != nil {
			return td, fmt.Errorf("invalid key_pattern regex: %w", err)
		}
		td.KeyPattern = re
	}

	if maxLen, ok := m["key_max_len"].(int); ok {
		td.KeyMaxLen = maxLen
	}

	return td, nil
}

// parseEdgeDef parses an EdgeDef from raw YAML data
func parseEdgeDef(raw any) (EdgeDef, error) {
	ed := EdgeDef{}

	m, ok := raw.(map[string]any)
	if !ok {
		return ed, fmt.Errorf("expected map, got %T", raw)
	}

	if desc, ok := m["description"].(string); ok {
		ed.Description = desc
	}

	if allowedFrom, ok := m["allowed_from"].([]any); ok {
		ed.AllowedFrom = make([]model.TypeName, 0, len(allowedFrom))
		for _, af := range allowedFrom {
			if s, ok := af.(string); ok {
				ed.AllowedFrom = append(ed.AllowedFrom, model.TypeName(s))
			}
		}
	}

	if allowedTo, ok := m["allowed_to"].([]any); ok {
		ed.AllowedTo = make([]model.TypeName, 0, len(allowedTo))
		for _, at := range allowedTo {
			if s, ok := at.(string); ok {
				ed.AllowedTo = append(ed.AllowedTo, model.TypeName(s))
			}
		}
	}

	return ed, nil
}

// validateConfig performs strict validation on the merged config
func validateConfig(cfg *Config) error {
	if cfg.Version != model.ConfigSchemaVersion {
		return fmt.Errorf("unsupported config version: %d (expected %d)", cfg.Version, model.ConfigSchemaVersion)
	}

	// Validate types
	for typeName, typeDef := range cfg.Types {
		if typeDef.DefaultState == "" {
			return fmt.Errorf("type %q: default_state must be non-empty", typeName)
		}
		if typeDef.KeyPattern == nil {
			return fmt.Errorf("type %q: key_pattern must be set", typeName)
		}
		if typeDef.KeyMaxLen <= 0 {
			return fmt.Errorf("type %q: key_max_len must be positive", typeName)
		}
	}

	return nil
}
