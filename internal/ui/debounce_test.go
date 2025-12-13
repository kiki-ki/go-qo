package ui

import (
	"database/sql"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

func TestDebounceDelay(t *testing.T) {
	// Verify debounce delay is reasonable (100-300ms range)
	if debounceDelay < 100*time.Millisecond || debounceDelay > 300*time.Millisecond {
		t.Errorf("debounceDelay should be between 100-300ms, got %v", debounceDelay)
	}
}

func TestNewDebounceMsg(t *testing.T) {
	query := "SELECT * FROM users"
	msg := NewDebounceMsg(query)

	// Verify the message is a debounceMsg with correct query
	dm, ok := msg.(debounceMsg)
	if !ok {
		t.Fatal("NewDebounceMsg should return a debounceMsg")
	}
	if dm.query != query {
		t.Errorf("expected query %q, got %q", query, dm.query)
	}
}

func TestHandleDebounceMsg_ExecutesMatchingQuery(t *testing.T) {
	db := setupTestDB(t)
	m := NewModel(db, []string{"test"})

	// Set pending query
	m.pendingQuery = "SELECT * FROM test"

	// Send matching debounce message
	m.handleDebounceMsg(debounceMsg{query: "SELECT * FROM test"})

	// Verify query was executed (lastExecQuery updated)
	if m.lastExecQuery != "SELECT * FROM test" {
		t.Errorf("expected lastExecQuery to be updated, got %q", m.lastExecQuery)
	}
}

func TestHandleDebounceMsg_SkipsNonMatchingQuery(t *testing.T) {
	db := setupTestDB(t)
	m := NewModel(db, []string{"test"})

	// Set pending query to something different
	m.pendingQuery = "SELECT * FROM test WHERE id = 1"

	// Send non-matching debounce message (simulates outdated debounce)
	m.handleDebounceMsg(debounceMsg{query: "SELECT * FROM test"})

	// Verify query was NOT executed
	if m.lastExecQuery != "" {
		t.Errorf("expected lastExecQuery to remain empty, got %q", m.lastExecQuery)
	}
}

func TestHandleDebounceMsg_SkipsDuplicateExecution(t *testing.T) {
	db := setupTestDB(t)
	m := NewModel(db, []string{"test"})

	query := "SELECT * FROM test"
	m.pendingQuery = query
	m.lastExecQuery = query // Already executed

	// Count should not change since it's a duplicate
	initialRows := len(m.tableState.Rows())

	m.handleDebounceMsg(debounceMsg{query: query})

	// Verify no re-execution (rows unchanged from initial state)
	if len(m.tableState.Rows()) != initialRows {
		t.Error("duplicate query should not be re-executed")
	}
}

func TestScheduleQueryExecution(t *testing.T) {
	db := setupTestDB(t)
	m := NewModel(db, []string{"test"})
	m.pendingQuery = "SELECT * FROM test"

	cmd := m.scheduleQueryExecution()
	if cmd == nil {
		t.Fatal("scheduleQueryExecution should return a command")
	}
}

// setupTestDB creates a test database with a simple table.
func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}

	_, err = db.Exec(`CREATE TABLE test (id INTEGER, name TEXT)`)
	if err != nil {
		t.Fatalf("failed to create table: %v", err)
	}

	_, err = db.Exec(`INSERT INTO test (id, name) VALUES (1, 'Alice'), (2, 'Bob')`)
	if err != nil {
		t.Fatalf("failed to insert data: %v", err)
	}

	t.Cleanup(func() { _ = db.Close() })

	return db
}
