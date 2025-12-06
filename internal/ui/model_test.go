package ui_test

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/kiki-ki/go-qo/internal/testutil"
	"github.com/kiki-ki/go-qo/internal/ui"
)

func TestNewModel(t *testing.T) {
	db := testutil.SetupTestDB(t)
	_, err := db.Exec(`
	CREATE TABLE test (id INTEGER, name TEXT);
	INSERT INTO test VALUES (1, 'Alice'), (2, 'Bob');
`)
	if err != nil {
		t.Fatalf("failed to setup test data: %v", err)
	}

	m := ui.NewModel(db, []string{"test_table"})

	// Verify initial state via View output
	view := m.View()
	if !strings.Contains(view, "QUERY") {
		t.Error("expected QUERY mode in view")
	}
}

func TestModel_View(t *testing.T) {
	db := testutil.SetupTestDB(t)
	_, err := db.Exec(`
	CREATE TABLE test (id INTEGER, name TEXT);
	INSERT INTO test VALUES (1, 'Alice'), (2, 'Bob');
`)
	if err != nil {
		t.Fatalf("failed to setup test data: %v", err)
	}

	m := ui.NewModel(db, []string{"test_table"})
	view := m.View()

	// Check view contains expected elements
	if !strings.Contains(view, "QUERY") {
		t.Error("expected 'QUERY' mode in view")
	}
	if !strings.Contains(view, "Tab") {
		t.Error("expected Tab hint in view")
	}
}

func TestModel_Update_Quit(t *testing.T) {
	db := testutil.SetupTestDB(t)
	_, err := db.Exec(`
	CREATE TABLE test (id INTEGER, name TEXT);
	INSERT INTO test VALUES (1, 'Alice'), (2, 'Bob');
`)
	if err != nil {
		t.Fatalf("failed to setup test data: %v", err)
	}

	m := ui.NewModel(db, []string{"test_table"})

	// Test Ctrl+C quits
	newModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if newModel == nil {
		t.Error("expected non-nil model")
	}
	if cmd == nil {
		t.Error("expected quit command")
	}

	// Test Esc quits
	m = ui.NewModel(db, []string{"test_table"})
	newModel, cmd = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if newModel == nil {
		t.Error("expected non-nil model")
	}
	if cmd == nil {
		t.Error("expected quit command")
	}
}

func TestModel_Update_TabTogglesFocus(t *testing.T) {
	db := testutil.SetupTestDB(t)
	_, err := db.Exec(`
	CREATE TABLE test (id INTEGER, name TEXT);
	INSERT INTO test VALUES (1, 'Alice'), (2, 'Bob');
`)
	if err != nil {
		t.Fatalf("failed to setup test data: %v", err)
	}

	m := ui.NewModel(db, []string{"test_table"})

	// Press Tab to toggle focus
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})

	// The model should have updated (we can't check internal state directly
	// in external test, but we verify no panic and model is returned)
	if newModel == nil {
		t.Error("expected non-nil model after Tab")
	}
}

func TestModel_Init(t *testing.T) {
	db := testutil.SetupTestDB(t)
	_, err := db.Exec(`
	CREATE TABLE test (id INTEGER, name TEXT);
	INSERT INTO test VALUES (1, 'Alice'), (2, 'Bob');
`)
	if err != nil {
		t.Fatalf("failed to setup test data: %v", err)
	}

	m := ui.NewModel(db, []string{"test_table"})
	cmd := m.Init()

	// Init should return a blink command for textinput
	if cmd == nil {
		t.Error("expected non-nil init command")
	}
}

func TestModel_View_TableMode(t *testing.T) {
	db := testutil.SetupTestDB(t)
	_, err := db.Exec(`
	CREATE TABLE test (id INTEGER, name TEXT);
	INSERT INTO test VALUES (1, 'Alice'), (2, 'Bob');
`)
	if err != nil {
		t.Fatalf("failed to setup test data: %v", err)
	}

	m := ui.NewModel(db, []string{"test"})

	// Switch to table mode
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})

	view := updated.View()
	if !strings.Contains(view, "TABLE") {
		t.Error("expected TABLE mode in view after Tab")
	}
}

func TestModel_View_ErrorDisplay(t *testing.T) {
	db := testutil.SetupTestDB(t)
	_, err := db.Exec(`
	CREATE TABLE test (id INTEGER, name TEXT);
	INSERT INTO test VALUES (1, 'Alice'), (2, 'Bob');
`)
	if err != nil {
		t.Fatalf("failed to setup test data: %v", err)
	}

	m := ui.NewModel(db, []string{"test"})

	// Enter invalid SQL to trigger error
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("INVALID SQL")})

	// Get model to access pending query
	model := updated.(ui.Model)

	// Trigger debounce to execute the query
	updated, _ = updated.Update(ui.NewDebounceMsg(model.PendingQuery()))

	view := updated.View()
	if !strings.Contains(view, "Error") {
		t.Error("expected error message in view for invalid SQL")
	}
}

func TestModel_Update_EnterReturnsResult(t *testing.T) {
	db := testutil.SetupTestDB(t)
	_, err := db.Exec(`
	CREATE TABLE test (id INTEGER, name TEXT);
	INSERT INTO test VALUES (1, 'Alice'), (2, 'Bob');
`)
	if err != nil {
		t.Fatalf("failed to setup test data: %v", err)
	}

	m := ui.NewModel(db, []string{"test"})

	// Press Enter in query mode should quit with result
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Error("expected quit command on Enter")
	}

	// Verify result contains query
	model := updated.(ui.Model)
	result := model.Result()
	if result == nil {
		t.Fatal("expected non-nil result")
		return
	}
	if !strings.Contains(result.Query, "SELECT") {
		t.Errorf("expected query to contain SELECT, got %q", result.Query)
	}
}

func TestModel_View_TableList(t *testing.T) {
	db := testutil.SetupTestDB(t)
	_, err := db.Exec(`
	CREATE TABLE test (id INTEGER, name TEXT);
	INSERT INTO test VALUES (1, 'Alice'), (2, 'Bob');
`)
	if err != nil {
		t.Fatalf("failed to setup test data: %v", err)
	}

	m := ui.NewModel(db, []string{"users", "orders"})
	view := m.View()

	// Check table list is displayed in query mode
	if !strings.Contains(view, "Tables:") {
		t.Error("expected 'Tables:' in view")
	}
	if !strings.Contains(view, "users") {
		t.Error("expected 'users' in table list")
	}
	if !strings.Contains(view, "orders") {
		t.Error("expected 'orders' in table list")
	}
}

func TestModel_View_CellDetail(t *testing.T) {
	db := testutil.SetupTestDB(t)
	_, err := db.Exec(`
	CREATE TABLE test (id INTEGER, name TEXT);
	INSERT INTO test VALUES (1, 'Alice'), (2, 'Bob');
`)
	if err != nil {
		t.Fatalf("failed to setup test data: %v", err)
	}

	m := ui.NewModel(db, []string{"test"})

	// Trigger window resize to initialize dimensions (needed for query execution)
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	// Get model to access pending query
	model := updated.(ui.Model)

	// Trigger debounce to execute the initial query
	updated, _ = updated.Update(ui.NewDebounceMsg(model.PendingQuery()))

	// Switch to table mode
	updated, _ = updated.Update(tea.KeyMsg{Type: tea.KeyTab})
	view := updated.View()

	// In table mode, should show TABLE header
	if !strings.Contains(view, "TABLE") {
		t.Error("expected TABLE mode indicator")
	}

	// Should show data (Alice and Bob from test table)
	if !strings.Contains(view, "Alice") {
		t.Error("expected data to be displayed")
	}
}

func TestModel_TableNavigation(t *testing.T) {
	db := testutil.SetupTestDB(t)
	_, err := db.Exec(`
	CREATE TABLE test (id INTEGER, name TEXT);
	INSERT INTO test VALUES (1, 'Alice'), (2, 'Bob');
`)
	if err != nil {
		t.Fatalf("failed to setup test data: %v", err)
	}

	m := ui.NewModel(db, []string{"test"})

	// Initialize dimensions and execute query
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	// Switch to table mode
	updated, _ = updated.Update(tea.KeyMsg{Type: tea.KeyTab})

	// Navigate right with arrow key
	updated, _ = updated.Update(tea.KeyMsg{Type: tea.KeyRight})
	view := updated.View()
	if !strings.Contains(view, "TABLE") {
		t.Error("expected TABLE mode after right navigation")
	}

	// Navigate left with arrow key
	updated, _ = updated.Update(tea.KeyMsg{Type: tea.KeyLeft})
	view = updated.View()
	if !strings.Contains(view, "TABLE") {
		t.Error("expected TABLE mode after left navigation")
	}

	// Navigate with vim keys (h/l)
	updated, _ = updated.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("l")})
	updated, _ = updated.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("h")})
	view = updated.View()
	if !strings.Contains(view, "TABLE") {
		t.Error("expected TABLE mode after vim navigation")
	}
}

func TestModel_ToggleMode_BackToQuery(t *testing.T) {
	db := testutil.SetupTestDB(t)
	_, err := db.Exec(`
	CREATE TABLE test (id INTEGER, name TEXT);
	INSERT INTO test VALUES (1, 'Alice'), (2, 'Bob');
`)
	if err != nil {
		t.Fatalf("failed to setup test data: %v", err)
	}

	m := ui.NewModel(db, []string{"test"})

	// Switch to table mode
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	view := updated.View()
	if !strings.Contains(view, "TABLE") {
		t.Error("expected TABLE mode")
	}

	// Switch back to query mode
	updated, _ = updated.Update(tea.KeyMsg{Type: tea.KeyTab})
	view = updated.View()
	if !strings.Contains(view, "QUERY") {
		t.Error("expected QUERY mode after second Tab")
	}
}
