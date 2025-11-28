package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// View renders the UI display.
func (m Model) View() string {
	var b strings.Builder

	parts := []string{
		m.renderHeader(),
		m.textInput.View(),
		m.renderError(),
		"",
		m.table.View(),
	}

	// Add cell detail view in table mode
	if m.mode == ModeTable {
		parts = append(parts, m.renderCellDetail())
	}

	// Add table list in query mode
	if m.mode == ModeQuery {
		parts = append(parts, m.renderTableList())
	}

	content := lipgloss.JoinVertical(lipgloss.Left, parts...)

	b.WriteString("\n") // ensure top border is visible in alt screen
	b.WriteString(baseStyle.Render(content))
	b.WriteString("\n")

	return b.String()
}

// renderHeader builds the header line with mode and hints.
func (m Model) renderHeader() string {
	header := fmt.Sprintf(" [%s] %s", modeStyle.Render(string(m.mode)), m.mode.CommandsHint())
	return headerStyle.Render(header)
}

// renderCellDetail returns the full content of the selected cell with position info.
func (m Model) renderCellDetail() string {
	if len(m.allRows) == 0 || len(m.allColumns) == 0 {
		return headerStyle.Render("\n (no data)")
	}

	rowIdx := m.table.Cursor()
	if rowIdx < 0 {
		rowIdx = 0 // default to first row
	}
	if rowIdx >= len(m.allRows) {
		return ""
	}

	row := m.allRows[rowIdx]
	if m.colCursor >= len(row) {
		return ""
	}

	colName := m.allColumns[m.colCursor].Title
	value := row[m.colCursor]
	pos := fmt.Sprintf("(%d/%d, %d/%d)", rowIdx+1, len(m.allRows), m.colCursor+1, len(m.allColumns))

	return fmt.Sprintf("\n %s %s: %s", headerStyle.Render(pos), modeStyle.Render(colName), value)
}

// renderError returns the error view. Always returns a line to prevent layout shift.
func (m Model) renderError() string {
	if m.err == nil {
		return "\n" // empty line to maintain consistent height
	}
	return errorStyle.Render(fmt.Sprintf("\nError: %v", m.err))
}

// renderTableList returns the list of available tables.
func (m Model) renderTableList() string {
	if len(m.tableNames) == 0 {
		return ""
	}
	return headerStyle.Render(fmt.Sprintf("\n Tables: %s", strings.Join(m.tableNames, ", ")))
}
