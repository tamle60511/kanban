package dto

// DepartmentResponse represents department data for API responses
type DepartmentResponse struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Code        string `json:"code"`
	Description string `json:"description,omitempty"`
	IsActive    bool   `json:"is_active"`
	UserCount   int    `json:"user_count,omitempty"`
}

// CreateDepartmentRequest represents request to create a new department
type CreateDepartmentRequest struct {
	Name        string `json:"name" validate:"required"`
	Code        string `json:"code" validate:"required,min=2,max=20"`
	Description string `json:"description" validate:"omitempty"`
	IsActive    *bool  `json:"is_active" validate:"omitempty"`
}

// UpdateDepartmentRequest represents request to update a department
type UpdateDepartmentRequest struct {
	Name        string `json:"name" validate:"omitempty"`
	Description string `json:"description" validate:"omitempty"`
	IsActive    *bool  `json:"is_active" validate:"omitempty"`
}
