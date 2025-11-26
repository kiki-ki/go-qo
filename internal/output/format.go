package output

import "slices"

type Format string

const (
	FormatTable Format = "table"
	FormatJSON  Format = "json"
	FormatCSV   Format = "csv"
)

func Formats() []string {
	return []string{string(FormatTable), string(FormatJSON), string(FormatCSV)}
}

func IsValidFormat(format string) bool {
	return slices.Contains(Formats(), string(format))
}
