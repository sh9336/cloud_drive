// internal/service/sync_token_service.go
package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"backend/internal/models"
	"backend/internal/repository"
	"backend/internal/utils"

	"github.com/google/uuid"
)

type SyncTokenService struct {
	syncTokenRepo *repository.SyncTokenRepository
	tenantRepo    *repository.TenantRepository
	auditRepo     *repository.AuditRepository
}

func NewSyncTokenService(
	syncTokenRepo *repository.SyncTokenRepository,
	tenantRepo *repository.TenantRepository,
	auditRepo *repository.AuditRepository,
) *SyncTokenService {
	return &SyncTokenService{
		syncTokenRepo: syncTokenRepo,
		tenantRepo:    tenantRepo,
		auditRepo:     auditRepo,
	}
}

// GenerateSyncToken creates a new sync token
func (s *SyncTokenService) GenerateSyncToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// HashToken hashes a token for storage
func (s *SyncTokenService) HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// CreateSyncToken creates a new sync token for a tenant
func (s *SyncTokenService) CreateSyncToken(ctx context.Context, req models.CreateSyncTokenRequest, adminID uuid.UUID) (*models.CreateSyncTokenResponse, error) {
	// Verify tenant exists
	tenant, err := s.tenantRepo.GetByID(ctx, req.TenantID)
	if err != nil {
		return nil, utils.ErrNotFound
	}

	// Generate random token
	randomPart, err := s.GenerateSyncToken(32)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// Format: sync_<tenant_id>_<random_32_chars>
	plainToken := fmt.Sprintf("sync_%s_%s", tenant.ID.String(), randomPart)
	tokenHash := s.HashToken(plainToken)

	// Create sync token
	syncToken := &models.SyncToken{
		ID:        uuid.New(),
		TenantID:  req.TenantID,
		TokenHash: tokenHash,
		Name:      req.Name,
		CanRead:   req.CanRead,
		CanWrite:  req.CanWrite,
		CanDelete: req.CanDelete,
		ExpiresAt: time.Now().AddDate(0, 0, req.ExpiresInDays),
		IsActive:  true,
		CreatedBy: adminID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.syncTokenRepo.Create(ctx, syncToken); err != nil {
		return nil, err
	}

	// Audit log
	_ = s.auditRepo.Create(ctx, models.CreateAuditLogRequest{
		ActorType:    "admin",
		ActorID:      &adminID,
		Action:       "sync_token_created",
		ResourceType: utils.StringPtr("sync_token"),
		ResourceID:   &syncToken.ID,
		Metadata: map[string]interface{}{
			"token_name":      req.Name,
			"tenant_id":       req.TenantID.String(),
			"tenant_email":    tenant.Email,
			"expires_in_days": req.ExpiresInDays,
			"permissions": map[string]bool{
				"can_read":   req.CanRead,
				"can_write":  req.CanWrite,
				"can_delete": req.CanDelete,
			},
		},
		Status: "success",
	})

	return &models.CreateSyncTokenResponse{
		Token:     plainToken,
		TokenInfo: syncToken,
	}, nil
}

// ValidateSyncToken validates and returns sync token claims
func (s *SyncTokenService) ValidateSyncToken(ctx context.Context, token string) (*models.SyncTokenClaims, error) {
	tokenHash := s.HashToken(token)

	syncToken, err := s.syncTokenRepo.GetByHash(ctx, tokenHash)
	if err != nil {
		return nil, utils.ErrInvalidSyncToken
	}

	// Check if active
	if !syncToken.IsActive {
		return nil, utils.ErrSyncTokenRevoked
	}

	// Check if revoked
	if syncToken.RevokedAt != nil {
		return nil, utils.ErrSyncTokenRevoked
	}

	// Check if expired
	if time.Now().After(syncToken.ExpiresAt) {
		return nil, utils.ErrSyncTokenExpired
	}

	// Update last used
	_ = s.syncTokenRepo.UpdateLastUsed(ctx, syncToken.ID)

	return &models.SyncTokenClaims{
		TokenID:   syncToken.ID,
		TenantID:  syncToken.TenantID,
		CanRead:   syncToken.CanRead,
		CanWrite:  syncToken.CanWrite,
		CanDelete: syncToken.CanDelete,
	}, nil
}

// ListAllSyncTokens lists all sync tokens (admin only)
func (s *SyncTokenService) ListAllSyncTokens(ctx context.Context) ([]*models.SyncTokenWithTenant, error) {
	return s.syncTokenRepo.ListAll(ctx)
}

// ListTenantSyncTokens lists sync tokens for a specific tenant
func (s *SyncTokenService) ListTenantSyncTokens(ctx context.Context, tenantID uuid.UUID) ([]*models.SyncToken, error) {
	return s.syncTokenRepo.ListByTenant(ctx, tenantID)
}

// GetSyncToken gets a sync token by ID
func (s *SyncTokenService) GetSyncToken(ctx context.Context, id uuid.UUID) (*models.SyncToken, error) {
	token, err := s.syncTokenRepo.GetByID(ctx, id)
	if err != nil {
		return nil, utils.ErrNotFound
	}
	return token, nil
}

// RotateSyncToken generates a new token and optionally keeps old one active for grace period
func (s *SyncTokenService) RotateSyncToken(ctx context.Context, tokenID, adminID uuid.UUID, req models.RotateSyncTokenRequest) (*models.RotateSyncTokenResponse, error) {
	// Get old token
	oldToken, err := s.syncTokenRepo.GetByID(ctx, tokenID)
	if err != nil {
		return nil, utils.ErrNotFound
	}

	// Generate new token
	randomPart, err := s.GenerateSyncToken(32)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	plainToken := fmt.Sprintf("sync_%s_%s", oldToken.TenantID.String(), randomPart)
	tokenHash := s.HashToken(plainToken)

	// Create new sync token
	newToken := &models.SyncToken{
		ID:        uuid.New(),
		TenantID:  oldToken.TenantID,
		TokenHash: tokenHash,
		Name:      oldToken.Name + " (Rotated)",
		CanRead:   oldToken.CanRead,
		CanWrite:  oldToken.CanWrite,
		CanDelete: oldToken.CanDelete,
		ExpiresAt: time.Now().AddDate(0, 0, req.ExpiresInDays),
		IsActive:  true,
		CreatedBy: adminID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.syncTokenRepo.Create(ctx, newToken); err != nil {
		return nil, err
	}

	// Calculate grace period
	gracePeriodDays := req.GracePeriodDays
	if gracePeriodDays == 0 {
		gracePeriodDays = 7 // Default 7 days
	}
	graceEndsAt := time.Now().AddDate(0, 0, gracePeriodDays)

	// Schedule old token revocation after grace period
	// In production, you'd use a job queue or cron for this
	// For now, we'll just leave it active with a note

	// Audit log
	_ = s.auditRepo.Create(ctx, models.CreateAuditLogRequest{
		ActorType:    "admin",
		ActorID:      &adminID,
		Action:       "sync_token_rotated",
		ResourceType: utils.StringPtr("sync_token"),
		ResourceID:   &newToken.ID,
		Metadata: map[string]interface{}{
			"old_token_id":      oldToken.ID.String(),
			"new_token_id":      newToken.ID.String(),
			"grace_period_days": gracePeriodDays,
			"grace_ends_at":     graceEndsAt,
		},
		Status: "success",
	})

	return &models.RotateSyncTokenResponse{
		OldTokenID:      oldToken.ID,
		NewToken:        plainToken,
		NewTokenID:      newToken.ID,
		ExpiresAt:       newToken.ExpiresAt,
		GracePeriodDays: gracePeriodDays,
		GraceEndsAt:     graceEndsAt,
	}, nil
}

// RevokeSyncToken revokes a sync token
func (s *SyncTokenService) RevokeSyncToken(ctx context.Context, tokenID, adminID uuid.UUID, reason string) error {
	token, err := s.syncTokenRepo.GetByID(ctx, tokenID)
	if err != nil {
		return utils.ErrNotFound
	}

	if err := s.syncTokenRepo.Revoke(ctx, tokenID, adminID, reason); err != nil {
		return err
	}

	// Audit log
	_ = s.auditRepo.Create(ctx, models.CreateAuditLogRequest{
		ActorType:    "admin",
		ActorID:      &adminID,
		Action:       "sync_token_revoked",
		ResourceType: utils.StringPtr("sync_token"),
		ResourceID:   &tokenID,
		Metadata: map[string]interface{}{
			"token_name": token.Name,
			"tenant_id":  token.TenantID.String(),
			"reason":     reason,
		},
		Status: "success",
	})

	return nil
}

// GetSyncTokenStats returns statistics for a sync token
func (s *SyncTokenService) GetSyncTokenStats(ctx context.Context, tokenID uuid.UUID) (*models.SyncTokenStats, error) {
	token, err := s.syncTokenRepo.GetByID(ctx, tokenID)
	if err != nil {
		return nil, utils.ErrNotFound
	}

	daysUntilExpiry := int(time.Until(token.ExpiresAt).Hours() / 24)
	if daysUntilExpiry < 0 {
		daysUntilExpiry = 0
	}

	stats := &models.SyncTokenStats{
		TokenID:              token.ID,
		TokenName:            token.Name,
		TotalRequests:        token.TotalRequests,
		TotalBytesUploaded:   token.TotalBytesUploaded,
		TotalBytesDownloaded: token.TotalBytesDownloaded,
		LastUsedAt:           token.LastUsedAt,
		CreatedAt:            token.CreatedAt,
		ExpiresAt:            token.ExpiresAt,
		DaysUntilExpiry:      daysUntilExpiry,
		IsActive:             token.IsActive,
	}

	return stats, nil
}

// UpdateStats updates usage statistics for a sync token
func (s *SyncTokenService) UpdateStats(ctx context.Context, tokenID uuid.UUID, bytesUploaded, bytesDownloaded int64) error {
	return s.syncTokenRepo.UpdateStats(ctx, tokenID, bytesUploaded, bytesDownloaded)
}

func (s *SyncTokenService) CleanupRevokedTokens(ctx context.Context, adminID uuid.UUID) (int64, error) {
	count, err := s.syncTokenRepo.DeleteRevoked(ctx)
	if err != nil {
		return 0, err
	}

	if count > 0 {
		_ = s.auditRepo.Create(ctx, models.CreateAuditLogRequest{
			ActorType: "admin",
			ActorID:   &adminID,
			Action:    "sync_tokens_cleanup",
			Metadata: map[string]interface{}{
				"deleted_count": count,
				"type":          "revoked",
			},
			Status: "success",
		})
	}

	return count, nil
}

func (s *SyncTokenService) DeleteSyncToken(ctx context.Context, tokenID, adminID uuid.UUID) error {
	// Verify token exists
	token, err := s.syncTokenRepo.GetByID(ctx, tokenID)
	if err != nil {
		return utils.ErrNotFound
	}

	err = s.syncTokenRepo.Delete(ctx, tokenID)
	if err != nil {
		return err
	}

	// Audit log
	_ = s.auditRepo.Create(ctx, models.CreateAuditLogRequest{
		ActorType:    "admin",
		ActorID:      &adminID,
		Action:       "sync_token_deleted",
		ResourceType: utils.StringPtr("sync_token"),
		ResourceID:   &tokenID,
		Metadata: map[string]interface{}{
			"token_name": token.Name,
			"tenant_id":  token.TenantID,
		},
		Status: "success",
	})

	return nil
}
