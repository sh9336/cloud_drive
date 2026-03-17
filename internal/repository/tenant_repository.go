// internal/repository/tenant_repository.go
package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"backend/internal/models"
	"github.com/google/uuid"
)

type TenantRepository struct {
	db *sql.DB
}

func NewTenantRepository(db *sql.DB) *TenantRepository {
	return &TenantRepository{db: db}
}

func (r *TenantRepository) Create(ctx context.Context, tenant *models.Tenant) error {
	query := `
		INSERT INTO tenants (id, email, password_hash, full_name, company_name, 
			password_changed_at, password_expires_at, must_change_password, 
			is_active, s3_prefix, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`
	_, err := r.db.ExecContext(ctx, query,
		tenant.ID, tenant.Email, tenant.PasswordHash, tenant.FullName,
		tenant.CompanyName, tenant.PasswordChangedAt, tenant.PasswordExpiresAt,
		tenant.MustChangePassword, tenant.IsActive, tenant.S3Prefix,
		tenant.CreatedBy, tenant.CreatedAt, tenant.UpdatedAt,
	)
	return err
}

func (r *TenantRepository) GetByEmail(ctx context.Context, email string) (*models.Tenant, error) {
	query := `
		SELECT id, email, password_hash, full_name, company_name, password_changed_at,
			password_expires_at, must_change_password, is_active, disabled_at,
			disabled_reason, s3_prefix, created_by, created_at, updated_at,
			last_login_at, last_login_ip
		FROM tenants
		WHERE email = $1
	`
	tenant := &models.Tenant{}
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&tenant.ID, &tenant.Email, &tenant.PasswordHash, &tenant.FullName,
		&tenant.CompanyName, &tenant.PasswordChangedAt, &tenant.PasswordExpiresAt,
		&tenant.MustChangePassword, &tenant.IsActive, &tenant.DisabledAt,
		&tenant.DisabledReason, &tenant.S3Prefix, &tenant.CreatedBy,
		&tenant.CreatedAt, &tenant.UpdatedAt, &tenant.LastLoginAt, &tenant.LastLoginIP,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("tenant not found")
	}
	return tenant, err
}

func (r *TenantRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Tenant, error) {
	query := `
		SELECT id, email, password_hash, full_name, company_name, password_changed_at,
			password_expires_at, must_change_password, is_active, disabled_at,
			disabled_reason, s3_prefix, created_by, created_at, updated_at,
			last_login_at, last_login_ip
		FROM tenants
		WHERE id = $1
	`
	tenant := &models.Tenant{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&tenant.ID, &tenant.Email, &tenant.PasswordHash, &tenant.FullName,
		&tenant.CompanyName, &tenant.PasswordChangedAt, &tenant.PasswordExpiresAt,
		&tenant.MustChangePassword, &tenant.IsActive, &tenant.DisabledAt,
		&tenant.DisabledReason, &tenant.S3Prefix, &tenant.CreatedBy,
		&tenant.CreatedAt, &tenant.UpdatedAt, &tenant.LastLoginAt, &tenant.LastLoginIP,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("tenant not found")
	}
	return tenant, err
}

func (r *TenantRepository) List(ctx context.Context) ([]*models.Tenant, error) {
	query := `
		SELECT id, email, full_name, company_name, is_active, s3_prefix,
			created_at, last_login_at
		FROM tenants
		ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tenants []*models.Tenant
	for rows.Next() {
		tenant := &models.Tenant{}
		err := rows.Scan(
			&tenant.ID, &tenant.Email, &tenant.FullName, &tenant.CompanyName,
			&tenant.IsActive, &tenant.S3Prefix, &tenant.CreatedAt, &tenant.LastLoginAt,
		)
		if err != nil {
			return nil, err
		}
		tenants = append(tenants, tenant)
	}
	return tenants, nil
}

func (r *TenantRepository) UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error {
	query := `
		UPDATE tenants 
		SET password_hash = $1, password_changed_at = $2, must_change_password = false
		WHERE id = $3
	`
	_, err := r.db.ExecContext(ctx, query, passwordHash, time.Now(), id)
	return err
}

func (r *TenantRepository) UpdateLastLogin(ctx context.Context, id uuid.UUID, ip string) error {
	query := `UPDATE tenants SET last_login_at = $1, last_login_ip = $2 WHERE id = $3`
	_, err := r.db.ExecContext(ctx, query, time.Now(), ip, id)
	return err
}

func (r *TenantRepository) UpdateStatus(ctx context.Context, id uuid.UUID, isActive bool, reason *string) error {
	var query string
	if isActive {
		query = `UPDATE tenants SET is_active = true, disabled_at = NULL, disabled_reason = NULL WHERE id = $1`
		_, err := r.db.ExecContext(ctx, query, id)
		return err
	}
	query = `UPDATE tenants SET is_active = false, disabled_at = $1, disabled_reason = $2 WHERE id = $3`
	_, err := r.db.ExecContext(ctx, query, time.Now(), reason, id)
	return err
}

func (r *TenantRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM tenants WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}
