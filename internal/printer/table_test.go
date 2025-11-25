package printer

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"strings"
	"testing"

	_ "modernc.org/sqlite"
)

func TestPrinter_PrintRows_Table(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	rows, err := db.Query("SELECT * FROM test ORDER BY id")
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}
	defer rows.Close()

	var buf bytes.Buffer
	p := New(&Options{
		Format: FormatTable,
		Output: &buf,
	})

	err = p.PrintRows(rows)
	if err != nil {
		t.Fatalf("PrintRows failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Alice") {
		t.Errorf("output should contain 'Alice': %s", output)
	}
	if !strings.Contains(output, "Bob") {
		t.Errorf("output should contain 'Bob': %s", output)
	}
}

func TestPrinter_PrintRows_JSON(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	rows, err := db.Query("SELECT * FROM test ORDER BY id")
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}
	defer rows.Close()

	var buf bytes.Buffer
	p := New(&Options{
		Format: FormatJSON,
		Output: &buf,
	})

	err = p.PrintRows(rows)
	if err != nil {
		t.Fatalf("PrintRows failed: %v", err)
	}

	// Parse JSON output
	var result []map[string]any
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON output: %v\nOutput: %s", err, buf.String())
	}

	if len(result) != 2 {
		t.Errorf("expected 2 records, got %d", len(result))
	}

	if result[0]["name"] != "Alice" {
		t.Errorf("expected first name to be Alice, got %v", result[0]["name"])
	}
}

func TestPrinter_PrintRows_CSV(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	rows, err := db.Query("SELECT * FROM test ORDER BY id")
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}
	defer rows.Close()

	var buf bytes.Buffer
	p := New(&Options{
		Format: FormatCSV,
		Output: &buf,
	})

	err = p.PrintRows(rows)
	if err != nil {
		t.Fatalf("PrintRows failed: %v", err)
	}

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	// Should have header + 2 data rows
	if len(lines) != 3 {
		t.Errorf("expected 3 lines, got %d: %s", len(lines), output)
	}

	// Check header
	if lines[0] != "id,name" {
		t.Errorf("unexpected header: %s", lines[0])
	}
}

func TestPrinter_PrintRows_NullValues(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(`
		CREATE TABLE test (id INTEGER, value TEXT);
		INSERT INTO test VALUES (1, NULL);
	`)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	rows, err := db.Query("SELECT * FROM test")
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}
	defer rows.Close()

	var buf bytes.Buffer
	p := New(&Options{
		Format: FormatTable,
		Output: &buf,
	})

	err = p.PrintRows(rows)
	if err != nil {
		t.Fatalf("PrintRows failed: %v", err)
	}

	if !strings.Contains(buf.String(), "NULL") {
		t.Errorf("NULL values should be displayed as 'NULL'")
	}
}

func TestPrinter_PrintRows_EmptyResult(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	_, err = db.Exec("CREATE TABLE test (id INTEGER)")
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	rows, err := db.Query("SELECT * FROM test")
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}
	defer rows.Close()

	var buf bytes.Buffer
	p := New(&Options{
		Format: FormatJSON,
		Output: &buf,
	})

	err = p.PrintRows(rows)
	if err != nil {
		t.Fatalf("PrintRows failed: %v", err)
	}

	// Should output empty array
	var result []map[string]any
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty array, got %d items", len(result))
	}
}

func TestNew_DefaultOptions(t *testing.T) {
	p := New(nil)
	if p.opts == nil {
		t.Error("opts should not be nil")
	}
	if p.opts.Format != FormatTable {
		t.Errorf("default format should be table, got %s", p.opts.Format)
	}
}

func TestDefaultOptions(t *testing.T) {
	opts := DefaultOptions()
	if opts.Format != FormatTable {
		t.Errorf("expected table format, got %s", opts.Format)
	}
	if opts.Output == nil {
		t.Error("output should not be nil")
	}
}

func TestPrint_Deprecated(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	rows, err := db.Query("SELECT * FROM test LIMIT 1")
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}
	defer rows.Close()

	// Just verify it doesn't panic
	// Note: This will print to stdout in tests
	err = Print(rows)
	if err != nil {
		t.Errorf("Print failed: %v", err)
	}
}

// Helper function to set up a test database
func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}

	_, err = db.Exec(`
		CREATE TABLE test (id INTEGER, name TEXT);
		INSERT INTO test VALUES (1, 'Alice');
		INSERT INTO test VALUES (2, 'Bob');
	`)
	if err != nil {
		db.Close()
		t.Fatalf("setup failed: %v", err)
	}

	return db
}
