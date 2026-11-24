package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testBook makes a three-chapter source file and returns its path, the
// book describing it, and a not-yet-existing output folder in a separate
// empty parent, so the tests can see exactly what gets created.
func testBook(t *testing.T) (src string, b book, out string) {
	t.Helper()
	dir := t.TempDir()
	src = filepath.Join(dir, "book.m4b")
	b = book{
		title:    "A Wizard of Earthsea",
		author:   "Ursula K. Le Guin",
		chapters: makeBook(t, src, "Opening Credits", "Chapter 1", "End Credits"),
	}
	parent := filepath.Join(dir, "out")
	require.NoError(t, os.Mkdir(parent, 0o755))
	return src, b, filepath.Join(parent, "book")
}

// names lists the entries in dir, including hidden ones.
func names(t *testing.T, dir string) (list []string) {
	t.Helper()
	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	for _, e := range entries {
		list = append(list, e.Name())
	}
	return list
}

func TestSplitBook(t *testing.T) {
	requireFFmpeg(t)
	src, b, out := testBook(t)

	require.NoError(t, splitBook(context.Background(), src, out, b))

	assert.Equal(t, []string{"book"}, names(t, filepath.Dir(out)), "no working folder left behind")
	assert.Equal(t, []string{"01 Opening Credits.m4a", "02 Chapter 1.m4a", "03 End Credits.m4a"}, names(t, out))

	info, err := os.Stat(out)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o755), info.Mode().Perm())
}

func TestSplitBookAlreadyExists(t *testing.T) {
	src, b, out := testBook(t)
	require.NoError(t, os.Mkdir(out, 0o755))

	err := splitBook(context.Background(), src, out, b)

	assert.ErrorIs(t, err, errAlreadySplit)
	assert.Equal(t, []string{"book"}, names(t, filepath.Dir(out)), "nothing else created")
}

func TestSplitBookFailureCleansUp(t *testing.T) {
	requireFFmpeg(t)
	_, b, out := testBook(t)

	err := splitBook(context.Background(), filepath.Join(t.TempDir(), "missing.m4b"), out, b)

	assert.Error(t, err)
	assert.Empty(t, names(t, filepath.Dir(out)), "working folder removed")
}

func TestSplitBookCancelled(t *testing.T) {
	requireFFmpeg(t)
	src, b, out := testBook(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := splitBook(ctx, src, out, b)

	assert.ErrorIs(t, err, context.Canceled)
	assert.Empty(t, names(t, filepath.Dir(out)))
}

func TestSplitBookPieceCountMismatch(t *testing.T) {
	requireFFmpeg(t)
	src, b, out := testBook(t)
	// The file is 3 s long, so cuts at 10 s and 20 s make no extra pieces
	b.chapters = []chapter{{start: 0, title: "A"}, {start: 10, title: "B"}, {start: 20, title: "C"}}

	err := splitBook(context.Background(), src, out, b)

	assert.ErrorIs(t, err, errPieceCount)
	assert.Empty(t, names(t, filepath.Dir(out)), "nothing renamed into place")
}
