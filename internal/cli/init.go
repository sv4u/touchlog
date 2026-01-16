package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/sv4u/touchlog/internal/config"
	"github.com/sv4u/touchlog/internal/model"
	cli3 "github.com/urfave/cli/v3"
	"gopkg.in/yaml.v3"
)

// TypeBundle represents a predefined set of note types
type TypeBundle struct {
	Name        string
	Description string
	Types       map[string]TypeDef
}

// TypeDef represents a type definition for the config
type TypeDef struct {
	Description  string
	DefaultState string
}

// Built-in type bundles
var typeBundles = map[string]TypeBundle{
	"note": {
		Name:        "note",
		Description: "Basic note-taking",
		Types: map[string]TypeDef{
			"note": {
				Description:  "A general note",
				DefaultState: "draft",
			},
		},
	},
	"log": {
		Name:        "log",
		Description: "Journaling and logging",
		Types: map[string]TypeDef{
			"log": {
				Description:  "A log entry",
				DefaultState: "draft",
			},
		},
	},
	"decision": {
		Name:        "decision",
		Description: "Decision records",
		Types: map[string]TypeDef{
			"decision": {
				Description:  "A decision record",
				DefaultState: "draft",
			},
		},
	},
	"task": {
		Name:        "task",
		Description: "Task management",
		Types: map[string]TypeDef{
			"task": {
				Description:  "A task",
				DefaultState: "draft",
			},
		},
	},
	"learning": {
		Name:        "learning",
		Description: "Learning notes",
		Types: map[string]TypeDef{
			"learning": {
				Description:  "A learning note",
				DefaultState: "draft",
			},
		},
	},
	"competitive": {
		Name:        "competitive",
		Description: "Competitive analysis",
		Types: map[string]TypeDef{
			"competitive": {
				Description:  "A competitive analysis note",
				DefaultState: "draft",
			},
		},
	},
}

// BuildInitCommand builds the init command
func BuildInitCommand() *cli3.Command {
	return &cli3.Command{
		Name:  "init",
		Usage: "Initialize a new touchlog vault",
		Action: func(ctx context.Context, cmd *cli3.Command) error {
			// Get vault root from context or use current directory
			vaultRoot, err := GetVaultFromContext(ctx, cmd)
			if err != nil {
				// If no vault found, use current directory
				cwd, cwdErr := os.Getwd()
				if cwdErr != nil {
					return fmt.Errorf("getting current directory: %w", cwdErr)
				}
				vaultRoot = cwd
			}

			return runInitWizard(vaultRoot)
		},
	}
}

// runInitWizard runs the interactive initialization wizard
func runInitWizard(vaultRoot string) error {
	// Check if vault already exists
	touchlogDir := filepath.Join(vaultRoot, ".touchlog")
	if _, statErr := os.Stat(touchlogDir); statErr == nil {
		configPath := filepath.Join(touchlogDir, "config.yaml")
		if _, statErr := os.Stat(configPath); statErr == nil {
			return fmt.Errorf("vault already initialized at %s", vaultRoot)
		}
	}

	fmt.Printf("Initializing touchlog vault at: %s\n\n", vaultRoot)

	// Step 1: Confirm vault root (for now, just use it)
	// In a full implementation, we'd prompt: "Is this correct? (Y/n)"

	// Step 2: Select type bundles
	fmt.Println("Available type bundles:")
	bundleNames := make([]string, 0, len(typeBundles))
	for name := range typeBundles {
		bundleNames = append(bundleNames, name)
		fmt.Printf("  - %s: %s\n", name, typeBundles[name].Description)
	}
	fmt.Println()

	// For Phase 1, we'll select all bundles by default
	// In a full implementation, we'd prompt for selection
	selectedBundles := bundleNames
	fmt.Printf("Selected bundles: %s\n\n", strings.Join(selectedBundles, ", "))

	// Step 3: Generate config
	cfg, err := generateConfig(selectedBundles)
	if err != nil {
		return fmt.Errorf("generating config: %w", err)
	}

	// Step 4: Write config.yaml
	// touchlogDir already declared above, reuse it
	if err := os.MkdirAll(touchlogDir, 0755); err != nil {
		return fmt.Errorf("creating .touchlog directory: %w", err)
	}

	configPath := filepath.Join(touchlogDir, "config.yaml")
	configYAML, err := marshalConfig(cfg)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	if err := AtomicWrite(configPath, configYAML); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}

	fmt.Printf("✓ Created %s\n", configPath)

	// Step 5: Optionally install templates (skip for Phase 1)
	// Step 6: Optionally create type directories (skip for Phase 1)

	fmt.Println("\nVault initialized successfully!")
	return nil
}

// generateConfig generates a config from selected type bundles
func generateConfig(selectedBundles []string) (*config.Config, error) {
	cfg := &config.Config{
		Version: model.ConfigSchemaVersion,
		Types:   make(map[model.TypeName]config.TypeDef),
		Tags:    config.TagConfig{Preferred: []string{}},
		Edges:   make(map[model.EdgeType]config.EdgeDef),
		Templates: config.TemplateConfig{
			Root: "templates",
		},
	}

	// Add types from selected bundles
	for _, bundleName := range selectedBundles {
		bundle, ok := typeBundles[bundleName]
		if !ok {
			return nil, fmt.Errorf("unknown bundle: %s", bundleName)
		}

		for typeName, typeDef := range bundle.Types {
			// Create default key pattern
			keyPattern, err := regexp.Compile(config.DefaultKeyPattern)
			if err != nil {
				return nil, fmt.Errorf("compiling default key pattern: %w", err)
			}

			cfg.Types[model.TypeName(typeName)] = config.TypeDef{
				Description:       typeDef.Description,
				DefaultState:      typeDef.DefaultState,
				RequiredFields:    []string{"id", "type", "key", "created", "updated", "title", "tags", "state"},
				RecommendedFields: []string{},
				KeyPattern:        keyPattern,
				KeyMaxLen:         config.DefaultKeyMaxLen,
			}
		}
	}

	// Add default edge type
	cfg.Edges[model.DefaultEdgeType] = config.EdgeDef{
		Description: "General relationship",
		AllowedFrom: []model.TypeName{},
		AllowedTo:   []model.TypeName{},
	}

	return cfg, nil
}

// marshalConfig marshals a config to YAML
func marshalConfig(cfg *config.Config) ([]byte, error) {
	// Build YAML structure
	configMap := map[string]any{
		"version": cfg.Version,
		"types":   make(map[string]any),
		"tags": map[string]any{
			"preferred": cfg.Tags.Preferred,
		},
		"edges": make(map[string]any),
		"templates": map[string]any{
			"root": cfg.Templates.Root,
		},
	}

	// Add types
	typesMap := configMap["types"].(map[string]any)
	for typeName, typeDef := range cfg.Types {
		typeMap := map[string]any{
			"description":    typeDef.Description,
			"default_state":  typeDef.DefaultState,
			"required_fields": typeDef.RequiredFields,
		}
		if typeDef.KeyPattern != nil {
			typeMap["key_pattern"] = typeDef.KeyPattern.String()
		}
		typeMap["key_max_len"] = typeDef.KeyMaxLen
		typesMap[string(typeName)] = typeMap
	}

	// Add edges
	edgesMap := configMap["edges"].(map[string]any)
	for edgeType, edgeDef := range cfg.Edges {
		edgesMap[string(edgeType)] = map[string]any{
			"description": edgeDef.Description,
		}
		if len(edgeDef.AllowedFrom) > 0 {
			allowedFrom := make([]string, len(edgeDef.AllowedFrom))
			for i, t := range edgeDef.AllowedFrom {
				allowedFrom[i] = string(t)
			}
			edgesMap[string(edgeType)].(map[string]any)["allowed_from"] = allowedFrom
		}
		if len(edgeDef.AllowedTo) > 0 {
			allowedTo := make([]string, len(edgeDef.AllowedTo))
			for i, t := range edgeDef.AllowedTo {
				allowedTo[i] = string(t)
			}
			edgesMap[string(edgeType)].(map[string]any)["allowed_to"] = allowedTo
		}
	}

	return yaml.Marshal(configMap)
}
