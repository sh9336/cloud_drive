// internal/service/file_service.go
package service

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"backend/internal/models"
	"backend/internal/repository"
	"backend/internal/storage"
	"backend/internal/utils"
	"backend/internal/validator"

	"github.com/google/uuid"
)

type FileService struct {
	fileRepo      *repository.FileRepository
	tenantRepo    *repository.TenantRepository
	auditRepo     *repository.AuditRepository
	s3Service     *storage.S3Service
	template      *models.FileTreeTemplate
	maxFileSize   int64
	allowedTypes  []string
	presignExpiry time.Duration
}

func NewFileService(
	fileRepo *repository.FileRepository,
	tenantRepo *repository.TenantRepository,
	auditRepo *repository.AuditRepository,
	s3Service *storage.S3Service,
	template *models.FileTreeTemplate,
	maxFileSize int64,
	allowedTypes []string,
	presignExpiry time.Duration,
) *FileService {
	return &FileService{
		fileRepo:      fileRepo,
		tenantRepo:    tenantRepo,
		auditRepo:     auditRepo,
		s3Service:     s3Service,
		template:      template,
		maxFileSize:   maxFileSize,
		allowedTypes:  allowedTypes,
		presignExpiry: presignExpiry,
	}
}

func (s *FileService) GenerateUploadURL(ctx context.Context, tenantID uuid.UUID, req models.UploadURLRequest) (*models.UploadURLResponse, error) {
	// Validate file
	if err := validator.ValidateFileUpload(req.Filename, req.FileSize, req.MimeType, s.maxFileSize, s.allowedTypes); err != nil {
		return nil, err
	}

	// Validate upload destination against template
	isValid, errMsg := s.template.ValidateUploadDestination(req.UploadTo)
	if !isValid {
		return nil, fmt.Errorf("invalid upload destination: %s", errMsg)
	}

	// Get tenant
	_, err := s.tenantRepo.GetByID(ctx, tenantID)
	if err != nil {
		return nil, utils.ErrNotFound
	}

	// Generate unique filename
	ext := filepath.Ext(req.Filename)
	storedFilename := fmt.Sprintf("%s%s", uuid.New().String(), ext)

	// Resolve S3 path using template
	s3Key := s.template.ResolveS3Path(tenantID.String(), req.UploadTo, storedFilename)

	// Create file record
	file := &models.File{
		ID:               uuid.New(),
		TenantID:         tenantID,
		OriginalFilename: req.Filename,
		StoredFilename:   storedFilename,
		S3Key:            s3Key,
		FileSize:         req.FileSize,
		MimeType:         req.MimeType,
		UploadTo:         req.UploadTo,
		UploadStatus:     "pending",
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	if err := s.fileRepo.Create(ctx, file); err != nil {
		return nil, err
	}

	// Generate presigned URL
	uploadURL, err := s.s3Service.GeneratePresignedPutURL(ctx, s3Key, s.presignExpiry)
	if err != nil {
		return nil, err
	}

	_ = s.auditRepo.Create(ctx, models.CreateAuditLogRequest{
		ActorType:    "tenant",
		ActorID:      &tenantID,
		Action:       "upload_url_generated",
		ResourceType: utils.StringPtr("file"),
		ResourceID:   &file.ID,
		Metadata: map[string]interface{}{
			"filename":  req.Filename,
			"size":      req.FileSize,
			"upload_to": req.UploadTo,
		},
		Status: "success",
	})

	return &models.UploadURLResponse{
		FileID:    file.ID,
		UploadURL: uploadURL,
		S3Key:     s3Key,
		ExpiresIn: int(s.presignExpiry.Seconds()),
	}, nil
}

func (s *FileService) CompleteUpload(ctx context.Context, tenantID uuid.UUID, req models.CompleteUploadRequest) error {
	file, err := s.fileRepo.GetByID(ctx, req.FileID, tenantID)
	if err != nil {
		return utils.ErrNotFound
	}

	// Verify file exists in S3
	if err := s.s3Service.HeadObject(ctx, file.S3Key); err != nil {
		return fmt.Errorf("file not found in storage")
	}

	if err := s.fileRepo.MarkUploadComplete(ctx, req.FileID); err != nil {
		return err
	}

	_ = s.auditRepo.Create(ctx, models.CreateAuditLogRequest{
		ActorType:    "tenant",
		ActorID:      &tenantID,
		Action:       "upload_completed",
		ResourceType: utils.StringPtr("file"),
		ResourceID:   &file.ID,
		Metadata: map[string]interface{}{
			"filename": file.OriginalFilename,
		},
		Status: "success",
	})

	return nil
}

func (s *FileService) ListFiles(ctx context.Context, tenantID uuid.UUID) ([]*models.File, error) {
	return s.fileRepo.ListByTenant(ctx, tenantID)
}

func (s *FileService) GenerateDownloadURL(ctx context.Context, tenantID, fileID uuid.UUID) (*models.DownloadURLResponse, error) {
	file, err := s.fileRepo.GetByID(ctx, fileID, tenantID)
	if err != nil {
		return nil, utils.ErrNotFound
	}

	if file.UploadStatus != "completed" {
		return nil, fmt.Errorf("file upload not completed")
	}

	downloadURL, err := s.s3Service.GeneratePresignedGetURL(ctx, file.S3Key, s.presignExpiry)
	if err != nil {
		return nil, err
	}

	_ = s.auditRepo.Create(ctx, models.CreateAuditLogRequest{
		ActorType:    "tenant",
		ActorID:      &tenantID,
		Action:       "download_url_generated",
		ResourceType: utils.StringPtr("file"),
		ResourceID:   &file.ID,
		Metadata: map[string]interface{}{
			"filename": file.OriginalFilename,
		},
		Status: "success",
	})

	return &models.DownloadURLResponse{
		DownloadURL: downloadURL,
		Filename:    file.OriginalFilename,
		ExpiresIn:   int(s.presignExpiry.Seconds()),
	}, nil
}

func (s *FileService) DeleteFile(ctx context.Context, tenantID, fileID uuid.UUID) error {
	file, err := s.fileRepo.GetByID(ctx, fileID, tenantID)
	if err != nil {
		return utils.ErrNotFound
	}

	// Delete from S3
	if err := s.s3Service.DeleteObject(ctx, file.S3Key); err != nil {
		return err
	}

	// Soft delete from database
	if err := s.fileRepo.SoftDelete(ctx, fileID, tenantID); err != nil {
		return err
	}

	_ = s.auditRepo.Create(ctx, models.CreateAuditLogRequest{
		ActorType:    "tenant",
		ActorID:      &tenantID,
		Action:       "file_deleted",
		ResourceType: utils.StringPtr("file"),
		ResourceID:   &file.ID,
		Metadata: map[string]interface{}{
			"filename": file.OriginalFilename,
		},
		Status: "success",
	})

	return nil
}

// GetSyncMetadata returns file metadata for sync devices (optimized for high-volume requests)
func (s *FileService) GetSyncMetadata(ctx context.Context, tenantID uuid.UUID) (*models.SyncMetadataResponse, error) {
	files, err := s.fileRepo.ListByTenant(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	// Only include completed uploads for sync
	syncFiles := make([]models.SyncFileMetadata, 0, len(files))
	var totalSize int64

	for _, file := range files {
		if file.UploadStatus != "completed" {
			continue
		}

		// Convert S3Key back to logical path
		// S3Key format: tenants/{tenant_id}/{folder}/{stored_filename}
		// We need to map this back to the original logical path
		path := s.resolveLogicalPath(file)

		syncFile := models.SyncFileMetadata{
			Path:         path,
			Size:         file.FileSize,
			LastModified: file.UpdatedAt,
			Hash:         "", // TODO: Implement file hashing if needed
			MimeType:     file.MimeType,
			FileID:       file.ID,
		}

		syncFiles = append(syncFiles, syncFile)
		totalSize += file.FileSize
	}

	return &models.SyncMetadataResponse{
		Files:      syncFiles,
		TotalFiles: len(syncFiles),
		TotalSize:  totalSize,
	}, nil
}

// GenerateSyncDownloadURLs creates download URLs for multiple files (batch operation)
func (s *FileService) GenerateSyncDownloadURLs(ctx context.Context, tenantID uuid.UUID, req models.SyncDownloadRequest) (*models.SyncDownloadResponse, error) {
	downloads := make([]models.SyncDownloadInfo, 0, len(req.Files))

	for _, fileReq := range req.Files {
		// Find file by logical path
		file, err := s.fileRepo.FindByPath(ctx, tenantID, fileReq.Path)
		if err != nil {
			// Skip files not found - device will handle missing files
			continue
		}

		if file.UploadStatus != "completed" {
			// Skip incomplete uploads
			continue
		}

		// Generate presigned download URL
		downloadURL, err := s.s3Service.GeneratePresignedGetURL(ctx, file.S3Key, s.presignExpiry)
		if err != nil {
			// Log error but continue with other files
			continue
		}

		downloadInfo := models.SyncDownloadInfo{
			Path:        fileReq.Path,
			DownloadURL: downloadURL,
			ExpiresIn:   int(s.presignExpiry.Seconds()),
		}

		downloads = append(downloads, downloadInfo)
	}

	return &models.SyncDownloadResponse{
		Downloads: downloads,
	}, nil
}

// resolveLogicalPath converts S3Key back to logical path for sync
func (s *FileService) resolveLogicalPath(file *models.File) string {
	// Extract path from S3Key if upload_to is empty
	// S3Key format: tenants/{tenant_id}/{folder}/{stored_filename}
	if file.UploadTo == "" || file.UploadTo == "root" {
		// Parse S3Key to extract folder
		s3Key := file.S3Key
		// Remove tenant prefix
		tenantPrefix := "tenants/" + file.TenantID.String() + "/"
		if strings.HasPrefix(s3Key, tenantPrefix) {
			remainingPath := s3Key[len(tenantPrefix):]
			// Extract folder (everything before the stored filename)
			if lastSlash := strings.LastIndex(remainingPath, "/"); lastSlash != -1 {
				folder := remainingPath[:lastSlash]
				// Return folder/original_filename
				return folder + "/" + file.OriginalFilename
			}
		}
		// Fallback to just filename
		return file.OriginalFilename
	}
	return file.UploadTo + "/" + file.OriginalFilename
}

// DirectUpload handles proxying the file upload through the backend
func (s *FileService) DirectUpload(ctx context.Context, tenantID uuid.UUID, filename string, fileSize int64, mimeType string, uploadTo string, body []byte) (*models.File, error) {
	// Validate file
	if err := validator.ValidateFileUpload(filename, fileSize, mimeType, s.maxFileSize, s.allowedTypes); err != nil {
		return nil, err
	}

	// Validate upload destination
	isValid, errMsg := s.template.ValidateUploadDestination(uploadTo)
	if !isValid {
		return nil, fmt.Errorf("invalid upload destination: %s", errMsg)
	}

	// Generate unique filename
	ext := filepath.Ext(filename)
	storedFilename := fmt.Sprintf("%s%s", uuid.New().String(), ext)

	// Resolve S3 path
	s3Key := s.template.ResolveS3Path(tenantID.String(), uploadTo, storedFilename)

	// Create file record (pending)
	file := &models.File{
		ID:               uuid.New(),
		TenantID:         tenantID,
		OriginalFilename: filename,
		StoredFilename:   storedFilename,
		S3Key:            s3Key,
		FileSize:         fileSize,
		MimeType:         mimeType,
		UploadTo:         uploadTo,
		UploadStatus:     "pending",
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	if err := s.fileRepo.Create(ctx, file); err != nil {
		return nil, err
	}

	// Upload directly to S3 from backend
	if err := s.s3Service.PutObject(ctx, s3Key, body, mimeType); err != nil {
		return nil, fmt.Errorf("failed to upload to storage: %w", err)
	}

	// Mark as complete immediately
	if err := s.fileRepo.MarkUploadComplete(ctx, file.ID); err != nil {
		return nil, err
	}

	// Refresh file state to get completion timestamps
	file, _ = s.fileRepo.GetByID(ctx, file.ID, tenantID)

	_ = s.auditRepo.Create(ctx, models.CreateAuditLogRequest{
		ActorType:    "tenant",
		ActorID:      &tenantID,
		Action:       "proxy_upload_completed",
		ResourceType: utils.StringPtr("file"),
		ResourceID:   &file.ID,
		Metadata: map[string]interface{}{
			"filename": filename,
			"size":     fileSize,
		},
		Status: "success",
	})

	return file, nil
}

