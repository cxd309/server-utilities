package main

import (
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
	makePlainMP3(t, mp3)

	want := trackTags{
		Title: "Feel So Different", Artist: "Sinéad O’Connor", Album: "I Do Not Want What I Haven’t Got",
		AlbumArtist: "Sinéad O’Connor", Genre: "Rock; Pop", Composer: "Sinéad O’Connor",
		Cover: &cover{mime: "image/jpeg", data: []byte{0xff, 0xd8, 0xff, 0xe0, 1, 2, 3}},
		Year:  "1990", Track: "1/10", Disc: "1/1", Compilation: true,
	}
	require.NoError(t, writeID3(mp3, want))

	got, err := readMP3(mp3)
	require.NoError(t, err)
	assert.Equal(t, want, got, "every field should survive write then read")
}

func TestReadMP3PrimaryArtist(t *testing.T) {
	requireFFmpeg(t)
	dir := t.TempDir()
	mp3 := filepath.Join(dir, "song.mp3")
	makePlainMP3(t, mp3)

	tag, err := id3v2.Open(mp3, id3v2.Options{Parse: false})
	require.NoError(t, err)
	tag.SetVersion(4)
	tag.AddTextFrame("TIT2", id3v2.EncodingUTF8, "We Are Young")
	tag.AddTextFrame("TPE1", id3v2.EncodingUTF8, "Fun feat. Janelle Monáe")
	tag.AddUserDefinedTextFrame(id3v2.UserDefinedTextFrame{
		Encoding: id3v2.EncodingUTF8, Description: "ARTISTS", Value: "Fun\x00Janelle Monáe",
	})
	require.NoError(t, tag.Save())
	require.NoError(t, tag.Close())

	got, err := readMP3(mp3)
	require.NoError(t, err)
	assert.Equal(t, "Fun", got.Artist)
	assert.Equal(t, "We Are Young (feat. Janelle Monáe)", got.Title)
}
