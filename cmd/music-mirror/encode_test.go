package main

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
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

// makeFLAC writes a one-second 96 kHz test tone to path.
func makeFLAC(t *testing.T, path string) {
	t.Helper()
	out, err := exec.Command("ffmpeg", "-nostdin", "-v", "error", "-y",
		"-f", "lavfi", "-i", "sine=duration=1:sample_rate=96000", path).CombinedOutput()
	require.NoError(t, err, string(out))
}

func TestEncode(t *testing.T) {
	requireFFmpeg(t)
	require.NoError(t, checkFFmpeg())

	dir := t.TempDir()
	src := filepath.Join(dir, "tone.flac")
	out := filepath.Join(dir, "tone.mp3.part")
	makeFLAC(t, src)

	require.NoError(t, encode(context.Background(), src, out, "320k"))

	data, err := os.ReadFile(out)
	require.NoError(t, err)
	require.Greater(t, len(data), 3)
	assert.NotEqual(t, "ID3", string(data[:3]), "ffmpeg should not write its own tag")
}

func TestEncodeFailure(t *testing.T) {
	requireFFmpeg(t)
	dir := t.TempDir()

	err := encode(context.Background(), filepath.Join(dir, "missing.flac"), filepath.Join(dir, "x.part"), "320k")

	var fe *ffmpegError
	require.ErrorAs(t, err, &fe)
	assert.Contains(t, fe.Stderr, "missing.flac")
}
