package ui_test

import (
	"testing"

	"github.com/kiki-ki/go-qo/internal/ui"
	"github.com/kiki-ki/go-qo/testutil"
)

func TestSQLRowsToTable(t *testing.T) {
	db := testutil.SetupTestDB(t)

	rows, err := db.Query("SELECT * FROM test ORDER BY id")
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}
	testutil.CloseRows(t, rows)

	cols, tableRows, err := ui.SQLRowsToTable(rows)
	if err != nil {
		t.Fatalf("SQLRowsToTable failed: %v", err)
	}

	// Verify columns
	if len(cols) != 2 {
		t.Errorf("expected 2 columns, got %d", len(cols))
	}
	if cols[0].Title != "id" {
		t.Errorf("expected first column 'id', got %s", cols[0].Title)
	}
	if cols[1].Title != "name" {
		t.Errorf("expected second column 'name', got %s", cols[1].Title)
	}

	// Verify rows
	if len(tableRows) != 2 {
		t.Errorf("expected 2 rows, got %d", len(tableRows))
	}
	if tableRows[0][1] != "Alice" {
		t.Errorf("expected 'Alice', got %s", tableRows[0][1])
	}
	if tableRows[1][1] != "Bob" {
		t.Errorf("expected 'Bob', got %s", tableRows[1][1])
	}
}

func TestSQLRowsToTable_EmptyResult(t *testing.T) {
	db := testutil.SetupTestDB(t)

	rows, err := db.Query("SELECT * FROM test WHERE id = 999")
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}
	testutil.CloseRows(t, rows)

	cols, tableRows, err := ui.SQLRowsToTable(rows)
	if err != nil {
		t.Fatalf("SQLRowsToTable failed: %v", err)
	}

	if len(cols) != 2 {
		t.Errorf("expected 2 columns, got %d", len(cols))
	}
	if len(tableRows) != 0 {
		t.Errorf("expected 0 rows, got %d", len(tableRows))
	}
}

func TestFormatValue(t *testing.T) {
	tests := []struct {
		name  string
		input any
		want  string
	}{
		{"nil", nil, "(NULL)"},
		{"string", "hello", "hello"},
		{"int", 42, "42"},
		{"int64", int64(100), "100"},
		{"float64 whole", float64(5), "5"},
		{"float64 decimal", 3.14, "3.14"},
		{"bytes", []byte("test"), "test"},
		{"bool true", true, "true"},
		{"bool false", false, "false"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ui.FormatValue(tt.input)
			if got != tt.want {
				t.Errorf("FormatValue(%v) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
