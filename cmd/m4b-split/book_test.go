package main

import (
	"testing"

	"github.com/cxd309/server-utilities/internal/ffmpeg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// earthsea is shaped like the Libation download, trimmed to a few chapters
var earthsea = ffmpeg.Info{
	Format: ffmpeg.Format{Tags: map[string]string{
		"title":        "A Wizard of Earthsea",
		"album":        "A Wizard of Earthsea: The First Book of Earthsea (Unabridged)",
		"artist":       "Ursula K. Le Guin",
		"album_artist": "Ursula K. Le Guin",
		"composer":     "Kobna Holdbrook-Smith",
		"date":         "2018",
	}},
	Streams: []ffmpeg.Stream{
		{CodecType: "audio", CodecName: "aac", Profile: "LC", SampleRate: "44100", Channels: 2},
		{CodecType: "data", CodecName: "bin_data"},
		{CodecType: "video", CodecName: "mjpeg"},
	},
	Chapters: []ffmpeg.Chapter{
		{StartTime: "0.000000", EndTime: "28.500998", Tags: map[string]string{"title": "Opening Credits"}},
		{StartTime: "28.500998", EndTime: "1924.446984", Tags: map[string]string{"title": "Chapter 1. Warriors in the Mist"}},
		{StartTime: "1924.446984", EndTime: "1924.446984", Tags: map[string]string{"title": "Empty"}},
		{StartTime: "1924.446984", EndTime: "4183.427982", Tags: map[string]string{"title": "Chapter  2. The Shadow"}},
		{StartTime: "4183.427982", EndTime: "4200.000000"},
	},
}

func TestBookFromInfo(t *testing.T) {
	b, err := bookFromInfo(earthsea, "fallback")
	require.NoError(t, err)

	assert.Equal(t, book{
		title:    "A Wizard of Earthsea: The First Book of Earthsea (Unabridged)",
		author:   "Ursula K. Le Guin",
		narrator: "Kobna Holdbrook-Smith",
		date:     "2018",
		chapters: []chapter{
			{start: 0, end: 28.500998, title: "Opening Credits"},
			{start: 28.500998, end: 1924.446984, title: "Chapter 1. Warriors in the Mist"},
			{start: 1924.446984, end: 4183.427982, title: "Chapter 2. The Shadow"}, // double space tidied, empty chapter dropped
			{start: 4183.427982, end: 4200, title: "Chapter 4"},                    // untitled: numbered by position
		},
	}, b)
}
func TestBookFromInfoFallbackTitle(t *testing.T) {
	info := ffmpeg.Info{
		Streams:  []ffmpeg.Stream{{CodecType: "audio"}},
		Chapters: []ffmpeg.Chapter{{StartTime: "0", EndTime: "10"}},
	}
	b, err := bookFromInfo(info, "Book From Filename")
	require.NoError(t, err)
	assert.Equal(t, "Book From Filename", b.title)
}

func TestBookFromInfoErrors(t *testing.T) {
	tests := []struct {
		name string
		info ffmpeg.Info
		want error
	}{
		{"no audio", ffmpeg.Info{
			Streams:  []ffmpeg.Stream{{CodecType: "video"}},
			Chapters: earthsea.Chapters,
		}, errNoAudio},
		{"no chapters", ffmpeg.Info{
			Streams: earthsea.Streams,
		}, errNoChapters},
		{"only empty chapters", ffmpeg.Info{
			Streams:  earthsea.Streams,
			Chapters: []ffmpeg.Chapter{{StartTime: "5", EndTime: "5"}},
		}, errNoChapters},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := bookFromInfo(tt.info, "x")
			assert.ErrorIs(t, err, tt.want)
		})
	}
}

func TestBookFromInfoBadTime(t *testing.T) {
	info := ffmpeg.Info{
		Streams:  earthsea.Streams,
		Chapters: []ffmpeg.Chapter{{StartTime: "soon", EndTime: "10"}},
	}
	_, err := bookFromInfo(info, "x")
	assert.ErrorContains(t, err, "chapter 1 start")
}
