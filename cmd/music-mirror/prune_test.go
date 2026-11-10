package main

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrune(t *testing.T) {
	dest := t.TempDir()
	touch(t, dest, "Kept/Album/01.mp3")
	touch(t, dest, "Gone/Album/01.mp3")      // source deleted
	touch(t, dest, "Kept/Album/02.mp3.part") // interrupted encode
	touch(t, dest, "Kept/Album/03.mp3.part-id3v2")
	touch(t, dest, "Kept/Album/notes.txt") // not ours: leave alone
	jobs := []job{{dst: filepath.Join(dest, "Kept/Album/01.mp3")}}

	removed, err := prune(dest, jobs, false)
	require.NoError(t, err)

	assert.ElementsMatch(t, []string{
		filepath.Join(dest, "Gone/Album/01.mp3"),
		filepath.Join(dest, "Kept/Album/02.mp3.part"),
		filepath.Join(dest, "Kept/Album/03.mp3.part-id3v2"),
	}, removed)
	assert.FileExists(t, filepath.Join(dest, "Kept/Album/01.mp3"))
	assert.FileExists(t, filepath.Join(dest, "Kept/Album/notes.txt"))
	assert.NoDirExists(t, filepath.Join(dest, "Gone"), "emptied folders are removed")
	assert.DirExists(t, dest, "the mirror root itself is never removed")
}

func TestPruneDryRun(t *testing.T) {
	dest := t.TempDir()
	touch(t, dest, "Gone/01.mp3")

	removed, err := prune(dest, nil, true)
	require.NoError(t, err)

	assert.Equal(t, []string{filepath.Join(dest, "Gone/01.mp3")}, removed)
	assert.FileExists(t, filepath.Join(dest, "Gone/01.mp3"), "dry run changes nothing")
}

func TestPruneMissingDest(t *testing.T) {
	removed, err := prune(filepath.Join(t.TempDir(), "never-created"), nil, false)
	require.NoError(t, err)
	assert.Empty(t, removed)
}
