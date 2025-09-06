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

type ReportHandler struct {
	BaseHandler

	reportService service.ReportService
	reportRepo    repository.InventoryRepository
}

func NewReportHandler(
	reportService service.ReportService,
	reportRepo repository.InventoryRepository,
) *ReportHandler {
	return &ReportHandler{
		reportService: reportService,
		reportRepo:    reportRepo,
	}
}

func (h *ReportHandler) GetInventoryReportData(c *fiber.Ctx) error {
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

	items, err := h.reportService.GetInventoryReportData(c.Context(), userID, departmentID, &request)
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
		reportTitle = fmt.Sprintf("Report: %s", formatPeriod(*request.Period))
	} else if request.FromDate != nil && !request.FromDate.IsZero() && request.ToDate != nil && !request.ToDate.IsZero() {
		reportTitle = fmt.Sprintf("Report from %s to %s",
			request.FromDate.Format("02/01/2006"),
			request.ToDate.Format("02/01/2006"),
		)
	}

	return c.Status(fiber.StatusOK).JSON(utils.SuccessResponse(
		dto.ReportDataResponse{
			ReportName:  reportTitle,
			GeneratedAt: time.Now(),
			Items:       items,
		},
		"Report data retrieved successfully",
	))
}

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

func (h *ReportHandler) ExportInventoryReport(c *fiber.Ctx) error {
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

	c.Attachment(reportFileResponse.FileName)
	return c.SendStream(reportFileResponse.FileDetal.(*bytes.Buffer))
}

func (h *ReportHandler) DownloadInventoryReport(c *fiber.Ctx) error {

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

func (h *ReportHandler) SetupRoutes(router fiber.Router) {
	reports := router.Group("/reports")

	reports.Post("/inventory", h.GetInventoryReportData)
	reports.Post("/inventory/export", h.ExportInventoryReport)
	reports.Get("/download/:fileName", h.DownloadInventoryReport)
}
