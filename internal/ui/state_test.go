package ui_test

import (
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/table"

	"github.com/kiki-ki/go-qo/internal/testutil"
	"github.com/kiki-ki/go-qo/internal/ui"
)

func TestNewTableState(t *testing.T) {
	s := ui.NewTableState()
	if s == nil {
		t.Fatal("expected non-nil TableState")
	}
	if len(s.Columns()) != 0 {
		t.Errorf("expected 0 columns, got %d", len(s.Columns()))
	}
	if len(s.Rows()) != 0 {
		t.Errorf("expected 0 rows, got %d", len(s.Rows()))
	}
}

func TestTableState_SetData(t *testing.T) {
	s := ui.NewTableState()
	cols := []table.Column{
		{Title: "id", Width: 10},
		{Title: "name", Width: 20},
	}
	rows := []table.Row{
		{"1", "Alice"},
		{"2", "Bob"},
	}

	s.SetData(cols, rows)

	if len(s.Columns()) != 2 {
		t.Errorf("expected 2 columns, got %d", len(s.Columns()))
	}
	if len(s.Rows()) != 2 {
		t.Errorf("expected 2 rows, got %d", len(s.Rows()))
	}
	if s.ColCursor() != 0 {
		t.Errorf("expected colCursor to be reset to 0, got %d", s.ColCursor())
	}
}

func TestTableState_MoveLeftRight(t *testing.T) {
	s := ui.NewTableState()
	s.SetData(
		[]table.Column{
			{Title: "a", Width: 10},
			{Title: "b", Width: 10},
			{Title: "c", Width: 10},
		},
		[]table.Row{
			{"1", "2", "3"},
		},
	)
	// Move right (1)
	if !s.MoveRight() {
		t.Error("MoveRight should return true")
	}
	if s.ColCursor() != 1 {
		t.Errorf("expected cursor 1, got %d", s.ColCursor())
	}
	// Move right (2)
	if !s.MoveRight() {
		t.Error("MoveRight should return true")
	}
	if s.ColCursor() != 2 {
		t.Errorf("expected cursor 2, got %d", s.ColCursor())
	}
	// Move right (3)
	if s.MoveRight() {
		t.Error("MoveRight at end should return false")
	}
	if s.ColCursor() != 2 {
		t.Errorf("cursor should stay at 2, got %d", s.ColCursor())
	}
	// Move left (1)
	if !s.MoveLeft() {
		t.Error("MoveLeft should return true")
	}
	if s.ColCursor() != 1 {
		t.Errorf("expected cursor 1, got %d", s.ColCursor())
	}
	// Move left (2)
	if !s.MoveLeft() {
		t.Error("MoveLeft should return true")
	}
	if s.ColCursor() != 0 {
		t.Errorf("expected cursor 0, got %d", s.ColCursor())
	}
	// Move left (3)
	if s.MoveLeft() {
		t.Error("MoveLeft at end should return false")
	}
	if s.ColCursor() != 0 {
		t.Errorf("cursor should stay at 0, got %d", s.ColCursor())
	}
}

func TestTableState_AdjustOffset(t *testing.T) {
	tests := []struct {
		name           string
		numCols        int
		moveRightCount int
		visibleCols    int
		wantChanged    bool
		wantStart      int
		wantEnd        int
	}{
		{
			name:           "offset shifts when cursor exceeds visible range",
			numCols:        10,
			moveRightCount: 3,
			visibleCols:    3,
			wantChanged:    true,
			wantStart:      1,
			wantEnd:        4,
		},
		{
			name:           "offset unchanged when cursor within range",
			numCols:        3,
			moveRightCount: 0,
			visibleCols:    3,
			wantChanged:    false,
			wantStart:      0,
			wantEnd:        3,
		},
		{
			name:           "offset unchanged after moving within range",
			numCols:        5,
			moveRightCount: 1,
			visibleCols:    3,
			wantChanged:    false,
			wantStart:      0,
			wantEnd:        3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := ui.NewTableState()
			cols := make([]table.Column, tt.numCols)
			row := make(table.Row, tt.numCols)
			for i := range cols {
				cols[i] = table.Column{Title: string(rune('a' + i)), Width: 10}
				row[i] = string(rune('1' + i))
			}
			s.SetData(cols, []table.Row{row})

			for i := 0; i < tt.moveRightCount; i++ {
				s.MoveRight()
			}

			changed := s.AdjustOffset(tt.visibleCols)
			if changed != tt.wantChanged {
				t.Errorf("AdjustOffset() changed = %v, want %v", changed, tt.wantChanged)
			}

			start, end := s.VisibleColumnRange(tt.visibleCols)
			if start != tt.wantStart {
				t.Errorf("VisibleColumnRange() start = %d, want %d", start, tt.wantStart)
			}
			if end != tt.wantEnd {
				t.Errorf("VisibleColumnRange() end = %d, want %d", end, tt.wantEnd)
			}
		})
	}
}

func TestTableState_SelectedCell(t *testing.T) {
	s := ui.NewTableState()
	cols := []table.Column{
		{Title: "id", Width: 10},
		{Title: "name", Width: 20},
	}
	rows := []table.Row{
		{"1", "Alice"},
		{"2", "Bob"},
	}
	s.SetData(cols, rows)

	tests := []struct {
		name        string
		moveRight   int
		rowIdx      int
		wantColName string
		wantValue   string
		wantOk      bool
	}{
		{"first cell", 0, 0, "id", "1", true},
		{"second column first row", 1, 0, "name", "Alice", true},
		{"second column second row", 1, 1, "name", "Bob", true},
		{"negative row", 0, -1, "", "", false},
		{"out of range row", 0, 10, "", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset state
			s.SetData(cols, rows)
			for i := 0; i < tt.moveRight; i++ {
				s.MoveRight()
			}

			colName, value, ok := s.SelectedCell(tt.rowIdx)
			if ok != tt.wantOk {
				t.Errorf("SelectedCell() ok = %v, want %v", ok, tt.wantOk)
			}
			if ok {
				if colName != tt.wantColName {
					t.Errorf("SelectedCell() colName = %s, want %s", colName, tt.wantColName)
				}
				if value != tt.wantValue {
					t.Errorf("SelectedCell() value = %s, want %s", value, tt.wantValue)
				}
			}
		})
	}
}

func TestTableState_SelectedCell_Empty(t *testing.T) {
	s := ui.NewTableState()

	_, _, ok := s.SelectedCell(0)
	if ok {
		t.Error("expected ok to be false for empty state")
	}
}

func TestTableState_Position(t *testing.T) {
	s := ui.NewTableState()
	cols := []table.Column{
		{Title: "a", Width: 10},
		{Title: "b", Width: 10},
		{Title: "c", Width: 10},
	}
	rows := []table.Row{
		{"1", "2", "3"},
		{"4", "5", "6"},
	}
	s.SetData(cols, rows)

	row, totalRows, col, totalCols := s.Position(0)
	if row != 1 || totalRows != 2 || col != 1 || totalCols != 3 {
		t.Errorf("expected (1, 2, 1, 3), got (%d, %d, %d, %d)", row, totalRows, col, totalCols)
	}

	s.MoveRight()
	s.MoveRight()
	row, totalRows, col, totalCols = s.Position(1)
	if row != 2 || totalRows != 2 || col != 3 || totalCols != 3 {
		t.Errorf("expected (2, 2, 3, 3), got (%d, %d, %d, %d)", row, totalRows, col, totalCols)
	}
}

func TestTableState_BuildVisibleColumns(t *testing.T) {
	s := ui.NewTableState()
	cols := []table.Column{
		{Title: "a", Width: 10},
		{Title: "b", Width: 10},
		{Title: "c", Width: 10},
		{Title: "d", Width: 10},
	}
	s.SetData(cols, []table.Row{})

	visible := s.BuildVisibleColumns(2, 20)
	if len(visible) != 2 {
		t.Errorf("expected 2 visible columns, got %d", len(visible))
	}
	if visible[0].Title != "a" || visible[1].Title != "b" {
		t.Errorf("expected columns a, b, got %s, %s", visible[0].Title, visible[1].Title)
	}
	if visible[0].Width != 20 {
		t.Errorf("expected width 20, got %d", visible[0].Width)
	}
}

func TestTableState_BuildVisibleRows(t *testing.T) {
	s := ui.NewTableState()
	cols := []table.Column{
		{Title: "a", Width: 10},
		{Title: "b", Width: 10},
	}
	rows := []table.Row{
		{"1", "2"},
		{"3", "4"},
	}
	s.SetData(cols, rows)

	visible := s.BuildVisibleRows(0, 2)
	if len(visible) != 2 {
		t.Errorf("expected 2 visible rows, got %d", len(visible))
	}
	if len(visible[0]) != 2 {
		t.Errorf("expected 2 cells in row, got %d", len(visible[0]))
	}

	// Selected cell (row 0, col 0) should contain marker
	if !strings.Contains(visible[0][0], "▶") {
		t.Errorf("expected marker in selected cell, got %s", visible[0][0])
	}
	// Non-selected cell should not contain marker
	if strings.Contains(visible[0][1], "▶") {
		t.Errorf("expected no marker in non-selected cell, got %s", visible[0][1])
	}
	if strings.Contains(visible[1][0], "▶") {
		t.Errorf("expected no marker in non-selected row, got %s", visible[1][0])
	}
}

func TestSQLRowsToTableData(t *testing.T) {
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

	cols, tableRows, err := ui.SQLRowsToTableData(rows)
	if err != nil {
		t.Fatalf("SQLRowsToTableData failed: %v", err)
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

func TestSQLRowsToTableData_EmptyResult(t *testing.T) {
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

	cols, tableRows, err := ui.SQLRowsToTableData(rows)
	if err != nil {
		t.Fatalf("SQLRowsToTableData failed: %v", err)
	}

	if len(cols) != 2 {
		t.Errorf("expected 2 columns, got %d", len(cols))
	}
	if len(tableRows) != 0 {
		t.Errorf("expected 0 rows, got %d", len(tableRows))
	}
}
