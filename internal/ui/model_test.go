package ui_test

import (
	"database/sql"
	"image/color"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/kiki-ki/go-qo/internal/testutil"
	"github.com/kiki-ki/go-qo/internal/ui"
)

// setupTestTable creates a test database with a simple table.
func setupTestTable(t *testing.T) *sql.DB {
	t.Helper()
	db := testutil.SetupTestDB(t)
	_, err := db.Exec(`
		CREATE TABLE test (id INTEGER, name TEXT);
		INSERT INTO test VALUES (1, 'Alice'), (2, 'Bob');
	`)
	if err != nil {
		t.Fatalf("failed to setup test data: %v", err)
	}
	return db
}

func TestNewModel(t *testing.T) {
	db := setupTestTable(t)
	m := ui.NewModel(db, []string{"test"}, nil)
	view := m.View().Content

	if !strings.Contains(view, "QUERY") {
		t.Error("expected QUERY mode in view")
	}
	if !strings.Contains(view, "Tab") {
		t.Error("expected Tab hint in view")
	}
	if !strings.Contains(view, "SELECT * FROM test") {
		t.Error("expected default query with first table name")
	}
}

func TestModel_Quit(t *testing.T) {
	db := setupTestTable(t)

	tests := []struct {
		name string
		key  tea.KeyPressMsg
	}{
		{"ctrl+c", tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl}},
		{"esc", tea.KeyPressMsg{Code: tea.KeyEscape}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := ui.NewModel(db, []string{"test"}, nil)
			_, cmd := m.Update(tt.key)
			if cmd == nil {
				t.Errorf("expected quit command for %s", tt.name)
			}
		})
	}
}

func TestModel_Init(t *testing.T) {
	db := setupTestTable(t)
	m := ui.NewModel(db, []string{"test"}, nil)
	cmd := m.Init()

	if cmd == nil {
		t.Error("expected non-nil init command (blink)")
	}
}

func TestModel_WindowResize(t *testing.T) {
	db := setupTestTable(t)
	m := ui.NewModel(db, []string{"test"}, nil)
	updated, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	sizes := []tea.WindowSizeMsg{
		{Width: 120, Height: 40},
		{Width: 40, Height: 10},
		{Width: 0, Height: 0},
	}

	for _, size := range sizes {
		updated, _ = updated.Update(size)
		_ = updated.View().Content
	}
}

func TestModel_ErrorDisplay(t *testing.T) {
	db := setupTestTable(t)
	m := ui.NewModel(db, []string{"test"}, nil)

	updated, _ := m.Update(tea.KeyPressMsg{Text: "INVALID SQL"})
	updated, _ = updated.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	view := updated.View().Content
	if !strings.Contains(view, "Error") {
		t.Error("expected error message in view for invalid SQL")
	}
}

func TestModel_EnterExecutesQuery(t *testing.T) {
	db := setupTestTable(t)
	m := ui.NewModel(db, []string{"test"}, nil)

	// Enter executes query but stays in TUI (no quit command)
	updated, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd != nil {
		t.Fatal("Enter should not return quit command")
	}

	// Result should be nil before Esc
	model := updated.(ui.Model)
	if model.Result() != nil {
		t.Error("Result should be nil before Esc")
	}
}

func TestModel_EscReturnsResult(t *testing.T) {
	db := setupTestTable(t)

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"default_query", "", "SELECT * FROM test LIMIT 10"},
		{"custom_query", "SELECT name FROM test", "SELECT name FROM test"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := ui.NewModel(db, []string{"test"}, nil)
			var updated tea.Model = m

			if tt.input != "" {
				for range "SELECT * FROM test LIMIT 10" {
					updated, _ = updated.Update(tea.KeyPressMsg{Code: tea.KeyBackspace})
				}
				updated, _ = updated.Update(tea.KeyPressMsg{Text: tt.input})
			}

			// Execute query first
			updated, _ = updated.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

			// Esc: output result & quit
			updated, cmd := updated.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
			if cmd == nil {
				t.Fatal("Esc should return quit command")
			}

			model := updated.(ui.Model)
			result := model.Result()
			if result == nil || result.Query != tt.want {
				t.Errorf("got %v, want %q", result, tt.want)
			}
		})
	}
}

func TestModel_TableList(t *testing.T) {
	db := setupTestTable(t)
	m := ui.NewModel(db, []string{"users", "orders"}, nil)
	view := m.View().Content

	if !strings.Contains(view, "Tables:") {
		t.Error("expected 'Tables:' in view")
	}
	if !strings.Contains(view, "users, orders") {
		t.Error("expected 'users, orders' in table list")
	}
}

func TestModel_CellDetail(t *testing.T) {
	db := setupTestTable(t)
	m := ui.NewModel(db, []string{"test"}, nil)

	updated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	updated, _ = updated.Update(tea.KeyPressMsg{Code: tea.KeyEnter}) // Execute query
	updated, _ = updated.Update(tea.KeyPressMsg{Code: tea.KeyTab})   // Switch to TABLE mode
	view := updated.View().Content

	if !strings.Contains(view, "TABLE") {
		t.Error("expected TABLE mode indicator")
	}
	if !strings.Contains(view, "(1/2, 1/2)") {
		t.Error("expected cell position in detail")
	}
	if !strings.Contains(view, "id:") {
		t.Error("expected column name 'id' in cell detail")
	}
	if !strings.Contains(view, "▶") {
		t.Error("expected cell marker on the first cell")
	}
}

func TestModel_TableNavigation(t *testing.T) {
	db := setupTestTable(t)
	m := ui.NewModel(db, []string{"test"}, nil)

	updated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	updated, _ = updated.Update(tea.KeyPressMsg{Code: tea.KeyEnter}) // Execute query
	updated, _ = updated.Update(tea.KeyPressMsg{Code: tea.KeyTab})   // Switch to TABLE mode

	view := updated.View().Content
	if !strings.Contains(view, "(1/2, 1/2)") {
		t.Error("expected initial position (1/2, 1/2)")
	}

	updated, _ = updated.Update(tea.KeyPressMsg{Code: tea.KeyRight})
	view = updated.View().Content
	if !strings.Contains(view, "(1/2, 2/2)") {
		t.Error("expected position (1/2, 2/2) after right")
	}
	if !strings.Contains(view, "name:") {
		t.Error("expected column 'name' after right navigation")
	}

	updated, _ = updated.Update(tea.KeyPressMsg{Text: "h"})
	view = updated.View().Content
	if !strings.Contains(view, "(1/2, 1/2)") {
		t.Error("expected position (1/2, 1/2) after h")
	}
	if !strings.Contains(view, "id:") {
		t.Error("expected column 'id' after h navigation")
	}
}

func TestModel_View_UsesAltScreen(t *testing.T) {
	db := setupTestTable(t)
	m := ui.NewModel(db, []string{"test"}, nil)

	if !m.View().AltScreen {
		t.Error("expected the view to use the alternate screen")
	}
}

func TestModel_BackgroundColor(t *testing.T) {
	const (
		white = "\x1b[97m" // base color on a dark background
		black = "\x1b[30m" // base color on a light background
	)
	db := setupTestTable(t)
	m := ui.NewModel(db, []string{"test"}, nil)

	if view := m.View().Content; !strings.Contains(view, white) || strings.Contains(view, black) {
		t.Error("expected dark background styles before the terminal answers")
	}

	updated, _ := m.Update(tea.BackgroundColorMsg{Color: color.White})
	if view := updated.View().Content; !strings.Contains(view, black) || strings.Contains(view, white) {
		t.Error("expected light background styles after a light background color")
	}

	updated, _ = updated.Update(tea.BackgroundColorMsg{Color: color.Black})
	if view := updated.View().Content; !strings.Contains(view, white) || strings.Contains(view, black) {
		t.Error("expected dark background styles after a dark background color")
	}
}

func TestModel_ToggleMode(t *testing.T) {
	db := setupTestTable(t)
	m := ui.NewModel(db, []string{"test"}, nil)

	view := m.View().Content
	if !strings.Contains(view, "QUERY") {
		t.Error("expected QUERY mode initially")
	}
	if !strings.Contains(view, "Tables:") {
		t.Error("expected table list in QUERY mode")
	}

	updated, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	view = updated.View().Content
	if !strings.Contains(view, "TABLE") {
		t.Error("expected TABLE mode after Tab")
	}
	if strings.Contains(view, "Tables:") {
		t.Error("should not show table list in TABLE mode")
	}

	// Tab -> back to QUERY mode
	updated, _ = updated.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	view = updated.View().Content
	if !strings.Contains(view, "QUERY") {
		t.Error("expected QUERY mode after second Tab")
	}
}

func TestExecuteQuery_UpdatesTable(t *testing.T) {
	db := setupTestTable(t)
	m := ui.NewModel(db, []string{"test"}, nil)

	// Execute query
	updated, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	// Switch to TABLE mode to see table data
	updated, _ = updated.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	view := updated.View().Content

	// Verify table contains expected data
	if !strings.Contains(view, "Alice") {
		t.Error("expected 'Alice' in table view")
	}
	if !strings.Contains(view, "Bob") {
		t.Error("expected 'Bob' in table view")
	}

	// Result should still be nil before Esc
	model := updated.(ui.Model)
	if model.Result() != nil {
		t.Error("expected no result before exit")
	}
}

func TestExecuteQuery_SetsErrorOnInvalidSQL(t *testing.T) {
	db := setupTestTable(t)
	m := ui.NewModel(db, []string{"test"}, nil)

	// Clear default query and set invalid one
	var updated tea.Model = m
	for range "SELECT * FROM test LIMIT 10" {
		updated, _ = updated.Update(tea.KeyPressMsg{Code: tea.KeyBackspace})
	}
	updated, _ = updated.Update(tea.KeyPressMsg{Text: "SELECT * FROM does_not_exist"})
	updated, _ = updated.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	view := updated.View().Content
	if !strings.Contains(view, "Error") {
		t.Error("expected error message in view for invalid SQL")
	}
}

func TestModel_ShowsWarnings(t *testing.T) {
	db := setupTestTable(t)
	warning := `column "Id" in table test renamed to "Id_2"`
	m := ui.NewModel(db, []string{"test"}, []string{warning})

	updated, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 24})
	updated, _ = updated.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if view := updated.View().Content; !strings.Contains(view, "Warning: "+warning) {
		t.Errorf("expected warning in view, got:\n%s", view)
	}

	updated, _ = updated.Update(tea.KeyPressMsg{Text: "x"})
	updated, _ = updated.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if view := updated.View().Content; !strings.Contains(view, "Error:") || strings.Contains(view, "Warning:") {
		t.Errorf("expected error instead of warning, got:\n%s", view)
	}
	updated, _ = updated.Update(tea.KeyPressMsg{Code: tea.KeyBackspace})
	updated, _ = updated.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if view := updated.View().Content; !strings.Contains(view, "Warning: "+warning) {
		t.Errorf("expected warning back after a successful query, got:\n%s", view)
	}
}
