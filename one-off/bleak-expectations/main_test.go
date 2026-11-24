package main

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChapterNumber(t *testing.T) {
	n, err := chapterNumber("/x/07 Chapter 7.m4a")
	require.NoError(t, err)
	assert.Equal(t, 7, n)

	for _, bad := range []string{"/x/Chapter 7.m4a", "/x/07 Episode 7.m4a", "/x/cover.m4a"} {
		_, err = chapterNumber(bad)
		assert.ErrorIs(t, err, errNotChapter, bad)
	}
}

func TestCheckOrder(t *testing.T) {
	assert.NoError(t, checkOrder([]string{"01 Chapter 1.m4a", "02 Chapter 2.m4a"}))

	var oe *orderError
	require.ErrorAs(t, checkOrder([]string{"01 Chapter 1.m4a", "02 Chapter 3.m4a"}), &oe)
	assert.Equal(t, 2, oe.Want)
	assert.Equal(t, 3, oe.Got)
}

func TestEmbeddedCSV(t *testing.T) {
	episodes, err := readEpisodes(strings.NewReader(episodesCSV))
	require.NoError(t, err)
	require.Len(t, episodes, 30)
	assert.Equal(t, "A Childhood Cruelly Kippered", episodes[0].title)
	for _, ep := range episodes {
		_, err := isoUTC(ep.releaseDate)
		assert.NoError(t, err, ep.title)
	}
}

func TestCopyArgsCover(t *testing.T) {
	ep := episode{title: "T", description: "D"}
	with := strings.Join(copyArgs("in.m4a", "cover.jpg", "out", ep, "2007-08-15T10:30:00Z", 1, 30), " ")
	assert.Contains(t, with, "-i cover.jpg -map 0:a:0 -map 1:v:0")

	without := strings.Join(copyArgs("in.m4a", "", "out", ep, "2007-08-15T10:30:00Z", 1, 30), " ")
	assert.NotContains(t, without, "1:v:0")
}
