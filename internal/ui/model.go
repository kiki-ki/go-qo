package ui

import (
	"database/sql"
	"fmt"
	"os"

	"charm.land/bubbles/v2/table"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
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
	styles     styles
	err        error
	width      int
	height     int
	tableNames []string
	warnings   []string

	// Table data state
	tableState *TableState

	// Result when exiting
	result *Result
}

// Result returns the final query result.
func (m Model) Result() *Result {
	return m.result
}

// NewModel creates a new UI model.
func NewModel(db *sql.DB, tableNames, warnings []string) Model {
	m := Model{
		db:         db,
		mode:       ModeQuery,
		table:      newTable(),
		textInput:  newTextInput(tableNames),
		tableState: NewTableState(),
		tableNames: tableNames,
		warnings:   warnings,
	}
	// Dark until the terminal answers the background color query.
	m.applyStyles(newStyles(true))
	return m
}

// applyStyles sets the styles of the model and its sub-components.
func (m *Model) applyStyles(s styles) {
	m.styles = s

	ts := textinput.DefaultStyles(true)
	for _, state := range []*textinput.StyleState{&ts.Focused, &ts.Blurred} {
		state.Text = s.textBase
		state.Placeholder = styleTextMuted
		state.Prompt = styleTextAccent
	}
	ts.Cursor.Color = colorAccent
	m.textInput.SetStyles(ts)

	tbl := table.DefaultStyles()
	tbl.Header = tbl.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(s.colorBase).
		BorderBottom(true).
		Bold(false)
	tbl.Selected = lipgloss.NewStyle()
	m.table.SetStyles(tbl)
}

// newTextInput creates a configured text input component.
func newTextInput(tableNames []string) textinput.Model {
	ti := textinput.New()
	ti.Placeholder = "SQL Query..."
	ti.Focus()
	ti.CharLimit = inputCharLimit
	ti.SetWidth(inputInitialWidth)

	if len(tableNames) > 0 {
		ti.SetValue(fmt.Sprintf("SELECT * FROM %s LIMIT %d", tableNames[0], defaultQueryLimit))
	}

	return ti
}

// newTable creates a configured table component.
func newTable() table.Model {
	return table.New(
		table.WithColumns([]table.Column{{Title: "Results", Width: initialColumnWidth}}),
		table.WithRows([]table.Row{}),
		table.WithFocused(false),
		table.WithHeight(initialTableHeight),
	)
}

// initQueryMsg is sent to trigger initial query execution.
type initQueryMsg struct{}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		textinput.Blink,
		tea.RequestBackgroundColor,
		func() tea.Msg { return initQueryMsg{} },
	)
}

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case initQueryMsg:
		query := m.textInput.Value()
		if query != "" {
			m.executeQuery()
		}
	case tea.BackgroundColorMsg:
		m.applyStyles(newStyles(msg.IsDark()))
	case tea.WindowSizeMsg:
		m.handleWindowResize(msg)
	case tea.KeyPressMsg:
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

	prevCursor := m.table.Cursor()

	m.table, _ = m.table.Update(msg)

	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	cmds = append(cmds, cmd)

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
	m.textInput.SetWidth(msg.Width - inputWidthOffset)
	m.syncTableView(true)
}

// handleKeyMsg processes key events and returns a command and quit flag.
func (m *Model) handleKeyMsg(msg tea.KeyPressMsg) (tea.Cmd, bool) {
	switch msg.String() {
	case "ctrl+c":
		return nil, true

	case "esc":
		query := m.textInput.Value()
		if query != "" {
			m.result = &Result{Query: query}
		}
		return nil, true

	case "enter":
		if m.mode != ModeQuery {
			return nil, false
		}
		query := m.textInput.Value()
		if query == "" {
			return nil, false
		}
		m.executeQuery()
		return nil, false

	case "tab":
		return m.toggleMode(), false

	default:
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
func (m *Model) handleTableNavigation(msg tea.KeyPressMsg) {
	moveLeft := msg.String() == "left" || msg.String() == "h"
	moveRight := msg.String() == "right" || msg.String() == "l"

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
		m.table.SetWidth(len(cols) * (colWidth + columnBorderWidth))
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
func Run(db *sql.DB, tableNames, warnings []string) (*Result, error) {
	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to open /dev/tty: %w", err)
	}
	defer func() { _ = tty.Close() }()

	p := tea.NewProgram(
		NewModel(db, tableNames, warnings),
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

	cols, tableRows, err := SQLRowsToTableData(rows)
	if err != nil {
		m.err = err
		return
	}

	m.tableState.SetData(cols, tableRows)
	m.syncTableView(true)
	// SetCursor clamps to the rows already in the table, so it must follow the sync.
	m.table.SetCursor(0)
	m.syncTableView(false)
	m.err = nil
}
