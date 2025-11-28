package ui

import "github.com/charmbracelet/lipgloss"

// Color palette using ANSI 256 colors.
// Reference: https://www.ditig.com/256-colors-cheat-sheet
var (
	colorNormal      = lipgloss.Color("15") // white - normal text
	colorPlaceholder = lipgloss.Color("8")  // gray - secondary text
	colorError       = lipgloss.Color("9")  // red - error messages
	colorAccent      = lipgloss.Color("6")  // cyan - highlights
)

// Component styles.
var (
	baseStyle   = lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()).BorderForeground(colorNormal).Padding(0, 1)
	headerStyle = lipgloss.NewStyle().Foreground(colorPlaceholder).Bold(true)
	modeStyle   = lipgloss.NewStyle().Foreground(colorAccent).Bold(true)
	errorStyle  = lipgloss.NewStyle().Foreground(colorError)
)

// Table column dimensions.
const (
	defaultColumnWidth = 15
	minColumnWidth     = 10
	maxColumnWidth     = 50
	defaultVisibleCols = 5
)

// Cell display settings.
const (
	maxCellDisplay    = 50 // max characters before truncation
	columnBorderWidth = 2  // border width per column
)

// Layout offsets for terminal dimensions.
const (
	framePadding      = 4  // padding for frame borders
	tableHeightOffset = 10 // subtract from terminal height for table
	inputWidthOffset  = 10 // subtract from terminal width for input
)
