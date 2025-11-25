package db

import (
	"testing"

	"github.com/kiki-ki/go-qo/internal/parser"
)

func TestNew(t *testing.T) {
	db, err := New()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer db.Close()

	// Test that database is functional
	var result int
	err = db.QueryRow("SELECT 1").Scan(&result)
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if result != 1 {
		t.Errorf("expected 1, got %d", result)
	}
}

func TestDB_LoadData(t *testing.T) {
	db, err := New()
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	defer db.Close()

	data := &parser.ParsedData{
		Columns: []parser.Column{
			{Name: "id", Type: parser.TypeInteger},
			{Name: "name", Type: parser.TypeText},
			{Name: "score", Type: parser.TypeReal},
		},
		Rows: [][]any{
			{int64(1), "Alice", 95.5},
			{int64(2), "Bob", 87.0},
		},
	}

	err = db.LoadData("users", data)
	if err != nil {
		t.Fatalf("LoadData failed: %v", err)
	}

	// Verify data was inserted
	rows, err := db.Query("SELECT * FROM users ORDER BY id")
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}
	defer rows.Close()

	var count int
	for rows.Next() {
		var id int
		var name string
		var score float64
		if err := rows.Scan(&id, &name, &score); err != nil {
			t.Fatalf("scan failed: %v", err)
		}
		count++
	}

	if count != 2 {
		t.Errorf("expected 2 rows, got %d", count)
	}
}

func TestDB_LoadData_EmptyRows(t *testing.T) {
	db, err := New()
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	defer db.Close()

	data := &parser.ParsedData{
		Columns: []parser.Column{
			{Name: "id", Type: parser.TypeInteger},
		},
		Rows: [][]any{},
	}

	err = db.LoadData("empty_table", data)
	if err != nil {
		t.Fatalf("LoadData with empty rows should not fail: %v", err)
	}

	// Verify table exists but is empty
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM empty_table").Scan(&count)
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 rows, got %d", count)
	}
}

func TestDB_LoadData_NullValues(t *testing.T) {
	db, err := New()
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	defer db.Close()

	data := &parser.ParsedData{
		Columns: []parser.Column{
			{Name: "id", Type: parser.TypeInteger},
			{Name: "value", Type: parser.TypeText},
		},
		Rows: [][]any{
			{int64(1), nil},
			{int64(2), "test"},
		},
	}

	err = db.LoadData("nullable", data)
	if err != nil {
		t.Fatalf("LoadData failed: %v", err)
	}

	var value *string
	err = db.QueryRow("SELECT value FROM nullable WHERE id = 1").Scan(&value)
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if value != nil {
		t.Errorf("expected nil, got %v", *value)
	}
}

func TestTableNameFromPath(t *testing.T) {
	tests := []struct {
		path     string
		expected string
	}{
		{"data.json", "data"},
		{"path/to/users.json", "users"},
		{"my-file.json", "my_file"},
		{"file with spaces.json", "file_with_spaces"},
		{"file.name.json", "file_name"},
		{"/absolute/path/data.json", "data"},
	}

	for _, tt := range tests {
		result := TableNameFromPath(tt.path)
		if result != tt.expected {
			t.Errorf("TableNameFromPath(%q) = %q, want %q", tt.path, result, tt.expected)
		}
	}
}

func TestSanitizeTableName_Deprecated(t *testing.T) {
	// Test that deprecated function still works
	result := SanitizeTableName("test-file.json")
	if result != "test_file" {
		t.Errorf("SanitizeTableName failed: got %q", result)
	}
}
