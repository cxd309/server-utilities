package podcast

import (
	"context"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cxd309/server-utilities/internal/ffmpeg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// requireTools skips the test on machines without ffmpeg, Python or mutagen
func requireTools(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg not installed")
	}
	if err := exec.Command("python3", "-c", "import mutagen").Run(); err != nil {
		t.Skip("python3 with mutagen not installed")
	}
}

// hasFlag reads the podcast flag back with mutagen
func hasFlag(t *testing.T, path string) (flagged bool) {
	t.Helper()
	out, err := exec.Command("python3", "-c",
		"import sys; from mutagen.mp4 import MP4; print(bool(MP4(sys.argv[1]).tags.get('pcst')))", path).Output()
	require.NoError(t, err)
	return strings.TrimSpace(string(out)) == "True"
}

func TestFlag(t *testing.T) {
	requireTools(t)
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "episode.m4a.part") // the scripts flag their .part files
	require.NoError(t, ffmpeg.Run(ctx, "-f", "lavfi", "-i", "sine=duration=1",
		"-c:a", "aac", "-metadata", "title=Gdansk", "-metadata", "media_type=21", "-f", "ipod", path))
	require.False(t, hasFlag(t, path), "ffmpeg doesn't write the flag")

	require.NoError(t, Flag(ctx, path))

	assert.True(t, hasFlag(t, path))
}

func TestFlagNothing(t *testing.T) {
	assert.NoError(t, Flag(context.Background()))
}

func TestFlagNoPython(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	assert.ErrorIs(t, Flag(context.Background(), "x.m4a"), ErrNoPython)
}

func TestFlagBadFile(t *testing.T) {
	requireTools(t)
	err := Flag(context.Background(), filepath.Join(t.TempDir(), "missing.m4a"))
	var pe *Error
	require.ErrorAs(t, err, &pe)
	assert.NotEmpty(t, pe.Stderr)
}
