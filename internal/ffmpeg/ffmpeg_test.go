package ffmpeg

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// requireFFmpeg skips the test on machines without ffmpeg.
func requireFFmpeg(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg not installed")
	}
}

func TestRunNotInstalled(t *testing.T) {
	t.Setenv("PATH", t.TempDir()) // a PATH with nothing on it

	err := Run(context.Background(), "-version")

	assert.ErrorIs(t, err, ErrNotInstalled)
	assert.EqualError(t, err, "ffmpeg: not found on PATH")
}

func TestRunFailure(t *testing.T) {
	requireFFmpeg(t)
	missing := filepath.Join(t.TempDir(), "missing.flac")

	err := Run(context.Background(), "-i", missing, "-f", "null", "-")

	var fe *Error
	require.ErrorAs(t, err, &fe)
	assert.Equal(t, "ffmpeg", fe.Tool)
	assert.Contains(t, fe.Stderr, "missing.flac")
}

func TestProbe(t *testing.T) {
	requireFFmpeg(t)
	dir := t.TempDir()
	meta := filepath.Join(dir, "meta.txt")
	book := filepath.Join(dir, "book.m4b")

	// ffmpeg's own metadata file format: global tags, then one block per chapter.
	var b strings.Builder
	b.WriteString(";FFMETADATA1\ntitle=Test Book\nartist=An Author\n")
	for i := range 3 {
		fmt.Fprintf(&b, "[CHAPTER]\nTIMEBASE=1/1000\nSTART=%d\nEND=%d\ntitle=Chapter %d\n", i*1000, (i+1)*1000, i+1)
	}
	require.NoError(t, os.WriteFile(meta, []byte(b.String()), 0o644))

	require.NoError(t, Run(context.Background(),
		"-f", "lavfi", "-i", "sine=duration=3:sample_rate=22050", "-i", meta,
		"-map", "0", "-map_metadata", "1", "-map_chapters", "1",
		"-c:a", "aac", "-f", "ipod", book))

	info, err := Probe(context.Background(), book)
	require.NoError(t, err)

	assert.Equal(t, "Test Book", info.Format.Tags["title"])
	require.NotEmpty(t, info.Streams)
	assert.Equal(t, "audio", info.Streams[0].CodecType)
	assert.Equal(t, "aac", info.Streams[0].CodecName)
	assert.Equal(t, "22050", info.Streams[0].SampleRate)
	require.Len(t, info.Chapters, 3)
	assert.Equal(t, "Chapter 2", info.Chapters[1].Tags["title"])
	assert.Equal(t, "1.000000", info.Chapters[1].StartTime)
}

func TestProbeMissingFile(t *testing.T) {
	requireFFmpeg(t)
	_, err := Probe(context.Background(), filepath.Join(t.TempDir(), "missing.m4b"))

	var fe *Error
	require.ErrorAs(t, err, &fe)
	assert.Equal(t, "ffprobe", fe.Tool)
}
