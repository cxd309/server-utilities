package main

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVorbisCommentsGet(t *testing.T) {
	v := vorbisComments{
		"GENRE":        {"Rock", "Pop"},
		"ALBUM ARTIST": {"Fallback"},
		"EMPTY":        {"  "},
	}
	assert.Equal(t, "Rock; Pop", v.get("GENRE"))
	assert.Equal(t, "Fallback", v.get("ALBUMARTIST", "ALBUM ARTIST"))
	assert.Equal(t, "", v.get("EMPTY"))
	assert.Equal(t, "", v.get("MISSING"))
}

func TestVorbisCommentsNumberPair(t *testing.T) {
	tests := []struct {
		v    vorbisComments
		want string
	}{
		{vorbisComments{"TRACKNUMBER": {"3"}, "TRACKTOTAL": {"12"}}, "3/12"},
		{vorbisComments{"TRACKNUMBER": {"3"}, "TOTALTRACKS": {"12"}}, "3/12"},
		{vorbisComments{"TRACKNUMBER": {"3/12"}}, "3/12"},
		{vorbisComments{"TRACKNUMBER": {"3"}}, "3"},
		{vorbisComments{}, ""},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, tt.v.numberPair("TRACKNUMBER", "TRACKTOTAL", "TOTALTRACKS"), "%v", tt.v)
	}
}

func TestFromVorbis(t *testing.T) {
	got := fromVorbis(vorbisComments{
		"TITLE":        {"Have You Ever Been (to Electric Ladyland)"},
		"DATE":         {"2012-09-10"},
		"ORIGINALDATE": {"1968-10-16"},
		"TRACKNUMBER":  {"2"},
		"TOTALTRACKS":  {"16"},
		"COMPILATION":  {"1"},
	})
	assert.Equal(t, trackTags{
		Title:       "Have You Ever Been (to Electric Ladyland)",
		Year:        "1968",
		Track:       "2/16",
		Compilation: true,
	}, got)
}

func TestReadFLAC(t *testing.T) {
	requireFFmpeg(t)
	src := filepath.Join(t.TempDir(), "song.flac")
	makeFLAC(t, src,
		"title=Café Song", // lower case: field names are case-insensitive
		"ARTIST=The Band",
		"ALBUMARTIST=Various Artists",
		"TRACKNUMBER=3", "TRACKTOTAL=12",
		"DISCNUMBER=1", "DISCTOTAL=2",
	)

	got, err := readFLAC(src)
	require.NoError(t, err)
	assert.Equal(t, trackTags{
		Title:       "Café Song",
		Artist:      "The Band",
		AlbumArtist: "Various Artists",
		Track:       "3/12",
		Disc:        "1/2",
	}, got)
}

func TestReadFLACNotAFLAC(t *testing.T) {
	dir := t.TempDir()
	touch(t, dir, "empty.flac")

	_, err := readFLAC(filepath.Join(dir, "empty.flac"))
	assert.Error(t, err)
}
