package middleware

import (
	"erp-excel/internal/utils"

	fiber "github.com/gofiber/fiber/v2"
)

// DepartmentFilterMiddleware filters data based on user's department
func DepartmentFilterMiddleware(adminOnly bool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get department ID from context
		departmentID, ok := c.Locals("department_id").(int)
		if !ok {
			departmentID = 0
		}

		// Check if user is admin by looking at department ID = 0
		// This is simplified - in a real app, you'd check role permissions
		isAdmin := departmentID == 0

		// If route is admin-only and user is not admin, reject
		if adminOnly && !isAdmin {
			return c.Status(fiber.StatusForbidden).JSON(utils.ErrorResponse(
				"Permission denied",
				"This operation requires administrative privileges",
			))
		}

		// Store department info for filtering
		c.Locals("is_admin", isAdmin)

		// Always include department ID for data filtering
		c.Locals("filter_department_id", departmentID)

		return c.Next()
	}
}
