package handlers

import (
	"erp-excel/internal/dto"
	"erp-excel/internal/service"
	"erp-excel/internal/utils"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

// DepartmentHandler handles department operations
type DepartmentHandler struct {
	departmentService service.DepartmentService
}

// NewDepartmentHandler creates a new department handler
func NewDepartmentHandler(departmentService service.DepartmentService) *DepartmentHandler {
	return &DepartmentHandler{
		departmentService: departmentService,
	}
}

// GetAll retrieves all departments
func (h *DepartmentHandler) GetAll(c *fiber.Ctx) error {
	// Parse pagination parameters
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))

	// Handle invalid pagination
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	// Calculate offset
	offset := (page - 1) * limit

	// Get departments
	departments, err := h.departmentService.GetAllDepartments(c.Context(), limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse(
			"Error retrieving departments",
			err.Error(),
		))
	}

	// Get total count for pagination
	total, err := h.departmentService.CountDepartments(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse(
			"Error counting departments",
			err.Error(),
		))
	}

	// Calculate pagination info
	totalPages := (total + limit - 1) / limit
	hasNext := page < totalPages
	hasPrev := page > 1

	return c.Status(fiber.StatusOK).JSON(utils.SuccessResponse(
		fiber.Map{
			"departments": departments,
			"pagination": fiber.Map{
				"total":       total,
				"page":        page,
				"limit":       limit,
				"total_pages": totalPages,
				"has_next":    hasNext,
				"has_prev":    hasPrev,
			},
		},
		"Departments retrieved successfully",
	))
}

// GetByID retrieves a department by ID
func (h *DepartmentHandler) GetByID(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse(
			"Invalid department ID",
			"Department ID must be a number",
		))
	}

	department, err := h.departmentService.GetDepartmentByID(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(utils.ErrorResponse(
			"Department not found",
			err.Error(),
		))
	}

	return c.Status(fiber.StatusOK).JSON(utils.SuccessResponse(
		department,
		"Department retrieved successfully",
	))
}

// Create creates a new department
func (h *DepartmentHandler) Create(c *fiber.Ctx) error {
	var request dto.CreateDepartmentRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse(
			"Invalid request",
			"Error parsing request body",
		))
	}

	// Validate request
	if err := utils.ValidateStruct(request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse(
			"Validation error",
			err.Error(),
		))
	}

	// Create department
	department, err := h.departmentService.CreateDepartment(c.Context(), request)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse(
			"Error creating department",
			err.Error(),
		))
	}

	return c.Status(fiber.StatusCreated).JSON(utils.SuccessResponse(
		department,
		"Department created successfully",
	))
}

// Update updates a department
func (h *DepartmentHandler) Update(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse(
			"Invalid department ID",
			"Department ID must be a number",
		))
	}

	var request dto.UpdateDepartmentRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse(
			"Invalid request",
			"Error parsing request body",
		))
	}

	// Validate request
	if err := utils.ValidateStruct(request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse(
			"Validation error",
			err.Error(),
		))
	}

	// Update department
	department, err := h.departmentService.UpdateDepartment(c.Context(), id, request)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse(
			"Error updating department",
			err.Error(),
		))
	}

	return c.Status(fiber.StatusOK).JSON(utils.SuccessResponse(
		department,
		"Department updated successfully",
	))
}

// Delete deactivates a department
func (h *DepartmentHandler) Delete(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse(
			"Invalid department ID",
			"Department ID must be a number",
		))
	}

	if err := h.departmentService.DeleteDepartment(c.Context(), id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse(
			"Error deleting department",
			err.Error(),
		))
	}

	return c.Status(fiber.StatusOK).JSON(utils.SuccessResponse(
		nil,
		"Department deleted successfully",
	))
}

// SetupRoutes sets up the handler routes
func (h *DepartmentHandler) SetupRoutes(router fiber.Router) {
	departments := router.Group("/departments")

	departments.Get("/", h.GetAll)
	departments.Get("/:id", h.GetByID)
	departments.Post("/", h.Create)
	departments.Put("/:id", h.Update)
	departments.Delete("/:id", h.Delete)
}
