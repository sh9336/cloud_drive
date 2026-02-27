// internal/config/template_loader.go
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"backend/internal/models"
)

// LoadFileTreeTemplate loads the JSON template from file
func LoadFileTreeTemplate(projectRoot string) (*models.FileTreeTemplate, error) {
	templatePath := filepath.Join(projectRoot, "default_dir_tree_template.json")

	data, err := os.ReadFile(templatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read template file: %w", err)
	}

	var template models.FileTreeTemplate
	if err := json.Unmarshal(data, &template); err != nil {
		return nil, fmt.Errorf("failed to parse template JSON: %w", err)
	}

	// Validate template
	if template.Version == "" {
		return nil, fmt.Errorf("template version is required")
	}

	if template.RootPath == "" {
		return nil, fmt.Errorf("template root_path is required")
	}

	if len(template.Nodes) == 0 {
		return nil, fmt.Errorf("template must have at least one node")
	}

	// Check for duplicate paths
	seenPaths := make(map[string]bool)
	for _, node := range template.Nodes {
		if node.Path == "" {
			return nil, fmt.Errorf("template node path cannot be empty")
		}
		if seenPaths[node.Path] {
			return nil, fmt.Errorf("duplicate node path in template: %s", node.Path)
		}
		seenPaths[node.Path] = true
	}

	// Build index for fast lookups
	template.BuildIndex()

	return &template, nil
}
