package input

import (
	"fmt"
	"io"
	"os"

	"github.com/kiki-ki/go-qo/internal/db"
	"github.com/kiki-ki/go-qo/internal/parser"
)

// LoaderOptions configures loader behavior.
type LoaderOptions struct {
	NoHeader bool // CSV: treat first row as data, not header
}

// Loader handles loading data into the database.
type Loader struct {
	db      *db.DB
	format  Format
	options *LoaderOptions
}

// NewLoader creates a new Loader.
func NewLoader(database *db.DB, format Format, options *LoaderOptions) *Loader {
	if options == nil {
		options = &LoaderOptions{}
	}
	return &Loader{
		db:      database,
		format:  format,
		options: options,
	}
}

// HasStdinData checks if there's data available on stdin.
func HasStdinData() (bool, error) {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return false, fmt.Errorf("failed to stat stdin: %w", err)
	}
	return (stat.Mode() & os.ModeCharDevice) == 0, nil
}

// LoadStdin loads data from stdin into the database.
func (l *Loader) LoadStdin(tableName string) error {
	return l.LoadReader(os.Stdin, tableName)
}

// LoadReader loads data from an io.Reader into the database.
func (l *Loader) LoadReader(r io.Reader, tableName string) error {
	data, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}

	parsed, err := l.parseBytes(data)
	if err != nil {
		return fmt.Errorf("failed to parse input: %w", err)
	}

	if err := l.db.LoadData(tableName, parsed); err != nil {
		return fmt.Errorf("failed to load data: %w", err)
	}

	return nil
}

// LoadFiles loads data from files into the database.
func (l *Loader) LoadFiles(filePaths []string) error {
	for _, path := range filePaths {
		parsed, err := l.parseFile(path)
		if err != nil {
			return fmt.Errorf("failed to parse %s: %w", path, err)
		}

		tableName := db.TableNameFromPath(path)

		if err := l.db.LoadData(tableName, parsed); err != nil {
			return fmt.Errorf("failed to load table %s: %w", tableName, err)
		}
	}

	return nil
}

// parseBytes parses byte data based on the format.
func (l *Loader) parseBytes(data []byte) (*parser.ParsedData, error) {
	switch l.format {
	case FormatJSON:
		return parser.ParseJSONBytes(data)
	case FormatCSV:
		return parser.ParseCSVBytes(data, parser.CSVOptions{NoHeader: l.options.NoHeader})
	case FormatTSV:
		return parser.ParseCSVBytes(data, parser.CSVOptions{NoHeader: l.options.NoHeader, Delimiter: '\t'})
	default:
		return nil, fmt.Errorf("unsupported format: %s", l.format)
	}
}

// parseFile parses a file based on the format.
func (l *Loader) parseFile(path string) (*parser.ParsedData, error) {
	switch l.format {
	case FormatJSON:
		return parser.ParseFile(path)
	case FormatCSV:
		return parser.ParseCSVFile(path, parser.CSVOptions{NoHeader: l.options.NoHeader})
	case FormatTSV:
		return parser.ParseCSVFile(path, parser.CSVOptions{NoHeader: l.options.NoHeader, Delimiter: '\t'})
	default:
		return nil, fmt.Errorf("unsupported format: %s", l.format)
	}
}
