package ui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/table"
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

	// Store original table columns ONLY on first load (if tableColumns is empty)
	if len(m.tableColumns) == 0 && len(m.tableNames) > 0 {
		// Get all columns from the first table
		tableName := m.tableNames[0]
		tableCols, err := m.getTableColumns(tableName)
		if err == nil {
			m.tableColumns = tableCols
		}
	}

	m.tableState.SetData(cols, tableRows)
	m.table.SetCursor(0)
	m.syncTableView(true)
	m.err = nil
}

func (m *Model) getTableColumns(tableName string) ([]table.Column, error) {
	// Query to get column information from SQLite
	query := fmt.Sprintf("PRAGMA table_info(`%s`)", tableName)
	rows, err := m.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var columns []table.Column
	for rows.Next() {
		var cid int
		var name string
		var colType string
		var notnull int
		var dfltValue interface{}
		var pk int

		if err := rows.Scan(&cid, &name, &colType, &notnull, &dfltValue, &pk); err != nil {
			continue
		}

		columns = append(columns, table.Column{
			Title: name,
			Width: defaultColumnWidth,
		})
	}

	return columns, rows.Err()
}
