package ui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

const debounceDelay = 150 * time.Millisecond

// debounceMsg is sent after the debounce delay to trigger query execution.
type debounceMsg struct {
	query string
}

// NewDebounceMsg creates a debounce message for testing purposes.
func NewDebounceMsg(query string) tea.Msg {
	return debounceMsg{query: query}
}

// scheduleQueryExecution returns a command that triggers query execution after debounce delay.
func (m *Model) scheduleQueryExecution() tea.Cmd {
	query := m.pendingQuery
	return tea.Tick(debounceDelay, func(time.Time) tea.Msg {
		return debounceMsg{query: query}
	})
}

// handleDebounceMsg executes the query if it matches the pending query.
func (m *Model) handleDebounceMsg(msg debounceMsg) {
	if msg.query == m.pendingQuery && msg.query != m.lastExecQuery {
		m.executeQuery()
		m.lastExecQuery = m.pendingQuery
	}
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
	m.table.SetCursor(0)
	m.syncTableView(true)
	m.err = nil
}
