// internal/repository/token_repository.go
package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"backend/internal/models"
	"github.com/google/uuid"
)

type TokenRepository struct {
	db *sql.DB
}

func NewTokenRepository(db *sql.DB) *TokenRepository {
	return &TokenRepository{db: db}
}

func (r *TokenRepository) Create(ctx context.Context, token *models.RefreshToken) error {
	query := `
		INSERT INTO refresh_tokens (id, token_hash, user_type, user_id, ip_address,
			user_agent, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.db.ExecContext(ctx, query,
		token.ID, token.TokenHash, token.UserType, token.UserID,
		token.IPAddress, token.UserAgent, token.ExpiresAt, token.CreatedAt,
	)
	return err
}

func (r *TokenRepository) GetByHash(ctx context.Context, tokenHash string) (*models.RefreshToken, error) {
	query := `
		SELECT id, token_hash, user_type, user_id, ip_address, user_agent,
			expires_at, revoked_at, revoked_reason, created_at
		FROM refresh_tokens
		WHERE token_hash = $1
	`
	token := &models.RefreshToken{}
	err := r.db.QueryRowContext(ctx, query, tokenHash).Scan(
		&token.ID, &token.TokenHash, &token.UserType, &token.UserID,
		&token.IPAddress, &token.UserAgent, &token.ExpiresAt,
		&token.RevokedAt, &token.RevokedReason, &token.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("token not found")
	}
	return token, err
}

func (r *TokenRepository) Revoke(ctx context.Context, tokenHash string, reason string) error {
	query := `UPDATE refresh_tokens SET revoked_at = $1, revoked_reason = $2 WHERE token_hash = $3`
	_, err := r.db.ExecContext(ctx, query, time.Now(), reason, tokenHash)
	return err
}

func (r *TokenRepository) RevokeAllForUser(ctx context.Context, userType string, userID uuid.UUID, reason string) error {
	query := `
		UPDATE refresh_tokens 
		SET revoked_at = $1, revoked_reason = $2
		WHERE user_type = $3 AND user_id = $4 AND revoked_at IS NULL
	`
	_, err := r.db.ExecContext(ctx, query, time.Now(), reason, userType, userID)
	return err
}

func (r *TokenRepository) DeleteExpired(ctx context.Context) error {
	query := `DELETE FROM refresh_tokens WHERE expires_at < $1`
	_, err := r.db.ExecContext(ctx, query, time.Now())
	return err
}

func (r *TokenRepository) DeleteAllForUser(ctx context.Context, userType string, userID uuid.UUID) error {
	query := `DELETE FROM refresh_tokens WHERE user_type = $1 AND user_id = $2`
	_, err := r.db.ExecContext(ctx, query, userType, userID)
	return err
}
