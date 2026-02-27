// internal/models/template_test.go
package models

import (
	"testing"
)

func TestFileTreeTemplateValidateUploadDestination(t *testing.T) {
	template := &FileTreeTemplate{
		Version:  "1.0.0",
		RootPath: "tenants/{tenant_id}/",
		Nodes: []TemplateNode{
			{
				Type:       "folder",
				Path:       "uploads",
				IsRequired: true,
				Permissions: TemplatePermissions{
					AllowFiles:   true,
					AllowFolders: false,
					CanUpload:    true,
					CanDelete:    true,
					CanReplace:   true,
					CanList:      true,
				},
			},
			{
				Type:       "folder",
				Path:       "assets",
				IsRequired: true,
				Permissions: TemplatePermissions{
					AllowFiles:   true,
					AllowFolders: false,
					CanUpload:    false, // Uploads disabled
					CanDelete:    true,
					CanReplace:   true,
					CanList:      true,
				},
			},
		},
	}
	template.BuildIndex()

	tests := []struct {
		name      string
		uploadTo  string
		wantValid bool
		wantError string
	}{
		{
			name:      "valid upload destination",
			uploadTo:  "uploads",
			wantValid: true,
		},
		{
			name:      "uploads disabled",
			uploadTo:  "assets",
			wantValid: false,
			wantError: "uploads not allowed",
		},
		{
			name:      "invalid destination",
			uploadTo:  "invalid",
			wantValid: false,
			wantError: "invalid upload destination",
		},
		{
			name:      "empty upload_to",
			uploadTo:  "",
			wantValid: false,
			wantError: "upload_to is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, msg := template.ValidateUploadDestination(tt.uploadTo)
			if valid != tt.wantValid {
				t.Errorf("ValidateUploadDestination() got %v, want %v", valid, tt.wantValid)
			}
			if !tt.wantValid && msg == "" {
				t.Errorf("ValidateUploadDestination() should return error message")
			}
		})
	}
}

func TestFileTreeTemplateResolveS3Path(t *testing.T) {
	template := &FileTreeTemplate{
		Version:  "1.0.0",
		RootPath: "tenants/{tenant_id}/",
		Nodes: []TemplateNode{
			{
				Type: "folder",
				Path: "uploads",
				Permissions: TemplatePermissions{
					AllowFiles: true,
					CanUpload:  true,
				},
			},
		},
	}

	tenantID := "abc-123-def"
	uploadTo := "uploads"
	filename := "abc123def.pdf"

	expected := "tenants/abc-123-def/uploads/abc123def.pdf"
	result := template.ResolveS3Path(tenantID, uploadTo, filename)

	if result != expected {
		t.Errorf("ResolveS3Path() got %s, want %s", result, expected)
	}
}

func TestFileTreeTemplateGetNode(t *testing.T) {
	template := &FileTreeTemplate{
		Version:  "1.0.0",
		RootPath: "tenants/{tenant_id}/",
		Nodes: []TemplateNode{
			{
				Type: "folder",
				Path: "uploads",
			},
			{
				Type: "folder",
				Path: "assets",
			},
		},
	}
	template.BuildIndex()

	tests := []struct {
		name     string
		path     string
		wantNode bool
	}{
		{
			name:     "existing node",
			path:     "uploads",
			wantNode: true,
		},
		{
			name:     "another existing node",
			path:     "assets",
			wantNode: true,
		},
		{
			name:     "non-existing node",
			path:     "invalid",
			wantNode: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := template.GetNode(tt.path)
			if (node != nil) != tt.wantNode {
				t.Errorf("GetNode() got node=%v, wantNode=%v", node != nil, tt.wantNode)
			}
		})
	}
}
