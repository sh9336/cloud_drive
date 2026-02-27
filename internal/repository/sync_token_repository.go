// internal/repository/sync_token_repository.go
package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"backend/internal/models"

	"github.com/google/uuid"
)

type SyncTokenRepository struct {
	db *sql.DB
}

func NewSyncTokenRepository(db *sql.DB) *SyncTokenRepository {
	return &SyncTokenRepository{db: db}
}

func (r *SyncTokenRepository) Create(ctx context.Context, token *models.SyncToken) error {
	query := `
		INSERT INTO sync_tokens (
			id, tenant_id, token_hash, name, can_read, can_write, can_delete,
			expires_at, is_active, created_by, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`
	_, err := r.db.ExecContext(ctx, query,
		token.ID, token.TenantID, token.TokenHash, token.Name,
		token.CanRead, token.CanWrite, token.CanDelete,
		token.ExpiresAt, token.IsActive, token.CreatedBy,
		token.CreatedAt, token.UpdatedAt,
	)
	return err
}

func (r *SyncTokenRepository) GetByHash(ctx context.Context, tokenHash string) (*models.SyncToken, error) {
	query := `
		SELECT id, tenant_id, token_hash, name, can_read, can_write, can_delete,
			expires_at, last_used_at, is_active, revoked_at, revoked_by, revoked_reason,
			created_by, created_at, updated_at, total_requests, 
			total_bytes_uploaded, total_bytes_downloaded
		FROM sync_tokens
		WHERE token_hash = $1
	`
	token := &models.SyncToken{}
	err := r.db.QueryRowContext(ctx, query, tokenHash).Scan(
		&token.ID, &token.TenantID, &token.TokenHash, &token.Name,
		&token.CanRead, &token.CanWrite, &token.CanDelete,
		&token.ExpiresAt, &token.LastUsedAt, &token.IsActive,
		&token.RevokedAt, &token.RevokedBy, &token.RevokedReason,
		&token.CreatedBy, &token.CreatedAt, &token.UpdatedAt,
		&token.TotalRequests, &token.TotalBytesUploaded, &token.TotalBytesDownloaded,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("sync token not found")
	}
	return token, err
}

func (r *SyncTokenRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.SyncToken, error) {
	query := `
		SELECT id, tenant_id, token_hash, name, can_read, can_write, can_delete,
			expires_at, last_used_at, is_active, revoked_at, revoked_by, revoked_reason,
			created_by, created_at, updated_at, total_requests,
			total_bytes_uploaded, total_bytes_downloaded
		FROM sync_tokens
		WHERE id = $1
	`
	token := &models.SyncToken{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&token.ID, &token.TenantID, &token.TokenHash, &token.Name,
		&token.CanRead, &token.CanWrite, &token.CanDelete,
		&token.ExpiresAt, &token.LastUsedAt, &token.IsActive,
		&token.RevokedAt, &token.RevokedBy, &token.RevokedReason,
		&token.CreatedBy, &token.CreatedAt, &token.UpdatedAt,
		&token.TotalRequests, &token.TotalBytesUploaded, &token.TotalBytesDownloaded,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("sync token not found")
	}
	return token, err
}

func (r *SyncTokenRepository) ListAll(ctx context.Context) ([]*models.SyncTokenWithTenant, error) {
	query := `
		SELECT 
			st.id, st.tenant_id, st.name, st.can_read, st.can_write, st.can_delete,
			st.expires_at, st.last_used_at, st.is_active, st.revoked_at,
			st.created_by, st.created_at, st.updated_at,
			st.total_requests, st.total_bytes_uploaded, st.total_bytes_downloaded,
			t.email, t.full_name, t.company_name
		FROM sync_tokens st
		JOIN tenants t ON st.tenant_id = t.id
		ORDER BY st.created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tokens []*models.SyncTokenWithTenant
	for rows.Next() {
		token := &models.SyncTokenWithTenant{}
		err := rows.Scan(
			&token.ID, &token.TenantID, &token.Name,
			&token.CanRead, &token.CanWrite, &token.CanDelete,
			&token.ExpiresAt, &token.LastUsedAt, &token.IsActive, &token.RevokedAt,
			&token.CreatedBy, &token.CreatedAt, &token.UpdatedAt,
			&token.TotalRequests, &token.TotalBytesUploaded, &token.TotalBytesDownloaded,
			&token.TenantEmail, &token.TenantName, &token.CompanyName,
		)
		if err != nil {
			return nil, err
		}
		tokens = append(tokens, token)
	}
	return tokens, nil
}

func (r *SyncTokenRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]*models.SyncToken, error) {
	query := `
		SELECT id, tenant_id, name, can_read, can_write, can_delete,
			expires_at, last_used_at, is_active, revoked_at, revoked_by, revoked_reason,
			created_by, created_at, updated_at, total_requests,
			total_bytes_uploaded, total_bytes_downloaded
		FROM sync_tokens
		WHERE tenant_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tokens []*models.SyncToken
	for rows.Next() {
		token := &models.SyncToken{}
		err := rows.Scan(
			&token.ID, &token.TenantID, &token.Name,
			&token.CanRead, &token.CanWrite, &token.CanDelete,
			&token.ExpiresAt, &token.LastUsedAt, &token.IsActive,
			&token.RevokedAt, &token.RevokedBy, &token.RevokedReason,
			&token.CreatedBy, &token.CreatedAt, &token.UpdatedAt,
			&token.TotalRequests, &token.TotalBytesUploaded, &token.TotalBytesDownloaded,
		)
		if err != nil {
			return nil, err
		}
		tokens = append(tokens, token)
	}
	return tokens, nil
}

func (r *SyncTokenRepository) UpdateLastUsed(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE sync_tokens 
		SET last_used_at = $1, total_requests = total_requests + 1
		WHERE id = $2
	`
	_, err := r.db.ExecContext(ctx, query, time.Now(), id)
	return err
}

func (r *SyncTokenRepository) UpdateStats(ctx context.Context, id uuid.UUID, bytesUploaded, bytesDownloaded int64) error {
	query := `
		UPDATE sync_tokens 
		SET total_bytes_uploaded = total_bytes_uploaded + $1,
		    total_bytes_downloaded = total_bytes_downloaded + $2
		WHERE id = $3
	`
	_, err := r.db.ExecContext(ctx, query, bytesUploaded, bytesDownloaded, id)
	return err
}

func (r *SyncTokenRepository) Revoke(ctx context.Context, id, revokedBy uuid.UUID, reason string) error {
	query := `
		UPDATE sync_tokens 
		SET is_active = false, revoked_at = $1, revoked_by = $2, revoked_reason = $3
		WHERE id = $4
	`
	_, err := r.db.ExecContext(ctx, query, time.Now(), revokedBy, reason, id)
	return err
}

func (r *SyncTokenRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM sync_tokens WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *SyncTokenRepository) DeleteExpired(ctx context.Context) error {
	query := `DELETE FROM sync_tokens WHERE expires_at < $1`
	_, err := r.db.ExecContext(ctx, query, time.Now())
	return err
}

func (r *SyncTokenRepository) DeleteRevoked(ctx context.Context) (int64, error) {
	query := `DELETE FROM sync_tokens WHERE is_active = false OR revoked_at IS NOT NULL`
	result, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
