package tui

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

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
