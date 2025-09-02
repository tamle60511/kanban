package app

import (
	"erp-excel/config"
	"erp-excel/database"
	"erp-excel/internal/handlers"
	"erp-excel/internal/middleware"
	"erp-excel/internal/repository"
	"erp-excel/internal/service"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

// App represents the application
type App struct {
	config *config.Config
	fiber  *fiber.App
	db     database.Database

	// Handlers
	handlers []handlers.BaseHandler // List of all handlers

	// Services
	authService service.AuthService

	// Repository
	userRepo       repository.UserRepository
	departmentRepo repository.DepartmentRepository
	roleRepo       repository.RoleRepository
	operationRepo  repository.OperationRepository
	reportRepo     repository.InventoryRepository
}

// New creates a new application instance
func New(cfg *config.Config, db database.Database) *App {
	app := &App{
		config: cfg,
		db:     db,
	}

	// Initialize Fiber
	app.fiber = fiber.New(fiber.Config{
		AppName:      cfg.Server.Name,
		ErrorHandler: errorHandler,
	})

	// Setup middleware
	app.fiber.Use(recover.New())
	app.fiber.Use(logger.New())
	app.fiber.Use(cors.New(cors.Config{
		AllowOrigins:     "*",
		AllowMethods:     "*",
		AllowHeaders:     "*",
		AllowCredentials: true,
	}))

	// Setup repositories
	app.userRepo = repository.NewUserRepository(app.db.DB())
	app.departmentRepo = repository.NewDepartmentRepository(app.db.DB())
	app.roleRepo = repository.NewRoleRepository(app.db.DB())
	app.operationRepo = repository.NewOperationRepository(app.db.DB())
	app.reportRepo = repository.NewInventoryRepository(app.db.ERPDatabase())

	// Setup services
	app.authService = service.NewAuthService(app.userRepo, app.config)
	userService := service.NewUserService(app.userRepo, app.departmentRepo, app.roleRepo, app.authService)
	departmentService := service.NewDepartmentService(app.departmentRepo)
	roleService := service.NewRoleService(app.roleRepo)
	operationService := service.NewOperationService(app.operationRepo, app.userRepo, app.roleRepo)
	reportService := service.NewReportService(
		app.db.ERPDatabase(),
		app.config,
		app.userRepo,
		app.operationRepo,
		app.reportRepo,
	)

	// Setup handlers
	authHandler := handlers.NewAuthHandler(app.authService)
	userHandler := handlers.NewUserHandler(userService)
	departmentHandler := handlers.NewDepartmentHandler(departmentService)
	roleHandler := handlers.NewRoleHandler(roleService)
	reportHandler := handlers.NewReportHandler(reportService, app.reportRepo)
	operationHandler := handlers.NewOperationHandler(operationService)
	adminHandler := handlers.NewAdminHandler(userService, departmentService, roleService, operationService)

	// Store handlers
	app.handlers = []handlers.BaseHandler{
		authHandler,
		userHandler,
		departmentHandler,
		roleHandler,
		reportHandler,
		adminHandler,
		operationHandler,
	}

	return app
}

// SetupRoutes configures the application routes
func (a *App) SetupRoutes() {
	// Health check endpoint
	a.fiber.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "ok",
			"name":   a.config.Server.Name,
			"env":    a.config.Server.Env,
		})
	})

	// API routes
	api := a.fiber.Group("/api")

	// white list routes
	whitelist := []string{
		"/api/auth/login",
	}

	// Protected routes
	protected := api.Group("/", middleware.JWTMiddleware(a.authService, whitelist))

	// Setup all handler routes
	for _, handler := range a.handlers {
		handler.SetupRoutes(protected)
	}

	// 404 handler
	a.fiber.Use(func(c *fiber.Ctx) error {
		return c.Status(404).JSON(fiber.Map{
			"success": false,
			"message": "Not Found",
			"error":   "The requested resource does not exist",
		})
	})
}

// Start starts the application
func (a *App) Start() {
	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start server in a goroutine
	go func() {
		addr := fmt.Sprintf(":%s", a.config.Server.Port)
		if err := a.fiber.Listen(addr); err != nil {
			log.Fatalf("Error starting server: %v", err)
		}
	}()

	log.Printf("Server started on port %s", a.config.Server.Port)

	// Wait for interrupt signal
	<-sigChan
	log.Println("Shutting down server...")

	// Close database connection
	if err := a.db.Close(); err != nil {
		log.Printf("Error closing database connection: %v", err)
	}

	// Shutdown server
	if err := a.fiber.Shutdown(); err != nil {
		log.Fatalf("Error shutting down server: %v", err)
	}

	log.Println("Server gracefully stopped")
}

// errorHandler handles API errors
func errorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	message := "Internal Server Error"

	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
		message = e.Message
	}

	return c.Status(code).JSON(fiber.Map{
		"success": false,
		"message": message,
		"error":   err.Error(),
	})
}
