package tui

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
		tCols[i] = table.Column{Title: c, Width: 15}
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
	if val == nil {
		return "(NULL)"
	}
	switch v := val.(type) {
	case []byte:
		return string(v)
	case float64:
		if float64(int64(v)) == v {
			return fmt.Sprintf("%d", int64(v))
		}
		return fmt.Sprintf("%g", v)
	default:
		return fmt.Sprintf("%v", v)
	}
}
