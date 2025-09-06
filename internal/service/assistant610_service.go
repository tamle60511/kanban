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

// Assistant610Service defines methods for report generation.
type Assistant610Service interface {
	GetAssistant610ReportData(ctx context.Context, userID int, departmentID int, request *dto.DateRangeRequest) ([]dto.Asisstant610ReportItem, error)
	ExportAssistant610Report(ctx context.Context, userID int, departmentID int, request *dto.DateRangeRequest) (*dto.ReportFileResponse, error)
}

type assistant610Service struct {
	erpDB            *sql.DB
	config           *config.Config
	userRepo         repository.UserRepository
	operationRepo    repository.OperationRepository
	assistant610Repo repository.Assistant610Repository
}

// NewAssistant610Service creates a new report service.
func NewAssistant610Service(
	erpDB *sql.DB,
	config *config.Config,
	userRepo repository.UserRepository,
	operationRepo repository.OperationRepository,
	assistant610Repo repository.Assistant610Repository,
) Assistant610Service {
	return &assistant610Service{
		erpDB:            erpDB,
		config:           config,
		userRepo:         userRepo,
		operationRepo:    operationRepo,
		assistant610Repo: assistant610Repo,
	}
}

// resolveDateRange calculates actual fromDate and toDate based on Period or uses provided dates.
func (s *assistant610Service) resolveDateRange(request *dto.DateRangeRequest) (time.Time, time.Time, error) {
	log.Printf("resolveDateRange called with request: %+v", request)

	now := time.Now()
	currentEndOfDay := now.Truncate(24 * time.Hour).Add(24*time.Hour - time.Nanosecond)
	var fromDate, toDate time.Time

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

		if request.FromDate.Year() < 1900 || request.ToDate.Year() < 1900 {
			log.Println("Invalid FromDate or ToDate (year < 1900)")
			return time.Time{}, time.Time{}, errors.New("invalid FromDate or ToDate (year < 1900)")
		}

		fromDate = request.FromDate.Truncate(24 * time.Hour)
		toDate = request.ToDate.Truncate(24 * time.Hour).Add(24*time.Hour - time.Nanosecond)
	} else {
		log.Println("No period or dates specified")
		return time.Time{}, time.Time{}, errors.New("fromDate and toDate are required if period is not specified")
	}

	log.Printf("resolveDateRange returning fromDate: %v, toDate: %v", fromDate, toDate)
	return fromDate, toDate, nil
}

// GetAssistant610ReportData retrieves inventory report data without generating a file.
func (s *assistant610Service) GetAssistant610ReportData(
	ctx context.Context,
	userID int,
	departmentID int,
	request *dto.DateRangeRequest,
) ([]dto.Asisstant610ReportItem, error) {
	log.Printf("GetAssistant610ReportData called with userID: %d, departmentID: %d, request: %+v", userID, departmentID, request)

	resolvedFromDate, resolvedToDate, err := s.resolveDateRange(request)
	if err != nil {
		log.Printf("Error resolving date range: %v", err)
		return nil, err
	}

	if err = s.validate610DateRange(resolvedFromDate, resolvedToDate); err != nil {
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

	var items []dto.Asisstant610ReportItem
	items, err = s.assistant610Repo.GetAssistant610Report(ctx, resolvedFromDate, resolvedToDate, departmentID) // Updated method name
	if err != nil {
		log.Printf("Error querying inventory data: %v", err)
		s.updateLogStatus(ctx, logID, "error")
		return nil, fmt.Errorf("error querying inventory data: %w", err)
	}

	if len(items) == 0 {
		log.Printf("No data found for date range from %s to %s", resolvedFromDate.Format("2006-01-02"), resolvedToDate.Format("2006-01-02"))
		s.updateLogStatus(ctx, logID, "success")
		return []dto.Asisstant610ReportItem{}, nil
	}

	s.updateLogStatus(ctx, logID, "success")
	return items, nil
}

// ExportAssistant610Report generates and exports the inventory report to an Excel file.
func (s *assistant610Service) ExportAssistant610Report( // Changed receiver type to match struct
	ctx context.Context,
	userID int,
	departmentID int,
	request *dto.DateRangeRequest,
) (*dto.ReportFileResponse, error) {
	log.Printf("ExportAssistant610Report called with userID: %d, departmentID: %d, request: %+v", userID, departmentID, request)

	resolvedFromDate, resolvedToDate, err := s.resolveDateRange(request)
	if err != nil {
		log.Printf("Error resolving date range: %v", err)
		return nil, err
	}

	if err = s.validate610DateRange(resolvedFromDate, resolvedToDate); err != nil {
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
		OperationID:  2,
		AccessTime:   time.Now(),
		SearchParams: string(searchParams),
		Status:       "pending",
	}

	logID, err := s.operationRepo.LogAccess(ctx, accessLog)
	if err != nil {
		log.Printf("Error logging access for export: %v", err)
	}

	items, err := s.assistant610Repo.GetAssistant610Report(ctx, resolvedFromDate, resolvedToDate, departmentID)
	if err != nil {
		log.Printf("Error getting inventory data for export: %v", err)
		s.updateLogStatus(ctx, logID, "error")
		return nil, fmt.Errorf("error getting inventory data for export: %w", err)
	}

	if len(items) == 0 {
		log.Println("No data found to export for the specified date range")
		s.updateLogStatus(ctx, logID, "success")
		return nil, errors.New("no data found to export for the specified date range")
	}

	title := fmt.Sprintf("Export Sales 610 from %s to %s", resolvedFromDate.Format("02/01/2006"), resolvedToDate.Format("02/01/2006"))

	headers := []string{
		"doc_date",
		"ar_type",
		"shipping_order",
		"customer_name",
		"total_amt_trasn",
		"total_amt",
		"order_no",
		"invoice_number",
		"notes",
	}

	data := make([]map[string]interface{}, len(items))
	for i, item := range items {
		data[i] = map[string]interface{}{
			"doc_date":        item.DocDate,
			"ar_type":         item.Ar_Type,
			"shipping_order":  item.ShippingOrder,
			"customer_name":   item.CustomerName,
			"total_amt_trasn": item.TotalAmtTrans,
			"total_amt":       item.TotalAmt,
			"order_no":        item.OrderNo,
			"invoice_number":  item.InvoiceNumber,
			"notes":           item.Notes,
		}
	}

	filePath, fileDetail, err := utils.ExportToExcel(data, headers, title)
	if err != nil {
		s.updateLogStatus(ctx, logID, "error")
		return nil, fmt.Errorf("error exporting to Excel: %w", err)
	}

	s.updateLogStatus(ctx, logID, "success")

	fileName := filepath.Base(filePath)

	return &dto.ReportFileResponse{
		ReportName:  title,
		FileName:    fileName,
		FileDetal:   fileDetail,
		GeneratedAt: time.Now(),
	}, nil
}

// validateDateRange validates date range for reports.
func (s *assistant610Service) validate610DateRange(fromDate, toDate time.Time) error {
	if fromDate.After(toDate) {
		return errors.New("from date must be before or equal to to date")
	}

	nowEndOfDay := time.Now().Truncate(24 * time.Hour).Add(24*time.Hour - time.Nanosecond)
	if toDate.After(nowEndOfDay) {
		return errors.New("to date cannot be in the future")
	}

	maxMonths := s.config.Excel.MaxSearchMonths
	oldestAllowed := time.Now().Truncate(24*time.Hour).AddDate(0, -maxMonths, 0)

	if fromDate.Before(oldestAllowed) {
		return fmt.Errorf("date range cannot exceed %d months from current date", maxMonths)
	}

	return nil
}

// updateLogStatus updates the status of an access log.
func (s *assistant610Service) updateLogStatus(ctx context.Context, logID int, status string) {
	if logID <= 0 {
		return // Skip updating if logID is invalid
	}

	if _, err := s.operationRepo.UpdateLogStatus(ctx, logID, status); err != nil {
		log.Printf("Error updating log status for logID %d: %v", logID, err)
	}
}
