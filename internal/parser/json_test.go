package parser

import (
	"os"
	"path/filepath"
	"testing"
)

func TestJSONParser_Parse_Array(t *testing.T) {
	content := `[
		{"id": 1, "name": "Alice", "active": true},
		{"id": 2, "name": "Bob", "active": false}
	]`

	path := createTempJSON(t, content)
	defer os.Remove(path)

	p := &JSONParser{}
	data, err := p.Parse(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check columns
	if len(data.Columns) != 3 {
		t.Errorf("expected 3 columns, got %d", len(data.Columns))
	}

	expectedCols := []string{"id", "name", "active"}
	for i, expected := range expectedCols {
		if data.Columns[i].Name != expected {
			t.Errorf("column %d: expected %q, got %q", i, expected, data.Columns[i].Name)
		}
	}

	// Check types
	if data.Columns[0].Type != TypeInteger {
		t.Errorf("expected id to be INTEGER, got %v", data.Columns[0].Type)
	}
	if data.Columns[1].Type != TypeText {
		t.Errorf("expected name to be TEXT, got %v", data.Columns[1].Type)
	}
	if data.Columns[2].Type != TypeBoolean {
		t.Errorf("expected active to be BOOLEAN, got %v", data.Columns[2].Type)
	}

	// Check rows
	if len(data.Rows) != 2 {
		t.Errorf("expected 2 rows, got %d", len(data.Rows))
	}

	// Check first row values
	if data.Rows[0][0] != int64(1) {
		t.Errorf("expected id=1, got %v", data.Rows[0][0])
	}
	if data.Rows[0][1] != "Alice" {
		t.Errorf("expected name=Alice, got %v", data.Rows[0][1])
	}
	if data.Rows[0][2] != true {
		t.Errorf("expected active=true, got %v", data.Rows[0][2])
	}
}

func TestJSONParser_Parse_Object(t *testing.T) {
	content := `{"id": 42, "value": 3.14}`

	path := createTempJSON(t, content)
	defer os.Remove(path)

	p := &JSONParser{}
	data, err := p.Parse(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(data.Rows) != 1 {
		t.Errorf("expected 1 row for single object, got %d", len(data.Rows))
	}

	if data.Columns[0].Type != TypeInteger {
		t.Errorf("expected id to be INTEGER, got %v", data.Columns[0].Type)
	}
	if data.Columns[1].Type != TypeReal {
		t.Errorf("expected value to be REAL, got %v", data.Columns[1].Type)
	}
}

func TestJSONParser_Parse_NullValues(t *testing.T) {
	content := `[
		{"id": 1, "name": "Alice"},
		{"id": 2, "name": null}
	]`

	path := createTempJSON(t, content)
	defer os.Remove(path)

	p := &JSONParser{}
	data, err := p.Parse(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if data.Rows[1][1] != nil {
		t.Errorf("expected null value, got %v", data.Rows[1][1])
	}
}

func TestJSONParser_Parse_MissingFields(t *testing.T) {
	content := `[
		{"id": 1, "name": "Alice"},
		{"id": 2}
	]`

	path := createTempJSON(t, content)
	defer os.Remove(path)

	p := &JSONParser{}
	data, err := p.Parse(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Second row should have nil for missing "name"
	if data.Rows[1][1] != nil {
		t.Errorf("expected nil for missing field, got %v", data.Rows[1][1])
	}
}

func TestJSONParser_Parse_NestedJSON(t *testing.T) {
	content := `[
		{"id": 1, "meta": {"key": "value"}}
	]`

	path := createTempJSON(t, content)
	defer os.Remove(path)

	p := &JSONParser{}
	data, err := p.Parse(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Nested JSON should be stored as raw string
	if data.Columns[1].Type != TypeJSON {
		t.Errorf("expected nested object to be JSON type, got %v", data.Columns[1].Type)
	}

	expected := `{"key": "value"}`
	if data.Rows[0][1] != expected {
		t.Errorf("expected %q, got %q", expected, data.Rows[0][1])
	}
}

func TestJSONParser_Parse_TypeWidening(t *testing.T) {
	content := `[
		{"value": 1},
		{"value": 2.5}
	]`

	path := createTempJSON(t, content)
	defer os.Remove(path)

	p := &JSONParser{}
	data, err := p.Parse(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Integer + Real should widen to Real
	if data.Columns[0].Type != TypeReal {
		t.Errorf("expected type widening to REAL, got %v", data.Columns[0].Type)
	}
}

func TestJSONParser_Parse_InvalidJSON(t *testing.T) {
	content := `{invalid json}`

	path := createTempJSON(t, content)
	defer os.Remove(path)

	p := &JSONParser{}
	_, err := p.Parse(path)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestJSONParser_Parse_EmptyArray(t *testing.T) {
	content := `[]`

	path := createTempJSON(t, content)
	defer os.Remove(path)

	p := &JSONParser{}
	_, err := p.Parse(path)
	if err == nil {
		t.Error("expected error for empty array")
	}
}

func TestJSONParser_Parse_FileNotFound(t *testing.T) {
	p := &JSONParser{}
	_, err := p.Parse("/nonexistent/file.json")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestJSONParser_SupportedExtensions(t *testing.T) {
	p := &JSONParser{}
	exts := p.SupportedExtensions()

	if len(exts) != 1 || exts[0] != ".json" {
		t.Errorf("expected [.json], got %v", exts)
	}
}

func TestParseFile(t *testing.T) {
	content := `[{"test": 1}]`

	path := createTempJSON(t, content)
	defer os.Remove(path)

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
		Columns: []Column{
			{Name: "id", Type: TypeInteger},
			{Name: "name", Type: TypeText},
		},
	}

	names := pd.ColumnNames()
	if len(names) != 2 {
		t.Errorf("expected 2 names, got %d", len(names))
	}
	if names[0] != "id" || names[1] != "name" {
		t.Errorf("expected [id, name], got %v", names)
	}
}

func TestDataType_String(t *testing.T) {
	tests := []struct {
		dt       DataType
		expected string
	}{
		{TypeText, "TEXT"},
		{TypeInteger, "INTEGER"},
		{TypeReal, "REAL"},
		{TypeBoolean, "INTEGER"},
		{TypeJSON, "TEXT"},
		{TypeNull, "TEXT"},
	}

	for _, tt := range tests {
		if got := tt.dt.String(); got != tt.expected {
			t.Errorf("DataType(%d).String() = %q, want %q", tt.dt, got, tt.expected)
		}
	}
}

// Helper function to create a temporary JSON file
func createTempJSON(t *testing.T, content string) string {
	t.Helper()

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.json")

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	return path
}
