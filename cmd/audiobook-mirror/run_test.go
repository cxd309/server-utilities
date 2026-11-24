package main

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRun(t *testing.T) {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg not installed")
	}
	ctx := context.Background()
	root := t.TempDir()
	o := options{src: filepath.Join(root, "lib"), dest: filepath.Join(root, "mirror"), prune: true}

	bookDir := filepath.Join(o.src, "An Author", "Test Book")
	require.NoError(t, os.MkdirAll(bookDir, 0o755))
	makeChapter(t, bookDir, "001 One.m4a", 1, 44100, "One")
	makeChapter(t, bookDir, "002 Two.m4a", 1, 44100, "Two")
	touch(t, o.src, "Pratchett/Mort/Disc 01/01 Track.flac")
	touch(t, o.src, "Pratchett/stray.m4a")

	// runOnce runs and returns what was printed, one line per entry
	runOnce := func(o options) (lines []string) {
		var out strings.Builder
		failed, err := run(ctx, o, &out)
		require.NoError(t, err)
		assert.Zero(t, failed)
		return strings.Split(strings.TrimSpace(out.String()), "\n")
	}

	assert.Equal(t, []string{
		"warning: Pratchett/stray.m4a is not in an Author/Title folder, skipping",
		"An Author/Test Book: built 1 part(s)",
		"Pratchett/Mort: skipped, no chapter files",
	}, runOnce(o), "first run")
	assert.FileExists(t, filepath.Join(o.dest, "An Author", "Test Book", "Part 01.m4b"))

	assert.Equal(t, []string{
		"warning: Pratchett/stray.m4a is not in an Author/Title folder, skipping",
		"Pratchett/Mort: skipped, no chapter files",
	}, runOnce(o), "second run has nothing to build")

	require.NoError(t, os.RemoveAll(bookDir))
	assert.Contains(t, runOnce(o), "prune An Author/Test Book", "book removed from library")
	assert.NoDirExists(t, filepath.Join(o.dest, "An Author"))
}
