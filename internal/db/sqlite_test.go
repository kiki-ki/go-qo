package db_test

import (
	"testing"

	"github.com/kiki-ki/go-qo/internal/db"
	"github.com/kiki-ki/go-qo/internal/parser"
)

func TestNew(t *testing.T) {
	database, err := db.New()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer database.Close()

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
	defer database.Close()

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
	}

	for _, tt := range tests {
		if got := db.TableNameFromPath(tt.path); got != tt.want {
			t.Errorf("TableNameFromPath(%q) = %q, want %q", tt.path, got, tt.want)
		}
	}
}
