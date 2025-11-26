package tui

import (
	"database/sql"
	"fmt"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

// Model represents the TUI application state.
type Model struct {
	db        *sql.DB
	table     table.Model
	textInput textinput.Model
	err       error
}

// NewModel creates a new TUI model.
func NewModel(db *sql.DB, tableNames []string) Model {
	ti := textinput.New()
	ti.Placeholder = "SQL Query..."
	ti.Focus()
	ti.CharLimit = 1000
	ti.Width = 100

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
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)

	return Model{
		db:        db,
		table:     t,
		textInput: ti,
	}
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyTab:
			// Toggle focus between input and table
			if m.textInput.Focused() {
				m.textInput.Blur()
				m.table.Focus()
			} else {
				m.table.Blur()
				m.textInput.Focus()
			}
		}
	}

	m.table, _ = m.table.Update(msg)
	m.textInput, cmd = m.textInput.Update(msg)

	// Execute query in real-time when input is focused
	if m.textInput.Focused() {
		query := m.textInput.Value()
		if query != "" {
			rows, err := m.db.Query(query)
			if err == nil {
				defer rows.Close()
				cols, tableRows, err := SQLRowsToTable(rows)
				if err == nil {
					// Clear rows first to avoid panic when column count changes
					m.table.SetRows([]table.Row{})
					m.table.SetColumns(cols)
					m.table.SetRows(tableRows)
					m.err = nil
				}
			} else {
				m.err = err
			}
		}
	}

	return m, cmd
}

func (m Model) View() string {
	errView := ""
	if m.err != nil {
		errView = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render(fmt.Sprintf("\nError: %v", m.err))
	}

	return baseStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			" SQL Editor (Tab to switch focus)",
			m.textInput.View(),
			errView,
			"\n",
			m.table.View(),
		),
	) + "\n"
}

// Run starts the TUI application.
func Run(db *sql.DB, tableNames []string) error {
	p := tea.NewProgram(NewModel(db, tableNames), tea.WithAltScreen())
	_, err := p.Run()
	return err
}
