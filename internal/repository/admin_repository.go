// internal/repository/admin_repository.go
package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"backend/internal/models"
	"github.com/google/uuid"
)

type AdminRepository struct {
	db *sql.DB
}

func NewAdminRepository(db *sql.DB) *AdminRepository {
	return &AdminRepository{db: db}
}

func (r *AdminRepository) Create(ctx context.Context, admin *models.Admin) error {
	query := `
		INSERT INTO admins (id, email, password_hash, full_name, password_changed_at, 
			password_expires_at, must_change_password, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err := r.db.ExecContext(ctx, query,
		admin.ID, admin.Email, admin.PasswordHash, admin.FullName,
		admin.PasswordChangedAt, admin.PasswordExpiresAt, admin.MustChangePassword,
		admin.IsActive, admin.CreatedAt, admin.UpdatedAt,
	)
	return err
}

func (r *AdminRepository) GetByEmail(ctx context.Context, email string) (*models.Admin, error) {
	query := `
		SELECT id, email, password_hash, full_name, password_changed_at, password_expires_at,
			must_change_password, is_active, disabled_at, disabled_reason,
			created_at, updated_at, last_login_at, last_login_ip
		FROM admins
		WHERE email = $1
	`
	admin := &models.Admin{}
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&admin.ID, &admin.Email, &admin.PasswordHash, &admin.FullName,
		&admin.PasswordChangedAt, &admin.PasswordExpiresAt, &admin.MustChangePassword,
		&admin.IsActive, &admin.DisabledAt, &admin.DisabledReason,
		&admin.CreatedAt, &admin.UpdatedAt, &admin.LastLoginAt, &admin.LastLoginIP,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("admin not found")
	}
	return admin, err
}

func (r *AdminRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Admin, error) {
	query := `
		SELECT id, email, password_hash, full_name, password_changed_at, password_expires_at,
			must_change_password, is_active, disabled_at, disabled_reason,
			created_at, updated_at, last_login_at, last_login_ip
		FROM admins
		WHERE id = $1
	`
	admin := &models.Admin{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&admin.ID, &admin.Email, &admin.PasswordHash, &admin.FullName,
		&admin.PasswordChangedAt, &admin.PasswordExpiresAt, &admin.MustChangePassword,
		&admin.IsActive, &admin.DisabledAt, &admin.DisabledReason,
		&admin.CreatedAt, &admin.UpdatedAt, &admin.LastLoginAt, &admin.LastLoginIP,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("admin not found")
	}
	return admin, err
}

func (r *AdminRepository) List(ctx context.Context) ([]*models.Admin, error) {
	query := `
		SELECT id, email, full_name, is_active, disabled_at, disabled_reason,
			created_at, updated_at, last_login_at
		FROM admins
		ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var admins []*models.Admin
	for rows.Next() {
		admin := &models.Admin{}
		err := rows.Scan(
			&admin.ID, &admin.Email, &admin.FullName, &admin.IsActive,
			&admin.DisabledAt, &admin.DisabledReason, &admin.CreatedAt,
			&admin.UpdatedAt, &admin.LastLoginAt,
		)
		if err != nil {
			return nil, err
		}
		admins = append(admins, admin)
	}
	return admins, nil
}

func (r *AdminRepository) UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error {
	query := `
		UPDATE admins 
		SET password_hash = $1, password_changed_at = $2, must_change_password = false
		WHERE id = $3
	`
	_, err := r.db.ExecContext(ctx, query, passwordHash, time.Now(), id)
	return err
}

func (r *AdminRepository) UpdateLastLogin(ctx context.Context, id uuid.UUID, ip string) error {
	query := `UPDATE admins SET last_login_at = $1, last_login_ip = $2 WHERE id = $3`
	_, err := r.db.ExecContext(ctx, query, time.Now(), ip, id)
	return err
}

func (r *AdminRepository) UpdateStatus(ctx context.Context, id uuid.UUID, isActive bool, reason *string) error {
	var query string
	if isActive {
		query = `UPDATE admins SET is_active = true, disabled_at = NULL, disabled_reason = NULL WHERE id = $1`
		_, err := r.db.ExecContext(ctx, query, id)
		return err
	}
	query = `UPDATE admins SET is_active = false, disabled_at = $1, disabled_reason = $2 WHERE id = $3`
	_, err := r.db.ExecContext(ctx, query, time.Now(), reason, id)
	return err
}

func (r *AdminRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM admins WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}
