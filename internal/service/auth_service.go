package service

import (
	"context"
	"erp-excel/config"
	"erp-excel/internal/dto"
	"erp-excel/internal/models"
	"erp-excel/internal/repository"
	"erp-excel/internal/utils"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// AuthService interface
type AuthService interface {
	Login(ctx context.Context, req dto.LoginRequest) (*dto.LoginResponse, error)
	ValidateToken(tokenString string) (*dto.TokenClaims, error)
	GenerateToken(user *models.User) (string, error)
	GetUserProfile(ctx context.Context, userID int) (*dto.UserResponse, error)
}

type authService struct {
	userRepo repository.UserRepository
	config   *config.Config
}

// NewAuthService creates a new auth service
func NewAuthService(userRepo repository.UserRepository, config *config.Config) AuthService {
	return &authService{
		userRepo: userRepo,
		config:   config,
	}
}

// Login authenticates a user
func (s *authService) Login(ctx context.Context, req dto.LoginRequest) (*dto.LoginResponse, error) {
	// Get user by username
	user, err := s.userRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		return nil, errors.New("invalid username or password")
	}

	// Verify password
	if !utils.CheckPasswordHash(req.Password, user.Password) {
		return nil, errors.New("invalid username or password")
	}

	// Check if user is active
	if !user.IsActive {
		return nil, errors.New("account is disabled")
	}

	// Update last login time
	if err := s.userRepo.UpdateLastLogin(ctx, user.ID); err != nil {
		// Just log this error, don't fail login
		fmt.Printf("Error updating last login: %v\n", err)
	}

	// Generate JWT token
	token, err := s.GenerateToken(user)
	if err != nil {
		return nil, fmt.Errorf("error generating token: %w", err)
	}

	// Get user roles for response
	roles, err := s.userRepo.GetUserRoles(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("error getting user roles: %w", err)
	}
	user.Roles = roles

	// Create response
	roleNames := make([]string, 0, len(roles))
	for _, role := range roles {
		roleNames = append(roleNames, role.Name)
	}

	userResp := &dto.UserResponse{
		ID:           user.ID,
		Username:     user.Username,
		FullName:     user.FullName,
		Email:        user.Email,
		Phone:        user.Phone,
		DepartmentID: user.DepartmentID,
		Department:   user.Department.Name,
		IsActive:     user.IsActive,
		CreatedAt:    user.CreatedAt,
		Roles:        roleNames,
	}

	return &dto.LoginResponse{
		User:  userResp,
		Token: token,
	}, nil
}

// ValidateToken validates a JWT token
func (s *authService) ValidateToken(tokenString string) (*dto.TokenClaims, error) {
	claims := &dto.TokenClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.config.JWT.Secret), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

// GenerateToken generates a JWT token for a user
func (s *authService) GenerateToken(user *models.User) (string, error) {
	// Set expiration time
	expirationTime := time.Now().Add(s.config.GetJWTExpiry())

	// Create claims
	claims := dto.TokenClaims{ // Sử dụng struct dto.TokenClaims
		UserID:       user.ID,
		Username:     user.Username,
		DepartmentID: user.DepartmentID,
		Exp:          expirationTime.Unix(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime), // Sử dụng ExpiresAt
			IssuedAt:  jwt.NewNumericDate(time.Now()),     // Thêm IssuedAt
		},
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign token with secret
	tokenString, err := token.SignedString([]byte(s.config.JWT.Secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// GetUserProfile retrieves the user profile by ID
func (s *authService) GetUserProfile(ctx context.Context, userID int) (*dto.UserResponse, error) {
	// Get user by ID
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("error getting user: %w", err)
	}

	// Get user roles for response
	roles, err := s.userRepo.GetUserRoles(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("error getting user roles: %w", err)
	}

	// Extract role names
	roleNames := make([]string, 0, len(roles))
	for _, role := range roles {
		roleNames = append(roleNames, role.Name)
	}

	departmentName := ""
	if user.Department != nil {
		departmentName = user.Department.Name
	}

	// Return user response DTO
	return &dto.UserResponse{
		ID:           user.ID,
		Username:     user.Username,
		FullName:     user.FullName,
		Email:        user.Email,
		Phone:        user.Phone,
		DepartmentID: user.DepartmentID,
		Department:   departmentName,
		IsActive:     user.IsActive,
		CreatedAt:    user.CreatedAt,
		Roles:        roleNames,
	}, nil
}
