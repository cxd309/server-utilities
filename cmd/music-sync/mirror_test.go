package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// touch creates an empty file (and its parent folders) under root.
func touch(t *testing.T, root, rel string) {
	t.Helper()
	p := filepath.Join(root, rel)
	require.NoError(t, os.MkdirAll(filepath.Dir(p), 0o755))
	require.NoError(t, os.WriteFile(p, nil, 0o644))
}

func TestListMirror(t *testing.T) {
	dir := t.TempDir()
	touch(t, dir, "A/Album/01.mp3")
	touch(t, dir, "A/Album/02.MP3")
	touch(t, dir, "A/Album/03.mp3.part") // interrupted encode
	touch(t, dir, ".music-sync")
	touch(t, dir, ".Trash/x.mp3")

	files, err := listMirror(dir)
	require.NoError(t, err)

	var paths []string
	for _, f := range files {
		paths = append(paths, f.path)
		assert.False(t, f.modTime.IsZero())
	}
	assert.Equal(t, []string{
		filepath.Join(dir, "A/Album/01.mp3"),
		filepath.Join(dir, "A/Album/02.MP3"),
	}, paths)
}

func TestLastSyncRoundTrip(t *testing.T) {
	dir := t.TempDir()

	last, err := readLastSync(dir)
	require.NoError(t, err)
	assert.True(t, last.IsZero(), "never synced")

	now := time.Now()
	require.NoError(t, writeLastSync(dir, now))

	last, err = readLastSync(dir)
	require.NoError(t, err)
	assert.True(t, now.Equal(last), "want %v, got %v", now, last)
}
