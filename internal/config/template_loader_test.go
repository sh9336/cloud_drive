// internal/config/template_loader_test.go
package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadFileTreeTemplate(t *testing.T) {
	// This test assumes the template file exists at the project root
	// In CI/CD, we'd use a temporary directory with a test template

	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Write test template
	templateContent := `{
  "version": "1.0.0",
  "root_path": "tenants/{tenant_id}/",
  "root_permissions": {
    "allow_files": true,
    "allow_folders": false,
    "can_upload": true,
    "can_delete": true,
    "can_replace": true,
    "can_list": true
  },
  "nodes": [
    {
      "type": "folder",
      "path": "uploads",
      "is_required": true,
      "permissions": {
        "allow_files": true,
        "allow_folders": false,
        "can_upload": true,
        "can_delete": true,
        "can_replace": true,
        "can_list": true
      },
      "metadata": {
        "description": "User-uploaded files"
      }
    }
  ]
}`

	templatePath := filepath.Join(tempDir, "default_dir_tree_template.json")
	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		t.Fatalf("Failed to write test template: %v", err)
	}

	t.Run("valid template", func(t *testing.T) {
		template, err := LoadFileTreeTemplate(tempDir)
		if err != nil {
			t.Errorf("LoadFileTreeTemplate() unexpected error: %v", err)
		}

		if template == nil {
			t.Errorf("LoadFileTreeTemplate() returned nil")
		}

		if template.Version != "1.0.0" {
			t.Errorf("LoadFileTreeTemplate() got version %s, want 1.0.0", template.Version)
		}

		if len(template.Nodes) != 1 {
			t.Errorf("LoadFileTreeTemplate() got %d nodes, want 1", len(template.Nodes))
		}
	})

	t.Run("file not found", func(t *testing.T) {
		_, err := LoadFileTreeTemplate("/nonexistent/path")
		if err == nil {
			t.Errorf("LoadFileTreeTemplate() expected error, got nil")
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		invalidPath := filepath.Join(tempDir, "invalid.json")
		if err := os.WriteFile(invalidPath, []byte("invalid json"), 0644); err != nil {
			t.Fatalf("Failed to write invalid template: %v", err)
		}

		os.Rename(invalidPath, filepath.Join(tempDir, "default_dir_tree_template.json"))

		_, err := LoadFileTreeTemplate(tempDir)
		if err == nil {
			t.Errorf("LoadFileTreeTemplate() expected error for invalid JSON, got nil")
		}
	})
}

func TestLoadFileTreeTemplateDuplicatePaths(t *testing.T) {
	tempDir := t.TempDir()

	// Template with duplicate paths
	templateContent := `{
  "version": "1.0.0",
  "root_path": "tenants/{tenant_id}/",
  "root_permissions": {
    "allow_files": true
  },
  "nodes": [
    {
      "type": "folder",
      "path": "uploads",
      "permissions": { "allow_files": true, "can_upload": true }
    },
    {
      "type": "folder",
      "path": "uploads",
      "permissions": { "allow_files": true, "can_upload": true }
    }
  ]
}`

	templatePath := filepath.Join(tempDir, "default_dir_tree_template.json")
	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		t.Fatalf("Failed to write test template: %v", err)
	}

	_, err := LoadFileTreeTemplate(tempDir)
	if err == nil {
		t.Errorf("LoadFileTreeTemplate() expected error for duplicate paths, got nil")
	}
}
