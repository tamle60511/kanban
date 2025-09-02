package repository

import (
	"context"
	"database/sql"
	"erp-excel/internal/models"
	"fmt"
)

// AccessLogRepository interface
type AccessLogRepository interface {
	Create(ctx context.Context, log *models.AccessLog) (int, error)
	UpdateStatus(ctx context.Context, id int, status string) (bool, error)
	GetRecentLogs(ctx context.Context, limit int) ([]*models.AccessLog, error)
	GetUserLogs(ctx context.Context, userID int, limit int) ([]*models.AccessLog, error)
}

type accessLogRepository struct {
	db *sql.DB
}

// NewAccessLogRepository creates a new access log repository
func NewAccessLogRepository(db *sql.DB) AccessLogRepository {
	return &accessLogRepository{
		db: db,
	}
}

// Create adds a new access log entry
func (r *accessLogRepository) Create(ctx context.Context, log *models.AccessLog) (int, error) {
	query := `
        INSERT INTO access_logs (user_id, operation_id, access_time, search_params, ip_address, status)
        OUTPUT INSERTED.id
        VALUES (@user_id, @operation_id, @access_time, @search_params, @ip_address, @status)
    `

	var id int
	err := r.db.QueryRowContext(
		ctx,
		query,
		sql.Named("user_id", log.UserID),
		sql.Named("operation_id", log.OperationID),
		sql.Named("access_time", log.AccessTime),
		sql.Named("search_params", log.SearchParams),
		sql.Named("ip_address", log.IPAddress),
		sql.Named("status", log.Status),
	).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("error creating access log: %w", err)
	}

	return id, nil
}

// UpdateStatus updates the status of an access log
func (r *accessLogRepository) UpdateStatus(ctx context.Context, id int, status string) (bool, error) {
	query := `
        UPDATE access_logs
        SET status = @status
        WHERE id = @id
    `

	result, err := r.db.ExecContext(
		ctx,
		query,
		sql.Named("status", status),
		sql.Named("id", id),
	)

	if err != nil {
		return false, fmt.Errorf("error updating access log status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("error getting rows affected: %w", err)
	}

	return rowsAffected > 0, nil
}

// GetRecentLogs gets recent access logs
func (r *accessLogRepository) GetRecentLogs(ctx context.Context, limit int) ([]*models.AccessLog, error) {
	query := `
        SELECT l.id, l.user_id, l.operation_id, l.access_time, l.search_params, l.ip_address, l.status,
               u.username, o.name as operation_name
        FROM access_logs l
        JOIN users u ON l.user_id = u.id
        JOIN operations o ON l.operation_id = o.id
        ORDER BY l.access_time DESC
        OFFSET 0 ROWS
        FETCH NEXT @limit ROWS ONLY
    `

	rows, err := r.db.QueryContext(ctx, query, sql.Named("limit", limit))
	if err != nil {
		return nil, fmt.Errorf("error getting recent logs: %w", err)
	}
	defer rows.Close()

	var logs []*models.AccessLog
	for rows.Next() {
		var log models.AccessLog
		var username, operationName string

		err := rows.Scan(
			&log.ID,
			&log.UserID,
			&log.OperationID,
			&log.AccessTime,
			&log.SearchParams,
			&log.IPAddress,
			&log.Status,
			&username,
			&operationName,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning log: %w", err)
		}

		logs = append(logs, &log)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating logs: %w", err)
	}

	return logs, nil
}

// GetUserLogs gets access logs for a specific user
func (r *accessLogRepository) GetUserLogs(ctx context.Context, userID int, limit int) ([]*models.AccessLog, error) {
	query := `
        SELECT l.id, l.user_id, l.operation_id, l.access_time, l.search_params, l.ip_address, l.status,
               o.name as operation_name
        FROM access_logs l
        JOIN operations o ON l.operation_id = o.id
        WHERE l.user_id = @user_id
        ORDER BY l.access_time DESC
        OFFSET 0 ROWS
        FETCH NEXT @limit ROWS ONLY
    `

	rows, err := r.db.QueryContext(
		ctx,
		query,
		sql.Named("user_id", userID),
		sql.Named("limit", limit),
	)
	if err != nil {
		return nil, fmt.Errorf("error getting user logs: %w", err)
	}
	defer rows.Close()

	var logs []*models.AccessLog
	for rows.Next() {
		var log models.AccessLog
		var operationName string

		err := rows.Scan(
			&log.ID,
			&log.UserID,
			&log.OperationID,
			&log.AccessTime,
			&log.SearchParams,
			&log.IPAddress,
			&log.Status,
			&operationName,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning log: %w", err)
		}

		logs = append(logs, &log)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating logs: %w", err)
	}

	return logs, nil
}
