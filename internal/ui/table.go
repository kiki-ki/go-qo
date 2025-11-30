package ui

import (
	"database/sql"

	"github.com/charmbracelet/bubbles/table"

	"github.com/kiki-ki/go-qo/internal/output"
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
			rowData[i] = output.FormatValueRaw(val)
		}
		tRows = append(tRows, rowData)
	}

	return tCols, tRows, nil
}
