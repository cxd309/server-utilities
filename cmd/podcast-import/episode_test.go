package main

import (
	"context"
	"errors"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cxd309/server-utilities/internal/ffmpeg"
	"github.com/cxd309/server-utilities/internal/podcast"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileName(t *testing.T) {
	assert.Equal(t, "01 Fit the First.m4a", fileName(1, "Fit the First"))
	assert.Equal(t, "12 Don't Adventures Ever Have an End_.m4a", fileName(12, "Don't Adventures Ever Have an End_"))
	assert.Equal(t, "03 AC-DC Part 2.m4a", fileName(3, "AC/DC: Part 2"))
}

func TestCopyArgs(t *testing.T) {
	ep := episode{title: "T", show: "S", artist: "A", releaseDate: "1978-03-08T22:30:00Z"}
	tests := []struct {
		name     string
		src      source
		want     []string
		dontWant []string
	}{
		{"AAC with its own cover", source{path: "in.m4a", codec: "aac", cover: "in.m4a"},
			[]string{"-map 0:a:0 -map 0:v:0", "-c:a copy", "attached_pic"}, []string{"alac", "-i in.m4a -i"}},
		{"FLAC with a cover.jpg", source{path: "in.flac", codec: "flac", cover: "cover.jpg"},
			[]string{"-i cover.jpg -map 0:a:0 -map 1:v:0", "-c:a alac -sample_fmt s16p"}, []string{"-c:a copy"}},
		{"no cover", source{path: "in.flac", codec: "flac"},
			[]string{"-map 0:a:0"}, []string{"v:0", "attached_pic"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := strings.Join(copyArgs(tt.src, "out.m4a", ep, 2, 26), " ")
			for _, w := range tt.want {
				assert.Contains(t, args, w)
			}
			for _, d := range tt.dontWant {
				assert.NotContains(t, args, d)
			}
			assert.Contains(t, args, "track=2/26")
			assert.Contains(t, args, "media_type=21")
		})
	}
}

func TestMakeEpisode(t *testing.T) {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg not installed")
	}
	ctx := context.Background()
	dir := t.TempDir()

	// A FLAC "CD track" with a cover.jpg beside it
	in := filepath.Join(dir, "1-01 Track.flac")
	require.NoError(t, ffmpeg.Run(ctx, "-f", "lavfi", "-i", "sine=duration=1:sample_rate=44100", "-c:a", "flac", in))
	require.NoError(t, ffmpeg.Run(ctx, "-f", "lavfi", "-i", "color=c=red:s=64x64", "-frames:v", "1", filepath.Join(dir, "cover.jpg")))

	src, err := probeSource(ctx, in)
	require.NoError(t, err)
	assert.Equal(t, source{path: in, codec: "flac", cover: filepath.Join(dir, "cover.jpg")}, src)

	out := filepath.Join(dir, "01 Track.m4a")
	ep := episode{title: "Track", show: "Show", artist: "Artist", releaseDate: "1981-03-08T12:00:00Z"}
	err = makeEpisode(ctx, src, out, ep, 1, 1)
	if errors.Is(err, podcast.ErrNoMutagen) || errors.Is(err, podcast.ErrNoPython) {
		t.Skip("mutagen not installed")
	}
	require.NoError(t, err)

	info, err := ffmpeg.Probe(ctx, out)
	require.NoError(t, err)
	assert.Equal(t, "alac", info.Streams[0].CodecName)
	assert.Equal(t, "Show", info.Format.Tags["album"])
	assert.Equal(t, "21", info.Format.Tags["media_type"])
	var hasCover bool
	for _, s := range info.Streams {
		hasCover = hasCover || s.CodecType == "video"
	}
	assert.True(t, hasCover)
}
