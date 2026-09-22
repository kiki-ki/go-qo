package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/kiki-ki/go-qo/internal/db"
	"github.com/kiki-ki/go-qo/internal/input"
	"github.com/kiki-ki/go-qo/internal/testutil"
)

// runCmd executes a fresh root command and returns everything it wrote.
// Each call gets its own flag state, so tests are order independent.
func runCmd(t *testing.T, args ...string) (string, error) {
	t.Helper()

	var out bytes.Buffer
	cmd := newRootCmd()
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs(args)

	err := cmd.Execute()
	return out.String(), err
}

// writeFile creates a file in a temp dir and returns its path.
func writeFile(t *testing.T, name, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write %s: %v", name, err)
	}
	return path
}

func TestValidateFormats(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		opts    options
		wantErr string
	}{
		{"defaults", options{inputFormat: "json", outputFormat: "json"}, ""},
		{"all input formats", options{inputFormat: "psv", outputFormat: "table"}, ""},
		{"unknown input", options{inputFormat: "xml", outputFormat: "json"}, "unsupported input format: xml"},
		{"unknown output", options{inputFormat: "json", outputFormat: "xml"}, "unsupported output format: xml"},
		{"input is case sensitive", options{inputFormat: "JSON", outputFormat: "json"}, "unsupported input format: JSON"},
		{"table is not an input format", options{inputFormat: "table", outputFormat: "json"}, "unsupported input format: table"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := validateFormats(&tt.opts)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("error = %q, want it to contain %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestLoadData_NoInput(t *testing.T) {
	t.Parallel()

	database, err := db.New()
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	testutil.CloseDB(t, database)

	loader := input.NewLoader(database, input.FormatJSON, nil)
	cfg := &runConfig{}

	err = loadData(loader, cfg, false)
	if err == nil {
		t.Fatal("expected error when there is no input, got nil")
	}
	if !strings.Contains(err.Error(), "no input data") {
		t.Errorf("error = %q, want it to mention no input data", err.Error())
	}
}

func TestLoadData_TableNames(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	paths := make([]string, 0, 3)
	// Names that need sanitizing, plus a collision, to pin what the TUI shows.
	for _, f := range []struct{ name, content string }{
		{"users.json", `[{"id": 1}]`},
		{"1-report.json", `[{"id": 2}]`},
		{"users.csv", "id\n3\n"},
	} {
		path := filepath.Join(dir, f.name)
		if err := os.WriteFile(path, []byte(f.content), 0o644); err != nil {
			t.Fatalf("failed to write %s: %v", f.name, err)
		}
		paths = append(paths, path)
	}

	database, err := db.New()
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	testutil.CloseDB(t, database)

	loader := input.NewLoader(database, input.FormatJSON, &input.LoaderOptions{DetectFormat: true})
	cfg := &runConfig{filePaths: paths}

	if err := loadData(loader, cfg, false); err != nil {
		t.Fatalf("loadData failed: %v", err)
	}

	want := []string{"users", "_1_report", "users_2"}
	if !slices.Equal(cfg.tableNames, want) {
		t.Fatalf("tableNames = %v, want %v", cfg.tableNames, want)
	}

	// The reported names must be the tables that actually exist.
	for _, name := range cfg.tableNames {
		var n int
		if err := database.QueryRow("SELECT COUNT(*) FROM " + name).Scan(&n); err != nil {
			t.Errorf("table %s is not queryable: %v", name, err)
		}
	}
}

func TestRootCmd_Query(t *testing.T) {
	t.Parallel()

	path := writeFile(t, "users.json", `[{"id": 2, "name": "bob"}, {"id": 1, "name": "alice"}]`)

	tests := []struct {
		name   string
		format string
		want   string
	}{
		{"json", "json", "[\n  {\n    \"id\": 1,\n    \"name\": \"alice\"\n  }\n]\n"},
		{"jsonl", "jsonl", "{\"id\":1,\"name\":\"alice\"}\n"},
		{"csv", "csv", "id,name\n1,alice\n"},
		{"tsv", "tsv", "id\tname\n1\talice\n"},
		{"psv", "psv", "id|name\n1|alice\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			out, err := runCmd(t, "-q", "SELECT * FROM users WHERE id = 1", "-o", tt.format, path)
			if err != nil {
				t.Fatalf("unexpected error: %v (output: %q)", err, out)
			}
			if out != tt.want {
				t.Errorf("output = %q, want %q", out, tt.want)
			}
		})
	}
}

func TestRootCmd_QueryTable(t *testing.T) {
	t.Parallel()

	path := writeFile(t, "users.json", `[{"id": 1, "name": "alice"}]`)

	out, err := runCmd(t, "-q", "SELECT * FROM users", "-o", "table", path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, want := range []string{"id", "name", "alice"} {
		if !strings.Contains(out, want) {
			t.Errorf("table output is missing %q:\n%s", want, out)
		}
	}
}

func TestRootCmd_Errors(t *testing.T) {
	t.Parallel()

	jsonPath := writeFile(t, "users.json", `[{"id": 1}]`)

	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{
			name:    "unsupported input format",
			args:    []string{"-i", "xml", "-q", "SELECT 1", jsonPath},
			wantErr: "unsupported input format: xml",
		},
		{
			name:    "unsupported output format",
			args:    []string{"-o", "xml", "-q", "SELECT 1", jsonPath},
			wantErr: "unsupported output format: xml",
		},
		{
			name:    "missing file",
			args:    []string{"-q", "SELECT 1", filepath.Join(t.TempDir(), "absent.json")},
			wantErr: "failed to read file",
		},
		{
			name:    "invalid SQL",
			args:    []string{"-q", "SELECT nope FROM users", jsonPath},
			wantErr: "no such column",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := runCmd(t, tt.args...)
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("error = %q, want it to contain %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestRootCmd_InputFormat(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	csvPath := filepath.Join(dir, "users.csv")
	if err := os.WriteFile(csvPath, []byte("id,name\n1,alice\n"), 0o644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	// Same CSV content behind a .json name, to tell inference from an override.
	mislabeledPath := filepath.Join(dir, "mislabeled.json")
	if err := os.WriteFile(mislabeledPath, []byte("id,name\n1,alice\n"), 0o644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	t.Run("inferred from extension when -i is absent", func(t *testing.T) {
		t.Parallel()

		out, err := runCmd(t, "-q", "SELECT name FROM users", "-o", "jsonl", csvPath)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if out != "{\"name\":\"alice\"}\n" {
			t.Errorf("output = %q", out)
		}
	})

	t.Run("explicit -i overrides the extension", func(t *testing.T) {
		t.Parallel()

		out, err := runCmd(t, "-i", "csv", "-q", "SELECT name FROM mislabeled", "-o", "jsonl", mislabeledPath)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if out != "{\"name\":\"alice\"}\n" {
			t.Errorf("output = %q", out)
		}
	})

	t.Run("explicit -i json still applies to a csv file", func(t *testing.T) {
		t.Parallel()

		if _, err := runCmd(t, "-i", "json", "-q", "SELECT 1", csvPath); err == nil {
			t.Error("expected error when forcing json on a csv file, got nil")
		}
	})
}
