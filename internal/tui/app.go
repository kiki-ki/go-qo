package tui

import (
	"database/sql"
	"fmt"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
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

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(colorFontNormal)

// Model represents the TUI application state.
type Model struct {
	db              *sql.DB
	mode            Mode
	table           table.Model
	textInput       textinput.Model
	viewport        viewport.Model
	err             error
	width           int
	height          int
	colScrollOffset int // column scroll offset (number of columns to skip)
	allColumns      []table.Column
	allRows         []table.Row
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

	vp := viewport.New(80, 10)

	return Model{
		db:              db,
		mode:            ModeQuery,
		table:           t,
		textInput:       ti,
		viewport:        vp,
		colScrollOffset: 0,
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
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.Width = msg.Width - 4   // account for borders
		m.viewport.Height = msg.Height - 8 // account for header, input, etc.
		m.table.SetHeight(m.viewport.Height - 2)

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc: // Quit
			return m, tea.Quit
		case tea.KeyTab: // Switch mode
			if m.mode == ModeQuery {
				m.mode = ModeTable
				m.textInput.Blur()
				m.table.Focus()
			} else {
				m.mode = ModeQuery
				m.table.Blur()
				m.textInput.Focus()
				cmds = append(cmds, textinput.Blink)
			}
		case tea.KeyLeft, tea.KeyRight, tea.KeyRunes:
			// Horizontal scroll when in table mode
			if m.mode == ModeTable {
				scrollLeft := msg.Type == tea.KeyLeft || msg.String() == "h"
				scrollRight := msg.Type == tea.KeyRight || msg.String() == "l"

				if scrollLeft && m.colScrollOffset > 0 {
					m.colScrollOffset--
					m.updateVisibleColumns()
				}
				if scrollRight && m.colScrollOffset < len(m.allColumns)-1 {
					m.colScrollOffset++
					m.updateVisibleColumns()
				}
			}
		}
	}

	m.table, _ = m.table.Update(msg)
	m.textInput, cmd = m.textInput.Update(msg)
	cmds = append(cmds, cmd)

	// Execute query in real-time when in query mode
	if m.mode == ModeQuery {
		query := m.textInput.Value()
		if query != "" {
			rows, err := m.db.Query(query)
			if err == nil {
				defer rows.Close()
				cols, tableRows, err := SQLRowsToTable(rows)
				if err == nil {
					// Store all data
					m.allColumns = cols
					m.allRows = tableRows
					m.colScrollOffset = 0
					m.updateVisibleColumns()
					m.err = nil
				}
			} else {
				m.err = err
			}
		}
	}

	return m, tea.Batch(cmds...)
}

// updateVisibleColumns updates the table with visible columns based on scroll offset.
func (m *Model) updateVisibleColumns() {
	if len(m.allColumns) == 0 {
		return
	}

	// Get visible columns from offset
	visibleCols := m.allColumns[m.colScrollOffset:]

	// Build visible rows with matching columns
	visibleRows := make([]table.Row, len(m.allRows))
	for i, row := range m.allRows {
		if m.colScrollOffset < len(row) {
			visibleRows[i] = row[m.colScrollOffset:]
		} else {
			visibleRows[i] = table.Row{}
		}
	}

	m.table.SetRows([]table.Row{})
	m.table.SetColumns(visibleCols)
	m.table.SetRows(visibleRows)
}

func (m Model) View() string {
	errView := ""
	if m.err != nil {
		errView = lipgloss.NewStyle().Foreground(colorFontError).Render(fmt.Sprintf("\nError: %v", m.err))
	}

	headerStyle := lipgloss.NewStyle().Foreground(colorFontPlaceholder).Bold(true)
	modeStyle := lipgloss.NewStyle().Foreground(colorFontAccent).Bold(true)

	header := fmt.Sprintf(" [%s] %s", modeStyle.Render(string(m.mode)), m.mode.CommandsHint())

	if m.mode == ModeTable {
		// position info
		if len(m.allRows) != 0 {
			row := m.table.Cursor() + 1
			col := m.colScrollOffset + 1
			header += fmt.Sprintf(" (row: %d/%d, col: %d/%d)", row, len(m.allRows), col, len(m.allColumns))
		} else {
			header += " (no data)"
		}
	}

	return baseStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			headerStyle.Render(header),
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
