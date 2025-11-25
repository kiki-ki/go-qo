package parser

import (
	"fmt"
	"os"

	"github.com/tidwall/gjson"
)

// JSONParser implements Parser interface for JSON files.
type JSONParser struct{}

// init registers the JSON parser.
func init() {
	Register(&JSONParser{})
}

// SupportedExtensions returns the file extensions this parser handles.
func (p *JSONParser) SupportedExtensions() []string {
	return []string{".json"}
}

// Parse parses a JSON file into ParsedData.
func (p *JSONParser) Parse(path string) (*ParsedData, error) {
	fileBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", path, err)
	}

	if !gjson.ValidBytes(fileBytes) {
		return nil, fmt.Errorf("invalid JSON format in %s", path)
	}

	result := gjson.ParseBytes(fileBytes)
	items := p.extractItems(result)

	if len(items) == 0 {
		return nil, fmt.Errorf("empty JSON data in %s", path)
	}

	columns := p.extractColumns(items)
	rows := p.extractRows(items, columns)

	return &ParsedData{
		Columns: columns,
		Rows:    rows,
	}, nil
}

// extractItems converts the JSON result to a slice of items.
func (p *JSONParser) extractItems(result gjson.Result) []gjson.Result {
	if result.IsArray() {
		return result.Array()
	}
	return []gjson.Result{result}
}

// extractColumns extracts column definitions from items.
func (p *JSONParser) extractColumns(items []gjson.Result) []Column {
	keyMap := make(map[string]int) // key -> index in columns
	typeMap := make(map[string]DataType)
	var columns []Column

	for _, item := range items {
		item.ForEach(func(key, value gjson.Result) bool {
			k := key.String()
			newType := p.inferType(value)

			if idx, exists := keyMap[k]; exists {
				// Upgrade type if needed (type widening)
				columns[idx].Type = p.widenType(typeMap[k], newType)
				typeMap[k] = columns[idx].Type
			} else {
				keyMap[k] = len(columns)
				typeMap[k] = newType
				columns = append(columns, Column{Name: k, Type: newType})
			}
			return true
		})
	}

	// Handle primitive arrays (e.g., [1, 2, 3])
	if len(columns) == 0 {
		columns = []Column{{Name: "value", Type: TypeText}}
	}

	return columns
}

// inferType infers the DataType from a gjson.Result.
func (p *JSONParser) inferType(val gjson.Result) DataType {
	switch val.Type {
	case gjson.String:
		return TypeText
	case gjson.Number:
		f := val.Float()
		if float64(int64(f)) == f {
			return TypeInteger
		}
		return TypeReal
	case gjson.True, gjson.False:
		return TypeBoolean
	case gjson.JSON:
		return TypeJSON
	case gjson.Null:
		return TypeNull
	default:
		return TypeText
	}
}

// widenType returns the wider type when two types conflict.
func (p *JSONParser) widenType(existing, new DataType) DataType {
	if existing == new {
		return existing
	}
	if existing == TypeNull {
		return new
	}
	if new == TypeNull {
		return existing
	}
	// Integer + Real = Real
	if (existing == TypeInteger && new == TypeReal) ||
		(existing == TypeReal && new == TypeInteger) {
		return TypeReal
	}
	// Default to Text for incompatible types
	return TypeText
}

// extractRows extracts row data from items based on columns.
func (p *JSONParser) extractRows(items []gjson.Result, columns []Column) [][]any {
	rows := make([][]any, 0, len(items))

	for _, item := range items {
		row := make([]any, len(columns))
		for i, col := range columns {
			row[i] = p.extractValue(item.Get(col.Name))
		}
		rows = append(rows, row)
	}

	return rows
}

// extractValue converts a gjson.Result to a Go value.
func (p *JSONParser) extractValue(val gjson.Result) any {
	if !val.Exists() {
		return nil
	}

	switch val.Type {
	case gjson.String:
		return val.String()
	case gjson.Number:
		f := val.Float()
		if float64(int64(f)) == f {
			return val.Int()
		}
		return f
	case gjson.True:
		return true
	case gjson.False:
		return false
	case gjson.Null:
		return nil
	case gjson.JSON:
		return val.Raw
	default:
		return val.String()
	}
}

// ParseJSON is a convenience function for backward compatibility.
// Deprecated: Use ParseFile or JSONParser.Parse instead.
func ParseJSON(path string) (*ParsedData, error) {
	return (&JSONParser{}).Parse(path)
}
