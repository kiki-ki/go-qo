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
	db            *sql.DB
	mode          Mode
	table         table.Model
	textInput     textinput.Model
	err           error
	width         int
	height        int
	colCursor     int // selected column index
	colOffset     int // column scroll offset for display
	allColumns    []table.Column
	allRows       []table.Row
	result        *Result // set when exiting with a query to execute
	tableNames    []string
	pendingQuery  string // query waiting for debounce
	lastExecQuery string // last executed query to avoid duplicate execution
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
		m.updateCellMarker()
	}

	return cmds
}

// handleWindowResize updates dimensions on window resize.
func (m *Model) handleWindowResize(msg tea.WindowSizeMsg) {
	m.width = msg.Width
	m.height = msg.Height
	m.table.SetHeight(msg.Height - tableHeightOffset)
	m.textInput.Width = msg.Width - inputWidthOffset
	m.updateVisibleColumns()
}

// Run starts the UI application and returns the final query if any.
// It uses /dev/tty for input/output so that stdin/stdout remain available for piping.
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
