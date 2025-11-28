package ui

import (
	"database/sql"
	"fmt"

	"github.com/charmbracelet/bubbles/table"
)

// SQLRowsToTable converts SQL rows to bubbles table format.
func SQLRowsToTable(rows *sql.Rows) ([]table.Column, []table.Row, error) {
	cols, err := rows.Columns()
	if err != nil {
		return nil, nil, err
	}

	tCols := make([]table.Column, len(cols))
	for i, c := range cols {
		tCols[i] = table.Column{Title: c, Width: defaultColumnWidth}
	}

	var tRows []table.Row

	values := make([]interface{}, len(cols))
	valuePtrs := make([]interface{}, len(cols))
	for i := range cols {
		valuePtrs[i] = &values[i]
	}

	for rows.Next() {
		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, nil, err
		}

		rowData := make(table.Row, len(cols))
		for i, val := range values {
			rowData[i] = FormatValue(val)
		}
		tRows = append(tRows, rowData)
	}

	return tCols, tRows, nil
}

// FormatValue converts a database value to string for display.
func FormatValue(val any) string {
	var s string
	if val == nil {
		return "(NULL)"
	}
	switch v := val.(type) {
	case []byte:
		s = string(v)
	case float64:
		if float64(int64(v)) == v {
			s = fmt.Sprintf("%d", int64(v))
		} else {
			s = fmt.Sprintf("%g", v)
		}
	default:
		s = fmt.Sprintf("%v", v)
	}
	return truncate(s, maxCellDisplay)
}

// truncate shortens a string to maxLen, adding "…" if truncated.
func truncate(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen-1]) + "…"
}
