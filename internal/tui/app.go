package tui

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Color palette for consistent theming.
// Using ANSI 256 color codes: https://www.ditig.com/256-colors-cheat-sheet
var (
	colorFontNormal      = lipgloss.Color("15")
	colorFontPlaceholder = lipgloss.Color("8")
	colorFontError       = lipgloss.Color("9")
	colorFontAccent      = lipgloss.Color("6")
)

// Styles
var (
	headerStyle = lipgloss.NewStyle().Foreground(colorFontPlaceholder).Bold(true)
	modeStyle   = lipgloss.NewStyle().Foreground(colorFontAccent).Bold(true)
	errorStyle  = lipgloss.NewStyle().Foreground(colorFontError)
)

// Result holds the final query when exiting TUI.
type Result struct {
	Query string
}

// Model represents the TUI application state.
type Model struct {
	db         *sql.DB
	mode       Mode
	table      table.Model
	textInput  textinput.Model
	err        error
	width      int
	height     int
	colCursor  int // selected column index
	colOffset  int // column scroll offset for display
	allColumns []table.Column
	allRows    []table.Row
	result     *Result // set when exiting with a query to execute
	tableNames []string
}

// Result returns the final query result.
func (m Model) Result() *Result {
	return m.result
}

// NewModel creates a new TUI model.
func NewModel(db *sql.DB, tableNames []string) Model {
	ti := textinput.New()
	ti.Placeholder = "SQL Query..."
	ti.Focus()
	ti.CharLimit = 1000
	ti.Width = 100

	ti.TextStyle = lipgloss.NewStyle().Foreground(colorFontNormal)
	ti.PlaceholderStyle = lipgloss.NewStyle().Foreground(colorFontPlaceholder)
	ti.PromptStyle = lipgloss.NewStyle().Foreground(colorFontAccent)
	ti.Cursor.Style = lipgloss.NewStyle().Foreground(colorFontAccent)

	// Set default query using first available table
	if len(tableNames) > 0 {
		ti.SetValue(fmt.Sprintf("SELECT * FROM %s LIMIT 10", tableNames[0]))
	}

	t := table.New(
		table.WithColumns([]table.Column{{Title: "Results", Width: 20}}),
		table.WithRows([]table.Row{}),
		table.WithFocused(false),
		table.WithHeight(10),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(colorFontNormal).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(colorFontAccent).
		Bold(true)
	t.SetStyles(s)

	return Model{
		db:         db,
		mode:       ModeQuery,
		table:      t,
		textInput:  ti,
		colCursor:  0,
		colOffset:  0,
		tableNames: tableNames,
	}
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.handleWindowResize(msg)

	case tea.KeyMsg:
		c, quit := m.handleKeyMsg(msg)
		if quit {
			return m, tea.Quit
		}
		if c != nil {
			cmds = append(cmds, c)
		}
	}

	m.table, _ = m.table.Update(msg)
	m.textInput, cmd = m.textInput.Update(msg)
	cmds = append(cmds, cmd)

	// Execute query in real-time when in query mode
	if m.mode == ModeQuery {
		m.executeQuery()
	}

	return m, tea.Batch(cmds...)
}

// handleWindowResize updates dimensions on window resize.
func (m *Model) handleWindowResize(msg tea.WindowSizeMsg) {
	m.width = msg.Width
	m.height = msg.Height
	m.table.SetHeight(msg.Height - 10)
	m.textInput.Width = msg.Width - 10
	m.updateVisibleColumns()
}

// handleKeyMsg processes key events and returns a command and quit flag.
func (m *Model) handleKeyMsg(msg tea.KeyMsg) (tea.Cmd, bool) {
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

	case tea.KeyLeft, tea.KeyRight, tea.KeyRunes:
		if m.mode == ModeTable {
			m.handleTableScroll(msg)
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

// handleTableScroll handles column cursor movement in table mode.
func (m *Model) handleTableScroll(msg tea.KeyMsg) {
	moveLeft := msg.Type == tea.KeyLeft || msg.String() == "h"
	moveRight := msg.Type == tea.KeyRight || msg.String() == "l"

	if moveLeft && m.colCursor > 0 {
		m.colCursor--
	} else if moveRight && m.colCursor < len(m.allColumns)-1 {
		m.colCursor++
	} else {
		return
	}

	// Adjust scroll offset to keep cursor visible
	visibleCols := m.visibleColumnCount()
	if m.colCursor < m.colOffset {
		m.colOffset = m.colCursor
	} else if m.colCursor >= m.colOffset+visibleCols {
		m.colOffset = m.colCursor - visibleCols + 1
	}
	m.updateVisibleColumns()
}

// visibleColumnCount returns the number of columns that can fit in the view.
func (m *Model) visibleColumnCount() int {
	if m.width == 0 {
		return 5 // default
	}
	// Approximate: each column is ~15 chars + border
	count := (m.width - 4) / 17
	if count < 1 {
		return 1
	}
	return count
}

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

// calculateColumnWidth returns the optimal column width based on terminal width.
func (m *Model) calculateColumnWidth(numCols int) int {
	if m.width == 0 || numCols == 0 {
		return 15 // default
	}
	// Available width minus borders and padding
	available := m.width - 4 - (numCols * 2) // 2 chars per column for borders
	width := available / numCols
	if width < 10 {
		return 10 // minimum width
	}
	if width > 50 {
		return 50 // maximum width
	}
	return width
}

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

	// Apply border style with padding
	style := lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(colorFontNormal).
		Padding(0, 1)

	b.WriteString("\n") // ensure top border is visible in alt screen
	b.WriteString(style.Render(content))
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
	if rowIdx < 0 || rowIdx >= len(m.allRows) {
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

// Run starts the TUI application and returns the final query if any.
func Run(db *sql.DB, tableNames []string) (*Result, error) {
	p := tea.NewProgram(NewModel(db, tableNames), tea.WithAltScreen())
	m, err := p.Run()
	if err != nil {
		return nil, err
	}
	if model, ok := m.(Model); ok {
		return model.result, nil
	}
	return nil, nil
}
