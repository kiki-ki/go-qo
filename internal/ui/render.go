package ui

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"

	"github.com/kiki-ki/go-qo/internal/output"
)

// View renders the UI display.
func (m Model) View() string {

	if m.columnPickerActive {
		return m.renderColumnPickerModal()
	}

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

// renderError returns the error view. Always returns a line to prevent layout shift.
func (m Model) renderError() string {
	if m.err == nil {
		return "\n"
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

// renderColumnPickerModal renders the column picker with fzf-like styling.
func (m Model) renderColumnPickerModal() string {
	modalWidth := m.width - 10
	modalHeight := m.height - 10

	if modalWidth < 40 {
		modalWidth = 40
	}
	if modalHeight < 10 {
		modalHeight = 10
	}

	m.columnPickerState.list.SetWidth(modalWidth - 4)
	m.columnPickerState.list.SetHeight(modalHeight - 8)

	// Detect if we are filtering
	isFiltering := m.columnPickerState.list.FilterState() == list.Filtering
	filterValue := m.columnPickerState.list.FilterInput.Value()

	// Get all items and filtered items
	allItems := m.columnPickerState.list.Items()
	filteredItems := m.columnPickerState.list.Items() // Already filtered by the list component

	// Count selected columns
	selectedCount := 0
	for _, item := range allItems {
		if colItem, ok := item.(ColumnItem); ok && colItem.selected {
			selectedCount++
		}
	}

	// Get total columns count
	totalCols := len(m.tableColumns)
	if totalCols == 0 {
		totalCols = len(allItems)
	}

	// Build search bar with better prompt
	var searchBar string
	searchPromptSymbol := styleTextAccent.Render("> ")

	if isFiltering {
		// Show filter value with blinking cursor
		value := styleTextBase.Render(filterValue)
		cursor := m.columnPickerState.list.FilterInput.Cursor.View()
		searchBar = searchPromptSymbol + value + cursor
	} else {
		searchBar = searchPromptSymbol + styleTextMuted.Render("Type '/' to search, or start typing...")
	}

	// Build header with selection count
	itemCount := fmt.Sprintf("%d/%d", len(filteredItems), totalCols)
	selectionCount := ""
	if selectedCount > 0 {
		selectionCount = styleTextAccent.Render(fmt.Sprintf(" • %d selected ✓", selectedCount))
	}

	headerRight := lipgloss.JoinHorizontal(
		lipgloss.Left,
		styleTextMuted.Render(itemCount),
		selectionCount,
	)

	// Build header
	header := lipgloss.JoinHorizontal(
		lipgloss.Left,
		searchBar,
		strings.Repeat(" ", max(0, modalWidth-lipgloss.Width(searchBar)-lipgloss.Width(headerRight)-6)),
		headerRight,
	)

	// Get list view
	listView := m.columnPickerState.list.View()

	// Check if no results
	noResultsMsg := ""
	if isFiltering && filterValue != "" && len(filteredItems) == 0 {
		noResultsMsg = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "240", Dark: "238"}).
			Align(lipgloss.Center).
			Padding(2, 0).
			Render(fmt.Sprintf("No columns match '%s'", filterValue))
	}

	// Combine header and list
	contentParts := []string{header, ""}
	if noResultsMsg != "" {
		contentParts = append(contentParts, noResultsMsg)
	} else {
		contentParts = append(contentParts, listView)
	}

	content := lipgloss.JoinVertical(lipgloss.Left, contentParts...)

	// Create modal with improved border
	borderColor := lipgloss.AdaptiveColor{Light: "240", Dark: "238"}
	if isFiltering {
		// Change border color when filtering
		borderColor = lipgloss.AdaptiveColor{Light: "13", Dark: "13"} // Pink accent when filtering
	}

	modalStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(1, 2).
		Width(modalWidth).
		Height(modalHeight)

	rendered := modalStyle.Render(content)

	// Footer with dynamic hints (Mejora #8)
	var footerLines []string

	if isFiltering {
		// When filtering, show filter-specific hints
		footerLines = []string{
			styleTextAccent.Render("🔍 Filtering") + " • " +
				styleTextBase.Render("Enter") + " " + styleTextMuted.Render("toggle") + " • " +
				styleTextBase.Render("Esc") + " " + styleTextMuted.Render("clear filter") + " • " +
				styleTextBase.Render("↑↓/jk") + " " + styleTextMuted.Render("navigate"),
		}
	} else {
		// Default hints
		keyHint := "Space"
		actionHint := "toggle"
		if selectedCount > 0 {
			actionHint = fmt.Sprintf("toggle • Enter to apply (%d)", selectedCount)
		}

		footerLines = []string{
			styleTextMuted.Render("Type '/' to search") + " • " +
				styleTextBase.Render("↑↓/jk") + " " + styleTextMuted.Render("navigate") + " • " +
				styleTextBase.Render(keyHint) + " " + styleTextMuted.Render(actionHint) + " • " +
				styleTextBase.Render("Esc") + " " + styleTextMuted.Render("cancel"),
		}
	}

	footer := strings.Join(footerLines, "\n")
	footerStyle := lipgloss.NewStyle().
		Width(modalWidth).
		Align(lipgloss.Center).
		PaddingTop(1)

	return lipgloss.JoinVertical(
		lipgloss.Center,
		rendered,
		footerStyle.Render(footer),
	)
}
