package parser

import (
	"bytes"
	"encoding/json"
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
	return []string{".json", ".jsonl", ".ndjson"}
}

// Parse parses a JSON file into ParsedData.
func (p *JSONParser) Parse(path string) (*ParsedData, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", path, err)
	}
	return p.ParseBytes(data)
}

// ParseBytes parses JSON or JSON Lines from a byte slice.
func (p *JSONParser) ParseBytes(data []byte) (*ParsedData, error) {
	var items []gjson.Result
	var err error

	if gjson.ValidBytes(data) {
		items = p.parseJSON(data)
	} else {
		items, err = p.parseJSONLines(data)
		if err != nil {
			return nil, err
		}
	}

	if len(items) == 0 {
		return nil, fmt.Errorf("empty JSON data")
	}

	columns := p.extractColumns(items)
	rows := p.extractRows(items, columns)

	return &ParsedData{
		Columns: columns,
		Rows:    rows,
	}, nil
}

// parseJSON parses standard JSON format.
func (p *JSONParser) parseJSON(data []byte) []gjson.Result {
	result := gjson.ParseBytes(data)
	if result.IsArray() {
		return result.Array()
	}
	return []gjson.Result{result}
}

// parseJSONLines parses JSON Lines format.
func (p *JSONParser) parseJSONLines(data []byte) ([]gjson.Result, error) {
	var items []gjson.Result
	var parseErr error
	gjson.ForEachLine(string(data), func(line gjson.Result) bool {
		if !gjson.Valid(line.Raw) {
			parseErr = fmt.Errorf("invalid JSON format")
			return false
		}
		items = append(items, line)
		return true
	})
	return items, parseErr
}

// extractColumns extracts column definitions from items.
func (p *JSONParser) extractColumns(items []gjson.Result) []Column {
	keyMap := make(map[string]int)
	typeMap := make(map[string]DataType)
	var columns []Column

	for _, item := range items {
		item.ForEach(func(key, value gjson.Result) bool {
			k := key.String()
			newType := p.inferType(value)

			if idx, exists := keyMap[k]; exists {
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
		if float64(int64(val.Float())) == val.Float() {
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
	if existing == new || new == TypeNull {
		return existing
	}
	if existing == TypeNull {
		return new
	}
	if (existing == TypeInteger && new == TypeReal) || (existing == TypeReal && new == TypeInteger) {
		return TypeReal
	}
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
		if float64(int64(val.Float())) == val.Float() {
			return val.Int()
		}
		return val.Float()
	case gjson.True:
		return true
	case gjson.False:
		return false
	case gjson.Null:
		return nil
	case gjson.JSON:
		// Compact the JSON to remove unnecessary whitespace and newlines
		var buf bytes.Buffer
		if err := json.Compact(&buf, []byte(val.Raw)); err != nil {
			return val.Raw
		}
		return buf.String()
	default:
		return val.String()
	}
}

// ParseJSONBytes parses JSON from a byte slice.
func ParseJSONBytes(data []byte) (*ParsedData, error) {
	return (&JSONParser{}).ParseBytes(data)
}
