// internal/models/tenant.go
package models

import (
	"time"

	"github.com/google/uuid"
)

type Tenant struct {
	ID                 uuid.UUID  `json:"id"`
	Email              string     `json:"email"`
	PasswordHash       string     `json:"-"` // Never expose in JSON
	FullName           string     `json:"full_name"`
	CompanyName        *string    `json:"company_name,omitempty"`
	PasswordChangedAt  time.Time  `json:"password_changed_at"`
	PasswordExpiresAt  *time.Time `json:"password_expires_at,omitempty"`
	MustChangePassword bool       `json:"must_change_password"`
	IsActive           bool       `json:"is_active"`
	DisabledAt         *time.Time `json:"disabled_at,omitempty"`
	DisabledReason     *string    `json:"disabled_reason,omitempty"`
	S3Prefix           string     `json:"s3_prefix"`
	CreatedBy          uuid.UUID  `json:"created_by"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
	LastLoginAt           *time.Time `json:"last_login_at,omitempty"`
	LastLoginIP           *string    `json:"last_login_ip,omitempty"`
	TemplateInitializedAt *time.Time `json:"template_initialized_at,omitempty"`
}

type CreateTenantRequest struct {
	Email       string  `json:"email" binding:"required,email"`
	FullName    string  `json:"full_name" binding:"required"`
	CompanyName *string `json:"company_name"`
}

type TenantLoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type UpdateTenantStatusRequest struct {
	// Use pointer so we can distinguish between "missing" and "false".
	IsActive       *bool   `json:"is_active" binding:"required"`
	DisabledReason *string `json:"disabled_reason"`
}
