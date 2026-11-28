package ui

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/charmbracelet/lipgloss"

	"github.com/kiki-ki/go-qo/internal/output"
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
	rowIdx := m.table.Cursor()
	if rowIdx < 0 {
		rowIdx = 0
	}

	colName, value, ok := m.tableState.SelectedCell(rowIdx)
	if !ok {
		return styleTextBase.Render("\n (no data)")
	}

	row, totalRows, col, totalCols := m.tableState.Position(rowIdx)
	pos := fmt.Sprintf("(%d/%d, %d/%d)", row, totalRows, col, totalCols)

	// Calculate available width for value
	prefix := fmt.Sprintf(" %s %s: ", pos, colName)
	prefixWidth := utf8.RuneCountInString(prefix)
	availableWidth := m.width - prefixWidth - cellDetailPadding
	if availableWidth < cellDetailMinWidth {
		availableWidth = cellDetailMinWidth
	}
	truncatedValue := output.Truncate(value, availableWidth)

	return styleTextBase.Render(fmt.Sprintf("\n%s%s", prefix, truncatedValue))
}

// renderError returns the error view, or the load warnings while there is no
// error. Always returns one line to prevent layout shift.
func (m Model) renderError() string {
	if m.err != nil {
		return styleTextError.Render(fmt.Sprintf("\nError: %v", m.err))
	}
	if len(m.warnings) == 0 {
		return "\n"
	}
	// Stays on screen because a query for a renamed column's original name
	// silently returns the column that kept that name.
	warning := "Warning: " + strings.Join(m.warnings, "; ")
	if width := m.width - cellDetailPadding; width > 0 {
		warning = output.Truncate(warning, width)
	}
	return styleTextError.Render("\n" + warning)
}

// renderTableList returns the list of available tables.
func (m Model) renderTableList() string {
	if len(m.tableNames) == 0 {
		return ""
	}
	return styleTextBase.Render(fmt.Sprintf("\n Tables: %s", strings.Join(m.tableNames, ", ")))
}
