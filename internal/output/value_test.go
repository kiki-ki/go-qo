package output_test

import (
	"testing"

	"github.com/kiki-ki/go-qo/internal/output"
)

func TestFormatValueForDisplay(t *testing.T) {
	tests := []struct {
		name  string
		input any
		want  string
	}{
		{"nil", nil, "(NULL)"},
		{"string", "hello", "hello"},
		{"int", 42, "42"},
		{"int64", int64(100), "100"},
		{"float64 whole", float64(5), "5"},
		{"float64 decimal", 3.14, "3.14"},
		{"bytes", []byte("test"), "test"},
		{"bool true", true, "true"},
		{"bool false", false, "false"},
		{"newline", "line1\nline2", "line1\\nline2"},
		{"tab", "col1\tcol2", "col1\\tcol2"},
		{"carriage return", "a\rb", "a\\rb"},
		{"crlf", "a\r\nb", "a\\r\\nb"},
		{"multiple spaces", "a    b", "a b"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := output.FormatValueForDisplay(tt.input)
			if got != tt.want {
				t.Errorf("FormatValueForDisplay(%v) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestFormatValueRaw(t *testing.T) {
	// FormatValueRaw is same as FormatValueForDisplay but without truncation
	long := string(make([]byte, 300))
	got := output.FormatValueRaw(long)
	if len(got) != 300 {
		t.Errorf("FormatValueRaw should not truncate, got length %d", len(got))
	}
}

func TestNormalizeValue(t *testing.T) {
	tests := []struct {
		name  string
		input any
		want  any
	}{
		{"nil", nil, nil},
		{"string", "hello", "hello"},
		{"bytes", []byte("test"), "test"},
		{"float64 whole", float64(5), int64(5)},
		{"float64 decimal", 3.14, 3.14},
		{"int", 42, 42},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := output.NormalizeValue(tt.input)
			if got != tt.want {
				t.Errorf("NormalizeValue(%v) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		maxLen int
		want   string
	}{
		{"short string", "hello", 10, "hello"},
		{"exact length", "hello", 5, "hello"},
		{"long string", "hello world", 8, "hello w…"},
		{"unicode", "こんにちは世界", 5, "こんにち…"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := output.Truncate(tt.input, tt.maxLen)
			if got != tt.want {
				t.Errorf("Truncate(%q, %d) = %q, want %q", tt.input, tt.maxLen, got, tt.want)
			}
		})
	}
}
