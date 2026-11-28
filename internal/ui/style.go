package ui

import (
	"image/color"

	"charm.land/lipgloss/v2"
)

// cellMarker is the character used to indicate the selected cell.
const cellMarker = "▶"

// Color palette.
// Reference: https://www.ditig.com/256-colors-cheat-sheet
var (
	colorError  = lipgloss.Color("9")  // red
	colorAccent = lipgloss.Color("13") // pink
)

var (
	styleTextMuted  = lipgloss.NewStyle().Faint(true)
	styleTextAccent = lipgloss.NewStyle().Foreground(colorAccent)
	styleTextError  = lipgloss.NewStyle().Foreground(colorError)
)

// styles holds the styles that depend on the terminal background.
type styles struct {
	colorBase color.Color
	textBase  lipgloss.Style
}

func newStyles(isDark bool) styles {
	base := lipgloss.LightDark(isDark)(lipgloss.Color("0"), lipgloss.Color("15")) // black/white
	return styles{
		colorBase: base,
		textBase:  lipgloss.NewStyle().Foreground(base),
	}
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

// Cell detail display settings.
const (
	cellDetailPadding  = 8  // total overhead: frame border/padding + safety margin
	cellDetailMinWidth = 20 // minimum width for cell value display
)
