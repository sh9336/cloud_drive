// internal/models/refresh_token.go
package models

import (
	"time"

	"github.com/google/uuid"
)

type RefreshToken struct {
	ID            uuid.UUID  `json:"id"`
	TokenHash     string     `json:"-"`         // Never expose
	UserType      string     `json:"user_type"` // 'admin' or 'tenant'
	UserID        uuid.UUID  `json:"user_id"`
	IPAddress     *string    `json:"ip_address,omitempty"`
	UserAgent     *string    `json:"user_agent,omitempty"`
	ExpiresAt     time.Time  `json:"expires_at"`
	RevokedAt     *time.Time `json:"revoked_at,omitempty"`
	RevokedReason *string    `json:"revoked_reason,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}
