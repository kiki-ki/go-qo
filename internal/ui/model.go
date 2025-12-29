package ui

import (
	"database/sql"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Result holds the final query when exiting UI.
type Result struct {
	Query string
}

// Model represents the UI application state.
type Model struct {
	db         *sql.DB
	mode       Mode
	table      table.Model
	textInput  textinput.Model
	err        error
	width      int
	height     int
	tableNames []string

	// Table data state
	tableState *TableState

	// Debounce state
	pendingQuery  string
	lastExecQuery string

	// Result when exiting
	result *Result

	// Original columns from the table (For column picker)
	tableColumns []table.Column

	columnPickerActive bool
	columnPickerState  *ColumnPickerState
}

// ColumnItem represents a column in the picker list.
type ColumnItem struct {
	name     string
	selected bool
}

// ColumnItemDelegate handles rendering of column items.
type ColumnItemDelegate struct{}

// Result returns the final query result.
func (m Model) Result() *Result {
	return m.result
}

// PendingQuery returns the pending query for testing.
func (m Model) PendingQuery() string {
	return m.pendingQuery
}

// NewModel creates a new UI model.
func NewModel(db *sql.DB, tableNames []string) Model {
	ti := newTextInput(tableNames)
	t := newTable()

	return Model{
		db:           db,
		mode:         ModeQuery,
		table:        t,
		textInput:    ti,
		tableState:   NewTableState(),
		tableNames:   tableNames,
		pendingQuery: ti.Value(),
	}
}

// newTextInput creates a configured text input component.
func newTextInput(tableNames []string) textinput.Model {
	ti := textinput.New()
	ti.Placeholder = "SQL Query..."
	ti.Focus()
	ti.CharLimit = inputCharLimit
	ti.Width = inputInitialWidth

	ti.TextStyle = styleTextBase
	ti.PlaceholderStyle = styleTextMuted
	ti.PromptStyle = styleTextAccent
	ti.Cursor.Style = styleTextAccent

	if len(tableNames) > 0 {
		ti.SetValue(fmt.Sprintf("SELECT * FROM %s LIMIT %d", tableNames[0], defaultQueryLimit))
	}

	return ti
}

// newTable creates a configured table component.
func newTable() table.Model {
	t := table.New(
		table.WithColumns([]table.Column{{Title: "Results", Width: initialColumnWidth}}),
		table.WithRows([]table.Row{}),
		table.WithFocused(false),
		table.WithHeight(initialTableHeight),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(colorBase).
		BorderBottom(true).
		Bold(false)
	s.Selected = lipgloss.NewStyle()
	t.SetStyles(s)

	return t
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		textinput.Blink,
		func() tea.Msg {
			return NewDebounceMsg(m.textInput.Value())
		},
	)
}

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	if m.columnPickerActive {
		return m.handleColumnPicker(msg)
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.handleWindowResize(msg)
	case debounceMsg:
		m.handleDebounceMsg(msg)
	case tea.KeyMsg:
		if cmd, quit := m.handleKeyMsg(msg); quit {
			return m, tea.Quit
		} else if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	cmds = append(cmds, m.updateComponents(msg)...)

	return m, tea.Batch(cmds...)
}

// updateComponents updates sub-components and handles state changes.
func (m *Model) updateComponents(msg tea.Msg) []tea.Cmd {
	var cmds []tea.Cmd

	prevQuery := m.textInput.Value()
	prevCursor := m.table.Cursor()

	m.table, _ = m.table.Update(msg)

	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	cmds = append(cmds, cmd)

	// Schedule debounced query execution when input changes in query mode
	if m.mode == ModeQuery && m.textInput.Value() != prevQuery {
		m.pendingQuery = m.textInput.Value()
		cmds = append(cmds, m.scheduleQueryExecution())
	}

	// Update cell marker when row cursor changes (lightweight, preserves viewport)
	if m.mode == ModeTable && m.table.Cursor() != prevCursor {
		m.syncTableView(false)
	}

	return cmds
}

// handleWindowResize updates dimensions on window resize.
func (m *Model) handleWindowResize(msg tea.WindowSizeMsg) {
	m.width = msg.Width
	m.height = msg.Height
	m.table.SetHeight(msg.Height - tableHeightOffset)
	m.textInput.Width = msg.Width - inputWidthOffset
	m.syncTableView(true)
}

// handleKeyMsg processes key events and returns a command and quit flag.
func (m *Model) handleKeyMsg(msg tea.KeyMsg) (tea.Cmd, bool) {

	// If columnPicker is active, don't handle keys
	if m.columnPickerActive {
		return nil, false
	}

	switch msg.Type {
	case tea.KeyCtrlC, tea.KeyEsc:
		return nil, true

	case tea.KeyEnter:
		if m.mode == ModeQuery && m.textInput.Value() != "" {
			m.result = &Result{Query: m.textInput.Value()}
			return nil, true
		}

	case tea.KeyTab:
		return m.toggleMode(), false

	case tea.KeyCtrlP:
		// Open column picker in query mode
		if m.mode == ModeQuery {
			// Load table columns if not already loaded
			if len(m.tableColumns) == 0 && len(m.tableNames) > 0 {
				tableName := m.tableNames[0]
				tableCols, err := m.getTableColumns(tableName)
				if err == nil {
					m.tableColumns = tableCols
				}
			}

			// Use tableColumns if available, otherwise fall back to allColumns
			colsToUse := m.tableColumns
			if len(colsToUse) == 0 {
				colsToUse = m.tableState.columns
			}

			if len(colsToUse) > 0 {
				// Extract currently selected columns from the query
				currentQuery := m.textInput.Value()
				selectedCols := extractColumnsFromSelect(currentQuery)

				m.columnPickerState = newColumnPicker(colsToUse, selectedCols)
				m.columnPickerActive = true

				// Set picker dimensions
				if m.width > 0 && m.height > 0 {
					m.columnPickerState.list.SetWidth(m.width - 10)
					m.columnPickerState.list.SetHeight(m.height - 10)
				}

				return nil, false
			}
		}

	case tea.KeyLeft, tea.KeyRight, tea.KeyRunes:
		if m.mode == ModeTable {
			m.handleTableNavigation(msg)
		}
	}
	return nil, false
}

// toggleMode switches between Query and Table modes.
func (m *Model) toggleMode() tea.Cmd {
	if m.mode == ModeQuery {
		m.mode = ModeTable
		m.textInput.Blur()
		m.table.Focus()
		return nil
	}
	m.mode = ModeQuery
	m.table.Blur()
	m.textInput.Focus()
	return textinput.Blink
}

// handleTableNavigation handles column cursor movement in table mode.
func (m *Model) handleTableNavigation(msg tea.KeyMsg) {
	moveLeft := msg.Type == tea.KeyLeft || msg.String() == "h"
	moveRight := msg.Type == tea.KeyRight || msg.String() == "l"

	var moved bool
	if moveLeft {
		moved = m.tableState.MoveLeft()
	} else if moveRight {
		moved = m.tableState.MoveRight()
	}

	if !moved {
		return
	}

	// Adjust offset and sync view
	offsetChanged := m.tableState.AdjustOffset(m.visibleColumnCount())
	m.syncTableView(offsetChanged)
}

// syncTableView updates the bubbles table with current state.
// If rebuildColumns is true, rebuilds both columns and rows (heavier).
// If false, only updates rows with cell marker (lighter).
func (m *Model) syncTableView(rebuildColumns bool) {
	if len(m.tableState.Columns()) == 0 {
		return
	}

	visibleCols := m.visibleColumnCount()
	start, end := m.tableState.VisibleColumnRange(visibleCols)
	actualCols := end - start // actual number of columns to display
	colWidth := m.calculateColumnWidth(actualCols)
	cursor := m.table.Cursor()

	if rebuildColumns {
		cols := m.tableState.BuildVisibleColumns(visibleCols, colWidth)
		m.table.SetRows([]table.Row{})
		m.table.SetColumns(cols)
	}

	rows := m.tableState.BuildVisibleRows(cursor, visibleCols)
	m.table.SetRows(rows)

	if rebuildColumns {
		// Restore cursor by moving from top (SetCursor alone doesn't update viewport)
		m.table.GotoTop()
		for i := 0; i < cursor; i++ {
			m.table.MoveDown(1)
		}
	}
}

// visibleColumnCount returns the number of columns that can fit in the view.
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

// Run starts the UI application and returns the final query if any.
func Run(db *sql.DB, tableNames []string) (*Result, error) {
	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to open /dev/tty: %w", err)
	}
	defer func() { _ = tty.Close() }()

	// Set lipgloss to use tty for color detection and reinitialize styles
	lipgloss.SetDefaultRenderer(lipgloss.NewRenderer(tty))
	initStyles()

	p := tea.NewProgram(
		NewModel(db, tableNames),
		tea.WithAltScreen(),
		tea.WithInput(tty),
		tea.WithOutput(tty),
	)

	m, err := p.Run()
	if err != nil {
		return nil, err
	}

	if model, ok := m.(Model); ok {
		return model.result, nil
	}

	return nil, nil
}

func (ci ColumnItem) FilterValue() string {
	return ci.name
}

// Returns the height of a column item.
func (cid ColumnItemDelegate) Height() int {
	return 1
}

// Spacing returns spacing between items.
func (cid ColumnItemDelegate) Spacing() int {
	return 0
}

// Update handles item updates (not used for simple items).
func (cid ColumnItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd {
	return nil
}

// Render renders a column item.
func (cid ColumnItemDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	colItem, ok := item.(ColumnItem)
	if !ok {
		return
	}

	// Check if this item is selected in the list
	isSelected := index == m.Index()

	// Show the column name
	checkbox := "[ ]"
	if colItem.selected {
		checkbox = "[x]"
	}

	name := colItem.name
	width := m.Width()
	if width > 0 && len(name) > width-5 {
		name = name[:width-5] + "..."
	}

	style := styleTextBase
	if isSelected {
		style = styleTextAccent
	}

	fmt.Fprintf(w, "%s %s", checkbox, style.Render(name))
}

// newColumnPicker creates a new list model for db column selection.
func newColumnPicker(columns []table.Column, selectedColumns []string) *ColumnPickerState {
	return newColumnPickerState(columns, selectedColumns)
}

// handleColumnPicker processes messages when column picker is active.
func (m Model) handleColumnPicker(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	// Check if we're filtering
	isFiltering := m.columnPickerState.list.FilterState() == list.Filtering
	prevFilterValue := m.columnPickerState.list.FilterInput.Value()

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.columnPickerState.list.SetWidth(msg.Width - 10)
		m.columnPickerState.list.SetHeight(msg.Height - 10)
		return m, nil

	case tea.KeyMsg:
		if isFiltering {
			switch msg.Type {
			case tea.KeyEsc:
				m.columnPickerState.list.ResetFilter()
				return m, nil

			case tea.KeyEnter:
				selectedItem := m.columnPickerState.list.SelectedItem()
				if selectedItem != nil {
					colItem := selectedItem.(ColumnItem)
					colItem.selected = !colItem.selected

					items := m.columnPickerState.list.Items()
					for i, item := range items {
						if it, ok := item.(ColumnItem); ok && it.name == colItem.name {
							cmd := m.columnPickerState.list.SetItem(i, colItem)
							return m, cmd
						}
					}
				}
				return m, nil

			case tea.KeyCtrlN, tea.KeyDown:
				m.columnPickerState.list.CursorDown()
				return m, nil

			case tea.KeyCtrlP, tea.KeyUp:
				m.columnPickerState.list.CursorUp()
				return m, nil
			}

			var cmd tea.Cmd
			m.columnPickerState.list, cmd = m.columnPickerState.list.Update(msg)
			
			// Auto-scroll to first result when filter changes
			newFilterValue := m.columnPickerState.list.FilterInput.Value()
			if newFilterValue != prevFilterValue && len(m.columnPickerState.list.Items()) > 0 {
				// Move cursor to first item
				m.columnPickerState.list.Select(0)
			}
			
			return m, cmd
		}

		switch msg.Type {
		case tea.KeyEsc:
			m.columnPickerActive = false
			return m, nil

		case tea.KeyEnter:
			selectedColumns := m.getSelectedColumns()
			if len(selectedColumns) > 0 {
				m.insertColumnsIntoQuery(selectedColumns)
				m.pendingQuery = m.textInput.Value()
				m.columnPickerActive = false
				return m, m.scheduleQueryExecution()
			}
			m.columnPickerActive = false
			return m, nil

		case tea.KeySpace:
			selectedItem := m.columnPickerState.list.SelectedItem()
			if selectedItem != nil {
				colItem := selectedItem.(ColumnItem)
				colItem.selected = !colItem.selected

				items := m.columnPickerState.list.Items()
				index := m.columnPickerState.list.Index()
				if index >= 0 && index < len(items) {
					items[index] = colItem
					m.columnPickerState.list.SetItems(items)
				}
			}
			return m, nil

		// Handle j/k, and / for navigation
		case tea.KeyRunes:
			switch msg.String() {
			case "j":
				m.columnPickerState.list.CursorDown()
				return m, nil

			case "k":
				m.columnPickerState.list.CursorUp()
				return m, nil

			case "/":
				var cmd tea.Cmd
				m.columnPickerState.list, cmd = m.columnPickerState.list.Update(msg)
				return m, tea.Batch(cmd, textinput.Blink)
			}
		}
	}

	// Update the list (this handles all other keys including filtering)
	var cmd tea.Cmd
	m.columnPickerState.list, cmd = m.columnPickerState.list.Update(msg)
	
	// Auto-scroll to first result when filter changes
	newFilterValue := m.columnPickerState.list.FilterInput.Value()
	if newFilterValue != prevFilterValue && len(m.columnPickerState.list.Items()) > 0 {
		m.columnPickerState.list.Select(0)
	}
	
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// insertColumnsIntoQuery replaces the SELECT clause with the selected columns.
func (m *Model) insertColumnsIntoQuery(columnNames []string) {
	if len(columnNames) == 0 {
		return
	}

	val := m.textInput.Value()
	cursor := m.textInput.Position()

	// Find SELECT clause position
	selectIdx := findSelectPosition(val, cursor)
	if selectIdx == -1 {
		// No SELECT found, just insert at cursor
		columnsStr := strings.Join(columnNames, ", ")
		before := val[:cursor]
		after := val[cursor:]
		newVal := before + columnsStr + after
		m.textInput.SetValue(newVal)
		m.textInput.SetCursor(cursor + len(columnsStr))
		return
	}

	// Find the position after SELECT keyword
	afterSelect := selectIdx + 6 // "SELECT" is 6 chars
	for afterSelect < len(val) && val[afterSelect] == ' ' {
		afterSelect++
	}

	// Find where the SELECT clause ends (before FROM/WHERE/etc)
	queryLower := strings.ToLower(val)
	keywords := []string{" from ", " where ", " group ", " order ", " having ", " limit "}

	selectEnd := len(val)
	for _, kw := range keywords {
		idx := indexOf(queryLower, kw, afterSelect)
		if idx != -1 && idx < selectEnd {
			selectEnd = idx
		}
	}

	// Build new SELECT clause
	columnsStr := strings.Join(columnNames, ", ")

	// Replace the SELECT clause
	before := val[:afterSelect]
	after := val[selectEnd:]
	newVal := before + columnsStr + after

	m.textInput.SetValue(newVal)
	// Set cursor after the inserted columns
	newCursor := afterSelect + len(columnsStr)
	if newCursor > len(newVal) {
		newCursor = len(newVal)
	}
	m.textInput.SetCursor(newCursor)
}

// getSelectedColumns returns all selected column names.
func (m *Model) getSelectedColumns() []string {
	var selected []string
	items := m.columnPickerState.list.Items()
	for _, item := range items {
		if colItem, ok := item.(ColumnItem); ok && colItem.selected {
			selected = append(selected, colItem.name)
		}
	}
	return selected
}

// insertColumnIntoQuery inserts a column name at the appropriate position in SELECT.
func (m *Model) insertColumnIntoQuery(columnName string) {
	val := m.textInput.Value()
	cursor := m.textInput.Position()

	// Find SELECT clause position
	selectIdx := findSelectPosition(val, cursor)
	if selectIdx == -1 {
		// No SELECT found, just insert at cursor
		before := val[:cursor]
		after := val[cursor:]
		newVal := before + columnName + after
		m.textInput.SetValue(newVal)
		m.textInput.SetCursor(cursor + len(columnName))
		return
	}

	// Find the position after SELECT keyword
	afterSelect := selectIdx + 6 // "SELECT" is 6 chars
	for afterSelect < len(val) && val[afterSelect] == ' ' {
		afterSelect++
	}

	// Check if we're replacing * or inserting
	if afterSelect < len(val) && val[afterSelect] == '*' {
		// Replace * with column name
		before := val[:afterSelect]
		after := val[afterSelect+1:]
		newVal := before + columnName + after
		m.textInput.SetValue(newVal)
		m.textInput.SetCursor(afterSelect + len(columnName))
	} else {
		// Insert column in the SELECT list
		// Find where to insert (after SELECT, before FROM/WHERE/etc)
		insertPos := findInsertPosition(val, afterSelect, cursor)

		before := val[:insertPos]
		after := val[insertPos:]

		// Add comma if needed
		insertText := columnName
		if insertPos > afterSelect && before[insertPos-1] != ' ' && before[insertPos-1] != ',' {
			insertText = ", " + columnName
		} else if insertPos > afterSelect && before[insertPos-1] != ' ' {
			insertText = " " + columnName
		}

		newVal := before + insertText + after
		m.textInput.SetValue(newVal)
		m.textInput.SetCursor(insertPos + len(insertText))
	}
}

// findSelectPosition finds the position of SELECT keyword before cursor.
func findSelectPosition(query string, cursor int) int {
	// Search backwards from cursor for SELECT (case insensitive)
	queryLower := toLower(query)
	searchStr := "select"

	// Look for "select" before cursor
	for i := cursor - 1; i >= 0; i-- {
		if i+len(searchStr) <= len(queryLower) {
			if queryLower[i:i+len(searchStr)] == searchStr {
				// Check it's a word boundary
				if i == 0 || !isAlphaNum(query[i-1]) {
					return i
				}
			}
		}
	}
	return -1
}

// findInsertPosition finds where to insert column in SELECT list.
func findInsertPosition(query string, afterSelect int, cursor int) int {
	// If cursor is in SELECT clause area, use cursor
	if cursor > afterSelect {
		// Check if cursor is before FROM/WHERE/GROUP/ORDER/etc
		queryLower := toLower(query)
		keywords := []string{" from ", " where ", " group ", " order ", " having ", " limit "}

		closestKeyword := len(query)
		for _, kw := range keywords {
			idx := indexOf(queryLower, kw, afterSelect)
			if idx != -1 && idx < closestKeyword {
				closestKeyword = idx
			}
		}

		if cursor < closestKeyword {
			return cursor
		}
		return closestKeyword
	}

	return afterSelect
}

// Helper functions
func toLower(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		if s[i] >= 'A' && s[i] <= 'Z' {
			result[i] = s[i] + 32
		} else {
			result[i] = s[i]
		}
	}
	return string(result)
}

func isAlphaNum(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_'
}

func indexOf(s, substr string, start int) int {
	for i := start; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// extractColumnsFromSelect parses the SELECT clause and returns column names.
func extractColumnsFromSelect(query string) []string {
	queryLower := toLower(query)

	// Find SELECT position
	selectIdx := indexOf(queryLower, "select", 0)
	if selectIdx == -1 {
		return nil
	}

	// Find FROM position (or end of query)
	fromIdx := len(query)
	keywords := []string{" from ", " where ", " group ", " order ", " having ", " limit "}
	for _, kw := range keywords {
		idx := indexOf(queryLower, kw, selectIdx)
		if idx != -1 && idx < fromIdx {
			fromIdx = idx
		}
	}

	// Extract the SELECT clause
	selectClause := strings.TrimSpace(query[selectIdx+6 : fromIdx])
	if selectClause == "" {
		return nil
	}

	// Handle SELECT *
	if strings.TrimSpace(selectClause) == "*" {
		return nil // Return nil to indicate "all columns"
	}

	// Split by comma and clean up column names
	columns := strings.Split(selectClause, ",")
	var result []string
	for _, col := range columns {
		col = strings.TrimSpace(col)
		// Remove AS aliases if present
		if idx := indexOf(toLower(col), " as ", 0); idx != -1 {
			col = strings.TrimSpace(col[:idx])
		}
		// Remove quotes if present
		col = strings.Trim(col, "`\"'")
		if col != "" {
			result = append(result, col)
		}
	}

	return result
}
