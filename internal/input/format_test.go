package input_test

import (
	"testing"

	"github.com/kiki-ki/go-qo/internal/input"
)

func TestFormats(t *testing.T) {
	formats := input.Formats()
	if len(formats) != 2 {
		t.Errorf("expected 2 formats, got %d", len(formats))
	}
	if formats[0] != "json" {
		t.Errorf("expected json, got %s", formats[0])
	}
	if formats[1] != "csv" {
		t.Errorf("expected csv, got %s", formats[1])
	}
}

func TestIsValidFormat(t *testing.T) {
	tests := []struct {
		in   string
		want bool
	}{
		{"json", true},
		{"csv", true},
		{"JSON", false}, // case sensitive
		{"CSV", false},  // case sensitive
		{"xml", false},
		{"", false},
	}

	for _, tt := range tests {
		if got := input.IsValidFormat(tt.in); got != tt.want {
			t.Errorf("IsValidFormat(%q) = %v, want %v", tt.in, got, tt.want)
		}
	}
}
