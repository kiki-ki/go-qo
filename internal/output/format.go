package output

import "slices"

type Format string

const (
	FormatTable Format = "table"
	FormatJSON  Format = "json"
	FormatJSONL Format = "jsonl"
	FormatCSV   Format = "csv"
	FormatTSV   Format = "tsv"
	FormatPSV   Format = "psv"
)

func Formats() []string {
	return []string{string(FormatTable), string(FormatJSON), string(FormatJSONL), string(FormatCSV), string(FormatTSV), string(FormatPSV)}
}

func IsValidFormat(format string) bool {
	return slices.Contains(Formats(), string(format))
}
