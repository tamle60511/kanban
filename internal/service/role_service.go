package service

import (
	"context"
	"erp-excel/internal/dto"
	"erp-excel/internal/models"
	"erp-excel/internal/repository"
	"fmt"
)

// RoleService interface
type RoleService interface {
	CreateRole(ctx context.Context, request dto.CreateRoleRequest) (*dto.RoleResponse, error)
	GetRoleByID(ctx context.Context, id int) (*dto.RoleResponse, error)
	UpdateRole(ctx context.Context, id int, request dto.UpdateRoleRequest) (*dto.RoleResponse, error)
	DeleteRole(ctx context.Context, id int) error
	GetAllRoles(ctx context.Context, limit, offset int) ([]*dto.RoleResponse, error)
	CountRoles(ctx context.Context) (int, error)
	AssignOperations(ctx context.Context, roleID int, operationIDs []int) error
}

type roleService struct {
	roleRepo repository.RoleRepository
}

// NewRoleService creates a new role service
func NewRoleService(roleRepo repository.RoleRepository) RoleService {
	return &roleService{
		roleRepo: roleRepo,
	}
}

// CreateRole creates a new role
func (s *roleService) CreateRole(ctx context.Context, request dto.CreateRoleRequest) (*dto.RoleResponse, error) {
	// Create role model
	role := &models.Role{
		Name:        request.Name,
		Description: request.Description,
	}

	// Save to database
	createdRole, err := s.roleRepo.Create(ctx, role)
	if err != nil {
		return nil, fmt.Errorf("error creating role: %w", err)
	}

	// Assign operations if provided
	if len(request.OperationIDs) > 0 {
		if err := s.roleRepo.AssignOperations(ctx, createdRole.ID, request.OperationIDs); err != nil {
			return nil, fmt.Errorf("error assigning operations: %w", err)
		}
	}

	// Return response
	return &dto.RoleResponse{
		ID:           createdRole.ID,
		Name:         createdRole.Name,
		Description:  createdRole.Description,
		CreatedAt:    createdRole.CreatedAt,
		UpdatedAt:    createdRole.UpdatedAt,
		OperationIDs: request.OperationIDs,
	}, nil
}

// GetRoleByID gets a role by ID
func (s *roleService) GetRoleByID(ctx context.Context, id int) (*dto.RoleResponse, error) {
	role, err := s.roleRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("error getting role: %w", err)
	}

	// Extract operation IDs
	operationIDs := make([]int, 0, len(role.Operations))
	for _, operation := range role.Operations {
		operationIDs = append(operationIDs, operation.ID)
	}

	return &dto.RoleResponse{
		ID:           role.ID,
		Name:         role.Name,
		Description:  role.Description,
		CreatedAt:    role.CreatedAt,
		UpdatedAt:    role.UpdatedAt,
		OperationIDs: operationIDs,
	}, nil
}

// UpdateRole updates a role
func (s *roleService) UpdateRole(ctx context.Context, id int, request dto.UpdateRoleRequest) (*dto.RoleResponse, error) {
	// Get existing role
	role, err := s.roleRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("error getting role: %w", err)
	}

	// Update fields if provided
	if request.Name != "" {
		role.Name = request.Name
	}

	if request.Description != "" {
		role.Description = request.Description
	}

	// Save to database
	if err := s.roleRepo.Update(ctx, role); err != nil {
		return nil, fmt.Errorf("error updating role: %w", err)
	}

	// Update operations if provided
	if len(request.OperationIDs) > 0 {
		if err := s.roleRepo.AssignOperations(ctx, role.ID, request.OperationIDs); err != nil {
			return nil, fmt.Errorf("error assigning operations: %w", err)
		}

		// Reload operations
		role, err = s.roleRepo.GetByID(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("error reloading role: %w", err)
		}
	}

	// Extract operation IDs
	operationIDs := make([]int, 0, len(role.Operations))
	for _, operation := range role.Operations {
		operationIDs = append(operationIDs, operation.ID)
	}

	// Return response
	return &dto.RoleResponse{
		ID:           role.ID,
		Name:         role.Name,
		Description:  role.Description,
		CreatedAt:    role.CreatedAt,
		UpdatedAt:    role.UpdatedAt,
		OperationIDs: operationIDs,
	}, nil
}

// DeleteRole deletes a role
func (s *roleService) DeleteRole(ctx context.Context, id int) error {
	return s.roleRepo.Delete(ctx, id)
}

// GetAllRoles gets all roles
func (s *roleService) GetAllRoles(ctx context.Context, limit, offset int) ([]*dto.RoleResponse, error) {
	roles, err := s.roleRepo.List(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("error listing roles: %w", err)
	}

	// Convert to response DTOs
	response := make([]*dto.RoleResponse, 0, len(roles))
	for _, role := range roles {
		// Get operations for this role
		operations, err := s.roleRepo.GetOperations(ctx, role.ID)
		if err != nil {
			return nil, fmt.Errorf("error getting operations for role: %w", err)
		}

		// Extract operation IDs
		operationIDs := make([]int, 0, len(operations))
		for _, operation := range operations {
			operationIDs = append(operationIDs, operation.ID)
		}

		response = append(response, &dto.RoleResponse{
			ID:           role.ID,
			Name:         role.Name,
			Description:  role.Description,
			CreatedAt:    role.CreatedAt,
			UpdatedAt:    role.UpdatedAt,
			OperationIDs: operationIDs,
		})
	}

	return response, nil
}

// CountRoles gets the total number of roles
func (s *roleService) CountRoles(ctx context.Context) (int, error) {
	return s.roleRepo.Count(ctx)
}

// AssignOperations assigns operations to a role
func (s *roleService) AssignOperations(ctx context.Context, roleID int, operationIDs []int) error {
	return s.roleRepo.AssignOperations(ctx, roleID, operationIDs)
}
