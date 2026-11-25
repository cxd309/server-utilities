package main

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cxd309/server-utilities/internal/ffmpeg"
)

func TestBuildDir(t *testing.T) {
	assert.Equal(t, "/m/Author/.Mort.building", buildDir("/m/Author/Mort"))
}

// names lists the entries in dir, hidden ones included
func names(t *testing.T, dir string) (list []string) {
	t.Helper()
	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	for _, e := range entries {
		list = append(list, e.Name())
	}
	return list
}

func TestSwapDir(t *testing.T) {
	t.Run("new book", func(t *testing.T) {
		root := t.TempDir()
		touch(t, root, ".Book.building/Part 01.m4b")
		dst := filepath.Join(root, "Author", "Book")

		require.NoError(t, swapDir(filepath.Join(root, ".Book.building"), dst))
		assert.Equal(t, []string{"Part 01.m4b"}, names(t, dst))
	})

	t.Run("replaces old book", func(t *testing.T) {
		root := t.TempDir()
		touch(t, root, "Book/Part 01.m4b")
		touch(t, root, "Book/Part 02.m4b")
		touch(t, root, ".Book.building/Part 01.m4b")
		dst := filepath.Join(root, "Book")

		require.NoError(t, swapDir(filepath.Join(root, ".Book.building"), dst))
		assert.Equal(t, []string{"Part 01.m4b"}, names(t, dst))
		assert.Equal(t, []string{"Book"}, names(t, root), "no build or old folder left behind")
	})
}

func TestBuildBook(t *testing.T) {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg not installed")
	}
	ctx := context.Background()
	src := t.TempDir()
	makeChapter(t, src, "001 One.m4a", 1, 44100, "One")
	makeChapter(t, src, "002 Two.m4a", 1, 44100, "Two")

	d, err := readDownload(ctx, src)
	require.NoError(t, err)

	root := t.TempDir()
	b := book{src: src, dst: filepath.Join(root, "Author", "Test Book (A Reader)")}
	touch(t, b.dst, "Part 03.m4b") // from an older, longer build

	// Each 1-second chapter is about 44 frames, so a limit of 50 gives one part each
	parts, err := buildBook(ctx, b, d, 50)
	require.NoError(t, err)
	assert.Equal(t, 2, parts)
	assert.Equal(t, []string{"Part 01.m4b", "Part 02.m4b"}, names(t, b.dst))

	info, err := ffmpeg.Probe(ctx, filepath.Join(b.dst, "Part 02.m4b"))
	require.NoError(t, err)
	assert.Equal(t, "Test Book (A Reader), Part 2 of 2", info.Format.Tags["title"])
	assert.Equal(t, "Test Book (A Reader)", info.Format.Tags["album"], "named after the folder, not the album tag")
	assert.Equal(t, "2/2", info.Format.Tags["track"])
}
