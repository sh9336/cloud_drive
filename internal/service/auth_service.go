// internal/service/auth_service.go
package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"backend/internal/auth"
	"backend/internal/models"
	"backend/internal/repository"
	"backend/internal/utils"

	"github.com/google/uuid"
)

type AuthService struct {
	adminRepo  *repository.AdminRepository
	tenantRepo *repository.TenantRepository
	tokenRepo  *repository.TokenRepository
	auditRepo  *repository.AuditRepository
	jwtService *auth.JWTService
}

func NewAuthService(
	adminRepo *repository.AdminRepository,
	tenantRepo *repository.TenantRepository,
	tokenRepo *repository.TokenRepository,
	auditRepo *repository.AuditRepository,
	jwtService *auth.JWTService,
) *AuthService {
	return &AuthService{
		adminRepo:  adminRepo,
		tenantRepo: tenantRepo,
		tokenRepo:  tokenRepo,
		auditRepo:  auditRepo,
		jwtService: jwtService,
	}
}

func (s *AuthService) AdminLogin(ctx context.Context, req models.AdminLoginRequest, ip, userAgent string) (*models.LoginResponse, error) {
	admin, err := s.adminRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		s.logFailedLogin(ctx, "admin", req.Email, ip, userAgent, "invalid_credentials")
		return nil, utils.ErrInvalidCredentials
	}

	if !admin.IsActive {
		s.logFailedLogin(ctx, "admin", req.Email, ip, userAgent, "account_disabled")
		return nil, utils.ErrAccountDisabled
	}

	if err := auth.CheckPassword(req.Password, admin.PasswordHash); err != nil {
		s.logFailedLogin(ctx, "admin", req.Email, ip, userAgent, "invalid_password")
		return nil, utils.ErrInvalidCredentials
	}

	if admin.PasswordExpiresAt != nil && time.Now().After(*admin.PasswordExpiresAt) {
		return nil, utils.ErrPasswordExpired
	}

	accessToken, err := s.jwtService.GenerateAccessTokenWithFlag(admin.ID, admin.Email, "admin", admin.MustChangePassword)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.jwtService.GenerateRefreshToken(admin.ID, admin.Email, "admin")
	if err != nil {
		return nil, err
	}

	if err := s.storeRefreshToken(ctx, refreshToken, "admin", admin.ID, ip, userAgent); err != nil {
		return nil, err
	}

	_ = s.adminRepo.UpdateLastLogin(ctx, admin.ID, ip)

	s.logSuccessfulLogin(ctx, "admin", admin.ID, admin.Email, ip, userAgent)

	return &models.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int(s.jwtService.GetAccessExpiry().Seconds()),
		User: models.UserInfo{
			ID:                 admin.ID,
			Email:              admin.Email,
			FullName:           admin.FullName,
			UserType:           "admin",
			MustChangePassword: admin.MustChangePassword,
		},
	}, nil
}

func (s *AuthService) TenantLogin(ctx context.Context, req models.TenantLoginRequest, ip, userAgent string) (*models.LoginResponse, error) {
	tenant, err := s.tenantRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		s.logFailedLogin(ctx, "tenant", req.Email, ip, userAgent, "invalid_credentials")
		return nil, utils.ErrInvalidCredentials
	}

	if !tenant.IsActive {
		s.logFailedLogin(ctx, "tenant", req.Email, ip, userAgent, "account_disabled")
		return nil, utils.ErrAccountDisabled
	}

	if err := auth.CheckPassword(req.Password, tenant.PasswordHash); err != nil {
		s.logFailedLogin(ctx, "tenant", req.Email, ip, userAgent, "invalid_password")
		return nil, utils.ErrInvalidCredentials
	}

	if tenant.PasswordExpiresAt != nil && time.Now().After(*tenant.PasswordExpiresAt) {
		return nil, utils.ErrPasswordExpired
	}

	accessToken, err := s.jwtService.GenerateAccessTokenWithFlag(tenant.ID, tenant.Email, "tenant", tenant.MustChangePassword)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.jwtService.GenerateRefreshToken(tenant.ID, tenant.Email, "tenant")
	if err != nil {
		return nil, err
	}

	if err := s.storeRefreshToken(ctx, refreshToken, "tenant", tenant.ID, ip, userAgent); err != nil {
		return nil, err
	}

	_ = s.tenantRepo.UpdateLastLogin(ctx, tenant.ID, ip)

	s.logSuccessfulLogin(ctx, "tenant", tenant.ID, tenant.Email, ip, userAgent)

	return &models.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int(s.jwtService.GetAccessExpiry().Seconds()),
		User: models.UserInfo{
			ID:                 tenant.ID,
			Email:              tenant.Email,
			FullName:           tenant.FullName,
			UserType:           "tenant",
			MustChangePassword: tenant.MustChangePassword,
		},
	}, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*models.RefreshTokenResponse, error) {
	claims, err := s.jwtService.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, utils.ErrInvalidToken
	}

	tokenHash := hashToken(refreshToken)
	storedToken, err := s.tokenRepo.GetByHash(ctx, tokenHash)
	if err != nil {
		return nil, utils.ErrInvalidToken
	}

	if storedToken.RevokedAt != nil {
		return nil, utils.ErrTokenRevoked
	}

	if time.Now().After(storedToken.ExpiresAt) {
		return nil, utils.ErrTokenExpired
	}

	// Fetch current user to get latest must_change_password flag
	var mustChangePassword bool

	if claims.UserType == "tenant" {
		tenant, err := s.tenantRepo.GetByID(ctx, claims.UserID)
		if err != nil {
			return nil, utils.ErrNotFound
		}
		mustChangePassword = tenant.MustChangePassword
	} else if claims.UserType == "admin" {
		admin, err := s.adminRepo.GetByID(ctx, claims.UserID)
		if err != nil {
			return nil, utils.ErrNotFound
		}
		mustChangePassword = admin.MustChangePassword
	}

	// Generate token with current state
	accessToken, err := s.jwtService.GenerateAccessTokenWithFlag(
		claims.UserID, claims.Email, claims.UserType, mustChangePassword,
	)
	if err != nil {
		return nil, err
	}

	return &models.RefreshTokenResponse{
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ExpiresIn:   int(s.jwtService.GetAccessExpiry().Seconds()),
	}, nil
}

func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	tokenHash := hashToken(refreshToken)
	return s.tokenRepo.Revoke(ctx, tokenHash, "user_logout")
}

func (s *AuthService) storeRefreshToken(ctx context.Context, token, userType string, userID uuid.UUID, ip, userAgent string) error {
	tokenHash := hashToken(token)
	refreshToken := &models.RefreshToken{
		ID:        uuid.New(),
		TokenHash: tokenHash,
		UserType:  userType,
		UserID:    userID,
		IPAddress: &ip,
		UserAgent: &userAgent,
		ExpiresAt: time.Now().Add(s.jwtService.GetRefreshExpiry()),
		CreatedAt: time.Now(),
	}
	return s.tokenRepo.Create(ctx, refreshToken)
}

func (s *AuthService) logSuccessfulLogin(ctx context.Context, userType string, userID uuid.UUID, email, ip, userAgent string) {
	_ = s.auditRepo.Create(ctx, models.CreateAuditLogRequest{
		ActorType:  userType,
		ActorID:    &userID,
		ActorEmail: &email,
		Action:     "login_success",
		IPAddress:  &ip,
		UserAgent:  &userAgent,
		Status:     "success",
	})
}

func (s *AuthService) logFailedLogin(ctx context.Context, userType, email, ip, userAgent, reason string) {
	_ = s.auditRepo.Create(ctx, models.CreateAuditLogRequest{
		ActorType:    userType,
		ActorEmail:   &email,
		Action:       "login_failed",
		IPAddress:    &ip,
		UserAgent:    &userAgent,
		Status:       "failure",
		ErrorMessage: &reason,
	})
}

func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
