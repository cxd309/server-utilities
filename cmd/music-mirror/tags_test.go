package main

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/bogem/id3v2/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVorbisTagsGet(t *testing.T) {
	tags := vorbisTags{
		"GENRE":        {"Rock", "Pop"},
		"ALBUM ARTIST": {"Fallback"},
		"EMPTY":        {"  "},
	}
	assert.Equal(t, "Rock; Pop", tags.get("GENRE"))
	assert.Equal(t, "Fallback", tags.get("ALBUMARTIST", "ALBUM ARTIST"))
	assert.Equal(t, "", tags.get("EMPTY"))
	assert.Equal(t, "", tags.get("MISSING"))
}

func TestVorbisTagsNumberPair(t *testing.T) {
	tests := []struct {
		tags vorbisTags
		want string
	}{
		{vorbisTags{"TRACKNUMBER": {"3"}, "TRACKTOTAL": {"12"}}, "3/12"},
		{vorbisTags{"TRACKNUMBER": {"3"}, "TOTALTRACKS": {"12"}}, "3/12"},
		{vorbisTags{"TRACKNUMBER": {"3/12"}}, "3/12"},
		{vorbisTags{"TRACKNUMBER": {"3"}}, "3"},
		{vorbisTags{}, ""},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, tt.tags.numberPair("TRACKNUMBER", "TRACKTOTAL", "TOTALTRACKS"), "%v", tt.tags)
	}
}

func TestVorbisTagsYear(t *testing.T) {
	assert.Equal(t, "1999", vorbisTags{"DATE": {"1999-03-04"}}.year())
	assert.Equal(t, "1999", vorbisTags{"DATE": {"1999"}}.year())
	assert.Equal(t, "1987", vorbisTags{"ORIGINALDATE": {"1987"}}.year())
	assert.Equal(t, "", vorbisTags{"DATE": {"n/a"}}.year())
}

func TestTagsRoundTrip(t *testing.T) {
	requireFFmpeg(t)
	dir := t.TempDir()
	src := filepath.Join(dir, "song.flac")
	mp3 := filepath.Join(dir, "song.mp3.part")

	makeFLAC(t, src,
		"TITLE=Café Song",
		"ARTIST=The Band",
		"ALBUMARTIST=Various Artists",
		"ALBUM=Compilation",
		"DATE=1999-03-04",
		"TRACKNUMBER=3", "TRACKTOTAL=12",
		"DISCNUMBER=1", "DISCTOTAL=2",
		"COMPILATION=1",
		"ARTISTSORT=Band, The",
	)

	tags, err := readFLACTags(src)
	require.NoError(t, err)
	require.NoError(t, encode(context.Background(), src, mp3, "320k"))
	require.NoError(t, writeID3(mp3, tags))

	tag, err := id3v2.Open(mp3, id3v2.Options{Parse: true})
	require.NoError(t, err)
	defer tag.Close()

	assert.Equal(t, byte(3), tag.Version())
	want := map[string]string{
		"TIT2": "Café Song",
		"TPE1": "The Band",
		"TPE2": "Various Artists",
		"TALB": "Compilation",
		"TYER": "1999",
		"TRCK": "3/12",
		"TPOS": "1/2",
		"TCMP": "1",
		"TSOP": "Band, The",
	}
	for frame, text := range want {
		assert.Equal(t, text, tag.GetTextFrame(frame).Text, frame)
	}
}
