package ui

import (
	"github.com/charmbracelet/bubbles/table"
)

// executeQuery runs the current query and updates the table.
func (m *Model) executeQuery() {
	query := m.textInput.Value()
	if query == "" {
		return
	}

	rows, err := m.db.Query(query)
	if err != nil {
		m.err = err
		return
	}
	defer func() { _ = rows.Close() }()

	cols, tableRows, err := SQLRowsToTable(rows)
	if err != nil {
		m.err = err
		return
	}

	// Reset table state before updating
	m.allColumns = cols
	m.allRows = tableRows
	m.colCursor = 0
	m.colOffset = 0
	m.table.SetCursor(0)
	m.updateVisibleColumns()
	m.err = nil
}

// updateVisibleColumns updates the table with visible columns based on scroll offset.
func (m *Model) updateVisibleColumns() {
	if len(m.allColumns) == 0 {
		return
	}

	visibleCount := m.visibleColumnCount()
	endIdx := m.colOffset + visibleCount
	if endIdx > len(m.allColumns) {
		endIdx = len(m.allColumns)
	}

	// Calculate column width to fill available space
	colWidth := m.calculateColumnWidth(endIdx - m.colOffset)

	// Get visible columns from offset with dynamic width
	visibleCols := make([]table.Column, endIdx-m.colOffset)
	for i, col := range m.allColumns[m.colOffset:endIdx] {
		visibleCols[i] = table.Column{Title: col.Title, Width: colWidth}
	}

	// Build visible rows with cell marker
	visibleRows := m.buildVisibleRows(endIdx)

	// Save cursor position before updating
	cursor := m.table.Cursor()

	// Clear rows first to avoid column/row mismatch during update
	m.table.SetRows([]table.Row{})
	m.table.SetColumns(visibleCols)
	m.table.SetRows(visibleRows)
	m.table.SetCursor(cursor)
}

// updateCellMarker updates only the cell marker without resetting columns.
// This is used when the row cursor changes to avoid viewport reset.
func (m *Model) updateCellMarker() {
	if len(m.allColumns) == 0 || len(m.allRows) == 0 {
		return
	}

	visibleCount := m.visibleColumnCount()
	endIdx := m.colOffset + visibleCount
	if endIdx > len(m.allColumns) {
		endIdx = len(m.allColumns)
	}

	// Build and set visible rows - preserves viewport scroll position
	m.table.SetRows(m.buildVisibleRows(endIdx))
}

// buildVisibleRows builds table rows with the cell marker applied to the selected cell.
func (m *Model) buildVisibleRows(endIdx int) []table.Row {
	selectedRow := m.table.Cursor()
	selectedColInView := m.colCursor - m.colOffset

	visibleRows := make([]table.Row, len(m.allRows))
	for i, row := range m.allRows {
		if m.colOffset < len(row) {
			end := endIdx
			if end > len(row) {
				end = len(row)
			}
			visibleRow := make(table.Row, end-m.colOffset)
			copy(visibleRow, row[m.colOffset:end])

			// Add accent colored marker to selected cell
			if i == selectedRow && selectedColInView >= 0 && selectedColInView < len(visibleRow) {
				visibleRow[selectedColInView] = styleTextAccent.Render(cellMarker) + " " + visibleRow[selectedColInView]
			}

			visibleRows[i] = visibleRow
		} else {
			visibleRows[i] = table.Row{}
		}
	}
	return visibleRows
}

// visibleColumnCount returns the number of columns that can fit in the view.
// Limited to maxVisibleCols for better readability.
func (m *Model) visibleColumnCount() int {
	if m.width == 0 {
		return maxVisibleCols
	}
	count := (m.width - framePadding) / (defaultColumnWidth + columnBorderWidth)
	if count < 1 {
		return 1
	}
	if count > maxVisibleCols {
		return maxVisibleCols
	}
	return count
}

// calculateColumnWidth returns the optimal column width based on terminal width.
func (m *Model) calculateColumnWidth(numCols int) int {
	if m.width == 0 || numCols == 0 {
		return defaultColumnWidth
	}
	available := m.width - framePadding - (numCols * columnBorderWidth)
	width := available / numCols
	if width < minColumnWidth {
		return minColumnWidth
	}
	if width > maxColumnWidth {
		return maxColumnWidth
	}
	return width
}
