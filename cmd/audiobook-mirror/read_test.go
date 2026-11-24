package main

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cxd309/server-utilities/internal/ffmpeg"
)

func TestListChapters(t *testing.T) {
	dir := t.TempDir()
	for _, rel := range []string{
		"002 Two.m4a",
		"001 One.M4A",
		"cover.jpg",
		".001 One.m4a.part",
		"._001 One.m4a",
		"Extras/003 Three.m4a",
		"1-01 Track.flac",
	} {
		touch(t, dir, rel)
	}

	paths, err := listChapters(dir)
	require.NoError(t, err)
	assert.Equal(t, []string{
		filepath.Join(dir, "001 One.M4A"),
		filepath.Join(dir, "002 Two.m4a"),
		filepath.Join(dir, "1-01 Track.flac"),
	}, paths)
}

func TestChapterFromInfo(t *testing.T) {
	audio := ffmpeg.Stream{CodecType: "audio", CodecName: "aac", SampleRate: "44100"}
	cover := ffmpeg.Stream{CodecType: "video"}

	tests := []struct {
		name     string
		info     ffmpeg.Info
		want     chapterFile
		wantRate int
	}{
		{
			"tagged",
			ffmpeg.Info{
				Format:  ffmpeg.Format{Duration: "7200.000000", Tags: map[string]string{"title": " Chapter 1 "}},
				Streams: []ffmpeg.Stream{cover, audio},
			},
			chapterFile{path: "/b/001 x.m4a", title: "Chapter 1", frames: 310_079, codec: "aac"},
			44100,
		},
		{
			"no title tag",
			ffmpeg.Info{
				Format:  ffmpeg.Format{Duration: "1.0"},
				Streams: []ffmpeg.Stream{{CodecType: "audio", CodecName: "flac", SampleRate: "22050"}},
			},
			chapterFile{path: "/b/001 x.m4a", title: "001 x", frames: 22, codec: "flac"},
			22050,
		},
		{
			"named medium",
			ffmpeg.Info{
				Format:  ffmpeg.Format{Duration: "1.0", Tags: map[string]string{"TITLE": "Chapter 1", "disc": "2/11", "DISCSUBTITLE": "Animal Farm"}},
				Streams: []ffmpeg.Stream{{CodecType: "audio", CodecName: "flac", SampleRate: "44100"}},
			},
			chapterFile{path: "/b/001 x.m4a", title: "Chapter 1", frames: 44, codec: "flac", section: "Animal Farm"},
			44100,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, rate, err := chapterFromInfo("/b/001 x.m4a", tt.info)
			require.NoError(t, err)
			assert.Equal(t, tt.want, c)
			assert.Equal(t, tt.wantRate, rate)
		})
	}
}

func TestChapterFromInfoErrors(t *testing.T) {
	_, _, err := chapterFromInfo("/b/x.m4a", ffmpeg.Info{Streams: []ffmpeg.Stream{{CodecType: "video"}}})
	assert.ErrorIs(t, err, errNoAudio)

	_, _, err = chapterFromInfo("/b/x.m4a", ffmpeg.Info{
		Format:  ffmpeg.Format{Duration: "N/A"},
		Streams: []ffmpeg.Stream{{CodecType: "audio", SampleRate: "44100"}},
	})
	var fe *probeFieldError
	require.ErrorAs(t, err, &fe)
	assert.Equal(t, "duration", fe.Field)
	assert.ErrorIs(t, err, strconv.ErrSyntax)
}

func TestListItem(t *testing.T) {
	tests := []struct {
		name string
		list string
		i    int
		want string
	}{
		{"first", "Terry Pratchett;Tony Robinson", 0, "Terry Pratchett"},
		{"second", "Terry Pratchett; Tony Robinson", 1, "Tony Robinson"},
		{"past the end", "Terry Pratchett", 1, ""},
		{"empty", "", 0, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, listItem(tt.list, tt.i))
		})
	}
}

func TestTagsFromInfo(t *testing.T) {
	tests := []struct {
		name string
		tags map[string]string
		want bookTags
	}{
		{
			"Picard tags",
			map[string]string{
				"album":        "Mort",
				"album_artist": "Terry Pratchett read by Tony Robinson",
				"artist":       "Terry Pratchett",
				"composer":     "Tony Robinson",
				"date":         "2004",
			},
			bookTags{title: "Mort", author: "Terry Pratchett read by Tony Robinson", narrator: "Tony Robinson", date: "2004"},
		},
		{
			"no album artist",
			map[string]string{"album": "Thud!", "artist": "Terry Pratchett"},
			bookTags{title: "Thud!", author: "Terry Pratchett"},
		},
		{
			"CD rip with upper case keys",
			map[string]string{
				"ALBUM":        "Eric",
				"ALBUMARTISTS": "Terry Pratchett;Tony Robinson",
				"album_artist": "Terry Pratchett read by Tony Robinson",
				"DATE":         "2006",
			},
			bookTags{title: "Eric", author: "Terry Pratchett", narrator: "Tony Robinson", date: "2006"},
		},
		{"no tags", nil, bookTags{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tagsFromInfo(ffmpeg.Info{Format: ffmpeg.Format{Tags: tt.tags}}))
		})
	}
}

// makeChapter encodes a short tagged AAC chapter file at dir/name
func makeChapter(t *testing.T, dir, name string, seconds, sampleRate int, title string) {
	t.Helper()
	require.NoError(t, ffmpeg.Run(context.Background(),
		"-f", "lavfi", "-i", fmt.Sprintf("sine=duration=%d:sample_rate=%d", seconds, sampleRate),
		"-c:a", "aac",
		"-metadata", "title="+title,
		"-metadata", "album=Test Book",
		"-metadata", "album_artist=An Author read by A Reader",
		"-f", "ipod", filepath.Join(dir, name)))
}

func TestReadDownload(t *testing.T) {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg not installed")
	}
	ctx := context.Background()

	t.Run("book", func(t *testing.T) {
		dir := t.TempDir()
		makeChapter(t, dir, "001 One.m4a", 1, 44100, "One")
		makeChapter(t, dir, "002 Two.m4a", 2, 44100, "Two")
		touch(t, dir, "cover.jpg")

		d, err := readDownload(ctx, dir)
		require.NoError(t, err)
		assert.Equal(t, 44100, d.sampleRate)
		assert.Equal(t, bookTags{title: "Test Book", author: "An Author read by A Reader"}, d.tags)
		assert.Equal(t, filepath.Join(dir, "cover.jpg"), d.cover)
		require.Len(t, d.chapters, 2)
		assert.Equal(t, "One", d.chapters[0].title)
		assert.Equal(t, "Two", d.chapters[1].title)
	})

	t.Run("no cover.jpg: the first chapter's embedded picture", func(t *testing.T) {
		dir := t.TempDir()
		jpg := filepath.Join(t.TempDir(), "art.jpg")
		require.NoError(t, ffmpeg.Run(ctx, "-f", "lavfi", "-i", "color=c=red:s=100x100", "-frames:v", "1", jpg))
		require.NoError(t, ffmpeg.Run(ctx,
			"-f", "lavfi", "-i", "sine=duration=1:sample_rate=44100", "-i", jpg,
			"-map", "0", "-map", "1", "-c:a", "aac", "-c:v", "copy", "-disposition:v", "attached_pic",
			"-f", "ipod", filepath.Join(dir, "001 One.m4a")))
		makeChapter(t, dir, "002 Two.m4a", 1, 44100, "Two")

		d, err := readDownload(ctx, dir)
		require.NoError(t, err)
		assert.Equal(t, filepath.Join(dir, "001 One.m4a"), d.cover)
	})

	t.Run("no cover at all", func(t *testing.T) {
		dir := t.TempDir()
		makeChapter(t, dir, "001 One.m4a", 1, 44100, "One")

		d, err := readDownload(ctx, dir)
		require.NoError(t, err)
		assert.Empty(t, d.cover)
	})

	t.Run("no chapters", func(t *testing.T) {
		dir := t.TempDir()
		touch(t, dir, "cover.jpg")
		_, err := readDownload(ctx, dir)
		assert.ErrorIs(t, err, errNoChapters)
	})

	t.Run("mixed sample rates", func(t *testing.T) {
		dir := t.TempDir()
		makeChapter(t, dir, "001 One.m4a", 1, 44100, "One")
		makeChapter(t, dir, "002 Two.m4a", 1, 22050, "Two")

		_, err := readDownload(ctx, dir)
		var se *sampleRateError
		require.ErrorAs(t, err, &se)
		assert.Equal(t, 44100, se.Want)
		assert.Equal(t, 22050, se.Got)
	})
}
