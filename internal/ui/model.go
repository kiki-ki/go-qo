package ui

import (
	"database/sql"
	"fmt"

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

// NewModel creates a new UI model.
func NewModel(db *sql.DB, tableNames []string) Model {
	ti := newTextInput(tableNames)
	t := newTable()

	return Model{
		db:         db,
		mode:       ModeQuery,
		table:      t,
		textInput:  ti,
		tableNames: tableNames,
	}
}

// newTextInput creates a configured text input component.
func newTextInput(tableNames []string) textinput.Model {
	ti := textinput.New()
	ti.Placeholder = "SQL Query..."
	ti.Focus()
	ti.CharLimit = 1000
	ti.Width = 100

	ti.TextStyle = lipgloss.NewStyle().Foreground(colorNormal)
	ti.PlaceholderStyle = lipgloss.NewStyle().Foreground(colorPlaceholder)
	ti.PromptStyle = lipgloss.NewStyle().Foreground(colorAccent)
	ti.Cursor.Style = lipgloss.NewStyle().Foreground(colorAccent)

	if len(tableNames) > 0 {
		ti.SetValue(fmt.Sprintf("SELECT * FROM %s LIMIT 10", tableNames[0]))
	}

	return ti
}

// newTable creates a configured table component.
func newTable() table.Model {
	t := table.New(
		table.WithColumns([]table.Column{{Title: "Results", Width: 20}}),
		table.WithRows([]table.Row{}),
		table.WithFocused(false),
		table.WithHeight(10),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(colorNormal).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(colorAccent).
		Bold(true)
	t.SetStyles(s)

	return t
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

// Update implements tea.Model.
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

	if m.mode == ModeQuery {
		m.executeQuery()
	}

	return m, tea.Batch(cmds...)
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
