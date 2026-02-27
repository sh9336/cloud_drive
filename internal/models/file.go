// internal/models/file.go
package models

import (
	"time"

	"github.com/google/uuid"
)

type File struct {
	ID                uuid.UUID  `json:"id"`
	TenantID          uuid.UUID  `json:"tenant_id"`
	OriginalFilename  string     `json:"original_filename"`
	StoredFilename    string     `json:"stored_filename"`
	S3Key             string     `json:"s3_key"`
	FileSize          int64      `json:"file_size"`
	MimeType          string     `json:"mime_type"`
	UploadTo          string     `json:"upload_to"` // Template folder destination
	UploadStatus      string     `json:"upload_status"`
	UploadCompletedAt *time.Time `json:"upload_completed_at,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
	DeletedAt         *time.Time `json:"deleted_at,omitempty"`
}

type UploadURLRequest struct {
	Filename string `json:"filename" binding:"required"`
	FileSize int64  `json:"file_size" binding:"required,gt=0"`
	MimeType string `json:"mime_type" binding:"required"`
	UploadTo string `json:"upload_to" binding:"required"` // Must match template node path
}

type UploadURLResponse struct {
	FileID    uuid.UUID `json:"file_id"`
	UploadURL string    `json:"upload_url"`
	S3Key     string    `json:"s3_key"`
	ExpiresIn int       `json:"expires_in"` // seconds
}

type CompleteUploadRequest struct {
	FileID uuid.UUID `json:"file_id" binding:"required"`
}

type DownloadURLResponse struct {
	DownloadURL string `json:"download_url"`
	Filename    string `json:"filename"`
	ExpiresIn   int    `json:"expires_in"` // seconds
}

// Sync metadata models for device synchronization

type SyncFileMetadata struct {
	Path         string    `json:"path"`          // Original filename path (folder/filename.ext)
	Size         int64     `json:"size"`          // File size in bytes
	LastModified time.Time `json:"last_modified"` // Last modification time
	Hash         string    `json:"hash"`          // File hash for change detection (optional)
	MimeType     string    `json:"mime_type"`     // MIME type
	FileID       uuid.UUID `json:"file_id"`       // Internal file ID for download URL generation
}

type SyncMetadataResponse struct {
	Files      []SyncFileMetadata `json:"files"`
	LastSyncAt *time.Time         `json:"last_sync_at,omitempty"` // When device last synced
	TotalFiles int                `json:"total_files"`
	TotalSize  int64              `json:"total_size"`
}

type SyncDownloadRequest struct {
	Files []SyncFileRequest `json:"files" binding:"required"`
}

type SyncFileRequest struct {
	Path string `json:"path" binding:"required"` // Original filename path
}

type SyncDownloadResponse struct {
	Downloads []SyncDownloadInfo `json:"downloads"`
}

type SyncDownloadInfo struct {
	Path        string `json:"path"`         // Original filename path
	DownloadURL string `json:"download_url"` // Presigned download URL
	ExpiresIn   int    `json:"expires_in"`   // URL expiration in seconds
}
