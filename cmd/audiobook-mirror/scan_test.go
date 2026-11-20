package main

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// touch creates an empty file at dir/rel, and any folders it needs
func touch(t *testing.T, dir, rel string) {
	t.Helper()
	path := filepath.Join(dir, rel)
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	require.NoError(t, os.WriteFile(path, nil, 0o644))
}

func TestScanBooks(t *testing.T) {
	src := t.TempDir()
	for _, rel := range []string{
		"A/Book One/001 Chapter.m4a",
		"A/Book One/cover.jpg",
		"A/.Hidden Book/001.m4a",
		"A/stray.m4a",
		"A/notes.txt",
		"B/Book Two/Disc 01/01 Track.flac",
		"loose.FLAC",
		".DS_Store",
	} {
		touch(t, src, rel)
	}

	books, stray, err := scanBooks(src, "/mirror")
	require.NoError(t, err)

	assert.Equal(t, []book{
		{src: filepath.Join(src, "A", "Book One"), dst: "/mirror/A/Book One"},
		{src: filepath.Join(src, "B", "Book Two"), dst: "/mirror/B/Book Two"},
	}, books)
	assert.Equal(t, []string{
		filepath.Join(src, "A", "stray.m4a"),
		filepath.Join(src, "loose.FLAC"),
	}, stray)
}

func TestScanBooksMissing(t *testing.T) {
	_, _, err := scanBooks(filepath.Join(t.TempDir(), "missing"), "/mirror")
	assert.ErrorIs(t, err, fs.ErrNotExist)
}
