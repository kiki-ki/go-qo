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

	if m.mode == ModeTable {
		parts = append(parts, m.renderCellDetail())
	}
	if m.mode == ModeQuery {
		parts = append(parts, m.renderTableList())
	}

	content := lipgloss.JoinVertical(lipgloss.Left, parts...)

	b.WriteString("\n")
	frame := lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(colorBase).
		Padding(0, 1)

	b.WriteString(frame.Render(content))
	b.WriteString("\n")

	return b.String()
}

// renderHeader builds the header line with mode and hints.
func (m Model) renderHeader() string {
	return fmt.Sprintf(
		" [%s] %s",
		styleTextAccent.Render(string(m.mode)),
		styleTextMuted.Render(m.mode.CommandsHint()),
	)
}

// renderCellDetail returns the full content of the selected cell with position info.
func (m Model) renderCellDetail() string {
	if len(m.allRows) == 0 || len(m.allColumns) == 0 {
		return styleTextMuted.Render("\n (no data)")
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

	return styleTextBase.Render(fmt.Sprintf("\n %s %s: %s", pos, colName, value))
}

// renderError returns the error view. Always returns a line to prevent layout shift.
func (m Model) renderError() string {
	if m.err == nil {
		return "\n" // empty line to maintain consistent height
	}
	return styleTextError.Render(fmt.Sprintf("\nError: %v", m.err))
}

// renderTableList returns the list of available tables.
func (m Model) renderTableList() string {
	if len(m.tableNames) == 0 {
		return ""
	}
	return styleTextBase.Render(fmt.Sprintf("\n Tables: %s", strings.Join(m.tableNames, ", ")))
}
