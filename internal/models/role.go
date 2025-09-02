package models

import "time"

// Role represents a user role with permissions
type Role struct {
	ID          int          `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description,omitempty"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
	Operations  []*Operation `json:"operations,omitempty"`
}

// RoleOperation represents the relationship between roles and operations
type RoleOperation struct {
	RoleID      int       `json:"role_id"`
	OperationID int       `json:"operation_id"`
	CanAccess   bool      `json:"can_access"`
	CreatedAt   time.Time `json:"created_at"`
}
