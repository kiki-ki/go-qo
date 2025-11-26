package output

import (
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

type Options struct {
	Format Format
	Output io.Writer
}

func DefaultOptions() *Options {
	return &Options{
		Format: FormatTable,
		Output: os.Stdout,
	}
}

type Printer struct {
	opts *Options
}

// Creates a new Printer with the given options.
func NewPrinter(opts *Options) *Printer {
	if opts == nil {
		opts = DefaultOptions()
	}
	if opts.Output == nil {
		opts.Output = os.Stdout
	}
	return &Printer{opts: opts}
}

// Prints SQL results.
func (p *Printer) PrintRows(rows *sql.Rows) error {
	columns, err := rows.Columns()
	if err != nil {
		return fmt.Errorf("failed to get columns: %w", err)
	}

	data, err := p.scanRows(rows, len(columns))
	if err != nil {
		return err
	}

	switch p.opts.Format {
	case FormatJSON:
		return p.printJSON(columns, data)
	case FormatCSV:
		return p.printCSV(columns, data)
	default:
		return p.printTable(columns, data)
	}
}

// Scans all rows into a slice of slices.
func (p *Printer) scanRows(rows *sql.Rows, numCols int) ([][]any, error) {
	var result [][]any

	values := make([]any, numCols)
	valuePtrs := make([]any, numCols)
	for i := range values {
		valuePtrs[i] = &values[i]
	}

	for rows.Next() {
		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		row := make([]any, numCols)
		for i, val := range values {
			row[i] = p.normalizeValue(val)
		}
		result = append(result, row)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return result, nil
}

// Converts database values to appropriate Go types.
func (p *Printer) normalizeValue(val any) any {
	if val == nil {
		return nil
	}

	switch v := val.(type) {
	case []byte:
		return string(v)
	case float64:
		if float64(int64(v)) == v {
			return int64(v)
		}
		return v
	default:
		return v
	}
}

// Print result formatted as a table.
func (p *Printer) printTable(columns []string, data [][]any) error {
	rows := make([][]string, len(data))

	for i, row := range data {
		r := make([]string, len(row))
		for j, val := range row {
			if val == nil {
				r[j] = "(NULL)"
			} else {
				r[j] = fmt.Sprintf("%v", val)
			}
		}
		rows[i] = r
	}

	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("240"))).
		Headers(columns...).
		Rows(rows...)

	t.StyleFunc(func(row, col int) lipgloss.Style {
		return lipgloss.NewStyle().Padding(0, 1)
	})

	_, _ = fmt.Fprintln(p.opts.Output, t.Render())

	return nil
}

// Print result formatted as a JSON.
func (p *Printer) printJSON(columns []string, data [][]any) error {
	result := make([]map[string]any, len(data))

	for i, row := range data {
		obj := make(map[string]any)
		for j, col := range columns {
			obj[col] = row[j]
		}
		result[i] = obj
	}

	encoder := json.NewEncoder(p.opts.Output)
	encoder.SetIndent("", "  ")
	return encoder.Encode(result)
}

// Print result formatted as a CSV.
func (p *Printer) printCSV(columns []string, data [][]any) error {
	w := csv.NewWriter(p.opts.Output)
	defer w.Flush()

	if err := w.Write(columns); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	for _, row := range data {
		record := make([]string, len(row))
		for i, val := range row {
			if val == nil {
				record[i] = ""
			} else {
				record[i] = fmt.Sprintf("%v", val)
			}
		}
		if err := w.Write(record); err != nil {
			return fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	return w.Error()
}
