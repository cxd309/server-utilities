package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/cxd309/server-utilities/internal/ffmpeg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// requireFFmpeg skips the test on machines without ffmpeg
func requireFFmpeg(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg not installed")
	}
}

// makeBook writes an AAC .m4b to path with one chapter per title,
// each one second long, and returns its chapters
func makeBook(t *testing.T, path string, titles ...string) (chapters []chapter) {
	t.Helper()
	meta := path + ".meta"
	var b strings.Builder
	b.WriteString(";FFMETADATA1\nalbum=Test Book\nartist=An Author\n")
	for i, title := range titles {
		fmt.Fprintf(&b, "[CHAPTER]\nTIMEBASE=1/1000\nSTART=%d\nEND=%d\ntitle=%s\n", i*1000, (i+1)*1000, title)
		chapters = append(chapters, chapter{start: float64(i), end: float64(i + 1), title: title})
	}
	require.NoError(t, os.WriteFile(meta, []byte(b.String()), 0o644))
	require.NoError(t, ffmpeg.Run(context.Background(),
		"-f", "lavfi", "-i", fmt.Sprintf("sine=duration=%d:sample_rate=44100", len(titles)), "-i", meta,
		"-map", "0", "-map_metadata", "1", "-map_chapters", "1",
		"-c:a", "aac", "-f", "ipod", path))
	return chapters
}

func TestSegmentArgs(t *testing.T) {
	chapters := []chapter{{start: 0}, {start: 28.500998}, {start: 1924.446984}}
	args := segmentArgs("in.m4b", "/tmp/out", chapters)

	assert.Contains(t, strings.Join(args, " "), "-segment_times 28.500998,1924.446984")
	assert.Equal(t, "/tmp/out/%04d.m4a", args[len(args)-1])
}

func TestSegmentArgsOneChapter(t *testing.T) {
	args := segmentArgs("in.m4b", "/tmp/out", []chapter{{start: 0}})
	assert.NotContains(t, args, "-segment_times")
	assert.Contains(t, strings.Join(args, " "), "-segment_time 1000000000", "no rule would mean a cut every 2 s")
}

func TestSplitAudio(t *testing.T) {
	requireFFmpeg(t)
	dir := t.TempDir()
	src := filepath.Join(dir, "book.m4b")
	chapters := makeBook(t, src, "One", "Two", "Three")
	out := filepath.Join(dir, "out")
	require.NoError(t, os.Mkdir(out, 0o755))

	require.NoError(t, splitAudio(context.Background(), src, out, chapters))

	for i := range 3 {
		info, err := ffmpeg.Probe(context.Background(), filepath.Join(out, fmt.Sprintf("%04d.m4a", i)))
		require.NoError(t, err)
		d, err := strconv.ParseFloat(info.Format.Duration, 64)
		require.NoError(t, err)
		assert.InDelta(t, 1.0, d, 0.05, "piece %d is about one second", i)
		assert.Empty(t, info.Chapters, "pieces carry no chapters")
	}
	assert.NoFileExists(t, filepath.Join(out, "0003.m4a"))
}

func TestSplitAudioOneChapter(t *testing.T) {
	requireFFmpeg(t)
	dir := t.TempDir()
	src := filepath.Join(dir, "book.m4b")
	// 5 s, longer than the segment muxer's 2 s default, which a 1 s book would hide
	require.NoError(t, ffmpeg.Run(context.Background(),
		"-f", "lavfi", "-i", "sine=duration=5:sample_rate=44100", "-c:a", "aac", "-f", "ipod", src))
	chapters := []chapter{{start: 0, end: 5, title: "Only"}}

	require.NoError(t, splitAudio(context.Background(), src, dir, chapters))

	assert.NoFileExists(t, filepath.Join(dir, "0001.m4a"), "one chapter, one piece")
	info, err := ffmpeg.Probe(context.Background(), filepath.Join(dir, "0000.m4a"))
	require.NoError(t, err)
	d, err := strconv.ParseFloat(info.Format.Duration, 64)
	require.NoError(t, err)
	assert.InDelta(t, 5.0, d, 0.05, "the whole book")
}
