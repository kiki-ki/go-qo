// Package printer provides output formatting for query results.
package printer

import (
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

// Format represents the output format type.
type Format string

const (
	FormatTable Format = "table"
	FormatJSON  Format = "json"
	FormatCSV   Format = "csv"
)

// Options configures the printer behavior.
type Options struct {
	Format Format
	Output io.Writer
}

// DefaultOptions returns default printer options.
func DefaultOptions() *Options {
	return &Options{
		Format: FormatTable,
		Output: os.Stdout,
	}
}

// Printer handles output formatting.
type Printer struct {
	opts *Options
}

// New creates a new Printer with the given options.
func New(opts *Options) *Printer {
	if opts == nil {
		opts = DefaultOptions()
	}
	if opts.Output == nil {
		opts.Output = os.Stdout
	}
	return &Printer{opts: opts}
}

// PrintRows prints the query results in the configured format.
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

// scanRows scans all rows into a slice of slices.
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

// normalizeValue converts database values to appropriate Go types.
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

// printTable outputs data in table format.
func (p *Printer) printTable(columns []string, data [][]any) error {
	t := table.NewWriter()
	t.SetOutputMirror(p.opts.Output)

	headerRow := make(table.Row, len(columns))
	for i, col := range columns {
		headerRow[i] = col
	}
	t.AppendHeader(headerRow)

	t.SetStyle(table.StyleLight)
	t.Style().Options.SeparateRows = true
	t.Style().Format.Header = text.FormatDefault

	for _, row := range data {
		tableRow := make(table.Row, len(row))
		for i, val := range row {
			if val == nil {
				tableRow[i] = "NULL"
			} else {
				tableRow[i] = val
			}
		}
		t.AppendRow(tableRow)
	}

	t.Render()
	return nil
}

// printJSON outputs data in JSON format.
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

// printCSV outputs data in CSV format.
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
