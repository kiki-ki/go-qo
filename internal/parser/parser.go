package parser

import (
	"fmt"
	"path/filepath"
	"strings"
)

// DataType represents the SQL column type.
type DataType int

const (
	TypeText DataType = iota
	TypeInteger
	TypeReal
	TypeBoolean
	TypeJSON
	TypeNull
)

// String returns the SQL type name.
func (dt DataType) String() string {
	switch dt {
	case TypeInteger:
		return "INTEGER"
	case TypeReal:
		return "REAL"
	case TypeBoolean:
		return "INTEGER" // SQLite stores booleans as integers
	default:
		return "TEXT" // TypeText, TypeJSON, TypeNull all map to TEXT
	}
}

// Column represents a table column with its name and type.
type Column struct {
	Name string
	Type DataType
}

// ParsedData holds parsed data from a file.
type ParsedData struct {
	Columns []Column
	Rows    [][]any
}

// ColumnNames returns column names as a string slice.
func (pd *ParsedData) ColumnNames() []string {
	names := make([]string, len(pd.Columns))
	for i, col := range pd.Columns {
		names[i] = col.Name
	}
	return names
}

// Parser defines the interface for file parsers.
type Parser interface {
	Parse(path string) (*ParsedData, error)
	SupportedExtensions() []string
}

var registry = make(map[string]Parser)

// Register adds a parser to the registry.
func Register(p Parser) {
	for _, ext := range p.SupportedExtensions() {
		registry[strings.ToLower(ext)] = p
	}
}

// GetParser returns the parser for the given file path.
func GetParser(path string) (Parser, error) {
	ext := strings.ToLower(filepath.Ext(path))
	if p, ok := registry[ext]; ok {
		return p, nil
	}
	return nil, fmt.Errorf("unsupported file format: %s", ext)
}

// ParseFile parses a file using the appropriate parser.
func ParseFile(path string) (*ParsedData, error) {
	p, err := GetParser(path)
	if err != nil {
		return nil, err
	}
	return p.Parse(path)
}
