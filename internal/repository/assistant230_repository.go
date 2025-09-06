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

type InventoryRepository interface {
	GetInventoryReport(
		ctx context.Context,
		fromDate time.Time,
		toDate time.Time,
		departmentID int,
	) ([]dto.Asisstant230ReportItem, error)
}

type inventoryRepository struct {
	erpDB *sql.DB
}

func NewInventoryRepository(erpDB *sql.DB) InventoryRepository {
	return &inventoryRepository{
		erpDB: erpDB,
	}
}

func (r *inventoryRepository) GetInventoryReport(
	ctx context.Context,
	fromDate time.Time,
	toDate time.Time,
	departmentID int,
) ([]dto.Asisstant230ReportItem, error) {
	log.Printf("GetInventoryReport called with fromDate: %v, toDate: %v, departmentID: %d", fromDate, toDate, departmentID)
	_, err := r.erpDB.ExecContext(ctx, "USE Leader")
	if err != nil {
		return nil, fmt.Errorf("error switching database: %w", err)
	}

	query := `
   SELECT DISTINCT
    CONVERT(VARCHAR(10), CONVERT(DATETIME, COPTG.TG042), 103) AS document_date,
    COPTG.TG001 + '-' + COPTG.TG002 AS sales_order_number,
    COPTG.TG007 AS customer_name,
    CASE
        WHEN COPTG.TG011 = 'VND' THEN 
            REPLACE(CONVERT(VARCHAR, CONVERT(MONEY, (ISNULL(COPTG.TG013, 0) + ISNULL(COPTG.TG025, 0))), 1), '.00', '')
        WHEN COPTG.TG011 = 'USD' THEN 
            CASE 
                WHEN (ISNULL(COPTG.TG013, 0) + ISNULL(COPTG.TG025, 0)) - FLOOR(ISNULL(COPTG.TG013, 0) + ISNULL(COPTG.TG025, 0)) = 0 THEN 
                    REPLACE(CONVERT(VARCHAR, CONVERT(MONEY, (ISNULL(COPTG.TG013, 0) + ISNULL(COPTG.TG025, 0))), 1), '.00', '')
                ELSE CONVERT(VARCHAR, CONVERT(MONEY, (ISNULL(COPTG.TG013, 0) + ISNULL(COPTG.TG025, 0))), 1)
            END
        ELSE CONVERT(VARCHAR, CONVERT(MONEY, (ISNULL(COPTG.TG013, 0) + ISNULL(COPTG.TG025, 0))), 1)
    END AS currency_type,
    REPLACE(CONVERT(VARCHAR, CONVERT(MONEY, (ISNULL(COPTG.TG045, 0) + ISNULL(COPTG.TG046, 0))), 1), '.00', '') AS currency,
    ISNULL(COPTD.TD001 + '-' + COPTD.TD002 + '-' + RIGHT('0' + CONVERT(VARCHAR, COPTD.TD003), 4), '') AS detailed_order_number,
    ISNULL(ACRTA.TA036, '') AS invoice_number,
    ISNULL(COPTG.TG020, '') AS notes
FROM 
    COPTG WITH (NOLOCK)
LEFT JOIN 
    ACRTB WITH (NOLOCK) ON ACRTB.TB005 = COPTG.TG001 AND ACRTB.TB006 = COPTG.TG002
LEFT JOIN 
    ACRTA WITH (NOLOCK) ON ACRTA.TA001 = ACRTB.TB001 AND ACRTA.TA002 = ACRTB.TB002
LEFT JOIN 
    COPTH WITH (NOLOCK) ON COPTH.TH001 = COPTG.TG001 AND COPTH.TH002 = COPTG.TG002
LEFT JOIN 
    COPTD WITH (NOLOCK) ON COPTD.TD001 = COPTH.TH014 AND COPTD.TD002 = COPTH.TH015 AND COPTD.TD003 = COPTH.TH016
WHERE 
    COPTG.TG023 <> 'V'  
    AND TG042 BETWEEN @FromDate AND @ToDate AND ACRTA.TA001 IS NULL
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

	var items []dto.Asisstant230ReportItem
	for rows.Next() {
		var item dto.Asisstant230ReportItem
		if err := rows.Scan(
			&item.DocumentDate,
			&item.SalesOrderNumber,
			&item.CustomerName,

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
