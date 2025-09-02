package dto

import "time"

// RoleResponse represents role data for API responses
type RoleResponse struct {
	ID           int       `json:"id"`
	Name         string    `json:"name"`
	Description  string    `json:"description,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	OperationIDs []int     `json:"operation_ids,omitempty"`
	UserCount    int       `json:"user_count,omitempty"`
}

// CreateRoleRequest represents request to create a new role
type CreateRoleRequest struct {
	Name         string `json:"name" validate:"required"`
	Description  string `json:"description" validate:"omitempty"`
	OperationIDs []int  `json:"operation_ids" validate:"omitempty,dive,min=1"`
}

// UpdateRoleRequest represents request to update a role
type UpdateRoleRequest struct {
	Name         string `json:"name" validate:"omitempty"`
	Description  string `json:"description" validate:"omitempty"`
	OperationIDs []int  `json:"operation_ids" validate:"omitempty,dive,min=1"`
}

// OperationResponse represents operation data for API responses
type OperationResponse struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Code        string `json:"code"`
	Description string `json:"description,omitempty"`
}
