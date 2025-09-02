// repository/inventory_repository.go
package repository

import (
	"context"
	"database/sql"
	"erp-excel/internal/dto"
	"fmt"
	"log"
	"time"
)

// InventoryRepository defines the interface for inventory-related database queries
type InventoryRepository interface {
	// GetInventoryReport retrieves inventory report data based on specified criteria
	GetInventoryReport(
		ctx context.Context,
		fromDate time.Time,
		toDate time.Time,
		departmentID int,
	) ([]dto.InventoryReportItem, error)
}

// inventoryRepository implements InventoryRepository
type inventoryRepository struct {
	erpDB *sql.DB
}

// NewInventoryRepository creates a new instance of InventoryRepository
func NewInventoryRepository(erpDB *sql.DB) InventoryRepository {
	return &inventoryRepository{
		erpDB: erpDB,
	}
}

// GetInventoryReport retrieves detailed inventory report
func (r *inventoryRepository) GetInventoryReport(
	ctx context.Context,
	fromDate time.Time,
	toDate time.Time,
	departmentID int, // departmentID not used in query, consider removing if not needed or add to query
) ([]dto.InventoryReportItem, error) {
	log.Printf("GetInventoryReport called with fromDate: %v, toDate: %v, departmentID: %d", fromDate, toDate, departmentID)
	_, err := r.erpDB.ExecContext(ctx, "USE Leader")
	if err != nil {
		return nil, fmt.Errorf("error switching database: %w", err)
	}

	query := `
    SELECT DISTINCT
        CONVERT(VARCHAR(10), CONVERT(DATETIME, TG042), 103) AS document_date,
        TG001 + '-' + TG002 AS sales_order_number,
        TG007 AS customer_name,
        ISNULL(TA001 + '-' + TA002, '') AS receipt_number,
        CASE
            WHEN TG011 = 'VND' THEN REPLACE(CONVERT(VARCHAR, CONVERT(MONEY, (ISNULL(TG013, 0) + ISNULL(TG025, 0))), 1), '.00', '')
            ELSE CONVERT(VARCHAR, CONVERT(MONEY, (ISNULL(TG013, 0) + ISNULL(TG025, 0))), 1)
        END AS currency_type,
        REPLACE(CONVERT(VARCHAR, CONVERT(MONEY, (ISNULL(TG045, 0) + ISNULL(TG046, 0))), 1), '.00', '') AS currency,
        ISNULL(TD001 + '-' + TD002 + '-' + RIGHT('0' + TD003, 4), '') AS detailed_order_number,
        ISNULL(TA036, '') AS invoice_number,
        ISNULL(TG020, '') AS notes
    FROM
        COPTG WITH (NOLOCK)
    LEFT JOIN
        ACRTA WITH (NOLOCK) ON TA001 = TG001 AND TA002 = TG002
    LEFT JOIN
        COPTD WITH (NOLOCK) ON TD001 = TG001 AND TD002 = TG002
    WHERE
        TG023 <> 'V'
    AND TG042 BETWEEN @FromDate AND @ToDate
    `
	log.Printf("Executing query: %s with FromDate: %v, ToDate: %v", query, fromDate, toDate)

	rows, err := r.erpDB.QueryContext(
		ctx,
		query,
		sql.Named("FromDate", fromDate),
		sql.Named("ToDate", toDate),
		// sql.Named("DepartmentID", departmentID), // Uncomment and use if needed in SQL query
	)
	if err != nil {
		return nil, fmt.Errorf("error querying inventory data: %w", err)
	}
	defer rows.Close()

	var items []dto.InventoryReportItem
	for rows.Next() {
		var item dto.InventoryReportItem
		if err := rows.Scan(
			&item.DocumentDate,
			&item.SalesOrderNumber,
			&item.CustomerName,
			&item.ReceiptNumber,
			&item.CurrencyType,
			&item.Currency,
			&item.DetailedOrderNumber,
			&item.InvoiceNumber,
			&item.Notes,
		); err != nil {
			return nil, fmt.Errorf("error scanning inventory data: %w", err)
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating inventory data: %w", err)
	}

	return items, nil
}
