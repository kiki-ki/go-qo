package ui

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ColumnPickerState holds the state for the column picker.
type ColumnPickerState struct {
	list          list.Model
	lastFilterLen int
}

// newColumnPickerState creates a new column picker state.
func newColumnPickerState(columns []table.Column, selectedColumns []string) *ColumnPickerState {
	items := make([]list.Item, len(columns))

	// Create a set of selected column names for quick lookup
	selectedSet := make(map[string]bool)
	for _, name := range selectedColumns {
		selectedSet[strings.ToLower(name)] = true
	}

	for i, col := range columns {
		isSelected := selectedSet[strings.ToLower(col.Title)]
		items[i] = ColumnItem{name: col.Title, selected: isSelected}
	}

	l := list.New(items, &ColumnPickerDelegate{}, 0, 0)
	l.Title = ""
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	l.SetShowFilter(false)
	l.SetFilteringEnabled(true)

	l.FilterInput.Cursor.Style = styleTextAccent
	l.FilterInput.Cursor.TextStyle = styleTextBase
	l.FilterInput.Prompt = ""
	l.FilterInput.Blur()

	l.FilterInput.Focus()

	// Custom styles for fzf-like appearance
	l.Styles.NoItems = lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "240", Dark: "238"}).
		Padding(0, 1)

	return &ColumnPickerState{
		list:          l,
		lastFilterLen: 0,
	}
}

// ColumnPickerDelegate handles rendering with highlighting.
type ColumnPickerDelegate struct{}

func (d *ColumnPickerDelegate) Height() int {
	return 1
}

func (d *ColumnPickerDelegate) Spacing() int {
	return 0
}

func (d *ColumnPickerDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd {
	return nil
}

// Render renders a column item with fuzzy match highlighting.
func (d *ColumnPickerDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	colItem, ok := item.(ColumnItem)
	if !ok {
		return
	}

	isSelected := index == m.Index()
	width := m.Width()
	if width <= 0 {
		width = 50
	}

	// Get filter value
	filterValue := m.FilterValue()

	// Render checkbox with better visibility
	checkbox := "☐"
	checkboxStyle := styleTextBase
	if colItem.selected {
		checkbox = "☑"
		checkboxStyle = styleTextAccent
	}
	renderedCheckbox := checkboxStyle.Render(checkbox)

	// Highlight matches if filtering
	name := colItem.name
	var renderedName string

	if filterValue != "" {
		renderedName = highlightMatches(name, filterValue, width-5)
	} else {
		if len(name) > width-5 {
			name = name[:width-5] + "..."
		}
		renderedName = name
	}

	// Style based on selection state - add visual "raised" effect
	var style lipgloss.Style
	var prefix string
	if isSelected {
		// Add visual indicator for "raised" effect without background
		prefix = styleTextAccent.Render("▶ ") // Arrow indicator
		style = styleTextAccent
	} else {
		prefix = "  " // Spacing to align non-selected items
		style = styleTextBase
	}

	result := fmt.Sprintf("%s%s %s", prefix, renderedCheckbox, renderedName)
	fmt.Fprint(w, style.Render(result))
}

// highlightMatches highlights matching characters in the text (fuzzy style).
func highlightMatches(text, query string, maxWidth int) string {
	if query == "" {
		if len(text) > maxWidth {
			return text[:maxWidth-3] + "..."
		}
		return text
	}

	textLower := strings.ToLower(text)
	queryLower := strings.ToLower(query)

	// Simple fuzzy matching: find positions of query characters
	matches := make([]bool, len(text))
	queryIdx := 0

	for i := 0; i < len(text) && queryIdx < len(queryLower); i++ {
		if textLower[i] == queryLower[queryIdx] {
			matches[i] = true
			queryIdx++
		}
	}

	// Build highlighted string
	var result strings.Builder
	for i, r := range text {
		if i >= maxWidth-3 {
			result.WriteString("...")
			break
		}
		if matches[i] {
			// Highlight matched character with better contrast
			highlighted := lipgloss.NewStyle().
				Foreground(lipgloss.Color("11")).
				Background(lipgloss.AdaptiveColor{Light: "255", Dark: "0"}).
				Bold(true).
				Render(string(r))
			result.WriteString(highlighted)
		} else {
			result.WriteRune(r)
		}
	}

	return result.String()
}
