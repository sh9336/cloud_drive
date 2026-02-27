// internal/repository/audit_repository.go
package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"backend/internal/models"
)

type AuditRepository struct {
	db *sql.DB
}

func NewAuditRepository(db *sql.DB) *AuditRepository {
	return &AuditRepository{db: db}
}

func (r *AuditRepository) Create(ctx context.Context, req models.CreateAuditLogRequest) error {
	var metadata interface{}
	if req.Metadata != nil {
		m, err := json.Marshal(req.Metadata)
		if err != nil {
			return err
		}
		metadata = m
	}

	query := `
		INSERT INTO audit_logs (actor_type, actor_id, actor_email, action, 
			resource_type, resource_id, metadata, ip_address, user_agent, status, error_message)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err := r.db.ExecContext(ctx, query,
		req.ActorType, req.ActorID, req.ActorEmail, req.Action,
		req.ResourceType, req.ResourceID, metadata, req.IPAddress,
		req.UserAgent, req.Status, req.ErrorMessage,
	)
	return err
}

func (r *AuditRepository) List(ctx context.Context, filter models.AuditLogFilter) ([]*models.AuditLog, int64, error) {
	where := "WHERE 1=1"
	args := []interface{}{}
	argIdx := 1

	if filter.ActorType != "" {
		where += fmt.Sprintf(" AND actor_type = $%d", argIdx)
		args = append(args, filter.ActorType)
		argIdx++
	}
	if filter.ActorEmail != "" {
		where += fmt.Sprintf(" AND actor_email = $%d", argIdx)
		args = append(args, filter.ActorEmail)
		argIdx++
	}
	if filter.Action != "" {
		where += fmt.Sprintf(" AND action = $%d", argIdx)
		args = append(args, filter.Action)
		argIdx++
	}
	if filter.Status != "" {
		where += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, filter.Status)
		argIdx++
	}
	if filter.StartDate != nil {
		where += fmt.Sprintf(" AND created_at >= $%d", argIdx)
		args = append(args, *filter.StartDate)
		argIdx++
	}
	if filter.EndDate != nil {
		where += fmt.Sprintf(" AND created_at <= $%d", argIdx)
		args = append(args, *filter.EndDate)
		argIdx++
	}

	// Count total records
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM audit_logs %s", where)
	var total int64
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Fetch records with pagination
	offset := (filter.Page - 1) * filter.Limit
	query := fmt.Sprintf(`
		SELECT id, actor_type, actor_id, actor_email, action, resource_type,
			resource_id, metadata, ip_address, user_agent, status, error_message, created_at
		FROM audit_logs
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, where, argIdx, argIdx+1)

	args = append(args, filter.Limit, offset)
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var logs []*models.AuditLog
	for rows.Next() {
		log := &models.AuditLog{}
		err := rows.Scan(
			&log.ID, &log.ActorType, &log.ActorID, &log.ActorEmail, &log.Action,
			&log.ResourceType, &log.ResourceID, &log.Metadata, &log.IPAddress,
			&log.UserAgent, &log.Status, &log.ErrorMessage, &log.CreatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		logs = append(logs, log)
	}

	return logs, total, nil
}

func (r *AuditRepository) Cleanup(ctx context.Context, days int) (int64, error) {
	query := "DELETE FROM audit_logs WHERE created_at < $1"
	threshold := time.Now().AddDate(0, 0, -days)
	result, err := r.db.ExecContext(ctx, query, threshold)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
