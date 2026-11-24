package main

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEpisodeName(t *testing.T) {
	tests := []struct{ path, want string }{
		{"/x/01-01 Abu Dhabi, TR 1.flac", "Abu Dhabi"},
		{"/x/13-10 Zurich, Part 1, TR 1.flac", "Zurich, Part 1"},
		{"/x/14-21 Interview with John Finnemore, TR 12.flac", "Interview with John Finnemore"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got, err := episodeName(tt.path)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}

	for _, bad := range []string{"/x/01-01 Abu Dhabi.flac", "/x/cover.flac", "/x/01-01 Abu Dhabi, TR one.flac"} {
		_, err := episodeName(bad)
		assert.ErrorIs(t, err, errNoTrackMark, bad)
	}
}

func TestGroupTracks(t *testing.T) {
	groups, err := groupTracks([]string{
		"01-01 Abu Dhabi, TR 1.flac", "01-02 Abu Dhabi, TR 2.flac",
		"01-10 Boston, TR 1.flac",
	})
	require.NoError(t, err)
	require.Len(t, groups, 2)
	assert.Equal(t, "Abu Dhabi", groups[0].name)
	assert.Len(t, groups[0].tracks, 2)
	assert.Equal(t, "Boston", groups[1].name)
}

func TestPair(t *testing.T) {
	groups := []group{{name: "Molokai"}, {name: "Yverdon-les-bains"}}
	assert.NoError(t, pair(groups, []episode{
		{title: "Molokai (Cabin Pressure at Christmas)"}, {title: "Yverdon-Les-Bains"},
	}), "contained, ignoring case")

	assert.ErrorIs(t, pair(groups, []episode{{title: "Molokai"}}), errEpisodeCount)

	var me *mismatchError
	assert.ErrorAs(t, pair(groups, []episode{{title: "Molokai"}, {title: "Zurich"}}), &me)
}

func TestEmbeddedCSV(t *testing.T) {
	episodes, err := readEpisodes(strings.NewReader(episodesCSV))
	require.NoError(t, err)
	require.Len(t, episodes, 28)
	for _, ep := range episodes {
		_, err := isoUTC(ep.releaseDate)
		assert.NoError(t, err, ep.title)
	}
}

func TestIsoUTC(t *testing.T) {
	got, err := isoUTC("2008-07-02T11:30:00+01:00")
	require.NoError(t, err)
	assert.Equal(t, "2008-07-02T10:30:00Z", got)
}
