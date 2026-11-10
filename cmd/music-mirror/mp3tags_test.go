package main

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/bogem/id3v2/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// writeV24 gives path a Picard-style ID3v2.4 tag, with cover art and a
// TXXX frame, to stand in for the MP3s in the library.
func writeV24(t *testing.T, path string) {
	t.Helper()
	tag, err := id3v2.Open(path, id3v2.Options{Parse: false})
	require.NoError(t, err)
	defer tag.Close()

	tag.SetVersion(4)
	enc := id3v2.EncodingUTF8
	for id, text := range map[string]string{
		"TIT2": "Jingle Bells",
		"TPE1": "The Torero Band",
		"TPE2": "The Torero Band",
		"TALB": "Tijuana Christmas",
		"TDRC": "2008-11-03", // reissue
		"TDOR": "1967",       // original release
		"TRCK": "1/12",
		"TCON": "Jazz\x00Christmas", // v2.4 multi-value
	} {
		tag.AddTextFrame(id, enc, text)
	}
	tag.AddAttachedPicture(id3v2.PictureFrame{
		Encoding: enc, MimeType: "image/jpeg", PictureType: id3v2.PTFrontCover,
		Picture: make([]byte, 50_000),
	})
	tag.AddUserDefinedTextFrame(id3v2.UserDefinedTextFrame{
		Encoding: enc, Description: "MusicBrainz Album Id", Value: "abc",
	})
	require.NoError(t, tag.Save())
}

// makeMP3 writes a short MP3 with a Picard-style v2.4 tag into dir.
func makeMP3(t *testing.T, dir string) (path string) {
	t.Helper()
	path = filepath.Join(dir, "song.mp3")
	require.NoError(t, encode(context.Background(), makeTone(t, dir), path, "128k"))
	writeV24(t, path)
	return path
}

func TestReadMP3(t *testing.T) {
	requireFFmpeg(t)
	src := makeMP3(t, t.TempDir())

	got, err := readMP3(src)
	require.NoError(t, err)
	assert.Equal(t, trackTags{
		Title:       "Jingle Bells",
		Artist:      "The Torero Band",
		AlbumArtist: "The Torero Band",
		Album:       "Tijuana Christmas",
		Genre:       "Jazz; Christmas",
		Year:        "1967", // TDOR beats TDRC
		Track:       "1/12",
	}, got)
}

func TestReadMP3Untagged(t *testing.T) {
	requireFFmpeg(t)
	dir := t.TempDir()
	src := filepath.Join(dir, "plain.mp3")
	require.NoError(t, encode(context.Background(), makeTone(t, dir), src, "128k"))

	got, err := readMP3(src)
	require.NoError(t, err)
	assert.Equal(t, trackTags{}, got)
}
