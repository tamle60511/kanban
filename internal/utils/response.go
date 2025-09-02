package utils

import "github.com/gofiber/fiber/v2"

// SuccessResponse returns a standardized success response
func SuccessResponse(data interface{}, message string) fiber.Map {
	return fiber.Map{
		"success": true,
		"message": message,
		"data":    data,
	}
}

// ErrorResponse returns a standardized error response
func ErrorResponse(message string, error string) fiber.Map {
	return fiber.Map{
		"success": false,
		"message": message,
		"error":   error,
	}
}

// PaginatedResponse returns a response with pagination metadata
func PaginatedResponse(data interface{}, page, limit, total int, message string) fiber.Map {
	totalPages := (total + limit - 1) / limit
	hasNext := page < totalPages
	hasPrev := page > 1

	return fiber.Map{
		"success": true,
		"message": message,
		"data":    data,
		"pagination": fiber.Map{
			"total":       total,
			"page":        page,
			"limit":       limit,
			"total_pages": totalPages,
			"has_next":    hasNext,
			"has_prev":    hasPrev,
		},
	}
}
