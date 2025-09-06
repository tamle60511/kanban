package service

import (
	"context"
	"erp-excel/internal/dto"
	"erp-excel/internal/models"
	"erp-excel/internal/repository"
	"erp-excel/internal/utils"
	"errors"
	"fmt"
)

// UserService interface
type UserService interface {
	CreateUser(ctx context.Context, request dto.CreateUserRequest) (*dto.UserResponse, error)
	GetUserByID(ctx context.Context, id int) (*dto.UserResponse, error)
	UpdateUser(ctx context.Context, id int, request dto.UpdateUserRequest) (*dto.UserResponse, error)
	UpdateUserPassword(ctx context.Context, id int, request dto.UpdatePasswordRequest) error
	DeleteUser(ctx context.Context, id int) error
	GetAllUsers(ctx context.Context, limit, offset int) ([]*dto.UserResponse, error)
	CountUsers(ctx context.Context) (int, error)
	AssignRolesToUser(ctx context.Context, userID int, roleIDs []int) error
}

type userService struct {
	userRepo       repository.UserRepository
	departmentRepo repository.DepartmentRepository
	roleRepo       repository.RoleRepository
	authService    AuthService
}

// NewUserService creates a new user service
func NewUserService(
	userRepo repository.UserRepository,
	departmentRepo repository.DepartmentRepository,
	roleRepo repository.RoleRepository,
	authService AuthService,
) UserService {
	return &userService{
		userRepo:       userRepo,
		departmentRepo: departmentRepo,
		roleRepo:       roleRepo,
		authService:    authService,
	}
}

// CreateUser creates a new user
func (s *userService) CreateUser(ctx context.Context, request dto.CreateUserRequest) (*dto.UserResponse, error) {
	// Validate department exists
	department, err := s.departmentRepo.GetByID(ctx, request.DepartmentID)
	if err != nil {
		return nil, fmt.Errorf("invalid department: %w", err)
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(request.Password)
	if err != nil {
		return nil, fmt.Errorf("error hashing password: %w", err)
	}

	// Create user model
	user := &models.User{
		Username:     request.Username,
		Password:     hashedPassword,
		FullName:     request.FullName,
		Email:        request.Email,
		DepartmentID: request.DepartmentID,
		IsActive:     true,
	}

	// Save to database
	createdUser, err := s.userRepo.Create(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("error creating user: %w", err)
	}

	// Assign roles
	if err := s.userRepo.AssignRoles(ctx, createdUser.ID, request.RoleIDs); err != nil {
		return nil, fmt.Errorf("error assigning roles: %w", err)
	}

	// Get roles for response
	roles, err := s.userRepo.GetUserRoles(ctx, createdUser.ID)
	if err != nil {
		return nil, fmt.Errorf("error getting user roles: %w", err)
	}

	// Extract role names
	roleNames := make([]string, 0, len(roles))
	for _, role := range roles {
		roleNames = append(roleNames, role.Name)
	}

	// Return response
	return &dto.UserResponse{
		ID:           createdUser.ID,
		Username:     createdUser.Username,
		FullName:     createdUser.FullName,
		Email:        createdUser.Email,
		DepartmentID: createdUser.DepartmentID,
		Department:   department.Name,
		IsActive:     createdUser.IsActive,
		CreatedAt:    createdUser.CreatedAt,
		Roles:        roleNames,
	}, nil
}

// GetUserByID gets a user by ID
func (s *userService) GetUserByID(ctx context.Context, id int) (*dto.UserResponse, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("error getting user: %w", err)
	}

	// Prepare response
	roleNames := make([]string, 0, len(user.Roles))
	for _, role := range user.Roles {
		roleNames = append(roleNames, role.Name)
	}

	return &dto.UserResponse{
		ID:           user.ID,
		Username:     user.Username,
		FullName:     user.FullName,
		Email:        user.Email,
		DepartmentID: user.DepartmentID,
		Department:   user.Department.Name,
		IsActive:     user.IsActive,
		CreatedAt:    user.CreatedAt,
		Roles:        roleNames,
	}, nil
}

// UpdateUser updates a user
func (s *userService) UpdateUser(ctx context.Context, id int, request dto.UpdateUserRequest) (*dto.UserResponse, error) {
	// Get existing user
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("error getting user: %w", err)
	}

	// Update fields if provided
	if request.FullName != "" {
		user.FullName = request.FullName
	}

	if request.Email != "" {
		user.Email = request.Email
	}

	if request.DepartmentID != 0 {
		// Validate department exists
		if _, err := s.departmentRepo.GetByID(ctx, request.DepartmentID); err != nil {
			return nil, fmt.Errorf("invalid department: %w", err)
		}
		user.DepartmentID = request.DepartmentID
	}

	if request.IsActive != nil {
		user.IsActive = *request.IsActive
	}

	// Save to database
	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("error updating user: %w", err)
	}

	// Get department name
	department, err := s.departmentRepo.GetByID(ctx, user.DepartmentID)
	if err != nil {
		// Log error but don't fail the operation
		fmt.Printf("Error getting department: %v\n", err)
	}

	// Get roles for response
	roles, err := s.userRepo.GetUserRoles(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("error getting user roles: %w", err)
	}

	// Extract role names
	roleNames := make([]string, 0, len(roles))
	for _, role := range roles {
		roleNames = append(roleNames, role.Name)
	}

	// Return response
	departmentName := ""
	if department != nil {
		departmentName = department.Name
	}

	return &dto.UserResponse{
		ID:           user.ID,
		Username:     user.Username,
		FullName:     user.FullName,
		Email:        user.Email,
		DepartmentID: user.DepartmentID,
		Department:   departmentName,
		IsActive:     user.IsActive,
		CreatedAt:    user.CreatedAt,
		Roles:        roleNames,
	}, nil
}

// UpdateUserPassword updates a user's password
func (s *userService) UpdateUserPassword(ctx context.Context, id int, request dto.UpdatePasswordRequest) error {
	// Get existing user
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("error getting user: %w", err)
	}

	// Verify current password
	if !utils.CheckPasswordHash(request.CurrentPassword, user.Password) {
		return errors.New("current password is incorrect")
	}

	// Check that new password and confirmation match
	if request.NewPassword != request.ConfirmPassword {
		return errors.New("new password and confirmation do not match")
	}

	// Hash new password
	hashedPassword, err := utils.HashPassword(request.NewPassword)
	if err != nil {
		return fmt.Errorf("error hashing password: %w", err)
	}

	// Update password
	if err := s.userRepo.UpdatePassword(ctx, id, hashedPassword); err != nil {
		return fmt.Errorf("error updating password: %w", err)
	}

	return nil
}

// DeleteUser deletes (deactivates) a user
func (s *userService) DeleteUser(ctx context.Context, id int) error {
	return s.userRepo.Delete(ctx, id)
}

// In UserService.GetAllUsers
func (s *userService) GetAllUsers(ctx context.Context, limit, offset int) ([]*dto.UserResponse, error) {
	users, err := s.userRepo.List(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("error listing users: %w", err)
	}

	// Convert to response DTOs
	response := make([]*dto.UserResponse, 0, len(users))
	for _, user := range users {
		// Extract role names
		roleNames := make([]string, 0, len(user.Roles))
		for _, role := range user.Roles {
			roleNames = append(roleNames, role.Name)
		}

		departmentName := ""
		if user.Department != nil {
			departmentName = user.Department.Name
		}

		response = append(response, &dto.UserResponse{
			ID:           user.ID,
			Username:     user.Username,
			FullName:     user.FullName,
			Email:        user.Email,
			DepartmentID: user.DepartmentID,
			Department:   departmentName,
			IsActive:     user.IsActive,
			CreatedAt:    user.CreatedAt,
			Roles:        roleNames,
		})
	}

	return response, nil
}

// CountUsers gets the total number of users
func (s *userService) CountUsers(ctx context.Context) (int, error) {
	return s.userRepo.Count(ctx)
}

// AssignRolesToUser assigns roles to a user
func (s *userService) AssignRolesToUser(ctx context.Context, userID int, roleIDs []int) error {
	return s.userRepo.AssignRoles(ctx, userID, roleIDs)
}
