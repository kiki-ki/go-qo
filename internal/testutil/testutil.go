// Package testutil provides common test utilities.
package testutil

import (
	"database/sql"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	_ "modernc.org/sqlite"
)

// SetupTestDB creates an in-memory SQLite database.
func SetupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	CloseDB(t, db)
	return db
}

// CreateTempJSON creates a temporary JSON file with the given content.
func CreateTempJSON(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.json")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	return path
}

// TestdataPath returns the absolute path to a file in the testdata directory.
// Use format subdirectories like "json/users.json".
func TestdataPath(filename string) string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "..", "..", "testdata", filename)
}

// JSONTestdataPath returns the absolute path to a JSON file in testdata/json.
func JSONTestdataPath(filename string) string {
	return TestdataPath(filepath.Join("json", filename))
}

// CloseDB registers a cleanup function to close the database when the test completes.
func CloseDB(t *testing.T, c io.Closer) {
	t.Helper()
	t.Cleanup(func() {
		if err := c.Close(); err != nil {
			t.Errorf("failed to close db: %v", err)
		}
	})
}

// CloseRows registers a cleanup function to close the rows when the test completes.
func CloseRows(t *testing.T, rows *sql.Rows) {
	t.Helper()
	t.Cleanup(func() {
		if err := rows.Close(); err != nil {
			t.Errorf("failed to close rows: %v", err)
		}
	})
}
