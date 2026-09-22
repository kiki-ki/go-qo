package parser_test

import (
	"testing"

	"github.com/kiki-ki/go-qo/internal/parser"
)

func TestJSONParser_ParseBytes(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantRows    int
		wantCols    int
		wantErr     bool
		checkValues func(t *testing.T, data *parser.ParsedData)
	}{
		{
			name:     "array of objects",
			input:    `[{"id": 1, "name": "Alice"}, {"id": 2, "name": "Bob"}]`,
			wantRows: 2,
			wantCols: 2,
			checkValues: func(t *testing.T, data *parser.ParsedData) {
				if data.Rows[0][1] != "Alice" {
					t.Errorf("expected Alice, got %v", data.Rows[0][1])
				}
			},
		},
		{
			name:     "single object",
			input:    `{"id": 42, "value": 3.14}`,
			wantRows: 1,
			wantCols: 2,
			checkValues: func(t *testing.T, data *parser.ParsedData) {
				if data.Columns[0].Type != parser.TypeInteger {
					t.Errorf("expected INTEGER, got %v", data.Columns[0].Type)
				}
				if data.Columns[1].Type != parser.TypeReal {
					t.Errorf("expected REAL, got %v", data.Columns[1].Type)
				}
			},
		},
		{
			name:     "null values",
			input:    `[{"id": 1, "name": "Alice"}, {"id": 2, "name": null}]`,
			wantRows: 2,
			wantCols: 2,
			checkValues: func(t *testing.T, data *parser.ParsedData) {
				if data.Rows[1][1] != nil {
					t.Errorf("expected nil, got %v", data.Rows[1][1])
				}
			},
		},
		{
			name:     "missing fields",
			input:    `[{"id": 1, "name": "Alice"}, {"id": 2}]`,
			wantRows: 2,
			wantCols: 2,
			checkValues: func(t *testing.T, data *parser.ParsedData) {
				if data.Rows[1][1] != nil {
					t.Errorf("expected nil for missing field, got %v", data.Rows[1][1])
				}
			},
		},
		{
			name:     "nested JSON",
			input:    `[{"id": 1, "meta": {"key": "value"}, "tags": ["a", "b"]}]`,
			wantRows: 1,
			wantCols: 3,
			checkValues: func(t *testing.T, data *parser.ParsedData) {
				for _, col := range data.Columns {
					if col.Name == "meta" && col.Type != parser.TypeJSON {
						t.Errorf("expected JSON type for meta")
					}
					if col.Name == "tags" && col.Type != parser.TypeJSON {
						t.Errorf("expected JSON type for tags")
					}
				}
			},
		},
		{
			name:     "nested JSON is compacted",
			input:    "[{\"id\": 1, \"meta\": {\n  \"key\": \"value\"\n}}]",
			wantRows: 1,
			wantCols: 2,
			checkValues: func(t *testing.T, data *parser.ParsedData) {
				for i, col := range data.Columns {
					if col.Name == "meta" {
						got, ok := data.Rows[0][i].(string)
						if !ok {
							t.Errorf("expected string, got %T", data.Rows[0][i])
							return
						}
						want := `{"key":"value"}`
						if got != want {
							t.Errorf("nested JSON not compacted: got %q, want %q", got, want)
						}
					}
				}
			},
		},
		{
			name:     "type widening int to real",
			input:    `[{"value": 1}, {"value": 2.5}]`,
			wantRows: 2,
			wantCols: 1,
			checkValues: func(t *testing.T, data *parser.ParsedData) {
				if data.Columns[0].Type != parser.TypeReal {
					t.Errorf("expected REAL after widening, got %v", data.Columns[0].Type)
				}
			},
		},
		{
			name:     "all types",
			input:    `[{"int": 1, "float": 3.14, "bool": true, "str": "hello", "null": null}]`,
			wantRows: 1,
			wantCols: 5,
			checkValues: func(t *testing.T, data *parser.ParsedData) {
				types := map[string]parser.DataType{
					"int": parser.TypeInteger, "float": parser.TypeReal,
					"bool": parser.TypeBoolean, "str": parser.TypeText, "null": parser.TypeNull,
				}
				for _, col := range data.Columns {
					if expected, ok := types[col.Name]; ok && col.Type != expected {
						t.Errorf("column %s: expected %v, got %v", col.Name, expected, col.Type)
					}
				}
			},
		},
		{
			name:     "JSON Lines format",
			input:    "{\"id\":1,\"name\":\"Alice\"}\n{\"id\":2,\"name\":\"Bob\"}\n",
			wantRows: 2,
			wantCols: 2,
			checkValues: func(t *testing.T, data *parser.ParsedData) {
				if data.Rows[0][1] != "Alice" {
					t.Errorf("expected Alice, got %v", data.Rows[0][1])
				}
				if data.Rows[1][1] != "Bob" {
					t.Errorf("expected Bob, got %v", data.Rows[1][1])
				}
			},
		},
		{
			name:     "JSON Lines with nested objects",
			input:    "{\"id\":1,\"meta\":{\"key\":\"value\"}}\n{\"id\":2,\"meta\":{\"key\":\"other\"}}\n",
			wantRows: 2,
			wantCols: 2,
			checkValues: func(t *testing.T, data *parser.ParsedData) {
				for i, col := range data.Columns {
					if col.Name == "meta" && col.Type != parser.TypeJSON {
						t.Errorf("expected JSON type for meta")
					}
					if col.Name == "meta" {
						if data.Rows[0][i] != `{"key":"value"}` {
							t.Errorf("expected nested JSON, got %v", data.Rows[0][i])
						}
					}
				}
			},
		},
		{
			name:     "array of numbers",
			input:    `[1, 2, 3]`,
			wantRows: 3,
			wantCols: 1,
			checkValues: func(t *testing.T, data *parser.ParsedData) {
				if data.Columns[0].Name != "value" {
					t.Errorf("expected column name value, got %q", data.Columns[0].Name)
				}
				if data.Columns[0].Type != parser.TypeInteger {
					t.Errorf("expected INTEGER, got %v", data.Columns[0].Type)
				}
				for i, want := range []any{int64(1), int64(2), int64(3)} {
					if data.Rows[i][0] != want {
						t.Errorf("row %d: expected %v, got %v", i, want, data.Rows[i][0])
					}
				}
			},
		},
		{
			name:     "array of strings",
			input:    `["a", "b"]`,
			wantRows: 2,
			wantCols: 1,
			checkValues: func(t *testing.T, data *parser.ParsedData) {
				if data.Rows[0][0] != "a" || data.Rows[1][0] != "b" {
					t.Errorf("expected a and b, got %v and %v", data.Rows[0][0], data.Rows[1][0])
				}
			},
		},
		{
			name:     "array of arrays",
			input:    `[[1, 2], [3, 4]]`,
			wantRows: 2,
			wantCols: 1,
			checkValues: func(t *testing.T, data *parser.ParsedData) {
				if data.Columns[0].Type != parser.TypeJSON {
					t.Errorf("expected JSON type, got %v", data.Columns[0].Type)
				}
				if data.Rows[0][0] != "[1,2]" {
					t.Errorf("expected [1,2], got %v", data.Rows[0][0])
				}
			},
		},
		{
			name:     "scalar array widens int to real",
			input:    `[1, 2.5]`,
			wantRows: 2,
			wantCols: 1,
			checkValues: func(t *testing.T, data *parser.ParsedData) {
				if data.Columns[0].Type != parser.TypeReal {
					t.Errorf("expected REAL after widening, got %v", data.Columns[0].Type)
				}
			},
		},
		{
			name:     "scalars mixed with objects",
			input:    `[1, {"a": 2}]`,
			wantRows: 2,
			wantCols: 2,
			checkValues: func(t *testing.T, data *parser.ParsedData) {
				cols := data.ColumnNames()
				if cols[0] != "value" || cols[1] != "a" {
					t.Errorf("expected columns [value a], got %v", cols)
				}
				if data.Rows[0][0] != int64(1) || data.Rows[0][1] != nil {
					t.Errorf("scalar row: got %v", data.Rows[0])
				}
				if data.Rows[1][0] != nil || data.Rows[1][1] != int64(2) {
					t.Errorf("object row: got %v", data.Rows[1])
				}
			},
		},
		{
			name:     "top level scalar",
			input:    `"hello"`,
			wantRows: 1,
			wantCols: 1,
			checkValues: func(t *testing.T, data *parser.ParsedData) {
				if data.Rows[0][0] != "hello" {
					t.Errorf("expected hello, got %v", data.Rows[0][0])
				}
			},
		},
		{
			name:     "JSON Lines of scalars",
			input:    "1\n2\n3\n",
			wantRows: 3,
			wantCols: 1,
			checkValues: func(t *testing.T, data *parser.ParsedData) {
				if data.Rows[2][0] != int64(3) {
					t.Errorf("expected 3, got %v", data.Rows[2][0])
				}
			},
		},
		{
			name:    "invalid JSON",
			input:   `{invalid}`,
			wantErr: true,
		},
		{
			name:    "empty array",
			input:   `[]`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := parser.ParseJSONBytes([]byte(tt.input))
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(data.Rows) != tt.wantRows {
				t.Errorf("expected %d rows, got %d", tt.wantRows, len(data.Rows))
			}
			if len(data.Columns) != tt.wantCols {
				t.Errorf("expected %d columns, got %d", tt.wantCols, len(data.Columns))
			}
			if tt.checkValues != nil {
				tt.checkValues(t, data)
			}
		})
	}
}

func TestParsedData_ColumnNames(t *testing.T) {
	pd := &parser.ParsedData{
		Columns: []parser.Column{{Name: "id"}, {Name: "name"}},
	}
	names := pd.ColumnNames()
	if len(names) != 2 || names[0] != "id" || names[1] != "name" {
		t.Errorf("expected [id, name], got %v", names)
	}
}

func TestDataType_String(t *testing.T) {
	tests := []struct {
		dt   parser.DataType
		want string
	}{
		{parser.TypeText, "TEXT"},
		{parser.TypeInteger, "INTEGER"},
		{parser.TypeReal, "REAL"},
		{parser.TypeBoolean, "INTEGER"},
		{parser.TypeJSON, "TEXT"},
		{parser.TypeNull, "TEXT"},
	}
	for _, tt := range tests {
		if got := tt.dt.String(); got != tt.want {
			t.Errorf("DataType(%d).String() = %q, want %q", tt.dt, got, tt.want)
		}
	}
}
