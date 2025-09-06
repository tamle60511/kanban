package handlers

import (
	"erp-excel/internal/dto"
	"erp-excel/internal/service"
	"erp-excel/internal/utils"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

// RoleHandler handles role operations
type RoleHandler struct {
	BaseHandler // Embedding BaseHandler

	roleService service.RoleService
}

// NewRoleHandler creates a new role handler
func NewRoleHandler(roleService service.RoleService) *RoleHandler {
	return &RoleHandler{
		roleService: roleService,
	}
}

// GetAll retrieves all roles
func (h *RoleHandler) GetAll(c *fiber.Ctx) error {
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

	// Get roles
	roles, err := h.roleService.GetAllRoles(c.Context(), limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse(
			"Error retrieving roles",
			err.Error(),
		))
	}

	// Get total count for pagination
	total, err := h.roleService.CountRoles(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse(
			"Error counting roles",
			err.Error(),
		))
	}

	// Calculate pagination info
	totalPages := (total + limit - 1) / limit
	hasNext := page < totalPages
	hasPrev := page > 1

	return c.Status(fiber.StatusOK).JSON(utils.SuccessResponse(
		fiber.Map{
			"roles": roles,
			"pagination": fiber.Map{
				"total":       total,
				"page":        page,
				"limit":       limit,
				"total_pages": totalPages,
				"has_next":    hasNext,
				"has_prev":    hasPrev,
			},
		},
		"Roles retrieved successfully",
	))
}

// GetByID retrieves a role by ID
func (h *RoleHandler) GetByID(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse(
			"Invalid role ID",
			"Role ID must be a number",
		))
	}

	role, err := h.roleService.GetRoleByID(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(utils.ErrorResponse(
			"Role not found",
			err.Error(),
		))
	}

	return c.Status(fiber.StatusOK).JSON(utils.SuccessResponse(
		role,
		"Role retrieved successfully",
	))
}

// Create creates a new role
func (h *RoleHandler) Create(c *fiber.Ctx) error {
	var request dto.CreateRoleRequest
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

	// Create role
	role, err := h.roleService.CreateRole(c.Context(), request)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse(
			"Error creating role",
			err.Error(),
		))
	}

	return c.Status(fiber.StatusCreated).JSON(utils.SuccessResponse(
		role,
		"Role created successfully",
	))
}

// Update updates a role
func (h *RoleHandler) Update(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse(
			"Invalid role ID",
			"Role ID must be a number",
		))
	}

	var request dto.UpdateRoleRequest
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

	// Update role
	role, err := h.roleService.UpdateRole(c.Context(), id, request)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse(
			"Error updating role",
			err.Error(),
		))
	}

	return c.Status(fiber.StatusOK).JSON(utils.SuccessResponse(
		role,
		"Role updated successfully",
	))
}

// Delete deletes a role
func (h *RoleHandler) Delete(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse(
			"Invalid role ID",
			"Role ID must be a number",
		))
	}

	if err := h.roleService.DeleteRole(c.Context(), id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse(
			"Error deleting role",
			err.Error(),
		))
	}

	return c.Status(fiber.StatusOK).JSON(utils.SuccessResponse(
		nil,
		"Role deleted successfully",
	))
}

// SetupRoutes sets up the handler routes
func (h *RoleHandler) SetupRoutes(router fiber.Router) {
	roles := router.Group("/roles")

	roles.Get("/", h.GetAll)
	roles.Get("/:id", h.GetByID)
	roles.Post("/", h.Create)
	roles.Put("/:id", h.Update)
	roles.Delete("/:id", h.Delete)
}
