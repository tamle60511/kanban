package repository

import (
	"context"
	"database/sql"
	"erp-excel/internal/models"
	"fmt"
	"time"
)

// UserRepository interface
type UserRepository interface {
	Create(ctx context.Context, user *models.User) (*models.User, error)
	GetByID(ctx context.Context, id int) (*models.User, error)
	GetByUsername(ctx context.Context, username string) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
	UpdatePassword(ctx context.Context, userID int, hashedPassword string) error
	Delete(ctx context.Context, id int) error
	List(ctx context.Context, limit, offset int) ([]*models.User, error)
	Count(ctx context.Context) (int, error)
	GetUserRoles(ctx context.Context, userID int) ([]*models.Role, error)
	AssignRoles(ctx context.Context, userID int, roleIDs []int) error
	RemoveRoles(ctx context.Context, userID int, roleIDs []int) error
	UpdateLastLogin(ctx context.Context, userID int) error
}

type userRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{
		db: db,
	}
}

// Create adds a new user to the database
func (r *userRepository) Create(ctx context.Context, user *models.User) (*models.User, error) {
	query := `
        INSERT INTO users (username, password, full_name, email, phone, department_id, is_active, created_at, updated_at)
        OUTPUT INSERTED.id
        VALUES (@username, @password, @full_name, @email, @phone, @department_id, @is_active, @created_at, @updated_at)
    `

	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error preparing statement: %w", err)
	}
	defer stmt.Close()

	var id int
	err = stmt.QueryRowContext(
		ctx,
		sql.Named("username", user.Username),
		sql.Named("password", user.Password),
		sql.Named("full_name", user.FullName),
		sql.Named("email", user.Email),
		sql.Named("phone", user.Phone),
		sql.Named("department_id", user.DepartmentID),
		sql.Named("is_active", user.IsActive),
		sql.Named("created_at", time.Now()),
		sql.Named("updated_at", time.Now()),
	).Scan(&id)

	if err != nil {
		return nil, fmt.Errorf("error creating user: %w", err)
	}

	user.ID = id
	return user, nil
}

// GetByID gets a user by ID
func (r *userRepository) GetByID(ctx context.Context, id int) (*models.User, error) {
	query := `
        SELECT u.id, u.username, u.password, u.full_name, u.email, u.phone, u.department_id, 
               u.is_active, u.last_login, u.created_at, u.updated_at,
               d.name as department_name
        FROM users u
        LEFT JOIN departments d ON u.department_id = d.id
        WHERE u.id = @id
    `

	var user models.User
	var department models.Department
	var lastLogin sql.NullTime

	err := r.db.QueryRowContext(ctx, query, sql.Named("id", id)).Scan(
		&user.ID,
		&user.Username,
		&user.Password,
		&user.FullName,
		&user.Email,
		&user.Phone,
		&user.DepartmentID,
		&user.IsActive,
		&lastLogin,
		&user.CreatedAt,
		&user.UpdatedAt,
		&department.Name,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found: %w", err)
		}
		return nil, fmt.Errorf("error getting user: %w", err)
	}

	if lastLogin.Valid {
		user.LastLogin = lastLogin.Time
	}

	department.ID = user.DepartmentID
	user.Department = &department

	// Get user roles
	roles, err := r.GetUserRoles(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("error getting user roles: %w", err)
	}
	user.Roles = roles

	return &user, nil
}

// GetByUsername gets a user by username
func (r *userRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	query := `
        SELECT u.id, u.username, u.password, u.full_name, u.email, u.phone, u.department_id, 
               u.is_active, u.last_login, u.created_at, u.updated_at,
               d.name as department_name
        FROM users u
        LEFT JOIN departments d ON u.department_id = d.id
        WHERE u.username = @username
    `

	var user models.User
	var department models.Department
	var lastLogin sql.NullTime

	err := r.db.QueryRowContext(ctx, query, sql.Named("username", username)).Scan(
		&user.ID,
		&user.Username,
		&user.Password,
		&user.FullName,
		&user.Email,
		&user.Phone,
		&user.DepartmentID,
		&user.IsActive,
		&lastLogin,
		&user.CreatedAt,
		&user.UpdatedAt,
		&department.Name,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found: %w", err)
		}
		return nil, fmt.Errorf("error getting user: %w", err)
	}

	if lastLogin.Valid {
		user.LastLogin = lastLogin.Time
	}

	department.ID = user.DepartmentID
	user.Department = &department

	return &user, nil
}

// Update updates a user
func (r *userRepository) Update(ctx context.Context, user *models.User) error {
	query := `
        UPDATE users
        SET full_name = @full_name,
            email = @email,
            phone = @phone,
            department_id = @department_id,
            is_active = @is_active,
            updated_at = @updated_at
        WHERE id = @id
    `

	_, err := r.db.ExecContext(
		ctx,
		query,
		sql.Named("full_name", user.FullName),
		sql.Named("email", user.Email),
		sql.Named("phone", user.Phone),
		sql.Named("department_id", user.DepartmentID),
		sql.Named("is_active", user.IsActive),
		sql.Named("updated_at", time.Now()),
		sql.Named("id", user.ID),
	)

	if err != nil {
		return fmt.Errorf("error updating user: %w", err)
	}

	return nil
}

// UpdatePassword updates a user's password
func (r *userRepository) UpdatePassword(ctx context.Context, userID int, hashedPassword string) error {
	query := `
        UPDATE users
        SET password = @password,
            updated_at = @updated_at
        WHERE id = @id
    `

	_, err := r.db.ExecContext(
		ctx,
		query,
		sql.Named("password", hashedPassword),
		sql.Named("updated_at", time.Now()),
		sql.Named("id", userID),
	)

	if err != nil {
		return fmt.Errorf("error updating password: %w", err)
	}

	return nil
}

// Delete soft deletes a user
func (r *userRepository) Delete(ctx context.Context, id int) error {
	query := `
        UPDATE users
        SET is_active = 0
        WHERE id = @id
    `

	_, err := r.db.ExecContext(ctx, query, sql.Named("id", id))
	if err != nil {
		return fmt.Errorf("error deleting user: %w", err)
	}

	return nil
}

// List gets a list of users
func (r *userRepository) List(ctx context.Context, limit, offset int) ([]*models.User, error) {
	query := `
        SELECT *
        FROM (
            SELECT 
                u.id, 
                u.username, 
                u.full_name, 
                u.email, 
                u.phone, 
                u.department_id, 
                u.is_active, 
                u.last_login, 
                u.created_at, 
                u.updated_at,
                d.name as department_name,
                ROW_NUMBER() OVER (ORDER BY u.id) AS RowNum
            FROM users u
            LEFT JOIN departments d ON u.department_id = d.id
        ) AS UsersWithRowNumbers
        WHERE RowNum BETWEEN @offset + 1 AND @offset + @limit
    `

	rows, err := r.db.QueryContext(
		ctx,
		query,
		sql.Named("limit", limit),
		sql.Named("offset", offset),
	)
	if err != nil {
		return nil, fmt.Errorf("error listing users: %w", err)
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		var user models.User
		var department models.Department
		var lastLogin sql.NullTime
		var rowNum int // Thêm biến để scan row number

		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.FullName,
			&user.Email,
			&user.Phone,
			&user.DepartmentID,
			&user.IsActive,
			&lastLogin,
			&user.CreatedAt,
			&user.UpdatedAt,
			&department.Name,
			&rowNum, // Scan row number
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning user: %w", err)
		}

		if lastLogin.Valid {
			user.LastLogin = lastLogin.Time
		}

		department.ID = user.DepartmentID
		user.Department = &department

		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating users: %w", err)
	}

	return users, nil
}

// Count gets the total number of users
func (r *userRepository) Count(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("error counting users: %w", err)
	}
	return count, nil
}

// GetUserRoles gets a user's roles
func (r *userRepository) GetUserRoles(ctx context.Context, userID int) ([]*models.Role, error) {
	query := `
        SELECT r.id, r.name, r.description
        FROM roles r
        JOIN user_roles ur ON r.id = ur.role_id
        WHERE ur.user_id = @user_id
    `

	rows, err := r.db.QueryContext(ctx, query, sql.Named("user_id", userID))
	if err != nil {
		return nil, fmt.Errorf("error getting user roles: %w", err)
	}
	defer rows.Close()

	var roles []*models.Role
	for rows.Next() {
		var role models.Role
		err := rows.Scan(&role.ID, &role.Name, &role.Description)
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

// AssignRoles assigns roles to a user
func (r *userRepository) AssignRoles(ctx context.Context, userID int, roleIDs []int) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error beginning transaction: %w", err)
	}
	defer tx.Rollback()

	// Delete existing roles first
	_, err = tx.ExecContext(
		ctx,
		"DELETE FROM user_roles WHERE user_id = @user_id",
		sql.Named("user_id", userID),
	)
	if err != nil {
		return fmt.Errorf("error deleting existing roles: %w", err)
	}

	// Insert new roles
	for _, roleID := range roleIDs {
		_, err = tx.ExecContext(
			ctx,
			"INSERT INTO user_roles (user_id, role_id, created_at) VALUES (@user_id, @role_id, @created_at)",
			sql.Named("user_id", userID),
			sql.Named("role_id", roleID),
			sql.Named("created_at", time.Now()),
		)
		if err != nil {
			return fmt.Errorf("error assigning role: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}

	return nil
}

// RemoveRoles removes roles from a user
func (r *userRepository) RemoveRoles(ctx context.Context, userID int, roleIDs []int) error {
	query := `
        DELETE FROM user_roles
        WHERE user_id = @user_id AND role_id IN (
    `

	params := []interface{}{sql.Named("user_id", userID)}
	for i, roleID := range roleIDs {
		if i > 0 {
			query += ", "
		}
		paramName := fmt.Sprintf("role_id_%d", i)
		query += "@" + paramName
		params = append(params, sql.Named(paramName, roleID))
	}
	query += ")"

	_, err := r.db.ExecContext(ctx, query, params...)
	if err != nil {
		return fmt.Errorf("error removing roles: %w", err)
	}

	return nil
}

// UpdateLastLogin updates a user's last login time
func (r *userRepository) UpdateLastLogin(ctx context.Context, userID int) error {
	query := `
        UPDATE users
        SET last_login = @last_login
        WHERE id = @id
    `

	_, err := r.db.ExecContext(
		ctx,
		query,
		sql.Named("last_login", time.Now()),
		sql.Named("id", userID),
	)

	if err != nil {
		return fmt.Errorf("error updating last login: %w", err)
	}

	return nil
}
