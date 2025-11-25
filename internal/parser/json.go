package parser

import (
	"fmt"
	"os"

	"github.com/tidwall/gjson"
)

type ParsedData struct {
	Headers []string
	Rows    [][]interface{}
}

// Parse JSON to intermediate representation
func ParseJSON(path string) (*ParsedData, error) {
	fileBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file error: %w", err)
	}
	if !gjson.ValidBytes(fileBytes) {
		return nil, fmt.Errorf("invalid json format")
	}

	result := gjson.ParseBytes(fileBytes)

	var items []gjson.Result
	if result.IsArray() {
		items = result.Array()
	} else {
		items = []gjson.Result{result}
	}

	if len(items) == 0 {
		return nil, fmt.Errorf("empty json data")
	}

	keyMap := make(map[string]bool)
	var headers []string

	for _, item := range items {
		item.ForEach(func(key, value gjson.Result) bool {
			k := key.String()
			if !keyMap[k] {
				keyMap[k] = true
				headers = append(headers, k)
			}
			return true
		})
	}

	if len(headers) == 0 {
		headers = []string{"value"}
	}

	var rows [][]interface{}
	for _, item := range items {
		row := make([]interface{}, len(headers))
		for i, h := range headers {
			val := item.Get(h)
			if !val.Exists() {
				row[i] = nil
			} else {
				// convert to Go types
				switch val.Type {
				case gjson.String:
					row[i] = val.String()
				case gjson.Number:
					f := val.Float()
					if float64(int64(f)) == f {
						row[i] = val.Int()
					} else {
						row[i] = f
					}
				case gjson.True:
					row[i] = true
				case gjson.False:
					row[i] = false
				case gjson.Null:
					row[i] = nil
				case gjson.JSON:
					row[i] = val.Raw
				default:
					row[i] = val.String()
				}
			}
		}
		rows = append(rows, row)
	}

	return &ParsedData{
		Headers: headers,
		Rows:    rows,
	}, nil
}
