package ui

import "github.com/charmbracelet/bubbles/table"

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

	// Build visible rows with matching columns
	visibleRows := make([]table.Row, len(m.allRows))
	for i, row := range m.allRows {
		if m.colOffset < len(row) {
			end := endIdx
			if end > len(row) {
				end = len(row)
			}
			visibleRows[i] = row[m.colOffset:end]
		} else {
			visibleRows[i] = table.Row{}
		}
	}

	m.table.SetRows([]table.Row{})
	m.table.SetColumns(visibleCols)
	m.table.SetRows(visibleRows)
}

// visibleColumnCount returns the number of columns that can fit in the view.
func (m *Model) visibleColumnCount() int {
	if m.width == 0 {
		return defaultVisibleCols
	}
	count := (m.width - framePadding) / (defaultColumnWidth + columnBorderWidth)
	if count < 1 {
		return 1
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
