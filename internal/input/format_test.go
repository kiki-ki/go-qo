package input_test

import (
	"testing"

	"github.com/kiki-ki/go-qo/internal/input"
)

func TestFormats(t *testing.T) {
	formats := input.Formats()
	if len(formats) != 4 {
		t.Errorf("expected 4 formats, got %d", len(formats))
	}
	if formats[0] != "json" {
		t.Errorf("expected json, got %s", formats[0])
	}
	if formats[1] != "csv" {
		t.Errorf("expected csv, got %s", formats[1])
	}
	if formats[2] != "tsv" {
		t.Errorf("expected tsv, got %s", formats[2])
	}
	if formats[3] != "psv" {
		t.Errorf("expected psv, got %s", formats[3])
	}
}

func TestIsValidFormat(t *testing.T) {
	tests := []struct {
		in   string
		want bool
	}{
		{"json", true},
		{"csv", true},
		{"tsv", true},
		{"psv", true},
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
