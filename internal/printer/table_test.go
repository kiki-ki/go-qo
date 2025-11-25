package printer

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"strings"
	"testing"

	_ "modernc.org/sqlite"
)

func TestPrinter_PrintRows(t *testing.T) {
	tests := []struct {
		name   string
		format Format
		check  func(t *testing.T, output string)
	}{
		{
			name:   "table format",
			format: FormatTable,
			check: func(t *testing.T, output string) {
				if !strings.Contains(output, "Alice") || !strings.Contains(output, "Bob") {
					t.Errorf("table output missing expected data: %s", output)
				}
			},
		},
		{
			name:   "json format",
			format: FormatJSON,
			check: func(t *testing.T, output string) {
				var result []map[string]any
				if err := json.Unmarshal([]byte(output), &result); err != nil {
					t.Fatalf("invalid JSON: %v", err)
				}
				if len(result) != 2 || result[0]["name"] != "Alice" {
					t.Errorf("unexpected JSON output: %s", output)
				}
			},
		},
		{
			name:   "csv format",
			format: FormatCSV,
			check: func(t *testing.T, output string) {
				lines := strings.Split(strings.TrimSpace(output), "\n")
				if len(lines) != 3 || lines[0] != "id,name" {
					t.Errorf("unexpected CSV output: %s", output)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupTestDB(t)
			defer db.Close()

			rows, err := db.Query("SELECT * FROM test ORDER BY id")
			if err != nil {
				t.Fatalf("query failed: %v", err)
			}
			defer rows.Close()

			var buf bytes.Buffer
			p := New(&Options{Format: tt.format, Output: &buf})

			if err := p.PrintRows(rows); err != nil {
				t.Fatalf("PrintRows failed: %v", err)
			}

			tt.check(t, buf.String())
		})
	}
}

func TestPrinter_PrintRows_NullValues(t *testing.T) {
	db, _ := sql.Open("sqlite", ":memory:")
	defer db.Close()
	db.Exec(`CREATE TABLE test (id INTEGER, value TEXT); INSERT INTO test VALUES (1, NULL);`)

	rows, _ := db.Query("SELECT * FROM test")
	defer rows.Close()

	var buf bytes.Buffer
	p := New(&Options{Format: FormatTable, Output: &buf})
	p.PrintRows(rows)

	if !strings.Contains(buf.String(), "NULL") {
		t.Error("NULL values should be displayed as 'NULL'")
	}
}

func TestPrinter_PrintRows_EmptyResult(t *testing.T) {
	db, _ := sql.Open("sqlite", ":memory:")
	defer db.Close()
	db.Exec("CREATE TABLE test (id INTEGER)")

	rows, _ := db.Query("SELECT * FROM test")
	defer rows.Close()

	var buf bytes.Buffer
	p := New(&Options{Format: FormatJSON, Output: &buf})
	p.PrintRows(rows)

	var result []map[string]any
	json.Unmarshal(buf.Bytes(), &result)
	if len(result) != 0 {
		t.Errorf("expected empty array, got %d items", len(result))
	}
}

func TestNew_DefaultOptions(t *testing.T) {
	p := New(nil)
	if p.opts.Format != FormatTable {
		t.Errorf("default format should be table, got %s", p.opts.Format)
	}
}

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	db.Exec(`CREATE TABLE test (id INTEGER, name TEXT); INSERT INTO test VALUES (1, 'Alice'), (2, 'Bob');`)
	return db
}
