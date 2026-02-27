// internal/models/audit_log.go
package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type AuditLog struct {
	ID           uuid.UUID       `json:"id"`
	ActorType    string          `json:"actor_type"` // 'admin', 'tenant', 'system'
	ActorID      *uuid.UUID      `json:"actor_id,omitempty"`
	ActorEmail   *string         `json:"actor_email,omitempty"`
	Action       string          `json:"action"`
	ResourceType *string         `json:"resource_type,omitempty"`
	ResourceID   *uuid.UUID      `json:"resource_id,omitempty"`
	Metadata     *json.RawMessage `json:"metadata,omitempty"`
	IPAddress    *string         `json:"ip_address,omitempty"`
	UserAgent    *string         `json:"user_agent,omitempty"`
	Status       string          `json:"status"` // 'success', 'failure'
	ErrorMessage *string         `json:"error_message,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
}

type CreateAuditLogRequest struct {
	ActorType    string
	ActorID      *uuid.UUID
	ActorEmail   *string
	Action       string
	ResourceType *string
	ResourceID   *uuid.UUID
	Metadata     map[string]interface{}
	IPAddress    *string
	UserAgent    *string
	Status       string
	ErrorMessage *string
}

type AuditLogFilter struct {
	ActorType  string
	ActorEmail string
	Action     string
	Status     string
	StartDate  *time.Time
	EndDate    *time.Time
	Page       int
	Limit      int
}

type PaginationResponse struct {
	CurrentPage  int   `json:"current_page"`
	PageSize     int   `json:"page_size"`
	TotalRecords int64 `json:"total_records"`
	TotalPages   int   `json:"total_pages"`
}

type AuditLogListResponse struct {
	Logs       []*AuditLog        `json:"logs"`
	Pagination PaginationResponse `json:"pagination"`
}
