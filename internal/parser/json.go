package parser

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/tidwall/gjson"
)

// scalarColumnName is the column non-object items (numbers, strings, arrays)
// are stored under, since they have no key of their own.
const scalarColumnName = "value"

// JSONParser parses JSON and JSON Lines input.
type JSONParser struct{}

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
	var columns []Column

	addKey := func(k string, t DataType) {
		if idx, exists := keyMap[k]; exists {
			columns[idx].Type = p.widenType(columns[idx].Type, t)
			return
		}
		keyMap[k] = len(columns)
		columns = append(columns, Column{Name: k, Type: t})
	}

	for _, item := range items {
		// gjson runs ForEach once on a non-object, yielding an empty key, so
		// scalars and arrays have to be routed to the scalar column instead.
		if !item.IsObject() {
			addKey(scalarColumnName, p.inferType(item))
			continue
		}
		item.ForEach(func(key, value gjson.Result) bool {
			addKey(key.String(), p.inferType(value))
			return true
		})
	}

	if len(columns) == 0 {
		columns = []Column{{Name: scalarColumnName, Type: TypeText}}
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
	index := make(map[string]int, len(columns))
	for i, col := range columns {
		index[col.Name] = i
	}

	rows := make([][]any, 0, len(items))
	for _, item := range items {
		row := make([]any, len(columns))

		if !item.IsObject() {
			if i, ok := index[scalarColumnName]; ok {
				row[i] = p.extractValue(item)
			}
			rows = append(rows, row)
			continue
		}

		// ForEach reports raw keys. Get would read the column name as a gjson
		// path instead, losing or mismatching any key holding "." or "\\".
		item.ForEach(func(key, value gjson.Result) bool {
			if i, ok := index[key.String()]; ok {
				row[i] = p.extractValue(value)
			}
			return true
		})
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
