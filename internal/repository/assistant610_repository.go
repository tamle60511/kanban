package repository

import (
	"context"
	"database/sql"
	"erp-excel/internal/dto"
	"fmt"
	"log"
	"time"
)

type Assistant610Repository interface {
	GetAssistant610Report(
		ctx context.Context,
		fromDate time.Time,
		toDate time.Time,
		departmentID int,
	) ([]dto.Asisstant610ReportItem, error)
}

type assistant610Repository struct {
	erpDB *sql.DB
}

func NewAssistant610Repository(erpDB *sql.DB) Assistant610Repository {
	return &assistant610Repository{
		erpDB: erpDB,
	}
}

func (r *assistant610Repository) GetAssistant610Report(
	ctx context.Context,
	fromDate time.Time,
	toDate time.Time,
	departmentID int,
) ([]dto.Asisstant610ReportItem, error) {
	log.Printf("GetAssistant610Report called with fromDate: %v, toDate: %v, departmentID: %d", fromDate, toDate, departmentID)
	_, err := r.erpDB.ExecContext(ctx, "USE Leader")
	if err != nil {
		return nil, fmt.Errorf("error switching database: %w", err)
	}

	query := `
	SELECT DISTINCT
    CONVERT(VARCHAR(10), ACRTB.TB008, 103) AS doc_date,
    ACRTA.TA001 + '-' + ACRTA.TA002 AS ar_type,
    ACRTB.TB005 + '-' + ACRTB.TB006 + '-' + ACRTB.TB007 AS shipping_order,
    ISNULL(COPTG.TG007, '') AS customer_name,
        CASE 
        WHEN ACRTA.TA009 = 'VND' THEN REPLACE(CONVERT(VARCHAR, CONVERT(MONEY, (ACRTA.TA029 + ACRTA.TA030)), 1), '.00', '')
        WHEN ACRTA.TA009 = 'USD' THEN 
            CASE 
                WHEN (ACRTA.TA029 + ACRTA.TA030) - FLOOR(ACRTA.TA029 + ACRTA.TA030) = 0 THEN 
                    REPLACE(CONVERT(VARCHAR, CONVERT(MONEY, (ACRTA.TA029 + ACRTA.TA030)), 1), '.00', '')
                ELSE CONVERT(VARCHAR, CONVERT(MONEY, (ACRTA.TA029 + ACRTA.TA030)), 1)
            END
        ELSE CONVERT(VARCHAR, CONVERT(MONEY, (ACRTA.TA029 + ACRTA.TA030)), 1)
    END AS 'total_amt_trans',
    REPLACE(CONVERT(VARCHAR, CONVERT(MONEY, (ACRTA.TA041 + ACRTA.TA042)), 1), '.00', '') AS 'total_amt',
      ISNULL(DetailOrder.order_no, '') AS order_no,
    ISNULL(ACRTA.TA036, '') AS invoice_number,
    ISNULL(COPTG.TG020, '') AS notes
FROM 
    ACRTA WITH (NOLOCK)
JOIN 
    ACRTB WITH (NOLOCK) ON ACRTA.TA001 = ACRTB.TB001 AND ACRTA.TA002 = ACRTB.TB002
LEFT JOIN 
    COPTG WITH (NOLOCK) ON ACRTB.TB005 = COPTG.TG001 AND ACRTB.TB006 = COPTG.TG002
OUTER APPLY (
    SELECT TOP 1
        REPLACE(RTRIM(COPTD.TD001) + '-' + RTRIM(COPTD.TD002) + '-' + RTRIM(COPTD.TD003), '--', '-') AS order_no
    FROM 
        COPTH WITH (NOLOCK)
    JOIN 
        COPTD WITH (NOLOCK) ON COPTD.TD001 = COPTH.TH014 AND COPTD.TD002 = COPTH.TH015 AND COPTD.TD003 = COPTH.TH016
    WHERE 
        COPTH.TH001 = COPTG.TG001 AND COPTH.TH002 = COPTG.TG002
    ORDER BY 
        COPTD.TD003
) AS DetailOrder
WHERE  ACRTB.TB008 BETWEEN @FromDate AND @ToDate
	
	`
	log.Printf("Executing query: %s with FromDate: %v, ToDate: %v, DepartmentID: %d", query, fromDate, toDate, departmentID)

	rows, err := r.erpDB.QueryContext(
		ctx,
		query,
		sql.Named("FromDate", fromDate),
		sql.Named("ToDate", toDate),
		sql.Named("DepartmentID", departmentID),
	)
	if err != nil {
		return nil, fmt.Errorf("error querying inventory data: %w", err)
	}
	defer rows.Close()

	var items []dto.Asisstant610ReportItem
	for rows.Next() {
		var item dto.Asisstant610ReportItem
		if err := rows.Scan(
			&item.DocDate,
			&item.Ar_Type,
			&item.ShippingOrder,
			&item.CustomerName,
			&item.TotalAmtTrans,
			&item.TotalAmt,
			&item.OrderNo,
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
