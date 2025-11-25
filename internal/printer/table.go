package printer

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

func Print(rows *sql.Rows) error {
	columns, err := rows.Columns()
	if err != nil {
		return fmt.Errorf("failed to get columns: %w", err)
	}

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)

	headerRow := table.Row{}
	for _, col := range columns {
		headerRow = append(headerRow, col)
	}
	t.AppendHeader(headerRow)

	t.SetStyle(table.StyleLight)
	t.Style().Options.SeparateRows = true
	t.Style().Format.Header = text.FormatDefault

	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	for i := range columns {
		valuePtrs[i] = &values[i]
	}

	for rows.Next() {
		if err := rows.Scan(valuePtrs...); err != nil {
			return fmt.Errorf("failed to scan row: %w", err)
		}

		row := table.Row{}

		for _, val := range values {
			if val == nil {
				row = append(row, "NULL")
			} else {
				switch v := val.(type) {
				case []byte:
					row = append(row, string(v))
				case float64:
					if float64(int64(v)) == v {
						row = append(row, int64(v))
					} else {
						row = append(row, v)
					}
				default:
					row = append(row, v)
				}
			}
		}
		t.AppendRow(row)
	}

	t.Render()

	return nil
}
