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
		name       string
		sampleRate int
		want       int64
	}{
		{"44.1 kHz: 79 minutes", 44100, 204_134},
		{"22 kHz: 79 minutes", 22050, 102_067},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, partLimit(tt.sampleRate))
		})
	}
}
