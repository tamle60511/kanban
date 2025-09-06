package handlers

import (
	"erp-excel/internal/service"
	"erp-excel/internal/utils"

	"github.com/gofiber/fiber/v2"
)

// AdminHandler handles admin operations
type AdminHandler struct {
	BaseHandler // Embedding BaseHandler

	userService       service.UserService
	departmentService service.DepartmentService
	roleService       service.RoleService
	operationService  service.OperationService
}

// NewAdminHandler creates a new admin handler
func NewAdminHandler(
	userService service.UserService,
	departmentService service.DepartmentService,
	roleService service.RoleService,
	operationService service.OperationService,
) *AdminHandler {
	return &AdminHandler{
		userService:       userService,
		departmentService: departmentService,
		roleService:       roleService,
		operationService:  operationService,
	}
}

// Dashboard returns admin dashboard statistics
func (h *AdminHandler) Dashboard(c *fiber.Ctx) error {
	userCount, err := h.userService.CountUsers(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse(
			"Error getting user count",
			err.Error(),
		))
	}

	deptCount, err := h.departmentService.CountDepartments(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse(
			"Error getting department count",
			err.Error(),
		))
	}

	roleCount, err := h.roleService.CountRoles(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse(
			"Error getting role count",
			err.Error(),
		))
	}

	// Get recent access logs
	logs, err := h.operationService.GetRecentLogs(c.Context(), 10)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse(
			"Error getting recent logs",
			err.Error(),
		))
	}

	return c.Status(fiber.StatusOK).JSON(utils.SuccessResponse(
		fiber.Map{
			"user_count":       userCount,
			"department_count": deptCount,
			"role_count":       roleCount,
			"recent_logs":      logs,
		},
		"Dashboard data retrieved successfully",
	))
}

// GetSystemOperations gets all system operations
func (h *AdminHandler) GetSystemOperations(c *fiber.Ctx) error {
	operations, err := h.operationService.GetAllOperations(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse(
			"Error getting operations",
			err.Error(),
		))
	}

	return c.Status(fiber.StatusOK).JSON(utils.SuccessResponse(
		operations,
		"Operations retrieved successfully",
	))
}

// SetupRoutes sets up the handler routes
func (h *AdminHandler) SetupRoutes(router fiber.Router) {
	admin := router.Group("/admin")

	admin.Get("/dashboard", h.Dashboard)
	admin.Get("/operations", h.GetSystemOperations)
}
