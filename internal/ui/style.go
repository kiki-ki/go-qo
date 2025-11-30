package ui

import "github.com/charmbracelet/lipgloss"

// Color palette with adaptive colors for light/dark terminal themes.
// Reference: https://www.ditig.com/256-colors-cheat-sheet
var (
	colorBase   = lipgloss.AdaptiveColor{Light: "0", Dark: "7"} // black/silver
	colorMuted  = lipgloss.AdaptiveColor{Light: "8", Dark: "8"} // gray
	colorError  = lipgloss.Color("9")                           // red
	colorAccent = lipgloss.Color("13")                          // pink
)

// Styles
var (
	styleTextBase   lipgloss.Style
	styleTextMuted  lipgloss.Style
	styleTextAccent lipgloss.Style
	styleTextError  lipgloss.Style
)

// initStyles initializes styles after the renderer is configured.
func initStyles() {
	styleTextBase = lipgloss.NewStyle().Foreground(colorBase)
	styleTextMuted = lipgloss.NewStyle().Foreground(colorMuted)
	styleTextAccent = lipgloss.NewStyle().Foreground(colorAccent)
	styleTextError = lipgloss.NewStyle().Foreground(colorError)
}

func init() {
	initStyles()
}

// Table column dimensions.
const (
	defaultColumnWidth = 15
	minColumnWidth     = 10
	maxColumnWidth     = 200
	maxVisibleCols     = 5 // max columns to display at once
	columnBorderWidth  = 2 // border width per column
)

// Layout offsets for terminal dimensions.
const (
	framePadding      = 4  // padding for frame borders
	tableHeightOffset = 10 // subtract from terminal height for table
	inputWidthOffset  = 10 // subtract from terminal width for input
)

// Input component settings.
const (
	inputCharLimit    = 1000 // max characters in query input
	inputInitialWidth = 100  // initial width before window resize
)

// Initial table settings (before data is loaded).
const (
	initialTableHeight = 10
	initialColumnWidth = 20
	defaultQueryLimit  = 10 // LIMIT value for default query
)
