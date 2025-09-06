package dto

import "time"

type Asisstant610ReportItem struct {
	DocDate       string `json:"doc_date"`        // Date of the document (ngày chứng từ)
	Ar_Type       string `json:"ar_type"`         // Accounts receivable type (loại chứng từ)
	ShippingOrder string `json:"shipping_order"`  // Sales order number (mã đơn bán hàng)
	CustomerName  string `json:"customer_name"`   // Customer name (tên khách hàng) - now a pointer
	TotalAmtTrans string `json:"total_amt_trans"` // Total amount in the transaction currency (nguyên tệ)
	TotalAmt      string `json:"total_amt"`       // Total amount in the local currency (nội tệ)
	OrderNo       string `json:"order_no"`        // Detailed order number (mã đơn hàng chi tiết)
	InvoiceNumber string `json:"invoice_number"`  // Invoice number (hoa đơn)
	Notes         string `json:"notes"`           // Notes (ghi chú)
}

type Assistant610DataResponse struct {
	ReportName  string                   `json:"report_name"`
	GeneratedAt time.Time                `json:"generated_at"`
	Items       []Asisstant610ReportItem `json:"items"`
}
