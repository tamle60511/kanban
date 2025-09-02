package handlers

import "github.com/gofiber/fiber/v2"

type BaseHandler interface {
	SetupRoutes(router fiber.Router)
}
