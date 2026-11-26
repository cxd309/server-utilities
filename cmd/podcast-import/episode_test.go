package main

import (
	"context"
	"errors"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/cxd309/server-utilities/internal/ffmpeg"
	"github.com/cxd309/server-utilities/internal/podcast"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileName(t *testing.T) {
	assert.Equal(t, "01 Fit the First.m4a", fileName(1, "Fit the First"))
	assert.Equal(t, "24 Don't Adventures Ever Have an End.m4a", fileName(24, "Don't Adventures Ever Have an End?"))
	assert.Equal(t, "05 Who What.m4a", fileName(5, `Who* "What"`))
	assert.Equal(t, "06 A-B.m4a", fileName(6, `A\B`))
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

func TestCopyArgsCut(t *testing.T) {
	src := source{path: "in.m4a", codec: "aac"}
	ep := episode{title: "T", start: 2355.04, end: 4543.94}
	args := strings.Join(copyArgs(src, "out.m4a", ep, 1, 1), " ")
	assert.True(t, strings.HasPrefix(args, "-ss 2355.040 -i in.m4a"), args)
	assert.Contains(t, args, "-t 2188.900 ")

	whole := strings.Join(copyArgs(src, "out.m4a", episode{title: "T"}, 1, 1), " ")
	assert.NotContains(t, whole, "-ss")
	assert.NotContains(t, whole, "-t ")
}

func TestMakeEpisodeCut(t *testing.T) {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg not installed")
	}
	ctx := context.Background()
	dir := t.TempDir()
	in := filepath.Join(dir, "book.m4a")
	require.NoError(t, ffmpeg.Run(ctx, "-f", "lavfi", "-i", "sine=duration=6:sample_rate=22050", "-c:a", "aac", "-f", "ipod", in))
	src, err := probeSource(ctx, in)
	require.NoError(t, err)

	out := filepath.Join(dir, "02 Two.m4a")
	ep := episode{title: "Two", show: "S", artist: "A", releaseDate: "2018-03-15T18:30:00Z", start: 2, end: 4}
	err = makeEpisode(ctx, src, out, ep, 2, 3)
	if errors.Is(err, podcast.ErrNoMutagen) || errors.Is(err, podcast.ErrNoPython) {
		t.Skip("mutagen not installed")
	}
	require.NoError(t, err)
	info, err := ffmpeg.Probe(ctx, out)
	require.NoError(t, err)
	seconds, err := strconv.ParseFloat(info.Format.Duration, 64)
	require.NoError(t, err)
	assert.InDelta(t, 2.0, seconds, 0.1, "two seconds, give or take a packet")
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
