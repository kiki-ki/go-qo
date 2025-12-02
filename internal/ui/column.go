package ui

import (
	"github.com/charmbracelet/bubbles/table"
)

// visibleColumnEndIndex returns the end index for visible columns.
func (m *Model) visibleColumnEndIndex() int {
	endIdx := m.colOffset + m.visibleColumnCount()
	if endIdx > len(m.allColumns) {
		return len(m.allColumns)
	}
	return endIdx
}

// updateVisibleColumns rebuilds the table with visible columns based on scroll offset.
// This resets the viewport, so cursor position must be restored afterwards.
func (m *Model) updateVisibleColumns() {
	if len(m.allColumns) == 0 {
		return
	}

	endIdx := m.visibleColumnEndIndex()
	numCols := endIdx - m.colOffset
	colWidth := m.calculateColumnWidth(numCols)

	visibleCols := make([]table.Column, numCols)
	for i, col := range m.allColumns[m.colOffset:endIdx] {
		visibleCols[i] = table.Column{Title: col.Title, Width: colWidth}
	}

	// Save cursor before SetRows resets it
	cursor := m.table.Cursor()
	visibleRows := m.buildVisibleRows(cursor)

	m.table.SetRows([]table.Row{})
	m.table.SetColumns(visibleCols)
	m.table.SetRows(visibleRows)

	// Restore cursor by moving from top (SetCursor alone doesn't update viewport)
	m.table.GotoTop()
	for i := 0; i < cursor; i++ {
		m.table.MoveDown(1)
	}
}

// updateCellMarker updates only the cell marker without rebuilding columns.
// This preserves the viewport scroll position.
func (m *Model) updateCellMarker() {
	if len(m.allColumns) == 0 || len(m.allRows) == 0 {
		return
	}
	m.table.SetRows(m.buildVisibleRows(m.table.Cursor()))
}

// buildVisibleRows builds table rows with the cell marker applied to the selected cell.
func (m *Model) buildVisibleRows(selectedRow int) []table.Row {
	endIdx := m.visibleColumnEndIndex()
	selectedColInView := m.colCursor - m.colOffset

	visibleRows := make([]table.Row, len(m.allRows))
	for i, row := range m.allRows {
		if m.colOffset >= len(row) {
			visibleRows[i] = table.Row{}
			continue
		}

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
