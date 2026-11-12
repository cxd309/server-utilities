package main

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/cxd309/server-utilities/internal/ffmpeg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTagArgsSkipsEmpty(t *testing.T) {
	b := book{title: "Book", author: "", chapters: []chapter{{title: "One"}, {title: "Two"}}}
	args := tagArgs("in.m4a", "out.m4a", b, 2)

	assert.Contains(t, args, "title=Two")
	assert.Contains(t, args, "track=2/2")
	assert.NotContains(t, args, "artist=")
	assert.Equal(t, "out.m4a", args[len(args)-1])
}

func TestTagPiece(t *testing.T) {
	requireFFmpeg(t)
	dir := t.TempDir()
	src := filepath.Join(dir, "book.m4b")
	chapters := makeBook(t, src, "Opening Credits", "Chapter 1: Warriors in the Mist")
	require.NoError(t, splitAudio(context.Background(), src, dir, chapters))
	b := book{
		title: "A Wizard of Earthsea", author: "Ursula K. Le Guin",
		narrator: "Kobna Holdbrook-Smith", date: "2018", chapters: chapters,
	}

	require.NoError(t, tagPiece(context.Background(), dir, b, 2))

	out := filepath.Join(dir, "02 Chapter 1 Warriors in the Mist.m4a")
	info, err := ffmpeg.Probe(context.Background(), out)
	require.NoError(t, err)
	assert.Equal(t, map[string]string{
		"title":        "Chapter 1: Warriors in the Mist",
		"album":        "A Wizard of Earthsea",
		"artist":       "Ursula K. Le Guin",
		"album_artist": "Ursula K. Le Guin",
		"composer":     "Kobna Holdbrook-Smith",
		"date":         "2018",
		"track":        "2/2",
	}, withoutContainerTags(info.Format.Tags))

	assert.NoFileExists(t, filepath.Join(dir, "0001.m4a"), "untagged piece removed")
	assert.FileExists(t, filepath.Join(dir, "0000.m4a"), "other pieces untouched")
}

// withoutContainerTags drops the tags ffmpeg adds to every MP4 itself,
// leaving only the ones we set.
func withoutContainerTags(tags map[string]string) (ours map[string]string) {
	ours = map[string]string{}
	for k, v := range tags {
		switch k {
		case "major_brand", "minor_version", "compatible_brands", "encoder":
		default:
			ours[k] = v
		}
	}
	return ours
}

func TestTagPieceMissing(t *testing.T) {
	requireFFmpeg(t)
	b := book{chapters: []chapter{{title: "One"}}}
	err := tagPiece(context.Background(), t.TempDir(), b, 1)

	var fe *ffmpeg.Error
	assert.ErrorAs(t, err, &fe)
}
