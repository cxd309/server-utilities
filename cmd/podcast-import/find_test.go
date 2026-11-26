package main

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/text/unicode/norm"
)

func TestFindFile(t *testing.T) {
	root := t.TempDir()
	// On disk: decomposed, "u" then a combining circumflex. In the manifest: one "û"
	disk := norm.NFD.String("Lothlórien/The Bridge of Khazad-dûm.flac")
	require.NoError(t, os.MkdirAll(filepath.Join(root, filepath.Dir(disk)), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(root, disk), nil, 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(root, "plain.flac"), nil, 0o644))

	path, err := findFile(root, norm.NFC.String("Lothlórien/The Bridge of Khazad-dûm.flac"))
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(root, disk), path, "the name as it is on disk")

	path, err = findFile(root, "plain.flac")
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(root, "plain.flac"), path)

	_, err = findFile(root, "Lothlórien/missing.flac")
	assert.ErrorIs(t, err, fs.ErrNotExist)
}
