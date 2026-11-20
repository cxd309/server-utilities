package main

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrune(t *testing.T) {
	// mirror builds the same test mirror for each run
	mirror := func(t *testing.T) (dest string, books []book) {
		dest = t.TempDir()
		for _, rel := range []string{
			"A/Keep/Part 01.m4b",
			"A/Gone/Part 01.m4b",
			"A/Gone/.DS_Store",
			"A/.Half.building/Part 01.m4b",
			"A/.Swapped.old/Part 01.m4b",
			"A/Mine/notes.txt",
			"B/Also Gone/Part 01.m4b",
		} {
			touch(t, dest, rel)
		}
		books = []book{{src: "/lib/A/Keep", dst: filepath.Join(dest, "A", "Keep")}}
		return dest, books
	}

	t.Run("removes", func(t *testing.T) {
		dest, books := mirror(t)
		removed, err := prune(dest, books, false)
		require.NoError(t, err)

		assert.Equal(t, []string{
			filepath.Join(dest, "A", ".Half.building"),
			filepath.Join(dest, "A", ".Swapped.old"),
			filepath.Join(dest, "A", "Gone"),
			filepath.Join(dest, "B", "Also Gone"),
		}, removed)
		assert.Equal(t, []string{"A"}, names(t, dest), "empty author folder B removed")
		assert.Equal(t, []string{"Keep", "Mine"}, names(t, filepath.Join(dest, "A")))
	})

	t.Run("dry run changes nothing", func(t *testing.T) {
		dest, books := mirror(t)
		removed, err := prune(dest, books, true)
		require.NoError(t, err)

		assert.Len(t, removed, 4)
		assert.Equal(t, []string{"A", "B"}, names(t, dest))
		assert.DirExists(t, filepath.Join(dest, "A", "Gone"))
	})

	t.Run("no mirror yet", func(t *testing.T) {
		removed, err := prune(filepath.Join(t.TempDir(), "missing"), nil, false)
		require.NoError(t, err)
		assert.Empty(t, removed)
	})
}

func TestOnlyParts(t *testing.T) {
	tests := []struct {
		name  string
		files []string
		want  bool
	}{
		{"parts", []string{"Part 01.m4b", "Part 02.M4B"}, true},
		{"parts and Finder junk", []string{"Part 01.m4b", ".DS_Store"}, true},
		{"empty", nil, true},
		{"something else", []string{"Part 01.m4b", "notes.txt"}, false},
		{"a subfolder", []string{"Part 01.m4b", "Extras/x.m4b"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			for _, f := range tt.files {
				touch(t, dir, f)
			}
			ours, err := onlyParts(dir)
			require.NoError(t, err)
			assert.Equal(t, tt.want, ours)
		})
	}
}
