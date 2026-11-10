package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseTracks(t *testing.T) {
	out := "A1B2C3\t/Users/chris/Music/music-mirror/A/01.mp3\n" +
		"D4E5F6\t\n" + // file missing
		"\n" +
		"G7H8I9\t/path/with\ttab.mp3"

	got, err := parseTracks(out)
	require.NoError(t, err)
	assert.Equal(t, []libraryTrack{
		{id: "A1B2C3", path: "/Users/chris/Music/music-mirror/A/01.mp3"},
		{id: "D4E5F6", path: ""},
		{id: "G7H8I9", path: "/path/with\ttab.mp3"},
	}, got)
}

func TestParseTracksEmpty(t *testing.T) {
	got, err := parseTracks("")
	require.NoError(t, err)
	assert.Empty(t, got)
}

func TestParseTracksBadLine(t *testing.T) {
	_, err := parseTracks("no tab here")
	var le *listingError
	require.ErrorAs(t, err, &le)
	assert.Equal(t, "no tab here", le.Line)
}
