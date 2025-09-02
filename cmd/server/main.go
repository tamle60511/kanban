package main

import (
	"erp-excel/config"
	"erp-excel/database"
	"erp-excel/internal/app"
)

func main() {
	// Load configuration
	cfg := config.MustConfig()

	// Connect to database
	db := database.MustDatabase(cfg)

	// Create application
	application := app.New(cfg, db)

	// Setup routes
	application.SetupRoutes()

	// Start application
	application.Start()
}
