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

type Assistant610Handler struct {
	BaseHandler
	assistant610Service service.Assistant610Service
	assistantRepo       repository.Assistant610Repository
}

// Corrected to match the field types
func NewAssistant610Handler(
	assistant610Service service.Assistant610Service,
	assistantRepo repository.Assistant610Repository,
) *Assistant610Handler {
	return &Assistant610Handler{
		assistant610Service: assistant610Service,
		assistantRepo:       assistantRepo,
	}
}

func (h *Assistant610Handler) GetAssistant610ReportData(c *fiber.Ctx) error {
	userID, _ := c.Locals("user_id").(int)
	departmentID, ok := c.Locals("department_id").(int)
	if !ok {
		departmentID = 0
	}

	var request dto.DateRangeRequest
	if err := c.BodyParser(&request); err != nil {
		log.Printf("Error parsing request body for inventory data: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse(
			"Invalid request",
			"Error parsing request body: "+err.Error(),
		))
	}

	if err := utils.ValidateStruct(&request); err != nil {
		log.Printf("Validation error for inventory data: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse(
			"Validation error",
			err.Error(),
		))
	}

	// Fixed method call to use assistant610Service
	items, err := h.assistant610Service.GetAssistant610ReportData(c.Context(), userID, departmentID, &request)
	if err != nil {
		log.Printf("Error getting inventory report data: %v", err)
		if err.Error() == "no data found to export for the specified date range" {
			return c.Status(fiber.StatusNotFound).JSON(utils.ErrorResponse(
				"No Data Found",
				"No data available for the selected period.",
			))
		}
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ErrorResponse(
			"Error retrieving report data",
			err.Error(),
		))
	}

	reportTitle := "Report "
	if request.Period != nil && *request.Period != "" {
		reportTitle = fmt.Sprintf("Report: %s", format610Period(*request.Period))
	} else if !request.FromDate.IsZero() && !request.ToDate.IsZero() {
		reportTitle = fmt.Sprintf("Report from %s to %s",
			request.FromDate.Format("02/01/2006"),
			request.ToDate.Format("02/01/2006"),
		)
	}

	return c.Status(fiber.StatusOK).JSON(utils.SuccessResponse(
		dto.Assistant610DataResponse{
			ReportName:  reportTitle,
			GeneratedAt: time.Now(),
			Items:       items,
		},
		"Report data retrieved successfully",
	))
}

func format610Period(period string) string {
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

func (h *Assistant610Handler) ExportAssistant610Report(c *fiber.Ctx) error {
	userID, _ := c.Locals("user_id").(int)
	departmentID, ok := c.Locals("department_id").(int)
	if !ok {
		departmentID = 0
	}

	var request dto.DateRangeRequest
	if err := c.BodyParser(&request); err != nil {
		log.Printf("Error parsing request body for inventory export: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse(
			"Invalid request",
			"Error parsing request body: "+err.Error(),
		))
	}

	if err := utils.ValidateStruct(&request); err != nil {
		log.Printf("Validation error for inventory export: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse(
			"Validation error",
			err.Error(),
		))
	}

	// Fixed method call to use assistant610Service
	reportFileResponse, err := h.assistant610Service.ExportAssistant610Report(c.Context(), userID, departmentID, &request)
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

	c.Attachment(reportFileResponse.FileName)
	return c.SendStream(reportFileResponse.FileDetal.(*bytes.Buffer)) // Ensure to check if FileDetal is not nil
}

func (h *Assistant610Handler) DownloadAssistant610Report(c *fiber.Ctx) error {
	fileName := c.Params("fileName")
	if fileName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse(
			"Invalid request",
			"Filename is required",
		))
	}

	fileName = filepath.Base(fileName)
	filePath := filepath.Join("public", "downloads", fileName)

	if !utils.FileExists(filePath) {
		return c.Status(fiber.StatusNotFound).JSON(utils.ErrorResponse(
			"File not found",
			"The requested file does not exist",
		))
	}

	return c.Download(filePath, fileName)
}

func (h *Assistant610Handler) SetupRoutes(router fiber.Router) {
	reports := router.Group("/assistants")

	reports.Post("/610", h.GetAssistant610ReportData) // Corrected to use correct method
	reports.Post("/610/export", h.ExportAssistant610Report)
	reports.Get("/download/:fileName", h.DownloadAssistant610Report)
}
