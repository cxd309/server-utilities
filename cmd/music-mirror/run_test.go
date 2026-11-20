package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/cxd309/server-utilities/internal/ffmpeg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProcessEncode(t *testing.T) {
	requireFFmpeg(t)
	dir := t.TempDir()
	src := filepath.Join(dir, "lib", "song.flac")
	require.NoError(t, os.MkdirAll(filepath.Dir(src), 0o755))
	makeFLAC(t, src, "TITLE=Song", "ORIGINALDATE=1968")
	dst := filepath.Join(dir, "mirror", "Artist", "Album", "song.m4a")

	require.NoError(t, process(context.Background(), job{src: src, dst: dst, kind: kindEncode}))

	info, err := ffmpeg.Probe(context.Background(), dst)
	require.NoError(t, err)
	require.NotEmpty(t, info.Streams)
	assert.Equal(t, "alac", info.Streams[0].CodecName)
	assert.Equal(t, "Song", info.Format.Tags["title"])
	assert.Equal(t, "1968", info.Format.Tags["date"])
	assert.NoFileExists(t, dst+".part")
}

func TestProcessCopy(t *testing.T) {
	requireFFmpeg(t)
	dir := t.TempDir()
	src := makeMP3(t, dir)
	original, err := os.ReadFile(src)
	require.NoError(t, err)
	dst := filepath.Join(dir, "mirror", "song.mp3")

	require.NoError(t, process(context.Background(), job{src: src, dst: dst, kind: kindCopy}))

	got, err := readMP3(dst)
	require.NoError(t, err)
	assert.Equal(t, "1967", got.Year, "copy is re-tagged with the original year")

	after, err := os.ReadFile(src)
	require.NoError(t, err)
	assert.Equal(t, original, after, "source MP3 must not change")
}

func TestProcessFailureCleansUp(t *testing.T) {
	dir := t.TempDir()
	touch(t, dir, "broken.flac") // empty, so tag reading fails
	dst := filepath.Join(dir, "mirror", "broken.m4a")

	err := process(context.Background(), job{src: filepath.Join(dir, "broken.flac"), dst: dst, kind: kindEncode})

	assert.Error(t, err)
	assert.NoFileExists(t, dst)
	assert.NoFileExists(t, dst+".part")
}

func TestProcessCancelled(t *testing.T) {
	requireFFmpeg(t)
	dir := t.TempDir()
	src := makeTone(t, dir)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := process(ctx, job{src: src, dst: filepath.Join(dir, "out.m4a"), kind: kindEncode})

	assert.ErrorIs(t, err, context.Canceled)
	assert.NoFileExists(t, filepath.Join(dir, "out.m4a"))
}

func TestRunJobs(t *testing.T) {
	requireFFmpeg(t)
	dir := t.TempDir()
	var jobs []job
	for _, name := range []string{"a", "b", "c", "d"} {
		src := filepath.Join(dir, name+".flac")
		makeFLAC(t, src)
		jobs = append(jobs, job{src: src, dst: filepath.Join(dir, "mirror", name+".m4a"), kind: kindEncode})
	}
	touch(t, dir, "bad.flac")
	jobs = append(jobs, job{src: filepath.Join(dir, "bad.flac"), dst: filepath.Join(dir, "mirror", "bad.m4a"), kind: kindEncode})

	var ok, failed int
	runJobs(context.Background(), jobs, 3, func(j job, err error) {
		if err != nil {
			failed++
		} else {
			ok++
		}
	})

	assert.Equal(t, 4, ok)
	assert.Equal(t, 1, failed)
}
