package cmd

import (
	"bytes"
	"fmt"
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
		{"unknown input", options{inputFormat: "xml", outputFormat: "json"}, "unsupported input format: xml"},
		{"unknown output", options{inputFormat: "json", outputFormat: "xml"}, "unsupported output format: xml"},
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

	loader := input.NewLoader(database, "", nil)
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

// The output formats themselves are covered in internal/output; what matters
// here is that -o reaches the printer and its bytes land on the command's
// writer, so one non-default format is enough.
func TestRootCmd_Query(t *testing.T) {
	t.Parallel()

	path := writeFile(t, "users.json", `[{"id": 2, "name": "bob"}, {"id": 1, "name": "alice"}]`)

	out, err := runCmd(t, "-q", "SELECT * FROM users WHERE id = 1", "-o", "csv", path)
	if err != nil {
		t.Fatalf("unexpected error: %v (output: %q)", err, out)
	}
	if out != "id,name\n1,alice\n" {
		t.Errorf("output = %q", out)
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
}

func runCmdWithStreams(t *testing.T, args ...string) (string, string, error) {
	t.Helper()

	var stdout, stderr bytes.Buffer
	cmd := newRootCmd()
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs(args)

	err := cmd.Execute()
	return stdout.String(), stderr.String(), err
}

func TestRootCmd_PrintCommand(t *testing.T) {
	t.Parallel()

	jsonPath := writeFile(t, "users.json", `[{"id": 1, "name": "alice"}]`)

	t.Run("disabled by default", func(t *testing.T) {
		t.Parallel()

		stdout, stderr, err := runCmdWithStreams(t, "-q", "SELECT name FROM users", jsonPath)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if stderr != "" {
			t.Errorf("stderr = %q, want empty", stderr)
		}
		if !strings.Contains(stdout, "alice") {
			t.Errorf("stdout = %q, want it to contain alice", stdout)
		}
	})

	t.Run("prints command to stderr without polluting stdout", func(t *testing.T) {
		t.Parallel()

		stdout, stderr, err := runCmdWithStreams(t, "--print-command", "-q", "SELECT name FROM users", jsonPath)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		wantCmd := fmt.Sprintf("qo -q \"SELECT name FROM users\" %s\n", jsonPath)
		if stderr != wantCmd {
			t.Errorf("stderr = %q, want %q", stderr, wantCmd)
		}
		if strings.Contains(stdout, "qo ") {
			t.Errorf("stdout was polluted with CLI command: %q", stdout)
		}
		if !strings.Contains(stdout, "alice") {
			t.Errorf("stdout = %q, want it to contain alice", stdout)
		}
	})

	t.Run("formats options correctly", func(t *testing.T) {
		t.Parallel()

		csvPath := writeFile(t, "data.csv", "name\nalice\n")
		stdout, stderr, err := runCmdWithStreams(t, "--print-command", "-i", "csv", "-o", "csv", "--no-header", "-q", "SELECT * FROM data", csvPath)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		wantCmd := fmt.Sprintf("qo -i csv -o csv --no-header -q \"SELECT * FROM data\" %s\n", csvPath)
		if stderr != wantCmd {
			t.Errorf("stderr = %q, want %q", stderr, wantCmd)
		}
		if !strings.Contains(stdout, "alice") {
			t.Errorf("stdout = %q, want it to contain alice", stdout)
		}
	})

	t.Run("escapes special characters", func(t *testing.T) {
		t.Parallel()

		specialFile := writeFile(t, "my file.json", `[{"val": "hello"}]`)
		_, stderr, err := runCmdWithStreams(t, "--print-command", "-q", `SELECT val, 'test' AS "alias" FROM my_file`, specialFile)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		wantEscapedQuery := `"SELECT val, 'test' AS \"alias\" FROM my_file"`
		if !strings.Contains(stderr, wantEscapedQuery) {
			t.Errorf("stderr %q does not contain properly escaped query %q", stderr, wantEscapedQuery)
		}
		wantQuotedFile := fmt.Sprintf(`"%s"`, specialFile)
		if !strings.Contains(stderr, wantQuotedFile) {
			t.Errorf("stderr %q does not contain properly escaped file %q", stderr, wantQuotedFile)
		}
	})

	t.Run("does not print command when stdin is used", func(t *testing.T) {
		t.Parallel()

		database, err := db.New()
		if err != nil {
			t.Fatalf("failed to create db: %v", err)
		}
		testutil.CloseDB(t, database)

		cfg := &runConfig{
			query:     "SELECT 1",
			filePaths: []string{jsonPath},
		}
		opts := &options{
			printCommand: true,
			outputFormat: "json",
		}
		var stdout, stderr bytes.Buffer
		if err := execute(database, cfg, opts, &stdout, &stderr, true); err != nil {
			t.Fatalf("execute failed: %v", err)
		}
		if stderr.String() != "" {
			t.Errorf("stderr = %q, want empty when useStdin is true", stderr.String())
		}
	})
}

func TestFormatCLICommand(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		cfg      *runConfig
		opts     *options
		expected string
	}{
		{
			name: "basic query",
			cfg: &runConfig{
				query:     "SELECT * FROM users",
				filePaths: []string{"users.json"},
			},
			opts:     &options{outputFormat: "json"},
			expected: `qo -q "SELECT * FROM users" users.json`,
		},
		{
			name: "with all options",
			cfg: &runConfig{
				query:     "SELECT id FROM data",
				filePaths: []string{"data.csv", "other.json"},
			},
			opts: &options{
				inputFormat:  "csv",
				outputFormat: "table",
				noHeader:     true,
			},
			expected: `qo -i csv -o table --no-header -q "SELECT id FROM data" data.csv other.json`,
		},
		{
			name: "special characters in query and path",
			cfg: &runConfig{
				query:     `SELECT "price" * 1.1, "$total", ` + "`col`" + ` FROM items`,
				filePaths: []string{"path with spaces/file.csv"},
			},
			opts:     &options{outputFormat: "json"},
			expected: "qo -q \"SELECT \\\"price\\\" * 1.1, \\\"\\$total\\\", \\`col\\` FROM items\" \"path with spaces/file.csv\"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := formatCLICommand(tt.cfg, tt.opts)
			if got != tt.expected {
				t.Errorf("formatCLICommand() =\n%q\nwant:\n%q", got, tt.expected)
			}
		})
	}
}
