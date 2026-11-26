package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFramesFor(t *testing.T) {
	tests := []struct {
		name       string
		seconds    float64
		sampleRate int
		want       int64
	}{
		{"zero", 0, 44100, 0},
		{"exactly one frame", 1, 1024, 1},
		{"part frame rounds up", 1.5, 1024, 2},
		{"2 h at 44.1 kHz", 2 * 3600, 44100, 310_079},
		{"Earthsea, 7 h at 22 kHz", 7 * 3600, 22050, 542_637},
		{"Hexagonal Phase chapter, 220 min", 220 * 60, 44100, 568_477},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, framesFor(tt.seconds, tt.sampleRate))
		})
	}
}

func TestCutChapter(t *testing.T) {
	tests := []struct {
		name   string
		frames int64
		limit  int64
		want   []piece
	}{
		{"fits", 10, 10, []piece{{3, 0, 10}}},
		{"empty", 0, 10, []piece{{3, 0, 0}}},
		{"one over", 11, 10, []piece{{3, 0, 6}, {3, 6, 11}}},
		{"exactly two", 20, 10, []piece{{3, 0, 10}, {3, 10, 20}}},
		{"three equal", 21, 10, []piece{{3, 0, 7}, {3, 7, 14}, {3, 14, 21}}},
		{"uneven", 25, 10, []piece{{3, 0, 9}, {3, 9, 17}, {3, 17, 25}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, cutChapter(3, tt.frames, tt.limit))
		})
	}
}

func TestPlanParts(t *testing.T) {
	tests := []struct {
		name   string
		frames []int64
		want   [][]piece
	}{
		{"no chapters", nil, nil},
		{"one chapter", []int64{5}, [][]piece{{{0, 0, 5}}}},
		{"all in one part", []int64{3, 4, 3}, [][]piece{{{0, 0, 3}, {1, 0, 4}, {2, 0, 3}}}},
		{"evened out, not 5+5 then 1", []int64{5, 5, 1}, [][]piece{{{0, 0, 5}}, {{1, 0, 5}, {2, 0, 1}}}},
		{"two halves, not 10 then 2", []int64{2, 2, 2, 2, 2, 2}, [][]piece{
			{{0, 0, 2}, {1, 0, 2}, {2, 0, 2}},
			{{3, 0, 2}, {4, 0, 2}, {5, 0, 2}},
		}},
		{"next won't fit", []int64{6, 5, 4}, [][]piece{{{0, 0, 6}}, {{1, 0, 5}, {2, 0, 4}}}},
		{"long chapter alone", []int64{25}, [][]piece{{{0, 0, 9}}, {{0, 9, 17}}, {{0, 17, 25}}}},
		{"long chapter after a short one", []int64{3, 11}, [][]piece{{{0, 0, 3}, {1, 0, 6}}, {{1, 6, 11}}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parts, err := planParts(tt.frames, 10)
			require.NoError(t, err)
			assert.Equal(t, tt.want, parts)
		})
	}
}

func TestPlanPartsBadLimit(t *testing.T) {
	for _, limit := range []int64{0, -1} {
		_, err := planParts([]int64{5}, limit)
		assert.ErrorIs(t, err, errBadLimit)
	}
}

func TestPartLimit(t *testing.T) {
	tests := []struct {
		name          string
		bytesPerFrame float64
		want          int64
	}{
		// (2³¹ - 1) / 1024 = 2,097,151 frames: about 13.5 h at 44.1 kHz
		{"size unknown, sample limit", 0, 2_097_151},
		// Audible AAC at ~364 bytes per frame: 3.9 GB allows ~10.7M frames
		{"AAC, sample limit wins", 364, 2_097_151},
		// spoken-word ALAC at ~900 bytes per frame: 3.9 GB allows ~4.3M frames
		{"speech ALAC, sample limit wins", 900, 2_097_151},
		// uncompressed PCM at 4,096 bytes per frame: 3.9 GB / 4096 = 952,148 frames, ~6.2 h
		{"PCM-sized ALAC, size limit wins", 4096, 952_148},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, partLimit(tt.bytesPerFrame))
		})
	}
}

func TestPlanBook(t *testing.T) {
	tests := []struct {
		name     string
		chapters []chapterFile
		want     [][]piece
	}{
		{
			"no sections: packed by the limit",
			[]chapterFile{{frames: 30}, {frames: 30}, {frames: 10}},
			[][]piece{{{0, 0, 30}}, {{1, 0, 30}, {2, 0, 10}}},
		},
		{
			"sections never share a part, even when they'd fit",
			[]chapterFile{{frames: 10, section: "1984"}, {frames: 10, section: "1984"}, {frames: 10, section: "Animal Farm"}},
			[][]piece{{{0, 0, 10}, {1, 0, 10}}, {{2, 0, 10}}},
		},
		{
			"a long section still splits by the limit",
			[]chapterFile{{frames: 10, section: "Essays"}, {frames: 40, section: "Poetry"}, {frames: 40, section: "Poetry"}},
			[][]piece{{{0, 0, 10}}, {{1, 0, 40}}, {{2, 0, 40}}},
		},
		{
			"unnamed chapters around a section each form their own run",
			[]chapterFile{{frames: 10}, {frames: 10, section: "Appendix"}, {frames: 10}},
			[][]piece{{{0, 0, 10}}, {{1, 0, 10}}, {{2, 0, 10}}},
		},
		{"no chapters", nil, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parts, err := planBook(tt.chapters, 50)
			require.NoError(t, err)
			assert.Equal(t, tt.want, parts)
		})
	}
}

func TestOutputBytesPerFrame(t *testing.T) {
	tests := []struct {
		name string
		d    download
		want float64
	}{
		{"copied AAC keeps its size", download{sampleRate: 44100, chapters: []chapterFile{
			{codec: "aac", frames: 600, bytes: 200_000},
			{codec: "aac", frames: 400, bytes: 164_000},
		}}, 364},
		{"FLAC to ALAC is a bit bigger", download{sampleRate: 44100, chapters: []chapterFile{
			{codec: "flac", frames: 1000, bytes: 2_000_000},
		}}, 2200},
		{"22 kHz AAC is copied too", download{sampleRate: 22050, chapters: []chapterFile{
			{codec: "aac", frames: 1000, bytes: 180_000},
		}}, 180},
		{"MP3 to ALAC counts as PCM", download{sampleRate: 44100, chapters: []chapterFile{
			{codec: "mp3", frames: 1000, bytes: 400_000},
		}}, 4096},
		{"no chapters is unknown", download{sampleRate: 44100}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.InDelta(t, tt.want, outputBytesPerFrame(tt.d), 0.001)
		})
	}
}
