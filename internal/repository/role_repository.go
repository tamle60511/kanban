package repository

import (
	"context"
	"database/sql"
	"erp-excel/internal/models"
	"fmt"
	"time"
)

// RoleRepository interface
type RoleRepository interface {
	Create(ctx context.Context, role *models.Role) (*models.Role, error)
	GetByID(ctx context.Context, id int) (*models.Role, error)
	Update(ctx context.Context, role *models.Role) error
	Delete(ctx context.Context, id int) error
	List(ctx context.Context, limit, offset int) ([]*models.Role, error)
	Count(ctx context.Context) (int, error)
	GetOperations(ctx context.Context, roleID int) ([]*models.Operation, error)
	AssignOperations(ctx context.Context, roleID int, operationIDs []int) error
	RemoveOperations(ctx context.Context, roleID int, operationIDs []int) error
	CheckUserOperationAccess(ctx context.Context, userID int, operationID int) (bool, error)
}

type roleRepository struct {
	db *sql.DB
}

// NewRoleRepository creates a new role repository
func NewRoleRepository(db *sql.DB) RoleRepository {
	return &roleRepository{
		db: db,
	}
}

// Create adds a new role
func (r *roleRepository) Create(ctx context.Context, role *models.Role) (*models.Role, error) {
	query := `  
        INSERT INTO roles (name, description, created_at, updated_at)  
        OUTPUT INSERTED.id  
        VALUES (@name, @description, @created_at, @updated_at)  
    `

	var id int
	err := r.db.QueryRowContext(
		ctx,
		query,
		sql.Named("name", role.Name),
		sql.Named("description", role.Description),
		sql.Named("created_at", time.Now()),
		sql.Named("updated_at", time.Now()),
	).Scan(&id)

	if err != nil {
		return nil, fmt.Errorf("error creating role: %w", err)
	}

	role.ID = id
	return role, nil
}

// GetByID gets a role by ID
func (r *roleRepository) GetByID(ctx context.Context, id int) (*models.Role, error) {
	query := `  
        SELECT id, name, description, created_at, updated_at  
        FROM roles  
        WHERE id = @id  
    `

	var role models.Role
	err := r.db.QueryRowContext(ctx, query, sql.Named("id", id)).Scan(
		&role.ID,
		&role.Name,
		&role.Description,
		&role.CreatedAt,
		&role.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("role not found: %w", err)
		}
		return nil, fmt.Errorf("error getting role: %w", err)
	}

	// Get operations for this role
	operations, err := r.GetOperations(ctx, role.ID)
	if err != nil {
		return nil, fmt.Errorf("error getting role operations: %w", err)
	}
	role.Operations = operations

	return &role, nil
}

// Update updates a role
func (r *roleRepository) Update(ctx context.Context, role *models.Role) error {
	query := `  
        UPDATE roles  
        SET name = @name,  
            description = @description,  
            updated_at = @updated_at  
        WHERE id = @id  
    `

	_, err := r.db.ExecContext(
		ctx,
		query,
		sql.Named("name", role.Name),
		sql.Named("description", role.Description),
		sql.Named("updated_at", time.Now()),
		sql.Named("id", role.ID),
	)

	if err != nil {
		return fmt.Errorf("error updating role: %w", err)
	}

	return nil
}

// Delete deletes a role
func (r *roleRepository) Delete(ctx context.Context, id int) error {
	// Start a transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error beginning transaction: %w", err)
	}
	defer tx.Rollback()

	// Delete role_operations first
	_, err = tx.ExecContext(
		ctx,
		"DELETE FROM role_operations WHERE role_id = @role_id",
		sql.Named("role_id", id),
	)
	if err != nil {
		return fmt.Errorf("error deleting role operations: %w", err)
	}

	// Delete user_roles
	_, err = tx.ExecContext(
		ctx,
		"DELETE FROM user_roles WHERE role_id = @role_id",
		sql.Named("role_id", id),
	)
	if err != nil {
		return fmt.Errorf("error deleting user roles: %w", err)
	}

	// Delete role
	_, err = tx.ExecContext(
		ctx,
		"DELETE FROM roles WHERE id = @id",
		sql.Named("id", id),
	)
	if err != nil {
		return fmt.Errorf("error deleting role: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}

	return nil
}

// List gets a list of roles
func (r *roleRepository) List(ctx context.Context, limit, offset int) ([]*models.Role, error) {
	query := `  
        SELECT *
        FROM (
            SELECT 
                id, 
                name, 
                description, 
                created_at, 
                updated_at,
                ROW_NUMBER() OVER (ORDER BY name) AS RowNum
            FROM roles
        ) AS RolesWithRowNumbers
        WHERE RowNum BETWEEN @offset + 1 AND @offset + @limit
    `

	rows, err := r.db.QueryContext(
		ctx,
		query,
		sql.Named("limit", limit),
		sql.Named("offset", offset),
	)
	if err != nil {
		return nil, fmt.Errorf("error listing roles: %w", err)
	}
	defer rows.Close()

	var roles []*models.Role
	for rows.Next() {
		var role models.Role
		var rowNum int // Thêm biến để scan row number

		err := rows.Scan(
			&role.ID,
			&role.Name,
			&role.Description,
			&role.CreatedAt,
			&role.UpdatedAt,
			&rowNum, // Scan row number
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning role: %w", err)
		}

		roles = append(roles, &role)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating roles: %w", err)
	}

	return roles, nil
}

// Count gets the total number of roles
func (r *roleRepository) Count(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM roles").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("error counting roles: %w", err)
	}
	return count, nil
}

// GetOperations gets operations assigned to a role
func (r *roleRepository) GetOperations(ctx context.Context, roleID int) ([]*models.Operation, error) {
	query := `
        SELECT o.id, o.name, o.code, o.description, o.created_at, o.updated_at
        FROM operations o
        JOIN role_operations ro ON o.id = ro.operation_id
        WHERE ro.role_id = @role_id AND ro.can_access = 1
        ORDER BY o.name
    `

	rows, err := r.db.QueryContext(ctx, query, sql.Named("role_id", roleID))
	if err != nil {
		return nil, fmt.Errorf("error getting role operations: %w", err)
	}
	defer rows.Close()

	var operations []*models.Operation
	for rows.Next() {
		var operation models.Operation
		err := rows.Scan(
			&operation.ID,
			&operation.Name,
			&operation.Code,
			&operation.Description,
			&operation.CreatedAt,
			&operation.UpdatedAt,
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

// AssignOperations assigns operations to a role
func (r *roleRepository) AssignOperations(ctx context.Context, roleID int, operationIDs []int) error {
	// Start a transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error beginning transaction: %w", err)
	}
	defer tx.Rollback()

	// Delete existing operations first
	_, err = tx.ExecContext(
		ctx,
		"DELETE FROM role_operations WHERE role_id = @role_id",
		sql.Named("role_id", roleID),
	)
	if err != nil {
		return fmt.Errorf("error deleting existing operations: %w", err)
	}

	// Insert new operations
	for _, operationID := range operationIDs {
		_, err = tx.ExecContext(
			ctx,
			"INSERT INTO role_operations (role_id, operation_id, can_access, created_at) VALUES (@role_id, @operation_id, 1, @created_at)",
			sql.Named("role_id", roleID),
			sql.Named("operation_id", operationID),
			sql.Named("created_at", time.Now()),
		)
		if err != nil {
			return fmt.Errorf("error assigning operation: %w", err)
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}

	return nil
}

// RemoveOperations removes operations from a role
func (r *roleRepository) RemoveOperations(ctx context.Context, roleID int, operationIDs []int) error {
	query := `
        DELETE FROM role_operations
        WHERE role_id = @role_id AND operation_id IN (
    `

	// Build the IN clause with named parameters
	params := []interface{}{sql.Named("role_id", roleID)}
	for i, operationID := range operationIDs {
		if i > 0 {
			query += ", "
		}
		paramName := fmt.Sprintf("operation_id_%d", i)
		query += "@" + paramName
		params = append(params, sql.Named(paramName, operationID))
	}
	query += ")"

	// Execute the query
	_, err := r.db.ExecContext(ctx, query, params...)
	if err != nil {
		return fmt.Errorf("error removing operations: %w", err)
	}

	return nil
}

// CheckUserOperationAccess checks if a user has access to an operation
func (r *roleRepository) CheckUserOperationAccess(ctx context.Context, userID int, operationID int) (bool, error) {
	query := `
        SELECT COUNT(*) 
        FROM user_roles ur
        JOIN role_operations ro ON ur.role_id = ro.role_id
        WHERE ur.user_id = @user_id 
          AND ro.operation_id = @operation_id 
          AND ro.can_access = 1
    `

	var count int
	err := r.db.QueryRowContext(
		ctx,
		query,
		sql.Named("user_id", userID),
		sql.Named("operation_id", operationID),
	).Scan(&count)

	if err != nil {
		return false, fmt.Errorf("error checking operation access: %w", err)
	}

	return count > 0, nil
}
