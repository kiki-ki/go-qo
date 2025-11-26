package input

import (
	"fmt"
	"io"
	"os"

	"github.com/kiki-ki/go-qo/internal/db"
	"github.com/kiki-ki/go-qo/internal/parser"
)

// Loader handles loading data into the database.
type Loader struct {
	db      *db.DB
	format  Format
	verbose bool
	output  io.Writer // for verbose output
}

// NewLoader creates a new Loader.
func NewLoader(database *db.DB, format Format, verbose bool) *Loader {
	return &Loader{
		db:      database,
		format:  format,
		verbose: verbose,
		output:  os.Stderr,
	}
}

// LoadStdin loads data from stdin into the database.
func (l *Loader) LoadStdin(tableName string) error {
	if l.verbose {
		fmt.Fprintln(l.output, "Loading stdin...")
	}

	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return fmt.Errorf("failed to read stdin: %w", err)
	}

	parsed, err := l.parseBytes(data)
	if err != nil {
		return fmt.Errorf("failed to parse stdin: %w", err)
	}

	if err := l.db.LoadData(tableName, parsed); err != nil {
		return fmt.Errorf("failed to load stdin data: %w", err)
	}

	if l.verbose {
		fmt.Fprintf(l.output, "  %s (%d rows, %d columns)\n",
			tableName, len(parsed.Rows), len(parsed.Columns))
	}

	return nil
}

// LoadFiles loads data from files into the database.
func (l *Loader) LoadFiles(filePaths []string) error {
	if l.verbose {
		fmt.Fprintln(l.output, "Loading files...")
	}

	for _, path := range filePaths {
		parsed, err := l.parseFile(path)
		if err != nil {
			return fmt.Errorf("failed to parse %s: %w", path, err)
		}

		tableName := db.TableNameFromPath(path)

		if err := l.db.LoadData(tableName, parsed); err != nil {
			return fmt.Errorf("failed to load table %s: %w", tableName, err)
		}

		if l.verbose {
			fmt.Fprintf(l.output, "  %s → %s (%d rows, %d columns)\n",
				path, tableName, len(parsed.Rows), len(parsed.Columns))
		}
	}

	return nil
}

// parseBytes parses byte data based on the format.
func (l *Loader) parseBytes(data []byte) (*parser.ParsedData, error) {
	switch l.format {
	case FormatJSON:
		return parser.ParseJSONBytes(data)
	default:
		return nil, fmt.Errorf("unsupported format: %s", l.format)
	}
}

// parseFile parses a file based on the format.
func (l *Loader) parseFile(path string) (*parser.ParsedData, error) {
	switch l.format {
	case FormatJSON:
		return parser.ParseFile(path)
	default:
		return nil, fmt.Errorf("unsupported format: %s", l.format)
	}
}
