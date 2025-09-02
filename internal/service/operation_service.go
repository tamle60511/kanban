package service

import (
	"context"
	"encoding/json"
	"erp-excel/internal/dto"
	"erp-excel/internal/models"
	"erp-excel/internal/repository"
	"fmt"
	"time"
)

// OperationService interface
type OperationService interface {
	GetAllOperations(ctx context.Context) ([]*dto.OperationResponse, error)
	CheckUserAccess(ctx context.Context, userID int, operationCode string) (bool, error)
	LogAccess(ctx context.Context, userID int, operationCode string, params interface{}, ipAddress string) (int, error)
	UpdateLogStatus(ctx context.Context, logID int, status string) (bool, error)
	GetRecentLogs(ctx context.Context, limit int) ([]*models.AccessLog, error)
}

type operationService struct {
	operationRepo repository.OperationRepository
	userRepo      repository.UserRepository
	roleRepo      repository.RoleRepository
}

// NewOperationService creates a new operation service
func NewOperationService(
	operationRepo repository.OperationRepository,
	userRepo repository.UserRepository,
	roleRepo repository.RoleRepository,
) OperationService {
	return &operationService{
		operationRepo: operationRepo,
		userRepo:      userRepo,
		roleRepo:      roleRepo,
	}
}

// GetAllOperations gets all operations
func (s *operationService) GetAllOperations(ctx context.Context) ([]*dto.OperationResponse, error) {
	return s.operationRepo.GetAll(ctx)
}

// CheckUserAccess checks if a user has access to an operation
func (s *operationService) CheckUserAccess(ctx context.Context, userID int, operationCode string) (bool, error) {
	// Find operation by code
	operation, err := s.operationRepo.FindByCode(ctx, operationCode)
	if err != nil {
		return false, fmt.Errorf("error finding operation: %w", err)
	}

	// Check if user has access
	return s.roleRepo.CheckUserOperationAccess(ctx, userID, operation.ID)
}

// LogAccess logs access to an operation
func (s *operationService) LogAccess(
	ctx context.Context,
	userID int,
	operationCode string,
	params interface{},
	ipAddress string,
) (int, error) {
	// Find operation by code
	operation, err := s.operationRepo.FindByCode(ctx, operationCode)
	if err != nil {
		return 0, fmt.Errorf("error finding operation: %w", err)
	}

	// Create access log
	log := &models.AccessLog{
		UserID:      userID,
		OperationID: operation.ID,
		AccessTime:  time.Now(),
		IPAddress:   ipAddress,
		Status:      "pending",
	}

	// Convert params to JSON string if provided
	if params != nil {
		jsonStr, err := json.Marshal(params)
		if err != nil {
			return 0, fmt.Errorf("error marshalling params: %w", err)
		}
		log.SearchParams = string(jsonStr)
	}

	// Save to database
	return s.operationRepo.LogAccess(ctx, log)
}

// UpdateLogStatus updates the status of an access log
func (s *operationService) UpdateLogStatus(ctx context.Context, logID int, status string) (bool, error) {
	if logID <= 0 {
		return false, fmt.Errorf("invalid log ID: %d", logID)
	}

	return s.operationRepo.UpdateLogStatus(ctx, logID, status)
}

// GetRecentLogs gets recent access logs
func (s *operationService) GetRecentLogs(ctx context.Context, limit int) ([]*models.AccessLog, error) {
	if limit <= 0 {
		limit = 10
	} else if limit > 100 {
		limit = 100
	}

	return s.operationRepo.GetRecentLogs(ctx, limit)
}
