package output_test

import (
	"bytes"
	"strings"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/kiki-ki/go-qo/internal/output"
	"github.com/kiki-ki/go-qo/internal/testutil"
)

func TestPrinter_PrintRows(t *testing.T) {
	tests := []struct {
		name   string
		setup  string
		query  string
		format output.Format
		want   string
	}{
		{
			name: "json format",
			setup: `
				CREATE TABLE test (id INTEGER, name TEXT);
				INSERT INTO test VALUES (1, 'Alice'), (2, 'Bob');
			`,
			query:  "SELECT * FROM test ORDER BY id",
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
			name: "jsonl format",
			setup: `
				CREATE TABLE test (id INTEGER, name TEXT);
				INSERT INTO test VALUES (1, 'Alice'), (2, 'Bob');
			`,
			query:  "SELECT * FROM test ORDER BY id",
			format: output.FormatJSONL,
			want:   "{\"id\":1,\"name\":\"Alice\"}\n{\"id\":2,\"name\":\"Bob\"}\n",
		},
		{
			name: "csv format",
			setup: `
				CREATE TABLE test (id INTEGER, name TEXT);
				INSERT INTO test VALUES (1, 'Alice'), (2, 'Bob');
			`,
			query:  "SELECT * FROM test ORDER BY id",
			format: output.FormatCSV,
			want:   "id,name\n1,Alice\n2,Bob\n",
		},
		{
			name: "tsv format",
			setup: `
				CREATE TABLE test (id INTEGER, name TEXT);
				INSERT INTO test VALUES (1, 'Alice'), (2, 'Bob');
			`,
			query:  "SELECT * FROM test ORDER BY id",
			format: output.FormatTSV,
			want:   "id\tname\n1\tAlice\n2\tBob\n",
		},
		{
			name: "json format with null value",
			setup: `
				CREATE TABLE test (id INTEGER, value TEXT);
				INSERT INTO test VALUES (1, NULL);
			`,
			query:  "SELECT * FROM test",
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
			name: "jsonl format with null value",
			setup: `
				CREATE TABLE test (id INTEGER, value TEXT);
				INSERT INTO test VALUES (1, NULL);
			`,
			query:  "SELECT * FROM test",
			format: output.FormatJSONL,
			want:   "{\"id\":1,\"value\":null}\n",
		},
		{
			name: "csv format with null value",
			setup: `
				CREATE TABLE test (id INTEGER, value TEXT);
				INSERT INTO test VALUES (1, NULL);
			`,
			query:  "SELECT * FROM test",
			format: output.FormatCSV,
			want:   "id,value\n1,\n",
		},
		{
			name: "tsv format with null value",
			setup: `
				CREATE TABLE test (id INTEGER, value TEXT);
				INSERT INTO test VALUES (1, NULL);
			`,
			query:  "SELECT * FROM test",
			format: output.FormatTSV,
			want:   "id\tvalue\n1\t\n",
		},
		{
			name:   "json format with empty result",
			setup:  `CREATE TABLE test (id INTEGER);`,
			query:  "SELECT * FROM test",
			format: output.FormatJSON,
			want:   "[]\n",
		},
		{
			name: "json format with nested object",
			setup: `
				CREATE TABLE test (id INTEGER, meta TEXT);
				INSERT INTO test VALUES (1, '{"name":"Alice","data":[1,2,3]}');
			`,
			query:  "SELECT * FROM test",
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
			name: "csv format with nested object",
			setup: `
				CREATE TABLE test (id INTEGER, meta TEXT);
				INSERT INTO test VALUES (1, '{"name":"Alice","data":[1,2,3]}');
			`,
			query:  "SELECT * FROM test",
			format: output.FormatCSV,
			want:   "id,meta\n1,\"{\"\"data\"\":[1,2,3],\"\"name\"\":\"\"Alice\"\"}\"\n",
		},
		{
			name: "tsv format with nested object",
			setup: `
				CREATE TABLE test (id INTEGER, meta TEXT);
				INSERT INTO test VALUES (1, '{"name":"Alice","data":[1,2,3]}');
			`,
			query:  "SELECT * FROM test",
			format: output.FormatTSV,
			want:   "id\tmeta\n1\t\"{\"\"data\"\":[1,2,3],\"\"name\"\":\"\"Alice\"\"}\"\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := testutil.SetupTestDB(t)
			_, err := db.Exec(tt.setup)
			if err != nil {
				t.Fatalf("failed to setup test data: %v", err)
			}

			rows, err := db.Query(tt.query)
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

func TestPrinter_PrintRows_TableFormat(t *testing.T) {
	tests := []struct {
		name     string
		setup    string
		query    string
		contains []string
	}{
		{
			name: "normal data",
			setup: `
				CREATE TABLE test (id INTEGER, name TEXT);
				INSERT INTO test VALUES (1, 'Alice'), (2, 'Bob');
			`,
			query:    "SELECT * FROM test ORDER BY id",
			contains: []string{"id", "name", "1", "2", "Alice", "Bob"},
		},
		{
			name: "null value",
			setup: `
				CREATE TABLE test (id INTEGER, value TEXT);
				INSERT INTO test VALUES (1, NULL);
			`,
			query:    "SELECT * FROM test",
			contains: []string{"id", "value", "1", "(NULL)"},
		},
		{
			name: "nested json",
			setup: `
				CREATE TABLE test (id INTEGER, data TEXT);
				INSERT INTO test VALUES (1, '{"key": "value", "roles": ["a", "b"]}');
			`,
			query:    "SELECT * FROM test",
			contains: []string{"id", "data", "1", `{"key":"value","roles":["a","b"]}`},
		},
		{
			name: "json array",
			setup: `
				CREATE TABLE test (id INTEGER, data TEXT);
				INSERT INTO test VALUES (1, '[1, 2, 3]');
			`,
			query:    "SELECT * FROM test",
			contains: []string{"id", "data", "1", `[1,2,3]`},
		},
		{
			name: "empty result",
			setup: `
				CREATE TABLE test (id INTEGER, name TEXT);
			`,
			query:    "SELECT * FROM test",
			contains: []string{"id", "name"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := testutil.SetupTestDB(t)
			_, err := db.Exec(tt.setup)
			if err != nil {
				t.Fatalf("failed to setup test data: %v", err)
			}

			rows, err := db.Query(tt.query)
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
			for _, s := range tt.contains {
				if !strings.Contains(out, s) {
					t.Errorf("table output missing %q: %s", s, out)
				}
			}
		})
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
