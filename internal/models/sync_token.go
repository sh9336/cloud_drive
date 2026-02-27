// internal/models/sync_token.go
package models

import (
	"time"

	"github.com/google/uuid"
)

// SyncToken represents a sync token in the database
type SyncToken struct {
	ID                   uuid.UUID  `json:"id"`
	TenantID             uuid.UUID  `json:"tenant_id"`
	TokenHash            string     `json:"-"` // Never expose
	Name                 string     `json:"name"`
	CanRead              bool       `json:"can_read"`
	CanWrite             bool       `json:"can_write"`
	CanDelete            bool       `json:"can_delete"`
	ExpiresAt            time.Time  `json:"expires_at"`
	LastUsedAt           *time.Time `json:"last_used_at,omitempty"`
	IsActive             bool       `json:"is_active"`
	RevokedAt            *time.Time `json:"revoked_at,omitempty"`
	RevokedBy            *uuid.UUID `json:"revoked_by,omitempty"`
	RevokedReason        *string    `json:"revoked_reason,omitempty"`
	CreatedBy            uuid.UUID  `json:"created_by"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
	TotalRequests        int64      `json:"total_requests"`
	TotalBytesUploaded   int64      `json:"total_bytes_uploaded"`
	TotalBytesDownloaded int64      `json:"total_bytes_downloaded"`
}

// SyncTokenWithTenant includes tenant information with the sync token
type SyncTokenWithTenant struct {
	ID                   uuid.UUID  `json:"id"`
	TenantID             uuid.UUID  `json:"tenant_id"`
	Name                 string     `json:"name"`
	CanRead              bool       `json:"can_read"`
	CanWrite             bool       `json:"can_write"`
	CanDelete            bool       `json:"can_delete"`
	ExpiresAt            time.Time  `json:"expires_at"`
	LastUsedAt           *time.Time `json:"last_used_at,omitempty"`
	IsActive             bool       `json:"is_active"`
	RevokedAt            *time.Time `json:"revoked_at,omitempty"`
	CreatedBy            uuid.UUID  `json:"created_by"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
	TotalRequests        int64      `json:"total_requests"`
	TotalBytesUploaded   int64      `json:"total_bytes_uploaded"`
	TotalBytesDownloaded int64      `json:"total_bytes_downloaded"`
	TenantEmail          string     `json:"tenant_email"`
	TenantName           string     `json:"tenant_name"`
	CompanyName          string     `json:"company_name"`
}

// SyncTokenClaims represents the validated claims from a sync token
type SyncTokenClaims struct {
	TokenID   uuid.UUID `json:"token_id"`
	TenantID  uuid.UUID `json:"tenant_id"`
	CanRead   bool      `json:"can_read"`
	CanWrite  bool      `json:"can_write"`
	CanDelete bool      `json:"can_delete"`
}

// SyncTokenStats represents usage statistics for a sync token
type SyncTokenStats struct {
	TokenID              uuid.UUID  `json:"token_id"`
	TokenName            string     `json:"token_name"`
	TotalRequests        int64      `json:"total_requests"`
	TotalBytesUploaded   int64      `json:"total_bytes_uploaded"`
	TotalBytesDownloaded int64      `json:"total_bytes_downloaded"`
	LastUsedAt           *time.Time `json:"last_used_at,omitempty"`
	CreatedAt            time.Time  `json:"created_at"`
	ExpiresAt            time.Time  `json:"expires_at"`
	DaysUntilExpiry      int        `json:"days_until_expiry"`
	IsActive             bool       `json:"is_active"`
}

// Request models

// CreateSyncTokenRequest is the request to create a new sync token
type CreateSyncTokenRequest struct {
	TenantID      uuid.UUID `json:"tenant_id" binding:"required"`
	Name          string    `json:"name" binding:"required"`
	CanRead       bool      `json:"can_read"`
	CanWrite      bool      `json:"can_write"`
	CanDelete     bool      `json:"can_delete"`
	ExpiresInDays int       `json:"expires_in_days" binding:"required,min=1,max=365"`
}

// RotateSyncTokenRequest is the request to rotate a sync token
type RotateSyncTokenRequest struct {
	TokenID         uuid.UUID `json:"token_id" binding:"required"`
	ExpiresInDays   int       `json:"expires_in_days" binding:"required,min=1,max=365"`
	GracePeriodDays int       `json:"grace_period_days" binding:"min=0,max=30"` // Optional, defaults to 7
}

// RevokeSyncTokenRequest is the request to revoke a sync token
type RevokeSyncTokenRequest struct {
	Reason string `json:"reason" binding:"required"`
}

// Response models

// CreateSyncTokenResponse is the response when creating a sync token
type CreateSyncTokenResponse struct {
	Token     string     `json:"token"` // Plain token (only shown once!)
	TokenInfo *SyncToken `json:"token_info"`
}

// RotateSyncTokenResponse is the response when rotating a sync token
type RotateSyncTokenResponse struct {
	OldTokenID      uuid.UUID `json:"old_token_id"`
	NewToken        string    `json:"new_token"` // Plain token (only shown once!)
	NewTokenID      uuid.UUID `json:"new_token_id"`
	ExpiresAt       time.Time `json:"expires_at"`
	GracePeriodDays int       `json:"grace_period_days"`
	GraceEndsAt     time.Time `json:"grace_ends_at"`
}
