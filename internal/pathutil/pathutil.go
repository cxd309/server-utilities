package pathutil

import (
	"path/filepath"
	"strings"
)

// within reports whether path is parent itself or somewhere beneath it.
// Both should be absolute.
func Within(parent, path string) (inside bool) {
	rel, err := filepath.Rel(parent, path)
	if err != nil {
		return false
	}
	if rel == "." {
		return true
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}
