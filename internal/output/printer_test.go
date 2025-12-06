package output_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"strings"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/kiki-ki/go-qo/internal/output"
	"github.com/kiki-ki/go-qo/internal/testutil"
)

func TestPrinter_PrintRows(t *testing.T) {
	tests := []struct {
		name   string
		format output.Format
		want   string
	}{
		{
			name:   "json format",
			format: output.FormatJSON,
			want: `[
  {
    "id": 1,
    "name": "Alice"
  },
  {
    "id": 2,
    "name": "Bob"
  }
]
`,
		},
		{
			name:   "jsonl format",
			format: output.FormatJSONL,
			want:   "{\"id\":1,\"name\":\"Alice\"}\n{\"id\":2,\"name\":\"Bob\"}\n",
		},
		{
			name:   "csv format",
			format: output.FormatCSV,
			want:   "id,name\n1,Alice\n2,Bob\n",
		},
		{
			name:   "tsv format",
			format: output.FormatTSV,
			want:   "id\tname\n1\tAlice\n2\tBob\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := testutil.SetupTestDB(t)
			_, err := db.Exec(`
				CREATE TABLE test (id INTEGER, name TEXT);
				INSERT INTO test VALUES (1, 'Alice'), (2, 'Bob');
			`)
			if err != nil {
				t.Fatalf("failed to setup test data: %v", err)
			}

			rows, err := db.Query("SELECT * FROM test ORDER BY id")
			if err != nil {
				t.Fatalf("query failed: %v", err)
			}
			testutil.CloseRows(t, rows)

			var buf bytes.Buffer
			p := output.NewPrinter(&output.Options{Format: tt.format, Output: &buf})

			if err := p.PrintRows(rows); err != nil {
				t.Fatalf("PrintRows failed: %v", err)
			}

			if got := buf.String(); got != tt.want {
				t.Errorf("output mismatch:\ngot:  %q\nwant: %q", got, tt.want)
			}
		})
	}
}

// table format is tested separately due to its unique output style.
func TestPrinter_PrintRows_TableFormat(t *testing.T) {
	db := testutil.SetupTestDB(t)
	_, err := db.Exec(`
		CREATE TABLE test (id INTEGER, name TEXT);
		INSERT INTO test VALUES (1, 'Alice'), (2, 'Bob');
	`)
	if err != nil {
		t.Fatalf("failed to setup test data: %v", err)
	}

	rows, err := db.Query("SELECT * FROM test ORDER BY id")
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}
	testutil.CloseRows(t, rows)

	var buf bytes.Buffer
	p := output.NewPrinter(&output.Options{Format: output.FormatTable, Output: &buf})

	if err := p.PrintRows(rows); err != nil {
		t.Fatalf("PrintRows failed: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "Alice") || !strings.Contains(out, "Bob") {
		t.Errorf("table output missing expected data: %s", out)
	}
	if !strings.Contains(out, "id") || !strings.Contains(out, "name") {
		t.Errorf("table output missing headers: %s", out)
	}
}

func TestPrinter_PrintRows_NullValues(t *testing.T) {
	tests := []struct {
		name   string
		format output.Format
		want   string
	}{
		{
			name:   "json format outputs null",
			format: output.FormatJSON,
			want: `[
  {
    "id": 1,
    "value": null
  }
]
`,
		},
		{
			name:   "jsonl format outputs null",
			format: output.FormatJSONL,
			want:   "{\"id\":1,\"value\":null}\n",
		},
		{
			name:   "csv format outputs empty string",
			format: output.FormatCSV,
			want:   "id,value\n1,\n",
		},
		{
			name:   "tsv format outputs empty string",
			format: output.FormatTSV,
			want:   "id\tvalue\n1\t\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, err := sql.Open("sqlite", ":memory:")
			if err != nil {
				t.Fatal(err)
			}
			testutil.CloseDB(t, db)

			_, err = db.Exec(`
				CREATE TABLE test (id INTEGER, value TEXT);
				INSERT INTO test VALUES (1, NULL);
			`)
			if err != nil {
				t.Fatal(err)
			}

			rows, err := db.Query("SELECT * FROM test")
			if err != nil {
				t.Fatal(err)
			}
			testutil.CloseRows(t, rows)

			var buf bytes.Buffer
			p := output.NewPrinter(&output.Options{Format: tt.format, Output: &buf})
			if err := p.PrintRows(rows); err != nil {
				t.Fatal(err)
			}

			if got := buf.String(); got != tt.want {
				t.Errorf("output mismatch:\ngot:  %q\nwant: %q", got, tt.want)
			}
		})
	}
}

// table format is tested separately due to lipgloss dependency.
func TestPrinter_PrintRows_NullValues_TableFormat(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	testutil.CloseDB(t, db)

	_, err = db.Exec(`
		CREATE TABLE test (id INTEGER, value TEXT);
		INSERT INTO test VALUES (1, NULL);
	`)
	if err != nil {
		t.Fatal(err)
	}

	rows, err := db.Query("SELECT * FROM test")
	if err != nil {
		t.Fatal(err)
	}
	testutil.CloseRows(t, rows)

	var buf bytes.Buffer
	p := output.NewPrinter(&output.Options{Format: output.FormatTable, Output: &buf})
	if err := p.PrintRows(rows); err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if !strings.Contains(out, "(NULL)") {
		t.Errorf("table output should contain '(NULL)' for null values: %s", out)
	}
}

func TestPrinter_PrintRows_EmptyResult(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	testutil.CloseDB(t, db)

	_, err = db.Exec("CREATE TABLE test (id INTEGER)")
	if err != nil {
		t.Fatal(err)
	}

	rows, err := db.Query("SELECT * FROM test")
	if err != nil {
		t.Fatal(err)
	}
	testutil.CloseRows(t, rows)

	var buf bytes.Buffer
	p := output.NewPrinter(&output.Options{Format: output.FormatJSON, Output: &buf})
	if err := p.PrintRows(rows); err != nil {
		t.Fatal(err)
	}

	var result []map[string]any
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
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

func TestPrinter_PrintRows_NoANSICodesWhenNotTTY(t *testing.T) {
	db := testutil.SetupTestDB(t)
	_, err := db.Exec(`
		CREATE TABLE test (id INTEGER, name TEXT);
		INSERT INTO test VALUES (1, 'Alice'), (2, 'Bob');
	`)
	if err != nil {
		t.Fatalf("failed to setup test data: %v", err)
	}

	rows, err := db.Query("SELECT * FROM test ORDER BY id")
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}
	testutil.CloseRows(t, rows)

	var buf bytes.Buffer
	p := output.NewPrinter(&output.Options{Format: output.FormatTable, Output: &buf})

	if err := p.PrintRows(rows); err != nil {
		t.Fatalf("PrintRows failed: %v", err)
	}

	out := buf.String()
	// ANSI escape codes start with ESC (0x1b or \033)
	if strings.Contains(out, "\x1b[") || strings.Contains(out, "\033[") {
		t.Error("table output should not contain ANSI escape codes when output is not a TTY")
	}
}

func TestPrinter_PrintRows_NestedJSON(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	testutil.CloseDB(t, db)

	_, err = db.Exec(`
		CREATE TABLE test (id INTEGER, meta TEXT);
		INSERT INTO test VALUES (1, '{"name":"Alice","data":[1,2,3]}');
	`)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name   string
		format output.Format
		want   string
	}{
		{
			name:   "json format preserves nested object",
			format: output.FormatJSON,
			want: `[
  {
    "id": 1,
    "meta": {
      "data": [
        1,
        2,
        3
      ],
      "name": "Alice"
    }
  }
]
`,
		},
		{
			name:   "csv format outputs JSON string",
			format: output.FormatCSV,
			want:   "id,meta\n1,\"{\"\"data\"\":[1,2,3],\"\"name\"\":\"\"Alice\"\"}\"\n",
		},
		{
			name:   "tsv format outputs JSON string",
			format: output.FormatTSV,
			want:   "id\tmeta\n1\t\"{\"\"data\"\":[1,2,3],\"\"name\"\":\"\"Alice\"\"}\"\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rows, err := db.Query("SELECT * FROM test")
			if err != nil {
				t.Fatalf("query failed: %v", err)
			}
			testutil.CloseRows(t, rows)

			var buf bytes.Buffer
			p := output.NewPrinter(&output.Options{Format: tt.format, Output: &buf})

			if err := p.PrintRows(rows); err != nil {
				t.Fatalf("PrintRows failed: %v", err)
			}

			if got := buf.String(); got != tt.want {
				t.Errorf("output mismatch:\ngot:  %q\nwant: %q", got, tt.want)
			}
		})
	}
}
