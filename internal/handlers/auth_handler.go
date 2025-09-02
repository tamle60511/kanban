package handlers

import (
	"erp-excel/internal/dto"
	"erp-excel/internal/service"
	"erp-excel/internal/utils"

	"github.com/gofiber/fiber/v2"
)

// AuthHandler handles authentication requests
type AuthHandler struct {
	BaseHandler // Embedding BaseHandler

	authService service.AuthService
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// Login handles user login
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var request dto.LoginRequest
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

	// Attempt login
	response, err := h.authService.Login(c.Context(), request)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(utils.ErrorResponse(
			"Login failed",
			err.Error(),
		))
	}

	return c.Status(fiber.StatusOK).JSON(utils.SuccessResponse(
		response,
		"Login successful",
	))
}

// GetProfile retrieves the current user's profile
func (h *AuthHandler) GetProfile(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(int)
	if !ok || userID == 0 {
		return c.Status(fiber.StatusUnauthorized).JSON(utils.ErrorResponse(
			"Authentication required",
			"User not authenticated",
		))
	}

	profile, err := h.authService.GetUserProfile(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse(
			"Error retrieving profile",
			err.Error(),
		))
	}

	return c.Status(fiber.StatusOK).JSON(utils.SuccessResponse(
		profile,
		"Profile retrieved successfully",
	))
}

// SetupRoutes sets up the handler routes
func (h *AuthHandler) SetupRoutes(router fiber.Router) {
	auth := router.Group("/auth")

	auth.Post("/login", h.Login)
	auth.Get("/profile", h.GetProfile)
}
