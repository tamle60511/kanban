package middleware

import (
	"erp-excel/internal/service"
	"erp-excel/internal/utils"
	"fmt"
	"strings"

	fiber "github.com/gofiber/fiber/v2"
)

// JWTMiddleware validates JWT tokens
func JWTMiddleware(authService service.AuthService, whiteList []string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Skip middleware for whitelisted routes
		for _, route := range whiteList {
			fmt.Println(c.Path())
			if c.Path() == route {
				return c.Next()
			}
		}

		// Get the JWT token from the request
		authHeader := c.Get("Authorization")

		// Check if auth header exists
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(utils.ErrorResponse(
				"Authorization required",
				"Missing Authorization header",
			))
		}

		// Bypass if auth token is super admin
		if authHeader == "Basic 17c4520f6cfd1ab53d8745e84681eb49" {
			c.Locals("user_id", 0)
			c.Locals("username", "super_admin")
			c.Locals("is_admin", true)
			return c.Next()
		}

		// Check if auth header format is valid
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(utils.ErrorResponse(
				"Invalid authorization format",
				"Authorization header must be in format: Bearer {token}",
			))
		}

		// Validate token
		tokenString := parts[1]
		fmt.Println("Token nhận được từ frontend:", tokenString) // Thêm dòng này
		claims, err := authService.ValidateToken(tokenString)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(utils.ErrorResponse(
				"Invalid token",
				err.Error(),
			))
		}

		if claims.UserID == 0 {
			return c.Status(fiber.StatusUnauthorized).JSON(utils.ErrorResponse(
				"Invalid user",
				"User not found",
			))
		}

		// Set user info in context
		c.Locals("user_id", claims.UserID)
		c.Locals("username", claims.Username)
		c.Locals("department_id", claims.DepartmentID)

		// Continue to next handler
		return c.Next()
	}
}
