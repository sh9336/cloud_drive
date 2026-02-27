// internal/models/template.go
package models

import "strings"

// FileTreeTemplate represents the JSON structure of the file tree template
type FileTreeTemplate struct {
	Version         string                   `json:"version"`
	RootPath        string                   `json:"root_path"`
	RootPermissions TemplatePermissions      `json:"root_permissions"`
	Nodes           []TemplateNode           `json:"nodes"`
	nodePathIndex   map[string]*TemplateNode // Internal index for fast lookup
}

// TemplateNode represents a folder or file node in the template
type TemplateNode struct {
	Type        string                 `json:"type"` // "folder" or "file"
	Path        string                 `json:"path"` // Logical path name
	IsRequired  bool                   `json:"is_required"`
	Permissions TemplatePermissions    `json:"permissions"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// TemplatePermissions defines what operations are allowed
type TemplatePermissions struct {
	AllowFiles   bool `json:"allow_files"`
	AllowFolders bool `json:"allow_folders"`
	CanUpload    bool `json:"can_upload"`
	CanDelete    bool `json:"can_delete"`
	CanReplace   bool `json:"can_replace"`
	CanList      bool `json:"can_list"`
}

// BuildIndex creates a fast lookup index for nodes by path
func (t *FileTreeTemplate) BuildIndex() {
	t.nodePathIndex = make(map[string]*TemplateNode)
	for i := range t.Nodes {
		t.nodePathIndex[t.Nodes[i].Path] = &t.Nodes[i]
	}
}

// GetNode returns a node by path name
func (t *FileTreeTemplate) GetNode(path string) *TemplateNode {
	if t.nodePathIndex == nil {
		t.BuildIndex()
	}
	return t.nodePathIndex[path]
}

// ValidateUploadDestination checks if upload_to is allowed
func (t *FileTreeTemplate) ValidateUploadDestination(uploadTo string) (bool, string) {
	if uploadTo == "" {
		return false, "upload_to is required"
	}

	node := t.GetNode(uploadTo)
	if node == nil {
		return false, "invalid upload destination: " + uploadTo
	}

	if node.Type != "folder" {
		return false, "upload destination must be a folder"
	}

	if !node.Permissions.CanUpload {
		return false, "uploads not allowed in: " + uploadTo
	}

	if !node.Permissions.AllowFiles {
		return false, "files not allowed in: " + uploadTo
	}

	return true, ""
}

// ResolveS3Path constructs the full S3 path for a file
func (t *FileTreeTemplate) ResolveS3Path(tenantID, uploadTo, storedFilename string) string {
	// Replace {tenant_id} in root_path
	rootPath := strings.ReplaceAll(t.RootPath, "{tenant_id}", tenantID)
	// If uploadTo is "root", store directly in tenant root without subfolder
	if uploadTo == "root" {
		return rootPath + storedFilename
	}
	return rootPath + uploadTo + "/" + storedFilename
}
