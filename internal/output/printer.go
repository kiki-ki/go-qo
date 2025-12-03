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
		Format: FormatJSON,
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
		return p.printCSV(columns, data, ',')
	case FormatTSV:
		return p.printCSV(columns, data, '\t')
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
			row[i] = NormalizeValue(val)
		}
		result = append(result, row)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return result, nil
}

// Print result formatted as a table.
func (p *Printer) printTable(columns []string, data [][]any) error {
	rows := make([][]string, len(data))

	for i, row := range data {
		r := make([]string, len(row))
		for j, val := range row {
			r[j] = FormatValueForDisplay(val)
		}
		rows[i] = r
	}

	// Create renderer that detects TTY for color support
	renderer := lipgloss.NewRenderer(p.opts.Output)

	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(renderer.NewStyle().Foreground(lipgloss.Color("240"))).
		Headers(columns...).
		Rows(rows...)

	t.StyleFunc(func(row, col int) lipgloss.Style {
		return renderer.NewStyle().Padding(0, 1)
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

// Print result formatted as a CSV (or TSV with tab delimiter).
func (p *Printer) printCSV(columns []string, data [][]any, delimiter rune) error {
	w := csv.NewWriter(p.opts.Output)
	w.Comma = delimiter
	defer w.Flush()

	if err := w.Write(columns); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	for _, row := range data {
		record := make([]string, len(row))
		for i, val := range row {
			switch v := val.(type) {
			case nil:
				record[i] = ""
			case map[string]any, []any:
				// Convert nested objects/arrays back to JSON string
				b, err := json.Marshal(v)
				if err != nil {
					record[i] = fmt.Sprintf("%v", v)
				} else {
					record[i] = string(b)
				}
			default:
				record[i] = fmt.Sprintf("%v", v)
			}
		}
		if err := w.Write(record); err != nil {
			return fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	return w.Error()
}
