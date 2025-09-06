package dto

import "github.com/golang-jwt/jwt/v4"

// LoginRequest represents login credentials
type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// LoginResponse represents login response with tokens
type LoginResponse struct {
	User  *UserResponse `json:"user"`
	Token string        `json:"token"`
}

// TokenClaims represents JWT claims
type TokenClaims struct {
	UserID       int    `json:"user_id"`
	Username     string `json:"username"`
	DepartmentID int    `json:"department_id"`
	Exp          int64  `json:"exp,omitempty"` // for compatibility
	jwt.RegisteredClaims
}
