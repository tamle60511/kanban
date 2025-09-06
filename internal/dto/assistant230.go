package dto

import "time"

type DateRangeRequest struct {
	FromDate *time.Time `json:"fromDate"`
	ToDate   *time.Time `json:"toDate"`
	Period   *string    `json:"period"`
}

type ReportRequest struct {
	DepartmentID string           `json:"department_id"`
	DateRange    DateRangeRequest `json:"date_range"`
}

type Asisstant230ReportItem struct {
	DocumentDate        string `json:"document_date"`         // Date of the document (ngày chứng từ)
	SalesOrderNumber    string `json:"sales_order_number"`    // Sales order number (mã đơn bán hàng)
	CustomerName        string `json:"customer_name"`         // Customer name (tên khách hàng)       // Result receipt number (mã phiếu kết sổ)
	CurrencyType        string `json:"currency_type"`         // Currency type (nguyên tệ)
	Currency            string `json:"currency"`              // Currency (nội tệ)
	DetailedOrderNumber string `json:"detailed_order_number"` // Detailed order number (mã đơn hàng chi tiết)
	InvoiceNumber       string `json:"invoice_number"`        // Invoice number (hoa đơn)
	Notes               string `json:"notes"`                 // Notes (ghi chú)
}

type ReportDataResponse struct {
	ReportName  string                   `json:"report_name"`
	GeneratedAt time.Time                `json:"generated_at"`
	Items       []Asisstant230ReportItem `json:"items"`
}

type ReportFileResponse struct {
	ReportName  string    `json:"report_name"`
	FileName    string    `json:"file_name"`    // Name of the file for download
	FileDetal   any       `json:"filed_detail"` // Detail of the file (e.g., excelize.File)
	GeneratedAt time.Time `json:"generated_at"`
}
