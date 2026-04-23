// internal/service/tenant_service.go
package service

import (
	"context"

	"encoding/json"
	"fmt"
	"time"

	"backend/internal/auth"
	"backend/internal/database"
	"backend/internal/models"
	"backend/internal/repository"
	"backend/internal/utils"
	"github.com/google/uuid"
)

type TenantService struct {
	tenantRepo  *repository.TenantRepository
	tokenRepo   *repository.TokenRepository
	auditRepo   *repository.AuditRepository
	redisClient *database.RedisClient
}

func NewTenantService(
	tenantRepo *repository.TenantRepository,
	tokenRepo *repository.TokenRepository,
	auditRepo *repository.AuditRepository,
	redisClient *database.RedisClient,
) *TenantService {
	return &TenantService{
		tenantRepo:  tenantRepo,
		tokenRepo:   tokenRepo,
		auditRepo:   auditRepo,
		redisClient: redisClient,
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

	if s.redisClient != nil {
		_ = s.redisClient.Del(ctx, fmt.Sprintf("tenant_profile:%s", tenantID.String())).Err()
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
	cacheKey := fmt.Sprintf("tenant_profile:%s", tenantID.String())

	if s.redisClient != nil {
		cachedData, err := s.redisClient.Get(ctx, cacheKey).Result()
		if err == nil {
			var tenant models.Tenant
			if err := json.Unmarshal([]byte(cachedData), &tenant); err == nil {
				return &tenant, nil
			}
		}
	}

	tenant, err := s.tenantRepo.GetByID(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	if s.redisClient != nil {
		if tenantData, err := json.Marshal(tenant); err == nil {
			_ = s.redisClient.Set(ctx, cacheKey, tenantData, 15*time.Minute).Err()
		}
	}

	return tenant, nil
}
