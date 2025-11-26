package input

import "slices"

type Format string

const (
	FormatJSON Format = "json"
)

func Formats() []string {
	return []string{string(FormatJSON)}
}

func IsValidFormat(format string) bool {
	return slices.Contains(Formats(), string(format))
}
