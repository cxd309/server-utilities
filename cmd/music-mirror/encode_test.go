package main

import (
	"context"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/cxd309/server-utilities/internal/ffmpeg"
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

// makeFLAC writes a one-second 96 kHz test tone to path,
// with optional Vorbis comments given as "KEY=value".
func makeFLAC(t *testing.T, path string, comments ...string) {
	t.Helper()
	args := []string{"-nostdin", "-v", "error", "-y",
		"-f", "lavfi", "-i", "sine=duration=1:sample_rate=96000"}
	for _, c := range comments {
		args = append(args, "-metadata", c)
	}
	out, err := exec.Command("ffmpeg", append(args, path)...).CombinedOutput()
	require.NoError(t, err, string(out))
}

func TestEncode(t *testing.T) {
	requireFFmpeg(t)
	require.NoError(t, checkFFmpeg())

	dir := t.TempDir()
	src := filepath.Join(dir, "tone.flac") // 96 kHz, so this also checks the resample
	out := filepath.Join(dir, "tone.m4a.part")
	makeFLAC(t, src)

	tags := trackTags{
		Title: "Song (feat. Rihanna)", Artist: "Calvin Harris", Album: "Motion",
		Year: "2014", Track: "3/12", Disc: "1/2", Compilation: true,
	}
	require.NoError(t, encode(context.Background(), src, out, tags))

	info, err := ffmpeg.Probe(context.Background(), out)
	require.NoError(t, err)
	require.Len(t, info.Streams, 1, "cover art and other streams are dropped")
	assert.Equal(t, "alac", info.Streams[0].CodecName)
	assert.Equal(t, "44100", info.Streams[0].SampleRate)
	assert.Equal(t, "Song (feat. Rihanna)", info.Format.Tags["title"])
	assert.Equal(t, "Calvin Harris", info.Format.Tags["artist"])
	assert.Equal(t, "3/12", info.Format.Tags["track"])
	assert.Equal(t, "1/2", info.Format.Tags["disc"])
	assert.Equal(t, "1", info.Format.Tags["compilation"])
}

func TestEncodeFailure(t *testing.T) {
	requireFFmpeg(t)
	dir := t.TempDir()

	err := encode(context.Background(), filepath.Join(dir, "missing.flac"), filepath.Join(dir, "x.part"), trackTags{})

	var fe *ffmpeg.Error
	require.ErrorAs(t, err, &fe)
	assert.Contains(t, fe.Stderr, "missing.flac")
}

func TestMetadataArgs(t *testing.T) {
	got := metadataArgs(trackTags{Title: "Song", Track: "1/10", Compilation: true})
	assert.Equal(t, []string{
		"-metadata", "title=Song",
		"-metadata", "track=1/10",
		"-metadata", "compilation=1",
	}, got, "empty fields are skipped")
}

// makeTone writes a short untagged FLAC into dir and returns its path.
func makeTone(t *testing.T, dir string) (path string) {
	t.Helper()
	path = filepath.Join(dir, "tone.flac")
	makeFLAC(t, path)
	return path
}

// makePlainMP3 writes a one-second untagged MP3 to path. encode makes ALAC
// now, so the MP3 tests need their own way to get a real MP3.
func makePlainMP3(t *testing.T, path string) {
	t.Helper()
	out, err := exec.Command("ffmpeg", "-nostdin", "-v", "error", "-y",
		"-f", "lavfi", "-i", "sine=duration=1",
		"-c:a", "libmp3lame", "-b:a", "128k",
		"-id3v2_version", "0", "-write_id3v1", "0",
		path).CombinedOutput()
	require.NoError(t, err, string(out))
}
