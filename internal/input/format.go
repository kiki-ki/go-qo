package input

import "slices"

type Format string

const (
	FormatJSON Format = "json"
	FormatCSV  Format = "csv"
	FormatTSV  Format = "tsv"
)

func Formats() []string {
	return []string{string(FormatJSON), string(FormatCSV), string(FormatTSV)}
}

func IsValidFormat(format string) bool {
	return slices.Contains(Formats(), string(format))
}
