package utils

import (
	"bytes"
	"erp-excel/internal/translate"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	excelize "github.com/xuri/excelize/v2"
)

// ExportToExcel exports data to Excel file
func ExportToExcel(data []map[string]interface{}, headers []string, title string) (string, *bytes.Buffer, error) {
	// Create a new Excel file
	f := excelize.NewFile()
	defer f.Close()

	// Get the default sheet
	sheetName := "Sheet1"

	// Set title
	f.SetCellValue(sheetName, "A1", title)

	// Set title style
	titleStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Size:  16,
			Bold:  true,
			Color: "1F497D",
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
	})
	if err != nil {
		return "", nil, fmt.Errorf("error creating title style: %w", err)
	}

	// Apply title style and merge cells for title
	f.SetCellStyle(sheetName, "A1", fmt.Sprintf("%c1", rune('A'+len(headers)-1)), titleStyle)
	f.MergeCell(sheetName, "A1", fmt.Sprintf("%c1", rune('A'+len(headers)-1)))

	// Set headers
	headerStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:  true,
			Color: "FFFFFF",
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"4472C4"},
			Pattern: 1,
		},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
	})
	if err != nil {
		return "", nil, fmt.Errorf("error creating header style: %w", err)
	}

	// Write headers
	for i, header := range headers {
		cellPos := fmt.Sprintf("%c3", rune('A'+i))
		f.SetCellValue(sheetName, cellPos, translate.TranslateKey(header))
	}

	// Apply header style
	headerRange := fmt.Sprintf("A3:%c3", rune('A'+len(headers)-1))
	f.SetCellStyle(sheetName, headerRange, headerRange, headerStyle)

	// Data cell styles
	dataStyle, err := f.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
		Alignment: &excelize.Alignment{
			Vertical: "center",
		},
	})
	if err != nil {
		return "", nil, fmt.Errorf("error creating data style: %w", err)
	}

	// TODO: currently, no longer using number format style
	// numberStyle, err := f.NewStyle(&excelize.Style{
	// 	Border: []excelize.Border{
	// 		{Type: "left", Color: "000000", Style: 1},
	// 		{Type: "top", Color: "000000", Style: 1},
	// 		{Type: "bottom", Color: "000000", Style: 1},
	// 		{Type: "right", Color: "000000", Style: 1},
	// 	},
	// 	Alignment: &excelize.Alignment{
	// 		Horizontal: "right",
	// 		Vertical:   "center",
	// 	},
	// 	NumFmt: 3, // #,##0 format
	// })
	// if err != nil {
	// 	return "", nil, fmt.Errorf("error creating number style: %w", err)
	// }

	// Write data
	for i, item := range data {
		row := i + 4 // Data starts from row 4

		for j, header := range headers {
			cellPos := fmt.Sprintf("%c%d", rune('A'+j), row)
			f.SetCellValue(sheetName, cellPos, item[header])

			// Apply style based on data type
			f.SetCellStyle(sheetName, cellPos, cellPos, dataStyle)
		}
	}

	// Set column width
	for i := range headers {
		colName := string(rune('A' + i))
		f.SetColWidth(sheetName, colName, colName, 15)
	}

	// Set row height
	f.SetRowHeight(sheetName, 1, 30)
	f.SetRowHeight(sheetName, 3, 25)

	// Generate timestamp for filename
	timestamp := time.Now().Format("20060102_150405")

	// Create sanitized filename
	safeTitlePart := sanitizeFilename(title)
	if len(safeTitlePart) > 30 {
		safeTitlePart = safeTitlePart[:30]
	}

	// Complete filename
	filename := fmt.Sprintf("%s_%s.xlsx", safeTitlePart, timestamp)

	// Write file to buffer and return
	buf, err := f.WriteToBuffer()
	if err != nil {
		return "", nil, fmt.Errorf("error writing Excel to buffer: %w", err)
	}
	return filename, buf, nil
}

// sanitizeFilename removes invalid characters from filename
func sanitizeFilename(name string) string {
	name = filepath.Clean(name)
	invalidChars := []rune{'<', '>', ':', '"', '/', '\\', '|', '?', '*'}

	for _, char := range invalidChars {
		name = strings.ReplaceAll(name, string(char), "_")
	}

	return name
}
