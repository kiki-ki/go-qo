package ui_test

import (
	"testing"

	"github.com/kiki-ki/go-qo/internal/testutil"
	"github.com/kiki-ki/go-qo/internal/ui"
)

func TestSQLRowsToTable(t *testing.T) {
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

	cols, tableRows, err := ui.SQLRowsToTable(rows)
	if err != nil {
		t.Fatalf("SQLRowsToTable failed: %v", err)
	}

	// Verify columns
	if len(cols) != 2 {
		t.Errorf("expected 2 columns, got %d", len(cols))
	}
	if cols[0].Title != "id" {
		t.Errorf("expected first column 'id', got %s", cols[0].Title)
	}
	if cols[1].Title != "name" {
		t.Errorf("expected second column 'name', got %s", cols[1].Title)
	}

	// Verify rows
	if len(tableRows) != 2 {
		t.Errorf("expected 2 rows, got %d", len(tableRows))
	}
	if tableRows[0][1] != "Alice" {
		t.Errorf("expected 'Alice', got %s", tableRows[0][1])
	}
	if tableRows[1][1] != "Bob" {
		t.Errorf("expected 'Bob', got %s", tableRows[1][1])
	}
}

func TestSQLRowsToTable_EmptyResult(t *testing.T) {
	db := testutil.SetupTestDB(t)
	_, err := db.Exec(`
	CREATE TABLE test (id INTEGER, name TEXT);
	INSERT INTO test VALUES (1, 'Alice'), (2, 'Bob');
`)
	if err != nil {
		t.Fatalf("failed to setup test data: %v", err)
	}

	rows, err := db.Query("SELECT * FROM test WHERE id = 999")
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}
	testutil.CloseRows(t, rows)

	cols, tableRows, err := ui.SQLRowsToTable(rows)
	if err != nil {
		t.Fatalf("SQLRowsToTable failed: %v", err)
	}

	if len(cols) != 2 {
		t.Errorf("expected 2 columns, got %d", len(cols))
	}
	if len(tableRows) != 0 {
		t.Errorf("expected 0 rows, got %d", len(tableRows))
	}
}
