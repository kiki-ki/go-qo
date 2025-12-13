package ui

import (
	"database/sql"

	"github.com/charmbracelet/bubbles/table"

	"github.com/kiki-ki/go-qo/internal/output"
)

// TableState manages table data and cell selection state.
type TableState struct {
	columns   []table.Column
	rows      []table.Row
	colCursor int // selected column index
	colOffset int // column scroll offset for horizontal scrolling
}

// NewTableState creates a new empty table state.
func NewTableState() *TableState {
	return &TableState{}
}

// SetData updates the table data and resets cursor positions.
func (s *TableState) SetData(cols []table.Column, rows []table.Row) {
	s.columns = cols
	s.rows = rows
	s.colCursor = 0
	s.colOffset = 0
}

// Columns returns all columns.
func (s *TableState) Columns() []table.Column {
	return s.columns
}

// Rows returns all rows.
func (s *TableState) Rows() []table.Row {
	return s.rows
}

// ColCursor returns the current column cursor position.
func (s *TableState) ColCursor() int {
	return s.colCursor
}

// MoveLeft moves the column cursor left if possible.
// Returns true if the cursor moved.
func (s *TableState) MoveLeft() bool {
	if s.colCursor > 0 {
		s.colCursor--
		return true
	}
	return false
}

// MoveRight moves the column cursor right if possible.
// Returns true if the cursor moved.
func (s *TableState) MoveRight() bool {
	if s.colCursor < len(s.columns)-1 {
		s.colCursor++
		return true
	}
	return false
}

// AdjustOffset adjusts the column offset to keep cursor visible.
// Returns true if the offset changed.
func (s *TableState) AdjustOffset(visibleCols int) bool {
	prevOffset := s.colOffset
	if s.colCursor < s.colOffset {
		s.colOffset = s.colCursor
	} else if s.colCursor >= s.colOffset+visibleCols {
		s.colOffset = s.colCursor - visibleCols + 1
	}
	return s.colOffset != prevOffset
}

// VisibleColumnRange returns the start and end indices for visible columns.
func (s *TableState) VisibleColumnRange(visibleCols int) (start, end int) {
	start = s.colOffset
	end = s.colOffset + visibleCols
	if end > len(s.columns) {
		end = len(s.columns)
	}
	return start, end
}

// BuildVisibleColumns builds the visible columns with the given width.
func (s *TableState) BuildVisibleColumns(visibleCols, colWidth int) []table.Column {
	start, end := s.VisibleColumnRange(visibleCols)
	numCols := end - start

	cols := make([]table.Column, numCols)
	for i, col := range s.columns[start:end] {
		cols[i] = table.Column{Title: col.Title, Width: colWidth}
	}
	return cols
}

// BuildVisibleRows builds visible rows with cell marker on selected cell.
func (s *TableState) BuildVisibleRows(selectedRow, visibleCols int) []table.Row {
	start, end := s.VisibleColumnRange(visibleCols)
	selectedColInView := s.colCursor - s.colOffset

	visibleRows := make([]table.Row, len(s.rows))
	for i, row := range s.rows {
		if start >= len(row) {
			visibleRows[i] = table.Row{}
			continue
		}

		rowEnd := end
		if rowEnd > len(row) {
			rowEnd = len(row)
		}
		visibleRow := make(table.Row, rowEnd-start)
		copy(visibleRow, row[start:rowEnd])

		// Add accent colored marker to selected cell
		if i == selectedRow && selectedColInView >= 0 && selectedColInView < len(visibleRow) {
			visibleRow[selectedColInView] = styleTextAccent.Render(cellMarker) + " " + visibleRow[selectedColInView]
		}

		visibleRows[i] = visibleRow
	}
	return visibleRows
}

// SelectedCell returns the column name and value of the selected cell.
func (s *TableState) SelectedCell(rowIdx int) (colName, value string, ok bool) {
	if len(s.rows) == 0 || len(s.columns) == 0 {
		return "", "", false
	}
	if rowIdx < 0 || rowIdx >= len(s.rows) {
		return "", "", false
	}
	if s.colCursor >= len(s.rows[rowIdx]) {
		return "", "", false
	}
	return s.columns[s.colCursor].Title, s.rows[rowIdx][s.colCursor], true
}

// Position returns the current position info (row, totalRows, col, totalCols).
func (s *TableState) Position(rowIdx int) (row, totalRows, col, totalCols int) {
	return rowIdx + 1, len(s.rows), s.colCursor + 1, len(s.columns)
}

// SQLRowsToTableData converts SQL rows to table columns and rows.
func SQLRowsToTableData(rows *sql.Rows) ([]table.Column, []table.Row, error) {
	cols, err := rows.Columns()
	if err != nil {
		return nil, nil, err
	}

	tCols := make([]table.Column, len(cols))
	for i, c := range cols {
		tCols[i] = table.Column{Title: c, Width: defaultColumnWidth}
	}

	var tRows []table.Row

	values := make([]interface{}, len(cols))
	valuePtrs := make([]interface{}, len(cols))
	for i := range cols {
		valuePtrs[i] = &values[i]
	}

	for rows.Next() {
		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, nil, err
		}

		rowData := make(table.Row, len(cols))
		for i, val := range values {
			rowData[i] = output.FormatValueRaw(val)
		}
		tRows = append(tRows, rowData)
	}

	return tCols, tRows, nil
}
