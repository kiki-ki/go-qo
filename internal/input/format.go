package input

import "slices"

type Format string

const (
	FormatJSON Format = "json"
	FormatCSV  Format = "csv"
)

func Formats() []string {
	return []string{string(FormatJSON), string(FormatCSV)}
}

func IsValidFormat(format string) bool {
	return slices.Contains(Formats(), string(format))
}
