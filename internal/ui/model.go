package ui

import (
	"database/sql"
	"fmt"
	"os"

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
}

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
