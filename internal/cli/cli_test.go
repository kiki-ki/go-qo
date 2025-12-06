package cli_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/kiki-ki/go-qo/internal/cli"
	"github.com/kiki-ki/go-qo/internal/output"
	"github.com/kiki-ki/go-qo/internal/testutil"
)

func TestRun(t *testing.T) {
	db := testutil.SetupTestDB(t)
	_, err := db.Exec(`
		CREATE TABLE test (id INTEGER, name TEXT);
		INSERT INTO test VALUES (1, 'Alice'), (2, 'Bob');
	`)
	if err != nil {
		t.Fatalf("failed to setup test data: %v", err)
	}

	tests := []struct {
		name   string
		query  string
		format output.Format
		check  func(t *testing.T, out string)
	}{
		{
			name:   "select all with table format",
			query:  "SELECT * FROM test ORDER BY id",
			format: output.FormatTable,
			check: func(t *testing.T, out string) {
				if !strings.Contains(out, "Alice") || !strings.Contains(out, "Bob") {
					t.Errorf("expected Alice and Bob in output: %s", out)
				}
			},
		},
		{
			name:   "select all with json format",
			query:  "SELECT * FROM test ORDER BY id",
			format: output.FormatJSON,
			check: func(t *testing.T, out string) {
				var result []map[string]any
				if err := json.Unmarshal([]byte(out), &result); err != nil {
					t.Fatalf("invalid JSON: %v", err)
				}
				if len(result) != 2 {
					t.Errorf("expected 2 rows, got %d", len(result))
				}
			},
		},
		{
			name:   "select with where clause",
			query:  "SELECT name FROM test WHERE id = 1",
			format: output.FormatCSV,
			check: func(t *testing.T, out string) {
				lines := strings.Split(strings.TrimSpace(out), "\n")
				if len(lines) != 2 { // header + 1 row
					t.Errorf("expected 2 lines, got %d", len(lines))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			opts := &cli.Options{
				Format: tt.format,
				Output: &buf,
			}

			if err := cli.Run(db, tt.query, opts); err != nil {
				t.Fatalf("Run failed: %v", err)
			}

			tt.check(t, buf.String())
		})
	}
}

func TestRun_DefaultOptions(t *testing.T) {
	db := testutil.SetupTestDB(t)
	_, err := db.Exec(`
		CREATE TABLE test (id INTEGER, name TEXT);
		INSERT INTO test VALUES (1, 'Alice'), (2, 'Bob');
	`)
	if err != nil {
		t.Fatalf("failed to setup test data: %v", err)
	}

	// Run with nil options should not panic
	err = cli.Run(db, "SELECT * FROM test", nil)
	if err != nil {
		t.Fatalf("Run with nil options failed: %v", err)
	}
}

func TestRun_InvalidQuery(t *testing.T) {
	db := testutil.SetupTestDB(t)
	_, err := db.Exec(`
		CREATE TABLE test (id INTEGER, name TEXT);
		INSERT INTO test VALUES (1, 'Alice'), (2, 'Bob');
	`)
	if err != nil {
		t.Fatalf("failed to setup test data: %v", err)
	}

	var buf bytes.Buffer
	err = cli.Run(db, "INVALID SQL", &cli.Options{Output: &buf})
	if err == nil {
		t.Error("expected error for invalid query")
	}
}
