package handlers

import (
	"erp-excel/internal/dto"
	"erp-excel/internal/service"
	"erp-excel/internal/utils"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

// UserHandler handles user operations
type UserHandler struct {
	BaseHandler // Embedding BaseHandler

	userService service.UserService
}

// NewUserHandler creates a new user handler
func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// GetAll retrieves all users
func (h *UserHandler) GetAll(c *fiber.Ctx) error {
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

	// Get users
	users, err := h.userService.GetAllUsers(c.Context(), limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse(
			"Error retrieving users",
			err.Error(),
		))
	}

	// Get total count for pagination
	total, err := h.userService.CountUsers(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse(
			"Error counting users",
			err.Error(),
		))
	}

	// Calculate pagination info
	totalPages := (total + limit - 1) / limit
	hasNext := page < totalPages
	hasPrev := page > 1

	return c.Status(fiber.StatusOK).JSON(utils.SuccessResponse(
		fiber.Map{
			"users": users,
			"pagination": fiber.Map{
				"total":       total,
				"page":        page,
				"limit":       limit,
				"total_pages": totalPages,
				"has_next":    hasNext,
				"has_prev":    hasPrev,
			},
		},
		"Users retrieved successfully",
	))
}

// GetByID retrieves a user by ID
func (h *UserHandler) GetByID(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse(
			"Invalid user ID",
			"User ID must be a number",
		))
	}

	user, err := h.userService.GetUserByID(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(utils.ErrorResponse(
			"User not found",
			err.Error(),
		))
	}

	return c.Status(fiber.StatusOK).JSON(utils.SuccessResponse(
		user,
		"User retrieved successfully",
	))
}

// Create creates a new user
func (h *UserHandler) Create(c *fiber.Ctx) error {
	var request dto.CreateUserRequest
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

	// Create user
	user, err := h.userService.CreateUser(c.Context(), request)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse(
			"Error creating user",
			err.Error(),
		))
	}

	return c.Status(fiber.StatusCreated).JSON(utils.SuccessResponse(
		user,
		"User created successfully",
	))
}

// Update updates a user
func (h *UserHandler) Update(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse(
			"Invalid user ID",
			"User ID must be a number",
		))
	}

	var request dto.UpdateUserRequest
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

	// Update user
	user, err := h.userService.UpdateUser(c.Context(), id, request)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse(
			"Error updating user",
			err.Error(),
		))
	}

	return c.Status(fiber.StatusOK).JSON(utils.SuccessResponse(
		user,
		"User updated successfully",
	))
}

// UpdatePassword updates a user's password
func (h *UserHandler) UpdatePassword(c *fiber.Ctx) error {
	var (
		userID int
		ok     bool
	)
	isAdmin, _ := c.Locals("is_admin").(bool)
	if !isAdmin {
		// Get current user ID
		userID, ok = c.Locals("user_id").(int)
		if !ok || userID == 0 {
			return c.Status(fiber.StatusUnauthorized).JSON(utils.ErrorResponse(
				"Authentication required",
				"User not authenticated",
			))
		}
	}

	var request dto.UpdatePasswordRequest
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

	// Update password
	if err := h.userService.UpdateUserPassword(c.Context(), userID, request); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse(
			"Error updating password",
			err.Error(),
		))
	}

	return c.Status(fiber.StatusOK).JSON(utils.SuccessResponse(
		nil,
		"Password updated successfully",
	))
}

// Delete deactivates a user
func (h *UserHandler) Delete(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse(
			"Invalid user ID",
			"User ID must be a number",
		))
	}

	if err := h.userService.DeleteUser(c.Context(), id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse(
			"Error deleting user",
			err.Error(),
		))
	}

	return c.Status(fiber.StatusOK).JSON(utils.SuccessResponse(
		nil,
		"User deleted successfully",
	))
}

// AssignRoles assigns roles to a user
func (h *UserHandler) AssignRoles(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse(
			"Invalid user ID",
			"User ID must be a number",
		))
	}

	var request struct {
		RoleIDs []int `json:"role_ids" validate:"required,min=1,dive,min=1"`
	}

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

	// Assign roles
	if err := h.userService.AssignRolesToUser(c.Context(), id, request.RoleIDs); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse(
			"Error assigning roles",
			err.Error(),
		))
	}

	return c.Status(fiber.StatusOK).JSON(utils.SuccessResponse(
		nil,
		"Roles assigned successfully",
	))
}

// SetupRoutes sets up the handler routes
func (h *UserHandler) SetupRoutes(router fiber.Router) {
	users := router.Group("/users")

	users.Get("/", h.GetAll)
	users.Get("/:id", h.GetByID)
	users.Post("/", h.Create)
	users.Put("/:id", h.Update)
	users.Delete("/:id", h.Delete)
	users.Post("/:id/roles", h.AssignRoles)
	users.Post("/password", h.UpdatePassword)
}
