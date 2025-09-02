package middleware

import (
	"erp-excel/internal/service"
	"erp-excel/internal/utils"

	"github.com/gofiber/fiber/v2"
)

// RoleCheckMiddleware checks if user has the required role
func RoleCheckMiddleware(operationService service.OperationService) func(string) fiber.Handler {
	return func(operationCode string) fiber.Handler {
		return func(c *fiber.Ctx) error {
			// Get user ID from context
			userID, ok := c.Locals("user_id").(int)
			if !ok || userID == 0 {
				return c.Status(fiber.StatusUnauthorized).JSON(utils.ErrorResponse(
					"Authentication required",
					"User not authenticated",
				))
			}

			// Check if user has permission for the operation
			hasAccess, err := operationService.CheckUserAccess(c.Context(), userID, operationCode)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse(
					"Error checking permissions",
					err.Error(),
				))
			}

			if !hasAccess {
				return c.Status(fiber.StatusForbidden).JSON(utils.ErrorResponse(
					"Permission denied",
					"You don't have permission to perform this operation",
				))
			}

			// Continue to next handler
			return c.Next()
		}
	}
}
