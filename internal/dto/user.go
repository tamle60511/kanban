package dto

import "time"

// UserResponse represents user data for API responses
type UserResponse struct {
	ID           int       `json:"id"`
	Username     string    `json:"username"`
	FullName     string    `json:"full_name"`
	Email        string    `json:"email"`
	Phone        string    `json:"phone,omitempty"`
	DepartmentID int       `json:"department_id"`
	Department   string    `json:"department,omitempty"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	Roles        []string  `json:"roles,omitempty"`
}

// CreateUserRequest represents request to create a new user
type CreateUserRequest struct {
	Username     string `json:"username" validate:"required,min=3,max=50"`
	Password     string `json:"password" validate:"required,min=6"`
	FullName     string `json:"full_name" validate:"required"`
	Email        string `json:"email" validate:"required,email"`
	Phone        string `json:"phone" validate:"omitempty"`
	DepartmentID int    `json:"department_id" validate:"required,min=1"`
	RoleIDs      []int  `json:"role_ids" validate:"required,min=1,dive,min=1"`
}

// UpdateUserRequest represents request to update a user
type UpdateUserRequest struct {
	FullName     string `json:"full_name" validate:"omitempty"`
	Email        string `json:"email" validate:"omitempty,email"`
	Phone        string `json:"phone" validate:"omitempty"`
	DepartmentID int    `json:"department_id" validate:"omitempty,min=1"`
	IsActive     *bool  `json:"is_active" validate:"omitempty"`
}

// UpdatePasswordRequest represents request to update user password
type UpdatePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=6"`
	ConfirmPassword string `json:"confirm_password" validate:"required,eqfield=NewPassword"`
}
