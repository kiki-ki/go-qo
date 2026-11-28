package db_test

import (
	"slices"
	"testing"

	"github.com/kiki-ki/go-qo/internal/db"
	"github.com/kiki-ki/go-qo/internal/parser"
	"github.com/kiki-ki/go-qo/internal/testutil"
)

func TestNew(t *testing.T) {
	database, err := db.New()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	testutil.CloseDB(t, database)

	var result int
	if err := database.QueryRow("SELECT 1").Scan(&result); err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if result != 1 {
		t.Errorf("expected 1, got %d", result)
	}
}

func TestDB_LoadData(t *testing.T) {
	database, err := db.New()
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	testutil.CloseDB(t, database)

	tests := []struct {
		name    string
		data    *parser.ParsedData
		wantErr bool
	}{
		{
			name: "normal data",
			data: &parser.ParsedData{
				Columns: []parser.Column{
					{Name: "id", Type: parser.TypeInteger},
					{Name: "name", Type: parser.TypeText},
				},
				Rows: [][]any{{int64(1), "Alice"}, {int64(2), "Bob"}},
			},
		},
		{
			name: "empty rows",
			data: &parser.ParsedData{
				Columns: []parser.Column{{Name: "id", Type: parser.TypeInteger}},
				Rows:    [][]any{},
			},
		},
		{
			name: "null values",
			data: &parser.ParsedData{
				Columns: []parser.Column{
					{Name: "id", Type: parser.TypeInteger},
					{Name: "value", Type: parser.TypeText},
				},
				Rows: [][]any{{int64(1), nil}, {int64(2), "test"}},
			},
		},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tableName := "test_" + string(rune('a'+i))
			err := database.LoadData(tableName, tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadData() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUniqueColumns(t *testing.T) {
	columns := []parser.Column{
		{Name: "a", Type: parser.TypeInteger},
		{Name: "A", Type: parser.TypeInteger},
		{Name: "a_2", Type: parser.TypeInteger},
		{Name: "", Type: parser.TypeText},
		{Name: "", Type: parser.TypeText},
	}

	got := db.UniqueColumns(columns)

	var names []string
	for _, col := range got {
		names = append(names, col.Name)
	}
	want := []string{"a", "A_3", "a_2", "", "_2"}
	if !slices.Equal(names, want) {
		t.Errorf("names = %q, want %q", names, want)
	}
	if columns[1].Name != "A" {
		t.Errorf("UniqueColumns mutated its input: %q", columns[1].Name)
	}
}

func TestDB_LoadData_ColumnNameWithBacktick(t *testing.T) {
	database, err := db.New()
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	testutil.CloseDB(t, database)

	parsed, err := parser.ParseJSONBytes([]byte("[{\"a`b\":1}]"))
	if err != nil {
		t.Fatalf("failed to parse test data: %v", err)
	}
	if err := database.LoadData("backtick_test", parsed); err != nil {
		t.Fatalf("LoadData() error = %v", err)
	}

	var got int64
	if err := database.QueryRow("SELECT \"a`b\" FROM backtick_test").Scan(&got); err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if got != 1 {
		t.Errorf("queried value = %d, want 1", got)
	}
}

func TestTableNameFromPath(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{"data.json", "data"},
		{"path/to/users.json", "users"},
		{"my-file.json", "my_file"},
		{"file with spaces.json", "file_with_spaces"},
		{"file.name.json", "file_name"},
		{"/absolute/path/data.json", "data"},
		// Unquoted SQLite identifiers cannot start with a digit.
		{"1-report.json", "_1_report"},
		{"report1.json", "report1"},
		// Anything else invalid in an identifier is replaced.
		{"we!rd@name.json", "we_rd_name"},
		// Non-ASCII passes through, combining marks included: macOS stores
		// filenames decomposed, so "ガス" can arrive as カ + U+3099 + ス.
		{"売上.json", "売上"},
		{"\u30ab\u3099\u30b9.json", "\u30ab\u3099\u30b9"},
		// Degenerate names still have to yield a usable identifier.
		{".json", "_"},
	}

	for _, tt := range tests {
		if got := db.TableNameFromPath(tt.path); got != tt.want {
			t.Errorf("TableNameFromPath(%q) = %q, want %q", tt.path, got, tt.want)
		}
	}
}
