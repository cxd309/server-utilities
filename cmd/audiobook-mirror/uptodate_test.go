package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setTime sets the modification time of dir/rel
func setTime(t *testing.T, dir, rel string, when time.Time) {
	t.Helper()
	require.NoError(t, os.Chtimes(filepath.Join(dir, rel), when, when))
}

func TestNewestIn(t *testing.T) {
	base := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	tests := []struct {
		name   string
		newest string // which path gets the newest time
		want   time.Time
	}{
		{"a chapter file", "001 One.m4a", base.Add(3 * time.Hour)},
		{"the cover", "cover.jpg", base.Add(3 * time.Hour)},
		{"a disc file", "Disc 01/01 Track.flac", base.Add(3 * time.Hour)},
		{"the book folder itself", ".", base.Add(3 * time.Hour)},
		{"a disc folder", "Disc 01", base.Add(3 * time.Hour)},
		{"only a hidden file", ".DS_Store", base},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			all := []string{"001 One.m4a", "cover.jpg", "Disc 01/01 Track.flac", ".DS_Store"}
			for _, rel := range all {
				touch(t, dir, rel)
			}
			// Everything old first (files before their folders, since creating a file touches its folder),
			// then the one under test made newer
			for _, rel := range append(all, "Disc 01", ".") {
				setTime(t, dir, rel, base)
			}
			setTime(t, dir, tt.newest, base.Add(3*time.Hour))

			newest, err := newestIn(dir)
			require.NoError(t, err)
			assert.True(t, tt.want.Equal(newest), "want %v, got %v", tt.want, newest)
		})
	}
}

func TestUpToDate(t *testing.T) {
	base := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	tests := []struct {
		name    string
		mirror  bool
		srcTime time.Time
		dstTime time.Time
		want    bool
	}{
		{"not mirrored yet", false, base, base, false},
		{"mirror newer", true, base, base.Add(time.Hour), true},
		{"same time", true, base, base, true},
		{"source changed since", true, base.Add(time.Hour), base, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			b := book{src: filepath.Join(root, "src"), dst: filepath.Join(root, "dst")}
			touch(t, b.src, "001 One.m4a")
			setTime(t, b.src, "001 One.m4a", tt.srcTime)
			setTime(t, b.src, ".", tt.srcTime)
			if tt.mirror {
				touch(t, b.dst, "Part 1.m4b")
				setTime(t, b.dst, ".", tt.dstTime)
			}

			done, err := upToDate(b)
			require.NoError(t, err)
			assert.Equal(t, tt.want, done)
		})
	}
}
