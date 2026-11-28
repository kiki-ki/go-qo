package db_test

import (
	"slices"
	"testing"

	"github.com/kiki-ki/go-qo/internal/db"
)

func TestNameSet_Claim(t *testing.T) {
	tests := []struct {
		name  string
		input []string
		want  []string
	}{
		{name: "ascii case collides", input: []string{"a", "A"}, want: []string{"a", "A_2"}},
		{name: "non-ascii case does not collide", input: []string{"ä", "Ä"}, want: []string{"ä", "Ä"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var names db.NameSet
			got := make([]string, len(tt.input))
			for i, name := range tt.input {
				got[i] = names.Claim(name)
			}
			if !slices.Equal(got, tt.want) {
				t.Errorf("Claim() = %q, want %q", got, tt.want)
			}
		})
	}
}
