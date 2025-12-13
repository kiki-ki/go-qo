package ui_test

import (
	"database/sql"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

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
	m := ui.NewModel(db, []string{"test"})
	view := m.View()

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
		name    string
		keyType tea.KeyType
	}{
		{"ctrl+c", tea.KeyCtrlC},
		{"esc", tea.KeyEsc},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := ui.NewModel(db, []string{"test"})
			_, cmd := m.Update(tea.KeyMsg{Type: tt.keyType})
			if cmd == nil {
				t.Errorf("expected quit command for %s", tt.name)
			}
		})
	}
}

func TestModel_Init(t *testing.T) {
	db := setupTestTable(t)
	m := ui.NewModel(db, []string{"test"})
	cmd := m.Init()

	if cmd == nil {
		t.Error("expected non-nil init command (blink)")
	}
}

func TestModel_WindowResize(t *testing.T) {
	db := setupTestTable(t)
	m := ui.NewModel(db, []string{"test"})
	updated, _ := m.Update(ui.NewDebounceMsg(m.PendingQuery()))

	sizes := []tea.WindowSizeMsg{
		{Width: 120, Height: 40},
		{Width: 40, Height: 10},
		{Width: 0, Height: 0},
	}

	for _, size := range sizes {
		updated, _ = updated.Update(size)
		_ = updated.View()
	}
}

func TestModel_ErrorDisplay(t *testing.T) {
	db := setupTestTable(t)
	m := ui.NewModel(db, []string{"test"})

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("INVALID SQL")})
	model := updated.(ui.Model)
	updated, _ = updated.Update(ui.NewDebounceMsg(model.PendingQuery()))

	view := updated.View()
	if !strings.Contains(view, "Error") {
		t.Error("expected error message in view for invalid SQL")
	}
}

func TestModel_EnterReturnsResult(t *testing.T) {
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
			m := ui.NewModel(db, []string{"test"})
			var updated tea.Model = m

			if tt.input != "" {
				for range "SELECT * FROM test LIMIT 10" {
					updated, _ = updated.Update(tea.KeyMsg{Type: tea.KeyBackspace})
				}
				updated, _ = updated.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.input)})
			}

			updated, cmd := updated.Update(tea.KeyMsg{Type: tea.KeyEnter})
			if cmd == nil {
				t.Fatal("expected quit command on Enter")
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
	m := ui.NewModel(db, []string{"users", "orders"})
	view := m.View()

	if !strings.Contains(view, "Tables:") {
		t.Error("expected 'Tables:' in view")
	}
	if !strings.Contains(view, "users, orders") {
		t.Error("expected 'users, orders' in table list")
	}
}

func TestModel_CellDetail(t *testing.T) {
	db := setupTestTable(t)
	m := ui.NewModel(db, []string{"test"})

	updated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	model := updated.(ui.Model)
	updated, _ = updated.Update(ui.NewDebounceMsg(model.PendingQuery()))
	updated, _ = updated.Update(tea.KeyMsg{Type: tea.KeyTab})
	view := updated.View()

	if !strings.Contains(view, "TABLE") {
		t.Error("expected TABLE mode indicator")
	}
	if !strings.Contains(view, "(1/2, 1/2)") {
		t.Error("expected cell position in detail")
	}
	if !strings.Contains(view, "id:") {
		t.Error("expected column name 'id' in cell detail")
	}
}

func TestModel_TableNavigation(t *testing.T) {
	db := setupTestTable(t)
	m := ui.NewModel(db, []string{"test"})

	updated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	model := updated.(ui.Model)
	updated, _ = updated.Update(ui.NewDebounceMsg(model.PendingQuery()))
	updated, _ = updated.Update(tea.KeyMsg{Type: tea.KeyTab})

	view := updated.View()
	if !strings.Contains(view, "(1/2, 1/2)") {
		t.Error("expected initial position (1/2, 1/2)")
	}

	updated, _ = updated.Update(tea.KeyMsg{Type: tea.KeyRight})
	view = updated.View()
	if !strings.Contains(view, "(1/2, 2/2)") {
		t.Error("expected position (1/2, 2/2) after right")
	}
	if !strings.Contains(view, "name:") {
		t.Error("expected column 'name' after right navigation")
	}

	updated, _ = updated.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("h")})
	view = updated.View()
	if !strings.Contains(view, "(1/2, 1/2)") {
		t.Error("expected position (1/2, 1/2) after h")
	}
	if !strings.Contains(view, "id:") {
		t.Error("expected column 'id' after h navigation")
	}
}

func TestModel_ToggleMode(t *testing.T) {
	db := setupTestTable(t)
	m := ui.NewModel(db, []string{"test"})

	view := m.View()
	if !strings.Contains(view, "QUERY") {
		t.Error("expected QUERY mode initially")
	}
	if !strings.Contains(view, "Tables:") {
		t.Error("expected table list in QUERY mode")
	}

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	view = updated.View()
	if !strings.Contains(view, "TABLE") {
		t.Error("expected TABLE mode after Tab")
	}
	if strings.Contains(view, "Tables:") {
		t.Error("should not show table list in TABLE mode")
	}

	// Tab -> back to QUERY mode
	updated, _ = updated.Update(tea.KeyMsg{Type: tea.KeyTab})
	view = updated.View()
	if !strings.Contains(view, "QUERY") {
		t.Error("expected QUERY mode after second Tab")
	}
}
