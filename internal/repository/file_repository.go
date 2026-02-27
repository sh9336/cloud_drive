// internal/repository/file_repository.go
package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"backend/internal/models"
	"github.com/google/uuid"
)

type FileRepository struct {
	db *sql.DB
}

func NewFileRepository(db *sql.DB) *FileRepository {
	return &FileRepository{db: db}
}

func (r *FileRepository) Create(ctx context.Context, file *models.File) error {
	query := `
		INSERT INTO files (id, tenant_id, original_filename, stored_filename, s3_key,
			file_size, mime_type, upload_to, upload_status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err := r.db.ExecContext(ctx, query,
		file.ID, file.TenantID, file.OriginalFilename, file.StoredFilename,
		file.S3Key, file.FileSize, file.MimeType, file.UploadTo, file.UploadStatus,
		file.CreatedAt, file.UpdatedAt,
	)
	return err
}

func (r *FileRepository) GetByID(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*models.File, error) {
	query := `
		SELECT id, tenant_id, original_filename, stored_filename, s3_key, file_size,
			mime_type, upload_to, upload_status, upload_completed_at, created_at, updated_at, deleted_at
		FROM files
		WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL
	`
	file := &models.File{}
	err := r.db.QueryRowContext(ctx, query, id, tenantID).Scan(
		&file.ID, &file.TenantID, &file.OriginalFilename, &file.StoredFilename,
		&file.S3Key, &file.FileSize, &file.MimeType, &file.UploadTo, &file.UploadStatus,
		&file.UploadCompletedAt, &file.CreatedAt, &file.UpdatedAt, &file.DeletedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("file not found")
	}
	return file, err
}

func (r *FileRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]*models.File, error) {
	query := `
		SELECT id, tenant_id, original_filename, stored_filename, s3_key, file_size,
			mime_type, upload_to, upload_status, upload_completed_at, created_at, updated_at
		FROM files
		WHERE tenant_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []*models.File
	for rows.Next() {
		file := &models.File{}
		err := rows.Scan(
			&file.ID, &file.TenantID, &file.OriginalFilename, &file.StoredFilename,
			&file.S3Key, &file.FileSize, &file.MimeType, &file.UploadTo, &file.UploadStatus,
			&file.UploadCompletedAt, &file.CreatedAt, &file.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		files = append(files, file)
	}
	return files, nil
}

func (r *FileRepository) MarkUploadComplete(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE files 
		SET upload_status = 'completed', upload_completed_at = $1
		WHERE id = $2
	`
	_, err := r.db.ExecContext(ctx, query, time.Now(), id)
	return err
}

func (r *FileRepository) SoftDelete(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) error {
	query := `UPDATE files SET deleted_at = $1 WHERE id = $2 AND tenant_id = $3`
	_, err := r.db.ExecContext(ctx, query, time.Now(), id, tenantID)
	return err
}

// FindByPath finds a file by its logical path (folder/filename.ext)
func (r *FileRepository) FindByPath(ctx context.Context, tenantID uuid.UUID, path string) (*models.File, error) {
	// Parse path to extract folder and filename
	// Path format: "folder/filename.ext" or just "filename.ext"
	var folder, filename string
	if lastSlash := strings.LastIndex(path, "/"); lastSlash != -1 {
		folder = path[:lastSlash]
		filename = path[lastSlash+1:]
	} else {
		filename = path
	}

	query := `
		SELECT id, tenant_id, original_filename, stored_filename, s3_key, file_size,
			mime_type, upload_to, upload_status, upload_completed_at, created_at, updated_at, deleted_at
		FROM files
		WHERE tenant_id = $1 AND original_filename = $2 AND deleted_at IS NULL
	`

	// Try exact match first with upload_to
	args := []interface{}{tenantID, filename}
	if folder != "" {
		query += " AND upload_to = $3"
		args = append(args, folder)
	}

	query += " LIMIT 1"

	file := &models.File{}
	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&file.ID, &file.TenantID, &file.OriginalFilename, &file.StoredFilename,
		&file.S3Key, &file.FileSize, &file.MimeType, &file.UploadTo, &file.UploadStatus,
		&file.UploadCompletedAt, &file.CreatedAt, &file.UpdatedAt, &file.DeletedAt,
	)

	if err == sql.ErrNoRows {
		// If not found with exact match, try matching by S3Key pattern
		// This handles cases where upload_to is empty but file is in a folder
		if folder != "" {
			s3Pattern := "%/" + folder + "/%"
			query2 := `
				SELECT id, tenant_id, original_filename, stored_filename, s3_key, file_size,
					mime_type, upload_to, upload_status, upload_completed_at, created_at, updated_at, deleted_at
				FROM files
				WHERE tenant_id = $1 AND original_filename = $2 AND s3_key LIKE $3 AND deleted_at IS NULL
				LIMIT 1
			`
			err = r.db.QueryRowContext(ctx, query2, tenantID, filename, s3Pattern).Scan(
				&file.ID, &file.TenantID, &file.OriginalFilename, &file.StoredFilename,
				&file.S3Key, &file.FileSize, &file.MimeType, &file.UploadTo, &file.UploadStatus,
				&file.UploadCompletedAt, &file.CreatedAt, &file.UpdatedAt, &file.DeletedAt,
			)
		}

		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("file not found")
		}
	}

	return file, err
}
