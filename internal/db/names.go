package db

import (
	"fmt"
	"strings"
)

// NameSet hands out identifiers that do not collide with the ones it has
// already handed out, appending "_2", "_3", ... to a taken name. The zero value
// is ready to use.
type NameSet struct {
	used map[string]bool
}

// Claim reserves name, or the first free suffixed form of it, and returns it.
func (s *NameSet) Claim(name string) string {
	if s.Reserve(name) {
		return name
	}
	for i := 2; ; i++ {
		candidate := fmt.Sprintf("%s_%d", name, i)
		if s.Reserve(candidate) {
			return candidate
		}
	}
}

// Reserve claims name as is and reports whether it was still free.
func (s *NameSet) Reserve(candidate string) bool {
	if s.used == nil {
		s.used = make(map[string]bool)
	}
	// SQLite folds only ASCII letters when comparing identifiers, so "a" and "A"
	// collide at CREATE while "ä" and "Ä" do not.
	key := strings.Map(func(r rune) rune {
		if r >= 'A' && r <= 'Z' {
			return r + ('a' - 'A')
		}
		return r
	}, candidate)
	if s.used[key] {
		return false
	}
	s.used[key] = true
	return true
}
