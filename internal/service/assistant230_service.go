package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"erp-excel/config"
	"erp-excel/internal/dto"
	"erp-excel/internal/models"
	"erp-excel/internal/repository"
	"erp-excel/internal/utils"
	"errors"
	"fmt"
	"log"
	"path/filepath"
	"time"
)

// ReportService interface defines methods for report generation.
type ReportService interface {
	GetInventoryReportData(ctx context.Context, userID int, departmentID int, request *dto.DateRangeRequest) ([]dto.Asisstant230ReportItem, error)
	ExportInventoryReport(ctx context.Context, userID int, departmentID int, request *dto.DateRangeRequest) (*dto.ReportFileResponse, error)
}

type reportService struct {
	erpDB         *sql.DB
	config        *config.Config
	userRepo      repository.UserRepository
	operationRepo repository.OperationRepository
	inventoryRepo repository.InventoryRepository
}

// NewReportService creates a new report service.
func NewReportService(
	erpDB *sql.DB,
	config *config.Config,
	userRepo repository.UserRepository,
	operationRepo repository.OperationRepository,
	inventoryRepo repository.InventoryRepository,
) ReportService {
	return &reportService{
		erpDB:         erpDB,
		config:        config,
		userRepo:      userRepo,
		operationRepo: operationRepo,
		inventoryRepo: inventoryRepo,
	}
}

// resolveDateRange calculates actual fromDate and toDate based on Period or uses provided dates.
func (s *reportService) resolveDateRange(request *dto.DateRangeRequest) (time.Time, time.Time, error) {
	log.Printf("resolveDateRange called with request: %+v", request)

	now := time.Now()
	currentEndOfDay := now.Truncate(24 * time.Hour).Add(24*time.Hour - time.Nanosecond)
	fromDate := time.Time{}
	toDate := time.Time{}

	if request.Period != nil && *request.Period != "" {
		period := *request.Period
		log.Printf("Using period: %s", period)
		switch period {
		case "7days":
			fromDate = currentEndOfDay.AddDate(0, 0, -6).Truncate(24 * time.Hour)
			toDate = currentEndOfDay
		case "30days":
			fromDate = currentEndOfDay.AddDate(0, 0, -29).Truncate(24 * time.Hour)
			toDate = currentEndOfDay
		case "3months":
			fromDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).AddDate(0, -2, 0).Truncate(24 * time.Hour)
			toDate = currentEndOfDay
		case "currentmonth":
			fromDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
			toDate = currentEndOfDay
		case "lastmonth":
			firstOfThisMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
			toDate = firstOfThisMonth.Add(-time.Nanosecond)
			fromDate = time.Date(toDate.Year(), toDate.Month(), 1, 0, 0, 0, 0, now.Location())
		default:
			return time.Time{}, time.Time{}, fmt.Errorf("invalid period specified: %s", period)
		}
	} else if !request.FromDate.IsZero() && !request.ToDate.IsZero() {
		log.Printf("Using FromDate and ToDate: %v - %v", request.FromDate, request.ToDate)

		// Check if FromDate and ToDate are valid dates before truncating
		if request.FromDate.Year() < 1900 || request.ToDate.Year() < 1900 {
			log.Println("Invalid FromDate or ToDate (year < 1900)")
			return time.Time{}, time.Time{}, errors.New("invalid FromDate or ToDate (year < 1900)")
		}

		fromDate = request.FromDate.Truncate(24 * time.Hour)
		toDate = request.ToDate.Truncate(24 * time.Hour).Add(24*time.Hour - time.Nanosecond) // End of day
	} else {
		log.Println("No period or dates specified")
		return time.Time{}, time.Time{}, errors.New("fromDate and toDate are required if period is not specified")
	}

	log.Printf("resolveDateRange returning fromDate: %v, toDate: %v", fromDate, toDate)
	return fromDate, toDate, nil
}

// GetInventoryReportData retrieves inventory report data without generating a file.
func (s *reportService) GetInventoryReportData(
	ctx context.Context,
	userID int,
	departmentID int,
	request *dto.DateRangeRequest,
) ([]dto.Asisstant230ReportItem, error) {
	log.Printf("GetInventoryReportData called with userID: %d, departmentID: %d, request: %+v", userID, departmentID, request)

	resolvedFromDate, resolvedToDate, err := s.resolveDateRange(request)
	if err != nil {
		log.Printf("Error resolving date range: %v", err)
		return nil, err
	}

	if err = s.validateDateRange(resolvedFromDate, resolvedToDate); err != nil {
		log.Printf("Error validating date range: %v", err)
		return nil, err
	}

	logRequest := *request
	logRequest.FromDate = &resolvedFromDate
	logRequest.ToDate = &resolvedToDate
	searchParams, err := json.Marshal(logRequest)
	if err != nil {
		log.Printf("Error marshalling search params: %v", err)
		searchParams = []byte(`{"error": "failed to marshal search parameters"}`)
	}

	accessLog := &models.AccessLog{
		UserID:       userID,
		OperationID:  1,
		AccessTime:   time.Now(),
		SearchParams: string(searchParams),
		Status:       "pending",
	}

	logID, err := s.operationRepo.LogAccess(ctx, accessLog)
	if err != nil {
		log.Printf("Error logging access: %v", err)
	}

	var items []dto.Asisstant230ReportItem
	items, err = s.inventoryRepo.GetInventoryReport(ctx, resolvedFromDate, resolvedToDate, departmentID)
	if err != nil {
		log.Printf("Error querying inventory data: %v", err)
		s.updateLogStatus(ctx, logID, "error")
		return nil, fmt.Errorf("error querying inventory data: %w", err)
	}

	if len(items) == 0 {
		log.Printf("No data found for date range from %s to %s",
			resolvedFromDate.Format("2006-01-02"),
			resolvedToDate.Format("2006-01-02"))
		s.updateLogStatus(ctx, logID, "success")
		return []dto.Asisstant230ReportItem{}, nil
	}

	s.updateLogStatus(ctx, logID, "success")

	return items, nil
}

// ExportInventoryReport generates and exports the inventory report to an Excel file.
func (s *reportService) ExportInventoryReport(
	ctx context.Context,
	userID int,
	departmentID int,
	request *dto.DateRangeRequest,
) (*dto.ReportFileResponse, error) {
	log.Printf("ExportInventoryReport called with userID: %d, departmentID: %d, request: %+v", userID, departmentID, request)

	// Resolve actual fromDate and toDate
	resolvedFromDate, resolvedToDate, err := s.resolveDateRange(request)
	if err != nil {
		log.Printf("Error resolving date range: %v", err)
		return nil, err
	}

	// Validate the resolved date range
	if err = s.validateDateRange(resolvedFromDate, resolvedToDate); err != nil {
		log.Printf("Error validating date range: %v", err)
		return nil, err
	}

	// Log access attempt for export
	logRequest := *request // Create a copy
	logRequest.FromDate = &resolvedFromDate
	logRequest.ToDate = &resolvedToDate

	searchParams, err := json.Marshal(logRequest)
	if err != nil {
		log.Printf("Error marshalling search params: %v", err)
		searchParams = []byte(`{"error": "failed to marshal search parameters"}`)
	}

	accessLog := &models.AccessLog{
		UserID:       userID,
		OperationID:  2, // Assuming operation ID 2 is for inventory report export
		AccessTime:   time.Now(),
		SearchParams: string(searchParams),
		Status:       "pending",
	}

	logID, err := s.operationRepo.LogAccess(ctx, accessLog)
	if err != nil {
		log.Printf("Error logging access for export: %v", err)
	}

	// Get data using the repository
	items, err := s.inventoryRepo.GetInventoryReport(ctx, resolvedFromDate, resolvedToDate, departmentID)
	if err != nil {
		log.Printf("Error getting inventory data for export: %v", err)
		s.updateLogStatus(ctx, logID, "error")
		return nil, fmt.Errorf("error getting inventory data for export: %w", err)
	}

	if len(items) == 0 {
		log.Println("No data found to export for the specified date range")
		s.updateLogStatus(ctx, logID, "success") // Exporting no data is also a success
		return nil, errors.New("no data found to export for the specified date range")
	}

	// Prepare title for the Excel file
	title := fmt.Sprintf("Export Sales 230 from %s to %s",
		resolvedFromDate.Format("02/01/2006"),
		resolvedToDate.Format("02/01/2006"),
	)

	headers := []string{
		"document_date",
		"sales_order_number",
		"customer_name",
		"currency_type",
		"currency",
		"detailed_order_number",
		"invoice_number",
		"notes",
	}

	// Prepare data for Excel export
	data := make([]map[string]interface{}, len(items))
	for i, item := range items {
		data[i] = map[string]interface{}{
			"document_date":         item.DocumentDate,
			"sales_order_number":    item.SalesOrderNumber,
			"customer_name":         item.CustomerName,
			"currency_type":         item.CurrencyType,
			"currency":              item.Currency,
			"detailed_order_number": item.DetailedOrderNumber,
			"invoice_number":        item.InvoiceNumber,
			"notes":                 item.Notes,
		}
	}
	// Generate Excel file using utils
	filePath, fileDetail, err := utils.ExportToExcel(data, headers, title)
	if err != nil {
		s.updateLogStatus(ctx, logID, "error")
		return nil, fmt.Errorf("error exporting to Excel: %w", err)
	}

	// Update log status to success
	s.updateLogStatus(ctx, logID, "success")

	// Prepare response for frontend
	fileName := filepath.Base(filePath)

	return &dto.ReportFileResponse{
		ReportName:  title,
		FileName:    fileName,
		FileDetal:   fileDetail,
		GeneratedAt: time.Now(),
	}, nil
}

// validateDateRange validates date range for reports.
// This is an internal helper, not exposed via interface.
func (s *reportService) validateDateRange(fromDate, toDate time.Time) error {
	// Check that fromDate is before or equal to toDate
	if fromDate.After(toDate) {
		return errors.New("from date must be before or equal to to date")
	}

	// Check that toDate is not in the future (compared to current end of day)
	nowEndOfDay := time.Now().Truncate(24 * time.Hour).Add(24*time.Hour - time.Nanosecond)
	if toDate.After(nowEndOfDay) {
		return errors.New("to date cannot be in the future")
	}

	// Check that date range is within allowed months
	maxMonths := s.config.Excel.MaxSearchMonths

	oldestAllowed := time.Now().Truncate(24*time.Hour).AddDate(0, -maxMonths, 0)

	if fromDate.Truncate(24 * time.Hour).Before(oldestAllowed) {
		return fmt.Errorf("date range cannot exceed %d months from current date", maxMonths)
	}

	return nil
}

// updateLogStatus updates the status of an access log.
func (s *reportService) updateLogStatus(ctx context.Context, logID int, status string) {
	if logID <= 0 {
		return // Do not attempt to update if logID is invalid
	}

	if _, err := s.operationRepo.UpdateLogStatus(ctx, logID, status); err != nil {
		log.Printf("Error updating log status for logID %d: %v", logID, err)
	}
}
