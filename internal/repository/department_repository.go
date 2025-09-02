package repository

import (
	"context"
	"database/sql"
	"erp-excel/internal/models"
	"fmt"
	"time"
)

// DepartmentRepository interface
type DepartmentRepository interface {
	Create(ctx context.Context, department *models.Department) (*models.Department, error)
	GetByID(ctx context.Context, id int) (*models.Department, error)
	Update(ctx context.Context, department *models.Department) error
	Delete(ctx context.Context, id int) error
	List(ctx context.Context, limit, offset int) ([]*models.Department, error)
	Count(ctx context.Context) (int, error)
	GetUserCount(ctx context.Context, departmentID int) (int, error)
}

type departmentRepository struct {
	db *sql.DB
}

// NewDepartmentRepository creates a new department repository
func NewDepartmentRepository(db *sql.DB) DepartmentRepository {
	return &departmentRepository{
		db: db,
	}
}

// Create adds a new department
func (r *departmentRepository) Create(ctx context.Context, department *models.Department) (*models.Department, error) {
	query := `
        INSERT INTO departments (name, code, description, is_active, created_at, updated_at)
        OUTPUT INSERTED.id
        VALUES (@name, @code, @description, @is_active, @created_at, @updated_at)
    `

	var id int
	err := r.db.QueryRowContext(
		ctx,
		query,
		sql.Named("name", department.Name),
		sql.Named("code", department.Code),
		sql.Named("description", department.Description),
		sql.Named("is_active", department.IsActive),
		sql.Named("created_at", time.Now()),
		sql.Named("updated_at", time.Now()),
	).Scan(&id)

	if err != nil {
		return nil, fmt.Errorf("error creating department: %w", err)
	}

	department.ID = id
	return department, nil
}

// GetByID gets a department by ID
func (r *departmentRepository) GetByID(ctx context.Context, id int) (*models.Department, error) {
	query := `
        SELECT id, name, code, description, is_active, created_at, updated_at
        FROM departments
        WHERE id = @id
    `

	var department models.Department
	err := r.db.QueryRowContext(ctx, query, sql.Named("id", id)).Scan(
		&department.ID,
		&department.Name,
		&department.Code,
		&department.Description,
		&department.IsActive,
		&department.CreatedAt,
		&department.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("department not found: %w", err)
		}
		return nil, fmt.Errorf("error getting department: %w", err)
	}

	// Get user count for this department
	if _, err := r.GetUserCount(ctx, department.ID); err == nil {
		// Just log the error, don't fail the operation
		fmt.Printf("Error getting user count: %v\n", err)
	}

	return &department, nil
}

// Update updates a department
func (r *departmentRepository) Update(ctx context.Context, department *models.Department) error {
	query := `
        UPDATE departments
        SET name = @name,
            description = @description,
            is_active = @is_active,
            updated_at = @updated_at
        WHERE id = @id
    `

	_, err := r.db.ExecContext(
		ctx,
		query,
		sql.Named("name", department.Name),
		sql.Named("description", department.Description),
		sql.Named("is_active", department.IsActive),
		sql.Named("updated_at", time.Now()),
		sql.Named("id", department.ID),
	)

	if err != nil {
		return fmt.Errorf("error updating department: %w", err)
	}

	return nil
}

// Delete deactivates a department
func (r *departmentRepository) Delete(ctx context.Context, id int) error {
	query := `
        UPDATE departments
        SET is_active = 0,
            updated_at = @updated_at
        WHERE id = @id
    `

	_, err := r.db.ExecContext(
		ctx,
		query,
		sql.Named("updated_at", time.Now()),
		sql.Named("id", id),
	)

	if err != nil {
		return fmt.Errorf("error deleting department: %w", err)
	}

	return nil
}

// List gets a list of departments
func (r *departmentRepository) List(ctx context.Context, limit, offset int) ([]*models.Department, error) {
	query := `
        SELECT id, name, code, description, is_active, created_at, updated_at
        FROM (
            SELECT 
                id, name, code, description, is_active, created_at, updated_at,
                ROW_NUMBER() OVER (ORDER BY name) AS RowNum
            FROM departments
        ) AS DepartmentWithRowNumbers
        WHERE RowNum BETWEEN @offset + 1 AND @offset + @limit
        ORDER BY name
    `

	rows, err := r.db.QueryContext(
		ctx,
		query,
		sql.Named("limit", limit),
		sql.Named("offset", offset),
	)
	if err != nil {
		return nil, fmt.Errorf("error listing departments: %w", err)
	}
	defer rows.Close()

	var departments []*models.Department
	for rows.Next() {
		var department models.Department
		err := rows.Scan(
			&department.ID,
			&department.Name,
			&department.Code,
			&department.Description,
			&department.IsActive,
			&department.CreatedAt,
			&department.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning department: %w", err)
		}

		departments = append(departments, &department)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating departments: %w", err)
	}

	return departments, nil
}

// Count gets the total number of departments
func (r *departmentRepository) Count(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM departments").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("error counting departments: %w", err)
	}
	return count, nil
}

// GetUserCount gets the number of users in a department
func (r *departmentRepository) GetUserCount(ctx context.Context, departmentID int) (int, error) {
	query := `
        SELECT COUNT(*) 
        FROM users 
        WHERE department_id = @department_id
    `

	var count int
	err := r.db.QueryRowContext(ctx, query, sql.Named("department_id", departmentID)).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("error counting users in department: %w", err)
	}

	return count, nil
}
