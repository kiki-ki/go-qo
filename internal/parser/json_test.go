package parser

import (
	"os"
	"path/filepath"
	"testing"
)

func TestJSONParser_ParseBytes(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantRows    int
		wantCols    int
		wantErr     bool
		checkValues func(t *testing.T, data *ParsedData)
	}{
		{
			name:     "array of objects",
			input:    `[{"id": 1, "name": "Alice"}, {"id": 2, "name": "Bob"}]`,
			wantRows: 2,
			wantCols: 2,
			checkValues: func(t *testing.T, data *ParsedData) {
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
			checkValues: func(t *testing.T, data *ParsedData) {
				if data.Columns[0].Type != TypeInteger {
					t.Errorf("expected INTEGER, got %v", data.Columns[0].Type)
				}
				if data.Columns[1].Type != TypeReal {
					t.Errorf("expected REAL, got %v", data.Columns[1].Type)
				}
			},
		},
		{
			name:     "null values",
			input:    `[{"id": 1, "name": "Alice"}, {"id": 2, "name": null}]`,
			wantRows: 2,
			wantCols: 2,
			checkValues: func(t *testing.T, data *ParsedData) {
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
			checkValues: func(t *testing.T, data *ParsedData) {
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
			checkValues: func(t *testing.T, data *ParsedData) {
				for _, col := range data.Columns {
					if col.Name == "meta" && col.Type != TypeJSON {
						t.Errorf("expected JSON type for meta")
					}
					if col.Name == "tags" && col.Type != TypeJSON {
						t.Errorf("expected JSON type for tags")
					}
				}
			},
		},
		{
			name:     "type widening int to real",
			input:    `[{"value": 1}, {"value": 2.5}]`,
			wantRows: 2,
			wantCols: 1,
			checkValues: func(t *testing.T, data *ParsedData) {
				if data.Columns[0].Type != TypeReal {
					t.Errorf("expected REAL after widening, got %v", data.Columns[0].Type)
				}
			},
		},
		{
			name:     "all types",
			input:    `[{"int": 1, "float": 3.14, "bool": true, "str": "hello", "null": null}]`,
			wantRows: 1,
			wantCols: 5,
			checkValues: func(t *testing.T, data *ParsedData) {
				types := map[string]DataType{
					"int": TypeInteger, "float": TypeReal,
					"bool": TypeBoolean, "str": TypeText, "null": TypeNull,
				}
				for _, col := range data.Columns {
					if expected, ok := types[col.Name]; ok && col.Type != expected {
						t.Errorf("column %s: expected %v, got %v", col.Name, expected, col.Type)
					}
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

	p := &JSONParser{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := p.ParseBytes([]byte(tt.input))
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

func TestJSONParser_Parse_File(t *testing.T) {
	content := `[{"id": 1, "name": "Alice"}, {"id": 2, "name": "Bob"}]`
	path := createTempJSON(t, content)

	p := &JSONParser{}
	data, err := p.Parse(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(data.Rows) != 2 {
		t.Errorf("expected 2 rows, got %d", len(data.Rows))
	}
}

func TestJSONParser_Parse_FileNotFound(t *testing.T) {
	p := &JSONParser{}
	_, err := p.Parse("/nonexistent/file.json")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestParseFile(t *testing.T) {
	path := createTempJSON(t, `[{"test": 1}]`)
	data, err := ParseFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(data.Rows) != 1 {
		t.Errorf("expected 1 row, got %d", len(data.Rows))
	}
}

func TestGetParser_UnsupportedFormat(t *testing.T) {
	_, err := GetParser("file.xml")
	if err == nil {
		t.Error("expected error for unsupported format")
	}
}

func TestParsedData_ColumnNames(t *testing.T) {
	pd := &ParsedData{
		Columns: []Column{{Name: "id"}, {Name: "name"}},
	}
	names := pd.ColumnNames()
	if len(names) != 2 || names[0] != "id" || names[1] != "name" {
		t.Errorf("expected [id, name], got %v", names)
	}
}

func TestDataType_String(t *testing.T) {
	tests := []struct {
		dt   DataType
		want string
	}{
		{TypeText, "TEXT"},
		{TypeInteger, "INTEGER"},
		{TypeReal, "REAL"},
		{TypeBoolean, "INTEGER"},
		{TypeJSON, "TEXT"},
		{TypeNull, "TEXT"},
	}
	for _, tt := range tests {
		if got := tt.dt.String(); got != tt.want {
			t.Errorf("DataType(%d).String() = %q, want %q", tt.dt, got, tt.want)
		}
	}
}

func createTempJSON(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.json")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	return path
}
