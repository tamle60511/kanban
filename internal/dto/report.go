package dto

import "time"

// DateRangeRequest defines the request body for report generation with flexible date input.
type DateRangeRequest struct {
	FromDate *time.Time `json:"fromDate,omitempty" validate:"omitempty,time"` // Optional: If Period is provided, these might be ignored
	ToDate   *time.Time `json:"toDate,omitempty"`                             // Optional: If Period is provided, these might be ignored
	Period   *string    `json:"period,omitempty"`                             // New: e.g., "7days", "30days", "3months", "currentmonth", "lastmonth"
}

type ReportRequest struct {
	DepartmentID string           `json:"department_id"`
	DateRange    DateRangeRequest `json:"date_range"`
}

// InventoryReportItem represents a single item in the inventory report.
type InventoryReportItem struct {
	NgayCT           string `json:"ngayCT"`
	MaDonBanHang     string `json:"maDonBanHang"`
	KhachHang        string `json:"khachHang"`
	MaPhieuKetSo     string `json:"maPhieuKetSo"`
	NguyenTe         string `json:"nguyenTe"`
	NoiTe            string `json:"noiTe"`
	MaDonHangChiTiet string `json:"maDonHangChiTiet"`
	HoaDon           string `json:"hoaDon"`
	GhiChu           string `json:"ghiChu"`
}

// ReportDataResponse is the structure for API response when viewing report data (JSON).
type ReportDataResponse struct {
	ReportName  string                `json:"reportName"`
	GeneratedAt time.Time             `json:"generatedAt"`
	Items       []InventoryReportItem `json:"items"`
}

// ReportFileResponse is the structure for API response when generating an Excel file.
type ReportFileResponse struct {
	ReportName  string    `json:"reportName"`
	FileName    string    `json:"fileName"`   // Name of the file for download
	FileDetal   any       `json:"fileDetail"` // Detail of the file (e.g., excelize.File)
	GeneratedAt time.Time `json:"generatedAt"`
}
