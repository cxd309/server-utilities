package main

import (
	"strings"

	"github.com/bogem/id3v2/v2"
)

// readMP3 reads an MP3 file's ID3v2 tag (v2.3 or v2.4)
func readMP3(path string) (tags trackTags, err error) {
	var tag *id3v2.Tag
	if tag, err = id3v2.Open(path, id3v2.Options{Parse: true}); err != nil {
		return trackTags{}, err
	}
	defer tag.Close()
	return fromID3(tag), nil
}

func fromID3(tag *id3v2.Tag) (tags trackTags) {
	// text returns the first listed frame that is set. v2.4 separates
	// multiple values with NUL bytes, so split and re-join them.
	text := func(ids ...string) (value string) {
		for _, id := range ids {
			if value = joinValues(strings.Split(tag.GetTextFrame(id).Text, "\x00")); value != "" {
				return value
			}
		}
		return ""
	}

	return trackTags{
		Title:           text("TIT2"),
		Artist:          text("TPE1"),
		Album:           text("TALB"),
		AlbumArtist:     text("TPE2"),
		Genre:           text("TCON"),
		Composer:        text("TCOM"),
		ArtistSort:      text("TSOP"),
		AlbumArtistSort: text("TSO2"),
		AlbumSort:       text("TSOA"),
		TitleSort:       text("TSOT"),
		// Original release (v2.4 TDOR, v2.3 TORY), then recording/release date.
		Year:        yearOf(text("TDOR", "TORY"), text("TDRC", "TYER")),
		Track:       text("TRCK"),
		Disc:        text("TPOS"),
		Compilation: text("TCMP") == "1",
	}
}
