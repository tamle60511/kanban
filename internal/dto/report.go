package dto

import "time"

// DateRangeRequest defines the request body for report generation with flexible date input.
type DateRangeRequest struct {
	FromDate *time.Time `json:"from_date,omitempty" validate:"required_without=Period"`                                                                          // Optional: If Period is provided, these might be ignored
	ToDate   *time.Time `json:"toDate,omitempty" validate:"required_without=Period"`                                                                             // Optional: If Period is provided, these might be ignored
	Period   *string    `json:"period,omitempty" validate:"required_without=FromDate,required_without=ToDate,oneof=7days 30days 3months currentmonth lastmonth"` // New: e.g., "7days", "30days", "3months", "currentmonth", "lastmonth"
}

type ReportRequest struct {
	DepartmentID string           `json:"department_id"`
	DateRange    DateRangeRequest `json:"date_range"`
}

type InventoryReportItem struct {
	DocumentDate        string `json:"document_date"`         // Date of the document (ngày chứng từ)
	SalesOrderNumber    string `json:"sales_order_number"`    // Sales order number (mã đơn bán hàng)
	CustomerName        string `json:"customer_name"`         // Customer name (tên khách hàng)
	ReceiptNumber       string `json:"receipt_number"`        // Result receipt number (mã phiếu kết sổ)
	CurrencyType        string `json:"currency_type"`         // Currency type (nguyên tệ)
	Currency            string `json:"currency"`              // Currency (nội tệ)
	DetailedOrderNumber string `json:"detailed_order_number"` // Detailed order number (mã đơn hàng chi tiết)
	InvoiceNumber       string `json:"invoice_number"`        // Invoice number (hoa đơn)
	Notes               string `json:"notes"`                 // Notes (ghi chú)
}

// ReportDataResponse is the structure for API response when viewing report data (JSON).
type ReportDataResponse struct {
	ReportName  string                `json:"report_name"`
	GeneratedAt time.Time             `json:"generated_at"`
	Items       []InventoryReportItem `json:"items"`
}

// ReportFileResponse is the structure for API response when generating an Excel file.
type ReportFileResponse struct {
	ReportName  string    `json:"report_name"`
	FileName    string    `json:"file_name"`    // Name of the file for download
	FileDetal   any       `json:"filed_detail"` // Detail of the file (e.g., excelize.File)
	GeneratedAt time.Time `json:"generated_at"`
}
