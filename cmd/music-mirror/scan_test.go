package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// touch creates an empty file (and its parent folders) under root
func touch(t *testing.T, root, rel string) {
	t.Helper()
	p := filepath.Join(root, rel)
	require.NoError(t, os.MkdirAll(filepath.Dir(p), 0o755))
	require.NoError(t, os.WriteFile(p, nil, 0o644))
}

func TestScan(t *testing.T) {
	src := t.TempDir()
	touch(t, src, "Artist/Album/01 Song.flac")
	touch(t, src, "Artist/Album/02 Song.MP3")
	touch(t, src, "Artist/Album/cover.jpg")
	touch(t, src, "Artist/Album/.hidden.flac")
	touch(t, src, ".stfolder/x.flac")

	jobs, err := scan(src, "/dest")
	require.NoError(t, err)

	assert.Equal(t, []job{
		{src: filepath.Join(src, "Artist/Album/01 Song.flac"), dst: "/dest/Artist/Album/01 Song.mp3", kind: kindEncode},
		{src: filepath.Join(src, "Artist/Album/02 Song.MP3"), dst: "/dest/Artist/Album/02 Song.mp3", kind: kindCopy},
	}, jobs)
}

func TestScanCollision(t *testing.T) {
	src := t.TempDir()
	touch(t, src, "a.flac")
	touch(t, src, "a.mp3")

	jobs, err := scan(src, "/dest")
	require.NoError(t, err)
	require.Len(t, jobs, 1)
	assert.Equal(t, kindEncode, jobs[0].kind) // a.flac sorts first and wins
}
