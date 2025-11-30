package parser

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// CSVOptions configures CSV parsing behavior.
type CSVOptions struct {
	NoHeader  bool // If true, first row is data, not header
	Delimiter rune // Field delimiter (default: ',')
}

// CSVParser implements Parser interface for CSV files.
type CSVParser struct {
	Options CSVOptions
}

// init registers the CSV parser.
func init() {
	Register(&CSVParser{})
}

// SupportedExtensions returns the file extensions this parser handles.
func (p *CSVParser) SupportedExtensions() []string {
	return []string{".csv"}
}

// Parse parses a CSV file into ParsedData.
func (p *CSVParser) Parse(path string) (*ParsedData, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", path, err)
	}
	return p.ParseBytes(data)
}

// ParseBytes parses CSV from a byte slice.
func (p *CSVParser) ParseBytes(data []byte) (*ParsedData, error) {
	reader := csv.NewReader(bytes.NewReader(data))

	// Set custom delimiter if specified
	if p.Options.Delimiter != 0 {
		reader.Comma = p.Options.Delimiter
	}

	// Read all records first
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV: %w", err)
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("empty CSV data")
	}

	var header []string
	var rawRows [][]string

	if p.Options.NoHeader {
		// Generate column names: col1, col2, ...
		numCols := len(records[0])
		header = make([]string, numCols)
		for i := range header {
			header[i] = fmt.Sprintf("col%d", i+1)
		}
		rawRows = records
	} else {
		header = records[0]
		if len(records) > 1 {
			rawRows = records[1:]
		}
	}

	if len(header) == 0 {
		return nil, fmt.Errorf("empty CSV header")
	}

	// Infer column types from data
	columns := p.inferColumns(header, rawRows)

	// Convert rows to typed values
	rows := p.convertRows(rawRows, columns)

	return &ParsedData{
		Columns: columns,
		Rows:    rows,
	}, nil
}

// inferColumns infers column types from header and data rows.
func (p *CSVParser) inferColumns(header []string, rawRows [][]string) []Column {
	columns := make([]Column, len(header))

	for i, name := range header {
		columns[i] = Column{
			Name: strings.TrimSpace(name),
			Type: p.inferColumnType(i, rawRows),
		}
	}

	return columns
}

// inferColumnType infers the type of a column by examining its values.
func (p *CSVParser) inferColumnType(colIdx int, rawRows [][]string) DataType {
	if len(rawRows) == 0 {
		return TypeText
	}

	hasInteger := false
	hasReal := false
	hasText := false

	for _, row := range rawRows {
		if colIdx >= len(row) {
			continue
		}

		val := strings.TrimSpace(row[colIdx])
		if val == "" {
			continue // Skip empty values for type inference
		}

		// Try integer
		if _, err := strconv.ParseInt(val, 10, 64); err == nil {
			hasInteger = true
			continue
		}

		// Try float
		if _, err := strconv.ParseFloat(val, 64); err == nil {
			hasReal = true
			continue
		}

		// It's text
		hasText = true
	}

	// Determine final type (widen as needed)
	if hasText {
		return TypeText
	}
	if hasReal {
		return TypeReal
	}
	if hasInteger {
		return TypeInteger
	}

	return TypeText
}

// convertRows converts raw string rows to typed values.
func (p *CSVParser) convertRows(rawRows [][]string, columns []Column) [][]any {
	rows := make([][]any, len(rawRows))

	for i, rawRow := range rawRows {
		row := make([]any, len(columns))
		for j, col := range columns {
			if j >= len(rawRow) {
				row[j] = nil
				continue
			}

			val := strings.TrimSpace(rawRow[j])
			if val == "" {
				row[j] = nil
				continue
			}

			row[j] = p.convertValue(val, col.Type)
		}
		rows[i] = row
	}

	return rows
}

// convertValue converts a string value to the appropriate type.
func (p *CSVParser) convertValue(val string, dataType DataType) any {
	switch dataType {
	case TypeInteger:
		if v, err := strconv.ParseInt(val, 10, 64); err == nil {
			return v
		}
		return val
	case TypeReal:
		if v, err := strconv.ParseFloat(val, 64); err == nil {
			return v
		}
		return val
	default:
		return val
	}
}
