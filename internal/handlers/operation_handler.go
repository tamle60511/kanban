package handlers

import (
	"erp-excel/internal/service"
	"erp-excel/internal/utils"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

// OperationHandler handles operation-related HTTP requests
type OperationHandler struct {
	BaseHandler // Embedding BaseHandler

	operationService service.OperationService
}

// NewOperationHandler creates a new operation handler
func NewOperationHandler(operationService service.OperationService) *OperationHandler {
	return &OperationHandler{
		operationService: operationService,
	}
}

// GetAllOperations retrieves all operations
func (h *OperationHandler) GetAllOperations(c *fiber.Ctx) error {
	operations, err := h.operationService.GetAllOperations(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse(
			"Error retrieving operations",
			err.Error(),
		))
	}

	return c.Status(fiber.StatusOK).JSON(utils.SuccessResponse(
		operations,
		"Operations retrieved successfully",
	))
}

// CheckUserAccess checks if a user has access to a specific operation
func (h *OperationHandler) CheckUserAccess(c *fiber.Ctx) error {
	// Parse user ID from request
	userID, err := strconv.Atoi(c.Params("userID"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse(
			"Invalid user ID",
			"User ID must be a number",
		))
	}

	// Get operation code from request
	operationCode := c.Params("operationCode")
	if operationCode == "" {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse(
			"Invalid operation code",
			"Operation code cannot be empty",
		))
	}

	// Check user access
	hasAccess, err := h.operationService.CheckUserAccess(c.Context(), userID, operationCode)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse(
			"Error checking user access",
			err.Error(),
		))
	}

	return c.Status(fiber.StatusOK).JSON(utils.SuccessResponse(
		fiber.Map{
			"has_access": hasAccess,
		},
		"User access checked successfully",
	))
}

// LogAccess logs access to an operation
func (h *OperationHandler) LogAccess(c *fiber.Ctx) error {
	// Parse request body
	var requestBody struct {
		UserID        int         `json:"user_id"`
		OperationCode string      `json:"operation_code"`
		Params        interface{} `json:"params,omitempty"`
	}

	if err := c.BodyParser(&requestBody); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse(
			"Invalid request body",
			"Error parsing request body",
		))
	}

	// Validate input
	if requestBody.UserID <= 0 || requestBody.OperationCode == "" {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse(
			"Invalid input",
			"User ID and operation code are required",
		))
	}

	// Get client IP address
	ipAddress := c.IP()

	// Log access
	logID, err := h.operationService.LogAccess(
		c.Context(),
		requestBody.UserID,
		requestBody.OperationCode,
		requestBody.Params,
		ipAddress,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse(
			"Error logging access",
			err.Error(),
		))
	}

	return c.Status(fiber.StatusCreated).JSON(utils.SuccessResponse(
		fiber.Map{
			"log_id": logID,
		},
		"Access logged successfully",
	))
}

// UpdateLogStatus updates the status of an access log
func (h *OperationHandler) UpdateLogStatus(c *fiber.Ctx) error {
	// Parse log ID from URL parameter
	logID, err := strconv.Atoi(c.Params("logID"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse(
			"Invalid log ID",
			"Log ID must be a number",
		))
	}

	// Parse request body
	var requestBody struct {
		Status string `json:"status"`
	}

	if err := c.BodyParser(&requestBody); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse(
			"Invalid request body",
			"Error parsing request body",
		))
	}

	// Validate status
	if requestBody.Status == "" {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse(
			"Invalid status",
			"Status cannot be empty",
		))
	}

	// Update log status
	updated, err := h.operationService.UpdateLogStatus(c.Context(), logID, requestBody.Status)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse(
			"Error updating log status",
			err.Error(),
		))
	}

	if !updated {
		return c.Status(fiber.StatusNotFound).JSON(utils.ErrorResponse(
			"Log not found",
			"No log found with the given ID",
		))
	}

	return c.Status(fiber.StatusOK).JSON(utils.SuccessResponse(
		nil,
		"Log status updated successfully",
	))
}

// GetRecentLogs retrieves recent access logs
func (h *OperationHandler) GetRecentLogs(c *fiber.Ctx) error {
	// Parse limit from query parameter
	limit, _ := strconv.Atoi(c.Query("limit", "10"))

	// Get recent logs
	logs, err := h.operationService.GetRecentLogs(c.Context(), limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse(
			"Error retrieving recent logs",
			err.Error(),
		))
	}

	return c.Status(fiber.StatusOK).JSON(utils.SuccessResponse(
		logs,
		"Recent logs retrieved successfully",
	))
}

// SetupRoutes sets up the routes for operation-related endpoints
func (h *OperationHandler) SetupRoutes(router fiber.Router) {
	operations := router.Group("/operations")

	// Get all operations
	operations.Get("/", h.GetAllOperations)

	// Check user access to an operation
	operations.Get("/access/:userID/:operationCode", h.CheckUserAccess)

	// Log access to an operation
	operations.Post("/log", h.LogAccess)

	// Update log status
	operations.Put("/log/:logID/status", h.UpdateLogStatus)

	// Get recent logs
	operations.Get("/logs/recent", h.GetRecentLogs)
}
