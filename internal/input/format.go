package input

import "slices"

type Format string

const (
	FormatJSON Format = "json"
	FormatCSV  Format = "csv"
	FormatTSV  Format = "tsv"
	FormatPSV  Format = "psv"
)

func Formats() []string {
	return []string{string(FormatJSON), string(FormatCSV), string(FormatTSV), string(FormatPSV)}
}

func IsValidFormat(format string) bool {
	return slices.Contains(Formats(), string(format))
}
