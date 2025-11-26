package input_test

import (
	"strings"
	"testing"

	"github.com/kiki-ki/go-qo/internal/db"
	"github.com/kiki-ki/go-qo/internal/input"
	"github.com/kiki-ki/go-qo/testutil"
)

func TestNewLoader(t *testing.T) {
	database, err := db.New()
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	testutil.CloseDB(t, database)

	loader := input.NewLoader(database, input.FormatJSON)
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			database, err := db.New()
			if err != nil {
				t.Fatalf("failed to create db: %v", err)
			}
			testutil.CloseDB(t, database)

			loader := input.NewLoader(database, tt.format)
			err = loader.LoadFiles([]string{tt.filePath})

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

	loader := input.NewLoader(database, input.FormatJSON)
	if err := loader.LoadReader(reader, "users"); err != nil {
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
			loader := input.NewLoader(database, input.FormatJSON)
			err = loader.LoadReader(reader, "test")

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
	loader := input.NewLoader(database, input.FormatJSON)
	if err := loader.LoadFiles(paths); err != nil {
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
