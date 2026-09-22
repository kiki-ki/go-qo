package input

import (
	"path/filepath"
	"slices"
	"strings"
)

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

// formatByExt maps file extensions to the parser that handles them.
// JSON Lines is parsed by the JSON parser, which detects it from the content.
var formatByExt = map[string]Format{
	".json":   FormatJSON,
	".jsonl":  FormatJSON,
	".ndjson": FormatJSON,
	".csv":    FormatCSV,
	".tsv":    FormatTSV,
	".psv":    FormatPSV,
}

// FormatFromPath resolves the input format from a file extension,
// returning an empty Format when the extension is not recognized.
func FormatFromPath(path string) Format {
	return formatByExt[strings.ToLower(filepath.Ext(path))]
}
