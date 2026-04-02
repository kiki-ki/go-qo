package parser

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

// ParseCSVBytes parses CSV data from a byte slice.
func ParseCSVBytes(data []byte, options CSVOptions) (*ParsedData, error) {
	p := &CSVParser{Options: options}
	return p.ParseBytes(data)
}
