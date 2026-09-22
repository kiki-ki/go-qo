package input_test

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

func TestNewLoader(t *testing.T) {
	database, err := db.New()
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	testutil.CloseDB(t, database)

	loader := input.NewLoader(database, input.FormatJSON, nil)
	if loader == nil {
		t.Fatal("expected non-nil loader")
	}
}

func TestLoader_LoadFiles(t *testing.T) {
	tests := []struct {
		name          string
		filePath      string
		format        input.Format
		wantErr       bool
		tableName     string
		expectedCount int
		checkName     bool
		expectedName  string
		nameQuery     string
	}{
		{
			name:          "valid multiple.json file",
			filePath:      testutil.JSONTestdataPath("multiple.json"),
			format:        input.FormatJSON,
			wantErr:       false,
			tableName:     "multiple",
			expectedCount: 3,
			checkName:     true,
			expectedName:  "Alice",
			nameQuery:     "SELECT name FROM multiple WHERE id = 1",
		},
		{
			name:     "invalid format",
			filePath: testutil.JSONTestdataPath("multiple.json"),
			format:   "invalid",
			wantErr:  true,
		},
		{
			name:     "file not found",
			filePath: "/nonexistent/file.json",
			format:   input.FormatJSON,
			wantErr:  true,
		},
		{
			name:          "nested JSON",
			filePath:      testutil.JSONTestdataPath("nested.json"),
			format:        input.FormatJSON,
			wantErr:       false,
			tableName:     "nested",
			expectedCount: 2,
			checkName:     false,
		},
		{
			name:     "empty JSON array",
			filePath: testutil.JSONTestdataPath("empty.json"),
			format:   input.FormatJSON,
			wantErr:  true,
		},
		{
			name:     "invalid JSON file",
			filePath: testutil.JSONTestdataPath("invalid.json"),
			format:   input.FormatJSON,
			wantErr:  true,
		},
		{
			name:          "single JSON object",
			filePath:      testutil.JSONTestdataPath("single.json"),
			format:        input.FormatJSON,
			wantErr:       false,
			tableName:     "single",
			expectedCount: 1,
			checkName:     true,
			expectedName:  "Alice",
			nameQuery:     "SELECT name FROM single",
		},
		{
			name:          "JSON with BOM",
			filePath:      testutil.JSONTestdataPath("bom.json"),
			format:        input.FormatJSON,
			wantErr:       false,
			tableName:     "bom",
			expectedCount: 3,
			checkName:     true,
			expectedName:  "Alice",
			nameQuery:     "SELECT name FROM bom WHERE id = 1",
		},
		{
			name:          "CSV with BOM",
			filePath:      testutil.CSVTestdataPath("bom.csv"),
			format:        input.FormatCSV,
			wantErr:       false,
			tableName:     "bom",
			expectedCount: 3,
			checkName:     true,
			expectedName:  "Alice",
			nameQuery:     "SELECT name FROM bom WHERE id = 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			database, err := db.New()
			if err != nil {
				t.Fatalf("failed to create db: %v", err)
			}
			testutil.CloseDB(t, database)

			loader := input.NewLoader(database, tt.format, nil)
			_, err = loader.LoadFiles([]string{tt.filePath})

			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("LoadFiles failed: %v", err)
			}

			// Verify row count
			var count int
			query := "SELECT COUNT(*) FROM " + tt.tableName
			if err := database.QueryRow(query).Scan(&count); err != nil {
				t.Fatalf("query failed: %v", err)
			}
			if count != tt.expectedCount {
				t.Errorf("expected %d rows, got %d", tt.expectedCount, count)
			}

			// Verify specific data if requested
			if tt.checkName {
				var name string
				if err := database.QueryRow(tt.nameQuery).Scan(&name); err != nil {
					t.Fatalf("query failed: %v", err)
				}
				if name != tt.expectedName {
					t.Errorf("expected %s, got %s", tt.expectedName, name)
				}
			}
		})
	}
}

func TestLoader_LoadReader(t *testing.T) {
	database, err := db.New()
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	testutil.CloseDB(t, database)

	jsonData := `[{"id": 1, "name": "Alice"}, {"id": 2, "name": "Bob"}]`
	reader := strings.NewReader(jsonData)

	loader := input.NewLoader(database, input.FormatJSON, nil)
	if _, err := loader.LoadReader(reader, "users"); err != nil {
		t.Fatalf("LoadReader failed: %v", err)
	}

	// Verify data was loaded
	var count int
	if err := database.QueryRow("SELECT COUNT(*) FROM users").Scan(&count); err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 rows, got %d", count)
	}

	// Verify data content
	var name string
	if err := database.QueryRow("SELECT name FROM users WHERE id = 1").Scan(&name); err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if name != "Alice" {
		t.Errorf("expected Alice, got %s", name)
	}
}

func TestLoader_LoadReader_InvalidJSON(t *testing.T) {
	tests := []struct {
		name     string
		jsonData string
		wantErr  bool
	}{
		{
			name:     "invalid JSON syntax",
			jsonData: `{invalid json}`,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			database, err := db.New()
			if err != nil {
				t.Fatalf("failed to create db: %v", err)
			}
			testutil.CloseDB(t, database)

			reader := strings.NewReader(tt.jsonData)
			loader := input.NewLoader(database, input.FormatJSON, nil)
			_, err = loader.LoadReader(reader, "test")

			if tt.wantErr && err == nil {
				t.Error("expected error for invalid JSON")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestLoader_LoadFiles_MultipleFiles(t *testing.T) {
	database, err := db.New()
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	testutil.CloseDB(t, database)

	paths := []string{
		testutil.JSONTestdataPath("multiple.json"),
		testutil.JSONTestdataPath("nested.json"),
	}
	loader := input.NewLoader(database, input.FormatJSON, nil)
	if _, err := loader.LoadFiles(paths); err != nil {
		t.Fatalf("LoadFiles failed: %v", err)
	}

	// Verify both tables exist
	var multipleCount, nestedCount int
	if err := database.QueryRow("SELECT COUNT(*) FROM multiple").Scan(&multipleCount); err != nil {
		t.Fatalf("query multiple failed: %v", err)
	}
	if err := database.QueryRow("SELECT COUNT(*) FROM nested").Scan(&nestedCount); err != nil {
		t.Fatalf("query nested failed: %v", err)
	}
	if multipleCount != 3 {
		t.Errorf("expected 3 multiple records, got %d", multipleCount)
	}
	if nestedCount != 2 {
		t.Errorf("expected 2 nested records, got %d", nestedCount)
	}
}

func TestLoader_LoadReader_BOM(t *testing.T) {
	bom := []byte{0xEF, 0xBB, 0xBF}

	tests := []struct {
		name         string
		data         []byte
		format       input.Format
		expectedName string
	}{
		{
			name:         "CSV with BOM",
			data:         append(bom, []byte("name,age\nAlice,30\n")...),
			format:       input.FormatCSV,
			expectedName: "Alice",
		},
		{
			name:         "JSON with BOM",
			data:         append(bom, []byte(`[{"name":"Alice","age":30}]`)...),
			format:       input.FormatJSON,
			expectedName: "Alice",
		},
		{
			name:         "CSV without BOM",
			data:         []byte("name,age\nAlice,30\n"),
			format:       input.FormatCSV,
			expectedName: "Alice",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			database, err := db.New()
			if err != nil {
				t.Fatalf("failed to create db: %v", err)
			}
			testutil.CloseDB(t, database)

			loader := input.NewLoader(database, tt.format, nil)
			if _, err := loader.LoadReader(bytes.NewReader(tt.data), "test"); err != nil {
				t.Fatalf("LoadReader failed: %v", err)
			}

			var name string
			if err := database.QueryRow("SELECT name FROM test LIMIT 1").Scan(&name); err != nil {
				t.Fatalf("query failed: %v", err)
			}
			if name != tt.expectedName {
				t.Errorf("expected %q, got %q", tt.expectedName, name)
			}
		})
	}
}

func TestSplitArgs(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		wantPaths     []string
		wantRequested bool
	}{
		{"no args", nil, nil, false},
		{"files only", []string{"a.json", "b.json"}, []string{"a.json", "b.json"}, false},
		{"stdin only", []string{"-"}, nil, true},
		{"stdin with files", []string{"-", "a.json"}, []string{"a.json"}, true},
		{"stdin after files", []string{"a.json", "-"}, []string{"a.json"}, true},
		{"repeated stdin", []string{"-", "-"}, nil, true},
		{"dash in filename is a path", []string{"-a.json"}, []string{"-a.json"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			paths, requested := input.SplitArgs(tt.args)
			if !slices.Equal(paths, tt.wantPaths) {
				t.Errorf("paths = %v, want %v", paths, tt.wantPaths)
			}
			if requested != tt.wantRequested {
				t.Errorf("stdinRequested = %v, want %v", requested, tt.wantRequested)
			}
		})
	}
}

func TestUseStdin(t *testing.T) {
	// File arguments must short-circuit before stdin is ever probed: an idle
	// pipe never reaches EOF, and reading it would hang the process.
	t.Run("file args do not use stdin", func(t *testing.T) {
		got, err := input.UseStdin([]string{"a.json"}, false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got {
			t.Error("expected false when file arguments are present")
		}
	})

	t.Run("explicit marker uses stdin alongside files", func(t *testing.T) {
		got, err := input.UseStdin([]string{"a.json"}, true)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !got {
			t.Error("expected true when stdin was explicitly requested")
		}
	})

	t.Run("explicit marker uses stdin without files", func(t *testing.T) {
		got, err := input.UseStdin(nil, true)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !got {
			t.Error("expected true when stdin was explicitly requested")
		}
	})
}

func TestLoader_LoadFiles_NameCollision(t *testing.T) {
	dir := t.TempDir()
	paths := make([]string, 3)
	for i, sub := range []string{"a", "b", "c"} {
		subDir := filepath.Join(dir, sub)
		if err := os.MkdirAll(subDir, 0o755); err != nil {
			t.Fatalf("failed to create dir: %v", err)
		}
		path := filepath.Join(subDir, "data.json")
		content := fmt.Sprintf(`[{"id": %d}]`, i)
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("failed to write file: %v", err)
		}
		paths[i] = path
	}

	database, err := db.New()
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	testutil.CloseDB(t, database)

	loader := input.NewLoader(database, input.FormatJSON, nil)
	names, err := loader.LoadFiles(paths)
	if err != nil {
		t.Fatalf("LoadFiles failed: %v", err)
	}

	want := []string{"data", "data_2", "data_3"}
	if !slices.Equal(names, want) {
		t.Fatalf("names = %v, want %v", names, want)
	}

	// Each table must hold the file it came from, in argument order.
	for i, name := range names {
		var id int
		if err := database.QueryRow("SELECT id FROM " + name).Scan(&id); err != nil {
			t.Fatalf("query %s failed: %v", name, err)
		}
		if id != i {
			t.Errorf("table %s: id = %d, want %d", name, id, i)
		}
	}
}

func TestLoader_LoadFiles_DetectFormat(t *testing.T) {
	dir := t.TempDir()
	write := func(name, content string) string {
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("failed to write %s: %v", name, err)
		}
		return path
	}
	csvPath := write("users.csv", "id,name\n1,alice\n")
	jsonPath := write("logs.json", `[{"uid": 1, "action": "login"}]`)
	unknownPath := write("plain.txt", `[{"n": 1}]`)

	t.Run("resolves each file by extension", func(t *testing.T) {
		database, err := db.New()
		if err != nil {
			t.Fatalf("failed to create db: %v", err)
		}
		testutil.CloseDB(t, database)

		loader := input.NewLoader(database, "", nil)
		if _, err := loader.LoadFiles([]string{csvPath, jsonPath}); err != nil {
			t.Fatalf("LoadFiles failed: %v", err)
		}

		var name, action string
		query := "SELECT users.name, logs.action FROM users JOIN logs ON users.id = logs.uid"
		if err := database.QueryRow(query).Scan(&name, &action); err != nil {
			t.Fatalf("join failed: %v", err)
		}
		if name != "alice" || action != "login" {
			t.Errorf("got %q/%q, want alice/login", name, action)
		}
	})

	t.Run("unknown extension falls back to JSON", func(t *testing.T) {
		database, err := db.New()
		if err != nil {
			t.Fatalf("failed to create db: %v", err)
		}
		testutil.CloseDB(t, database)

		loader := input.NewLoader(database, "", nil)
		if _, err := loader.LoadFiles([]string{unknownPath}); err != nil {
			t.Fatalf("LoadFiles failed: %v", err)
		}

		var n int
		if err := database.QueryRow("SELECT n FROM plain").Scan(&n); err != nil {
			t.Fatalf("query failed: %v", err)
		}
		if n != 1 {
			t.Errorf("n = %d, want 1", n)
		}
	})

	t.Run("explicit format ignores the extension", func(t *testing.T) {
		database, err := db.New()
		if err != nil {
			t.Fatalf("failed to create db: %v", err)
		}
		testutil.CloseDB(t, database)

		// A forced format must win over the extension, so the CSV file is
		// parsed as JSON and fails.
		loader := input.NewLoader(database, input.FormatJSON, nil)
		if _, err := loader.LoadFiles([]string{csvPath}); err == nil {
			t.Error("expected error when parsing CSV as JSON, got nil")
		}
	})
}

func TestLoader_LoadFiles_CaseInsensitiveCollision(t *testing.T) {
	dir := t.TempDir()
	paths := make([]string, 0, 2)
	// SQLite identifiers are case-insensitive, so these two basenames collide
	// even though they differ as Go map keys.
	for i, name := range []string{"Data.json", "data.json"} {
		subDir := filepath.Join(dir, fmt.Sprintf("d%d", i))
		if err := os.MkdirAll(subDir, 0o755); err != nil {
			t.Fatalf("failed to create dir: %v", err)
		}
		path := filepath.Join(subDir, name)
		if err := os.WriteFile(path, []byte(fmt.Sprintf(`[{"id": %d}]`, i)), 0o644); err != nil {
			t.Fatalf("failed to write file: %v", err)
		}
		paths = append(paths, path)
	}

	database, err := db.New()
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	testutil.CloseDB(t, database)

	loader := input.NewLoader(database, input.FormatJSON, nil)
	names, err := loader.LoadFiles(paths)
	if err != nil {
		t.Fatalf("LoadFiles failed: %v", err)
	}

	want := []string{"Data", "data_2"}
	if !slices.Equal(names, want) {
		t.Fatalf("names = %v, want %v", names, want)
	}

	for i, name := range names {
		var id int
		if err := database.QueryRow("SELECT id FROM " + name).Scan(&id); err != nil {
			t.Fatalf("query %s failed: %v", name, err)
		}
		if id != i {
			t.Errorf("table %s: id = %d, want %d", name, id, i)
		}
	}
}
