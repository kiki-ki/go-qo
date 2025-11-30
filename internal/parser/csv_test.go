package parser_test

import (
	"testing"

	"github.com/kiki-ki/go-qo/internal/parser"
)

func TestCSVParser_ParseBytes(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantColumns []string
		wantRows    int
		wantErr     bool
	}{
		{
			name:        "simple csv",
			input:       "id,name,age\n1,Alice,30\n2,Bob,25\n",
			wantColumns: []string{"id", "name", "age"},
			wantRows:    2,
			wantErr:     false,
		},
		{
			name:        "with float values",
			input:       "id,score\n1,95.5\n2,87.3\n",
			wantColumns: []string{"id", "score"},
			wantRows:    2,
			wantErr:     false,
		},
		{
			name:        "empty data (header only)",
			input:       "id,name\n",
			wantColumns: []string{"id", "name"},
			wantRows:    0,
			wantErr:     false,
		},
		{
			name:    "completely empty",
			input:   "",
			wantErr: true,
		},
		{
			name:        "with quoted values",
			input:       "id,name,desc\n1,Alice,\"Hello, World\"\n",
			wantColumns: []string{"id", "name", "desc"},
			wantRows:    1,
			wantErr:     false,
		},
	}

	p := &parser.CSVParser{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := p.ParseBytes([]byte(tt.input))

			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(result.Columns) != len(tt.wantColumns) {
				t.Errorf("columns: got %d, want %d", len(result.Columns), len(tt.wantColumns))
			}

			for i, col := range result.Columns {
				if col.Name != tt.wantColumns[i] {
					t.Errorf("column %d: got %s, want %s", i, col.Name, tt.wantColumns[i])
				}
			}

			if len(result.Rows) != tt.wantRows {
				t.Errorf("rows: got %d, want %d", len(result.Rows), tt.wantRows)
			}
		})
	}
}

func TestCSVParser_TypeInference(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		colIndex int
		wantType parser.DataType
	}{
		{
			name:     "integer column",
			input:    "id,name\n1,Alice\n2,Bob\n",
			colIndex: 0,
			wantType: parser.TypeInteger,
		},
		{
			name:     "text column",
			input:    "id,name\n1,Alice\n2,Bob\n",
			colIndex: 1,
			wantType: parser.TypeText,
		},
		{
			name:     "float column",
			input:    "id,score\n1,95.5\n2,87.3\n",
			colIndex: 1,
			wantType: parser.TypeReal,
		},
		{
			name:     "mixed int and float becomes float",
			input:    "id,score\n1,95\n2,87.3\n",
			colIndex: 1,
			wantType: parser.TypeReal,
		},
		{
			name:     "mixed with text becomes text",
			input:    "id,value\n1,100\n2,hello\n",
			colIndex: 1,
			wantType: parser.TypeText,
		},
	}

	p := &parser.CSVParser{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := p.ParseBytes([]byte(tt.input))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.Columns[tt.colIndex].Type != tt.wantType {
				t.Errorf("column type: got %v, want %v", result.Columns[tt.colIndex].Type, tt.wantType)
			}
		})
	}
}

func TestCSVParser_ParseFile(t *testing.T) {
	p := &parser.CSVParser{}

	result, err := p.Parse("../../testdata/csv/simple.csv")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Columns) != 3 {
		t.Errorf("columns: got %d, want 3", len(result.Columns))
	}

	if len(result.Rows) != 3 {
		t.Errorf("rows: got %d, want 3", len(result.Rows))
	}
}

func TestCSVParser_NoHeader(t *testing.T) {
	p := &parser.CSVParser{Options: parser.CSVOptions{NoHeader: true}}

	result, err := p.ParseBytes([]byte("1,Alice,30\n2,Bob,25\n"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have auto-generated column names
	expectedCols := []string{"col1", "col2", "col3"}
	for i, col := range result.Columns {
		if col.Name != expectedCols[i] {
			t.Errorf("column %d: got %s, want %s", i, col.Name, expectedCols[i])
		}
	}

	// All rows should be data rows
	if len(result.Rows) != 2 {
		t.Errorf("rows: got %d, want 2", len(result.Rows))
	}
}

func TestCSVParser_TSV(t *testing.T) {
	p := &parser.CSVParser{Options: parser.CSVOptions{Delimiter: '\t'}}

	result, err := p.ParseBytes([]byte("id\tname\tage\n1\tAlice\t30\n2\tBob\t25\n"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedCols := []string{"id", "name", "age"}
	for i, col := range result.Columns {
		if col.Name != expectedCols[i] {
			t.Errorf("column %d: got %s, want %s", i, col.Name, expectedCols[i])
		}
	}

	if len(result.Rows) != 2 {
		t.Errorf("rows: got %d, want 2", len(result.Rows))
	}

	// Verify data
	if result.Rows[0][1] != "Alice" {
		t.Errorf("first row name: got %v, want Alice", result.Rows[0][1])
	}
}

func TestCSVParser_TSV_NoHeader(t *testing.T) {
	p := &parser.CSVParser{Options: parser.CSVOptions{
		NoHeader:  true,
		Delimiter: '\t',
	}}

	result, err := p.ParseBytes([]byte("1\tAlice\t30\n2\tBob\t25\n"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have auto-generated column names
	expectedCols := []string{"col1", "col2", "col3"}
	for i, col := range result.Columns {
		if col.Name != expectedCols[i] {
			t.Errorf("column %d: got %s, want %s", i, col.Name, expectedCols[i])
		}
	}

	if len(result.Rows) != 2 {
		t.Errorf("rows: got %d, want 2", len(result.Rows))
	}
}
