package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cxd309/server-utilities/internal/ffmpeg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCoverArgs(t *testing.T) {
	tests := []struct {
		codec string
		want  string
	}{
		{"mjpeg", "-c:v copy"},
		{"png", "-c:v mjpeg"},
	}
	for _, tt := range tests {
		t.Run(tt.codec, func(t *testing.T) {
			args := coverArgs("in.m4b", "/out/cover.jpg", tt.codec)
			assert.Contains(t, strings.Join(args, " "), tt.want)
			assert.Equal(t, "/out/cover.jpg", args[len(args)-1])
		})
	}
}

func TestSplitBookSavesCover(t *testing.T) {
	requireFFmpeg(t)
	ctx := context.Background()
	src, b, out := testBook(t)

	// Embed a picture in the test book, the way Audible does
	dir := filepath.Dir(src)
	picture := filepath.Join(dir, "picture.jpg")
	require.NoError(t, ffmpeg.Run(ctx, "-f", "lavfi", "-i", "color=c=red:s=64x64", "-frames:v", "1", picture))
	withCover := filepath.Join(dir, "with-cover.m4b")
	require.NoError(t, ffmpeg.Run(ctx, "-i", src, "-i", picture, "-map", "0:a", "-map", "1",
		"-c", "copy", "-disposition:v", "attached_pic", "-f", "ipod", withCover))
	b.cover = "mjpeg"

	require.NoError(t, splitBook(ctx, withCover, out, b))

	assert.Contains(t, names(t, out), coverName)
	want, err := os.ReadFile(picture)
	require.NoError(t, err)
	got, err := os.ReadFile(filepath.Join(out, coverName))
	require.NoError(t, err)
	assert.Equal(t, want, got, "a JPEG is copied byte for byte")
}
