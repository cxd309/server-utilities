package main

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckPaths(t *testing.T) {
	root := t.TempDir()
	lib := filepath.Join(root, "music")
	require.NoError(t, os.Mkdir(lib, 0o755))
	touch(t, root, "file.flac")

	tests := []struct {
		name      string
		src, dest string
		wantErr   error
	}{
		{"siblings", lib, filepath.Join(root, "music-mirror"), nil},
		{"src missing", filepath.Join(root, "nope"), filepath.Join(root, "m"), fs.ErrNotExist},
		{"src is a file", filepath.Join(root, "file.flac"), filepath.Join(root, "m"), errSrcNotDir},
		{"dest inside src", lib, filepath.Join(lib, "mirror"), errDestInSrc},
		{"dest is src", lib, lib, errDestInSrc},
		{"src inside dest", lib, root, errSrcInDest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := checkPaths(tt.src, tt.dest)
			if tt.wantErr == nil {
				assert.NoError(t, err)
			} else {
				assert.ErrorIs(t, err, tt.wantErr)
			}
		})
	}
}
