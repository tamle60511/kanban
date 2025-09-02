package repository

import (
	"context"
	"database/sql"
	"erp-excel/internal/dto"
	"erp-excel/internal/models"
	"fmt"
)

// OperationRepository interface
type OperationRepository interface {
	GetAll(ctx context.Context) ([]*dto.OperationResponse, error)
	FindByCode(ctx context.Context, code string) (*models.Operation, error)
	GetByID(ctx context.Context, id int) (*models.Operation, error)
	LogAccess(ctx context.Context, log *models.AccessLog) (int, error)
	UpdateLogStatus(ctx context.Context, logID int, status string) (bool, error)
	GetRecentLogs(ctx context.Context, limit int) ([]*models.AccessLog, error)
}

type operationRepository struct {
	db *sql.DB
}

// NewOperationRepository creates a new operation repository
func NewOperationRepository(db *sql.DB) OperationRepository {
	return &operationRepository{
		db: db,
	}
}

// GetAll gets all operations
func (r *operationRepository) GetAll(ctx context.Context) ([]*dto.OperationResponse, error) {
	query := `
        SELECT id, name, code, description
        FROM operations
        ORDER BY name
    `

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error getting operations: %w", err)
	}
	defer rows.Close()

	var operations []*dto.OperationResponse
	for rows.Next() {
		var operation dto.OperationResponse
		err := rows.Scan(
			&operation.ID,
			&operation.Name,
			&operation.Code,
			&operation.Description,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning operation: %w", err)
		}

		operations = append(operations, &operation)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating operations: %w", err)
	}

	return operations, nil
}

// FindByCode gets an operation by code
func (r *operationRepository) FindByCode(ctx context.Context, code string) (*models.Operation, error) {
	query := `
        SELECT id, name, code, description, created_at, updated_at
        FROM operations
        WHERE code = @code
    `

	var operation models.Operation
	err := r.db.QueryRowContext(ctx, query, sql.Named("code", code)).Scan(
		&operation.ID,
		&operation.Name,
		&operation.Code,
		&operation.Description,
		&operation.CreatedAt,
		&operation.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("operation not found: %w", err)
		}
		return nil, fmt.Errorf("error getting operation: %w", err)
	}

	return &operation, nil
}

// GetByID gets an operation by ID
func (r *operationRepository) GetByID(ctx context.Context, id int) (*models.Operation, error) {
	query := `
        SELECT id, name, code, description, created_at, updated_at
        FROM operations
        WHERE id = @id
    `

	var operation models.Operation
	err := r.db.QueryRowContext(ctx, query, sql.Named("id", id)).Scan(
		&operation.ID,
		&operation.Name,
		&operation.Code,
		&operation.Description,
		&operation.CreatedAt,
		&operation.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("operation not found: %w", err)
		}
		return nil, fmt.Errorf("error getting operation: %w", err)
	}

	return &operation, nil
}

// LogAccess logs access to an operation
func (r *operationRepository) LogAccess(ctx context.Context, log *models.AccessLog) (int, error) {
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
		return 0, fmt.Errorf("error logging access: %w", err)
	}

	return id, nil
}

// UpdateLogStatus updates the status of an access log
func (r *operationRepository) UpdateLogStatus(ctx context.Context, logID int, status string) (bool, error) {
	query := `
        UPDATE access_logs
        SET status = @status
        WHERE id = @id
    `

	result, err := r.db.ExecContext(
		ctx,
		query,
		sql.Named("status", status),
		sql.Named("id", logID),
	)

	if err != nil {
		return false, fmt.Errorf("error updating log status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("error getting rows affected: %w", err)
	}

	return rowsAffected > 0, nil
}

// GetRecentLogs gets recent access logs
func (r *operationRepository) GetRecentLogs(ctx context.Context, limit int) ([]*models.AccessLog, error) {
	query := `
        SELECT *
        FROM (
            SELECT 
                l.id, 
                l.user_id, 
                l.operation_id, 
                l.access_time, 
                l.search_params, 
                l.ip_address, 
                l.status,
                u.username, 
                o.name as operation_name,
                ROW_NUMBER() OVER (ORDER BY l.access_time DESC) AS RowNum
            FROM access_logs l
            JOIN users u ON l.user_id = u.id
            JOIN operations o ON l.operation_id = o.id
        ) AS LogsWithRowNumbers
        WHERE RowNum BETWEEN 1 AND @limit
    `

	rows, err := r.db.QueryContext(
		ctx,
		query,
		sql.Named("limit", limit),
	)
	if err != nil {
		return nil, fmt.Errorf("error getting recent logs: %w", err)
	}
	defer rows.Close()

	var logs []*models.AccessLog
	for rows.Next() {
		var log models.AccessLog
		var username, operationName string
		var rowNum int

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
			&rowNum,
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
