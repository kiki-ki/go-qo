package output_test

import (
	"testing"

	"github.com/kiki-ki/go-qo/internal/output"
)

func TestFormats(t *testing.T) {
	formats := output.Formats()
	if len(formats) != 5 {
		t.Errorf("expected 5 formats, got %d", len(formats))
	}
}

func TestIsValidFormat(t *testing.T) {
	tests := []struct {
		in   string
		want bool
	}{
		{"table", true},
		{"json", true},
		{"jsonl", true},
		{"csv", true},
		{"tsv", true},
		{"TABLE", false}, // case sensitive
		{"yaml", false},
		{"", false},
	}

	for _, tt := range tests {
		if got := output.IsValidFormat(tt.in); got != tt.want {
			t.Errorf("IsValidFormat(%q) = %v, want %v", tt.in, got, tt.want)
		}
	}
}
