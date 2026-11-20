package main

import (
	"io/fs"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckOptions(t *testing.T) {
	root := t.TempDir()
	lib := filepath.Join(root, "lib")
	touch(t, root, "lib/A/B/001.m4a")
	touch(t, root, "file.txt")

	tests := []struct {
		name    string
		src     string
		dest    string
		wantErr error
	}{
		{"separate", lib, filepath.Join(root, "mirror"), nil},
		{"mirror inside library", lib, filepath.Join(lib, "mirror"), errDestInSrc},
		{"library inside mirror", lib, root, errSrcInDest},
		{"same folder", lib, lib, errDestInSrc},
		{"src is a file", filepath.Join(root, "file.txt"), filepath.Join(root, "mirror"), errSrcNotDir},
		{"src missing", filepath.Join(root, "missing"), filepath.Join(root, "mirror"), fs.ErrNotExist},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o, err := checkOptions(options{src: tt.src, dest: tt.dest, prune: true})
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, options{src: tt.src, dest: tt.dest, prune: true}, o)
		})
	}

	t.Run("relative paths made absolute", func(t *testing.T) {
		t.Chdir(root)
		o, err := checkOptions(options{src: "lib", dest: "mirror"})
		require.NoError(t, err)
		assert.Equal(t, lib, o.src)
		assert.Equal(t, filepath.Join(root, "mirror"), o.dest)
	})
}
