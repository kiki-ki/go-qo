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

func TestFormatFromPath(t *testing.T) {
	tests := []struct {
		path      string
		want      input.Format
		wantFound bool
	}{
		{"data.json", input.FormatJSON, true},
		{"data.jsonl", input.FormatJSON, true},
		{"data.ndjson", input.FormatJSON, true},
		{"data.csv", input.FormatCSV, true},
		{"data.tsv", input.FormatTSV, true},
		{"data.psv", input.FormatPSV, true},
		{"path/to/data.CSV", input.FormatCSV, true}, // case insensitive
		{"archive.csv.gz", "", false},               // only the last extension counts
		{"data.txt", "", false},
		{"data", "", false},
		{"", "", false},
	}

	for _, tt := range tests {
		got, found := input.FormatFromPath(tt.path)
		if found != tt.wantFound {
			t.Errorf("FormatFromPath(%q) found = %v, want %v", tt.path, found, tt.wantFound)
			continue
		}
		if found && got != tt.want {
			t.Errorf("FormatFromPath(%q) = %q, want %q", tt.path, got, tt.want)
		}
	}
}
