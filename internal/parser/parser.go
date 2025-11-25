// Package parser provides functionality to parse various file formats
// into a common intermediate representation for SQL querying.
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
	case TypeJSON:
		return "TEXT"
	case TypeNull:
		return "TEXT"
	default:
		return "TEXT"
	}
}

// Column represents a table column with its name and inferred type.
type Column struct {
	Name string
	Type DataType
}

// ParsedData holds the parsed data from a file.
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

// registry holds registered parsers.
var registry = make(map[string]Parser)

// Register adds a parser to the registry.
func Register(p Parser) {
	for _, ext := range p.SupportedExtensions() {
		registry[strings.ToLower(ext)] = p
	}
}

// GetParser returns the appropriate parser for the given file path.
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
