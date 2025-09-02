package handlers

import (
	"bytes"
	"fmt"
	"log"
	"path/filepath"
	"time"

	"erp-excel/internal/dto"
	"erp-excel/internal/repository"
	"erp-excel/internal/service"
	"erp-excel/internal/utils"

	"github.com/gofiber/fiber/v2"
)

// ReportHandler handles report generation requests.
type ReportHandler struct {
	reportService service.ReportService
	reportRepo    repository.InventoryRepository // Keep if needed elsewhere, though service wraps it.
}

// NewReportHandler creates a new report handler.
func NewReportHandler(
	reportService service.ReportService,
	reportRepo repository.InventoryRepository,
) *ReportHandler {
	return &ReportHandler{
		reportService: reportService,
		reportRepo:    reportRepo,
	}
}

// GetInventoryReportData handles requests to get inventory report data (JSON response).
func (h *ReportHandler) GetInventoryReportData(c *fiber.Ctx) error {
	// Get user from context
	userID, ok := c.Locals("user_id").(int)
	if !ok || userID == 0 {
		return c.Status(fiber.StatusUnauthorized).JSON(utils.ErrorResponse(
			"Authentication required",
			"User not authenticated",
		))
	}

	departmentID, ok := c.Locals("department_id").(int)
	if !ok {
		departmentID = 0 // Default to all departments if not specified
	}

	// Parse request body
	var request dto.DateRangeRequest
	if err := c.BodyParser(&request); err != nil {
		log.Printf("Error parsing request body for inventory data: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse(
			"Invalid request",
			"Error parsing request body: "+err.Error(),
		))
	}

	// Validate request (using the enhanced ValidateStruct)
	if err := utils.ValidateStruct(&request); err != nil {
		log.Printf("Validation error for inventory data: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse(
			"Validation error",
			err.Error(),
		))
	}

	// Get report data from service
	items, err := h.reportService.GetInventoryReportData(c.Context(), userID, departmentID, &request)
	if err != nil {
		log.Printf("Error getting inventory report data: %v", err)
		// Provide more user-friendly messages for specific errors
		if err.Error() == "no data found to export for the specified date range" { // This exact message is from service
			return c.Status(fiber.StatusNotFound).JSON(utils.ErrorResponse(
				"No Data Found",
				"No data available for the selected period.",
			))
		}
		// Generic internal server error
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse(
			"Error retrieving report data",
			err.Error(),
		))
	}

	// Determine report title based on the request (period or specific dates)
	reportTitle := "Báo cáo tồn kho"
	if request.Period != nil && *request.Period != "" {
		reportTitle = fmt.Sprintf("Báo cáo tồn kho: %s", formatPeriod(*request.Period))
	} else if request.FromDate != nil && !request.FromDate.IsZero() && request.ToDate != nil && !request.ToDate.IsZero() {
		reportTitle = fmt.Sprintf("Báo cáo tồn kho từ %s đến %s",
			request.FromDate.Format("02/01/2006"),
			request.ToDate.Format("02/01/2006"),
		)
	}

	// Return response with data
	return c.Status(fiber.StatusOK).JSON(utils.SuccessResponse(
		dto.ReportDataResponse{
			ReportName:  reportTitle,
			GeneratedAt: time.Now(),
			Items:       items,
		},
		"Report data retrieved successfully",
	))
}

// formatPeriod helper for display purposes
func formatPeriod(period string) string {
	switch period {
	case "7days":
		return "7 ngày gần nhất"
	case "30days":
		return "30 ngày gần nhất"
	case "3months":
		return "3 tháng gần nhất"
	case "currentmonth":
		return "Tháng hiện tại"
	case "lastmonth":
		return "Tháng trước"
	default:
		return period
	}
}

// ExportInventoryReport handles requests to export inventory report to Excel.
func (h *ReportHandler) ExportInventoryReport(c *fiber.Ctx) error {
	// Get user from context
	userID, ok := c.Locals("user_id").(int)
	if !ok || userID == 0 {
		return c.Status(fiber.StatusUnauthorized).JSON(utils.ErrorResponse(
			"Authentication required",
			"User not authenticated",
		))
	}

	departmentID, ok := c.Locals("department_id").(int)
	if !ok {
		departmentID = 0 // Default to all departments if not specified
	}

	// Parse request body
	var request dto.DateRangeRequest
	if err := c.BodyParser(&request); err != nil {
		log.Printf("Error parsing request body for inventory export: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse(
			"Invalid request",
			"Error parsing request body: "+err.Error(),
		))
	}

	// Validate request
	if err := utils.ValidateStruct(&request); err != nil {
		log.Printf("Validation error for inventory export: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse(
			"Validation error",
			err.Error(),
		))
	}

	// Generate and export report file
	reportFileResponse, err := h.reportService.ExportInventoryReport(c.Context(), userID, departmentID, &request)
	if err != nil {
		log.Printf("Error exporting inventory report: %v", err)
		if err.Error() == "no data found to export for the specified date range" {
			return c.Status(fiber.StatusNotFound).JSON(utils.ErrorResponse(
				"No Data Found",
				"No data found for the specified date range to export.",
			))
		}
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse(
			"Error exporting report",
			err.Error(),
		))
	}

	c.Attachment(reportFileResponse.FileName)                         // Suggest filename for download
	return c.SendStream(reportFileResponse.FileDetal.(*bytes.Buffer)) // Stream the file content directly
}

// DownloadInventoryReport downloads an inventory report file.
func (h *ReportHandler) DownloadInventoryReport(c *fiber.Ctx) error {
	// Get filename from request parameters
	fileName := c.Params("fileName")
	if fileName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse(
			"Invalid request",
			"Filename is required",
		))
	}

	// Ensure no path manipulation by using filepath.Base
	fileName = filepath.Base(fileName)
	filePath := filepath.Join("public", "downloads", fileName) // Correct path structure

	// Check if file exists
	if !utils.FileExists(filePath) {
		return c.Status(fiber.StatusNotFound).JSON(utils.ErrorResponse(
			"File not found",
			"The requested file does not exist",
		))
	}

	// Return file for download
	return c.Download(filePath, fileName)
}

// SetupRoutes sets up the handler routes.
func (h *ReportHandler) SetupRoutes(router fiber.Router) {
	reports := router.Group("/reports")

	// Inventory reports
	reports.Post("/inventory", h.GetInventoryReportData)          // Returns JSON data for display
	reports.Post("/inventory/export", h.ExportInventoryReport)    // Generates Excel file and returns file info
	reports.Get("/download/:fileName", h.DownloadInventoryReport) // Downloads the specific file
}
