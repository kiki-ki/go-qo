package output_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"strings"
	"testing"

	"github.com/kiki-ki/go-qo/internal/output"
	"github.com/kiki-ki/go-qo/testutil"
	_ "modernc.org/sqlite"
)

func TestPrinter_PrintRows(t *testing.T) {
	tests := []struct {
		name   string
		format output.Format
		check  func(t *testing.T, out string)
	}{
		{
			name:   "table format",
			format: output.FormatTable,
			check: func(t *testing.T, out string) {
				if !strings.Contains(out, "Alice") || !strings.Contains(out, "Bob") {
					t.Errorf("table output missing expected data: %s", out)
				}
			},
		},
		{
			name:   "json format",
			format: output.FormatJSON,
			check: func(t *testing.T, out string) {
				var result []map[string]any
				if err := json.Unmarshal([]byte(out), &result); err != nil {
					t.Fatalf("invalid JSON: %v", err)
				}
				if len(result) != 2 || result[0]["name"] != "Alice" {
					t.Errorf("unexpected JSON output: %s", out)
				}
			},
		},
		{
			name:   "csv format",
			format: output.FormatCSV,
			check: func(t *testing.T, out string) {
				lines := strings.Split(strings.TrimSpace(out), "\n")
				if len(lines) != 3 || lines[0] != "id,name" {
					t.Errorf("unexpected CSV output: %s", out)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := testutil.SetupTestDB(t)
			defer db.Close()

			rows, err := db.Query("SELECT * FROM test ORDER BY id")
			if err != nil {
				t.Fatalf("query failed: %v", err)
			}
			defer rows.Close()

			var buf bytes.Buffer
			p := output.NewPrinter(&output.Options{Format: tt.format, Output: &buf})

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
	p := output.NewPrinter(&output.Options{Format: output.FormatTable, Output: &buf})
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
	p := output.NewPrinter(&output.Options{Format: output.FormatJSON, Output: &buf})
	p.PrintRows(rows)

	var result []map[string]any
	json.Unmarshal(buf.Bytes(), &result)
	if len(result) != 0 {
		t.Errorf("expected empty array, got %d items", len(result))
	}
}

func TestNewPrinter_DefaultOptions(t *testing.T) {
	p := output.NewPrinter(nil)
	if p == nil {
		t.Error("expected non-nil printer")
	}
}
