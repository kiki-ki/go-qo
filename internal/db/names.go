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
	if s.take(name) {
		return name
	}
	for i := 2; ; i++ {
		candidate := fmt.Sprintf("%s_%d", name, i)
		if s.take(candidate) {
			return candidate
		}
	}
}

func (s *NameSet) take(candidate string) bool {
	if s.used == nil {
		s.used = make(map[string]bool)
	}
	// SQLite identifiers are case-insensitive, so names differing only in case
	// would pass a case-sensitive check here and still collide at CREATE.
	key := strings.ToLower(candidate)
	if s.used[key] {
		return false
	}
	s.used[key] = true
	return true
}
