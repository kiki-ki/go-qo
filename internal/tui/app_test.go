package tui_test

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kiki-ki/go-qo/internal/tui"
	"github.com/kiki-ki/go-qo/testutil"
)

func TestNewModel(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()

	m := tui.NewModel(db, []string{"test_table"})

	// Verify initial state via View output
	view := m.View()
	if !strings.Contains(view, "QUERY") {
		t.Error("expected QUERY mode in view")
	}
}

func TestModel_View(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()

	m := tui.NewModel(db, []string{"test_table"})
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
	defer db.Close()

	m := tui.NewModel(db, []string{"test_table"})

	// Test Ctrl+C quits
	newModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if newModel == nil {
		t.Error("expected non-nil model")
	}
	if cmd == nil {
		t.Error("expected quit command")
	}

	// Test Esc quits
	m = tui.NewModel(db, []string{"test_table"})
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
	defer db.Close()

	m := tui.NewModel(db, []string{"test_table"})

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
	defer db.Close()

	m := tui.NewModel(db, []string{"test_table"})
	cmd := m.Init()

	// Init should return a blink command for textinput
	if cmd == nil {
		t.Error("expected non-nil init command")
	}
}

func TestModel_View_TableMode(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()

	m := tui.NewModel(db, []string{"test"})

	// Switch to table mode
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})

	view := updated.View()
	if !strings.Contains(view, "TABLE") {
		t.Error("expected TABLE mode in view after Tab")
	}
}

func TestModel_View_ErrorDisplay(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()

	m := tui.NewModel(db, []string{"test"})

	// Enter invalid SQL to trigger error
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("INVALID SQL")})

	view := updated.View()
	if !strings.Contains(view, "Error") {
		t.Error("expected error message in view for invalid SQL")
	}
}
