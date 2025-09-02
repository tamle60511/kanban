package models

import "time"

// User represents a system user
type User struct {
	ID           int         `json:"id"`
	Username     string      `json:"username"`
	Password     string      `json:"-"` // Don't expose password
	FullName     string      `json:"full_name"`
	Email        string      `json:"email"`
	Phone        string      `json:"phone,omitempty"`
	DepartmentID int         `json:"department_id"`
	Department   *Department `json:"department,omitempty"`
	IsActive     bool        `json:"is_active"`
	LastLogin    time.Time   `json:"last_login,omitempty"`
	CreatedAt    time.Time   `json:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at"`
	Roles        []*Role     `json:"roles,omitempty"`
}

// UserRole represents the relationship between users and roles
type UserRole struct {
	UserID    int       `json:"user_id"`
	RoleID    int       `json:"role_id"`
	CreatedAt time.Time `json:"created_at"`
}
