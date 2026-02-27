// internal/service/tenant_service.go
package service

import (
	"context"

	"backend/internal/auth"
	"backend/internal/models"
	"backend/internal/repository"
	"backend/internal/utils"
	"github.com/google/uuid"
)

type TenantService struct {
	tenantRepo *repository.TenantRepository
	tokenRepo  *repository.TokenRepository
	auditRepo  *repository.AuditRepository
}

func NewTenantService(
	tenantRepo *repository.TenantRepository,
	tokenRepo *repository.TokenRepository,
	auditRepo *repository.AuditRepository,
) *TenantService {
	return &TenantService{
		tenantRepo: tenantRepo,
		tokenRepo:  tokenRepo,
		auditRepo:  auditRepo,
	}
}

func (s *TenantService) ChangePassword(ctx context.Context, tenantID uuid.UUID, req models.ChangePasswordRequest) error {
	tenant, err := s.tenantRepo.GetByID(ctx, tenantID)
	if err != nil {
		return utils.ErrNotFound
	}

	// Verify current password
	if err := auth.CheckPassword(req.CurrentPassword, tenant.PasswordHash); err != nil {
		return utils.ErrInvalidCredentials
	}

	// Validate new password strength
	if err := auth.ValidatePasswordStrength(req.NewPassword); err != nil {
		return err
	}

	// Hash new password
	newPasswordHash, err := auth.HashPassword(req.NewPassword)
	if err != nil {
		return err
	}

	// Update password
	if err := s.tenantRepo.UpdatePassword(ctx, tenantID, newPasswordHash); err != nil {
		return err
	}

	// Revoke all existing tokens
	_ = s.tokenRepo.RevokeAllForUser(ctx, "tenant", tenantID, "password_changed")

	// Audit log
	_ = s.auditRepo.Create(ctx, models.CreateAuditLogRequest{
		ActorType:    "tenant",
		ActorID:      &tenantID,
		Action:       "password_changed",
		ResourceType: utils.StringPtr("tenant"),
		ResourceID:   &tenantID,
		Status:       "success",
	})

	return nil
}

func (s *TenantService) GetProfile(ctx context.Context, tenantID uuid.UUID) (*models.Tenant, error) {
	return s.tenantRepo.GetByID(ctx, tenantID)
}
