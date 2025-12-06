package output

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/tidwall/gjson"
)

var multiSpaceRegex = regexp.MustCompile(`\s{2,}`)

const maxCellDisplay = 256

// NormalizeValue converts database values to appropriate Go types.
// This is used for data output (JSON, CSV, TSV) where values should be preserved.
func NormalizeValue(val any) any {
	if val == nil {
		return nil
	}

	switch v := val.(type) {
	case []byte:
		s := string(v)
		if parsed := tryParseJSON(s); parsed != nil {
			return parsed
		}
		return s
	case string:
		if parsed := tryParseJSON(v); parsed != nil {
			return parsed
		}
		return v
	case float64:
		if float64(int64(v)) == v {
			return int64(v)
		}
		return v
	default:
		return v
	}
}

// tryParseJSON attempts to parse a string as a JSON object or array.
// Returns nil if the string is not valid JSON or is a primitive value.
func tryParseJSON(s string) any {
	result := gjson.Parse(s)
	if result.IsObject() || result.IsArray() {
		return result.Value()
	}
	return nil
}

// FormatValueForDisplay converts a database value to string for CLI table display.
// This applies escape notation for control characters and truncation.
func FormatValueForDisplay(val any) string {
	s := FormatValueRaw(val)
	return Truncate(s, maxCellDisplay)
}

// FormatValueRaw converts a database value to string without truncation.
// This is used for TUI where truncation is done dynamically based on column width.
func FormatValueRaw(val any) string {
	var s string
	if val == nil {
		return "(NULL)"
	}
	switch v := val.(type) {
	case []byte:
		s = string(v)
	case float64:
		if float64(int64(v)) == v {
			s = fmt.Sprintf("%d", int64(v))
		} else {
			s = fmt.Sprintf("%g", v)
		}
	case map[string]any, []any:
		if b, err := json.Marshal(v); err == nil {
			s = string(b)
		} else {
			s = fmt.Sprintf("%v", v)
		}
	default:
		s = fmt.Sprintf("%v", v)
	}
	// Replace control characters with escape notation to prevent table display from breaking
	s = strings.ReplaceAll(s, "\r\n", "\\r\\n")
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "\r", "\\r")
	s = strings.ReplaceAll(s, "\t", "\\t")
	// Compress multiple spaces into one
	s = multiSpaceRegex.ReplaceAllString(s, " ")
	return s
}

// Truncate shortens a string to maxLen, adding "…" if truncated.
func Truncate(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen-1]) + "…"
}
