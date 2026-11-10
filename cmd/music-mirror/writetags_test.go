package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/bogem/id3v2/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteID3ReplacesExistingTag(t *testing.T) {
	requireFFmpeg(t)
	mp3 := makeMP3(t, t.TempDir())
	before, err := os.Stat(mp3)
	require.NoError(t, err)

	require.NoError(t, writeID3(mp3, trackTags{Title: "Café Song", Year: "1967", Compilation: true}))

	tag, err := id3v2.Open(mp3, id3v2.Options{Parse: true})
	require.NoError(t, err)
	defer tag.Close()

	assert.Equal(t, byte(3), tag.Version())
	assert.Equal(t, "Café Song", tag.GetTextFrame("TIT2").Text)
	assert.Equal(t, "1967", tag.GetTextFrame("TYER").Text)
	assert.Equal(t, "1", tag.GetTextFrame("TCMP").Text)
	assert.Empty(t, tag.GetFrames("TPE1"), "unset fields should not be written")
	assert.Empty(t, tag.GetFrames("APIC"), "cover art should be gone")
	assert.Empty(t, tag.GetFrames("TXXX"), "old frames should be gone")

	after, err := os.Stat(mp3)
	require.NoError(t, err)
	assert.Less(t, after.Size(), before.Size()-40_000, "dropping the art should shrink the file")
}

func TestWriteID3RoundTrip(t *testing.T) {
	requireFFmpeg(t)
	dir := t.TempDir()
	mp3 := filepath.Join(dir, "song.mp3")
	require.NoError(t, encode(context.Background(), makeTone(t, dir), mp3, "128k"))

	want := trackTags{
		Title: "Feel So Different", Artist: "Sinéad O’Connor", Album: "I Do Not Want What I Haven’t Got",
		AlbumArtist: "Sinéad O’Connor", Genre: "Rock; Pop", Composer: "Sinéad O’Connor",
		ArtistSort: "O’Connor, Sinéad", AlbumArtistSort: "O’Connor, Sinéad",
		AlbumSort: "I Do Not Want", TitleSort: "Feel So Different",
		Year: "1990", Track: "1/10", Disc: "1/1", Compilation: true,
	}
	require.NoError(t, writeID3(mp3, want))

	got, err := readMP3(mp3)
	require.NoError(t, err)
	assert.Equal(t, want, got, "every field should survive write then read")
}

func TestFLACToTaggedMP3(t *testing.T) {
	requireFFmpeg(t)
	dir := t.TempDir()
	src := filepath.Join(dir, "song.flac")
	mp3 := filepath.Join(dir, "song.mp3.part")
	makeFLAC(t, src,
		"TITLE=Café Song",
		"ARTIST=The Band",
		"ALBUMARTIST=Various Artists",
		"DATE=2012-09-10",
		"ORIGINALDATE=1968-10-16",
		"TRACKNUMBER=3", "TRACKTOTAL=12",
		"COMPILATION=1",
	)

	tags, err := readFLAC(src)
	require.NoError(t, err)
	require.NoError(t, encode(context.Background(), src, mp3, "320k"))
	require.NoError(t, writeID3(mp3, tags))

	got, err := readMP3(mp3)
	require.NoError(t, err)
	assert.Equal(t, trackTags{
		Title:       "Café Song",
		Artist:      "The Band",
		AlbumArtist: "Various Artists",
		Year:        "1968",
		Track:       "3/12",
		Compilation: true,
	}, got)
}
