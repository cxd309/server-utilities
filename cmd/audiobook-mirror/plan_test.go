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
		{"exactly full", []int64{5, 5, 1}, [][]piece{{{0, 0, 5}, {1, 0, 5}}, {{2, 0, 1}}}},
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
		sampleRate    int
		bytesPerFrame float64
		want          int64
	}{
		// 13.5 h × 44100 / 1024 = 2,093,027 frames; 2³¹ samples would allow 2,097,151
		{"44.1 kHz, time limit wins", 44100, 0, 2_093_027},
		// 13.5 h at 22.05 kHz is 1,046,513 frames, well inside 2³¹ samples
		{"22 kHz, time limit wins", 22050, 0, 1_046_513},
		// at 48 kHz 13.5 h is 2,278,125 frames, so the 32-bit sample limit wins
		{"48 kHz, sample limit wins", 48000, 0, 2_097_151},
		// ALAC at ~2,000 bytes per frame: 800 MB / 2000 = 400,000 frames, about 2.6 h
		{"ALAC, size limit wins", 44100, 2000, 400_000},
		// Audible AAC at ~364 bytes per frame: 800 MB allows ~2.2M frames, more than time
		{"AAC, size below time", 44100, 364, 2_093_027},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, partLimit(tt.sampleRate, tt.bytesPerFrame))
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

func TestMeaningfulSection(t *testing.T) {
	tests := []struct {
		name, want string
	}{
		{"Animal Farm", "Animal Farm"},
		{"  1984 ", "1984"}, // a number, but too long to be a disc count
		{"The Annals of the Kings and Rulers", "The Annals of the Kings and Rulers"},
		{"Civil", "Civil"},
		{"Discworld", "Discworld"},
		{"Part One", "Part One"}, // words aren't caught: kept, and harmless
		{"", ""},
		{"Disc 1", ""},
		{"disc 2 of 3", ""},
		{"CD2", ""},
		{"Disk 3/4", ""},
		{"Part II", ""},
		{"Vol. 2", ""},
		{"Volume 1", ""},
		{"Side A", ""},
		{"3", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, meaningfulSection(tt.name))
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
		{"22 kHz AAC to ALAC counts as PCM", download{sampleRate: 22050, chapters: []chapterFile{
			{codec: "aac", frames: 1000, bytes: 180_000},
		}}, 4096},
		{"no chapters is unknown", download{sampleRate: 44100}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.InDelta(t, tt.want, outputBytesPerFrame(tt.d), 0.001)
		})
	}
}
