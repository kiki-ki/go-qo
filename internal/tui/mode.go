package tui

import "slices"

type Mode string

const (
	ModeQuery Mode = "QUERY"
	ModeTable Mode = "TABLE"
)

func (m Mode) Commands() []modeCommand {
	switch m {
	case ModeQuery:
		return queryModeCommands
	case ModeTable:
		return tableModeCommands
	default:
		return baseModeCommands
	}
}

// CommandsHint returns a formatted string of all commands for the mode.
func (m Mode) CommandsHint() string {
	cmds := m.Commands()
	result := ""
	for i, cmd := range cmds {
		if i > 0 {
			result += ", "
		}
		result += cmd.String()
	}
	return result
}

var (
	baseModeCommands = []modeCommand{
		{key: "Tab", message: "switch mode"},
		{key: "Esc", message: "quit"},
	}
	queryModeCommands = slices.Concat(baseModeCommands, []modeCommand{
		{key: "Enter", message: "execute query"},
	})
	tableModeCommands = slices.Concat(baseModeCommands, []modeCommand{
		{key: "Arrows", message: "scroll"},
	})
)

type modeCommand struct {
	key     string
	message string
}

func (c modeCommand) String() string {
	return c.key + ": " + c.message
}
