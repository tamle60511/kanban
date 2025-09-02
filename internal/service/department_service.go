package service

import (
	"context"
	"erp-excel/internal/dto"
	"erp-excel/internal/models"
	"erp-excel/internal/repository"
	"errors"
	"fmt"
)

// DepartmentService interface
type DepartmentService interface {
	CreateDepartment(ctx context.Context, request dto.CreateDepartmentRequest) (*dto.DepartmentResponse, error)
	GetDepartmentByID(ctx context.Context, id int) (*dto.DepartmentResponse, error)
	UpdateDepartment(ctx context.Context, id int, request dto.UpdateDepartmentRequest) (*dto.DepartmentResponse, error)
	DeleteDepartment(ctx context.Context, id int) error
	GetAllDepartments(ctx context.Context, limit, offset int) ([]*dto.DepartmentResponse, error)
	CountDepartments(ctx context.Context) (int, error)
}

type departmentService struct {
	departmentRepo repository.DepartmentRepository
}

// NewDepartmentService creates a new department service
func NewDepartmentService(departmentRepo repository.DepartmentRepository) DepartmentService {
	return &departmentService{
		departmentRepo: departmentRepo,
	}
}

// CreateDepartment creates a new department
func (s *departmentService) CreateDepartment(ctx context.Context, request dto.CreateDepartmentRequest) (*dto.DepartmentResponse, error) {
	// Create department model
	isActive := true
	if request.IsActive != nil {
		isActive = *request.IsActive
	}

	department := &models.Department{
		Name:        request.Name,
		Code:        request.Code,
		Description: request.Description,
		IsActive:    isActive,
	}

	// Save to database
	createdDepartment, err := s.departmentRepo.Create(ctx, department)
	if err != nil {
		return nil, fmt.Errorf("error creating department: %w", err)
	}

	// Return response
	return &dto.DepartmentResponse{
		ID:          createdDepartment.ID,
		Name:        createdDepartment.Name,
		Code:        createdDepartment.Code,
		Description: createdDepartment.Description,
		IsActive:    createdDepartment.IsActive,
	}, nil
}

// GetDepartmentByID gets a department by ID
func (s *departmentService) GetDepartmentByID(ctx context.Context, id int) (*dto.DepartmentResponse, error) {
	department, err := s.departmentRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("error getting department: %w", err)
	}

	// Get user count
	userCount, err := s.departmentRepo.GetUserCount(ctx, department.ID)
	if err != nil {
		// Log the error but continue
		fmt.Printf("Error getting user count: %v\n", err)
		userCount = 0
	}

	return &dto.DepartmentResponse{
		ID:          department.ID,
		Name:        department.Name,
		Code:        department.Code,
		Description: department.Description,
		IsActive:    department.IsActive,
		UserCount:   userCount,
	}, nil
}

// UpdateDepartment updates a department
func (s *departmentService) UpdateDepartment(ctx context.Context, id int, request dto.UpdateDepartmentRequest) (*dto.DepartmentResponse, error) {
	// Get existing department
	department, err := s.departmentRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("error getting department: %w", err)
	}

	// Update fields if provided
	if request.Name != "" {
		department.Name = request.Name
	}

	if request.Description != "" {
		department.Description = request.Description
	}

	if request.IsActive != nil {
		department.IsActive = *request.IsActive
	}

	// Save to database
	if err := s.departmentRepo.Update(ctx, department); err != nil {
		return nil, fmt.Errorf("error updating department: %w", err)
	}

	// Get user count
	userCount, err := s.departmentRepo.GetUserCount(ctx, department.ID)
	if err != nil {
		// Log the error but continue
		fmt.Printf("Error getting user count: %v\n", err)
		userCount = 0
	}

	// Return response
	return &dto.DepartmentResponse{
		ID:          department.ID,
		Name:        department.Name,
		Code:        department.Code,
		Description: department.Description,
		IsActive:    department.IsActive,
		UserCount:   userCount,
	}, nil
}

// DeleteDepartment deletes a department
func (s *departmentService) DeleteDepartment(ctx context.Context, id int) error {
	// Check if department has users
	userCount, err := s.departmentRepo.GetUserCount(ctx, id)
	if err != nil {
		return fmt.Errorf("error checking department users: %w", err)
	}

	if userCount > 0 {
		return errors.New("cannot delete department with assigned users")
	}

	// Delete department
	if err := s.departmentRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("error deleting department: %w", err)
	}

	return nil
}

// GetAllDepartments gets all departments
func (s *departmentService) GetAllDepartments(ctx context.Context, limit, offset int) ([]*dto.DepartmentResponse, error) {
	departments, err := s.departmentRepo.List(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("error listing departments: %w", err)
	}

	// Convert to response DTOs
	response := make([]*dto.DepartmentResponse, 0, len(departments))
	for _, department := range departments {
		// Get user count
		userCount, err := s.departmentRepo.GetUserCount(ctx, department.ID)
		if err != nil {
			// Log the error but continue
			fmt.Printf("Error getting user count: %v\n", err)
			userCount = 0
		}

		response = append(response, &dto.DepartmentResponse{
			ID:          department.ID,
			Name:        department.Name,
			Code:        department.Code,
			Description: department.Description,
			IsActive:    department.IsActive,
			UserCount:   userCount,
		})
	}

	return response, nil
}

// CountDepartments gets the total number of departments
func (s *departmentService) CountDepartments(ctx context.Context) (int, error) {
	return s.departmentRepo.Count(ctx)
}
