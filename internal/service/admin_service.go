// internal/service/admin_service.go
package service

import (
	"context"
	"fmt"
	"time"

	"backend/internal/auth"
	"backend/internal/database"
	"backend/internal/models"
	"backend/internal/repository"
	"backend/internal/utils"

	"github.com/google/uuid"
)

type AdminService struct {
	adminRepo  *repository.AdminRepository
	tenantRepo  *repository.TenantRepository
	tokenRepo   *repository.TokenRepository
	auditRepo   *repository.AuditRepository
	redisClient *database.RedisClient
}

func NewAdminService(
	adminRepo *repository.AdminRepository,
	tenantRepo *repository.TenantRepository,
	tokenRepo *repository.TokenRepository,
	auditRepo *repository.AuditRepository,
	redisClient *database.RedisClient,
) *AdminService {
	return &AdminService{
		adminRepo:   adminRepo,
		tenantRepo:  tenantRepo,
		tokenRepo:   tokenRepo,
		auditRepo:   auditRepo,
		redisClient: redisClient,
	}
}

func (s *AdminService) CreateTenant(ctx context.Context, req models.CreateTenantRequest, adminID uuid.UUID) (*models.Tenant, string, error) {
	existing, _ := s.tenantRepo.GetByEmail(ctx, req.Email)
	if existing != nil {
		return nil, "", utils.ErrEmailAlreadyExists
	}

	tempPassword, err := auth.GenerateTemporaryPassword()
	if err != nil {
		return nil, "", err
	}

	passwordHash, err := auth.HashPassword(tempPassword)
	if err != nil {
		return nil, "", err
	}

	tenant := &models.Tenant{
		ID:                 uuid.New(),
		Email:              req.Email,
		PasswordHash:       passwordHash,
		FullName:           req.FullName,
		CompanyName:        req.CompanyName,
		PasswordChangedAt:  time.Now(),
		MustChangePassword: true,
		IsActive:           true,
		S3Prefix:           fmt.Sprintf("tenants/%s/", uuid.New().String()),
		CreatedBy:          adminID,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	if err := s.tenantRepo.Create(ctx, tenant); err != nil {
		return nil, "", err
	}

	_ = s.auditRepo.Create(ctx, models.CreateAuditLogRequest{
		ActorType:    "admin",
		ActorID:      &adminID,
		Action:       "tenant_created",
		ResourceType: utils.StringPtr("tenant"),
		ResourceID:   &tenant.ID,
		Metadata: map[string]interface{}{
			"email": tenant.Email,
			"name":  tenant.FullName,
		},
		Status: "success",
	})

	return tenant, tempPassword, nil
}

func (s *AdminService) ListTenants(ctx context.Context) ([]*models.Tenant, error) {
	return s.tenantRepo.List(ctx)
}

func (s *AdminService) GetTenant(ctx context.Context, id uuid.UUID) (*models.Tenant, error) {
	return s.tenantRepo.GetByID(ctx, id)
}

func (s *AdminService) ResetTenantPassword(ctx context.Context, tenantID, adminID uuid.UUID) (string, error) {
	tenant, err := s.tenantRepo.GetByID(ctx, tenantID)
	if err != nil {
		return "", utils.ErrNotFound
	}

	tempPassword, err := auth.GenerateTemporaryPassword()
	if err != nil {
		return "", err
	}

	passwordHash, err := auth.HashPassword(tempPassword)
	if err != nil {
		return "", err
	}

	if err := s.tenantRepo.UpdatePassword(ctx, tenantID, passwordHash); err != nil {
		return "", err
	}

	if s.redisClient != nil {
		_ = s.redisClient.Del(ctx, fmt.Sprintf("tenant_profile:%s", tenantID.String())).Err()
	}

	_ = s.tokenRepo.RevokeAllForUser(ctx, "tenant", tenantID, "password_reset_by_admin")

	_ = s.auditRepo.Create(ctx, models.CreateAuditLogRequest{
		ActorType:    "admin",
		ActorID:      &adminID,
		Action:       "tenant_password_reset",
		ResourceType: utils.StringPtr("tenant"),
		ResourceID:   &tenantID,
		Metadata: map[string]interface{}{
			"tenant_email": tenant.Email,
		},
		Status: "success",
	})

	return tempPassword, nil
}

func (s *AdminService) UpdateTenantStatus(ctx context.Context, tenantID, adminID uuid.UUID, req models.UpdateTenantStatusRequest) error {
	tenant, err := s.tenantRepo.GetByID(ctx, tenantID)
	if err != nil {
		return utils.ErrNotFound
	}

	if req.IsActive == nil {
		return utils.ErrBadRequest
	}

	if err := s.tenantRepo.UpdateStatus(ctx, tenantID, *req.IsActive, req.DisabledReason); err != nil {
		return err
	}

	if s.redisClient != nil {
		_ = s.redisClient.Del(ctx, fmt.Sprintf("tenant_profile:%s", tenantID.String())).Err()
	}

	if !*req.IsActive {
		_ = s.tokenRepo.RevokeAllForUser(ctx, "tenant", tenantID, "account_disabled_by_admin")
	}

	action := "tenant_enabled"
	if !*req.IsActive {
		action = "tenant_disabled"
	}

	_ = s.auditRepo.Create(ctx, models.CreateAuditLogRequest{
		ActorType:    "admin",
		ActorID:      &adminID,
		Action:       action,
		ResourceType: utils.StringPtr("tenant"),
		ResourceID:   &tenantID,
		Metadata: map[string]interface{}{
			"tenant_email": tenant.Email,
			"reason":       req.DisabledReason,
		},
		Status: "success",
	})

	return nil
}

func (s *AdminService) ChangePassword(ctx context.Context, adminID uuid.UUID, req models.ChangePasswordRequest) error {
	admin, err := s.adminRepo.GetByID(ctx, adminID)
	if err != nil {
		return utils.ErrNotFound
	}

	// Verify current password
	if err := auth.CheckPassword(req.CurrentPassword, admin.PasswordHash); err != nil {
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
	if err := s.adminRepo.UpdatePassword(ctx, adminID, newPasswordHash); err != nil {
		return err
	}

	// Revoke all existing tokens
	_ = s.tokenRepo.RevokeAllForUser(ctx, "admin", adminID, "password_changed")

	// Audit log
	_ = s.auditRepo.Create(ctx, models.CreateAuditLogRequest{
		ActorType:    "admin",
		ActorID:      &adminID,
		Action:       "password_changed",
		ResourceType: utils.StringPtr("admin"),
		ResourceID:   &adminID,
		Status:       "success",
	})

	return nil
}

func (s *AdminService) GetAuditLogs(ctx context.Context, filter models.AuditLogFilter) (*models.AuditLogListResponse, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.Limit <= 0 {
		filter.Limit = 20
	}

	logs, total, err := s.auditRepo.List(ctx, filter)
	if err != nil {
		return nil, err
	}

	totalPages := int(total / int64(filter.Limit))
	if total%int64(filter.Limit) > 0 {
		totalPages++
	}

	return &models.AuditLogListResponse{
		Logs: logs,
		Pagination: models.PaginationResponse{
			CurrentPage:  filter.Page,
			PageSize:     filter.Limit,
			TotalRecords: total,
			TotalPages:   totalPages,
		},
	}, nil
}

func (s *AdminService) RotateAuditLogs(ctx context.Context, days int, adminID uuid.UUID) (int64, error) {
	if days <= 0 {
		days = 30 // Default 30 days
	}

	count, err := s.auditRepo.Cleanup(ctx, days)
	if err != nil {
		return 0, err
	}

	_ = s.auditRepo.Create(ctx, models.CreateAuditLogRequest{
		ActorType: "admin",
		ActorID:   &adminID,
		Action:    "audit_logs_rotated",
		Metadata: map[string]interface{}{
			"retention_days": days,
			"deleted_count":  count,
		},
		Status: "success",
	})

	return count, nil
}

func (s *AdminService) DeleteTenant(ctx context.Context, tenantID, adminID uuid.UUID) error {
	tenant, err := s.tenantRepo.GetByID(ctx, tenantID)
	if err != nil {
		return utils.ErrNotFound
	}

	// 1. Delete all refresh tokens (they don't have FK cascade)
	_ = s.tokenRepo.DeleteAllForUser(ctx, "tenant", tenantID)

	// 2. Delete tenant (this will cascade delete files and sync tokens in DB)
	if err := s.tenantRepo.Delete(ctx, tenantID); err != nil {
		return err
	}

	if s.redisClient != nil {
		_ = s.redisClient.Del(ctx, fmt.Sprintf("tenant_profile:%s", tenantID.String())).Err()
	}

	// 3. Audit log
	_ = s.auditRepo.Create(ctx, models.CreateAuditLogRequest{
		ActorType:    "admin",
		ActorID:      &adminID,
		Action:       "tenant_deleted",
		ResourceType: utils.StringPtr("tenant"),
		ResourceID:   &tenantID,
		Metadata: map[string]interface{}{
			"tenant_email": tenant.Email,
			"tenant_name":  tenant.FullName,
		},
		Status: "success",
	})

	return nil
}

